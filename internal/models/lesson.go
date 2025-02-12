package models

import (
	"errors"
	"time"
)

var (
	ErrLessonsAreEmpty = errors.New("lessons are empty")
	ErrLessonNotFound  = errors.New("lesson not found")
)

type Lesson struct {
	ID        string
	Name      string
	Cabinet   string
	Type      string
	Teacher   string
	Stream    string
	Substream string
	DateStart time.Time
	DateEnd   time.Time
}
