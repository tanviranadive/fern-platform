# Test Summaries and Visualization Use Cases

## Overview

The Test Summaries page is the primary dashboard for viewing test analytics across projects. It provides multiple visualization modes including card view, treemap view, and historical trends. Users can see aggregated test data, project health metrics, and drill down into specific project details based on their team membership and role permissions.

## Actors

- **Team Member**: A regular user who belongs to a team group and can view team projects
- **Manager**: A user with manager role who can view and manage team projects
- **Platform Admin**: A user with admin role who can view all projects across all teams
- **System**: The Fern Platform backend providing test data

## Prerequisites

- User must be authenticated (see [00-authentication.md](./00-authentication.md))
- User must belong to at least one team group AND (user or manager) group
- Test data must exist in the database (inserted via `scripts/insert-test-data.sh`)
- User has navigated to the Test Summaries page

## Use Cases

### UC-02-01: View Test Summary Dashboard

**As a** Team Member, Manager, or Admin  
**I want to** view test summaries for my accessible projects  
**So that** I can monitor test health and metrics across projects

#### Acceptance Criteria

```gherkin
Feature: Test Summary Dashboard
  As an authenticated user
  I want to view test summaries for projects I have access to
  So that I can monitor testing metrics

  Background:
    Given I have completed authentication as per UC-00-01
    And test data has been loaded via scripts/insert-test-data.sh
    And I have navigated to the Test Summaries page

  Scenario: Team member views team projects
    Given I am logged in as a member of "<team_name>"
    And I belong to the "user" group
    And there are <total_projects> projects in the database
    And <team_projects> projects belong to "<team_name>"
    When the Test Summaries page loads
    Then I should see exactly <team_projects> project cards
    And I should not see projects from other teams
    And each project card should display:
      | Field | Description |
      | Project Name | The name of the project |
      | Total Tests | Accurate count from database |
      | Test Runs | Accurate count from database |
      | Pass Rate | Calculated percentage |
      | Last Run | Timestamp of most recent run |

  Scenario: Manager views team projects with management indicators
    Given I am logged in as a manager of "<team_name>"
    And I belong to both "manager" and "<team_name>" groups
    When the Test Summaries page loads
    Then I should see exactly <team_projects> project cards for my team
    And each project card should show management capabilities

  Scenario: Platform admin views all projects
    Given I am logged in as a platform admin
    When the Test Summaries page loads
    Then I should see all <total_projects> project cards
    And I should see projects from all teams
    And project cards should be grouped or labeled by team

  Scenario: User without team assignment sees no projects
    Given I am logged in but not assigned to any team
    When the Test Summaries page loads
    Then I should see an empty state message
    And the message should say "No projects available. Please contact your administrator to be assigned to a team."
    And I should not see any project cards
    And the summary metrics should show:
      | Metric | Value |
      | Active Projects | 0 |
      | Total Runs | 0 |
      | Recent Runs | 0 |

  Scenario: User not in team with projects sees no data
    Given I am logged in as a member of "<team_name>"
    And "<team_name>" has no projects assigned
    And other teams have projects in the system
    When the Test Summaries page loads
    Then I should not see any project cards from other teams
    And I should see an empty state message
    And the message should say "No projects found for your team."
    And I should not be able to access Treemap View
    And the view toggle buttons should be disabled

  Scenario: Summary metrics accuracy
    Given I can see <visible_projects> projects
    And the database contains specific test data
    When I look at the top-right summary metrics
    Then "Active Projects" should show <active_projects> (matching database)
    And "Total Runs" should show <total_runs> (sum of all visible project runs)
    And "Recent Runs" should show <recent_runs> (runs in last 24 hours)
```

### UC-02-02: Toggle Between Card and Treemap Views

**As a** User viewing test summaries  
**I want to** switch between card and treemap visualizations  
**So that** I can analyze test data in different ways

#### Acceptance Criteria

```gherkin
Feature: View Mode Toggle
  As a user on the Test Summaries page
  I want to switch between card and treemap views
  So that I can visualize data differently

  Background:
    Given I am viewing the Test Summaries page
    And I can see <project_count> project cards

  Scenario: View toggle disabled when no projects
    Given I am logged in as a member of "<team_name>"
    And my team has no projects
    When the Test Summaries page loads
    Then the "Card View" button should be disabled
    And the "Treemap View" button should be disabled
    And clicking these buttons should have no effect
    And I should not be able to switch views

  Scenario: Switch from Card View to Treemap View
    Given I am in "Card View" mode (default)
    When I click the "Treemap View" button
    Then all <project_count> cards should flip simultaneously
    And each card should animate with a vertical flip
    And the flip animation should take approximately 0.6 seconds
    And after flipping, each card should show a treemap visualization

  Scenario: Treemap visualization accuracy
    Given I have switched to Treemap View
    When I examine a project treemap for "<project_name>"
    Then the treemap should display test suites as rectangles
    And rectangle sizes should be proportional to suite duration
    And rectangle colors should represent pass rates:
      | Pass Rate | Color |
      | 90-100% | Green (#10b981) |
      | 70-89% | Yellow-Green |
      | 50-69% | Yellow (#f59e0b) |
      | 30-49% | Orange |
      | 0-29% | Red (#ef4444) |
    And hovering over a rectangle should show:
      | Field | Value |
      | Suite Name | <suite_name> |
      | Total Tests | <test_count> |
      | Pass Rate | <pass_percentage>% |
      | Duration | <duration_ms>ms |

  Scenario: Switch from Treemap View back to Card View
    Given I am in "Treemap View" mode
    When I click the "Card View" button
    Then all cards should flip back simultaneously
    And the flip animation should reverse
    And after flipping, cards should show the original card content
    And all test metrics should be visible again

  Scenario: Maintain view preference during session
    Given I have switched to "Treemap View"
    When I navigate away from Test Summaries
    And I return to the Test Summaries page
    Then the view should still be in "Treemap View"
    And the preference should persist for my session
```

### UC-02-03: Interact with Treemap Visualization

**As a** User in treemap view  
**I want to** interact with the treemap elements  
**So that** I can drill down into test suite details

#### Acceptance Criteria

```gherkin
Feature: Treemap Interaction
  As a user viewing treemaps
  I want to interact with treemap elements
  So that I can explore test data hierarchically

  Background:
    Given I am in Treemap View
    And I am viewing the treemap for "<project_name>"

  Scenario: Hover over test suite
    When I hover over a test suite rectangle
    Then a tooltip should appear
    And the tooltip should display:
      | Field | Example |
      | Suite | authentication.spec.js |
      | Tests | 15 |
      | Passed | 13 |
      | Failed | 2 |
      | Duration | 1250ms |
      | Pass Rate | 86.7% |

  Scenario: Click to drill down to test specs
    Given the treemap shows test suites
    When I click on a suite "<suite_name>"
    Then the treemap should animate and zoom into that suite
    And I should now see individual test specs as rectangles
    And spec rectangle sizes should be based on execution duration
    And spec colors should be:
      | Status | Color |
      | Passed | Green (#10b981) |
      | Failed | Red (#ef4444) |
      | Skipped | Gray (#6b7280) |

  Scenario: Navigate back from spec view
    Given I have drilled down to see test specs
    When I click the back button or breadcrumb
    Then the treemap should zoom out
    And I should return to the suite-level view
    And the animation should be smooth

  Scenario: Treemap responsiveness
    Given I am viewing a treemap
    When I resize my browser window
    Then the treemap should responsively adjust
    And maintain aspect ratio within the card
    And text labels should remain readable
```

### UC-02-04: View Test History

**As a** User viewing project cards  
**I want to** view historical test trends  
**So that** I can understand test stability over time

#### Acceptance Criteria

```gherkin
Feature: Test History Visualization
  As a user viewing project summaries
  I want to see test execution history
  So that I can track trends over time

  Background:
    Given I am in Card View
    And I am viewing the project "<project_name>"

  Scenario: Access test history
    When I click the "View Test History" button on a project card
    Then a modal or expanded view should appear
    And I should see a stacked area chart
    And the chart should show the last <time_range> of data

  Scenario: Stacked chart accuracy
    Given I am viewing the test history chart
    Then the chart should display:
      | Layer | Color | Data |
      | Passed Tests | Green | Count of passed tests per run |
      | Failed Tests | Red | Count of failed tests per run |
      | Skipped Tests | Gray | Count of skipped tests per run |
    And the X-axis should show dates/times
    And the Y-axis should show test counts
    And hovering over a point should show:
      | Field | Value |
      | Date | <run_date> |
      | Passed | <passed_count> |
      | Failed | <failed_count> |
      | Skipped | <skipped_count> |
      | Total | <total_count> |

  Scenario: Chart data matches database
    Given the database contains <run_count> test runs for the project
    When I view the test history chart
    Then the chart should plot exactly <run_count> data points
    And each data point's values should match the database records
    And the stacked totals should equal the total tests per run

  Scenario: Time range selection in history
    Given I am viewing the test history
    When I see the time range selector
    Then I should be able to select:
      | Option | Duration |
      | 7 days | Last 7 days (default) |
      | 1 month | Last 30 days |
      | 6 months | Last 180 days |
      | 1 year | Last 365 days |
    When I change the time range to "<selected_range>"
    Then the chart should update to show only that period's data
    And the data points should be appropriately spaced
```

### UC-02-05: Mark Projects as Favorites

**As a** User with multiple projects  
**I want to** mark projects as favorites  
**So that** I can quickly identify important projects

#### Acceptance Criteria

```gherkin
Feature: Project Favorites
  As a user viewing test summaries
  I want to mark projects as favorites
  So that I can prioritize important projects

  Background:
    Given I am viewing the Test Summaries page
    And I can see project cards

  Scenario: Mark project as favorite
    Given I see a project "<project_name>" that is not favorited
    And the star icon shows only an outline (☆)
    When I click the star icon
    Then the star should fill and turn yellow (★)
    And the project should be marked as favorite
    And this preference should be saved immediately

  Scenario: Unmark favorite project
    Given I see a project "<project_name>" that is favorited
    And the star icon is filled and yellow (★)
    When I click the star icon
    Then the star should return to outline only (☆)
    And the project should no longer be marked as favorite

  Scenario: Favorites persist across sessions
    Given I have marked <favorite_count> projects as favorites
    When I log out and log back in
    And I navigate to Test Summaries
    Then the same <favorite_count> projects should show filled stars
    And my favorite selections should be preserved

  Scenario: Favorites are user-specific
    Given User A has marked "Project Alpha" as favorite
    When User B logs in and views Test Summaries
    Then User B should see "Project Alpha" without favorite marking
    And User B's favorite selections should be independent

  Scenario: Sort by favorites
    Given I have marked some projects as favorites
    When I apply sorting or filtering
    Then I should have an option to "Show favorites first"
    And selecting this should reorder the cards
    With favorited projects appearing at the top
```

### UC-02-06: Time Range Filtering

**As a** User analyzing test trends  
**I want to** filter data by time ranges  
**So that** I can focus on relevant time periods

#### Acceptance Criteria

```gherkin
Feature: Time Range Filtering
  As a user on Test Summaries page
  I want to filter test data by time range
  So that I can analyze specific periods

  Background:
    Given I am on the Test Summaries page
    And test data exists for multiple time periods

  Scenario: Default time range
    When the page loads
    Then the time range selector should show "7 days" as default
    And all metrics should reflect the last 7 days of data

  Scenario: Change time range
    Given I see the time range selector
    When I select "<time_range>" from the dropdown
    Then all project cards should update their metrics
    And only test runs within <time_range> should be counted
    And the following should update:
      | Metric | Update Behavior |
      | Total Tests | Sum of tests in period |
      | Test Runs | Count of runs in period |
      | Pass Rate | Recalculated for period |
      | Last Run | Most recent in period |

  Scenario: Time range affects all views
    Given I have selected "<time_range>"
    When I switch to Treemap View
    Then the treemap should only visualize data from <time_range>
    And suite sizes should reflect durations in that period
    And pass rates should be calculated for that period only

  Scenario: No data in selected range
    Given I select a time range with no test data
    Then project cards should show:
      | Metric | Display |
      | Total Tests | 0 |
      | Test Runs | 0 |
      | Pass Rate | No data |
      | Last Run | No runs in period |
    And cards should have a muted appearance
```

## Test Data Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `<team_name>` | Team the user belongs to | `frontend-team`, `backend-team` |
| `<total_projects>` | Total projects in database | `12` |
| `<team_projects>` | Projects belonging to user's team | `4` |
| `<visible_projects>` | Projects visible to current user | `4` or `12` (admin) |
| `<active_projects>` | Projects with recent activity | `10` |
| `<total_runs>` | Sum of all test runs visible | `1,247` |
| `<recent_runs>` | Test runs in last 24 hours | `23` |
| `<project_name>` | Name of a specific project | `E-Commerce Frontend` |
| `<suite_name>` | Name of a test suite | `authentication.spec.js` |
| `<time_range>` | Selected time period | `7 days`, `1 month`, `6 months`, `1 year` |
| `<run_count>` | Number of test runs | `50` |
| `<favorite_count>` | Number of favorited projects | `3` |

### Example Test Execution

```bash
# Running test summaries tests with Ginkgo
ginkgo -v \
  -team-name="frontend-team" \
  -total-projects=12 \
  -team-projects=4 \
  -project-name="E-Commerce Frontend" \
  -base-url="http://fern-platform.local:8080"
```

## Data Validation Rules

1. **Test Counts**: Must match exact counts from database test_runs and spec_runs tables
2. **Pass Rates**: Calculated as (passed_tests / total_tests) * 100, rounded to 1 decimal
3. **Duration**: Sum of all spec durations within a suite, displayed in milliseconds
4. **Time Filtering**: Uses test_run.start_time for filtering, inclusive of selected range
5. **Team Filtering**: Based on project.team field matching user's team groups

## Visual Design Specifications

### Card Flip Animation
- Duration: 600ms
- Transform: rotateY(180deg)
- Perspective: 1000px
- Timing: ease-in-out

### Treemap Colors (Pass Rate)
- 90-100%: `#10b981` (Green)
- 70-89%: Linear gradient green to yellow
- 50-69%: `#f59e0b` (Yellow)
- 30-49%: Linear gradient yellow to red
- 0-29%: `#ef4444` (Red)

### Favorite Star States
- Not favorited: `☆` (outline only)
- Favorited: `★` (filled, color: #f59e0b)

## Error Scenarios

1. **No Team Assignment**: Show helpful empty state with admin contact
2. **Database Connection Lost**: Show cached data with stale indicator
3. **Incomplete Test Data**: Handle gracefully, show available metrics
4. **Animation Performance**: Disable animations on low-end devices
5. **Large Dataset**: Implement pagination or virtualization for many projects

## Security Considerations

1. **Team Isolation**: Strict filtering based on authenticated user's teams
2. **Data Access**: GraphQL queries must respect team boundaries
3. **Admin Override**: Only platform admins can see cross-team data
4. **Favorite Storage**: User preferences stored per user ID, not shared

## Future Enhancements

1. **Custom Time Ranges**: Date picker for specific period selection
2. **Export Capabilities**: Download charts and data as PNG/CSV
3. **Comparison Mode**: Compare multiple projects side by side
4. **Alerting**: Notifications for pass rate drops
5. **AI Insights**: Automated trend analysis and recommendations
6. **Performance Benchmarks**: Compare against historical baselines