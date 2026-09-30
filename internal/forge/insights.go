package forge

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/unbound-force/replicator/internal/db"
)

// GetStrategyInsights queries the events table for historical success rates by strategy.
func GetStrategyInsights(store *db.Store, task string) (result map[string]any, err error) {
	rows, err := store.DB.Query(
		"SELECT payload FROM events WHERE type = 'forge_outcome'",
	)
	if err != nil {
		return nil, fmt.Errorf("query outcomes: %w", err)
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			result = nil
			err = errors.Join(err, fmt.Errorf("close outcome rows: %w", closeErr))
		}
	}()

	type strategyStats struct {
		Total   int `json:"total"`
		Success int `json:"success"`
	}
	stats := map[string]*strategyStats{}

	for rows.Next() {
		var payloadStr string
		if err := rows.Scan(&payloadStr); err != nil {
			continue
		}
		var payload map[string]any
		if err := json.Unmarshal([]byte(payloadStr), &payload); err != nil {
			continue
		}

		strategy, _ := payload["strategy"].(string)
		if strategy == "" {
			strategy = "unknown"
		}

		if _, ok := stats[strategy]; !ok {
			stats[strategy] = &strategyStats{}
		}
		stats[strategy].Total++

		if success, ok := payload["success"].(bool); ok && success {
			stats[strategy].Success++
		}
	}
	if rowsErr := rows.Err(); rowsErr != nil {
		return nil, fmt.Errorf("iterate outcomes: %w", rowsErr)
	}

	// Calculate success rates.
	rates := map[string]any{}
	for strategy, s := range stats {
		rate := 0.0
		if s.Total > 0 {
			rate = float64(s.Success) / float64(s.Total) * 100
		}
		rates[strategy] = map[string]any{
			"total":        s.Total,
			"success":      s.Success,
			"success_rate": rate,
		}
	}

	recommendation := "feature-based" // Default recommendation.
	bestRate := 0.0
	for strategy, s := range stats {
		if s.Total >= 2 { // Need at least 2 data points.
			rate := float64(s.Success) / float64(s.Total)
			if rate > bestRate {
				bestRate = rate
				recommendation = strategy
			}
		}
	}

	return map[string]any{
		"task":           task,
		"strategies":     rates,
		"recommendation": recommendation,
	}, nil
}

// GetFileInsights queries the events table for file-specific gotchas.
func GetFileInsights(store *db.Store, files []string) (result map[string]any, err error) {
	if len(files) == 0 {
		return map[string]any{
			"files":    files,
			"insights": map[string]any{},
		}, nil
	}

	rows, err := store.DB.Query(
		"SELECT payload FROM events WHERE type = 'forge_outcome'",
	)
	if err != nil {
		return nil, fmt.Errorf("query outcomes: %w", err)
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			result = nil
			err = errors.Join(err, fmt.Errorf("close outcome rows: %w", closeErr))
		}
	}()

	// Build a set of target files for fast lookup.
	fileSet := map[string]bool{}
	for _, f := range files {
		fileSet[f] = true
	}

	type fileStats struct {
		Total      int `json:"total"`
		Failures   int `json:"failures"`
		ErrorCount int `json:"error_count"`
	}
	insights := map[string]*fileStats{}

	for rows.Next() {
		var payloadStr string
		if err := rows.Scan(&payloadStr); err != nil {
			continue
		}
		var payload map[string]any
		if err := json.Unmarshal([]byte(payloadStr), &payload); err != nil {
			continue
		}

		touchedFiles, _ := payload["files_touched"].([]any)
		success, _ := payload["success"].(bool)
		errorCount := 0
		if ec, ok := payload["error_count"].(float64); ok {
			errorCount = int(ec)
		}

		for _, tf := range touchedFiles {
			fStr, ok := tf.(string)
			if !ok {
				continue
			}
			if !fileSet[fStr] {
				continue
			}
			if _, ok := insights[fStr]; !ok {
				insights[fStr] = &fileStats{}
			}
			insights[fStr].Total++
			if !success {
				insights[fStr].Failures++
			}
			insights[fStr].ErrorCount += errorCount
		}
	}
	if rowsErr := rows.Err(); rowsErr != nil {
		return nil, fmt.Errorf("iterate outcomes: %w", rowsErr)
	}

	return map[string]any{
		"files":    files,
		"insights": insights,
	}, nil
}

// GetPatternInsights queries the events table for the top 5 most frequent failure patterns.
func GetPatternInsights(store *db.Store) (result map[string]any, err error) {
	rows, err := store.DB.Query(
		"SELECT payload FROM events WHERE type = 'forge_outcome'",
	)
	if err != nil {
		return nil, fmt.Errorf("query outcomes: %w", err)
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			result = nil
			err = errors.Join(err, fmt.Errorf("close outcome rows: %w", closeErr))
		}
	}()

	patternCounts := map[string]int{}
	totalOutcomes := 0
	totalFailures := 0

	for rows.Next() {
		var payloadStr string
		if err := rows.Scan(&payloadStr); err != nil {
			continue
		}
		var payload map[string]any
		if err := json.Unmarshal([]byte(payloadStr), &payload); err != nil {
			continue
		}

		totalOutcomes++
		success, _ := payload["success"].(bool)
		if !success {
			totalFailures++

			// Extract failure criteria as patterns.
			if criteria, ok := payload["criteria"].([]any); ok {
				for _, c := range criteria {
					if cStr, ok := c.(string); ok {
						patternCounts[cStr]++
					}
				}
			}

			// Count error types.
			if ec, ok := payload["error_count"].(float64); ok && ec > 0 {
				patternCounts["errors_encountered"]++
			}
			if rc, ok := payload["retry_count"].(float64); ok && rc > 0 {
				patternCounts["required_retries"]++
			}
		}
	}
	if rowsErr := rows.Err(); rowsErr != nil {
		return nil, fmt.Errorf("iterate outcomes: %w", rowsErr)
	}

	// Sort by count and take top 5.
	type patternEntry struct {
		Pattern string `json:"pattern"`
		Count   int    `json:"count"`
	}
	var patterns []patternEntry
	for p, c := range patternCounts {
		patterns = append(patterns, patternEntry{Pattern: p, Count: c})
	}

	// Simple sort -- top 5 by count.
	for i := 0; i < len(patterns); i++ {
		for j := i + 1; j < len(patterns); j++ {
			if patterns[j].Count > patterns[i].Count {
				patterns[i], patterns[j] = patterns[j], patterns[i]
			}
		}
	}
	if len(patterns) > 5 {
		patterns = patterns[:5]
	}

	return map[string]any{
		"total_outcomes": totalOutcomes,
		"total_failures": totalFailures,
		"top_patterns":   patterns,
	}, nil
}
