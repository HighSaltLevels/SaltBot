package cache

import (
	"encoding/json"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Reminder struct {
	Author  string      `json:"author"`
	Channel string      `json:"channel"`
	Expiry  int64       `json:"expiry"`
	Message interface{} `json:"msg"`
	Id      string      `json:"id"`
}

func (r *Reminder) FromConfigMap(configMap *corev1.ConfigMap) error {
	jsonData, ok := configMap.Data["json"]
	if !ok {
		return fmt.Errorf("could not find json data in reminder configmap")
	}

	err := json.Unmarshal([]byte(jsonData), &r)
	if err != nil {
		return fmt.Errorf("failed to unmarshal configmap to reminder: %v", err)
	}

	return nil
}

func (r *Reminder) ToConfigMap() (*corev1.ConfigMap, error) {
	bytes, err := json.Marshal(r)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal reminder: %v", err)
	}

	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: "reminder-" + r.Id,
			Labels: map[string]string{
				"author": r.Author,
			},
		},
		Data: map[string]string{
			"json": string(bytes),
		},
	}, nil
}
