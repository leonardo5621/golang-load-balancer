package serverpool

import (
	"context"
	"log"
	"sync/atomic"
	"time"

	"github.com/leonardo5621/golang-load-balancer/backend"
)

type ServerPool interface {
	HealthCheck(ctx context.Context)
	NextIndex() int
	GetNextPeer() backend.Backend
	AddBackend(backend.Backend)
}

type serverPool struct {
	backends []backend.Backend
	current  uint64
}

func (s *serverPool) NextIndex() int {
	return int(atomic.AddUint64(&s.current, uint64(1)) % uint64(len(s.backends)))
}

func (s *serverPool) GetNextPeer() backend.Backend {
	// loop entire backends to find out an Alive backend
	next := s.NextIndex()
	l := len(s.backends) + next // start from next and move a full cycle
	for i := next; i < l; i++ {
		idx := i % len(s.backends) // take an index by modding with length
		// if we have an alive backend, use it and store if its not the original one
		if s.backends[idx].IsAlive() {
			if i != next {
				atomic.StoreUint64(&s.current, uint64(idx)) // mark the current one
			}
			return s.backends[idx]
		}
	}
	return nil
}

func (s *serverPool) HealthCheck(ctx context.Context) {
	aliveChannel := make(chan bool, 1)

	for _, b := range s.backends {
		b := b
		requestCtx, stop := context.WithTimeout(ctx, 10*time.Second)
		defer stop()
		status := "up"
		go backend.IsBackendAlive(requestCtx, aliveChannel, b.GetURL())

		select {
		case <-ctx.Done():
			log.Println("Gracefully shutting down health check")
			return
		case alive := <-aliveChannel:
			b.SetAlive(alive)
			if !alive {
				status = "down"
			}
		}
		log.Printf("%s [%s]\n", b.GetURL(), status)
	}
}

func (s *serverPool) AddBackend(b backend.Backend) {
	s.backends = append(s.backends, b)
}

func NewServerPool() ServerPool {
	return &serverPool{
		backends: make([]backend.Backend, 0),
		current:  uint64(0),
	}
}
