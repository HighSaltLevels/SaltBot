package youtube

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
	youtubeResponse interface{}

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

	data, _ := json.Marshal(c.youtubeResponse)

	return &http.Response{
		StatusCode: c.responseCode,
		Body:       io.NopCloser(strings.NewReader(string(data))),
	}, nil
}

func TestGet(t *testing.T) {
	tests := []struct {
		name            string
		commandStr      string
		youtubeResponse interface{}
		getResponseCode int
		// Should the client return an error instead. Omitting == false
		shouldClientError bool
		// Should the ReadCloser response body error. Omitting == false
		shouldReadCloserError bool
		expectedResponse      string
		expectedError         error
	}{
		{
			name:       "Test successful youtube video fetch",
			commandStr: "!youtube query",
			youtubeResponse: YoutubeResponse{
				Items: []YoutubeVideo{
					YoutubeVideo{
						Id: YoutubeId{
							VideoId: "foo",
						},
					},
				},
			},
			getResponseCode:  http.StatusOK,
			expectedResponse: "https://www.youtube.com/watch?v=foo",
			expectedError:    nil,
		},
		{
			name:       "Test successful youtube video fetch with short form",
			commandStr: "!y query",
			youtubeResponse: YoutubeResponse{
				Items: []YoutubeVideo{
					YoutubeVideo{
						Id: YoutubeId{
							VideoId: "foo",
						},
					},
				},
			},
			getResponseCode:  http.StatusOK,
			expectedResponse: "https://www.youtube.com/watch?v=foo",
			expectedError:    nil,
		},
		{
			name:       "Test successful youtube video fetch with specific index",
			commandStr: "!youtube query -i 0",
			youtubeResponse: YoutubeResponse{
				Items: []YoutubeVideo{
					YoutubeVideo{
						Id: YoutubeId{
							VideoId: "foo",
						},
					},
				},
			},
			getResponseCode:  http.StatusOK,
			expectedResponse: "https://www.youtube.com/watch?v=foo",
			expectedError:    nil,
		},
		{
			name:              "Test failed youtube video fetch",
			commandStr:        "!youtube query",
			youtubeResponse:   YoutubeResponse{},
			getResponseCode:   http.StatusOK,
			shouldClientError: true,
			expectedError:     errors.New(expectedError),
		},
		{
			name:            "Test youtube video fetch returns non-200",
			commandStr:      "!youtube query",
			youtubeResponse: YoutubeResponse{},
			getResponseCode: http.StatusInternalServerError,
			expectedError:   fmt.Errorf("received status code %d from youtube", http.StatusInternalServerError),
		},
		{
			name:       "Test successful youtube video fetch but 0 videos returned",
			commandStr: "!youtube query",
			youtubeResponse: YoutubeResponse{
				Items: []YoutubeVideo{},
			},
			getResponseCode:  http.StatusOK,
			expectedResponse: "```No videos for that query :(```",
			expectedError:    nil,
		},
		{
			name:             "Test invalid command string",
			commandStr:       "!youtube",
			youtubeResponse:  YoutubeResponse{},
			getResponseCode:  http.StatusOK,
			expectedResponse: "```Must specify a query like: \"!youtube dog\"```",
			expectedError:    nil,
		},
		{
			name:             "Test pick video by index with negative index",
			commandStr:       "!youtube -i -1",
			youtubeResponse:  YoutubeResponse{},
			getResponseCode:  http.StatusOK,
			expectedResponse: "```Must use a valid number between 0 and 14```",
			expectedError:    nil,
		},
		{
			name:             "Test pick video by index with disallowed index",
			commandStr:       "!youtube -i 15",
			youtubeResponse:  YoutubeResponse{},
			getResponseCode:  http.StatusOK,
			expectedResponse: "```Must use a valid number between 0 and 14```",
			expectedError:    nil,
		},
		{
			name:            "Test youtube video fetch unmarshalable response",
			commandStr:      "!youtube query",
			youtubeResponse: []byte("this can't be marshaled"),
			getResponseCode: http.StatusOK,
			expectedError:   errors.New("failed to unmarshal youtube response"),
		},
		{
			name:                  "Test youtube video fetch unreadable response",
			commandStr:            "!youtube query",
			youtubeResponse:       YoutubeResponse{},
			getResponseCode:       http.StatusOK,
			shouldReadCloserError: true,
			expectedError:         errors.New("failed to read youtube resp"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client = &MockHttpClient{
				expectError:     tt.shouldClientError,
				expectIOError:   tt.shouldReadCloserError,
				responseCode:    tt.getResponseCode,
				youtubeResponse: tt.youtubeResponse,
			}

			msg, err := Get(tt.commandStr)
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
				if msg.Content != tt.expectedResponse {
					t.Errorf("expected message: '%s', but got '%s'", tt.expectedResponse, msg.Content)
				}
			}
		})

	}
}
