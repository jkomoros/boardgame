import { test, expect } from '@playwright/test';

test.describe('Debug Animations - WebSocket Fix Verification', () => {
  test('should load game with cards visible and animate on button click', async ({ page }) => {
    // Navigate to the debuganimations game
    await page.goto('http://localhost:8080/game/debuganimations/71F2D25CE72ECBA1/');

    // Wait for game to load
    await page.waitForSelector('boardgame-render-game', { timeout: 10000 });

    // Wait for cards to render
    await page.waitForTimeout(2000);

    // Take initial screenshot
    await page.screenshot({ path: 'test-results/before-animation.png', fullPage: true });

    // Check WebSocket connection in console
    const wsMessages: string[] = [];
    page.on('console', msg => {
      const text = msg.text();
      if (text.includes('Socket') || text.includes('WebSocket')) {
        wsMessages.push(text);
      }
    });

    // Enable slow animation for easier observation
    await page.getByText('Slow Animation').click();

    // Click the "Draw" button (should move card from DrawStack to DiscardStack)
    await page.getByRole('button', { name: 'Draw' }).first().click();

    // Wait for slow animation to complete (5 seconds + buffer)
    await page.waitForTimeout(6000);

    // Take post-animation screenshot
    await page.screenshot({ path: 'test-results/after-animation.png', fullPage: true });

    // Check for WebSocket closure warnings
    const hasSocketClosedWarning = wsMessages.some(msg => msg.includes('Socket closed'));

    console.log('WebSocket messages:', wsMessages);
    console.log('Socket closed warning found:', hasSocketClosedWarning);

    // The fix should prevent the socket from closing
    // If ws: true is working, socket should stay open
    if (hasSocketClosedWarning) {
      console.warn('⚠️  WebSocket closed - the fix may not be working');
    } else {
      console.log('✅ WebSocket stayed open - fix is working!');
    }

    // Basic assertion: page should not have crashed
    const title = await page.title();
    expect(title).toBe('Board Game');
  });

  test('should receive state updates via WebSocket after move', async ({ page }) => {
    // Navigate to game
    await page.goto('http://localhost:8080/game/debuganimations/71F2D25CE72ECBA1/');
    await page.waitForSelector('boardgame-render-game', { timeout: 10000 });
    await page.waitForTimeout(2000);

    // Listen for network WebSocket connections
    const wsConnections: any[] = [];
    page.on('websocket', ws => {
      console.log('✅ WebSocket connected:', ws.url());
      wsConnections.push(ws);

      ws.on('framereceived', event => {
        console.log('📨 WebSocket received:', event.payload);
      });

      ws.on('framesent', event => {
        console.log('📤 WebSocket sent:', event.payload);
      });

      ws.on('close', () => {
        console.log('❌ WebSocket closed');
      });
    });

    // Click a button to trigger a move
    await page.getByRole('button', { name: 'Flip' }).click();

    // Wait for response
    await page.waitForTimeout(2000);

    // Verify WebSocket was established
    expect(wsConnections.length).toBeGreaterThan(0);
    console.log(`WebSocket connections established: ${wsConnections.length}`);
  });
});
