package serverpool

import (
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"sync"
	"testing"
	"time"

	"github.com/leonardo5621/golang-load-balancer/backend"
	"github.com/leonardo5621/golang-load-balancer/utils"
	"github.com/stretchr/testify/assert"
)

func SleepHandler(w http.ResponseWriter, r *http.Request) {
	time.Sleep(5 * time.Second)
}

var (
	h   = http.HandlerFunc(SleepHandler)
	req = httptest.NewRequest(http.MethodGet, "/test", nil)
	w   = httptest.NewRecorder()
)

func TestLeastConnectionLB(t *testing.T) {
	dummyServer1 := httptest.NewServer(h)
	defer dummyServer1.Close()
	backend1URL, err := url.Parse(dummyServer1.URL)
	if err != nil {
		t.Fatal(err)
	}

	dummyServer2 := httptest.NewServer(h)
	defer dummyServer2.Close()
	backend2URL, err := url.Parse(dummyServer2.URL)
	if err != nil {
		t.Fatal(err)
	}

	rp1 := httputil.NewSingleHostReverseProxy(backend1URL)
	backend1 := backend.NewBackend(backend1URL, rp1)

	rp2 := httputil.NewSingleHostReverseProxy(backend2URL)
	backend2 := backend.NewBackend(backend2URL, rp2)

	serverPool, err := NewServerPool(utils.LeastConnected)
	if err != nil {
		t.Fatal(err)
	}

	serverPool.AddBackend(backend1)
	serverPool.AddBackend(backend2)

	assert.Equal(t, 2, serverPool.GetServerPoolSize())

	var wg sync.WaitGroup
	wg.Add(1)

	peer := serverPool.GetNextValidPeer()
	t.Log(peer.GetURL().String())
	assert.NotNil(t, peer)
	go func() {
		defer wg.Done()
		peer.Serve(w, req)
	}()
	time.Sleep(1 * time.Second)
	peer2 := serverPool.GetNextValidPeer()
	t.Log(peer2.GetURL().String())
	connPeer2 := peer2.GetActiveConnections()

	assert.NotNil(t, peer)
	assert.Equal(t, 0, connPeer2)
	assert.NotEqual(t, peer, peer2)

	wg.Wait()
}
