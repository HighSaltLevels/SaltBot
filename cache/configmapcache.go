package cache

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"sync"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	infcorev1 "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	k8scache "k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

const namespace = "saltbot"

type ConfigMapCache struct {
	informer  k8scache.SharedIndexInformer
	polls     map[string]Poll
	reminders map[string]Reminder
	lock      sync.Mutex
	stopCh    <-chan struct{}
}

// Only create a single instance of config map cache
var Cache *ConfigMapCache
var client *kubernetes.Clientset

func init() {
	config, err := getClientConfig()
	if err != nil {
		log.Fatalf("failed to create k8s client config: %v", err)
	}

	client, err = kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("failed to create k8s client: %v", err)
	}

	if Cache == nil {
		Cache = newConfigMapCache()
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

func newConfigMapCache() *ConfigMapCache {
	var informer k8scache.SharedIndexInformer
	c := ConfigMapCache{
		informer:  informer,
		polls:     make(map[string]Poll, 1),
		reminders: make(map[string]Reminder, 1),
		stopCh:    make(chan struct{}),
	}

	informer = infcorev1.NewConfigMapInformer(client, namespace, time.Hour*24, nil)
	_, err := informer.AddEventHandler(
		k8scache.ResourceEventHandlerFuncs{
			AddFunc:    c.addConfigMap,
			UpdateFunc: c.updateConfigMap,
			DeleteFunc: c.deleteConfigMap,
		},
	)
	if err != nil {
		log.Fatalf("failed to create informer handler: %v", err)
	}

	log.Println("starting informer and waiting for it to sync")
	go informer.Run(c.stopCh)
	k8scache.WaitForCacheSync(c.stopCh, informer.HasSynced)
	log.Println("informer cache has synced")

	return &c
}

// Add handler for the configmap informer
func (c *ConfigMapCache) addConfigMap(obj interface{}) {
	configMap := obj.(*corev1.ConfigMap)
	name := configMap.ObjectMeta.Name

	c.lock.Lock()
	defer c.lock.Unlock()

	if strings.Contains(name, "poll") {
		p := Poll{}
		err := p.FromConfigMap(configMap)
		if err != nil {
			log.Printf("failed to parse poll: %v", err)
		} else {
			log.Printf("adding poll with id: %s", p.Id)
			c.polls[p.Id] = p
		}
	}

	if strings.Contains(name, "reminder") {
		r := Reminder{}
		err := r.FromConfigMap(configMap)
		if err != nil {
			log.Printf("failed to parse reminder: %v", err)
		} else {
			log.Printf("adding reminder with id: %s", r.Id)
			c.reminders[r.Id] = r
		}
	}
}

// Update handler for the configmap informer
func (c *ConfigMapCache) updateConfigMap(oldObj interface{}, newObj interface{}) {
	// We don't care about the older resource version
	configMap := newObj.(*corev1.ConfigMap)
	name := configMap.ObjectMeta.Name

	c.lock.Lock()
	defer c.lock.Unlock()

	if strings.Contains(name, "poll") {
		p := Poll{}
		err := p.FromConfigMap(configMap)
		if err != nil {
			log.Printf("failed to parse poll: %v", err)
		} else {
			log.Printf("updating poll with id: %s", p.Id)
			c.polls[p.Id] = p
		}
	}

	if strings.Contains(name, "reminder") {
		r := Reminder{}
		err := r.FromConfigMap(configMap)
		if err != nil {
			log.Printf("failed to parse poll: %v", err)
		} else {
			log.Printf("updating reminder with id: %s", r.Id)
			c.reminders[r.Id] = r
		}
	}
}

// Delete handler for the configmap informer
func (c *ConfigMapCache) deleteConfigMap(obj interface{}) {
	configMap := obj.(*corev1.ConfigMap)
	nameParts := strings.Split(configMap.ObjectMeta.Name, "-")
	if len(nameParts) < 2 {
		fmt.Printf("unparseable configmap name %s. Ignoring deletion\n", configMap.ObjectMeta.Name)
		return
	}

	c.lock.Lock()
	defer c.lock.Unlock()

	if nameParts[0] == "poll" {
		delete(c.polls, nameParts[1])
	}

	if nameParts[0] == "reminder" {
		delete(c.reminders, nameParts[1])
	}
}

// Getter for polls in the cache
func (c *ConfigMapCache) ListPolls() map[string]Poll {
	c.lock.Lock()
	defer c.lock.Unlock()
	return c.polls
}

func (c *ConfigMapCache) GetPoll(id, author string) *Poll {
	c.lock.Lock()
	defer c.lock.Unlock()
	var poll Poll
	var ok bool
	if poll, ok = c.polls[id]; !ok {
		return nil
	}

	return &poll
}

// Add a poll configmap, this in turn triggers the informer handler which
// adds it to the in-mem cache.
func (c *ConfigMapCache) AddPoll(p *Poll) error {
	cmClient := client.CoreV1().ConfigMaps(namespace)

	configMap, err := p.ToConfigMap()
	if err != nil {
		return err
	}

	_, err = cmClient.Create(context.TODO(), configMap, metav1.CreateOptions{})
	return err
}

func (c *ConfigMapCache) UpdatePoll(p *Poll) error {
	cmClient := client.CoreV1().ConfigMaps(namespace)

	configMap, err := p.ToConfigMap()
	if err != nil {
		return err
	}

	_, err = cmClient.Update(context.TODO(), configMap, metav1.UpdateOptions{})
	return err
}

// Getter for reminders in the cache
func (c *ConfigMapCache) ListReminders() map[string]Reminder {
	c.lock.Lock()
	defer c.lock.Unlock()
	return c.reminders
}

func (c *ConfigMapCache) GetReminder(id, author string) *Reminder {
	c.lock.Lock()
	defer c.lock.Unlock()
	var reminder Reminder
	var ok bool
	if reminder, ok = c.reminders[id]; !ok {
		return nil
	}

	if reminder.Author != author {
		return nil
	}

	return &reminder
}

// Add a reminder configmap, this in turn triggers the informer handler which
// adds it to the in-mem cache.
func (c *ConfigMapCache) AddReminder(r *Reminder, user string) error {
	cmClient := client.CoreV1().ConfigMaps(namespace)
	configMap, err := r.ToConfigMap()
	if err != nil {
		return fmt.Errorf("failed to convert reminder to configMap: %v", err)
	}

	_, err = cmClient.Create(context.TODO(), configMap, metav1.CreateOptions{})
	return err
}

/*
Delete the configmap from the cluster which in turn triggers

	the delete handler to remove it from the in-mem cache
*/
func (c *ConfigMapCache) Delete(name string) {
	cmClient := client.CoreV1().ConfigMaps(namespace)
	err := cmClient.Delete(context.TODO(), name, metav1.DeleteOptions{})
	if err != nil {
		log.Printf("warning: failed to delete configmap: %v\n", err)
		log.Println("deleting from in-mem cache anyway")
	}
}
