package models

import "errors"

var (
	ErrStudentNotFound = errors.New("student not found")
)

type Student struct {
	ID        int64
	Nickname  *string
	Stream    *string
	Substream *string
}
