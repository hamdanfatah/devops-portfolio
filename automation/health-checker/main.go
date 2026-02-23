package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds the health checker configuration
type Config struct {
	Endpoints []Endpoint `yaml:"endpoints"`
	Interval  int        `yaml:"interval_seconds"`
	Webhook   Webhook    `yaml:"webhook"`
}

// Endpoint defines a service to monitor
type Endpoint struct {
	Name    string `yaml:"name"`
	URL     string `yaml:"url"`
	Method  string `yaml:"method"`
	Timeout int    `yaml:"timeout_seconds"`
	Expect  int    `yaml:"expected_status"`
}

// Webhook holds notification configuration
type Webhook struct {
	URL     string `yaml:"url"`
	Enabled bool   `yaml:"enabled"`
}

// CheckResult holds the result of a health check
type CheckResult struct {
	Name       string        `json:"name"`
	URL        string        `json:"url"`
	Status     string        `json:"status"`
	StatusCode int           `json:"status_code"`
	Latency    time.Duration `json:"latency_ms"`
	Error      string        `json:"error,omitempty"`
	Timestamp  time.Time     `json:"timestamp"`
}

func main() {
	configFile := "config.yaml"
	if len(os.Args) > 1 {
		configFile = os.Args[1]
	}

	cfg, err := loadConfig(configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to load config: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("🏥 Health Checker started — monitoring %d endpoints\n", len(cfg.Endpoints))
	fmt.Printf("📡 Check interval: %ds\n\n", cfg.Interval)

	// Run initial check
	results := checkAll(cfg.Endpoints)
	printResults(results)
	notifyUnhealthy(cfg.Webhook, results)

	// Continuous monitoring
	ticker := time.NewTicker(time.Duration(cfg.Interval) * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		results := checkAll(cfg.Endpoints)
		printResults(results)
		notifyUnhealthy(cfg.Webhook, results)
	}
}

func loadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	// Set defaults
	if cfg.Interval <= 0 {
		cfg.Interval = 30
	}
	for i := range cfg.Endpoints {
		if cfg.Endpoints[i].Method == "" {
			cfg.Endpoints[i].Method = "GET"
		}
		if cfg.Endpoints[i].Timeout <= 0 {
			cfg.Endpoints[i].Timeout = 5
		}
		if cfg.Endpoints[i].Expect <= 0 {
			cfg.Endpoints[i].Expect = 200
		}
	}

	return &cfg, nil
}

// checkAll performs health checks concurrently using goroutines
func checkAll(endpoints []Endpoint) []CheckResult {
	results := make([]CheckResult, len(endpoints))
	var wg sync.WaitGroup

	for i, ep := range endpoints {
		wg.Add(1)
		go func(idx int, endpoint Endpoint) {
			defer wg.Done()
			results[idx] = checkEndpoint(endpoint)
		}(i, ep)
	}

	wg.Wait()
	return results
}

func checkEndpoint(ep Endpoint) CheckResult {
	result := CheckResult{
		Name:      ep.Name,
		URL:       ep.URL,
		Timestamp: time.Now(),
	}

	client := &http.Client{
		Timeout: time.Duration(ep.Timeout) * time.Second,
	}

	start := time.Now()

	req, err := http.NewRequest(ep.Method, ep.URL, nil)
	if err != nil {
		result.Status = "ERROR"
		result.Error = err.Error()
		return result
	}

	resp, err := client.Do(req)
	result.Latency = time.Since(start)

	if err != nil {
		result.Status = "DOWN"
		result.Error = err.Error()
		return result
	}
	defer resp.Body.Close()

	result.StatusCode = resp.StatusCode
	if resp.StatusCode == ep.Expect {
		result.Status = "HEALTHY"
	} else {
		result.Status = "UNHEALTHY"
		result.Error = fmt.Sprintf("expected %d, got %d", ep.Expect, resp.StatusCode)
	}

	return result
}

func printResults(results []CheckResult) {
	fmt.Printf("\n─── Health Check Report [%s] ───\n", time.Now().Format("15:04:05"))
	for _, r := range results {
		icon := "✅"
		if r.Status != "HEALTHY" {
			icon = "❌"
		}
		fmt.Printf("%s %-25s | %-10s | %6dms | %s\n",
			icon, r.Name, r.Status, r.Latency.Milliseconds(), r.URL)
		if r.Error != "" {
			fmt.Printf("   └─ Error: %s\n", r.Error)
		}
	}
}

func notifyUnhealthy(webhook Webhook, results []CheckResult) {
	if !webhook.Enabled || webhook.URL == "" {
		return
	}

	var unhealthy []string
	for _, r := range results {
		if r.Status != "HEALTHY" {
			unhealthy = append(unhealthy, fmt.Sprintf("• %s (%s): %s", r.Name, r.Status, r.Error))
		}
	}

	if len(unhealthy) == 0 {
		return
	}

	message := fmt.Sprintf("🚨 *Health Check Alert*\n%d service(s) unhealthy:\n%s",
		len(unhealthy), strings.Join(unhealthy, "\n"))

	payload := map[string]string{"text": message}
	data, _ := json.Marshal(payload)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(webhook.URL, "application/json", strings.NewReader(string(data)))
	if err != nil {
		fmt.Fprintf(os.Stderr, "⚠️  Failed to send webhook: %v\n", err)
		return
	}
	defer resp.Body.Close()
}
