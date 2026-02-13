# Playwright E2E Testing Guide

This directory contains end-to-end tests using Playwright. The setup is designed to work seamlessly with `boardgame-util serve`.

## Quick Start

### Running Tests

```bash
# Terminal 1: Start the dev server
boardgame-util serve

# Terminal 2: Run tests (in server/static directory)
cd server/static
npm run test:e2e              # Run tests headlessly
npm run test:e2e:headed       # Run with visible browser
npm run test:e2e:ui           # Run with Playwright UI
npm run test:e2e:debug        # Run in debug mode
npm run test:e2e:report       # View test results
```

## Configuration

The Playwright configuration is in `playwright.config.ts` with these key settings:

- **baseURL**: `http://localhost:8080` - Connects to Vite server from boardgame-util serve
- **reuseExistingServer**: `true` - Uses existing Vite server, doesn't start/stop it
- **workers**: `1` - Sequential test execution for predictable animations
- **headless**: By default yes, override with `HEADED=1` environment variable

## Test Structure

```
tests/
├── README.md           # This file
├── fixtures.ts         # Helper functions for tests
├── global.d.ts         # TypeScript declarations
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

1. **Always start boardgame-util serve first** - Tests expect the server to be running on port 8080
2. **Single worker mode** - Tests run sequentially to avoid animation timing issues
3. **Headless by default** - Use `HEADED=1` to see the browser
4. **Reuses existing server** - Won't interfere with your running dev server
5. **Screenshots on failure** - Automatically captured to help debugging

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
Make sure `boardgame-util serve` is running before executing tests.

### Tests Timing Out
Increase timeout in `playwright.config.ts` or use `waitForAnimationQueue()` to wait for animations.

### Custom Elements Not Found
Use `waitForCustomElement()` to ensure the element is registered before interacting with it.

## CI/CD Integration

For continuous integration, set the `CI` environment variable:

```bash
CI=1 npm run test:e2e
```

This enables:
- GitHub Actions reporter
- Automatic retries (2 attempts)
- `forbidOnly` check to prevent `.only` in tests
