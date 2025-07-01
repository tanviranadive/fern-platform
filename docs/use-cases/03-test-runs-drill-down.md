# Test Runs and Drill-Down Use Cases

## Overview

The Test Runs page provides a detailed view of all test executions across team projects. Users can explore test results at multiple levels - from high-level run summaries down to individual test spec details. The page enforces team-based access control and provides hierarchical navigation through test runs, suites, and specs with accurate metrics at each level.

## Actors

- **Team Member**: A regular user who belongs to a team group and can view team test runs
- **Manager**: A user with manager role who can view and analyze team test runs
- **Platform Admin**: A user with admin role who can view test runs across all teams
- **System**: The Fern Platform backend providing test execution data

## Prerequisites

- User must be authenticated (see [00-authentication.md](./00-authentication.md))
- User must belong to at least one team group AND (user or manager) group
- Test data must exist in the database (inserted via `scripts/insert-test-data.sh`)
- User has navigated to the Test Runs page

## Use Cases

### UC-03-01: View Test Runs List

**As a** Team Member, Manager, or Admin  
**I want to** view all test runs for my accessible projects  
**So that** I can monitor test execution history and results

#### Acceptance Criteria

```gherkin
Feature: Test Runs List View
  As an authenticated user
  I want to view test runs for projects I have access to
  So that I can analyze test execution history

  Background:
    Given I have completed authentication as per UC-00-01
    And test data has been loaded via scripts/insert-test-data.sh
    And I have navigated to the Test Runs page

  Scenario: Team member views team test runs
    Given I am logged in as a member of "<team_name>"
    And there are <total_runs> test runs in the database
    And <team_runs> test runs belong to projects in "<team_name>"
    When the Test Runs page loads
    Then I should see exactly <team_runs> test run entries
    And I should not see test runs from other teams' projects
    And each test run should display columns in this order:
      | Column | Description | Example |
      | Project | Project name | E-Commerce Frontend |
      | Run ID | Unique test run identifier | run-123abc |
      | Branch | Git branch name | main |
      | Test Results | Total Failed Passed format | 47 2 45 |
      | Status | Overall run status | Passed/Failed |
      | Duration | Run duration in ms or seconds | 1,234ms |
      | Started | Start timestamp | 2024-01-15 10:30:45 |

  Scenario: Test results accuracy
    Given I can see a test run for project "<project_name>"
    And the database shows:
      | Metric | Value |
      | Total Tests | <total_tests> |
      | Passed Tests | <passed_tests> |
      | Failed Tests | <failed_tests> |
    When I look at the Test Results column
    Then it should display "<total_tests> <failed_tests> <passed_tests>"
    And the status should show "Passed" if <failed_tests> equals 0
    And the status should show "Failed" if <failed_tests> is greater than 0

  Scenario: Duration data accuracy
    Given a test run has start_time "<start_time>"
    And end_time "<end_time>"
    When I look at the Duration column
    Then it should show the accurate duration in milliseconds
    And format as "1,234ms" for durations under 60 seconds
    And format as "1m 23s" for durations over 60 seconds

  Scenario: User without team projects sees no runs
    Given I am logged in as a member of "<team_name>"
    And no projects are assigned to "<team_name>"
    When the Test Runs page loads
    Then I should see an empty state message
    And the message should say "No test runs found for your team's projects."
    And I should not see any test run entries

  Scenario: Platform admin views all test runs
    Given I am logged in as a platform admin
    When the Test Runs page loads
    Then I should see all <total_runs> test runs
    And I should see test runs from all teams' projects
    And the project column should indicate which team owns each project

  Scenario: Test runs sorted by recency
    Given I can see multiple test runs
    When the page loads
    Then test runs should be sorted by Started time descending
    And the most recent test run should appear first
    And older test runs should appear below
```

### UC-03-02: Navigate to Test Suite Details

**As a** User viewing test runs  
**I want to** click on a test run to see suite-level details  
**So that** I can understand which test suites passed or failed

#### Acceptance Criteria

```gherkin
Feature: Test Suite Details View
  As a user viewing test runs
  I want to drill down to suite-level details
  So that I can analyze test results by suite

  Background:
    Given I am on the Test Runs page
    And I can see test runs for my team's projects

  Scenario: Click test run to view suites
    Given I see a test run with ID "<run_id>"
    When I click on the test run row
    Then I should be navigated to the suite details page
    And the URL should include the run ID
    And I should see a list of all test suites for that run

  Scenario: Suite details display
    Given I have clicked on test run "<run_id>"
    And the run contains <suite_count> test suites
    When the suite details page loads
    Then I should see exactly <suite_count> suite entries
    And each suite should display columns in this order:
      | Column | Description | Example |
      | Suite Name | Name of the test suite | authentication.spec.js |
      | Test Results | Total Failed Passed format | 13 1 12 |
      | Status | Suite execution status | Passed/Failed |
      | Duration | Suite duration in ms | 456ms |

  Scenario: Suite metrics accuracy
    Given I am viewing suites for test run "<run_id>"
    And a suite has the following database data:
      | Metric | Value |
      | Total Specs | <total_specs> |
      | Passed Specs | <passed_specs> |
      | Failed Specs | <failed_specs> |
      | Duration | <duration_ms> |
    When I look at that suite's row
    Then Test Results should show "<total_specs> <failed_specs> <passed_specs>"
    And Status should show "Passed" if <failed_specs> equals 0
    And Status should show "Failed" if <failed_specs> is greater than 0
    And Duration should show "<duration_ms>ms"

  Scenario: Sum of suite metrics equals run metrics
    Given I am viewing suites for a test run
    And the test run showed "<run_total> <run_failed> <run_passed>" tests
    When I sum all suite Test Results
    Then the total passed specs should equal <run_passed>
    And the total failed specs should equal <run_failed>
    And the total specs should equal <run_total>

  Scenario: Breadcrumb navigation appears
    Given I have navigated to suite details
    When I look at the top of the page
    Then I should see breadcrumbs showing:
      | Level | Text | Clickable |
      | 1 | Test Runs | Yes |
      | 2 | <project_name> - <run_id> | No |
    When I click "Test Runs" in the breadcrumb
    Then I should return to the Test Runs list page
```

### UC-03-03: Navigate to Test Spec Details

**As a** User viewing test suites  
**I want to** click on a test suite to see spec-level details  
**So that** I can see individual test results and failure reasons

#### Acceptance Criteria

```gherkin
Feature: Test Spec Details View
  As a user viewing test suites
  I want to drill down to spec-level details
  So that I can see individual test results

  Background:
    Given I am viewing suite details for test run "<run_id>"
    And I can see a list of test suites

  Scenario: Click suite to view specs
    Given I see a suite named "<suite_name>"
    When I click on the suite row
    Then I should be navigated to the spec details page
    And I should see all test specs for that suite

  Scenario: Spec details display
    Given I have clicked on suite "<suite_name>"
    And the suite contains <spec_count> test specs
    When the spec details page loads
    Then I should see exactly <spec_count> spec entries
    And each spec should display columns in this order:
      | Column | Description | Example |
      | Test Name | Name of the test spec | should authenticate user |
      | Status | Test execution status | Passed/Failed/Skipped |
      | Duration | Spec duration in ms | 123ms |
      | Error Message | Failure reason if failed | Expected true but got false |
      | Started | Start timestamp | 10:30:45.123 |

  Scenario: Failed test shows error message
    Given I am viewing specs for suite "<suite_name>"
    And a spec has status "Failed"
    When I look at that spec's row
    Then the Error Message column should contain text
    And the error message should match the database error_message field
    And the text should be readable and not truncated in the table

  Scenario: Passed test has no error message
    Given I am viewing specs for suite "<suite_name>"
    And a spec has status "Passed"
    When I look at that spec's row
    Then the Error Message column should be empty
    And no error text should be displayed

  Scenario: Spec data accuracy
    Given a spec in the database has:
      | Field | Value |
      | spec_name | <spec_name> |
      | status | <status> |
      | duration | <duration_ms> |
      | error_message | <error_msg> |
      | start_time | <start_time> |
    When I view this spec in the list
    Then Test Name should show "<spec_name>"
    And Status should show "<status>"
    And Duration should show "<duration_ms>ms"
    And Error Message should show "<error_msg>" or be empty
    And Started should show the time portion of "<start_time>"

  Scenario: Breadcrumb navigation updated
    Given I have navigated to spec details
    When I look at the breadcrumbs
    Then I should see:
      | Level | Text | Clickable |
      | 1 | Test Runs | Yes |
      | 2 | <project_name> - <run_id> | Yes |
      | 3 | <suite_name> | No |
    When I click "<project_name> - <run_id>"
    Then I should return to the suite details page
```

### UC-03-04: Multi-Level Navigation

**As a** User exploring test results  
**I want to** navigate between different levels using breadcrumbs  
**So that** I can move efficiently through the hierarchy

#### Acceptance Criteria

```gherkin
Feature: Hierarchical Navigation
  As a user analyzing test results
  I want to navigate between levels easily
  So that I can explore data efficiently

  Background:
    Given I am logged in and have access to test runs

  Scenario: Navigate from runs to suites to specs
    Given I am on the Test Runs page
    When I click on test run "<run_id>"
    Then I should see the suite list
    When I click on suite "<suite_name>"
    Then I should see the spec list
    And the navigation should maintain context

  Scenario: Use breadcrumbs to go back one level
    Given I am viewing spec details
    And breadcrumbs show "Test Runs > Project - Run > Suite"
    When I click on "Project - Run" in breadcrumbs
    Then I should return to the suite list
    And I should see all suites for that test run

  Scenario: Use breadcrumbs to go back to top level
    Given I am viewing spec details
    When I click on "Test Runs" in breadcrumbs
    Then I should return to the test runs list
    And I should see all test runs for my team

  Scenario: Breadcrumb state preservation
    Given I navigated from Test Runs to Suites to Specs
    When I use breadcrumbs to go back
    Then each level should show the same data as before
    And scroll positions should be preserved
    And any filters or sorting should be maintained

  Scenario: Direct URL navigation
    Given I have a direct URL to suite details "/test-runs/<run_id>/suites"
    When I navigate to this URL
    Then I should see the suite details page
    And breadcrumbs should be properly populated
    And I should be able to navigate up or down the hierarchy
```

### UC-03-05: Access Control at Each Level

**As a** System  
**I want to** enforce access control at every navigation level  
**So that** users only see data from their team's projects

#### Acceptance Criteria

```gherkin
Feature: Test Run Access Control
  As the system
  I want to enforce team-based access
  So that data remains secure

  Background:
    Given test runs exist for multiple teams

  Scenario: Cannot access other team's test run
    Given I am a member of "<team_name>"
    And a test run "<run_id>" belongs to "<other_team>"
    When I try to directly access "/test-runs/<run_id>"
    Then I should see an access denied error
    And the error should say "You don't have permission to view this test run"

  Scenario: Cannot see other team's data in lists
    Given I am a member of "<team_name>"
    When I view any level (runs, suites, or specs)
    Then I should only see data from "<team_name>" projects
    And no data from "<other_team>" should be visible
    And this should be verified at each navigation level

  Scenario: Admin can access all team data
    Given I am logged in as a platform admin
    When I navigate to test run "<run_id>" from any team
    Then I should be able to access it
    And I should be able to drill down to all levels
    And I should see accurate data at each level
```

### UC-03-06: Empty States and Error Handling

**As a** User  
**I want to** see helpful messages when data is unavailable  
**So that** I understand the system state

#### Acceptance Criteria

```gherkin
Feature: Empty States and Errors
  As a user viewing test data
  I want clear messaging for empty or error states
  So that I know what's happening

  Scenario: No test runs available
    Given I am a member of a team with projects
    But no test runs have been executed yet
    When I visit the Test Runs page
    Then I should see "No test runs found"
    And a message "Run your tests to see results here"

  Scenario: Test run has no suites
    Given a test run exists but has no suite data
    When I click on that test run
    Then I should see "No test suites found for this run"
    And a message indicating the run may have failed to start

  Scenario: Suite has no specs
    Given a suite exists but has no spec data
    When I click on that suite
    Then I should see "No test specs found for this suite"

  Scenario: Data loading errors
    Given there's a database connection issue
    When I try to load any test data page
    Then I should see an error message
    And the message should be user-friendly
    And I should see a "Retry" button
```

## Test Data Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `<team_name>` | Team the user belongs to | `frontend-team`, `backend-team` |
| `<other_team>` | A different team for access control | `mobile-team`, `qa-team` |
| `<total_runs>` | Total test runs in database | `156` |
| `<team_runs>` | Test runs for user's team | `43` |
| `<project_name>` | Name of the project | `E-Commerce Frontend` |
| `<run_id>` | Unique test run identifier | `run-abc123def` |
| `<total_tests>` | Total tests in a run | `147` |
| `<passed_tests>` | Passed tests in a run | `142` |
| `<failed_tests>` | Failed tests in a run | `5` |
| `<start_time>` | Run start timestamp | `2024-01-15T10:30:45Z` |
| `<end_time>` | Run end timestamp | `2024-01-15T10:32:19Z` |
| `<suite_count>` | Number of suites in a run | `12` |
| `<suite_name>` | Name of a test suite | `authentication.spec.js` |
| `<total_specs>` | Total specs in a suite | `15` |
| `<passed_specs>` | Passed specs in a suite | `14` |
| `<failed_specs>` | Failed specs in a suite | `1` |
| `<duration_ms>` | Duration in milliseconds | `1234` |
| `<spec_count>` | Number of specs in a suite | `15` |
| `<spec_name>` | Name of a test spec | `should authenticate valid user` |
| `<status>` | Test status | `Passed`, `Failed`, `Skipped` |
| `<error_msg>` | Error message for failed test | `Expected true but got false` |

### Example Test Execution

```bash
# Running test runs page tests with Ginkgo
ginkgo -v \
  -team-name="frontend-team" \
  -other-team="backend-team" \
  -total-runs=156 \
  -team-runs=43 \
  -run-id="run-abc123def" \
  -project-name="E-Commerce Frontend" \
  -base-url="http://fern-platform.local:8080"
```

## Data Display Formats

### Test Results Format
- Display: `{total} {failed} {passed}`
- Example: `47 2 45`
- Color coding:
  - All passed (0 failed): Green text
  - Some failed: Red text
  - Some skipped: Include in total but not in passed/failed

### Duration Format
- Under 1 second: `{ms}ms` (e.g., `756ms`)
- 1-59 seconds: `{s}s` (e.g., `45s`)
- Over 60 seconds: `{m}m {s}s` (e.g., `2m 15s`)

### Timestamp Format
- Full timestamp: `YYYY-MM-DD HH:mm:ss`
- Time only (for specs): `HH:mm:ss.SSS`

### Status Display
- **Passed**: Green badge/text
- **Failed**: Red badge/text
- **Skipped**: Gray badge/text
- **Running**: Blue badge/text with spinner (if applicable)

## Column Specifications

### Test Runs Table
| Column | Width | Alignment | Sortable |
|--------|-------|-----------|----------|
| Project | 20% | Left | Yes |
| Run ID | 15% | Left | Yes |
| Branch | 15% | Left | Yes |
| Test Results | 15% | Center | Yes |
| Status | 10% | Center | Yes |
| Duration | 10% | Right | Yes |
| Started | 15% | Left | Yes |

### Suite Details Table
| Column | Width | Alignment | Sortable |
|--------|-------|-----------|----------|
| Suite Name | 40% | Left | Yes |
| Test Results | 20% | Center | Yes |
| Status | 20% | Center | Yes |
| Duration | 20% | Right | Yes |

### Spec Details Table
| Column | Width | Alignment | Sortable |
|--------|-------|-----------|----------|
| Test Name | 30% | Left | Yes |
| Status | 15% | Center | Yes |
| Duration | 15% | Right | Yes |
| Error Message | 25% | Left | No |
| Started | 15% | Left | Yes |

## Performance Considerations

1. **Pagination**: Implement pagination for large test run lists (>100 items)
2. **Lazy Loading**: Load suite/spec details only when requested
3. **Caching**: Cache recently viewed run/suite data for quick navigation
4. **Virtual Scrolling**: For very large spec lists (>500 items)

## Security Validations

1. **Team Membership**: Validate user belongs to project's team at each level
2. **Run Ownership**: Verify run belongs to an accessible project
3. **URL Tampering**: Protect against direct URL manipulation
4. **Data Consistency**: Ensure child data belongs to parent (specs -> suite -> run)

## Future Enhancements

1. **Filtering**: Add filters for status, date range, branch
2. **Search**: Full-text search across test names
3. **Bulk Operations**: Select multiple runs for comparison
4. **Export**: Download test results as CSV/JSON
5. **Real-time Updates**: Show live test execution progress
6. **Failure Analysis**: Group similar failures across runs