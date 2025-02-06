package portal

import (
	"time"
)

type Lesson struct {
	ID        string    `json:"id,omitempty"`
	Name      string    `json:"discipline_name,omitempty"`
	Teacher   string    `json:"teacher_fio,omitempty"`
	StreamID  int       `json:"stream_id,omitempty"`
	Substream string    `json:"subgroup_name,omitempty"`
	Cabinet   string    `json:"cabinet_fullnumber_wotype,omitempty"`
	Type      string    `json:"classtype_name,omitempty"`
	DateStart time.Time `json:"date_start,omitempty"`
	DateEnd   time.Time `json:"date_end,omitempty"`
	TimeStart string    `json:"daytime_start,omitempty"`
	TimeEnd   string    `json:"daytime_end,omitempty"`
}
