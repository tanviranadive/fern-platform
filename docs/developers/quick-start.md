# 🚀 Quick Start Guide

**Get Fern Platform running in 15 minutes with OAuth authentication and sample data**

<div align="center">
  <img src="../images/quick-start-flow.png" alt="Quick Start Flow" width="600"/>
</div>

## ⚡ Choose Your Path

| Time Available | Method | Best For |
|----------------|---------|----------|
| **15 minutes** | [Full Local Setup](#-15-minute-full-setup) | Developers ready to explore |
| **30 minutes** | [Production Ready](#-30-minute-production-setup) | Teams preparing for deployment |

---

## 🔥 15-Minute Full Setup

**Perfect for:** Developers who want to explore features and APIs

### Prerequisites (2 minutes)
```bash
# Install dependencies (macOS with Homebrew)
brew install k3d kubectl

# Install KubeVela
curl -fsSl https://kubevela.io/script/install.sh | bash

# For other OS, see: https://kubevela.io/docs/installation/
```

### One-Command Setup (13 minutes)
```bash
# Clone and setup everything
git clone https://github.com/guidewire-oss/fern-platform
cd fern-platform

# This single command:
# 1. Creates k3d cluster with port mappings
# 2. Installs KubeVela 
# 3. Deploys PostgreSQL, Redis, Keycloak, and Fern Platform
# 4. Configures OAuth with test users
# 5. Sets up DNS resolution
make quick-start

# After 10-15 minutes, visit:
# http://fern-platform.local:8080
```

### Test Credentials
```bash
# Admin user (full access)
Username: admin@fern.com
Password: admin123

# Regular user (read-only)  
Username: user@fern.com
Password: user123
```

### What You Get
- ✅ **Full OAuth flow** with Keycloak
- ✅ **Admin vs user roles** and permissions
- ✅ **Real database** with migrations
- ✅ **All APIs** (REST + GraphQL)
- ✅ **Modern web UI** with treemap visualization
- ✅ **Multi-framework test data** support
- 🚧 **AI features planned** (not yet implemented)

### Verify Everything Works
```bash
# 1. Check all pods are running
kubectl get pods -n fern-platform

# 2. Test health endpoint
curl http://fern-platform.local:8080/health

# 3. Test OAuth login (browser)
open http://fern-platform.local:8080

# 4. Test admin API
curl -H "Accept: application/json" \
  http://fern-platform.local:8080/api/v1/admin/users \
  # (This should redirect to OAuth login)
```

---

## 🏢 30-Minute Production Setup

**Perfect for:** Teams ready to deploy to their Kubernetes cluster

### Step 1: Customize Configuration (10 minutes)

1. **Copy the deployment template:**
   ```bash
   cp deployments/fern-platform-kubevela.yaml deployments/production.yaml
   ```

2. **Update OAuth settings:**
   ```yaml
   # In deployments/production.yaml
   env:
     - name: OAUTH_CLIENT_ID
       value: "your-production-client-id"
     - name: OAUTH_CLIENT_SECRET
       value: "your-production-client-secret"
     - name: OAUTH_AUTH_URL
       value: "https://auth.yourcompany.com/oauth2/authorize"
     # ... other OAuth endpoints
   ```

3. **Configure admin users:**
   ```yaml
   env:
     - name: OAUTH_ADMIN_USERS
       value: "admin@yourcompany.com,platform-team@yourcompany.com"
     - name: OAUTH_ADMIN_GROUPS
       value: "platform-admins,engineering-leads"
   ```

### Step 2: Deploy to Production (15 minutes)

```bash
# 1. Create namespace
kubectl create namespace fern-platform-prod

# 2. Deploy with production config
kubectl apply -f deployments/production.yaml

# 3. Wait for deployment
kubectl wait --for=condition=Available \
  deployment/fern-platform -n fern-platform-prod \
  --timeout=300s

# 4. Set up ingress/load balancer (depends on your cluster)
kubectl apply -f deployments/ingress-production.yaml
```

### Step 3: Verify Production Setup (5 minutes)

```bash
# 1. Check all components
kubectl get pods -n fern-platform-prod

# 2. Test health endpoint
kubectl port-forward -n fern-platform-prod service/fern-platform 8080:8080 &
curl http://localhost:8080/health

# 3. Test with your OAuth provider
# Visit your production URL and test login
```

**[📖 Complete production deployment guide →](../operations/production-setup.md)**

---

## 🔧 Next Steps After Setup

### 1. Explore the Platform (5 minutes)
- **Dashboard**: Overview of platform status and recent activity
- **Test Summaries**: Grid and treemap views of all projects  
- **Test Runs**: Detailed test execution data
- **Admin Panel** (admin users only): User and project management

### 2. Send Your First Test Data (10 minutes)
```bash
# Example: Submit a test run via API
curl -X POST http://fern-platform.local:8080/api/v1/test-runs \
  -H "Content-Type: application/json" \
  -d '{
    "projectId": "my-project",
    "status": "passed",
    "duration": 1234,
    "gitCommit": "abc123",
    "gitBranch": "main"
  }'

# See it appear in the dashboard!
```

### 3. Integrate with Your CI/CD (20 minutes)
Choose your integration method:

#### GitHub Actions
```yaml
- name: Report test results
  uses: guidewire-oss/fern-ginkgo-action@v1
  with:
    fern-url: http://fern-platform.local:8080
    project-id: my-project
```

#### Jenkins Pipeline
```groovy
post {
  always {
    sh 'fern-junit-client submit results.xml --url=http://fern-platform.local:8080'
  }
}
```

#### Generic cURL
```bash
# Submit JUnit XML results
curl -X POST http://fern-platform.local:8080/api/v1/test-runs/junit \
  -F "file=@test-results.xml" \
  -F "projectId=my-project"
```

**[📖 Complete integration guide →](../developers/api-reference.md)**

### 4. Prepare for AI Features (Coming Soon)
The platform is designed with AI integration in mind, but these features are not yet implemented:

**Planned AI capabilities:**
- 🤖 **Flaky test detection** using statistical analysis
- 📊 **Failure pattern recognition** with ML models  
- 💡 **Smart test recommendations** powered by LLMs
- 🔍 **Automated root cause analysis**

**[📋 See our AI roadmap →](../../issues?q=is%3Aissue+is%3Aopen+label%3Aai)**

---

## 🆘 Troubleshooting

### Common Issues

#### "Pod stuck in Pending state"
```bash
# Check node resources
kubectl describe nodes
kubectl top nodes

# Usually means insufficient CPU/memory
# Solution: Free up resources or use smaller resource requests
```

#### "DNS resolution failed"
```bash
# Check /etc/hosts entries
cat /etc/hosts | grep fern-platform

# Should show:
# 127.0.0.1 fern-platform.local
# 127.0.0.1 keycloak

# Add if missing:
echo "127.0.0.1 fern-platform.local" | sudo tee -a /etc/hosts
echo "127.0.0.1 keycloak" | sudo tee -a /etc/hosts
```

#### "OAuth redirect error"
```bash
# Make sure you're accessing via the correct URL
# ✅ Good: http://fern-platform.local:8080
# ❌ Bad:  http://localhost:8080

# Check Keycloak client configuration matches
kubectl logs -n fern-platform deployment/keycloak | grep -i redirect
```

### Get Help
- 🐛 **Found a bug?** [Open an issue](../../issues)
- ❓ **Have questions?** [Start a discussion](../../discussions)  
- 📖 **Need more details?** [Browse full documentation](../README.md)
- 🔧 **Production issues?** [Troubleshooting guide](../operations/troubleshooting.md)

---

## ✅ Quick Start Checklist

Copy this checklist to track your progress:

```markdown
## Fern Platform Quick Start

### Setup
- [ ] Prerequisites installed (k3d, kubectl, vela)
- [ ] Repository cloned
- [ ] `make quick-start` completed successfully
- [ ] All pods showing "Running" status

### Testing  
- [ ] Health endpoint responds: `curl http://fern-platform.local:8080/health`
- [ ] Web UI accessible: http://fern-platform.local:8080
- [ ] Admin login works: admin@fern.com / admin123
- [ ] User login works: user@fern.com / user123
- [ ] Can view dashboard, test summaries, and test runs

### Integration
- [ ] Submitted first test data via API
- [ ] Test run appears in dashboard
- [ ] Configured CI/CD integration (optional)
- [ ] Added AI API keys (optional)

### Next Steps
- [ ] Read [Architecture Overview](../architecture/overview.md)
- [ ] Explore [API Documentation](api-reference.md)
- [ ] Join [Community Discussions](../../discussions)
```

---

**🎉 Congratulations!** You now have a fully functional Fern Platform. Ready to dive deeper? Check out our [Developer Guide](local-development.md) or [Architecture Overview](../architecture/overview.md).