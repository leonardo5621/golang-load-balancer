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
	"github.com/leonardo5621/golang-load-balancer/frontend"
	"github.com/leonardo5621/golang-load-balancer/serverpool"
)

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
	serverPool := serverpool.NewServerPool()
	loadBalancer := frontend.NewLoadBalancer(serverPool)
	for _, u := range backendEndpoints {
		endpoint, err := url.Parse(u)
		if err != nil {
			log.Fatal(err)
		}

		rp := httputil.NewSingleHostReverseProxy(endpoint)
		backendServer := backend.NewBackend(endpoint, rp)
		rp.ErrorHandler = func(writer http.ResponseWriter, request *http.Request, e error) {
			log.Printf("[%s] %s\n", endpoint.Host, e.Error())
			retries := frontend.GetRetryFromContext(request)
			if retries < 3 {
				select {
				case <-time.After(10 * time.Millisecond):
					ctx := context.WithValue(request.Context(), frontend.Retry, retries+1)
					rp.ServeHTTP(writer, request.WithContext(ctx))
				}
				return
			}
			backendServer.SetAlive(false)

			attempts := frontend.GetAttemptsFromContext(request)
			log.Printf("%s(%s) Attempting retry %d\n", request.RemoteAddr, request.URL.Path, attempts)
			ctx := context.WithValue(request.Context(), frontend.Attempts, attempts+1)
			loadBalancer.Serve(writer, request.WithContext(ctx))
		}

		serverPool.AddBackend(backendServer)
	}

	server := http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: http.HandlerFunc(loadBalancer.Serve),
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
