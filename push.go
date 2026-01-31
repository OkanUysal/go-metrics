package metrics

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/prometheus/common/expfmt"
)

// StartGrafanaPush starts pushing metrics to Grafana Cloud
func (m *Metrics) StartGrafanaPush(ctx context.Context) {
	if m.config.GrafanaCloudURL == "" || m.config.GrafanaCloudAPIKey == "" {
		return
	}

	interval := m.config.PushInterval
	if interval == 0 {
		interval = 15 * time.Second
	}

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		// Push immediately on start
		if err := m.pushToGrafana(); err != nil {
			fmt.Printf("Failed to push metrics to Grafana: %v\n", err)
		}

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := m.pushToGrafana(); err != nil {
					fmt.Printf("Failed to push metrics to Grafana: %v\n", err)
				}
			}
		}
	}()
}

// pushToGrafana pushes metrics to Grafana Cloud using Prometheus remote write
func (m *Metrics) pushToGrafana() error {
	// Gather metrics
	metricFamilies, err := m.registry.Gather()
	if err != nil {
		return fmt.Errorf("failed to gather metrics: %w", err)
	}

	// Encode metrics in Prometheus text format
	var buf bytes.Buffer
	encoder := expfmt.NewEncoder(&buf, expfmt.NewFormat(expfmt.TypeTextPlain))
	
	for _, mf := range metricFamilies {
		if err := encoder.Encode(mf); err != nil {
			return fmt.Errorf("failed to encode metric: %w", err)
		}
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", m.config.GrafanaCloudURL, &buf)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/openmetrics-text; version=1.0.0; charset=utf-8")
	req.Header.Set("User-Agent", "go-metrics/1.0")
	
	// Set basic auth
	req.SetBasicAuth(m.config.GrafanaCloudUser, m.config.GrafanaCloudAPIKey)

	// Send request
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to push metrics: %w", err)
	}
	defer resp.Body.Close()

	// Check response
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("push failed with status %d: %s", resp.StatusCode, string(body))
	}

	fmt.Printf("Successfully pushed %d metrics to Grafana Cloud\n", len(metricFamilies))
	return nil
}
