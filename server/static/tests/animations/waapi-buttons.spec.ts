import { test, expect } from '@playwright/test';
import { createOfflineGame, gateSnapshot, settleInitialLoad } from './helpers';

// Covers #721: boardgame-render-game reflects isAnimating (is-animating
// attribute) while an animation cycle is in flight, and
// boardgame-game-view swallows propose-move events that arrive while it is
// true -- so a move button double-clicked during the default animation
// window enqueues only one move instead of two (a move judged against
// stale on-screen state).

// deepQueryFirst walks into every shadowRoot (mirrors the deepQueryAll
// helper used by waapi-attrs.spec.ts) to find the first match for
// `selector`, since boardgame-render-game lives inside
// boardgame-game-view's shadow root.
function deepQueryFirstScript() {
  function deepQueryFirst(root: Document | ShadowRoot | Element, selector: string): Element | null {
    const direct = root.querySelector(selector);
    if (direct) return direct;
    for (const el of Array.from(root.querySelectorAll('*'))) {
      if ((el as any).shadowRoot) {
        const found = deepQueryFirst((el as any).shadowRoot, selector);
        if (found) return found;
      }
    }
    return null;
  }
  return deepQueryFirst;
}

test('move buttons disable during animation and re-enable after', async ({ page }) => {
  await createOfflineGame(page, 'debuganimations');
  // The initial-load bundles animate too (gated player-info/roster), so
  // is-animating is legitimately true for a few seconds after the page
  // mounts. The "closed before any click" baseline below is only meaningful
  // once that initial cascade has settled.
  await settleInitialLoad(page);

  const isAnimatingAttr = async () => page.evaluate((fnSrc: string) => {
    // eslint-disable-next-line no-eval
    const deepQueryFirst = eval(`(${fnSrc})`);
    const rg = deepQueryFirst(document, 'boardgame-render-game');
    return rg ? rg.hasAttribute('is-animating') : null;
  }, `(${deepQueryFirstScript.toString()})()`);

  // Before clicking anything, the gate should be closed (no animation in flight).
  expect(await isAnimatingAttr()).toBe(false);

  await page.getByRole('button', { name: 'To Hidden' }).click();

  // Immediately after the click, the render-game should reflect is-animating.
  // The click -> propose-move -> server round-trip -> gate-open is
  // asynchronous, so poll rather than sampling exactly once. Use the same
  // generous timeout as the gate-open wait elsewhere in this suite (see
  // helpers.ts's expectCleanGate) -- this round-trip can take noticeably
  // longer than a few ms under load.
  await expect.poll(isAnimatingAttr, { timeout: 20000 }).toBe(true);

  // Wait for the gate to fully close (animation cycle complete).
  await page.waitForFunction(() => {
    const h = (window as any).__bgAnimTestHooks;
    return h.gateCloses >= h.gateOpens;
  }, undefined, { timeout: 20000 });

  await expect.poll(isAnimatingAttr, { timeout: 20000 }).toBe(false);
});

test('a move proposed while isAnimating is true is swallowed, not enqueued', async ({ page }) => {
  await createOfflineGame(page, 'debuganimations');
  // Settle the initial-load cascade first: otherwise the `gateOpens >
  // before.gateOpens` wait below is satisfied by an *initial-load* gate
  // cycle (opened between the snapshot and the click's server round-trip)
  // rather than the clicked move's own cycle, and the mid-animation
  // dispatch lands after that spurious cycle already closed
  // (isAnimating=false precondition failure).
  await settleInitialLoad(page);

  // This test isolates the game-view guard itself (boardgame-game-view's
  // propose-move listener returning early while
  // this._renderEle.isAnimating) from the surrounding real-move-button UI.
  // Driving it via two real rapid clicks on debuganimations' buttons is
  // unreliable as an end-to-end test: a click's move can flip
  // isMoveCurrentlyLegal (server-driven, via the moveLegality prop) well
  // before the *animation* actually completes, which disables the button
  // and blocks the second click for a reason unrelated to the #721 guard
  // being tested here -- confirmed while developing this test (verified via
  // direct propose-move-count instrumentation: with real clicks the second
  // click's event sometimes never dispatches at all, purely because of
  // legality-driven disabling, not the animation guard). Dispatching a
  // synthetic propose-move event directly is exactly what
  // boardgame-base-game-renderer's click handler does under the hood (see
  // _handleButtonTapped in boardgame-base-game-renderer.ts), so this
  // exercises the real consumer-facing contract without that race.
  const before = await gateSnapshot(page);

  await page.getByRole('button', { name: 'To Hidden' }).click();

  // Wait for the gate to actually open (async: click -> propose-move ->
  // server round-trip -> gate-open) before firing the synthetic second
  // proposal, so it lands while isAnimating is genuinely true.
  await page.waitForFunction((sinceGateOpens: number) => {
    const h = (window as any).__bgAnimTestHooks;
    return h.gateOpens > sinceGateOpens;
  }, before.gateOpens, { timeout: 20000 });

  const warnings: string[] = [];
  page.on('console', (msg) => {
    if (msg.type() === 'warning') warnings.push(msg.text());
  });

  const midAnimationResult = await page.evaluate((fnSrc: string) => {
    // eslint-disable-next-line no-eval
    const deepQueryFirst = eval(`(${fnSrc})`);
    const rg = deepQueryFirst(document, 'boardgame-render-game') as any;
    const isAnimating = rg ? rg.isAnimating : null;
    const h = (window as any).__bgAnimTestHooks;
    const gateOpensBefore = h.gateOpens;
    // Dispatch the exact same event shape the real button-click handler
    // sends (see boardgame-base-game-renderer.ts's _handleButtonTapped),
    // directly on the render-game element -- bubbles+composed carries it up
    // through boardgame-game-view's shadow-DOM-crossing listener exactly as
    // a real in-game button click would (boardgame-game-view itself lives
    // inside boardgame-app's shadow root, so document.querySelector can't
    // find it directly).
    rg.dispatchEvent(new CustomEvent('propose-move', {
      bubbles: true,
      composed: true,
      detail: { name: 'Start Move All Components To Visible', arguments: {} }
    }));
    return { isAnimating, gateOpensAtDispatch: gateOpensBefore };
  }, `(${deepQueryFirstScript.toString()})()`);

  expect(midAnimationResult.isAnimating, 'precondition: gate must be open when the synthetic move is dispatched').toBe(true);

  // Give any (incorrectly) accepted move a moment to reach the server and
  // open a second gate cycle before asserting it didn't.
  await page.waitForTimeout(500);
  const afterSwallowAttempt = await gateSnapshot(page);
  expect(afterSwallowAttempt.gateOpens, 'swallowed move must not open a second animation gate cycle')
    .toBe(midAnimationResult.gateOpensAtDispatch);
  expect(warnings.some((w) => w.includes('propose-move ignored while animations are running'))).toBe(true);

  // Let the original (first) animation cycle settle cleanly so this test
  // doesn't leak an in-flight animation into the next one.
  await page.waitForFunction(() => {
    const h = (window as any).__bgAnimTestHooks;
    return h.gateCloses >= h.gateOpens;
  }, undefined, { timeout: 20000 });
  // The diagnostic close counter increments at the start of the close
  // notification. Wait for the reflected renderer state to consume it before
  // testing a post-gate dispatch; otherwise this assertion races Lit's update.
  await page.waitForFunction((fnSrc: string) => {
    // eslint-disable-next-line no-eval
    const deepQueryFirst = eval(`(${fnSrc})`);
    const rg = deepQueryFirst(document, 'boardgame-render-game') as any;
    return rg ? !rg.isAnimating : false;
  }, `(${deepQueryFirstScript.toString()})()`, { timeout: 20000 });
  const settled = await gateSnapshot(page);
  expect(settled.watchdogFirings, 'animation watchdog must never fire').toBe(before.watchdogFirings);

  // Once the gate has closed, the same synthetic move IS accepted (proves
  // the guard is scoped to isAnimating, not a permanent block).
  const afterCloseResult = await page.evaluate((fnSrc: string) => {
    // eslint-disable-next-line no-eval
    const deepQueryFirst = eval(`(${fnSrc})`);
    const rg = deepQueryFirst(document, 'boardgame-render-game') as any;
    const h = (window as any).__bgAnimTestHooks;
    const gateOpensBefore = h.gateOpens;
    const isAnimatingBefore = rg ? rg.isAnimating : null;
    rg.dispatchEvent(new CustomEvent('propose-move', {
      bubbles: true,
      composed: true,
      detail: { name: 'Start Move All Components To Visible', arguments: {} }
    }));
    return { isAnimating: isAnimatingBefore, gateOpensAtDispatch: gateOpensBefore };
  }, `(${deepQueryFirstScript.toString()})()`);
  expect(afterCloseResult.isAnimating, 'precondition: gate must be closed for this second dispatch').toBe(false);

  await page.waitForFunction((sinceGateOpens: number) => {
    const h = (window as any).__bgAnimTestHooks;
    return h.gateOpens > sinceGateOpens;
  }, afterCloseResult.gateOpensAtDispatch, { timeout: 20000 });
});
