package portal

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const (
	baseUrl     = "https://portal.pgtk-perm.ru/shedule/public"
	scheduleUrl = baseUrl + "/public_shedule"
	weeksUrl    = baseUrl + "/get_weekdates_actual"
	gridUrl     = baseUrl + "/public_shedule_spo_grid"

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
	mu          sync.RWMutex
}

func NewPortal() *portal {
	return &portal{}
}

func (p *portal) Update() error {
	resp, err := http.Get(baseUrl)
	if err != nil {
		return fmt.Errorf("unable to get site: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("bad status code: %d", resp.StatusCode)
	}

	html, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("unable to read body: %w", err)
	}

	stringHtml := string(html)
	doc, err := goquery.NewDocumentFromReader(resp.Body)
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

	p.mu.Lock()
	p.studyYearId = studyYearId
	p.term = term
	p.weeks = weeks
	p.mu.Unlock()

	// Substreams
	week, err := p.CurrentWeek()
	if err != nil {
		return fmt.Errorf("unable to collect substreams: %w", err)
	}

	var wg sync.WaitGroup
	for i, s := range streams {
		wg.Add(1)
		go func() {
			defer wg.Done()
			substreams, err := p.collectSubstreams(s.Value, p.term, p.studyYearId, week.Value)
			if err != nil {
				log.Println(err.Error(), s)
				return
			}

			streams[i].Substreams = substreams
		}()
	}

	wg.Wait()

	p.mu.Lock()
	p.streams = streams
	p.mu.Unlock()

	return nil
}

func (p *portal) CurrentWeek() (Week, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return Week{}, fmt.Errorf("unable to load timezone: %w", err)
	}

	now := time.Now().In(loc)
	index := 0
	for i, w := range p.weeks {
		if w.Selected {
			index = i
			break
		}
	}

	if index == len(p.weeks)-1 {
		return p.weeks[index], nil
	}

	weekday := now.Weekday()
	if weekday == time.Saturday && now.Hour() >= saturdayNextDayHours {
		return p.weeks[index+1], nil
	}

	if weekday == time.Sunday {
		return p.weeks[index+1], nil
	}

	return Week{}, errors.New("current week not found")
}

func (p *portal) extractStudyYearId(html string) (string, error) {
	match := regexStudyYearId.FindStringSubmatch(html)
	if len(match) <= 1 {
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

	req, err := http.NewRequest(http.MethodPost, weeksUrl, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("unable to create weeks request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	c := http.DefaultClient
	c.Timeout = 30 * time.Second

	resp, err := c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("unable to get weeks: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("bad status code: %d", resp.StatusCode)
	}

	var weeks []Week
	if err := json.NewDecoder(resp.Body).Decode(&weeks); err != nil {
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

	req, err := http.NewRequest(http.MethodPost, gridUrl, strings.NewReader(encoded))
	if err != nil {
		return nil, fmt.Errorf("unable to create weeks request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	c := http.DefaultClient
	c.Timeout = 30 * time.Second

	resp, err := c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("unable to get grid: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("bad status code: %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
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
