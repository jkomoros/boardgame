import { Page } from '@playwright/test';

/**
 * Test fixtures and helper functions for Playwright tests
 */

/**
 * Expose the Redux store to tests via window.__TEST_STORE__
 * This allows tests to inspect application state
 */
export async function exposeStore(page: Page, timeout = 10000): Promise<void> {
  // Wait for the app element to be available (indicating the app is loaded)
  await page.waitForSelector('boardgame-app', { timeout });

  // Import and expose the store module
  await page.evaluate(async () => {
    try {
      // Dynamically import the store module
      const storeModule = await import('/src/store.ts');
      (window as any).__TEST_STORE__ = storeModule.store;
    } catch (error) {
      console.error('Failed to import store:', error);
    }
  });

  // Verify the store was successfully exposed
  await page.waitForFunction(
    () => {
      return (window as any).__TEST_STORE__ !== undefined;
    },
    { timeout: 5000 }
  );
}

/**
 * Get the current Redux store state
 * Must call exposeStore() first
 */
export async function getStoreState(page: Page): Promise<any> {
  return await page.evaluate(() => {
    const store = (window as any).__TEST_STORE__;
    return store ? store.getState() : null;
  });
}

/**
 * Wait for the animation queue to drain
 * This ensures all animations have completed before continuing
 */
export async function waitForAnimationQueue(page: Page, timeout = 5000): Promise<void> {
  await page.waitForFunction(
    () => {
      const store = (window as any).__TEST_STORE__;
      if (!store) return false;

      const state = store.getState();
      const bundleCount = state?.app?.pendingBundleCount || 0;

      return bundleCount === 0;
    },
    { timeout }
  );
}

/**
 * Get the current pending bundle count (animations in progress)
 */
export async function getPendingBundleCount(page: Page): Promise<number> {
  return await page.evaluate(() => {
    const store = (window as any).__TEST_STORE__;
    if (!store) return 0;

    const state = store.getState();
    return state?.app?.pendingBundleCount || 0;
  });
}

/**
 * Login helper - waits for auth state to be loaded
 * Note: This doesn't perform actual login, just waits for auth initialization
 */
export async function waitForAuth(page: Page, timeout = 10000): Promise<void> {
  await page.waitForFunction(
    () => {
      const store = (window as any).__TEST_STORE__;
      if (!store) return false;

      const state = store.getState();
      // Check if user state is initialized (could be null for anonymous)
      return state?.user !== undefined;
    },
    { timeout }
  );
}

/**
 * Navigate to a game by name
 */
export async function navigateToGame(page: Page, gameName: string): Promise<void> {
  await page.goto(`/game/${gameName}`);
  await page.waitForLoadState('networkidle');
}

/**
 * Wait for a custom element to be defined and upgraded
 */
export async function waitForCustomElement(page: Page, tagName: string, timeout = 5000): Promise<void> {
  await page.waitForFunction(
    (tag) => {
      const element = document.querySelector(tag);
      return element && customElements.get(tag);
    },
    tagName,
    { timeout }
  );
}

/**
 * Take a screenshot with a descriptive name
 */
export async function takeScreenshot(page: Page, name: string): Promise<void> {
  await page.screenshot({ path: `screenshots/${name}.png`, fullPage: true });
}
