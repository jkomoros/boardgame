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
    // (unlike componentAttrs), so nothing re-stamps children automatically
    // when it's set after the fact. Re-assign componentAttrs to a fresh
    // object (same content) to trigger Lit's changed-property machinery,
    // which re-applies _attributesForComponents() -- including the new
    // post-animation-delay attribute -- to every already-stamped child
    // (see _applyComponentAttrsToChildren in boardgame-component-stack.ts).
    stack.componentAttrs = { ...stack.componentAttrs };
    await stack.updateComplete;

    const comp = stack.querySelector('[boardgame-component]') as any;
    return { found: true, value: comp ? comp.postAnimationDelay : undefined };
  });

  expect(v.found).toBe(true);
  expect(v.value).toBe(150);
});
