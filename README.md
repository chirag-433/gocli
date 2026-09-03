# gocli

A lightweight, fast, and practical **Website & REST API Health Monitor** built in Go.

No bloated third-party dependencies—built entirely using the Go standard library (`net/http`, `crypto/tls`, goroutines, and channels).

---

## Features

- **Instant Status Checks**: Check status codes (200, 404, 500), response latency, and payload sizes.
- **SSL/TLS Certificate Tracker**: Automatically inspects HTTPS certificates and calculates days remaining until expiration.
- **Concurrent Multi-Endpoint Probing**: Uses Go goroutines (`sync.WaitGroup`) to check multiple saved endpoints simultaneously in milliseconds.
- **Persistent JSON Storage**: Easily add and remove target endpoints saved in `~/.gocli/sites.json`.
- **Live Watch Mode**: Continuously monitor your APIs and microservices with customizable intervals (`5s`, `10s`, `1m`).
- **Colorized Terminal Output**: Clean ANSI-formatted status badges and tab-aligned summary tables.
- **Zero External Dependencies**: Compiles to a single static standalone binary.

---

## Installation & Build

Ensure you have [Go](https://go.dev/dl/) installed (1.20+ recommended).

```bash
# Clone the repository
git clone https://github.com/chirag-433/gocli.git
cd gocli

# Build the executable
go build -o gocli .
```

On Windows, this produces `gocli.exe`.

---

## Usage & Commands

### 1. Check a Single Endpoint
Check any URL on the fly:
```bash
gocli check https://google.com
# or check a local server
gocli check localhost:8080/health
```

**Output:**
```text
Pinging https://google.com...

Target URL:    https://google.com
Status:        [200 OK]
Latency:       112ms
Payload Size:  83.8 KB
SSL Cert:      Valid (expires in 59 days)
```

---

### 2. List & Probe All Saved Endpoints
Concurrently tests all saved endpoints and displays an aligned status table:
```bash
gocli list
```

**Output:**
```text
Checking 3 endpoint(s)...

STATUS      NAME             LATENCY    SSL EXPIRY    URL
------      ----             -------    ----------    ---
[200 OK]    google           115ms      59 days       https://google.com
[200 OK]    github           209ms      87 days       https://github.com
[200 OK]    cloudflare-dns   45ms       109 days      https://1.1.1.1

Summary: 3/3 online
```

---

### 3. Add an Endpoint to Monitor
Save endpoints that you want to monitor regularly:
```bash
gocli add my-api https://api.github.com
```

---

### 4. Remove an Endpoint
```bash
gocli remove my-api
```

---

### 5. Live Monitoring (Watch Mode)
Continuously poll all endpoints at regular intervals:
```bash
# Default interval is 5 seconds
gocli watch

# Custom interval
gocli watch 10s
```

Press `Ctrl+C` at any time to gracefully stop monitoring.

---

## Running Tests

Unit tests are included and use Go's built-in `httptest.Server`:

```bash
go test -v ./...
```

---

## Project Structure

```text
gocli/
├── main.go          # CLI entrypoint, argument parsing, commands (check, list, add, remove, watch)
├── checker.go       # HTTP client probe logic, SSL certificate inspector, and metrics
├── storage.go       # JSON persistence for saved endpoints in ~/.gocli/sites.json
├── checker_test.go  # Unit tests using standard net/http/httptest
├── go.mod           # Go module definition
└── README.md        # Documentation
```

---

## License
MIT