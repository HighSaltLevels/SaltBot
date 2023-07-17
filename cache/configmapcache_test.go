package cache

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/highsaltlevels/saltbot/testutil"
)

func TestAddConfigMap(t *testing.T) {
	tests := []struct {
		name             string
		configMap        interface{}
		cache            *ConfigMapCache
		expectedPoll     *Poll
		expectedReminder *Reminder
	}{
		{
			name: "Test successfully adding poll",
			configMap: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name: "poll-foo",
				},
				Data: map[string]string{
					"json": `{"author":"1234","channel":"1234","prompt":"prompt","choices":["choice1","choice2"],"expiry":1234,"id":"1234","votes":{"0":["chooser"],"1":[]}}`,
				},
			},
			cache: &ConfigMapCache{
				polls:     map[string]Poll{},
				reminders: map[string]Reminder{},
			},
			expectedPoll: &Poll{
				Author:  "1234",
				Channel: "1234",
				Prompt:  "prompt",
				Choices: []string{"choice1", "choice2"},
				Expiry:  1234,
				Id:      "1234",
				Votes: map[string][]interface{}{
					"0": []interface{}{"chooser"},
					"1": []interface{}{},
				},
			},
		},
		{
			name: "Test failure adding poll invalid JSON",
			configMap: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name: "poll-foo",
				},
				Data: map[string]string{
					"json": "i am not json :)",
				},
			},
			cache: &ConfigMapCache{
				polls:     map[string]Poll{},
				reminders: map[string]Reminder{},
			},
		},
		{
			name: "Test failure adding poll json missing",
			configMap: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name: "poll-foo",
				},
				Data: map[string]string{},
			},
			cache: &ConfigMapCache{
				polls:     map[string]Poll{},
				reminders: map[string]Reminder{},
			},
		},
		{
			name: "Test successfully adding reminder",
			configMap: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name: "reminder-foo",
				},
				Data: map[string]string{
					"json": `{"author":"1234","channel":"1234","expiry":1234,"msg":"bloop","id":"1234"}`,
				},
			},
			cache: &ConfigMapCache{
				polls:     map[string]Poll{},
				reminders: map[string]Reminder{},
			},
			expectedReminder: &Reminder{
				Author:  "1234",
				Channel: "1234",
				Expiry:  1234,
				Message: "bloop",
				Id:      "1234",
			},
		},
		{
			name: "Test failure adding reminder invalid JSON",
			configMap: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name: "reminder-foo",
				},
				Data: map[string]string{
					"json": "i am not json :)",
				},
			},
			cache: &ConfigMapCache{
				polls:     map[string]Poll{},
				reminders: map[string]Reminder{},
			},
		},
		{
			name: "Test failure adding reminder json missing",
			configMap: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name: "reminder-foo",
				},
				Data: map[string]string{},
			},
			cache: &ConfigMapCache{
				polls:     map[string]Poll{},
				reminders: map[string]Reminder{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set the cache before adding
			Cache = tt.cache
			Cache.addConfigMap(tt.configMap)
			if tt.expectedPoll != nil {
				if len(Cache.polls) != 1 {
					t.Fatalf("should have 1 poll, but got no polls")
				}
				for _, poll := range Cache.polls {
					validatePoll(t, tt.expectedPoll, &poll)
				}
			} else {
				if len(Cache.polls) != 0 {
					t.Fatalf("expected no polls but got: %d", len(Cache.polls))
				}
			}
			if tt.expectedReminder != nil {
				if len(Cache.reminders) != 1 {
					t.Fatalf("should have 1 reminder, but got no reminders")
				}
				for _, reminder := range Cache.reminders {
					validateReminder(t, tt.expectedReminder, &reminder)
				}
			} else {
				if len(Cache.reminders) != 0 {
					t.Fatalf("expected no reminders but got: %d", len(Cache.reminders))
				}
			}
		})
	}
}

func TestUpdateConfigMap(t *testing.T) {
	tests := []struct {
		name             string
		configMap        interface{}
		cache            *ConfigMapCache
		expectedPoll     *Poll
		expectedReminder *Reminder
	}{
		{
			name: "Test successfully updating poll",
			configMap: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name: "poll-foo",
				},
				Data: map[string]string{
					"json": `{"author":"1234","channel":"1234","prompt":"prompt","choices":["choice1","choice2"],"expiry":1234,"id":"1234","votes":{"0":["chooser"],"1":[]}}`,
				},
			},
			cache: &ConfigMapCache{
				polls: map[string]Poll{
					"1234": Poll{
						Author:  "1234",
						Channel: "1234",
						Prompt:  "prompt",
						Choices: []string{"choice1", "choice2"},
						Expiry:  1234,
						Id:      "1234",
						Votes: map[string][]interface{}{
							"0": []interface{}{},
							"1": []interface{}{"chooser"},
						},
					},
				},
				reminders: map[string]Reminder{},
			},
			expectedPoll: &Poll{
				Author:  "1234",
				Channel: "1234",
				Prompt:  "prompt",
				Choices: []string{"choice1", "choice2"},
				Expiry:  1234,
				Id:      "1234",
				Votes: map[string][]interface{}{
					"0": []interface{}{"chooser"},
					"1": []interface{}{},
				},
			},
		},
		{
			name: "Test failure updating poll invalid JSON",
			configMap: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name: "poll-foo",
				},
				Data: map[string]string{
					"json": "i am not json :)",
				},
			},
			cache: &ConfigMapCache{
				polls:     map[string]Poll{},
				reminders: map[string]Reminder{},
			},
		},
		{
			name: "Test successfully updating reminder",
			configMap: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name: "reminder-foo",
				},
				Data: map[string]string{
					"json": `{"author":"1234","channel":"1234","expiry":1234,"msg":"bloop","id":"1234"}`,
				},
			},
			cache: &ConfigMapCache{
				polls: map[string]Poll{},
				reminders: map[string]Reminder{
					"1234": Reminder{
						Author:  "1234",
						Channel: "1234",
						Expiry:  1234,
						Message: "something else",
						Id:      "1234",
					},
				},
			},
			expectedReminder: &Reminder{
				Author:  "1234",
				Channel: "1234",
				Expiry:  1234,
				Message: "bloop",
				Id:      "1234",
			},
		},
		{
			name: "Test failure updating reminder invalid JSON",
			configMap: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name: "reminder-foo",
				},
				Data: map[string]string{
					"json": "i am not json :)",
				},
			},
			cache: &ConfigMapCache{
				polls:     map[string]Poll{},
				reminders: map[string]Reminder{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set the cache before adding
			Cache = tt.cache
			// We ignore the old object, so let's just reuse tt.configMap
			Cache.updateConfigMap(tt.configMap, tt.configMap)
			if tt.expectedPoll != nil {
				if len(Cache.polls) != 1 {
					t.Fatalf("should have 1 poll, but got no polls")
				}
				for _, poll := range Cache.polls {
					validatePoll(t, tt.expectedPoll, &poll)
				}
			} else {
				if len(Cache.polls) != 0 {
					t.Fatalf("expected no polls but got: %d", len(Cache.polls))
				}
			}
			if tt.expectedReminder != nil {
				if len(Cache.reminders) != 1 {
					t.Fatalf("should have 1 reminder, but got no reminders")
				}
				for _, reminder := range Cache.reminders {
					validateReminder(t, tt.expectedReminder, &reminder)
				}
			} else {
				if len(Cache.reminders) != 0 {
					t.Fatalf("expected no reminders but got: %d", len(Cache.reminders))
				}
			}
		})
	}
}

func TestDeleteConfigMap(t *testing.T) {
	tests := []struct {
		name      string
		configMap interface{}
		cache     *ConfigMapCache
	}{
		{
			name: "Test successfully deleting poll",
			configMap: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name: "poll-foo",
				},
				Data: map[string]string{
					"json": `{"author":"1234","channel":"1234","prompt":"prompt","choices":["choice1","choice2"],"expiry":1234,"id":"1234","votes":{"0":["chooser"],"1":[]}}`,
				},
			},
			cache: &ConfigMapCache{
				polls: map[string]Poll{
					"1234": Poll{
						Author:  "1234",
						Channel: "1234",
						Prompt:  "prompt",
						Choices: []string{"choice1", "choice2"},
						Expiry:  1234,
						Id:      "1234",
						Votes: map[string][]interface{}{
							"0": []interface{}{"chooser"},
							"1": []interface{}{},
						},
					},
				},
				reminders: map[string]Reminder{},
			},
		},
		{
			name: "Test failure deleting poll invalid name",
			configMap: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name: "poll",
				},
				Data: map[string]string{
					"json": "",
				},
			},
			cache: &ConfigMapCache{
				polls:     map[string]Poll{},
				reminders: map[string]Reminder{},
			},
		},
		{
			name: "Test successfully deleting reminder",
			configMap: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name: "reminder-foo",
				},
				Data: map[string]string{
					"json": `{"author":"1234","channel":"1234","expiry":1234,"msg":"bloop","id":"1234"}`,
				},
			},
			cache: &ConfigMapCache{
				polls: map[string]Poll{},
				reminders: map[string]Reminder{
					"1234": Reminder{
						Author:  "1234",
						Channel: "1234",
						Expiry:  1234,
						Message: "something else",
						Id:      "1234",
					},
				},
			},
		},
		{
			name: "Test failure deleting reminder invalid name",
			configMap: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name: "reminder",
				},
				Data: map[string]string{
					"json": "",
				},
			},
			cache: &ConfigMapCache{
				polls:     map[string]Poll{},
				reminders: map[string]Reminder{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set the cache before adding
			Cache = tt.cache
			// We ignore the old object, so let's just reuse tt.configMap
			Cache.deleteConfigMap(tt.configMap)
		})
	}
}

func TestAddPoll(t *testing.T) {
	tests := []struct {
		name          string
		client        kubernetes.Interface
		poll          *Poll
		expectedError error
	}{
		{
			name:   "Test adding poll successfully",
			client: &testutil.MockK8sClient{},
			poll: &Poll{
				Id:     "1234",
				Author: "1234",
			},
			expectedError: nil,
		},
		{
			name:   "Test failed adding poll to k8s",
			client: &testutil.MockErrorK8sClient{},
			poll: &Poll{
				Id:     "1234",
				Author: "1234",
			},
			expectedError: errors.New(testutil.ExpectedError),
		},
		{
			name:   "Test failed adding invalid poll",
			client: &testutil.MockErrorK8sClient{},
			poll: &Poll{
				Id:     "1234",
				Author: "1234",
				Votes: map[string][]interface{}{
					"1234": []interface{}{
						make(chan bool),
					},
				},
			},
			expectedError: errors.New("failed to marshal poll"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Client = tt.client
			c := ConfigMapCache{}

			err := c.AddPoll(tt.poll)
			if tt.expectedError == nil {
				if err != nil {
					t.Errorf("expected nil error, but got: %v", err)
				}
			} else {
				if err == nil {
					t.Fatalf("got nil error but expected: %v", err)
				}
				if !strings.Contains(err.Error(), tt.expectedError.Error()) {
					t.Errorf("expected: \"%v\" but got: \"%v\"", tt.expectedError, err)
				}
			}
		})
	}
}

func TestUpdatePoll(t *testing.T) {
	tests := []struct {
		name          string
		client        kubernetes.Interface
		poll          *Poll
		expectedError error
	}{
		{
			name:   "Test updating poll successfully",
			client: &testutil.MockK8sClient{},
			poll: &Poll{
				Id:     "1234",
				Author: "1234",
			},
			expectedError: nil,
		},
		{
			name:   "Test failed updating poll in k8s",
			client: &testutil.MockErrorK8sClient{},
			poll: &Poll{
				Id:     "1234",
				Author: "1234",
			},
			expectedError: errors.New(testutil.ExpectedError),
		},
		{
			name:   "Test failed updating invalid poll",
			client: &testutil.MockErrorK8sClient{},
			poll: &Poll{
				Id:     "1234",
				Author: "1234",
				Votes: map[string][]interface{}{
					"1234": []interface{}{
						make(chan bool),
					},
				},
			},
			expectedError: errors.New("failed to marshal poll"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Client = tt.client
			c := ConfigMapCache{}

			err := c.UpdatePoll(tt.poll)
			if tt.expectedError == nil {
				if err != nil {
					t.Errorf("expected nil error, but got: %v", err)
				}
			} else {
				if err == nil {
					t.Fatalf("got nil error but expected: %v", err)
				}
				if !strings.Contains(err.Error(), tt.expectedError.Error()) {
					t.Errorf("expected: \"%v\" but got: \"%v\"", tt.expectedError, err)
				}
			}
		})
	}
}

func TestAddReminder(t *testing.T) {
	tests := []struct {
		name          string
		client        kubernetes.Interface
		reminder      *Reminder
		expectedError error
	}{
		{
			name:          "Test adding reminder successfully",
			client:        &testutil.MockK8sClient{},
			reminder:      &Reminder{},
			expectedError: nil,
		},
		{
			name:          "Test failed adding reminder to k8s",
			client:        &testutil.MockErrorK8sClient{},
			reminder:      &Reminder{},
			expectedError: errors.New(testutil.ExpectedError),
		},
		{
			name:   "Test failed adding invalid reminder",
			client: &testutil.MockErrorK8sClient{},
			reminder: &Reminder{
				Message: make(chan bool),
			},
			expectedError: errors.New("failed to convert reminder to configMap"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Client = tt.client
			c := ConfigMapCache{}

			err := c.AddReminder(tt.reminder, "user")
			if tt.expectedError == nil {
				if err != nil {
					t.Errorf("expected nil error, but got: %v", err)
				}
			} else {
				if err == nil {
					t.Fatalf("got nil error but expected: %v", err)
				}
				if !strings.Contains(err.Error(), tt.expectedError.Error()) {
					t.Errorf("expected: \"%v\" but got: \"%v\"", tt.expectedError, err)
				}
			}
		})
	}
}

func TestListPolls(t *testing.T) {
	expected := map[string]Poll{
		"1234": Poll{
			Id: "1234",
		},
		"5678": Poll{
			Id: "5678",
		},
	}
	c := NewInMemConfigMapCache(expected, map[string]Reminder{})
	actual := c.ListPolls()
	for _, id := range []string{"1234", "5678"} {
		if actual[id].Id != expected[id].Id {
			t.Errorf("expected Id %s but got %s", expected[id].Id, actual[id].Id)
		}
	}
}

func TestGetPoll(t *testing.T) {
	tests := []struct {
		name         string
		cache        *ConfigMapCache
		id           string
		author       string
		expectedPoll *Poll
	}{
		{
			name: "test successfully getting poll",
			cache: NewInMemConfigMapCache(
				map[string]Poll{
					"1234": Poll{
						Id:     "1234",
						Author: "5678",
					},
				},
				map[string]Reminder{},
			),
			expectedPoll: &Poll{
				Id: "1234",
			},
			id:     "1234",
			author: "5678",
		},
		{
			name: "test failing to get poll",
			cache: NewInMemConfigMapCache(
				map[string]Poll{},
				map[string]Reminder{},
			),
			expectedPoll: nil,
			id:           "1234",
			author:       "5678",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.cache.GetPoll(tt.id, tt.author)
			if tt.expectedPoll == nil {
				if actual != nil {
					t.Errorf("expected nil poll but got: %v", actual)
				}
			} else {
				if actual == nil {
					t.Fatalf("expected poll: %v, but got nil poll", tt.expectedPoll)
				}
				if tt.expectedPoll.Id != actual.Id {
					t.Errorf("expected id: %s, but got: %s", tt.expectedPoll.Id, actual.Id)
				}
			}
		})
	}
}

func TestListReminders(t *testing.T) {
	expected := map[string]Reminder{
		"1234": Reminder{
			Id: "1234",
		},
		"5678": Reminder{
			Id: "5678",
		},
	}
	c := NewInMemConfigMapCache(map[string]Poll{}, expected)
	actual := c.ListReminders()
	for _, id := range []string{"1234", "5678"} {
		if actual[id].Id != expected[id].Id {
			t.Errorf("expected Id %s but got %s", expected[id].Id, actual[id].Id)
		}
	}
}

func TestGetReminder(t *testing.T) {
	tests := []struct {
		name             string
		cache            *ConfigMapCache
		id               string
		author           string
		expectedReminder *Reminder
	}{
		{
			name: "test successfully getting reminder",
			cache: NewInMemConfigMapCache(
				map[string]Poll{},
				map[string]Reminder{
					"1234": Reminder{
						Id:     "1234",
						Author: "5678",
					},
				},
			),
			expectedReminder: &Reminder{
				Id: "1234",
			},
			id:     "1234",
			author: "5678",
		},
		{
			name: "test failing to get reminder",
			cache: NewInMemConfigMapCache(
				map[string]Poll{},
				map[string]Reminder{},
			),
			expectedReminder: nil,
			id:               "1234",
			author:           "5678",
		},
		{
			name: "test failing to get reminder with different author",
			cache: NewInMemConfigMapCache(
				map[string]Poll{},
				map[string]Reminder{
					"1234": Reminder{
						Id:     "1234",
						Author: "5678",
					},
				},
			),
			expectedReminder: nil,
			id:               "1234",
			author:           "not the right one",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.cache.GetReminder(tt.id, tt.author)
			if tt.expectedReminder == nil {
				if actual != nil {
					t.Errorf("expected nil poll but got: %v", actual)
				}
			} else {
				if actual == nil {
					t.Fatalf("expected poll: %v, but got nil poll", tt.expectedReminder)
				}
				if tt.expectedReminder.Id != actual.Id {
					t.Errorf("expected id: %s, but got: %s", tt.expectedReminder.Id, actual.Id)
				}
			}
		})
	}
}

func TestDelete(t *testing.T) {
	tests := []struct {
		name   string
		client kubernetes.Interface
		cache  *ConfigMapCache
	}{
		{
			name:   "Test delete poll from cache successfully",
			client: &testutil.MockK8sClient{},
			cache: NewInMemConfigMapCache(
				map[string]Poll{
					"1234": Poll{
						Id:     "1234",
						Author: "5678",
					},
				},
				map[string]Reminder{},
			),
		},
		{
			name:   "Test failed to delete poll from cache",
			client: &testutil.MockErrorK8sClient{},
			cache: NewInMemConfigMapCache(
				map[string]Poll{
					"1234": Poll{
						Id:     "1234",
						Author: "5678",
					},
				},
				map[string]Reminder{},
			),
		},
		{
			name:   "Test delete reminder from cache successfully",
			client: &testutil.MockK8sClient{},
			cache: NewInMemConfigMapCache(
				map[string]Poll{},
				map[string]Reminder{
					"1234": Reminder{
						Id:     "1234",
						Author: "5678",
					},
				},
			),
		},
		{
			name:   "Test failed to delete reminder from cache",
			client: &testutil.MockErrorK8sClient{},
			cache: NewInMemConfigMapCache(
				map[string]Poll{},
				map[string]Reminder{
					"1234": Reminder{
						Id:     "1234",
						Author: "5678",
					},
				},
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Client = tt.client
			Cache = tt.cache

			Cache.Delete("1234")
		})
	}
}

func validatePoll(t *testing.T, expected, actual *Poll) {
	if actual.Author != expected.Author {
		t.Errorf("incorrect author. Expected: %s, got: %s", expected.Author, actual.Author)
	}
	if actual.Channel != expected.Channel {
		t.Errorf("incorrect channel. Expected: %s, got: %s", expected.Channel, actual.Channel)
	}
	if actual.Prompt != expected.Prompt {
		t.Errorf("incorrect prompt. Expected: %s, got: %s", expected.Prompt, actual.Prompt)
	}
	if !reflect.DeepEqual(actual.Choices, expected.Choices) {
		t.Errorf("incorrect choices. Expected: %v, got %v", expected.Choices, actual.Choices)
	}
	if actual.Expiry != expected.Expiry {
		t.Errorf("incorrect expiry. Expected: %d, got: %d", expected.Expiry, actual.Expiry)
	}
	if actual.Id != expected.Id {
		t.Errorf("incorrect id. Expected: %s, got: %s", expected.Id, actual.Id)
	}
	if !reflect.DeepEqual(actual.Votes, expected.Votes) {
		t.Errorf("incorrect votes. Expected: %v, got %v", expected.Votes, actual.Votes)
	}
}

func validateReminder(t *testing.T, expected, actual *Reminder) {
	if actual.Author != expected.Author {
		t.Errorf("incorrect author. Expected: %s, got: %s", expected.Author, actual.Author)
	}
	if actual.Channel != expected.Channel {
		t.Errorf("incorrect channel. Expected: %s, got: %s", expected.Channel, actual.Channel)
	}
	if actual.Expiry != expected.Expiry {
		t.Errorf("incorrect expiry. Expected: %d, got: %d", expected.Expiry, actual.Expiry)
	}
	if actual.Message != expected.Message {
		t.Errorf("incorrect message. Expected: %s, got: %s", expected.Message, actual.Message)
	}
	if actual.Id != expected.Id {
		t.Errorf("incorrect id. Expected: %s, got: %s", expected.Id, actual.Id)
	}
}
