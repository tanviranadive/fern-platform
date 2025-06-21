/**
 * Custom Jest Test Environment for Fern Platform Acceptance Tests
 * 
 * This environment handles:
 * - Kubernetes cluster setup/teardown
 * - Service deployment and health checking
 * - Test isolation and cleanup
 * - Performance monitoring
 */

import { NodeEnvironment } from 'jest-environment-node';
import type { EnvironmentContext, JestEnvironmentConfig } from '@jest/environment';
import { ClusterManager } from './cluster-manager';
import { DataFixtures } from './data-fixtures';

interface TestEnvironmentGlobal extends NodeJS.Global {
  __FERN_TEST_CONTEXT__: {
    clusterManager: ClusterManager;
    dataFixtures: DataFixtures;
    namespace: string;
    baseUrls: {
      reporter: string;
      mycelium: string;
      ui: string;
    };
    testId: string;
    startTime: number;
  };
}

export default class FernAcceptanceTestEnvironment extends NodeEnvironment {
  private clusterManager: ClusterManager;
  private dataFixtures: DataFixtures;
  private namespace: string;
  private testId: string;

  constructor(config: JestEnvironmentConfig, context: EnvironmentContext) {
    super(config, context);
    
    // Generate unique test identifier
    this.testId = `test-${Date.now()}-${Math.random().toString(36).substring(7)}`;
    this.namespace = `fern-test-${this.testId}`;
    
    // Initialize managers
    this.clusterManager = new ClusterManager({
      namespace: this.namespace,
      testId: this.testId,
      config: config.projectConfig
    });
    
    this.dataFixtures = new DataFixtures({
      namespace: this.namespace,
      testId: this.testId
    });
  }

  async setup(): Promise<void> {
    await super.setup();
    
    console.log(`🚀 Setting up test environment: ${this.testId}`);
    const startTime = Date.now();
    
    try {
      // Create isolated namespace
      await this.clusterManager.createNamespace();
      
      // Deploy core services
      await this.clusterManager.deployServices();
      
      // Wait for services to be ready
      await this.clusterManager.waitForServicesReady();
      
      // Setup test data
      await this.dataFixtures.setupTestData();
      
      // Expose test context to global scope
      const testContext = {
        clusterManager: this.clusterManager,
        dataFixtures: this.dataFixtures,
        namespace: this.namespace,
        baseUrls: await this.clusterManager.getServiceUrls(),
        testId: this.testId,
        startTime
      };
      
      (this.global as TestEnvironmentGlobal).__FERN_TEST_CONTEXT__ = testContext;
      
      const setupTime = Date.now() - startTime;
      console.log(`✅ Test environment ready in ${setupTime}ms`);
      
    } catch (error) {
      console.error(`❌ Failed to setup test environment:`, error);
      await this.cleanup();
      throw error;
    }
  }

  async teardown(): Promise<void> {
    console.log(`🧹 Tearing down test environment: ${this.testId}`);
    
    try {
      await this.cleanup();
      console.log(`✅ Test environment cleaned up: ${this.testId}`);
    } catch (error) {
      console.error(`⚠️  Cleanup failed for ${this.testId}:`, error);
    }
    
    await super.teardown();
  }

  private async cleanup(): Promise<void> {
    if (this.dataFixtures) {
      await this.dataFixtures.cleanup();
    }
    
    if (this.clusterManager) {
      await this.clusterManager.cleanup();
    }
  }

  getVmContext() {
    return super.getVmContext();
  }
}