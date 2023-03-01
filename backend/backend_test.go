package backend

import (
	"net/http/httputil"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBackendCreation(t *testing.T) {
	url, _ := url.Parse("http://localhost:3333")
	b := NewBackend(url, httputil.NewSingleHostReverseProxy(url))
	assert.Equal(t, "http://localhost:3333", b.GetURL().String())
	assert.Equal(t, true, b.IsAlive())
}

func TestBackendAlive(t *testing.T) {
	url, _ := url.Parse("http://localhost:3333")
	b := NewBackend(url, httputil.NewSingleHostReverseProxy(url))
	b.SetAlive(b.IsAlive())
	assert.Equal(t, false, b.IsAlive())
}
