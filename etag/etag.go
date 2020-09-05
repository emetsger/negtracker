// Provides for the creation of HTTP ETags by providing an encoder for types commonly used to generate ETags
package etag

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"time"
)

// Provides for the encoding of common data types as ETags.
// The caller is responsible for determining the tokens that comprise a given ETag, and encodes them using this
// interface.  The result of calling an Encode* method will be a base64 string derived from values supplied by
// the Add* or EncodeWith(...) methods.
//
// It is considered an error execute EncodeWith(...) or Encode() with any empty state.
type Encoder interface {

	// Encodes the supplied tokens as a base64 string
	EncodeWith(tokens ...string) string

	// Encodes any state supplied by the Add methods as a base64 string
	// It is an error to call Encode() when no tokens have been supplied.
	Encode() string

	// Add a byte token to be considered as an ETag input
	AddByte(token []byte) Encoder

	// Add a string token to be considered as an ETag input
	AddString(token string) Encoder

	// Add an int64 token to be considered as an ETag input
	AddInt64(token int64) Encoder

	// Add a time token to be considered as an ETag input
	AddTime(token time.Time) Encoder
}

// Simply stores the state to be encoded as a slice of byte slices.  All methods (e.g. Add* or EncodeWith(...)) must
// convert their inputs to []byte.  Note the initState() function provides for creating the initial [][]byte slice.
type naiveEncoder struct {
	state [][]byte
}

// Initializes a new Encoder
func NewEncoder() Encoder {
	return &naiveEncoder{initState()}
}

func (e *naiveEncoder) EncodeWith(tokens ...string) string {
	// convert each attribute to a byte slice and add to the state
	for i := range tokens {
		e.AddByte([]byte(tokens[i]))
	}

	return e.Encode()
}

func (e *naiveEncoder) Encode() string {
	switch len(e.state) {
	case 0:
		panic("etag: encoder has no state, this may be buggy behavior in the caller")
	default:
		for i := range e.state {
			if len(e.state[i]) == 0 {
				panic(fmt.Sprintf("etag: encoded field %d was empty, this may be buggy behavior in the caller", i))
			}
		}
	}

	buf := bytes.Buffer{}
	encoder := base64.NewEncoder(base64.StdEncoding, &buf)
	defer func() { encoder.Close() }()
	for i := range e.state {
		encoder.Write(e.state[i])
	}

	e.state = initState()
	return buf.String()
}

func (e *naiveEncoder) AddByte(b []byte) Encoder {
	e.state = append(e.state, b)
	return e
}

func (e *naiveEncoder) AddString(s string) Encoder {
	return e.AddByte([]byte(s))
}

func (e *naiveEncoder) AddTime(t time.Time) Encoder {
	return e.AddInt64(t.UnixNano())
}

func (e *naiveEncoder) AddInt64(i int64) Encoder {
	b := make([]byte, binary.MaxVarintLen64)
	binary.PutVarint(b, i)
	e.AddByte(b)
	return e
}

// Allocates a new [][]byte and returns a copy (not a pointer)
func initState() [][]byte {
	return *new([][]byte)
}
