package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNormalizeURL(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"example.com", "https://example.com"},
		{"http://example.com", "http://example.com"},
		{"https://example.com", "https://example.com"},
		{"  api.github.com  ", "https://api.github.com"},
	}

	for _, tc := range tests {
		got := NormalizeURL(tc.input)
		if got != tc.expected {
			t.Errorf("NormalizeURL(%q) = %q; expected %q", tc.input, got, tc.expected)
		}
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		input    int64
		expected string
	}{
		{500, "500 B"},
		{1024, "1.0 KB"},
		{2048, "2.0 KB"},
		{1048576, "1.0 MB"},
	}

	for _, tc := range tests {
		got := FormatBytes(tc.input)
		if got != tc.expected {
			t.Errorf("FormatBytes(%d) = %q; expected %q", tc.input, got, tc.expected)
		}
	}
}

func TestCheckSite_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello, World!"))
	}))
	defer server.Close()

	result := CheckSite("test-server", server.URL, 2*time.Second)

	if result.Error != nil {
		t.Fatalf("expected no error, got: %v", result.Error)
	}
	if result.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got: %d", result.StatusCode)
	}
	if result.SizeBytes != int64(len("Hello, World!")) {
		t.Errorf("expected size %d, got: %d", len("Hello, World!"), result.SizeBytes)
	}
}

func TestCheckSite_ErrorStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("server error"))
	}))
	defer server.Close()

	result := CheckSite("error-server", server.URL, 2*time.Second)

	if result.Error != nil {
		t.Fatalf("expected HTTP response, got transport error: %v", result.Error)
	}
	if result.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected status 500, got: %d", result.StatusCode)
	}
}

func TestCheckSite_Unreachable(t *testing.T) {
	result := CheckSite("down-server", "http://127.0.0.1:59999", 100*time.Millisecond)

	if result.Error == nil {
		t.Errorf("expected error for unreachable address, got nil")
	}
}
