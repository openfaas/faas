package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

func main() {
	input, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.Fatalf("Unable to read standard input: %s", err.Error())
	}
	fmt.Println(string(input))
}
