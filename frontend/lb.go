package frontend

import (
	"log"
	"net/http"

	"github.com/leonardo5621/golang-load-balancer/serverpool"
)

const (
	Attempts int = iota
	Retry
)

func GetAttemptsFromContext(r *http.Request) int {
	if attempts, ok := r.Context().Value(Attempts).(int); ok {
		return attempts
	}
	return 1
}

func GetRetryFromContext(r *http.Request) int {
	if retry, ok := r.Context().Value(Retry).(int); ok {
		return retry
	}
	return 0
}

type LoadBalancer interface {
	Serve(http.ResponseWriter, *http.Request)
}

type loadBalancer struct {
	serverPool serverpool.ServerPool
}

func (lb *loadBalancer) Serve(w http.ResponseWriter, r *http.Request) {
	attempts := GetAttemptsFromContext(r)
	if attempts > 3 {
		log.Printf("%s(%s) Max attempts reached, terminating\n", r.RemoteAddr, r.URL.Path)
		http.Error(w, "Service not available", http.StatusServiceUnavailable)
		return
	}

	peer := lb.serverPool.GetNextPeer()
	if peer != nil {
		peer.ServeThoughReverseProxy(w, r)
		return
	}
	http.Error(w, "Service not available", http.StatusServiceUnavailable)
}

func NewLoadBalancer(serverPool serverpool.ServerPool) LoadBalancer {
	return &loadBalancer{
		serverPool: serverPool,
	}
}
