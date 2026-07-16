import { defineConfig, devices } from '@playwright/test';
import { execFileSync } from 'node:child_process';
import { resolve } from 'node:path';
import { tmpdir } from 'node:os';
import { fileURLToPath } from 'node:url';

const repoRoot = fileURLToPath(new URL('../..', import.meta.url));

function allocatePorts(): [number, number] {
  // Playwright loads TypeScript config synchronously, so allocate both ports in
  // a short synchronous child process rather than using top-level await here.
  // The child holds both listeners at once, preventing duplicate allocation.
  const script = `
    const net = require('node:net');
    const servers = [net.createServer(), net.createServer()];
    Promise.all(servers.map(server => new Promise((resolve, reject) => {
      server.once('error', reject);
      server.listen(0, '127.0.0.1', () => resolve(server.address().port));
    }))).then(ports => Promise.all(servers.map(server => new Promise(resolve => server.close(resolve)))).then(() => {
      process.stdout.write(JSON.stringify(ports));
    })).catch(error => { console.error(error); process.exit(1); });
  `;
  return JSON.parse(execFileSync(process.execPath, ['-e', script], { encoding: 'utf8' })) as [number, number];
}

const inheritedStaticPort = Number(process.env.BOARDGAME_RENDERER_STATIC_PORT || 0);
const inheritedApiPort = Number(process.env.BOARDGAME_RENDERER_API_PORT || 0);
const [staticPort, apiPort] = inheritedStaticPort && inheritedApiPort
  ? [inheritedStaticPort, inheritedApiPort]
  : allocatePorts();
// Playwright evaluates the config in worker processes too. Persist the first
// allocation so every worker uses the ports owned by the main webServer.
process.env.BOARDGAME_RENDERER_STATIC_PORT = String(staticPort);
process.env.BOARDGAME_RENDERER_API_PORT = String(apiPort);
const baseURL = `http://127.0.0.1:${staticPort}`;

export default defineConfig({
  testDir: './tests/renderer',
  fullyParallel: true,
  workers: process.env.CI ? 2 : undefined,
  retries: 0,
  forbidOnly: !!process.env.CI,
  reporter: process.env.CI ? 'github' : 'line',
  use: {
    baseURL,
    ...devices['Desktop Chrome'],
    headless: !process.env.HEADED,
    reducedMotion: 'reduce',
    trace: 'retain-on-failure',
    screenshot: 'only-on-failure',
    video: 'retain-on-failure',
  },
  webServer: {
    command: `go run ./boardgame-util serve --storage memory --offline-dev-mode --port ${apiPort} --static-port ${staticPort}`,
    cwd: repoRoot,
    env: {
      ...process.env,
      GOCACHE: process.env.GOCACHE || resolve(tmpdir(), 'boardgame-renderer-go-cache'),
    },
    // Readiness must prove the API is listening through Vite's proxy, not only
    // that Vite has served its first static file.
    url: `${baseURL}/api/list/manager`,
    reuseExistingServer: false,
    timeout: 180_000,
    gracefulShutdown: { signal: 'SIGTERM', timeout: 10_000 },
  },
  projects: [{ name: 'renderer-chromium', use: { ...devices['Desktop Chrome'] } }],
});
