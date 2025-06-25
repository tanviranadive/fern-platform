# Fern Platform Complete Setup Guide

This guide provides step-by-step instructions to deploy and run the Fern Platform with OAuth authentication using Keycloak, k3d, and KubeVela.

## Prerequisites

### Required Software
- **Docker Desktop** - For running k3d clusters
- **k3d** - Lightweight Kubernetes distribution
- **kubectl** - Kubernetes command-line tool
- **vela CLI** - KubeVela application platform

### Installation Commands

```bash
# Install k3d (macOS with Homebrew)
brew install k3d

# Install kubectl (if not already installed)
brew install kubectl

# Install vela CLI
curl -fsSl https://kubevela.io/script/install.sh | bash
```

For other operating systems, refer to the official installation guides:
- [k3d Installation](https://k3d.io/v5.4.6/#installation)
- [kubectl Installation](https://kubernetes.io/docs/tasks/tools/)
- [KubeVela Installation](https://kubevela.io/docs/installation/kubernetes)

## Step 1: Create k3d Cluster

Create a k3d cluster with port mappings for web access:

```bash
# Create cluster with port forwarding
k3d cluster create fern-platform \
  --port "8080:80@loadbalancer" \
  --port "8081:8080@loadbalancer" \
  --agents 1

# Verify cluster is running
kubectl cluster-info
kubectl get nodes
```

**Important**: The port mappings allow:
- `8080:80` - Access to applications via Traefik ingress on port 8080
- `8081:8080` - Direct access to services on port 8081 (if needed)

## Step 2: Install KubeVela

Install KubeVela in your k3d cluster:

```bash
# Install KubeVela core
vela install

# Wait for installation to complete (may take 2-3 minutes)
kubectl wait --for=condition=Ready pod -l app.kubernetes.io/name=vela-core -n vela-system --timeout=300s

# Verify installation
vela version
kubectl get pods -n vela-system
```

## Step 3: Create Namespace

Create the namespace for the Fern Platform:

```bash
# Create namespace
kubectl create namespace fern-platform

# Verify namespace creation
kubectl get namespaces | grep fern-platform
```

## Step 4: **CRITICAL** - Configure Host Resolution

Add the following entries to your `/etc/hosts` file to enable local DNS resolution:

```bash
# Open hosts file for editing (requires sudo)
sudo vim /etc/hosts

# Add these lines to the end of the file:
127.0.0.1 fern-platform.local
127.0.0.1 keycloak
```

**Alternative using echo commands:**
```bash
echo "127.0.0.1 fern-platform.local" | sudo tee -a /etc/hosts
echo "127.0.0.1 keycloak" | sudo tee -a /etc/hosts
```

**Why this is required**: 
- The application uses host-based routing via Traefik ingress
- Keycloak needs to be accessible at `keycloak:8080` from both browser and pods
- Without these entries, OAuth redirects will fail

## Step 5: Deploy the Application

Deploy the complete application stack using KubeVela:

```bash
# Navigate to the project directory
cd fern-platform

# Deploy the application
vela up -f deployments/fern-platform-kubevela.yaml

# Monitor deployment progress
kubectl get pods -n fern-platform -w
```

**Expected deployment order:**
1. PostgreSQL database
2. Redis cache  
3. Keycloak realm configuration (ConfigMap)
4. Keycloak OAuth provider
5. Fern Platform application (after 90-second wait for infrastructure)

## Step 6: Verify Deployment

Wait for all pods to be ready (this may take 5-10 minutes):

```bash
# Check pod status
kubectl get pods -n fern-platform

# Expected output (all pods should show "Running" and "1/1" ready):
# NAME                             READY   STATUS    RESTARTS   AGE
# fern-platform-xxxxxxxxxx-xxxxx   1/1     Running   0          5m
# keycloak-xxxxxxxxxx-xxxxx        1/1     Running   0          8m
# postgres-1                       1/1     Running   0          10m
# redis-xxxxxxxxxx-xxxxx           1/1     Running   0          10m

# Check application logs
kubectl logs -n fern-platform deployment/fern-platform --tail=20

# Check Keycloak logs  
kubectl logs -n fern-platform deployment/keycloak --tail=20
```

## Step 7: Test Application Access

### 7.1 Health Check
Verify the application is responding:

```bash
curl http://fern-platform.local:8080/health
# Expected: {"status":"healthy","timestamp":"..."}
```

### 7.2 Keycloak Admin Access
Access Keycloak admin console to verify OAuth provider:

- **URL**: http://keycloak:8080/admin
- **Username**: admin
- **Password**: admin123

Verify the `fern-platform` realm exists with the configured client.

### 7.3 Main Application Access
Open your browser and navigate to:

**URL**: http://fern-platform.local:8080

You should be redirected to the login page with a "Sign In with OAuth" button.

## Step 8: Test OAuth Authentication

### Pre-configured Test Users

The application comes with two pre-configured users:

#### Admin User
- **Email**: admin@fern.com
- **Password**: admin123
- **Groups**: fern-platform-admins
- **Role**: admin
- **Permissions**: Full access to admin panel, user management, system settings

#### Regular User  
- **Email**: user@fern.com
- **Password**: user123
- **Groups**: fern-platform-users
- **Role**: user
- **Permissions**: Read-only access to dashboards and test data

### Testing the Authentication Flow

1. **Navigate to Application**: http://fern-platform.local:8080
2. **Click "Sign In with OAuth"** - redirects to Keycloak
3. **Enter credentials** - use either admin@fern.com or user@fern.com
4. **Successful login** - redirected back to dashboard
5. **Test logout** - click user menu → "Sign Out"
6. **Verify logout** - should return to login page, requiring re-authentication

### Testing Admin Features (admin@fern.com only)

After logging in as admin, verify access to:
- **User Management**: View and manage all users
- **Project Management**: Create, edit, activate/deactivate projects  
- **System Settings**: View system statistics and maintenance tools
- **Admin Navigation**: Additional menu items for admin functions

## OAuth Configuration Details

The OAuth integration is fully configured in the KubeVela deployment:

### OAuth Provider Endpoints
- **Authorization**: http://keycloak:8080/realms/fern-platform/protocol/openid-connect/auth
- **Token**: http://keycloak:8080/realms/fern-platform/protocol/openid-connect/token
- **UserInfo**: http://keycloak:8080/realms/fern-platform/protocol/openid-connect/userinfo
- **Logout**: http://keycloak:8080/realms/fern-platform/protocol/openid-connect/logout

### OAuth Client Configuration
- **Client ID**: fern-platform-web
- **Client Secret**: fern-platform-client-secret  
- **Redirect URI**: http://fern-platform.local:8080/auth/callback
- **Post-logout URI**: http://fern-platform.local:8080/auth/login
- **Scopes**: openid, profile, email, groups

### Role Mapping
- **Admin Users**: admin@fern.com
- **Admin Groups**: fern-platform-admins
- **User Groups**: fern-platform-users

## Troubleshooting

### Common Issues and Solutions

#### 1. DNS Resolution Issues
**Symptoms**: 404 errors, "site can't be reached"
**Solution**: 
```bash
# Verify hosts file entries
cat /etc/hosts | grep -E "(fern-platform.local|keycloak)"

# Should show:
# 127.0.0.1 fern-platform.local  
# 127.0.0.1 keycloak
```

#### 2. Pod Startup Issues
**Symptoms**: Pods stuck in "Pending" or "CrashLoopBackOff"
**Solution**:
```bash
# Check pod details
kubectl describe pod -n fern-platform <pod-name>

# Check resource usage
kubectl top nodes
kubectl top pods -n fern-platform

# Common fix: Restart Docker Desktop if resource constraints
```

#### 3. Keycloak Import Issues  
**Symptoms**: Keycloak starts but realm not configured
**Solution**:
```bash
# Check Keycloak logs for import errors
kubectl logs -n fern-platform deployment/keycloak | grep -i import

# Restart Keycloak pod if needed
kubectl delete pod -n fern-platform -l app.oam.dev/component=keycloak
```

#### 4. OAuth Redirect Errors
**Symptoms**: "Invalid redirect URI" after login
**Solution**: Verify hosts file and that you're accessing via http://fern-platform.local:8080 (not localhost)

#### 5. Database Connection Issues
**Symptoms**: Application fails to start, database errors in logs
**Solution**:
```bash
# Check PostgreSQL pod
kubectl logs -n fern-platform postgres-1

# Restart if needed
kubectl delete pod -n fern-platform postgres-1
```

### Log Debugging Commands

```bash
# View application logs
kubectl logs -n fern-platform deployment/fern-platform -f

# View Keycloak logs  
kubectl logs -n fern-platform deployment/keycloak -f

# View all pod events
kubectl get events -n fern-platform --sort-by='.lastTimestamp'

# Check ingress status
kubectl get ingress -n fern-platform
```

### Port Forwarding (Alternative Access)

If ingress isn't working, you can use port forwarding:

```bash
# Forward fern-platform port
kubectl port-forward -n fern-platform service/fern-platform 8080:8080 &

# Forward Keycloak port
kubectl port-forward -n fern-platform service/keycloak 8081:8080 &

# Access via:
# - Fern Platform: http://localhost:8080
# - Keycloak: http://localhost:8081
```

## Cleanup

To remove the entire deployment:

```bash
# Delete the application
vela delete fern-platform-app

# Delete the namespace  
kubectl delete namespace fern-platform

# Delete the k3d cluster (optional)
k3d cluster delete fern-platform

# Remove hosts file entries (manual)
sudo vim /etc/hosts
# Remove the fern-platform.local and keycloak lines
```

## Next Steps

After successful deployment:

1. **Test API endpoints** - Use the admin credentials to test the REST API
2. **Submit test data** - Configure your test suites to report to the platform
3. **Customize configuration** - Modify OAuth settings for your specific provider
4. **Scale deployment** - Adjust resource limits and replica counts as needed

For API testing and test data submission, refer to the [OAuth Setup and Testing Documentation](docs/OAuth-Setup-and-Testing.md).