import { test, expect } from '@playwright/test';
import { exposeStore, waitForAuth, waitForCustomElement } from '../fixtures';

/**
 * Basic smoke tests for the homepage
 */

test.describe('Homepage', () => {
  test('loads successfully', async ({ page }) => {
    // Navigate to homepage
    await page.goto('/');

    // Wait for the main app element to be visible
    await expect(page.locator('boardgame-app')).toBeVisible();

    // Check that the page title is set
    await expect(page).toHaveTitle(/Board Game/);
  });

  test('displays main navigation', async ({ page }) => {
    await page.goto('/');

    // Wait for custom elements to load
    await waitForCustomElement(page, 'boardgame-app');

    // Check for main navigation elements
    const app = page.locator('boardgame-app');
    await expect(app).toBeVisible();

    // The app should have loaded without errors
    const errors = await page.locator('.error').count();
    expect(errors).toBe(0);
  });

  test('initializes Redux store', async ({ page }) => {
    await page.goto('/');

    // Expose the store for testing
    await exposeStore(page);

    // Wait for auth to initialize
    await waitForAuth(page);

    // Verify store is accessible
    const hasStore = await page.evaluate(() => {
      return (window as any).__TEST_STORE__ !== undefined;
    });
    expect(hasStore).toBe(true);
  });

  test('renders without console errors', async ({ page }) => {
    const consoleErrors: string[] = [];

    // Listen for console errors
    page.on('console', (msg) => {
      if (msg.type() === 'error') {
        consoleErrors.push(msg.text());
      }
    });

    await page.goto('/');
    await page.waitForLoadState('networkidle');

    // Filter out known harmless errors (if any)
    const criticalErrors = consoleErrors.filter((error) => {
      // Add filters here if there are known non-critical errors
      return !error.includes('favicon'); // Ignore favicon errors
    });

    // Verify no critical console errors
    expect(criticalErrors).toHaveLength(0);
  });
});
