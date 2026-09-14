import { expect, test } from '@playwright/test';

test('real Memory replay renders reveal, mismatch, and timer cleanup for each viewer', async ({ page }) => {
  const failures: string[] = [];
  page.on('pageerror', error => failures.push(error.message));
  await page.goto('/scenario-review.html');
  await expect(page.locator('body')).toHaveAttribute('data-frame', '0');
  for (const viewer of ['-1', '0', '1']) {
    await page.locator('#viewer').selectOption(viewer);
    for (const [frame, visible] of [0, 1, 2, 0].entries()) {
      await expect(page.locator('body')).toHaveAttribute('data-frame', String(frame));
      const cards = page.locator('boardgame-render-game-memory boardgame-component-stack').first().locator(':scope > boardgame-card[boardgame-component]');
      await expect(cards).toHaveCount(20);
      await expect.poll(() => cards.evaluateAll(elements => elements.filter(card => card.querySelector('div') !== null).length)).toBe(visible);
      await expect(page.locator('#error')).toBeEmpty();
      if (viewer === '0' && frame === 1) {
        await expect(cards.nth(0)).toHaveAttribute('aria-disabled', 'true');
        await expect(cards.nth(1)).toHaveAttribute('aria-disabled', 'false');
      }
      if (frame < 3) await page.getByRole('button', { name: 'Next', exact: true }).click();
    }
    await expect(page.getByText(viewer === "1" ? "Your turn" : "Player 2's turn", { exact: true })).toBeVisible();
    await expect(page.getByRole('button', { name: 'Next', exact: true })).toBeDisabled();
    for (let i = 0; i < 3; i++) await page.getByRole('button', { name: 'Previous', exact: true }).click();
  }
  expect(failures).toEqual([]);
});


test('timed Werewolf replay exposes native votes and the partial-vote deadline', async ({ page }) => {
  const failures: string[] = [];
  page.on('pageerror', error => failures.push(error.message));
  await page.goto('/scenario-review.html?game=werewolf&surface=hand&viewer=0');
  await expect(page.locator('body')).toHaveAttribute('data-frame', '0');
  for (let i = 0; i < 4; i++) await page.getByRole('button', { name: 'Next', exact: true }).click();
  await expect(page.locator('body')).toHaveAttribute('data-frame', '4');
  await expect(page.getByText('Vote to eliminate', { exact: true })).toBeVisible();
  await expect(page.locator('boardgame-target-list button')).toHaveCount(4);
  await expect(page.locator('boardgame-target-list button').nth(0)).toBeDisabled();
  await expect(page.locator('boardgame-target-list button').nth(1)).toBeEnabled();
  await expect(page.getByText('45s', { exact: true })).toBeVisible();
  await page.getByRole('button', { name: 'Next', exact: true }).click();
  await expect(page.getByText(/Vote cast/)).toBeVisible();
  await page.getByRole('button', { name: 'Next', exact: true }).click();
  await expect(page.getByText(/Sleep tight/)).toBeVisible();
  expect(failures).toEqual([]);
});

test('recorded transports reject stale actors and inputs without exact legal evidence', async ({ page }) => {
  await page.goto('/scenario-review.html');
  await expect(page.locator('body')).toHaveAttribute('data-frame', '0');
  const result = await page.evaluate(async () => {
    const { scenarioFixtureSnapshot } = await import('/src/testing/scenario-fixture.ts');
    const { mountRendererFixture } = await import('/src/testing/renderer-fixture.ts');
    const replay = await (await fetch('/game-src/memory/scenario-replay.json')).json();
    const snapshot = scenarioFixtureSnapshot(replay, 1, 0, { gameName: 'memory' });
    const handle = await mountRendererFixture({ tagName: 'boardgame-render-game-memory', snapshot });
    const request = {
      requestID: 'review', snapshotVersion: snapshot.version, viewingAsPlayer: 0,
      proposingAsPlayer: 0, proposingAsAdmin: false, name: 'Reveal Card', arguments: { CardIndex: '1' },
    };
    const invalid = [
      { ...request, snapshotVersion: snapshot.version - 1 },
      { ...request, viewingAsPlayer: 1 },
      { ...request, proposingAsPlayer: 1 },
      { ...request, proposingAsAdmin: true },
      { ...request, arguments: { CardIndex: '0' } },
      { ...request, arguments: {} },
      { ...request, arguments: { CardIndex: '1', Unexpected: 'yes' } },
      { ...request, arguments: { CardIndex: 1 } },
    ];
    const rejected = await Promise.all(invalid.map(candidate => handle.renderer.moveTransport.submit(candidate)));
    const preview = await handle.renderer.movePreviewTransport.preview({ ...request, proposingAsPlayer: 1, candidateKey: 'wrong', signal: new AbortController().signal });
    const targets = await handle.renderer.targetPreviewTransport.previewTargets({ ...request, snapshotVersion: snapshot.version - 1, candidates: [{ id: 'one', arguments: request.arguments }], signal: new AbortController().signal });
    const accepted = await handle.renderer.moveTransport.submit(request);
    const proposals = handle.proposals.length;
    handle.dispose();
    const disposed = await handle.renderer.moveTransport.submit(request);
    return { rejected: rejected.map(value => value.kind), preview: preview.kind, targets: targets.kind, accepted: accepted.kind, disposed: disposed.kind, proposals };
  });
  expect(result).toEqual({ rejected: Array(8).fill('server-rejection'), preview: 'failure', targets: 'failure', accepted: 'success', disposed: 'server-rejection', proposals: 1 });
});
