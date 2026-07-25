import { test, expect } from '@playwright/test';
import { createOfflineGame, expectCleanGate, gateSnapshot, waitForAnimationCounterStability } from '../helpers.js';
import { captureTrace, expectTraceMatchesGolden } from './trace-helpers.js';

// createOfflineGame alone (create + fake sign-in + admin-mode dance) measures
// 30-60s on a loaded dev laptop, and quiescence waits add more; the 30s
// default test timeout cannot cover any of these scenarios (all four timed
// out inside setup when run with the default budget).
const PARITY_TIMEOUT_MS = 180_000;

test.describe('animation parity traces', () => {
  test('debuganimations: card move cycle', async ({ page }) => {
    test.setTimeout(PARITY_TIMEOUT_MS);
    await createOfflineGame(page, 'debuganimations');
    // Drain the game-creation pipeline (a 200+-play auto-deal) completely
    // before opening the capture window, or the Swap cycle overlaps it and
    // the window cuts creation animations mid-flight (observed as plays >
    // settles inside the window).
    const setup = await gateSnapshot(page);
    await expectCleanGate(page, setup, 60000, { allowAlreadySettled: true });
    // Setup-drain stability: the clean-gate check is point-in-time and the
    // creation pipeline can start another wave right after it passes (the
    // per-player info renderers mounting again shifted creation timing and
    // exposed exactly that race). Hold until counters are stable+balanced.
    await waitForAnimationCounterStability(page, { balance: 'plays' });
    const trace = await captureTrace(page, async () => {
      // debuganimations has no button literally labeled "Move Card" -- the
      // task brief's placeholder name doesn't match the real UI (verified
      // against examples/debuganimations/client/boardgame-render-game-
      // debuganimations.ts). The "Swap" button inside the #shortstacks
      // section drives MoveCardBetweenShortStacks, flying one card between
      // two visible stacks -- the deterministic single-card move this
      // scenario needs. Scoped to #shortstacks because "Swap" is reused
      // (ambiguously) by the token-swap section elsewhere on the page.
      const before = await gateSnapshot(page);
      await page.locator('#shortstacks').getByRole('button', { name: 'Swap' }).click();
      // Close the capture window deterministically: a bare quiescence wait
      // right after a click is vacuously true before the move's round-trip
      // opens the gate, so whether the cycle lands inside the trace would
      // otherwise be a race.
      await expectCleanGate(page, before);
    });
    // Structural with exact cycle counts: one Swap is always exactly one
    // gated cycle, but per-component play counts vary per game -- FLIP
    // skips no-op transforms, and the messy-stack rotations are hashed
    // from per-game random component ids, so how many of the re-laid-out
    // cards actually animate is not deterministic (observed 92-138 plays
    // for the identical action across two fresh games).
    expectTraceMatchesGolden(trace, 'debuganimations-card-move',
      { structural: { requiredKinds: ['boardgame-card'], exactCycles: true } });
  });

  test('memory: reveal one card', async ({ page }) => {
    test.setTimeout(PARITY_TIMEOUT_MS);
    // adminMode:false is DETERMINISM-LOAD-BEARING. With admin mode on, the
    // admin-state install races the initial player-view fetch inside
    // createOfflineGame (root-caused in the bimodality investigation:
    // whichever full-state response lands first wins and no refetch happens
    // when only admin flips), so the grid rests face-up (reveal = no flip
    // plays) or face-down (20 real flips) per run — a 21-vs-41 bimodal
    // golden. The seated creator (auto-seat + ActivateInactivePlayer fixes)
    // proposes reveals legally as player 0, so admin is unnecessary; the
    // sanitized player view deterministically pins the face-down state and
    // the reveal's REAL flip animation.
    await createOfflineGame(page, 'memory', { adminMode: false });
    // createOfflineGame only waits for <boardgame-render-game> to mount, not
    // for memory's own card grid to finish its async render (same caveat
    // documented in waapi-gate.spec.ts's memory test).
    await expect(page.locator('boardgame-card').first()).toBeAttached({ timeout: 15000 });
    // Drain the creation deal completely before the capture window opens
    // (same rationale as the debuganimations scenario).
    const setup = await gateSnapshot(page);
    await expectCleanGate(page, setup, 60000, { allowAlreadySettled: true });
    // Setup-drain stability: the clean-gate check is point-in-time and the
    // creation pipeline can start another wave right after it passes (the
    // per-player info renderers mounting again shifted creation timing and
    // exposed exactly that race). Hold until counters are stable+balanced.
    await waitForAnimationCounterStability(page, { balance: 'plays' });
    // Layout stability precondition. The reveal's FLIP measures every card;
    // if the grid is still settling (font swap, late image decode), all 20
    // survivors pick up sub-pixel deltas and play real host animations
    // (observed 41 plays) instead of being skipped as no-ops (21 plays) --
    // a bimodal flake in the exact-count golden. Wait for fonts plus a
    // stable first-card rect across consecutive frames before clicking.
    await page.evaluate(async () => {
      await (document as Document & { fonts: FontFaceSet }).fonts.ready;
      const cardRect = (): string => {
        const walk = (root: Document | ShadowRoot): Element | null => {
          for (const el of Array.from(root.querySelectorAll('*'))) {
            if (el.tagName === 'BOARDGAME-CARD') return el;
            const sr = (el as Element & { shadowRoot: ShadowRoot | null }).shadowRoot;
            if (sr) { const hit = walk(sr); if (hit) return hit; }
          }
          return null;
        };
        const el = walk(document);
        return el ? JSON.stringify(el.getBoundingClientRect()) : '';
      };
      const frame = () => new Promise<void>((r) => requestAnimationFrame(() => r()));
      const deadline = performance.now() + 10000;
      let last = cardRect();
      let stable = 0;
      while (performance.now() < deadline && stable < 5) {
        await frame();
        const cur = cardRect();
        if (cur === last && cur !== '') stable++; else { stable = 0; last = cur; }
      }
    });
    // A SINGLE reveal is the deterministic memory scenario. Two arbitrary
    // reveals branch on game logic the test cannot control: card Values are
    // sanitized out of the client state (even in the admin view -- items
    // carry only ID), so a second click either matches (capture + score
    // path) or doesn't (a timer-driven hide whose cycle races any capture
    // window) and the trace differs per branch. One reveal exercises the
    // card-flip animation cycle -- the stack-motion subject this golden
    // exists to pin -- with a single move-driven bundle pipeline. The
    // match/score path is covered separately (Phase 1 fading-text spec)
    // with branch-tolerant assertions.
    const trace = await captureTrace(page, async () => {
      const beforeClick = await gateSnapshot(page);
      await page.locator('boardgame-card:not([disabled])').first().click();
      await expectCleanGate(page, beforeClick);
    });
    expectTraceMatchesGolden(trace, 'memory-reveal-one');
  });

  test('blackjack: initial deal settles cleanly', async ({ page }) => {
    test.setTimeout(PARITY_TIMEOUT_MS);
    // Deal auto-runs on creation; captureTrace tolerates the page not
    // having animation hooks installed yet when the scenario itself is what
    // navigates to the game (see captureTrace's hooksExisted handling).
    const trace = await captureTrace(page, async () => {
      await createOfflineGame(page, 'blackjack');
      // The auto-deal may already be fully settled by the time game setup
      // finishes (allowAlreadySettled), or still mid-flight -- either way
      // this drains the whole deal pipeline before the window closes. The
      // snapshot must come from gateSnapshot (never a hand-built zero
      // object): it stamps the reload-detection probe token expectCleanGate
      // checks, and a missing token reads as "page reloaded".
      const afterCreate = await gateSnapshot(page);
      await expectCleanGate(page,
        { ...afterCreate, gateOpens: 0 },
        60000, { allowAlreadySettled: true });
    });
    // Structural comparison: blackjack's deal length depends on the shuffled
    // deck (the dealer's draws vary per game), so exact play counts can
    // never match across recordings -- see expectTraceMatchesGolden.
    expectTraceMatchesGolden(trace, 'blackjack-deal',
      { structural: { requiredKinds: ['boardgame-card'] } });
  });

  test('pig: die roll', async ({ page }) => {
    test.setTimeout(PARITY_TIMEOUT_MS);
    await createOfflineGame(page, 'pig');
    // The die mounts disabled until the initial state install and legality
    // preview land; wait for interactivity before opening the trace window.
    await expect(page.getByRole('button', { name: 'Roll die' }))
      .toBeEnabled({ timeout: 30000 });
    const trace = await captureTrace(page, async () => {
      // pig has no "Roll" button -- rolling is driven by clicking the die
      // itself (boardgame-die), whose accessible name is "Roll die" while
      // interactive (see server/static/src/components/boardgame-die.ts).
      // A roll landing on the SAME face animates nothing (~1-in-6), which
      // would fail the required-kinds contract below -- re-roll (bounded)
      // until the die visibly animates. P(4 same-face rolls) ~ 0.08%.
      for (let attempt = 0; attempt < 4; attempt++) {
        const logStart: number = await page.evaluate(
          () => (window as any).__bgAnimTestHooks.log.length);
        const before = await gateSnapshot(page);
        const die = page.getByRole('button', { name: 'Roll die' });
        await expect(die).toBeEnabled({ timeout: 20000 });
        await die.click();
        // Deterministic capture-window close (see debuganimations note).
        await expectCleanGate(page, before);
        const dieAnimated = await page.evaluate((start: number) => {
          const h = (window as any).__bgAnimTestHooks;
          return h.log.slice(start)
            .some((e: any) => e.ev === 'play' && String(e.detail).startsWith('boardgame-die'));
        }, logStart).catch(() => false);
        if (dieAnimated) break;
      }
    });
    // Structural: pig's post-roll cycle count branches on the rolled value
    // (a 1 busts the turn; 2-6 score), which the test cannot control. The
    // die itself must animate in at least one captured roll.
    expectTraceMatchesGolden(trace, 'pig-roll',
      { structural: { requiredKinds: ['boardgame-die'] } });
  });
});
