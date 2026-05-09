# Simple API Gateway

A lightweight API gateway with request forwarding, load balancing, and health checking.

## Features

- Request forwarding to backend services
- Round-robin load balancing
- Health checking for backend services
- Request/Response logging
- Configurable backend services

## Project Structure

```
day2/
├── main.go           # Main entry point
├── gateway/          # Gateway implementation
│   ├── proxy.go      # Proxy/forwarding logic
│   ├── balancer.go   # Load balancer
│   └── health.go     # Health checker
└── config.go        # Configuration
```

## Getting Started

### Run the gateway

```bash
go run main.go
```

Gateway will start on `http://localhost:8080`

### Forward request to backend

```bash
curl http://localhost:8080/api/v1/users
```

### Health check

```bash
curl http://localhost:8080/health
```

## Configuration

Edit `config.go` to configure backend services:

```go
Backends: []Backend{
    {URL: "http://localhost:3001", Name: "service-1"},
    {URL: "http://localhost:3002", Name: "service-2"},
    {URL: "http://localhost:3003", Name: "service-3"},
}
```

## API Endpoints

| Endpoint | Description |
|----------|-------------|
| `/health` | Gateway health check |
| `/api/*` | Proxied requests to backends |
| `/metrics` | Basic metrics |

## Load Balancing

Currently uses round-robin algorithm. Requests are distributed evenly across all healthy backends.

## Health Checking

- Periodically checks backend health
- Removes unhealthy backends from rotation
- Automatically adds backends when they recover
- Check interval: 10 seconds
- Timeout: 5 seconds
