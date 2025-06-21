package utils

import (
	"context"
	"fmt"
	"os"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// WaitForCondition waits for a condition to be true with timeout and polling
func WaitForCondition(condition func() bool, timeout time.Duration, message ...string) {
	GinkgoHelper()
	
	msg := "Condition should become true"
	if len(message) > 0 {
		msg = message[0]
	}
	
	Eventually(func() bool {
		return condition()
	}, timeout, 1*time.Second).Should(BeTrue(), msg)
}

// WaitForConditionWithContext waits for a condition with context support
func WaitForConditionWithContext(ctx context.Context, condition func(context.Context) bool, timeout time.Duration, message ...string) {
	GinkgoHelper()
	
	msg := "Condition should become true"
	if len(message) > 0 {
		msg = message[0]
	}
	
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	
	Eventually(func() bool {
		return condition(timeoutCtx)
	}, timeout, 1*time.Second).Should(BeTrue(), msg)
}

// WaitForConditionWithError waits for a condition that returns an error
func WaitForConditionWithError(conditionFunc func() error, timeout time.Duration, message ...string) {
	GinkgoHelper()
	
	msg := "Condition should succeed without error"
	if len(message) > 0 {
		msg = message[0]
	}
	
	Eventually(func() error {
		return conditionFunc()
	}, timeout, 1*time.Second).Should(Succeed(), msg)
}

// RetryOperation retries an operation with exponential backoff
func RetryOperation(operation func() error, maxRetries int, initialDelay time.Duration) error {
	GinkgoHelper()
	
	var lastErr error
	delay := initialDelay
	
	for i := 0; i < maxRetries; i++ {
		lastErr = operation()
		if lastErr == nil {
			return nil
		}
		
		if i < maxRetries-1 {
			By(fmt.Sprintf("Operation failed (attempt %d/%d), retrying in %v: %v", 
				i+1, maxRetries, delay, lastErr))
			time.Sleep(delay)
			delay *= 2 // Exponential backoff
		}
	}
	
	return fmt.Errorf("operation failed after %d attempts: %w", maxRetries, lastErr)
}

// GenerateUniqueID generates a unique identifier for test resources
func GenerateUniqueID(prefix string) string {
	timestamp := time.Now().UnixNano() / int64(time.Millisecond)
	randomSeed := GinkgoRandomSeed()
	parallelProcess := GinkgoParallelProcess()
	
	return fmt.Sprintf("%s-%d-%d-%d", prefix, timestamp, randomSeed, parallelProcess)
}

// SafeStringPtr safely converts a string to a pointer
func SafeStringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// SafeTimePtr safely converts time to pointer
func SafeTimePtr(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	return &t
}

// StringPtrValue safely gets value from string pointer
func StringPtrValue(ptr *string) string {
	if ptr == nil {
		return ""
	}
	return *ptr
}

// TimePtrValue safely gets value from time pointer
func TimePtrValue(ptr *time.Time) time.Time {
	if ptr == nil {
		return time.Time{}
	}
	return *ptr
}

// Contains checks if a slice contains a specific element
func Contains[T comparable](slice []T, element T) bool {
	for _, item := range slice {
		if item == element {
			return true
		}
	}
	return false
}

// Filter filters a slice based on a predicate function
func Filter[T any](slice []T, predicate func(T) bool) []T {
	var result []T
	for _, item := range slice {
		if predicate(item) {
			result = append(result, item)
		}
	}
	return result
}

// Map transforms a slice using a mapping function
func Map[T any, R any](slice []T, mapFunc func(T) R) []R {
	result := make([]R, len(slice))
	for i, item := range slice {
		result[i] = mapFunc(item)
	}
	return result
}

// Find finds the first element in a slice that matches a predicate
func Find[T any](slice []T, predicate func(T) bool) (*T, bool) {
	for _, item := range slice {
		if predicate(item) {
			return &item, true
		}
	}
	return nil, false
}

// GetEnvOrDefault gets an environment variable or returns a default value
func GetEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// ParseDuration safely parses a duration string
func ParseDuration(s string, defaultDuration time.Duration) time.Duration {
	if s == "" {
		return defaultDuration
	}
	
	duration, err := time.ParseDuration(s)
	if err != nil {
		By(fmt.Sprintf("Failed to parse duration '%s', using default %v", s, defaultDuration))
		return defaultDuration
	}
	
	return duration
}

// FormatDuration formats a duration in a human-readable way
func FormatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Nanoseconds()/int64(time.Millisecond))
	} else if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	} else if d < time.Hour {
		return fmt.Sprintf("%.1fm", d.Minutes())
	} else {
		return fmt.Sprintf("%.1fh", d.Hours())
	}
}

// MustNotError is a helper that fails the test if an error is present
func MustNotError(err error, message ...string) {
	GinkgoHelper()
	
	msg := "Operation should not fail"
	if len(message) > 0 {
		msg = message[0]
	}
	
	Expect(err).NotTo(HaveOccurred(), msg)
}

// MustSucceed is a helper that fails the test if a condition is false
func MustSucceed(condition bool, message ...string) {
	GinkgoHelper()
	
	msg := "Condition should be true"
	if len(message) > 0 {
		msg = message[0]
	}
	
	Expect(condition).To(BeTrue(), msg)
}

// LogWithTimestamp logs a message with timestamp for debugging
func LogWithTimestamp(message string, args ...interface{}) {
	timestamp := time.Now().Format("15:04:05.000")
	fullMessage := fmt.Sprintf("[%s] %s", timestamp, message)
	By(fmt.Sprintf(fullMessage, args...))
}

// Benchmark measures the execution time of a function
func Benchmark(name string, fn func()) time.Duration {
	GinkgoHelper()
	
	LogWithTimestamp("Starting benchmark: %s", name)
	start := time.Now()
	fn()
	duration := time.Since(start)
	LogWithTimestamp("Completed benchmark: %s in %s", name, FormatDuration(duration))
	
	return duration
}

// Sleep sleeps for a duration while logging the reason
func Sleep(duration time.Duration, reason string) {
	GinkgoHelper()
	
	LogWithTimestamp("Sleeping for %s: %s", FormatDuration(duration), reason)
	time.Sleep(duration)
}

// ValidateStringNotEmpty validates that a string is not empty
func ValidateStringNotEmpty(value, fieldName string) {
	GinkgoHelper()
	Expect(value).NotTo(BeEmpty(), "%s should not be empty", fieldName)
}

// ValidatePositiveNumber validates that a number is positive
func ValidatePositiveNumber(value int, fieldName string) {
	GinkgoHelper()
	Expect(value).To(BeNumerically(">", 0), "%s should be positive", fieldName)
}

// ValidateTimeNotZero validates that a time is not zero
func ValidateTimeNotZero(value time.Time, fieldName string) {
	GinkgoHelper()
	Expect(value).NotTo(BeZero(), "%s should not be zero time", fieldName)
}

// ValidateSliceNotEmpty validates that a slice is not empty
func ValidateSliceNotEmpty[T any](slice []T, fieldName string) {
	GinkgoHelper()
	Expect(len(slice)).To(BeNumerically(">", 0), "%s should not be empty", fieldName)
}

// CreateTestContext creates a context for tests with timeout
func CreateTestContext(timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), timeout)
}

// WaitGroup helps manage concurrent operations in tests
type WaitGroup struct {
	operations []func()
	errors     []error
}

// NewWaitGroup creates a new wait group for test operations
func NewWaitGroup() *WaitGroup {
	return &WaitGroup{
		operations: make([]func(), 0),
		errors:     make([]error, 0),
	}
}

// Add adds an operation to the wait group
func (wg *WaitGroup) Add(operation func()) {
	wg.operations = append(wg.operations, operation)
}

// AddWithError adds an operation that returns an error
func (wg *WaitGroup) AddWithError(operation func() error) {
	wg.operations = append(wg.operations, func() {
		if err := operation(); err != nil {
			wg.errors = append(wg.errors, err)
		}
	})
}

// Wait executes all operations concurrently and waits for completion
func (wg *WaitGroup) Wait() error {
	GinkgoHelper()
	
	if len(wg.operations) == 0 {
		return nil
	}
	
	done := make(chan bool, len(wg.operations))
	
	for _, operation := range wg.operations {
		go func(op func()) {
			defer GinkgoRecover()
			op()
			done <- true
		}(operation)
	}
	
	// Wait for all operations to complete
	for i := 0; i < len(wg.operations); i++ {
		<-done
	}
	
	if len(wg.errors) > 0 {
		return fmt.Errorf("wait group had %d errors: %v", len(wg.errors), wg.errors)
	}
	
	return nil
}