package reminder

import (
	"errors"
	"strings"
	"testing"

	"github.com/bwmarrin/discordgo"
	"k8s.io/client-go/kubernetes"

	"github.com/highsaltlevels/saltbot/cache"
	"github.com/highsaltlevels/saltbot/testutil"
)

const expectedError string = "expect me"

func TestHandle(t *testing.T) {
	tests := []struct {
		name            string
		reminders       map[string]cache.Reminder
		commandStr      string
		client          kubernetes.Interface
		expectedMessage string
		expectedError   error
	}{
		{
			name:            "test set reminder",
			reminders:       map[string]cache.Reminder{},
			commandStr:      "!remind set do something in 1 second",
			client:          &testutil.MockK8sClient{},
			expectedMessage: "Created reminder with id:",
		},
		{
			name:            "test set reminder missing 'in'",
			reminders:       map[string]cache.Reminder{},
			commandStr:      "!remind set do something 1 second",
			client:          &testutil.MockK8sClient{},
			expectedMessage: helpMessage,
		},
		{
			name:            "test set reminder invalid expiry",
			reminders:       map[string]cache.Reminder{},
			commandStr:      "!remind set do something in 1 minit",
			client:          &testutil.MockK8sClient{},
			expectedMessage: helpMessage,
		},
		{
			name:            "test asking for help",
			reminders:       map[string]cache.Reminder{},
			commandStr:      "!remind help",
			client:          &testutil.MockK8sClient{},
			expectedMessage: helpMessage,
		},
		{
			name:            "test not enough args",
			reminders:       map[string]cache.Reminder{},
			commandStr:      "!remind",
			client:          &testutil.MockK8sClient{},
			expectedMessage: helpMessage,
		},
		{
			name:          "test adding to k8s returns error",
			reminders:     map[string]cache.Reminder{},
			commandStr:    "!remind set do something in 1 second",
			client:        &testutil.MockErrorK8sClient{},
			expectedError: errors.New("error adding reminder to k8s"),
		},
		{
			name:            "test invalid command",
			reminders:       map[string]cache.Reminder{},
			commandStr:      "!remind something else",
			client:          &testutil.MockK8sClient{},
			expectedMessage: helpMessage,
		},
		{
			name: "test list reminders",
			reminders: map[string]cache.Reminder{
				"1234": cache.Reminder{
					Author:  "1234",
					Channel: "1234",
					Expiry:  12345,
					Message: "foo",
					Id:      "1234",
				},
				"5678": cache.Reminder{
					Author:  "1234",
					Channel: "1234",
					Expiry:  12345,
					Message: "bar",
					Id:      "5687",
				},
			},
			commandStr: "!remind list",
			client:     &testutil.MockK8sClient{},
			// "on" will not be in the message if there's no reminders
			expectedMessage: "on",
		},
		{
			name: "test list reminders mismatching author",
			reminders: map[string]cache.Reminder{
				"1234": cache.Reminder{
					Author:  "not the expected one",
					Channel: "1234",
					Expiry:  12345,
					Message: "foo",
					Id:      "1234",
				},
				"5678": cache.Reminder{
					Author:  "not the expected one",
					Channel: "1234",
					Expiry:  12345,
					Message: "bar",
					Id:      "5687",
				},
			},
			commandStr:      "!remind list",
			client:          &testutil.MockK8sClient{},
			expectedMessage: "```Reminders:\n```",
		},
		{
			name:            "test list reminders empty cache",
			reminders:       map[string]cache.Reminder{},
			commandStr:      "!remind list",
			client:          &testutil.MockK8sClient{},
			expectedMessage: "```Reminders:\n```",
		},
		{
			name: "test delete reminder successfully",
			reminders: map[string]cache.Reminder{
				"1234": cache.Reminder{
					Author:  "1234",
					Channel: "1234",
					Expiry:  12345,
					Message: "foo",
					Id:      "1234",
				},
			},
			commandStr:      "!remind delete 1234",
			client:          &testutil.MockK8sClient{},
			expectedMessage: "```Deleted reminder 1234```",
		},
		{
			name:            "test delete not enough args",
			reminders:       map[string]cache.Reminder{},
			commandStr:      "!remind delete",
			client:          &testutil.MockK8sClient{},
			expectedMessage: "To delete a reminder, you must specify the id",
		},
		{
			name:            "test delete reminder doesn't exist",
			reminders:       map[string]cache.Reminder{},
			commandStr:      "!remind delete 1234",
			client:          &testutil.MockK8sClient{},
			expectedMessage: "Either that reminder doesn't exist or you don't",
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
			cache.Cache = cache.NewInMemConfigMapCache(map[string]cache.Poll{}, tt.reminders)
			cache.Client = tt.client.(kubernetes.Interface)

			resp, err := Handle(&msg)

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
