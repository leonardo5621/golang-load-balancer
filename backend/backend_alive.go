package backend

import (
	"context"
	"net"
	"net/url"

	"github.com/leonardo5621/golang-load-balancer/utils"
	"go.uber.org/zap"
)

func IsBackendAlive(ctx context.Context, aliveChannel chan bool, u *url.URL) {
	var d net.Dialer
	conn, err := d.DialContext(ctx, "tcp", u.Host)
	if err != nil {
		utils.Logger.Debug("Site unreachable", zap.Error(err))
		aliveChannel <- false
		return
	}
	_ = conn.Close()
	aliveChannel <- true
}
