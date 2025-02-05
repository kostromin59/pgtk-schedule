package request

import (
	"errors"
	"io"
	"net/http"
	"time"
)

var (
	ErrBadStatusCode = errors.New("bad status code")
)

var (
	headerContentType = "Content-Type"
)

type Request struct {
	url         string
	method      string
	contentType string
	body        io.Reader
}

type Result struct {
	body       []byte
	statusCode int
}

func (r *Result) StatusCode() int {
	return r.statusCode
}

func (r *Result) Body() []byte {
	return r.body
}

func New(url string) *Request {
	return &Request{
		url:    url,
		method: http.MethodGet,
	}
}

func (r *Request) Method(method string) *Request {
	r.method = method
	return r
}

func (r *Request) ContentType(ct string) *Request {
	r.contentType = ct
	return r
}

func (r *Request) Body(body io.Reader) *Request {
	r.body = body
	return r
}

func (r *Request) Do() (*Result, error) {
	req, err := http.NewRequest(r.method, r.url, r.body)
	if err != nil {
		return nil, err
	}

	if r.contentType != "" {
		req.Header.Set(headerContentType, r.contentType)
	}

	c := http.DefaultClient
	c.Timeout = 30 * time.Second

	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	result := &Result{
		statusCode: resp.StatusCode,
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return result, err
	}
	result.body = body

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return result, ErrBadStatusCode
	}

	return result, nil
}
