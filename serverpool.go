package main

import (
	"net/url"
	"sync/atomic"
)

type ServerPoolObject struct {
	backends []*Backend
	current  uint64
}

var ServerPool = &ServerPoolObject{
	backends: []*Backend{},
}

func InitServerPool(backends []string) {
	for _, backend := range backends {
		url, err := url.Parse(backend)
		if err != nil {
			panic(err) // Handle error gracefully in production
		}

		proxy := NewReverseProxy(url)
		ServerPool.backends = append(ServerPool.backends, &Backend{
			URL:          url,
			Alive:        true,
			ReverseProxy: proxy,
		})
	}
}

func (s *ServerPoolObject) NextIndex() int {
	if len(s.backends) == 0 {
		return -1
	}
	return int(atomic.AddUint64(&s.current, uint64(1)) % uint64(len(s.backends)))
}

func (s *ServerPoolObject) GetNextPeer() *Backend {
	if len(s.backends) == 0 {
		return nil
	}

	next := s.NextIndex()
	totalBackends := len(s.backends)

	for i := 0; i < totalBackends; i++ {
		idx := (next + i) % totalBackends
		if s.backends[idx].IsAlive() {
			if idx != next {
				atomic.StoreUint64(&s.current, uint64(idx))
			}
			return s.backends[idx]
		}
	}
	return nil
}

func (s *ServerPoolObject) MarkBackendStatus(url *url.URL, alive bool) {
	for _, backend := range s.backends {
		if backend.URL.String() == url.String() {
			backend.SetAlive(alive)
			return
		}
	}
}
