package graphql

import (
	"strconv"
	"testing"
	"time"

	tagsDomain "github.com/guidewire-oss/fern-platform/internal/domains/tags/domain"
	testingDomain "github.com/guidewire-oss/fern-platform/internal/domains/testing/domain"
	"github.com/guidewire-oss/fern-platform/internal/reporter/graphql/model"
	"github.com/guidewire-oss/fern-platform/pkg/config"
	"github.com/guidewire-oss/fern-platform/pkg/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestResolver creates a test resolver with minimal dependencies
func setupTestResolver(t *testing.T) *Resolver {
	logger, err := logging.NewLogger(&config.LoggingConfig{
		Level:      "info",
		Format:     "json",
		Output:     "stdout",
		Structured: true,
	})
	require.NoError(t, err)

	return &Resolver{
		logger: logger,
	}
}

func TestConvertTestRunToGraphQL_TagConversion(t *testing.T) {
	resolver := setupTestResolver(t)
	now := time.Now()

	tests := []struct {
		name            string
		testRunTags     []testingDomain.Tag
		expectedTagsLen int
		validateTags    func(t *testing.T, tags []*model.Tag)
	}{
		{
			name:            "empty tags",
			testRunTags:     []testingDomain.Tag{},
			expectedTagsLen: 0,
			validateTags: func(t *testing.T, tags []*model.Tag) {
				assert.Empty(t, tags)
			},
		},
		{
			name: "single tag without category",
			testRunTags: []testingDomain.Tag{
				{ID: 1, Name: "smoke", Category: "", Value: "smoke"},
			},
			expectedTagsLen: 1,
			validateTags: func(t *testing.T, tags []*model.Tag) {
				require.Len(t, tags, 1)
				assert.Equal(t, "1", tags[0].ID)
				assert.Equal(t, "smoke", tags[0].Name)
				assert.Nil(t, tags[0].Category)
				assert.NotNil(t, tags[0].Value)
				assert.Equal(t, "smoke", *tags[0].Value)
			},
		},
		{
			name: "single tag with category",
			testRunTags: []testingDomain.Tag{
				{ID: 2, Name: "priority:high", Category: "priority", Value: "high"},
			},
			expectedTagsLen: 1,
			validateTags: func(t *testing.T, tags []*model.Tag) {
				require.Len(t, tags, 1)
				assert.Equal(t, "2", tags[0].ID)
				assert.Equal(t, "priority:high", tags[0].Name)
				assert.NotNil(t, tags[0].Category)
				assert.Equal(t, "priority", *tags[0].Category)
				assert.NotNil(t, tags[0].Value)
				assert.Equal(t, "high", *tags[0].Value)
			},
		},
		{
			name: "multiple tags with mixed categories",
			testRunTags: []testingDomain.Tag{
				{ID: 1, Name: "smoke", Category: "", Value: "smoke"},
				{ID: 2, Name: "priority:high", Category: "priority", Value: "high"},
				{ID: 3, Name: "env:staging", Category: "env", Value: "staging"},
			},
			expectedTagsLen: 3,
			validateTags: func(t *testing.T, tags []*model.Tag) {
				require.Len(t, tags, 3)

				// First tag - no category
				assert.Equal(t, "1", tags[0].ID)
				assert.Equal(t, "smoke", tags[0].Name)
				assert.Nil(t, tags[0].Category)
				assert.NotNil(t, tags[0].Value)
				assert.Equal(t, "smoke", *tags[0].Value)

				// Second tag - with category
				assert.Equal(t, "2", tags[1].ID)
				assert.Equal(t, "priority:high", tags[1].Name)
				assert.NotNil(t, tags[1].Category)
				assert.Equal(t, "priority", *tags[1].Category)
				assert.NotNil(t, tags[1].Value)
				assert.Equal(t, "high", *tags[1].Value)

				// Third tag - with category
				assert.Equal(t, "3", tags[2].ID)
				assert.Equal(t, "env:staging", tags[2].Name)
				assert.NotNil(t, tags[2].Category)
				assert.Equal(t, "env", *tags[2].Category)
				assert.NotNil(t, tags[2].Value)
				assert.Equal(t, "staging", *tags[2].Value)
			},
		},
		{
			name: "tag with empty category string",
			testRunTags: []testingDomain.Tag{
				{ID: 10, Name: "test", Category: "", Value: "test"},
			},
			expectedTagsLen: 1,
			validateTags: func(t *testing.T, tags []*model.Tag) {
				require.Len(t, tags, 1)
				// Empty string should convert to nil
				assert.Nil(t, tags[0].Category)
				assert.NotNil(t, tags[0].Value)
			},
		},
		{
			name: "tag with empty value string",
			testRunTags: []testingDomain.Tag{
				{ID: 11, Name: "category:", Category: "category", Value: ""},
			},
			expectedTagsLen: 1,
			validateTags: func(t *testing.T, tags []*model.Tag) {
				require.Len(t, tags, 1)
				assert.NotNil(t, tags[0].Category)
				// Empty string should convert to nil
				assert.Nil(t, tags[0].Value)
			},
		},
		{
			name: "tags with large IDs",
			testRunTags: []testingDomain.Tag{
				{ID: 999999, Name: "large-id", Category: "test", Value: "large"},
			},
			expectedTagsLen: 1,
			validateTags: func(t *testing.T, tags []*model.Tag) {
				require.Len(t, tags, 1)
				assert.Equal(t, "999999", tags[0].ID)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testRun := &testingDomain.TestRun{
				ID:           1,
				RunID:        "test-run-1",
				ProjectID:    "proj-1",
				Status:       "completed",
				StartTime:    now,
				TotalTests:   10,
				PassedTests:  8,
				FailedTests:  2,
				SkippedTests: 0,
				Duration:     5 * time.Second,
				Tags:         tt.testRunTags,
				SuiteRuns:    []testingDomain.SuiteRun{},
			}

			result := resolver.convertTestRunToGraphQL(testRun)

			require.NotNil(t, result)
			assert.Equal(t, tt.expectedTagsLen, len(result.Tags))
			tt.validateTags(t, result.Tags)

			// Verify tags use zero time for timestamps (as per comment in code)
			for _, tag := range result.Tags {
				assert.Equal(t, time.Time{}, tag.CreatedAt)
				assert.Equal(t, time.Time{}, tag.UpdatedAt)
			}
		})
	}
}

func TestConvertSuiteRunToGraphQL_TagConversion(t *testing.T) {
	resolver := setupTestResolver(t)
	now := time.Now()

	tests := []struct {
		name            string
		suiteRunTags    []testingDomain.Tag
		expectedTagsLen int
		validateTags    func(t *testing.T, tags []*model.Tag)
	}{
		{
			name:            "empty tags",
			suiteRunTags:    []testingDomain.Tag{},
			expectedTagsLen: 0,
			validateTags: func(t *testing.T, tags []*model.Tag) {
				assert.Empty(t, tags)
			},
		},
		{
			name: "single tag",
			suiteRunTags: []testingDomain.Tag{
				{ID: 5, Name: "integration", Category: "type", Value: "integration"},
			},
			expectedTagsLen: 1,
			validateTags: func(t *testing.T, tags []*model.Tag) {
				require.Len(t, tags, 1)
				assert.Equal(t, "5", tags[0].ID)
				assert.Equal(t, "integration", tags[0].Name)
				assert.NotNil(t, tags[0].Category)
				assert.Equal(t, "type", *tags[0].Category)
			},
		},
		{
			name: "multiple tags",
			suiteRunTags: []testingDomain.Tag{
				{ID: 1, Name: "fast", Category: "speed", Value: "fast"},
				{ID: 2, Name: "critical", Category: "importance", Value: "critical"},
			},
			expectedTagsLen: 2,
			validateTags: func(t *testing.T, tags []*model.Tag) {
				require.Len(t, tags, 2)
				assert.Equal(t, "1", tags[0].ID)
				assert.Equal(t, "2", tags[1].ID)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			suite := &testingDomain.SuiteRun{
				ID:           1,
				TestRunID:    1,
				Name:         "test-suite",
				Status:       "passed",
				StartTime:    now,
				TotalTests:   5,
				PassedTests:  5,
				FailedTests:  0,
				SkippedTests: 0,
				Duration:     2 * time.Second,
				Tags:         tt.suiteRunTags,
				SpecRuns:     []*testingDomain.SpecRun{},
			}

			result := resolver.convertSuiteRunToGraphQL(suite)

			require.NotNil(t, result)
			assert.Equal(t, tt.expectedTagsLen, len(result.Tags))
			tt.validateTags(t, result.Tags)

			// Verify tags use zero time for timestamps
			for _, tag := range result.Tags {
				assert.Equal(t, time.Time{}, tag.CreatedAt)
				assert.Equal(t, time.Time{}, tag.UpdatedAt)
			}
		})
	}
}

func TestConvertSpecRunToGraphQL_TagConversion(t *testing.T) {
	resolver := setupTestResolver(t)
	now := time.Now()

	tests := []struct {
		name            string
		specRunTags     []testingDomain.Tag
		expectedTagsLen int
		validateTags    func(t *testing.T, tags []*model.Tag)
	}{
		{
			name:            "empty tags",
			specRunTags:     []testingDomain.Tag{},
			expectedTagsLen: 0,
			validateTags: func(t *testing.T, tags []*model.Tag) {
				assert.Empty(t, tags)
			},
		},
		{
			name: "single tag",
			specRunTags: []testingDomain.Tag{
				{ID: 7, Name: "flaky", Category: "", Value: "flaky"},
			},
			expectedTagsLen: 1,
			validateTags: func(t *testing.T, tags []*model.Tag) {
				require.Len(t, tags, 1)
				assert.Equal(t, "7", tags[0].ID)
				assert.Equal(t, "flaky", tags[0].Name)
			},
		},
		{
			name: "multiple tags with special characters",
			specRunTags: []testingDomain.Tag{
				{ID: 1, Name: "browser:chrome-v120", Category: "browser", Value: "chrome-v120"},
				{ID: 2, Name: "os:linux_ubuntu-22.04", Category: "os", Value: "linux_ubuntu-22.04"},
			},
			expectedTagsLen: 2,
			validateTags: func(t *testing.T, tags []*model.Tag) {
				require.Len(t, tags, 2)
				assert.Equal(t, "browser:chrome-v120", tags[0].Name)
				assert.Equal(t, "os:linux_ubuntu-22.04", tags[1].Name)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec := &testingDomain.SpecRun{
				ID:         1,
				SuiteRunID: 1,
				Name:       "test-spec",
				Status:     "passed",
				StartTime:  now,
				Duration:   1 * time.Second,
				Tags:       tt.specRunTags,
			}

			result := resolver.convertSpecRunToGraphQL(spec)

			require.NotNil(t, result)
			assert.Equal(t, tt.expectedTagsLen, len(result.Tags))
			tt.validateTags(t, result.Tags)

			// Verify tags use zero time for timestamps
			for _, tag := range result.Tags {
				assert.Equal(t, time.Time{}, tag.CreatedAt)
				assert.Equal(t, time.Time{}, tag.UpdatedAt)
			}
		})
	}
}

func TestConvertTagToGraphQL(t *testing.T) {
	resolver := setupTestResolver(t)
	now := time.Now()

	tests := []struct {
		name     string
		domainTag *tagsDomain.Tag
		validate func(t *testing.T, result *model.Tag)
	}{
		{
			name: "tag with category and value",
			domainTag: tagsDomain.ReconstructTag(
				tagsDomain.TagID("tag-123"),
				"priority:high",
				"priority",
				"high",
				now,
			),
			validate: func(t *testing.T, result *model.Tag) {
				assert.Equal(t, "tag-123", result.ID)
				assert.Equal(t, "priority:high", result.Name)
				assert.NotNil(t, result.Category)
				assert.Equal(t, "priority", *result.Category)
				assert.NotNil(t, result.Value)
				assert.Equal(t, "high", *result.Value)
				assert.Equal(t, now, result.CreatedAt)
				assert.Equal(t, now, result.UpdatedAt) // Should be same as CreatedAt (immutable)
				assert.Nil(t, result.Description)
				assert.Nil(t, result.Color)
			},
		},
		{
			name: "tag without category",
			domainTag: tagsDomain.ReconstructTag(
				tagsDomain.TagID("tag-456"),
				"smoke",
				"",
				"smoke",
				now,
			),
			validate: func(t *testing.T, result *model.Tag) {
				assert.Equal(t, "tag-456", result.ID)
				assert.Equal(t, "smoke", result.Name)
				assert.Nil(t, result.Category)
				assert.NotNil(t, result.Value)
				assert.Equal(t, "smoke", *result.Value)
			},
		},
		{
			name: "tag with empty value",
			domainTag: tagsDomain.ReconstructTag(
				tagsDomain.TagID("tag-789"),
				"category:",
				"category",
				"",
				now,
			),
			validate: func(t *testing.T, result *model.Tag) {
				assert.Equal(t, "tag-789", result.ID)
				assert.NotNil(t, result.Category)
				assert.Equal(t, "category", *result.Category)
				assert.Nil(t, result.Value) // Empty string converts to nil
			},
		},
		{
			name: "tag with long name",
			domainTag: tagsDomain.ReconstructTag(
				tagsDomain.TagID("tag-long"),
				"environment:production-us-east-1-cluster-a",
				"environment",
				"production-us-east-1-cluster-a",
				now,
			),
			validate: func(t *testing.T, result *model.Tag) {
				assert.Equal(t, "tag-long", result.ID)
				assert.Equal(t, "environment:production-us-east-1-cluster-a", result.Name)
				assert.NotNil(t, result.Category)
				assert.Equal(t, "environment", *result.Category)
				assert.NotNil(t, result.Value)
				assert.Equal(t, "production-us-east-1-cluster-a", *result.Value)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := resolver.convertTagToGraphQL(tt.domainTag)
			require.NotNil(t, result)
			tt.validate(t, result)
		})
	}
}

func TestConvertTestRunToGraphQL_NestedTagsInSuites(t *testing.T) {
	resolver := setupTestResolver(t)
	now := time.Now()

	// Create test run with nested suite runs and spec runs, all with tags
	testRun := &testingDomain.TestRun{
		ID:           1,
		RunID:        "test-run-nested",
		ProjectID:    "proj-1",
		Status:       "completed",
		StartTime:    now,
		TotalTests:   10,
		PassedTests:  8,
		FailedTests:  2,
		SkippedTests: 0,
		Duration:     10 * time.Second,
		Tags: []testingDomain.Tag{
			{ID: 1, Name: "run-tag", Category: "", Value: "run-tag"},
		},
		SuiteRuns: []testingDomain.SuiteRun{
			{
				ID:           1,
				TestRunID:    1,
				Name:         "suite-1",
				Status:       "passed",
				StartTime:    now,
				TotalTests:   5,
				PassedTests:  5,
				FailedTests:  0,
				SkippedTests: 0,
				Duration:     5 * time.Second,
				Tags: []testingDomain.Tag{
					{ID: 2, Name: "suite-tag:value1", Category: "suite-tag", Value: "value1"},
				},
				SpecRuns: []*testingDomain.SpecRun{
					{
						ID:         1,
						SuiteRunID: 1,
						Name:       "spec-1",
						Status:     "passed",
						StartTime:  now,
						Duration:   1 * time.Second,
						Tags: []testingDomain.Tag{
							{ID: 3, Name: "spec-tag", Category: "", Value: "spec-tag"},
						},
					},
				},
			},
		},
	}

	result := resolver.convertTestRunToGraphQL(testRun)

	require.NotNil(t, result)

	// Verify test run tags
	require.Len(t, result.Tags, 1)
	assert.Equal(t, "1", result.Tags[0].ID)
	assert.Equal(t, "run-tag", result.Tags[0].Name)

	// Verify suite run tags
	require.Len(t, result.SuiteRuns, 1)
	require.Len(t, result.SuiteRuns[0].Tags, 1)
	assert.Equal(t, "2", result.SuiteRuns[0].Tags[0].ID)
	assert.Equal(t, "suite-tag:value1", result.SuiteRuns[0].Tags[0].Name)
	assert.NotNil(t, result.SuiteRuns[0].Tags[0].Category)
	assert.Equal(t, "suite-tag", *result.SuiteRuns[0].Tags[0].Category)

	// Verify spec run tags
	require.Len(t, result.SuiteRuns[0].SpecRuns, 1)
	require.Len(t, result.SuiteRuns[0].SpecRuns[0].Tags, 1)
	assert.Equal(t, "3", result.SuiteRuns[0].SpecRuns[0].Tags[0].ID)
	assert.Equal(t, "spec-tag", result.SuiteRuns[0].SpecRuns[0].Tags[0].Name)
}

func TestConvertStringPtr(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected *string
	}{
		{
			name:     "empty string returns nil",
			input:    "",
			expected: nil,
		},
		{
			name:     "non-empty string returns pointer",
			input:    "test",
			expected: strPtr("test"),
		},
		{
			name:     "whitespace string returns pointer",
			input:    "   ",
			expected: strPtr("   "),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertStringPtr(tt.input)
			if tt.expected == nil {
				assert.Nil(t, result)
			} else {
				require.NotNil(t, result)
				assert.Equal(t, *tt.expected, *result)
			}
		})
	}
}

func TestConvertSpecRunToGraphQL_ErrorAndStackTrace(t *testing.T) {
	resolver := setupTestResolver(t)
	now := time.Now()

	tests := []struct {
		name             string
		errorMessage     string
		stackTrace       string
		expectErrorMsg   bool
		expectStackTrace bool
	}{
		{
			name:             "both empty",
			errorMessage:     "",
			stackTrace:       "",
			expectErrorMsg:   false,
			expectStackTrace: false,
		},
		{
			name:             "only error message",
			errorMessage:     "test failed",
			stackTrace:       "",
			expectErrorMsg:   true,
			expectStackTrace: false,
		},
		{
			name:             "only stack trace",
			errorMessage:     "",
			stackTrace:       "at line 10",
			expectErrorMsg:   false,
			expectStackTrace: true,
		},
		{
			name:             "both present",
			errorMessage:     "assertion failed",
			stackTrace:       "at test.go:42",
			expectErrorMsg:   true,
			expectStackTrace: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec := &testingDomain.SpecRun{
				ID:           1,
				SuiteRunID:   1,
				Name:         "test-spec",
				Status:       "failed",
				StartTime:    now,
				Duration:     1 * time.Second,
				ErrorMessage: tt.errorMessage,
				StackTrace:   tt.stackTrace,
				Tags:         []testingDomain.Tag{},
			}

			result := resolver.convertSpecRunToGraphQL(spec)

			require.NotNil(t, result)

			if tt.expectErrorMsg {
				require.NotNil(t, result.ErrorMessage)
				assert.Equal(t, tt.errorMessage, *result.ErrorMessage)
			} else {
				assert.Nil(t, result.ErrorMessage)
			}

			if tt.expectStackTrace {
				require.NotNil(t, result.StackTrace)
				assert.Equal(t, tt.stackTrace, *result.StackTrace)
			} else {
				assert.Nil(t, result.StackTrace)
			}
		})
	}
}

func TestConvertTestRunToGraphQL_IDConversion(t *testing.T) {
	resolver := setupTestResolver(t)
	now := time.Now()

	tests := []struct {
		name          string
		testRunID     uint
		expectedIDStr string
	}{
		{
			name:          "small ID",
			testRunID:     1,
			expectedIDStr: "1",
		},
		{
			name:          "large ID",
			testRunID:     999999999,
			expectedIDStr: "999999999",
		},
		{
			name:          "medium ID",
			testRunID:     12345,
			expectedIDStr: "12345",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testRun := &testingDomain.TestRun{
				ID:           tt.testRunID,
				RunID:        "run-id",
				ProjectID:    "proj-1",
				Status:       "completed",
				StartTime:    now,
				TotalTests:   10,
				PassedTests:  10,
				FailedTests:  0,
				SkippedTests: 0,
				Duration:     5 * time.Second,
				Tags:         []testingDomain.Tag{},
				SuiteRuns:    []testingDomain.SuiteRun{},
			}

			result := resolver.convertTestRunToGraphQL(testRun)

			require.NotNil(t, result)
			assert.Equal(t, tt.expectedIDStr, result.ID)

			// Verify the string can be parsed back to uint
			parsedID, err := strconv.ParseUint(result.ID, 10, 32)
			require.NoError(t, err)
			assert.Equal(t, uint64(tt.testRunID), parsedID)
		})
	}
}

// Helper function to create string pointer
func strPtr(s string) *string {
	return &s
}
