import { test, expect } from '@playwright/test';
// (no live-game helpers needed: every probe mounts its own component)

/**
 * DASHED ATTRIBUTES THAT WERE NEVER OBSERVED.
 *
 * Lit derives a reactive property's observed attribute by LOWERCASING the
 * property name -- not by dash-casing it. `noDefaultSpacer` therefore observed
 * `nodefaultspacer`, `autoMessage` observed `automessage`, `isAgent` observed
 * `isagent`, and every dashed spelling this codebase actually writes fell on
 * the floor in silence. `stack-faux-components.spec.ts` covers the first
 * instance found; this covers the rest, and
 * `src/components/property-attribute-names.test.ts` covers the class so there
 * is no next one.
 *
 * Every assertion here sets the ATTRIBUTE, never the property. Setting the
 * property always worked, which is precisely why the existing unit fixtures
 * could not see any of this.
 */
test.describe('dashed attributes reach their properties', () => {
  /**
   * `no-default-spacer` is the instance that was live in production code
   * rather than in an example game: the animator, `boardgame-game-board` and
   * `boardgame-spatial-board` all write it, and none of the three was heard.
   *
   * The spacer hosts are counted by walking `#container`'s own children and
   * reading the `.spacer` PROPERTY, which is the measurement that does not
   * depend on the other half of this story: `spacer` used to be declared
   * without `reflect`, so the `[spacer]` attribute selector the stack's own
   * bookkeeping uses matched nothing, `haveSpacer` never fired, and the
   * control here accumulated THREE placeholder hosts rather than one. That was
   * a separate defect -- a missing `reflect`, not a wrong attribute name --
   * and it is fixed; `stack-spacer-reflect.spec.ts` owns it. The control's
   * count is asserted exactly now that it can be.
   */
  test('no-default-spacer suppresses the placeholder an empty stack would draw', async ({ page }) => {
    await page.goto('/');
    const result = await page.evaluate(async () => {
      const view = await import('/src/components/component-view.ts');
      await import('/src/components/boardgame-component-stack.ts');
      document.body.innerHTML = '';

      const spacerCount = (stack: any): number =>
        [...stack.shadowRoot.querySelectorAll('#container > [boardgame-component]')]
          .filter((el: any) => el.spacer === true).length;

      const build = async (suppress: boolean) => {
        const stack = document.createElement('boardgame-component-stack') as any;
        stack.setAttribute('layout', 'stack');
        // The spelling game-board, spatial-board and the animator all write.
        if (suppress) stack.setAttribute('no-default-spacer', '');
        // Configure BEFORE mounting: the spacer decision is made in
        // firstUpdated, and a stack with no componentView yet bails out of it.
        stack.componentView = view.tokenView({
          properties: () => ({ type: 'disc', color: 'red' }),
        });
        // An EMPTY stack: the spacer only exists to hold the empty slot open.
        stack.stack = { Deck: 'd', Size: 0, Components: [], IDs: [] };
        document.body.appendChild(stack);
        const settle = async () => {
          await stack.updateComplete;
          await new Promise<void>((r) => requestAnimationFrame(() => requestAnimationFrame(() => r())));
          await new Promise<void>((r) => setTimeout(r, 50));
        };
        await settle();
        const initial = spacerCount(stack);
        // A second empty state: re-evaluating an empty stack must not stack up
        // placeholders.
        stack.stack = { Deck: 'd', Size: 0, Components: [], IDs: [] };
        stack.fauxComponents = 0;
        (stack as any).requestUpdate();
        await settle();
        const out = {
          property: stack.noDefaultSpacer,
          initial,
          afterReevaluation: spacerCount(stack),
        };
        stack.remove();
        return out;
      };

      return { suppressed: await build(true), control: await build(false) };
    });

    // THE BUG: the attribute landed in the DOM and the property stayed false.
    expect(result.suppressed.property, 'the dashed attribute must reach the property').toBe(true);
    expect(result.suppressed.initial, 'so no placeholder is drawn').toBe(0);
    expect(result.suppressed.afterReevaluation, 'and none accumulates').toBe(0);
    // The control proves the assertions above are not vacuous: a stack that
    // does NOT ask for suppression still builds its placeholder, so 0 above
    // means the attribute did something.
    expect(result.control.property, 'the control must not be suppressed').toBe(false);
    expect(result.control.initial, 'and must still draw its placeholder -- exactly one')
      .toBe(1);
    expect(result.control.afterReevaluation, 'and still exactly one after re-evaluation')
      .toBe(1);
  });

  /**
   * The production witness for the same fix. `boardgame-spatial-board` writes
   * `no-default-spacer` on every token-overlay stack it renders, and a token
   * overlay is routinely EMPTY -- which is the only state in which the spacer
   * is ever built. (The other two writers turn out to be unreachable either
   * way: `boardgame-game-board` validates that its stack holds exactly
   * `cols * rows` components, so it can never be empty, and the animator's own
   * stack is never given a `componentView`, so its spacer branch bails out.
   * Their declarations are now honest rather than newly effective.)
   */
  test('a spatial board suppresses the spacer on an empty token overlay', async ({ page }) => {
    await page.goto('/');
    const result = await page.evaluate(async () => {
      const ensure = async (tag: string, path: string) => {
        if (!customElements.get(tag)) await import(/* @vite-ignore */ path);
      };
      const view = await import('/src/components/component-view.ts');
      await ensure('boardgame-spatial-board', '/src/components/boardgame-spatial-board.ts');
      document.body.innerHTML = '';

      const board = document.createElement('boardgame-spatial-board') as any;
      board.componentView = view.tokenView({ properties: () => ({ type: 'disc', color: 'red' }) });
      board.stacks = [{ Deck: 'd', Size: 0, Components: [], IDs: [] }];
      document.body.appendChild(board);
      await board.updateComplete;
      await new Promise<void>((r) => requestAnimationFrame(() => requestAnimationFrame(() => r())));
      await new Promise<void>((r) => setTimeout(r, 100));

      const stacks = [...board.shadowRoot.querySelectorAll('boardgame-component-stack')] as any[];
      const out = stacks.map((s) => ({
        suppressed: s.noDefaultSpacer,
        spacers: [...s.shadowRoot.querySelectorAll('#container > [boardgame-component]')]
          .filter((el: any) => el.spacer === true).length,
      }));
      board.remove();
      return out;
    });

    expect(result.length, 'the spatial board renders a token-overlay stack').toBeGreaterThan(0);
    for (const stack of result) {
      expect(stack.suppressed, 'the overlay stack must have heard no-default-spacer').toBe(true);
      expect(stack.spacers, 'and must draw no placeholder').toBe(0);
    }
  });

  /**
   * `auto-message` on `boardgame-fading-text`. debuganimations writes it on
   * its draw-stack callout and got the `'fixed'` default instead, so that
   * callout announced the static `message` ("Point Scored") every time the
   * stack count moved, rather than the signed delta it asked for.
   */
  test('auto-message selects the callout policy', async ({ page }) => {
    await page.goto('/');
    const result = await page.evaluate(async () => {
      await import('/src/components/boardgame-fading-text.ts');
      document.body.innerHTML = '';

      const build = async (dashed: boolean) => {
        const el = document.createElement('boardgame-fading-text') as any;
        if (dashed) el.setAttribute('auto-message', 'diff');
        el.trigger = 3;
        document.body.appendChild(el);
        await el.updateComplete;
        el.trigger = 5;
        await el.updateComplete;
        const out = { policy: el.autoMessage, message: el.message };
        el.remove();
        return out;
      };

      return { dashed: await build(true), control: await build(false) };
    });

    expect(result.dashed.policy, 'the dashed attribute must reach the property').toBe('diff');
    expect(result.dashed.message, '3 -> 5 is a +2 callout').toBe('+2');
    expect(result.control.policy, 'the control keeps the default policy').toBe('fixed');
    expect(result.control.message, 'and so never rewrites its message').toBe('Point Scored');
  });

  /**
   * The roster and lobby boolean-attribute bindings. `?is-agent`,
   * `?is-empty`, `?game-open`, `?game-visible` and `?is-owner` are all
   * ATTRIBUTE bindings written by `boardgame-player-roster`,
   * `boardgame-player-roster-item` and `boardgame-game-item`, and none of
   * them were observed.
   */
  test('the roster and lobby boolean attributes reach their properties', async ({ page }) => {
    await page.goto('/');
    const result = await page.evaluate(async () => {
      const ensure = async (tag: string, path: string) => {
        if (!customElements.get(tag)) await import(/* @vite-ignore */ path);
      };
      await ensure('boardgame-player-chip', '/src/components/boardgame-player-chip.ts');
      await ensure('boardgame-player-roster-item', '/src/components/boardgame-player-roster-item.ts');
      await ensure('boardgame-configure-game-properties',
        '/src/components/boardgame-configure-game-properties.ts');
      document.body.innerHTML = '';

      const mount = async (tag: string, attributes: Record<string, string>) => {
        const el = document.createElement(tag) as any;
        for (const [name, value] of Object.entries(attributes)) el.setAttribute(name, value);
        document.body.appendChild(el);
        await el.updateComplete;
        return el;
      };

      const chip = await mount('boardgame-player-chip', { 'is-agent': '' });
      // The avatar the chip actually draws, which is the visible consequence.
      const chipSrc = chip.shadowRoot.querySelector('img')?.getAttribute('src') ?? '';

      const item = await mount('boardgame-player-roster-item', { 'is-empty': '' });
      const emptyDescription = item.shadowRoot.textContent.replace(/\s+/g, ' ').trim();

      const agentItem = await mount('boardgame-player-roster-item', { 'is-agent': '' });
      const agentDescription = agentItem.shadowRoot.textContent.replace(/\s+/g, ' ').trim();

      const properties = await mount('boardgame-configure-game-properties', {
        'game-open': '', 'game-visible': '', 'is-owner': '',
      });
      const icons = [...properties.shadowRoot.querySelectorAll('md-icon')]
        .map((i: any) => i.textContent.trim());

      const out = {
        chipIsAgent: chip.isAgent,
        chipSrc,
        itemIsEmpty: item.isEmpty,
        emptyDescription,
        agentItemIsAgent: agentItem.isAgent,
        agentDescription,
        gameOpen: properties.gameOpen,
        gameVisible: properties.gameVisible,
        isOwner: properties.isOwner,
        icons,
      };
      document.body.innerHTML = '';
      return out;
    });

    expect(result.chipIsAgent, 'boardgame-player-chip observes is-agent').toBe(true);
    expect(result.chipSrc, 'so an agent draws the robot avatar').toContain('agent.svg');

    expect(result.itemIsEmpty, 'boardgame-player-roster-item observes is-empty').toBe(true);
    expect(result.emptyDescription, 'an empty seat is described as such').toContain('No one');

    expect(result.agentItemIsAgent, 'boardgame-player-roster-item observes is-agent').toBe(true);
    expect(result.agentDescription, 'an agent seat is described as such').toContain('Robot');

    expect(result.gameOpen, 'boardgame-configure-game-properties observes game-open').toBe(true);
    expect(result.gameVisible, 'and game-visible').toBe(true);
    expect(result.isOwner, 'and is-owner').toBe(true);
    // The visible consequence: the two toggles show the OPEN and VISIBLE
    // icons. Before the fix both properties were pinned false, so the header
    // always showed "unlisted, invite-only" whatever the game really was --
    // and clicking either toggle submitted from those wrong values.
    expect(result.icons, 'the toggles must render the open/visible state')
      .toEqual(['people', 'visibility']);
  });
});
