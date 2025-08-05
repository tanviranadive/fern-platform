package acceptance_test

import (
	"fmt"
	"os"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/playwright-community/playwright-go"

	"github.com/guidewire-oss/fern-platform/acceptance/helpers"
)

var _ = Describe("JIRA Connection Management", Label("acceptance", "jira", "e2e"), func() {
	var (
		browser    playwright.Browser
		ctx        playwright.BrowserContext
		page       playwright.Page
		auth       *helpers.LoginHelper
		mockJira   *helpers.MockJiraServer
	)

	BeforeEach(func() {
		var err error

		// Start mock JIRA server
		mockJira = helpers.NewMockJiraServer()

		// Create a new browser for each test
		browser = CreateBrowser()

		// Create browser context options
		contextOptions := playwright.BrowserNewContextOptions{
			BaseURL: playwright.String(baseURL),
		}

		// Add video recording if enabled
		if recordVideo {
			contextOptions.RecordVideo = &playwright.RecordVideo{
				Dir:  "./videos",
				Size: &playwright.Size{Width: 1280, Height: 720},
			}
		}

		ctx, err = browser.NewContext(contextOptions)
		Expect(err).NotTo(HaveOccurred())

		page, err = ctx.NewPage()
		Expect(err).NotTo(HaveOccurred())

		auth = helpers.NewLoginHelper(page, baseURL, username, password)
	})

	AfterEach(func() {
		// Stop mock JIRA server
		if mockJira != nil {
			mockJira.Close()
		}

		// Handle cleanup even if test fails
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("Recovered from panic in AfterEach: %v\n", r)
			}

			// Force cleanup in correct order
			// 1. Close page first
			if page != nil {
				_ = page.Close()
				page = nil
			}

			// 2. Close context
			if ctx != nil {
				_ = ctx.Close()
				ctx = nil
			}

			// 3. Close browser last
			if browser != nil {
				_ = browser.Close()
				browser = nil
			}
		}()

		// Save video if recording is enabled
		if recordVideo && page != nil {
			video := page.Video()
			if video != nil {
				// Get the test name for the video file
				testName := CurrentSpecReport().LeafNodeText
				testName = strings.ReplaceAll(testName, " ", "_")
				testName = strings.ReplaceAll(testName, "/", "-")

				// Save the video with a descriptive name
				timestamp := time.Now().Format("20060102_150405")
				newPath := fmt.Sprintf("../videos/jira/%s_%s.webm", testName, timestamp)

				// Get the original video path before closing
				originalPath, _ := video.Path()

				// Close the page first (not the context) to finalize video
				if page != nil {
					page.Close()
					page = nil
				}

				// Wait a moment for video to be saved
				time.Sleep(1 * time.Second)

				// Create directory if it doesn't exist
				os.MkdirAll("../videos/jira", 0755)

				// Try to move the video file
				if originalPath != "" {
					if err := os.Rename(originalPath, newPath); err != nil {
						GinkgoWriter.Printf("Warning: Could not rename video file: %v\n", err)
					} else {
						GinkgoWriter.Printf("Video saved to: %s\n", newPath)
					}
				}
			}
		}
	})

	Describe("Managing JIRA Connections", func() {
		It("should successfully add a JIRA connection during project creation", func() {
			By("Logging in as a project manager")
			auth.Login()

			By("Navigating to the projects page")
			Expect(page.Goto("/#/projects")).To(Succeed())
			
			By("Clicking the Create Project button")
			createButton := page.Locator("button:has-text('Create Project')")
			Expect(createButton.WaitFor(playwright.LocatorWaitForOptions{
				Timeout: playwright.Float(10000),
			})).To(Succeed())
			Expect(createButton.Click()).To(Succeed())

			By("Waiting for the modal to appear")
			time.Sleep(500 * time.Millisecond)

			By("Filling in project details")
			Expect(page.Fill("input[placeholder='My Project']", "Test JIRA Project")).To(Succeed())
			Expect(page.Fill("textarea[placeholder='Project description...']", "A test project with JIRA integration")).To(Succeed())
			
			By("Enabling JIRA integration")
			jiraCheckbox := page.Locator("input[type='checkbox']").First()
			Expect(jiraCheckbox.Click()).To(Succeed())

			By("Filling in JIRA connection details")
			Expect(page.Fill("input[placeholder='Production JIRA']", "Test JIRA Connection")).To(Succeed())
			Expect(page.Fill("input[placeholder='https://mycompany.atlassian.net']", "http://mock-jira.fern-platform.local:8080")).To(Succeed())
			Expect(page.Fill("input[placeholder='PROJ']", "TEST")).To(Succeed())
			Expect(page.Fill("input[placeholder='user@example.com']", "test@fern.com")).To(Succeed())
			Expect(page.Fill("input[type='password']", "test-api-token-123")).To(Succeed())

			By("Submitting the form")
			submitButton := page.Locator("button[type='submit']")
			Expect(submitButton.Click()).To(Succeed())

			By("Waiting for project creation to complete")
			time.Sleep(2 * time.Second)

			By("Verifying the project was created")
			projectName := page.Locator("text=Test JIRA Project")
			Expect(projectName.WaitFor(playwright.LocatorWaitForOptions{
				Timeout: playwright.Float(10000),
			})).To(Succeed())
		})

		It("should successfully add a JIRA connection to an existing project", func() {
			By("Logging in as a project manager")
			auth.Login()

			By("Navigating to the projects page")
			Expect(page.Goto("/#/projects")).To(Succeed())
			
			By("Waiting for projects to load")
			Expect(page.WaitForSelector("text=Fern Test Automation", playwright.PageWaitForSelectorOptions{
				Timeout: playwright.Float(10000),
			})).NotTo(BeNil())

			By("Clicking the settings button for Fern Test Automation project")
			// Find the row containing "Fern Test Automation" and click its settings button
			projectRow := page.Locator("tr:has-text('Fern Test Automation')")
			Expect(projectRow.Count()).To(BeNumerically(">", 0))
			
			settingsButton := projectRow.Locator("button[title='Project Settings']")
			Expect(settingsButton.Count()).To(BeNumerically(">", 0))
			Expect(settingsButton.Click()).To(Succeed())

			By("Waiting for project settings page to load")
			time.Sleep(1 * time.Second) // Give time for navigation
			currentURL := page.URL()
			Expect(currentURL).To(ContainSubstring("#/project/"))
			Expect(currentURL).To(ContainSubstring("/settings"))

			By("Clicking on the Integrations tab")
			integrationsTab := page.Locator("text=Integrations")
			Expect(integrationsTab.Count()).To(BeNumerically(">", 0))
			Expect(integrationsTab.Click()).To(Succeed())

			By("Clicking Add JIRA Connection button")
			addButton := page.Locator("button:has-text('Add JIRA Connection')")
			Expect(addButton.Count()).To(BeNumerically(">", 0))
			Expect(addButton.Click()).To(Succeed())

			By("Waiting for the modal to appear")
			Expect(page.WaitForSelector(".modal", playwright.PageWaitForSelectorOptions{
				Timeout: playwright.Float(5000),
			})).NotTo(BeNil())

			By("Filling in the connection details")
			Expect(page.Fill("input[name='name']", "Test JIRA Connection")).To(Succeed())
			Expect(page.Fill("input[name='jiraUrl']", "http://mock-jira.fern-platform.local:8080")).To(Succeed())
			Expect(page.SelectOption("select[name='authenticationType']", playwright.SelectOptionValues{Values: &[]string{"api_token"}})).To(Succeed())
			Expect(page.Fill("input[name='projectKey']", "TEST")).To(Succeed())
			Expect(page.Fill("input[name='username']", "test@fern.com")).To(Succeed())
			Expect(page.Fill("input[name='credential']", "test-api-token-123")).To(Succeed())

			By("Clicking the Test Connection button")
			testButton := page.Locator("button:has-text('Test Connection')")
			Expect(testButton.Count()).To(BeNumerically(">", 0))
			Expect(testButton.Click()).To(Succeed())

			By("Waiting for connection test to complete")
			successMessage := page.Locator("text=Connection test successful")
			Expect(successMessage.WaitFor(playwright.LocatorWaitForOptions{
				Timeout: playwright.Float(5000),
			})).To(Succeed())

			By("Saving the connection")
			saveButton := page.Locator("button:has-text('Save Connection')")
			Expect(saveButton.Count()).To(BeNumerically(">", 0))
			Expect(saveButton.Click()).To(Succeed())

			By("Verifying the connection appears in the list")
			connectionItem := page.Locator("text=Test JIRA Connection")
			Expect(connectionItem.WaitFor(playwright.LocatorWaitForOptions{
				Timeout: playwright.Float(5000),
			})).To(Succeed())
			
			// Verify status badge
			activeStatus := page.Locator(".status.active:has-text('Active')")
			Expect(activeStatus.Count()).To(BeNumerically(">", 0))
		})

		It("should successfully edit a JIRA connection", func() {
			By("Setting up an existing connection")
			// First create a connection
			auth.Login()
			Expect(page.Goto("/#/projects")).To(Succeed())
			
			// Navigate to settings and create a connection (simplified for brevity)
			projectRow := page.Locator("tr:has-text('Fern Test Automation')")
			settingsButton := projectRow.Locator("button[title='Project Settings']")
			Expect(settingsButton.Click()).To(Succeed())
			
			Expect(page.WaitForURL("**/project/*/settings")).To(Succeed())
			Expect(page.Locator("text=Integrations").Click()).To(Succeed())
			
			// If no connections exist, add one first
			if count, _ := page.Locator(".connection-item").Count(); count == 0 {
				// Add a connection first
				Expect(page.Locator("button:has-text('Add JIRA Connection')").Click()).To(Succeed())
				Expect(page.Fill("input[name='name']", "Original Connection")).To(Succeed())
				Expect(page.Fill("input[name='jiraUrl']", "http://mock-jira.fern-platform.local:8080")).To(Succeed())
				Expect(page.SelectOption("select[name='authenticationType']", playwright.SelectOptionValues{Values: &[]string{"api_token"}})).To(Succeed())
				Expect(page.Fill("input[name='projectKey']", "TEST")).To(Succeed())
				Expect(page.Fill("input[name='username']", "testuser@example.com")).To(Succeed())
				Expect(page.Fill("input[name='credential']", "test-api-token")).To(Succeed())
				Expect(page.Locator("button:has-text('Save Connection')").Click()).To(Succeed())
				time.Sleep(1 * time.Second) // Wait for save
			}

			By("Clicking edit on the first connection")
			editButton := page.Locator(".connection-item button:has-text('Edit')").First()
			Expect(editButton.Click()).To(Succeed())

			By("Updating the connection name")
			nameInput := page.Locator("input[name='name']")
			Expect(nameInput.Fill("")).To(Succeed())
			Expect(nameInput.Fill("Updated JIRA Connection")).To(Succeed())

			By("Saving the changes")
			Expect(page.Locator("button:has-text('Update Connection')").Click()).To(Succeed())

			By("Verifying the updated name appears")
			Expect(page.Locator("text=Updated JIRA Connection").WaitFor()).To(Succeed())
		})

		It("should successfully delete a JIRA connection", func() {
			By("Setting up an existing connection")
			auth.Login()
			Expect(page.Goto("/#/projects")).To(Succeed())
			
			// Navigate to settings and create a connection
			projectRow := page.Locator("tr:has-text('Fern Test Automation')")
			settingsButton := projectRow.Locator("button[title='Project Settings']")
			Expect(settingsButton.Click()).To(Succeed())
			
			Expect(page.WaitForURL("**/project/*/settings")).To(Succeed())
			Expect(page.Locator("text=Integrations").Click()).To(Succeed())
			
			// Add a connection to delete
			Expect(page.Locator("button:has-text('Add JIRA Connection')").Click()).To(Succeed())
			Expect(page.Fill("input[name='name']", "Connection to Delete")).To(Succeed())
			Expect(page.Fill("input[name='jiraUrl']", "http://mock-jira.fern-platform.local:8080")).To(Succeed())
			Expect(page.SelectOption("select[name='authenticationType']", playwright.SelectOptionValues{Values: &[]string{"api_token"}})).To(Succeed())
			Expect(page.Fill("input[name='projectKey']", "TEST")).To(Succeed())
			Expect(page.Fill("input[name='username']", "test@fern.com")).To(Succeed())
			Expect(page.Fill("input[name='credential']", "test-api-token-123")).To(Succeed())
			Expect(page.Locator("button:has-text('Save Connection')").Click()).To(Succeed())
			time.Sleep(1 * time.Second) // Wait for save

			By("Clicking delete on the connection")
			connectionItem := page.Locator(".connection-item:has-text('Connection to Delete')")
			deleteButton := connectionItem.Locator("button:has-text('Delete')")
			Expect(deleteButton.Click()).To(Succeed())

			By("Confirming the deletion")
			// Handle confirmation dialog
			page.On("dialog", func(dialog playwright.Dialog) {
				Expect(dialog.Accept()).To(Succeed())
			})

			By("Verifying the connection is removed")
			Expect(page.Locator("text=Connection to Delete").WaitFor(playwright.LocatorWaitForOptions{
				State: playwright.WaitForSelectorStateDetached,
				Timeout: playwright.Float(5000),
			})).To(Succeed())
		})

		It("should display error for invalid JIRA URL", func() {
			By("Logging in and navigating to JIRA connections")
			auth.Login()
			Expect(page.Goto("/#/projects")).To(Succeed())
			
			projectRow := page.Locator("tr:has-text('Fern Test Automation')")
			settingsButton := projectRow.Locator("button[title='Project Settings']")
			Expect(settingsButton.Click()).To(Succeed())
			
			Expect(page.WaitForURL("**/project/*/settings")).To(Succeed())
			Expect(page.Locator("text=Integrations").Click()).To(Succeed())
			Expect(page.Locator("button:has-text('Add JIRA Connection')").Click()).To(Succeed())

			By("Entering invalid connection details")
			Expect(page.Fill("input[name='name']", "Invalid Connection")).To(Succeed())
			Expect(page.Fill("input[name='jiraUrl']", "http://invalid-jira-url")).To(Succeed())
			Expect(page.SelectOption("select[name='authenticationType']", playwright.SelectOptionValues{Values: &[]string{"api_token"}})).To(Succeed())
			Expect(page.Fill("input[name='projectKey']", "TEST")).To(Succeed())
			Expect(page.Fill("input[name='username']", "testuser@example.com")).To(Succeed())
			Expect(page.Fill("input[name='credential']", "invalid-token")).To(Succeed())

			By("Testing the connection")
			Expect(page.Locator("button:has-text('Test Connection')").Click()).To(Succeed())

			By("Verifying error message is displayed")
			errorMessage := page.Locator("text=Connection test failed")
			Expect(errorMessage.WaitFor(playwright.LocatorWaitForOptions{
				Timeout: playwright.Float(5000),
			})).To(Succeed())
		})
	})
})