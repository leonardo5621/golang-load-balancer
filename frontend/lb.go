package frontend

import (
	"net/http"

	"github.com/leonardo5621/golang-load-balancer/serverpool"
)

const (
	RETRY_ATTEMPTED int = 0
)

func AllowRetry(r *http.Request) bool {
	if _, ok := r.Context().Value(RETRY_ATTEMPTED).(bool); ok {
		return false
	}
	return true
}

type LoadBalancer interface {
	Serve(http.ResponseWriter, *http.Request)
}

type loadBalancer struct {
	serverPool serverpool.ServerPool
}

func (lb *loadBalancer) Serve(w http.ResponseWriter, r *http.Request) {
	peer := lb.serverPool.GetNextValidPeer()
	if peer != nil {
		peer.Serve(w, r)
		return
	}
	http.Error(w, "Service not available", http.StatusServiceUnavailable)
}

func NewLoadBalancer(serverPool serverpool.ServerPool) LoadBalancer {
	return &loadBalancer{
		serverPool: serverPool,
	}
}
