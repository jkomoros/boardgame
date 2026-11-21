import { expect, test, type Frame, type Page } from '@playwright/test';
import { readFileSync } from 'node:fs';
import { dirname, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';

const here = dirname(fileURLToPath(import.meta.url));
const HARNESS = '/fabric/guest/harness.html';

interface Entry {
  game: { CurrentState: { Game: { VisibleCards: { Indexes: number[] }; HiddenCards: { Indexes: number[] } } } };
}
const replay = JSON.parse(readFileSync(resolve(here, '../replays/memory-full.json'), 'utf8')) as {
  chest: { Decks: { cards: { Values: { Type: string } }[] } };
  versions: { version: number; settled: boolean; viewers: Record<string, Entry> }[];
};
const lastVersion = replay.versions.length - 1;

interface GuestMessage {
  type: string;
  version?: number;
  how?: string;
  viewer?: number;
  stage?: string;
  name?: string;
  arguments?: Record<string, string>;
  snapshotVersion?: number;
  message?: string;
  timerId?: string;
  jumpedFrom?: number;
  remaining?: number;
}

declare const harness: {
  ready: Promise<void>;
  received: GuestMessage[];
  deliver(count: number): boolean;
  sendWindow(from: number, to: number): void;
  playAll(): Promise<void>;
  setViewer(seat: number): void;
  waitFor(predicate: (m: GuestMessage) => boolean, timeoutMs?: number): Promise<GuestMessage>;
};

/** Everything the page asked the network for, and which frame asked. */
function watchNetwork(page: Page) {
  const requests: { url: string; guest: boolean }[] = [];
  const sockets: string[] = [];
  page.on('request', request => {
    const frame = request.frame();
    requests.push({ url: request.url(), guest: frame !== page.mainFrame() && !request.isNavigationRequest() });
  });
  page.on('websocket', socket => sockets.push(socket.url()));
  return { requests, sockets };
}

async function open(page: Page, query = ''): Promise<Frame> {
  const errors: string[] = [];
  page.on('pageerror', error => errors.push(String(error)));
  page.on('console', message => { if (message.type() === 'error') errors.push(message.text()); });
  await page.goto(HARNESS + query);
  await page.evaluate(() => harness.ready);
  const guest = page.frames().find(frame => frame !== page.mainFrame());
  if (!guest) throw new Error('the guest frame did not load');
  (page as Page & { guestErrors: string[] }).guestErrors = errors;
  return guest;
}

const errorsOf = (page: Page) => (page as Page & { guestErrors: string[] }).guestErrors;
const received = (page: Page) => page.evaluate(() => harness.received);
const done = async (page: Page) => (await received(page)).filter(m => m.type === 'animations-done');
const waitDone = (page: Page, version: number) =>
  page.evaluate(v => harness.waitFor(m => m.type === 'animations-done' && m.version === v, 60_000), version);
const gridCards = (page: Page) =>
  page.frameLocator('#guest').locator('boardgame-component-stack[layout="grid"] boardgame-card');

test('the guest runs in an opaque origin, mounts the real renderer host, and renders the cards', async ({ page }) => {
  const guest = await open(page);

  // The sandbox really is opaque: this is the environment storage throws in.
  expect(await guest.evaluate(() => window.origin)).toBe('null');
  expect(await guest.evaluate(() => {
    try { void window.sessionStorage; return 'readable'; } catch { return 'throws'; }
  })).toBe('throws');
  expect(await page.locator('#guest').getAttribute('sandbox')).toBe('allow-scripts');

  const messages = await received(page);
  expect(messages.slice(0, 3).map(m => [m.type, m.stage ?? m.how])).toEqual([
    ['ready', 'loaded'], ['ready', 'mounted'], ['animations-done', 'initial'],
  ]);

  // The real element tree: render-game hosting Memory's own renderer.
  const tree = await guest.evaluate(() => {
    const render = document.querySelector('boardgame-render-game') as HTMLElement & { renderer: HTMLElement | null };
    return {
      renderer: render.renderer?.localName,
      animator: !!render.shadowRoot?.querySelector('boardgame-component-animator'),
      effects: !!render.shadowRoot?.querySelector('boardgame-effect-layer'),
    };
  });
  expect(tree).toEqual({ renderer: 'boardgame-render-game-memory', animator: true, effects: true });

  await expect(gridCards(page)).toHaveCount(10);
  // Nothing is revealed at version 0, so no card shows a face.
  expect((await gridCards(page).allTextContents()).map(text => text.trim()).filter(Boolean)).toEqual([]);
  // The player-info renderer is mounted per seat, without the Redux roster.
  await expect(page.frameLocator('#guest').locator('boardgame-render-player-info-memory')).toHaveCount(2);
  await expect(page.frameLocator('#guest').locator('.guest-seat-name')).toHaveText(['Ada', 'Grace']);
  expect(errorsOf(page)).toEqual([]);
});

test('a revealed card shows its face, and the reveal really animates', async ({ page }) => {
  const guest = await open(page);
  const state = replay.versions[1]!.viewers['0']!.game.CurrentState.Game;
  const slot = state.VisibleCards.Indexes.findIndex(index => index >= 0);
  const face = replay.chest.Decks.cards[state.VisibleCards.Indexes[slot]!]!.Values.Type;
  expect(slot).toBeGreaterThanOrEqual(0);

  await page.evaluate(() => { harness.deliver(1); });
  // While the cycle is open the gate is up and WAAPI animations are running.
  await expect.poll(() => guest.evaluate(() =>
    (document.querySelector('boardgame-render-game') as HTMLElement & { isAnimating: boolean }).isAnimating)).toBe(true);
  const running = await guest.evaluate(() => {
    const renderer = (document.querySelector('boardgame-render-game') as HTMLElement & { renderer: HTMLElement }).renderer;
    let count = 0;
    for (const card of renderer.shadowRoot!.querySelectorAll('boardgame-card')) {
      count += card.getAnimations({ subtree: true }).filter(animation => animation.playState === 'running').length;
    }
    return count;
  });
  expect(running).toBeGreaterThan(0);

  const message = await waitDone(page, 1);
  expect(message.how).toBe('played');
  const texts = (await gridCards(page).allTextContents()).map(text => text.trim());
  expect(texts[slot]).toBe(face);
  expect(texts.filter(Boolean)).toEqual([face]);
  await expect(gridCards(page).nth(slot)).toBeVisible();

  const hooks = await guest.evaluate(() => {
    const h = (window as unknown as { __bgAnimTestHooks: { gateOpens: number; plays: number; watchdogFirings: number } }).__bgAnimTestHooks;
    return { gateOpens: h.gateOpens, plays: h.plays, watchdogFirings: h.watchdogFirings };
  });
  expect(hooks.gateOpens).toBe(1);
  expect(hooks.plays).toBeGreaterThan(0);
  expect(hooks.watchdogFirings).toBe(0);
  expect(errorsOf(page)).toEqual([]);
});

test('stepping the whole game reports animations-done for every version, in order, with no network', async ({ page }) => {
  const network = watchNetwork(page);
  const guest = await open(page);
  await page.evaluate(() => harness.playAll());

  const versions = (await done(page)).map(m => m.version);
  expect(versions).toEqual(replay.versions.map(v => v.version));
  expect((await done(page)).slice(1).every(m => m.how === 'played')).toBe(true);
  expect((await received(page)).filter(m => m.type === 'error')).toEqual([]);

  // The finished game: every card captured, only empty slots in the grid.
  await expect(page.frameLocator('#guest')
    .locator('boardgame-component-stack[layout="grid"] boardgame-card:not([spacer])')).toHaveCount(0);
  await expect(page.frameLocator('#guest').locator('.discard-pile boardgame-card')).toHaveCount(10);

  // No watchdog ever had to force a cycle closed.
  expect(await guest.evaluate(() =>
    (window as unknown as { __bgAnimTestHooks: { watchdogFirings: number } }).__bgAnimTestHooks.watchdogFirings)).toBe(0);

  // The page asked for exactly its own three files, and the guest for nothing.
  const paths = network.requests.map(request => new URL(request.url).pathname).sort();
  expect(paths).toEqual([
    '/fabric/guest/dist/guest.html',
    '/fabric/guest/harness.html',
    '/fabric/guest/replays/memory-full.json',
  ]);
  expect(network.requests.filter(request => request.guest)).toEqual([]);
  expect(network.requests.every(request => new URL(request.url).origin === new URL(page.url()).origin)).toBe(true);
  expect(network.sockets).toEqual([]);
  expect(errorsOf(page)).toEqual([]);
});

test('a window of several versions plays each one in order, and a repeated window is ignored', async ({ page }) => {
  await open(page);
  // One message carrying versions 0 through 6: 0 was already shown.
  await page.evaluate(() => harness.sendWindow(0, 6));
  await waitDone(page, 6);
  let events = await done(page);
  expect(events.map(m => m.version)).toEqual([0, 1, 2, 3, 4, 5, 6]);
  // `remaining` counts down as the queue drains.
  expect(events.slice(1).map(m => m.remaining)).toEqual([5, 4, 3, 2, 1, 0]);
  expect(events.some(m => m.jumpedFrom !== undefined)).toBe(false);

  // The same cell value delivered again, and an older overlapping one.
  await page.evaluate(() => { harness.sendWindow(0, 6); harness.sendWindow(2, 4); });
  // A window that overlaps and extends: only the new versions play.
  await page.evaluate(() => harness.sendWindow(4, 9));
  await waitDone(page, 9);
  events = await done(page);
  expect(events.map(m => m.version)).toEqual([0, 1, 2, 3, 4, 5, 6, 7, 8, 9]);
  expect(errorsOf(page)).toEqual([]);
});

test('clicking a hidden card emits the right propose-move, and only for the seat to move', async ({ page }) => {
  await open(page);
  await gridCards(page).nth(3).click();
  const proposal = await page.evaluate(() => harness.waitFor(m => m.type === 'propose-move'));
  expect(proposal).toMatchObject({
    type: 'propose-move',
    name: 'Reveal Card',
    arguments: { CardIndex: '3' },
    snapshotVersion: 0,
    viewer: 0,
  });
  await expect(page.locator('#proposals li')).toHaveText([/Reveal Card \{"CardIndex":"3"\} at v0 as seat 0/]);

  // The move is unanswered, so the board is unchanged and nothing else is sent.
  await gridCards(page).nth(5).click({ force: true });
  await page.waitForTimeout(300);
  expect((await received(page)).filter(m => m.type === 'propose-move')).toHaveLength(1);

  // The other seat is not to move: its clicks propose nothing.
  await page.evaluate(() => harness.setViewer(1));
  await page.evaluate(() => harness.waitFor(m => m.type === 'animations-done' && m.viewer === 1));
  await gridCards(page).nth(2).click({ force: true });
  await page.waitForTimeout(300);
  expect((await received(page)).filter(m => m.type === 'propose-move')).toHaveLength(1);
});

test('a refused move frees the board for another, and an accepted one advances', async ({ page }) => {
  await open(page);
  await page.locator('#answer').selectOption('reject');
  await gridCards(page).nth(1).click();
  await page.evaluate(() => harness.waitFor(m => m.type === 'propose-move'));
  await expect.poll(async () => {
    await gridCards(page).nth(2).click({ force: true });
    return (await received(page)).filter(m => m.type === 'propose-move').length;
  }).toBeGreaterThanOrEqual(2);

  await page.locator('#answer').selectOption('advance');
  const before = (await received(page)).filter(m => m.type === 'propose-move').length;
  await expect.poll(async () => {
    if ((await received(page)).filter(m => m.type === 'propose-move').length === before) {
      await gridCards(page).nth(4).click({ force: true });
    }
    return (await done(page)).some(m => m.version === 1);
  }).toBe(true);
});

test('switching seat cuts to that seat\'s view without animating', async ({ page }) => {
  await open(page);
  await page.evaluate(() => harness.sendWindow(0, 1));
  await waitDone(page, 1);
  const status = page.frameLocator('#guest').locator('boardgame-turn-status');
  await expect(status).toContainText('Your turn');

  await page.evaluate(() => harness.setViewer(1));
  const cut = await page.evaluate(() => harness.waitFor(m => m.type === 'animations-done' && m.viewer === 1));
  expect(cut).toMatchObject({ version: 1, how: 'initial', viewer: 1 });
  await expect(status).not.toContainText('Your turn');
  await expect(gridCards(page)).toHaveCount(10);

  // The observer sees the same public board.
  await page.evaluate(() => harness.setViewer(-1));
  const observer = await page.evaluate(() => harness.waitFor(m => m.type === 'animations-done' && m.viewer === -1));
  expect(observer).toMatchObject({ version: 1, how: 'initial' });
  await expect(gridCards(page)).toHaveCount(10);
  expect(errorsOf(page)).toEqual([]);
});

test('a running timer asks the host for a tick when it reaches zero', async ({ page }) => {
  await open(page);
  // Version 3 starts the four-second hide timer.
  await page.evaluate(() => harness.sendWindow(0, 3));
  await waitDone(page, 3);
  await expect(page.frameLocator('#guest').locator('boardgame-timer')).toBeVisible();
  const tick = await page.evaluate(() => harness.waitFor(m => m.type === 'tick', 15_000));
  expect(tick.version).toBe(3);
  expect(typeof tick.timerId).toBe('string');
  expect((await received(page)).filter(m => m.type === 'tick')).toHaveLength(1);
});

test('the checked-in scenario replay plays, reporting the versions it never received', async ({ page }) => {
  const network = watchNetwork(page);
  await open(page, '?replay=scenario');
  await page.evaluate(() => harness.playAll());
  const events = await done(page);
  // scenario-replay.json holds decision points only: versions 0, 1, 3, 5.
  expect(events.map(m => [m.version, m.jumpedFrom])).toEqual([
    [0, undefined], [1, undefined], [3, 1], [5, 3],
  ]);
  await expect(gridCards(page)).toHaveCount(20);
  expect(network.requests.map(request => new URL(request.url).pathname).sort()).toEqual([
    '/examples/memory/client/scenario-replay.json',
    '/fabric/guest/dist/guest.html',
    '/fabric/guest/harness.html',
  ]);
  expect(errorsOf(page)).toEqual([]);
});

test('the guest works when embedded by srcdoc', async ({ page }) => {
  const guest = await open(page, '?srcdoc=1');
  expect(await guest.evaluate(() => window.origin)).toBe('null');
  await page.evaluate(() => harness.sendWindow(0, 2));
  await waitDone(page, 2);
  await expect(gridCards(page)).toHaveCount(10);
  expect(errorsOf(page)).toEqual([]);
});

test('the guest plays under a policy that forbids every network request', async ({ page }) => {
  await open(page, '?csp=network');
  await page.evaluate(() => harness.sendWindow(0, 5));
  await waitDone(page, 5);
  expect((await received(page)).filter(m => m.type === 'error')).toEqual([]);
  expect(errorsOf(page)).toEqual([]);
});

test('under a hash-only script policy nothing needs eval; inline styles are what it costs', async ({ page }) => {
  page.on('console', () => {});
  await page.goto(HARNESS + '?csp=strict');
  await page.evaluate(() => harness.ready);
  await page.evaluate(() => harness.sendWindow(0, 3));
  await waitDone(page, 3);
  const violations = (await received(page))
    .filter(m => m.type === 'error' && m.message?.startsWith('csp:'))
    .map(m => m.message!);
  test.info().annotations.push({ type: 'csp-violations', description: JSON.stringify([...new Set(violations)]) });
  // The script itself runs by hash, and nothing asks for eval or a fetch.
  expect(violations.filter(v => /script-src|connect-src|img-src|font-src/.test(v))).toEqual([]);
  // What a no-inline-style policy blocks is recorded in FINDINGS.md.
  expect(violations.every(v => v.startsWith('csp: style-src'))).toBe(true);
});

test('a guest that joins late cuts to the newest version instead of replaying the window', async ({ page }) => {
  await open(page, '?autostart=0');
  await page.evaluate(async () => {
    const mounted = harness.waitFor(m => m.type === 'ready' && m.stage === 'mounted');
    (harness as unknown as { sendInit(): void }).sendInit();
    await mounted;
    harness.sendWindow(8, 12);
  });
  const first = await page.evaluate(() => harness.waitFor(m => m.type === 'animations-done'));
  expect(first).toMatchObject({ version: 12, how: 'initial', remaining: 0 });
  // From there on, versions play one at a time as usual.
  await page.evaluate(() => harness.sendWindow(10, 14));
  await waitDone(page, 14);
  expect((await done(page)).map(m => [m.version, m.how])).toEqual([
    [12, 'initial'], [13, 'played'], [14, 'played'],
  ]);
  expect(errorsOf(page)).toEqual([]);
});
