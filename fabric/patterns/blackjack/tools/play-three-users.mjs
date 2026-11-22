// Plays a whole game of blackjack in a browser as three different people and
// one watcher who owns the table's space, against a piece already deployed
// to a local Common Fabric toolshed, and checks at every decision that each
// of them was handed their own hole card and nobody else's.
//
//   node tools/play-three-users.mjs <piece url> [screenshot dir]
//
// Each person is a separate browser context with its own storage, and so its
// own identity: the script registers a fresh passphrase identity in each, as
// a person would on the login screen, and grants it access to the table's
// space with `cf acl set`, run as the key that created the space: WRITE for
// the three players, and OWNER for the watcher, who never sits. The guest in
// each context is the unmodified one.
//
// Environment: as for ../../memory/tools/play-two-users.mjs
//   CF_LABS_ROOT, CF_IDENTITY, CF_SPACE, CF_API_URL, HEADED=1, PLAYWRIGHT
//   ROUNDS        how many rounds to play (default 1)
//   WATCHER       the watcher's access to the space (default OWNER)
//
// "Handed" is checked in two places in each unmodified client: the state the
// guest's renderer host was given (so, what crossed into the frame for that
// identity's seat), and what the cards on screen show.
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
const rounds = process.env.ROUNDS ?? '1';
if (!url) {
  console.error('usage: node play-three-users.mjs <piece url> [screenshot dir]');
  process.exit(2);
}
for (const name of ['CF_LABS_ROOT', 'CF_IDENTITY', 'CF_SPACE', 'CF_API_URL']) {
  if (!process.env[name]) {
    console.error(`${name} is not set`);
    process.exit(2);
  }
}
mkdirSync(shots, { recursive: true });

const sleep = ms => new Promise(done => setTimeout(done, ms));

function grant(did, access) {
  execFileSync(resolve(process.env.CF_LABS_ROOT, 'bin/cf'), ['acl', 'set', did, access], {
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

async function person(browser, name, access = 'WRITE') {
  const context = await browser.newContext({ viewport: { width: 1000, height: 900 } });
  const page = await context.newPage();
  const problems = [];
  page.on('console', message => {
    if (message.type() === 'error' || message.text().includes('boardgame guest')) problems.push(message.text().slice(0, 400));
  });
  page.on('pageerror', error => problems.push(String(error).slice(0, 400)));
  const did = await register(page);
  grant(did, access);
  await page.reload();
  const frame = await guestFrame(page);
  return { name, page, frame, did, problems };
}

const statusOf = who => who.frame.evaluate(() => ({
  who: document.querySelector('#fabric-status .who')?.textContent ?? '',
  note: document.querySelector('#fabric-status .note')?.textContent ?? '',
  error: document.querySelector('#fabric-status .error')?.textContent ?? '',
  version: document.querySelector('#guest-root')?.dataset.version ?? '',
  viewer: document.querySelector('#guest-root')?.dataset.viewer ?? '',
}));

async function waitFor(check, what, timeout = 60_000) {
  const started = Date.now();
  for (;;) {
    const value = await check();
    if (value) return value;
    if (Date.now() - started > timeout) throw new Error(`timed out waiting for ${what}`);
    await sleep(60);
  }
}

/** Waits until a guest shows the newest version its bridge has received. */
async function settled(who) {
  await waitFor(() => who.frame.evaluate(() => {
    const views = globalThis.__boardgameFabricLog.filter(entry => entry.what === 'view');
    const newest = views[views.length - 1]?.version;
    return String(newest) === document.querySelector('#guest-root')?.dataset.version;
  }), `${who.name}'s guest to catch up`);
  await sleep(900);
}

/**
 * What this identity's unmodified client holds for every seat's hand.
 * `given` is the state its renderer host was handed: for each card, the
 * card's values, or null where the view holds no face. `shown` is whether
 * each card on screen is face up.
 */
const handsOf = who => who.frame.evaluate(() => {
  const render = document.querySelector('boardgame-render-game');
  const state = render?.state;
  if (!state) return null;
  const given = state.Players.map(player => (player.Hand?.Components ?? []).map(component =>
    component && component.Values ? `${component.Values.Rank}/${component.Values.Suit}` : null));
  const panels = [...(render.renderer?.shadowRoot?.querySelectorAll('boardgame-player-panel') ?? [])];
  const cardsIn = root => {
    const out = [];
    const walk = node => {
      for (const child of node.querySelectorAll('*')) {
        if (child.localName === 'boardgame-card') out.push(child);
        if (child.shadowRoot) walk(child.shadowRoot);
      }
    };
    walk(root);
    return out;
  };
  const shown = panels.map(panel => cardsIn(panel).map(card => card.faceUp === true));
  return {
    version: Number(document.querySelector('#guest-root')?.dataset.version),
    viewer: Number(document.querySelector('#guest-root')?.dataset.viewer),
    currentPlayer: render.currentPlayerIndex,
    hidden: state.Players.map(player => (player.HiddenHand?.Indexes ?? []).length),
    values: state.Players.map(player => player.Computed?.HandValue ?? null),
    given,
    shown,
  };
});

const shot = (who, label) => who.page.screenshot({ path: resolve(shots, `${label}-${who.name.toLowerCase()}.png`) });

// ---------------------------------------------------------------------------

const browser = await chromium.launch({ headless: !process.env.HEADED });
const players = [];
for (const name of ['Alice', 'Bob', 'Carol']) players.push(await person(browser, name));
const watcher = await person(browser, 'Dave', process.env.WATCHER ?? 'OWNER');
const everyone = [...players, watcher];
console.log(JSON.stringify(Object.fromEntries(everyone.map(who => [who.name, who.did]))));

/** One line per check; the run fails at the end if any is false. */
const checks = [];
function check(ok, what, detail) {
  checks.push({ ok, what });
  if (!ok) console.log(JSON.stringify({ FAILED: what, detail }));
}

/**
 * At one version, everyone's hands side by side: a seat's still-hidden hole
 * card is given, and shown, to that seat and to nobody else.
 */
async function checkSecrecy(label) {
  for (const who of everyone) await settled(who);
  const hands = await Promise.all(everyone.map(handsOf));
  const version = hands[0].version;
  check(hands.every(hand => hand.version === version), `${label}: everyone is on the same version`, hands.map(hand => hand.version));
  check(hands.every((hand, index) => hand.viewer === (index < players.length ? index : -1)),
    `${label}: each guest is drawing its own seat's view`, hands.map(hand => hand.viewer));
  let holeCards = 0;
  for (let seat = 0; seat < players.length; seat++) {
    const own = hands[seat];
    // The owner's view says whether this seat's first card is still hidden.
    if (own.hidden[seat] === 0) continue;
    holeCards++;
    const face = own.given[seat][0];
    check(face !== null, `${label}: seat ${seat} was given its own hole card`, own.given[seat]);
    check(own.shown[seat]?.[0] === true, `${label}: seat ${seat} sees its own hole card face up`, own.shown[seat]);
    everyone.forEach((other, index) => {
      if (index === seat) return;
      const theirs = hands[index];
      check(theirs.given[seat][0] === null,
        `${label}: ${other.name} was not given seat ${seat}'s hole card`, theirs.given[seat]);
      check(theirs.shown[seat]?.[0] === false,
        `${label}: ${other.name} sees seat ${seat}'s hole card face down`, theirs.shown[seat]);
      // Nor anywhere else in what they were given.
      // Nor anywhere else in what they were given. One deck is dealt, so a
      // rank and suit names one card.
      check(!theirs.given.flat().includes(face),
        `${label}: seat ${seat}'s hole card (${face}) is nowhere in what ${other.name} was given`, theirs.given);
    });
  }
  console.log(JSON.stringify({ secrecy: label, version, holeCardsStillHidden: holeCards,
    given: Object.fromEntries(everyone.map((who, index) => [who.name, hands[index].given])) }));
  return hands;
}

const latencies = [];
let moves = 0;
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
  await waitFor(() => watcher.frame.evaluate(() => document.querySelectorAll('#fabric-lobby li').length === 3),
    'the watcher to see three seats');
  check(await players[0].frame.evaluate(() => document.querySelector('#fabric-lobby h1')?.textContent) === 'Blackjack',
    'the lobby is titled from the table\'s info');
  for (const who of everyone) await shot(who, '01-lobby');

  // The deal.
  await players[0].frame.locator('#fabric-lobby select[name="maxrounds"]').selectOption(rounds);
  const dealt = Date.now();
  await players[0].frame.getByRole('button', { name: 'Deal' }).click();
  for (const who of everyone) {
    await waitFor(async () => (await handsOf(who))?.given.every(hand => hand.length === 2), `${who.name} to see the deal`, 180_000);
  }
  console.log(JSON.stringify({ dealToEveryoneMs: Date.now() - dealt }));
  for (const [seat, who] of players.entries()) {
    await waitFor(async () => (await statusOf(who)).viewer === String(seat), `${who.name}'s own view`, 120_000);
  }
  await checkSecrecy('after the deal');
  for (const who of everyone) await shot(who, '02-dealt');

  // Play: hit below 15, otherwise stand.
  for (let guard = 0; guard < 80; guard++) {
    const notes = await Promise.all(everyone.map(async who => (await statusOf(who)).note));
    if (notes[0].includes('won') || notes[0].includes('tie')) break;
    const seat = await waitFor(async () => {
      const now = await Promise.all(players.map(async who => (await statusOf(who)).note));
      const mine = now.map((note, index) => note === 'Your turn.' ? index : -1).filter(index => index >= 0);
      if (now.some(note => note.includes('won') || note.includes('tie'))) return 'over';
      return mine.length === 1 ? String(mine[0]) : undefined;
    }, 'someone to have the turn', 120_000);
    if (seat === 'over') break;
    const mover = players[Number(seat)];
    const hands = await checkSecrecy(`before move ${moves + 1}`);
    const value = hands[Number(seat)].values[Number(seat)] ?? 0;
    const action = value < 15 ? 'Hit' : 'Stand';
    const before = (await statusOf(mover)).version;
    const start = Date.now();
    let advanced = false;
    for (let clicks = 0; clicks < 6 && !advanced; clicks++) {
      await mover.frame.getByRole('button', { name: action, exact: true }).click({ force: true });
      advanced = await waitFor(async () => (await statusOf(mover)).version !== before, `${mover.name}'s ${action} to land`, 20_000)
        .catch(() => false);
    }
    if (!advanced) throw new Error(`${mover.name}'s ${action} never landed; ${JSON.stringify(await statusOf(mover))}`);
    const own = Date.now() - start;
    await waitFor(async () => (await Promise.all(everyone.map(async who => (await statusOf(who)).version !== before))).every(Boolean),
      'everyone to see the move', 120_000);
    const everyone_ = Date.now() - start;
    // How many versions the proposal produced: the move and its fix-ups.
    for (const who of everyone) await settled(who);
    const versions = Number((await statusOf(mover)).version) - Number(before);
    latencies.push({ move: ++moves, by: mover.name, action, handValue: value, versions, ownMs: own, everyoneMs: everyone_ });
    console.log(JSON.stringify(latencies[latencies.length - 1]));
    if (moves === 1) {
      await sleep(400);
      for (const who of everyone) await shot(who, '03-first-move');
    }
  }
  await waitFor(async () => (await Promise.all(everyone.map(async who => {
    const note = (await statusOf(who)).note;
    return note.includes('won') || note.includes('tie');
  }))).every(Boolean), 'everyone to see the end', 120_000);
} catch (error) {
  for (const who of everyone) {
    await who.page.screenshot({ path: resolve(shots, `error-${who.name.toLowerCase()}.png`) }).catch(() => {});
    console.error(JSON.stringify({
      who: who.name,
      status: await statusOf(who).catch(() => null),
      console: who.problems.slice(0, 8),
      log: await who.frame.evaluate(() => globalThis.__boardgameFabricLog.slice(-6)).catch(cause => String(cause)),
    }));
  }
  await browser.close();
  throw error;
}

for (const who of everyone) await settled(who);
await sleep(1500);
for (const who of everyone) await shot(who, '04-finished');
const end = await Promise.all(everyone.map(statusOf));
check(new Set(end.map(status => status.note)).size === 1, 'everyone is told the same result', end.map(status => status.note));
const motion = await Promise.all(everyone.map(who => who.frame.evaluate(() => {
  const hooks = globalThis.__bgAnimTestHooks ?? {};
  return { gateOpens: hooks.gateOpens, plays: hooks.plays, watchdogFirings: hooks.watchdogFirings };
})));
check(motion.every(counters => counters.watchdogFirings === 0), 'no animation watchdog fired', motion);
check(motion.every(counters => counters.plays > 0), 'animations played in every guest', motion);

// Proposal sent to the view carrying its version reaching the mover's guest,
// and to `send()` resolving, which is the table's commit settling. Each pair
// also says how many versions the proposal produced.
const logs = await Promise.all(players.map(who => who.frame.evaluate(() => globalThis.__boardgameFabricLog)));
const toOwnView = [];
const toSettled = [];
const split = [];
for (const log of logs) {
  log.forEach((entry, at) => {
    if (entry.what !== 'sent' || entry.kind !== 'play') return;
    const lastSeen = log.slice(0, at).filter(prior => prior.what === 'view').pop()?.version ?? -1;
    const mine = log.slice(at).find(later => later.what === 'view' && later.version > lastSeen);
    const settledAt = log.slice(at).find(later => later.what === 'settled' || later.what === 'refused');
    if (!mine) return;
    toOwnView.push(Math.round(mine.wall - entry.wall));
    if (settledAt?.what === 'settled') {
      toSettled.push(Math.round(settledAt.wall - entry.wall));
      split.push({ versions: mine.version - lastSeen, settledMs: Math.round(settledAt.wall - entry.wall), viewMs: Math.round(mine.wall - entry.wall) });
    }
  });
}
toOwnView.sort((a, b) => a - b);
toSettled.sort((a, b) => a - b);
// Every proposal's `send()` round trip, by kind: sitting writes a table of a
// few hundred bytes, a move one of tens of kilobytes.
const everyLog = await Promise.all(everyone.map(who => who.frame.evaluate(() => globalThis.__boardgameFabricLog)));
const sendMsByKind = {};
for (const log of everyLog) {
  log.forEach((entry, at) => {
    if (entry.what !== 'sent') return;
    const done = log.slice(at).find(later => later.what === 'settled' || later.what === 'refused');
    if (done?.what === 'settled') (sendMsByKind[entry.kind] ??= []).push(Math.round(done.wall - entry.wall));
  });
}
const failed = checks.filter(item => !item.ok);
console.log(JSON.stringify({
  moves,
  result: end.map(status => `${status.who}: ${status.note}`),
  motion,
  proposalToOwnViewMs: { n: toOwnView.length, min: toOwnView[0], median: toOwnView[Math.floor(toOwnView.length / 2)], max: toOwnView[toOwnView.length - 1] },
  proposalToSendSettledMs: { n: toSettled.length, min: toSettled[0], median: toSettled[Math.floor(toSettled.length / 2)], max: toSettled[toSettled.length - 1] },
  split,
  sendMsByKind,
  checks: { passed: checks.length - failed.length, failed: failed.length, failures: failed.map(item => item.what) },
  consoleErrors: Object.fromEntries(everyone.map(who => [who.name, who.problems.slice(0, 5)])),
}));
await browser.close();
if (failed.length > 0) process.exit(1);
