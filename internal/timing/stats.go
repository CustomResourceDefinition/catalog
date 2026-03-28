package timing

import (
	"context"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
)

type Category string

const (
	CategoryHTTP       Category = "http"
	CategoryGit        Category = "git"
	CategoryHelm       Category = "helm"
	CategoryOCI        Category = "oci"
	CategoryGeneration Category = "generation"
	CategoryMisc       Category = "misc"
)

type OperationType string

const (
	OperationTypeFetch     OperationType = "fetch"
	OperationTypeGenerate  OperationType = "generate"
	OperationTypeDownload  OperationType = "download"
	OperationTypeRender    OperationType = "render"
	OperationTypeAPIFetch  OperationType = "api_fetch"
	OperationTypeClone     OperationType = "clone"
	OperationTypeCheckout  OperationType = "checkout"
	OperationTypeSubmodule OperationType = "submodule"
	OperationTypeTags      OperationType = "tags"
	OperationTypeBranches  OperationType = "branches"
	OperationTypeWrite     OperationType = "write"
	OperationTypeStatus    OperationType = "status"
	OperationTypeUpdate    OperationType = "update"
	OperationTypePull      OperationType = "pull"
	OperationTypeAuth      OperationType = "auth"
	OperationTypeIndex     OperationType = "index"
)

type Operation struct {
	Category  Category
	Type      OperationType
	Name      string
	Duration  time.Duration
	Success   bool
	StartTime time.Time
}

type categoryStats struct {
	operations []Operation
	mu         sync.RWMutex
}

type Stats struct {
	categories map[Category]*categoryStats
	totalTime  time.Duration
	totalOps   int
	mu         sync.RWMutex
	logFile    *os.File
	logMu      sync.Mutex
}

var contextKey = struct{}{}

func NewStats() *Stats {
	return &Stats{
		categories: make(map[Category]*categoryStats),
	}
}

func FromContext(ctx context.Context) (*Stats, bool) {
	s, ok := ctx.Value(contextKey).(*Stats)
	return s, ok
}

func WithContext(ctx context.Context, stats *Stats) context.Context {
	return context.WithValue(ctx, contextKey, stats)
}

func (s *Stats) OpenLogFile(path string) error {
	if path == "" {
		return nil
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	s.logFile = f
	return nil
}

func (s *Stats) CloseLogFile() {
	if s.logFile != nil {
		s.logFile.Close()
		s.logFile = nil
	}
}

func (s *Stats) Record(category Category, opType OperationType, name string, duration time.Duration, success bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.totalTime += duration
	s.totalOps++

	cat, ok := s.categories[category]
	if !ok {
		cat = &categoryStats{}
		s.categories[category] = cat
	}

	cat.mu.Lock()
	defer cat.mu.Unlock()

	cat.operations = append(cat.operations, Operation{
		Category:  category,
		Type:      opType,
		Name:      name,
		Duration:  duration,
		Success:   success,
		StartTime: time.Now().Add(-duration),
	})

	if s.logFile != nil {
		successStr := "success"
		if !success {
			successStr = "failure"
		}
		s.logMu.Lock()
		defer s.logMu.Unlock()
		fmt.Fprintf(s.logFile, "%s %s %s %s %s (%s)\n",
			time.Now().Format(time.RFC3339),
			category,
			opType,
			name,
			formatDuration(duration),
			successStr,
		)
	}
}

func (s *Stats) RecordFunc(category Category, opType OperationType, name string, success bool, fn func() error) error {
	start := time.Now()
	err := fn()
	duration := time.Since(start)
	s.Record(category, opType, name, duration, err == nil)
	return err
}

func (s *Stats) GetCategoryStats(category Category) []Operation {
	s.mu.RLock()
	defer s.mu.RUnlock()

	cat, ok := s.categories[category]
	if !ok {
		return nil
	}

	cat.mu.RLock()
	defer cat.mu.RUnlock()

	result := make([]Operation, len(cat.operations))
	copy(result, cat.operations)
	return result
}

func (s *Stats) TotalTime() time.Duration {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.totalTime
}

func (s *Stats) TotalOperations() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.totalOps
}

func (s *Stats) RecordDuration(category Category, opType OperationType, name string) func() {
	start := time.Now()
	return func() {
		s.Record(category, opType, name, time.Since(start), true)
	}
}

func calculatePercentiles(durations []float64, ps []float64) map[float64]time.Duration {
	if len(durations) == 0 {
		return nil
	}

	sorted := make([]float64, len(durations))
	copy(sorted, durations)
	sort.Float64s(sorted)

	result := make(map[float64]time.Duration)
	for _, p := range ps {
		idx := int(float64(len(sorted)) * p)
		if idx >= len(sorted) {
			idx = len(sorted) - 1
		}
		if idx < 0 {
			idx = 0
		}
		result[p] = time.Duration(sorted[idx] * float64(time.Second))
	}
	return result
}

func formatDuration(d time.Duration) string {
	if d < time.Millisecond {
		return fmt.Sprintf("%.0fµs", float64(d.Microseconds()))
	}
	if d < time.Second {
		return fmt.Sprintf("%.0fms", float64(d.Milliseconds()))
	}
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	if d < time.Hour {
		m := d.Minutes()
		if m < 10 {
			return fmt.Sprintf("%.1fm", m)
		}
		return fmt.Sprintf("%.0fm", m)
	}
	h := d.Hours()
	return fmt.Sprintf("%.1fh", h)
}

func (s *Stats) PrintSummary(writer io.Writer) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	fmt.Fprintf(writer, "\n=== Update Statistics ===\n\n")
	fmt.Fprintf(writer, "Overall: %s (%d operations)\n\n", formatDuration(s.totalTime), s.totalOps)

	categoryOrder := []Category{CategoryHTTP, CategoryGit, CategoryHelm, CategoryOCI, CategoryGeneration, CategoryMisc}

	for _, cat := range categoryOrder {
		catStats, ok := s.categories[cat]
		if !ok {
			continue
		}

		catStats.mu.RLock()
		ops := catStats.operations
		catStats.mu.RUnlock()

		if len(ops) == 0 {
			continue
		}

		typeStats := make(map[OperationType][]time.Duration)

		for _, op := range ops {
			typeStats[op.Type] = append(typeStats[op.Type], op.Duration)
		}

		fmt.Fprintf(writer, "%s:\n", strings.ToUpper(string(cat)[:1])+string(cat)[1:])

		for opType, durations := range typeStats {
			if len(durations) == 0 {
				continue
			}

			var total time.Duration
			durationsSecs := make([]float64, len(durations))
			for i, d := range durations {
				total += d
				durationsSecs[i] = d.Seconds()
			}

			percs := calculatePercentiles(durationsSecs, []float64{0.75, 0.90, 0.95})

			typeLabel := string(opType)
			fmt.Fprintf(writer, "  %-10s %4d operations  total: %s",
				typeLabel+":",
				len(durations),
				formatDuration(total),
			)

			if len(percs) > 0 {
				fmt.Fprintf(writer, "  p75: %s  p90: %s  p95: %s",
					formatDuration(percs[0.75]),
					formatDuration(percs[0.90]),
					formatDuration(percs[0.95]),
				)
			}
			fmt.Fprintln(writer)
		}
		fmt.Fprintln(writer)
	}
}
