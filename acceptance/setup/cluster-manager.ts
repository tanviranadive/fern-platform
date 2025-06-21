/**
 * Kubernetes Cluster Manager for Acceptance Tests
 * 
 * Handles:
 * - K8s namespace creation and isolation
 * - Service deployment via KubeVela
 * - Health checking and readiness verification
 * - Service URL discovery and port forwarding
 * - Resource cleanup
 */

import { exec } from 'child_process';
import { promisify } from 'util';
import * as path from 'path';
import * as fs from 'fs/promises';

const execAsync = promisify(exec);

interface ClusterManagerConfig {
  namespace: string;
  testId: string;
  config: any;
}

interface ServiceUrls {
  reporter: string;
  mycelium: string;
  ui: string;
}

export class ClusterManager {
  private namespace: string;
  private testId: string;
  private config: any;
  private deployedServices: string[] = [];
  private portForwards: { pid: number; port: number; service: string }[] = [];

  constructor(config: ClusterManagerConfig) {
    this.namespace = config.namespace;
    this.testId = config.testId;
    this.config = config.config;
  }

  async createNamespace(): Promise<void> {
    console.log(`📦 Creating namespace: ${this.namespace}`);
    
    try {
      // Create namespace with labels for cleanup
      const namespaceManifest = {
        apiVersion: 'v1',
        kind: 'Namespace',
        metadata: {
          name: this.namespace,
          labels: {
            'fern.test/managed': 'true',
            'fern.test/id': this.testId,
            'fern.test/created': new Date().toISOString()
          }
        }
      };

      const manifestPath = `/tmp/namespace-${this.testId}.yaml`;
      await fs.writeFile(manifestPath, JSON.stringify(namespaceManifest));
      
      await execAsync(`kubectl apply -f ${manifestPath}`);
      
      // Cleanup temp file
      await fs.unlink(manifestPath);
      
      console.log(`✅ Namespace created: ${this.namespace}`);
    } catch (error) {
      console.error(`❌ Failed to create namespace:`, error);
      throw error;
    }
  }

  async deployServices(): Promise<void> {
    console.log(`🚀 Deploying services in namespace: ${this.namespace}`);
    
    try {
      // Deploy using KubeVela with test-specific configuration
      const velaAppPath = path.join(__dirname, '../../deployments/local/vela.yaml');
      const testVelaApp = await this.customizeVelaApp(velaAppPath);
      
      const testAppPath = `/tmp/test-app-${this.testId}.yaml`;
      await fs.writeFile(testAppPath, testVelaApp);
      
      // Apply the application
      await execAsync(`vela up -f ${testAppPath} -n ${this.namespace}`);
      
      this.deployedServices = ['fern-reporter', 'fern-mycelium', 'fern-ui', 'postgresql', 'redis'];
      
      // Cleanup temp file
      await fs.unlink(testAppPath);
      
      console.log(`✅ Services deployment initiated`);
    } catch (error) {
      console.error(`❌ Failed to deploy services:`, error);
      throw error;
    }
  }

  async waitForServicesReady(timeout: number = 300000): Promise<void> {
    console.log(`⏳ Waiting for services to be ready (timeout: ${timeout/1000}s)`);
    
    const startTime = Date.now();
    const services = ['fern-reporter', 'fern-mycelium', 'fern-ui'];
    
    while (Date.now() - startTime < timeout) {
      try {
        const readyServices = await Promise.all(
          services.map(service => this.checkServiceReady(service))
        );
        
        if (readyServices.every(ready => ready)) {
          console.log(`✅ All services are ready`);
          return;
        }
        
        // Wait before next check
        await new Promise(resolve => setTimeout(resolve, 5000));
        
      } catch (error) {
        console.log(`⏳ Services not ready yet, continuing to wait...`);
        await new Promise(resolve => setTimeout(resolve, 5000));
      }
    }
    
    throw new Error(`Services failed to become ready within ${timeout/1000}s`);
  }

  async getServiceUrls(): Promise<ServiceUrls> {
    console.log(`🔍 Discovering service URLs`);
    
    // Setup port forwarding for each service
    const basePort = 38000 + parseInt(this.testId.slice(-3), 36) % 1000;
    
    const reporterPort = await this.setupPortForward('fern-reporter', 8080, basePort);
    const myceliumPort = await this.setupPortForward('fern-mycelium', 8081, basePort + 1);
    const uiPort = await this.setupPortForward('fern-ui', 3000, basePort + 2);
    
    const urls = {
      reporter: `http://localhost:${reporterPort}`,
      mycelium: `http://localhost:${myceliumPort}`,
      ui: `http://localhost:${uiPort}`
    };
    
    console.log(`✅ Service URLs configured:`, urls);
    return urls;
  }

  async cleanup(): Promise<void> {
    console.log(`🧹 Cleaning up cluster resources for: ${this.testId}`);
    
    try {
      // Kill port forwards
      await Promise.all(
        this.portForwards.map(pf => this.killPortForward(pf.pid))
      );
      
      // Delete namespace (cascades to all resources)
      if (this.namespace) {
        try {
          await execAsync(`kubectl delete namespace ${this.namespace} --timeout=60s`);
          console.log(`✅ Namespace deleted: ${this.namespace}`);
        } catch (error) {
          console.warn(`⚠️  Failed to delete namespace ${this.namespace}:`, error);
          // Force delete if stuck
          await execAsync(`kubectl delete namespace ${this.namespace} --force --grace-period=0`).catch(() => {});
        }
      }
      
    } catch (error) {
      console.error(`❌ Cleanup failed:`, error);
    }
  }

  private async customizeVelaApp(originalPath: string): Promise<string> {
    try {
      const originalContent = await fs.readFile(originalPath, 'utf-8');
      
      // Customize the VelaApp for test environment
      const customized = originalContent
        .replace(/namespace: fern/g, `namespace: ${this.namespace}`)
        .replace(/name: fern-local/g, `name: fern-test-${this.testId}`)
        // Add test-specific labels and annotations
        .replace(
          /metadata:/g, 
          `metadata:\n  labels:\n    fern.test/managed: "true"\n    fern.test/id: "${this.testId}"`
        );
      
      return customized;
    } catch (error) {
      console.error(`❌ Failed to customize VelaApp:`, error);
      throw error;
    }
  }

  private async checkServiceReady(serviceName: string): Promise<boolean> {
    try {
      // Check if pods are running
      const { stdout } = await execAsync(
        `kubectl get pods -n ${this.namespace} -l app=${serviceName} -o jsonpath='{.items[*].status.phase}'`
      );
      
      const phases = stdout.trim().split(' ').filter(p => p);
      return phases.length > 0 && phases.every(phase => phase === 'Running');
      
    } catch (error) {
      return false;
    }
  }

  private async setupPortForward(serviceName: string, servicePort: number, localPort: number): Promise<number> {
    try {
      // Start port forward in background
      const command = `kubectl port-forward -n ${this.namespace} service/${serviceName} ${localPort}:${servicePort}`;
      const child = exec(command);
      
      // Store process for cleanup
      if (child.pid) {
        this.portForwards.push({
          pid: child.pid,
          port: localPort,
          service: serviceName
        });
      }
      
      // Wait for port forward to be ready
      await new Promise(resolve => setTimeout(resolve, 2000));
      
      return localPort;
      
    } catch (error) {
      console.error(`❌ Failed to setup port forward for ${serviceName}:`, error);
      throw error;
    }
  }

  private async killPortForward(pid: number): Promise<void> {
    try {
      process.kill(pid, 'SIGTERM');
    } catch (error) {
      // Process may already be dead
    }
  }
}