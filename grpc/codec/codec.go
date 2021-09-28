package codec

import "google.golang.org/grpc/encoding"

type RawMessage []byte

// By default, gRPC registers and uses the "proto" codec, so it is not necessary to do this in your own code to send and receive proto messages.
// https://github.com/grpc/grpc-go/blob/master/Documentation/encoding.md#using-a-codec
const Name string = "proto"
const rm string = "rawMessage"

func (r RawMessage) Reset()       {}
func (RawMessage) ProtoMessage()  {}
func (RawMessage) String() string { return rm }

type Codec struct {
	Base encoding.Codec
}

// Marshal returns the wire format of v. rawMessages would be returned without encoding.
func (c *Codec) Marshal(v interface{}) ([]byte, error) {
	if raw, ok := v.(RawMessage); ok {
		return raw, nil
	}

	return c.Base.Marshal(v)
}

// Unmarshal parses the wire format into v. rawMessages would not be unmarshalled.
func (c *Codec) Unmarshal(data []byte, v interface{}) error {
	if raw, ok := v.(*RawMessage); ok {
		*raw = data
		return nil
	}

	return c.Base.Unmarshal(data, v)
}

func (c *Codec) Name() string {
	return Name
}

// String return codec name.
func (c *Codec) String() string {
	return "raw:" + c.Base.Name()
}
