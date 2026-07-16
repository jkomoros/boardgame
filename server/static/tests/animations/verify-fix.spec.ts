import { test, expect } from '@playwright/test';
import { createOfflineGame, expectCleanGate, gateSnapshot } from './helpers';

test.describe('Debug Animations - WebSocket Fix Verification', () => {
  test('should load game with cards visible and animate on button click', async ({ page }) => {
    await createOfflineGame(page, 'debuganimations');
    await expect(page.locator('boardgame-card').first()).toBeVisible();
    const before = await gateSnapshot(page);

    // Enable slow animation for easier observation
    await page.getByText('Slow Animation').click();

    // Click the "Draw" button (should move card from DrawStack to DiscardStack)
    await page.getByRole('button', { name: 'Draw' }).first().click();

    await expectCleanGate(page, before, 15000);
  });

  test('should receive state updates via WebSocket after move', async ({ page }) => {
    // Listen for network WebSocket connections
    let gameSocketConnected = false;
    let gameSocketClosed = false;
    const versionFrames: number[] = [];
    page.on('websocket', ws => {
      if (!ws.url().includes('/api/game/debuganimations/')) return;
      gameSocketConnected = true;

      ws.on('framereceived', event => {
        try {
          const message = JSON.parse(String(event.payload));
          if (message.type === 'version' && typeof message.data === 'number') {
            versionFrames.push(message.data);
          }
        } catch {
          // Non-JSON frames are irrelevant to the version assertion.
        }
      });

      ws.on('close', () => {
        gameSocketClosed = true;
      });
    });

    await createOfflineGame(page, 'debuganimations');
    await expect.poll(() => gameSocketConnected).toBe(true);
    await expect.poll(() => versionFrames.length).toBeGreaterThan(0);
    const versionBeforeMove = Math.max(...versionFrames);

    // Click a button to trigger a move
    await page.getByRole('button', { name: 'To Hidden' }).click();

    // The socket sends an initial version during registration. Requiring a
    // strictly newer frame proves this click produced a post-handshake state
    // update rather than accidentally accepting that initial frame.
    await expect.poll(() => Math.max(...versionFrames), { timeout: 10000 }).toBeGreaterThan(versionBeforeMove);
    expect(gameSocketClosed).toBe(false);
  });
});
