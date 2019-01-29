package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/openfaas/faas-provider/auth"
)

//AddBasicAuth to a request by reading secrets
func AddBasicAuth(req *http.Request) error {
	if os.Getenv("basic_auth") == "true" {
		reader := auth.ReadBasicAuthFromDisk{}

		if len(os.Getenv("secret_mount_path")) > 0 {
			reader.SecretMountPath = os.Getenv("secret_mount_path")
		}

		credentials, err := reader.Read()
		if err != nil {
			return fmt.Errorf("Unable to read basic auth: %s", err.Error())
		}

		req.SetBasicAuth(credentials.User, credentials.Password)
	}
	return nil
}

//LoadCredentials load credentials from dis
func LoadCredentials() (*auth.BasicAuthCredentials, error) {
	reader := auth.ReadBasicAuthFromDisk{}

	if len(os.Getenv("secret_mount_path")) > 0 {
		reader.SecretMountPath = os.Getenv("secret_mount_path")
	}

	credentials, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("Unable to read basic auth: %s", err.Error())
	}
	return credentials, nil
}
