package etag

import (
	"github.com/emetsger/negtracker/id"
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
	"time"
)

var encUnderTest = NewEncoder()

func Test_EtagEncodeWith(t *testing.T) {
	log.Printf("Using encoder %p", &encUnderTest)
	input := "moo"
	result := encUnderTest.EncodeWith(input)
	log.Printf("result: %s encoded as %s", input, result)
	assert.NotEqual(t, "", result)
}

func Test_EtagWithTimestamp(t *testing.T) {
	log.Printf("Using encoder %p", &encUnderTest)
	input := time.Now()
	result := encUnderTest.AddTime(input).Encode()
	log.Printf("result: %s encoded as %s", input, result)
	assert.NotEqual(t, "", result)
}

func Test_EtagWithUuid(t *testing.T) {
	log.Printf("Using encoder %p", &encUnderTest)
	input := (&id.UuidMinter{}).Mint("foo")
	result := encUnderTest.AddString(input).Encode()

	log.Printf("result: %s encoded as %s", input, result)
	assert.NotEqual(t, "", result)
}

func Test_EncodeWithCreateUpdateAndUuid(t *testing.T) {
	log.Printf("Using encoder %p", &encUnderTest)
	res := encUnderTest.AddString((&id.UuidMinter{}).Mint("foo")).AddTime(time.Now()).AddTime(time.Now()).Encode()
	log.Printf("result: encoded as %s", res)
	assert.NotEqual(t, "", res)
}

func Test_EncodeMixWithAndAdd(t *testing.T) {
	log.Printf("Using encoder %p", &encUnderTest)
	created := time.Now()
	modified := time.Now()
	uuid := (&id.UuidMinter{}).Mint("foo")
	encUnderTest.AddTime(created).AddTime(modified)
	res := encUnderTest.EncodeWith(uuid)
	log.Printf("result: encoded as %s", res)
	assert.NotEqual(t, "", res)
}

func Test_EncoderIdempotent(t *testing.T) {
	// same inputs, same outputs
	log.Printf("Using encoder %p", &encUnderTest)
	created := time.Now()
	modified := time.Now()
	res1 := encUnderTest.AddTime(created).AddTime(modified).Encode()

	res2 := encUnderTest.AddTime(created).AddTime(modified).Encode()

	assert.Equal(t, res1, res2)
}

func Test_EncoderNoState(t *testing.T) {
	assert.Panics(t, func() { NewEncoder().Encode() })
}

func Test_EncoderOneEmptyState(t *testing.T) {
	assert.Panics(t, func() { NewEncoder().EncodeWith("foo", "") })
}
