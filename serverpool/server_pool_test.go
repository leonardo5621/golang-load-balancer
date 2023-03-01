package serverpool

import (
	"net/http/httputil"
	"net/url"
	"sync"
	"testing"

	"github.com/leonardo5621/golang-load-balancer/backend"
	"github.com/leonardo5621/golang-load-balancer/utils"
	"github.com/stretchr/testify/assert"
)

func TestPoolCreation(t *testing.T) {
	sp, _ := NewServerPool(utils.RoundRobin)
	url, _ := url.Parse("http://localhost:3333")
	b := backend.NewBackend(url, httputil.NewSingleHostReverseProxy(url))
	sp.AddBackend(b)

	assert.Equal(t, 1, sp.GetServerPoolSize())
}

func TestNextIndexIteration(t *testing.T) {
	sp, _ := NewServerPool(utils.RoundRobin)
	url, _ := url.Parse("http://localhost:3333")
	b := backend.NewBackend(url, httputil.NewSingleHostReverseProxy(url))
	sp.AddBackend(b)

	url, _ = url.Parse("http://localhost:3334")
	b2 := backend.NewBackend(url, httputil.NewSingleHostReverseProxy(url))
	sp.AddBackend(b2)

	url, _ = url.Parse("http://localhost:3335")
	b3 := backend.NewBackend(url, httputil.NewSingleHostReverseProxy(url))
	sp.AddBackend(b3)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		for i := 0; i < 3; i++ {
			sp.GetNextValidPeer()
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < 2; i++ {
			sp.GetNextValidPeer()
		}
	}()

	wg.Wait()
	assert.Equal(t, b.GetURL().String(), sp.GetNextValidPeer().GetURL().String())
}
