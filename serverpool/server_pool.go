package serverpool

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/leonardo5621/golang-load-balancer/backend"
	"github.com/leonardo5621/golang-load-balancer/utils"
	"go.uber.org/zap"
)

type ServerPool interface {
	GetBackends() []backend.Backend
	GetNextPeer() backend.Backend
	AddBackend(backend.Backend)
	GetServerPoolSize() int
}

type roundRobinServerPool struct {
	backends []backend.Backend
	current  uint64
}

func (s *roundRobinServerPool) NextIndex() int {
	return int(atomic.AddUint64(&s.current, uint64(1)) % uint64(s.GetServerPoolSize()))
}

func (s *roundRobinServerPool) GetNextPeer() backend.Backend {
	next := s.NextIndex()
	backends := s.GetBackends()
	spLen := s.GetServerPoolSize()
	l := spLen + next

	for i := next; i < l; i++ {
		idx := i % spLen
		if backends[idx].IsAlive() {
			if i != next {
				atomic.StoreUint64(&s.current, uint64(idx))
			}
			return backends[idx]
		}
	}
	return nil
}

func (s *roundRobinServerPool) GetBackends() []backend.Backend {
	return s.backends
}

func (s *roundRobinServerPool) AddBackend(b backend.Backend) {
	s.backends = append(s.backends, b)
}

func (s *roundRobinServerPool) GetServerPoolSize() int {
	return len(s.backends)
}

func HealthCheck(ctx context.Context, s ServerPool) {
	aliveChannel := make(chan bool, 1)

	for _, b := range s.GetBackends() {
		b := b
		requestCtx, stop := context.WithTimeout(ctx, 10*time.Second)
		defer stop()
		status := "up"
		go backend.IsBackendAlive(requestCtx, aliveChannel, b.GetURL())

		select {
		case <-ctx.Done():
			utils.Logger.Info("Gracefully shutting down health check")
			return
		case alive := <-aliveChannel:
			b.SetAlive(alive)
			if !alive {
				status = "down"
			}
		}
		utils.Logger.Debug(
			"URL Status",
			zap.String("URL", b.GetURL().String()),
			zap.String("status", status),
		)
	}
}

func NewServerPool() ServerPool {
	return &roundRobinServerPool{
		backends: make([]backend.Backend, 0),
		current:  uint64(0),
	}
}
