package cache

import (
	"log"
	"path/filepath"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

const (
	namespace  = "saltbot"
	apiGroup   = "saltbot.greeson.xyz"
	apiVersion = "v1"
)

var client *dynamic.DynamicClient
var Polls *PollCache

type Cache interface {
	// Informer handlers
	addHandler(obj interface{})
	updateHandler(oldObj interface{}, newObj interface{})
	deleteHandler(obj interface{})

	List() map[string]interface{}
	Get(id string) interface{}

	Add(interface{}) error
	Update(interface{}) error
	Delete(id string) error
}

func init() {
	config, err := getClientConfig()
	if err != nil {
		log.Fatalf("failed to create k8s client config: %v", err)
	}

	client, err = dynamic.NewForConfig(config)
	if err != nil {
		log.Fatalf("failed to create k8s client: %v", err)
	}

	if Polls == nil {
		Polls = newPollCache()
	}
}

func getClientConfig() (config *rest.Config, err error) {
	// If we can't get the in-cluster config, try the ~/.kube/config
	if config, err = rest.InClusterConfig(); err == nil {
		return config, nil
	}

	log.Printf("failed to initialize k8s client config: %v", err)
	log.Println("attempting to use ~/.kube/config")

	kubeconfig := filepath.Join(homedir.HomeDir(), ".kube", "config")
	config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	return config, err
}
