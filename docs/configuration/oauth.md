# OAuth Authentication and Authorization Guide

This guide explains how to set up OAuth authentication for the Fern Platform using any OAuth 2.0/OpenID Connect provider, with a focus on the hybrid group and scope-based authorization system.

## Overview

The Fern Platform uses a flexible authentication and authorization system that supports:
- **Provider-agnostic OAuth 2.0/OIDC** - Works with any OAuth provider (Keycloak, Auth0, Okta, Google, etc.)
- **Team-based access control** - Projects are organized by teams with isolated access
- **Role-based permissions** - Configurable role groups (admin, manager, user)
- **Scope-based fine-grained access** - Additional permissions can be granted via scopes
- **Session management** - Secure session handling with proper logout flow

## Authentication Flow

### 1. Initial Authentication
1. User accesses Fern Platform
2. If not authenticated, redirected to OAuth provider
3. User logs in with their credentials
4. OAuth provider returns user to Fern Platform with authorization code
5. Fern Platform exchanges code for tokens
6. User information and groups are extracted from tokens
7. Session is created with user data and permissions

### 2. Authorization Model

Fern Platform uses a **hybrid authorization model** that combines:

#### Group-Based Authorization
Users are members of two types of groups:
- **Team Groups**: Represent which teams a user belongs to (e.g., `fern`, `atmos`, `engineering`)
- **Role Groups**: Represent what role a user has (e.g., `admin`, `manager`, `user`)

The combination determines base permissions:
- `fern` + `manager` = Can create/update/delete projects in the Fern team
- `fern` + `user` = Can only view projects in the Fern team
- `admin` (role) = Full access to all resources regardless of team

#### Scope-Based Authorization
For fine-grained access control beyond group membership:
- Format: `resource:action:target` (e.g., `project:write:project-123`)
- Can grant temporary or permanent additional permissions
- Useful for cross-team collaboration or specific project access

## Configuration Options

### Basic OAuth Configuration

```yaml
auth:
  enabled: true
  oauth:
    enabled: true
    clientId: "your-client-id"
    clientSecret: "your-client-secret"
    redirectUrl: "https://your-domain.com/auth/callback"
    scopes: ["openid", "profile", "email", "groups"]
    
    # OAuth endpoints (required)
    authUrl: "https://provider.com/oauth2/authorize"
    tokenUrl: "https://provider.com/oauth2/token"
    userInfoUrl: "https://provider.com/oauth2/userinfo"
    jwksUrl: "https://provider.com/.well-known/jwks.json"
    logoutUrl: "https://provider.com/oauth2/logout"  # Optional
    
    # Admin configuration
    adminUsers: ["admin@company.com"]  # Users who are always admins
    adminGroups: ["platform-admins"]   # Groups that grant admin access
```

### Role Group Configuration

You can customize the names of role groups to match your organization:

```yaml
auth:
  oauth:
    # Configurable role group names (with defaults)
    adminGroupName: "admin"      # Default: "admin"
    managerGroupName: "manager"  # Default: "manager"
    userGroupName: "user"        # Default: "user"
    
    # Or use custom names that match your identity provider
    adminGroupName: "platform-administrators"
    managerGroupName: "team-leads"
    userGroupName: "developers"
```

Environment variables:
```bash
export OAUTH_ADMIN_GROUP_NAME=platform-administrators
export OAUTH_MANAGER_GROUP_NAME=team-leads
export OAUTH_USER_GROUP_NAME=developers
```

### Token Field Mapping

Configure how to extract user information from OAuth tokens:

```yaml
auth:
  oauth:
    # Field mappings in token/userinfo
    userIdField: "sub"        # User ID field (default: "sub")
    emailField: "email"       # Email field (default: "email")
    nameField: "name"         # Display name field (default: "name")
    groupsField: "groups"     # Groups array field (default: "groups")
    rolesField: "roles"       # Roles field (optional)
```

## Setting Up Teams in Your Identity Provider

### Team Structure

Teams in Fern Platform are simply group names in your identity provider. The platform distinguishes between:
- **Role groups**: The configurable groups (`admin`, `manager`, `user` by default)
- **Team groups**: Any other group names become team identifiers

### Example: Setting Up in Keycloak

1. **Create Groups** in Keycloak:
   ```
   /admin           (role group - platform administrators)
   /manager         (role group - can manage team resources)
   /user            (role group - read-only access)
   /fern            (team group)
   /atmos           (team group)
   /engineering     (team group)
   /marketing       (team group)
   ```

2. **Assign Users to Groups**:
   - User needs ONE role group AND one or more team groups
   - Example: John is in groups `manager` and `fern` → Can manage Fern team projects
   - Example: Jane is in groups `user`, `fern`, and `atmos` → Can view projects from both teams

3. **Ensure Groups are in Tokens**:
   - Configure your OAuth client to include groups in tokens
   - In Keycloak: Add "groups" mapper to client

### Example: Setting Up in Okta

1. **Create Groups**:
   - `platform-admins` (if using custom role names)
   - `team-leads` 
   - `developers`
   - `team-fern`
   - `team-atmos`

2. **Configure App to Include Groups**:
   - In Okta app settings, add groups claim
   - Set groups claim filter (e.g., Regex: .*)

### Example: Setting Up in Auth0

1. **Create Roles and Groups** via Auth0 Dashboard
2. **Add Groups to Tokens** via Auth0 Rules:
   ```javascript
   function addGroupsToToken(user, context, callback) {
     context.idToken.groups = user.groups || [];
     context.accessToken.groups = user.groups || [];
     callback(null, user, context);
   }
   ```

## Permission Model

### How Permissions Are Checked

When a user tries to perform an action, the system checks in this order:

1. **Admin Role Check**
   - Is the user in the admin role group? → Allow all actions
   - Is the user in `adminUsers` list? → Allow all actions

2. **Group-Based Check**
   - Does the user have the required team + role combination?
   - Example: Creating a project in team "fern" requires: `fern` + `manager` groups

3. **Scope-Based Check**
   - Does the user have a matching scope?
   - Scopes can override group-based permissions

4. **Explicit Database Permissions**
   - Check project_permissions table for specific access

### Scope Format

Scopes follow the pattern: `resource:action:target`

**Examples:**
```
project:create:fern         # Can create projects in Fern team
project:write:project-123   # Can update specific project
project:delete:project-456  # Can delete specific project
project:*:project-789      # All actions on specific project
project:*:fern:*          # All actions on all Fern team projects
project:*:*               # Global project admin
```

### Team Visibility

- Users only see projects from their teams
- Admin users see all projects regardless of team
- No cross-team visibility unless explicitly granted via scopes

## Practical Examples

### Example 1: Basic Team Setup

Your company has two development teams: Frontend and Backend

1. **Create groups in your identity provider:**
   - Role groups: `admin`, `manager`, `user`
   - Team groups: `frontend`, `backend`

2. **Assign users:**
   - Alice: `frontend` + `manager` → Manages frontend projects
   - Bob: `frontend` + `user` → Views frontend projects
   - Carol: `backend` + `manager` → Manages backend projects
   - Dave: `admin` → Manages everything

### Example 2: Cross-Team Collaboration

Bob needs temporary write access to a backend project:

1. **Admin grants scope** (via API or database):
   ```sql
   INSERT INTO user_scopes (user_id, scope, granted_by, expires_at)
   VALUES ('bob-user-id', 'project:write:backend-project-123', 'admin-id', '2024-12-31');
   ```

2. **Bob can now:**
   - View all frontend projects (via groups)
   - Edit the specific backend project (via scope)
   - Cannot see other backend projects

### Example 3: Custom Role Names

Your organization uses different terminology:

```yaml
auth:
  oauth:
    adminGroupName: "platform-owner"
    managerGroupName: "team-lead"
    userGroupName: "contributor"
```

Users in `platform-owner` group have admin access, `team-lead` can manage their team's projects, etc.

## GraphQL API Usage

### Check Current User Permissions

```graphql
query MyPermissions {
  currentUser {
    id
    email
    role
    groups
  }
  
  projects {
    edges {
      node {
        projectId
        name
        team
        canManage  # true if user can edit/delete this project
      }
    }
  }
}
```

### Create Project (Managers Only)

```graphql
mutation CreateProject {
  createProject(input: {
    name: "New Frontend App"
    description: "React application"
    team: "frontend"  # Optional, defaults to user's team
  }) {
    projectId
    name
    team
  }
}
```

## Provider-Specific Examples

### Keycloak Configuration

```yaml
auth:
  oauth:
    enabled: true
    clientId: "fern-platform"
    clientSecret: "your-secret"
    redirectUrl: "https://app.company.com/auth/callback"
    
    authUrl: "https://keycloak.company.com/realms/company/protocol/openid-connect/auth"
    tokenUrl: "https://keycloak.company.com/realms/company/protocol/openid-connect/token"
    userInfoUrl: "https://keycloak.company.com/realms/company/protocol/openid-connect/userinfo"
    jwksUrl: "https://keycloak.company.com/realms/company/protocol/openid-connect/certs"
    logoutUrl: "https://keycloak.company.com/realms/company/protocol/openid-connect/logout"
```

### Auth0 Configuration

```yaml
auth:
  oauth:
    enabled: true
    clientId: "your-auth0-client-id"
    clientSecret: "your-auth0-secret"
    redirectUrl: "https://app.company.com/auth/callback"
    
    authUrl: "https://company.auth0.com/authorize"
    tokenUrl: "https://company.auth0.com/oauth/token"
    userInfoUrl: "https://company.auth0.com/userinfo"
    jwksUrl: "https://company.auth0.com/.well-known/jwks.json"
```

### Okta Configuration

```yaml
auth:
  oauth:
    enabled: true
    clientId: "your-okta-client-id"
    clientSecret: "your-okta-secret"
    redirectUrl: "https://app.company.com/auth/callback"
    
    authUrl: "https://company.okta.com/oauth2/default/v1/authorize"
    tokenUrl: "https://company.okta.com/oauth2/default/v1/token"
    userInfoUrl: "https://company.okta.com/oauth2/default/v1/userinfo"
    jwksUrl: "https://company.okta.com/oauth2/default/v1/keys"
    
    groupsField: "groups"  # Okta includes groups in tokens
```

## Troubleshooting

### User Can't Create Projects

1. **Check group membership:**
   ```graphql
   query CheckGroups {
     currentUser {
       email
       groups
     }
   }
   ```

2. **Verify they have BOTH:**
   - A team group (e.g., `fern`)
   - The manager role group (e.g., `manager`)

3. **Check role group configuration matches:**
   ```yaml
   # Your config
   managerGroupName: "manager"  # Must match actual group name
   ```

### User Sees No Projects

1. **Verify team membership:**
   - User must be in at least one team group
   - Team group must match project's team field

2. **Check if projects exist for their team:**
   ```graphql
   query AllProjects {
     projects {
       edges {
         node {
           name
           team
         }
       }
     }
   }
   ```

### Groups Not Appearing

1. **Verify OAuth scope includes groups:**
   ```yaml
   scopes: ["openid", "profile", "email", "groups"]
   ```

2. **Check token contents:**
   - Enable debug logging
   - Check if groups are in the token
   - Verify `groupsField` configuration matches your provider

3. **Provider-specific checks:**
   - Keycloak: Ensure groups mapper is added to client
   - Okta: Verify groups claim is configured
   - Auth0: Check if groups rule is active

## Security Best Practices

1. **Use HTTPS in Production**
   - OAuth requires secure connections
   - Protect tokens in transit

2. **Secure Client Secrets**
   - Use environment variables or secrets management
   - Never commit secrets to version control

3. **Configure Token Expiration**
   - Set appropriate session timeouts
   - Implement token refresh if needed

4. **Audit Permissions**
   - Regularly review admin users
   - Monitor scope assignments
   - Check for unused permissions

5. **Principle of Least Privilege**
   - Users should only have necessary permissions
   - Use time-limited scopes for temporary access
   - Prefer group-based permissions over individual scopes

## Migration from Previous Versions

If upgrading from a version using `{team}-managers` and `{team}-users` groups:

1. **Update group structure** in your identity provider:
   - Old: `fern-managers`, `fern-users`
   - New: `fern` (team) + `manager`/`user` (role)

2. **Migrate users** to new groups:
   - User in `fern-managers` → Add to `fern` and `manager`
   - User in `atmos-users` → Add to `atmos` and `user`

3. **Update configuration** if using custom role names:
   ```yaml
   auth:
     oauth:
       managerGroupName: "managers"  # If keeping suffix
       userGroupName: "users"
   ```

This completes the comprehensive OAuth documentation with the current authentication and authorization model.