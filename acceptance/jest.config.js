/** @type {import('jest').Config} */
module.exports = {
  displayName: 'Fern Platform Acceptance Tests',
  
  // Test environment configuration
  testEnvironment: './setup/test-environment.ts',
  
  // Test file patterns
  testMatch: [
    '<rootDir>/features/**/*.test.ts',
    '<rootDir>/features/**/*.spec.ts'
  ],
  
  // Setup files
  setupFilesAfterEnv: [
    '<rootDir>/setup/test-helpers.ts'
  ],
  
  // Module resolution
  moduleNameMapping: {
    '^@acceptance/(.*)$': '<rootDir>/$1',
    '^@fixtures/(.*)$': '<rootDir>/fixtures/$1',
    '^@utils/(.*)$': '<rootDir>/utils/$1',
    '^@features/(.*)$': '<rootDir>/features/$1'
  },
  
  // TypeScript support
  preset: 'ts-jest',
  extensionsToTreatAsEsm: ['.ts'],
  globals: {
    'ts-jest': {
      useESM: true,
      tsconfig: {
        module: 'esnext'
      }
    }
  },
  
  // Timeout configuration (acceptance tests can take longer)
  testTimeout: 120000, // 2 minutes
  
  // Parallel execution (careful with resource conflicts)
  maxWorkers: process.env.CI ? 1 : '50%',
  
  // Coverage configuration
  collectCoverage: false, // Coverage handled by unit tests
  
  // Reporter configuration
  reporters: [
    'default',
    [
      'jest-html-reporters',
      {
        publicPath: './acceptance-test-reports',
        filename: 'acceptance-test-report.html',
        expand: true,
        hideIcon: false,
        pageTitle: 'Fern Platform Acceptance Test Results',
        logoImgPath: undefined,
        inlineSource: false
      }
    ],
    [
      'jest-junit',
      {
        outputDirectory: './acceptance-test-reports',
        outputName: 'acceptance-junit.xml',
        ancestorSeparator: ' › ',
        uniqueOutputName: 'false',
        suiteNameTemplate: '{filepath}',
        classNameTemplate: '{classname}',
        titleTemplate: '{title}'
      }
    ]
  ],
  
  // Test result processor for custom handling
  testResultsProcessor: './utils/test-results-processor.js',
  
  // Global setup and teardown
  globalSetup: '<rootDir>/setup/global-setup.ts',
  globalTeardown: '<rootDir>/setup/global-teardown.ts',
  
  // Error handling
  errorOnDeprecated: true,
  
  // Verbose output in CI
  verbose: process.env.CI === 'true',
  
  // Test categories via runner configuration
  runner: process.env.TEST_CATEGORY ? 
    `./utils/category-runner.js` : 
    '@jest/runner',
  
  // Environment variables
  setupFiles: ['<rootDir>/setup/env-setup.ts'],
  
  // Transform configuration for ESM modules
  transform: {
    '^.+\\.tsx?$': 'ts-jest'
  },
  
  // Module file extensions
  moduleFileExtensions: ['ts', 'tsx', 'js', 'jsx', 'json'],
  
  // Ignore patterns
  testPathIgnorePatterns: [
    '/node_modules/',
    '/dist/',
    '/build/',
    '/__fixtures__/',
    '/temp/'
  ],
  
  // Force exit after tests complete
  forceExit: true,
  
  // Clear mocks between tests
  clearMocks: true,
  
  // Restore mocks after each test
  restoreMocks: true
};