package timing

import (
	"io"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewStats(t *testing.T) {
	s := NewStats()
	assert.NotNil(t, s)
	assert.Zero(t, s.TotalTime())
	assert.Zero(t, s.TotalOperations())
}

func TestRecord(t *testing.T) {
	s := NewStats()

	s.Record(CategoryHTTP, OperationTypeFetch, "test-op", 100*time.Millisecond, true, time.Now())

	assert.Equal(t, 1, s.TotalOperations())
	assert.Equal(t, 100*time.Millisecond, s.TotalTime())

	ops := s.GetCategoryStats(CategoryHTTP)
	assert.Len(t, ops, 1)
	assert.Equal(t, OperationTypeFetch, ops[0].Type)
	assert.Equal(t, "test-op", ops[0].Name)
	assert.Equal(t, 100*time.Millisecond, ops[0].Duration)
	assert.True(t, ops[0].Success)
}

func TestRecordMultipleCategories(t *testing.T) {
	s := NewStats()

	s.Record(CategoryHTTP, OperationTypeFetch, "http-op", 50*time.Millisecond, true, time.Now())
	s.Record(CategoryGit, OperationTypeClone, "git-op", 100*time.Millisecond, true, time.Now())
	s.Record(CategoryHelm, OperationTypeDownload, "helm-op", 75*time.Millisecond, true, time.Now())

	assert.Equal(t, 3, s.TotalOperations())
	assert.Equal(t, 225*time.Millisecond, s.TotalTime())

	assert.Len(t, s.GetCategoryStats(CategoryHTTP), 1)
	assert.Len(t, s.GetCategoryStats(CategoryGit), 1)
	assert.Len(t, s.GetCategoryStats(CategoryHelm), 1)
}

func TestRecordMultipleOperationsSameCategory(t *testing.T) {
	s := NewStats()

	s.Record(CategoryHTTP, OperationTypeFetch, "op1", 10*time.Millisecond, true, time.Now())
	s.Record(CategoryHTTP, OperationTypeFetch, "op2", 20*time.Millisecond, true, time.Now())
	s.Record(CategoryHTTP, OperationTypeFetch, "op3", 30*time.Millisecond, true, time.Now())

	ops := s.GetCategoryStats(CategoryHTTP)
	assert.Len(t, ops, 3)
}

func TestRecordFailure(t *testing.T) {
	s := NewStats()

	s.Record(CategoryHTTP, OperationTypeFetch, "failed-op", 100*time.Millisecond, false, time.Now())

	ops := s.GetCategoryStats(CategoryHTTP)
	assert.Len(t, ops, 1)
	assert.False(t, ops[0].Success)
}

func TestGetCategoryStatsEmpty(t *testing.T) {
	s := NewStats()

	ops := s.GetCategoryStats(CategoryHTTP)
	assert.Nil(t, ops)
}

func TestGetCategoryStatsUnknownCategory(t *testing.T) {
	s := NewStats()
	s.Record(CategoryHTTP, OperationTypeFetch, "test", 100*time.Millisecond, true, time.Now())

	ops := s.GetCategoryStats(CategoryOCI)
	assert.Nil(t, ops)
}

func TestCalculatePercentiles(t *testing.T) {
	durations := []float64{10, 20, 30, 40, 50, 60, 70, 80, 90, 100}

	result := calculatePercentiles(durations, []float64{0.75, 0.90, 0.95})

	assert.NotNil(t, result)
	assert.Equal(t, time.Duration(77.5*float64(time.Second)), result[0.75])
	assert.Equal(t, time.Duration(91*float64(time.Second)), result[0.90])
	assert.InDelta(t, time.Duration(95.5*float64(time.Second)), result[0.95], float64(1*time.Millisecond))
}

func TestCalculatePercentilesEmpty(t *testing.T) {
	result := calculatePercentiles([]float64{}, []float64{0.75})
	assert.Nil(t, result)
}

func TestCalculatePercentilesSingle(t *testing.T) {
	result := calculatePercentiles([]float64{50}, []float64{0.75, 0.90, 0.95})

	assert.NotNil(t, result)
	assert.Equal(t, 50*time.Second, result[0.75])
	assert.Equal(t, 50*time.Second, result[0.90])
	assert.Equal(t, 50*time.Second, result[0.95])
}

func TestCalculatePercentilesComprehensive(t *testing.T) {
	durations := make([]float64, 100)
	for i := range 100 {
		durations[i] = float64(i + 1)
	}

	result := calculatePercentiles(durations, []float64{0.01, 0.25, 0.50, 0.75, 0.95, 0.99})

	assert.NotNil(t, result)
	assert.Equal(t, time.Duration(1.99*float64(time.Second)), result[0.01])
	assert.Equal(t, time.Duration(25.75*float64(time.Second)), result[0.25])
	assert.Equal(t, time.Duration(50.5*float64(time.Second)), result[0.50])
	assert.Equal(t, time.Duration(75.25*float64(time.Second)), result[0.75])
	assert.Equal(t, time.Duration(95.05*float64(time.Second)), result[0.95])
	assert.Equal(t, time.Duration(99.01*float64(time.Second)), result[0.99])
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		input    time.Duration
		expected string
	}{
		{100 * time.Microsecond, "100µs"},
		{500 * time.Microsecond, "500µs"},
		{100 * time.Millisecond, "100ms"},
		{500 * time.Millisecond, "500ms"},
		{1500 * time.Millisecond, "1.5s"},
		{10 * time.Second, "10.0s"},
		{90 * time.Second, "1.5m"},
		{5 * time.Minute, "5.0m"},
		{10 * time.Minute, "10m"},
		{90 * time.Minute, "1.5h"},
		{2 * time.Hour, "2.0h"},
	}

	for _, tt := range tests {
		result := formatDuration(tt.input)
		assert.Equal(t, tt.expected, result, "input: %v", tt.input)
	}
}

func TestPrintSummaryEmpty(t *testing.T) {
	s := NewStats()

	s.PrintSummary(io.Discard)
}

func TestPrintSummaryWithData(t *testing.T) {
	s := NewStats()

	s.Record(CategoryHTTP, OperationTypeFetch, "op1", 10*time.Millisecond, true, time.Now())
	s.Record(CategoryHTTP, OperationTypeFetch, "op2", 20*time.Millisecond, true, time.Now())
	s.Record(CategoryHTTP, OperationTypeFetch, "op3", 30*time.Millisecond, true, time.Now())
	s.Record(CategoryGit, OperationTypeClone, "repo1", 100*time.Millisecond, true, time.Now())
	s.Record(CategoryGit, OperationTypeClone, "repo2", 200*time.Millisecond, true, time.Now())

	s.PrintSummary(io.Discard)
}

func TestPrintSummaryAllCategories(t *testing.T) {
	s := NewStats()

	s.Record(CategoryHTTP, OperationTypeFetch, "http-op", 10*time.Millisecond, true, time.Now())
	s.Record(CategoryGit, OperationTypeClone, "git-op", 20*time.Millisecond, true, time.Now())
	s.Record(CategoryHelm, OperationTypeDownload, "helm-op", 30*time.Millisecond, true, time.Now())
	s.Record(CategoryOCI, OperationTypePull, "oci-op", 40*time.Millisecond, true, time.Now())
	s.Record(CategoryGeneration, OperationTypeWrite, "gen-op", 50*time.Millisecond, true, time.Now())
	s.Record(CategoryMisc, OperationTypeUpdate, "misc-op", 60*time.Millisecond, true, time.Now())

	s.PrintSummary(io.Discard)
}

func TestConcurrentRecording(t *testing.T) {
	s := NewStats()

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			s.Record(CategoryHTTP, OperationTypeFetch, "op", time.Duration(n)*time.Millisecond, true, time.Now())
		}(i)
	}
	wg.Wait()

	assert.Equal(t, 100, s.TotalOperations())
	assert.Len(t, s.GetCategoryStats(CategoryHTTP), 100)
}

func TestOpenLogFileEmpty(t *testing.T) {
	s := NewStats()

	err := s.OpenLogFile("")
	assert.NoError(t, err)
	assert.Nil(t, s.logFile)
}

func TestOpenLogFile(t *testing.T) {
	s := NewStats()

	tmpFile := t.TempDir() + "/perf.log"
	err := s.OpenLogFile(tmpFile)
	assert.NoError(t, err)
	assert.NotNil(t, s.logFile)

	s.CloseLogFile()

	_, err = os.Stat(tmpFile)
	assert.NoError(t, err)
}

func TestOpenLogFileThenRecord(t *testing.T) {
	s := NewStats()

	tmpFile := t.TempDir() + "/perf.log"
	err := s.OpenLogFile(tmpFile)
	assert.NoError(t, err)

	s.Record(CategoryHTTP, OperationTypeFetch, "test-op", 100*time.Millisecond, true, time.Now())
	s.CloseLogFile()

	content, err := os.ReadFile(tmpFile)
	assert.NoError(t, err)
	assert.Contains(t, string(content), "http")
	assert.Contains(t, string(content), "fetch")
	assert.Contains(t, string(content), "test-op")
}

func TestOpenLogFileInvalidPath(t *testing.T) {
	s := NewStats()

	err := s.OpenLogFile("/invalid/path/that/does/not/exist/perf.log")
	assert.Error(t, err)
}

func TestPrintSummaryWriter(t *testing.T) {
	s := NewStats()

	s.Record(CategoryHTTP, OperationTypeFetch, "op1", 10*time.Millisecond, true, time.Now())

	var buf strings.Builder
	s.PrintSummary(&buf)

	assert.Contains(t, buf.String(), "Update Statistics")
	assert.Contains(t, buf.String(), "**Overall:**")
	assert.Contains(t, buf.String(), "#### Http")
}
