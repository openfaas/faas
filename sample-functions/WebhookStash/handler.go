package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"time"
)

func main() {
	input, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.Fatalf("Cannot read input %s.\n", err)
		return
	}
	now := time.Now()
	stamp := strconv.FormatInt(now.UnixNano(), 10)

	writeErr := ioutil.WriteFile(stamp+".txt", input, 0644)
	if writeErr != nil {
		log.Fatalf("Cannot write input %s.\n", err)
		return
	}

	fmt.Printf("Stashing request: %s.txt\n", stamp)
}
