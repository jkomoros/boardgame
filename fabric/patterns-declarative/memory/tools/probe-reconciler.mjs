// Drives `../../reconciler/main.tsx` as one identity: presses Next six times
// and reports what the reconciler did to the elements and whether cards flew.
//   node tools/probe-reconciler.mjs <piece url> [screenshot dir]
import { mkdirSync } from 'node:fs';
import { resolve } from 'node:path';
import { chromium, hooksOf, player, requireEnv, runningOf, sleep, traceOf, waitFor } from './lib.mjs';
requireEnv();
const shots = process.argv[3] ? resolve(process.argv[3]) : null;
if (shots) mkdirSync(shots, { recursive: true });
const browser = await chromium.launch();
const who = await player(browser, 'Prober', process.argv[2]);
const state = () => who.page.evaluate(() => {
  const table = globalThis.__deep('bg-table')[0];
  const stacks = globalThis.__deep('bg-stack', table);
  return {
    step: Number(globalThis.__deep('#step')[0]?.textContent ?? -1),
    stats: { ...table.stats },
    animating: table.animating,
    stacks: Object.fromEntries(stacks.map(stack => [stack.key,
      [...stack.renderRoot.querySelectorAll('boardgame-card:not([spacer])')].map(card => card.id)])),
  };
});
await waitFor(async () => Object.keys((await state()).stacks).length === 3, 'three stacks', 60_000);
await sleep(1500);
const steps = [];
for (let step = 1; step <= 6; step++) {
  let running = 0;
  const watch = setInterval(() => {
    runningOf(who).then(count => { running = Math.max(running, count); }, () => {});
  }, 30);
  await who.page.locator('#next').click();
  await waitFor(async () => (await state()).step === step, `step ${step}`, 20_000);
  if (shots && step === 2) { await sleep(150); await who.page.screenshot({ path: resolve(shots, 'mid-flight.png') }); }
  await sleep(1200);
  clearInterval(watch);
  const now = await state();
  steps.push({ step, runningAnimations: running, cycles: now.stats.cycles, animatedCycles: now.stats.animatedCycles, interrupted: now.stats.interrupted });
  console.log(JSON.stringify(steps[steps.length - 1]));
}
if (shots) await who.page.screenshot({ path: resolve(shots, 'end.png') });
console.log(JSON.stringify({ end: await state(), animatorHooks: await hooksOf(who), reconciler: await traceOf(who),
  consoleProblems: who.problems.filter(text => !/lacks READ|event attention|slug|selected piece/.test(text)).slice(0, 6) }, null, 1));
await browser.close();
