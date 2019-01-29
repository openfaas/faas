package nats

import "regexp"

var supportedCharacters = regexp.MustCompile("[^a-zA-Z0-9-_]+")
func GetClientID(value string) string {
	return supportedCharacters.ReplaceAllString(value, "_")
}