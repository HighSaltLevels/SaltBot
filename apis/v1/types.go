package v1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type Poll struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec PollSpec `json:"spec,omitempty"`
}

type PollList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items []Poll `json:"items"`
}

type PollSpec struct {
	Author  string   `json:"author"`
	Channel string   `json:"channel"`
	Prompt  string   `json:"prompt"`
	Choices []string `json:"choices"`
	Expiry  int64    `json:"expiry"`
	Id      string   `json:"unique_id"`
	Votes   []Vote   `json:"votes"`
}

type Vote struct {
	Choice   string   `json:"choice"`
	Choosers []string `json:"choosers"`
}

