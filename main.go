package main

import (
	"fmt"
	"log"
	"net/http"
)

var port int = 8080

func main() {
	// Define backend servers
	backends := []string{
		"http://localhost:8081",
		"http://localhost:8082",
		"http://localhost:8083",
	}

	// Initialize the server pool
	InitServerPool(backends)

	// Create the server with the load balancer handler
	server := http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: http.HandlerFunc(lb), // lb function is defined here
	}

	// Log server start
	log.Printf("Starting load balancer on port %d...", port)

	// Start the server
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// lb is the load balancer's main request handler
func lb(w http.ResponseWriter, r *http.Request) {
	peer := ServerPool.GetNextPeer()
	if peer != nil {
		log.Printf("Forwarding request to: %s", peer.URL)
		peer.ReverseProxy.ServeHTTP(w, r)
		return
	}
	http.Error(w, "Service not available", http.StatusServiceUnavailable)
}
