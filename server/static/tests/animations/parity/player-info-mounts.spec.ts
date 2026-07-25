import { test, expect } from '@playwright/test';
import { createOfflineGame } from '../helpers.js';

// Regression test for the roster renderer-loaded binding (pre-existing bug,
// present since a87137230): boardgame-player-roster forwarded its loaded
// flag as `?renderer-loaded=`, an ATTRIBUTE binding, but the roster-item's
// `rendererLoaded` property declares no `attribute:` option, so Lit observes
// `rendererloaded` (lowercased camelCase, no hyphen). The flag never
// arrived, so no per-game player-info renderer (pig's Round Score, memory's
// pair counts, etc.) ever mounted in the live app. The binding is now a
// property binding (`.rendererLoaded=`), matching the roster-item's own
// downstream forwarding style.
test('per-game player-info renderers mount in the roster', async ({ page }) => {
  test.setTimeout(180_000);
  await createOfflineGame(page, 'pig');
  const found = await page.waitForFunction(() => {
    const walk = (root: Document | ShadowRoot): boolean => {
      for (const el of Array.from(root.querySelectorAll('*'))) {
        if (el.tagName.toLowerCase() === 'boardgame-render-player-info-pig') return true;
        const sr = (el as Element & { shadowRoot: ShadowRoot | null }).shadowRoot;
        if (sr && walk(sr)) return true;
      }
      return false;
    };
    return walk(document);
  }, undefined, { timeout: 20000 }).then(() => true).catch(() => false);
  expect(found, 'the pig player-info renderer must mount inside the roster').toBe(true);
});
