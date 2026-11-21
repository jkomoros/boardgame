// Plays a whole game of Memory in a browser as two different people, against
// a piece already deployed to a local Common Fabric toolshed.
//
//   node tools/play-two-users.mjs <piece url> [screenshot dir]
//
// Each player is a separate browser context with its own storage, and so its
// own identity: the script registers a fresh passphrase identity in each, as
// a person would on the login screen, and grants it WRITE on the table's
// space with `cf acl set`, run as the key that created the space.
//
// Environment:
//   CF_LABS_ROOT   the labs checkout whose `bin/cf` to run
//   CF_IDENTITY    the key that created the space
//   CF_SPACE       the space's DID
//   CF_API_URL     the toolshed
//   HEADED=1       show the browsers
//   PLAYWRIGHT     where to import Playwright from (default: ../../../guest)
//
// It prints one JSON line per move with how long the move took to show up
// for the player who made it and for the other player, and a summary.
import { execFileSync } from 'node:child_process';
import { mkdirSync } from 'node:fs';
import { dirname, resolve } from 'node:path';
import { fileURLToPath, pathToFileURL } from 'node:url';

const here = dirname(fileURLToPath(import.meta.url));
const playwrightPath = process.env.PLAYWRIGHT ??
  resolve(here, '../../../guest/node_modules/@playwright/test/index.mjs');
const { chromium } = await import(pathToFileURL(playwrightPath).href);

const url = process.argv[2];
const shots = resolve(process.argv[3] ?? resolve(here, '../screenshots'));
if (!url) {
  console.error('usage: node play-two-users.mjs <piece url> [screenshot dir]');
  process.exit(2);
}
for (const name of ['CF_LABS_ROOT', 'CF_IDENTITY', 'CF_SPACE', 'CF_API_URL']) {
  if (!process.env[name]) {
    console.error(`${name} is not set`);
    process.exit(2);
  }
}
mkdirSync(shots, { recursive: true });

const GRID = 'boardgame-component-stack[layout="grid"] boardgame-card';
const sleep = ms => new Promise(done => setTimeout(done, ms));

function grant(did) {
  execFileSync(resolve(process.env.CF_LABS_ROOT, 'bin/cf'), ['acl', 'set', did, 'WRITE'], {
    env: process.env,
    stdio: ['ignore', 'ignore', 'inherit'],
  });
}

/** Registers a new identity on the shell's login screen; returns its DID. */
async function register(page) {
  await page.goto(url);
  await page.getByText('Register', { exact: false }).first().click();
  await page.getByText('Use Passphrase', { exact: false }).click();
  await page.getByText('Generate Passphrase', { exact: false }).click();
  await page.getByText("I've Saved It", { exact: false }).click();
  await page.waitForFunction(() => {
    try {
      return typeof globalThis.app?.state?.().identity?.did?.() === 'string';
    } catch {
      return false;
    }
  }, null, { timeout: 30_000 });
  return page.evaluate(() => globalThis.app.state().identity.did());
}

/** The guest's frame: the one holding the bridge's status bar. */
async function guestFrame(page) {
  for (let attempt = 0; attempt < 300; attempt++) {
    for (const frame of page.frames()) {
      if (frame === page.mainFrame()) continue;
      const found = await frame.evaluate(() => !!document.querySelector('#fabric-status')).catch(() => false);
      if (found) return frame;
    }
    await sleep(200);
  }
  throw new Error('the guest never loaded');
}

async function player(browser, name) {
  const context = await browser.newContext({ viewport: { width: 900, height: 760 } });
  const page = await context.newPage();
  const problems = [];
  const denials = [];
  page.on('console', async message => {
    if (message.type() === 'error' || message.text().includes('boardgame guest')) problems.push(message.text().slice(0, 400));
    // CFC_DEBUG=1: keep what each gate says about a refusal (see below).
    if (process.env.CFC_DEBUG && /ceiling|denial|write-policy/.test(message.text())) {
      const parts = await Promise.all(message.args().map(arg => arg.jsonValue().catch(() => '?')));
      denials.push(JSON.stringify(parts).slice(0, 3000));
    }
  });
  page.on('pageerror', error => problems.push(String(error).slice(0, 400)));
  const did = await register(page);
  grant(did);
  await page.reload();
  const frame = await guestFrame(page);
  if (process.env.CFC_DEBUG) {
    // A gate's reasons go to the runtime worker's console at debug level.
    await page.evaluate(async () => {
      globalThis.commonfabric.forwardWorkerConsole(true);
      await globalThis.commonfabric.rt.setLoggerEnabled(true, 'cfc');
      await globalThis.commonfabric.rt.setLoggerLevel('debug', 'cfc');
    });
  }
  return { name, page, frame, did, problems, denials, seen: [] };
}

const statusOf = who => who.frame.evaluate(() => ({
  who: document.querySelector('#fabric-status .who')?.textContent ?? '',
  note: document.querySelector('#fabric-status .note')?.textContent ?? '',
  error: document.querySelector('#fabric-status .error')?.textContent ?? '',
  version: document.querySelector('#guest-root')?.dataset.version ?? '',
}));

/** Each grid slot: its face if one shows, and whether a card is there. */
const boardOf = who => who.frame.evaluate(selector => {
  const renderer = document.querySelector('boardgame-render-game')?.renderer;
  const root = renderer?.shadowRoot ?? document;
  const cards = [...root.querySelectorAll(selector)];
  const deep = [];
  const walk = node => {
    for (const child of node.querySelectorAll('*')) {
      if (child.matches?.(selector)) deep.push(child);
      if (child.shadowRoot) walk(child.shadowRoot);
    }
  };
  if (cards.length === 0) walk(document);
  return (cards.length ? cards : deep).map(card => ({
    face: (card.innerText || '').trim(),
    present: !card.hasAttribute('spacer'),
  }));
}, GRID);

/**
 * Waits until a guest shows the newest version its bridge has received and
 * has had time to finish animating into it.
 */
async function settled(who) {
  await waitFor(() => who.frame.evaluate(() => {
    const views = globalThis.__boardgameFabricLog.filter(entry => entry.what === 'view');
    const newest = views[views.length - 1]?.version;
    return String(newest) === document.querySelector('#guest-root')?.dataset.version;
  }), `${who.name}'s guest to catch up`);
  await sleep(700);
}

async function waitFor(check, what, timeout = 30_000) {
  const started = Date.now();
  for (;;) {
    const value = await check();
    if (value) return value;
    if (Date.now() - started > timeout) throw new Error(`timed out waiting for ${what}`);
    await sleep(40);
  }
}

const shot = (who, label) => who.page.screenshot({ path: resolve(shots, `${label}-${who.name.toLowerCase()}.png`) });

// ---------------------------------------------------------------------------

const browser = await chromium.launch({ headless: !process.env.HEADED });
const alice = await player(browser, 'Alice');
const bob = await player(browser, 'Bob');
const players = [alice, bob];
console.log(JSON.stringify({ alice: alice.did, bob: bob.did }));

const latencies = [];
let moves = 0;
let shotReveal = false;
let shotTimer = false;
let shotCapture = false;

try {
// Seating.
for (const who of players) {
  await waitFor(async () => (await statusOf(who)).who !== 'Loading…', `${who.name}'s guest to hydrate`);
  await waitFor(() => who.frame.evaluate(() => !document.querySelector('#fabric-lobby button')?.disabled),
    `${who.name} to be identified`);
  await who.frame.locator('#fabric-lobby input').fill(who.name);
  await who.frame.getByRole('button', { name: 'Sit down' }).click();
  await waitFor(async () => (await statusOf(who)).who === `You are ${who.name}`, `${who.name} to be seated`);
}
await waitFor(() => alice.frame.evaluate(() => document.querySelectorAll('#fabric-lobby li').length === 2),
  'Alice to see both seats');
await shot(alice, '01-lobby');
await shot(bob, '01-lobby');

// The deal. Record every view the dealer's bridge sees, to tell whether a
// locally speculated deal was shown before the server's.
await alice.frame.getByRole('button', { name: 'Deal' }).click();
for (const who of players) {
  await waitFor(async () => (await boardOf(who)).length > 0, `${who.name} to see the board`);
}
await sleep(1500);
await shot(alice, '02-dealt');
await shot(bob, '02-dealt');
const dealViews = await alice.frame.evaluate(() =>
  globalThis.__boardgameFabricLog.filter(entry => entry.what === 'view' && entry.version === 0));
console.log(JSON.stringify({ dealerSawDeals: dealViews.map(entry => ({ epoch: entry.epoch, at: Math.round(entry.at) })) }));

// Play. Each player remembers every face either of them has turned.

/** The one player whose guest says it is their turn, once both agree. */
async function turnOf() {
  const notes = await Promise.all(players.map(async who => (await statusOf(who)).note));
  const mine = notes.map((note, index) => note === 'Your turn.' ? index : -1).filter(index => index >= 0);
  if (mine.length !== 1) return undefined;
  const waiting = players[1 - mine[0]];
  return notes[1 - mine[0]] === `${players[mine[0]].name}'s turn.` && waiting ? players[mine[0]] : undefined;
}

function choose(board, seen, turned) {
  const hidden = board.map((card, slot) => ({ card, slot }))
    .filter(({ card }) => card.present && card.face === '').map(({ slot }) => slot);
  if (turned >= 0) {
    const partner = hidden.find(slot => seen[slot] === board[turned].face);
    if (partner !== undefined) return partner;
  } else {
    const known = hidden.find(slot => seen[slot] && hidden.some(other => other !== slot && seen[other] === seen[slot]));
    if (known !== undefined) return known;
  }
  return hidden.find(slot => !seen[slot]) ?? hidden[0];
}

const seen = [];
for (let guard = 0; guard < 200; guard++) {
  const finished = (await statusOf(alice)).note.includes('won') || (await statusOf(alice)).note.includes('tie');
  if (finished) break;
  const mover = await waitFor(turnOf, 'someone to have the turn', 20_000);
  const other = mover === alice ? bob : alice;
  await settled(mover);
  await settled(other);
  // Wait for the mover's guest to settle on a board with a card to turn.
  const board = await waitFor(async () => {
    const now = await boardOf(mover);
    return now.some(card => card.present && card.face === '') ? now : undefined;
  }, 'a hidden card');
  board.forEach((card, slot) => { if (card.face) seen[slot] = card.face; });
  const turned = board.findIndex(card => card.present && card.face !== '');
  const slot = choose(board, seen, turned);
  const before = (await statusOf(mover)).version;
  // The guest takes no proposal while an earlier version is still
  // animating, so a click can land on a card that is not listening yet. Do
  // what a person does: click again.
  const start = Date.now();
  let face;
  let clicks = 0;
  for (; clicks < 8 && !face; clicks++) {
    await mover.frame.locator(GRID).nth(slot).click({ force: true });
    if (!shotReveal) {
      await sleep(220);
      await shot(mover, '03-flipping');
      shotReveal = true;
    }
    face = await waitFor(async () => {
      const mine = (await boardOf(mover))[slot];
      if (mine?.face) return mine.face;
      return (await statusOf(mover)).version !== before && mine?.present === false ? 'captured' : undefined;
    }, `${mover.name}'s card to turn`, 4000).catch(() => undefined);
  }
  if (!face) throw new Error(`${mover.name}'s card never turned; status ${JSON.stringify(await statusOf(mover))}`);
  const own = Date.now() - start;
  await waitFor(async () => {
    const theirs = (await boardOf(other))[slot];
    return theirs?.face || theirs?.present === false;
  }, `${other.name} to see the card`);
  const others = Date.now() - start;
  if (face && face !== 'captured') seen[slot] = face;
  latencies.push({ move: ++moves, by: mover.name, slot, clicks, ownMs: own, otherMs: others });
  console.log(JSON.stringify(latencies[latencies.length - 1]));

  if (turned >= 0) {
    const matched = seen[slot] === board[turned].face;
    if (matched && !shotCapture) {
      await sleep(350);
      await shot(mover, '05-capturing');
      await shot(other, '05-capturing');
      shotCapture = true;
    }
    if (!matched && !shotTimer) {
      await sleep(900);
      await shot(mover, '04-mismatch-timer');
      await shot(other, '04-mismatch-timer');
      shotTimer = true;
    }
    // A match is captured at once. A mismatch waits for the hide-cards
    // timer: nobody clicks, a guest's tick asks the table to look at its
    // clock, and the table hides the cards and passes the turn.
    const hideStart = Date.now();
    await waitFor(async () => (await statusOf(other)).note === 'Your turn.' ||
      (await statusOf(other)).note.includes('won') || (await statusOf(other)).note.includes('tie'),
    'the turn to pass', 30_000);
    if (!matched) console.log(JSON.stringify({ hiddenByTimerAfterMs: Date.now() - hideStart }));
  }
}

} catch (error) {
  // Leave something to look at: what each player had on screen, and what
  // their consoles said.
  for (const who of players) {
    await who.page.screenshot({ path: resolve(shots, `error-${who.name.toLowerCase()}.png`) }).catch(() => {});
    console.error(JSON.stringify({ who: who.name, status: await statusOf(who).catch(() => null), console: who.problems.slice(0, 8), denials: who.denials.slice(-3), guest: await who.frame.evaluate(() => ({ log: globalThis.__boardgameFabricLog.slice(-6), dataset: { ...document.querySelector('#guest-root').dataset }, stacks: document.querySelector('boardgame-render-game')?.renderer?.shadowRoot?.querySelectorAll('boardgame-component-stack').length, cards: document.querySelector('boardgame-render-game')?.renderer?.shadowRoot?.querySelectorAll('boardgame-card').length })).catch(error => String(error)) }));
  }
  await browser.close();
  throw error;
}

await sleep(1500);
await shot(alice, '06-finished');
await shot(bob, '06-finished');
const end = await Promise.all(players.map(statusOf));

// The same moves as the two bridges saw them, without the animation: from a
// proposal leaving the mover's guest to the view that carries its version
// reaching that guest, and reaching the other one.
const logs = await Promise.all(players.map(who => who.frame.evaluate(() => globalThis.__boardgameFabricLog)));
const bridge = [];
logs.forEach((log, index) => {
  const theirs = logs[1 - index].filter(entry => entry.what === 'view');
  log.forEach((entry, at) => {
    if (entry.what !== 'sent' || entry.kind !== 'play') return;
    const lastSeen = log.slice(0, at).filter(prior => prior.what === 'view').pop()?.version ?? -1;
    const mine = log.slice(at).find(later => later.what === 'view' && later.version > lastSeen);
    const settledAt = log.slice(at).find(later => later.what === 'settled' || later.what === 'refused');
    const other = mine && theirs.find(view => view.version >= mine.version);
    if (!mine) return;
    bridge.push({
      toOwnViewMs: Math.round(mine.wall - entry.wall),
      toSendSettledMs: settledAt ? Math.round(settledAt.wall - entry.wall) : null,
      toOtherViewMs: other ? Math.round(other.wall - entry.wall) : null,
    });
  });
});
const spread = list => {
  const sorted = list.filter(value => value !== null).sort((a, b) => a - b);
  return { n: sorted.length, min: sorted[0], median: sorted[Math.floor(sorted.length / 2)], max: sorted[sorted.length - 1] };
};
const ownMs = latencies.map(entry => entry.ownMs).sort((a, b) => a - b);
const otherMs = latencies.map(entry => entry.otherMs).sort((a, b) => a - b);
const median = list => list[Math.floor(list.length / 2)];
console.log(JSON.stringify({
  moves,
  result: end.map(status => `${status.who}: ${status.note}`),
  // Click to the card's face showing, animation included.
  clickToFaceMs: { own: { median: median(ownMs), max: ownMs[ownMs.length - 1] }, other: { median: median(otherMs), max: otherMs[otherMs.length - 1] } },
  // Proposal sent to the new view arriving, as each guest's bridge timed it.
  proposalToViewMs: {
    own: spread(bridge.map(entry => entry.toOwnViewMs)),
    sendSettled: spread(bridge.map(entry => entry.toSendSettledMs)),
    other: spread(bridge.map(entry => entry.toOtherViewMs)),
  },
  consoleErrors: { alice: alice.problems.slice(0, 5), bob: bob.problems.slice(0, 5) },
}));
await browser.close();
