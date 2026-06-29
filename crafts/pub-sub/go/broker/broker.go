package broker

import (
	"sync"
)

type Broker struct {
	mu          sync.RWMutex
	subscribers map[string][]*Subscriber
}

func NewBroker() *Broker {
	return &Broker{subscribers: make(map[string][]*Subscriber)}
}

func (b *Broker) Subscribe(topic string, sub *Subscriber) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.subscribers[topic] = append(b.subscribers[topic], sub)
}

func (b *Broker) Publish(topic string, message string) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	subs := b.subscribers[topic]
	for _, sub := range subs {
		sub.ch <- message
	}
}
