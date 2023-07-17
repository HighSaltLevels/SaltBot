package poll

import (
	"errors"
	"strings"
	"testing"

	"github.com/bwmarrin/discordgo"
	"k8s.io/client-go/kubernetes"

	"github.com/highsaltlevels/saltbot/cache"
	"github.com/highsaltlevels/saltbot/testutil"
)

func TestCreate(t *testing.T) {
	tests := []struct {
		name             string
		cache            *cache.ConfigMapCache
		client           kubernetes.Interface
		commandStr       string
		expectedMessages []string
		expectedError    error
	}{
		{
			name:       "Test creating a poll successfully",
			cache:      &cache.ConfigMapCache{},
			client:     &testutil.MockK8sClient{},
			commandStr: "!poll prompt ; choice1 ; choice2 ; ends in 1 minute",
			expectedMessages: []string{
				"prompt",
				"choice1",
				"choice2",
				"Type or DM me \"!vote",
			},
			expectedError: nil,
		},
		{
			name:       "Test creating a poll successfully using short form",
			cache:      &cache.ConfigMapCache{},
			client:     &testutil.MockK8sClient{},
			commandStr: "!p prompt ; choice1 ; choice2 ; ends in 1 minute",
			expectedMessages: []string{
				"prompt",
				"choice1",
				"choice2",
				"Type or DM me \"!vote",
			},
			expectedError: nil,
		},
		{
			name:             "Test sending not enough args",
			cache:            &cache.ConfigMapCache{},
			client:           &testutil.MockK8sClient{},
			commandStr:       "!poll",
			expectedMessages: []string{helpMessage},
			expectedError:    nil,
		},
		{
			name:             "Test poll help",
			cache:            &cache.ConfigMapCache{},
			client:           &testutil.MockK8sClient{},
			commandStr:       "!poll help",
			expectedMessages: []string{helpMessage},
			expectedError:    nil,
		},
		{
			name:             "Test missing ends",
			cache:            &cache.ConfigMapCache{},
			client:           &testutil.MockK8sClient{},
			commandStr:       "!poll prompt ; choice1 ; choice2 ; in 1 minute",
			expectedMessages: []string{helpMessage},
			expectedError:    nil,
		},
		{
			name:             "Test missing in",
			cache:            &cache.ConfigMapCache{},
			client:           &testutil.MockK8sClient{},
			commandStr:       "!poll prompt ; choice1 ; choice2 ; ends 1 minute",
			expectedMessages: []string{helpMessage},
			expectedError:    nil,
		},
		{
			name:             "Test not enough choice args",
			cache:            &cache.ConfigMapCache{},
			client:           &testutil.MockK8sClient{},
			commandStr:       "!poll prompt ; choice1 ; ends in 1 minute",
			expectedMessages: []string{helpMessage},
			expectedError:    nil,
		},
		{
			name:             "Test cannot add poll",
			cache:            &cache.ConfigMapCache{},
			client:           &testutil.MockErrorK8sClient{},
			commandStr:       "!poll prompt ; choice1 ; choice2 ; ends in 1 minute",
			expectedMessages: []string{},
			expectedError:    errors.New("error adding poll to k8s"),
		},
		{
			name:             "Test invalid expiry",
			cache:            &cache.ConfigMapCache{},
			client:           &testutil.MockK8sClient{},
			commandStr:       "!poll prompt ; choice1 ; choice2 ; ends in 1 minit",
			expectedMessages: []string{helpMessage},
			expectedError:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache.Client = tt.client.(kubernetes.Interface)
			cache.Cache = tt.cache
			msg := discordgo.MessageCreate{
				Message: &discordgo.Message{
					Content:   tt.commandStr,
					ChannelID: "1234",
					Author: &discordgo.User{
						ID: "1234",
					},
				},
			}

			resp, err := Create(&msg)
			if tt.expectedError == nil {
				if err != nil {
					t.Fatalf("expected nil error but got: %v", err)
				}

				for _, expectedMessage := range tt.expectedMessages {
					if !strings.Contains(resp.Content, expectedMessage) {
						t.Errorf("expected: '%s' to be in: '%s'", expectedMessage, resp.Content)
					}
				}

			} else {
				if err == nil {
					t.Fatalf("expected error: '%v', but got nil error", tt.expectedError)
				}

				if !strings.Contains(err.Error(), tt.expectedError.Error()) {
					t.Errorf("expected error: '%v', but got: '%v'", tt.expectedError, err)
				}
			}
		})
	}
}

func TestVote(t *testing.T) {
	tests := []struct {
		name            string
		polls           map[string]cache.Poll
		commandStr      string
		client          kubernetes.Interface
		expectedMessage string
		expectedError   error
	}{
		{
			name: "Test vote successfully",
			polls: map[string]cache.Poll{
				"1234": cache.Poll{
					Choices: []string{
						"choice1",
					},
				},
			},
			commandStr:      "!vote 1234 1",
			client:          &testutil.MockK8sClient{},
			expectedMessage: "You have voted for choice1",
			expectedError:   nil,
		},
		{
			name: "Test vote successfully using short form",
			polls: map[string]cache.Poll{
				"1234": cache.Poll{
					Choices: []string{
						"choice1",
					},
				},
			},
			commandStr:      "!v 1234 1",
			client:          &testutil.MockK8sClient{},
			expectedMessage: "You have voted for choice1",
			expectedError:   nil,
		},
		{
			name:            "Test vote not enough args",
			polls:           map[string]cache.Poll{},
			commandStr:      "!v 1234",
			client:          &testutil.MockK8sClient{},
			expectedMessage: voteHelpMessage,
			expectedError:   nil,
		},
		{
			name:            "Test vote invalid arg",
			polls:           map[string]cache.Poll{},
			commandStr:      "!v 1234 foo",
			client:          &testutil.MockK8sClient{},
			expectedMessage: voteHelpMessage,
			expectedError:   nil,
		},
		{
			name:            "Test vote poll doesn't exist",
			polls:           map[string]cache.Poll{},
			commandStr:      "!v 1234 1",
			client:          &testutil.MockK8sClient{},
			expectedMessage: "Poll 1234 does not exist!",
			expectedError:   nil,
		},
		{
			name: "Test vote invalid choice number",
			polls: map[string]cache.Poll{
				"1234": cache.Poll{
					Choices: []string{
						"choice1",
					},
				},
			},
			commandStr:      "!v 1234 2",
			client:          &testutil.MockK8sClient{},
			expectedMessage: "No such choice number: 2",
			expectedError:   nil,
		},
		{
			name: "Test vote selecting different choice",
			polls: map[string]cache.Poll{
				"1234": cache.Poll{
					Choices: []string{
						"choice1",
						"choice2",
					},
					Votes: map[string][]interface{}{
						"0": []interface{}{
							"user",
							"other user",
						},
					},
				},
			},
			commandStr:      "!vote 1234 2",
			client:          &testutil.MockK8sClient{},
			expectedMessage: "You have voted for choice2",
			expectedError:   nil,
		},
		{
			name: "Test vote kubernetes threw error",
			polls: map[string]cache.Poll{
				"1234": cache.Poll{
					Choices: []string{
						"choice1",
					},
				},
			},
			commandStr:      "!v 1234 1",
			client:          &testutil.MockErrorK8sClient{},
			expectedMessage: "",
			expectedError:   errors.New("failed to update poll"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := discordgo.MessageCreate{
				Message: &discordgo.Message{
					Content:   tt.commandStr,
					ChannelID: "1234",
					Author: &discordgo.User{
						ID:       "1234",
						Username: "user",
					},
				},
			}

			cache.Cache = cache.NewInMemConfigMapCache(tt.polls, map[string]cache.Reminder{})
			cache.Client = tt.client.(kubernetes.Interface)
			resp, err := Vote(&msg)

			if tt.expectedError == nil {
				if err != nil {
					t.Fatalf("expected nil error but got: %v", err)
				}

				if !strings.Contains(resp.Content, tt.expectedMessage) {
					t.Errorf("expected: '%s' to be in: '%s'", tt.expectedMessage, resp.Content)
				}

			} else {
				if err == nil {
					t.Fatalf("expected error: '%v', but got nil error", tt.expectedError)
				}

				if !strings.Contains(err.Error(), tt.expectedError.Error()) {
					t.Errorf("expected error: '%v', but got: '%v'", tt.expectedError, err)
				}
			}
		})
	}
}
