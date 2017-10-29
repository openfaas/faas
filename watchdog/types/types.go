// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package types

import (
	"encoding/json"
	"io"
	"io/ioutil"
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

func MarshalRequest(data io.Reader, header *http.Header) (io.Reader, error) {
	bs, err := ioutil.ReadAll(data)
	if err != nil {
		return nil, err
	}
	req := MarshalReq{
		Body: MarshalBody{
			Raw: bs,
		},
		Header: *header,
	}
	r, w := io.Pipe()
	encoder := json.NewEncoder(w)
	go encoder.Encode(req)
	return r, nil
}
