package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/leonardo5621/golang-load-balancer/backend"
	"github.com/leonardo5621/golang-load-balancer/frontend"
	"github.com/leonardo5621/golang-load-balancer/serverpool"
	"github.com/leonardo5621/golang-load-balancer/utils"
)

func main() {
	logger := utils.InitLogger()
	defer logger.Sync()

	config, err := utils.GetLBConfig()
	if err != nil {
		utils.Logger.Fatal(err.Error())
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	serverPool := serverpool.NewServerPool()
	loadBalancer := frontend.NewLoadBalancer(serverPool)

	for _, u := range config.Backends {
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
		Addr:    fmt.Sprintf(":%d", config.Port),
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
		zap.Int("port", config.Port),
	)
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		logger.Fatal("ListenAndServe() error", zap.Error(err))
	}
}
