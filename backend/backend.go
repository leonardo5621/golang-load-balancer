package backend

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
)

type Backend interface {
	SetAlive(bool)
	IsAlive() bool
	GetURL() *url.URL
	GetActiveConnections() int
	ServeThoughReverseProxy(http.ResponseWriter, *http.Request)
}

type backend struct {
	URL          *url.URL
	Alive        bool
	mux          sync.RWMutex
	connections  int
	ReverseProxy *httputil.ReverseProxy
}

func (b *backend) GetActiveConnections() int {
	b.mux.RLock()
	connections := b.connections
	b.mux.RUnlock()
	return connections
}

func (b *backend) SetAlive(alive bool) {
	b.mux.Lock()
	b.Alive = alive
	b.mux.Unlock()
}

func (b *backend) IsAlive() bool {
	b.mux.RLock()
	alive := b.Alive
	defer b.mux.RUnlock()
	return alive
}

func (b *backend) GetURL() *url.URL {
	return b.URL
}

func (b *backend) ServeThoughReverseProxy(rw http.ResponseWriter, req *http.Request) {
	defer func() {
		b.mux.Lock()
		b.connections--
		b.mux.Unlock()
	}()

	b.mux.Lock()
	b.connections++
	b.mux.Unlock()
	b.ReverseProxy.ServeHTTP(rw, req)
}

func NewBackend(u *url.URL, rp *httputil.ReverseProxy) Backend {
	return &backend{
		URL:          u,
		Alive:        true,
		ReverseProxy: rp,
	}
}
