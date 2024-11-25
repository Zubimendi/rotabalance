package main

import (
    "fmt"
    "log"
    "net/http"
)

func handler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Hello from backend server: %s", r.URL.Path)
}

func main() {
    http.HandleFunc("/", handler)

    // Change the port number to run multiple servers
    log.Println("Starting backend server on port 8082...")
    if err := http.ListenAndServe(":8082", nil); err != nil {
        log.Fatal("Error starting server: ", err)
    }
}
