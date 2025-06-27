# Scope-Based Permissions: Common Scenarios and Examples

This guide provides practical examples of using the scope-based permission system in Fern Platform.

## Common Scenarios

### 1. Cross-Team Collaboration

**Scenario**: A developer from the Atmos team needs to update a Fern team project for a joint initiative.

**Solution**: Grant temporary write access to the specific project.

```sql
-- Grant write access to specific project with 30-day expiration
INSERT INTO user_scopes (user_id, scope, granted_by, expires_at)
VALUES (
    'atmos-dev-123',
    'project:write:fern-api-gateway',
    'fern-manager-456',
    NOW() + INTERVAL '30 days'
);
```

### 2. External Contractor Access

**Scenario**: An external contractor needs read-only access to specific projects.

**Solution**: Grant read permissions to specific projects only.

```sql
-- Grant read access to multiple specific projects
INSERT INTO user_scopes (user_id, scope, granted_by, expires_at)
VALUES 
    ('contractor-789', 'project:read:project-abc', 'admin', NOW() + INTERVAL '90 days'),
    ('contractor-789', 'project:read:project-xyz', 'admin', NOW() + INTERVAL '90 days');
```

### 3. Team Lead Permissions

**Scenario**: A team lead needs to manage all projects for their team plus create projects in a shared space.

**Solution**: Combine team-wide management with specific create permissions.

```sql
-- Grant full team management plus shared space creation
INSERT INTO user_scopes (user_id, scope, granted_by)
VALUES 
    ('lead-111', 'project:*:fern:*', 'admin'),           -- All Fern team projects
    ('lead-111', 'project:create:shared', 'admin');      -- Create in shared team
```

### 4. Project Ownership Transfer

**Scenario**: Transfer project ownership from one team to another, updating permissions accordingly.

```go
// In your service code
func TransferProjectOwnership(projectID, fromTeam, toTeam, newOwnerID string) error {
    // Start transaction
    tx := db.Begin()
    
    // Update project team
    if err := tx.Model(&ProjectDetails{}).
        Where("project_id = ?", projectID).
        Update("team", toTeam).Error; err != nil {
        tx.Rollback()
        return err
    }
    
    // Grant new owner full access
    newScope := UserScope{
        UserID:    newOwnerID,
        Scope:     fmt.Sprintf("project:*:%s", projectID),
        GrantedBy: "system",
    }
    if err := tx.Create(&newScope).Error; err != nil {
        tx.Rollback()
        return err
    }
    
    // Optionally revoke old team's access
    if err := tx.Where("scope LIKE ?", fmt.Sprintf("project:%%:%s:%%", fromTeam)).
        Delete(&UserScope{}).Error; err != nil {
        tx.Rollback()
        return err
    }
    
    return tx.Commit().Error
}
```

### 5. Audit Reviewer Access

**Scenario**: An auditor needs read-only access to all projects for compliance review.

**Solution**: Grant global read access with expiration.

```sql
-- Grant read access to all projects for audit period
INSERT INTO user_scopes (user_id, scope, granted_by, expires_at)
VALUES (
    'auditor-222',
    'project:read:*',
    'compliance-admin',
    NOW() + INTERVAL '14 days'
);
```

## GraphQL Usage Examples

### Checking User Permissions

```graphql
query GetMyProjects {
    projects {
        edges {
            node {
                id
                projectId
                name
                team
                canManage  # Will be true if user has write/delete permissions
            }
        }
    }
}
```

### Creating a Project with Team Assignment

```graphql
mutation CreateTeamProject {
    createProject(input: {
        projectId: "new-analytics-dashboard"
        name: "Analytics Dashboard"
        description: "Team analytics and reporting dashboard"
        team: "fern"
        defaultBranch: "main"
    }) {
        id
        projectId
        name
        team
        canManage
    }
}
```

### Updating Project with Permission Check

```graphql
mutation UpdateProject($id: ID!, $input: UpdateProjectInput!) {
    updateProject(id: $id, input: $input) {
        id
        name
        team
        canManage
    }
}
```

The mutation will fail with "insufficient permissions" if the user lacks the necessary scope.

## API Integration Examples

### REST API with Scope Checks

```go
// API endpoint that checks scopes
func (h *Handler) UpdateProjectHandler(c *gin.Context) {
    projectID := c.Param("id")
    
    // Use the middleware function to check permissions
    if !h.oauth.CanManageProject(c, projectID, "write") {
        c.JSON(403, gin.H{"error": "insufficient permissions"})
        return
    }
    
    // Proceed with update...
}
```

### Programmatic Scope Checking

```go
// Service layer example
func (s *ProjectService) CanUserManageProject(userID, projectID string) (bool, error) {
    // Get user with scopes
    var user User
    if err := s.db.Preload("UserScopes").Where("user_id = ?", userID).First(&user).Error; err != nil {
        return false, err
    }
    
    // Check if admin
    if user.Role == "admin" {
        return true, nil
    }
    
    // Get project details
    var project ProjectDetails
    if err := s.db.Where("project_id = ?", projectID).First(&project).Error; err != nil {
        return false, err
    }
    
    // Check scopes
    requiredScopes := []string{
        fmt.Sprintf("project:write:%s", projectID),
        fmt.Sprintf("project:*:%s", projectID),
        fmt.Sprintf("project:write:%s:*", project.Team),
        fmt.Sprintf("project:*:%s:*", project.Team),
        "project:*:*",
    }
    
    for _, userScope := range user.UserScopes {
        for _, required := range requiredScopes {
            if matchScope(userScope.Scope, required) {
                return true, nil
            }
        }
    }
    
    return false, nil
}
```

## Scope Management UI Examples

### React Component for Scope Management

```javascript
const ScopeManager = ({ userId }) => {
    const [scopes, setScopes] = useState([]);
    const [newScope, setNewScope] = useState('');
    
    const grantScope = async () => {
        const response = await fetch('/api/admin/scopes', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                userId,
                scope: newScope,
                expiresIn: 30 * 24 * 60 * 60 // 30 days in seconds
            })
        });
        
        if (response.ok) {
            const granted = await response.json();
            setScopes([...scopes, granted]);
            setNewScope('');
        }
    };
    
    const revokeScope = async (scopeId) => {
        const response = await fetch(`/api/admin/scopes/${scopeId}`, {
            method: 'DELETE'
        });
        
        if (response.ok) {
            setScopes(scopes.filter(s => s.id !== scopeId));
        }
    };
    
    return (
        <div>
            <h3>User Scopes</h3>
            <div>
                <input
                    type="text"
                    value={newScope}
                    onChange={(e) => setNewScope(e.target.value)}
                    placeholder="project:write:project-123"
                />
                <button onClick={grantScope}>Grant Scope</button>
            </div>
            
            <ul>
                {scopes.map(scope => (
                    <li key={scope.id}>
                        {scope.scope}
                        {scope.expiresAt && ` (expires: ${new Date(scope.expiresAt).toLocaleDateString()})`}
                        <button onClick={() => revokeScope(scope.id)}>Revoke</button>
                    </li>
                ))}
            </ul>
        </div>
    );
};
```

## Troubleshooting Common Issues

### Issue: User Can't See Projects They Should Access

**Check**:
1. Verify user's groups in Keycloak
2. Check UserScopes table for relevant scopes
3. Ensure scopes haven't expired
4. Verify project team assignment is correct

```sql
-- Debug query to see all user's access
SELECT 
    u.user_id,
    u.email,
    u.role,
    array_agg(DISTINCT ug.group_name) as groups,
    array_agg(DISTINCT us.scope) as scopes
FROM users u
LEFT JOIN user_groups ug ON u.user_id = ug.user_id
LEFT JOIN user_scopes us ON u.user_id = us.user_id
WHERE u.email = 'user@example.com'
GROUP BY u.user_id, u.email, u.role;
```

### Issue: Permission Denied Despite Having Scope

**Check**:
1. Scope format is correct (no typos)
2. Wildcard matching is working
3. No case sensitivity issues
4. Transaction committed if scope was just granted

```go
// Test scope matching
fmt.Println(matchScope("project:*:fern:*", "project:write:fern:api"))      // Should be true
fmt.Println(matchScope("project:write:*", "project:write:project-123"))    // Should be true
fmt.Println(matchScope("project:write:abc", "project:read:abc"))          // Should be false
```

## Performance Considerations

### Caching User Scopes

```go
// Simple in-memory cache for user scopes
type ScopeCache struct {
    cache map[string][]string
    mutex sync.RWMutex
    ttl   time.Duration
}

func (c *ScopeCache) GetUserScopes(userID string) ([]string, bool) {
    c.mutex.RLock()
    defer c.mutex.RUnlock()
    
    scopes, exists := c.cache[userID]
    return scopes, exists
}

func (c *ScopeCache) SetUserScopes(userID string, scopes []string) {
    c.mutex.Lock()
    defer c.mutex.Unlock()
    
    c.cache[userID] = scopes
    
    // Set expiration
    go func() {
        time.Sleep(c.ttl)
        c.mutex.Lock()
        delete(c.cache, userID)
        c.mutex.Unlock()
    }()
}
```

### Database Indexes

Ensure proper indexes for performance:

```sql
-- Indexes for scope lookups
CREATE INDEX idx_user_scopes_user_id ON user_scopes(user_id);
CREATE INDEX idx_user_scopes_scope ON user_scopes(scope);
CREATE INDEX idx_user_scopes_expires ON user_scopes(expires_at) WHERE expires_at IS NOT NULL;

-- Indexes for project permissions
CREATE INDEX idx_project_perms_user ON project_permissions(user_id);
CREATE INDEX idx_project_perms_project ON project_permissions(project_id);
```