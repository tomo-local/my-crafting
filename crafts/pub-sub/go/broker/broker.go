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

func (b *Broker) Subscribe(topic string, sub *Subscriber) chan string {
	b.mu.Lock()
	defer b.mu.Unlock()

	sub.mu.Lock()
	if existing, ok := sub.subscriptions[topic]; ok {
		sub.mu.Unlock()
		return existing.ch
	}
	sub.mu.Unlock()

	subscription := &Subscription{ch: make(chan string, 64), topic: topic}
	sub.subscriptions[topic] = subscription
	b.subscribers[topic] = append(b.subscribers[topic], sub)
	return subscription.ch
}

func (b *Broker) Unsubscribe(topic string, sub *Subscriber) {
	b.mu.Lock()
	defer b.mu.Unlock()

	sub.mu.Lock()
	delete(sub.subscriptions, topic)
	sub.mu.Unlock()

	subs := b.subscribers[topic]
	for i, s := range subs {
		if s == sub {
			b.subscribers[topic] = append(subs[:i], subs[i+1:]...)
			break
		}
	}
}

func (b *Broker) UnsubscribeAll(sub *Subscriber) {
	sub.mu.Lock()
	topics := make([]string, 0, len(sub.subscriptions))
	for topic := range sub.subscriptions {
		topics = append(topics, topic)
	}
	sub.mu.Unlock()

	for _, topic := range topics {
		b.Unsubscribe(topic, sub)
	}
}

func (b *Broker) Publish(topic string, message string) {
	b.mu.RLock()
	subs := make([]*Subscriber, len(b.subscribers[topic]))
	copy(subs, b.subscribers[topic])
	defer b.mu.RUnlock()

	for _, sub := range subs {
		sub.mu.Lock()
		subscription, ok := sub.subscriptions[topic]
		sub.mu.Unlock()
		if ok {
			subscription.ch <- message
		}
	}
}
