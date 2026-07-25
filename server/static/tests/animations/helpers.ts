import { Page, expect } from '@playwright/test';

// Display names shown in the "Game Type" combobox on /list-games, keyed by
// the internal game name used in URLs (/game/<name>/<id>/).
const GAME_TYPE_LABELS: Record<string, string> = {
  debuganimations: 'Animations Debugger',
  blackjack: 'Blackjack',
  memory: 'Memory',
  // pig's gameDelegate doesn't override DisplayName(), so base.GameDelegate's
  // default (title-case of Name()) applies -- see base/game_delegate.go.
  pig: 'Pig',
};

const FAKE_EMAIL = 'animtest@example.com';
const FAKE_PASSWORD = 'animtest-password';

// Creates a fresh offline-dev-mode game and lands on its game page, signed
// in and with Admin Mode enabled (view-as Admin, "Make Moves As
// ViewingAsPlayer" unchecked) so moves can be proposed deterministically as
// AdminPlayerIndex. Mirrors the real UI flow (there is no working reference
// flow in verify-fix.spec.ts to transcribe -- see task brief correction):
//   /list-games -> pick game type card -> Create Game -> (if not yet signed
//   in) Email/Password -> fill fake credentials -> Sign In.
export async function createOfflineGame(
  page: Page,
  gameName: string,
  options: { companionMode?: boolean; adminMode?: boolean } = {},
): Promise<void> {
  const label = GAME_TYPE_LABELS[gameName];
  if (!label) {
    throw new Error(`Unknown game type "${gameName}"; add it to GAME_TYPE_LABELS in helpers.ts`);
  }

  await page.goto('/list-games');

  await page.getByRole('combobox', { name: 'Game Type', exact: true }).click();
  await page.getByRole('option', { name: new RegExp(`^${label}`) }).click();

  if (options.companionMode) {
    const companionMode = page.locator('md-switch[name="companionMode"]');
    await expect(companionMode).toBeVisible();
    if (!await companionMode.evaluate((el) => (el as any).selected === true)) {
      await companionMode.click();
    }
  }

  await page.getByRole('button', { name: 'Create Game' }).click();

  // If this is a fresh (unauthenticated) browser context, offline-dev-mode
  // shows a fake-login dialog. Sign in with Email/Password and a fake
  // account; this is idempotent across repeated calls within the same test
  // (same fake credentials -> same fake account).
  await signInOffline(page);

  // Now on /game/<gameName>/<id>/. Wait for the game renderer to mount.
  await page.waitForURL(new RegExp(`/game/${gameName}/`));
  await page.waitForSelector('boardgame-render-game', { timeout: 15000 });
  await page.waitForFunction(() => (window as any).__bgAnimTestHooks !== undefined, undefined, { timeout: 15000 });

  if (options.adminMode === false) return;

  // The game creator is not automatically seated as an *active* player in
  // the roster UI sense (both slots render "Sitting out" even though the
  // server has auto-seated the creator -- offline-dev-mode's
  // DisableAdminChecking auto-seat, server/api/main.go ~line 586), and
  // proposing a move as an unseated observer is rejected server-side ("The
  // proposer was not valid."). Rather than fight the join-dialog's
  // shadow-DOM Material buttons (which don't respond to scripted .click()
  // -- only real pointer events reliably trigger their form-submitter
  // logic), use the built-in admin panel: Admin Mode -> View as Admin ->
  // uncheck "Make Moves As ViewingAsPlayer" makes every subsequent move
  // proposal use AdminPlayerIndex, which game.go's applyMove() always
  // accepts for moves.Default-based moves. Player-scoped moves should be
  // exercised through a real seated-player surface instead.
  await enableAdminMode(page);
  await page.getByText('Admin', { exact: true }).click();

  // Uncheck "Make Moves As ViewingAsPlayer" (id="move-as-player") so
  // subsequent move proposals use AdminPlayerIndex instead of player 0.
  // md-checkbox's `checked` also does not reflect to an attribute (same
  // issue as md-switch's `selected`, see enableAdminMode) -- read the
  // live property.
  const makeMovesCheckbox = page.locator('md-checkbox#move-as-player');
  await expect(makeMovesCheckbox).toBeVisible();
  const isChecked = await makeMovesCheckbox.evaluate((el) => (el as any).checked === true);
  if (isChecked) {
    await makeMovesCheckbox.click();
  }
}

// Authenticates the current page as the stable offline-dev test identity.
// Useful when a test needs to join an existing game without creating a
// throwaway game (and without enabling the admin-only controls).
export async function signInOffline(page: Page): Promise<void> {
  const emailPasswordButton = page.getByRole('button', { name: 'Email/Password' });
  if (await emailPasswordButton.isVisible().catch(() => false)) {
    await emailPasswordButton.click();
    await page.getByRole('textbox', { name: 'Email' }).fill(FAKE_EMAIL);
    await page.getByRole('textbox', { name: 'Password' }).fill(FAKE_PASSWORD);
    await page.getByRole('button', { name: 'Sign In' }).click();
  }
}

// Joins an existing companion-mode room through the same guest flow a phone
// uses after scanning the Table QR code. Symmetric games auto-assign a seat;
// asymmetric games expose a seat grid, where this helper chooses the first
// open slot.
export async function joinCompanionAsGuest(
  page: Page,
  roomCode: string,
  gameName: string,
): Promise<void> {
  await page.goto(`/join?code=${encodeURIComponent(roomCode)}`);
  await page.getByRole('button', { name: 'Use a new guest identity' }).click();
  await page.getByRole('button', { name: 'Looks good — join!' }).click();
  const openSeat = page.locator('.slot:not(.filled)').first();
  if (await openSeat.isVisible({ timeout: 1000 }).catch(() => false)) {
    await openSeat.click();
  }
  await page.waitForURL(new RegExp(`/game/${gameName}/`), { timeout: 20000 });
  await page.waitForSelector('boardgame-render-game', { timeout: 15000 });
}

// Turns on the header's "Admin Mode" switch, revealing the admin debug
// panel (View as / Moves / State / Chest) at the bottom of the game page.
// Idempotent -- safe to call even if already enabled.
export async function enableAdminMode(page: Page): Promise<void> {
  const adminSwitch = page.locator('.admin-toggle md-switch');
  if (await adminSwitch.count() === 0) {
    throw new Error('Admin Mode switch not found; offline-dev-mode may not be enabled on the server');
  }
  // md-switch's `selected` is a plain Lit @property that does NOT reflect
  // to an attribute (@material/web/switch/internal/switch.js: "The
  // selected property does not reflect..."), so `getAttribute('selected')`
  // is always null -- checking it would re-click (and thus toggle back
  // off) an already-enabled switch. Read the live JS property instead.
  const isSelected = await adminSwitch.evaluate((el) => (el as any).selected === true);
  if (!isSelected) {
    // This is harness setup, not a switch interaction test. The fixed game
    // surface can cover the off-canvas drawer at some viewport/scroll
    // positions, so drive the switch's public property/change contract
    // directly and deterministically.
    await adminSwitch.evaluate((el) => {
      (el as any).selected = true;
      el.dispatchEvent(new Event('change', { bubbles: true, composed: true }));
    });
    await expect.poll(() => adminSwitch.evaluate((el) => (el as any).selected === true)).toBe(true);
  }
}

// Waits until the animHooks counters have been unchanged for `stableMs`
// (and optionally balanced). Point-in-time quiescence checks race the
// trailing edge -- a late fix-up bundle or next-frame settle can land right
// after they pass -- so callers that assert on cumulative counter equality
// must sample only after sustained stability.
export async function waitForAnimationCounterStability(
  page: Page,
  opts: { stableMs?: number; timeoutMs?: number; balance?: 'plays' | 'all' | 'none' } = {},
): Promise<void> {
  const { stableMs = 1500, timeoutMs = 30000, balance = 'plays' } = opts;
  await page.waitForFunction(([stable, bal]) => {
    const h = (window as any).__bgAnimTestHooks;
    if (!h) return false;
    const w = (window as any).__parityStability ??= { last: '', since: 0 };
    const now = performance.now();
    const key = `${h.gateOpens}|${h.gateCloses}|${h.plays}|${h.settles}`;
    if (key !== w.last) { w.last = key; w.since = now; return false; }
    if (bal === 'plays' && h.plays !== h.settles) return false;
    if (bal === 'all' && (h.plays !== h.settles || h.gateOpens !== h.gateCloses)) return false;
    return (now - w.since) >= (stable as number);
  }, [stableMs, balance] as [number, string], { timeout: timeoutMs, polling: 100 });
}

export interface GateSnapshot {
  gateOpens: number;
  gateCloses: number;
  watchdogFirings: number;
  plays: number;
  settles: number;
}

export async function waitForClientQuiescence(page: Page, timeoutMs = 20000): Promise<void> {
  await page.waitForFunction(async () => {
    const hooks = (window as any).__bgAnimTestHooks;
    if (!hooks || hooks.gateCloses < hooks.gateOpens) return false;
    // Read the APP's store instance via the always-installed window handle
    // (src/store.ts). Re-importing '/src/store.ts' here would construct a
    // second, permanently-empty store whenever the dev server's HMR graph
    // has rewritten the app's import to /src/store.ts?t=<timestamp>.
    const store = (window as any).__bgReduxStore
      ?? (await import('/src/store.ts')).store;
    return (store.getState().game?.animation?.pendingBundles?.length ?? 0) === 0;
  }, undefined, { timeout: timeoutMs });
  const renderGame = page.locator('boardgame-render-game').first();
  await expect.poll(
    () => renderGame.evaluate((element) => (element as any).isAnimating === false),
    { timeout: timeoutMs },
  ).toBe(true);
  // Let listeners reacting to the final all-animations-done/dequeue events
  // finish their Lit update before the next user gesture.
  await page.evaluate(() => new Promise<void>((resolve) => {
    requestAnimationFrame(() => requestAnimationFrame(() => resolve()));
  }));
}

export async function installedGameVersion(page: Page): Promise<number> {
  return page.evaluate(async () => {
    // See waitForClientQuiescence for why the window handle is required.
    const store = (window as any).__bgReduxStore
      ?? (await import('/src/store.ts')).store;
    return store.getState().game?.animation?.lastFiredBundle?.game?.Version ?? -1;
  });
}

export async function gateSnapshot(page: Page): Promise<GateSnapshot> {
  for (let attempt = 0; attempt < 3; attempt++) {
    await page.waitForFunction(() => (window as any).__bgAnimTestHooks !== undefined);
    try {
      return await page.evaluate(() => {
        // Stamp a reload sentinel so a later navigation cannot masquerade as
        // successful gate completion with freshly reset counters.
        (window as any).__gateProbeToken = ((window as any).__gateProbeToken || 0) + 1;
        const h = (window as any).__bgAnimTestHooks;
        return {
          gateOpens: h.gateOpens, gateCloses: h.gateCloses,
          watchdogFirings: h.watchdogFirings, plays: h.plays, settles: h.settles,
        };
      });
    } catch (error) {
      const navigated = error instanceof Error && error.message.includes('Execution context was destroyed');
      if (!navigated || attempt === 2) throw error;
      await page.waitForLoadState('domcontentloaded');
    }
  }
  throw new Error('animation hooks did not stabilize');
}

// Waits for the animation gate to be quiescent (closes caught up with
// opens) then asserts no watchdog fired since `since`.
//
// IMPORTANT: this must first wait for the gate to actually *open* past
// `since.gateOpens` before checking that closes have caught up. A move's
// proposal -> server round-trip -> Redux update -> gate-open is
// asynchronous and can take noticeably longer than a few ms; checking
// `gateCloses >= gateOpens` immediately after a click, before that
// round-trip lands, is trivially (and wrongly) true because neither
// counter has moved yet. Without this, a slow-to-open gate is
// indistinguishable from an already-settled one, and the true opening
// (and its close) get silently missed by the caller's next assertion.
// allowAlreadySettled is for scenarios where the animation that opens the
// gate fires on its own (e.g. blackjack's fresh deal, which auto-runs the
// moment the game is created) rather than in response to an action this
// helper's caller just performed. Setup after game creation
// (createOfflineGame's admin-panel dance) can take long enough that the
// entire deal -- many rapid open/close cycles -- completes before
// gateSnapshot is even taken, so there is no fresh open past `since` left
// to observe. In that case a quiescent gate (closes caught up to opens)
// with an unchanged watchdog count IS the clean outcome; requiring a new
// open would hang until timeout. The click-driven scenarios
// (debuganimations) leave this false so a slow-to-open gate can't be
// mistaken for an already-settled one.
export async function expectCleanGate(
  page: Page,
  since: GateSnapshot,
  timeoutMs = 20000,
  opts: { allowAlreadySettled?: boolean } = {}
) {
  const allowAlreadySettled = opts.allowAlreadySettled ?? false;
  // Wait for the hooks before sampling, but never treat their reset after a
  // reload as successful completion: a reload destroys the evidence this
  // assertion exists to verify.
  await page.waitForFunction(() => (window as any).__bgAnimTestHooks !== undefined, undefined, { timeout: timeoutMs });
  const openedOrReset = await page.waitForFunction(([sinceGateOpens, allowSettled]) => {
    if ((window as any).__gateProbeToken === undefined) return 'reset';
    const h = (window as any).__bgAnimTestHooks;
    if (h.gateOpens < sinceGateOpens) return 'reset';
    if (h.gateOpens > sinceGateOpens) return 'opened';
    // No fresh open past `since`. For auto-firing scenarios, a gate that is
    // already quiescent (closes caught up) is the clean, already-completed
    // outcome -- accept it rather than hanging on an open that won't come.
    if (allowSettled && h.gateOpens > 0 && h.gateCloses >= h.gateOpens) return 'settled';
    return false;
  }, [since.gateOpens, allowAlreadySettled] as [number, boolean], { timeout: timeoutMs }).then((h) => h.jsonValue());
  await page.waitForFunction(() => {
    const h = (window as any).__bgAnimTestHooks;
    return h.gateCloses >= h.gateOpens;
  }, undefined, { timeout: timeoutMs });
  if (openedOrReset === 'reset') {
    throw new Error('page reloaded while waiting for animation completion; gate evidence was lost');
  }
  // A gate can close while a later server bundle is still queued. Returning
  // before that queue drains lets a caller's next click race the next
  // animation and be intentionally swallowed by game-view (#721). "Clean"
  // therefore means the whole client pipeline is quiescent, not merely that
  // one gate-close event has been recorded.
  await waitForClientQuiescence(page, timeoutMs);
  const now = await gateSnapshot(page);
  // This definitive sample comes after every queued bundle drains: a later
  // bundle cannot wedge, be watchdog-closed, and disappear behind an earlier
  // clean gate assertion.
  expect(now.watchdogFirings, 'animation watchdog must never fire').toBe(since.watchdogFirings);
  if (allowAlreadySettled) {
    expect(now.gateOpens, 'an auto-fired animation must open the gate at least once').toBeGreaterThan(0);
    if (now.plays > 0) {
      expect(now.settles, 'every instrumented play must settle').toBeGreaterThanOrEqual(now.plays);
    }
  }
}
