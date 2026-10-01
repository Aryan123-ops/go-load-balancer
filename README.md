# Go Load Balancer — Round Robin

A simple load balancer implementation in Go using the **Round Robin algorithm**.

This project demonstrates:

* Multiple Go backend servers
* Round Robin request distribution
* Reverse proxy using `httputil.ReverseProxy`
* Backend health checks
* Automatic detection of unhealthy servers
* Automatic recovery of backend servers
* Graceful shutdown
* HTTP connection pooling using `http.Transport`

## Project Structure

```text
go-load-balancer/
│
├── go.mod
│
├── backend/
│   └── main.go
│
├── loadbalancer/
│   └── main.go
│
└── README.md
```

## How It Works

The load balancer runs on port `8080` and distributes incoming requests between three backend servers.

```text
                         Client
                           |
                           v
                  +-------------------+
                  |   Load Balancer   |
                  |      :8080        |
                  +-------------------+
                    /       |       \
                   /        |        \
                  v         v         v
             Server-1   Server-2   Server-3
               :9001      :9002      :9003
```

The Round Robin algorithm distributes requests in this order:

```text
Request 1 → Server-1
Request 2 → Server-2
Request 3 → Server-3
Request 4 → Server-1
Request 5 → Server-2
Request 6 → Server-3
```

## 1. Start Backend Servers

Open **three separate Command Prompt/PowerShell windows**.

### Backend Server 1

```cmd
set PORT=9001
set SERVER_NAME=server-1
go run .\backend\main.go
```

Server 1 will run on:

```text
http://127.0.0.1:9001
```

### Backend Server 2

```cmd
set PORT=9002
set SERVER_NAME=server-2
go run .\backend\main.go
```

Server 2 will run on:

```text
http://127.0.0.1:9002
```

### Backend Server 3

```cmd
set PORT=9003
set SERVER_NAME=server-3
go run .\backend\main.go
```

Server 3 will run on:

```text
http://127.0.0.1:9003
```

## 2. Start the Load Balancer

Open another Command Prompt/PowerShell window:

```cmd
go run .\loadbalancer\main.go
```

The load balancer will run on:

```text
http://localhost:8080
```

## 3. Check Backend Health

Each backend provides a `/health` endpoint.

### Server 1

```text
http://127.0.0.1:9001/health
```

### Server 2

```text
http://127.0.0.1:9002/health
```

### Server 3

```text
http://127.0.0.1:9003/health
```

Example response:

```json
{
    "server": "server-1",
    "status": "UP"
}
```

These health endpoints are also used by the load balancer to determine whether a backend server is available.

## 4. Test the Load Balancer

Send requests to:

```text
http://localhost:8080/api/users
```

The load balancer forwards the request to one of the available backend servers.

Example response:

```json
{
    "server": "server-3",
    "message": "Request handled successfully",
    "time": "2026-10-01T13:08:04+05:30"
}
```

Send the request multiple times and the response will come from different backend servers.

Example:

```text
Request 1 → server-1
Request 2 → server-2
Request 3 → server-3
Request 4 → server-1
Request 5 → server-2
Request 6 → server-3
```

## 5. Backend Failure Detection

The load balancer performs periodic health checks against each backend.

For example:

```text
Server-1 :9001 → UP
Server-2 :9002 → UP
Server-3 :9003 → UP
```

If Server 2 is stopped using:

```text
CTRL + C
```

the load balancer will detect that Server 2 is unavailable through its `/health` endpoint.

The load balancer will log:

```text
Backend DOWN: http://localhost:9002
```

Server 2 will then be removed from request rotation.

Requests will continue to be distributed between the remaining healthy servers:

```text
Request 1 → server-1
Request 2 → server-3
Request 3 → server-1
Request 4 → server-3
```

## 6. Backend Recovery

Start Server 2 again:

```cmd
set PORT=9002
set SERVER_NAME=server-2
go run .\backend\main.go
```

After the next successful health check, the load balancer will detect the recovery:

```text
Backend RECOVERED: http://localhost:9002
```

Server 2 will automatically participate in Round Robin routing again.

```text
Request 1 → server-1
Request 2 → server-2
Request 3 → server-3
```

## 7. HTTP Connection Pooling

The load balancer uses Go's `http.Transport` for connection management:

```go
transport := &http.Transport{
    MaxIdleConns:        100,
    MaxIdleConnsPerHost: 20,
    IdleConnTimeout:     90 * time.Second,
}
```

This allows the reverse proxy to reuse HTTP connections instead of creating a new connection for every request.

## Technologies Used

* Go
* `net/http`
* `httputil.ReverseProxy`
* Round Robin algorithm
* HTTP health checks
* HTTP connection pooling
* Goroutines
* Atomic operations
* Graceful shutdown

## API Endpoints

| Service       | Endpoint          | Purpose                    |
| ------------- | ----------------- | -------------------------- |
| Server 1      | `:9001/health`    | Check Server 1 health      |
| Server 2      | `:9002/health`    | Check Server 2 health      |
| Server 3      | `:9003/health`    | Check Server 3 health      |
| Load Balancer | `:8080/api/users` | Route request to a backend |

## Key Concepts Demonstrated

### Round Robin

Requests are distributed sequentially across healthy backend servers.

### Health Check

The load balancer periodically checks whether each backend is available.

### Reverse Proxy

The load balancer uses Go's `httputil.ReverseProxy` to forward requests to backend servers.

### Fault Tolerance

If a backend server becomes unavailable, traffic is automatically routed to the remaining healthy servers.

### Recovery

When the failed server becomes available again, it is automatically added back into the load-balancing rotation.

## Future Improvements

Possible improvements for this project include:

* Retry mechanism
* Circuit breaker
* Weighted Round Robin
* Least Connections algorithm
* Prometheus metrics
* Request tracing
* Dynamic backend registration
* Configuration through environment variables
* Docker support
* Kubernetes deployment
* Distributed load balancing
