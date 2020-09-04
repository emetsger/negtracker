package model

import (
	"github.com/emetsger/negtracker/index"
	"github.com/emetsger/negtracker/store"
)

// Business objects implementing this interface can satisfy storage operations.
type StoreAware interface {

	// Implementations will retrieve the copy referenced by the supplied storage layer id,
	// and the state of the implementing object will be overwritten with the state
	// retrieved from the storage layer.
	Retrieve(api store.Api, id string) (err error)

	// Implementations store a copy of themselves using the store.Api.
	//
	// The returned identifier is a persistence layer id.
	Store(api store.Api) (id string, err error)
}

// Business objects implementing this interface can satisfy index operations
type IndexAware interface {
	Add(api index.Api)
	Update(api index.Api)
	Del(api index.Api)
}
