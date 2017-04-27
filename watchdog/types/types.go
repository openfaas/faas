// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package types

import (
	"encoding/json"
	"net/http"
	"os"
)

// OsEnv implements interface to wrap os.Getenv
type OsEnv struct {
}

// Getenv wraps os.Getenv
func (OsEnv) Getenv(key string) string {
	return os.Getenv(key)
}

type MarshalBody struct {
	Raw []byte `json:"raw"`
}

type MarshalReq struct {
	Header http.Header `json:"header"`
	Body   MarshalBody `json:"body"`
}

func UnmarshalRequest(data []byte) (*MarshalReq, error) {
	request := MarshalReq{}
	err := json.Unmarshal(data, &request)
	return &request, err
}

func MarshalRequest(data []byte, header *http.Header) ([]byte, error) {
	req := MarshalReq{
		Body: MarshalBody{
			Raw: data,
		},
		Header: *header,
	}

	res, marshalErr := json.Marshal(&req)
	return res, marshalErr
}
