package balancer

type Balancer interface {
	Next() (string, error)
	StartHealthCheck()
}
