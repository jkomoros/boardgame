import { test, expect } from '@playwright/test';
import { createOfflineGame, gateSnapshot, expectCleanGate } from '../helpers.js';

// Task 10 (plan Task 10, #714's second Phase 2 gap): boardgame-player-roster
// is a DOM SIBLING of boardgame-render-game (both children of
// boardgame-game-view), so the `will-animate`/`animation-done` events
// bubbling out of a player-info renderer's boardgame-status-text /
// boardgame-fading-text (gate participants since Task 4/Phase 1) never reach
// render-game's own gate listeners -- those are installed on render-game
// itself, and a sibling's bubble path never crosses it. This suite is the
// roster-gating deliverable's only parity witness (design doc Phase 2 item 2
// / plan Task 10), asserting BOTH directions:
//
//   (a) a roster animatable's will-animate, forwarded while a real board
//       cycle is open, HOLDS the gate -- close waits for the roster
//       participant's settle, not just the board's own animation.
//   (b) a roster animation with NO board cycle open must NOT open, extend,
//       or otherwise disturb the gate (the HARNESS-CRITIC REQUIREMENT
//       (gap 3) guard: forwarding will-animate only while
//       renderGame.isAnimating is true, so e.g. a hover-triggered roster
//       fade can never wedge/queue the gate).
//
// Vehicle choice (documented per plan instructions): a real memory scoring
// flow (two matching reveals) or a real per-game player-info renderer (e.g.
// pig's "Round Score" boardgame-status-text) were both considered first --
// the most literal restatement of #714's own checklist wording. Memory's is
// ruled out because its deck is shuffled with an unseeded RNG and client
// state sanitizes card Values even in the admin view (see trace.spec.ts's
// "memory: reveal one card" comment, which deliberately limits itself to a
// single non-matching reveal for exactly this reason) -- there is no
// scripted way to guarantee a match lands inside a capture window. Pig's
// player-info renderer was the more promising vehicle (a real move,
// RoundScore genuinely scoring on 5 of 6 rolls, status-text bound to a
// real per-player counter) until investigation surfaced a pre-existing,
// unrelated bug that makes it a dead end: boardgame-player-roster.ts's
// template forwards its loaded flag to <boardgame-player-roster-item> via
// `?renderer-loaded="${this.rendererLoaded}"` (a boolean-ATTRIBUTE
// binding), but boardgame-player-roster-item.ts's `rendererLoaded`
// property has no explicit `attribute:` option, so Lit derives the
// observed attribute name `rendererloaded` (no hyphen) -- confirmed
// empirically via `customElements.get('boardgame-player-roster-item')
// .observedAttributes`. The mismatch means `rendererLoaded` never reaches
// true on the roster item, so `boardgame-render-player-info.ts`'s
// `instantiateRenderer()` guard (`if (!this.active || !this.rendererLoaded
// || ...) return;`) never passes -- NO per-game player-info renderer
// mounts in the live app currently, for any game (confirmed: waiting 15s+
// on a fresh pig game, <boardgame-render-player-info>'s shadow root still
// only contains its placeholder comment). This is a real, separate,
// pre-existing bug (introduced 2026-02-07, long before this plan), flagged
// separately -- not something Task 10 owns or should fix incidentally.
//
// The vehicle actually used, per the plan's own sanctioned fallback
// ("a fixture-shaped test that dispatches a genuine gated play() on an
// animatable INSIDE the roster subtree while a real board cycle runs"):
// mount a REAL boardgame-fading-text element (the actual production class,
// not a stand-in) as a light-DOM child of the real, live, already-mounted
// <boardgame-player-roster> element from a real pig game, and drive a
// genuine `.trigger` change -- exactly boardgame-status-text's own
// mechanism -- while a real board cycle (a real Roll-die move) is open.
// boardgame-player-roster's shadow template has no <slot>, so this child
// renders nothing visible, but connectedCallback / composed event bubbling
// (the only things this test needs) are unaffected by slot assignment.
// This is the closest deterministic approximation of "a roster-hosted
// animatable, participating in a real board cycle" available while the
// separate bug above stands.
const PARITY_TIMEOUT_MS = 180_000;

// Deep-walks shadow roots to find the first match (mirrors registry.spec.ts /
// trace.spec.ts's established pattern -- boardgame-player-roster lives
// behind at least one shadow boundary, inside boardgame-game-view).
function deepQueryFirst(root: Document | ShadowRoot | Element, selector: string): Element | null {
  const direct = root.querySelector(selector);
  if (direct) return direct;
  for (const el of Array.from(root.querySelectorAll('*'))) {
    const sr = (el as Element & { shadowRoot: ShadowRoot | null }).shadowRoot;
    if (sr) {
      const hit = deepQueryFirst(sr, selector);
      if (hit) return hit;
    }
  }
  return null;
}

async function mountRosterFadingText(page: import('@playwright/test').Page): Promise<void> {
  await page.evaluate(async (deepQueryFirstSrc: string) => {
    await import('/src/components/boardgame-fading-text.ts');
    const deepQueryFirst = eval(deepQueryFirstSrc) as
      (root: Document | ShadowRoot | Element, selector: string) => Element | null;
    const roster = deepQueryFirst(document, 'boardgame-player-roster');
    if (!roster) throw new Error('boardgame-player-roster not found');
    const el = document.createElement('boardgame-fading-text') as any;
    el.id = 'task10-roster-probe';
    el.autoMessage = 'fixed';
    el.message = 'Task 10 probe';
    roster.appendChild(el);
    // Stashed on window: the probe lives inside boardgame-game-view's
    // shadow root (as a light-DOM child of the roster), so
    // document.getElementById cannot find it later -- shadow roots are not
    // part of the document's flat ID index.
    (window as any).__task10RosterProbe = el;
    el.trigger = 1; // establish the baseline; no fade fires on this first set
    await el.updateComplete;
  }, `(${deepQueryFirst.toString()})`);
}

test.describe('player-info gate participation', () => {
  test('roster fading-text holds the gate: close waits for its settle during a real board cycle', async ({ page }) => {
    test.setTimeout(PARITY_TIMEOUT_MS);
    await createOfflineGame(page, 'pig');
    await expect(page.getByRole('button', { name: 'Roll die' })).toBeEnabled({ timeout: 30000 });

    // Drain game-creation setup completely before measuring (same rationale
    // as trace.spec.ts / registry.spec.ts's pig scenarios).
    const setup = await gateSnapshot(page);
    await expectCleanGate(page, setup, 60000, { allowAlreadySettled: true });

    await mountRosterFadingText(page);

    const openedBefore = (await gateSnapshot(page)).gateOpens;

    // All timing below is measured with in-page performance.now()
    // timestamps and in-page rAF polling loops -- deliberately avoiding
    // Node-side round-trip polling (locator.evaluate() per iteration),
    // whose latency is comparable to the short (sub-second) windows being
    // measured and would make the comparison noisy or miss them entirely.
    let attempt = 0;
    let rollResult: { diePlayed: boolean } = { diePlayed: false };
    for (; attempt < 5 && !rollResult.diePlayed; attempt++) {
      const logStart = await page.evaluate(() => (window as any).__bgAnimTestHooks.log.length);
      await expect(page.getByRole('button', { name: 'Roll die' })).toBeEnabled({ timeout: 15000 });
      await page.getByRole('button', { name: 'Roll die' }).click();

      // Wait (in-page) for the die's own real board cycle to genuinely
      // start (a 'play' hook for boardgame-die -- see registry.spec.ts's
      // precedent: a roll landing on the same face already showing plays
      // no animation, 1-in-6, hence the retry loop), then IMMEDIATELY
      // (same evaluate call, no round-trip gap) fire a real gated play()
      // on the roster-mounted probe while the board cycle is provably
      // still open -- this is the exact window the plan's guard
      // (`if (!renderGame.isAnimating) return;`) must keep passable.
      rollResult = await page.evaluate(async (start: number) => {
        const hooks = (window as any).__bgAnimTestHooks;
        const diePlayed = await new Promise<boolean>((resolve) => {
          const deadline = performance.now() + 2000;
          const check = () => {
            if (hooks.log.slice(start).some((e: any) => e.ev === 'play' && e.detail === 'boardgame-die')) {
              resolve(true);
              return;
            }
            if (performance.now() > deadline) { resolve(false); return; }
            requestAnimationFrame(check);
          };
          check();
        });
        if (!diePlayed) return { diePlayed: false };
        const el = (window as any).__task10RosterProbe;
        el.postAnimationDelay = 2500;
        el.trigger = 2; // real change from the established baseline -> fires the fade
        await el.updateComplete;
        return { diePlayed: true };
      }, logStart);

      if (!rollResult.diePlayed) {
        // Same face rolled again (no die animation this attempt): drain
        // whatever tiny cycle this no-op roll produced before retrying.
        const snap = await gateSnapshot(page);
        await expectCleanGate(page, { ...snap, gateOpens: 0 }, 20000, { allowAlreadySettled: true });
      }
    }
    expect(rollResult.diePlayed, 'the die must genuinely animate within 5 attempts').toBe(true);

    // Wait (in-page) for BOTH the roster probe's own settle AND the gate's
    // close, recording each one's performance.now() timestamp so the
    // comparison below is immune to Node-side scheduling noise.
    const result = await page.evaluate(async () => {
      function deepQueryFirst(root: any, selector: string): any {
        const direct = root.querySelector(selector);
        if (direct) return direct;
        for (const el of Array.from(root.querySelectorAll('*'))) {
          const sr = (el as any).shadowRoot;
          if (sr) { const hit = deepQueryFirst(sr, selector); if (hit) return hit; }
        }
        return null;
      }
      const renderGame = deepQueryFirst(document, 'boardgame-render-game');
      const hooks = (window as any).__bgAnimTestHooks;
      const startIdx = hooks.log.length; // probe trigger already landed above
      const deadline = performance.now() + 20000;
      let rosterSettleAt = -1;
      let gateCloseAt = -1;
      while (performance.now() < deadline && (rosterSettleAt < 0 || gateCloseAt < 0)) {
        if (rosterSettleAt < 0) {
          // Search from the beginning of the log (not startIdx): the
          // probe's own play()/settle may have already been recorded by
          // the time this evaluate call starts running.
          const hit = (hooks.log as Array<{ ev: string; detail?: string; t: number }>)
            .find((e) => e.ev === 'settle' && e.detail === 'boardgame-fading-text#task10-roster-probe');
          if (hit) rosterSettleAt = hit.t;
        }
        if (gateCloseAt < 0 && renderGame.isAnimating === false) {
          gateCloseAt = performance.now();
        }
        await new Promise<void>((r) => requestAnimationFrame(() => r()));
      }
      return { rosterSettleAt, gateCloseAt, timedOut: rosterSettleAt < 0 || gateCloseAt < 0 };
    });

    expect(result.timedOut, 'both the roster probe settle and gate-close must be observed within 20s').toBe(false);
    expect(result.rosterSettleAt, 'the roster probe fade must have actually played and settled').toBeGreaterThan(0);
    expect((await gateSnapshot(page)).gateOpens, 'the die roll must have opened a gate cycle')
      .toBeGreaterThan(openedBefore);

    // THE assertion: the gate must not report closed (isAnimating false)
    // strictly before the roster participant's own settle -- proof it
    // waited for the roster participant, not merely for the die's own
    // ~250ms spin. A small tolerance absorbs rAF-frame granularity on
    // both sides of the comparison, not systematic slack: pre-fix, the
    // gate closes ~250ms after the click while the roster probe (a 2500ms
    // postAnimationDelay) hasn't settled for another ~2.25s -- a gap far
    // larger than any frame-timing tolerance. See docs/superpowers/specs/
    // evidence/2026-07-25-player-info-ungated.md for the recorded
    // before/after underlying this exact mechanism.
    expect(result.gateCloseAt,
      `gate reported closed at ${result.gateCloseAt}ms (page-relative) but the roster probe did not ` +
      `settle until ${result.rosterSettleAt}ms; expected the gate to wait for the roster participant`)
      .toBeGreaterThanOrEqual(result.rosterSettleAt - 100);
  });

  test('a roster animation with no board cycle open leaves the gate untouched (non-wedging guard)', async ({ page }) => {
    test.setTimeout(PARITY_TIMEOUT_MS);
    await createOfflineGame(page, 'pig');
    await expect(page.getByRole('button', { name: 'Roll die' })).toBeEnabled({ timeout: 30000 });

    const setup = await gateSnapshot(page);
    await expectCleanGate(page, setup, 60000, { allowAlreadySettled: true });

    await mountRosterFadingText(page);

    const renderGame = page.locator('boardgame-render-game').first();
    expect(await renderGame.evaluate((el: any) => el.isAnimating === false)).toBe(true);

    const before = await gateSnapshot(page);

    // Directly drive the roster-mounted fading-text's real play() with NO
    // board cycle open -- modeling the plan's own example of an
    // out-of-band roster animation (e.g. a hover-triggered fade) that must
    // never open/wedge the gate.
    const result = await page.evaluate(async () => {
      const el = (window as any).__task10RosterProbe;
      if (!el) return { found: false as const };
      const done = new Promise<void>((resolve) => {
        el.addEventListener('animation-done', () => resolve(), { once: true });
        setTimeout(resolve, 5000);
      });
      el.trigger = 2; // real out-of-cycle change, no board move involved
      await done;
      return { found: true as const };
    });
    expect(result.found, 'roster-mounted fading-text probe must exist').toBe(true);

    // isAnimating must never have flipped true: the guard forwards
    // will-animate only while a board cycle is already open, so an
    // out-of-cycle roster animation must be invisible to render-game.
    expect(await renderGame.evaluate((el: any) => el.isAnimating === false)).toBe(true);

    const after = await gateSnapshot(page);
    expect(after.gateOpens, 'gate must not have opened for an out-of-cycle roster animation').toBe(before.gateOpens);
    expect(after.gateCloses, 'gate must not have closed for an out-of-cycle roster animation').toBe(before.gateCloses);
    expect(after.watchdogFirings, 'watchdog must never fire').toBe(before.watchdogFirings);
  });
});
