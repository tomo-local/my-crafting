package broker

import "net"

type Subscriber struct {
	conn  net.Conn
	ch    chan string
	topic string
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
