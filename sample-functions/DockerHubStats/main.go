package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

type dockerHubStatsType struct {
	Count int `json:"count"`
}

func sanitizeInput(input string) string {
	parts := strings.Split(input, "\n")
	return strings.Trim(parts[0], " ")
}

func requestHubStats(org string) dockerHubStatsType {
	client := http.Client{}
	res, err := client.Get("https://hub.docker.com/v2/repositories/" + org)
	if err != nil {
		log.Fatalln("Unable to reach Docker Hub server.")
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatalln("Unable to parse response from server.")
	}

	dockerHubStats := dockerHubStatsType{}
	json.Unmarshal(body, &dockerHubStats)
	return dockerHubStats
}

func main() {
	input, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.Fatal("Unable to read standard input:", err)
	}
	org := string(input)
	if len(input) == 0 {
		log.Fatalln("A username or organisation is required.")
	}

	org = sanitizeInput(org)
	dockerHubStats := requestHubStats(org)

	fmt.Printf("The organisation or user %s has %d repositories on the Docker hub.\n", org, dockerHubStats.Count)
}
