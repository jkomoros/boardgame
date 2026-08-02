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
  // Self-starting renderer contract fixtures have their own parallel shard.
  testIgnore: '**/renderer/**',

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

    // Maximum time for navigation.
    //
    // Deliberately generous, and NOT covered by `test.setTimeout()`. Twenty-three
    // animation specs raise their own timeout to 180s because a full-game scenario
    // legitimately takes minutes -- but `test.setTimeout` governs the TEST clock and
    // leaves navigation on this one, so a `page.goto('/')` still had 30s no matter
    // what the spec asked for. Under a ten-minute suite run the dev server is warm
    // but busy, and that was enough to lose one test per five full runs to a bare
    // navigation timeout with no assertion involved.
    //
    // Raising it cannot mask a product regression: a navigation timeout is not an
    // assertion about behavior, and every real check in these specs runs after the
    // page is up. A genuinely hung page still fails, just later.
    navigationTimeout: 60000,
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
