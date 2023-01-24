package serverpool

import (
	"context"
	"time"

	"github.com/leonardo5621/golang-load-balancer/utils"
)

func LauchHealthCheck(ctx context.Context, sp ServerPool) {
	t := time.NewTicker(time.Second * 20)
	for {
		select {
		case <-t.C:
			utils.Logger.Info("Starting health check...")
			go HealthCheck(ctx, sp)
		case <-ctx.Done():
			utils.Logger.Info("Closing Health Check")
			return
		}
	}
}
