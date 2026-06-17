package balancer

import (
	"net"
	"sync"
	"time"
)

type Upstream struct {
	Addr  string
	alive bool
	mu    sync.RWMutex
}

func NewUpstream(addr string) *Upstream {
	return &Upstream{
		Addr:  addr,
		alive: true,
	}
}

func (u *Upstream) SetAlive(alive bool) {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.alive = alive
}

func (u *Upstream) IsAlive() bool {
	u.mu.RLock()
	defer u.mu.RUnlock()
	return u.alive
}

func (u *Upstream) CheckHealth() {
	conn, err := net.DialTimeout("tcp", u.Addr,
		2*time.Second)
	if err != nil {
		u.SetAlive(false)
		return
	}
	conn.Close()
	u.SetAlive(true)
}
