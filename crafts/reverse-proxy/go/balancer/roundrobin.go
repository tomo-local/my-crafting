package balancer

import "sync/atomic"

type RoundRobin struct {
	upstreams []string
	counter   uint64
}

func NewRoundRobin(upstreams []string) *RoundRobin {
	return &RoundRobin{
		upstreams: upstreams,
		counter:   0,
	}
}

func (r *RoundRobin) Next() string {
	n := atomic.AddUint64(&r.counter, 1)
	return r.upstreams[(n-1)%uint64(len(r.upstreams))]
}
