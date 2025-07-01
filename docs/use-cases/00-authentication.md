# Authentication Use Cases

## Overview

Authentication is the entry point to Fern Platform. The system uses OAuth 2.0 with Keycloak as the identity provider, supporting single sign-on (SSO) across the organization. Users must authenticate before accessing any platform features.

## Actors

- **Anonymous User**: An unauthenticated user attempting to access the platform
- **Authenticated User**: A user who has successfully logged in via OAuth/Keycloak
- **System**: The Fern Platform backend and authentication services

## Prerequisites

- Keycloak server is running and accessible
- User has valid credentials in the organization's identity provider
- User account is active and not locked

## Use Cases

### UC-00-01: User Login

**As an** Anonymous User  
**I want to** log in to Fern Platform  
**So that** I can access test analytics and project data

#### Acceptance Criteria

```gherkin
Feature: User Authentication
  As an Anonymous User
  I want to authenticate with the platform
  So that I can access my team's test data

  Scenario: Accessing platform redirects to login
    Given I am not authenticated
    When I navigate to "http://fern-platform.local:8080"
    Then I should be redirected to the login page
    And the URL should be "http://fern-platform.local:8080/auth/login"
    And I should see a "Sign in with OAuth" button

  Scenario: Successful login with valid credentials
    Given I am on the login page at "/auth/login"
    When I click the "Sign in with OAuth" button
    Then I should be redirected to Keycloak login page
    And the URL should contain "keycloak:8080/realms/fern-platform/protocol/openid-connect/auth"
    And the URL should contain "client_id=fern-platform-web"
    And the URL should contain "redirect_uri=http%3A%2F%2Ffern-platform.local%3A8080%2Fauth%2Fcallback"
    When I enter my username "<username>" in the Keycloak form
    And I enter my password "<password>" in the Keycloak form
    And I click the "Sign In" button on Keycloak
    Then I should be redirected back to Fern Platform via "/auth/callback"
    And I should see the main dashboard at "/"
    And I should see my username in the navigation bar

  Scenario: Existing Keycloak session bypasses login
    Given I have an active Keycloak session
    When I navigate to "http://fern-platform.local:8080"
    Then I should be redirected to "/auth/login"
    When I click the "Sign in with OAuth" button
    Then I should be automatically redirected through Keycloak
    And I should land on the main dashboard at "/"
    Without needing to enter credentials

  Scenario: Failed login with invalid credentials
    Given I am on the login page at "/auth/login"
    When I click the "Sign in with OAuth" button
    Then I should be redirected to Keycloak login page
    When I enter an invalid username "invalid@example.com"
    And I enter an invalid password "wrongpassword"
    And I click the "Sign In" button on Keycloak
    Then I should see an error message "Invalid username or password" on Keycloak
    And I should remain on the Keycloak login page

  Scenario: Login with SSO provider
    Given I am on the login page
    And my organization uses SSO
    When I click "Sign in with <sso_provider>"
    Then I should be redirected to the SSO provider login
    When I complete SSO authentication
    Then I should be redirected back to Fern Platform
    And I should be logged in automatically

  Scenario: Session persistence after login
    Given I have successfully logged in
    When I close my browser
    And I reopen the browser within <session_timeout> minutes
    And I navigate to "http://fern-platform.local:8080"
    Then I should still be logged in
    And I should not see the login page

  Scenario: Deep link authentication
    Given I am not authenticated
    When I navigate to a specific project URL "http://fern-platform.local:8080/projects/<project_id>"
    Then I should be redirected to the login page
    And the return URL should be saved
    When I successfully log in
    Then I should be redirected to the originally requested project page
```

### UC-00-02: User Logout

**As an** Authenticated User  
**I want to** log out of Fern Platform  
**So that** I can securely end my session

#### Acceptance Criteria

```gherkin
Feature: User Logout
  As an Authenticated User
  I want to log out of the platform
  So that my session is securely terminated

  Background:
    Given I am logged in as "<username>"
    And I am on any page of the platform

  Scenario: Logout from user menu
    When I click on my username in the navigation bar
    Then I should see a dropdown menu
    And I should see a "Logout" option
    When I click "Logout"
    Then I should be redirected to the logout confirmation page

  Scenario: Confirm logout
    Given I clicked logout and see the confirmation page
    When I click "Confirm Logout"
    Then my session should be terminated
    And I should be redirected to the login page
    And I should see a message "You have been successfully logged out"

  Scenario: Cancel logout
    Given I clicked logout and see the confirmation page
    When I click "Cancel" or "Stay Logged In"
    Then I should be redirected back to the platform
    And I should still be logged in

  Scenario: Logout clears all sessions
    Given I am logged in on multiple browser tabs
    When I log out from one tab
    Then all other tabs should detect the logout
    And redirect to the login page when refreshed

  Scenario: Accessing protected resources after logout
    Given I have successfully logged out
    When I try to access a protected URL like "/projects"
    Then I should be redirected to the login page
    And I should not be able to access any project data

  Scenario: Browser back button after logout
    Given I have successfully logged out
    When I press the browser back button
    Then I should not see any cached protected content
    And I should remain on the login page or be redirected to it
```

### UC-00-03: Session Management

**As an** Authenticated User  
**I want** my session to be managed securely  
**So that** I don't have to log in repeatedly but my account remains secure

#### Acceptance Criteria

```gherkin
Feature: Session Management
  As an Authenticated User
  I want secure session management
  So that I have a good balance of security and convenience

  Background:
    Given I am logged in as "<username>"

  Scenario: Session timeout warning
    Given my session timeout is set to <session_timeout> minutes
    When I am inactive for <warning_time> minutes
    Then I should see a session timeout warning
    And the warning should say "Your session will expire in <remaining_time> minutes"
    And I should see options to "Stay Logged In" or "Log Out Now"

  Scenario: Extend session from warning
    Given I see the session timeout warning
    When I click "Stay Logged In"
    Then the warning should disappear
    And my session should be extended
    And the timeout timer should reset

  Scenario: Session expires during inactivity
    Given I have been inactive for <session_timeout> minutes
    When my session expires
    Then I should be automatically logged out
    And redirected to the login page
    And see a message "Your session has expired. Please log in again."

  Scenario: Active use prevents timeout
    Given I am actively using the platform
    When I perform actions like clicking, scrolling, or navigating
    Then my session timeout should be continuously reset
    And I should not see timeout warnings

  Scenario: Remember me functionality
    Given I am on the login page
    When I check the "Remember me" checkbox
    And I successfully log in
    Then my session should persist for <remember_me_duration> days
    And I should stay logged in even after closing the browser

  Scenario: Multiple device sessions
    Given I am logged in on my primary device
    When I log in from a different device
    Then both sessions should be active
    And I should see a list of active sessions in my account settings
    And I should be able to terminate other sessions
```

### UC-00-04: Authentication Error Handling

**As a** User  
**I want** clear error messages during authentication  
**So that** I can resolve issues and access the platform

#### Acceptance Criteria

```gherkin
Feature: Authentication Error Handling
  As a User
  I want helpful error messages
  So that I can troubleshoot authentication issues

  Scenario: Keycloak service unavailable
    Given Keycloak service is down
    When I navigate to the login page
    Then I should see a maintenance message
    And the message should say "Authentication service is temporarily unavailable"
    And I should see "Please try again in a few minutes"

  Scenario: OAuth callback error
    Given I am in the OAuth flow
    When the callback fails with an invalid state
    Then I should be redirected to the login page
    And I should see "Authentication failed. Please try logging in again"
    And the previous OAuth state should be cleared

  Scenario: Expired authentication token
    Given I have an expired authentication token
    When I try to access a protected resource
    Then I should be redirected to the login page
    And I should see "Your session has expired. Please log in again"

  Scenario: Account locked
    Given my account has been locked
    When I try to log in with correct credentials
    Then I should see "Your account has been locked. Please contact your administrator"
    And I should not be able to proceed with login

  Scenario: Account not found in any team
    Given I successfully authenticate with Keycloak
    But my account is not assigned to any team group
    When the system tries to log me in
    Then I should see "Your account is not assigned to any team. Please contact your administrator"
    And I should be logged out automatically
```

## Security Considerations

1. **OAuth 2.0 Flow**: All authentication flows through Keycloak's OAuth 2.0 implementation
2. **HTTPS Required**: All authentication traffic must be over HTTPS
3. **CSRF Protection**: Anti-CSRF tokens must be validated for all state-changing operations
4. **Session Security**: 
   - Sessions stored server-side only
   - Session IDs regenerated on login
   - Secure, HttpOnly, SameSite cookies
5. **Password Security**: Passwords never stored by Fern Platform (handled by Keycloak)
6. **Rate Limiting**: Login attempts should be rate-limited to prevent brute force
7. **Audit Logging**: All authentication events logged with timestamp, IP, and user agent

## Test Data Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `<username>` | User's login username | `john.doe@company.com`, `jdoe` |
| `<password>` | User's password | `SecurePass123!` |
| `<sso_provider>` | Name of SSO provider | `Google`, `Okta`, `Azure AD` |
| `<session_timeout>` | Minutes until session expires | `30`, `60`, `120` |
| `<warning_time>` | Minutes before showing timeout warning | `25`, `55`, `115` |
| `<remaining_time>` | Minutes left before timeout | `5`, `10` |
| `<remember_me_duration>` | Days to persist "remember me" session | `7`, `14`, `30` |
| `<team_name>` | Team the user belongs to | `frontend-team`, `backend-team` |
| `<project_id>` | UUID of a specific project | `550e8400-e29b-41d4-a716-446655440001` |

### Example Test Execution

```bash
# Running authentication tests with Ginkgo
ginkgo -v \
  -username="test.user@company.com" \
  -password="TestPass123!" \
  -sso-provider="Google" \
  -session-timeout=30 \
  -base-url="http://fern-platform.local:8080"
```

## Integration Points

### Keycloak Configuration

The platform expects Keycloak to be configured with:

1. **Realm**: `fern-platform`
2. **Client ID**: `fern-platform-web`
3. **Redirect URIs**: 
   - `http://fern-platform.local:8080/auth/callback`
   - `http://fern-platform.local:8080/*`
4. **Logout URL**: `http://fern-platform.local:8080/auth/logout`
5. **Required Groups**:
   - `manager` - For project management permissions
   - `admin` - For system administration
   - Team groups (e.g., `frontend-team`, `backend-team`)

### Session Storage

Sessions are stored in the backend with the following structure:
- Session ID (UUID)
- User ID
- Groups/Roles
- Creation timestamp
- Last activity timestamp
- Expiration timestamp

## Error Scenarios

1. **Keycloak Unavailable**: Show maintenance message, prevent login
2. **Invalid OAuth State**: Clear session, redirect to fresh login
3. **Group Sync Failure**: Allow login but limit to read-only access
4. **Session Store Failure**: Graceful degradation with temporary sessions
5. **Network Timeout**: Show retry options with helpful messaging

## Future Enhancements

1. **Multi-Factor Authentication**: Support for TOTP/SMS codes
2. **Passwordless Login**: Magic links via email
3. **Biometric Authentication**: TouchID/FaceID for mobile
4. **OAuth Provider Choice**: Support multiple OAuth providers
5. **Session Sharing**: Single session across multiple Fern services
6. **Adaptive Authentication**: Risk-based authentication requirements