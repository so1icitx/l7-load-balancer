
#  Go-Balancer: 

This is a lightweight HTTP Load Balancer built in Go. It was created as a pedagogical project to explore the nuances of Go's concurrency model, thread safety, and network proxying.

> **Disclaimer:** This is **not** an enterprise-grade load balancer (use NGINX or HAProxy for that!). This is a "learning-out-loud" project focused on training Go-specific concepts like Mutexes, Goroutines, and Context.

##  Key Learning Objectives

Throughout this project, I tackled several critical "bottlenecks" in concurrent programming:

* **Thread-Safe State:** Using `sync.Mutex` to prevent race conditions when multiple goroutines update server health status.
* **The Round-Robin Closure:** Implementing a stateful closure that remembers the last server used without relying on global variables.
* **Goroutine Leak Prevention:** Using **Buffered Channels** and `context.WithTimeout` to ensure background health checks never hang or leak memory.
* **Reverse Proxying:** Utilizing `net/http/httputil` to forward requests while maintaining a custom health-check loop.

##  How it Works

1. **The Registry:** On startup, the program spins up 6 mock backend servers on ports `8081-8086`.
2. **The Health Worker:** A background "Doctor" goroutine picks a random server every 5 seconds to check its heartbeat. If a server takes longer than 3 seconds to respond, it is marked as `Unhealthy`.
3. **The Balancer:** The main entry point on port `8080` intercepts incoming traffic and routes it to the next available healthy server using a Round-Robin algorithm.

##  Getting Started

### Prerequisites

* Go 1.25+

### Running the Project

```bash
# Clone the repo
git clone https://github.com/so1icitx/l7-load-balancer.git

# Run the program
go run main.go

```

Once running, you can visit `http://localhost:8080`. You will see the Load Balancer distributing your requests across the various backend servers.
