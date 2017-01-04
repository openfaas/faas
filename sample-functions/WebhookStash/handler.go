package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"time"
)

func main() {
	input, _ := ioutil.ReadAll(os.Stdin)
	fmt.Println("Stashing request")
	now := time.Now()
	stamp := strconv.FormatInt(now.UnixNano(), 10)

	ioutil.WriteFile(stamp+".txt", input, 0644)
}
