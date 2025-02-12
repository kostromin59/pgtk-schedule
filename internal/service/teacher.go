package service

import (
	"pgtk-schedule/internal/models"
	"strings"
	"time"
)

type teacherPortal interface {
	Streams() []models.Stream
	Timezone() string
}

type scheduleService interface {
	TodayLessons(stream, substream string) ([]models.Lesson, error)
}

type teacher struct {
	portal          teacherPortal
	scheduleService scheduleService
}

func NewTeacher(portal teacherPortal, scheduleService scheduleService) *teacher {
	return &teacher{
		portal:          portal,
		scheduleService: scheduleService,
	}
}

func (t *teacher) TodayList() ([]string, error) {
	loc, err := time.LoadLocation(t.portal.Timezone())
	if err != nil {
		return nil, err
	}

	now := time.Now().In(loc)

	streams := t.portal.Streams()

	teacherSet := make(map[string]struct{}, 0)

	for _, stream := range streams {
		if len(stream.Substreams) == 0 {
			lessons, err := t.scheduleService.TodayLessons(stream.ID, "")
			if err != nil {
				return nil, err
			}

			for _, lesson := range lessons {
				if now.Before(lesson.DateEnd) {
					teacherSet[lesson.Teacher] = struct{}{}
				}
			}
		}

		for _, substream := range stream.Substreams {
			lessons, err := t.scheduleService.TodayLessons(stream.ID, substream)
			if err != nil {
				return nil, err
			}

			for _, lesson := range lessons {
				if now.Before(lesson.DateEnd) {
					teacherSet[lesson.Teacher] = struct{}{}
				}
			}
		}
	}

	teachers := make([]string, 0, len(teacherSet))

	for teacher := range teacherSet {
		teachers = append(teachers, teacher)
	}

	return teachers, nil
}

func (t *teacher) Find(teacher string) (models.Lesson, error) {
	loc, err := time.LoadLocation(t.portal.Timezone())
	if err != nil {
		return models.Lesson{}, err
	}

	now := time.Now().In(loc)

	streams := t.portal.Streams()

	for _, stream := range streams {
		if len(stream.Substreams) == 0 {
			lessons, err := t.scheduleService.TodayLessons(stream.ID, "")
			if err != nil {
				return models.Lesson{}, err
			}

			for _, lesson := range lessons {
				if now.Before(lesson.DateEnd) && strings.Contains(lesson.Teacher, teacher) {
					return lesson, nil
				}
			}
		}

		for _, substream := range stream.Substreams {
			lessons, err := t.scheduleService.TodayLessons(stream.ID, substream)
			if err != nil {
				return models.Lesson{}, err
			}

			for _, lesson := range lessons {
				if now.Before(lesson.DateEnd) && strings.Contains(lesson.Teacher, teacher) {
					return lesson, nil
				}
			}
		}
	}

	return models.Lesson{}, models.ErrLessonNotFound
}
