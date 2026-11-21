// Plays a whole game of Memory as two different people, against the
// declarative piece deployed to a local Common Fabric toolshed whose shell
// loaded the boardgame element bundle (see ../../README.md).
//
//   node tools/play-two-users.mjs <piece url> [screenshot dir]
//
// Each player is a separate browser context, and so a separate identity,
// registered on the shell's login screen and granted WRITE with `cf acl set`.
// Every click is a real pointer click at the card's position: the elements
// emit `propose-move` only for a trusted gesture.
//
// Environment: CF_LABS_ROOT, CF_IDENTITY, CF_SPACE, CF_API_URL as for the
// guest's script; HEADED=1 shows the browsers.
import { mkdirSync } from 'node:fs';
import { dirname, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';
import {
  cardPoint, chromium, hooksOf, player, requireEnv, runningOf, sleep, tableOf, traceOf, waitFor,
} from './lib.mjs';

const here = dirname(fileURLToPath(import.meta.url));
const url = process.argv[2];
const shots = resolve(process.argv[3] ?? resolve(here, '../screenshots'));
if (!url) {
  console.error('usage: node play-two-users.mjs <piece url> [screenshot dir]');
  process.exit(2);
}
requireEnv();
mkdirSync(shots, { recursive: true });
const shot = (who, label) =>
  who.page.screenshot({ path: resolve(shots, `${label}-${who.name.toLowerCase()}.png`) });

const browser = await chromium.launch({ headless: !process.env.HEADED });
const alice = await player(browser, 'Alice', url);
const bob = await player(browser, 'Bob', url);
const players = [alice, bob];
console.log(JSON.stringify({ alice: alice.did, bob: bob.did }));

const latencies = [];
let moves = 0;
let shotReveal = false, shotTimer = false, shotCapture = false;
const timerWaits = [];

try {
  // Seating. The table releases nothing before its first move, so the lobby
  // is two buttons that are always there.
  for (const who of players) {
    await who.page.locator('cf-input input').fill(who.name);
    await who.page.locator('#sit').click();
    await waitFor(async () => (await tableOf(who)).seats.includes(who.name), `${who.name} to be seated`);
    await waitFor(async () => (await tableOf(who)).viewer >= 0, `${who.name}'s table to know its seat`);
  }
  await waitFor(async () => (await tableOf(alice)).seats.length === 2, 'Alice to see both seats');
  await shot(alice, '01-lobby');
  await shot(bob, '01-lobby');

  await alice.page.locator('#deal').click();
  for (const who of players) {
    await waitFor(async () => (await tableOf(who)).board.length > 0, `${who.name} to see the board`);
  }
  await sleep(1500);
  await shot(alice, '02-dealt');
  await shot(bob, '02-dealt');

  const settled = who => waitFor(async () => {
    const now = await tableOf(who);
    return !now.animating && now.version === now.queued ? now : undefined;
  }, `${who.name}'s table to settle`);

  const choose = (board, seen, turned) => {
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
  };

  const seen = [];
  for (let guard = 0; guard < 200; guard++) {
    const states = await Promise.all(players.map(settled));
    if (states.every(state => state.finished)) break;
    if (states[0].version !== states[1].version || states[0].turn === null) { await sleep(100); continue; }
    const mover = players.find((who, index) => states[index].viewer === states[index].turn);
    if (!mover) throw new Error(`nobody holds the turn: ${JSON.stringify(states.map(s => [s.viewer, s.turn]))}`);
    const other = mover === alice ? bob : alice;
    // A hidden card the table says this viewer may turn. With none, the
    // version on screen is one a timer is about to replace.
    const board = (await tableOf(mover)).board;
    if (!board.some(card => card.enabled)) {
      const started = Date.now();
      const from = states[0].version;
      await waitFor(async () => (await tableOf(mover)).queued !== from, 'the timer to hide the cards', 30_000);
      timerWaits.push(Date.now() - started);
      continue;
    }
    board.forEach((card, slot) => { if (card.face) seen[slot] = card.face; });
    const turned = board.findIndex(card => card.present && card.face !== '');
    const slot = choose(board, seen, turned);
    if (!board[slot].enabled) throw new Error(`slot ${slot} is not enabled: ${JSON.stringify(board[slot])}`);

    let running = 0;
    const watch = setInterval(() => {
      runningOf(mover).then(count => { running = Math.max(running, count); }, () => {});
    }, 50);
    const start = Date.now();
    const point = await cardPoint(mover, slot);
    await mover.page.mouse.click(point.x, point.y);
    if (!shotReveal) {
      await waitFor(async () => (await runningOf(mover)) > 0, 'the flip to start', 15_000).catch(() => {});
      await shot(mover, '03-flipping');
      shotReveal = true;
    }
    const shown = slotNow => slotNow?.face || slotNow?.present === false;
    const face = await waitFor(async () => {
      const mine = (await tableOf(mover)).board[slot];
      return mine?.face || (mine?.present === false ? 'captured' : undefined);
    }, `${mover.name}'s card to turn`, 20_000);
    const own = Date.now() - start;
    await waitFor(async () => shown((await tableOf(other)).board[slot]), `${other.name} to see the card`, 20_000);
    const others = Date.now() - start;
    if (face !== 'captured') seen[slot] = face;
    const matched = turned >= 0 && seen[slot] === board[turned].face;
    if (matched && !shotCapture) {
      for (let index = 1; index <= 8; index++) {
        await sleep(250);
        await shot(mover, `05-capturing-${index}`);
      }
      await shot(other, '05-capturing-8');
      shotCapture = true;
    }
    if (turned >= 0 && !matched && !shotTimer) {
      await sleep(900);
      await shot(mover, '04-mismatch-timer');
      await shot(other, '04-mismatch-timer');
      shotTimer = true;
    }
    await sleep(500);
    clearInterval(watch);
    latencies.push({ move: ++moves, by: mover.name, slot, face, ownMs: own, otherMs: others, runningAnimations: running });
    console.log(JSON.stringify(latencies[latencies.length - 1]));
  }
} catch (error) {
  for (const who of players) {
    await who.page.screenshot({ path: resolve(shots, `error-${who.name.toLowerCase()}.png`) }).catch(() => {});
    console.error(JSON.stringify({ who: who.name, table: await tableOf(who).catch(() => null), console: who.problems.slice(0, 8) }));
  }
  await browser.close();
  throw error;
}

await sleep(1500);
await shot(alice, '06-finished');
await shot(bob, '06-finished');
const end = await Promise.all(players.map(tableOf));
const hooks = await Promise.all(players.map(hooksOf));
const traces = await Promise.all(players.map(traceOf));
const median = list => [...list].sort((a, b) => a - b)[Math.floor(list.length / 2)];
console.log(JSON.stringify({
  moves,
  result: end.map(state => ({ seats: state.seats, finished: state.finished, winners: state.winners, won: state.won, status: state.status })),
  // The table element's own count of motion cycles, and the animator's
  // counters (boardgame-animatable-item records one play per animation).
  tableStats: end.map(state => state.stats),
  animatorHooks: hooks,
  // What the reconciler did to the elements over the whole game.
  reconciler: traces,
  blockedByPolicy: end.map(state => state.blocked),
  hiddenByTimerAfterMs: timerWaits,
  clickToFaceMs: {
    own: { median: median(latencies.map(entry => entry.ownMs)), max: Math.max(...latencies.map(entry => entry.ownMs)) },
    other: { median: median(latencies.map(entry => entry.otherMs)), max: Math.max(...latencies.map(entry => entry.otherMs)) },
  },
  consoleProblems: { alice: alice.problems.slice(0, 5), bob: bob.problems.slice(0, 5) },
}));
await browser.close();
