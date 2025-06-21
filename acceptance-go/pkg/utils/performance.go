package utils

import (
	"sync"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// PerformanceMonitor provides utilities for measuring performance in tests
type PerformanceMonitor struct {
	measurements map[string][]time.Duration
	mutex        sync.RWMutex
}

// NewPerformanceMonitor creates a new performance monitor
func NewPerformanceMonitor() *PerformanceMonitor {
	return &PerformanceMonitor{
		measurements: make(map[string][]time.Duration),
	}
}

// StartMeasurement starts a performance measurement and returns a function to end it
func (p *PerformanceMonitor) StartMeasurement(name string) func() time.Duration {
	GinkgoHelper()
	
	startTime := time.Now()
	
	return func() time.Duration {
		duration := time.Since(startTime)
		p.recordMeasurement(name, duration)
		return duration
	}
}

// MeasureOperation measures the time taken to execute an operation
func (p *PerformanceMonitor) MeasureOperation(name string, operation func()) time.Duration {
	GinkgoHelper()
	
	startTime := time.Now()
	operation()
	duration := time.Since(startTime)
	
	p.recordMeasurement(name, duration)
	return duration
}

// MeasureOperationWithReturn measures the time taken to execute an operation that returns a value
func (p *PerformanceMonitor) MeasureOperationWithReturn(name string, operation func() interface{}) (interface{}, time.Duration) {
	GinkgoHelper()
	
	startTime := time.Now()
	result := operation()
	duration := time.Since(startTime)
	
	p.recordMeasurement(name, duration)
	return result, duration
}

// recordMeasurement records a measurement
func (p *PerformanceMonitor) recordMeasurement(name string, duration time.Duration) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	
	if p.measurements[name] == nil {
		p.measurements[name] = make([]time.Duration, 0)
	}
	
	p.measurements[name] = append(p.measurements[name], duration)
}

// GetMeasurements returns all measurements for a given name
func (p *PerformanceMonitor) GetMeasurements(name string) []time.Duration {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	
	measurements := p.measurements[name]
	if measurements == nil {
		return []time.Duration{}
	}
	
	// Return a copy to avoid data races
	result := make([]time.Duration, len(measurements))
	copy(result, measurements)
	return result
}

// GetAverageDuration returns the average duration for a given measurement name
func (p *PerformanceMonitor) GetAverageDuration(name string) time.Duration {
	measurements := p.GetMeasurements(name)
	if len(measurements) == 0 {
		return 0
	}
	
	var total time.Duration
	for _, duration := range measurements {
		total += duration
	}
	
	return total / time.Duration(len(measurements))
}

// GetMaxDuration returns the maximum duration for a given measurement name
func (p *PerformanceMonitor) GetMaxDuration(name string) time.Duration {
	measurements := p.GetMeasurements(name)
	if len(measurements) == 0 {
		return 0
	}
	
	var maxDuration time.Duration
	for _, duration := range measurements {
		if duration > maxDuration {
			maxDuration = duration
		}
	}
	
	return maxDuration
}

// GetMinDuration returns the minimum duration for a given measurement name
func (p *PerformanceMonitor) GetMinDuration(name string) time.Duration {
	measurements := p.GetMeasurements(name)
	if len(measurements) == 0 {
		return 0
	}
	
	minDuration := measurements[0]
	for _, duration := range measurements[1:] {
		if duration < minDuration {
			minDuration = duration
		}
	}
	
	return minDuration
}

// AssertPerformance asserts that a measurement meets performance requirements
func (p *PerformanceMonitor) AssertPerformance(name string, maxDuration time.Duration) {
	GinkgoHelper()
	
	measurements := p.GetMeasurements(name)
	Expect(len(measurements)).To(BeNumerically(">", 0), 
		"No measurements found for: %s", name)
	
	latestMeasurement := measurements[len(measurements)-1]
	Expect(latestMeasurement).To(BeNumerically("<=", maxDuration),
		"Performance requirement failed for %s: %v > %v", name, latestMeasurement, maxDuration)
}

// AssertAveragePerformance asserts that the average performance meets requirements
func (p *PerformanceMonitor) AssertAveragePerformance(name string, maxAverageDuration time.Duration) {
	GinkgoHelper()
	
	avgDuration := p.GetAverageDuration(name)
	Expect(avgDuration).To(BeNumerically(">", 0),
		"No measurements found for: %s", name)
	
	Expect(avgDuration).To(BeNumerically("<=", maxAverageDuration),
		"Average performance requirement failed for %s: %v > %v", name, avgDuration, maxAverageDuration)
}

// PrintStatistics prints performance statistics for all measurements
func (p *PerformanceMonitor) PrintStatistics() {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	
	if len(p.measurements) == 0 {
		GinkgoWriter.Printf("No performance measurements recorded\n")
		return
	}
	
	GinkgoWriter.Printf("\n=== Performance Statistics ===\n")
	for name, measurements := range p.measurements {
		if len(measurements) == 0 {
			continue
		}
		
		var total time.Duration
		var min, max time.Duration
		min = measurements[0]
		max = measurements[0]
		
		for _, duration := range measurements {
			total += duration
			if duration < min {
				min = duration
			}
			if duration > max {
				max = duration
			}
		}
		
		avg := total / time.Duration(len(measurements))
		
		GinkgoWriter.Printf("%-30s: count=%d, avg=%v, min=%v, max=%v\n",
			name, len(measurements), avg, min, max)
	}
	GinkgoWriter.Printf("================================\n\n")
}

// Clear clears all measurements
func (p *PerformanceMonitor) Clear() {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	
	p.measurements = make(map[string][]time.Duration)
}

// Global performance monitor instance
var globalPerformanceMonitor = NewPerformanceMonitor()

// StartMeasurement starts a performance measurement using the global monitor
func StartMeasurement(name string) func() time.Duration {
	return globalPerformanceMonitor.StartMeasurement(name)
}

// MeasureOperation measures an operation using the global monitor
func MeasureOperation(name string, operation func()) time.Duration {
	return globalPerformanceMonitor.MeasureOperation(name, operation)
}

// AssertPerformance asserts performance using the global monitor
func AssertPerformance(name string, maxDuration time.Duration) {
	globalPerformanceMonitor.AssertPerformance(name, maxDuration)
}

// PrintGlobalStatistics prints statistics from the global monitor
func PrintGlobalStatistics() {
	globalPerformanceMonitor.PrintStatistics()
}