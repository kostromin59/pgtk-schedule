package portal

import "time"

type Week struct {
	Value     int      `json:"value,omitempty"`
	Text      string   `json:"text,omitempty"`
	StartDate WeekDate `json:"start_date,omitempty"`
	EndDate   WeekDate `json:"end_date,omitempty"`
	Selected  bool     `json:"selected,omitempty"`
}

type WeekDate struct {
	time.Time
}

func (wd *WeekDate) UnmarshalJSON(b []byte) error {
	t, err := time.Parse(`"02.01.2006"`, string(b))
	if err != nil {
		return err
	}

	wd.Time = t

	return nil
}
