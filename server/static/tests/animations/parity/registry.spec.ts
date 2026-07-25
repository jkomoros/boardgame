import { test, expect } from '@playwright/test';
import { createOfflineGame, gateSnapshot, expectCleanGate } from '../helpers.js';

// Task 9: the ambient AnimatableRegistry closes #714's non-component
// discovery gap -- animatable items outside the shared component animator's
// own stack bookkeeping (a standalone die, status-text, fading-text, a
// game-authored token, ...) were previously invisible to render-game's
// cycle-start reset. See src/motion/animatable-registry.ts,
// src/components/boardgame-animatable-item.ts's connectedCallback/
// disconnectedCallback, and boardgame-render-game.ts's _resetAnimating.
const PARITY_TIMEOUT_MS = 180_000;

// deepQueryFirst/deepQueryAll walk into every shadowRoot (mirrors the
// pattern already used by waapi-buttons.spec.ts / waapi-attrs.spec.ts /
// trace.spec.ts's memory test), needed because <boardgame-die> and
// <boardgame-render-game> both live behind at least one shadow boundary.
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

test.describe('animatable registry', () => {
  test('pig: the standalone die is discovered by the ambient registry', async ({ page }) => {
    test.setTimeout(PARITY_TIMEOUT_MS);
    await createOfflineGame(page, 'pig');
    await expect(page.getByRole('button', { name: 'Roll die' })).toBeEnabled({ timeout: 30000 });

    const result = await page.evaluate((deepQueryFirstSrc: string) => {
      const deepQueryFirst = eval(`(${deepQueryFirstSrc})`) as
        (root: Document | ShadowRoot | Element, selector: string) => Element | null;
      const renderGame = deepQueryFirst(document, 'boardgame-render-game') as
        (HTMLElement & { animatableRegistry: { items(): readonly { tagName: string }[] } }) | null;
      if (!renderGame) return { renderGameFound: false, hasDie: false, tagNames: [] as string[] };
      const items = renderGame.animatableRegistry.items();
      return {
        renderGameFound: true,
        hasDie: items.some((item) => item.tagName === 'BOARDGAME-DIE'),
        tagNames: items.map((item) => item.tagName),
      };
    }, `(${deepQueryFirst.toString()})`);

    expect(result.renderGameFound).toBe(true);
    expect(result.hasDie, `registered tags were: ${result.tagNames.join(', ')}`).toBe(true);
  });

  test('pig: a second roll cycle finishes the first roll\'s die animation instead of leaving it running', async ({ page }) => {
    test.setTimeout(PARITY_TIMEOUT_MS);
    await createOfflineGame(page, 'pig');
    await expect(page.getByRole('button', { name: 'Roll die' })).toBeEnabled({ timeout: 30000 });

    // Drain game-creation setup completely before measuring, same rationale
    // as trace.spec.ts's debuganimations/memory scenarios.
    const setup = await gateSnapshot(page);
    await expectCleanGate(page, setup, 60000, { allowAlreadySettled: true });

    // Real double-clicking the die cannot exercise the race this fixes:
    // move proposal is gated on `!animating` (moves/action.ts's `reason`
    // getter reports 'animation-running' while the gate is open), and in
    // pig that flips back to enabled only once cycle 1's gate has fully
    // closed -- which, with no motion-release configured for this
    // renderer, coincides with the die's own animation having already
    // settled naturally. So a literal second click can never land while
    // cycle 1's spin is still running; there is no UI-observable race to
    // click through.
    //
    // Instead, drive the exact mechanism directly: perform ONE real roll
    // (a genuine WAAPI Animation on the die, not a synthetic one), confirm
    // it is actually running, then invoke render-game's own
    // _resetAnimating() -- the method _stateChanged() calls to open a new
    // cycle, including on the interrupted-cycle path (two state installs
    // landing before the first settles). It is TS-private, not
    // JS-private, and the surrounding code comments already document tests
    // reaching it directly for exactly this kind of gate/cycle isolation
    // (see the "memory: same-cycle state reinstall" test in
    // waapi-gate.spec.ts for the established precedent of reaching into
    // render-game internals rather than fighting realistic network/UI
    // timing for a race the client's own gating makes otherwise
    // unreachable). Calling it simulates "a second cycle's start lands
    // while the die is still animating" precisely, deterministically, and
    // without contending with the live game-view/Redux pipeline that a
    // synthetic .state= reassignment would collide with.
    const result = await page.evaluate(async () => {
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

      const renderGame = deepQueryFirst(document, 'boardgame-render-game') as
        (HTMLElement & { _resetAnimating(): void }) | null;
      const die = deepQueryFirst(document, 'boardgame-die') as HTMLElement | null;
      if (!renderGame || !die || !die.shadowRoot) return { found: false as const };
      const inner = die.shadowRoot.querySelector('#inner');
      const button = die.shadowRoot.querySelector('#main') as HTMLButtonElement | null;
      if (!inner || !button) return { found: false as const };

      const frame = () => new Promise<void>((r) => requestAnimationFrame(() => r()));
      const overallDeadline = performance.now() + 15000;

      // The die's own spin guards on an ACTUAL face change (boardgame-die.ts's
      // _selectedFaceChanged: "if (oldValue === newValue) return"), so a roll
      // that happens to land on the same face already showing (1-in-6, since
      // faces are 1-6) plays no animation at all -- nothing to interrupt.
      // Retry the roll (waiting for the button to re-enable between
      // attempts) until one genuinely animates.
      let anim1: Animation | undefined;
      let attempts = 0;
      while (!anim1) {
        if (performance.now() > overallDeadline) return { found: true as const, timedOut: true as const };
        while (button.disabled) {
          if (performance.now() > overallDeadline) return { found: true as const, timedOut: true as const };
          await frame();
        }
        attempts++;
        button.click();
        const perAttemptDeadline = Math.min(overallDeadline, performance.now() + 2000);
        while (performance.now() < perAttemptDeadline) {
          anim1 = inner.getAnimations().find((a) => a.playState === 'running');
          if (anim1) break;
          await frame();
        }
      }
      const playStateBeforeReset = anim1.playState;

      // Simulate the next cycle starting while roll 1's spin is still
      // genuinely running -- the exact moment the registry sweep must
      // force-finish it.
      renderGame._resetAnimating();
      const playStateAfterReset = anim1.playState;

      return {
        found: true as const, timedOut: false as const,
        playStateBeforeReset, playStateAfterReset,
      };
    });

    expect(result.found).toBe(true);
    if (!result.found) return;
    expect(result.timedOut, 'roll 1 must produce a genuinely running die animation').toBe(false);
    if (result.timedOut) return;
    // Confirms the animation was actually live before the reset -- without
    // this, a passing "no longer running" assertion below would be vacuous.
    expect(result.playStateBeforeReset).toBe('running');
    // The real assertion: render-game starting a new cycle must not leave a
    // standalone animatable's animation from the PRIOR cycle still running.
    expect(result.playStateAfterReset).not.toBe('running');
  });

  test('a fixture-mounted animatable registers under a provider and unregisters on detach', async ({ page }) => {
    await page.goto('/');

    const result = await page.evaluate(async () => {
      const { AnimatableRegistry } = await import('/src/motion/animatable-registry.ts');
      await import('/src/components/boardgame-fading-text.ts');

      const provider = document.createElement('div') as HTMLDivElement & {
        animatableRegistry: InstanceType<typeof AnimatableRegistry>;
      };
      provider.animatableRegistry = new AnimatableRegistry();
      document.body.appendChild(provider);

      const el = document.createElement('boardgame-fading-text');
      // Registration happens in connectedCallback, so check right after
      // the synchronous appendChild -- no need to wait for updateComplete.
      provider.appendChild(el);

      const registeredAfterAttach = provider.animatableRegistry.items().includes(el as any);

      el.remove();

      const registeredAfterDetach = provider.animatableRegistry.items().includes(el as any);

      return { registeredAfterAttach, registeredAfterDetach };
    });

    expect(result.registeredAfterAttach).toBe(true);
    expect(result.registeredAfterDetach).toBe(false);
  });
});
