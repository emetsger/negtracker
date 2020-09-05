package id

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

var uuidMinter = UuidMinter{}

func Test_UuidMinter(t *testing.T) {
	id := uuidMinter.Mint("moo")
	assert.NotNil(t, id)
}

func Test_BusinessObjectRequired(t *testing.T) {
	assert.Panics(t, func() { uuidMinter.Mint(nil) })
}
