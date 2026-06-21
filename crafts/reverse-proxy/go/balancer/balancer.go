package balancer

import (
	"errors"
	"time"
)

type Balancer interface {
	Next() (string, error)
	StartHealthCheck()
}

type Kind int

const (
	KindRoundRobin Kind = iota
	KindLeastConn  Kind = iota
)

func NewBalancer(kind Kind, strUpstreams []string, interval time.Duration) (Balancer, error) {
	if len(strUpstreams) == 0 {
		return nil, errors.New("upstreams must not be empty")
	}

	upstreams := make([]*Upstream, len(strUpstreams))
	for i, addr := range strUpstreams {
		upstreams[i] = NewUpstream(addr)
	}

	switch kind {
	case KindRoundRobin:
		return &RoundRobin{
			upstreams: upstreams,
			interval:  interval,
			counter:   0,
		}, nil
	case KindLeastConn:
		return &LeastConn{
			upstreams: upstreams,
			interval:  interval,
		}, nil
	default:
		return nil, errors.New("not kind")
	}
}
