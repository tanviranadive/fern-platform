package domain

import (
	"errors"
	"time"
)

// SpecStatus represents the status of a spec run
type SpecStatus string

const (
	SpecStatusRunning SpecStatus = "running"
	SpecStatusPassed  SpecStatus = "passed"
	SpecStatusFailed  SpecStatus = "failed"
	SpecStatusSkipped SpecStatus = "skipped"
	SpecStatusError   SpecStatus = "error"
)

// SpecRun represents an individual test spec execution
type SpecRun struct {
	specName     string
	status       SpecStatus
	startTime    time.Time
	endTime      *time.Time
	errorMessage string
	stackTrace   string
	retryCount   int
	isFlaky      bool
}

// NewSpecRun creates a new spec run
func NewSpecRun(specName string) (*SpecRun, error) {
	if specName == "" {
		return nil, errors.New("spec name cannot be empty")
	}

	return &SpecRun{
		specName:  specName,
		status:    SpecStatusRunning,
		startTime: time.Now(),
	}, nil
}

// SpecName returns the spec name
func (sr *SpecRun) SpecName() string {
	return sr.specName
}

// Status returns the current status
func (sr *SpecRun) Status() SpecStatus {
	return sr.status
}

// Pass marks the spec as passed
func (sr *SpecRun) Pass() error {
	if sr.status != SpecStatusRunning {
		return errors.New("can only pass a running spec")
	}

	now := time.Now()
	sr.endTime = &now
	sr.status = SpecStatusPassed
	return nil
}

// Fail marks the spec as failed with an error
func (sr *SpecRun) Fail(errorMessage, stackTrace string) error {
	if sr.status != SpecStatusRunning {
		return errors.New("can only fail a running spec")
	}

	now := time.Now()
	sr.endTime = &now
	sr.status = SpecStatusFailed
	sr.errorMessage = errorMessage
	sr.stackTrace = stackTrace
	return nil
}

// Skip marks the spec as skipped
func (sr *SpecRun) Skip() error {
	if sr.status != SpecStatusRunning {
		return errors.New("can only skip a running spec")
	}

	now := time.Now()
	sr.endTime = &now
	sr.status = SpecStatusSkipped
	return nil
}

// MarkFlaky marks the spec as flaky
func (sr *SpecRun) MarkFlaky() {
	sr.isFlaky = true
}

// IncrementRetry increments the retry count
func (sr *SpecRun) IncrementRetry() {
	sr.retryCount++
}

// Duration returns the duration in milliseconds
func (sr *SpecRun) Duration() int64 {
	if sr.endTime == nil {
		return int64(time.Since(sr.startTime).Milliseconds())
	}
	return int64(sr.endTime.Sub(sr.startTime).Milliseconds())
}

// ErrorMessage returns the error message if failed
func (sr *SpecRun) ErrorMessage() string {
	return sr.errorMessage
}

// StackTrace returns the stack trace if failed
func (sr *SpecRun) StackTrace() string {
	return sr.stackTrace
}

// IsFlaky returns whether the spec is marked as flaky
func (sr *SpecRun) IsFlaky() bool {
	return sr.isFlaky
}

// RetryCount returns the number of retries
func (sr *SpecRun) RetryCount() int {
	return sr.retryCount
}