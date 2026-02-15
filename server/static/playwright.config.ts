import { defineConfig, devices } from '@playwright/test';

/**
 * Playwright configuration for boardgame-util serve testing
 *
 * This configuration is designed to work with the existing Vite dev server
 * started by `boardgame-util serve`, which runs on port 8080.
 *
 * CRITICAL: reuseExistingServer is set to true to prevent Playwright from
 * killing the Vite server that boardgame-util started.
 */
export default defineConfig({
  // Test directory
  testDir: './tests',

  // Run tests sequentially to avoid animation timing issues
  fullyParallel: false,

  // Single worker ensures predictable test execution order
  // and prevents race conditions with animations
  workers: 1,

  // Fail the build on CI if you accidentally left test.only in the source code
  forbidOnly: !!process.env.CI,

  // Retry on CI only
  retries: process.env.CI ? 2 : 0,

  // Reporter to use
  reporter: process.env.CI ? 'github' : 'html',

  // Shared settings for all projects
  use: {
    // Base URL to use in actions like `await page.goto('/')`
    baseURL: 'http://localhost:8080',

    // Headless mode by default, can be overridden with HEADED env var
    headless: !process.env.HEADED,

    // Collect trace when retrying the failed test
    trace: 'on-first-retry',

    // Screenshot on failure
    screenshot: 'only-on-failure',

    // Video on failure
    video: 'retain-on-failure',

    // Maximum time each action such as `click()` can take
    actionTimeout: 10000,

    // Maximum time for navigation
    navigationTimeout: 30000,
  },

  // Configure web server that tests will connect to
  webServer: {
    // Connect to the Vite server started by boardgame-util serve
    url: 'http://localhost:8080',

    // CRITICAL: Reuse existing server, don't try to start/stop it
    // This prevents Playwright from killing the Vite server
    reuseExistingServer: true,

    // How long to wait for the server to be ready (2 minutes)
    timeout: 120000,

    // Placeholder command - with reuseExistingServer, Playwright checks
    // the URL first and skips launching if the server is already running
    command: 'echo "Waiting for server"',
  },

  // Configure projects for major browsers
  projects: [
    {
      name: 'chromium',
      use: {
        ...devices['Desktop Chrome'],
        // Use viewport size that matches typical desktop usage
        viewport: { width: 1280, height: 720 },
      },
    },

    // Uncomment to test on Firefox
    // {
    //   name: 'firefox',
    //   use: { ...devices['Desktop Firefox'] },
    // },

    // Uncomment to test on WebKit (Safari)
    // {
    //   name: 'webkit',
    //   use: { ...devices['Desktop Safari'] },
    // },

    // Uncomment to test on mobile viewports
    // {
    //   name: 'Mobile Chrome',
    //   use: { ...devices['Pixel 5'] },
    // },
    // {
    //   name: 'Mobile Safari',
    //   use: { ...devices['iPhone 12'] },
    // },
  ],
});
