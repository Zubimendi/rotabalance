package loadbalancer

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"sync/atomic"
)

var port int = 8080

type Backend struct {
	// URL of the backend server
	URL *url.URL

	// Alive indicates if the backend server is reachable
	Alive bool

	// mux protects access to the Alive flag for concurrency
	mux sync.RWMutex

	// ReverseProxy forwards requests to the backend server
	ReverseProxy *httputil.ReverseProxy
}

type ServerPool struct {
	// A Slice used to keep track of backends in our load balancer
	backends []*Backend
	// A counter variable
	current uint64
}

var serverPool = &ServerPool{
	backends: []*Backend{},
}

func main() {
	// Define backend servers
	backends := []string{
		"http://localhost:8081",
		"http://localhost:8082",
		"http://localhost:8083",
	}

	// Add backends to the server pool
	for _, backend := range backends {
		url, err := url.Parse(backend)
		if err != nil {
			log.Fatalf("Failed to parse backend URL %s: %v", backend, err)
		}

		serverPool.backends = append(serverPool.backends, &Backend{
			URL:          url,
			Alive:        true, // Assume all backends are alive initially
			ReverseProxy: httputil.NewSingleHostReverseProxy(url),
		})
	}

	// Create the server with the lb handler
	server := http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: http.HandlerFunc(lb),
	}

	// Log server start
	log.Printf("Starting load balancer on port %d...", port)

	// Start the server
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func (s *ServerPool) NextIndex() int {
	if len(s.backends) == 0 {
		return -1 // Indicate no backends are available
	}
	return int(atomic.AddUint64(&s.current, uint64(1)) % uint64(len(s.backends)))
}

// ResetCounter resets the counter safely to prevent overflow
func (s *ServerPool) ResetCounter() {
	atomic.StoreUint64(&s.current, 0)
}

// GetNextPeer returns the next active backend to handle a connection
func (s *ServerPool) GetNextPeer() *Backend {
	if len(s.backends) == 0 {
		return nil // No backends available
	}

	next := s.NextIndex()
	totalBackends := len(s.backends)

	// Loop through backends in a circular fashion
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

// SetAlive updates the Alive status of the Backend in a thread-safe manner.
func (b *Backend) SetAlive(alive bool) {
	b.mux.Lock()
	b.Alive = alive
	b.mux.Unlock()
}

// IsAlive checks if the Backend is alive in a thread-safe manner.
func (b *Backend) IsAlive() bool {
	b.mux.RLock()
	defer b.mux.RUnlock()
	return b.Alive
}

// lb load balances the incoming request
func lb(w http.ResponseWriter, r *http.Request) {
	peer := serverPool.GetNextPeer()
	if peer != nil {
		log.Printf("Forwarding request to: %s", peer.URL)
		peer.ReverseProxy.ServeHTTP(w, r)
		return
	}
	http.Error(w, "Service not available", http.StatusServiceUnavailable)
}