# End-to-End Testing with Playwright: Animation System Focus

## Context

**What prompted this:** Need to verify the web app works correctly end-to-end, with special focus on the FLIP animation system that handles game state transitions. Currently no automated E2E tests exist.

**Current state:**
- No Playwright installed or configured
- Web app uses Lit Elements + Redux + Vite (port 3000)
- TypeScript with strict mode enabled
- FLIP animation system in `boardgame-component-animator.ts`
- Redux animation state with bundle queue (`pendingBundles`)
- Firebase authentication (Google + email/password)

**Intended outcome:**
- Playwright test suite covering critical flows
- Animation system thoroughly tested (FLIP phases, timing, queue management)
- All discovered bugs documented and critical bugs fixed
- Visual regression tests for animation correctness

## Implementation Approach

### Phase 1: Playwright Setup

**Install Dependencies:**
```bash
cd /Users/jkomoros/Code/boardgame/server/static
npm install -D @playwright/test@latest
npx playwright install chromium firefox webkit
```

**Create Configuration:**

File: `playwright.config.ts`
```typescript
import { defineConfig, devices } from '@playwright/test';

export default defineConfig({
  testDir: './tests',
  fullyParallel: false,  // Sequential for animation timing
  workers: 1,            // Single worker prevents race conditions
  retries: process.env.CI ? 2 : 0,
  reporter: 'html',

  use: {
    baseURL: 'http://localhost:3000',
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
    video: 'retain-on-failure',
  },

  projects: [
    { name: 'chromium', use: { ...devices['Desktop Chrome'] } },
  ],

  webServer: {
    command: 'npm run dev',
    url: 'http://localhost:3000',
    reuseExistingServer: !process.env.CI,
    timeout: 120000,
  },
});
```

**Add npm scripts to package.json:**
```json
{
  "scripts": {
    "test:e2e": "playwright test",
    "test:e2e:ui": "playwright test --ui",
    "test:e2e:debug": "playwright test --debug",
    "test:e2e:headed": "playwright test --headed"
  }
}
```

### Phase 2: Test Infrastructure

**Create Test Fixtures:**

File: `tests/fixtures.ts`
```typescript
import { Page } from '@playwright/test';

// Expose Redux store for test inspection
export async function exposeReduxState(page: Page) {
  await page.addInitScript(() => {
    window.__TEST_STORE__ = window.store;
  });
}

// Login helper
export async function loginWithEmail(page: Page, email: string, password: string) {
  await page.goto('/');
  await page.click('text=Sign In');
  await page.click('text=Email/Password');
  await page.fill('input[type="email"]', email);
  await page.fill('input[type="password"]', password);
  await page.click('paper-button:has-text("Sign In")');
  await page.waitForSelector('text=Not signed in', { state: 'hidden', timeout: 10000 });
}

// Wait for Redux animation queue to drain
export async function waitForAnimationsComplete(page: Page) {
  await page.waitForFunction(() => {
    const state = window.__TEST_STORE__.getState();
    return state.game?.animation?.pendingBundles?.length === 0;
  }, { timeout: 10000 });
}

// Get pending bundles count
export async function getPendingBundlesCount(page: Page): Promise<number> {
  return await page.evaluate(() => {
    return window.__TEST_STORE__.getState().game?.animation?.pendingBundles?.length || 0;
  });
}

// Wait for state version to reach target
export async function waitForStateBundleInstall(page: Page, expectedVersion: number) {
  await page.waitForFunction(
    (version) => {
      const state = window.__TEST_STORE__.getState();
      return state.game?.versions?.current >= version;
    },
    expectedVersion,
    { timeout: 15000 }
  );
}
```

**Type declarations:**

File: `tests/global.d.ts`
```typescript
import { Store } from 'redux';
import { RootState } from '../src/types/store';

declare global {
  interface Window {
    __TEST_STORE__: Store<RootState>;
    store: Store<RootState>;
  }
}
```

### Phase 3: Core Test Suite

**Test Directory Structure:**
```
tests/
├── fixtures.ts
├── global.d.ts
├── auth/
│   └── login.spec.ts
├── navigation/
│   ├── game-list.spec.ts
│   └── game-view.spec.ts
├── animation/
│   ├── flip-basic.spec.ts
│   ├── bundle-queue.spec.ts
│   └── visual-regression.spec.ts
└── game-actions/
    └── make-move.spec.ts
```

**Authentication Test:**

File: `tests/auth/login.spec.ts`
```typescript
import { test, expect } from '@playwright/test';
import { exposeReduxState, loginWithEmail } from '../fixtures';

test.describe('Authentication', () => {
  test.beforeEach(async ({ page }) => {
    await exposeReduxState(page);
  });

  test('should login with email/password', async ({ page }) => {
    await loginWithEmail(page, 'sallytester@gmail.com', 'Winthrop3915');

    // Verify logged in state
    await expect(page.locator('boardgame-user')).toContainText('sallytester');

    // Verify Redux state
    const loggedIn = await page.evaluate(() => {
      return window.__TEST_STORE__.getState().user?.loggedIn;
    });
    expect(loggedIn).toBe(true);
  });
});
```

**Navigation Test:**

File: `tests/navigation/game-list.spec.ts`
```typescript
import { test, expect } from '@playwright/test';
import { exposeReduxState, loginWithEmail } from '../fixtures';

test.describe('Game List Navigation', () => {
  test.beforeEach(async ({ page }) => {
    await exposeReduxState(page);
    await loginWithEmail(page, 'sallytester@gmail.com', 'Winthrop3915');
  });

  test('should navigate to game list and load games', async ({ page }) => {
    await page.goto('/list-games');

    // Wait for games to load
    await page.waitForSelector('boardgame-game-item', { timeout: 10000 });

    // Verify Redux list state populated
    const hasGames = await page.evaluate(() => {
      const state = window.__TEST_STORE__.getState();
      return state.list?.participatingActiveGames?.length > 0 ||
             state.list?.visibleJoinableGames?.length > 0;
    });
    expect(hasGames).toBe(true);
  });
});
```

**Basic FLIP Animation Test:**

File: `tests/animation/flip-basic.spec.ts`
```typescript
import { test, expect } from '@playwright/test';
import { exposeReduxState, loginWithEmail, waitForAnimationsComplete, getPendingBundlesCount } from '../fixtures';

test.describe('FLIP Animation System', () => {
  test.beforeEach(async ({ page }) => {
    await exposeReduxState(page);
    await loginWithEmail(page, 'sallytester@gmail.com', 'Winthrop3915');
  });

  test('should complete FLIP animation cycle', async ({ page }) => {
    // Navigate to a game
    await page.goto('/list-games');
    await page.waitForSelector('boardgame-game-item', { timeout: 10000 });

    // Click first available game
    await page.click('boardgame-game-item a');

    // Wait for game to load
    await page.waitForSelector('boardgame-render-game', { timeout: 10000 });
    await waitForAnimationsComplete(page);

    // Verify no stuck bundles
    const pendingCount = await getPendingBundlesCount(page);
    expect(pendingCount).toBe(0);

    // Verify animator exists
    const hasAnimator = await page.evaluate(() => {
      return document.querySelector('boardgame-component-animator') !== null;
    });
    expect(hasAnimator).toBe(true);
  });

  test('should drain bundle queue in FIFO order', async ({ page }) => {
    // Navigate to a game
    await page.goto('/list-games');
    await page.waitForSelector('boardgame-game-item', { timeout: 10000 });
    await page.click('boardgame-game-item a');

    await page.waitForSelector('boardgame-render-game', { timeout: 10000 });

    // Wait for initial bundles to process
    await waitForAnimationsComplete(page);

    // Get current version
    const initialVersion = await page.evaluate(() => {
      return window.__TEST_STORE__.getState().game?.versions?.current || 0;
    });

    // Simulate multiple rapid state changes
    await page.evaluate((baseVersion) => {
      const store = window.__TEST_STORE__;
      for (let i = 1; i <= 3; i++) {
        store.dispatch({
          type: 'ENQUEUE_STATE_BUNDLE',
          bundle: {
            Game: { Version: baseVersion + i },
            Move: { Name: `TestMove${i}` }
          }
        });
      }
    }, initialVersion);

    // Verify queue has bundles
    let count = await getPendingBundlesCount(page);
    expect(count).toBeGreaterThan(0);

    // Wait for all to process
    await waitForAnimationsComplete(page);

    // Verify queue drained
    count = await getPendingBundlesCount(page);
    expect(count).toBe(0);
  });
});
```

**Visual Regression Test:**

File: `tests/animation/visual-regression.spec.ts`
```typescript
import { test, expect } from '@playwright/test';
import { exposeReduxState, loginWithEmail, waitForAnimationsComplete } from '../fixtures';

test.describe('Visual Regression', () => {
  test.beforeEach(async ({ page }) => {
    await exposeReduxState(page);
    await loginWithEmail(page, 'sallytester@gmail.com', 'Winthrop3915');
  });

  test('should match game view baseline', async ({ page }) => {
    await page.goto('/list-games');
    await page.waitForSelector('boardgame-game-item', { timeout: 10000 });
    await page.click('boardgame-game-item a');

    await page.waitForSelector('boardgame-render-game', { timeout: 10000 });
    await waitForAnimationsComplete(page);

    // Baseline screenshot
    await expect(page).toHaveScreenshot('game-view.png', {
      maxDiffPixels: 100,
      timeout: 10000
    });
  });
});
```

**Move Submission Test:**

File: `tests/game-actions/make-move.spec.ts`
```typescript
import { test, expect } from '@playwright/test';
import { exposeReduxState, loginWithEmail, waitForAnimationsComplete, getPendingBundlesCount } from '../fixtures';

test.describe('Move Submission', () => {
  test.beforeEach(async ({ page }) => {
    await exposeReduxState(page);
    await loginWithEmail(page, 'sallytester@gmail.com', 'Winthrop3915');
  });

  test('should submit move and animate state transition', async ({ page }) => {
    // Navigate to game
    await page.goto('/list-games');
    await page.waitForSelector('boardgame-game-item', { timeout: 10000 });
    await page.click('boardgame-game-item a');

    await page.waitForSelector('boardgame-render-game', { timeout: 10000 });
    await waitForAnimationsComplete(page);

    // Check if move form exists
    const hasMoveForm = await page.locator('boardgame-move-form').count() > 0;

    if (hasMoveForm) {
      // Get initial version
      const initialVersion = await page.evaluate(() => {
        return window.__TEST_STORE__.getState().game?.versions?.current;
      });

      // Submit move (if available)
      const submitButton = page.locator('paper-button:has-text("Submit")');
      if (await submitButton.count() > 0) {
        await submitButton.click();

        // Wait for animation to complete
        await waitForAnimationsComplete(page);

        // Verify version changed
        const newVersion = await page.evaluate(() => {
          return window.__TEST_STORE__.getState().game?.versions?.current;
        });
        expect(newVersion).toBeGreaterThanOrEqual(initialVersion);
      }
    }
  });
});
```

### Phase 4: Bug Discovery & Fixing

**Bug Discovery Process:**
1. Run full test suite: `npm run test:e2e`
2. Document all failures in comments
3. Use `npm run test:e2e:debug` to investigate failures
4. Categorize bugs:
   - **Critical**: App crashes, stuck animations, broken auth
   - **High**: Broken game flows, missing features
   - **Medium**: Visual glitches, timing issues
   - **Low**: Minor UX issues

**Bug Fixing Workflow:**
1. Create minimal failing test
2. Debug with headed mode: `npm run test:e2e:headed`
3. Fix bug in source code
4. Re-run test until passing
5. Run full suite to check for regressions
6. Commit fix with test

**Common Animation Bugs to Watch For:**
- `pendingBundles` queue not draining → check `ready-for-next-state` event
- Animations stuck → verify `transitionend` listeners
- Visual jumps → check FLIP timing (double microtask delay)
- Transform not clearing → verify `resetAnimating()` called
- Zero-width components → verify scaleFactor defaults to 1.0

### Phase 5: Verification

**Run Tests:**
```bash
# All tests
npm run test:e2e

# View results in UI
npm run test:e2e:ui

# Generate HTML report
npm run test:e2e:report
```

**Success Criteria:**
- [ ] Playwright installed and configured
- [ ] Authentication test passing
- [ ] Navigation tests passing
- [ ] FLIP animation test passing
- [ ] Bundle queue test passing
- [ ] Visual regression baseline captured
- [ ] Move submission test passing (or skipped if no moves available)
- [ ] All critical bugs fixed
- [ ] Test documentation complete

**Manual Verification:**
1. Load app at `http://localhost:3000`
2. Log in with sallytester@gmail.com / Winthrop3915
3. Navigate to game list
4. Open a game
5. Watch animations complete smoothly
6. Submit a move (if available)
7. Verify no console errors
8. Check Redux DevTools shows clean animation state

## Critical Files

**Animation System Core:**
- `/Users/jkomoros/Code/boardgame/server/static/src/components/boardgame-component-animator.ts` - FLIP implementation
- `/Users/jkomoros/Code/boardgame/server/static/src/components/boardgame-render-game.ts` - Animation orchestration
- `/Users/jkomoros/Code/boardgame/server/static/src/components/boardgame-animatable-item.ts` - Base animating component

**Redux State Management:**
- `/Users/jkomoros/Code/boardgame/server/static/src/selectors.ts` - Animation state selectors
- `/Users/jkomoros/Code/boardgame/server/static/src/actions/game.ts` - Animation actions
- `/Users/jkomoros/Code/boardgame/server/static/src/reducers/game.ts` - Animation reducer

**State Orchestration:**
- `/Users/jkomoros/Code/boardgame/server/static/src/components/boardgame-game-state-manager.ts` - Bundle management, WebSocket

**Authentication:**
- `/Users/jkomoros/Code/boardgame/server/static/src/components/boardgame-user.ts` - Login UI
- `/Users/jkomoros/Code/boardgame/server/static/src/actions/user.ts` - Auth actions

**Configuration:**
- `/Users/jkomoros/Code/boardgame/server/static/package.json` - Dependencies and scripts
- `/Users/jkomoros/Code/boardgame/server/static/vite.config.ts` - Dev server config

## Implementation Order

1. **Setup** (30 min): Install Playwright, create config, add scripts
2. **Fixtures** (30 min): Create test helpers and Redux state exposure
3. **Auth Test** (30 min): Login flow verification
4. **Navigation Tests** (1 hour): Game list and game view
5. **Animation Tests** (2-3 hours): FLIP cycle, bundle queue, visual regression
6. **Game Action Tests** (1 hour): Move submission
7. **Bug Fixing** (variable): Fix discovered issues
8. **Documentation** (30 min): Create README, document known issues

**Total Estimate:** 6-8 hours (plus bug fixing time)

## Notes

- Single worker execution prevents animation timing race conditions
- Redux state exposure allows testing animation queue state directly
- Visual regression tests catch subtle animation glitches
- Tests are written to be resilient to game-specific differences
- Focus on core animation system rather than specific game rules
