# Rotabalance

Rotabalance is a lightweight and efficient load balancer written in [Go](https://golang.org/). It uses a thread-safe round-robin algorithm to evenly distribute traffic across backend servers while dynamically handling server availability. 

With simplicity and high performance as its core principles, Rotabalance is perfect for developers who need a reliable load-balancing solution without the complexity of heavyweight tools.

## Features

- **Round-Robin Load Balancing:** Distributes requests evenly across available servers.
- **Health Checks:** Automatically skips unhealthy servers during traffic distribution.
- **Failover Support:** Ensures high availability by dynamically handling server outages.
- **Concurrency-Safe:** Uses Go's `atomic` package for efficient and safe multi-threaded operation.
- **Configurable and Extendable:** Easily add custom logic for routing and server selection.

## Getting Started

### Prerequisites

- Go 1.18 or later installed on your machine.
- Basic understanding of Go modules.

### Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/yourusername/rotabalance.git
   ```
## Navigate to the Project Directory

To get started, navigate to the Rotabalance project directory:

```bash
cd rotabalance
```
### Build the Application
Build the load balancer application using the following command:

```bash
go build -o rotabalance
```
### Usage
Example Configuration
To configure your backend servers, modify the main.go file as shown below:

```go
package main

import (
    "log"
    "net/http"
    "net/http/httputil"
    "net/url"
)

func main() {
    // Backend servers
    backends := []*Backend{
        {URL: mustParseURL("http://localhost:8081"), Alive: true},
        {URL: mustParseURL("http://localhost:8082"), Alive: true},
        {URL: mustParseURL("http://localhost:8083"), Alive: true},
    }

    // Initialize ServerPool
    serverPool := &ServerPool{backends: backends}

    // Start the server
    log.Println("Starting Rotabalance on :8080")
    log.Fatal(http.ListenAndServe(":8080", serverPool))
}

func mustParseURL(rawURL string) *url.URL {
    u, err := url.Parse(rawURL)
    if err != nil {
        log.Fatalf("Failed to parse URL: %v", err)
    }
    return u
}
```

### Start the Application
Run the load balancer application:

```bash
./rotabalance
```

Send traffic to the load balancer, and it will automatically distribute requests to the backend servers.

### Health Checks
Rotabalance includes built-in health checks to ensure traffic is only routed to "alive" backend servers. Unavailable servers are automatically skipped until they are back online.

### Contributing
We welcome contributions from the community! If you have an idea, bug fix, or enhancement, feel free to submit an issue or pull request on the GitHub repository.

## Steps to Contribute
Fork the repository to your GitHub account.
Create a new branch for your feature or fix:
```bash
git checkout -b feature-name
```
Commit your changes:

```bash
git commit -m "Add feature-name"
```
Push your branch to your fork:
```bash
git push origin feature-name
```
Open a pull request to the main repository.

### License
This project is licensed under the MIT License. See the LICENSE file for details.

### Acknowledgments
Rotabalance is inspired by the simplicity of round-robin algorithms and the power of Go's concurrency primitives. We extend special thanks to the open-source community for their invaluable contributions and support.

Happy balancing! 🚀
