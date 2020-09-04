package model

import (
	"github.com/emetsger/negtracker/store"
)

type Neg struct {
	ID          string
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
	if obj, err := s.Retrieve(id); err != nil {
		return err
	} else {
		res := obj.(Neg)
		n.copyFrom(res)
	}

	return nil
}

func (dst *Neg) copyFrom(src Neg) {
	dst.ID = src.ID
	dst.Film = src.Film
	dst.Developer = src.Developer
	dst.FrameNumber = src.FrameNumber
	dst.Tags = src.Tags
	dst.Developer = src.Description
	dst.Format = src.Format
}
