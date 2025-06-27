# User Permissions Configuration

## Overview

Fern Platform uses a team-based role system with scope-based permissions. Users need specific group memberships in Keycloak to access different features.

## Permission Model

### Group Types

1. **Team Groups**: Define which projects a user can access (e.g., "fern", "qa", "dev")
2. **Role Groups**: Define what actions a user can perform
   - `admin`: Full system access
   - `manager`: Can create, edit, and delete projects in their teams
   - `user`: Can only view projects in their teams

### Access Rules

- **View Projects**: User must be in the team group (e.g., "fern")
- **Create/Edit/Delete Projects**: User must be in BOTH the team group AND "manager" role group
- **Admin Access**: User in "admin" role group can access everything

## Configuring fern-user@fern.com for View-Only Access

To give fern-user@fern.com view-only access to projects in the "fern" team:

### 1. Access Keycloak Admin Console

```bash
# Access at http://keycloak:8080/admin
# Or use the ingress: http://keycloak.local:8080/admin
```

### 2. Navigate to the User

1. Go to **Users** in the left menu
2. Search for `fern-user@fern.com`
3. Click on the user to open their details

### 3. Configure Group Memberships

1. Click on the **Groups** tab
2. Click **Join Group** and add the following groups:
   - `/fern` - Team group for accessing "fern" team projects
   - `/user` - Role group for view-only permissions

**Important**: Do NOT add the `/manager` group, as this would grant edit/delete permissions.

### 4. Verify Permissions

After configuration, fern-user@fern.com will have:
- ✅ View access to all projects in the "fern" team
- ❌ Cannot create new projects
- ❌ Cannot edit existing projects
- ❌ Cannot delete projects

### 5. Test the Configuration

1. Log out if currently logged in as fern-user@fern.com
2. Log in again to refresh the token with new group memberships
3. Navigate to the Projects page
4. Verify that:
   - Projects with team="fern" are visible
   - Edit and Delete buttons are disabled or hidden
   - Create Project button is disabled or hidden

## Troubleshooting

### User Can't See Projects
- Verify the user is in the correct team group (e.g., "/fern")
- Check that projects have the correct team assigned
- Ensure the user has logged out and back in after group changes

### User Can Still Edit/Delete
- Verify the user is NOT in the "/manager" group
- Check for any custom scopes assigned to the user
- Clear browser cache and re-login

### Groups Not Showing in Token
- Check Keycloak client configuration has group membership mapper
- Verify the mapper includes the groups in the token
- Token introspection should show: `"groups": ["/fern", "/user"]`

## Example Group Configurations

### Read-Only User (fern-user@fern.com)
```
Groups: /fern, /user
Result: Can view "fern" team projects only
```

### Team Manager
```
Groups: /fern, /manager
Result: Can view, create, edit, delete "fern" team projects
```

### Multi-Team User
```
Groups: /fern, /qa, /user
Result: Can view projects in both "fern" and "qa" teams
```

### Admin User
```
Groups: /admin
Result: Full access to all projects and features
```