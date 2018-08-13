// Copyright (c) OpenFaaS Author(s) 2018. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

// API source - https://haveibeenpwned.com/API/v2#SearchingPwnedPasswordsByRange
// Idea from Matthew Holt (@mholt6)

package function

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

// Handle a serverless request
func Handle(payload []byte) string {

	if len(payload) == 0 {
		return "Enter a number of characters."
	}

	hashed := fmt.Sprintf("%X", sha1.Sum(payload))
	prefix := hashed[:5]

	c := http.Client{}

	req, _ := http.NewRequest(http.MethodGet,
		fmt.Sprintf("https://api.pwnedpasswords.com/range/%s", prefix), nil)

	res, err := c.Do(req)
	if err != nil {
		return err.Error()
	}

	var bytesOut []byte
	if res.Body != nil {
		defer res.Body.Close()
		bytesOut, _ = ioutil.ReadAll(res.Body)
	}

	passwords := string(bytesOut)
	foundTimes, findErr := findPassword(passwords, prefix, hashed)
	if findErr != nil {
		return findErr.Error()
	}

	result := result{Found: foundTimes}
	output, _ := json.Marshal(result)
	return string(output)
}

type result struct {
	Found int `json:"found"`
}

func findPassword(passwords string, prefix string, hashed string) (int, error) {
	foundTimes := 0

	for _, passwordLine := range strings.Split(passwords, "\r\n") {
		if passwordLine != "" {
			parts := strings.Split(passwordLine, ":")

			if fmt.Sprintf("%s%s", prefix, parts[0]) == hashed {
				var convErr error
				foundTimes, convErr = strconv.Atoi(parts[1])
				if convErr != nil {
					return 0, fmt.Errorf(`Cannot convert value: "%s", error: "%s"\n`, parts[1], convErr)
				}
				break
			}
		}
	}
	return foundTimes, nil
}
