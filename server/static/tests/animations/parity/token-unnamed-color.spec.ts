import { test, expect } from '@playwright/test';
import { createOfflineGame, settleInitialLoad } from '../helpers.js';

/**
 * A TOKEN WHOSE COLOUR HAS NO NAME MUST KEEP RENDERING — and until this spec
 * was written, one silently stopped for good, taking every checkers piece on
 * the board with it.
 *
 * `_computeClasses` puts the colour in the class map under its own name:
 * `{ [this.color.toLowerCase()]: true }`. A stack is entitled to pass no colour
 * at all — `boardgame-render-game-checkers.ts` passes `''` for any component
 * whose colour this player is not allowed to see, which on the first render pass
 * is EVERY component, because the view runs before the state arrives. That makes
 * the key the empty string, and Lit's `classMap` records it as a class it has
 * applied.
 *
 * On the NEXT update the colour has a name, so `''` is no longer in the map, and
 * `classMap` does what it does for any class that went away:
 * `classList.remove('')` — which is a `SyntaxError`, thrown from inside Lit's
 * update. The exception aborts `performUpdate`, so nothing after it runs: the
 * class list, the `item`, the `spacer` flag and the rendered content are frozen
 * at the values they had in the pass that poisoned them, FOREVER. Lit never
 * retries a failed update, and no later property change can dislodge it.
 *
 * What that looked like: a checkers game with all 24 pieces present in its state
 * and an empty board on screen, every one of the 64 hosts still `spacer` and
 * still `visibility: hidden`. Nothing logged anywhere a player would look —
 * the throw surfaces as an unhandled rejection from a microtask.
 *
 * Two tests, deliberately: the mechanism, at the level a fix has to be right at,
 * and the consequence, in the real game, because the mechanism test would still
 * pass if some other part of the chain broke the same board.
 */
test.describe('a token with an unnamed colour', () => {
  test('keeps updating after its colour is named', async ({ page }) => {
    await page.goto('/');

    const result = await page.evaluate(async () => {
      await import('/src/components/boardgame-token.ts');
      const el = document.createElement('boardgame-token') as any;
      // Exactly what a stack does with a component whose colour is hidden: the
      // property is set, to the empty string.
      el.color = '';
      el.type = 'disc';
      document.body.appendChild(el);
      await el.updateComplete;
      const before = el.renderRoot.querySelector('#outer').className;

      // ...and then the state arrives and the component becomes visible.
      let failure: string | null = null;
      el.color = 'Red';
      el.item = { ID: 'a-real-component' };
      try {
        await el.updateComplete;
        // The spacer flag is set from `updated()`, so it schedules a second
        // pass; wait for that one too or the assertion races it.
        await el.updateComplete;
      } catch (error) {
        failure = String(error);
      }
      const after = {
        className: el.renderRoot.querySelector('#outer').className,
        spacer: el.spacer,
        failure,
      };
      el.remove();
      return { before, after };
    });

    expect(result.before, 'an unnamed colour contributes no class').not.toContain('red');
    expect(result.after.failure, 'naming the colour must not throw out of the update').toBeNull();
    expect(result.after.className, 'the named colour must reach the class list').toContain('red');
    expect(result.after.spacer, 'a token with an item is not a spacer').toBe(false);
  });

  test('does not empty a checkers board', async ({ page }) => {
    test.setTimeout(120000);
    await createOfflineGame(page, 'checkers', { adminMode: false });
    await settleInitialLoad(page);

    const board = await page.evaluate(() => {
      const found: Element[] = [];
      const walk = (node: ParentNode) => {
        for (const element of node.querySelectorAll('boardgame-token')) found.push(element);
        for (const element of node.querySelectorAll('*')) {
          if (element.shadowRoot) walk(element.shadowRoot);
        }
      };
      walk(document);
      const tokens = found as unknown as { item: unknown; spacer: boolean }[];
      return {
        hosts: tokens.length,
        withItem: tokens.filter((token) => token.item).length,
        drawn: tokens.filter((token) => token.item && !token.spacer).length,
      };
    });

    // The state's own count, so this cannot pass by the board having no pieces.
    expect(board.withItem, 'checkers starts with 24 pieces on the board').toBe(24);
    expect(board.drawn, 'every piece the state has must actually be drawn').toBe(24);
    expect(board.hosts, 'and the empty squares are still hosted').toBeGreaterThanOrEqual(64);
  });
});
