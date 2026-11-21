// Drives the per-stack experiment (`../per-stack.tsx`) as two identities
// until a pair has been captured, and reports what the reconciler did.
//
//   node tools/probe-per-stack.mjs <piece url> [screenshot dir]
import { mkdirSync } from 'node:fs';
import { dirname, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';
import {
  cardPoint, chromium, hooksOf, player, requireEnv, runningOf, sleep, tableOf, traceOf, waitFor,
} from './lib.mjs';

const here = dirname(fileURLToPath(import.meta.url));
const url = process.argv[2];
const shots = resolve(process.argv[3] ?? resolve(here, '../screenshots/per-stack'));
requireEnv();
mkdirSync(shots, { recursive: true });

const browser = await chromium.launch({ headless: !process.env.HEADED });
const alice = await player(browser, 'Alice', url);
const bob = await player(browser, 'Bob', url);
const players = [alice, bob];
const versionOf = async who => Number(await who.page.locator('#version').innerText());
const log = [];

try {
  for (const who of players) {
    await who.page.locator('cf-input input').fill(who.name);
    await who.page.locator('#sit').click();
    await sleep(2500);
  }
  await alice.page.locator('#deal').click();
  for (const who of players) {
    await waitFor(async () => (await tableOf(who)).board.length > 0, `${who.name} to see the board`, 60_000);
  }
  await sleep(1500);
  await alice.page.screenshot({ path: resolve(shots, '01-dealt-alice.png') });

  // Nothing here knows whose turn it is: a click by the wrong player is
  // refused by the table and changes nothing, so the other one tries.
  const act = async (label, perform) => {
    for (const who of [...players]) {
      const before = await versionOf(alice);
      let running = 0;
      const watch = setInterval(() => {
        runningOf(alice).then(count => { running = Math.max(running, count); }, () => {});
      }, 40);
      await perform(who);
      const moved = await waitFor(async () => (await versionOf(alice)) !== before, 'a new version', 6000)
        .then(() => true, () => false);
      await sleep(1200);
      clearInterval(watch);
      if (moved) {
        log.push({ step: label, by: who.name, version: await versionOf(alice), runningAnimationsSeenByAlice: running });
        console.log(JSON.stringify(log[log.length - 1]));
        if (who !== players[0]) players.reverse();
        return true;
      }
    }
    return false;
  };
  const click = slot => async who => {
    const point = await cardPoint(who, slot);
    await who.page.mouse.click(point.x, point.y);
  };

  const seen = [];
  for (let guard = 0; guard < 24; guard++) {
    const state = await tableOf(alice);
    if (state.won.some(count => count > 0)) break;
    const board = state.board;
    board.forEach((card, slot) => { if (card.face) seen[slot] = card.face; });
    const up = board.map((card, slot) => ({ card, slot })).filter(({ card }) => card.present && card.face);
    if (up.length >= 2) {
      if (!await act('hide', who => who.page.locator('#hide').click())) players.reverse();
      continue;
    }
    const hidden = board.map((card, slot) => ({ card, slot }))
      .filter(({ card }) => card.present && !card.face).map(({ slot }) => slot);
    let slot;
    if (up.length === 1) slot = hidden.find(candidate => seen[candidate] === up[0].card.face);
    else slot = hidden.find(candidate => seen[candidate] && hidden.some(other => other !== candidate && seen[other] === seen[candidate]));
    slot ??= hidden.find(candidate => !seen[candidate]) ?? hidden[0];
    if (!await act(`reveal ${slot}`, click(slot))) throw new Error('nobody could move');
  }
  await sleep(1500);
  await alice.page.screenshot({ path: resolve(shots, '02-captured-alice.png') });
  await bob.page.screenshot({ path: resolve(shots, '02-captured-bob.png') });
} catch (error) {
  await alice.page.screenshot({ path: resolve(shots, 'error-alice.png') }).catch(() => {});
  console.error(JSON.stringify({ table: await tableOf(alice).catch(() => null), console: alice.problems.slice(0, 8) }));
  await browser.close();
  throw error;
}

const end = await Promise.all([alice, bob].map(tableOf));
console.log(JSON.stringify({
  steps: log.length,
  won: end.map(state => state.won),
  tableStats: end.map(state => state.stats),
  animatorHooks: await Promise.all([alice, bob].map(hooksOf)),
  reconciler: await Promise.all([alice, bob].map(traceOf)),
  blockedByPolicy: end.map(state => state.blocked),
  consoleProblems: alice.problems.filter(text => !/lacks READ|event attention/.test(text)).slice(0, 6),
}));
await browser.close();
