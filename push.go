package metrics

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/golang/snappy"
	"github.com/prometheus/prometheus/prompb"
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

	// Convert to Prometheus remote write format
	var timeseries []prompb.TimeSeries
	now := time.Now().UnixMilli()

	for _, mf := range metricFamilies {
		for _, metric := range mf.GetMetric() {
			// Create labels
			labels := []prompb.Label{
				{Name: "__name__", Value: mf.GetName()},
			}
			for _, label := range metric.GetLabel() {
				labels = append(labels, prompb.Label{
					Name:  label.GetName(),
					Value: label.GetValue(),
				})
			}

			// Get metric value
			var value float64
			switch mf.GetType() {
			case 0: // COUNTER
				if metric.Counter != nil {
					value = metric.Counter.GetValue()
				}
			case 1: // GAUGE
				if metric.Gauge != nil {
					value = metric.Gauge.GetValue()
				}
			case 2: // SUMMARY
				if metric.Summary != nil {
					value = metric.Summary.GetSampleSum()
				}
			case 4: // HISTOGRAM
				if metric.Histogram != nil {
					value = metric.Histogram.GetSampleSum()
				}
			}

			timeseries = append(timeseries, prompb.TimeSeries{
				Labels: labels,
				Samples: []prompb.Sample{
					{
						Value:     value,
						Timestamp: now,
					},
				},
			})
		}
	}

	// Create write request
	writeRequest := &prompb.WriteRequest{
		Timeseries: timeseries,
	}

	// Marshal to protobuf
	data, err := proto.Marshal(writeRequest)
	if err != nil {
		return fmt.Errorf("failed to marshal protobuf: %w", err)
	}

	// Compress with Snappy
	compressed := snappy.Encode(nil, data)

	// Create HTTP request
	req, err := http.NewRequest("POST", m.config.GrafanaCloudURL, bytes.NewReader(compressed))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Encoding", "snappy")
	req.Header.Set("Content-Type", "application/x-protobuf")
	req.Header.Set("X-Prometheus-Remote-Write-Version", "0.1.0")
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
