package balancer

import (
	"net"
	"reverse-proxy/pool"
	"sync"
	"sync/atomic"
	"time"
)

type Upstream struct {
	Addr        string
	alive       bool
	mu          sync.RWMutex
	connections int64
	Pool        *pool.ConnPool
}

func NewUpstream(addr string, maxIdle int) *Upstream {
	return &Upstream{
		Addr:  addr,
		alive: true,
		Pool:  pool.NewConnPool(addr, 10),
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

func (u *Upstream) Connections() int64 {
	return atomic.LoadInt64(&u.connections)
}

func (u *Upstream) IncrConnections() {
	atomic.AddInt64(&u.connections, 1)
}

func (u *Upstream) DecrConnections() {
	atomic.AddInt64(&u.connections, -1)
}
