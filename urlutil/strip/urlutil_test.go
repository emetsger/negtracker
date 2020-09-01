package strip

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTrailingSlashesSingleSlash(t *testing.T) {
	assert.Equal(t, "http://example.com", TrailingSlashes(("http://example.com/")))
}


func TestTrailingSlashesMultipleSlashes(t *testing.T) {
	assert.Equal(t, "http://example.com", TrailingSlashes(("http://example.com///")))
}

func TestTrailingSlashesOnlySlashes(t *testing.T) {
	assert.Equal(t, "", TrailingSlashes(("///")))
}

func TestTrailingSlashesZeroLengthString(t *testing.T) {
	assert.Equal(t, "", TrailingSlashes(("")))
}

func TestTrailingSlashesEmptyString(t *testing.T) {
	assert.Equal(t, " ", TrailingSlashes((" ")))
}