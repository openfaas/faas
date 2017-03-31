package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/alexellis/faas/watchdog/types"
)

func handle(header http.Header, body []byte) {
	key := header.Get("X-Api-Key")
	if key == os.Getenv("secret_api_key") {
		fmt.Println("Unlocked the function!")
	} else {
		fmt.Println("Access denied!")
	}
}

func main() {
	bytes, _ := ioutil.ReadAll(os.Stdin)
	req, err := types.UnmarshalRequest(bytes)
	if err != nil {
		log.Fatal(err)
	}
	handle(req.Header, req.Body.Raw)
}
