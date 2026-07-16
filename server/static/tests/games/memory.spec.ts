import { test, expect } from '@playwright/test';
import { exposeStore, waitForAnimationQueue, takeScreenshot, navigateToGameByName } from '../fixtures';

/**
 * Tests for the Memory game
 *
 * This game should display a grid of face-down cards that can be flipped
 * to find matching pairs.
 */

test.describe('Memory Game', () => {
  test('loads game page', async ({ page }) => {
    await navigateToGameByName(page, 'memory');

    // Wait for the game renderer to be present
    await page.waitForSelector('boardgame-render-game', { timeout: 10000 });

    // Take screenshot of initial state
    await takeScreenshot(page, 'memory-game-initial');

    // Verify the game renderer is visible
    const gameRenderer = page.locator('boardgame-render-game');
    await expect(gameRenderer).toBeVisible();
  });

  test('renders game board with cards', async ({ page }) => {
    await navigateToGameByName(page, 'memory');
    await page.waitForSelector('boardgame-render-game', { timeout: 10000 });

    // Expose store to inspect game state
    await exposeStore(page);

    // Wait for animations to complete
    await waitForAnimationQueue(page, 10000);

    // Playwright locators intentionally pierce open shadow roots; a plain
    // document.querySelector does not, because boardgame-game-view owns the
    // generic renderer inside its shadow tree.
    const renderer = page.locator('boardgame-render-game');
    await expect(renderer).toBeVisible();

    // Look for the actual game components
    // Memory game should have a deck or stack of cards
    const gameContent = {
      hasContent: await renderer.evaluate((element) => (element.shadowRoot?.innerHTML.length || 0) > 0),
      innerHTML: await renderer.evaluate((element) => element.shadowRoot?.innerHTML.substring(0, 500) || ''),
      stackCount: await page.locator('boardgame-component-stack').count(),
      deckCount: await page.locator('boardgame-deck').count(),
      cardCount: await page.locator('boardgame-card').count(),
    };

    console.log('Game content:', JSON.stringify(gameContent, null, 2));

    // Take screenshot showing the issue
    await takeScreenshot(page, 'memory-game-with-cards');

    // The game should have rendered some content
    expect(gameContent.hasContent).toBe(true);

    // Memory game should have cards visible (either as stacks or cards)
    const hasGamePieces = gameContent.stackCount > 0 ||
                         gameContent.deckCount > 0 ||
                         gameContent.cardCount > 0;

    if (!hasGamePieces) {
      console.error('No game pieces found!');
      console.error('Shadow DOM content:', gameContent.innerHTML);
    }

    expect(hasGamePieces).toBe(true);
  });

  test('game renderer module is loaded', async ({ page }) => {
    await navigateToGameByName(page, 'memory');
    await page.waitForSelector('boardgame-render-game', { timeout: 10000 });

    // Wait for the dynamic import to complete
    await page.waitForTimeout(2000);

    // Check if the memory game renderer module was successfully imported
    const moduleStatus = await page.evaluate(async () => {
      try {
        // Try to import the memory game renderer (use .ts in dev mode with Vite)
        await import('/game-src/memory/boardgame-render-game-memory.ts');

        // The module registers a custom element, check if it's defined
        const isRegistered = customElements.get('boardgame-render-game-memory') !== undefined;

        return {
          success: true,
          hasRenderer: isRegistered,
          customElementDefined: isRegistered,
        };
      } catch (error) {
        return {
          success: false,
          error: error instanceof Error ? error.message : String(error),
        };
      }
    });

    console.log('Module status:', JSON.stringify(moduleStatus, null, 2));

    expect(moduleStatus.success).toBe(true);
    if (moduleStatus.success) {
      expect(moduleStatus.hasRenderer).toBe(true);
    }
  });

  test('checks for console errors during game load', async ({ page }) => {
    const errors: string[] = [];
    const warnings: string[] = [];

    page.on('console', (msg) => {
      if (msg.type() === 'error') {
        errors.push(msg.text());
      } else if (msg.type() === 'warning') {
        warnings.push(msg.text());
      }
    });

    await navigateToGameByName(page, 'memory');
    await page.waitForSelector('boardgame-render-game', { timeout: 10000 });
    await page.waitForTimeout(3000);

    console.log('Errors:', errors);
    console.log('Warnings:', warnings);

    // Filter out known harmless warnings
    const criticalErrors = errors.filter(error => {
      return !error.includes('favicon') &&
             !error.includes('Lit is in dev mode');
    });

    // There should be no critical errors
    if (criticalErrors.length > 0) {
      console.error('Critical errors found:', criticalErrors);
    }

    expect(criticalErrors.length).toBe(0);
  });

  test('verifies game state is initialized', async ({ page }) => {
    await navigateToGameByName(page, 'memory');
    await page.waitForSelector('boardgame-render-game', { timeout: 10000 });

    await exposeStore(page);

    const gameState = await page.evaluate(() => {
      const store = (window as any).__TEST_STORE__;
      if (!store) return { error: 'No store' };

      const state = store.getState();
      const game = state.game;

      return {
        hasGame: !!game,
        gameName: game?.name,
        gameId: game?.id,  // State uses 'id' not 'gameId'
        hasState: !!game?.currentState,  // State uses 'currentState' not 'state'
        version: game?.versions?.current,
      };
    });

    console.log('Game state:', JSON.stringify(gameState, null, 2));

    expect(gameState.hasGame).toBe(true);
    expect(gameState.gameName).toBe('memory'); // Should now be populated
    expect(gameState.gameId).toBeTruthy(); // Should have a valid game ID
    expect(gameState.hasState).toBe(true);
  });
});
