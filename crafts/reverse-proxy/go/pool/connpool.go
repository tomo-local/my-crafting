package pool

import "net"

type ConnPool struct {
	addr    string
	idle    chan net.Conn
	maxIdle int
}

func NewConnPool(addr string, maxIdle int) *ConnPool {
	return &ConnPool{
		addr:    addr,
		idle:    make(chan net.Conn, maxIdle),
		maxIdle: maxIdle,
	}
}

func (p *ConnPool) Get() (net.Conn, error) {
	select {
	case conn := <-p.idle:
		return conn, nil
	default:
		return net.Dial("tcp", p.addr)
	}
}

func (p *ConnPool) Put(conn net.Conn) {
	select {
	case p.idle <- conn:
	default:
		conn.Close()
	}
}
