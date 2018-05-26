package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

func getAPISecret(secretName string) (secretBytes []byte, err error) {
	// read from the openfaas secrets folder
	secretBytes, err = ioutil.ReadFile("/var/openfaas/secrets/" + secretName)
	if err != nil {
		// read from the original location for backwards compatibility with openfaas <= 0.8.2
		secretBytes, err = ioutil.ReadFile("/run/secrets/" + secretName)
	}

	return secretBytes, err
}

func handle(body []byte) {
	key := os.Getenv("Http_X_Api_Key")

	secretBytes, err := getAPISecret("secret_api_key")
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
