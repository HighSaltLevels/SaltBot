package jeopardy

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/highsaltlevels/saltbot/util"
)

const expectedError = "expect me"

type MockFailedReadCloser struct {
	io.Reader
	io.Closer
}

func (rc MockFailedReadCloser) Read(p []byte) (n int, err error) {
	return 0, errors.New(expectedError)
}

func (rc MockFailedReadCloser) Close() error {
	return errors.New(expectedError)
}

type MockHttpClient struct {
	util.HttpClientInterface

	// The response to return if successful
	jeopardyResponse interface{}

	// Used to determine if the mock client should return error or not
	expectError bool

	// Used to determine if mock client should return object that can't be read
	expectIOError bool

	// The response code to be used in the http response object
	responseCode int
}

func (c *MockHttpClient) Get(url string) (*http.Response, error) {
	if c.expectError {
		return nil, errors.New(expectedError)
	}

	if c.expectIOError {
		return &http.Response{
			StatusCode: c.responseCode,
			Body:       MockFailedReadCloser{},
		}, nil
	}

	data, _ := json.Marshal(c.jeopardyResponse)

	return &http.Response{
		StatusCode: c.responseCode,
		Body:       io.NopCloser(strings.NewReader(string(data))),
	}, nil
}

func TestGet(t *testing.T) {
	tests := []struct {
		name             string
		jeopardyResponse interface{}
		getResponseCode  int
		// Should the client return an error instead. Omitting == false
		shouldClientError bool
		// Should the ReadCloser response body error. Omitting == false
		shouldReadCloserError bool
		expectedRespStrings   []string
		expectedError         error
	}{
		{
			name: "Test successful jeopardy question fetch",
			jeopardyResponse: JeopardyResponse{
				Title: "expected title",
				Clues: []Clue{
					Clue{
						Question: "expected question 1",
						Answer:   "expected answer 1",
					},
					Clue{
						Question: "expected question 2",
						Answer:   "expected answer 2",
					},
				},
			},
			getResponseCode: http.StatusOK,
			expectedRespStrings: []string{
				"expected title",
				"expected question 1",
				"expected answer 1",
				"expected question 2",
				"expected answer 2",
			},
			expectedError: nil,
		},
		{
			name:              "Test failed jeopardy question fetch",
			jeopardyResponse:  JeopardyResponse{},
			getResponseCode:   http.StatusOK,
			shouldClientError: true,
			expectedError:     errors.New("error getting jeopardy question"),
		},
		{
			name:                  "Test failed readon jeopardy response body",
			jeopardyResponse:      JeopardyResponse{},
			getResponseCode:       http.StatusOK,
			shouldReadCloserError: true,
			expectedError:         errors.New("failed to read jeopardy response"),
		},
		{
			name:             "Test jeopardy returns non-200 statuscode",
			jeopardyResponse: JeopardyResponse{},
			getResponseCode:  http.StatusInternalServerError,
			expectedError:    fmt.Errorf("got %d status code getting jeopardy question", http.StatusInternalServerError),
		},
		{
			name:             "Test jeopardy failed unmarshaling response",
			jeopardyResponse: []byte("you can't unmarshal me :)"),
			getResponseCode:  http.StatusOK,
			expectedError:    errors.New("failed to unmarshal jeopardy response"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client = &MockHttpClient{
				expectError:      tt.shouldClientError,
				expectIOError:    tt.shouldReadCloserError,
				responseCode:     tt.getResponseCode,
				jeopardyResponse: tt.jeopardyResponse,
			}

			msg, err := Get()
			if tt.expectedError != nil {
				if err == nil {
					t.Errorf("expected error '%v' to be returned but was nil", tt.expectedError)
				}
				if !strings.Contains(err.Error(), tt.expectedError.Error()) {
					t.Errorf("expected error '%v', but got error '%v'", tt.expectedError, err)
				}
			} else {
				if err != nil {
					t.Fatalf("expected no error but got error: '%v'", err)
				}
				for _, resp := range tt.expectedRespStrings {
					if !strings.Contains(msg.Content, resp) {
						t.Errorf("expected message: '%s', but got '%s'", resp, msg.Content)
					}
				}
			}
		})
	}
}
