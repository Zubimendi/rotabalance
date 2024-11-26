package main

import (
	"context"
	"log"
	"net"
	"net/url"
	"time"
)

func isBackendAlive(u *url.URL) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	d := net.Dialer{}
	conn, err := d.DialContext(ctx, "tcp", u.Host)
	if err != nil {
		return false
	}

	_ = conn.Close()
	return true
}

func (s *ServerPoolObject) HealthCheck() {
	for _, b := range s.backends {
		alive := isBackendAlive(b.URL)
		b.SetAlive(alive)
		status := "up"
		if !alive {
			status = "down"
		}
		log.Printf("%s [%s]\n", b.URL, status)
	}
}
