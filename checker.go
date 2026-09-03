package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorCyan   = "\033[36m"
	colorGray   = "\033[90m"
	colorBold   = "\033[1m"
)

type CheckResult struct {
	Name          string
	URL           string
	StatusCode    int
	StatusText    string
	Latency       time.Duration
	SizeBytes     int64
	SSLExpiryDays int
	Error         error
}

func NormalizeURL(rawURL string) string {
	rawURL = strings.TrimSpace(rawURL)
	if !strings.HasPrefix(rawURL, "http://") && !strings.HasPrefix(rawURL, "https://") {
		return "https://" + rawURL
	}
	return rawURL
}

func CheckSite(name, rawURL string, timeout time.Duration) CheckResult {
	targetURL := NormalizeURL(rawURL)

	client := &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: false,
			},
		},
	}

	start := time.Now()
	resp, err := client.Get(targetURL)
	latency := time.Since(start)

	if err != nil {
		return CheckResult{
			Name:          name,
			URL:           targetURL,
			Latency:       latency,
			SSLExpiryDays: -1,
			Error:         err,
		}
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 10*1024*1024))
	size := int64(len(bodyBytes))

	sslExpiryDays := -1
	if resp.TLS != nil && len(resp.TLS.PeerCertificates) > 0 {
		cert := resp.TLS.PeerCertificates[0]
		remaining := time.Until(cert.NotAfter)
		sslExpiryDays = int(remaining.Hours() / 24)
	}

	return CheckResult{
		Name:          name,
		URL:           targetURL,
		StatusCode:    resp.StatusCode,
		StatusText:    resp.Status,
		Latency:       latency,
		SizeBytes:     size,
		SSLExpiryDays: sslExpiryDays,
		Error:         nil,
	}
}

func FormatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func PrintCheckResult(res CheckResult) {
	fmt.Printf("%sTarget URL:%s    %s\n", colorBold, colorReset, res.URL)

	if res.Error != nil {
		fmt.Printf("%sStatus:%s        %s[DOWN] %v%s\n", colorBold, colorReset, colorRed, res.Error, colorReset)
		fmt.Printf("%sResponse Time:%s %v\n", colorBold, colorReset, res.Latency.Round(time.Millisecond))
		return
	}

	var statusColor string
	switch {
	case res.StatusCode >= 200 && res.StatusCode < 300:
		statusColor = colorGreen
	case res.StatusCode >= 300 && res.StatusCode < 400:
		statusColor = colorCyan
	case res.StatusCode >= 400 && res.StatusCode < 500:
		statusColor = colorYellow
	default:
		statusColor = colorRed
	}

	var latencyColor string
	switch {
	case res.Latency < 200*time.Millisecond:
		latencyColor = colorGreen
	case res.Latency < 800*time.Millisecond:
		latencyColor = colorYellow
	default:
		latencyColor = colorRed
	}

	fmt.Printf("%sStatus:%s        %s[%s]%s\n", colorBold, colorReset, statusColor, res.StatusText, colorReset)
	fmt.Printf("%sLatency:%s       %s%v%s\n", colorBold, colorReset, latencyColor, res.Latency.Round(time.Millisecond), colorReset)
	fmt.Printf("%sPayload Size:%s  %s\n", colorBold, colorReset, FormatBytes(res.SizeBytes))

	if res.SSLExpiryDays >= 0 {
		var sslColor string
		switch {
		case res.SSLExpiryDays > 30:
			sslColor = colorGreen
		case res.SSLExpiryDays > 7:
			sslColor = colorYellow
		default:
			sslColor = colorRed
		}
		fmt.Printf("%sSSL Cert:%s      %sValid (expires in %d days)%s\n", colorBold, colorReset, sslColor, res.SSLExpiryDays, colorReset)
	} else if strings.HasPrefix(res.URL, "https://") {
		fmt.Printf("%sSSL Cert:%s      %sNot available%s\n", colorBold, colorReset, colorGray, colorReset)
	} else {
		fmt.Printf("%sSSL Cert:%s      %sN/A (HTTP)%s\n", colorBold, colorReset, colorGray, colorReset)
	}
}
