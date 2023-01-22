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

	"go.uber.org/zap"

	"github.com/leonardo5621/golang-load-balancer/backend"
	"github.com/leonardo5621/golang-load-balancer/frontend"
	"github.com/leonardo5621/golang-load-balancer/serverpool"
	"github.com/leonardo5621/golang-load-balancer/utils"
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

	logger := utils.InitLogger()
	defer logger.Sync()

	backendEndpoints := strings.Split(backendList, ",")
	if len(backendEndpoints) == 0 {
		utils.Logger.Fatal("empty server pool")
	}

	serverPool := serverpool.NewServerPool()
	loadBalancer := frontend.NewLoadBalancer(serverPool)

	for _, u := range backendEndpoints {
		endpoint, err := url.Parse(u)
		if err != nil {
			logger.Fatal(err.Error(), zap.String("URL", u))
		}

		rp := httputil.NewSingleHostReverseProxy(endpoint)
		backendServer := backend.NewBackend(endpoint, rp)
		rp.ErrorHandler = func(writer http.ResponseWriter, request *http.Request, e error) {
			logger.Error("error handling the request",
				zap.String("host", endpoint.Host),
				zap.Error(e),
			)
			retries := frontend.GetRetryFromContext(request)

			if retries < 3 {
				<-time.After(50 * time.Duration(retries+1) * time.Millisecond)
				ctx := context.WithValue(request.Context(), frontend.Retry, retries+1)
				rp.ServeHTTP(writer, request.WithContext(ctx))
				return
			}

			backendServer.SetAlive(false)
			attempts := frontend.GetAttemptsFromContext(request)
			logger.Info(
				"Attempting retry",
				zap.String("address", request.RemoteAddr),
				zap.String("URL", request.URL.Path),
				zap.Int("attempts", attempts),
			)
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

	logger.Info(
		"Load Balancer started",
		zap.Int("port", port),
	)
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		logger.Fatal("ListenAndServe() error", zap.Error(err))
	}
}
