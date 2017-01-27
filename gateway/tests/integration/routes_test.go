package inttests

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"testing"
	"time"
)

// Before running these tests do a Docker stack deploy.

func fireRequest(url string, method string, reqBody string) (string, int, error) {
	return fireRequestWithHeader(url, method, reqBody, "")
}

func fireRequestWithHeader(url string, method string, reqBody string, xheader string) (string, int, error) {
	httpClient := http.Client{
		Timeout: time.Second * 2, // Maximum of 2 secs
	}

	req, err := http.NewRequest(method, url, bytes.NewBufferString(reqBody))
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("User-Agent", "spacecount-tutorial")
	if len(xheader) != 0 {
		req.Header.Set("X-Function", xheader)
	}
	res, getErr := httpClient.Do(req)
	if getErr != nil {
		log.Fatal(getErr)
	}

	body, readErr := ioutil.ReadAll(res.Body)
	defer req.Body.Close()
	if readErr != nil {
		log.Fatal(readErr)
	}
	return string(body), res.StatusCode, readErr
}

func Test_Get_Rejected(t *testing.T) {
	var reqBody string
	_, code, err := fireRequest("http://localhost:8080/function/func_echoit", http.MethodGet, reqBody)
	if code != http.StatusInternalServerError {
		t.Log("Failed")
	}

	if err != nil {
		t.Log(err)
		t.Fail()
	}

}

func Test_EchoIt_Post_Route_Handler(t *testing.T) {
	reqBody := "test message"
	body, code, err := fireRequest("http://localhost:8080/function/func_echoit", http.MethodPost, reqBody)

	if err != nil {
		t.Log(err)
		t.Fail()
	}
	if code != http.StatusOK {
		t.Log("Failed")
	}
	if body != reqBody {
		t.Log("Expected body returned")
		t.Fail()
	}
}

func Test_EchoIt_Post_Header_Handler(t *testing.T) {
	reqBody := "test message"
	body, code, err := fireRequestWithHeader("http://localhost:8080/", http.MethodPost, reqBody, "func_echoit")

	if err != nil {
		t.Log(err)
		t.Fail()
	}
	if code != http.StatusOK {
		t.Log("Failed")
	}
	if body != reqBody {
		t.Log("Expected body returned")
		t.Fail()
	}

}
