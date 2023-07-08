package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"path/filepath"
	"strconv"
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

// TODO Use custom resources instead of configmaps
type Poll struct {
	Author  string              `json:"author"`
	Channel string              `json:"channel"`
	Prompt  string              `json:"prompt"`
	Choices []string            `json:"choices"`
	Expiry  int64               `json:"expiry"`
	Id      string              `json:"unique_id"`
	Votes   map[string][]string `json:"votes"`
}

type Reminder struct {
	Author  string `json:"author"`
	Channel string `json:"channel"`
	Expiry  int64  `json:"expiry"`
	Message string `json:"msg"`
	Id      string `json:"unique_id"`
}

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
	id := configMap.Data["unique_id"]

	c.lock.Lock()
	defer c.lock.Unlock()

	if strings.Contains(name, "poll") {
		poll, err := parsePoll(configMap)
		if err != nil {
			log.Printf("failed to parse poll: %v", err)
		} else {
			log.Printf("adding poll with id: %s", id)
			c.polls[id] = *poll
		}
	}

	if strings.Contains(name, "reminder") {
		reminder, err := parseReminder(configMap)
		if err != nil {
			log.Printf("failed to parse poll: %v", err)
		} else {
			log.Printf("adding reminder with id: %s", id)
			c.reminders[id] = *reminder
		}
	}
}

// Update handler for the configmap informer
func (c *ConfigMapCache) updateConfigMap(oldObj interface{}, newObj interface{}) {
	// We don't care about the older resource version
	configMap := newObj.(*corev1.ConfigMap)
	name := configMap.ObjectMeta.Name
	id := configMap.Data["unique_id"]

	c.lock.Lock()
	defer c.lock.Unlock()

	if strings.Contains(name, "poll") {
		poll, err := parsePoll(configMap)
		if err != nil {
			log.Printf("failed to parse poll: %v", err)
		} else {
			log.Printf("updating poll with id: %s", id)
			c.polls[id] = *poll
		}
	}

	if strings.Contains(name, "reminder") {
		reminder, err := parseReminder(configMap)
		if err != nil {
			log.Printf("failed to parse poll: %v", err)
		} else {
			log.Printf("updating reminder with id: %s", id)
			c.reminders[id] = *reminder
		}
	}
}

// Delete handler for the configmap informer
func (c *ConfigMapCache) deleteConfigMap(obj interface{}) {
	configMap := obj.(*corev1.ConfigMap)
	name := configMap.ObjectMeta.Name
	id := configMap.Data["unique_id"]

	c.lock.Lock()
	defer c.lock.Unlock()

	if strings.Contains(name, "poll") {
		delete(c.polls, id)
	}

	if strings.Contains(name, "reminder") {
		delete(c.reminders, id)
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

	configMap := pollToConfigMap(p)
	_, err := cmClient.Create(context.TODO(), configMap, metav1.CreateOptions{})
	return err
}

func (c *ConfigMapCache) UpdatePoll(p *Poll) error {
	cmClient := client.CoreV1().ConfigMaps(namespace)

	configMap := pollToConfigMap(p)
	_, err := cmClient.Update(context.TODO(), configMap, metav1.UpdateOptions{})
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
	// First convert to a configmap, then use the client to create it
	cmClient := client.CoreV1().ConfigMaps(namespace)
	configMap := corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: "reminder-" + r.Id,
			Labels: map[string]string{
				"user": user,
			},
		},
		Data: map[string]string{
			"author":    r.Author,
			"channel":   r.Channel,
			"expiry":    fmt.Sprintf("%d", r.Expiry),
			"msg":       r.Message,
			"unique_id": r.Id,
		},
	}
	_, err := cmClient.Create(context.TODO(), &configMap, metav1.CreateOptions{})
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

func parsePoll(configMap *corev1.ConfigMap) (*Poll, error) {
	// Parse choices to a list and votes to a map
	var choices []string
	err := json.Unmarshal([]byte(configMap.Data["choices"]), &choices)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal choices: %v", err)
	}
	var votes map[string][]string
	err = json.Unmarshal([]byte(configMap.Data["votes"]), &votes)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal votes: %v", err)
	}

	// Parse the expiry to an int
	expiry, err := strconv.ParseInt(configMap.Data["expiry"], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal expiry: %v", err)
	}

	return &Poll{
		Channel: configMap.Data["channel"],
		Choices: choices,
		Expiry:  expiry,
		Id:      configMap.Data["unique_id"],
		Votes:   votes,
	}, nil
}

func parseReminder(configMap *corev1.ConfigMap) (*Reminder, error) {
	// Parse the expiry to an int
	expiry, err := strconv.ParseInt(configMap.Data["expiry"], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal expiry: %v", err)
	}

	return &Reminder{
		Author:  configMap.Data["author"],
		Channel: configMap.Data["channel"],
		Expiry:  expiry,
		Message: configMap.Data["msg"],
		Id:      configMap.Data["unique_id"],
	}, nil
}

func pollToConfigMap(p *Poll) *corev1.ConfigMap {
	choices := "["
	votes := "{"
	for idx, choice := range p.Choices {
		choices += fmt.Sprintf("\"%s\", ", choice)

		votesForChoice := p.Votes[fmt.Sprintf("%d", idx)]
		votes += fmt.Sprintf("\"%d\": [", idx)
		for _, voteForChoice := range votesForChoice {
			votes += fmt.Sprintf("\"%s\", ", voteForChoice)
		}
		votes = strings.TrimSuffix(votes, ", ") + "], "
	}
	choices = strings.TrimSuffix(choices, ", ") + "]"
	votes = strings.TrimSuffix(votes, ", ") + "}"

	configMap := corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: "poll-" + p.Id,
		},
		Data: map[string]string{
			"author":    p.Author,
			"channel":   p.Channel,
			"expiry":    fmt.Sprintf("%d", p.Expiry),
			"choices":   choices,
			"votes":     votes,
			"unique_id": p.Id,
		},
	}
	return &configMap
}
