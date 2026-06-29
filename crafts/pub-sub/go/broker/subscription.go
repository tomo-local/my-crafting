package broker

type Subscription struct {
	ch    chan string
	topic string
}

func (s *Subscription) Messages() chan string {
	return s.ch
}

func (s *Subscription) Topic() string {
	return s.topic
}

func (s *Subscription) SetTopic(topic string) {
	s.topic = topic
}
