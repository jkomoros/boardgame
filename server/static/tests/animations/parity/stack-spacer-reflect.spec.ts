import { test, expect } from '@playwright/test';
import { createOfflineGame, settleInitialLoad } from '../helpers';

/**
 * THE SPACER ATTRIBUTE NOBODY WROTE.
 *
 * A stack keeps ONE placeholder host -- a `spacer` -- so an empty stack still
 * occupies its slot. `boardgame-component-stack` finds that host with the
 * ATTRIBUTE selector `#container>[boardgame-component][spacer]`, in three
 * places:
 *
 *   `_refreshShadowViewComponents()`  re-run the view recipe over it
 *   `_removeShadowViewComponents()`   tear it down when the recipe changes shape
 *   `_slotChanged()`                  `haveSpacer` (don't build a second one)
 *                                     and the `while (spacers.length > target)`
 *                                     loop (remove surplus / remove it once the
 *                                     stack fills)
 *
 * `BoardgameComponent.spacer` was declared `@property({ type: Boolean })` with
 * no `reflect: true`, so the value lived only in Lit state and that attribute
 * was NEVER written. All three selectors matched nothing, always. Consequences,
 * all measured on this fixture before the fix:
 *
 *   - `haveSpacer` was permanently false, so every `_slotChanged` on an empty
 *     stack built ANOTHER placeholder;
 *   - the removal loop's `spacers` NodeList was permanently empty, so it never
 *     removed one -- not the surplus, and not the last one when the stack
 *     filled;
 *   - the two shadow-view methods silently skipped the spacer entirely, so a
 *     changed `componentView` left a stale host behind.
 *
 * Measured before the fix, cycling one raw stack empty/full six times:
 * 3, 3, 4, 4, 5, 5, 6, 6, 7, 7, 8, 8, 9 hosts -- monotone, +1 per empty phase,
 * never falling to 0 when full, and `[spacer]` matching 0 at every single step.
 * On a freshly loaded `debuganimations`: 60 placeholder hosts across 15 stacks
 * where 15 were intended, and 0 selector matches.
 *
 * Note this is NOT the attribute-NAME trap `property-attribute-names.test.ts`
 * lints -- `spacer` is one word, so its observed name was right. It was the
 * reflection direction that was missing, the same one `id` a few lines above
 * carries a comment about for the same class of reason.
 */
test.describe('the spacer attribute reflects', () => {
  test('an empty stack keeps exactly one placeholder, and a full one keeps none',
    async ({ page }) => {
      await page.goto('/');
      const series = await page.evaluate(async () => {
        const view = await import('/src/components/component-view.ts');
        await import('/src/components/boardgame-component-stack.ts');
        document.body.innerHTML = '';
        const stack = document.createElement('boardgame-component-stack') as any;
        stack.setAttribute('layout', 'stack');
        document.body.appendChild(stack);
        stack.componentView = view.tokenView({
          properties: () => ({ type: 'disc', color: 'red' }),
        });

        const settle = async () => {
          await stack.updateComplete;
          await new Promise<void>((r) =>
            requestAnimationFrame(() => requestAnimationFrame(() => r())));
          await new Promise<void>((r) => setTimeout(r, 40));
        };
        const hosts = (): any[] => [
          ...stack.shadowRoot
            .querySelector('#container')
            .querySelectorAll(':scope > [boardgame-component]'),
        ];
        const snapshot = (phase: string) => ({
          phase,
          // Every direct host in #container, however it was made.
          hosts: hosts().length,
          // ...that believe they are spacers, read from the PROPERTY.
          byProperty: hosts().filter((el: any) => el.spacer === true).length,
          // ...and what the stack's own selector can actually see. Before the
          // fix this was 0 in every phase.
          bySelector: stack.shadowRoot
            .querySelectorAll('#container>[boardgame-component][spacer]').length,
        });
        const component = (id: string) => ({
          ID: id, Index: 0, Deck: 'd', GameName: 'g', Values: {}, DynamicValues: {},
        });

        const out: any[] = [];
        // Six full/empty cycles. One is enough to see the missing removal;
        // six is what makes the growth undeniably unbounded rather than a
        // one-off off-by-one.
        for (let i = 0; i < 6; i++) {
          stack.stack = { Deck: 'd', Size: 1, Components: [component('a' + i)], IDs: ['a' + i] };
          await settle();
          out.push(snapshot('full'));
          stack.stack = { Deck: 'd', Size: 0, Components: [], IDs: [] };
          await settle();
          out.push(snapshot('empty'));
        }
        stack.remove();
        return out;
      });

      for (const step of series) {
        const want = step.phase === 'empty' ? 1 : 0;
        expect(step.hosts, `${step.phase}: exactly ${want} placeholder host`).toBe(want);
        // The property and the selector must agree. They disagreed in every
        // phase before the fix, which is the whole defect in one line.
        expect(step.bySelector, `${step.phase}: the [spacer] selector must see it`)
          .toBe(step.byProperty);
        expect(step.bySelector).toBe(want);
      }
      // Not vacuous: the sequence must actually have exercised both phases.
      expect(series.filter((s: any) => s.phase === 'empty').length).toBe(6);
      expect(series.filter((s: any) => s.phase === 'full').length).toBe(6);
    });

  /**
   * The production witness. A game with sixteen stacks is where the leak was
   * visible: 60 hidden placeholder hosts, each a full custom element with its
   * own shadow root and its own view recipe run over it, for 15 that were
   * wanted. Nothing threw, and nothing was visible -- placeholders are
   * `visibility: hidden` -- which is exactly why it survived.
   */
  test('a real game carries one placeholder per empty stack, not a pile',
    async ({ page }) => {
      await createOfflineGame(page, 'debuganimations');
      await settleInitialLoad(page);
      const measured = await page.evaluate(() => {
        const walk = (root: Document | ShadowRoot, acc: Element[]) => {
          for (const el of root.querySelectorAll('boardgame-component-stack')) acc.push(el);
          for (const el of root.querySelectorAll('*')) {
            if ((el as any).shadowRoot) walk((el as any).shadowRoot, acc);
          }
        };
        const stacks: Element[] = [];
        walk(document, stacks);
        const per = stacks.map((stack: any) => {
          const container = stack.shadowRoot?.querySelector('#container');
          if (!container) return null;
          const hosts = [...container.querySelectorAll(':scope > [boardgame-component]')];
          return {
            byProperty: hosts.filter((el: any) => el.spacer === true).length,
            bySelector: stack.shadowRoot
              .querySelectorAll('#container>[boardgame-component][spacer]').length,
          };
        }).filter(Boolean) as { byProperty: number; bySelector: number }[];
        return {
          stacks: per.length,
          worst: Math.max(0, ...per.map((p) => p.byProperty)),
          totalByProperty: per.reduce((sum, p) => sum + p.byProperty, 0),
          totalBySelector: per.reduce((sum, p) => sum + p.bySelector, 0),
        };
      });

      expect(measured.stacks, 'debuganimations must still render its stacks')
        .toBeGreaterThan(10);
      // A stack wants at most ONE. Before the fix every empty stack carried 4.
      expect(measured.worst, 'no stack may carry more than one placeholder')
        .toBeLessThanOrEqual(1);
      // And every one of them must be findable by the selector the stack's own
      // bookkeeping uses. This was 0 against 60 before the fix.
      expect(measured.totalBySelector, 'the selector must see every placeholder')
        .toBe(measured.totalByProperty);
      // Not vacuous: there must be at least one placeholder to have seen.
      expect(measured.totalByProperty, 'at least one empty stack must draw one')
        .toBeGreaterThanOrEqual(1);
    });
});
