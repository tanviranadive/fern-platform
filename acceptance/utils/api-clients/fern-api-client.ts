/**
 * Fern API Client - Unified client for all Fern services
 * 
 * Provides a unified interface for interacting with:
 * - fern-reporter (GraphQL and REST)
 * - fern-mycelium (AI intelligence)
 * - fern-ui (frontend API)
 * 
 * Handles authentication, error handling, and service discovery
 */

import axios, { AxiosInstance, AxiosResponse } from 'axios';

export interface ServiceUrls {
  reporter: string;
  mycelium: string;
  ui: string;
}

export interface TestProject {
  id: string;
  name: string;
  description?: string;
  tags: string[];
  createdAt?: string;
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
  startTime: string;
  endTime: string;
  duration: number;
  tags: string[];
  specRuns?: SpecRun[];
}

export interface SpecRun {
  id: string;
  testRunId: string;
  specDescription: string;
  status: 'passed' | 'failed' | 'skipped';
  startTime: string;
  endTime: string;
  duration: number;
  errorMessage?: string;
  stackTrace?: string;
}

export interface GraphQLResponse<T = any> {
  data?: T;
  errors?: Array<{
    message: string;
    locations?: Array<{ line: number; column: number }>;
    path?: string[];
  }>;
}

export interface TestRunsQueryResult {
  testRuns: TestRun[];
  pageInfo?: {
    hasNextPage: boolean;
    hasPreviousPage: boolean;
    startCursor?: string;
    endCursor?: string;
  };
}

export class ApiClient {
  private reporterClient: AxiosInstance;
  private myceliumClient: AxiosInstance;
  private uiClient: AxiosInstance;
  private serviceUrls: ServiceUrls;

  constructor(serviceUrls: ServiceUrls) {
    this.serviceUrls = serviceUrls;
    
    // Initialize HTTP clients for each service
    this.reporterClient = axios.create({
      baseURL: serviceUrls.reporter,
      timeout: 30000,
      headers: {
        'Content-Type': 'application/json',
        'Accept': 'application/json'
      }
    });

    this.myceliumClient = axios.create({
      baseURL: serviceUrls.mycelium,
      timeout: 30000,
      headers: {
        'Content-Type': 'application/json',
        'Accept': 'application/json'
      }
    });

    this.uiClient = axios.create({
      baseURL: serviceUrls.ui,
      timeout: 30000,
      headers: {
        'Content-Type': 'application/json',
        'Accept': 'application/json'
      }
    });

    // Add response interceptors for error handling
    this.setupInterceptors();
  }

  private setupInterceptors(): void {
    const responseInterceptor = (response: AxiosResponse) => response;
    const errorInterceptor = (error: any) => {
      if (error.response) {
        // Server responded with error status
        return Promise.reject(new Error(`API Error ${error.response.status}: ${error.response.data?.message || error.response.statusText}`));
      } else if (error.request) {
        // Request made but no response
        return Promise.reject(new Error('Network Error: No response from server'));
      } else {
        // Request setup error
        return Promise.reject(new Error(`Request Error: ${error.message}`));
      }
    };

    [this.reporterClient, this.myceliumClient, this.uiClient].forEach(client => {
      client.interceptors.response.use(responseInterceptor, errorInterceptor);
    });
  }

  // Health and Service Discovery
  async healthCheck(): Promise<boolean> {
    try {
      const responses = await Promise.all([
        this.reporterClient.get('/health'),
        this.myceliumClient.get('/health'),
        this.uiClient.get('/api/health')
      ]);
      
      return responses.every(response => response.status === 200);
    } catch (error) {
      return false;
    }
  }

  async checkChatService(): Promise<boolean> {
    try {
      const response = await this.uiClient.get('/api/chat/health');
      return response.status === 200;
    } catch (error) {
      return false;
    }
  }

  // GraphQL Operations
  async graphqlQuery<T = any>(query: string, variables?: Record<string, any>): Promise<GraphQLResponse<T>> {
    const response = await this.reporterClient.post('/query', {
      query,
      variables
    });
    
    return response.data;
  }

  // Test Runs
  async getTestRuns(options?: {
    limit?: number;
    projectId?: string;
    status?: string;
    after?: string;
    before?: string;
  }): Promise<GraphQLResponse<TestRunsQueryResult>> {
    const query = `
      query GetTestRuns(
        $limit: Int,
        $projectId: ID,
        $status: TestStatus,
        $after: String,
        $before: String
      ) {
        testRuns(
          limit: $limit,
          projectId: $projectId,
          status: $status,
          after: $after,
          before: $before
        ) {
          id
          projectId
          suiteId
          branch
          buildUrl
          buildActor
          gitSha
          status
          startTime
          endTime
          duration
          tags
          pageInfo {
            hasNextPage
            hasPreviousPage
            startCursor
            endCursor
          }
        }
      }
    `;
    
    return this.graphqlQuery<TestRunsQueryResult>(query, options);
  }

  async getTestRunById(id: string): Promise<GraphQLResponse<{ testRun: TestRun }>> {
    const query = `
      query GetTestRun($id: ID!) {
        testRun(id: $id) {
          id
          projectId
          suiteId
          branch
          buildUrl
          buildActor
          gitSha
          status
          startTime
          endTime
          duration
          tags
          specRuns {
            id
            specDescription
            status
            startTime
            endTime
            duration
            errorMessage
            stackTrace
          }
        }
      }
    `;
    
    return this.graphqlQuery<{ testRun: TestRun }>(query, { id });
  }

  async createTestRun(testRun: Omit<TestRun, 'id'>): Promise<AxiosResponse> {
    return this.reporterClient.post('/api/test-runs', testRun);
  }

  // Projects
  async getProjects(): Promise<TestProject[]> {
    const response = await this.reporterClient.get('/api/projects');
    return response.data;
  }

  async getProjectById(id: string): Promise<TestProject> {
    const response = await this.reporterClient.get(`/api/projects/${id}`);
    return response.data;
  }

  async createProject(project: Omit<TestProject, 'id' | 'createdAt'>): Promise<AxiosResponse> {
    return this.reporterClient.post('/api/projects', project);
  }

  // AI Intelligence (fern-mycelium)
  async getFlakyTests(projectId: string, limit: number = 10): Promise<any> {
    const query = `
      query GetFlakyTests($projectId: ID!, $limit: Int!) {
        flakyTests(projectID: $projectId, limit: $limit) {
          testID
          testName
          passRate
          failureRate
          lastFailure
          runCount
        }
      }
    `;
    
    return this.myceliumClient.post('/query', {
      query,
      variables: { projectId, limit }
    });
  }

  async sendChatMessage(message: string, sessionId?: string): Promise<any> {
    return this.uiClient.post('/api/chat', {
      message,
      sessionId
    });
  }

  // Test Data Management
  async clearTestData(namespace: string): Promise<void> {
    // This would typically be a privileged operation for testing
    await this.reporterClient.delete(`/api/test-data?namespace=${namespace}`);
  }

  async seedTestData(testData: any): Promise<void> {
    await this.reporterClient.post('/api/test-data/seed', testData);
  }

  // Error Simulation (for testing error handling)
  async simulateFailure(): Promise<void> {
    // Temporarily break the client to simulate failure
    this.reporterClient.defaults.baseURL = 'http://non-existent-server:9999';
    this.myceliumClient.defaults.baseURL = 'http://non-existent-server:9999';
  }

  async restoreConnection(): Promise<void> {
    // Restore proper URLs
    this.reporterClient.defaults.baseURL = this.serviceUrls.reporter;
    this.myceliumClient.defaults.baseURL = this.serviceUrls.mycelium;
  }

  async simulateSlowNetwork(delayMs: number): Promise<void> {
    // Add artificial delay to requests
    const delayInterceptor = (config: any) => {
      return new Promise(resolve => {
        setTimeout(() => resolve(config), delayMs);
      });
    };

    this.reporterClient.interceptors.request.use(delayInterceptor);
    this.myceliumClient.interceptors.request.use(delayInterceptor);
  }

  async simulateLLMFailure(): Promise<void> {
    // Simulate LLM service unavailability
    this.myceliumClient.defaults.baseURL = 'http://non-existent-llm:9999';
  }

  async restoreLLMService(): Promise<void> {
    this.myceliumClient.defaults.baseURL = this.serviceUrls.mycelium;
  }

  async simulateNetworkDisconnection(): Promise<void> {
    // Simulate network issues
    const networkErrorInterceptor = () => {
      return Promise.reject(new Error('Network disconnected'));
    };

    this.reporterClient.interceptors.request.use(networkErrorInterceptor);
    this.myceliumClient.interceptors.request.use(networkErrorInterceptor);
    this.uiClient.interceptors.request.use(networkErrorInterceptor);
  }

  async restoreNetworkConnection(): Promise<void> {
    // Clear error interceptors and restore normal operation
    this.reporterClient.interceptors.request.clear();
    this.myceliumClient.interceptors.request.clear();
    this.uiClient.interceptors.request.clear();
    
    this.setupInterceptors();
  }

  async disableAIFeatures(): Promise<void> {
    // Simulate AI features being disabled
    await this.myceliumClient.post('/api/config', { 
      aiEnabled: false 
    });
  }

  async enableAIFeatures(): Promise<void> {
    await this.myceliumClient.post('/api/config', { 
      aiEnabled: true 
    });
  }

  async injectMalformedData(namespace: string): Promise<void> {
    // Inject malformed test data for error handling tests
    const malformedData = {
      testRuns: [
        {
          // Missing required fields
          id: 'malformed-1',
          status: 'invalid-status'
        },
        {
          // Invalid data types
          id: 123,
          projectId: null,
          duration: 'not-a-number'
        }
      ]
    };
    
    await this.reporterClient.post('/api/test-data/inject-malformed', {
      namespace,
      data: malformedData
    });
  }

  async cleanMalformedData(namespace: string): Promise<void> {
    await this.reporterClient.delete(`/api/test-data/malformed?namespace=${namespace}`);
  }

  // Summary and Reports
  async getProjectSummaries(): Promise<any> {
    const response = await this.reporterClient.get('/api/reports/projects');
    return response.data;
  }

  async getProjectSummary(projectId: string, options?: {
    startDate?: string;
    endDate?: string;
  }): Promise<any> {
    const params = new URLSearchParams();
    if (options?.startDate) params.append('startDate', options.startDate);
    if (options?.endDate) params.append('endDate', options.endDate);
    
    const response = await this.reporterClient.get(
      `/api/reports/summary/${projectId}?${params.toString()}`
    );
    return response.data;
  }

  async getSystemStats(): Promise<any> {
    const response = await this.reporterClient.get('/api/reports/stats');
    return response.data;
  }

  // User Preferences
  async getUserPreferences(userId: string): Promise<any> {
    const response = await this.reporterClient.get(`/api/user-preferences/${userId}`);
    return response.data;
  }

  async setUserPreferences(userId: string, preferences: any): Promise<any> {
    const response = await this.reporterClient.put(`/api/user-preferences/${userId}`, {
      preferences
    });
    return response.data;
  }

  async addFavoriteProject(userId: string, projectId: string): Promise<any> {
    const response = await this.reporterClient.post(`/api/user-preferences/${userId}/favorites`, {
      projectId
    });
    return response.data;
  }

  async removeFavoriteProject(userId: string, projectId: string): Promise<any> {
    const response = await this.reporterClient.delete(
      `/api/user-preferences/${userId}/favorites/${projectId}`
    );
    return response.data;
  }

  // Utility methods
  async waitForService(serviceName: 'reporter' | 'mycelium' | 'ui', timeout: number = 30000): Promise<boolean> {
    const startTime = Date.now();
    const client = serviceName === 'reporter' ? this.reporterClient :
                  serviceName === 'mycelium' ? this.myceliumClient : this.uiClient;
    
    while (Date.now() - startTime < timeout) {
      try {
        await client.get('/health');
        return true;
      } catch (error) {
        await new Promise(resolve => setTimeout(resolve, 1000));
      }
    }
    
    return false;
  }

  async getServiceVersion(serviceName: 'reporter' | 'mycelium' | 'ui'): Promise<string> {
    const client = serviceName === 'reporter' ? this.reporterClient :
                  serviceName === 'mycelium' ? this.myceliumClient : this.uiClient;
    
    const response = await client.get('/info');
    return response.data.version;
  }

  async getServiceMetrics(serviceName: 'reporter' | 'mycelium' | 'ui'): Promise<any> {
    const client = serviceName === 'reporter' ? this.reporterClient :
                  serviceName === 'mycelium' ? this.myceliumClient : this.uiClient;
    
    const response = await client.get('/metrics');
    return response.data;
  }
}