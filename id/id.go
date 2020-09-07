package id

import "github.com/google/uuid"

// Responsible for minting unique identifiers for business model objects.
// Business model objects receive exactly one, unique, immutable identifier
// which is used to retrieve, update or link to an object.
//
// Identifiers are guaranteed to be unique within an instance of negtracker.
// Specifically, they are not guaranteed to be globally unique.
type Minter interface {
	Mint(interface{}) string
}

func Mint() string {
	return uuid.New().String()
}
