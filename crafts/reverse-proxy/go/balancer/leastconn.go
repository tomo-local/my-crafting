package balancer

import (
	"errors"
	"math"
	"time"
)

type LeastConn struct {
	upstreams []*Upstream
	interval  time.Duration
	maxIdle   int
}

func (lc *LeastConn) Next() (*Upstream, error) {
	var candidate *Upstream
	minConn := int64(math.MaxInt64)

	for _, u := range lc.upstreams {
		if !u.IsAlive() {
			continue
		}

		c := u.Connections()
		if c < minConn {
			minConn = c
			candidate = u
		}
	}

	if candidate == nil {
		return nil, errors.New("no alive upstream")
	}

	return candidate, nil
}

func (lc *LeastConn) StartHealthCheck() {
	for _, u := range lc.upstreams {
		go func(u *Upstream) {
			for {
				u.CheckHealth()
				time.Sleep(lc.interval)
			}
		}(u)
	}
}
