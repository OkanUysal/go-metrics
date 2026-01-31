package metrics

import (
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

func TestNewMetrics(t *testing.T) {
	t.Run("with nil config", func(t *testing.T) {
		m := NewMetrics(nil)
		if m == nil {
			t.Fatal("Expected metrics to be created")
		}
		if m.config == nil {
			t.Fatal("Expected default config to be used")
		}
		if m.config.ServiceName != "app" {
			t.Errorf("Expected default service name 'app', got '%s'", m.config.ServiceName)
		}
	})

	t.Run("with custom config", func(t *testing.T) {
		config := &Config{
			ServiceName:       "test-service",
			Namespace:         "test",
			EnableHTTPMetrics: true,
		}
		m := NewMetrics(config)
		if m.config.ServiceName != "test-service" {
			t.Errorf("Expected service name 'test-service', got '%s'", m.config.ServiceName)
		}
		if m.config.Namespace != "test" {
			t.Errorf("Expected namespace 'test', got '%s'", m.config.Namespace)
		}
	})

	t.Run("HTTP metrics enabled", func(t *testing.T) {
		config := &Config{
			ServiceName:       "test",
			Namespace:         "test",
			EnableHTTPMetrics: true,
		}
		m := NewMetrics(config)
		if m.httpMetrics == nil {
			t.Error("Expected HTTP metrics to be initialized")
		}
	})

	t.Run("HTTP metrics disabled", func(t *testing.T) {
		config := &Config{
			ServiceName:       "test",
			Namespace:         "test",
			EnableHTTPMetrics: false,
		}
		m := NewMetrics(config)
		if m.httpMetrics != nil {
			t.Error("Expected HTTP metrics to be nil when disabled")
		}
	})
}

func TestCounterMetrics(t *testing.T) {
	m := NewMetrics(&Config{
		ServiceName: "test",
		Namespace:   "test",
	})

	t.Run("increment counter", func(t *testing.T) {
		m.IncrementCounter("test_counter", nil)
		// Verify counter exists
		if _, exists := m.counters["test_counter"]; !exists {
			t.Error("Expected counter to be created")
		}
	})

	t.Run("increment counter with labels", func(t *testing.T) {
		m.IncrementCounter("test_counter_labeled", MetricLabels{
			"status": "success",
		})
		if _, exists := m.counters["test_counter_labeled"]; !exists {
			t.Error("Expected labeled counter to be created")
		}
	})

	t.Run("increment counter by value", func(t *testing.T) {
		m.IncrementCounterBy("test_counter_by", 5.0, nil)
		if _, exists := m.counters["test_counter_by"]; !exists {
			t.Error("Expected counter to be created")
		}
	})
}

func TestGaugeMetrics(t *testing.T) {
	m := NewMetrics(&Config{
		ServiceName: "test",
		Namespace:   "test",
	})

	t.Run("set gauge", func(t *testing.T) {
		m.SetGauge("test_gauge", 42.0, nil)
		if _, exists := m.gauges["test_gauge"]; !exists {
			t.Error("Expected gauge to be created")
		}
	})

	t.Run("increment gauge", func(t *testing.T) {
		m.IncrementGauge("test_gauge_inc", nil)
		if _, exists := m.gauges["test_gauge_inc"]; !exists {
			t.Error("Expected gauge to be created")
		}
	})

	t.Run("decrement gauge", func(t *testing.T) {
		m.DecrementGauge("test_gauge_dec", nil)
		if _, exists := m.gauges["test_gauge_dec"]; !exists {
			t.Error("Expected gauge to be created")
		}
	})

	t.Run("gauge with labels", func(t *testing.T) {
		m.SetGauge("test_gauge_labeled", 100.0, MetricLabels{
			"type": "memory",
		})
		if _, exists := m.gauges["test_gauge_labeled"]; !exists {
			t.Error("Expected labeled gauge to be created")
		}
	})
}

func TestHistogramMetrics(t *testing.T) {
	m := NewMetrics(&Config{
		ServiceName: "test",
		Namespace:   "test",
	})

	t.Run("record histogram", func(t *testing.T) {
		m.RecordHistogram("test_histogram", 0.5, nil)
		if _, exists := m.histograms["test_histogram"]; !exists {
			t.Error("Expected histogram to be created")
		}
	})

	t.Run("record histogram with labels", func(t *testing.T) {
		m.RecordHistogram("test_histogram_labeled", 1.5, MetricLabels{
			"operation": "query",
		})
		if _, exists := m.histograms["test_histogram_labeled"]; !exists {
			t.Error("Expected labeled histogram to be created")
		}
	})
}

func TestWebSocketMetrics(t *testing.T) {
	m := NewMetrics(&Config{
		ServiceName: "test",
		Namespace:   "test",
	})
	ws := m.NewWebSocketMetrics()

	t.Run("connection opened", func(t *testing.T) {
		ws.ConnectionOpened()
		if _, exists := m.gauges["websocket_connections_active"]; !exists {
			t.Error("Expected websocket_connections_active gauge to be created")
		}
	})

	t.Run("connection closed", func(t *testing.T) {
		ws.ConnectionClosed()
	})

	t.Run("message sent", func(t *testing.T) {
		ws.MessageSent("chat")
		if _, exists := m.counters["websocket_messages_sent_total"]; !exists {
			t.Error("Expected message sent counter to be created")
		}
	})

	t.Run("message received", func(t *testing.T) {
		ws.MessageReceived("game_event")
		if _, exists := m.counters["websocket_messages_received_total"]; !exists {
			t.Error("Expected message received counter to be created")
		}
	})

	t.Run("room metrics", func(t *testing.T) {
		ws.RoomCreated("match")
		ws.SetActiveRooms(5)
		ws.SetRoomClients("room_123", 10)
		ws.RoomClosed("match")

		if _, exists := m.counters["websocket_rooms_created_total"]; !exists {
			t.Error("Expected room created counter to be created")
		}
	})
}

func TestCacheMetrics(t *testing.T) {
	m := NewMetrics(&Config{
		ServiceName: "test",
		Namespace:   "test",
	})
	cache := m.NewCacheMetrics()

	t.Run("cache hit", func(t *testing.T) {
		cache.Hit("redis")
		if _, exists := m.counters["cache_hits_total"]; !exists {
			t.Error("Expected cache hits counter to be created")
		}
	})

	t.Run("cache miss", func(t *testing.T) {
		cache.Miss("memory")
		if _, exists := m.counters["cache_misses_total"]; !exists {
			t.Error("Expected cache misses counter to be created")
		}
	})

	t.Run("cache hit ratio", func(t *testing.T) {
		cache.SetHitRatio("redis", 0.95)
		if _, exists := m.gauges["cache_hit_ratio"]; !exists {
			t.Error("Expected hit ratio gauge to be created")
		}
	})

	t.Run("cache eviction", func(t *testing.T) {
		cache.Eviction("memory")
		if _, exists := m.counters["cache_evictions_total"]; !exists {
			t.Error("Expected evictions counter to be created")
		}
	})

	t.Run("cache size", func(t *testing.T) {
		cache.SetSize("redis", 1024000)
		if _, exists := m.gauges["cache_size_bytes"]; !exists {
			t.Error("Expected cache size gauge to be created")
		}
	})
}

func TestDatabaseMetrics(t *testing.T) {
	m := NewMetrics(&Config{
		ServiceName: "test",
		Namespace:   "test",
	})
	db := m.NewDatabaseMetrics()

	t.Run("query executed", func(t *testing.T) {
		db.QueryExecuted("SELECT", 0.05, true)
		if _, exists := m.histograms["database_query_duration_seconds"]; !exists {
			t.Error("Expected query duration histogram to be created")
		}
		if _, exists := m.counters["database_queries_total"]; !exists {
			t.Error("Expected queries counter to be created")
		}
	})

	t.Run("connection management", func(t *testing.T) {
		db.ConnectionOpened()
		db.ConnectionClosed()
		db.SetConnectionPoolSize(20)

		if _, exists := m.gauges["database_connections_active"]; !exists {
			t.Error("Expected connections gauge to be created")
		}
	})
}

func TestBusinessMetrics(t *testing.T) {
	m := NewMetrics(&Config{
		ServiceName: "test",
		Namespace:   "test",
	})
	business := m.NewBusinessMetrics()

	t.Run("user metrics", func(t *testing.T) {
		business.UserRegistered()
		business.UserLoggedIn()
		business.SetActiveUsers(100)

		if _, exists := m.counters["users_registered_total"]; !exists {
			t.Error("Expected user registrations counter to be created")
		}
	})

	t.Run("match metrics", func(t *testing.T) {
		business.MatchStarted("ranked")
		business.SetActiveMatches(5)
		business.MatchCompleted("ranked", 450.5)

		if _, exists := m.counters["matches_started_total"]; !exists {
			t.Error("Expected matches started counter to be created")
		}
		if _, exists := m.histograms["match_duration_seconds"]; !exists {
			t.Error("Expected match duration histogram to be created")
		}
	})

	t.Run("leaderboard updates", func(t *testing.T) {
		business.LeaderboardUpdated()
		if _, exists := m.counters["leaderboard_updates_total"]; !exists {
			t.Error("Expected leaderboard updates counter to be created")
		}
	})
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.ServiceName != "app" {
		t.Errorf("Expected default service name 'app', got '%s'", config.ServiceName)
	}
	if config.Namespace != "app" {
		t.Errorf("Expected default namespace 'app', got '%s'", config.Namespace)
	}
	if !config.EnableHTTPMetrics {
		t.Error("Expected HTTP metrics to be enabled by default")
	}
	if config.PushInterval != 15*time.Second {
		t.Errorf("Expected default push interval 15s, got %v", config.PushInterval)
	}
	if len(config.HTTPBuckets) == 0 {
		t.Error("Expected default HTTP buckets to be set")
	}
}

func TestRegistry(t *testing.T) {
	m := NewMetrics(&Config{
		ServiceName: "test",
		Namespace:   "test",
	})

	registry := m.Registry()
	if registry == nil {
		t.Error("Expected registry to be returned")
	}
}

func TestHandler(t *testing.T) {
	m := NewMetrics(&Config{
		ServiceName: "test",
		Namespace:   "test",
	})

	handler := m.Handler()
	if handler == nil {
		t.Error("Expected handler to be returned")
	}
}

func TestGetLabelKeys(t *testing.T) {
	t.Run("nil labels", func(t *testing.T) {
		keys := getLabelKeys(nil)
		if len(keys) != 0 {
			t.Errorf("Expected 0 label keys, got %d", len(keys))
		}
	})

	t.Run("with labels", func(t *testing.T) {
		labels := MetricLabels{
			"status": "success",
			"method": "GET",
		}
		keys := getLabelKeys(labels)
		if len(keys) != 2 {
			t.Errorf("Expected 2 label keys, got %d", len(keys))
		}
	})
}

func TestConstLabels(t *testing.T) {
	config := &Config{
		ServiceName: "test",
		Namespace:   "test",
		ConstLabels: prometheus.Labels{
			"environment": "test",
			"region":      "us-east-1",
		},
	}

	m := NewMetrics(config)
	if m.config.ConstLabels["environment"] != "test" {
		t.Error("Expected const labels to be preserved")
	}
}
