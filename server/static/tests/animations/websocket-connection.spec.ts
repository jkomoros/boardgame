import { test, expect } from '@playwright/test';

test.describe('WebSocket Connection Fix', () => {
  test('WebSocket should stay connected after page load', async ({ page }) => {
    // Track console messages about WebSocket
    const consoleMessages: string[] = [];
    page.on('console', msg => {
      const text = msg.text();
      consoleMessages.push(text);
    });

    // Navigate to the game
    await page.goto('http://localhost:8080/game/debuganimations/71F2D25CE72ECBA1/');

    // Wait for game to load
    await page.waitForSelector('boardgame-render-game', { timeout: 10000 });

    // Wait 5 seconds to see if socket closes
    await page.waitForTimeout(5000);

    // Check for "Socket closed" warnings
    const socketClosedMessages = consoleMessages.filter(msg =>
      msg.includes('Socket closed') || msg.includes('WebSocket')
    );

    console.log('All console messages:', consoleMessages.length);
    console.log('Socket-related messages:', socketClosedMessages);

    // The fix should prevent "Socket closed" warnings
    const hasSocketClosed = socketClosedMessages.some(msg => msg.includes('Socket closed'));

    if (hasSocketClosed) {
      console.log('❌ FAILED: WebSocket closed (fix not working)');
      throw new Error('WebSocket connection closed - ws:true fix may not be applied');
    } else {
      console.log('✅ PASSED: WebSocket stayed connected');
    }

    expect(hasSocketClosed).toBe(false);
  });

  test('Game should load with cards visible', async ({ page }) => {
    await page.goto('http://localhost:8080/game/debuganimations/71F2D25CE72ECBA1/');
    await page.waitForSelector('boardgame-render-game', { timeout: 10000 });
    await page.waitForTimeout(2000);

    // Take screenshot
    await page.screenshot({ path: 'test-results/game-loaded.png', fullPage: true });

    // Verify buttons are present (game loaded)
    const swapButton = await page.getByRole('button', { name: 'Swap' }).first();
    await expect(swapButton).toBeVisible();

    const drawButton = await page.getByRole('button', { name: 'Draw' }).first();
    await expect(drawButton).toBeVisible();

    const flipButton = await page.getByRole('button', { name: 'Flip' });
    await expect(flipButton).toBeVisible();

    console.log('✅ Game loaded successfully with all buttons visible');
  });
});
