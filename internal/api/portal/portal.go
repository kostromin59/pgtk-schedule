package portal

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"pgtk-schedule/internal/models"
	"pgtk-schedule/pkg/request"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const (
	baseUrl     = "https://psi.thinkery.ru/shedule/public"
	scheduleUrl = baseUrl + "/public_shedule"
	weeksUrl    = baseUrl + "/get_weekdates_actual"
	gridUrl     = baseUrl + "/public_shedule_spo_grid"
	lessonsUrl  = baseUrl + "/public_getsheduleclasses_spo"

	saturdayNextDayHours = 14
	timezone             = "Asia/Yekaterinburg"
)

var (
	regexStudyYearId = regexp.MustCompile(`studyyear_id\s*:\s*'(\d+)'`)
)

type portal struct {
	studyYearId string
	term        string
	streams     []Stream
	weeks       []Week
	lessons     map[string][]Lesson
	mu          sync.RWMutex
}

func New() *portal {
	return &portal{}
}

func (*portal) Timezone() string {
	return timezone
}

func (p *portal) Update() error {
	res, err := request.New(scheduleUrl).Do()
	if err != nil {
		return err
	}

	stringHtml := string(res.Body())
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(res.Body()))
	if err != nil {
		return fmt.Errorf("unable to create document: %w", err)
	}

	studyYearId, err := p.extractStudyYearId(stringHtml)
	if err != nil {
		return fmt.Errorf("unable to extract study year id: %w", err)
	}

	term, err := p.extractTerm(doc)
	if err != nil {
		return fmt.Errorf("unable to extract term: %w", err)
	}

	streams := p.extractStreams(doc)
	if len(streams) == 0 {
		return errors.New("streams not found")
	}

	weeks, err := p.collectWeeks(studyYearId)
	if err != nil {
		return fmt.Errorf("unable to collect weeks: %w", err)
	}

	// Substreams
	week, err := p.currentWeek(weeks)
	if err != nil {
		return fmt.Errorf("unable to collect substreams: %w", err)
	}

	var wg sync.WaitGroup
	for i, s := range streams {
		wg.Add(1)
		go func() {
			defer wg.Done()
			substreams, err := p.collectSubstreams(s.Value, term, studyYearId, week.Value)
			if err != nil {
				log.Println(err.Error(), s)
				return
			}

			streams[i].Substreams = substreams
		}()
	}

	wg.Wait()

	p.mu.Lock()
	defer p.mu.Unlock()
	p.studyYearId = studyYearId
	p.term = term
	p.weeks = weeks
	p.streams = streams

	wg = sync.WaitGroup{}
	mu := sync.Mutex{}
	lessons := make(map[string][]Lesson, len(streams))
	for _, stream := range streams {
		wg.Add(1)
		go func() {
			defer wg.Done()
			l, err := p.collectSchedule(stream.Value, term, studyYearId, week.StartDate.Time, week.EndDate.Time)
			if err != nil {
				log.Println(err.Error(), stream.Name)
				return
			}

			mu.Lock()
			lessons[stream.Value] = l
			mu.Unlock()
		}()
	}

	wg.Wait()
	p.lessons = lessons

	return nil
}

func (p *portal) CurrentWeekLessons(stream, substream string) ([]models.Lesson, error) {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return nil, err
	}

	p.mu.RLock()
	defer p.mu.RUnlock()

	if len(p.lessons) == 0 {
		return nil, models.ErrLessonsAreEmpty
	}

	l, ok := p.lessons[stream]
	if !ok {
		return nil, models.ErrStreamIsUnknown
	}

	lessons := make([]models.Lesson, 0, len(l))
	for _, lesson := range l {
		if strings.Contains(lesson.Type, "подгрупп") && lesson.Substream != substream {
			continue
		}

		startTimeFormatted := strings.Replace(lesson.TimeStart, ".", ":", 1)
		endTimeFormatted := strings.Replace(lesson.TimeEnd, ".", ":", 1)

		parsedStartTime, err := time.Parse("15:04", startTimeFormatted)
		if err != nil {
			return nil, err
		}

		parsedEndTime, err := time.Parse("15:04", endTimeFormatted)
		if err != nil {
			return nil, err
		}

		combinedStartDateTime := time.Date(
			lesson.DateStart.Year(),
			lesson.DateStart.Month(),
			lesson.DateStart.Day(),
			parsedStartTime.Hour(),
			parsedStartTime.Minute(),
			0, 0,
			loc,
		)

		combinedEndDateTime := time.Date(
			lesson.DateEnd.Year(),
			lesson.DateEnd.Month(),
			lesson.DateEnd.Day(),
			parsedEndTime.Hour(),
			parsedEndTime.Minute(),
			0, 0,
			loc,
		)

		lessons = append(lessons, models.Lesson{
			ID:        lesson.ID,
			Name:      lesson.Name,
			Teacher:   lesson.Teacher,
			Substream: lesson.Substream,
			Cabinet:   lesson.Cabinet,
			Type:      lesson.Type,
			Stream:    fmt.Sprintf("%d", lesson.StreamID),
			DateStart: combinedStartDateTime,
			DateEnd:   combinedEndDateTime,
		})
	}

	if len(lessons) == 0 {
		return nil, models.ErrLessonsAreEmpty
	}

	return lessons, nil
}

func (p *portal) Streams() []models.Stream {
	p.mu.RLock()
	defer p.mu.RUnlock()

	streams := make([]models.Stream, len(p.streams))
	for i, s := range p.streams {
		streams[i] = models.Stream{
			ID:         s.Value,
			Name:       s.Name,
			Substreams: s.Substreams,
		}
	}

	return streams
}

func (p *portal) currentWeek(weeks []Week) (Week, error) {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return Week{}, err
	}

	now := time.Now().In(loc)
	index := -1
	for i, w := range weeks {
		if w.Selected {
			index = i
			break
		}
	}

	if index == -1 {
		return Week{}, errors.New("current week not found")
	}

	if index == len(p.weeks)-1 {
		return weeks[index], nil
	}

	weekday := now.Weekday()
	if weekday == time.Saturday && now.Hour() >= saturdayNextDayHours {
		return weeks[index+1], nil
	}

	if weekday == time.Sunday {
		return weeks[index+1], nil
	}

	return weeks[index], nil
}

func (p *portal) extractStudyYearId(html string) (string, error) {
	match := regexStudyYearId.FindStringSubmatch(html)
	if len(match) < 2 {
		return "", errors.New("study year id not found")
	}

	return match[1], nil
}

func (p *portal) extractTerm(doc *goquery.Document) (string, error) {
	opt := doc.Find("#termdiv").Find("select#term").Find("option[selected]").First()

	term, ok := opt.Attr("value")
	if !ok {
		return "", errors.New("term not found")
	}

	return term, nil
}

func (p *portal) extractStreams(doc *goquery.Document) []Stream {
	opts := doc.Find("#stream_iddiv").Find("select#stream_id").Find("option")

	streams := make([]Stream, 0, len(opts.Nodes))
	opts.Each(func(i int, s *goquery.Selection) {
		v, ok := s.Attr("value")
		if !ok {
			return
		}

		v = strings.TrimSpace(v)

		if v == "" {
			return
		}

		text := s.Text()
		text = strings.TrimSpace(text)

		stream := Stream{
			Name:  text,
			Value: v,
		}

		streams = append(streams, stream)
	})

	return streams
}

func (p *portal) collectWeeks(studyYearId string) ([]Week, error) {
	body := map[string]string{
		"studyyear_id": studyYearId,
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("unable to marhal weeks body: %w", err)
	}

	res, err := request.New(weeksUrl).
		Method(http.MethodPost).
		ContentType("application/json").
		Body(bytes.NewReader(jsonBody)).
		Do()
	if err != nil {
		return nil, err
	}

	var weeks []Week
	if err := json.Unmarshal(res.Body(), &weeks); err != nil {
		return nil, fmt.Errorf("unable to decode weeks: %w", err)
	}

	return weeks, nil
}

func (p *portal) collectSubstreams(stream, term, studyYearId string, dateweek int) ([]string, error) {
	v := url.Values{}
	v.Set("studyyear_id", studyYearId)
	v.Set("stream_id", stream)
	v.Set("term", term)
	v.Set("dateweek", fmt.Sprintf("%d", dateweek))
	encoded := v.Encode()

	res, err := request.New(gridUrl).
		Method(http.MethodPost).
		Body(strings.NewReader(encoded)).
		ContentType("application/x-www-form-urlencoded").
		Do()

	if err != nil {
		return nil, err
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(res.Body()))
	if err != nil {
		return nil, fmt.Errorf("unable to create document: %w", err)
	}

	rows := doc.Find("#timetable").Find("thead").First().Find("tr")
	if len(rows.Nodes) < 2 {
		return nil, errors.New("substreams not found")
	}

	cells := rows.Last().Find("th")
	substreams := make([]string, 0, len(cells.Nodes))
	cells.Each(func(i int, s *goquery.Selection) {
		label := strings.TrimSpace(s.Text())
		if label != "" {
			substreams = append(substreams, label)
		}
	})

	return substreams, nil
}

func (p *portal) collectSchedule(stream, term, studyYearId string, startDate, endDate time.Time) ([]Lesson, error) {
	v := url.Values{}
	v.Set("studyyear_id", studyYearId)
	v.Set("stream_id", stream)
	v.Set("term", term)
	v.Set("start_date", startDate.Format("02.01.2006"))
	v.Set("end_date", endDate.Format("02.01.2006"))
	encoded := v.Encode()

	res, err := request.New(lessonsUrl).
		Method(http.MethodPost).
		Body(strings.NewReader(encoded)).
		ContentType("application/x-www-form-urlencoded").
		Do()

	if err != nil {
		return nil, err
	}

	var lessons []Lesson
	if err := json.Unmarshal(res.Body(), &lessons); err != nil {
		return nil, fmt.Errorf("unable to decode lessons: %w", err)
	}

	return lessons, nil
}
