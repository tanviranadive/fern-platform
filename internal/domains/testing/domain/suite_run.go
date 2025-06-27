package domain

import (
	"errors"
	"time"
)

// SuiteRunStatus represents the status of a suite run
type SuiteRunStatus string

const (
	SuiteRunStatusRunning SuiteRunStatus = "running"
	SuiteRunStatusPassed  SuiteRunStatus = "passed"
	SuiteRunStatusFailed  SuiteRunStatus = "failed"
	SuiteRunStatusError   SuiteRunStatus = "error"
)

// SuiteRun represents a test suite execution
type SuiteRun struct {
	suiteName     string
	status        SuiteRunStatus
	startTime     time.Time
	endTime       *time.Time
	totalSpecs    int
	passedSpecs   int
	failedSpecs   int
	skippedSpecs  int
	specs         []SpecRun
}

// NewSuiteRun creates a new suite run
func NewSuiteRun(suiteName string) (*SuiteRun, error) {
	if suiteName == "" {
		return nil, errors.New("suite name cannot be empty")
	}

	return &SuiteRun{
		suiteName: suiteName,
		status:    SuiteRunStatusRunning,
		startTime: time.Now(),
		specs:     make([]SpecRun, 0),
	}, nil
}

// SuiteName returns the suite name
func (sr *SuiteRun) SuiteName() string {
	return sr.suiteName
}

// Status returns the current status
func (sr *SuiteRun) Status() SuiteRunStatus {
	return sr.status
}

// TotalSpecs returns the total number of specs
func (sr *SuiteRun) TotalSpecs() int {
	return sr.totalSpecs
}

// PassedSpecs returns the number of passed specs
func (sr *SuiteRun) PassedSpecs() int {
	return sr.passedSpecs
}

// FailedSpecs returns the number of failed specs
func (sr *SuiteRun) FailedSpecs() int {
	return sr.failedSpecs
}

// SkippedSpecs returns the number of skipped specs
func (sr *SuiteRun) SkippedSpecs() int {
	return sr.skippedSpecs
}

// AddSpec adds a spec run to the suite
func (sr *SuiteRun) AddSpec(spec SpecRun) error {
	if sr.status != SuiteRunStatusRunning {
		return errors.New("cannot add spec to completed suite")
	}

	sr.specs = append(sr.specs, spec)
	sr.totalSpecs++

	// Update counts based on spec status
	switch spec.Status() {
	case SpecStatusPassed:
		sr.passedSpecs++
	case SpecStatusFailed:
		sr.failedSpecs++
	case SpecStatusSkipped:
		sr.skippedSpecs++
	}

	return nil
}

// Complete marks the suite run as completed
func (sr *SuiteRun) Complete() error {
	if sr.status != SuiteRunStatusRunning {
		return errors.New("can only complete a running suite")
	}

	now := time.Now()
	sr.endTime = &now

	// Determine final status
	if sr.failedSpecs > 0 {
		sr.status = SuiteRunStatusFailed
	} else if sr.totalSpecs == sr.passedSpecs+sr.skippedSpecs {
		sr.status = SuiteRunStatusPassed
	} else {
		sr.status = SuiteRunStatusError
	}

	return nil
}

// Duration returns the duration in milliseconds
func (sr *SuiteRun) Duration() int64 {
	if sr.endTime == nil {
		return int64(time.Since(sr.startTime).Milliseconds())
	}
	return int64(sr.endTime.Sub(sr.startTime).Milliseconds())
}

// GetSpecs returns all spec runs
func (sr *SuiteRun) GetSpecs() []SpecRun {
	// Return a copy to prevent external modification
	specs := make([]SpecRun, len(sr.specs))
	copy(specs, sr.specs)
	return specs
}