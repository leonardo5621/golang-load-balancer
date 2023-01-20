package serverpool

import (
	"context"
	"log"
	"time"
)

func LauchHealthCheck(ctx context.Context, sp ServerPool) {
	t := time.NewTicker(time.Second * 20)
	for {
		select {
		case <-t.C:
			log.Println("Starting health check...")
			go sp.HealthCheck(ctx)
		case <-ctx.Done():
			log.Println("Closing Health Check")
			return
		}
	}
}
