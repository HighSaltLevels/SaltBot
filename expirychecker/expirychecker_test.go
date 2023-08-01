package expirychecker

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/bwmarrin/discordgo"
	"k8s.io/client-go/kubernetes"

	"github.com/highsaltlevels/saltbot/cache"
	"github.com/highsaltlevels/saltbot/testutil"
)

type MockDiscordSession struct {
	// error to return. Leave this as nil to return nil as error
	err error

	// used to save what would have been sent as a message
	SentMessage string
}

func (m *MockDiscordSession) ChannelMessageSend(channelID string, content string, options ...discordgo.RequestOption) (*discordgo.Message, error) {
	if m.err != nil {
		return nil, m.err
	}

	m.SentMessage = content
	return nil, m.err
}

func TestPollerLoop(t *testing.T) {
	tests := []struct {
		name                 string
		session              MockDiscordSession
		client               kubernetes.Interface
		cache                *cache.ConfigMapCache
		expectedMessageParts []string
	}{
		{
			name:    "Test send poll successfully",
			session: MockDiscordSession{},
			client:  &testutil.MockK8sClient{},
			cache: cache.NewInMemConfigMapCache(
				map[string]cache.Poll{
					"1234": cache.Poll{
						Prompt: "prompt",
						Votes: map[string][]interface{}{
							"0": []interface{}{"person1", "persion2"},
							"1": []interface{}{"person3"},
						},
						Choices: []string{
							"choice1",
							"choice2",
						},
						Expiry: 0,
					},
				},
				map[string]cache.Reminder{},
			),
			expectedMessageParts: []string{
				"prompt",
				"choice1",
				"choice2",
				"67%",
				"33%",
			},
		},
		{
			name:    "Test send poll successfully but no one voted",
			session: MockDiscordSession{},
			client:  &testutil.MockK8sClient{},
			cache: cache.NewInMemConfigMapCache(
				map[string]cache.Poll{
					"1234": cache.Poll{
						Prompt: "prompt",
						Votes: map[string][]interface{}{
							"0": []interface{}{},
							"1": []interface{}{},
						},
						Choices: []string{
							"choice1",
							"choice2",
						},
						Expiry: 0,
					},
				},
				map[string]cache.Reminder{},
			),
			expectedMessageParts: []string{"No one voted on this poll"},
		},
		{
			name:    "Test send reminder successfully",
			session: MockDiscordSession{},
			client:  &testutil.MockK8sClient{},
			cache: cache.NewInMemConfigMapCache(
				map[string]cache.Poll{},
				map[string]cache.Reminder{
					"1234": cache.Reminder{
						Message: "message",
						Expiry:  0,
					},
				},
			),
			expectedMessageParts: []string{"message"},
		},
		{
			name:    "Test send poll discord session error",
			session: MockDiscordSession{err: errors.New("foo")},
			client:  &testutil.MockK8sClient{},
			cache: cache.NewInMemConfigMapCache(
				map[string]cache.Poll{
					"1234": cache.Poll{
						Prompt: "prompt",
						Votes: map[string][]interface{}{
							"0": []interface{}{},
							"1": []interface{}{},
						},
						Choices: []string{
							"choice1",
							"choice2",
						},
						Expiry: 0,
					},
				},
				map[string]cache.Reminder{},
			),
			expectedMessageParts: []string{},
		},
		{
			name:    "Test send reminder discord session error",
			session: MockDiscordSession{err: errors.New("foo")},
			client:  &testutil.MockK8sClient{},
			cache: cache.NewInMemConfigMapCache(
				map[string]cache.Poll{},
				map[string]cache.Reminder{
					"1234": cache.Reminder{
						Message: "message",
						Expiry:  0,
					},
				},
			),
			expectedMessageParts: []string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache.Cache = tt.cache
			cache.Client = tt.client
			ctx, cancel := context.WithCancel(context.Background())

			poller := NewPoller(&tt.session, ctx)
			go poller.Loop()
			// Give 1.1 seconds to make sure the poll/reminder gets picked up.
			time.Sleep(1100 * time.Millisecond)
			cancel()

			if len(tt.expectedMessageParts) == 0 && tt.session.SentMessage != "" {
				t.Fatalf("expected no message to be sent, but got: %s", tt.session.SentMessage)
			}

			for _, msg := range tt.expectedMessageParts {
				if !strings.Contains(tt.session.SentMessage, msg) {
					t.Errorf("expected \"%s\" to be in \"%s\"", msg, tt.session.SentMessage)
				}
			}
		})
	}
}
