package store

import "github.com/emetsger/negtracker/model"

type Store interface {
	Retrieve(id string) (res model.Neg, err error)
	Store(n model.Neg) (id string, err error)
}
