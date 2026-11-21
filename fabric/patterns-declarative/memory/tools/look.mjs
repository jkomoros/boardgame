// Opens the piece as one new identity and reports what is on the page.
//   node tools/look.mjs <piece url> [screenshot]
import { chromium, player, requireEnv, sleep, tableOf } from './lib.mjs';
requireEnv();
const browser = await chromium.launch();
const who = await player(browser, 'Looker', process.argv[2]);
await sleep(3000);
console.log(JSON.stringify(await tableOf(who), null, 1));
console.log(JSON.stringify(who.problems.slice(0, 10), null, 1));
if (process.argv[3]) await who.page.screenshot({ path: process.argv[3] });
await browser.close();
