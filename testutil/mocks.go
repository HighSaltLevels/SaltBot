package testutil

import (
	"context"
	"errors"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

const ExpectedError string = "expect me"

type MockConfigMapClient struct {
	corev1.ConfigMapInterface
}

func (m MockConfigMapClient) Create(ctx context.Context, configMap *v1.ConfigMap, opts metav1.CreateOptions) (*v1.ConfigMap, error) {
	return nil, nil
}

func (m MockConfigMapClient) Update(ctx context.Context, configMap *v1.ConfigMap, opts metav1.UpdateOptions) (*v1.ConfigMap, error) {
	return nil, nil
}

func (m MockConfigMapClient) Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error {
	return nil
}

type MockCoreV1 struct {
	corev1.CoreV1Interface
}

func (m MockCoreV1) ConfigMaps(namespace string) corev1.ConfigMapInterface {
	return MockConfigMapClient{}
}

type MockK8sClient struct {
	kubernetes.Interface
}

func (m MockK8sClient) CoreV1() corev1.CoreV1Interface {
	return MockCoreV1{}
}

type MockErrorConfigMapClient struct {
	corev1.ConfigMapInterface
}

func (m MockErrorConfigMapClient) Create(ctx context.Context, configMap *v1.ConfigMap, opts metav1.CreateOptions) (*v1.ConfigMap, error) {
	return nil, errors.New(ExpectedError)
}

func (m MockErrorConfigMapClient) Update(ctx context.Context, configMap *v1.ConfigMap, opts metav1.UpdateOptions) (*v1.ConfigMap, error) {
	return nil, errors.New(ExpectedError)
}

func (m MockErrorConfigMapClient) Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error {
	return errors.New(ExpectedError)
}

type MockErrorCoreV1 struct {
	corev1.CoreV1Interface
}

func (m MockErrorCoreV1) ConfigMaps(namespace string) corev1.ConfigMapInterface {
	return MockErrorConfigMapClient{}
}

type MockErrorK8sClient struct {
	kubernetes.Interface
}

func (m MockErrorK8sClient) CoreV1() corev1.CoreV1Interface {
	return MockErrorCoreV1{}
}
