import { test, expect } from '@playwright/test';
import { createOfflineGame, joinCompanionAsGuest } from './helpers';

// Companion-mode cross-screen animation sync (#798, spec §8.4). The
// Table surface (shared projector) and Hand surfaces (players' phones)
// must launch the same card flight at roughly the same wall-clock
// instant. Two mechanisms cooperate:
//
//   1. boardgame-game-state-manager installs each version early with its
//      server-anchored animation context, but only on companion surfaces.
//   2. boardgame-component-animator receives that version-bound start from
//      render-game and uses it as animateBetween's default, so semantic
//      Table/Hand/game code contains no transport timing plumbing.
//
// These specs drive the real production singletons (exposed on
// window.__bgCompanionSync, mirroring window.__bgAnimTestHooks) and the
// real animateBetween, then assert the cross-context skew is well under
// the 250ms coherence threshold (drift was ~1000ms before #798). Skew
// alone doesn't discriminate a correct sync from both surfaces ignoring
// the version context and firing immediately (they'd still agree),
// so the test also asserts each surface's absolute play instant lands
// near the shared target time — see the per-page offset assertions below.

test.describe('companion-sync estimator', () => {
  test('localEquivalent needs >=3 samples, then applies the min offset', async ({ page }) => {
    await createOfflineGame(page, 'blackjack'); // any page with the module loaded
    const r = await page.evaluate(() => {
      const sync = (window as any).__bgCompanionSync;
      if (!sync) return { missing: true } as any;
      const est = sync.estimator;
      const server = Date.now() + 100000; // arbitrary future server epoch
      // NOTE: this is the live production singleton, so real socket
      // version-timing frames from the running game may already have fed
      // it samples — we assert its CONTRACT, not a from-empty offset.
      const offsetBefore = est.minOffset();
      const beforeWarm = est.localEquivalent(server);
      // Feed >=3 samples with a large, unambiguous one-way latency (250ms)
      // so the min-wins estimator has a committed, non-null offset even if
      // it was cold before. ingest records (Date.now() - serverSentAt).
      const now = Date.now();
      for (let i = 0; i < 4; i++) {
        est.ingest({
          version: 1000 + i,
          serverSentAt: now - 250,
          serverPlayAt: now,
          slotDurationMs: 800,
          maxAnimationDurationMs: 600,
        });
      }
      const offsetAfter = est.minOffset();
      const afterWarm = est.localEquivalent(server);
      return { missing: false, offsetBefore, beforeWarm, offsetAfter, afterWarm, server };
    });
    expect(r.missing, 'window.__bgCompanionSync must be exposed').toBe(false);
    // Cold-estimator contract: with <3 samples minOffset is null and
    // localEquivalent is a passthrough (callers then play immediately).
    if (r.offsetBefore === null) {
      expect(r.beforeWarm).toBe(r.server);
    } else {
      // Already warm from live traffic: the conversion identity must hold.
      expect(r.beforeWarm).toBe(r.server + r.offsetBefore);
    }
    // Warm-estimator contract: >=3 samples ⇒ non-null offset, and
    // localEquivalent(x) === x + minOffset() exactly (the core server→
    // local conversion the schedulers depend on).
    expect(r.offsetAfter).not.toBeNull();
    expect(r.afterWarm).toBe(r.server + r.offsetAfter);
    // The min-wins offset can never exceed our injected 250ms sample
    // (variance only ever adds latency; the minimum is the floor).
    expect(r.offsetAfter).toBeLessThanOrEqual(250);
  });
});

test.describe('cross-screen synced auto-fly', () => {
  test.setTimeout(120000);

  test('common play policy covers composed-tree providers and the full remaining budget', async ({ page }) => {
    await createOfflineGame(page, 'blackjack');
    const result = await page.evaluate(() => {
      const provider = document.createElement('div') as any;
      const item = document.createElement('boardgame-animatable-item') as any;
      provider.appendChild(item);
      document.body.appendChild(provider);
      const now = Date.now();
      provider.animationContext = {
        version: 77,
        startAtMs: now + 300,
        slotDurationMs: 800,
        maxAnimationDurationMs: 600,
      };
      const inherited = item.play(item, [{ opacity: '0' }, { opacity: '1' }],
        { duration: 200, delay: 50 });
      const inheritedTiming = inherited.effect.getTiming();
      inherited.cancel();
      const local = item.play(item, [{ opacity: '0' }, { opacity: '1' }],
        { duration: 20 }, { timing: 'immediate' });
      const localTiming = local.effect.getTiming();
      local.cancel();

      delete provider.animationContext;
      item.animationContext = {
        version: 78,
        startAtMs: Date.now() - 500,
        slotDurationMs: 800,
        maxAnimationDurationMs: 600,
      };
      const skipped = item.play(item, [{ opacity: '0' }, { opacity: '1' }],
        { duration: 200, delay: 150 });

      item.animationContext = {
        version: 79,
        startAtMs: Date.now(),
        slotDurationMs: 800,
        maxAnimationDurationMs: 600,
      };
      item.postAnimationDelay = 1000;
      const held = item.play(item, [{ opacity: '0' }, { opacity: '1' }],
        { duration: 400 });
      const heldTiming = held.effect.getTiming();
      held.cancel();
      provider.remove();
      return {
        inheritedDelay: Number(inheritedTiming.delay),
        inheritedDuration: Number(inheritedTiming.duration),
        localDelay: Number(localTiming.delay),
        skipped: skipped === null,
        heldDuration: Number(heldTiming.duration),
        heldEndDelay: Number(heldTiming.endDelay),
      };
    });
    expect(result.inheritedDelay).toBeGreaterThan(250);
    expect(result.inheritedDelay).toBeLessThan(450);
    expect(result.inheritedDuration).toBe(200);
    expect(result.localDelay).toBe(0);
    expect(result.skipped).toBe(true);
    expect(result.heldDuration).toBe(0);
    expect(result.heldEndDelay).toBe(600);
  });

  // Two independent browser contexts stand in for the Table projector and
  // a player's phone. Each drives the REAL animateBetween with the SAME
  // absolute startAtMs (a shared wall-clock target on this one machine).
  // We then read each context's first 'play' hook — recorded by
  // animateBetween at the moment the flight VISUALLY begins — normalize it
  // to absolute epoch-comparable ms via performance.timeOrigin + t, and
  // assert the two launches land within the coherence threshold.
  test('deal flights start within sync threshold on both surfaces', async ({ browser }) => {
    const tableCtx = await browser.newContext();
    const handCtx = await browser.newContext();
    const table = await tableCtx.newPage();
    const hand = await handCtx.newPage();

    try {
      // Surface routing is the ?display=table|hand query param (see
      // utils/companion-surface.ts — it takes precedence over the
      // per-game surface_<gameId> cookie and is the dev-testing seam).
      await createOfflineGame(table, 'blackjack');
      const gameUrl = new URL(table.url());

      const tableUrl = new URL(gameUrl.toString());
      tableUrl.searchParams.set('display', 'table');
      const handUrl = new URL(gameUrl.toString());
      handUrl.searchParams.set('display', 'hand');

      // Join the same game on the hand context, then land both pages on
      // their surface-routed URLs. (The -table/-hand renderer variants are
      // not shipped for any game yet, so the loader falls back to the solo
      // renderer with a console warning; that is fine here — we are testing
      // the surface-agnostic animator + estimator, and the animator element
      // is created directly below, exactly as waapi-play.spec.ts does.)
      await createOfflineGame(hand, 'blackjack');
      await hand.goto(handUrl.toString());
      await table.goto(tableUrl.toString());

      await table.waitForFunction(() => (window as any).__bgAnimTestHooks !== undefined, undefined, { timeout: 30000 });
      await hand.waitForFunction(() => (window as any).__bgAnimTestHooks !== undefined, undefined, { timeout: 30000 });
      await table.waitForFunction(() => (window as any).__bgCompanionSync !== undefined, undefined, { timeout: 30000 });
      await hand.waitForFunction(() => (window as any).__bgCompanionSync !== undefined, undefined, { timeout: 30000 });

      // Shared wall-clock target. Both contexts run on this same machine,
      // so Date.now() is a common clock — the same absolute startAtMs on
      // each page is exactly what companionSync.localEquivalent(playAt)
      // produces once the estimators on both surfaces agree. We give a
      // comfortable lead (800ms, comfortably longer than the 250ms
      // coherence threshold) so that if a page ignored startAtMs and
      // played immediately instead, its play instant would land ~800ms
      // early and fail the per-page target assertion below — this is what
      // makes the test discriminating rather than trivially true.
      const sharedStartAtMs = Date.now() + 800;

      // Reset hooks on BOTH pages right before launching, so the first
      // 'play' entry we read is unambiguously our flight.
      await table.evaluate(() => (window as any).__bgAnimTestHooks.reset());
      await hand.evaluate(() => (window as any).__bgAnimTestHooks.reset());

      const launch = async (page: typeof table, startAtMs: number) => {
        return page.evaluate(async (startAt) => {
          const animator = document.createElement('boardgame-component-animator') as any;
          document.body.appendChild(animator);
          await animator.updateComplete;
          const a = document.createElement('div');
          const b = document.createElement('div');
          a.id = 'fly-real';
          a.style.cssText = 'position:fixed;top:10px;left:10px;width:20px;height:20px';
          b.style.cssText = 'position:fixed;top:400px;left:400px;width:20px;height:20px';
          document.body.append(a, b);
          // This is the production path: render-game assigns the installed
          // version's context to the animator and semantic renderer code calls
          // animateBetween without passing timing.
          animator.animationContext = {
            version: 42,
            startAtMs: startAt,
            slotDurationMs: 800,
            maxAnimationDurationMs: 600,
          };
          // Fire-and-forget: the flight resolves ~600ms after its delay
          // elapses; we only need the 'play' hook it records at visual
          // start, which we read below after awaiting the shared instant.
          void animator.animateBetween(a, b, 300);
        }, startAtMs);
      };

      // Launch on both surfaces. Order/skew here is absorbed by the shared
      // absolute startAtMs — each side's WAAPI delay self-corrects to the
      // common instant.
      await Promise.all([launch(table, sharedStartAtMs), launch(hand, sharedStartAtMs)]);

      // Wait until both pages have recorded their fly 'play' hook (the
      // delay has elapsed and the flight began).
      const readFirstFly = async (page: typeof table) => {
        await page.waitForFunction(() => {
          const h = (window as any).__bgAnimTestHooks;
          return h.log.some((e: any) => e.ev === 'active' && typeof e.detail === 'string' && e.detail.startsWith('fly:'));
        }, undefined, { timeout: 15000 });
        return page.evaluate(() => {
          const h = (window as any).__bgAnimTestHooks;
          const entry = h.log.find((e: any) => e.ev === 'active' && typeof e.detail === 'string' && e.detail.startsWith('fly:'));
          // Normalize to an absolute, cross-context-comparable epoch-ms
          // instant. performance.now() (entry.t) is relative to each
          // page's own timeOrigin; adding timeOrigin makes them shareable.
          return performance.timeOrigin + entry.t;
        });
      };

      const [tableStart, handStart] = await Promise.all([
        readFirstFly(table),
        readFirstFly(hand),
      ]);

      const skew = Math.abs(tableStart - handStart);
      // eslint-disable-next-line no-console
      console.log(`[waapi-companion] measured cross-surface skew: ${skew.toFixed(1)}ms`);
      test.info().annotations.push({ type: 'skew', description: `${skew.toFixed(1)}ms` });

      // Coherence threshold (spec §8.4): the two surfaces must launch the
      // same flight within 250ms. Do NOT weaken this.
      expect(skew).toBeLessThan(250);

      // Skew alone doesn't discriminate: if both pages ignored startAtMs
      // and played immediately, they'd still be in sync with each other
      // (both near "now"), and the skew assertion above would trivially
      // pass. What actually proves startAtMs was honored is that EACH
      // page's play instant lands near the shared absolute target — not
      // ~800ms early, which is what "played immediately" would produce.
      const tableOffset = tableStart - sharedStartAtMs;
      const handOffset = handStart - sharedStartAtMs;
      // eslint-disable-next-line no-console
      console.log(`[waapi-companion] table offset from target: ${tableOffset.toFixed(1)}ms, hand offset from target: ${handOffset.toFixed(1)}ms`);
      test.info().annotations.push({ type: 'table-offset', description: `${tableOffset.toFixed(1)}ms` });
      test.info().annotations.push({ type: 'hand-offset', description: `${handOffset.toFixed(1)}ms` });

      expect(Math.abs(tableOffset)).toBeLessThan(200);
      expect(Math.abs(handOffset)).toBeLessThan(200);
    } finally {
      await tableCtx.close();
      await handCtx.close();
    }
  });

  test('real blackjack move stays synchronized through socket, bundle queue, renderers, and auto-fly', async ({ browser }) => {
    const tableCtx = await browser.newContext();
    const handCtx = await browser.newContext();
    const playerTwoCtx = await browser.newContext();
    const controller = await tableCtx.newPage();
    const hand = await handCtx.newPage();
    const playerTwo = await playerTwoCtx.newPage();
    try {
      await createOfflineGame(controller, 'blackjack', { companionMode: true, adminMode: false });
      const sharedGame = new URL(controller.url());
      const roomCode = (await controller.locator('.room-code-giant').textContent())?.trim();
      expect(roomCode).toMatch(/^[A-Z]{4,5}$/);
      // Blackjack needs two seated players to leave the gathering phase.
      // Join both through the real QR-code guest path; the first phone is
      // seat zero and supplies the measured Hit below.
      await joinCompanionAsGuest(hand, roomCode!, 'blackjack');
      await joinCompanionAsGuest(playerTwo, roomCode!, 'blackjack');
      const tableUrl = new URL(sharedGame);
      tableUrl.searchParams.set('display', 'table');
      // Keep the creator page open for lifecycle coverage; a second page in
      // its authenticated context is the actual shared Table surface.
      const table = await tableCtx.newPage();
      await table.goto(tableUrl.toString());

      const rendererReady = async (page: typeof table, tag: string) => {
        await page.waitForFunction(([fnSrc, selector]) => {
          // eslint-disable-next-line no-eval
          const deepQueryFirst = eval(`(${fnSrc})`);
          return !!deepQueryFirst(document, selector)
            && (window as any).__bgCompanionSync?.estimator?.sampleCount() >= 3;
        }, [`(${deepQueryFirstScript.toString()})()`, tag], { timeout: 20000 });
      };
      await Promise.all([
        rendererReady(table, 'boardgame-render-game-blackjack-table'),
        rendererReady(hand, 'boardgame-render-game-blackjack-hand'),
      ]);
      const hit = hand.getByRole('button', { name: 'Hit', exact: true });
      await expect(hit).toBeEnabled({ timeout: 20000 });
      // A new Blackjack game deals through a sequence of automatic moves.
      // Measure a deliberate Hit only after that sequence has reached a legal,
      // quiescent player turn on every surface. Hit becomes legal just before
      // the last reveal fix-up, so allow two protocol slots for that final
      // automatic version before requiring the gates to be closed.
      const waitForQuiescence = (page: typeof table) => page.waitForFunction((fnSrc) => {
        // eslint-disable-next-line no-eval
        const deepQueryFirst = eval(`(${fnSrc})`);
        const renderGame = deepQueryFirst(document, 'boardgame-render-game') as any;
        return renderGame && !renderGame.isAnimating;
      }, `(${deepQueryFirstScript.toString()})()`, { timeout: 20000 });
      await controller.waitForTimeout(1600);
      await Promise.all([controller, table, hand].map(waitForQuiescence));
      await Promise.all([
        table.evaluate(() => (window as any).__bgAnimTestHooks.reset()),
        hand.evaluate(() => (window as any).__bgAnimTestHooks.reset()),
      ]);
      await hit.click();

      const readCycle = async (page: typeof table) => {
        await page.waitForFunction(() => {
          const hooks = (window as any).__bgAnimTestHooks.log;
          const fly = hooks.find((entry: any) =>
            entry.ev === 'active' && entry.detail?.startsWith('fly:') && Number.isInteger(entry.version));
          return !!fly
            && hooks.some((entry: any) => entry.ev === 'active'
              && !entry.detail?.startsWith('fly:') && entry.version === fly.version)
            && hooks.some((entry: any) => entry.ev === 'install' && entry.version === fly.version);
        }, undefined, { timeout: 15000 });
        return page.evaluate(() => {
          const hooks = (window as any).__bgAnimTestHooks.log;
          const fly = hooks.find(
            (item: any) => item.ev === 'active' && item.detail?.startsWith('fly:')
              && Number.isInteger(item.version),
          );
          const ordinary = hooks.find((item: any) => item.ev === 'active'
            && !item.detail?.startsWith('fly:') && item.version === fly.version);
          const install = hooks.find((item: any) =>
            item.ev === 'install' && item.version === fly.version);
          return {
            version: fly.version,
            targetAt: fly.targetAtMs,
            flyAt: performance.timeOrigin + fly.t,
            ordinaryAt: performance.timeOrigin + ordinary.t,
            installAt: performance.timeOrigin + install.t,
            ordinaryTargetAt: ordinary.targetAtMs,
          };
        });
      };
      const [tableCycle, handCycle] = await Promise.all([readCycle(table), readCycle(hand)]);
      expect(Math.abs(tableCycle.flyAt - handCycle.flyAt)).toBeLessThan(250);
      expect(Math.abs(tableCycle.ordinaryAt - handCycle.ordinaryAt)).toBeLessThan(250);
      for (const cycle of [tableCycle, handCycle]) {
        // Cross-screen flights and the ordinary FLIP/property pipeline must
        // each launch near their own page's clock-converted server target.
        // Merely being equally immediate on both pages cannot pass this.
        expect(Math.abs(cycle.flyAt - cycle.targetAt)).toBeLessThan(250);
        expect(Math.abs(cycle.ordinaryAt - cycle.targetAt)).toBeLessThan(250);
        expect(cycle.ordinaryTargetAt).toBe(cycle.targetAt);
        expect(cycle.installAt).toBeLessThanOrEqual(cycle.targetAt + 50);
        expect(cycle.installAt).toBeGreaterThanOrEqual(cycle.targetAt - 350);
      }
    } finally {
      await tableCtx.close();
      await handCtx.close();
      await playerTwoCtx.close();
    }
  });

  test('a deliberately local flight can opt out of the version timeline', async ({ page }) => {
    await createOfflineGame(page, 'blackjack');
    await page.waitForFunction(() => (window as any).__bgAnimTestHooks !== undefined);
    const result = await page.evaluate(async () => {
      const animator = document.createElement('boardgame-component-animator') as any;
      document.body.appendChild(animator);
      await animator.updateComplete;
      const real = document.createElement('div');
      const stub = document.createElement('div');
      real.style.cssText = 'position:fixed;top:10px;left:10px;width:20px;height:20px';
      stub.style.cssText = 'position:fixed;top:300px;left:300px;width:20px;height:20px';
      document.body.append(real, stub);
      animator.animationContext = {
        version: 42,
        startAtMs: Date.now() + 800,
        slotDurationMs: 800,
        maxAnimationDurationMs: 600,
      };
      (window as any).__bgAnimTestHooks.reset();
      const before = Date.now();
      void animator.animateBetween(real, stub, 50, { timing: 'immediate' });
      await new Promise((resolve) => requestAnimationFrame(() => resolve(null)));
      return {
        elapsed: Date.now() - before,
        played: (window as any).__bgAnimTestHooks.log.some(
          (entry: any) => entry.ev === 'active' && entry.detail?.startsWith('fly:'),
        ),
      };
    });
    expect(result.played).toBe(true);
    expect(result.elapsed).toBeLessThan(300);
  });
});

// Verdict-gating (#798 final piece): the outcome/verdict text must never
// appear while animations are still running -- cards must visually land
// before any surface announces "Game over". boardgame-render-game mirrors
// its isAnimating gate onto the renderer as `animating` (both at the gate
// flips and at renderer instantiation -- see _applyAnimatingToRenderer in
// boardgame-render-game.ts), and BoardgameTableViewBase.renderGameOverBanner
// / BoardgameHandViewBase.renderHandHeader gate their outcome markup on
// `!this.animating` in addition to `gameFinished`. Blackjack ships real
// -table/-hand renderers (unlike the games this task's original brief
// named), so this drives the REAL component tree via ?display=table.
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

test.describe('verdict gating on animation completion', () => {
  test('game outcome is suppressed while animating, appears once the gate closes', async ({ page }) => {
    await createOfflineGame(page, 'blackjack');
    const tableUrl = new URL(page.url());
    tableUrl.searchParams.set('display', 'table');
    await page.goto(tableUrl.toString());
    await page.waitForSelector('boardgame-render-game', { timeout: 15000 });
    await page.waitForFunction(() => (window as any).__bgAnimTestHooks !== undefined, undefined, { timeout: 15000 });

    // Wait for the real -table renderer (boardgame-render-game-blackjack-table)
    // to be instantiated and reachable via deep shadow-DOM query.
    const findTableRenderer = () => page.evaluate((fnSrc: string) => {
      // eslint-disable-next-line no-eval
      const deepQueryFirst = eval(`(${fnSrc})`);
      const r = deepQueryFirst(document, 'boardgame-render-game-blackjack-table');
      return !!r;
    }, `(${deepQueryFirstScript.toString()})()`);
    await expect.poll(findTableRenderer, { timeout: 15000 }).toBe(true);

    // Inject gameFinished/gameWinners/animating directly onto the real
    // renderer instance -- this exercises the actual guard in the actual
    // component (renderGameOverBanner), not a synthetic stand-in.
    const setRendererProps = (props: Record<string, unknown>) => page.evaluate(
      ([fnSrc, p]) => {
        // eslint-disable-next-line no-eval
        const deepQueryFirst = eval(`(${fnSrc})`);
        const r = deepQueryFirst(document, 'boardgame-render-game-blackjack-table') as any;
        Object.assign(r, p);
        return r.updateComplete;
      },
      [`(${deepQueryFirstScript.toString()})()`, props] as const,
    );

    const outcomeVisible = () => page.evaluate((fnSrc: string) => {
      // eslint-disable-next-line no-eval
      const deepQueryFirst = eval(`(${fnSrc})`);
      const r = deepQueryFirst(document, 'boardgame-render-game-blackjack-table') as any;
      const outcome = r?.shadowRoot?.querySelector('boardgame-game-outcome') as HTMLElement | null;
      return !!outcome?.shadowRoot?.querySelector('#outcome');
    }, `(${deepQueryFirstScript.toString()})()`);

    // gameFinished + animating=true: the verdict must stay hidden.
    await setRendererProps({ gameFinished: true, gameWinners: [0], animating: true });
    expect(await outcomeVisible(), 'outcome must be absent while animating=true, even though gameFinished=true').toBe(false);

    // Gate closes: the verdict must now appear.
    await setRendererProps({ animating: false });
    expect(await outcomeVisible(), 'outcome must appear once animating=false and gameFinished=true').toBe(true);
  });

  test('animateBetween flight on a real card holds the render-game gate until it settles', async ({ page }) => {
    // #798: a synced deal flight (animateBetween) on a REAL animatable card
    // must keep the render-game completion gate OPEN until the flight
    // settles, so the game-over verdict can't appear while the winning card
    // is still in the air. Before this fix animateBetween used a raw
    // real.animate() that never registered with the gate, so the gate could
    // close (and the banner appear) mid-flight. This drives the REAL
    // production render-game + its REAL animator against a REAL card, and
    // asserts the render-game[is-animating] gate is open mid-flight and
    // closed only after settlement.
    // This assertion drives the renderer directly and does not need the
    // admin panel; skipping that setup also avoids its view-as transition
    // racing this page.evaluate with a navigation.
    await createOfflineGame(page, 'debuganimations', { adminMode: false });

    // Quiescent baseline: gate closed.
    await page.waitForFunction(() => {
      const h = (window as any).__bgAnimTestHooks;
      return h.gateCloses >= h.gateOpens;
    }, undefined, { timeout: 20000 });
    // The initial info install can close one cycle immediately before a
    // queued automatic bundle opens the next. Require a short quiet period so
    // that unrelated state-cycle reset cannot overwrite this isolated probe's
    // gate while its flight is running.
    await page.waitForTimeout(500);
    await page.waitForFunction((fnSrc: string) => {
      // eslint-disable-next-line no-eval
      const deepQueryFirst = eval(`(${fnSrc})`);
      const rg = deepQueryFirst(document, 'boardgame-render-game') as any;
      return rg && !rg.hasAttribute('is-animating');
    }, `(${deepQueryFirstScript.toString()})()`, { timeout: 20000 });

    // The wrapper mounts before the game-specific renderer has necessarily
    // stamped its first card. Wait for the actual test prerequisites rather
    // than racing the initial Lit render.
    const waitForProbeTargets = () => page.waitForFunction((fnSrc: string) => {
        // eslint-disable-next-line no-eval
        const deepQueryFirst = eval(`(${fnSrc})`);
        const rg = deepQueryFirst(document, 'boardgame-render-game') as any;
        const animator = rg?.shadowRoot?.querySelector('#animator');
        const card = deepQueryFirst(document, 'boardgame-card')
          || deepQueryFirst(document, 'boardgame-component');
        return !!rg && !!animator && !!card;
      }, `(${deepQueryFirstScript.toString()})()`, { timeout: 15000 });
    await waitForProbeTargets();

    const runProbe = () => page.evaluate(async (fnSrc: string) => {
      // eslint-disable-next-line no-eval
      const deepQueryFirst = eval(`(${fnSrc})`);
      const rg = deepQueryFirst(document, 'boardgame-render-game') as any;
      // The production animator lives in render-game's shadow root.
      const animator = rg?.shadowRoot?.querySelector('#animator') as any;
      // Any real animatable card/component connected inside the renderer's
      // tree. It extends BoardgameAnimatableItem, so it has play().
      const card = deepQueryFirst(document, 'boardgame-card')
        || deepQueryFirst(document, 'boardgame-component');
      if (!rg || !animator || !card) {
        return { setupOk: false, hasPlay: false, midFlightAnimating: null, afterAnimating: null };
      }
      const hasPlay = typeof (card as any).play === 'function';

      // The counter-based page baseline can become quiescent a microtask
      // before this particular card's settlement bookkeeping. Make the probe
      // itself independent of any initial-deal animation still owned by the
      // card; otherwise its pre-existing gated count can prevent this flight
      // from emitting a fresh will-animate/animation-done pair.
      (card as any).finishAllAnimations();
      await (card as any).settled();

      // Open the completion gate exactly as a state-change cycle would, so
      // there is a live gate for the flight to hold. (_resetAnimating is
      // private; we reach it deliberately to isolate the flight's effect on
      // the gate from any incidental FLIP animations.)
      (rg as any)._resetAnimating();
      await rg.updateComplete;

      // Fly the real card from a shifted position to its resting spot. Use
      // an explicit HTMLElement stub with a non-overlapping rect so dx/dy
      // are non-zero (animateBetween early-returns on a zero delta).
      const stub = document.createElement('div');
      stub.style.cssText = 'position:fixed;top:500px;left:500px;width:40px;height:60px';
      document.body.appendChild(stub);

      // Fire-and-forget so we can sample the gate WHILE the flight is live.
      const flightDone = animator.animateBetween(card, stub, 1000);

      // Let the will-animate bubble to render-game and flip the gate. Poll
      // within the deliberately long flight instead of assuming one frame is
      // sufficient under a loaded full-suite browser worker.
      const gateDeadline = performance.now() + 500;
      while (!rg.hasAttribute('is-animating') && performance.now() < gateDeadline) {
        await new Promise((r) => requestAnimationFrame(() => r(null)));
      }
      const midFlightAnimating = rg.hasAttribute('is-animating');

      // Now let the flight settle; the animation-done must close the gate.
      await flightDone;
      await rg.updateComplete;
      const afterAnimating = rg.hasAttribute('is-animating');

      stub.remove();
      return { setupOk: true, hasPlay, midFlightAnimating, afterAnimating };
    }, `(${deepQueryFirstScript.toString()})()`);
    let result: Awaited<ReturnType<typeof runProbe>> | undefined;
    for (let attempt = 0; attempt < 3; attempt++) {
      try {
        result = await runProbe();
        break;
      } catch (error) {
        const navigated = error instanceof Error && error.message.includes('Execution context was destroyed');
        if (!navigated || attempt === 2) throw error;
        await page.waitForLoadState('domcontentloaded');
        await waitForProbeTargets();
      }
    }
    if (!result) throw new Error('gate probe did not produce a result');

    expect(result.setupOk, 'render-game, its animator, and a real card must all be present').toBe(true);
    expect(result.hasPlay, 'the real card must be an animatable item with play()').toBe(true);
    // The load-bearing #798 assertions: gate OPEN while the flight is in the
    // air, CLOSED only once it settles.
    expect(result.midFlightAnimating, 'render-game[is-animating] must be true while the flight is airborne').toBe(true);
    expect(result.afterAnimating, 'render-game[is-animating] must be false once the flight settles').toBe(false);
  });

  test('renderer.animating mirrors render-game.isAnimating through a real move animation', async ({ page }) => {
    // debuganimations (not blackjack) drives this: it exposes a reliable
    // button-triggered move (waapi-buttons.spec.ts uses the same "To
    // Hidden" trigger for the analogous isAnimating-attribute check), so
    // the plumbing assertion isn't coupled to blackjack's move set or to
    // the initial deal's own animation burst racing the sample.
    await createOfflineGame(page, 'debuganimations');

    const sample = () => page.evaluate((fnSrc: string) => {
      // eslint-disable-next-line no-eval
      const deepQueryFirst = eval(`(${fnSrc})`);
      const rg = deepQueryFirst(document, 'boardgame-render-game') as any;
      const renderer = deepQueryFirst(document, 'boardgame-render-game-debuganimations') as any;
      return {
        isAnimating: rg ? rg.isAnimating : null,
        rendererAnimating: renderer ? renderer.animating : null,
      };
    }, `(${deepQueryFirstScript.toString()})()`);

    // Quiescent baseline: gate closed, renderer mirrors it.
    await page.waitForFunction(() => {
      const h = (window as any).__bgAnimTestHooks;
      return h.gateCloses >= h.gateOpens;
    }, undefined, { timeout: 20000 });
    const before = await sample();
    expect(before.isAnimating).toBe(false);
    expect(before.rendererAnimating).toBe(false);

    await page.getByRole('button', { name: 'To Hidden' }).click();

    // DURING: poll until the gate opens, then sample both properties
    // together -- they must agree at every observed instant, not just at
    // the two quiescent endpoints.
    await expect.poll(async () => (await sample()).isAnimating, { timeout: 20000 }).toBe(true);
    const during = await sample();
    expect(during.rendererAnimating).toBe(true);

    // AFTER: wait for the gate to close, then re-sample.
    await page.waitForFunction(() => {
      const h = (window as any).__bgAnimTestHooks;
      return h.gateCloses >= h.gateOpens;
    }, undefined, { timeout: 20000 });
    await expect.poll(async () => (await sample()).isAnimating, { timeout: 20000 }).toBe(false);
    const after = await sample();
    expect(after.rendererAnimating).toBe(false);
  });
});
