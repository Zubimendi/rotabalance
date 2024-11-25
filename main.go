package loadbalancer

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"sync/atomic"
	"time"
)

type contextKey string

const (
	Retry    contextKey = "Retry"
	Attempts contextKey = "Attempts"
)

var port int = 8080

type Backend struct {
	URL          *url.URL
	Alive        bool
	mux          sync.RWMutex
	ReverseProxy *httputil.ReverseProxy
}

type ServerPool struct {
	backends []*Backend
	current  uint64
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

		proxy := httputil.NewSingleHostReverseProxy(url)

		// Assign the error handler for the proxy
		proxy.ErrorHandler = func(writer http.ResponseWriter, request *http.Request, e error) {
			log.Printf("[%s] %s\n", url.Host, e.Error())
			retries := GetRetryFromContext(request)
			if retries < 3 {
				select {
				case <-time.After(10 * time.Millisecond):
					ctx := context.WithValue(request.Context(), Retry, retries+1)
					proxy.ServeHTTP(writer, request.WithContext(ctx))
				}
				return
			}

			// After 3 retries, mark this backend as down
			serverPool.MarkBackendStatus(url, false)

			// Retry with another backend
			attempts := GetAttemptsFromContext(request)
			log.Printf("%s(%s) Attempting retry %d\n", request.RemoteAddr, request.URL.Path, attempts)
			ctx := context.WithValue(request.Context(), Attempts, attempts+1)
			lb(writer, request.WithContext(ctx))
		}

		serverPool.backends = append(serverPool.backends, &Backend{
			URL:          url,
			Alive:        true, // Assume all backends are alive initially
			ReverseProxy: proxy,
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
		return -1
	}
	return int(atomic.AddUint64(&s.current, uint64(1)) % uint64(len(s.backends)))
}

func (s *ServerPool) ResetCounter() {
	atomic.StoreUint64(&s.current, 0)
}

func (s *ServerPool) GetNextPeer() *Backend {
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

func (s *ServerPool) MarkBackendStatus(url *url.URL, alive bool) {
	for _, backend := range s.backends {
		if backend.URL.String() == url.String() {
			backend.SetAlive(alive)
			log.Printf("Backend %s marked as %v", url, alive)
			return
		}
	}
	log.Printf("Backend %s not found in the server pool", url)
}

func (b *Backend) SetAlive(alive bool) {
	b.mux.Lock()
	b.Alive = alive
	b.mux.Unlock()
}

func (b *Backend) IsAlive() bool {
	b.mux.RLock()
	defer b.mux.RUnlock()
	return b.Alive
}

func lb(w http.ResponseWriter, r *http.Request) {
	peer := serverPool.GetNextPeer()
	if peer != nil {
		log.Printf("Forwarding request to: %s", peer.URL)
		peer.ReverseProxy.ServeHTTP(w, r)
		return
	}
	http.Error(w, "Service not available", http.StatusServiceUnavailable)
}

// GetRetryFromContext retrieves the retry count from the request context
func GetRetryFromContext(r *http.Request) int {
	if retries, ok := r.Context().Value(Retry).(int); ok {
		return retries
	}
	return 0
}

// GetAttemptsFromContext retrieves the attempt count from the request context
func GetAttemptsFromContext(r *http.Request) int {
	if attempts, ok := r.Context().Value(Attempts).(int); ok {
		return attempts
	}
	return 0
}

//TODO: 
// 1. Make Application dynamic by accepting user inputs
// 2. Refactor codebase and organize functions in different files