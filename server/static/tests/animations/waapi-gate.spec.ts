import { test, expect } from '@playwright/test';
import { createOfflineGame, gateSnapshot, expectCleanGate, installedGameVersion, joinCompanionAsGuest, waitForClientQuiescence } from './helpers';

// The reliability gate (spec Testing section). These scenarios are the
// historical wedge repros for #720. They must run clean N times.
//
// This suite is written against the CURRENT (pre-WAAPI-rewrite) animation
// system and documents its baseline reliability. It is expected that
// watchdogFirings assertions may fail or flake here -- see the commit
// message for observed baseline numbers. A later task rewrites the
// animation completion gate to fix the underlying wedge (#720); this
// suite is what that rewrite must pass cleanly.
const ROUNDS = 10;

test.describe('animation completion gate', () => {
  test('debuganimations: To Hidden / To Visible toggle never wedges', async ({ page }) => {
    test.setTimeout(120_000);

    await createOfflineGame(page, 'debuganimations');
    let animatedCycles = 0;

    const moveAndCheck = async (buttonName: 'To Hidden' | 'To Visible') => {
      const before = await gateSnapshot(page);
      const versionBefore = await installedGameVersion(page);
      const button = page.getByRole('button', { name: buttonName });
      await expect(button).toBeEnabled();
      // The intentionally messy debug stack can visually overlap these
      // controls. Pointer hit-testing is not under test here; activate the
      // enabled control's native button contract so random card placement
      // cannot turn this animation lifecycle test into a layout lottery.
      await button.evaluate((element) => (element as HTMLButtonElement).click());
      await expect.poll(() => installedGameVersion(page), { timeout: 10_000 })
        .toBeGreaterThan(versionBefore);

      // Moving between a rendered and hidden stack can legitimately install
      // without a visual animation when there is no matching rendered item.
      // That is not a wedged gate. Give Lit two frames to emit will-animate;
      // if it does, require a complete clean cycle. In all cases the accepted
      // version must fully drain before the next move.
      await page.evaluate(() => new Promise<void>((resolve) => {
        requestAnimationFrame(() => requestAnimationFrame(() => resolve()));
      }));
      const afterInstall = await gateSnapshot(page);
      if (afterInstall.gateOpens > before.gateOpens) {
        animatedCycles++;
        await expectCleanGate(page, before);
      } else {
        await waitForClientQuiescence(page);
        const afterQuiescence = await gateSnapshot(page);
        expect(afterQuiescence.watchdogFirings).toBe(before.watchdogFirings);
      }
    };

    for (let i = 0; i < ROUNDS; i++) {
      // debuganimations exposes "To Hidden" / "To Visible" buttons that
      // toggle moveStartMoveAllComponentsToHidden / ...ToVisible -- moving
      // every component in AllVisibleStack/AllHiddenStack in one Apply().
      // This all-at-once bulk move is the canonical #720 wedge repro (the
      // brief's guessed "Move All Components"/"Undo" labels don't exist in
      // the real renderer; these are the real button labels).
      //
      // A fresh snapshot is taken immediately before each click (not once
      // outside the loop): expectCleanGate needs to observe the gate
      // opening *after* the snapshot it's given, and a click's move
      // proposal -> server round-trip -> gate-open is asynchronous, so a
      // snapshot from several moves ago would let it race ahead.
      await moveAndCheck('To Hidden');
      await moveAndCheck('To Visible');
    }
    expect(animatedCycles, 'bulk moves must exercise at least one real animation gate cycle').toBeGreaterThan(0);
  });

  test('blackjack: player hit completes cleanly', async ({ browser }) => {
    // Three complete rooms, each with the two real guest seats Blackjack
    // requires before it can leave Gathering. This deliberately measures a
    // player-facing Hit; a newly-created room alone has no deal animation and
    // therefore cannot provide positive gate evidence.
    //
    // Budget: one full room (create + sign-in + two guest join flows + deal
    // + hit + clean-gate wait) measures ~80s on a loaded dev laptop — a
    // single guest join alone is ~40s of avatar/seat round-trips. 120s
    // could not cover three rooms on such hardware and timed out mid-
    // iteration 2 with the room genuinely healthy.
    test.setTimeout(360_000);

    for (let i = 0; i < 3; i++) {
      const controllerContext = await browser.newContext();
      const firstPlayerContext = await browser.newContext();
      const secondPlayerContext = await browser.newContext();
      try {
        const controller = await controllerContext.newPage();
        const firstPlayer = await firstPlayerContext.newPage();
        const secondPlayer = await secondPlayerContext.newPage();
        await createOfflineGame(controller, 'blackjack', { companionMode: true, adminMode: false });
        const roomCode = (await controller.locator('.room-code-giant').textContent())?.trim();
        expect(roomCode).toMatch(/^[A-Z]{4,5}$/);
        await joinCompanionAsGuest(firstPlayer, roomCode!, 'blackjack');
        await joinCompanionAsGuest(secondPlayer, roomCode!, 'blackjack');

        const hit = firstPlayer.getByRole('button', { name: 'Hit', exact: true });
        await expect(hit).toBeEnabled({ timeout: 20_000 });
        // Hit becomes legal just before Blackjack's final automatic reveal
        // bundle finishes. Wait for that initial deal pipeline to drain so
        // game-view does not correctly swallow our measured click as a
        // mid-animation gesture (#721).
        await waitForClientQuiescence(firstPlayer, 20_000);
        const before = await gateSnapshot(firstPlayer);
        await hit.click();
        await expectCleanGate(firstPlayer, before, 30_000);
      } finally {
        await Promise.all([
          controllerContext.close(),
          firstPlayerContext.close(),
          secondPlayerContext.close(),
        ]);
      }
    }
  });

  test('memory: interrupted cycles at game creation close every gate they open', async ({ page }) => {
    test.setTimeout(60_000);

    // Game creation installs several state bundles in rapid succession (the
    // initial deal), and the motion-release / legacy-overlap cutover admits
    // each successor while the previous animation cycle's gate is still
    // open. An interrupted cycle must still complete its lifecycle: if its
    // gate-open is never matched by a gate-close, the completion accounting
    // wedges permanently (gateCloses lags gateOpens forever) even though the
    // board eventually looks settled. Regression coverage for the
    // _stateChanged interrupted-cycle close in boardgame-render-game.
    await createOfflineGame(page, 'memory');
    await expect(page.locator('boardgame-card').first()).toBeAttached({ timeout: 15000 });

    await waitForClientQuiescence(page);
    const snapshot = await gateSnapshot(page);
    expect(snapshot.gateOpens, 'the creation deal must open at least one gate').toBeGreaterThan(0);
    expect(snapshot.gateCloses, 'every gate-open (including interrupted cycles) must be matched by a close')
      .toBe(snapshot.gateOpens);
    expect(snapshot.watchdogFirings, 'animation watchdog must never fire').toBe(0);
  });

  test('memory: same-cycle state reinstall mid-gate must not close the gate early', async ({ page }) => {
    test.setTimeout(60_000);

    // Guard scope regression (code review of the interrupted-cycle close):
    // state also installs WITHOUT a new motion cycle -- doGameInfo refreshes
    // (refresh-data / requested-player / admin-mode changes) reinstall a
    // fresh state object while motionCycleId is unchanged. If such an
    // install closed the still-open gate, the close would carry the
    // STILL-CURRENT cycleId; game-view's _forwardCycleRelease would forward
    // it and release a queued successor bundle early, cutting the in-flight
    // animation short. This drives exactly that shape at the component
    // contract: a new state object, same motionCycleId, landing mid-gate.
    await createOfflineGame(page, 'memory');
    await expect(page.locator('boardgame-card').first()).toBeAttached({ timeout: 15000 });
    await waitForClientQuiescence(page);

    // Open a real gate cycle with a card reveal, then perform the whole
    // wait-for-open + reinstall + sample sequence inside ONE in-page
    // evaluation so no CDP round-trip can race the ~500ms animation window.
    await page.locator('boardgame-card:not([disabled])').first().click();
    const result = await page.locator('boardgame-render-game').first().evaluate(async (el: any) => {
      const hooks = (window as any).__bgAnimTestHooks;
      const opened = await new Promise<boolean>((resolve) => {
        const deadline = performance.now() + 10_000;
        const poll = () => {
          if (el.isAnimating) return resolve(true);
          if (performance.now() > deadline) return resolve(false);
          setTimeout(poll, 10);
        };
        poll();
      });
      if (!opened) return { opened, closedEarly: false };
      const closesBefore = hooks.gateCloses;
      // The doGameInfo shape: fresh state object identity, unchanged cycle.
      el.state = { ...el.state };
      await el.updateComplete;
      return { opened, closedEarly: hooks.gateCloses !== closesBefore };
    });
    expect(result.opened, 'the card reveal must open an observable gate cycle').toBe(true);
    expect(result.closedEarly,
      'a same-cycle reinstall must not close the open gate (it would release the successor early)')
      .toBe(false);

    // The reinstall itself resets the gate for the same cycle; it must still
    // settle on its own (and never via the watchdog). Deliberately no
    // closes==opens assertion here: a same-cycle reset re-opens the gate
    // without a bundle handoff, which the cumulative-equality invariant
    // (covered by the creation-interrupt test above) does not model.
    const renderGame = page.locator('boardgame-render-game').first();
    await expect.poll(
      () => renderGame.evaluate((element) => (element as any).isAnimating === false),
      { timeout: 20_000 },
    ).toBe(true);
    const after = await gateSnapshot(page);
    expect(after.watchdogFirings, 'animation watchdog must never fire').toBe(0);
  });

  test('memory: card reveal completes cleanly', async ({ page }) => {
    test.setTimeout(60_000);

    await createOfflineGame(page, 'memory');
    // createOfflineGame only waits for <boardgame-render-game> to mount, not
    // for the memory-specific card grid inside its shadow tree to finish
    // its own async render -- wait for at least one card to actually exist.
    await expect(page.locator('boardgame-card').first()).toBeAttached({ timeout: 15000 });

    // Drive the actual card surface. The offline creator is auto-seated in a
    // fresh Memory game, and the renderer supplies CardIndex through the
    // card's data attribute. Unlike the admin move form, this proposal stays
    // on the document whose animation hooks we sampled.
    const before = await gateSnapshot(page);
    await page.locator('boardgame-card:not([disabled])').first().click();
    await expectCleanGate(page, before, 30000);
  });
});
