package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/microcosm-cc/bluemonday"
	"github.com/russross/blackfriday"
)

func main() {
	input, _ := ioutil.ReadAll(os.Stdin)
	unsafe := blackfriday.Run([]byte(input), blackfriday.WithNoExtensions())
	html := bluemonday.UGCPolicy().SanitizeBytes(unsafe)
	fmt.Println(string(html))
}
