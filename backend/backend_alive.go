package backend

import (
	"context"
	"log"
	"net"
	"net/url"
)

func IsBackendAlive(ctx context.Context, aliveChannel chan bool, u *url.URL) {
	var d net.Dialer
	conn, err := d.DialContext(ctx, "tcp", u.Host)
	if err != nil {
		log.Println("Site unreachable, error: ", err)
		aliveChannel <- false
	}
	_ = conn.Close()
	aliveChannel <- true
}
