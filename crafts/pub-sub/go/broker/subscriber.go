package broker

import (
	"net"
	"sync"
)

type Subscriber struct {
	conn          net.Conn
	mu            sync.Mutex
	subscriptions map[string]*Subscription
}

func NewSubscriber(conn net.Conn) *Subscriber {
	return &Subscriber{
		conn:          conn,
		subscriptions: make(map[string]*Subscription),
	}
}

func (s *Subscriber) Conn() net.Conn {
	return s.conn
}
