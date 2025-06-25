# Fern Platform Setup Instructions

## Prerequisites
- k3d cluster running
- kubectl configured
- vela CLI installed

## 1. Add Host Entries

Add the following entries to your `/etc/hosts` file:

```bash
127.0.0.1 fern-platform.local
127.0.0.1 keycloak
```

## 2. Deploy Application

```bash
vela up -f deployments/fern-platform-kubevela.yaml
```

## 3. Access Application

- **Fern Platform**: http://fern-platform.local:8080
- **Keycloak Admin**: http://keycloak:8080/admin
- **Health Check**: http://fern-platform.local:8080/health

## 4. Test Users

The application comes with pre-configured test users:

- **Admin User**: 
  - Email: admin@fern.com
  - Password: admin123
  - Groups: fern-platform-admins
  - Role: admin

- **Regular User**:
  - Email: user@fern.com  
  - Password: user123
  - Groups: fern-platform-users
  - Role: user

## 5. OAuth Configuration

All OAuth configuration is included in the KubeVela application:
- Client ID: fern-platform-web
- Client Secret: fern-platform-client-secret
- Redirect URI: http://fern-platform.local:8080/auth/callback
- Scopes: openid, profile, email, groups
- Admin Groups: fern-platform-admins

## Troubleshooting

1. **404 Errors**: Ensure hosts entries are added and ingress class is traefik
2. **Connection Refused**: Check that k3d cluster is running and pods are ready
3. **OAuth Issues**: Check that Keycloak is accessible at http://keycloak:8080