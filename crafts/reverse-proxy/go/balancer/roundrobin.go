package balancer

import (
	"errors"
	"sync/atomic"
	"time"
)

type RoundRobin struct {
	upstreams []*Upstream
	counter   uint64
}

func NewRoundRobin(strUpstreams []string) (*RoundRobin, error) {
	if len(strUpstreams) == 0 {
		return nil, errors.New("upstreams must not be empty")
	}

	upstreams := make([]*Upstream, len(strUpstreams))
	for i, addr := range strUpstreams {
		upstreams[i] = NewUpstream(addr)
	}

	return &RoundRobin{
		upstreams: upstreams,
		counter:   0,
	}, nil
}

func (r *RoundRobin) Next() (string, error) {
	for i := 0; i < len(r.upstreams); i++ {
		n := atomic.AddUint64(&r.counter, 1)
		u := r.upstreams[(n-1)%uint64(len(r.upstreams))]
		if u.IsAlive() {
			return u.Addr, nil
		}
	}
	return "", errors.New("no alive upstream")
}

func (r *RoundRobin) StartHealthCheck(interval time.Duration) {
	for _, u := range r.upstreams {
		go func(u *Upstream) {
			for {
				u.CheckHealth()
				time.Sleep(interval)
			}
		}(u)
	}
}
