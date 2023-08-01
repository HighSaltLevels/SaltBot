package cache

import (
	"encoding/json"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Poll struct {
	Author  string                   `json:"author"`
	Channel string                   `json:"channel"`
	Prompt  string                   `json:"prompt"`
	Choices []string                 `json:"choices"`
	Expiry  int64                    `json:"expiry"`
	Id      string                   `json:"id"`
	Votes   map[string][]interface{} `json:"votes"`
}

func (p *Poll) FromConfigMap(configMap *corev1.ConfigMap) error {
	jsonData, ok := configMap.Data["json"]
	if !ok {
		return fmt.Errorf("could not find json data in poll configmap")
	}

	err := json.Unmarshal([]byte(jsonData), &p)
	if err != nil {
		return fmt.Errorf("failed to unmarshal configmap to poll: %v", err)
	}

	return nil
}

func (p *Poll) ToConfigMap() (*corev1.ConfigMap, error) {
	bytes, err := json.Marshal(p)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal poll: %v", err)
	}

	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: "poll-" + p.Id,
			Labels: map[string]string{
				"author": p.Author,
			},
		},
		Data: map[string]string{
			"json": string(bytes),
		},
	}, nil
}
