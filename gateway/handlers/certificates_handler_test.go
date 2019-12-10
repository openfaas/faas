package handlers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
)

func TestMakeCertificatesHandler(t *testing.T) {
	tests := []struct {
		name   string
		keyID  string
		eval   func(r *http.Response) error
		reader SecretsReader
	}{
		{
			name:  "Get certificate that exists",
			keyID: "callback",
			eval: func(r *http.Response) error {
				if r.StatusCode != 200 {
					return fmt.Errorf("expected 200")
				}

				body, _ := ioutil.ReadAll(r.Body)
				keyData := &KeyType{}
				if err := json.Unmarshal(body, keyData); err != nil {
					return fmt.Errorf("error unmarshalling result")
				}

				if keyData.PEM != testPublicKey {
					return fmt.Errorf("PEM want %s got %s", testPublicKey, keyData.PEM)
				}
				return nil
			},
			reader: &TestSecretsReader{readCallBack: func(key string) (s string, e error) {
				return testPublicKey, nil
			}},
		},
		{

			name:  "Attempt to get certificate that doesn't exist",
			keyID: "missingID",
			eval: func(r *http.Response) error {
				if r.StatusCode != 404 {
					return fmt.Errorf("expected 404")
				}

				return nil
			},
			reader: &TestSecretsReader{readCallBack: func(key string) (s string, e error) {
				return "", fmt.Errorf("key not found")
			}},
		},
	}

	r := mux.NewRouter()
	ts := httptest.NewServer(r)
	defer ts.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r.HandleFunc("/certificates/{id}", MakeCertificatesHandler(tt.reader))

			url := ts.URL + "/certificates/" + tt.keyID
			resp, err := http.Get(url)
			if err != nil {
				t.Errorf("MakeCertificatesHandler() call = %v", err)
			}

			if err := tt.eval(resp); err != nil {
				t.Errorf("MakeCertificatesHandler() eval = %v", err)
			}
		})
	}
}

type TestSecretsReader struct {
	readCallBack func(key string) (string, error)
}

func (r *TestSecretsReader) Read(key string) (string, error) {
	return r.readCallBack(key)
}

const testPublicKey = `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA7pEUKQ28pI5N3g/zG6OJ
100N/DV2Q8Ob+gzRjd7HjXgVgZyjS3nA8FAYrxTLSihcIhXuQrYxyk2vp6YMNmSB
fOptkdmj4UgLYskfeqEt8JjS6ExBxSWEDgr1IXOPPDP61on8F65/ZYGnp2JF2wHY
k0OeD4ppNUV+mIHj/wXf7VLHGflwFQH/+mfUn+tVQRgX7hTadcYmGJ+1XP0py4kU
gJDHfw8eBsFurHWr2mXu3BdraSKKf1G9i+SifmOUUul6mBONmlvzQdKtDCr48o1H
QndRHcMWjKhlBhKz4qrmqku8oGBh6iHhGVVYf8D3mU1nzyjH4rOUXZwzj+SaqgGk
vQIDAQAB
-----END PUBLIC KEY-----`
