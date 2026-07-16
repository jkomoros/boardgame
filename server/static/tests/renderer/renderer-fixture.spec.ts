import AxeBuilder from '@axe-core/playwright';
import { expect, test } from '@playwright/test';
import {
  RENDERER_VIEWPORTS,
  focusWithKeyboard,
  prepareRendererFixturePage,
  retryRendererEvaluation,
} from './renderer-fixture-helpers.js';

test('Pig fixture installs a typed snapshot and correlates zero-input proposals', async ({ page }) => {
  const diagnostics = await prepareRendererFixturePage(page);
  try {
    const result = await page.evaluate(async () => {
      await import('/game-src/pig/boardgame-render-game-pig.ts');
      const { pigRendererFixture } = await import(
        '/game-src/pig/boardgame-render-fixtures-pig.ts'
      );
      const { mountRendererFixture } = await import('/src/testing/renderer-fixture.ts');
      const handle = await mountRendererFixture(pigRendererFixture);
      const actionButton = handle.renderer.shadowRoot?.querySelector('boardgame-action-button');
      const button = actionButton?.shadowRoot?.querySelector('button');
      if (!(button instanceof HTMLButtonElement)) throw new Error('Pig fixture did not render Done');
      button.click();
      await Promise.resolve();
      await handle.update({ ...pigRendererFixture.snapshot, version: 4 });
      button.click();
      await Promise.resolve();
      await handle.update({ ...pigRendererFixture.snapshot, version: 5 });
      (globalThis as unknown as { __pigFixtureHandle: typeof handle }).__pigFixtureHandle = handle;
      return {
        host: { ...handle.host.dataset },
        proposals: handle.proposals,
      };
    });

    expect(result.host).toMatchObject({
      fixtureSchemaVersion: '1',
      fixtureVersion: '5',
      fixtureSurface: 'game',
    });
    expect(result.proposals).toEqual([
      {
        requestID: 'fixture-v3-move-1',
        snapshotVersion: 3,
        name: 'Done Turn',
        arguments: {},
      },
      {
        requestID: 'fixture-v4-move-2',
        snapshotVersion: 4,
        name: 'Done Turn',
        arguments: {},
      },
    ]);

    const pending = await page.evaluate(async () => {
      const { pigRendererFixture } = await import('/game-src/pig/boardgame-render-fixtures-pig.ts');
      const handle = (globalThis as unknown as {
        __pigFixtureHandle: {
          readonly renderer: HTMLElement & {
            moveTransport: unknown;
            updateComplete: Promise<unknown>;
          };
          update(snapshot: typeof pigRendererFixture.snapshot): Promise<void>;
        };
      }).__pigFixtureHandle;
      const requests: unknown[] = [];
      let finish: (() => void) | undefined;
      handle.renderer.moveTransport = {
        submit: (request: unknown) => new Promise(resolve => {
          requests.push(request);
          finish = () => resolve({ kind: 'success' });
        }),
      };
      await handle.update({ ...pigRendererFixture.snapshot, version: 6 });
      const actionButton = handle.renderer.shadowRoot?.querySelector('boardgame-action-button') as (
        HTMLElement & { updateComplete: Promise<unknown> }
      ) | null;
      const button = actionButton?.shadowRoot?.querySelector('button');
      if (!(button instanceof HTMLButtonElement) || !actionButton) {
        throw new Error('Pig action button was unavailable');
      }
      button.click();
      button.click();
      await actionButton.updateComplete;
      const bounds = button.getBoundingClientRect();
      const during = {
        requestCount: requests.length,
        disabled: button.disabled,
        busy: button.getAttribute('aria-busy'),
        width: bounds.width,
        height: bounds.height,
      };
      finish?.();
      await Promise.resolve();
      await actionButton.updateComplete;
      await handle.update({ ...pigRendererFixture.snapshot, version: 7 });
      return during;
    });
    expect(pending).toMatchObject({ requestCount: 1, disabled: true, busy: 'true' });
    expect(pending.width).toBeGreaterThanOrEqual(44);
    expect(pending.height).toBeGreaterThanOrEqual(44);

    const host = page.locator('[data-renderer-fixture]');
    const axeResult = await new AxeBuilder({ page })
      .include('[data-renderer-fixture]')
      .withRules(['aria-required-children', 'aria-required-parent', 'button-name'])
      .analyze();
    expect(axeResult.violations).toEqual([]);
    await focusWithKeyboard(page, host.locator('boardgame-action-button button'));
    const disposal = await page.evaluate(() => {
      const handle = (globalThis as unknown as {
        __pigFixtureHandle: {
          readonly host: HTMLElement;
          readonly renderer: HTMLElement;
          readonly proposals: readonly unknown[];
          dispose(): void;
        };
      }).__pigFixtureHandle;
      const proposalsBefore = handle.proposals.length;
      handle.dispose();
      handle.renderer.dispatchEvent(new CustomEvent('propose-move', {
        detail: { name: 'Done Turn', arguments: {} },
      }));
      return {
        hostConnected: handle.host.isConnected,
        proposalsBefore,
        proposalsAfter: handle.proposals.length,
      };
    });
    expect(disposal).toEqual({ hostConnected: false, proposalsBefore: 2, proposalsAfter: 2 });
    diagnostics.assertEmpty();
  } finally {
    diagnostics.stop();
  }
});

test('component stacks bind typed actions by slot and reject ambiguous wiring', async ({ page }) => {
  const diagnostics = await prepareRendererFixturePage(page);
  try {
    const result = await page.evaluate(async () => {
      await import('/src/components/boardgame-card.ts');
      await import('/src/components/boardgame-component-stack.ts');
      await import('/game-src/pig/boardgame-render-game-pig.ts');
      const { pigRendererFixture } = await import('/game-src/pig/boardgame-render-fixtures-pig.ts');
      const { mountRendererFixture } = await import('/src/testing/renderer-fixture.ts');
      const { html } = await import('/src/client.ts');
      const PigRenderer = customElements.get('boardgame-render-game-pig');
      if (!PigRenderer) throw new Error('Pig renderer was not registered');

      class ComponentActionsRenderer extends PigRenderer {
        override render() {
          const renderer = this as unknown as { move(name: string): object };
          const done = renderer.move('Done Turn');
          return html`
            <boardgame-component-stack .componentActions=${[done, null]}>
              <boardgame-card
                boardgame-component
                .item=${{ ID: 'first', Values: { Rank: 'A' } }}>
              </boardgame-card>
              <boardgame-card
                boardgame-component
                .item=${{ ID: 'second', Values: { Rank: 'K' } }}>
              </boardgame-card>
            </boardgame-component-stack>`;
        }
      }
      customElements.define('boardgame-render-game-pig-component-actions', ComponentActionsRenderer);
      const handle = await mountRendererFixture({
        ...pigRendererFixture,
        tagName: 'boardgame-render-game-pig-component-actions',
      } as never);
      const stack = handle.renderer.shadowRoot?.querySelector('boardgame-component-stack') as (
        HTMLElement & {
          componentActions: readonly unknown[];
          updateComplete: Promise<unknown>;
        }
      ) | null;
      if (!stack) throw new Error('Typed component stack did not render');
      await stack.updateComplete;
      const cards = [...stack.querySelectorAll('boardgame-card')];
      if (cards.length !== 2) throw new Error(`Expected two cards, received ${cards.length}`);
      const initial = cards.map(card => ({
        disabled: (card as HTMLElement & { disabled?: boolean }).disabled,
        ariaDisabled: card.getAttribute('aria-disabled'),
        role: card.getAttribute('role'),
        tabIndex: (card as HTMLElement).tabIndex,
      }));
      (cards[1].shadowRoot?.querySelector('#outer') as HTMLElement | null)?.click();
      await Promise.resolve();
      const proposalsAfterNullSlot = handle.proposals.length;
      (cards[0].shadowRoot?.querySelector('#outer') as HTMLElement | null)?.click();
      await Promise.resolve();
      await handle.update({ ...pigRendererFixture.snapshot, version: 4 });
      const refreshedCard = handle.renderer.shadowRoot
        ?.querySelector('boardgame-component-stack boardgame-card') as HTMLElement | null;
      if (!refreshedCard) throw new Error('Typed component stack lost its first card after refresh');
      const enter = new KeyboardEvent('keydown', {
        key: 'Enter', bubbles: true, composed: true, cancelable: true,
      });
      const keyboardDefaultPrevented = !refreshedCard.dispatchEvent(enter);
      await Promise.resolve();
      const activeActions = stack.componentActions;
      stack.componentActions = [];
      await stack.updateComplete;
      const released = [...stack.querySelectorAll('boardgame-card')].map(card => ({
        disabled: (card as HTMLElement & { disabled?: boolean }).disabled,
        ariaDisabled: card.getAttribute('aria-disabled'),
        role: card.getAttribute('role'),
        tabIndex: (card as HTMLElement).getAttribute('tabindex'),
      }));
      stack.componentActions = activeActions;
      await stack.updateComplete;
      (globalThis as unknown as { __componentActionsTest: { handle: typeof handle; stack: typeof stack } })
        .__componentActionsTest = { handle, stack };
      return { initial, released, proposalsAfterNullSlot, keyboardDefaultPrevented, proposals: handle.proposals };
    });

    expect(result.initial).toEqual([
      { disabled: false, ariaDisabled: 'false', role: 'button', tabIndex: 0 },
      { disabled: true, ariaDisabled: 'true', role: null, tabIndex: -1 },
    ]);
    expect(result.released).toEqual([
      { disabled: false, ariaDisabled: null, role: null, tabIndex: null },
      { disabled: false, ariaDisabled: null, role: null, tabIndex: null },
    ]);
    expect(result.proposalsAfterNullSlot).toBe(0);
    expect(result.keyboardDefaultPrevented).toBe(true);
    expect(result.proposals).toHaveLength(2);
    expect(result.proposals[0]).toMatchObject({ name: 'Done Turn', arguments: {} });
    expect(result.proposals[1]).toMatchObject({ snapshotVersion: 4, name: 'Done Turn', arguments: {} });
    diagnostics.assertEmpty();
    diagnostics.stop();

    const cardinalityError = await page.evaluate(async () => {
      const testState = (globalThis as unknown as {
        __componentActionsTest: {
          handle: { dispose(): void };
          stack: { componentActions: readonly unknown[]; updateComplete: Promise<unknown> };
        };
      }).__componentActionsTest;
      testState.stack.componentActions = [testState.stack.componentActions[0]];
      try {
        await testState.stack.updateComplete;
        return '<resolved>';
      } catch (error) {
        return error instanceof Error ? error.message : String(error);
      } finally {
        testState.handle.dispose();
      }
    });
    expect(cardinalityError).toContain('componentActions has 1 entries');
  } finally {
    diagnostics.stop();
  }
});

test('renderer-scoped component views preserve hosts and distinguish visible, hidden, and empty slots', async ({ page }) => {
  const diagnostics = await prepareRendererFixturePage(page);
  try {
    const result = await page.evaluate(async () => {
      await import('/src/components/boardgame-component-stack.ts');
      const { cardView, componentView, html } = await import('/src/client.ts');
      const { BoardgameComponent } = await import('/src/components/boardgame-component.ts');
      const visible = (id: string, rank: string) => ({
        Index: 0, Values: { rank }, Deck: 'cards', GameName: 'view-test', ID: id,
      });
      const makeStack = (components: readonly unknown[], ids: readonly string[]) => ({
        Deck: 'cards', Indexes: components.map((_item, index) => index), IDs: ids,
        IDsLastSeen: {}, ShuffleCount: 0, Size: components.length,
        GameName: 'view-test', Components: components,
      });
      const view = cardView({
        render: (context: { kind: string; component: { Values?: { rank?: string } } | null }) =>
          context.kind === 'visible' ? html`<span class="rank">${context.component?.Values?.rank}</span>` : null,
        properties: (context: { kind: string }) => context.kind === 'visible' ? { rotated: true } : {},
      });
      const stack = document.createElement('boardgame-component-stack') as HTMLElement & {
        stack: unknown;
        componentView: unknown;
        componentsDisabled: boolean;
        fauxComponents: number;
        newComponent(): unknown;
        updateComplete: Promise<unknown>;
      };
      // Match normal Lit source order: .stack is commonly committed first.
      stack.stack = makeStack([visible('a', 'A'), {}, null], ['a', 'hidden', '']);
      stack.componentView = view;
      document.body.append(stack);
      await stack.updateComplete;
      const firstCards = [...stack.querySelectorAll('boardgame-card')] as Array<HTMLElement & {
        rotated: boolean;
        spacer: boolean;
        item: unknown;
        updateComplete: Promise<unknown>;
      }>;
      await Promise.all(firstCards.map(card => card.updateComplete));
      const before = firstCards.map(card => ({
        text: card.textContent?.trim() ?? '',
        rotated: card.rotated,
        spacer: card.spacer,
      }));

      stack.stack = makeStack([{}, visible('b', 'K'), null], ['hidden', 'b', '']);
      await stack.updateComplete;
      const secondCards = [...stack.querySelectorAll('boardgame-card')] as typeof firstCards;
      await Promise.all(secondCards.map(card => card.updateComplete));
      const after = secondCards.map(card => ({
        text: card.textContent?.trim() ?? '',
        rotated: card.rotated,
        spacer: card.spacer,
      }));
      const stableHosts = firstCards.every((card, index) => card === secondCards[index]);

      stack.fauxComponents = 4;
      await stack.updateComplete;
      const fauxBefore = stack.shadowRoot?.querySelector('#faux-components boardgame-card') as
        (HTMLElement & { rotated: boolean; disabled: boolean; updateComplete: Promise<unknown> }) | null;
      await fauxBefore?.updateComplete;

      stack.componentView = view.withProperties({ rotated: false });
      await stack.updateComplete;
      const reboundCards = [...stack.querySelectorAll('boardgame-card')] as typeof firstCards;
      await Promise.all(reboundCards.map(card => card.updateComplete));
      const reboundKeepsHosts = secondCards.every((card, index) => card === reboundCards[index]);
      const reboundRotated = reboundCards.map(card => card.rotated);
      const fauxAfter = stack.shadowRoot?.querySelector('#faux-components boardgame-card') as
        (HTMLElement & { rotated: boolean; disabled: boolean; updateComplete: Promise<unknown> }) | null;
      await fauxAfter?.updateComplete;
      const fauxBinding = {
        created: Boolean(fauxBefore),
        stableHost: fauxBefore === fauxAfter,
        rotated: fauxAfter?.rotated,
      };
      stack.componentsDisabled = true;
      await stack.updateComplete;
      const displayOnlyDisabled = reboundCards.map(card => card.disabled);
      const fauxDisplayOnlyDisabled = fauxAfter?.disabled;

      class FixturePiece extends BoardgameComponent {
        label = '';
      }
      customElements.define('boardgame-fixture-piece', FixturePiece);
      const customView = componentView(
        () => document.createElement('boardgame-fixture-piece'),
        {
          render: (context: { kind: string; component: { DynamicValues?: { marked?: boolean } } | null }) =>
            context.kind === 'visible' && context.component?.DynamicValues?.marked
              ? html`<strong>Marked custom piece</strong>`
              : null,
        },
      );
      const customStack = document.createElement('boardgame-component-stack') as typeof stack;
      customStack.componentView = customView;
      customStack.stack = makeStack([{
        ...visible('custom', 'Q'),
        DynamicValues: { marked: true },
      }], ['custom']);
      document.body.append(customStack);
      await customStack.updateComplete;
      const customText = customStack.textContent?.trim() ?? '';

      const missingViewStack = document.createElement('boardgame-component-stack') as typeof stack;
      let missingViewError = '';
      try {
        missingViewStack.newComponent();
      } catch (error) {
        missingViewError = error instanceof Error ? error.message : String(error);
      }

      stack.remove();
      customStack.remove();
      return {
        before, after, stableHosts, reboundKeepsHosts, reboundRotated,
        fauxBinding, displayOnlyDisabled, fauxDisplayOnlyDisabled, customText, missingViewError,
      };
    });

    expect(result.before).toEqual([
      { text: 'A', rotated: true, spacer: false },
      { text: '', rotated: false, spacer: false },
      { text: '', rotated: false, spacer: true },
    ]);
    expect(result.after).toEqual([
      { text: '', rotated: false, spacer: false },
      { text: 'K', rotated: true, spacer: false },
      { text: '', rotated: false, spacer: true },
    ]);
    expect(result.stableHosts).toBe(true);
    expect(result.reboundKeepsHosts).toBe(true);
    expect(result.reboundRotated).toEqual([false, false, false]);
    expect(result.fauxBinding).toEqual({ created: true, stableHost: true, rotated: false });
    expect(result.displayOnlyDisabled).toEqual([true, true, true]);
    expect(result.fauxDisplayOnlyDisabled).toBe(true);
    expect(result.customText).toBe('Marked custom piece');
    expect(result.missingViewError).toContain('set .componentView');
    diagnostics.assertEmpty();
  } finally {
    diagnostics.stop();
  }
});

test('status text uses typed values, accessible live updates, and loud authoring errors', async ({ page }) => {
  const diagnostics = await prepareRendererFixturePage(page);
  try {
    const result = await page.evaluate(async () => {
      await import('/src/components/boardgame-status-text.ts');
      const status = document.createElement('boardgame-status-text');
      status.value = 3;
      document.body.append(status);
      await status.updateComplete;
      const first = status.shadowRoot?.querySelector('strong');
      const initial = {
        text: first?.textContent,
        role: first?.getAttribute('role'),
        live: first?.getAttribute('aria-live'),
        atomic: first?.getAttribute('aria-atomic'),
      };
      status.value = 7;
      await status.updateComplete;
      const fading = status.shadowRoot?.querySelector('boardgame-fading-text') as
        (HTMLElement & { trigger: unknown }) | null;
      const updated = {
        text: status.shadowRoot?.querySelector('strong')?.textContent,
        fadingTrigger: fading?.trigger,
        fadingHidden: fading?.getAttribute('aria-hidden'),
      };

      const renderError = (configure: (element: HTMLElement & { value?: unknown }) => void) => {
        const element = document.createElement('boardgame-status-text') as HTMLElement & {
          value?: unknown;
          render(): unknown;
        };
        configure(element);
        try {
          element.render();
          return '<resolved>';
        } catch (error) {
          return error instanceof Error ? error.message : String(error);
        }
      };
      const invalidValue = renderError(element => { element.value = { score: 7 }; });
      const legacyContent = renderError(element => { element.textContent = '7'; });
      const legacyAttribute = renderError(element => { element.setAttribute('message', '7'); });
      status.remove();
      return { initial, updated, invalidValue, legacyContent, legacyAttribute };
    });

    expect(result.initial).toEqual({ text: '3', role: 'status', live: 'polite', atomic: 'true' });
    expect(result.updated).toEqual({ text: '7', fadingTrigger: 7, fadingHidden: 'true' });
    expect(result.invalidValue).toContain('.value must be a string, number, null, or undefined');
    expect(result.legacyContent).toContain('slotted content is not supported');
    expect(result.legacyAttribute).toContain('value/message attributes are not supported');
    diagnostics.assertEmpty();
  } finally {
    diagnostics.stop();
  }
});

test('fading text handles typed scalar policies, decimal diffs, restarts, and invalid configuration', async ({ page }) => {
  const diagnostics = await prepareRendererFixturePage(page);
  try {
    // Animation restart is the behavior under test; the rest of this shard uses reduced motion.
    await page.emulateMedia({ reducedMotion: 'no-preference' });
    const result = await page.evaluate(async () => {
      await import('/src/components/boardgame-fading-text.ts');
      const fading = document.createElement('boardgame-fading-text');
      fading.style.setProperty('--animation-length', '10s');
      fading.autoMessage = 'diff';
      fading.trigger = 1.25;
      document.body.append(fading);
      await fading.updateComplete;
      fading.trigger = 2.75;
      await fading.updateComplete;
      await new Promise<void>(resolve => requestAnimationFrame(() => requestAnimationFrame(() => resolve())));
      await fading.updateComplete;
      const container = fading.shadowRoot?.querySelector('#container');
      const initial = {
        message: fading.shadowRoot?.querySelector('#message')?.textContent?.trim(),
        animating: container?.classList.contains('animating'),
        role: container?.getAttribute('role'),
        live: container?.getAttribute('aria-live'),
      };

      fading.trigger = 4;
      await fading.updateComplete;
      fading.trigger = 5;
      await fading.updateComplete;
      await new Promise<void>(resolve => requestAnimationFrame(() => requestAnimationFrame(() => resolve())));
      await fading.updateComplete;
      const restarted = {
        message: fading.shadowRoot?.querySelector('#message')?.textContent?.trim(),
        animating: fading.shadowRoot?.querySelector('#container')?.classList.contains('animating'),
      };

      const renderError = (configure: (element: HTMLElement & Record<string, unknown>) => void) => {
        const element = document.createElement('boardgame-fading-text') as HTMLElement &
          Record<string, unknown> & { render(): unknown };
        configure(element);
        try {
          element.render();
          return '<resolved>';
        } catch (error) {
          return error instanceof Error ? error.message : String(error);
        }
      };
      const invalidMessage = renderError(element => { element.message = 7; });
      const invalidTrigger = renderError(element => { element.trigger = Number.POSITIVE_INFINITY; });
      const invalidMessagePolicy = renderError(element => { element.autoMessage = 'difference'; });
      const invalidSuppressPolicy = renderError(element => { element.suppress = 'empty'; });
      fading.remove();
      return { initial, restarted, invalidMessage, invalidTrigger, invalidMessagePolicy, invalidSuppressPolicy };
    });

    expect(result.initial).toEqual({ message: '+1.5', animating: true, role: 'status', live: 'polite' });
    expect(result.restarted).toEqual({ message: '+1', animating: true });
    expect(result.invalidMessage).toContain('.message must be a string');
    expect(result.invalidTrigger).toContain('.trigger numbers must be finite');
    expect(result.invalidMessagePolicy).toContain('unknown autoMessage');
    expect(result.invalidSuppressPolicy).toContain('unknown suppress policy');
    diagnostics.assertEmpty();
  } finally {
    diagnostics.stop();
  }
});

test('action button provides accessible names, pending feedback, styling parts, and loud wiring errors', async ({ page }) => {
  const diagnostics = await prepareRendererFixturePage(page);
  try {
    const result = await page.evaluate(async () => {
      await import('/game-src/pig/boardgame-render-game-pig.ts');
      const { pigRendererFixture } = await import('/game-src/pig/boardgame-render-fixtures-pig.ts');
      const { mountRendererFixture } = await import('/src/testing/renderer-fixture.ts');
      const handle = await mountRendererFixture(pigRendererFixture);
      const actionButton = handle.renderer.shadowRoot?.querySelector('boardgame-action-button') as
        (HTMLElement & { action: unknown; updateComplete: Promise<unknown> }) | null;
      if (!actionButton) throw new Error('Pig did not render its action button');
      await actionButton.updateComplete;
      const native = actionButton.shadowRoot?.querySelector('button');
      const ordinary = {
        text: actionButton.textContent?.trim(),
        buttonPart: native?.getAttribute('part'),
        labelPart: actionButton.shadowRoot?.querySelector('[part="label"]')?.getAttribute('part'),
      };

      let resolveSubmission: ((result: { readonly kind: 'success' }) => void) | undefined;
      (handle.renderer as unknown as { moveTransport: unknown }).moveTransport = {
        submit: () => new Promise(resolve => { resolveSubmission = resolve; }),
      };
      native?.click();
      await actionButton.updateComplete;
      const pending = {
        busy: native?.getAttribute('aria-busy'),
        spinnerHidden: (actionButton.shadowRoot?.querySelector('[part="spinner"]') as HTMLElement | null)?.hidden,
      };
      resolveSubmission?.({ kind: 'success' });
      await Promise.resolve();
      await actionButton.updateComplete;
      const settled = {
        busy: native?.getAttribute('aria-busy'),
        spinnerHidden: (actionButton.shadowRoot?.querySelector('[part="spinner"]') as HTMLElement | null)?.hidden,
      };

      const icon = document.createElement('boardgame-action-button');
      icon.label = 'Draw a card';
      icon.action = actionButton.action as never;
      icon.append(document.createElement('svg'));
      document.body.append(icon);
      await icon.updateComplete;
      const iconLabel = icon.shadowRoot?.querySelector('button')?.getAttribute('aria-label');

      const renderError = (configure: (element: HTMLElement & Record<string, unknown>) => void) => {
        const element = document.createElement('boardgame-action-button') as HTMLElement &
          Record<string, unknown> & { render(): unknown };
        configure(element);
        try {
          element.render();
          return '<resolved>';
        } catch (error) {
          return error instanceof Error ? error.message : String(error);
        }
      };
      const emptyName = renderError(() => {});
      const invalidAction = renderError(element => {
        element.textContent = 'Move';
        element.action = {};
      });
      icon.remove();
      handle.dispose();
      return { ordinary, pending, settled, iconLabel, emptyName, invalidAction };
    });

    expect(result.ordinary).toEqual({ text: 'Done', buttonPart: 'button', labelPart: 'label' });
    expect(result.pending).toEqual({ busy: 'true', spinnerHidden: false });
    expect(result.settled).toEqual({ busy: 'false', spinnerHidden: true });
    expect(result.iconLabel).toBe('Draw a card');
    expect(result.emptyName).toContain('provide visible text or a non-empty label');
    expect(result.invalidAction).toContain('.action must be a bound move action');
    diagnostics.assertEmpty();
  } finally {
    diagnostics.stop();
  }
});

test('action bar supplies named responsive grouping, styling hooks, and closed layout policies', async ({ page }) => {
  const diagnostics = await prepareRendererFixturePage(page);
  try {
    const result = await page.evaluate(async () => {
      await import('/src/components/boardgame-action-bar.ts');
      await import('/src/components/boardgame-action-button.ts');
      const bar = document.createElement('boardgame-action-bar');
      bar.style.width = '50rem';
      for (const text of ['Draw', 'Pass']) {
        const action = document.createElement('boardgame-action-button');
        action.textContent = text;
        bar.append(action);
      }
      document.body.append(bar);
      await bar.updateComplete;
      const inner = bar.shadowRoot?.querySelector('#bar') as HTMLElement | null;
      const wide = {
        role: inner?.getAttribute('role'),
        label: inner?.getAttribute('aria-label'),
        part: inner?.getAttribute('part'),
        direction: inner ? getComputedStyle(inner).flexDirection : null,
      };

      bar.style.width = '20rem';
      await new Promise<void>(resolve => requestAnimationFrame(() => resolve()));
      const firstAction = bar.querySelector('boardgame-action-button');
      const narrow = {
        direction: inner ? getComputedStyle(inner).flexDirection : null,
        actionWidth: firstAction
          ? getComputedStyle(firstAction).getPropertyValue('--boardgame-action-width').trim()
          : null,
      };

      bar.style.width = '50rem';
      bar.orientation = 'vertical';
      bar.alignment = 'start';
      await bar.updateComplete;
      const vertical = {
        direction: inner ? getComputedStyle(inner).flexDirection : null,
        alignment: inner ? getComputedStyle(inner).alignItems : null,
      };

      const renderError = (configure: (element: HTMLElement & Record<string, unknown>) => void) => {
        const element = document.createElement('boardgame-action-bar') as HTMLElement &
          Record<string, unknown> & { render(): unknown };
        configure(element);
        try {
          element.render();
          return '<resolved>';
        } catch (error) {
          return error instanceof Error ? error.message : String(error);
        }
      };
      const emptyLabel = renderError(element => { element.label = '  '; });
      const invalidOrientation = renderError(element => { element.orientation = 'diagonal'; });
      const invalidAlignment = renderError(element => { element.alignment = 'around'; });
      bar.remove();
      return { wide, narrow, vertical, emptyLabel, invalidOrientation, invalidAlignment };
    });

    expect(result.wide).toEqual({ role: 'group', label: 'Game actions', part: 'bar', direction: 'row' });
    expect(result.narrow).toEqual({ direction: 'column', actionWidth: '100%' });
    expect(result.vertical).toEqual({ direction: 'column', alignment: 'flex-start' });
    expect(result.emptyLabel).toContain('label must be a non-empty accessible group name');
    expect(result.invalidOrientation).toContain('unknown orientation');
    expect(result.invalidAlignment).toContain('unknown alignment');
    diagnostics.assertEmpty();
  } finally {
    diagnostics.stop();
  }
});

test('component stack exposes its real closed layout contract and rejects invalid geometry', async ({ page }) => {
  const diagnostics = await prepareRendererFixturePage(page);
  try {
    const result = await page.evaluate(async () => {
      const { isStackLayout } = await import('/src/client.ts');
      const valid = document.createElement('boardgame-component-stack');
      valid.layout = 'fan';
      document.body.append(valid);
      await valid.updateComplete;
      const container = valid.shadowRoot?.querySelector('#container');

      const renderError = (name: string, value: unknown) => {
        const element = document.createElement('boardgame-component-stack') as HTMLElement &
          Record<string, unknown> & { render(): unknown };
        element[name] = value;
        try {
          element.render();
          return '<resolved>';
        } catch (error) {
          return error instanceof Error ? error.message : String(error);
        }
      };

      const output = {
        className: container?.getAttribute('class'),
        guard: ['stack', 'grid', 'fan', 'pile', 'spread', 'board', 'spatial', 'carousel']
          .map(value => [value, isStackLayout(value)]),
        invalidLayout: renderError('layout', 'carousel'),
        invalidColumns: renderError('boardCols', 0),
        invalidRows: renderError('boardRows', 1.5),
        invalidFaux: renderError('fauxComponents', -1),
        invalidStagger: renderError('stagger', Number.NaN),
        invalidPosition: renderError('spatialPositions', [{ top: 10, left: Number.POSITIVE_INFINITY }]),
      };
      valid.remove();
      return output;
    });

    expect(result.className).toBe('fan');
    expect(result.guard).toEqual([
      ['stack', true], ['grid', true], ['fan', true], ['pile', true],
      ['spread', true], ['board', true], ['spatial', true], ['carousel', false],
    ]);
    expect(result.invalidLayout).toContain('unknown layout');
    expect(result.invalidColumns).toContain('boardCols must be a positive safe integer');
    expect(result.invalidRows).toContain('boardRows must be a positive safe integer');
    expect(result.invalidFaux).toContain('fauxComponents must be a nonnegative safe integer');
    expect(result.invalidStagger).toContain('stagger must be a finite nonnegative number');
    expect(result.invalidPosition).toContain('spatialPositions[0] must be null or finite');
    diagnostics.assertEmpty();
  } finally {
    diagnostics.stop();
  }
});

test('component zone makes a named stack region, count, and empty state the default', async ({ page }) => {
  const diagnostics = await prepareRendererFixturePage(page);
  try {
    const result = await page.evaluate(async () => {
      await import('/src/client.ts');
      const zone = document.createElement('boardgame-component-zone');
      zone.label = 'Draw pile';
      zone.layout = 'fan';
      zone.headingLevel = 3;
      zone.stack = null;
      zone.messy = true;
      const headingAction = document.createElement('button');
      headingAction.slot = 'heading-actions';
      headingAction.textContent = 'Inspect';
      zone.append(headingAction);
      document.body.append(zone);
      await zone.updateComplete;
      const stack = zone.shadowRoot?.querySelector('boardgame-component-stack');
      if (!stack) throw new Error('component zone did not render its stack');
      await stack.updateComplete;
      const section = zone.shadowRoot?.querySelector('section');
      const heading = zone.shadowRoot?.querySelector('#label');
      const count = zone.shadowRoot?.querySelector('#count');
      const empty = zone.shadowRoot?.querySelector('#empty');
      const headingSlot = zone.shadowRoot?.querySelector('slot[name="heading-actions"]') as HTMLSlotElement | null;
      const defaults = {
        sectionLabel: section?.getAttribute('aria-labelledby'),
        zonePart: section?.getAttribute('part'),
        heading: heading?.textContent,
        headingRole: heading?.getAttribute('role'),
        headingLevel: heading?.getAttribute('aria-level'),
        count: count?.textContent,
        countLabel: count?.getAttribute('aria-label'),
        empty: empty?.textContent,
        assignedHeadingActions: headingSlot?.assignedElements().length,
        stack: {
          layout: stack.layout,
          messy: stack.messy,
          disabled: stack.componentsDisabled,
          part: stack.getAttribute('part'),
        },
      };

      zone.hideCount = true;
      zone.hideEmptyState = true;
      await zone.updateComplete;
      const suppressed = {
        count: zone.shadowRoot?.querySelector('#count') !== null,
        empty: zone.shadowRoot?.querySelector('#empty') !== null,
      };

      const renderError = (name: string, value: unknown) => {
        const element = document.createElement('boardgame-component-zone') as HTMLElement &
          Record<string, unknown> & { render(): unknown };
        element.label = 'Zone';
        element[name] = value;
        try {
          element.render();
          return '<resolved>';
        } catch (error) {
          return error instanceof Error ? error.message : String(error);
        }
      };
      const blankLabel = renderError('label', '  ');
      const invalidHeading = renderError('headingLevel', 7);
      const boardLayout = renderError('layout', 'board');
      const blankEmpty = (() => {
        const element = document.createElement('boardgame-component-zone');
        element.label = 'Zone';
        element.emptyLabel = ' ';
        try {
          element.render();
          return '<resolved>';
        } catch (error) {
          return error instanceof Error ? error.message : String(error);
        }
      })();
      zone.remove();
      return { defaults, suppressed, blankLabel, invalidHeading, boardLayout, blankEmpty };
    });

    expect(result.defaults).toEqual({
      sectionLabel: 'label',
      zonePart: 'zone',
      heading: 'Draw pile',
      headingRole: 'heading',
      headingLevel: '3',
      count: '0',
      countLabel: '0 items',
      empty: 'Empty',
      assignedHeadingActions: 1,
      stack: { layout: 'fan', messy: true, disabled: true, part: 'stack' },
    });
    expect(result.suppressed).toEqual({ count: false, empty: false });
    expect(result.blankLabel).toContain('label must be a non-empty visible and accessible name');
    expect(result.invalidHeading).toContain('headingLevel must be a safe integer from 1 through 6');
    expect(result.boardLayout).toContain('board and spatial layouts use their dedicated components');
    expect(result.blankEmpty).toContain('emptyLabel must be non-empty');
    diagnostics.assertEmpty();
  } finally {
    diagnostics.stop();
  }
});

test('bindMoveAction adapts a typed action to md-filled-button semantics', async ({ page }) => {
  const diagnostics = await prepareRendererFixturePage(page);
  try {
    const result = await page.evaluate(async () => {
      await import('/node_modules/@material/web/button/filled-button.js');
      await import('/game-src/pig/boardgame-render-game-pig.ts');
      const { pigRendererFixture } = await import('/game-src/pig/boardgame-render-fixtures-pig.ts');
      const { mountRendererFixture } = await import('/src/testing/renderer-fixture.ts');
      const { html, bindMoveAction } = await import('/src/client.ts');
      const PigRenderer = customElements.get('boardgame-render-game-pig');
      if (!PigRenderer) throw new Error('Pig renderer was not registered');
      class AdapterRenderer extends PigRenderer {
        localDisabled = false;

        override render() {
          const renderer = this as unknown as { move(name: string): object };
          return html`<md-filled-button
            title="Author title"
            aria-description="Author description"
            ${bindMoveAction(renderer.move('Done Turn') as never, { disabled: this.localDisabled })}>
            Done
          </md-filled-button>`;
        }
      }
      customElements.define('boardgame-render-game-pig-action-adapter', AdapterRenderer);
      const handle = await mountRendererFixture({
        ...pigRendererFixture,
        tagName: 'boardgame-render-game-pig-action-adapter',
      } as never);
      const button = handle.renderer.shadowRoot?.querySelector('md-filled-button') as (
        HTMLElement & { disabled: boolean }
      ) | null;
      if (!button) throw new Error('Adapter renderer did not render a Material button');
      const initial = {
        disabled: button.disabled,
        ariaDisabled: button.getAttribute('aria-disabled'),
        ariaBusy: button.getAttribute('aria-busy'),
      };
      button.click();
      await Promise.resolve();
      const typedRenderer = handle.renderer as HTMLElement & {
        localDisabled: boolean;
        requestUpdate(): void;
        updateComplete: Promise<unknown>;
      };
      typedRenderer.localDisabled = true;
      typedRenderer.requestUpdate();
      await typedRenderer.updateComplete;
      const locallyDisabled = {
        disabled: button.disabled,
        title: button.getAttribute('title'),
      };
      typedRenderer.remove();
      await Promise.resolve();
      const detached = {
        disabled: button.disabled,
        title: button.getAttribute('title'),
        description: button.getAttribute('aria-description'),
        ariaDisabled: button.getAttribute('aria-disabled'),
      };
      handle.host.append(typedRenderer);
      await typedRenderer.updateComplete;
      const reconnected = {
        disabled: button.disabled,
        title: button.getAttribute('title'),
      };
      const proposals = handle.proposals;
      handle.dispose();
      return { initial, locallyDisabled, detached, reconnected, proposals };
    });
    expect(result.initial).toEqual({ disabled: false, ariaDisabled: 'false', ariaBusy: 'false' });
    expect(result.locallyDisabled).toEqual({
      disabled: true,
      title: 'This action is disabled by the renderer',
    });
    expect(result.detached).toEqual({
      disabled: false,
      title: 'Author title',
      description: 'Author description',
      ariaDisabled: null,
    });
    expect(result.reconnected).toEqual({
      disabled: true,
      title: 'This action is disabled by the renderer',
    });
    expect(result.proposals).toHaveLength(1);
    expect(result.proposals[0]).toMatchObject({ name: 'Done Turn', arguments: {} });
    diagnostics.assertEmpty();
  } finally {
    diagnostics.stop();
  }
});

test('fixture host rejects stale schemas and unregistered renderers loudly', async ({ page }) => {
  await page.goto('/client_config.js');
  await page.evaluate(() => {
    document.open();
    document.write('<!doctype html><html lang="en"><body></body></html>');
    document.close();
  });
  const result = await page.evaluate(async () => {
    const { mountRendererFixture } = await import('/src/testing/renderer-fixture.ts');
    const base = {
      tagName: 'missing-renderer',
      snapshot: {
        schemaVersion: 1,
        state: {},
        viewingAsPlayer: 0,
        currentPlayerIndex: 0,
        moveLegality: {},
        version: 0,
        outcome: { finished: false, winners: [] },
        surface: 'game',
        serverMoveInputSchemaFingerprint: `sha256:${'0'.repeat(64)}`,
      },
    } as const;
    const messages: string[] = [];
    try {
      await mountRendererFixture(base as never);
    } catch (error) {
      messages.push(error instanceof Error ? error.message : String(error));
    }
    try {
      await mountRendererFixture({
        ...base,
        snapshot: { ...base.snapshot, schemaVersion: 2 },
      } as never);
    } catch (error) {
      messages.push(error instanceof Error ? error.message : String(error));
    }
    try {
      await mountRendererFixture({
        tagName: 'boardgame-render-game-missing-table',
        snapshot: base.snapshot,
      } as never);
    } catch (error) {
      messages.push(error instanceof Error ? error.message : String(error));
    }
    try {
      await mountRendererFixture({
        ...base,
        snapshot: {
          ...base.snapshot,
          moveLegality: { Act: { legalForPlayer: true, legalForAnyone: false } },
        },
      } as never);
    } catch (error) {
      messages.push(error instanceof Error ? error.message : String(error));
    }
    let proposalListenersAdded = 0;
    let proposalListenersRemoved = 0;
    class RejectingRenderer extends HTMLElement {
      state: object | null = null;
      viewingAsPlayer = 0;
      currentPlayerIndex = 0;
      moveLegality = {};
      gameFinished = false;
      gameWinners: number[] = [];
      serverMoveInputSchemaFingerprint: string | null = null;
      previewDisabledSpaces: number[] = [];
      readonly updateComplete = Promise.reject(new Error('deliberate render failure'));

      override addEventListener(
        type: string,
        listener: EventListenerOrEventListenerObject | null,
        options?: boolean | AddEventListenerOptions,
      ): void {
        if (type === 'propose-move') proposalListenersAdded += 1;
        super.addEventListener(type, listener, options);
      }

      override removeEventListener(
        type: string,
        listener: EventListenerOrEventListenerObject | null,
        options?: boolean | EventListenerOptions,
      ): void {
        if (type === 'propose-move') proposalListenersRemoved += 1;
        super.removeEventListener(type, listener, options);
      }
    }
    customElements.define('boardgame-render-game-rejecting', RejectingRenderer);
    try {
      await mountRendererFixture({
        tagName: 'boardgame-render-game-rejecting',
        snapshot: base.snapshot,
      } as never);
    } catch (error) {
      messages.push(error instanceof Error ? error.message : String(error));
    }
    class AcceptingRenderer extends HTMLElement {
      state: object | null = null;
      viewingAsPlayer = 0;
      currentPlayerIndex = 0;
      moveLegality = {};
      gameFinished = false;
      gameWinners: number[] = [];
      serverMoveInputSchemaFingerprint: string | null = null;
      previewDisabledSpaces: number[] = [];
      readonly updateComplete = Promise.resolve();
    }
    customElements.define('boardgame-render-game-accepting', AcceptingRenderer);
    const accepting = await mountRendererFixture({
      tagName: 'boardgame-render-game-accepting',
      snapshot: {
        ...base.snapshot,
        moveLegality: { Act: { legalForPlayer: true, legalForAnyone: true } },
      },
    } as never);
    const onPrototypeProposal = (event: ErrorEvent) => {
      messages.push(event.message);
      event.preventDefault();
    };
    window.addEventListener('error', onPrototypeProposal);
    accepting.renderer.dispatchEvent(new CustomEvent('propose-move', {
      detail: { name: 'toString', arguments: {} },
    }));
    window.removeEventListener('error', onPrototypeProposal);
    accepting.dispose();
    return {
      messages,
      proposalListenersAdded,
      proposalListenersRemoved,
      leakedHosts: document.querySelectorAll('[data-renderer-fixture]').length,
    };
  });
  expect(result.messages).toEqual([
    'missing-renderer is not registered; import the renderer module before mounting it',
    'Unsupported renderer fixture schema version: 2',
    'Renderer fixture surface game does not match renderer tag boardgame-render-game-missing-table',
    'Renderer fixture legality is contradictory for move Act',
    'deliberate render failure',
    'Uncaught Error: Renderer fixture received unknown move proposal: toString',
  ]);
  expect(result).toMatchObject({
    proposalListenersAdded: 1,
    proposalListenersRemoved: 1,
    leakedHosts: 0,
  });
});

test('Checkers composes source selection with typed destination actions', async ({ page }) => {
  const diagnostics = await prepareRendererFixturePage(page);
  try {
    await page.evaluate(async () => {
      await import('/game-src/checkers/boardgame-render-game-checkers.ts');
      const { checkersRendererFixture } = await import(
        '/game-src/checkers/boardgame-render-fixtures-checkers.ts'
      );
      const { mountRendererFixture } = await import('/src/testing/renderer-fixture.ts');
      const handle = await mountRendererFixture(checkersRendererFixture);
      (globalThis as unknown as { __checkersFixtureHandle: typeof handle }).__checkersFixtureHandle = handle;
    });

    const board = page.locator('boardgame-game-board');
    const source = board.locator('.cell[data-index="17"]');
    const sourceCell = source.locator('xpath=..');
    const destination = board.locator('.cell[data-index="26"]');
    await expect(source).toHaveAttribute('aria-label', /Selectable source/);
    await source.click();
    await expect(sourceCell).toHaveAttribute('aria-selected', 'true');
    await expect(destination).toHaveAttribute('aria-disabled', 'false');
    await destination.click();

    const proposal = await page.evaluate(() => {
      const handle = (globalThis as unknown as {
        __checkersFixtureHandle: { readonly proposals: readonly unknown[] };
      }).__checkersFixtureHandle;
      return handle.proposals[0];
    });
    expect(proposal).toMatchObject({
      snapshotVersion: 3,
      name: 'Move Token',
      arguments: { TokenIndexToMove: '17', SpaceIndex: '26' },
    });
    await expect(sourceCell).toHaveAttribute('aria-selected', 'false');

    await source.click();
    await expect(sourceCell).toHaveAttribute('aria-selected', 'true');
    await source.press('Escape');
    await expect(sourceCell).toHaveAttribute('aria-selected', 'false');
    await source.click();
    await expect(sourceCell).toHaveAttribute('aria-selected', 'true');
    await page.evaluate(async () => {
      const { checkersRendererFixture } = await import(
        '/game-src/checkers/boardgame-render-fixtures-checkers.ts'
      );
      const handle = (globalThis as unknown as {
        __checkersFixtureHandle: {
          update(snapshot: typeof checkersRendererFixture.snapshot): Promise<void>;
          dispose(): void;
        };
      }).__checkersFixtureHandle;
      await handle.update({ ...checkersRendererFixture.snapshot, version: 4 });
    });
    await expect(sourceCell).toHaveAttribute('aria-selected', 'false');
    const axeResult = await new AxeBuilder({ page })
      .include('[data-renderer-fixture]')
      .withRules(['aria-required-children', 'aria-required-parent', 'button-name'])
      .analyze();
    expect(axeResult.violations).toEqual([]);
    diagnostics.assertEmpty();
    await page.evaluate(() => {
      (globalThis as unknown as { __checkersFixtureHandle: { dispose(): void } })
        .__checkersFixtureHandle.dispose();
    });
  } finally {
    diagnostics.stop();
  }
});

test('Tic-tac-toe fixture proposes native numeric targets and stays bounded at canonical widths', async ({ page }) => {
  const diagnostics = await prepareRendererFixturePage(page);
  try {
    const proposals = await retryRendererEvaluation(page, () => page.evaluate(async () => {
      await import('/game-src/tictactoe/boardgame-render-game-tictactoe.ts');
      const { tictactoeRendererFixture } = await import(
        '/game-src/tictactoe/boardgame-render-fixtures-tictactoe.ts'
      );
      const { mountRendererFixture } = await import('/src/testing/renderer-fixture.ts');
      const handle = await mountRendererFixture(tictactoeRendererFixture);
      const board = handle.renderer.shadowRoot?.querySelector('boardgame-game-board');
      if (!(board instanceof HTMLElement)) throw new Error('Tic-tac-toe fixture did not render its board');
      const button = board.shadowRoot?.querySelector<HTMLButtonElement>('.cell[data-index="1"]');
      if (!button) throw new Error('Tic-tac-toe fixture did not render target buttons');
      button.click();
      for (let attempt = 0; attempt < 20 && handle.proposals.length === 0; attempt++) {
        await new Promise(resolve => setTimeout(resolve, 0));
      }
      return handle.proposals;
    }));

    expect(proposals).toEqual([{
      requestID: 'fixture-v4-move-2',
      snapshotVersion: 4,
      name: 'Place Token',
      arguments: { Slot: '1' },
    }]);

    const buttons = page.locator('boardgame-game-board .cell');
    await expect(buttons).toHaveCount(9);
    await expect(buttons.nth(0)).toHaveAttribute('aria-disabled', 'true');
    // A successful proposal consumes the snapshot, but unavailable cells remain
    // keyboard-discoverable instead of disappearing from the roving grid.
    await expect(buttons.nth(1)).toHaveAttribute('aria-disabled', 'true');
    await expect(buttons.nth(1)).toHaveAttribute('title', /Waiting for the accepted move/);
    await buttons.nth(1).focus();
    await page.keyboard.press('ArrowRight');
    await expect(buttons.nth(2)).toBeFocused();
    await page.keyboard.press('ArrowRight');
    await expect(buttons.nth(2)).toBeFocused();
    await page.keyboard.press('End');
    await expect(buttons.nth(2)).toBeFocused();
    await page.keyboard.press('Control+End');
    await expect(buttons.nth(8)).toBeFocused();

    const labelGeometry = await page.locator('boardgame-game-board').evaluate(async board => {
      const typed = board as HTMLElement & { labels: boolean; updateComplete: Promise<unknown> };
      typed.labels = true;
      await typed.updateComplete;
      const host = typed.getBoundingClientRect();
      const labels = [...(typed.shadowRoot?.querySelectorAll('.label') ?? [])]
        .map(label => label.getBoundingClientRect());
      const result = {
        count: labels.length,
        contained: labels.every(label => label.left >= host.left - 0.5
          && label.right <= host.right + 0.5
          && label.top >= host.top - 0.5
          && label.bottom <= host.bottom + 0.5),
      };
      typed.labels = false;
      await typed.updateComplete;
      return result;
    });
    expect(labelGeometry).toEqual({ count: 6, contained: true });

    const axeResult = await new AxeBuilder({ page }).include('[data-renderer-fixture]').analyze();
    expect(axeResult.violations).toEqual([]);

    for (const viewport of Object.values(RENDERER_VIEWPORTS)) {
      await page.setViewportSize(viewport);
      const geometry = await page.locator('boardgame-game-board').evaluate((board) => {
        const bounds = board.getBoundingClientRect();
        return {
          pageWidth: document.documentElement.scrollWidth,
          viewportWidth: document.documentElement.clientWidth,
          boardWidth: bounds.width,
          boardHeight: bounds.height,
        };
      });
      expect(geometry.pageWidth).toBeLessThanOrEqual(geometry.viewportWidth);
      expect(Math.abs(geometry.boardWidth - geometry.boardHeight)).toBeLessThanOrEqual(2);
    }
    const rectangular = await page.locator('boardgame-game-board').evaluate(async (element) => {
      const board = element as HTMLElement & {
        rows: number;
        cols: number;
        stack: object;
        action: object | null;
        updateComplete: Promise<unknown>;
      };
      board.rows = 2;
      board.cols = 3;
      board.action = null;
      const components = Array.from({ length: 6 }, (_, index) => ({
        Index: index,
        Values: { Value: index % 2 === 0 ? 'X' : 'O' },
        Deck: 'tokens',
        GameName: 'tictactoe',
        ID: `rectangular-token-${index}`,
      }));
      board.stack = {
        Deck: 'tokens', Indexes: [0, 1, 2, 3, 4, 5], IDs: components.map(({ ID }) => ID),
        IDsLastSeen: {}, ShuffleCount: 0, Size: 6, MaxSize: 6, GameName: 'tictactoe',
        Components: components,
      };
      await board.updateComplete;
      const componentStack = board.shadowRoot?.querySelector('boardgame-component-stack') as (
        HTMLElement & { boardRows: number; boardCols: number; updateComplete: Promise<unknown> }
      ) | null;
      if (!componentStack) throw new Error('Board did not render its component stack');
      await componentStack.updateComplete;
      const area = board.shadowRoot?.querySelector('.board-area');
      const cells = [...(board.shadowRoot?.querySelectorAll('.cell') ?? [])];
      const pieces = [...componentStack.querySelectorAll('[boardgame-component]')];
      if (!(area instanceof HTMLElement) || cells.length !== 6 || pieces.length !== 6) {
        throw new Error(`Rectangular board geometry was incomplete: ${cells.length} cells, ${pieces.length} pieces`);
      }
      const areaBounds = area.getBoundingClientRect();
      const cellBounds = cells.map((cell) => cell.getBoundingClientRect());
      const pieceBounds = pieces.map((piece) => piece.getBoundingClientRect());
      const centerErrors = cellBounds.map((cell, index) => {
        const piece = pieceBounds[index];
        if (!piece) return Number.POSITIVE_INFINITY;
        return Math.max(
          Math.abs((cell.left + cell.width / 2) - (piece.left + piece.width / 2)),
          Math.abs((cell.top + cell.height / 2) - (piece.top + piece.height / 2)),
        );
      });
      return {
        innerRatio: areaBounds.width / areaBounds.height,
        cellSquareError: Math.max(...cellBounds.map((cell) => Math.abs(cell.width - cell.height))),
        containmentError: Math.max(...cellBounds.map((cell) => (
          Math.max(cell.right - areaBounds.right, cell.bottom - areaBounds.bottom)
        ))),
        centerError: Math.max(...centerErrors),
        componentRows: componentStack.boardRows,
        componentCols: componentStack.boardCols,
      };
    });
    expect(rectangular.innerRatio).toBeCloseTo(1.5, 2);
    expect(rectangular.cellSquareError).toBeLessThanOrEqual(1);
    expect(rectangular.containmentError).toBeLessThanOrEqual(1);
    expect(rectangular.centerError).toBeLessThanOrEqual(1);
    expect(rectangular).toMatchObject({ componentRows: 2, componentCols: 3 });
    diagnostics.assertEmpty();
  } finally {
    diagnostics.stop();
  }
});

test('game board rejects contradictory geometry and accessibility configuration loudly', async ({ page }) => {
  await page.goto('/client_config.js');
  const messages = await page.evaluate(async () => {
    await import('/src/components/boardgame-game-board.ts');
    const rejects = async (properties: Record<string, unknown>): Promise<string> => {
      const board = document.createElement('boardgame-game-board') as HTMLElement & {
        updateComplete: Promise<unknown>;
      };
      Object.assign(board, properties);
      document.body.append(board);
      try {
        await board.updateComplete;
        return '<resolved>';
      } catch (error) {
        return error instanceof Error ? error.message : String(error);
      } finally {
        board.remove();
      }
    };
    return Promise.all([
      rejects({ rows: 0, cols: 3 }),
      rejects({ rows: 100, cols: 100 }),
      rejects({ rows: 33, cols: 32, action: { candidates: [] } }),
      rejects({ rows: 1, cols: 1, stack: { Components: null } }),
      rejects({ rows: 2, cols: 2, stack: { Components: [null] } }),
      rejects({ rows: 1, cols: 1, labelFor: () => { throw new Error('boom'); } }),
      rejects({ rows: 1, cols: 1, labelFor: () => 42 }),
      rejects({ rows: 1, cols: 2, labelFor: () => 'same' }),
      rejects({
        rows: 1,
        cols: 1,
        action: { candidates: [], subscribe: () => () => undefined },
        sourceDestination: { selectedSource: null, sources: [], action: null, selectSource: () => undefined, clear: () => undefined },
      }),
      rejects({
        rows: 1,
        cols: 2,
        sourceDestination: { selectedSource: null, sources: [2], action: null, selectSource: () => undefined, clear: () => undefined },
      }),
      rejects({
        rows: 1,
        cols: 2,
        action: {
          candidates: [{ key: 0 }, { key: 0 }],
          preview: { kind: 'ready' },
          get: () => undefined,
          subscribe: () => () => undefined,
        },
      }),
    ]);
  });
  expect(messages).toEqual([
    expect.stringMatching(/positive integers/),
    expect.stringMatching(/maximum is 4096/),
    expect.stringMatching(/target actions support at most 1024 cells/),
    expect.stringMatching(/stack\.Components must be an array/),
    expect.stringMatching(/expected 4 stack components/),
    expect.stringMatching(/labelFor failed for cell 0: boom/),
    expect.stringMatching(/labelFor must return a string for cell 0/),
    expect.stringMatching(/labels must be unique/),
    expect.stringMatching(/mutually exclusive/),
    expect.stringMatching(/source key 2 is outside 0 through 1/),
    expect.stringMatching(/keys must cover exactly 0 through 1/),
  ]);
});

test('spatial board sanitizes authored geometry and shares typed target activation', async ({ page }) => {
  await page.goto('/client_config.js');
  const result = await page.evaluate(async () => {
    const source = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="10 20 200 100" preserveAspectRatio="xMaxYMin meet"
      onload="globalThis.__unsafeBoardRoot = true" style="fill:url(https://attacker.invalid/root)">
      <script>globalThis.__unsafeBoardScript = true</script>
      <g transform="translate(20 10)">
        <g data-board-space="room:one/?" data-board-label="Library" data-board-order="3"
          onclick="globalThis.__unsafeBoardClick = true" fill="url(#safe) url(https://attacker.invalid/paint)">
          <path d="M 0 0 h 80 v 60 h -80 z" />
          <circle data-inner="" cx="20" cy="20" r="10" />
          <animate attributeName="href" values="javascript:alert(1)" />
        </g>
        <rect data-board-focus-anchor="room:one/?" x="10" y="10" width="4" height="4" />
        <rect data-board-piece-anchor="room:one/?" x="50" y="30" width="6" height="6" />
        <image href="https://attacker.invalid/tracker.png" x="0" y="0" width="2" height="2" />
      </g>
      <rect data-decorative-overlay="" x="20" y="10" width="80" height="60" fill="transparent" />
    </svg>`;
    const originalFetch = globalThis.fetch;
    globalThis.fetch = async () => new Response(source, {
      status: 200,
      headers: { 'content-type': 'image/svg+xml' },
    });
    try {
      await import('/src/components/boardgame-spatial-board.ts');
      const geometryHelpers = await import('/src/components/spatial-board-geometry.ts');
      const componentViews = await import('/src/components/component-view.ts');
      let duplicateAnchorError = '';
      try {
        const duplicateAnchorSvg = geometryHelpers.parseTrustedBoardSvg(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 10 10">
          <rect data-board-space="room" data-board-label="Room" width="10" height="10" />
          <circle data-board-piece-anchor="room" cx="2" cy="2" r="1" />
          <circle data-board-piece-anchor="room" cx="3" cy="3" r="1" />
        </svg>`);
        geometryHelpers.geometryFromSvg(duplicateAnchorSvg);
      } catch (error) {
        duplicateAnchorError = error instanceof Error ? error.message : String(error);
      }
      let activations = 0;
      const actionState = {
        canActivate: true,
        reason: null,
        activate: async () => {
          activations++;
          return { kind: 'success', requestID: 'spatial-test' };
        },
      };
      const candidate = { key: 'room:one/?', action: actionState };
      const action = {
        candidates: [candidate],
        preview: { kind: 'ready' },
        get: (key: string | number) => key === candidate.key ? candidate : undefined,
        ensurePreview: async () => ({ kind: 'ready' }),
        refreshPreview: async () => ({ kind: 'ready' }),
        subscribe: () => () => undefined,
      };
      const board = document.createElement('boardgame-spatial-board') as HTMLElement & {
        svgUrl: string;
        action: unknown;
        pieces: readonly unknown[];
        tokenSize: number;
        geometryInspector: boolean;
        geometry: ((svg: SVGSVGElement) => unknown) | null;
        componentView: unknown;
        svgLoaded: boolean;
        updateComplete: Promise<unknown>;
      };
      board.svgUrl = '/authored-board.svg';
      board.action = action;
      board.tokenSize = 20;
      board.geometryInspector = true;
      const token = { Index: 0, Values: {}, Deck: 'tokens', GameName: 'fixture', ID: 'token-0' };
      const stack = {
        Deck: 'tokens', Indexes: [0], IDs: [token.ID], IDsLastSeen: {}, ShuffleCount: 0,
        GameName: 'fixture', Components: [token],
      };
      const tokenView = componentViews.tokenView({
        properties: () => ({ type: 'meeple', color: 'teal' }),
      });
      board.componentView = tokenView;
      board.pieces = [{ id: token.ID, space: candidate.key, stack, slot: 0, component: token }];
      document.body.append(board);
      for (let attempt = 0; attempt < 20 && !board.svgLoaded; attempt++) {
        await new Promise(resolve => setTimeout(resolve, 0));
      }
      await board.updateComplete;
      const root = board.shadowRoot;
      const region = root?.querySelector('[data-board-space]');
      const inner = root?.querySelector('[data-inner]');
      const button = root?.querySelector('#space-list button');
      const focusButton = root?.querySelector('.space-focus');
      const pieceAnchor = root?.querySelector('[data-board-piece-anchor]');
      const focusAnchor = root?.querySelector('[data-board-focus-anchor]');
      const decorativeOverlay = root?.querySelector('[data-decorative-overlay]');
      const artwork = root?.querySelector('#container > svg');
      const componentStack = root?.querySelector('boardgame-component-stack') as (
        HTMLElement & {
          spatialPositions: readonly ({ top: number; left: number } | null)[];
          componentView: unknown;
          updateComplete: Promise<unknown>;
        }
      ) | null;
      if (!(region instanceof SVGElement) || !(inner instanceof SVGElement)
        || !(pieceAnchor instanceof SVGGraphicsElement) || !(focusAnchor instanceof SVGGraphicsElement)
        || !(decorativeOverlay instanceof SVGGraphicsElement) || !(artwork instanceof SVGSVGElement)
        || !(button instanceof HTMLButtonElement)
        || !(focusButton instanceof HTMLButtonElement)
        || !componentStack) throw new Error('Spatial fixture did not render');
      await componentStack.updateComplete;
      const position = componentStack.spatialPositions[0];
      if (!position) throw new Error('Spatial fixture did not position its piece');
      const anchorBounds = pieceAnchor.getBoundingClientRect();
      const focusAnchorBounds = focusAnchor.getBoundingClientRect();
      const focusButtonBounds = focusButton.getBoundingClientRect();
      const containerBounds = root?.querySelector('#container')?.getBoundingClientRect();
      if (!containerBounds) throw new Error('Spatial fixture had no container geometry');
      const jitter = (axis: number) => {
        let hash = axis * 41;
        hash = ((hash >>> 16) ^ hash) * 0x45d9f3b;
        hash = ((hash >>> 16) ^ hash) * 0x45d9f3b;
        hash = (hash >>> 16) ^ hash;
        return ((hash & 0xFFFF) / 0x7FFF) - 1;
      };
      const expectedLeft = anchorBounds.left + anchorBounds.width / 2 - containerBounds.left - 10 + jitter(0) * 20;
      const expectedTop = anchorBounds.top + anchorBounds.height / 2 - containerBounds.top - 10 + jitter(1) * 20;
      inner.dispatchEvent(new MouseEvent('click', { bubbles: true, composed: true }));
      const overlayBounds = decorativeOverlay.getBoundingClientRect();
      decorativeOverlay.dispatchEvent(new MouseEvent('click', {
        bubbles: true, composed: true,
        clientX: overlayBounds.left + overlayBounds.width / 2,
        clientY: overlayBounds.top + overlayBounds.height / 2,
      }));
      await Promise.resolve();
      focusButton.focus();

      // A custom sidecar uses a real sanitized SVG element that has no
      // data-board-space attribute; activation must follow the returned region.
      const sidecarBoard = document.createElement('boardgame-spatial-board') as typeof board;
      sidecarBoard.svgUrl = '/authored-board.svg';
      sidecarBoard.action = action;
      sidecarBoard.geometry = svg => ({ spaces: [{
        key: candidate.key,
        label: 'Sidecar Library',
        region: svg.querySelector('path') as SVGGraphicsElement,
      }] });
      document.body.append(sidecarBoard);
      for (let attempt = 0; attempt < 20 && !sidecarBoard.svgLoaded; attempt++) {
        await new Promise(resolve => setTimeout(resolve, 0));
      }
      await sidecarBoard.updateComplete;
      const sidecarRegion = sidecarBoard.shadowRoot?.querySelector('path');
      if (!(sidecarRegion instanceof SVGGraphicsElement)) throw new Error('Sidecar path geometry did not render');
      sidecarRegion.dispatchEvent(new MouseEvent('click', { bubbles: true, composed: true }));
      await Promise.resolve();
      sidecarBoard.remove();

      let responsiveAnchorError = 0;
      let pieceOutsideRegion = false;
      const settleGeometry = async () => {
        await new Promise<void>(resolve => requestAnimationFrame(() => requestAnimationFrame(() => resolve())));
        await board.updateComplete;
        await componentStack.updateComplete;
      };
      const verifyGeometry = () => {
        const currentFocus = focusAnchor.getBoundingClientRect();
        const currentButton = focusButton.getBoundingClientRect();
        responsiveAnchorError = Math.max(responsiveAnchorError,
          Math.abs((currentButton.left + currentButton.width / 2) - (currentFocus.left + currentFocus.width / 2)),
          Math.abs((currentButton.top + currentButton.height / 2) - (currentFocus.top + currentFocus.height / 2)));
        const currentPosition = componentStack.spatialPositions[0];
        const currentContainer = root?.querySelector('#container')?.getBoundingClientRect();
        const currentRegion = region.getBoundingClientRect();
        if (!currentPosition || !currentContainer) throw new Error('Responsive geometry disappeared');
        const tokenLeft = currentContainer.left + currentPosition.left;
        const tokenTop = currentContainer.top + currentPosition.top;
        pieceOutsideRegion ||= tokenLeft < currentRegion.left - 1 || tokenTop < currentRegion.top - 1
          || tokenLeft + 20 > currentRegion.right + 1 || tokenTop + 20 > currentRegion.bottom + 1;
      };
      for (const width of [320, 768, 1280]) {
        board.style.width = `${width}px`;
        await settleGeometry();
        verifyGeometry();
      }
      document.documentElement.style.fontSize = '200%';
      board.style.setProperty('zoom', '2');
      await settleGeometry();
      verifyGeometry();
      board.style.removeProperty('zoom');
      document.documentElement.style.fontSize = '';
      board.style.width = '330px';
      board.style.width = '900px';
      board.style.width = '480px';
      await settleGeometry();
      verifyGeometry();

      const successful = {
        activations,
        label: button.textContent?.trim(),
        focused: region.classList.contains('focused'),
        scriptCount: root?.querySelectorAll('script').length,
        onclick: region.getAttribute('onclick'),
        unsafeFill: region.getAttribute('fill'),
        animationCount: root?.querySelectorAll('animate').length,
        imageHref: root?.querySelector('image')?.getAttribute('href'),
        unsafeScriptRan: Boolean((globalThis as unknown as { __unsafeBoardScript?: boolean }).__unsafeBoardScript),
        unsafeRootRan: Boolean((globalThis as unknown as { __unsafeBoardRoot?: boolean }).__unsafeBoardRoot),
        rootOnload: artwork.getAttribute('onload'),
        rootStyle: artwork.getAttribute('style'),
        artworkHidden: artwork.getAttribute('aria-hidden'),
        duplicateAnchorError,
        responsiveAnchorError,
        pieceOutsideRegion,
        componentViewForwarded: componentStack.componentView === tokenView,
        anchorError: Math.max(Math.abs(position.left - expectedLeft), Math.abs(position.top - expectedTop)),
        focusAnchorError: Math.max(
          Math.abs((focusButtonBounds.left + focusButtonBounds.width / 2)
            - (focusAnchorBounds.left + focusAnchorBounds.width / 2)),
          Math.abs((focusButtonBounds.top + focusButtonBounds.height / 2)
            - (focusAnchorBounds.top + focusAnchorBounds.height / 2)),
        ),
        inspector: root?.querySelector('#geometry-inspector')?.textContent?.replace(/\s+/g, ' ').trim(),
      };

      let resolveSlow: ((response: Response) => void) | undefined;
      globalThis.fetch = async input => {
        const url = String(input);
        if (url.includes('slow.svg')) return new Promise<Response>(resolve => { resolveSlow = resolve; });
        if (url.includes('missing.svg')) return new Response('missing', { status: 404 });
        return new Response(source.replace('Library', 'Fast Library'), {
          status: 200,
          headers: { 'content-type': 'image/svg+xml' },
        });
      };
      board.svgUrl = '/slow.svg';
      await board.updateComplete;
      board.svgUrl = '/fast.svg';
      await board.updateComplete;
      for (let attempt = 0; attempt < 20 && !board.svgLoaded; attempt++) {
        await new Promise(resolve => setTimeout(resolve, 0));
      }
      resolveSlow?.(new Response(source.replace('Library', 'Stale Library'), {
        status: 200,
        headers: { 'content-type': 'image/svg+xml' },
      }));
      await new Promise(resolve => setTimeout(resolve, 0));
      const postRaceLabel = root?.querySelector('#space-list button')?.textContent?.trim();

      board.svgUrl = '/missing.svg';
      await board.updateComplete;
      for (let attempt = 0; attempt < 20 && !root?.querySelector('#status')?.textContent?.includes('HTTP 404'); attempt++) {
        await new Promise(resolve => setTimeout(resolve, 0));
      }
      return {
        ...successful,
        postRaceLabel,
        failedSvgCount: root?.querySelectorAll('#container > svg').length,
        failureStatus: root?.querySelector('#status')?.textContent?.replace(/\s+/g, ' ').trim(),
      };
    } finally {
      globalThis.fetch = originalFetch;
    }
  });
  expect(result).toMatchObject({
    activations: 3,
    label: 'Library',
    focused: true,
    scriptCount: 0,
    onclick: null,
    unsafeFill: null,
    animationCount: 0,
    imageHref: null,
    unsafeScriptRan: false,
    unsafeRootRan: false,
    rootOnload: null,
    rootStyle: null,
    artworkHidden: 'true',
    duplicateAnchorError: expect.stringContaining('duplicate data-board-piece-anchor'),
    pieceOutsideRegion: false,
    componentViewForwarded: true,
    inspector: expect.stringContaining('room:one/? — Library'),
    postRaceLabel: 'Fast Library',
    failedSvgCount: 0,
    failureStatus: expect.stringContaining('HTTP 404'),
  });
  expect(result.anchorError).toBeLessThanOrEqual(1);
  expect(result.focusAnchorError).toBeLessThanOrEqual(1);
  expect(result.responsiveAnchorError).toBeLessThanOrEqual(1);
  const axeResult = await new AxeBuilder({ page })
    .include('boardgame-spatial-board')
    .withRules(['button-name', 'aria-allowed-attr', 'aria-valid-attr-value', 'nested-interactive'])
    .analyze();
  expect(axeResult.violations).toEqual([]);
});

test('spatial board rejects ambiguous or misaligned component views loudly', async ({ page }) => {
  await page.goto('/client_config.js');
  const messages = await page.evaluate(async () => {
    await import('/src/components/boardgame-spatial-board.ts');
    const { tokenView } = await import('/src/components/component-view.ts');
    const view = tokenView({ properties: () => ({ type: 'meeple' }) });
    const stack = {
      Deck: 'tokens', Indexes: [], IDs: [], IDsLastSeen: {}, ShuffleCount: 0,
      GameName: 'fixture', Components: [],
    };
    const rejects = async (properties: Record<string, unknown>): Promise<string> => {
      const board = document.createElement('boardgame-spatial-board');
      Object.assign(board, properties);
      document.body.append(board);
      try {
        await (board as typeof board & { updateComplete: Promise<unknown> }).updateComplete;
        return '<resolved>';
      } catch (error) {
        return error instanceof Error ? error.message : String(error);
      } finally {
        board.remove();
      }
    };
    return Promise.all([
      rejects({ stack, componentView: view, componentViews: [view] }),
      rejects({ stacks: [stack, stack], componentViews: [view] }),
      rejects({ componentViews: [view] }),
    ]);
  });
  expect(messages).toEqual([
    expect.stringContaining('choose componentView or componentViews, not both'),
    expect.stringContaining('componentViews has 1 entries for 2 effective stack layers'),
    expect.stringContaining('componentViews has 1 entries for 0 effective stack layers'),
  ]);
});

test('game board reconnects, exposes reasons, avoids double retry, and labels wide coordinates', async ({ page }) => {
  await page.goto('/client_config.js');
  const result = await page.evaluate(async () => {
    await import('/src/components/boardgame-game-board.ts');
    let subscriptions = 0;
    let unsubscriptions = 0;
    let activations = 0;
    let collectionRefreshes = 0;
    const candidateAction = {
      canPropose: false,
      canActivate: true,
      preview: { kind: 'failed' },
      reason: { code: 'preview-failed', message: 'Network unavailable' },
      activate: async () => { activations++; return { kind: 'blocked' }; },
    };
    const action = {
      candidates: [{ key: 1, action: candidateAction }, { key: 0, action: candidateAction }],
      preview: { kind: 'failed', retryable: true, reason: candidateAction.reason },
      get: (key: number) => ({ key, action: candidateAction }),
      refreshPreview: async () => { collectionRefreshes++; return action.preview; },
      subscribe: () => {
        subscriptions++;
        return () => { unsubscriptions++; };
      },
    };
    const board = document.createElement('boardgame-game-board') as HTMLElement & {
      rows: number;
      cols: number;
      action: object;
      labels: boolean;
      updateComplete: Promise<unknown>;
    };
    board.rows = 1;
    board.cols = 2;
    board.action = action;
    document.body.append(board);
    await board.updateComplete;
    const retry = board.shadowRoot?.querySelector<HTMLButtonElement>('.cell[data-index="0"]');
    const retryLabel = retry?.getAttribute('aria-label') ?? '';
    retry?.click();
    await Promise.resolve();
    board.remove();
    document.body.append(board);
    await board.updateComplete;
    board.remove();

    const wide = document.createElement('boardgame-game-board') as HTMLElement & {
      rows: number;
      cols: number;
      labels: boolean;
      updateComplete: Promise<unknown>;
    };
    wide.rows = 1;
    wide.cols = 27;
    wide.labels = true;
    document.body.append(wide);
    await wide.updateComplete;
    const lastButton = wide.shadowRoot?.querySelector<HTMLButtonElement>('.cell[data-index="26"]');
    const visibleColumns = [...(wide.shadowRoot?.querySelectorAll('.labels-row .label') ?? [])]
      .map(label => label.textContent?.trim());
    const visibleRows = [...(wide.shadowRoot?.querySelectorAll('.labels-col .label') ?? [])]
      .map(label => label.textContent?.trim());
    wide.remove();
    return {
      subscriptions,
      unsubscriptions,
      activations,
      collectionRefreshes,
      retryLabel,
      lastLabel: lastButton?.getAttribute('aria-label'),
      visibleColumns,
      visibleRows,
    };
  });
  expect(result).toMatchObject({
    subscriptions: 2,
    unsubscriptions: 2,
    activations: 1,
    collectionRefreshes: 0,
    lastLabel: 'AA1, empty',
    visibleRows: ['1'],
  });
  expect(result.retryLabel).toContain('Retry available: Network unavailable');
  expect(result.visibleColumns.at(-1)).toBe('AA');
});
