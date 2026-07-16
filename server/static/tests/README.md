# Playwright Testing Guide

This directory contains two deliberately separate Playwright shards. Renderer
authoring fixtures own their server lifecycle and can run in parallel. The
real-time animation/companion suite reuses a developer server and stays
sequential because animation timing and shared game state make parallelism
misleading.

## Quick Start

### Running Tests

```bash
# Self-contained renderer-authoring fixtures (no server setup required)
cd server/static
npm run test:renderer

# Real-time animation/companion tests: Terminal 1
boardgame-util serve

# Real-time animation/companion tests: Terminal 2
cd server/static
npm run test:e2e              # Run tests headlessly
npm run test:e2e:headed       # Run with visible browser
npm run test:e2e:ui           # Run with Playwright UI
npm run test:e2e:debug        # Run in debug mode
npm run test:e2e:report       # View test results
```

## Configurations

`playwright.renderer.config.ts` is the reliable renderer-authoring gate:

- starts and stops `boardgame-util serve` itself;
- allocates isolated API and Vite ports on `127.0.0.1`;
- uses in-memory storage and offline development mode;
- runs only `tests/renderer`, permits parallel workers, and uses zero retries;
- preserves traces, screenshots, and videos for the first failing attempt.
- runs the pinned axe integration; its deliberately inaccessible control
  fixture proves accessibility violations fail detection instead of silently
  passing because the analyzer was misconfigured.

`src/testing/renderer-fixture.ts` mounts a registered game renderer without a
live game. Define fixtures against the generated `GameClientContract`; that one
type parameter binds the snapshot to the exact generated `State`, complete move
name set, and matching registered renderer tag. The snapshot also carries the
generated move-schema fingerprint, player perspective, legality, outcome, and
surface:

```ts
export const pigRendererFixture = defineRendererFixture<GameClientContract>({
  tagName: 'boardgame-render-game-pig',
  snapshot: { /* state, every move's legality, version, outcome, ... */ },
});
```

Misspelled or omitted move names, another game's renderer tag, and state drift
fail the package-isolated `check-client` gate. At runtime the host rejects stale
fixture schemas, contradictory legality, surface/tag mismatches, malformed
proposals, and unregistered elements immediately. It records valid proposals
with snapshot-version request IDs and supports deterministic snapshot
replacement and cleanup. Pig and Tic-tac-toe keep their `satisfies State`
builders beside their renderers so failures point to creator-owned fixture code
before Chromium runs.

`tests/renderer/renderer-fixture-helpers.ts` supplies canonical 320px, 768px,
and 1280px viewports, keyboard reachability, and console/page-error capture.
Renderer tests run with reduced motion by default; animation-specific behavior
belongs in the separate real-time shard.

`playwright.config.ts` drives the existing real-time suite:

- **baseURL**: `http://localhost:8080` - Connects to Vite server from boardgame-util serve
- **reuseExistingServer**: `true` - Uses existing Vite server, doesn't start/stop it
- **workers**: `1` - Sequential test execution for predictable animations
- **headless**: By default yes, override with `HEADED=1` environment variable

## Test Structure

```
tests/
├── README.md           # This file
├── BASELINE.md         # Known-green gates and classified pre-existing debt
├── fixtures.ts         # Helper functions for tests
├── global.d.ts         # TypeScript declarations
├── renderer/           # Self-contained authoring/integration fixtures
├── basic/              # Basic smoke tests
│   └── homepage.spec.ts
└── navigation/         # Navigation tests (add as needed)
```

## Writing Tests

### Basic Test Example

```typescript
import { test, expect } from '@playwright/test';

test('loads homepage', async ({ page }) => {
  await page.goto('/');
  await expect(page.locator('boardgame-app')).toBeVisible();
});
```

### Using Test Fixtures

The `fixtures.ts` file provides helpful utilities:

```typescript
import { exposeStore, waitForAnimationQueue, waitForAuth } from '../fixtures';

test('game with animations', async ({ page }) => {
  await page.goto('/game/memory');

  // Expose Redux store for inspection
  await exposeStore(page);

  // Wait for auth to initialize
  await waitForAuth(page);

  // Wait for all animations to complete
  await waitForAnimationQueue(page);

  // Now interact with the game
});
```

### Available Fixtures

- `exposeStore(page)` - Makes Redux store available as `window.__TEST_STORE__`
- `getStoreState(page)` - Gets current Redux state
- `waitForAnimationQueue(page)` - Waits for all animations to complete
- `getPendingBundleCount(page)` - Gets count of pending animations
- `waitForAuth(page)` - Waits for authentication to initialize
- `navigateToGame(page, gameName)` - Navigate to a specific game
- `waitForCustomElement(page, tagName)` - Wait for custom element to be defined
- `takeScreenshot(page, name)` - Take a labeled screenshot

## Key Points

1. `test:renderer` starts its own server; do not start one first or hard-code its
   ports in a test.
2. Renderer fixtures must be isolated enough to pass with parallel workers and
   `retries: 0`.
3. Real-time tests require `boardgame-util serve` on the default ports and run
   with one worker.
4. Both shards are headless by default; use `HEADED=1` to see the browser.
5. Failure artifacts are retained for debugging.

## Debugging Tips

### View Test Report
```bash
npm run test:e2e:report
```

### Run with Visible Browser
```bash
npm run test:e2e:headed
```

### Debug Specific Test
```bash
npx playwright test tests/basic/homepage.spec.ts --debug
```

### Check Console Logs
Console output from the browser is captured in `.playwright-mcp/` directory.

## Common Issues

### Port 8080 Not Available
This only affects `test:e2e`. Make sure its manually started server is running
on the default port. `test:renderer` allocates its own ports.

### Tests Timing Out
Increase timeout in `playwright.config.ts` or use `waitForAnimationQueue()` to wait for animations.

### Custom Elements Not Found
Use `waitForCustomElement()` to ensure the element is registered before interacting with it.

## CI/CD Integration

For continuous integration, set the `CI` environment variable:

```bash
CI=1 npm run test:e2e
```

For the legacy real-time shard this enables:
- GitHub Actions reporter
- Automatic retries (2 attempts)
- `forbidOnly` check to prevent `.only` in tests

The renderer shard intentionally keeps retries disabled in CI so flakes are
visible rather than converted into apparent passes.

The repository's `Client quality` workflow installs dependencies with
`npm ci`, installs the lockfile-selected Chromium build, and runs this shard as
an independent zero-retry job. Failure traces, screenshots, and videos are
uploaded for seven days.
