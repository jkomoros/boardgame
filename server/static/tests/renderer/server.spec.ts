import { expect, test } from '@playwright/test';

test('self-started renderer server uses offline mode and same-origin API proxy', async ({ page, request }) => {
  const configResponse = await request.get('/client_config.js');
  expect(configResponse.ok()).toBe(true);
  const config = await configResponse.text();
  expect(config).toContain('"offline_dev_mode": true');
  expect(config).toContain('"host": ""');
  expect(config).toContain('"dev_host": ""');

  const apiResponse = await request.get('/api/list/manager');
  expect(apiResponse.ok()).toBe(true);
  expect(apiResponse.headers()['content-type']).toContain('application/json');

  await page.goto('/');
  await expect(page.locator('boardgame-app')).toBeAttached();
  expect(await page.evaluate(() => (window as unknown as { API_HOST: string }).API_HOST)).toBe('');
});

test('assembled Pig renderer resolves the experimental client facade', async ({ page, request }) => {
  const auth = await request.post('/api/auth', {
    form: {
      uid: 'facade-test@example.com',
      token: 'offline-test-token',
      email: 'facade-test@example.com',
      displayname: 'Facade Test',
    },
  });
  expect(auth.ok()).toBe(true);

  const created = await request.post('/api/new/game', {
    form: { manager: 'pig', numplayers: '2' },
  });
  expect(created.ok()).toBe(true);
  const result = await created.json() as { GameID: string; Status: string };
  expect(result.Status).toBe('Success');

  const failedRendererRequests: string[] = [];
  page.on('requestfailed', (failed) => {
    if (failed.url().includes('/game-src/pig/') || failed.url().endsWith('/src/client.js')) {
      failedRendererRequests.push(`${failed.url()}: ${failed.failure()?.errorText || 'unknown failure'}`);
    }
  });
  await page.goto(`/game/pig/${result.GameID}`);
  await expect(page.locator('boardgame-render-game-pig')).toBeAttached({ timeout: 15_000 });
  await expect(page.locator('boardgame-die')).toBeAttached();
  expect(failedRendererRequests).toEqual([]);
});
