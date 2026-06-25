package broker

import (
	"net"
	"sync"
)

type Subscriber struct {
	conn  net.Conn
	ch    chan string
	topic string
}

type Broker struct {
	mu          sync.RWMutex
	subscribers map[string][]*Subscriber
}

func NewBroker() *Broker {
	return &Broker{subscribers: make(map[string][]*Subscriber)}
}

func NewSubscriber(conn net.Conn) *Subscriber {
	return &Subscriber{
		conn: conn,
		ch:   make(chan string, 64),
	}
}

func (s *Subscriber) Conn() net.Conn {
	return s.conn
}

func (s *Subscriber) Messages() chan string {
	return s.ch
}

func (s *Subscriber) Topic() string {
	return s.topic
}

func (s *Subscriber) SetTopic(topic string) {
	s.topic = topic
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
