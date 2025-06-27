# Authorization System for Fern Platform

This document describes the authorization system in Fern Platform, which combines group-based and scope-based permissions for flexible access control.

## Overview

Fern Platform uses a hybrid authorization model with three levels of permissions:

1. **Roles**: System-wide access levels (admin vs user)
2. **Groups**: Team-based permissions that provide default access
3. **Scopes**: Fine-grained permissions for specific resources

## Functionality-Based Permissions

### Project Management

| Functionality | Required Permission | Who Has Access |
|---------------|-------------------|----------------|
| View all projects | Authenticated user | All logged-in users |
| View team projects | Group: `{team}-users` or `{team}-managers` | Team members |
| Create project | Group: `{team}-managers` OR Scope: `project:create:{team}` | Team managers |
| Edit project | Group: `{team}-managers` (for team projects) OR Scope: `project:write:{projectId}` | Project owners/managers |
| Delete project | Group: `{team}-managers` (for team projects) OR Scope: `project:delete:{projectId}` | Project owners/managers |
| Transfer project | Admin role OR Scope: `project:*:{projectId}` | Admins only |

### Test Results Access

| Functionality | Required Permission | Who Has Access |
|---------------|-------------------|----------------|
| View test results | Group: `{team}-users` or `{team}-managers` | Team members |
| Submit test results | API Key (no user auth required) | CI/CD systems |
| Delete test runs | Group: `{team}-managers` OR Admin role | Team managers |

### User Management

| Functionality | Required Permission | Who Has Access |
|---------------|-------------------|----------------|
| View users | Admin role | Platform admins |
| Grant scopes | Admin role | Platform admins |
| Manage groups | Keycloak admin access | Keycloak admins |

### Analytics & Reports

| Functionality | Required Permission | Who Has Access |
|---------------|-------------------|----------------|
| View team analytics | Group: `{team}-users` or `{team}-managers` | Team members |
| View global analytics | Admin role | Platform admins |
| Export reports | Group: `{team}-managers` | Team managers |

## Permission Hierarchy

### 1. Admin Role
Users with the `admin` role have full access to all resources and operations.

### 2. Team-Based Groups
Users are organized into teams through Keycloak groups:
- `fern-managers`: Can create/edit/delete projects for the Fern team
- `fern-users`: Can view projects for the Fern team
- `atmos-managers`: Can create/edit/delete projects for the Atmos team
- `atmos-users`: Can view projects for the Atmos team

### 3. Scope-Based Permissions
Scopes provide fine-grained control over specific resources and actions.

## Scope Format

Scopes follow the pattern: `resource:action:target`

For project management:
- `project:action:identifier`
- `project:action:team:identifier`

### Wildcards
The `*` wildcard can be used in any position:
- `project:*:project-123`: All actions on project-123
- `project:write:*`: Write access to all projects
- `project:*:fern:*`: All actions on all Fern team projects

## Project Management Scopes

### Actions
- `create`: Create new projects
- `write`: Update existing projects
- `delete`: Delete projects
- `read`: View projects (implicit for authenticated users)
- `*`: All actions

### Examples

#### Project-Specific Scopes
```
project:write:project-123     # Update project-123
project:delete:project-123    # Delete project-123
project:*:project-123         # All actions on project-123
```

#### Team-Based Scopes
```
project:create:fern           # Create projects for Fern team
project:write:fern:*          # Update any Fern team project
project:*:atmos:*             # All actions on Atmos team projects
```

#### Global Scopes
```
project:create:*              # Create projects in any team
project:*:*                   # All project management actions
```

## Database Schema

### UserScope Table
Stores scope grants for users:
```sql
CREATE TABLE user_scopes (
    id SERIAL PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    scope VARCHAR(255) NOT NULL,
    granted_by VARCHAR(255),
    expires_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE KEY idx_user_scope (user_id, scope)
);
```

### ProjectPermission Table
Stores explicit project permissions:
```sql
CREATE TABLE project_permissions (
    id SERIAL PRIMARY KEY,
    project_id VARCHAR(255) NOT NULL,
    user_id VARCHAR(255) NOT NULL,
    permission VARCHAR(50) NOT NULL, -- read, write, delete, admin
    granted_by VARCHAR(255),
    expires_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE KEY idx_project_user_perm (project_id, user_id, permission)
);
```

## Implementation Examples

### Checking Project Management Permission

```go
func (m *OAuthMiddleware) CanManageProject(c *gin.Context, projectID string, action string) bool {
    user, exists := GetOAuthUser(c)
    if !exists {
        return false
    }
    
    // Admin can do anything
    if user.Role == string(database.RoleAdmin) {
        return true
    }
    
    // Get project to check team
    var project database.ProjectDetails
    if err := m.db.Where("project_id = ?", projectID).First(&project).Error; err != nil {
        return false
    }
    
    // Check scopes
    requiredScopes := []string{
        fmt.Sprintf("project:%s:%s", action, projectID),        // Specific project
        fmt.Sprintf("project:%s:%s:*", action, project.Team),   // Team wildcard
        fmt.Sprintf("project:*:%s", projectID),                 // All actions on project
        fmt.Sprintf("project:*:%s:*", project.Team),            // All actions on team
        "project:*:*",                                           // Global project admin
    }
    
    userScopes := GetUserScopes(c)
    for _, scope := range userScopes {
        for _, required := range requiredScopes {
            if matchScope(scope, required) {
                return true
            }
        }
    }
    
    // Check explicit project permissions
    var perm database.ProjectPermission
    now := time.Now()
    err := m.db.Where("project_id = ? AND user_id = ? AND permission IN ? AND (expires_at IS NULL OR expires_at > ?)", 
        projectID, user.UserID, []string{action, "admin"}, now).First(&perm).Error
    
    return err == nil
}
```

### Granting Scopes

```go
// Grant a user write access to a specific project
scope := database.UserScope{
    UserID:    "user-123",
    Scope:     "project:write:project-456",
    GrantedBy: "admin-user",
    ExpiresAt: nil, // No expiration
}
db.Create(&scope)

// Grant a manager access to all team projects
scope := database.UserScope{
    UserID:    "manager-123",
    Scope:     "project:*:fern:*",
    GrantedBy: "admin-user",
}
db.Create(&scope)
```

## GraphQL Integration

### Project Type
The GraphQL Project type includes a `canManage` field that is resolved based on the user's permissions:

```graphql
type Project {
    id: ID!
    projectId: String!
    name: String!
    team: String
    canManage: Boolean!  # Resolved based on user's scopes
    # ... other fields
}
```

### Mutations with Permission Checks

```graphql
mutation CreateProject($input: CreateProjectInput!) {
    createProject(input: $input) {
        id
        projectId
        name
        team
        canManage
    }
}
```

The resolver checks:
1. If the user is an admin (can create in any team)
2. If the user has a relevant scope (e.g., `project:create:fern`)
3. If the team is not specified, uses the user's primary team

## UI Integration

The React UI components check permissions at multiple levels:

1. **Route Level**: Only managers see the Project Management menu item
2. **Component Level**: Edit/Delete buttons only shown if `project.canManage` is true
3. **Action Level**: Mutations will fail server-side if permissions are insufficient

```javascript
// Check if user is a manager
const isManager = user?.groups?.some(group => 
    group.endsWith('-managers') || group === 'admin'
) || user?.role === 'admin';

// Show management UI only for managers
{isManager && (
    <Link to="/projects/manage">
        <i className="fas fa-cog"></i> Manage Projects
    </Link>
)}

// Show edit/delete buttons based on canManage
{project.canManage && (
    <>
        <button onClick={() => handleEdit(project)}>Edit</button>
        <button onClick={() => handleDelete(project)}>Delete</button>
    </>
)}
```

## Security Considerations

1. **Principle of Least Privilege**: Users should only be granted the minimum scopes necessary
2. **Scope Expiration**: Use `expires_at` for temporary permissions
3. **Audit Trail**: The `granted_by` field tracks who granted each permission
4. **Double Authorization**: Permissions are checked both client-side (UI) and server-side (API)

## How the Hybrid System Works

The permission system checks access in the following order:

1. **Admin Role**: If the user has the `admin` role, they have full access
2. **Group-Based Access**: Team membership provides default permissions:
   - `{team}-managers` groups can create/update/delete their team's projects
   - `{team}-users` groups can view their team's projects
3. **Scope-Based Access**: Specific scopes override or extend group permissions
4. **Explicit Permissions**: Database-stored permissions for individual project access

### Example Permission Flow

When a user tries to update a project:

```go
// 1. Check if admin
if user.Role == "admin" {
    return true // Allow
}

// 2. Check if user has relevant scopes
scopes := []string{
    "project:write:project-123",     // Specific project
    "project:write:fern:*",          // All fern team projects
    "project:*:project-123",         // All actions on specific project
    "project:*:*",                   // Global project admin
}

// 3. If no scopes match, check if user is in team-managers group
if user.Groups.Contains("fern-managers") && project.Team == "fern" {
    return true // Allow based on group membership
}

// 4. Finally, check explicit database permissions
if hasProjectPermission(user.ID, project.ID, "write") {
    return true // Allow
}

return false // Deny
```

## Best Practices

1. **Use Team Scopes**: Prefer team-based scopes over individual project scopes when possible
2. **Regular Audits**: Periodically review granted scopes and remove unnecessary ones
3. **Explicit Denies**: The system uses allow-only; there are no explicit deny scopes
4. **Scope Naming**: Use consistent, descriptive scope patterns
5. **Documentation**: Document why specific scopes were granted in your ticketing system