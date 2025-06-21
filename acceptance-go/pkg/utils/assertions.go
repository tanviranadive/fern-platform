package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/onsi/gomega/types"
)

// Custom Gomega matchers for Fern Platform acceptance tests

// HaveValidApiResponse checks if a response has a valid API response structure
func HaveValidApiResponse() types.GomegaMatcher {
	return &validApiResponseMatcher{}
}

type validApiResponseMatcher struct{}

func (m *validApiResponseMatcher) Match(actual interface{}) (success bool, err error) {
	switch v := actual.(type) {
	case *http.Response:
		return v.StatusCode >= 200 && v.StatusCode < 300, nil
	case map[string]interface{}:
		// For GraphQL responses
		_, hasData := v["data"]
		errors, hasErrors := v["errors"]
		
		if hasErrors {
			if errorSlice, ok := errors.([]interface{}); ok && len(errorSlice) > 0 {
				return false, nil
			}
		}
		
		return hasData, nil
	default:
		return false, fmt.Errorf("expected http.Response or map[string]interface{}, got %T", actual)
	}
}

func (m *validApiResponseMatcher) FailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Expected %v to be a valid API response", actual)
}

func (m *validApiResponseMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Expected %v not to be a valid API response", actual)
}

// HaveValidTestRunStructure checks if an object has a valid test run structure
func HaveValidTestRunStructure() types.GomegaMatcher {
	return &validTestRunStructureMatcher{}
}

type validTestRunStructureMatcher struct{}

func (m *validTestRunStructureMatcher) Match(actual interface{}) (success bool, err error) {
	v, ok := actual.(map[string]interface{})
	if !ok {
		return false, fmt.Errorf("expected map[string]interface{}, got %T", actual)
	}
	
	requiredFields := []string{"id", "projectId", "status", "startTime", "duration"}
	
	for _, field := range requiredFields {
		if _, exists := v[field]; !exists {
			return false, nil
		}
	}
	
	// Validate status is one of the expected values
	if status, ok := v["status"].(string); ok {
		validStatuses := []string{"passed", "failed", "skipped", "PASSED", "FAILED", "SKIPPED"}
		statusValid := false
		for _, validStatus := range validStatuses {
			if status == validStatus {
				statusValid = true
				break
			}
		}
		if !statusValid {
			return false, nil
		}
	}
	
	return true, nil
}

func (m *validTestRunStructureMatcher) FailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Expected %v to have valid test run structure", actual)
}

func (m *validTestRunStructureMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Expected %v not to have valid test run structure", actual)
}

// BeWithinTimeRange checks if a duration is within a specified range
func BeWithinTimeRange(min, max time.Duration) types.GomegaMatcher {
	return &timeRangeMatcher{min: min, max: max}
}

type timeRangeMatcher struct {
	min, max time.Duration
}

func (m *timeRangeMatcher) Match(actual interface{}) (success bool, err error) {
	var duration time.Duration
	
	switch v := actual.(type) {
	case time.Duration:
		duration = v
	case int64:
		duration = time.Duration(v) * time.Millisecond
	case float64:
		duration = time.Duration(v) * time.Millisecond
	default:
		return false, fmt.Errorf("expected time.Duration, int64, or float64, got %T", actual)
	}
	
	return duration >= m.min && duration <= m.max, nil
}

func (m *timeRangeMatcher) FailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Expected %v to be within time range %v - %v", actual, m.min, m.max)
}

func (m *timeRangeMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Expected %v not to be within time range %v - %v", actual, m.min, m.max)
}

// RespondWithinTimeout checks if an operation completes within a timeout
func RespondWithinTimeout(timeout time.Duration) types.GomegaMatcher {
	return &respondWithinTimeoutMatcher{timeout: timeout}
}

type respondWithinTimeoutMatcher struct {
	timeout time.Duration
}

func (m *respondWithinTimeoutMatcher) Match(actual interface{}) (success bool, err error) {
	switch v := actual.(type) {
	case func():
		done := make(chan bool, 1)
		go func() {
			v()
			done <- true
		}()
		
		select {
		case <-done:
			return true, nil
		case <-time.After(m.timeout):
			return false, nil
		}
	case func() error:
		done := make(chan error, 1)
		go func() {
			done <- v()
		}()
		
		select {
		case err := <-done:
			return err == nil, err
		case <-time.After(m.timeout):
			return false, fmt.Errorf("operation timed out after %v", m.timeout)
		}
	default:
		return false, fmt.Errorf("expected function, got %T", actual)
	}
}

func (m *respondWithinTimeoutMatcher) FailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Expected operation to complete within %v", m.timeout)
}

func (m *respondWithinTimeoutMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Expected operation not to complete within %v", m.timeout)
}

// HaveValidJSONStructure checks if a string is valid JSON with expected structure
func HaveValidJSONStructure(expectedKeys []string) types.GomegaMatcher {
	return &validJSONStructureMatcher{expectedKeys: expectedKeys}
}

type validJSONStructureMatcher struct {
	expectedKeys []string
}

func (m *validJSONStructureMatcher) Match(actual interface{}) (success bool, err error) {
	jsonStr, ok := actual.(string)
	if !ok {
		return false, fmt.Errorf("expected string, got %T", actual)
	}
	
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return false, fmt.Errorf("invalid JSON: %v", err)
	}
	
	for _, key := range m.expectedKeys {
		if _, exists := data[key]; !exists {
			return false, nil
		}
	}
	
	return true, nil
}

func (m *validJSONStructureMatcher) FailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Expected %v to be valid JSON with keys %v", actual, m.expectedKeys)
}

func (m *validJSONStructureMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Expected %v not to be valid JSON with keys %v", actual, m.expectedKeys)
}

// HaveHTTPStatusCode checks if an HTTP response has a specific status code
func HaveHTTPStatusCode(expectedStatus int) types.GomegaMatcher {
	return &httpStatusMatcher{expectedStatus: expectedStatus}
}

type httpStatusMatcher struct {
	expectedStatus int
}

func (m *httpStatusMatcher) Match(actual interface{}) (success bool, err error) {
	switch v := actual.(type) {
	case *http.Response:
		return v.StatusCode == m.expectedStatus, nil
	case int:
		return v == m.expectedStatus, nil
	default:
		return false, fmt.Errorf("expected *http.Response or int, got %T", actual)
	}
}

func (m *httpStatusMatcher) FailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Expected HTTP status %d, got %v", m.expectedStatus, actual)
}

func (m *httpStatusMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Expected HTTP status not to be %d, got %v", m.expectedStatus, actual)
}

// ContainElementWithProperty checks if a slice contains an element with a specific property value
func ContainElementWithProperty(propertyName string, expectedValue interface{}) types.GomegaMatcher {
	return &containElementWithPropertyMatcher{
		propertyName:  propertyName,
		expectedValue: expectedValue,
	}
}

type containElementWithPropertyMatcher struct {
	propertyName  string
	expectedValue interface{}
}

func (m *containElementWithPropertyMatcher) Match(actual interface{}) (success bool, err error) {
	value := reflect.ValueOf(actual)
	if value.Kind() != reflect.Slice && value.Kind() != reflect.Array {
		return false, fmt.Errorf("expected slice or array, got %T", actual)
	}
	
	for i := 0; i < value.Len(); i++ {
		element := value.Index(i).Interface()
		
		if hasProperty(element, m.propertyName, m.expectedValue) {
			return true, nil
		}
	}
	
	return false, nil
}

func (m *containElementWithPropertyMatcher) FailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Expected %v to contain element with %s = %v", actual, m.propertyName, m.expectedValue)
}

func (m *containElementWithPropertyMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Expected %v not to contain element with %s = %v", actual, m.propertyName, m.expectedValue)
}

func hasProperty(element interface{}, propertyName string, expectedValue interface{}) bool {
	switch v := element.(type) {
	case map[string]interface{}:
		if value, exists := v[propertyName]; exists {
			return reflect.DeepEqual(value, expectedValue)
		}
	default:
		// Try to access property using reflection
		value := reflect.ValueOf(element)
		if value.Kind() == reflect.Ptr {
			value = value.Elem()
		}
		
		if value.Kind() == reflect.Struct {
			field := value.FieldByName(strings.Title(propertyName))
			if field.IsValid() {
				return reflect.DeepEqual(field.Interface(), expectedValue)
			}
		}
	}
	
	return false
}

// BeValidUUID checks if a string is a valid UUID
func BeValidUUID() types.GomegaMatcher {
	return &validUUIDMatcher{}
}

type validUUIDMatcher struct{}

func (m *validUUIDMatcher) Match(actual interface{}) (success bool, err error) {
	str, ok := actual.(string)
	if !ok {
		return false, fmt.Errorf("expected string, got %T", actual)
	}
	
	// Simple UUID validation - matches standard UUID format
	uuidPattern := regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
	matched := uuidPattern.MatchString(str)
	
	return matched, nil
}

func (m *validUUIDMatcher) FailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Expected %v to be a valid UUID", actual)
}

func (m *validUUIDMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Expected %v not to be a valid UUID", actual)
}