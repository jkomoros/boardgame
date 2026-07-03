import { Page, expect, test } from '@playwright/test';

// Display names shown in the "Game Type" combobox on /list-games, keyed by
// the internal game name used in URLs (/game/<name>/<id>/).
const GAME_TYPE_LABELS: Record<string, string> = {
  debuganimations: 'Animations Debugger',
  blackjack: 'Blackjack',
  memory: 'Memory',
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
export async function createOfflineGame(page: Page, gameName: string): Promise<void> {
  const label = GAME_TYPE_LABELS[gameName];
  if (!label) {
    throw new Error(`Unknown game type "${gameName}"; add it to GAME_TYPE_LABELS in helpers.ts`);
  }

  await page.goto('/list-games');

  await page.getByRole('combobox', { name: 'Game Type', exact: true }).click();
  await page.getByRole('option', { name: new RegExp(`^${label}`) }).click();

  await page.getByRole('button', { name: 'Create Game' }).click();

  // If this is a fresh (unauthenticated) browser context, offline-dev-mode
  // shows a fake-login dialog. Sign in with Email/Password and a fake
  // account; this is idempotent across repeated calls within the same test
  // (same fake credentials -> same fake account).
  const emailPasswordButton = page.getByRole('button', { name: 'Email/Password' });
  if (await emailPasswordButton.isVisible().catch(() => false)) {
    await emailPasswordButton.click();
    await page.getByRole('textbox', { name: 'Email' }).fill(FAKE_EMAIL);
    await page.getByRole('textbox', { name: 'Password' }).fill(FAKE_PASSWORD);
    await page.getByRole('button', { name: 'Sign In' }).click();
  }

  // Now on /game/<gameName>/<id>/. Wait for the game renderer to mount.
  await page.waitForURL(new RegExp(`/game/${gameName}/`));
  await page.waitForSelector('boardgame-render-game', { timeout: 15000 });
  await page.waitForFunction(() => (window as any).__bgAnimTestHooks !== undefined, undefined, { timeout: 15000 });

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
  // accepts for moves.Default-based moves (moves.CurrentPlayer-based moves,
  // e.g. examples/memory's RevealCard, need a different approach -- see
  // makeAdminFormMove).
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
    await adminSwitch.click();
  }
}

// Submits a move via the admin panel's "Moves" section instead of clicking
// in-game UI as a specific seated player.
//
// Why this exists: examples/memory's RevealCard move embeds
// moves.CurrentPlayer (examples/memory/moves.go), which requires the
// proposer to equal state.CurrentPlayerIndex() exactly -- unlike
// moves.Default-based moves (e.g. debuganimations'), AdminPlayerIndex does
// NOT satisfy this as a wildcard proposer here; proposing as our own seat
// (player 0) fails whenever it isn't actually our turn, and proposing as
// AdminPlayerIndex fails moves.CurrentPlayer's TargetPlayerIndex.Valid()
// check outright. The admin panel's own move form defaults
// TargetPlayerIndex (and other fields, e.g. CardIndex) correctly for
// whichever player is actually current (server-side DefaultsForState), so
// submitting through it sidesteps needing to track turn order from the
// test.
//
// Returns false (without throwing) if the move was illegal -- e.g. no
// cards left, wrong phase -- since that's an expected "nothing left to do"
// end state for a test loop, not an infrastructure failure. Returns true
// if the move was accepted (HTTP 200 without a "Couldn't..." error
// dialog).
export async function makeAdminFormMove(page: Page, moveDisplayName: string): Promise<boolean> {
  // Admin Mode is a client-only UI toggle (not persisted server-side), and
  // some move round-trips in this app cause a full page reload that resets
  // it -- re-enable defensively so the "Moves" panel is actually present
  // before trying to use it.
  await enableAdminMode(page);

  const moveRow = page.getByText(moveDisplayName, { exact: true });
  await expect(moveRow).toBeVisible({ timeout: 5000 });
  await moveRow.click();
  const makeMoveButton = page.getByText('Make Move', { exact: true }).first();
  await expect(makeMoveButton).toBeVisible({ timeout: 5000 });
  await makeMoveButton.click();

  const errorDialog = page.locator('md-dialog[open]').filter({ hasText: "Couldn't" });
  const sawError = await errorDialog.first().isVisible({ timeout: 3000 }).catch(() => false);
  if (sawError) {
    await page.getByRole('button', { name: 'OK' }).click().catch(() => {});
    return false;
  }
  return true;
}

export interface GateSnapshot {
  gateOpens: number;
  gateCloses: number;
  watchdogFirings: number;
  plays: number;
  settles: number;
}

export async function gateSnapshot(page: Page): Promise<GateSnapshot> {
  return page.evaluate(() => {
    const h = (window as any).__bgAnimTestHooks;
    return {
      gateOpens: h.gateOpens, gateCloses: h.gateCloses,
      watchdogFirings: h.watchdogFirings, plays: h.plays, settles: h.settles,
    };
  });
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
export async function expectCleanGate(page: Page, since: GateSnapshot, timeoutMs = 20000) {
  // Defensive: some move round-trips in this app (observed on
  // examples/memory) trigger a page reload, which briefly makes
  // __bgAnimTestHooks undefined and would throw inside the waitForFunction
  // predicates below with a raw property access. Wait for it to exist
  // (again) first.
  await page.waitForFunction(() => (window as any).__bgAnimTestHooks !== undefined, undefined, { timeout: timeoutMs });
  // If a page reload happened, __bgAnimTestHooks was reinstalled fresh at
  // 0 -- `gateOpens` can now be *below* since.gateOpens (a counter reset,
  // not a wedge). Treat that as "already opened" (the reload itself proves
  // the state round-trip that would have opened the gate happened) rather
  // than waiting forever for a monotonic increase that will never come; a
  // reset also makes the watchdogFirings delta meaningless, so skip that
  // assertion for this call (a reset gate is definitionally not wedged).
  const openedOrReset = await page.waitForFunction((sinceGateOpens) => {
    const h = (window as any).__bgAnimTestHooks;
    if (h.gateOpens < sinceGateOpens) return 'reset';
    return h.gateOpens > sinceGateOpens ? 'opened' : false;
  }, since.gateOpens, { timeout: timeoutMs }).then((h) => h.jsonValue());
  await page.waitForFunction(() => {
    const h = (window as any).__bgAnimTestHooks;
    return h.gateCloses >= h.gateOpens;
  }, undefined, { timeout: timeoutMs });
  if (openedOrReset === 'reset') {
    console.warn('[expectCleanGate] page reloaded mid-check — watchdog assertion SKIPPED for this cycle');
    try {
      test.info().annotations.push({
        type: 'warning',
        description: 'expectCleanGate: reload detected, watchdog assertion skipped',
      });
    } catch {
      // Called outside a test context; annotation is optional
    }
    // Best-effort check: after reload, counters restarted at 0, so any
    // non-zero watchdogFirings indicates a post-reload wedge.
    const afterReload = await gateSnapshot(page);
    expect(afterReload.watchdogFirings, 'animation watchdog must never fire (post-reload check)').toBe(0);
    return;
  }
  const now = await gateSnapshot(page);
  expect(now.watchdogFirings, 'animation watchdog must never fire').toBe(since.watchdogFirings);
}
