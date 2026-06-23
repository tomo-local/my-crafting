package balancer

import (
	"errors"
	"sync/atomic"
	"time"
)

type RoundRobin struct {
	upstreams []*Upstream
	interval  time.Duration
	counter   uint64
	maxIdle   int
}

func (r *RoundRobin) Next() (*Upstream, error) {
	for i := 0; i < len(r.upstreams); i++ {
		n := atomic.AddUint64(&r.counter, 1)
		u := r.upstreams[(n-1)%uint64(len(r.upstreams))]
		if u.IsAlive() {
			return u, nil
		}
	}
	return nil, errors.New("no alive upstream")
}

func (r *RoundRobin) StartHealthCheck() {
	for _, u := range r.upstreams {
		go func(u *Upstream) {
			for {
				u.CheckHealth()
				time.Sleep(r.interval)
			}
		}(u)
	}
}
