package timing

import (
	"context"
	"fmt"
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

func findWorst(durations []float64) (int, time.Duration) {
	if len(durations) == 0 {
		return -1, 0
	}

	maxIdx := 0
	maxVal := durations[0]
	for i, v := range durations {
		if v > maxVal {
			maxVal = v
			maxIdx = i
		}
	}
	return maxIdx, time.Duration(maxVal * float64(time.Second))
}

func sanitizeName(name string) string {
	name = strings.ReplaceAll(name, "/", "_")
	name = strings.ReplaceAll(name, ":", "_")
	name = strings.ReplaceAll(name, ".", "_")
	name = strings.ReplaceAll(name, "-", "_")
	name = strings.ReplaceAll(name, " ", "_")
	return name
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

func (s *Stats) PrintSummary() {
	s.mu.RLock()
	defer s.mu.RUnlock()

	fmt.Printf("\n=== Update Statistics ===\n\n")
	fmt.Printf("Overall: %s (%d operations)\n\n", formatDuration(s.totalTime), s.totalOps)

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

		fmt.Printf("%s:\n", strings.ToUpper(string(cat)[:1])+string(cat)[1:])

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
			fmt.Printf("  %-10s %4d operations  total: %s",
				typeLabel+":",
				len(durations),
				formatDuration(total),
			)

			if len(percs) > 0 {
				fmt.Printf("  p75: %s  p90: %s  p95: %s",
					formatDuration(percs[0.75]),
					formatDuration(percs[0.90]),
					formatDuration(percs[0.95]),
				)
			}
			fmt.Println()

			worstIdx, worstDur := findWorst(durationsSecs)
			if worstIdx >= 0 {
				for _, op := range ops {
					if op.Type == opType && op.Duration == worstDur {
						fmt.Printf("  Worst:   %s (%s)\n", op.Name, formatDuration(worstDur))
						break
					}
				}
			}
		}
		fmt.Println()
	}
}
