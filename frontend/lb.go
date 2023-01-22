package frontend

import (
	"net/http"

	"github.com/leonardo5621/golang-load-balancer/serverpool"
	"github.com/leonardo5621/golang-load-balancer/utils"
	"go.uber.org/zap"
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
	GetAttemptLimit() int
}

type loadBalancer struct {
	serverPool   serverpool.ServerPool
	attemptLimit int
}

func (lb *loadBalancer) GetAttemptLimit() int {
	if lb.attemptLimit == 0 {
		spLen := lb.serverPool.GetServerPoolSize()
		if spLen >= 3 {
			lb.attemptLimit = 3
		} else {
			lb.attemptLimit = spLen
		}
	}
	return lb.attemptLimit
}

func (lb *loadBalancer) Serve(w http.ResponseWriter, r *http.Request) {
	attempts := GetAttemptsFromContext(r)
	if attempts > lb.GetAttemptLimit() {
		utils.Logger.Info(
			"Max attempts reached, terminating",
			zap.String("address", r.RemoteAddr),
			zap.String("path", r.URL.Path),
		)
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
