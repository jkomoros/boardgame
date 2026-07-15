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
