# Test Users and Credentials

After running `make deploy-all`, the following test users are automatically created in Keycloak for testing different permission levels:

## Available Test Users

All test users have the same password: **`test123`**

### Administrator

| Email | Password | Groups | Permissions |
|-------|----------|--------|-------------|
| `admin@fern.com` | `test123` | `/admin` | Full system access, can manage all projects and users |

### Fern Team Users

| Email | Password | Groups | Permissions |
|-------|----------|--------|-------------|
| `fern-manager@fern.com` | `test123` | `/fern`, `/manager` | Can create, view, edit, and delete projects in the "fern" team |
| `fern-user@fern.com` | `test123` | `/fern`, `/user` | Can only view projects in the "fern" team (read-only access) |

### Atmos Team Users

| Email | Password | Groups | Permissions |
|-------|----------|--------|-------------|
| `atmos-manager@fern.com` | `test123` | `/atmos`, `/manager` | Can create, view, edit, and delete projects in the "atmos" team |
| `atmos-user@fern.com` | `test123` | `/atmos`, `/user` | Can only view projects in the "atmos" team (read-only access) |

## Access URLs

After deployment, you can access the application at:
- **Application**: http://fern-platform.local:8080
- **Keycloak Admin**: http://keycloak:8080/admin (username: `admin`, password: `admin123`)

## Testing Different Permission Levels

### 1. Admin User (admin@fern.com)
- Can see all projects across all teams
- Has access to admin menu items (User Management, All Projects, System Settings)
- Can create projects for any team
- Can edit/delete any project

### 2. Team Manager (e.g., fern-manager@fern.com)
- Can see "Projects" menu item
- Can view all projects in their team
- Can create new projects (assigned to their team)
- Can edit/delete projects in their team
- Cannot see projects from other teams

### 3. Team User (e.g., fern-user@fern.com)
- Can see "Projects" menu item
- Can view all projects in their team
- Cannot create new projects (Create button is disabled)
- Cannot edit/delete projects (buttons are hidden)
- Cannot see projects from other teams

## Quick Test Scenarios

### Scenario 1: Test Read-Only Access
1. Login as `fern-user@fern.com` (password: `test123`)
2. Navigate to Projects menu
3. Verify you can see projects with team="fern"
4. Verify Create Project button is disabled
5. Verify no Edit/Delete buttons appear in the project table

### Scenario 2: Test Team Manager Access
1. Login as `fern-manager@fern.com` (password: `test123`)
2. Navigate to Projects menu
3. Create a new project (it will be assigned to "fern" team)
4. Verify you can edit and delete the project
5. Logout and login as `atmos-manager@fern.com`
6. Verify you cannot see the project created by fern-manager

### Scenario 3: Test Team Isolation
1. Login as `atmos-manager@fern.com` (password: `test123`)
2. Create a project (assigned to "atmos" team)
3. Logout and login as `fern-user@fern.com`
4. Verify you cannot see the atmos team project

### Scenario 4: Test Admin Override
1. Login as `admin@fern.com` (password: `test123`)
2. Navigate to All Projects (under admin menu)
3. Verify you can see projects from all teams
4. Verify you can edit/delete any project regardless of team

## Troubleshooting

### Cannot Login
- Ensure Keycloak is running: `kubectl get pods -n fern-platform`
- Check if realm was imported correctly: `kubectl logs -n fern-platform deployment/keycloak`
- Try accessing Keycloak directly at http://keycloak:8080

### Projects Not Visible
- Verify the user's groups in their token by checking browser developer tools
- Ensure projects have the correct team field set in the database
- Check GraphQL network requests for any errors

### Permissions Not Working
- Clear browser cache and cookies
- Logout and login again to refresh the token
- Verify group memberships in Keycloak admin console