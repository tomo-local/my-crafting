package server

import (
	"errors"
	b "reverse-proxy/balancer"
)

type Server interface {
	ListenAndServe() error
}

func NewHTTPServer(addr string, handler HttpHandler) Server {
	return &HttpServer{
		baseServer: baseServer{Addr: addr},
		Handler:    handler,
	}
}

func NewReverseProxyServer(addr string, handler ReverseProxyHandler, balancer b.Balancer) Server {
	return &ReverseProxyServer{
		Handler:    handler,
		baseServer: baseServer{Addr: addr},
		balancer:   balancer,
	}
}

var ErrServerClosed = errors.New("server: Server closed")
