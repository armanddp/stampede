package reporter

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"stampede-shooter/internal/metrics"
)

// Reporter handles progress reporting and final results
type Reporter struct {
	collector *metrics.Collector
	startTime time.Time
	verbose   bool
}

// New creates a new reporter
func New(collector *metrics.Collector, verbose bool) *Reporter {
	return &Reporter{
		collector: collector,
		startTime: time.Now(),
		verbose:   verbose,
	}
}

// StartLiveReporting begins showing live progress updates
func (r *Reporter) StartLiveReporting() {
	if !r.verbose {
		return
	}

	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()

		for range ticker.C {
			r.showProgress()
		}
	}()
}

// showProgress displays current test progress
func (r *Reporter) showProgress() {
	stats := r.collector.GetStats()

	totalOK := int64(0)
	totalErr := int64(0)
	currentRPS := float64(0)

	for _, stat := range stats {
		totalOK += stat.TotalOK
		totalErr += stat.TotalErrors
	}

	elapsed := time.Since(r.startTime).Seconds()
	if elapsed > 0 {
		currentRPS = float64(totalOK) / elapsed
	}

	successRate := float64(100)
	if totalOK+totalErr > 0 {
		successRate = float64(totalOK) / float64(totalOK+totalErr) * 100
	}

	fmt.Printf("\rElapsed: %.0fs | Requests: %d | Errors: %d | Success: %.1f%% | RPS: %.1f",
		elapsed, totalOK, totalErr, successRate, currentRPS)
}

// PrintFinalReport displays the final test results
func (r *Reporter) PrintFinalReport() {
	fmt.Println("\n\nFinal Test Results:")
	fmt.Println("==================")

	stats := r.collector.GetStats()
	if len(stats) == 0 {
		fmt.Println("No requests were made.")
		return
	}

	// Sort actions by name for consistent output
	var actionNames []string
	for name := range stats {
		actionNames = append(actionNames, name)
	}
	sort.Strings(actionNames)

	// Print header
	fmt.Printf("%-15s %8s %8s %8s %8s %8s %8s %8s\n",
		"Action", "OK", "ERR", "p50", "p90", "p95", "p99", "RPS")
	fmt.Println(strings.Repeat("─", 88))

	totalOK := int64(0)
	totalErr := int64(0)
	totalBytes := int64(0)
	elapsed := time.Since(r.startTime).Seconds()

	// Print stats for each action
	for _, name := range actionNames {
		stat := stats[name]

		p50 := stat.GetLatencyPercentile(50.0)
		p90 := stat.GetLatencyPercentile(90.0)
		p95 := stat.GetLatencyPercentile(95.0)
		p99 := stat.GetLatencyPercentile(99.0)

		actionRPS := float64(stat.TotalOK) / elapsed

		fmt.Printf("%-15s %8d %8d %8s %8s %8s %8s %8.1f\n",
			truncateString(name, 15),
			stat.TotalOK,
			stat.TotalErrors,
			formatDuration(p50),
			formatDuration(p90),
			formatDuration(p95),
			formatDuration(p99),
			actionRPS)

		totalOK += stat.TotalOK
		totalErr += stat.TotalErrors
		totalBytes += stat.BytesTotal
	}

	// Print totals
	fmt.Println(strings.Repeat("─", 88))

	totalRequests := totalOK + totalErr
	successRate := float64(100)
	if totalRequests > 0 {
		successRate = float64(totalOK) / float64(totalRequests) * 100
	}

	avgRPS := float64(totalOK) / elapsed
	avgLatency := time.Duration(0)

	// Calculate overall average latency
	if len(stats) > 0 {
		totalLatency := int64(0)
		totalCount := int64(0)

		for _, stat := range stats {
			if stat.TotalOK > 0 {
				// Use p50 as approximation for average
				p50Micros := stat.Histogram.ValueAtQuantile(50.0)
				totalLatency += p50Micros * stat.TotalOK
				totalCount += stat.TotalOK
			}
		}

		if totalCount > 0 {
			avgLatency = time.Duration(totalLatency/totalCount) * time.Microsecond
		}
	}

	fmt.Printf("\nTotals: %d requests, %.1f%% success, %.0fs, %.1f rps, avg %s\n",
		totalRequests, successRate, elapsed, avgRPS, formatDuration(avgLatency))

	if totalBytes > 0 {
		mbTransferred := float64(totalBytes) / (1024 * 1024)
		fmt.Printf("Data transferred: %.2f MB (%.2f MB/s)\n",
			mbTransferred, mbTransferred/elapsed)
	}
}

// SaveReport saves the results to a JSON file
func (r *Reporter) SaveReport(filename string) error {
	if filename == "" {
		return nil
	}

	stats := r.collector.GetStats()
	elapsed := time.Since(r.startTime).Seconds()

	// Build report structure
	report := map[string]interface{}{
		"timestamp":    r.startTime.Format(time.RFC3339),
		"duration_sec": elapsed,
		"actions":      make(map[string]interface{}),
	}

	totalOK := int64(0)
	totalErr := int64(0)
	totalBytes := int64(0)

	for name, stat := range stats {
		actionReport := map[string]interface{}{
			"total_ok":     stat.TotalOK,
			"total_errors": stat.TotalErrors,
			"bytes_total":  stat.BytesTotal,
			"p50_ms":       stat.GetLatencyPercentile(50.0).Milliseconds(),
			"p90_ms":       stat.GetLatencyPercentile(90.0).Milliseconds(),
			"p95_ms":       stat.GetLatencyPercentile(95.0).Milliseconds(),
			"p99_ms":       stat.GetLatencyPercentile(99.0).Milliseconds(),
			"rps":          float64(stat.TotalOK) / elapsed,
		}

		report["actions"].(map[string]interface{})[name] = actionReport

		totalOK += stat.TotalOK
		totalErr += stat.TotalErrors
		totalBytes += stat.BytesTotal
	}

	// Add summary
	totalRequests := totalOK + totalErr
	successRate := float64(100)
	if totalRequests > 0 {
		successRate = float64(totalOK) / float64(totalRequests) * 100
	}

	report["summary"] = map[string]interface{}{
		"total_requests": totalRequests,
		"total_ok":       totalOK,
		"total_errors":   totalErr,
		"success_rate":   successRate,
		"avg_rps":        float64(totalOK) / elapsed,
		"bytes_total":    totalBytes,
	}

	// Write to file
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(report); err != nil {
		return fmt.Errorf("failed to write JSON: %w", err)
	}

	fmt.Printf("Results saved to %s\n", filename)
	return nil
}

// formatDuration formats a duration for display
func formatDuration(d time.Duration) string {
	if d < time.Microsecond {
		return "0µs"
	} else if d < time.Millisecond {
		return fmt.Sprintf("%dµs", d.Microseconds())
	} else if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	} else {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
}

// truncateString truncates a string to the specified length
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
