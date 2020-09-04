// Responsible for the durable storage of business objects and their retrieval from the
// persistence layer.
package store

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
