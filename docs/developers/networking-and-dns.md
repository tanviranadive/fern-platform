# Networking and DNS Configuration for Local Development

This document explains the networking setup for Fern Platform in local development environments, particularly the DNS configuration required for OAuth authentication to work properly.

## Overview

When running Fern Platform locally with k3d, we use a specific DNS configuration that allows:
1. **Browser access** to both Keycloak and Fern Platform
2. **Service-to-service communication** within Kubernetes
3. **OAuth redirects** to work correctly

## DNS Configuration

### Required /etc/hosts Entries

Add these entries to your `/etc/hosts` file:

```bash
127.0.0.1 keycloak
127.0.0.1 fern-platform.local
```

### Why These Specific Names?

#### `keycloak` (without .local)
- **Kubernetes Limitation**: Service names in Kubernetes cannot contain dots (.)
- **Dual Purpose**: 
  - Internal: Kubernetes service name for pod-to-pod communication
  - External: Browser access via Traefik ingress
- **Used by**: 
  - Fern Platform backend (for token validation)
  - Your browser (for OAuth flow)

#### `fern-platform.local`
- **User-Friendly**: Clear indication this is a local development URL
- **Separation**: Distinguishes platform access from OAuth provider
- **Browser Access**: Main entry point for the application

## How It Works

### 1. Browser OAuth Flow
```
Browser → http://fern-platform.local:8080 
       → Redirect to http://keycloak:8080/auth
       → User logs in
       → Redirect back to http://fern-platform.local:8080/auth/callback
```

### 2. Backend Token Validation
```
Fern Platform Pod → http://keycloak:8080/realms/fern-platform/protocol/openid-connect/userinfo
                  → (Internal Kubernetes service resolution)
```

### 3. Traefik Ingress Routing
- **keycloak** → Routes to Keycloak service on port 8080
- **fern-platform.local** → Routes to Fern Platform service on port 8080

## Configuration Details

### OAuth URLs in fern-platform-kubevela.yaml

```yaml
# Browser-facing URLs (in Keycloak client config)
redirectUris:
  - "http://fern-platform.local:8080/auth/callback"
  - "http://localhost:8080/auth/callback"  # Alternative access

# Server-to-server URLs (in Fern Platform env vars)
OAUTH_TOKEN_URL: "http://keycloak:8080/realms/fern-platform/protocol/openid-connect/token"
OAUTH_USERINFO_URL: "http://keycloak:8080/realms/fern-platform/protocol/openid-connect/userinfo"
```

### Why Not keycloak.local?

1. **Kubernetes Services**: Cannot have dots in their names
2. **Internal Communication**: Pods resolve `keycloak` to the service ClusterIP
3. **Simplicity**: One name works for both internal and external access

## Troubleshooting

### Common Issues

#### 1. "Cannot resolve keycloak"
- **Cause**: Missing /etc/hosts entry
- **Fix**: Add `127.0.0.1 keycloak` to /etc/hosts

#### 2. OAuth redirect fails
- **Cause**: Mismatch between OAuth URLs and actual access URLs
- **Fix**: Ensure you're accessing via http://fern-platform.local:8080

#### 3. Token validation fails
- **Cause**: Backend can't reach Keycloak
- **Fix**: Check that Keycloak service is running: `kubectl get svc -n fern-platform`

### Verification Commands

```bash
# Check DNS resolution from your host
ping keycloak
ping fern-platform.local

# Check services are accessible
curl http://keycloak:8080/health
curl http://fern-platform.local:8080/health

# Check Kubernetes services
kubectl get svc -n fern-platform

# Check ingress rules
kubectl get ingress -n fern-platform
```

## Production Considerations

In production environments, you would typically:

1. **Use Real Domain Names**: 
   - `auth.company.com` for Keycloak
   - `fern.company.com` for the platform

2. **Configure Proper SSL/TLS**:
   - HTTPS for all endpoints
   - Valid SSL certificates

3. **Service Mesh or Internal DNS**:
   - Kubernetes internal DNS (keycloak.fern-platform.svc.cluster.local)
   - Service mesh for secure service-to-service communication

4. **Environment-Specific OAuth URLs**:
   - Use environment variables or ConfigMaps
   - Different URLs for internal vs external access

## Alternative Approaches

### Using Different Ports
Instead of domain-based routing, you could use:
- http://localhost:8080 → Fern Platform
- http://localhost:8081 → Keycloak

However, this requires additional port mapping configuration in k3d.

### Using Subdomains
With more complex DNS setup (like dnsmasq), you could use:
- http://app.fern.local
- http://auth.fern.local

But this adds complexity for local development.

## Summary

The current configuration strikes a balance between:
- **Simplicity**: Minimal DNS configuration required
- **Compatibility**: Works with Kubernetes naming restrictions
- **Functionality**: Supports complete OAuth flow
- **Portability**: Same setup works across different development machines

Remember: This configuration is specifically for local development. Production deployments should use proper domain names and SSL/TLS certificates.