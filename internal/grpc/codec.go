package grpc

import (
	"encoding/json"

	"google.golang.org/grpc/encoding"
)

func init() {
	encoding.RegisterCodec(&jsonCodec{})
}

type jsonCodec struct{}

func (c *jsonCodec) Name() string { return "json" }

func (c *jsonCodec) Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func (c *jsonCodec) Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}
