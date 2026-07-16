import { test, expect } from '@playwright/test';
import { createOfflineGame } from './helpers';

test('post-animation-delay defers animation-done', async ({ page }) => {
  await createOfflineGame(page, 'blackjack');
  const elapsed = await page.evaluate(async () => {
    const ele = document.createElement('boardgame-component') as any;
    ele.postAnimationDelay = 300;
    document.body.appendChild(ele);
    await ele.updateComplete;
    const start = performance.now();
    ele.play(ele, [{ opacity: 0 }, { opacity: 1 }], { duration: 50 });
    await ele.settled();
    return performance.now() - start;
  });
  expect(elapsed).toBeGreaterThanOrEqual(340); // 50ms anim + 300ms endDelay, minus jitter
});

test('wait-for-animation=false items do not hold the gate', async ({ page }) => {
  await createOfflineGame(page, 'blackjack');
  const r = await page.evaluate(async () => {
    const ele = document.createElement('boardgame-component') as any;
    ele.waitForAnimation = false;
    document.body.appendChild(ele);
    await ele.updateComplete;
    let done = false;
    ele.addEventListener('animation-done', () => { done = true; });
    const anim = ele.play(ele, [{ opacity: 0 }, { opacity: 1 }], { duration: 60000 });
    await ele.settled(); // resolves immediately: nothing gated
    return { settledImmediately: true, doneFired: done, running: anim.playState === 'running' };
  });
  expect(r.settledImmediately).toBe(true);
  expect(r.doneFired).toBe(false);
  expect(r.running).toBe(true);
});

test('stack forwards post-animation-delay to stamped components', async ({ page }) => {
  await createOfflineGame(page, 'blackjack');

  // All of blackjack's <boardgame-component-stack> elements (deck, hands,
  // discard) live nested inside other components' shadow roots
  // (boardgame-render-game-blackjack-table etc. don't override
  // createRenderRoot), so a plain document.querySelectorAll can't see them
  // -- walk into every shadowRoot to find them. On a fresh deal the deck
  // stack (layout="stack") holds all 52 real, stamped <boardgame-component>
  // children, so pick the first stack that actually has stamped
  // components rather than trusting a fixed selector/index.
  const v = await page.evaluate(async () => {
    function deepQueryAll(root: Document | ShadowRoot | Element, selector: string): Element[] {
      const results: Element[] = [...root.querySelectorAll(selector)];
      for (const el of Array.from(root.querySelectorAll('*'))) {
        if ((el as any).shadowRoot) {
          results.push(...deepQueryAll((el as any).shadowRoot, selector));
        }
      }
      return results;
    }

    const stacks = deepQueryAll(document, 'boardgame-component-stack') as any[];
    const stack = stacks.find((s) => (s.Components?.length ?? 0) > 0);
    if (!stack) return { found: false, value: undefined };

    stack.setAttribute('post-animation-delay', '150');
    // Plain HTML attributes on the stack aren't a reactive Lit @property
    // (unlike unsafeComponentAttrs), so nothing re-stamps children automatically
    // when it's set after the fact. Re-assign unsafeComponentAttrs to a fresh
    // object (same content) to trigger Lit's changed-property machinery,
    // which re-applies _attributesForComponents() -- including the new
    // post-animation-delay attribute -- to every already-stamped child
    // (see _applyComponentAttrsToChildren in boardgame-component-stack.ts).
    stack.unsafeComponentAttrs = { ...stack.unsafeComponentAttrs };
    await stack.updateComplete;

    const comp = stack.querySelector('[boardgame-component]') as any;
    return { found: true, value: comp ? comp.postAnimationDelay : undefined };
  });

  expect(v.found).toBe(true);
  expect(v.value).toBe(150);
});

test('stagger produces strictly increasing per-index animation delays', async ({ page }) => {
  await createOfflineGame(page, 'debuganimations');

  function deepQueryAllScript() {
    function deepQueryAll(root: Document | ShadowRoot | Element, selector: string): Element[] {
      const results: Element[] = [...root.querySelectorAll(selector)];
      for (const el of Array.from(root.querySelectorAll('*'))) {
        if ((el as any).shadowRoot) {
          results.push(...deepQueryAll((el as any).shadowRoot, selector));
        }
      }
      return results;
    }
    return deepQueryAll;
  }

  // Widen the timing window so the live-animation snapshot below (taken
  // right after the click) has plenty of time to catch every card's
  // animation before any of them finish -- avoids flakiness from the
  // default (short) --animation-length. Set directly on
  // <boardgame-render-game> (which forwards it as an inline --animation-length
  // custom property that everything inside it inherits, same mechanism as
  // its own defaultAnimationLength handling) since that's the closest
  // ancestor in the cascade -- a document.documentElement-level override is
  // shadowed by it and has no effect (verified while developing this test).
  await page.evaluate((deepQueryAllFn: string) => {
    // eslint-disable-next-line no-eval
    const deepQueryAll = eval(`(${deepQueryAllFn})`);
    const renderGames = deepQueryAll(document, 'boardgame-render-game') as HTMLElement[];
    for (const rg of renderGames) rg.style.setProperty('--animation-length', '3s');
  }, `(${deepQueryAllScript.toString()})()`);

  // Set stagger on every stack that actually holds >1 stamped component --
  // in particular this covers the AllVisibleStack/AllHiddenStack pair
  // driven by the "To Hidden"/"To Visible" buttons (see
  // boardgame-render-game-debuganimations.ts's #all div) -- so the
  // animator's per-collection animating index has something to stagger.
  const setup = await page.evaluate((deepQueryAllFn: string) => {
    // eslint-disable-next-line no-eval
    const deepQueryAll = eval(`(${deepQueryAllFn})`);
    const stacks = deepQueryAll(document, 'boardgame-component-stack') as any[];
    const target = stacks.filter((s) => (s.Components?.length ?? 0) > 1);
    for (const s of target) s.stagger = 0.2;
    return { count: target.length };
  }, `(${deepQueryAllScript.toString()})()`);
  expect(setup.count).toBeGreaterThan(0);

  // Ground truth for the actual `delay` WAAPI applied. Neither
  // document.getAnimations() nor element.getAnimations({subtree: true})
  // cross shadow boundaries in this browser (verified while developing this
  // test, including with a minimal open-shadow-root repro) -- a card's flip
  // animation targets its shadow #inner div, and the host transform/opacity
  // animations target the card itself, so getAnimations() must be called
  // per-element while walking every shadow root to find them all. Reset the
  // hooks counter and poll until several plays have been recorded (the
  // whole per-cycle play() loop runs synchronously once started, so by the
  // time a handful have landed, all of this cycle's animations -- including
  // the higher-index, higher-delay ones -- already exist as live Animation
  // objects, just not yet in their active phase).
  await page.evaluate(() => { (window as any).__bgAnimTestHooks.reset(); });
  await page.getByRole('button', { name: 'To Hidden' }).click();
  await page.waitForFunction(() => (window as any).__bgAnimTestHooks.plays > 5, undefined, { timeout: 10000 });

  const delaysByIndex = await page.evaluate((deepQueryAllFn: string) => {
    // eslint-disable-next-line no-eval
    const deepQueryAll = eval(`(${deepQueryAllFn})`);
    const allElements = [document.body, ...deepQueryAll(document, '*')] as Element[];
    const withDelay: number[] = [];
    for (const el of allElements) {
      for (const anim of el.getAnimations()) {
        const timing = (anim.effect as KeyframeEffect)?.getTiming?.();
        const delay = timing?.delay;
        if (typeof delay === 'number' && delay > 0) {
          withDelay.push(delay);
        }
      }
    }
    return withDelay;
  }, `(${deepQueryAllScript.toString()})()`);

  // Wait for the cycle to fully settle before ending the test so later
  // tests (and the regression suite) don't inherit an in-flight animation.
  await page.waitForFunction(() => {
    const h = (window as any).__bgAnimTestHooks;
    return h.gateCloses >= h.gateOpens;
  }, undefined, { timeout: 20000 });

  // The watchdog must NOT have fired across this staggered cycle. With
  // --animation-length forced to 3s and stagger 0.2, later cards start
  // well past 4s (index 2 alone starts at 1200ms, and a real deal cascade
  // of many cards pushes the last card's *settle* far past a flat 4s
  // watchdog). Before the event-driven watchdog extension this could
  // silently force-close mid-cascade (watchdogFirings > 0); the extension
  // scales the deadline to each play's declared settle budget, so a clean
  // cycle records zero firings. A non-zero count here means the watchdog
  // wrongly force-closed a legitimate long animation -- a regression.
  const watchdogFirings = await page.evaluate(
    () => (window as any).__bgAnimTestHooks.watchdogFirings as number,
  );
  expect(watchdogFirings, 'watchdog must not fire during a legitimate long staggered cycle').toBe(0);

  // At least two distinct non-zero delays observed (one stack's worth of
  // staggered cards), and they form a strictly increasing sequence once
  // deduped+sorted -- i.e. index*stagger*animationLength for index
  // 1,2,3,... (index 0 has delay 0, which is legitimately "no delay" and
  // excluded by the `delay > 0` filter above).
  const uniqueSorted = [...new Set(delaysByIndex)].sort((a, b) => a - b);
  expect(uniqueSorted.length).toBeGreaterThan(1);
  for (let i = 1; i < uniqueSorted.length; i++) {
    expect(uniqueSorted[i]).toBeGreaterThan(uniqueSorted[i - 1]);
  }

  // Strict check against the formula: delay = index * stagger * animLength.
  // animLength was forced to 3000ms above; stagger is 0.2 -> step = 600ms.
  const expectedStep = 0.2 * 3000;
  for (let i = 0; i < uniqueSorted.length; i++) {
    const nearestIndex = Math.round(uniqueSorted[i] / expectedStep);
    expect(nearestIndex).toBeGreaterThan(0);
    expect(Math.abs(uniqueSorted[i] - nearestIndex * expectedStep)).toBeLessThan(expectedStep * 0.1);
  }
});

test('wait-for-animation="false" set directly on an item parses as false and ungates', async ({ page }) => {
  await createOfflineGame(page, 'blackjack');
  const r = await page.evaluate(async () => {
    const ele = document.createElement('boardgame-component') as any;
    ele.setAttribute('wait-for-animation', 'false');
    document.body.appendChild(ele);
    await ele.updateComplete;
    const parsedFalse = ele.waitForAnimation === false;
    ele.play(ele, [{ opacity: 0 }, { opacity: 1 }], { duration: 60000 });
    await ele.settled(); // must resolve immediately since ungated
    const bare = document.createElement('boardgame-component') as any;
    bare.setAttribute('wait-for-animation', '');
    document.body.appendChild(bare);
    await bare.updateComplete;
    return { parsedFalse, bareParsedTrue: bare.waitForAnimation === true };
  });
  expect(r.parsedFalse).toBe(true);
  expect(r.bareParsedTrue).toBe(true);
});
