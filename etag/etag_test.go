package etag

import (
	"github.com/emetsger/negtracker/id"
	"github.com/stretchr/testify/assert"
	"log"
	"strings"
	"testing"
	"time"
)

var encUnderTest = NewEncoder()

func Test_EtagEncodeWith(t *testing.T) {
	log.Printf("Using encoder %p", &encUnderTest)
	input := "moo"
	result := encUnderTest.EncodeWith(true, input)
	log.Printf("result: %s encoded as %s", input, result)
	assert.NotEqual(t, "", result)
}

func Test_EtagWithTimestamp(t *testing.T) {
	log.Printf("Using encoder %p", &encUnderTest)
	input := time.Now()
	result := encUnderTest.AddTime(input).Encode(true)
	log.Printf("result: %s encoded as %s", input, result)
	assert.NotEqual(t, "", result)
}

func Test_EtagWithUuid(t *testing.T) {
	log.Printf("Using encoder %p", &encUnderTest)
	input := (&id.UuidMinter{}).Mint("foo")
	result := encUnderTest.AddString(input).Encode(true)

	log.Printf("result: %s encoded as %s", input, result)
	assert.NotEqual(t, "", result)
}

func Test_EncodeWithCreateUpdateAndUuid(t *testing.T) {
	log.Printf("Using encoder %p", &encUnderTest)
	res := encUnderTest.AddString((&id.UuidMinter{}).Mint("foo")).AddTime(time.Now()).AddTime(time.Now()).Encode(true)
	log.Printf("result: encoded as %s", res)
	assert.NotEqual(t, "", res)
}

func Test_EncodeMixWithAndAdd(t *testing.T) {
	log.Printf("Using encoder %p", &encUnderTest)
	created := time.Now()
	modified := time.Now()
	uuid := (&id.UuidMinter{}).Mint("foo")
	encUnderTest.AddTime(created).AddTime(modified)
	res := encUnderTest.EncodeWith(true, uuid)
	log.Printf("result: encoded as %s", res)
	assert.NotEqual(t, "", res)
}

func Test_EncoderIdempotent(t *testing.T) {
	// same inputs, same outputs
	log.Printf("Using encoder %p", &encUnderTest)
	created := time.Now()
	modified := time.Now()
	res1 := encUnderTest.AddTime(created).AddTime(modified).Encode(true)

	res2 := encUnderTest.AddTime(created).AddTime(modified).Encode(true)

	assert.Equal(t, res1, res2)
}

func Test_EncoderNoState(t *testing.T) {
	assert.Panics(t, func() { NewEncoder().Encode(true) })
}

func Test_EncoderOneEmptyState(t *testing.T) {
	assert.Panics(t, func() { NewEncoder().EncodeWith(true, "", "moo") })
}

func Test_EncodeWeak(t *testing.T) {
	assert.True(t, strings.HasPrefix(encUnderTest.EncodeWith(false, "moo"), "W/"))
}

func Test_EncodeQuotedProperlyWeakTag(t *testing.T) {
	encoded := encUnderTest.EncodeWith(false, "moo")[2:]
	assert.True(t, strings.HasPrefix(encoded, "\""))
	assert.True(t, strings.HasSuffix(encoded, "\""))
}

func Test_EncodeQuotedProperlyStrongTag(t *testing.T) {
	encoded := encUnderTest.EncodeWith(true, "moo")
	assert.True(t, strings.HasPrefix(encoded, "\""))
	assert.True(t, strings.HasSuffix(encoded, "\""))
}
