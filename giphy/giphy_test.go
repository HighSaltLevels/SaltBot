package giphy

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
	giphyResponse interface{}

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

	data, _ := json.Marshal(c.giphyResponse)

	return &http.Response{
		StatusCode: c.responseCode,
		Body:       io.NopCloser(strings.NewReader(string(data))),
	}, nil
}

func TestGet(t *testing.T) {
	tests := []struct {
		name            string
		commandStr      string
		giphyResponse   interface{}
		getResponseCode int
		// Should the client return an error instead. Omitting == false
		shouldClientError bool
		// Should the ReadCloser response body error. Omitting == false
		shouldReadCloserError bool
		expectedResponse      string
		expectedError         error
	}{
		{
			name:       "Test successful giphy fetch",
			commandStr: "!giphy query",
			giphyResponse: GiphyResponse{
				Data: []GiphyData{
					GiphyData{
						Url: "foo",
					},
				},
			},
			getResponseCode:  http.StatusOK,
			expectedResponse: "foo",
			expectedError:    nil,
		},
		{
			name:       "Test successful giphy fetch with short form",
			commandStr: "!g query",
			giphyResponse: GiphyResponse{
				Data: []GiphyData{
					GiphyData{
						Url: "foo",
					},
				},
			},
			getResponseCode:  http.StatusOK,
			expectedResponse: "foo",
			expectedError:    nil,
		},
		{
			name:       "Test successful giphy fetch with specific index",
			commandStr: "!giphy query -i 0",
			giphyResponse: GiphyResponse{
				Data: []GiphyData{
					GiphyData{
						Url: "foo",
					},
				},
			},
			getResponseCode:  http.StatusOK,
			expectedResponse: "foo",
			expectedError:    nil,
		},
		{
			name:       "Test successful giphy fetch all gifs",
			commandStr: "!giphy query -a",
			giphyResponse: GiphyResponse{
				Data: []GiphyData{
					GiphyData{
						Url: "foo",
					},
					GiphyData{
						Url: "bar",
					},
				},
			},
			getResponseCode:  http.StatusOK,
			expectedResponse: "Here's all the gifs for that query:\nfoo\nbar\n",
			expectedError:    nil,
		},
		{
			name:       "Test successful giphy fetch all gifs with different placement of flag",
			commandStr: "!giphy -a query",
			giphyResponse: GiphyResponse{
				Data: []GiphyData{
					GiphyData{
						Url: "foo",
					},
					GiphyData{
						Url: "bar",
					},
				},
			},
			getResponseCode:  http.StatusOK,
			expectedResponse: "Here's all the gifs for that query:\nfoo\nbar\n",
			expectedError:    nil,
		},
		{
			name:              "Test failed giphy fetch",
			commandStr:        "!giphy query",
			giphyResponse:     GiphyResponse{},
			getResponseCode:   http.StatusOK,
			shouldClientError: true,
			expectedError:     errors.New(expectedError),
		},
		{
			name:              "Test failed giphy all gifs fetch",
			commandStr:        "!giphy query -a",
			giphyResponse:     GiphyResponse{},
			getResponseCode:   http.StatusOK,
			shouldClientError: true,
			expectedError:     errors.New(expectedError),
		},
		{
			name:            "Test giphy fetch non-200 status code",
			commandStr:      "!giphy query",
			giphyResponse:   GiphyResponse{},
			getResponseCode: http.StatusInternalServerError,
			expectedError:   fmt.Errorf("received status code %d from giphy", http.StatusInternalServerError),
		},
		{
			name:             "Test missing command string",
			commandStr:       "!giphy",
			giphyResponse:    GiphyResponse{},
			getResponseCode:  http.StatusOK,
			expectedResponse: "```Must specify giphy query like: \"!giphy dog\"```",
			expectedError:    nil,
		},
		{
			name:             "Test invalid command string",
			commandStr:       "!giphy -i foo",
			giphyResponse:    GiphyResponse{},
			getResponseCode:  http.StatusOK,
			expectedResponse: "```Must use a valid number between 0 and 24```",
			expectedError:    nil,
		},
		{
			name:             "Test index is negative",
			commandStr:       "!giphy -i -1",
			giphyResponse:    GiphyResponse{},
			getResponseCode:  http.StatusOK,
			expectedResponse: "```Must use a valid number between 0 and 24```",
			expectedError:    nil,
		},
		{
			name:             "Test index is too large",
			commandStr:       "!giphy -i 25",
			giphyResponse:    GiphyResponse{},
			getResponseCode:  http.StatusOK,
			expectedResponse: "```Must use a valid number between 0 and 24```",
			expectedError:    nil,
		},
		{
			name:            "Test unmarshalable giphy response",
			commandStr:      "!giphy query",
			giphyResponse:   []byte("this can't be marshaled"),
			getResponseCode: http.StatusOK,
			expectedError:   errors.New("failed to unmarshal giphy response"),
		},
		{
			name:                  "Test giphy unreadable response",
			commandStr:            "!giphy query",
			giphyResponse:         GiphyResponse{},
			getResponseCode:       http.StatusOK,
			shouldReadCloserError: true,
			expectedError:         errors.New("failed to read giphy response"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client = &MockHttpClient{
				expectError:   tt.shouldClientError,
				expectIOError: tt.shouldReadCloserError,
				responseCode:  tt.getResponseCode,
				giphyResponse: tt.giphyResponse,
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
