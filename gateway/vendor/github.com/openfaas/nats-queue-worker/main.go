package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"time"

	"net/http"

	"github.com/nats-io/go-nats-streaming"
	"github.com/openfaas/faas-provider/auth"
	"github.com/openfaas/faas/gateway/queue"
	"github.com/openfaas/nats-queue-worker/nats"
)

// AsyncReport is the report from a function executed on a queue worker.
type AsyncReport struct {
	FunctionName string  `json:"name"`
	StatusCode   int     `json:"statusCode"`
	TimeTaken    float64 `json:"timeTaken"`
}

func printMsg(m *stan.Msg, i int) {
	log.Printf("[#%d] Received on [%s]: '%s'\n", i, m.Subject, m)
}

func makeClient() http.Client {
	proxyClient := http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 0,
			}).DialContext,
			MaxIdleConns:          1,
			DisableKeepAlives:     true,
			IdleConnTimeout:       120 * time.Millisecond,
			ExpectContinueTimeout: 1500 * time.Millisecond,
		},
	}
	return proxyClient
}

func main() {
	readConfig := ReadConfig{}
	config := readConfig.Read()
	log.SetFlags(0)

	clusterID := "faas-cluster"
	val, _ := os.Hostname()
	clientID := "faas-worker-" + nats.GetClientID(val)

	var durable string
	var qgroup string
	var unsubscribe bool
	var credentials *auth.BasicAuthCredentials
	var err error

	if os.Getenv("basic_auth") == "true" {
		log.Printf("Loading basic authentication credentials")
		credentials, err = LoadCredentials()
		if err != nil {
			log.Printf("Error with LoadCredentials: %s ", err.Error())
		}
	}

	client := makeClient()
	sc, err := stan.Connect(clusterID, clientID, stan.NatsURL("nats://"+config.NatsAddress+":4222"))
	if err != nil {
		log.Fatalf("Can't connect to %s: %v\n", "nats://"+config.NatsAddress+":4222", err)
	}

	startOpt := stan.StartWithLastReceived()

	i := 0
	mcb := func(msg *stan.Msg) {
		i++

		printMsg(msg, i)

		started := time.Now()

		req := queue.Request{}
		unmarshalErr := json.Unmarshal(msg.Data, &req)

		if unmarshalErr != nil {
			log.Printf("Unmarshal error: %s with data %s", unmarshalErr, msg.Data)
			return
		}

		xCallID := req.Header.Get("X-Call-Id")

		fmt.Printf("Request for %s.\n", req.Function)

		if config.DebugPrintBody {
			fmt.Println(string(req.Body))
		}

		queryString := ""
		if len(req.QueryString) > 0 {
			queryString = fmt.Sprintf("?%s", strings.TrimLeft(req.QueryString, "?"))
		}

		functionURL := fmt.Sprintf("http://%s%s:8080/%s", req.Function, config.FunctionSuffix, queryString)

		request, err := http.NewRequest(http.MethodPost, functionURL, bytes.NewReader(req.Body))
		defer request.Body.Close()

		copyHeaders(request.Header, &req.Header)

		res, err := client.Do(request)
		var status int
		var functionResult []byte

		if err != nil {
			status = http.StatusServiceUnavailable

			log.Println(err)
			timeTaken := time.Since(started).Seconds()

			if req.CallbackURL != nil {
				log.Printf("Callback to: %s\n", req.CallbackURL.String())

				resultStatusCode, resultErr := postResult(&client,
					res,
					functionResult,
					req.CallbackURL.String(),
					xCallID,
					status)
				if resultErr != nil {
					log.Println(resultErr)
				} else {
					log.Printf("Posted result: %d", resultStatusCode)
				}
			}

			statusCode, reportErr := postReport(&client, req.Function, status, timeTaken, config.GatewayAddress, credentials)
			if reportErr != nil {
				log.Println(reportErr)
			} else {
				log.Printf("Posting report - %d\n", statusCode)
			}
			return
		}

		if res.Body != nil {
			defer res.Body.Close()

			resData, err := ioutil.ReadAll(res.Body)
			functionResult = resData

			if err != nil {
				log.Println(err)
			}

			if config.WriteDebug {
				fmt.Println(string(functionResult))
			} else {
				fmt.Printf("Wrote %d Bytes\n", len(string(functionResult)))
			}
		}

		timeTaken := time.Since(started).Seconds()

		fmt.Println(res.Status)

		if req.CallbackURL != nil {
			log.Printf("Callback to: %s\n", req.CallbackURL.String())
			resultStatusCode, resultErr := postResult(&client,
				res,
				functionResult,
				req.CallbackURL.String(),
				xCallID,
				res.StatusCode)
			if resultErr != nil {
				log.Println(resultErr)
			} else {
				log.Printf("Posted result: %d", resultStatusCode)
			}
		}

		statusCode, reportErr := postReport(&client, req.Function, res.StatusCode, timeTaken, config.GatewayAddress, credentials)

		if reportErr != nil {
			log.Println(reportErr)
		} else {
			log.Printf("Posting report - %d\n", statusCode)
		}
	}

	subj := "faas-request"
	qgroup = "faas"

	log.Println("Wait for ", config.AckWait)
	sub, err := sc.QueueSubscribe(subj, qgroup, mcb, startOpt, stan.DurableName(durable), stan.MaxInflight(config.MaxInflight), stan.AckWait(config.AckWait))
	if err != nil {
		log.Panicln(err)
	}

	log.Printf("Listening on [%s], clientID=[%s], qgroup=[%s] durable=[%s]\n", subj, clientID, qgroup, durable)

	// Wait for a SIGINT (perhaps triggered by user with CTRL-C)
	// Run cleanup when signal is received
	signalChan := make(chan os.Signal, 1)
	cleanupDone := make(chan bool)
	signal.Notify(signalChan, os.Interrupt)
	go func() {
		for range signalChan {
			fmt.Printf("\nReceived an interrupt, unsubscribing and closing connection...\n\n")
			// Do not unsubscribe a durable on exit, except if asked to.
			if durable == "" || unsubscribe {
				sub.Unsubscribe()
			}
			sc.Close()
			cleanupDone <- true
		}
	}()
	<-cleanupDone
}

func postResult(client *http.Client, functionRes *http.Response, result []byte, callbackURL string, xCallID string, statusCode int) (int, error) {
	var reader io.Reader

	if result != nil {
		reader = bytes.NewReader(result)
	}

	request, err := http.NewRequest(http.MethodPost, callbackURL, reader)

	if functionRes != nil {
		copyHeaders(request.Header, &functionRes.Header)
	}

	request.Header.Set("X-Function-Status", fmt.Sprintf("%d", statusCode))

	if len(xCallID) > 0 {
		request.Header.Set("X-Call-Id", xCallID)
	}

	res, err := client.Do(request)

	if err != nil {
		return http.StatusBadGateway, fmt.Errorf("error posting result to URL %s %s", callbackURL, err.Error())
	}

	if request.Body != nil {
		defer request.Body.Close()
	}

	if res.Body != nil {
		defer res.Body.Close()
	}
	return res.StatusCode, nil
}

func copyHeaders(destination http.Header, source *http.Header) {
	for k, v := range *source {
		vClone := make([]string, len(v))
		copy(vClone, v)
		(destination)[k] = vClone
	}
}

func postReport(client *http.Client, function string, statusCode int, timeTaken float64, gatewayAddress string, credentials *auth.BasicAuthCredentials) (int, error) {
	req := AsyncReport{
		FunctionName: function,
		StatusCode:   statusCode,
		TimeTaken:    timeTaken,
	}

	targetPostback := "http://" + gatewayAddress + ":8080/system/async-report"
	reqBytes, _ := json.Marshal(req)
	request, err := http.NewRequest(http.MethodPost, targetPostback, bytes.NewReader(reqBytes))

	if os.Getenv("basic_auth") == "true" && credentials != nil {
		request.SetBasicAuth(credentials.User, credentials.Password)
	}

	defer request.Body.Close()

	res, err := client.Do(request)

	if err != nil {
		return http.StatusGatewayTimeout, fmt.Errorf("cannot post report to %s: %s", targetPostback, err)
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	return res.StatusCode, nil
}
