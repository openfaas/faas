package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"time"

	"github.com/openfaas/faas-provider/auth"
	"github.com/pkg/errors"
)

func main() {
	port := 8080

	if val, ok := os.LookupEnv("port"); ok {
		intOut, err := strconv.Atoi(val)
		if err != nil {
			panic(errors.Wrap(err, fmt.Sprintf("value of `port`: %s, not a valid port", val)))
		}
		port = intOut
	}

	s := &http.Server{
		Addr:           fmt.Sprintf(":%d", port),
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   5 * time.Second,
		MaxHeaderBytes: 1 << 20, // Max header of 1MB
	}

	credentialsReader := auth.ReadBasicAuthFromDisk{
		SecretMountPath:  os.Getenv("secret_mount_path"),
		UserFilename:     os.Getenv("user_filename"),
		PasswordFilename: os.Getenv("pass_filename"),
	}

	credentials, err := credentialsReader.Read()
	if err != nil {
		panic(errors.Wrap(err, "unable to read basic auth credentials, check `secret_mount_path`"))
	}

	authHandler := auth.DecorateWithBasicAuth(func(w http.ResponseWriter, r *http.Request) {
	}, credentials)
	http.HandleFunc("/validate", makeLogger(authHandler))

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	log.Printf("Listening on: %d\n", port)
	log.Fatal(s.ListenAndServe())
}

func makeLogger(next http.Handler) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		rr := httptest.NewRecorder()

		next.ServeHTTP(rr, r)
		log.Printf("Validated request %d.\n", rr.Code)

		resHeader := rr.Header()
		copyHeaders(w.Header(), &resHeader)

		w.WriteHeader(rr.Code)
		if rr.Body != nil {
			w.Write(rr.Body.Bytes())
		}
	}
}

func copyHeaders(destination http.Header, source *http.Header) {
	for k, v := range *source {
		vClone := make([]string, len(v))
		copy(vClone, v)
		(destination)[k] = vClone
	}
}
