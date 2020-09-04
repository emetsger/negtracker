package store

type Api interface {
	Retrieve(id string, t interface{}) (err error)
	Store(obj interface{}) (id string, err error)
}
