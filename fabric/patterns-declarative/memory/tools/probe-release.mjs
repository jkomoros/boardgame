// Seats two identities at `../probe-release.tsx`, deals, and reports which
// reads of the released view reached the screen.
//   node tools/probe-release.mjs <piece url> [screenshot]
import { chromium, player, requireEnv, sleep } from './lib.mjs';
requireEnv();
const browser = await chromium.launch();
const url = process.argv[2];
const one = await player(browser, 'One', url);
const two = await player(browser, 'Two', url);
await one.page.locator('#sit').click();
await sleep(3000);
await two.page.locator('#sit').click();
await sleep(3000);
await one.page.locator('#deal').click();
await sleep(6000);
const read = who => who.page.evaluate(() => {
  const text = selector => globalThis.__deep(selector).map(node => node.textContent.trim());
  const stack = globalThis.__deep('bg-stack')[0];
  return {
    direct: text('#direct'),
    derived: text('#derived'),
    seats: text('#seats .seat'),
    seatsBox: text('#seats'),
    cardsInStack: stack?.renderRoot.querySelectorAll('boardgame-card').length ?? null,
    blocked: globalThis.__deep('cf-cfc-blocked').map(node =>
      node.parentElement?.id || node.parentElement?.localName),
  };
});
console.log(JSON.stringify({ one: await read(one), two: await read(two) }, null, 1));
if (process.argv[3]) await one.page.screenshot({ path: process.argv[3] });
await browser.close();
