package main

import (
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestStrictExitCodeForHTTPStatuses(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		wantExit   int
	}{
		{name: "2xx response is healthy", statusCode: http.StatusOK, wantExit: 0},
		{name: "last 2xx response is healthy", statusCode: 299, wantExit: 0},
		{name: "redirect is unhealthy", statusCode: http.StatusFound, wantExit: 1},
		{name: "client error is unhealthy", statusCode: http.StatusNotFound, wantExit: 1},
		{name: "server error is unhealthy", statusCode: http.StatusInternalServerError, wantExit: 1},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(test.statusCode)
			}))
			defer server.Close()

			result := CheckSite(test.name, server.URL, time.Second)
			if result.Error != nil {
				t.Fatalf("CheckSite returned unexpected error: %v", result.Error)
			}
			if got := strictExitCode([]CheckResult{result}); got != test.wantExit {
				t.Errorf("strictExitCode() = %d, want %d for status %d", got, test.wantExit, test.statusCode)
			}
		})
	}
}

func TestStrictExitCodeForNetworkError(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen on local port: %v", err)
	}
	address := listener.Addr().String()
	if err := listener.Close(); err != nil {
		t.Fatalf("close local listener: %v", err)
	}

	result := CheckSite("unreachable", "http://"+address, time.Second)
	if result.Error == nil {
		t.Fatal("CheckSite returned no error for an unreachable endpoint")
	}
	if got := strictExitCode([]CheckResult{result}); got != 1 {
		t.Errorf("strictExitCode() = %d, want 1 for a network error", got)
	}
}
