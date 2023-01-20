package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/leonardo5621/golang-load-balancer/backend"
	"github.com/leonardo5621/golang-load-balancer/serverpool"
)

const (
	Attempts int = iota
	Retry
)

var (
	serverPool serverpool.ServerPool
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

func lb(w http.ResponseWriter, r *http.Request) {
	attempts := GetAttemptsFromContext(r)
	if attempts > 3 {
		log.Printf("%s(%s) Max attempts reached, terminating\n", r.RemoteAddr, r.URL.Path)
		http.Error(w, "Service not available", http.StatusServiceUnavailable)
		return
	}

	peer := serverPool.GetNextPeer()
	if peer != nil {
		peer.ServeThoughReverseProxy(w, r)
		return
	}
	http.Error(w, "Service not available", http.StatusServiceUnavailable)
}

func main() {
	var (
		port        int
		backendList string
	)
	flag.IntVar(&port, "port", 3333, "Port to serve the LB")
	flag.StringVar(&backendList, "backends", "", "Load balanced backends, use commas to separate")
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	backendEndpoints := strings.Split(backendList, ",")
	serverPool = serverpool.NewServerPool()
	for _, u := range backendEndpoints {
		endpoint, err := url.Parse(u)
		if err != nil {
			log.Fatal(err)
		}

		rp := httputil.NewSingleHostReverseProxy(endpoint)
		rp.ErrorHandler = func(writer http.ResponseWriter, request *http.Request, e error) {
			log.Printf("[%s] %s\n", endpoint.Host, e.Error())
			retries := GetRetryFromContext(request)
			if retries < 3 {
				select {
				case <-time.After(10 * time.Millisecond):
					ctx := context.WithValue(request.Context(), Retry, retries+1)
					rp.ServeHTTP(writer, request.WithContext(ctx))
				}
				return
			}

			// after 3 retries, mark this backend as down
			serverPool.MarkBackendStatus(endpoint, false)

			// if the same request routing for few attempts with different backends, increase the count
			attempts := GetAttemptsFromContext(request)
			log.Printf("%s(%s) Attempting retry %d\n", request.RemoteAddr, request.URL.Path, attempts)
			ctx := context.WithValue(request.Context(), Attempts, attempts+1)
			lb(writer, request.WithContext(ctx))
		}

		serverPool.AddBackend(backend.NewBackend(endpoint, rp))
	}

	server := http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: http.HandlerFunc(lb),
	}

	go serverpool.LauchHealthCheck(ctx, serverPool)

	go func() {
		<-ctx.Done()
		shutdownCtx, _ := context.WithTimeout(context.Background(), 10*time.Second)
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Fatal(err)
		}
	}()

	log.Printf("Load Balancer started at :%d\n", port)
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("ListenAndServe(): %v", err)
	}
}
