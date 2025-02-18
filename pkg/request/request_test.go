package request

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequest(t *testing.T) {
	testCases := []struct {
		name           string
		method         string
		url            string
		contentType    string
		body           io.Reader
		expectedStatus int
		expectedBody   []byte
		expectedError  error
	}{
		{
			name:           "Successful GET request",
			method:         http.MethodGet,
			url:            "/test",
			contentType:    "",
			body:           nil,
			expectedStatus: http.StatusOK,
			expectedBody:   []byte(`{"message":"success"}`),
			expectedError:  nil,
		},
		{
			name:           "POST request with body",
			method:         http.MethodPost,
			url:            "/test",
			contentType:    "application/json",
			body:           bytes.NewBufferString(`{"key":"value"}`),
			expectedStatus: http.StatusCreated,
			expectedBody:   []byte(`{"message":"created"}`),
			expectedError:  nil,
		},
		{
			name:           "Bad status code",
			method:         http.MethodGet,
			url:            "/bad",
			contentType:    "",
			body:           nil,
			expectedStatus: http.StatusNotFound,
			expectedBody:   []byte(`{"error":"not found"}`),
			expectedError:  ErrBadStatusCode,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.expectedStatus)
				w.Write(tc.expectedBody)
			}))
			defer ts.Close()

			req := New(ts.URL + tc.url).
				Method(tc.method).
				ContentType(tc.contentType).
				Body(tc.body)

			result, err := req.Do()

			if tc.expectedError != nil {
				require.Error(t, err)
				assert.Equal(t, tc.expectedError, err)
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, tc.expectedStatus, result.StatusCode())
			assert.Equal(t, tc.expectedBody, result.Body())
		})
	}
}
