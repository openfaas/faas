package function

import (
	"io/ioutil"
	"log"
	"os"
	"strings"
)

// Handle a serverless request
func Handle(req []byte) string {

	key := os.Getenv("Http_X_Api_Key") // converted via the Header: X-Api-Key

	secretBytes, err := ioutil.ReadFile("/run/secrets/secret_api_key") // You must create a secret ahead of time named `secret_api_key`
	if err != nil {
		log.Fatal(err)
	}

	secret := strings.TrimSpace(string(secretBytes))

	message := "Access was denied."
	if key == secret {
		message = "You unlocked the function."
	}

	return message
}
