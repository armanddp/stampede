package metrics

import (
	"sync"
	"time"

	"github.com/HdrHistogram/hdrhistogram-go"
)

// RequestMetric represents a single HTTP request measurement
type RequestMetric struct {
	Name       string
	Method     string
	URL        string
	StartTime  time.Time
	EndTime    time.Time
	StatusCode int
	BytesRead  int64
	Error      string
}

// ActionStats holds aggregated statistics for a specific action
type ActionStats struct {
	Name        string
	TotalOK     int64
	TotalErrors int64
	Histogram   *hdrhistogram.Histogram
	BytesTotal  int64
	mu          sync.RWMutex
}

// Collector aggregates metrics from multiple workers
type Collector struct {
	metrics   chan RequestMetric
	actions   map[string]*ActionStats
	startTime time.Time
	mu        sync.RWMutex
	done      chan struct{}
}

// NewCollector creates a new metrics collector
func NewCollector() *Collector {
	return &Collector{
		metrics:   make(chan RequestMetric, 10000),
		actions:   make(map[string]*ActionStats),
		startTime: time.Now(),
		done:      make(chan struct{}),
	}
}

// Record sends a metric to the collector
func (c *Collector) Record(metric RequestMetric) {
	select {
	case c.metrics <- metric:
	default:
		// Drop metric if channel is full to avoid blocking workers
	}
}

// Start begins collecting metrics in a goroutine
func (c *Collector) Start() {
	go c.collect()
}

// Stop stops the collector and closes channels
func (c *Collector) Stop() {
	close(c.metrics)
	<-c.done
}

// GetStats returns current aggregated statistics
func (c *Collector) GetStats() map[string]*ActionStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make(map[string]*ActionStats)
	for name, stats := range c.actions {
		result[name] = stats
	}
	return result
}

// collect processes incoming metrics
func (c *Collector) collect() {
	defer close(c.done)

	for metric := range c.metrics {
		c.mu.Lock()

		// Get or create action stats
		stats, exists := c.actions[metric.Name]
		if !exists {
			hist := hdrhistogram.New(1, 60000000, 3) // 1µs to 60s, 3 significant digits
			stats = &ActionStats{
				Name:      metric.Name,
				Histogram: hist,
			}
			c.actions[metric.Name] = stats
		}

		// Update stats
		stats.mu.Lock()
		latencyMicros := metric.EndTime.Sub(metric.StartTime).Microseconds()

		if metric.Error == "" && metric.StatusCode >= 200 && metric.StatusCode < 400 {
			stats.TotalOK++
			stats.Histogram.RecordValue(latencyMicros)
		} else {
			stats.TotalErrors++
		}

		stats.BytesTotal += metric.BytesRead
		stats.mu.Unlock()

		c.mu.Unlock()
	}
}

// GetLatencyPercentile returns the specified percentile from the histogram
func (as *ActionStats) GetLatencyPercentile(percentile float64) time.Duration {
	as.mu.RLock()
	defer as.mu.RUnlock()

	micros := as.Histogram.ValueAtQuantile(percentile)
	return time.Duration(micros) * time.Microsecond
}
