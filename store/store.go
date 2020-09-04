package store

type Api interface {
	Retrieve(id string) (obj interface{}, err error)
	Store(obj interface{}) (id string, err error)
}
