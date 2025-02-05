package models

type Stream struct {
	ID         string
	Name       string
	Substreams []string
}
