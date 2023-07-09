package cache

import (
	"log"
	"sync"
	"time"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic/dynamicinformer"
	k8scache "k8s.io/client-go/tools/cache"

	v1 "github.com/highsaltlevels/saltbot/apis/v1"
)

type PollCache struct {
	// Implement the cache interface
	Cache

	informer *k8scache.SharedIndexInformer
	polls    map[string]v1.Poll
	lock     sync.Mutex
	stopCh   <-chan struct{}
}

func newPollCache() *PollCache {
	var informer k8scache.SharedIndexInformer
	pc := PollCache{
		informer: &informer,
		polls:    map[string]v1.Poll{},
		lock:     sync.Mutex{},
		stopCh:   make(chan struct{}),
	}
	poll := schema.GroupVersionResource{
		Group:    apiGroup,
		Version:  apiVersion,
		Resource: "Poll",
	}
	factory := dynamicinformer.NewFilteredDynamicSharedInformerFactory(client, time.Minute, namespace, nil)
	informer = factory.ForResource(poll).Informer()
	informer.AddEventHandler(
		k8scache.ResourceEventHandlerFuncs{
			AddFunc:    pc.addHandler,
			UpdateFunc: pc.updateHandler,
			DeleteFunc: pc.deleteHandler,
		},
	)

	log.Println("starting poll informer and waiting for it to sync")
	go informer.Run(pc.stopCh)
	k8scache.WaitForCacheSync(pc.stopCh, informer.HasSynced)
	log.Println("poll informer cache is synced")

	return &pc
}

func (pc *PollCache) addHandler(obj interface{}) {
	_, ok := obj.(v1.Poll)
	if !ok {
		log.Printf("failed to parse Poll CR. Not adding to cache.")
	}

	log.Printf("added poll to cache")
}

func (pc *PollCache) updateHandler(oldObj interface{}, newObj interface{}) {}

func (pc *PollCache) deleteHandler(obj interface{}) {}

func (pc *PollCache) List() map[string]v1.Poll {
	pc.lock.Lock()
	defer pc.lock.Unlock()

	return pc.polls
}

func (pc *PollCache) Get(id string) *v1.Poll { return nil }

func (pc *PollCache) Add(p *v1.Poll) error { return nil }

func (pc *PollCache) Update(p *v1.Poll) error { return nil }

func (pc *PollCache) Delete(id string) error { return nil }
