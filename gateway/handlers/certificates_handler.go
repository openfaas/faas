package handlers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"

	"github.com/gorilla/mux"
)

var keyMap = map[string]string{
	"callback": "http-signing-public-key",
}

type KeyType struct {
	Id  string `json:"id"`
	PEM string `json:"pem"`
}

type SecretsReader interface {
	Read(key string) (string, error)
}

type FileSecretsReader struct {
	SecretMountPath string
}

func (f *FileSecretsReader) Read(key string) (string, error) {
	if len(f.SecretMountPath) == 0 {
		return "", fmt.Errorf("invalid SecretMountPath specified for reading secrets used for certificates")
	}

	certificatePath := path.Join(f.SecretMountPath, key)
	if _, err := os.Stat(certificatePath); os.IsNotExist(err) {
		return "", fmt.Errorf("unable to find secret %s", key)
	}

	value, err := ioutil.ReadFile(certificatePath)
	if err != nil {
		return "", fmt.Errorf("error reading find secret %s", certificatePath)
	}

	return string(value), nil
}

func MakeCertificatesHandler(reader SecretsReader) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		keyID := vars["id"]
		secretKey, ok := keyMap[keyID]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			message := fmt.Sprintf("Unable to find certificate %s.", keyID)
			w.Write([]byte(message))
			log.Println(message)
			return
		}

		publicKey, err := reader.Read(secretKey)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			message := fmt.Sprintf("Unable to find secret for key %s.", keyID)
			w.Write([]byte(message))
			log.Println(message)
			return
		}

		key := &KeyType{
			Id:  secretKey,
			PEM: publicKey,
		}

		bytesOut, marshalErr := json.Marshal(key)
		if marshalErr != nil {
			w.WriteHeader(http.StatusInternalServerError)
			message := fmt.Sprintf("error marshalling json for key %s. %v", keyID, err)
			w.Write([]byte(message))
			log.Println(message)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(bytesOut)
	}
}
