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

type dockerHubOrgStatsType struct {
	Count int `json:"count"`
}

type dockerHubRepoStatsType struct {
	PullCount int `json:"pull_count"`
}

func sanitizeInput(input string) string {
	parts := strings.Split(input, "\n")
	return strings.Trim(parts[0], " ")
}

func requestStats(repo string) []byte {
	client := http.Client{}
	res, err := client.Get("https://hub.docker.com/v2/repositories/" + repo)
	if err != nil {
		log.Fatalln("Unable to reach Docker Hub server.")
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatalln("Unable to parse response from server.")
	}

	return body
}

func parseOrgStats(response []byte) (dockerHubOrgStatsType, error) {
	dockerHubOrgStats := dockerHubOrgStatsType{}
	err := json.Unmarshal(response, &dockerHubOrgStats)
	return dockerHubOrgStats, err
}

func parseRepoStats(response []byte) (dockerHubRepoStatsType, error) {
	dockerHubRepoStats := dockerHubRepoStatsType{}
	err := json.Unmarshal(response, &dockerHubRepoStats)
	return dockerHubRepoStats, err
}

func main() {
	input, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.Fatal("Unable to read standard input:", err)
	}
	request := string(input)
	if len(input) == 0 {
		log.Fatalln("A username/organisation or repository is required.")
	}

	request = sanitizeInput(request)
	response := requestStats(request)

	if strings.Contains(request, "/") {
		dockerHubRepoStats, err := parseRepoStats(response)
		if err != nil {
			log.Fatalln("Unable to parse response from Docker Hub for repository")
		} else {
			fmt.Printf("Repo: %s has been pulled %d times from the Docker Hub", request, dockerHubRepoStats.PullCount)
		}
	} else {
		dockerHubOrgStats, err := parseOrgStats(response)
		if err != nil {
			log.Fatalln("Unable to parse response from Docker Hub for user/organisation")
		} else {
			fmt.Printf("The organisation or user %s has %d repositories on the Docker hub.\n", request, dockerHubOrgStats.Count)
		}
	}
}
