package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"text/tabwriter"
	"time"
)

const version = "1.0.0"

func printBanner() {
	fmt.Printf("%s%sgocli%s - Simple & Fast API / Website Health Monitor (v%s)\n\n", colorBold, colorCyan, colorReset, version)
}

func printUsage() {
	printBanner()
	fmt.Printf("%sUSAGE:%s\n", colorBold, colorReset)
	fmt.Println("  gocli check <url>             Quickly check status, latency & SSL of a single URL")
	fmt.Println("  gocli list                    Check and display status for all saved endpoints")
	fmt.Println("  gocli add <name> <url>        Save a new endpoint to monitor")
	fmt.Println("  gocli remove <name>           Remove an endpoint from the saved list")
	fmt.Println("  gocli watch [interval]        Continuously monitor saved endpoints (e.g. 5s, 10s)")
	fmt.Println("  gocli help                    Show this help message")
	fmt.Println()
	fmt.Printf("%sEXAMPLES:%s\n", colorBold, colorReset)
	fmt.Println("  gocli check https://google.com")
	fmt.Println("  gocli check localhost:8080/health")
	fmt.Println("  gocli add my-api https://api.github.com")
	fmt.Println("  gocli list")
	fmt.Println("  gocli watch 5s")
}

func handleCheck(args []string) {
	if len(args) < 1 {
		fmt.Printf("%sError: Missing URL to check.%s\n", colorRed, colorReset)
		fmt.Println("Usage: gocli check <url>")
		os.Exit(1)
	}

	url := args[0]
	fmt.Printf("Pinging %s%s%s...\n\n", colorCyan, url, colorReset)
	result := CheckSite("target", url, 8*time.Second)
	PrintCheckResult(result)
}

func handleList() {
	sites, err := LoadSites()
	if err != nil {
		fmt.Printf("%sFailed to load sites: %v%s\n", colorRed, err, colorReset)
		os.Exit(1)
	}

	if len(sites) == 0 {
		fmt.Printf("%sNo saved endpoints found.%s Use 'gocli add <name> <url>' to add one.\n", colorYellow, colorReset)
		return
	}

	fmt.Printf("Checking %d endpoint(s)...\n\n", len(sites))

	results := make([]CheckResult, len(sites))
	var wg sync.WaitGroup
	for i, site := range sites {
		wg.Add(1)
		go func(idx int, s Site) {
			defer wg.Done()
			results[idx] = CheckSite(s.Name, s.URL, 5*time.Second)
		}(i, site)
	}
	wg.Wait()

	printResultsTable(results)
}

func printResultsTable(results []CheckResult) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintf(w, "%sSTATUS\tNAME\tLATENCY\tSSL EXPIRY\tURL%s\n", colorBold, colorReset)
	fmt.Fprintf(w, "%s------\t----\t-------\t----------\t---%s\n", colorGray, colorReset)

	upCount := 0
	for _, res := range results {
		var statusStr string
		var latencyStr string
		var sslStr string

		if res.Error != nil {
			statusStr = fmt.Sprintf("%s[DOWN]%s", colorRed, colorReset)
			latencyStr = fmt.Sprintf("%s%v%s", colorRed, res.Latency.Round(time.Millisecond), colorReset)
			sslStr = fmt.Sprintf("%s-%s", colorGray, colorReset)
		} else {
			upCount++
			if res.StatusCode >= 200 && res.StatusCode < 300 {
				statusStr = fmt.Sprintf("%s[%d OK]%s", colorGreen, res.StatusCode, colorReset)
			} else if res.StatusCode >= 300 && res.StatusCode < 400 {
				statusStr = fmt.Sprintf("%s[%d]%s", colorCyan, res.StatusCode, colorReset)
			} else {
				statusStr = fmt.Sprintf("%s[%d]%s", colorYellow, res.StatusCode, colorReset)
			}

			if res.Latency < 200*time.Millisecond {
				latencyStr = fmt.Sprintf("%s%v%s", colorGreen, res.Latency.Round(time.Millisecond), colorReset)
			} else if res.Latency < 800*time.Millisecond {
				latencyStr = fmt.Sprintf("%s%v%s", colorYellow, res.Latency.Round(time.Millisecond), colorReset)
			} else {
				latencyStr = fmt.Sprintf("%s%v%s", colorRed, res.Latency.Round(time.Millisecond), colorReset)
			}

			if res.SSLExpiryDays >= 0 {
				if res.SSLExpiryDays > 30 {
					sslStr = fmt.Sprintf("%s%d days%s", colorGreen, res.SSLExpiryDays, colorReset)
				} else if res.SSLExpiryDays > 7 {
					sslStr = fmt.Sprintf("%s%d days%s", colorYellow, res.SSLExpiryDays, colorReset)
				} else {
					sslStr = fmt.Sprintf("%s%d days!%s", colorRed, res.SSLExpiryDays, colorReset)
				}
			} else if strings.HasPrefix(res.URL, "https://") {
				sslStr = fmt.Sprintf("%sUnknown%s", colorGray, colorReset)
			} else {
				sslStr = fmt.Sprintf("%sHTTP%s", colorGray, colorReset)
			}
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", statusStr, res.Name, latencyStr, sslStr, res.URL)
	}
	w.Flush()

	fmt.Printf("\n%sSummary:%s %d/%d online\n", colorBold, colorReset, upCount, len(results))
}

func handleAdd(args []string) {
	if len(args) < 2 {
		fmt.Printf("%sError: 'add' requires <name> and <url>.%s\n", colorRed, colorReset)
		fmt.Println("Usage: gocli add <name> <url>")
		os.Exit(1)
	}

	name := args[0]
	url := args[1]

	err := AddSite(name, url)
	if err != nil {
		fmt.Printf("%sFailed to add site: %v%s\n", colorRed, err, colorReset)
		os.Exit(1)
	}

	fmt.Printf("%s✓ Added endpoint %s%q%s (%s)%s\n", colorGreen, colorBold, name, colorReset, NormalizeURL(url), colorReset)
}

func handleRemove(args []string) {
	if len(args) < 1 {
		fmt.Printf("%sError: 'remove' requires <name>.%s\n", colorRed, colorReset)
		fmt.Println("Usage: gocli remove <name>")
		os.Exit(1)
	}

	name := args[0]
	err := RemoveSite(name)
	if err != nil {
		fmt.Printf("%sFailed to remove site: %v%s\n", colorRed, err, colorReset)
		os.Exit(1)
	}

	fmt.Printf("%s✓ Removed endpoint %s%q%s\n", colorGreen, colorBold, name, colorReset)
}

func handleWatch(args []string) {
	interval := 5 * time.Second
	if len(args) > 0 {
		parsed, err := time.ParseDuration(args[0])
		if err != nil {
			fmt.Printf("%sInvalid duration format %q. Using 5s default.%s\n", colorYellow, args[0], colorReset)
		} else {
			interval = parsed
		}
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	fmt.Printf("Starting live monitor (refreshing every %v). Press Ctrl+C to stop.\n\n", interval)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	runWatchCycle()

	for {
		select {
		case <-ticker.C:
			runWatchCycle()
		case <-sigChan:
			fmt.Println("\nMonitoring stopped.")
			return
		}
	}
}

func runWatchCycle() {
	timestamp := time.Now().Format("15:04:05")
	fmt.Printf("\n[%s] Probing endpoints...\n", timestamp)

	sites, err := LoadSites()
	if err != nil || len(sites) == 0 {
		fmt.Println("No sites configured to watch.")
		return
	}

	results := make([]CheckResult, len(sites))
	var wg sync.WaitGroup
	for i, site := range sites {
		wg.Add(1)
		go func(idx int, s Site) {
			defer wg.Done()
			results[idx] = CheckSite(s.Name, s.URL, 4*time.Second)
		}(i, site)
	}
	wg.Wait()

	printResultsTable(results)
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		return
	}

	cmd := strings.ToLower(os.Args[1])
	args := os.Args[2:]

	switch cmd {
	case "check":
		handleCheck(args)
	case "list", "ls":
		handleList()
	case "add":
		handleAdd(args)
	case "remove", "rm", "delete":
		handleRemove(args)
	case "watch", "monitor":
		handleWatch(args)
	case "help", "-h", "--help":
		printUsage()
	case "version", "-v", "--version":
		fmt.Printf("gocli version %s\n", version)
	default:
		fmt.Printf("%sUnknown command: %q%s\n\n", colorRed, cmd, colorReset)
		printUsage()
		os.Exit(1)
	}
}
