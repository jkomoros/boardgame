import { test, expect } from '@playwright/test';
import { createOfflineGame, settleInitialLoad } from '../helpers';

/**
 * `faux-components`, THE ATTRIBUTE, ON A RAW STACK.
 *
 * `boardgame-component-stack` declares `@property({ type: Number })
 * fauxComponents` with no `attribute:` option, so Lit's default lowercasing
 * makes its observed attribute `fauxcomponents` -- and nothing writes that.
 * `boardgame-component-zone` is the only place the dashed spelling was ever
 * mapped, and it maps it for ITSELF and then forwards the value down as a
 * property.
 *
 * So a markup author who writes `faux-components="5"` directly on a stack gets
 * silence: the attribute lands in the DOM, `fauxComponents` stays 0, and no
 * faux host is ever built. Measured before the fix, on a real stack:
 * `getAttribute('faux-components') === "5"`, `stack.fauxComponents === 0`, zero
 * faux elements. Both of `debuganimations`' uses are that spelling, so that
 * faux path had never run in the one game written to exercise it.
 *
 * The unit fixtures could not catch it: `tests/renderer/renderer-fixture.spec.ts`
 * sets `stack.fauxComponents = 4` as a PROPERTY, which was always fine. Only an
 * attribute can see an attribute bug.
 */
test.describe('faux-components on a raw stack', () => {
  test('the dashed attribute reaches the property and builds the hosts', async ({ page }) => {
    await page.goto('/');
    const result = await page.evaluate(async () => {
      const view = await import('/src/components/component-view.ts');
      await import('/src/components/boardgame-component-stack.ts');
      document.body.innerHTML = '';
      const stack = document.createElement('boardgame-component-stack') as any;
      // The spelling a markup author writes, and the spelling debuganimations
      // uses. Set as an ATTRIBUTE, deliberately: the property always worked.
      stack.setAttribute('faux-components', '5');
      stack.setAttribute('layout', 'stack');
      document.body.appendChild(stack);
      stack.componentView = view.tokenView({
        properties: () => ({ type: 'disc', color: 'red' }),
      });
      const component = (id: string) => ({
        ID: id, Index: 0, Deck: 'd', GameName: 'g', Values: {}, DynamicValues: {},
      });
      const real = [component('a'), component('b')];
      stack.stack = { Deck: 'd', Size: real.length, Components: real, IDs: real.map((c) => c.ID) };
      await stack.updateComplete;
      await new Promise<void>((r) => requestAnimationFrame(() => requestAnimationFrame(() => r())));
      await new Promise<void>((r) => setTimeout(r, 50));
      const faux = stack.shadowRoot.querySelectorAll('#faux-components [boardgame-component]');
      const out = {
        attribute: stack.getAttribute('faux-components'),
        property: stack.fauxComponents,
        fauxHosts: faux.length,
        fauxTypes: [...faux].map((el: any) => el.type),
      };
      stack.remove();
      return out;
    });

    expect(result.attribute, 'the attribute is in the DOM either way').toBe('5');
    // THE BUG. Lit observed `fauxcomponents`, so this stayed 0.
    expect(result.property, 'the dashed attribute must reach the property').toBe(5);
    // fauxComponents is a FLOOR, not an addition: 5 wanted minus 2 real = 3.
    expect(result.fauxHosts, 'five wanted, two real, so three faux').toBe(3);
    expect(result.fauxTypes, 'and they are built from the stack\'s own view')
      .toEqual(['disc', 'disc', 'disc']);
  });

  /**
   * The other half, and the reason this is worth a browser test rather than a
   * unit one: making the attribute work CHANGES WHAT DEBUGANIMATIONS DRAWS.
   * Its `#hidden` stack and its `#tokens-sanitized` destination stack both
   * carry `faux-components="5"` and both had been rendering nothing extra. A
   * fix that works is a fix that is visible, so it is asserted here rather
   * than discovered later.
   */
  test('debuganimations\' own faux-components stacks finally build hosts', async ({ page }) => {
    await createOfflineGame(page, 'debuganimations');
    await settleInitialLoad(page);
    const counts = await page.evaluate(() => {
      const walk = (root: Document | ShadowRoot, out: Element[]) => {
        for (const el of root.querySelectorAll('boardgame-component-stack')) out.push(el);
        for (const el of root.querySelectorAll('*')) {
          if ((el as any).shadowRoot) walk((el as any).shadowRoot, out);
        }
      };
      const stacks: Element[] = [];
      walk(document, stacks);
      return stacks
        .filter((s) => s.hasAttribute('faux-components'))
        .map((s: any) => ({
          want: s.fauxComponents,
          real: s.querySelectorAll('[boardgame-component]').length,
          faux: s.shadowRoot.querySelectorAll('#faux-components [boardgame-component]').length,
        }));
    });

    expect(counts.length, 'debuganimations declares faux-components on two stacks').toBe(2);
    for (const stack of counts) {
      expect(stack.want, 'the attribute must have reached the property').toBe(5);
      expect(stack.faux, 'and the faux hosts must exist')
        .toBe(Math.max(0, 5 - stack.real));
    }
    // Not vacuous: at least one of the two must actually be short of five, or
    // this proves nothing about the hosts being built.
    expect(counts.some((s) => s.faux > 0),
      'at least one of debuganimations\' stacks must draw faux hosts').toBe(true);
  });
});
