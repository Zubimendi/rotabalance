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
func (s *ServerPool) GetNextPeer() *Backend {
	// loop entire backends to find out an Alive backend
	next := s.NextIndex()
	l := len(s.backends) + next // start from next and move a full cycle
	for i := next; i < l; i++ {
	  idx := i % len(s.backends) // take an index by modding with length
	  // if we have an alive backend, use it and store if its not the original one
	  if s.backends[idx].Alive {
		if i != next {
		  atomic.StoreUint64(&s.current, uint64(idx)) // mark the current one
		}
		return s.backends[idx]
	  }
	}
	return nil
  }
  