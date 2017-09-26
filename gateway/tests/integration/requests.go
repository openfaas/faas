package inttests

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

func fireRequest(url string, method string, reqBody string) (string, int) {
	headers := make(map[string]string)
	return fireRequestWithHeaders(url, method, reqBody, headers)
}

func fireRequestWithHeaders(url string, method string, reqBody string, headers map[string]string) (string, int) {
	httpClient := http.Client{
		Timeout: time.Second * 2, // Maximum of 2 secs
	}

	req, err := http.NewRequest(method, url, bytes.NewBufferString(reqBody))
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("User-Agent", "go-integration")
	for kk, vv := range headers {
		req.Header.Set(kk, vv)
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

	return string(body), res.StatusCode
}
