package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

func handle(body []byte) {
	key := os.Getenv("Http_X_Api_Key")

	secretBytes, err := ioutil.ReadFile("/run/secrets/secret_api_key")
	if err != nil {
		log.Fatal(err)
	}

	secret := strings.TrimSpace(string(secretBytes))

	if key == secret {
		fmt.Println("Unlocked the function!")
	} else {
		fmt.Println("Access denied!")
	}
}

func main() {
	bytes, _ := ioutil.ReadAll(os.Stdin)
	handle(bytes)
}
