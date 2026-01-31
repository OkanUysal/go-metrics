# Go Metrics

A powerful, production-ready metrics library for Go with Prometheus integration, automatic HTTP metrics collection, and specialized helpers for WebSocket, cache, database, and business metrics. Includes comprehensive Railway and Grafana Cloud setup guides.

## Features

- **Prometheus Integration**: Industry-standard metrics format
- **Auto HTTP Metrics**: Automatic collection via Gin middleware
- **Custom Metrics**: Counters, Gauges, Histograms with labels
- **Specialized Helpers**: WebSocket, Cache, Database, Business metrics
- **Railway Ready**: Easy deployment and monitoring setup
- **Grafana Cloud**: Free tier integration with visual dashboards
- **Production Ready**: Thread-safe, performant, low overhead

## Installation

```bash
go get github.com/OkanUysal/go-metrics
```

## Quick Start

### Basic Setup

```go
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/OkanUysal/go-metrics"
)

func main() {
    // Create metrics instance
    m := metrics.NewMetrics(&metrics.Config{
        ServiceName: "your-app",
        Namespace:   "myapp",
    })
    
    r := gin.Default()
    
    // Enable automatic HTTP metrics
    r.Use(m.Middleware())
    
    // Expose metrics endpoint
    r.GET("/metrics", gin.WrapH(m.Handler()))
    
    // Your API routes
    r.GET("/api/users", func(c *gin.Context) {
        c.JSON(200, gin.H{"status": "ok"})
    })
    
    r.Run(":8080")
}
```

Now visit `http://localhost:8080/metrics` to see Prometheus metrics!

## HTTP Metrics (Automatic)

When you use `m.Middleware()`, these metrics are automatically collected:

```
# Request count by method, path, and status
myapp_http_requests_total{method="GET",path="/api/users",status="200"} 150

# Request duration (histogram with p50, p95, p99)
myapp_http_request_duration_seconds_bucket{method="GET",path="/api/users",status="200",le="0.05"} 145
myapp_http_request_duration_seconds_sum{method="GET",path="/api/users",status="200"} 4.2
myapp_http_request_duration_seconds_count{method="GET",path="/api/users",status="200"} 150

# Request/response sizes
myapp_http_request_size_bytes{method="POST",path="/api/users"} 1024
myapp_http_response_size_bytes{method="GET",path="/api/users"} 512

# In-flight requests
myapp_http_requests_in_flight 5
```

## Custom Metrics

### Counters

```go
// Simple counter
m.IncrementCounter("user_registrations_total", nil)

// Counter with labels
m.IncrementCounter("api_calls_total", metrics.MetricLabels{
    "service": "auth",
    "method":  "login",
})

// Increment by specific value
m.IncrementCounterBy("bytes_processed_total", 1024, nil)
```

### Gauges

```go
// Set gauge value
m.SetGauge("active_users", 1523, nil)

// Increment/decrement
m.IncrementGauge("cache_size", nil)
m.DecrementGauge("queue_length", nil)

// With labels
m.SetGauge("server_temperature", 45.3, metrics.MetricLabels{
    "location": "datacenter1",
})
```

### Histograms

```go
// Record observation (for durations, sizes, etc.)
m.RecordHistogram("query_duration_seconds", 0.042, metrics.MetricLabels{
    "query_type": "select",
})

m.RecordHistogram("file_size_bytes", 2048576, metrics.MetricLabels{
    "file_type": "image",
})
```

## WebSocket Metrics

```go
import (
    "github.com/OkanUysal/go-metrics"
    "github.com/OkanUysal/go-websocket"
)

ws := m.NewWebSocketMetrics()

// Track connections
func onConnect(client *websocket.Client) {
    ws.ConnectionOpened()
}

func onDisconnect(client *websocket.Client) {
    ws.ConnectionClosed()
}

// Track messages
ws.MessageSent("chat_message")
ws.MessageReceived("game_event")

// Track rooms
ws.RoomCreated("match_room")
ws.SetActiveRooms(42)
ws.SetRoomClients("room_123", 5)
ws.RoomClosed("match_room")
```

**Metrics generated:**
```
websocket_connections_active 156
websocket_connections_total 2341
websocket_messages_sent_total{type="chat_message"} 5234
websocket_rooms_active 42
websocket_room_clients{room_id="room_123"} 5
```

## Cache Metrics

```go
cache := m.NewCacheMetrics()

// Track hits/misses
cache.Hit("redis")
cache.Miss("redis")

// Calculate and set hit ratio
hitRatio := float64(hits) / float64(hits + misses)
cache.SetHitRatio("redis", hitRatio)

// Track evictions
cache.Eviction("memory")

// Track cache size
cache.SetSize("redis", 1024000) // bytes
```

**Metrics generated:**
```
cache_hits_total{type="redis"} 8523
cache_misses_total{type="redis"} 234
cache_hit_ratio{type="redis"} 0.973
cache_size_bytes{type="redis"} 1024000
```

## Database Metrics

```go
db := m.NewDatabaseMetrics()

// Track queries
start := time.Now()
err := database.Query()
duration := time.Since(start).Seconds()

db.QueryExecuted("SELECT", duration, err == nil)

// Track connections
db.ConnectionOpened()
db.ConnectionClosed()
db.SetConnectionPoolSize(20)
```

**Metrics generated:**
```
database_queries_total{operation="SELECT",status="success"} 15234
database_query_duration_seconds{operation="SELECT",status="success"} 0.042
database_connections_active 15
database_connection_pool_size 20
```

## Business Metrics

```go
business := m.NewBusinessMetrics()

// User metrics
business.UserRegistered()
business.UserLoggedIn()
business.SetActiveUsers(1523)

// Match metrics
business.MatchStarted("ranked")
business.SetActiveMatches(42)

// When match ends
duration := 450.5 // seconds
business.MatchCompleted("ranked", duration)

// Leaderboard updates
business.LeaderboardUpdated()
```

**Metrics generated:**
```
users_registered_total 5234
users_logged_in_total 12456
users_active 1523
matches_started_total{type="ranked"} 3421
matches_active 42
matches_completed_total{type="ranked"} 3398
match_duration_seconds{type="ranked"} 450.5
leaderboard_updates_total 523
```

## Custom Configuration

```go
config := &metrics.Config{
    ServiceName:       "your-app",
    Namespace:         "myapp",
    Subsystem:         "backend", // optional
    EnableHTTPMetrics: true,
    
    // Custom HTTP histogram buckets (seconds)
    HTTPBuckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1, 5},
    
    // Add constant labels to all metrics
    ConstLabels: prometheus.Labels{
        "environment": "production",
        "region":      "us-east-1",
    },
}

m := metrics.NewMetrics(config)
```

## Skip Metrics for Specific Endpoints

```go
// Don't collect metrics for health check or metrics endpoint
r.Use(m.MiddlewareWithSkipper(func(c *gin.Context) bool {
    return c.Request.URL.Path == "/health" || 
           c.Request.URL.Path == "/metrics"
}))
```

## Railway Deployment Setup

### 1. Deploy Your App to Railway

```bash
# In your app repository
railway init
railway up
```

### 2. Expose Metrics Endpoint

Make sure your app has:
```go
r.GET("/metrics", gin.WrapH(metrics.Handler()))
```

### 3. View Metrics in Railway

```
https://your-app.up.railway.app/metrics
```

You'll see Prometheus-format metrics in text.

### 4. Railway Built-in Monitoring

Railway automatically:
- ✅ Shows CPU/Memory usage
- ✅ Displays request logs
- ✅ Tracks deployment health

Your custom metrics from `/metrics` endpoint can be:
- Viewed directly in browser
- Scraped by external monitoring (Grafana, Datadog)
- Used for custom alerts

## Grafana Cloud Setup (Free Tier)

### Step 1: Create Grafana Cloud Account

1. Go to [grafana.com](https://grafana.com)
2. Sign up for **Free Forever** plan (10K series, 50GB logs)
3. Create a new stack

### Step 2: Get Prometheus Remote Write Credentials

1. In Grafana Cloud → **Connections** → **Add Integration**
2. Select **Prometheus**
3. Copy these values:
   ```
   Remote Write URL: https://prometheus-xxx.grafana.net/api/prom/push
   Username: 123456
   Password: glc_xxx (API Key)
   ```

### Step 3: Install Prometheus in Railway

Create `prometheus.yml`:
```yaml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'your-app'
    static_configs:
      - targets: ['your-app.railway.internal:8080']

remote_write:
  - url: https://prometheus-xxx.grafana.net/api/prom/push
    basic_auth:
      username: 123456
      password: glc_xxx
```

Deploy Prometheus to Railway:
```bash
# Create new service in Railway
railway service create prometheus

# Deploy Prometheus image
railway service deploy --image prom/prometheus:latest
```

### Step 4: Configure Environment Variables

In Railway, set these for your app:
```bash
GRAFANA_CLOUD_URL=https://prometheus-xxx.grafana.net/api/prom/push
GRAFANA_CLOUD_USER=123456
GRAFANA_CLOUD_KEY=glc_xxx
```

### Step 5: View Dashboards in Grafana

1. Go to your Grafana Cloud dashboard
2. **Explore** → Select your Prometheus datasource
3. Query your metrics:
   ```promql
   rate(myapp_http_requests_total[5m])
   histogram_quantile(0.95, rate(myapp_http_request_duration_seconds_bucket[5m]))
   ```

### Step 6: Import Pre-built Dashboards

1. **Dashboards** → **New** → **Import**
2. Use dashboard ID `1860` (Node Exporter Full)
3. Or create custom dashboard for your metrics

### Example Grafana Queries

```promql
# Request rate (requests per second)
rate(myapp_http_requests_total[5m])

# 95th percentile response time
histogram_quantile(0.95, rate(myapp_http_request_duration_seconds_bucket[5m]))

# Error rate
rate(myapp_http_requests_total{status=~"5.."}[5m])

# Active WebSocket connections
websocket_connections_active

# Cache hit ratio
cache_hit_ratio

# Active matches
matches_active
```

## Alternative: Simple Railway + Grafana Setup

If you don't want to run Prometheus:

### Option A: Grafana Agent (Lightweight)

```yaml
# grafana-agent.yaml
metrics:
  configs:
    - name: myapp
      scrape_configs:
        - job_name: your-app
          static_configs:
            - targets: ['localhost:8080']
      remote_write:
        - url: https://prometheus-xxx.grafana.net/api/prom/push
          basic_auth:
            username: 123456
            password: glc_xxx
```

Deploy alongside your app in Railway.

### Option B: Direct Push (No scraping needed)

Use Prometheus Pushgateway:
```go
// In your app
import "github.com/prometheus/client_golang/prometheus/push"

pusher := push.New("http://pushgateway:9091", "your-app").
    Collector(m.Registry())

// Push every 15 seconds
ticker := time.NewTicker(15 * time.Second)
go func() {
    for range ticker.C {
        if err := pusher.Push(); err != nil {
            log.Printf("Push error: %v", err)
        }
    }
}()
```

## Complete Example

```go
package main

import (
    "time"
    
    "github.com/gin-gonic/gin"
    "github.com/OkanUysal/go-metrics"
    "github.com/OkanUysal/go-websocket"
)

func main() {
    // Initialize metrics
    m := metrics.NewMetrics(&metrics.Config{
        ServiceName: "your-app",
        Namespace:   "myapp",
        ConstLabels: prometheus.Labels{
            "environment": "production",
        },
    })
    
    // Create helpers
    wsMetrics := m.NewWebSocketMetrics()
    cacheMetrics := m.NewCacheMetrics()
    dbMetrics := m.NewDatabaseMetrics()
    business := m.NewBusinessMetrics()
    
    // Setup Gin
    r := gin.Default()
    r.Use(m.Middleware())
    
    // Metrics endpoint
    r.GET("/metrics", gin.WrapH(m.Handler()))
    
    // WebSocket setup
    hub := websocket.NewHub(nil)
    hub.SetOnConnect(func(client *websocket.Client) {
        wsMetrics.ConnectionOpened()
        business.SetActiveUsers(float64(hub.GetOnlineCount()))
    })
    hub.SetOnDisconnect(func(client *websocket.Client) {
        wsMetrics.ConnectionClosed()
        business.SetActiveUsers(float64(hub.GetOnlineCount()))
    })
    go hub.Run()
    
    // API Routes
    r.POST("/api/auth/register", func(c *gin.Context) {
        // Register user...
        business.UserRegistered()
        c.JSON(200, gin.H{"status": "registered"})
    })
    
    r.POST("/api/matches", func(c *gin.Context) {
        // Start match...
        business.MatchStarted("ranked")
        business.SetActiveMatches(float64(getActiveMatchCount()))
        
        // Track match duration
        start := time.Now()
        // ... match logic ...
        duration := time.Since(start).Seconds()
        business.MatchCompleted("ranked", duration)
        
        c.JSON(200, gin.H{"match_id": "12345"})
    })
    
    r.Run(":8080")
}
```

## Metric Naming Best Practices

```go
// ✅ Good naming
"http_requests_total"          // Clear, includes unit (_total)
"database_query_duration_seconds"  // Includes unit (_seconds)
"cache_size_bytes"             // Includes unit (_bytes)
"users_active"                 // Clear gauge name

// ❌ Avoid
"requests"                     // Too vague
"db_time"                      // No unit
"CacheSize"                    // Use snake_case
```

## Performance

- **Overhead**: < 1ms per request
- **Memory**: ~100KB for 1000 unique metrics
- **CPU**: Negligible (concurrent-safe operations)

## Requirements

- Go 1.21 or higher
- [prometheus/client_golang](https://github.com/prometheus/client_golang) v1.19.0+

## Optional Dependencies

- [gin-gonic/gin](https://github.com/gin-gonic/gin) - For HTTP middleware

## Free Monitoring Options

1. **Railway Built-in**: Basic CPU/Memory/Logs
2. **Grafana Cloud Free**: 10K series, 50GB logs, 50GB traces
3. **Prometheus (self-hosted)**: Unlimited, free
4. **Alternative**: Datadog (14-day trial), New Relic (100GB/month free)

## License

MIT License

## Author

Okan Uysal - [GitHub](https://github.com/OkanUysal)
