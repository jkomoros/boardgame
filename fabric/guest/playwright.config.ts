import { defineConfig, devices } from '@playwright/test';

// A port of its own, so this never collides with the webapp's suites
// (8080/8888) or another session's dev server.
const port = Number(process.env.GUEST_PORT || 8719);

export default defineConfig({
  testDir: './tests',
  timeout: 120_000,
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: 0,
  reporter: [['list']],
  use: {
    baseURL: `http://127.0.0.1:${port}`,
    viewport: { width: 1300, height: 900 },
    trace: 'retain-on-failure',
  },
  projects: [{ name: 'chromium', use: { ...devices['Desktop Chrome'], viewport: { width: 1300, height: 900 } } }],
  webServer: {
    command: 'node tools/serve.mjs',
    url: `http://127.0.0.1:${port}/fabric/guest/harness.html`,
    reuseExistingServer: true,
    env: { GUEST_PORT: String(port) },
  },
});
