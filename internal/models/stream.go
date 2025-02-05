package models

import "errors"

var (
	ErrStreamIsUnknown = errors.New("unknown stream")
)

type Stream struct {
	ID         string
	Name       string
	Substreams []string
}
