package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Site struct {
	Name    string    `json:"name"`
	URL     string    `json:"url"`
	AddedAt time.Time `json:"added_at"`
}

func getStoragePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "sites.json", nil
	}
	dir := filepath.Join(homeDir, ".gocli")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	return filepath.Join(dir, "sites.json"), nil
}

func LoadSites() ([]Site, error) {
	path, err := getStoragePath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		defaultSites := []Site{
			{Name: "google", URL: "https://google.com", AddedAt: time.Now()},
			{Name: "github", URL: "https://github.com", AddedAt: time.Now()},
			{Name: "cloudflare-dns", URL: "https://1.1.1.1", AddedAt: time.Now()},
		}
		_ = SaveSites(defaultSites)
		return defaultSites, nil
	} else if err != nil {
		return nil, fmt.Errorf("failed to read sites file: %w", err)
	}

	var sites []Site
	if err := json.Unmarshal(data, &sites); err != nil {
		return nil, fmt.Errorf("failed to parse sites json: %w", err)
	}

	return sites, nil
}

func SaveSites(sites []Site) error {
	path, err := getStoragePath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(sites, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to encode sites: %w", err)
	}

	return os.WriteFile(path, data, 0644)
}

func AddSite(name, url string) error {
	name = strings.TrimSpace(name)
	url = NormalizeURL(strings.TrimSpace(url))

	if name == "" || url == "" {
		return fmt.Errorf("name and url must not be empty")
	}

	sites, err := LoadSites()
	if err != nil {
		return err
	}

	for _, s := range sites {
		if strings.EqualFold(s.Name, name) {
			return fmt.Errorf("site with name %q already exists (url: %s)", name, s.URL)
		}
	}

	sites = append(sites, Site{
		Name:    name,
		URL:     url,
		AddedAt: time.Now(),
	})

	return SaveSites(sites)
}

func RemoveSite(name string) error {
	name = strings.TrimSpace(name)
	sites, err := LoadSites()
	if err != nil {
		return err
	}

	found := false
	newSites := make([]Site, 0, len(sites))
	for _, s := range sites {
		if strings.EqualFold(s.Name, name) {
			found = true
			continue
		}
		newSites = append(newSites, s)
	}

	if !found {
		return fmt.Errorf("site %q not found in saved list", name)
	}

	return SaveSites(newSites)
}
