package function

import (
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

// Handle a serverless request
func Handle(req []byte) string {

	key := os.Getenv("Http_X_Api_Key") // converted via the Header: X-Api-Key

	secretBytes, err := getAPISecret("secret_api_key") // You must create a secret ahead of time named `secret_api_key`
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
