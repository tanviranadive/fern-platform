/**
 * AI Chatbot - UI Acceptance Tests
 * 
 * Tests the AI-powered chatbot functionality including:
 * - Chat session management
 * - Real-time messaging and WebSocket communication
 * - LLM integration and response handling
 * - Test data queries and insights
 * - Error handling and fallback scenarios
 * - Performance with concurrent sessions
 * 
 * Based on fern-ui chatbot features and fern-mycelium integration
 */

import { TestUtils, HttpUtils, performanceMonitor } from '@acceptance/setup/test-helpers';
import { ChatbotPage } from '@acceptance/utils/page-objects/chatbot-page';
import { ApiClient } from '@acceptance/utils/api-clients/fern-api-client';

describe('AI Chatbot Interface', () => {
  let chatbotPage: ChatbotPage;
  let apiClient: ApiClient;
  let context: any;

  beforeAll(async () => {
    context = TestUtils.getTestContext();
    apiClient = new ApiClient(context.baseUrls);
    chatbotPage = new ChatbotPage(context.baseUrls.ui);
    
    // Ensure services are ready
    await TestUtils.waitForCondition(
      async () => {
        try {
          await apiClient.healthCheck();
          await apiClient.checkChatService();
          return true;
        } catch {
          return false;
        }
      },
      60000
    );
  });

  describe('Chat Session Management', () => {
    test('should open and close chatbot window', async () => {
      await chatbotPage.navigate();
      
      // Initially closed
      expect(await chatbotPage.isChatbotOpen()).toBe(false);
      
      // Open chatbot
      await chatbotPage.openChatbot();
      expect(await chatbotPage.isChatbotOpen()).toBe(true);
      
      // Verify chat window elements
      expect(await chatbotPage.hasChatHeader()).toBe(true);
      expect(await chatbotPage.hasChatInput()).toBe(true);
      expect(await chatbotPage.hasMessageArea()).toBe(true);
      
      // Close chatbot
      await chatbotPage.closeChatbot();
      expect(await chatbotPage.isChatbotOpen()).toBe(false);
    });

    test('should start new chat session', async () => {
      await chatbotPage.navigate();
      await chatbotPage.openChatbot();
      
      // Should show welcome message
      const welcomeMessage = await chatbotPage.getWelcomeMessage();
      expect(welcomeMessage).toBeTruthy();
      expect(welcomeMessage).toContain('Hello');
      
      // Should have empty message history
      const messageCount = await chatbotPage.getMessageCount();
      expect(messageCount).toBe(1); // Only welcome message
    });

    test('should maintain session state across page navigation', async () => {
      await chatbotPage.navigate();
      await chatbotPage.openChatbot();
      
      // Send a message
      await chatbotPage.sendMessage('Hello, can you help me with test analysis?');
      await chatbotPage.waitForResponse();
      
      const initialMessageCount = await chatbotPage.getMessageCount();
      expect(initialMessageCount).toBeGreaterThan(1);
      
      // Navigate to different page
      await chatbotPage.navigateToTestRuns();
      await chatbotPage.waitForPageLoad();
      
      // Open chatbot again
      await chatbotPage.openChatbot();
      
      // Session should be maintained
      const maintainedMessageCount = await chatbotPage.getMessageCount();
      expect(maintainedMessageCount).toBe(initialMessageCount);
    });

    test('should clear chat history when requested', async () => {
      await chatbotPage.navigate();
      await chatbotPage.openChatbot();
      
      // Send some messages
      await chatbotPage.sendMessage('Test message 1');
      await chatbotPage.waitForResponse();
      
      await chatbotPage.sendMessage('Test message 2');
      await chatbotPage.waitForResponse();
      
      const messageCountBeforeClear = await chatbotPage.getMessageCount();
      expect(messageCountBeforeClear).toBeGreaterThan(2);
      
      // Clear chat
      await chatbotPage.clearChat();
      
      // Should only have welcome message
      const messageCountAfterClear = await chatbotPage.getMessageCount();
      expect(messageCountAfterClear).toBe(1);
    });
  });

  describe('Real-time Messaging', () => {
    beforeEach(async () => {
      await chatbotPage.navigate();
      await chatbotPage.openChatbot();
    });

    test('should send and receive messages', async () => {
      const testMessage = 'What are my recent test failures?';
      
      const endMeasurement = performanceMonitor.startMeasurement('message_round_trip');
      
      await chatbotPage.sendMessage(testMessage);
      
      // Should show message immediately
      const lastUserMessage = await chatbotPage.getLastUserMessage();
      expect(lastUserMessage.text).toBe(testMessage);
      expect(lastUserMessage.timestamp).toBeTruthy();
      
      // Should show typing indicator
      expect(await chatbotPage.isTypingIndicatorVisible()).toBe(true);
      
      // Wait for response
      await chatbotPage.waitForResponse();
      
      const roundTripTime = endMeasurement();
      expect(roundTripTime).toBeWithinTimeRange(0, 10000); // 10 second max
      
      // Should receive AI response
      const lastBotMessage = await chatbotPage.getLastBotMessage();
      expect(lastBotMessage.text).toBeTruthy();
      expect(lastBotMessage.text.length).toBeGreaterThan(10);
      
      // Typing indicator should be hidden
      expect(await chatbotPage.isTypingIndicatorVisible()).toBe(false);
    });

    test('should handle message formatting and markdown', async () => {
      await chatbotPage.sendMessage('Show me test statistics with charts');
      await chatbotPage.waitForResponse();
      
      const response = await chatbotPage.getLastBotMessage();
      
      // Should support markdown formatting
      const hasFormatting = await chatbotPage.hasMarkdownFormatting(response.element);
      expect(hasFormatting).toBe(true);
      
      // Should render code blocks properly
      if (response.text.includes('```')) {
        const hasCodeBlock = await chatbotPage.hasCodeBlock(response.element);
        expect(hasCodeBlock).toBe(true);
      }
    });

    test('should handle long messages and scrolling', async () => {
      const longMessage = 'Can you provide a detailed analysis of '.repeat(20) + 'my test suite performance?';
      
      await chatbotPage.sendMessage(longMessage);
      await chatbotPage.waitForResponse();
      
      // Should auto-scroll to latest message
      expect(await chatbotPage.isScrolledToBottom()).toBe(true);
      
      // Send multiple messages to test scrolling
      for (let i = 0; i < 5; i++) {
        await chatbotPage.sendMessage(`Test message ${i + 1}`);
        await chatbotPage.waitForResponse();
      }
      
      // Should still be scrolled to bottom
      expect(await chatbotPage.isScrolledToBottom()).toBe(true);
    });

    test('should show timestamp for messages', async () => {
      await chatbotPage.sendMessage('What time is it?');
      await chatbotPage.waitForResponse();
      
      const userMessage = await chatbotPage.getLastUserMessage();
      const botMessage = await chatbotPage.getLastBotMessage();
      
      expect(userMessage.timestamp).toBeTruthy();
      expect(botMessage.timestamp).toBeTruthy();
      
      // Timestamps should be recent
      const userTime = new Date(userMessage.timestamp);
      const botTime = new Date(botMessage.timestamp);
      const now = new Date();
      
      expect(now.getTime() - userTime.getTime()).toBeLessThan(60000); // Within 1 minute
      expect(now.getTime() - botTime.getTime()).toBeLessThan(60000);
    });
  });

  describe('Test Data Queries and Insights', () => {
    beforeEach(async () => {
      await chatbotPage.navigate();
      await chatbotPage.openChatbot();
    });

    test('should answer questions about test failures', async () => {
      await chatbotPage.sendMessage('What are my recent test failures?');
      await chatbotPage.waitForResponse();
      
      const response = await chatbotPage.getLastBotMessage();
      expect(response.text).toContainTestInsight();
      
      // Should mention specific projects or test names
      const mentionsTests = response.text.toLowerCase().includes('test') ||
                           response.text.toLowerCase().includes('fail') ||
                           response.text.toLowerCase().includes('error');
      expect(mentionsTests).toBe(true);
    });

    test('should provide flaky test analysis', async () => {
      await chatbotPage.sendMessage('Which tests are flaky?');
      await chatbotPage.waitForResponse();
      
      const response = await chatbotPage.getLastBotMessage();
      expect(response.text).toContainTestInsight();
      
      // Should mention flakiness concepts
      const mentionsFlakiness = response.text.toLowerCase().includes('flaky') ||
                               response.text.toLowerCase().includes('intermittent') ||
                               response.text.toLowerCase().includes('unstable');
      expect(mentionsFlakiness).toBe(true);
    });

    test('should analyze test performance trends', async () => {
      await chatbotPage.sendMessage('Show me test performance trends over the last week');
      await chatbotPage.waitForResponse();
      
      const response = await chatbotPage.getLastBotMessage();
      expect(response.text).toContainTestInsight();
      
      // Should mention performance or timing
      const mentionsPerformance = response.text.toLowerCase().includes('performance') ||
                                 response.text.toLowerCase().includes('duration') ||
                                 response.text.toLowerCase().includes('time') ||
                                 response.text.toLowerCase().includes('speed');
      expect(mentionsPerformance).toBe(true);
    });

    test('should provide project-specific insights', async () => {
      // Get available projects from test data
      const projects = await apiClient.getProjects();
      expect(projects.length).toBeGreaterThan(0);
      
      const firstProject = projects[0];
      
      await chatbotPage.sendMessage(`Tell me about the ${firstProject.name} project tests`);
      await chatbotPage.waitForResponse();
      
      const response = await chatbotPage.getLastBotMessage();
      expect(response.text).toContainTestInsight();
      expect(response.text).toContain(firstProject.name);
    });

    test('should handle complex multi-part queries', async () => {
      const complexQuery = `I need help understanding:
      1. Which tests failed in the last 24 hours?
      2. Are there any patterns in the failures?
      3. What should I prioritize fixing first?`;
      
      await chatbotPage.sendMessage(complexQuery);
      await chatbotPage.waitForResponse();
      
      const response = await chatbotPage.getLastBotMessage();
      expect(response.text).toContainTestInsight();
      expect(response.text.length).toBeGreaterThan(100); // Substantial response
      
      // Should address multiple aspects
      const addressesFailures = response.text.toLowerCase().includes('fail');
      const addressesPatterns = response.text.toLowerCase().includes('pattern');
      const addressesPriority = response.text.toLowerCase().includes('priorit');
      
      expect(addressesFailures || addressesPatterns || addressesPriority).toBe(true);
    });

    test('should provide actionable recommendations', async () => {
      await chatbotPage.sendMessage('What should I do to improve my test reliability?');
      await chatbotPage.waitForResponse();
      
      const response = await chatbotPage.getLastBotMessage();
      expect(response.text).toContainTestInsight();
      
      // Should contain actionable language
      const hasActionableContent = response.text.toLowerCase().includes('should') ||
                                  response.text.toLowerCase().includes('recommend') ||
                                  response.text.toLowerCase().includes('suggest') ||
                                  response.text.toLowerCase().includes('consider');
      expect(hasActionableContent).toBe(true);
    });
  });

  describe('Error Handling and Fallback Scenarios', () => {
    beforeEach(async () => {
      await chatbotPage.navigate();
      await chatbotPage.openChatbot();
    });

    test('should handle LLM service unavailability', async () => {
      // Simulate LLM service failure
      await apiClient.simulateLLMFailure();
      
      await chatbotPage.sendMessage('Help me analyze my tests');
      await chatbotPage.waitForResponse();
      
      const response = await chatbotPage.getLastBotMessage();
      expect(response.text).toContain('currently unavailable');
      expect(response.text).toContain('try again');
      
      // Restore service
      await apiClient.restoreLLMService();
    });

    test('should handle network connection issues', async () => {
      await chatbotPage.sendMessage('What are my test results?');
      
      // Simulate network disconnection during response
      await apiClient.simulateNetworkDisconnection();
      
      // Should show connection error
      const errorMessage = await chatbotPage.getErrorMessage();
      expect(errorMessage).toContain('connection');
      
      // Should provide retry option
      const retryButton = await chatbotPage.getRetryButton();
      expect(retryButton).toBeTruthy();
      
      // Restore connection
      await apiClient.restoreNetworkConnection();
    });

    test('should handle invalid or unclear queries gracefully', async () => {
      const unclearQueries = [
        'abcdefg',
        '???',
        'help me with the thing',
        ''
      ];
      
      for (const query of unclearQueries) {
        await chatbotPage.sendMessage(query);
        await chatbotPage.waitForResponse();
        
        const response = await chatbotPage.getLastBotMessage();
        
        // Should respond politely and ask for clarification
        const isPoliteResponse = response.text.toLowerCase().includes('clarif') ||
                               response.text.toLowerCase().includes('specific') ||
                               response.text.toLowerCase().includes('help you');
        expect(isPoliteResponse).toBe(true);
      }
    });

    test('should handle rate limiting gracefully', async () => {
      // Send many messages rapidly to trigger rate limiting
      const promises = [];
      for (let i = 0; i < 10; i++) {
        promises.push(chatbotPage.sendMessage(`Test message ${i}`));
      }
      
      await Promise.all(promises);
      
      // Should show rate limit message
      const rateLimitMessage = await chatbotPage.getRateLimitMessage();
      if (rateLimitMessage) {
        expect(rateLimitMessage).toContain('rate limit');
        expect(rateLimitMessage).toContain('please wait');
      }
    });

    test('should provide fallback responses when AI is unavailable', async () => {
      await apiClient.disableAIFeatures();
      
      await chatbotPage.sendMessage('Analyze my test results');
      await chatbotPage.waitForResponse();
      
      const response = await chatbotPage.getLastBotMessage();
      
      // Should provide basic help without AI insights
      expect(response.text).toBeTruthy();
      expect(response.text).toContain('help');
      
      await apiClient.enableAIFeatures();
    });
  });

  describe('WebSocket Communication', () => {
    beforeEach(async () => {
      await chatbotPage.navigate();
      await chatbotPage.openChatbot();
    });

    test('should establish WebSocket connection', async () => {
      // Verify WebSocket connection is established
      const isConnected = await chatbotPage.isWebSocketConnected();
      expect(isConnected).toBe(true);
      
      // Should show connection status
      const connectionStatus = await chatbotPage.getConnectionStatus();
      expect(connectionStatus).toBe('connected');
    });

    test('should handle WebSocket reconnection', async () => {
      // Simulate connection drop
      await chatbotPage.simulateConnectionDrop();
      
      // Should show disconnected state
      expect(await chatbotPage.getConnectionStatus()).toBe('disconnected');
      
      // Should attempt reconnection
      await TestUtils.waitForCondition(
        async () => (await chatbotPage.getConnectionStatus()) === 'reconnecting',
        5000
      );
      
      // Should eventually reconnect
      await TestUtils.waitForCondition(
        async () => (await chatbotPage.getConnectionStatus()) === 'connected',
        15000
      );
    });

    test('should queue messages during disconnection', async () => {
      // Simulate disconnection
      await chatbotPage.simulateConnectionDrop();
      
      // Send messages while disconnected
      await chatbotPage.sendMessage('Message during disconnection 1');
      await chatbotPage.sendMessage('Message during disconnection 2');
      
      // Messages should be queued
      const queuedMessages = await chatbotPage.getQueuedMessageCount();
      expect(queuedMessages).toBe(2);
      
      // Restore connection
      await chatbotPage.restoreConnection();
      
      // Queued messages should be sent
      await TestUtils.waitForCondition(
        async () => (await chatbotPage.getQueuedMessageCount()) === 0,
        10000
      );
    });

    test('should handle message ordering correctly', async () => {
      const messages = [
        'Message 1',
        'Message 2',
        'Message 3'
      ];
      
      // Send messages rapidly
      for (const message of messages) {
        await chatbotPage.sendMessage(message);
      }
      
      // Wait for all responses
      await chatbotPage.waitForMessageCount(8); // 3 user + 3 bot + 1 welcome + 1 initial
      
      const allMessages = await chatbotPage.getAllMessages();
      
      // Verify order is maintained
      let messageIndex = 0;
      for (const message of allMessages) {
        if (message.type === 'user' && messages.includes(message.text)) {
          expect(message.text).toBe(messages[messageIndex]);
          messageIndex++;
        }
      }
    });
  });

  describe('Performance and Concurrent Sessions', () => {
    test('should handle multiple concurrent chat sessions', async () => {
      const sessionCount = 3;
      const chatbotPages = [];
      
      // Open multiple sessions
      for (let i = 0; i < sessionCount; i++) {
        const page = new ChatbotPage(context.baseUrls.ui);
        await page.navigate();
        await page.openChatbot();
        chatbotPages.push(page);
      }
      
      // Send messages from all sessions simultaneously
      const sendPromises = chatbotPages.map((page, index) => 
        page.sendMessage(`Test message from session ${index + 1}`)
      );
      
      await Promise.all(sendPromises);
      
      // Wait for responses in all sessions
      const responsePromises = chatbotPages.map(page => page.waitForResponse());
      await Promise.all(responsePromises);
      
      // Verify all sessions received responses
      for (const page of chatbotPages) {
        const lastResponse = await page.getLastBotMessage();
        expect(lastResponse.text).toBeTruthy();
      }
      
      // Cleanup
      for (const page of chatbotPages) {
        await page.closeChatbot();
      }
    });

    test('should maintain performance with chat history', async () => {
      await chatbotPage.navigate();
      await chatbotPage.openChatbot();
      
      // Build up chat history
      for (let i = 0; i < 20; i++) {
        await chatbotPage.sendMessage(`Test message ${i + 1}`);
        await chatbotPage.waitForResponse();
      }
      
      const endMeasurement = performanceMonitor.startMeasurement('chat_with_history');
      
      // Send another message
      await chatbotPage.sendMessage('Final test message');
      await chatbotPage.waitForResponse();
      
      const responseTime = endMeasurement();
      
      // Should still be responsive with long history
      expect(responseTime).toBeWithinTimeRange(0, 15000); // 15 second max
    });

    test('should handle large message payloads', async () => {
      await chatbotPage.navigate();
      await chatbotPage.openChatbot();
      
      // Send very long message
      const longMessage = 'Please analyze '.repeat(100) + 'my test suite in detail';
      
      const endMeasurement = performanceMonitor.startMeasurement('large_message_handling');
      
      await chatbotPage.sendMessage(longMessage);
      await chatbotPage.waitForResponse();
      
      const handlingTime = endMeasurement();
      
      expect(handlingTime).toBeWithinTimeRange(0, 20000); // 20 second max
      
      const response = await chatbotPage.getLastBotMessage();
      expect(response.text).toBeTruthy();
    });
  });

  describe('UI/UX and Accessibility', () => {
    beforeEach(async () => {
      await chatbotPage.navigate();
      await chatbotPage.openChatbot();
    });

    test('should support keyboard navigation', async () => {
      // Focus on chat input
      await chatbotPage.focusChatInput();
      
      // Type message
      await chatbotPage.typeMessage('Test keyboard input');
      
      // Send with Enter key
      await chatbotPage.pressKey('Enter');
      
      // Should send message
      const lastMessage = await chatbotPage.getLastUserMessage();
      expect(lastMessage.text).toBe('Test keyboard input');
      
      // Test Escape to close
      await chatbotPage.pressKey('Escape');
      expect(await chatbotPage.isChatbotOpen()).toBe(false);
    });

    test('should support screen readers', async () => {
      // Check ARIA labels and roles
      const chatRole = await chatbotPage.getChatRole();
      expect(chatRole).toBe('log');
      
      const inputAriaLabel = await chatbotPage.getInputAriaLabel();
      expect(inputAriaLabel).toContain('message');
      
      // Send message and check message accessibility
      await chatbotPage.sendMessage('Accessibility test');
      await chatbotPage.waitForResponse();
      
      const lastUserMessage = await chatbotPage.getLastUserMessage();
      const lastBotMessage = await chatbotPage.getLastBotMessage();
      
      expect(await chatbotPage.hasAriaLabel(lastUserMessage.element)).toBe(true);
      expect(await chatbotPage.hasAriaLabel(lastBotMessage.element)).toBe(true);
    });

    test('should work in different themes', async () => {
      // Test light theme
      await chatbotPage.setTheme('light');
      await chatbotPage.refresh();
      await chatbotPage.openChatbot();
      
      expect(await chatbotPage.getChatTheme()).toBe('light');
      
      // Test dark theme
      await chatbotPage.setTheme('dark');
      await chatbotPage.refresh();
      await chatbotPage.openChatbot();
      
      expect(await chatbotPage.getChatTheme()).toBe('dark');
      
      // Functionality should work in both themes
      await chatbotPage.sendMessage('Theme test message');
      await chatbotPage.waitForResponse();
      
      const response = await chatbotPage.getLastBotMessage();
      expect(response.text).toBeTruthy();
    });

    test('should be responsive on different screen sizes', async () => {
      const viewports = [
        { width: 1920, height: 1080, name: 'desktop' },
        { width: 768, height: 1024, name: 'tablet' },
        { width: 375, height: 667, name: 'mobile' }
      ];
      
      for (const viewport of viewports) {
        await chatbotPage.setViewportSize(viewport.width, viewport.height);
        await chatbotPage.openChatbot();
        
        // Should be usable at all sizes
        expect(await chatbotPage.isChatInputVisible()).toBe(true);
        expect(await chatbotPage.isMessageAreaVisible()).toBe(true);
        
        // Send test message
        await chatbotPage.sendMessage(`Test on ${viewport.name}`);
        await chatbotPage.waitForResponse();
        
        const response = await chatbotPage.getLastBotMessage();
        expect(response.text).toBeTruthy();
        
        await chatbotPage.closeChatbot();
      }
    });
  });
});