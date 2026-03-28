package timing

import (
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
	CategoryGeneration Category = "generation"
	CategoryGit        Category = "git"
	CategoryHelm       Category = "helm"
	CategoryHTTP       Category = "http"
	CategoryMisc       Category = "misc"
	CategoryOCI        Category = "oci"
)

type OperationType string

const (
	OperationTypeAPIFetch  OperationType = "api_fetch"
	OperationTypeAuth      OperationType = "auth"
	OperationTypeBranches  OperationType = "branches"
	OperationTypeCheckout  OperationType = "checkout"
	OperationTypeClone     OperationType = "clone"
	OperationTypeDownload  OperationType = "download"
	OperationTypeFetch     OperationType = "fetch"
	OperationTypeGC        OperationType = "gc"
	OperationTypeGenerate  OperationType = "generate"
	OperationTypeIndex     OperationType = "index"
	OperationTypePull      OperationType = "pull"
	OperationTypeRender    OperationType = "render"
	OperationTypeStatus    OperationType = "status"
	OperationTypeSubmodule OperationType = "submodule"
	OperationTypeTags      OperationType = "tags"
	OperationTypeUpdate    OperationType = "update"
	OperationTypeWrite     OperationType = "write"
)

type Operation struct {
	Category  Category
	Type      OperationType
	Name      string
	Duration  time.Duration
	Success   bool
	StartTime time.Time
}

type Stats struct {
	categories map[Category][]Operation
	totalTime  time.Duration
	totalOps   int
	mu         sync.RWMutex
	logFile    *os.File
	logMu      sync.Mutex
}

func NewStats() *Stats {
	return &Stats{
		categories: make(map[Category][]Operation),
	}
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
	s.Flush()
	if s.logFile != nil {
		s.logFile.Close()
		s.logFile = nil
	}
}

func (s *Stats) Flush() {
	if s.logFile == nil {
		return
	}

	s.mu.RLock()
	var allOps []Operation
	for _, ops := range s.categories {
		allOps = append(allOps, ops...)
	}
	s.mu.RUnlock()

	sort.Slice(allOps, func(i, j int) bool {
		return allOps[i].StartTime.Before(allOps[j].StartTime)
	})

	s.logMu.Lock()
	defer s.logMu.Unlock()
	for _, op := range allOps {
		successStr := "success"
		if !op.Success {
			successStr = "failure"
		}
		fmt.Fprintf(s.logFile, "%s %s %s %s %s (%s)\n",
			op.StartTime.Format(time.RFC3339Nano),
			op.Category,
			op.Type,
			op.Name,
			formatDuration(op.Duration),
			successStr,
		)
	}
}

func (s *Stats) Record(category Category, opType OperationType, name string, duration time.Duration, success bool, startTime time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.totalTime += duration
	s.totalOps++

	s.categories[category] = append(s.categories[category], Operation{
		Category:  category,
		Type:      opType,
		Name:      name,
		Duration:  duration,
		Success:   success,
		StartTime: startTime,
	})
}

func (s *Stats) GetCategoryStats(category Category) []Operation {
	s.mu.RLock()
	defer s.mu.RUnlock()

	ops, ok := s.categories[category]
	if !ok {
		return nil
	}

	result := make([]Operation, len(ops))
	copy(result, ops)
	return result
}

func (s *Stats) GetAllStats() []Operation {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var allOps []Operation
	for _, ops := range s.categories {
		allOps = append(allOps, ops...)
	}
	return allOps
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
		ops, ok := s.categories[cat]
		if !ok || len(ops) == 0 {
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
