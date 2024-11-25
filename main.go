package loadbalancer

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"sync/atomic"
)

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

func main() {
	u, err := url.Parse("http://localhost:8080")
	rp := httputil.NewSingleHostReverseProxy(u)

	if err != nil {
		log.Fatalf("Failed to parse backend URL: %v", err)
	}
	// initialize your server and add this as handler
	http.Handle("/load-balancer", http.HandlerFunc(rp.ServeHTTP))
}

func (s *ServerPool) NextIndex() int {
	if len(s.backends) == 0 {
		return -1 // Indicate no backends are available
	}
	return int(atomic.AddUint64(&s.current, uint64(1)) % uint64(len(s.backends)))
}

// Reset the counter to prevent the counter from becoming too large
func (s *ServerPool) ResetCounter() {
	atomic.StoreUint64(&s.current, 0) // Reset the counter safely
}

// GetNextPeer returns next active peer to take a connection
// GetNextPeer returns the next active backend to handle a connection
func (s *ServerPool) GetNextPeer() *Backend {
    // Ensure backends exist
    if len(s.backends) == 0 {
        return nil // No backends available
    }

    // Start from the next index
    next := s.NextIndex()
    totalBackends := len(s.backends)

    // Loop through backends in a circular fashion
    for i := 0; i < totalBackends; i++ {
        idx := (next + i) % totalBackends

        // If backend is alive, mark and return it
        if s.backends[idx].Alive {
            // Update the current index if it changed
            if idx != next {
                atomic.StoreUint64(&s.current, uint64(idx))
            }
            return s.backends[idx]
        }
    }

    // No alive backends found
    return nil
}
