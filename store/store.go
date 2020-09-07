// Responsible for the durable storage of business objects and their retrieval from the
// persistence layer.
package store

import (
	"errors"
	"fmt"
)

// Presents an API for durably storing business objects.
type Api interface {

	// Retrieve the identified object from the store and unmarshal it to t.
	// The underlying value of t must be a pointer to a model struct, e.g.:
	//   negative := model.Neg
	//   _ = impl.Retrieve("1", &negative)
	//
	// The identifier is a persistence layer id, which may change to a business layer
	// id in the future.
	Retrieve(id string, t interface{}) (err error)

	// Durably persist the supplied object in the storage layer.
	// The returned id will be a persistence layer id, which may change to a business
	// layer id in the future.
	Store(obj interface{}) (id string, err error)
}

const (
	EnvDbUri           = "DB_URI"
	EnvDbName          = "DB_NAME"
	EnvDbNegCollection = "DB_NEG_COLLECTION"
)

type StorageError struct {
	sentinel   error
	msg, cause string
}

func SentinelErr(sentinel error, msg, cause string) StorageError {
	if sentinel == nil {
		panic("store: error creating error, sentinel value required")
	}
	return StorageError{sentinel, msg, cause}
}

func GenericErr(msg, cause string) error {
	return StorageError{GeneralErr, msg, cause}
}

func (e StorageError) Error() string {
	if len(e.msg) > 0 {
		if len(e.cause) > 0 {
			return fmt.Sprintf("%s, %s: %s", e.sentinel.Error(), e.msg, e.cause)
		} else {
			return fmt.Sprintf("%s, %s", e.sentinel.Error(), e.msg)
		}
	}

	return e.sentinel.Error()
}

var GeneralErr = errors.New("store: error occurred interacting with the storage layer")
var DuplicateKeyErr = errors.New("store: attempt to insert a duplicate key")
var DecodingErr = errors.New("store: error decoding object")

func (e StorageError) Is(target error) bool {
	return target == GeneralErr || target == DuplicateKeyErr || target == DecodingErr
}
