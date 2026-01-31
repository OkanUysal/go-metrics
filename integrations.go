package metrics

// WebSocketMetrics provides WebSocket-specific metrics helpers
type WebSocketMetrics struct {
	m *Metrics
}

// NewWebSocketMetrics creates WebSocket metrics helper
func (m *Metrics) NewWebSocketMetrics() *WebSocketMetrics {
	return &WebSocketMetrics{m: m}
}

// ConnectionOpened increments active WebSocket connections
func (ws *WebSocketMetrics) ConnectionOpened() {
	ws.m.IncrementGauge("websocket_connections_active", nil)
	ws.m.IncrementCounter("websocket_connections_total", nil)
}

// ConnectionClosed decrements active WebSocket connections
func (ws *WebSocketMetrics) ConnectionClosed() {
	ws.m.DecrementGauge("websocket_connections_active", nil)
}

// MessageSent increments sent messages counter
func (ws *WebSocketMetrics) MessageSent(messageType string) {
	ws.m.IncrementCounter("websocket_messages_sent_total", MetricLabels{
		"type": messageType,
	})
}

// MessageReceived increments received messages counter
func (ws *WebSocketMetrics) MessageReceived(messageType string) {
	ws.m.IncrementCounter("websocket_messages_received_total", MetricLabels{
		"type": messageType,
	})
}

// RoomCreated increments room creation counter
func (ws *WebSocketMetrics) RoomCreated(roomType string) {
	ws.m.IncrementCounter("websocket_rooms_created_total", MetricLabels{
		"type": roomType,
	})
}

// RoomClosed increments room closure counter
func (ws *WebSocketMetrics) RoomClosed(roomType string) {
	ws.m.IncrementCounter("websocket_rooms_closed_total", MetricLabels{
		"type": roomType,
	})
}

// SetActiveRooms sets the active rooms gauge
func (ws *WebSocketMetrics) SetActiveRooms(count float64) {
	ws.m.SetGauge("websocket_rooms_active", count, nil)
}

// SetRoomClients sets the number of clients in a specific room
func (ws *WebSocketMetrics) SetRoomClients(roomID string, count float64) {
	ws.m.SetGauge("websocket_room_clients", count, MetricLabels{
		"room_id": roomID,
	})
}

// CacheMetrics provides cache-specific metrics helpers
type CacheMetrics struct {
	m *Metrics
}

// NewCacheMetrics creates cache metrics helper
func (m *Metrics) NewCacheMetrics() *CacheMetrics {
	return &CacheMetrics{m: m}
}

// Hit increments cache hit counter
func (cm *CacheMetrics) Hit(cacheType string) {
	cm.m.IncrementCounter("cache_hits_total", MetricLabels{
		"type": cacheType,
	})
}

// Miss increments cache miss counter
func (cm *CacheMetrics) Miss(cacheType string) {
	cm.m.IncrementCounter("cache_misses_total", MetricLabels{
		"type": cacheType,
	})
}

// SetHitRatio sets the cache hit ratio gauge
func (cm *CacheMetrics) SetHitRatio(cacheType string, ratio float64) {
	cm.m.SetGauge("cache_hit_ratio", ratio, MetricLabels{
		"type": cacheType,
	})
}

// Eviction increments cache eviction counter
func (cm *CacheMetrics) Eviction(cacheType string) {
	cm.m.IncrementCounter("cache_evictions_total", MetricLabels{
		"type": cacheType,
	})
}

// SetSize sets the cache size gauge
func (cm *CacheMetrics) SetSize(cacheType string, size float64) {
	cm.m.SetGauge("cache_size_bytes", size, MetricLabels{
		"type": cacheType,
	})
}

// DatabaseMetrics provides database-specific metrics helpers
type DatabaseMetrics struct {
	m *Metrics
}

// NewDatabaseMetrics creates database metrics helper
func (m *Metrics) NewDatabaseMetrics() *DatabaseMetrics {
	return &DatabaseMetrics{m: m}
}

// QueryExecuted records a database query execution
func (dm *DatabaseMetrics) QueryExecuted(operation string, duration float64, success bool) {
	status := "success"
	if !success {
		status = "error"
	}

	dm.m.RecordHistogram("database_query_duration_seconds", duration, MetricLabels{
		"operation": operation,
		"status":    status,
	})

	dm.m.IncrementCounter("database_queries_total", MetricLabels{
		"operation": operation,
		"status":    status,
	})
}

// ConnectionOpened increments active database connections
func (dm *DatabaseMetrics) ConnectionOpened() {
	dm.m.IncrementGauge("database_connections_active", nil)
}

// ConnectionClosed decrements active database connections
func (dm *DatabaseMetrics) ConnectionClosed() {
	dm.m.DecrementGauge("database_connections_active", nil)
}

// SetConnectionPoolSize sets the connection pool size
func (dm *DatabaseMetrics) SetConnectionPoolSize(size float64) {
	dm.m.SetGauge("database_connection_pool_size", size, nil)
}

// BusinessMetrics provides business-specific metrics helpers
type BusinessMetrics struct {
	m *Metrics
}

// NewBusinessMetrics creates business metrics helper
func (m *Metrics) NewBusinessMetrics() *BusinessMetrics {
	return &BusinessMetrics{m: m}
}

// UserRegistered increments user registration counter
func (bm *BusinessMetrics) UserRegistered() {
	bm.m.IncrementCounter("users_registered_total", nil)
}

// UserLoggedIn increments user login counter
func (bm *BusinessMetrics) UserLoggedIn() {
	bm.m.IncrementCounter("users_logged_in_total", nil)
}

// SetActiveUsers sets the active users gauge
func (bm *BusinessMetrics) SetActiveUsers(count float64) {
	bm.m.SetGauge("users_active", count, nil)
}

// MatchStarted increments match started counter
func (bm *BusinessMetrics) MatchStarted(matchType string) {
	bm.m.IncrementCounter("matches_started_total", MetricLabels{
		"type": matchType,
	})
}

// MatchCompleted records match completion with duration
func (bm *BusinessMetrics) MatchCompleted(matchType string, duration float64) {
	bm.m.IncrementCounter("matches_completed_total", MetricLabels{
		"type": matchType,
	})

	bm.m.RecordHistogram("match_duration_seconds", duration, MetricLabels{
		"type": matchType,
	})
}

// SetActiveMatches sets the active matches gauge
func (bm *BusinessMetrics) SetActiveMatches(count float64) {
	bm.m.SetGauge("matches_active", count, nil)
}

// LeaderboardUpdated increments leaderboard update counter
func (bm *BusinessMetrics) LeaderboardUpdated() {
	bm.m.IncrementCounter("leaderboard_updates_total", nil)
}
