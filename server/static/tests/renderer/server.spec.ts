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
  const gameList = page.locator('boardgame-list-games-view');
  await expect(gameList).toBeAttached();
  await expect.poll(() => gameList.evaluate(element => (
    element as unknown as { _managers: unknown[] }
  )._managers.length)).toBeGreaterThan(0);
  expect(await page.evaluate(() => (window as unknown as { API_HOST: string }).API_HOST)).toBe('');
});

test('assembled Pig renderer reports a real server rejection without advancing state', async ({ page }) => {
  const email = 'typed-action@example.com';
  const auth = await page.request.post('/api/auth', {
    form: {
      uid: email,
      token: 'offline-test-token',
      email,
      displayname: 'Typed Action',
    },
  });
  expect(auth.ok()).toBe(true);

  const created = await page.request.post('/api/new/game', {
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
  const renderer = page.locator('boardgame-render-game-pig');
  await expect(renderer).toBeAttached({ timeout: 15_000 });
  // Let Pig's initial fix-up chain settle before installing the presentation
  // perspective used by this transport-focused test.
  await expect.poll(() => renderer.evaluate(element => (
    element as unknown as { gameVersion: number }
  ).gameVersion)).toBeGreaterThanOrEqual(2);
  // This server-backed smoke game intentionally has no occupied seats. Make
  // only the presentation verdict available so the test exercises the action
  // and HTTP transport; server legality remains authoritative and may reject.
  const live = await renderer.evaluate(element => {
    const typed = element as unknown as { currentPlayerIndex: number; gameVersion: number };
    return { proposer: typed.currentPlayerIndex, version: typed.gameVersion };
  });

  const moveRequestPromise = page.waitForRequest(candidate =>
    candidate.method() === 'POST' && candidate.url().endsWith(`/api/game/pig/${result.GameID}/move`));
  // Install and consume the test perspective atomically so a websocket render
  // cannot replace it between setup and action creation.
  const proposalPromise = renderer.evaluate(element => {
    const typed = element as unknown as {
      currentPlayerIndex: number;
      gameVersion: number;
      viewingAsPlayer: number;
      proposingAsPlayer: number;
      proposingAsAdmin: boolean;
      moveLegality: Record<string, { legalForPlayer: boolean; legalForAnyone: boolean }>;
      move(name: 'Roll Dice'): { propose(): Promise<unknown> };
    };
    typed.viewingAsPlayer = typed.currentPlayerIndex;
    typed.proposingAsPlayer = typed.currentPlayerIndex;
    typed.proposingAsAdmin = true;
    typed.moveLegality = {
      'Roll Dice': { legalForPlayer: true, legalForAnyone: true },
      'Done Turn': { legalForPlayer: false, legalForAnyone: true },
    };
    return typed.move('Roll Dice').propose();
  });
  const moveRequest = await moveRequestPromise;
  const body = new URLSearchParams(moveRequest.postData() ?? '');
  expect(body.get('MoveType')).toBe('Roll Dice');
  expect(body.get('player')).toBe(String(live.proposer));
  expect(body.get('admin')).toBe('1');
  expect(body.get('ExpectedVersion')).toMatch(/^\d+$/);
  const expectedVersion = Number(body.get('ExpectedVersion'));

  expect((await moveRequest.response())?.ok()).toBe(true);
  await expect(proposalPromise).resolves.toMatchObject({
    kind: 'server-rejection',
    requestID: expect.any(String),
    error: expect.any(String),
  });
  expect(expectedVersion).toBe(live.version);
  await expect.poll(() => renderer.evaluate(element => (
    element as unknown as { gameVersion: number }
  ).gameVersion)).toBe(live.version);
  expect(failedRendererRequests).toEqual([]);
});

test('legacy declarative renderer controls are inert', async ({ page }) => {
  await page.goto('/');
  const result = await page.evaluate(async () => {
    await import('/src/components/boardgame-base-game-renderer.ts');
    const renderer = document.createElement('boardgame-base-game-renderer') as HTMLElement & {
      updateComplete: Promise<unknown>;
    };
    let proposals = 0;
    renderer.addEventListener('propose-move', () => proposals++);
    document.body.append(renderer);
    await renderer.updateComplete;

    const button = document.createElement('button');
    button.setAttribute('propose-move', 'Choose');
    renderer.append(button);
    button.click();
    await Promise.resolve();
    const hasProposalShortcut = 'proposeMove' in renderer;
    renderer.remove();
    return { proposals, hasProposalShortcut };
  });

  expect(result).toEqual({ proposals: 0, hasProposalShortcut: false });
});
