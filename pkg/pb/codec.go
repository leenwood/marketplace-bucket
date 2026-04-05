// Package pb provides gRPC service definitions and request/response types
// for the cart service.
//
// Because this service does not rely on code generation, a JSON codec is
// registered to replace gRPC's default protobuf codec. All gRPC messages
// are therefore serialised as JSON. Clients must connect without TLS and
// may use tools such as grpcurl with --plaintext.
package pb

import (
	"encoding/json"

	"google.golang.org/grpc/encoding"
)

func init() {
	encoding.RegisterCodec(JSONCodec{})
}

// JSONCodec is a gRPC codec that uses encoding/json.
// Registering it under the name "proto" overrides gRPC's default protobuf
// codec so that plain Go structs can be used as message types without
// running protoc.
type JSONCodec struct{}

// Marshal serialises v to JSON.
func (JSONCodec) Marshal(v any) ([]byte, error) {
	return json.Marshal(v)
}

// Unmarshal deserialises JSON data into v.
func (JSONCodec) Unmarshal(data []byte, v any) error {
	return json.Unmarshal(data, v)
}

// Name returns "proto" to override gRPC's default codec.
func (JSONCodec) Name() string {
	return "proto"
}
