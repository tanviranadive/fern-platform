# OAuth Setup and Testing Guide

This guide explains how to set up OAuth authentication for the Fern Platform using any OAuth 2.0/OpenID Connect provider, with Keycloak as an example for testing.

## Overview

The Fern Platform supports configurable OAuth 2.0 authentication with:
- **Provider-agnostic configuration** - works with any OAuth 2.0/OIDC provider
- **Role-based access control** - admin vs regular user roles
- **Project-level permissions** - granular access control per project
- **Session management** - secure session handling and logout

## Admin Functions

### System-Level Admin Functions
Administrators have access to these platform-wide capabilities:

1. **User Management**
   - View all users (`GET /api/v1/admin/users`)
   - Update user roles (`PUT /api/v1/admin/users/:userId/role`)
   - Suspend/activate users (`POST /api/v1/admin/users/:userId/suspend`)
   - Delete user accounts (`DELETE /api/v1/admin/users/:userId`)

2. **Project Management**
   - Create/update/delete projects (`POST/PUT/DELETE /api/v1/admin/projects`)
   - Activate/deactivate projects
   - Grant/revoke project access (`POST/DELETE /api/v1/admin/projects/:projectId/users/:userId/access`)

3. **Data Management**
   - Bulk delete test runs (`POST /api/v1/admin/test-runs/bulk-delete`)
   - System cleanup operations (`POST /api/v1/admin/system/cleanup`)
   - Export system data

4. **System Monitoring**
   - View system statistics (`GET /api/v1/admin/system/stats`)
   - Check system health (`GET /api/v1/admin/system/health`)
   - Access audit logs (`GET /api/v1/admin/audit-logs`)

5. **Tag Management**
   - Create/update/delete tags (`POST/PUT/DELETE /api/v1/admin/tags`)

### Regular User Functions
Regular users have access to:

1. **Test Result Viewing**
   - View test runs for accessible projects
   - Filter and search test results
   - View test run details and drill-down

2. **Personal Settings**
   - Manage user preferences (`GET/PUT /api/v1/user/preferences`)
   - Manage project favorites
   - Update profile settings

3. **Project Access**
   - View assigned projects (`GET /api/v1/user/projects`)
   - Export test data for accessible projects

## Quick Start with Keycloak

### 1. Deploy Fern Platform with Keycloak

The Keycloak OAuth provider is now included as a component in the main Fern Platform KubeVela application. It uses the `dev-file` database (file-based H2) which is suitable for development and testing.

```bash
# Deploy the entire Fern Platform stack (includes Keycloak)
kubectl apply -f deployments/fern-platform-kubevela.yaml

# Wait for all components to be ready
kubectl wait --for=condition=available deployment/keycloak -n fern-platform --timeout=300s
kubectl wait --for=condition=available deployment/fern-platform -n fern-platform --timeout=300s

# Port forward Keycloak to access admin console (Keycloak runs on port 8080 inside cluster)
kubectl port-forward service/keycloak 8081:8080 -n fern-platform

# Port forward Fern Platform
kubectl port-forward service/fern-platform 8080:8080 -n fern-platform

# Check the deployment
kubectl get pods -n fern-platform
```

### 2. Access Keycloak Admin Console

1. **Port forward to access Keycloak locally:**
   ```bash
   kubectl port-forward service/keycloak 8081:8080 -n fern-platform
   ```

2. **Access admin console:**
   - URL: http://localhost:8081/admin
   - Username: `admin`
   - Password: `admin123`

### 3. Configure Keycloak Realm

The deployment includes a pre-configured realm with test users and client setup. The realm configuration is automatically applied via ConfigMap when Keycloak starts.

**Pre-configured realm details:**

1. **Create Realm:** `fern-platform`
2. **Create Client:** `fern-platform-web`
   - Client Type: OpenID Connect
   - Client authentication: On
   - Valid redirect URIs: `http://localhost:8080/auth/callback`
   - Web origins: `http://localhost:8080`

3. **Create Groups:**
   - `fern-platform-admins` (admin role)
   - `fern-platform-users` (user role)

4. **Create Test Users:**
   - Admin user: `admin@fern-platform.local` / `admin123`
   - Regular user: `user@fern-platform.local` / `user123`

### 4. Configure Fern Platform

The Fern Platform is automatically configured with OAuth settings via environment variables in the KubeVela deployment. The configuration uses internal Kubernetes service names for communication between components.

**Current OAuth configuration (set via environment variables):**

```bash
# OAuth is automatically enabled
OAUTH_ENABLED=true
OAUTH_CLIENT_ID=fern-platform-web
OAUTH_CLIENT_SECRET=fern-platform-client-secret
OAUTH_REDIRECT_URL=http://localhost:8080/auth/callback

# Keycloak endpoints (using internal Kubernetes service names)
OAUTH_AUTH_URL=http://keycloak:8080/realms/fern-platform/protocol/openid-connect/auth
OAUTH_TOKEN_URL=http://keycloak:8080/realms/fern-platform/protocol/openid-connect/token
OAUTH_USERINFO_URL=http://keycloak:8080/realms/fern-platform/protocol/openid-connect/userinfo
OAUTH_JWKS_URL=http://keycloak:8080/realms/fern-platform/protocol/openid-connect/certs
```

**For local development** (if running outside Kubernetes), use the example config file `config/oauth-keycloak-example.yaml` which has localhost URLs.

### 5. Test Authentication

1. **Access the applications:**
   ```bash
   # In separate terminals:
   
   # Port forward Keycloak admin console (port 8081)
   kubectl port-forward service/keycloak 8081:8080 -n fern-platform
   
   # Port forward Fern Platform (port 8080)
   kubectl port-forward service/fern-platform 8080:8080 -n fern-platform
   ```

2. **Test Login Flow:**
   - Go to: http://localhost:8080
   - Click "Login" or access any protected resource
   - You'll be redirected to Keycloak
   - Login with test credentials:
     - Admin: `admin@fern-platform.local` / `admin123`
     - User: `user@fern-platform.local` / `user123`
   - You'll be redirected back to Fern Platform

## Configuration for Other OAuth Providers

### Auth0 Example

```yaml
auth:
  oauth:
    enabled: true
    clientId: "your-auth0-client-id"
    clientSecret: "your-auth0-client-secret"
    redirectUrl: "http://localhost:8080/auth/callback"
    scopes: ["openid", "profile", "email"]
    
    authUrl: "https://your-domain.auth0.com/authorize"
    tokenUrl: "https://your-domain.auth0.com/oauth/token"
    userInfoUrl: "https://your-domain.auth0.com/userinfo"
    jwksUrl: "https://your-domain.auth0.com/.well-known/jwks.json"
    
    adminUsers: ["admin@yourcompany.com"]
```

### Google OAuth Example

```yaml
auth:
  oauth:
    enabled: true
    clientId: "your-google-client-id.googleusercontent.com"
    clientSecret: "your-google-client-secret"
    redirectUrl: "http://localhost:8080/auth/callback"
    scopes: ["openid", "profile", "email"]
    
    authUrl: "https://accounts.google.com/o/oauth2/v2/auth"
    tokenUrl: "https://oauth2.googleapis.com/token"
    userInfoUrl: "https://openidconnect.googleapis.com/v1/userinfo"
    jwksUrl: "https://www.googleapis.com/oauth2/v3/certs"
    
    adminUsers: ["admin@yourcompany.com"]
```

### Okta Example

```yaml
auth:
  oauth:
    enabled: true
    clientId: "your-okta-client-id"
    clientSecret: "your-okta-client-secret"
    redirectUrl: "http://localhost:8080/auth/callback"
    scopes: ["openid", "profile", "email", "groups"]
    
    authUrl: "https://your-domain.okta.com/oauth2/default/v1/authorize"
    tokenUrl: "https://your-domain.okta.com/oauth2/default/v1/token"
    userInfoUrl: "https://your-domain.okta.com/oauth2/default/v1/userinfo"
    jwksUrl: "https://your-domain.okta.com/oauth2/default/v1/keys"
    
    adminUsers: ["admin@yourcompany.com"]
    groupsField: "groups"  # Okta includes groups in token
```

## Environment Variables

All OAuth settings can be configured via environment variables:

```bash
# OAuth Configuration
export OAUTH_ENABLED=true
export OAUTH_CLIENT_ID="fern-platform-web"
export OAUTH_CLIENT_SECRET="fern-platform-client-secret"
export OAUTH_REDIRECT_URL="http://localhost:8080/auth/callback"
export OAUTH_AUTH_URL="http://localhost:8080/realms/fern-platform/protocol/openid-connect/auth"
export OAUTH_TOKEN_URL="http://localhost:8080/realms/fern-platform/protocol/openid-connect/token"
export OAUTH_USERINFO_URL="http://localhost:8080/realms/fern-platform/protocol/openid-connect/userinfo"
export OAUTH_JWKS_URL="http://localhost:8080/realms/fern-platform/protocol/openid-connect/certs"
```

## Testing the OAuth Flow

### 1. Manual Testing

1. **Access the application:** http://localhost:8080
2. **Trigger authentication:** Click login or access `/admin` endpoints
3. **OAuth redirect:** You'll be redirected to your OAuth provider
4. **Login:** Use test credentials
5. **Callback:** You'll be redirected back with authentication

### 2. API Testing

```bash
# Test protected endpoint (should redirect to OAuth)
curl -v http://localhost:8080/api/v1/admin/users

# Test with session cookie (after web login)
curl -v -b "session_id=your-session-id" http://localhost:8080/api/v1/admin/users

# Get current user info
curl -v -b "session_id=your-session-id" http://localhost:8080/auth/user
```

### 3. Role Testing

```bash
# Admin user should have access
curl -v -b "admin-session-id" http://localhost:8080/api/v1/admin/users

# Regular user should get 403
curl -v -b "user-session-id" http://localhost:8080/api/v1/admin/users
```

## Troubleshooting

### Common Issues

1. **"OAuth not enabled" error**
   - Check `auth.oauth.enabled: true` in config
   - Verify environment variables are set

2. **"Invalid redirect URI" error**
   - Ensure `redirectUrl` matches OAuth provider configuration
   - Check for trailing slashes and protocol mismatches

3. **"Token exchange failed" error**
   - Verify `clientSecret` is correct
   - Check `tokenUrl` is accessible from server
   - Ensure OAuth provider accepts the redirect URI

4. **"User not found" error**
   - Check user exists in OAuth provider
   - Verify email/username format
   - Check claim field mappings (`userIdField`, `emailField`)

5. **"Access denied" error**
   - Verify user roles and groups
   - Check admin user/group configuration
   - Ensure groups are included in OAuth token claims

### Debug Logging

Enable debug logging to troubleshoot OAuth issues:

```yaml
logging:
  level: "debug"
```

This will log OAuth token exchanges, user info requests, and role assignments.

### Health Checks

Check OAuth provider connectivity:

```bash
# Check OAuth provider endpoints
curl -v $OAUTH_AUTH_URL
curl -v $OAUTH_JWKS_URL

# Check Keycloak health (if using Keycloak)
curl http://localhost:8081/health/ready
```

## Security Considerations

1. **Use HTTPS in production** - OAuth requires secure connections
2. **Secure client secrets** - Store secrets in environment variables or secret management
3. **Validate redirect URIs** - Ensure only authorized URIs are configured
4. **Token expiration** - Configure appropriate token lifetimes
5. **Session management** - Implement proper session timeout and cleanup
6. **Audit logging** - Enable audit logs for security monitoring

## Production Deployment

### Database Configuration

For production, configure Keycloak with PostgreSQL:

```yaml
# In keycloak deployment
env:
  - name: KC_DB
    value: "postgres"
  - name: KC_DB_URL
    value: "jdbc:postgresql://postgres.fern-platform:5432/keycloak"
  - name: KC_DB_USERNAME
    value: "keycloak"
  - name: KC_DB_PASSWORD
    valueFrom:
      secretKeyRef:
        name: keycloak-db-secret
        key: password
```

### TLS/SSL Configuration

```yaml
# Enable HTTPS
auth:
  oauth:
    redirectUrl: "https://fern-platform.yourcompany.com/auth/callback"
    authUrl: "https://auth.yourcompany.com/realms/fern-platform/protocol/openid-connect/auth"
    # ... other HTTPS URLs
```

### High Availability

Deploy multiple replicas for both Keycloak and Fern Platform:

```yaml
traits:
  - type: scaler
    properties:
      replicas: 3
```

This completes the OAuth implementation with provider-agnostic configuration, comprehensive admin functions, and detailed testing documentation.