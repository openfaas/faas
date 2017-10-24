package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/mlabouardy/9gag"
)

func main() {
	tag, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.Fatalf("Unable to read standard input: %s", err.Error())
	}
	gag9 := gag9.New()
	memes := gag9.FindByTag(string(tag))
	rawJson, _ := json.Marshal(memes)
	fmt.Println(string(rawJson))
}
