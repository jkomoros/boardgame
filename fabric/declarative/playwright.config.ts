import { defineConfig, devices } from '@playwright/test';

// Own port, so the suite never meets the webapp's dev server (8080) or a
// developer's `npm run dev` here (5197).
const port = Number(process.env.DECLARATIVE_TEST_PORT || 5199);

export default defineConfig({
  testDir: './tests',
  fullyParallel: true,
  retries: 0,
  reporter: 'line',
  use: {
    baseURL: `http://localhost:${port}`,
    ...devices['Desktop Chrome'],
    // Motion is the subject under test: do not let the browser reduce it.
    contextOptions: { reducedMotion: 'no-preference' },
    trace: 'retain-on-failure',
  },
  webServer: {
    command: `npx vite --port ${port} --strictPort`,
    url: `http://localhost:${port}/memory.html`,
    reuseExistingServer: false,
    timeout: 60_000,
  },
  projects: [{ name: 'chromium', use: { ...devices['Desktop Chrome'] } }],
});
