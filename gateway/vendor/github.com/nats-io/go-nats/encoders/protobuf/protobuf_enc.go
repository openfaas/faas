// Copyright 2015-2018 The NATS Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package protobuf

import (
	"errors"

	"github.com/golang/protobuf/proto"
	"github.com/nats-io/go-nats"
)

// Additional index for registered Encoders.
const (
	PROTOBUF_ENCODER = "protobuf"
)

func init() {
	// Register protobuf encoder
	nats.RegisterEncoder(PROTOBUF_ENCODER, &ProtobufEncoder{})
}

// ProtobufEncoder is a protobuf implementation for EncodedConn
// This encoder will use the builtin protobuf lib to Marshal
// and Unmarshal structs.
type ProtobufEncoder struct {
	// Empty
}

var (
	ErrInvalidProtoMsgEncode = errors.New("nats: Invalid protobuf proto.Message object passed to encode")
	ErrInvalidProtoMsgDecode = errors.New("nats: Invalid protobuf proto.Message object passed to decode")
)

// Encode
func (pb *ProtobufEncoder) Encode(subject string, v interface{}) ([]byte, error) {
	if v == nil {
		return nil, nil
	}
	i, found := v.(proto.Message)
	if !found {
		return nil, ErrInvalidProtoMsgEncode
	}

	b, err := proto.Marshal(i)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// Decode
func (pb *ProtobufEncoder) Decode(subject string, data []byte, vPtr interface{}) error {
	if _, ok := vPtr.(*interface{}); ok {
		return nil
	}
	i, found := vPtr.(proto.Message)
	if !found {
		return ErrInvalidProtoMsgDecode
	}

	return proto.Unmarshal(data, i)
}
