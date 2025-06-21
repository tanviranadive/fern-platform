/**
 * Test Helpers and Global Setup
 * 
 * Provides:
 * - Custom Jest matchers
 * - Common test utilities
 * - Global test configuration
 * - Performance monitoring
 */

import { jest } from '@jest/globals';

// Extend Jest matchers
declare global {
  namespace jest {
    interface Matchers<R> {
      toBeWithinTimeRange(min: number, max: number): R;
      toHaveValidTestRunStructure(): R;
      toRespondWithinTimeout(timeout: number): R;
      toHaveValidApiResponse(): R;
      toContainTestInsight(): R;
    }
  }
}

// Global test context interface
interface GlobalTestContext {
  clusterManager: any;
  dataFixtures: any;
  namespace: string;
  baseUrls: {
    reporter: string;
    mycelium: string;
    ui: string;
  };
  testId: string;
  startTime: number;
}

declare global {
  var __FERN_TEST_CONTEXT__: GlobalTestContext;
}

// Performance monitoring
export class PerformanceMonitor {
  private measurements: Map<string, number[]> = new Map();
  
  startMeasurement(name: string): () => number {
    const startTime = performance.now();
    return () => {
      const duration = performance.now() - startTime;
      this.addMeasurement(name, duration);
      return duration;
    };
  }
  
  addMeasurement(name: string, duration: number): void {
    if (!this.measurements.has(name)) {
      this.measurements.set(name, []);
    }
    this.measurements.get(name)!.push(duration);
  }
  
  getStats(name: string): { avg: number; min: number; max: number; count: number } | null {
    const measurements = this.measurements.get(name);
    if (!measurements || measurements.length === 0) {
      return null;
    }
    
    return {
      avg: measurements.reduce((a, b) => a + b, 0) / measurements.length,
      min: Math.min(...measurements),
      max: Math.max(...measurements),
      count: measurements.length
    };
  }
  
  getAllStats(): Record<string, ReturnType<typeof this.getStats>> {
    const stats: Record<string, ReturnType<typeof this.getStats>> = {};
    for (const [name] of this.measurements) {
      stats[name] = this.getStats(name);
    }
    return stats;
  }
}

export const performanceMonitor = new PerformanceMonitor();

// Common test utilities
export class TestUtils {
  static getTestContext(): GlobalTestContext {
    if (!global.__FERN_TEST_CONTEXT__) {
      throw new Error('Test context not available. Ensure tests are running with proper setup.');
    }
    return global.__FERN_TEST_CONTEXT__;
  }
  
  static async waitForCondition(
    condition: () => Promise<boolean>, 
    timeout: number = 30000,
    interval: number = 1000
  ): Promise<void> {
    const startTime = Date.now();
    
    while (Date.now() - startTime < timeout) {
      if (await condition()) {
        return;
      }
      await new Promise(resolve => setTimeout(resolve, interval));
    }
    
    throw new Error(`Condition not met within ${timeout}ms`);
  }
  
  static async retryOperation<T>(
    operation: () => Promise<T>,
    maxRetries: number = 3,
    delay: number = 1000
  ): Promise<T> {
    let lastError: Error;
    
    for (let i = 0; i <= maxRetries; i++) {
      try {
        return await operation();
      } catch (error) {
        lastError = error as Error;
        if (i < maxRetries) {
          await new Promise(resolve => setTimeout(resolve, delay * (i + 1)));
        }
      }
    }
    
    throw lastError!;
  }
  
  static generateTestData<T>(
    generator: (index: number) => T,
    count: number
  ): T[] {
    return Array.from({ length: count }, (_, i) => generator(i));
  }
}

// HTTP utilities for API testing
export class HttpUtils {
  static async makeRequest(
    url: string, 
    options: RequestInit = {}
  ): Promise<Response> {
    const endMeasurement = performanceMonitor.startMeasurement(`http_${options.method || 'GET'}_${url}`);
    
    try {
      const response = await fetch(url, {
        headers: {
          'Content-Type': 'application/json',
          ...options.headers
        },
        ...options
      });
      
      endMeasurement();
      return response;
    } catch (error) {
      endMeasurement();
      throw error;
    }
  }
  
  static async makeGraphQLRequest(
    baseUrl: string,
    query: string,
    variables?: Record<string, any>
  ): Promise<any> {
    const response = await this.makeRequest(`${baseUrl}/query`, {
      method: 'POST',
      body: JSON.stringify({ query, variables })
    });
    
    if (!response.ok) {
      throw new Error(`GraphQL request failed: ${response.status} ${response.statusText}`);
    }
    
    return await response.json();
  }
}

// Custom Jest matchers
expect.extend({
  toBeWithinTimeRange(received: number, min: number, max: number) {
    const pass = received >= min && received <= max;
    return {
      message: () => 
        pass 
          ? `Expected ${received} not to be within range ${min}-${max}ms`
          : `Expected ${received} to be within range ${min}-${max}ms`,
      pass
    };
  },
  
  toHaveValidTestRunStructure(received: any) {
    const requiredFields = ['id', 'projectId', 'suiteId', 'status', 'startTime', 'endTime'];
    const hasAllFields = requiredFields.every(field => field in received);
    const hasValidStatus = ['passed', 'failed', 'skipped'].includes(received.status);
    
    const pass = hasAllFields && hasValidStatus;
    return {
      message: () => 
        pass
          ? `Expected object not to have valid test run structure`
          : `Expected object to have valid test run structure with fields: ${requiredFields.join(', ')} and valid status`,
      pass
    };
  },
  
  async toRespondWithinTimeout(received: Promise<any>, timeout: number) {
    const startTime = Date.now();
    let responseTime: number;
    let success = false;
    
    try {
      await received;
      responseTime = Date.now() - startTime;
      success = true;
    } catch (error) {
      responseTime = Date.now() - startTime;
    }
    
    const pass = success && responseTime <= timeout;
    return {
      message: () =>
        pass
          ? `Expected request not to respond within ${timeout}ms (responded in ${responseTime}ms)`
          : `Expected request to respond within ${timeout}ms (took ${responseTime}ms)`,
      pass
    };
  },
  
  toHaveValidApiResponse(received: any) {
    const isObject = typeof received === 'object' && received !== null;
    const hasNoErrors = !received.errors || received.errors.length === 0;
    const hasData = received.data !== undefined;
    
    const pass = isObject && hasNoErrors && hasData;
    return {
      message: () =>
        pass
          ? `Expected response not to be a valid API response`
          : `Expected response to be a valid API response with data and no errors`,
      pass
    };
  },
  
  toContainTestInsight(received: any) {
    const hasInsight = received && (
      received.summary || 
      received.recommendations || 
      received.analysis ||
      received.insights
    );
    
    const pass = !!hasInsight;
    return {
      message: () =>
        pass
          ? `Expected response not to contain test insights`
          : `Expected response to contain test insights (summary, recommendations, analysis, or insights)`,
      pass
    };
  }
});

// Global error handler
process.on('unhandledRejection', (reason, promise) => {
  console.error('Unhandled Rejection at:', promise, 'reason:', reason);
});

// Test timeout warning
const originalTimeout = jest.setTimeout;
jest.setTimeout = (timeout: number) => {
  if (timeout > 120000) { // 2 minutes
    console.warn(`⚠️  Long test timeout set: ${timeout}ms. Consider optimizing test or breaking it down.`);
  }
  return originalTimeout(timeout);
};

// Performance reporting
afterEach(() => {
  const stats = performanceMonitor.getAllStats();
  const slowOperations = Object.entries(stats).filter(([_, stat]) => 
    stat && stat.avg > 5000 // 5 seconds
  );
  
  if (slowOperations.length > 0) {
    console.warn('⚠️  Slow operations detected:', 
      slowOperations.map(([name, stat]) => `${name}: ${stat!.avg.toFixed(2)}ms avg`)
    );
  }
});

// Global teardown logging
afterAll(() => {
  const context = global.__FERN_TEST_CONTEXT__;
  if (context) {
    const totalTime = Date.now() - context.startTime;
    console.log(`🏁 Test suite completed in ${totalTime}ms for ${context.testId}`);
  }
});