import { expect, test, type Frame, type Page } from '@playwright/test';
import { readFileSync } from 'node:fs';
import { dirname, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';
import { gunzipSync } from 'node:zlib';

// Blackjack in the standalone guest, driven by a game recorded on the Go
// engine: conformance/blackjack/scenario-bust-and-finish.json.gz. Three
// players, two rounds, to game over. The harness turns each viewer's view in
// that document into protocol entries; nothing here is hand-written state.

const here = dirname(fileURLToPath(import.meta.url));
const HARNESS = '/fabric/guest/harness.html?replay=blackjack';

interface Stack { Indexes: number[] }
interface View {
  viewer: number;
  game: {
    CurrentPlayerIndex: number;
    Finished: boolean;
    Winners: number[];
    CurrentState: { Players: { Hand: Stack }[] };
  };
  legality: { name: string; legalForPlayer: boolean }[];
}
const recording = JSON.parse(gunzipSync(readFileSync(
  resolve(here, '../../../conformance/blackjack/scenario-bust-and-finish.json.gz'),
)).toString('utf8')) as {
  game: { name: string; numPlayers: number; finalVersion: number };
  viewers: number[];
  versions: { version: number; settled: boolean; move: { name: string; trigger: string } | null; views: View[] }[];
};
const viewOf = (version: number, viewer: number): View =>
  recording.versions[version]!.views[recording.viewers.indexOf(viewer)]!;
/** How many cards of each seat's hand this viewer is sent a face for. */
const visibleInHands = (version: number, viewer: number): number[] =>
  viewOf(version, viewer).game.CurrentState.Players.map(player => player.Hand.Indexes.filter(index => index >= 0).length);
const handSizes = (version: number): number[] =>
  viewOf(version, -1).game.CurrentState.Players.map(player => player.Hand.Indexes.length);

interface GuestMessage {
  type: string;
  version?: number;
  how?: string;
  viewer?: number;
  stage?: string;
  games?: string[];
  name?: string;
  arguments?: Record<string, string>;
  snapshotVersion?: number;
  message?: string;
}

declare const harness: {
  ready: Promise<void>;
  received: GuestMessage[];
  sendWindow(from: number, to: number): void;
  playAll(): Promise<void>;
  setViewer(seat: number): void;
  waitFor(predicate: (m: GuestMessage) => boolean, timeoutMs?: number): Promise<GuestMessage>;
};

async function open(page: Page, viewer = 0): Promise<Frame> {
  const errors: string[] = [];
  page.on('pageerror', error => errors.push(String(error)));
  page.on('console', message => { if (message.type() === 'error') errors.push(message.text()); });
  await page.goto(`${HARNESS}&viewer=${viewer}`);
  await page.evaluate(() => harness.ready);
  const guest = page.frames().find(frame => frame !== page.mainFrame());
  if (!guest) throw new Error('the guest frame did not load');
  (page as Page & { guestErrors: string[] }).guestErrors = errors;
  return guest;
}
const errorsOf = (page: Page) => (page as Page & { guestErrors: string[] }).guestErrors;
const received = (page: Page) => page.evaluate(() => harness.received);
/** Deliver every version through `version` and wait until it has been shown. */
async function playThrough(page: Page, version: number): Promise<void> {
  await page.evaluate(v => {
    harness.sendWindow(0, v);
    return harness.waitFor(m => m.type === 'animations-done' && m.version === v, 90_000);
  }, version);
}
/** Per seat: how many cards its hand shows, and how many of them show a face. */
async function handsShown(page: Page): Promise<{ cards: number; faces: number }[]> {
  const panels = page.frameLocator('#guest').locator('boardgame-player-panel');
  const out: { cards: number; faces: number }[] = [];
  for (let seat = 0; seat < await panels.count(); seat++) {
    // A card draws its face in its own shadow tree, so ask the element.
    const faceUp = await panels.nth(seat).locator('boardgame-component-zone[label="Hand"] boardgame-card')
      .evaluateAll(cards => cards.map(card => (card as HTMLElement & { faceUp: boolean }).faceUp === true));
    out.push({ cards: faceUp.length, faces: faceUp.filter(Boolean).length });
  }
  return out;
}

test('the recording is the game these tests assume', () => {
  expect(recording.game).toMatchObject({ name: 'blackjack', numPlayers: 3, finalVersion: 46 });
  expect(recording.versions).toHaveLength(47);
  // Version 10: the deal is done and nobody's hidden card has been revealed.
  expect(visibleInHands(10, 0)).toEqual([2, 1, 1]);
  expect(visibleInHands(10, 1)).toEqual([1, 2, 1]);
  expect(visibleInHands(10, -1)).toEqual([1, 1, 1]);
  expect(viewOf(11, 0).game.CurrentPlayerIndex).toBe(0);
  expect(viewOf(46, 0).game).toMatchObject({ Finished: true, Winners: [2] });
});

test('the renderer is chosen by the game name of init, from a guest that bundles two games', async ({ page }) => {
  const guest = await open(page);
  const messages = await received(page);
  expect(messages[0]).toMatchObject({ type: 'ready', stage: 'loaded', games: ['blackjack', 'memory'] });
  expect(messages.slice(0, 3).map(m => [m.type, m.stage ?? m.how])).toEqual([
    ['ready', 'loaded'], ['ready', 'mounted'], ['animations-done', 'initial'],
  ]);
  const tree = await guest.evaluate(() => {
    const render = document.querySelector('boardgame-render-game') as HTMLElement & { renderer: HTMLElement | null; gameName: string };
    return {
      gameName: render.gameName,
      renderer: render.renderer?.localName,
      memoryMounted: !!document.querySelector('boardgame-render-game-memory'),
    };
  });
  expect(tree).toEqual({ gameName: 'blackjack', renderer: 'boardgame-render-game-blackjack', memoryMounted: false });
  // One player-info renderer per seat, mounted through the host, for three seats.
  await expect(page.frameLocator('#guest').locator('boardgame-render-player-info-blackjack')).toHaveCount(3);
  await expect(page.frameLocator('#guest').locator('.guest-seat-name')).toHaveText(['Ada', 'Grace', 'Lin']);
  expect(errorsOf(page)).toEqual([]);
});

for (const viewer of [0, 1, -1]) {
  test(`after the deal, viewer ${viewer} sees faces only for the cards that viewer was sent`, async ({ page }) => {
    await open(page, viewer);
    await playThrough(page, 10);
    const expectedFaces = visibleInHands(10, viewer);
    await expect.poll(() => handsShown(page)).toEqual(
      handSizes(10).map((cards, seat) => ({ cards, faces: expectedFaces[seat] })),
    );
    expect(errorsOf(page)).toEqual([]);
  });
}

test('the whole recorded game plays through, every version in order, with no network', async ({ page }) => {
  test.setTimeout(240_000);
  const requests: { url: string; guest: boolean }[] = [];
  const sockets: string[] = [];
  page.on('request', request => {
    requests.push({ url: request.url(), guest: request.frame() !== page.mainFrame() && !request.isNavigationRequest() });
  });
  page.on('websocket', socket => sockets.push(socket.url()));
  const guest = await open(page);
  await page.evaluate(() => harness.playAll());
  const messages = await received(page);
  const done = messages.filter(m => m.type === 'animations-done');
  expect(done.map(m => m.version)).toEqual(recording.versions.map(v => v.version));
  expect(done.filter(m => m.how !== 'played').map(m => [m.version, m.how])).toEqual([[0, 'initial']]);
  expect(messages.filter(m => m.type === 'error')).toEqual([]);
  // The real animator ran, and its gate never had to be forced shut.
  const counters = await guest.evaluate(() => {
    const hooks = (window as unknown as { __bgAnimTestHooks: { gateOpens: number; gateCloses: number; plays: number; watchdogFirings: number } }).__bgAnimTestHooks;
    return { opens: hooks.gateOpens, closes: hooks.gateCloses, plays: hooks.plays, watchdog: hooks.watchdogFirings };
  });
  expect(counters.watchdog).toBe(0);
  expect(counters.opens).toBeGreaterThanOrEqual(recording.game.finalVersion);
  expect(counters.closes).toBe(counters.opens);
  expect(counters.plays).toBeGreaterThan(0);
  // The game is over and the renderer says so.
  // Player 3 won; this viewer is player 1.
  expect(await page.frameLocator('#guest').locator('boardgame-game-outcome')
    .evaluate(outcome => outcome.shadowRoot?.textContent?.replace(/\s+/g, ' ').trim())).toBe('Game over! You lost.');
  expect(requests.filter(request => request.guest)).toEqual([]);
  expect(sockets).toEqual([]);
  expect(errorsOf(page)).toEqual([]);
});

test('the seat to move proposes Hit; another seat cannot', async ({ page }) => {
  await open(page, 0);
  await playThrough(page, 11);
  expect(viewOf(11, 0).legality.find(item => item.name === 'Current Player Hit')?.legalForPlayer).toBe(true);
  const frame = page.frameLocator('#guest');
  await frame.getByRole('button', { name: 'Hit', exact: true }).click();
  const proposal = await page.evaluate(() => harness.waitFor(m => m.type === 'propose-move', 10_000));
  expect(proposal).toMatchObject({
    type: 'propose-move',
    name: 'Current Player Hit',
    // The renderer names no target; the engine binds the current player.
    arguments: {},
    snapshotVersion: 11,
    viewer: 0,
  });

  // Player 2 is shown the same version, where it is not their turn.
  await page.evaluate(() => harness.setViewer(1));
  await page.evaluate(() => harness.waitFor(m => m.type === 'animations-done' && m.viewer === 1 && m.version === 11, 30_000));
  expect(viewOf(11, 1).legality.find(item => item.name === 'Current Player Hit')?.legalForPlayer).toBe(false);
  await expect(frame.getByRole('button', { name: 'Hit', exact: true })).toBeDisabled();
  expect((await received(page)).filter(m => m.type === 'propose-move')).toHaveLength(1);
});
