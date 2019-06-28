package httputil

import (
	"fmt"
	"net/http"
)

// Errorf sets the response status code and write formats the provided message as the
// response body
func Errorf(w http.ResponseWriter, statusCode int, msg string, args ...interface{}) {
	http.Error(w, fmt.Sprintf(msg, args...), statusCode)
}
