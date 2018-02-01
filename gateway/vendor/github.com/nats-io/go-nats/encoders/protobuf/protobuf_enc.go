// Copyright 2015 Apcera Inc. All rights reserved.

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
