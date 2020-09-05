package id

import (
	"github.com/google/uuid"
)

// Mints identifiers according to RFC 4122.
type UuidMinter struct {
}

// Mints a globally unique identifier, or UUID, according to RFC 4122.
func (m *UuidMinter) Mint(obj interface{}) string {
	if obj == nil {
		panic("id: cannot mint id for null business object")
	}
	return uuid.New().String()
}
