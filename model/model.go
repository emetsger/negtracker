package model

import (
	"github.com/emetsger/negtracker/etag"
	"github.com/emetsger/negtracker/store"
	"strings"
	"time"
)

// Encapsulates a business object that is exposed as a web resource
//
// Every business object should implement this interface, so that a type assertion with this interface may be used by
// functions accepting the `interface{}` type.
type WebResource interface {
	// Obtain the business identifier of the resource, which is guaranteed to be immutable once the resource has been
	// durably persisted.
	GetId() string
	// Obtain the UTC time the resource was created.
	GetCreated() time.Time
	// Obtain the UTC time the state of the resource was last updated.
	GetUpdated() time.Time
	// Obtain the HTTP ETag of the resource, suitable for performing validation of resource state
	GetEtag() Etag
	// Set the business identifier of the resource, which is guaranteed to be immutable once the resource has been
	// durably persisted
	SetId(id string)
	// Set the UTC time the resource was created.
	SetCreated(t time.Time)
	// Set the UTC time the state of the resource was last updated
	SetUpdated(t time.Time)
}

// Encapsulates an ETag and whether it is a weak or strong validator.  The zero value of an ETag is the same as that
// of the type `string`: a zero-length empty string.
type Etag string

// Returns true if the ETag is a strong validator.  Panics if the ETag is an empty string.
func (e Etag) strong() bool {
	if string(e) == "" {
		panic("model: etag is empty")
	}
	return !strings.HasPrefix(string(e), "W/")
}

func (e *Neg) GetId() string {
	return e.Id
}

func (e *Neg) GetCreated() time.Time {
	return e.Created
}

func (e *Neg) GetUpdated() time.Time {
	return e.Updated
}

func (e *Neg) SetId(id string) {
	e.Id = id
}

func (e *Neg) SetCreated(t time.Time) {
	e.Created = t
}

func (e *Neg) SetUpdated(t time.Time) {
	e.Updated = t
}

func (e *Neg) GetEtag() Etag {
	return Etag(etag.NewEncoder().AddString(e.Id).AddTime(e.Created).AddTime(e.Updated).Encode(true))
}

type Neg struct {
	Id          string
	Created     time.Time
	Updated     time.Time
	Film        string
	EI          int
	Developer   string
	FrameNumber string
	Tags        []string
	Description string
	Format      string
}

func (n *Neg) Store(s store.Api) (id string, err error) {
	return s.Store(*n)
}

func (n *Neg) Retrieve(s store.Api, id string) (err error) {
	return s.Retrieve(id, n)
}
