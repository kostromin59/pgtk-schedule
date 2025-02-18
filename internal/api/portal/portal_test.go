package portal

import (
	"pgtk-schedule/internal/models"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCurrentWeekLessons(t *testing.T) {
	loc, err := time.LoadLocation(timezone)
	require.NoError(t, err)
	now := time.Now().In(loc)
	testLessons := map[string][]Lesson{
		"stream1": {
			{
				ID:        "1",
				Name:      "Go",
				Teacher:   "ФИО",
				Substream: "A",
				Cabinet:   "101",
				Type:      "занятие на подгруппу",
				StreamID:  1,
				TimeStart: "09.00",
				TimeEnd:   "10.30",
				DateStart: now,
				DateEnd:   now,
			},
			{
				ID:        "2",
				Name:      "Physics",
				Teacher:   "ФИО2",
				Substream: "B",
				Cabinet:   "102",
				Type:      "занятие на подгруппу",
				StreamID:  1,
				TimeStart: "11.00",
				TimeEnd:   "12.30",
				DateStart: now,
				DateEnd:   now,
			},
		},
		"stream2": {
			{
				ID:        "3",
				Name:      "Chemistry",
				Teacher:   "ФИО",
				Substream: "A",
				Cabinet:   "103",
				Type:      "лекция",
				StreamID:  2,
				TimeStart: "13.00",
				TimeEnd:   "14.30",
				DateStart: now,
				DateEnd:   now,
			},
		},
	}

	p := &portal{
		lessons: testLessons,
	}

	tests := []struct {
		name            string
		stream          string
		substream       string
		expectedLessons []models.Lesson
		expectedError   error
	}{
		{
			name:      "Success: Get lessons for stream1 and substream A",
			stream:    "stream1",
			substream: "A",
			expectedLessons: []models.Lesson{
				{
					ID:        "1",
					Name:      "Go",
					Teacher:   "ФИО",
					Substream: "A",
					Cabinet:   "101",
					Type:      "занятие на подгруппу",
					Stream:    "1",
					DateStart: time.Date(now.Year(), now.Month(), now.Day(), 9, 0, 0, 0, loc),
					DateEnd:   time.Date(now.Year(), now.Month(), now.Day(), 10, 30, 0, 0, loc),
				},
			},
			expectedError: nil,
		},
		{
			name:            "Error: Unknown stream",
			stream:          "unknown",
			substream:       "A",
			expectedLessons: nil,
			expectedError:   models.ErrStreamIsUnknown,
		},
		{
			name:            "Error: No lessons for substream",
			stream:          "stream1",
			substream:       "C",
			expectedLessons: nil,
			expectedError:   models.ErrLessonsAreEmpty,
		},
		{
			name:      "Success: Get all lessons for stream2 (no substream filter)",
			stream:    "stream2",
			substream: "",
			expectedLessons: []models.Lesson{
				{
					ID:        "3",
					Name:      "Chemistry",
					Teacher:   "ФИО",
					Substream: "A",
					Cabinet:   "103",
					Type:      "лекция",
					Stream:    "2",
					DateStart: time.Date(now.Year(), now.Month(), now.Day(), 13, 0, 0, 0, loc),
					DateEnd:   time.Date(now.Year(), now.Month(), now.Day(), 14, 30, 0, 0, loc),
				},
			},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lessons, err := p.CurrentWeekLessons(tt.stream, tt.substream)

			if tt.expectedError != nil {
				require.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedLessons, lessons)
			}
		})
	}
}
