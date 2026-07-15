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

test('legacy declarative controls cannot bypass generated schema freshness', async ({ page }) => {
  await page.goto('/');
  const result = await page.evaluate(async () => {
    await import('/src/components/boardgame-base-game-renderer.ts');
    const renderer = document.createElement('boardgame-base-game-renderer') as HTMLElement & {
      updateComplete: Promise<unknown>;
      moveInputSchema: unknown;
      moveInputSchemaFingerprint: string;
      serverMoveInputSchemaFingerprint: string;
    };
    renderer.moveInputSchema = [{ name: 'Choose', fields: [] }];
    renderer.moveInputSchemaFingerprint = 'client-fingerprint';
    renderer.serverMoveInputSchemaFingerprint = 'server-fingerprint';

    let proposals = 0;
    let errorMessage = '';
    renderer.addEventListener('propose-move', () => proposals++);
    const captureError = (event: ErrorEvent) => {
      errorMessage = event.error instanceof Error ? event.error.message : event.message;
      event.preventDefault();
    };
    window.addEventListener('error', captureError);
    document.body.append(renderer);
    await renderer.updateComplete;

    const button = document.createElement('button');
    button.setAttribute('propose-move', 'Choose');
    renderer.append(button);
    button.click();
    await Promise.resolve();

    window.removeEventListener('error', captureError);
    renderer.remove();
    return { proposals, errorMessage };
  });

  expect(result.proposals).toBe(0);
  expect(result.errorMessage).toContain('Generated move inputs are stale');
});
