/**
 * Test Data Fixtures Manager
 * 
 * Handles:
 * - Test data generation and seeding
 * - Database state management
 * - Fixture cleanup and isolation
 * - Performance test data creation
 */

import { exec } from 'child_process';
import { promisify } from 'util';
import * as path from 'path';
import * as fs from 'fs/promises';

const execAsync = promisify(exec);

interface DataFixturesConfig {
  namespace: string;
  testId: string;
}

export interface TestProject {
  id: string;
  name: string;
  description?: string;
  tags: string[];
}

export interface TestRun {
  id: string;
  projectId: string;
  suiteId: string;
  branch: string;
  buildUrl?: string;
  buildActor?: string;
  gitSha: string;
  status: 'passed' | 'failed' | 'skipped';
  startTime: Date;
  endTime: Date;
  duration: number;
  tags: string[];
  specRuns: SpecRun[];
}

export interface SpecRun {
  id: string;
  testRunId: string;
  specDescription: string;
  status: 'passed' | 'failed' | 'skipped';
  startTime: Date;
  endTime: Date;
  duration: number;
  errorMessage?: string;
  stackTrace?: string;
}

export class DataFixtures {
  private namespace: string;
  private testId: string;
  private createdData: {
    projects: TestProject[];
    testRuns: TestRun[];
    specRuns: SpecRun[];
  } = {
    projects: [],
    testRuns: [],
    specRuns: []
  };

  constructor(config: DataFixturesConfig) {
    this.namespace = config.namespace;
    this.testId = config.testId;
  }

  async setupTestData(): Promise<void> {
    console.log(`📊 Setting up test data for: ${this.testId}`);
    
    try {
      // Wait for database to be ready
      await this.waitForDatabase();
      
      // Create base test projects
      await this.createTestProjects();
      
      // Create test runs with various scenarios
      await this.createTestRuns();
      
      // Create user preferences data
      await this.createUserPreferences();
      
      console.log(`✅ Test data setup completed`);
    } catch (error) {
      console.error(`❌ Failed to setup test data:`, error);
      throw error;
    }
  }

  async createTestProjects(): Promise<TestProject[]> {
    const projects: TestProject[] = [
      {
        id: `proj-auth-${this.testId}`,
        name: 'Authentication Service',
        description: 'Core authentication and authorization tests',
        tags: ['auth', 'security', 'core']
      },
      {
        id: `proj-api-${this.testId}`,
        name: 'API Gateway',
        description: 'API gateway and routing tests',
        tags: ['api', 'gateway', 'networking']
      },
      {
        id: `proj-ui-${this.testId}`,
        name: 'User Interface',
        description: 'Frontend user interface tests',
        tags: ['ui', 'frontend', 'react']
      },
      {
        id: `proj-perf-${this.testId}`,
        name: 'Performance Tests',
        description: 'Load and performance validation',
        tags: ['performance', 'load', 'stress']
      },
      {
        id: `proj-flaky-${this.testId}`,
        name: 'Flaky Test Showcase',
        description: 'Demonstrates flaky test patterns',
        tags: ['flaky', 'unstable', 'debug']
      }
    ];

    this.createdData.projects = projects;
    
    // Insert projects via SQL
    await this.executeSql(`
      INSERT INTO projects (id, name, description, tags, created_at)
      VALUES ${projects.map(p => 
        `('${p.id}', '${p.name}', '${p.description}', '${JSON.stringify(p.tags)}', NOW())`
      ).join(', ')}
      ON CONFLICT (id) DO NOTHING;
    `);

    return projects;
  }

  async createTestRuns(): Promise<TestRun[]> {
    const testRuns: TestRun[] = [];
    const projects = this.createdData.projects;
    
    // Create test runs for each project with different patterns
    for (const project of projects) {
      // Recent successful runs
      testRuns.push(...this.generateSuccessfulRuns(project, 15));
      
      // Some failed runs
      testRuns.push(...this.generateFailedRuns(project, 5));
      
      // Flaky test patterns (for flaky project)
      if (project.name.includes('Flaky')) {
        testRuns.push(...this.generateFlakyRuns(project, 20));
      }
      
      // Performance degradation pattern (for performance project)
      if (project.name.includes('Performance')) {
        testRuns.push(...this.generatePerformanceRuns(project, 10));
      }
    }

    this.createdData.testRuns = testRuns;
    
    // Insert test runs via SQL
    for (const testRun of testRuns) {
      await this.insertTestRun(testRun);
    }

    return testRuns;
  }

  async createUserPreferences(): Promise<void> {
    const userPreferences = [
      {
        userId: `user-test-${this.testId}`,
        preferences: {
          timezone: 'America/New_York',
          theme: 'dark',
          favoriteProjects: [
            this.createdData.projects[0].id,
            this.createdData.projects[2].id
          ],
          projectGroups: [
            {
              name: 'Core Services',
              projects: [
                this.createdData.projects[0].id,
                this.createdData.projects[1].id
              ]
            }
          ]
        }
      }
    ];

    // Insert user preferences
    await this.executeSql(`
      INSERT INTO user_preferences (user_id, preferences, created_at, updated_at)
      VALUES ${userPreferences.map(up => 
        `('${up.userId}', '${JSON.stringify(up.preferences)}', NOW(), NOW())`
      ).join(', ')}
      ON CONFLICT (user_id) DO UPDATE SET 
        preferences = EXCLUDED.preferences,
        updated_at = NOW();
    `);
  }

  async cleanup(): Promise<void> {
    console.log(`🧹 Cleaning up test data for: ${this.testId}`);
    
    try {
      // Clean up in reverse dependency order
      await this.executeSql(`DELETE FROM spec_runs WHERE test_run_id LIKE '%${this.testId}%';`);
      await this.executeSql(`DELETE FROM test_runs WHERE id LIKE '%${this.testId}%';`);
      await this.executeSql(`DELETE FROM projects WHERE id LIKE '%${this.testId}%';`);
      await this.executeSql(`DELETE FROM user_preferences WHERE user_id LIKE '%${this.testId}%';`);
      
      console.log(`✅ Test data cleaned up`);
    } catch (error) {
      console.error(`❌ Failed to cleanup test data:`, error);
    }
  }

  // Helper methods for generating different test patterns
  private generateSuccessfulRuns(project: TestProject, count: number): TestRun[] {
    const runs: TestRun[] = [];
    const baseTime = new Date();
    
    for (let i = 0; i < count; i++) {
      const startTime = new Date(baseTime.getTime() - (i * 60 * 60 * 1000)); // Each run 1 hour apart
      const duration = Math.floor(Math.random() * 300000) + 30000; // 30s to 5min
      const endTime = new Date(startTime.getTime() + duration);
      
      runs.push({
        id: `run-success-${project.id}-${i}-${this.testId}`,
        projectId: project.id,
        suiteId: `suite-${project.id}-${i}`,
        branch: i < 3 ? 'main' : `feature/branch-${i}`,
        buildUrl: `https://ci.example.com/builds/${i}`,
        buildActor: `developer-${i % 3}`,
        gitSha: this.generateGitSha(),
        status: 'passed',
        startTime,
        endTime,
        duration,
        tags: [...project.tags, 'success'],
        specRuns: this.generateSpecRuns(`run-success-${project.id}-${i}-${this.testId}`, 'passed', 10)
      });
    }
    
    return runs;
  }

  private generateFailedRuns(project: TestProject, count: number): TestRun[] {
    const runs: TestRun[] = [];
    const baseTime = new Date();
    
    for (let i = 0; i < count; i++) {
      const startTime = new Date(baseTime.getTime() - (i * 2 * 60 * 60 * 1000)); // Failed runs spread out
      const duration = Math.floor(Math.random() * 200000) + 60000; // 1min to 3.5min
      const endTime = new Date(startTime.getTime() + duration);
      
      runs.push({
        id: `run-failed-${project.id}-${i}-${this.testId}`,
        projectId: project.id,
        suiteId: `suite-${project.id}-failed-${i}`,
        branch: 'main',
        buildUrl: `https://ci.example.com/builds/failed-${i}`,
        buildActor: `developer-${i % 3}`,
        gitSha: this.generateGitSha(),
        status: 'failed',
        startTime,
        endTime,
        duration,
        tags: [...project.tags, 'failed'],
        specRuns: this.generateSpecRuns(`run-failed-${project.id}-${i}-${this.testId}`, 'mixed', 8)
      });
    }
    
    return runs;
  }

  private generateFlakyRuns(project: TestProject, count: number): TestRun[] {
    const runs: TestRun[] = [];
    const baseTime = new Date();
    
    for (let i = 0; i < count; i++) {
      const startTime = new Date(baseTime.getTime() - (i * 30 * 60 * 1000)); // Every 30 minutes
      const duration = Math.floor(Math.random() * 180000) + 45000; // 45s to 3min
      const endTime = new Date(startTime.getTime() + duration);
      
      // 70% pass rate to simulate flakiness
      const status = Math.random() < 0.7 ? 'passed' : 'failed';
      
      runs.push({
        id: `run-flaky-${project.id}-${i}-${this.testId}`,
        projectId: project.id,
        suiteId: `suite-${project.id}-flaky-${i}`,
        branch: 'main',
        buildUrl: `https://ci.example.com/builds/flaky-${i}`,
        buildActor: `developer-${i % 2}`,
        gitSha: this.generateGitSha(),
        status,
        startTime,
        endTime,
        duration,
        tags: [...project.tags, 'flaky'],
        specRuns: this.generateSpecRuns(`run-flaky-${project.id}-${i}-${this.testId}`, 'flaky', 6)
      });
    }
    
    return runs;
  }

  private generatePerformanceRuns(project: TestProject, count: number): TestRun[] {
    const runs: TestRun[] = [];
    const baseTime = new Date();
    
    for (let i = 0; i < count; i++) {
      const startTime = new Date(baseTime.getTime() - (i * 60 * 60 * 1000));
      // Simulate performance degradation over time
      const baseDuration = 120000; // 2 minutes base
      const degradation = i * 10000; // 10s degradation per run
      const duration = baseDuration + degradation + (Math.random() * 30000);
      const endTime = new Date(startTime.getTime() + duration);
      
      runs.push({
        id: `run-perf-${project.id}-${i}-${this.testId}`,
        projectId: project.id,
        suiteId: `suite-${project.id}-perf-${i}`,
        branch: 'main',
        buildUrl: `https://ci.example.com/builds/perf-${i}`,
        buildActor: 'performance-bot',
        gitSha: this.generateGitSha(),
        status: 'passed',
        startTime,
        endTime,
        duration,
        tags: [...project.tags, 'regression'],
        specRuns: this.generateSpecRuns(`run-perf-${project.id}-${i}-${this.testId}`, 'passed', 15)
      });
    }
    
    return runs;
  }

  private generateSpecRuns(testRunId: string, pattern: 'passed' | 'mixed' | 'flaky', count: number): SpecRun[] {
    const specs: SpecRun[] = [];
    const baseTime = new Date();
    
    for (let i = 0; i < count; i++) {
      const startTime = new Date(baseTime.getTime() + (i * 5000)); // 5s intervals
      const duration = Math.floor(Math.random() * 10000) + 1000; // 1-11s
      const endTime = new Date(startTime.getTime() + duration);
      
      let status: 'passed' | 'failed' | 'skipped';
      let errorMessage: string | undefined;
      
      switch (pattern) {
        case 'passed':
          status = 'passed';
          break;
        case 'mixed':
          status = i < count / 2 ? 'passed' : (Math.random() < 0.7 ? 'failed' : 'skipped');
          errorMessage = status === 'failed' ? `Test failure in spec ${i}: Assertion failed` : undefined;
          break;
        case 'flaky':
          status = Math.random() < 0.7 ? 'passed' : 'failed';
          errorMessage = status === 'failed' ? `Flaky test failure: Timeout waiting for element` : undefined;
          break;
      }
      
      specs.push({
        id: `spec-${testRunId}-${i}`,
        testRunId,
        specDescription: `Test Case ${i + 1}: ${this.generateSpecDescription(i)}`,
        status,
        startTime,
        endTime,
        duration,
        errorMessage,
        stackTrace: errorMessage ? `Error stack trace for spec ${i}` : undefined
      });
    }
    
    return specs;
  }

  private generateSpecDescription(index: number): string {
    const descriptions = [
      'User authentication flow',
      'API endpoint validation',
      'Database transaction handling',
      'UI component rendering',
      'Error handling scenarios',
      'Performance benchmark',
      'Security validation',
      'Integration test suite',
      'Regression test case',
      'Smoke test verification'
    ];
    
    return descriptions[index % descriptions.length];
  }

  private generateGitSha(): string {
    return Math.random().toString(36).substring(2, 10);
  }

  private async insertTestRun(testRun: TestRun): Promise<void> {
    // Insert test run
    await this.executeSql(`
      INSERT INTO test_runs (id, project_id, suite_id, branch, build_url, build_actor, git_sha, status, start_time, end_time, duration, tags)
      VALUES ('${testRun.id}', '${testRun.projectId}', '${testRun.suiteId}', '${testRun.branch}', 
              '${testRun.buildUrl}', '${testRun.buildActor}', '${testRun.gitSha}', '${testRun.status}',
              '${testRun.startTime.toISOString()}', '${testRun.endTime.toISOString()}', 
              ${testRun.duration}, '${JSON.stringify(testRun.tags)}')
      ON CONFLICT (id) DO NOTHING;
    `);
    
    // Insert spec runs
    for (const specRun of testRun.specRuns) {
      await this.executeSql(`
        INSERT INTO spec_runs (id, test_run_id, spec_description, status, start_time, end_time, duration, error_message, stack_trace)
        VALUES ('${specRun.id}', '${specRun.testRunId}', '${specRun.specDescription}', '${specRun.status}',
                '${specRun.startTime.toISOString()}', '${specRun.endTime.toISOString()}', 
                ${specRun.duration}, ${specRun.errorMessage ? `'${specRun.errorMessage}'` : 'NULL'}, 
                ${specRun.stackTrace ? `'${specRun.stackTrace}'` : 'NULL'})
        ON CONFLICT (id) DO NOTHING;
      `);
    }
  }

  private async waitForDatabase(timeout: number = 60000): Promise<void> {
    console.log(`⏳ Waiting for database to be ready...`);
    
    const startTime = Date.now();
    while (Date.now() - startTime < timeout) {
      try {
        await this.executeSql('SELECT 1;');
        console.log(`✅ Database is ready`);
        return;
      } catch (error) {
        await new Promise(resolve => setTimeout(resolve, 2000));
      }
    }
    
    throw new Error(`Database failed to become ready within ${timeout/1000}s`);
  }

  private async executeSql(sql: string): Promise<void> {
    try {
      // Execute SQL via kubectl
      const command = `kubectl exec -n ${this.namespace} deployment/postgresql -c postgresql -- psql -U postgres -d fern -c "${sql}"`;
      await execAsync(command);
    } catch (error) {
      console.error(`❌ SQL execution failed:`, error);
      throw error;
    }
  }
}