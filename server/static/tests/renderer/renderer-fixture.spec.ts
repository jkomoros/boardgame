import AxeBuilder from '@axe-core/playwright';
import { expect, test } from '@playwright/test';
import type { TargetAction } from '../../src/moves/target-action.js';
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
        player: handle.renderer.playerPresentation(0),
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
    expect(result.player).toEqual({ playerIndex: 0, label: 'Alice', color: '#7c3aed' });

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

test('chat preserves rejected drafts, retries visibly, and deduplicates notifications', async ({ page }) => {
  const diagnostics = await prepareRendererFixturePage(page);
  let postCount = 0;
  const postedBodies: string[] = [];
  await page.route('**/api/game/chatgame/GAME/chat*', async route => {
    if (route.request().method() === 'POST') {
      postCount += 1;
      postedBodies.push(route.request().postData() ?? '');
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(postCount === 1
          ? { Status: 'Failure', Error: 'blocked', FriendlyError: 'Chat is paused' }
          : { Status: 'Success', MessageID: '1' }),
      });
      return;
    }
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        Status: 'Success',
        Messages: postCount >= 2
          ? [{ id: '1', channel: 'all', sender: 0, body: 'Keep this draft', timestamp: 100 }]
          : null,
        ViewChannels: ['all'],
        SendChannels: ['all'],
        UserIDMap: { ada: 0 },
        ChatConfig: { Enabled: true, PrebakedOnly: false, AllowedMessages: null },
      }),
    });
  });

  await page.evaluate(async () => {
    await import('/src/components/boardgame-chat-panel.ts');
    const panel = document.createElement('boardgame-chat-panel') as HTMLElement & {
      gameRoute: { name: string; id: string };
      viewingAsPlayer: number;
      playersInfo: Array<{ IsEmpty: boolean; IsAgent: boolean; DisplayName: string }>;
      updateComplete: Promise<unknown>;
    };
    panel.gameRoute = { name: 'chatgame', id: 'GAME' };
    panel.viewingAsPlayer = 0;
    panel.playersInfo = [{ IsEmpty: false, IsAgent: false, DisplayName: 'Ada' }];
    document.body.append(panel);
    await panel.updateComplete;
  });

  const panel = page.locator('boardgame-chat-panel');
  await expect(panel.locator('.chat-container')).toBeVisible();
  const input = panel.locator('md-filled-text-field');
  await input.evaluate((element: Element) => {
    (element as Element & { value: string }).value = 'Keep this draft';
  });
  await panel.getByRole('button', { name: 'Send message' }).click();
  await expect(panel.getByRole('alert')).toContainText('Chat is paused');
  await expect.poll(() => input.evaluate((element: Element) => (
    element as Element & { value: string }
  ).value)).toBe('Keep this draft');

  await panel.getByRole('button', { name: 'Send message' }).click();
  await expect(panel.locator('.message')).toHaveCount(1);
  await expect(panel.locator('.message')).toContainText('Keep this draft');
  await expect.poll(() => input.evaluate((element: Element) => (
    element as Element & { value: string }
  ).value)).toBe('');

  await page.evaluate(() => window.dispatchEvent(new CustomEvent('chat-notification')));
  await expect(panel.locator('.message')).toHaveCount(1);
  expect(postedBodies.map(body => Object.fromEntries(new URLSearchParams(body)))).toEqual([
    { channel: 'all', body: 'Keep this draft' },
    { channel: 'all', body: 'Keep this draft' },
  ]);
  diagnostics.assertEmpty();
  diagnostics.stop();
});

test('game state manager isolates socket routes and owns the socket lifecycle', async ({ page }) => {
  const diagnostics = await prepareRendererFixturePage(page);
  try {
    const result = await page.evaluate(async () => {
      const NativeWebSocket = window.WebSocket;
      class FakeWebSocket {
        static readonly instances: FakeWebSocket[] = [];
        readonly url: string;
        readyState = 0;
        closeCount = 0;
        onclose: ((event: CloseEvent) => void) | null = null;
        onerror: ((event: Event) => void) | null = null;
        onmessage: ((event: MessageEvent) => void) | null = null;
        onopen: ((event: Event) => void) | null = null;

        constructor(url: string | URL) {
          this.url = String(url);
          FakeWebSocket.instances.push(this);
        }

        close() {
          this.closeCount += 1;
          this.readyState = 3;
        }

        send(_data: string | ArrayBufferLike | Blob | ArrayBufferView) {}
      }

      window.WebSocket = FakeWebSocket as unknown as typeof WebSocket;
      try {
        (globalThis as unknown as {
          CONFIG: { offline_dev_mode: boolean; firebase: Record<string, string> };
        }).CONFIG = {
          offline_dev_mode: true,
          firebase: {
            apiKey: 'fixture',
            authDomain: 'fixture.invalid',
            projectId: 'fixture',
            appId: '1:fixture:web:fixture',
          },
        };
        await import('/src/components/boardgame-game-state-manager.ts');
        type TestManager = HTMLElement & {
          _socketUrl: string;
          updateComplete: Promise<unknown>;
        };
        const manager = document.createElement('boardgame-game-state-manager') as TestManager;
        let activeEvents = 0;
        manager.addEventListener('socket-active', () => activeEvents += 1);

        document.body.append(manager);
        await manager.updateComplete;
        manager._socketUrl = 'ws://fixture.test/first';
        await manager.updateComplete;
        const first = FakeWebSocket.instances[0];
        if (!first) throw new Error('Initial attachment did not connect');

        manager._socketUrl = 'ws://fixture.test/second';
        await manager.updateComplete;
        const second = FakeWebSocket.instances[1];
        if (!second) throw new Error('Route replacement did not connect');

        // Browser callbacks already queued for the replaced connection must be inert.
        first.onmessage?.(new MessageEvent('message', { data: '{"type":"version","data":99}' }));
        first.onclose?.(new CloseEvent('close'));
        await new Promise(resolve => window.setTimeout(resolve, 300));
        const afterStaleCallbacks = FakeWebSocket.instances.length;

        second.onmessage?.(new MessageEvent('message', { data: '{"type":"version","data":1}' }));
        manager.remove();
        await new Promise(resolve => window.setTimeout(resolve, 300));
        const afterRemoval = FakeWebSocket.instances.length;

        // Reinsertion without a property change must restore the owned resource.
        document.body.append(manager);
        const third = FakeWebSocket.instances[2];
        if (!third) throw new Error('Reattachment did not reconnect');

        // Losing the route must close without reconnecting; restoring it may
        // connect once, and removal must cancel a natural-close retry.
        manager._socketUrl = '';
        await manager.updateComplete;
        await new Promise(resolve => window.setTimeout(resolve, 300));
        const afterRouteClear = FakeWebSocket.instances.length;
        manager._socketUrl = 'ws://fixture.test/second';
        await manager.updateComplete;
        const fourth = FakeWebSocket.instances[3];
        if (!fourth) throw new Error('Restored route did not reconnect');
        fourth.onclose?.(new CloseEvent('close'));
        manager.remove();
        await new Promise(resolve => window.setTimeout(resolve, 300));

        const nativeFetch = globalThis.fetch;
        globalThis.fetch = () => new Promise<Response>(() => undefined);
        const routed = document.createElement('boardgame-game-state-manager') as TestManager & {
          active: boolean;
          gameRoute: { name: string; id: string } | null;
          _infoInstalled: boolean;
          targetVersion: number;
        };
        routed.gameRoute = { name: 'alpha', id: 'A' };
        routed.active = true;
        document.body.append(routed);
        await routed.updateComplete;
        routed._infoInstalled = true;
        await routed.updateComplete;
        await routed.updateComplete;
        const alpha = FakeWebSocket.instances[4];
        if (!alpha) throw new Error('Reactive route did not connect');
        routed.gameRoute = { name: 'beta', id: 'B' };
        await routed.updateComplete;
        const targetAfterChange = routed.targetVersion;
        alpha.onmessage?.(new MessageEvent('message', { data: '{"type":"version","data":99}' }));
        await Promise.resolve();
        const targetAfterStaleFrame = routed.targetVersion;
        routed._infoInstalled = true;
        await routed.updateComplete;
        await routed.updateComplete;
        const beta = FakeWebSocket.instances[5];
        if (!beta) throw new Error('Replacement route did not reconnect');
        const routeUrls = [alpha.url, beta.url];
        const alphaCloseCount = alpha.closeCount;
        routed.remove();
        await new Promise(resolve => window.setTimeout(resolve, 0));
        globalThis.fetch = nativeFetch;

        return {
          urls: FakeWebSocket.instances.slice(0, 4).map(socket => socket.url),
          closeCounts: FakeWebSocket.instances.slice(0, 4).map(socket => socket.closeCount),
          activeEvents,
          afterStaleCallbacks,
          afterRemoval,
          afterRouteClear,
          routeUrls,
          alphaCloseCount,
          targetAfterChange,
          targetAfterStaleFrame,
        };
      } finally {
        window.WebSocket = NativeWebSocket;
      }
    });

    expect(result).toEqual({
      urls: [
        'ws://fixture.test/first',
        'ws://fixture.test/second',
        'ws://fixture.test/second',
        'ws://fixture.test/second',
      ],
      closeCounts: [1, 1, 1, 0],
      activeEvents: 1,
      afterStaleCallbacks: 2,
      afterRemoval: 2,
      afterRouteClear: 3,
      routeUrls: ['/api/game/alpha/A/socket', '/api/game/beta/B/socket'],
      alphaCloseCount: 1,
      targetAfterChange: -1,
      targetAfterStaleFrame: -1,
    });
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

test('removed component proposal configuration is inert or fails loudly', async ({ page }) => {
  await page.goto('/client_config.js');
  const result = await page.evaluate(async () => {
    await import('/src/components/boardgame-die.ts');
    await import('/src/components/boardgame-component-stack.ts');
    const die = document.createElement('boardgame-die') as HTMLElement & { updateComplete: Promise<unknown> };
    die.setAttribute('propose-move', 'Roll');
    document.body.append(die);
    await die.updateComplete;
    const dieButton = die.shadowRoot?.querySelector<HTMLButtonElement>('button');

    const errorFor = async (unsafeComponentAttrs: Record<string, unknown>) => {
      const stack = document.createElement('boardgame-component-stack') as HTMLElement & {
        unsafeComponentAttrs: Record<string, unknown>;
        updateComplete: Promise<unknown>;
      };
      stack.unsafeComponentAttrs = unsafeComponentAttrs;
      document.body.append(stack);
      try { await stack.updateComplete; return '<missing error>'; }
      catch (error) { return error instanceof Error ? error.message : String(error); }
      finally { stack.remove(); }
    };
    return {
      dieDisabled: dieButton?.disabled,
      dieLabel: dieButton?.getAttribute('aria-label'),
      errors: await Promise.all([
        errorFor({ proposeMove: 'Choose' }),
        errorFor({ indexAttributes: 'card-index' }),
        errorFor({ 'data-arg-card-index': 2 }),
      ]),
    };
  });
  expect(result).toEqual({
    dieDisabled: true,
    dieLabel: 'Die',
    errors: [
      expect.stringContaining('componentActions for moves'),
      expect.stringContaining('view.withProperties()'),
      expect.stringContaining('componentActions for moves'),
    ],
  });
});

test('target lists turn exact target collections into accessible previewed choices', async ({ page }) => {
  const diagnostics = await prepareRendererFixturePage(page);
  try {
    const result = await page.evaluate(async () => {
      await import('/game-src/tictactoe/boardgame-render-game-tictactoe.ts');
      const { tictactoeRendererFixture } = await import('/game-src/tictactoe/boardgame-render-fixtures-tictactoe.ts');
      const { mountRendererFixture } = await import('/src/testing/renderer-fixture.ts');
      const { html, targetList } = await import('/src/client.ts');
      const TicTacToeRenderer = customElements.get('boardgame-render-game-tictactoe');
      if (!TicTacToeRenderer) throw new Error('Tic-tac-toe renderer was not registered');

      class TargetListRenderer extends TicTacToeRenderer {
        override render() {
          const renderer = this as unknown as {
            move(name: 'Place Token'): {
              targets(keys: readonly number[], inputFor: (key: number) => { Slot: number }):
                TargetAction<number, 'Place Token', { Slot: number }>;
            };
          };
          const targets = renderer.move('Place Token').targets([1, 2], Slot => ({ Slot }));
          return html`<boardgame-target-list
            label="Choose a square"
            .choices=${targetList(targets, Slot => `Square ${Slot}`)}>
          </boardgame-target-list>`;
        }
      }
      customElements.define('boardgame-render-game-tictactoe-target-list', TargetListRenderer);
      const handle = await mountRendererFixture({
        ...tictactoeRendererFixture,
        tagName: 'boardgame-render-game-tictactoe-target-list',
      } as never);
      const list = handle.renderer.shadowRoot?.querySelector('boardgame-target-list') as (
        HTMLElement & {
          choices: ReturnType<typeof targetList>;
          updateComplete: Promise<unknown>;
        }
      ) | null;
      if (!list) throw new Error('Target list did not render');
      await list.updateComplete;
      await list.choices.target.ensurePreview();
      await list.updateComplete;
      const heading = list.shadowRoot?.querySelector('#heading');
      const controls = [...(list.shadowRoot?.querySelectorAll('boardgame-action-button') ?? [])];
      await Promise.all(controls.map(control => (control as HTMLElement & { updateComplete: Promise<unknown> }).updateComplete));
      const labels = controls.map(control => control.textContent?.trim());
      const activeButton = controls[1]?.shadowRoot?.querySelector<HTMLButtonElement>('button');
      activeButton?.focus();
      activeButton?.click();
      for (let attempt = 0; attempt < 20 && handle.proposals.length === 0; attempt++) {
        await new Promise(resolve => setTimeout(resolve, 0));
      }

      const malformed = document.createElement('boardgame-target-list') as HTMLElement & {
        choices: unknown;
        updateComplete: Promise<unknown>;
      };
      malformed.choices = { target: list.choices.target, choices: [] };
      document.body.append(malformed);
      let malformedError = '<resolved>';
      try { await malformed.updateComplete; }
      catch (error) { malformedError = error instanceof Error ? error.message : String(error); }
      malformed.remove();

      const configurationError = async (properties: Record<string, unknown>): Promise<string> => {
        const candidate = document.createElement('boardgame-target-list') as HTMLElement & {
          choices: ReturnType<typeof targetList>;
          updateComplete: Promise<unknown>;
        };
        candidate.choices = list.choices;
        Object.assign(candidate, properties);
        document.body.append(candidate);
        try { await candidate.updateComplete; return '<resolved>'; }
        catch (error) { return error instanceof Error ? error.message : String(error); }
        finally { candidate.remove(); }
      };

      let labelError = '<resolved>';
      try { targetList(list.choices.target, () => '   '); }
      catch (error) { labelError = error instanceof Error ? error.message : String(error); }
      const output = {
        heading: heading?.textContent?.trim(),
        ariaLevel: heading?.getAttribute('aria-level'),
        labels,
        proposals: handle.proposals,
        malformedError,
        labelError,
        configurationErrors: await Promise.all([
          configurationError({ label: ' ' }),
          configurationError({ headingLevel: 0 }),
          configurationError({ emptyLabel: '' }),
          configurationError({ layout: 'columns' }),
        ]),
      };
      (globalThis as unknown as { __targetListHandle: typeof handle }).__targetListHandle = handle;
      return output;
    });
    expect(result).toMatchObject({
      heading: 'Choose a square',
      ariaLevel: '2',
      labels: ['Square 1', 'Square 2'],
      proposals: [{ name: 'Place Token', arguments: { Slot: '2' } }],
      malformedError: expect.stringContaining('must come from targetList'),
      labelError: expect.stringContaining('must return a non-empty string'),
      configurationErrors: [
        expect.stringContaining('label must be a non-empty'),
        expect.stringContaining('headingLevel must be a safe integer'),
        expect.stringContaining('emptyLabel must be non-empty'),
        expect.stringContaining('layout must be "stack" or "grid"'),
      ],
    });
    const axeResult = await new AxeBuilder({ page }).include('[data-renderer-fixture]').analyze();
    expect(axeResult.violations).toEqual([]);
    await page.evaluate(() => {
      (globalThis as unknown as { __targetListHandle: { dispose(): void } }).__targetListHandle.dispose();
    });
    diagnostics.assertEmpty();
  } finally {
    diagnostics.stop();
  }
});

test('player badges consume explicit typed presentations without ambient store state', async ({ page }) => {
  await page.goto('/client_config.js');
  const result = await page.evaluate(async () => {
    await import('/src/client.ts');
    const badge = document.createElement('boardgame-player-badge') as HTMLElement & {
      player: { playerIndex: number; label: string; color?: string } | null;
      compact: boolean;
      updateComplete: Promise<unknown>;
    };
    badge.player = { playerIndex: 0, label: 'Alice', color: '#123456' };
    badge.compact = true;
    document.body.append(badge);
    await badge.updateComplete;
    const avatar = badge.shadowRoot?.querySelector<HTMLElement>('.avatar');

    const errorFor = async (player: unknown): Promise<string> => {
      const candidate = document.createElement('boardgame-player-badge') as HTMLElement & {
        player: unknown;
        updateComplete: Promise<unknown>;
      };
      candidate.player = player;
      document.body.append(candidate);
      try { await candidate.updateComplete; return '<resolved>'; }
      catch (error) { return error instanceof Error ? error.message : String(error); }
      finally { candidate.remove(); }
    };
    return {
      text: avatar?.textContent?.trim(),
      role: avatar?.getAttribute('role'),
      label: avatar?.getAttribute('aria-label'),
      color: avatar?.style.backgroundColor,
      errors: await Promise.all([
        errorFor(null),
        errorFor({ playerIndex: -1, label: 'Alice' }),
        errorFor({ playerIndex: 0, label: '' }),
        errorFor({ playerIndex: 0, label: 'Alice', color: 'not a color(' }),
      ]),
    };
  });
  expect(result).toEqual({
    text: 'A',
    role: 'img',
    label: 'Alice',
    color: 'rgb(18, 52, 86)',
    errors: [
      expect.stringContaining('.player must come from renderer.playerPresentation'),
      expect.stringContaining('playerIndex must be a non-negative'),
      expect.stringContaining('player label must be a non-empty'),
      expect.stringContaining('player color is not valid CSS'),
    ],
  });
  const axeResult = await new AxeBuilder({ page }).include('boardgame-player-badge').analyze();
  expect(axeResult.violations).toEqual([]);
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
      stack.componentView = view.withProperties({ faceUp: true }).withProperties({ rotated: false });
      await stack.updateComplete;
      const chainedCard = stack.querySelector('boardgame-card') as
        (HTMLElement & { faceUp: boolean; rotated: boolean; updateComplete: Promise<unknown> }) | null;
      await chainedCard?.updateComplete;
      const chainedBinding = { faceUp: chainedCard?.faceUp, rotated: chainedCard?.rotated };
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
        fauxBinding, chainedBinding, displayOnlyDisabled, fauxDisplayOnlyDisabled, customText, missingViewError,
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
    expect(result.chainedBinding).toEqual({ faceUp: true, rotated: false });
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

test('placement draft controls make keyboard drafting, undo, rebase, and exact commit visible', async ({ page }) => {
  await page.goto('/client_config.js');
  const result = await page.evaluate(async () => {
    const { PlacementDraftController } = await import('/src/moves/placement-draft.ts');
    const { MoveSubmissionGate, createMoveAction } = await import('/src/moves/action.ts');
    await import('/src/components/boardgame-draft-controls.ts');
    await import('/src/components/boardgame-placement-item.ts');
    const controls = document.createElement('boardgame-draft-controls');
    controls.style.display = 'block';
    controls.style.width = '300px';
    controls.label = 'Word placement';
    controls.commitLabel = 'Play word';
    let submissions = 0;
    let request = 0;
    let placementItems: HTMLElementTagNameMap['boardgame-placement-item'][] = [];
    const gate = new MoveSubmissionGate();
    const host = {
      state: {}, gameName: 'words', gameId: 'one', gameVersion: 1, snapshotEpoch: 1,
      viewingAsPlayer: 0, proposingAsPlayer: 0, proposingAsAdmin: false,
      addController: () => undefined,
      removeController: () => undefined,
      requestUpdate: () => {
        controls.draft = controller.bind(options);
        placementItems.forEach((element, index) => {
          element.item = controls.draft.item(index === 0 ? 'a' : 'c');
        });
      },
      updateComplete: Promise.resolve(true),
    };
    const snapshotKey = () => [
      host.gameName, host.gameId, host.gameVersion, host.snapshotEpoch,
      host.viewingAsPlayer, host.proposingAsPlayer, host.proposingAsAdmin ? 1 : 0,
    ].join('\u0000');
    const service = {
      currentClientSchemaFingerprint: () => 'schema',
      currentServerSchemaFingerprint: () => 'schema',
      currentTransport: () => ({ submit: async () => { submissions++; return { kind: 'success' as const }; } }),
      currentPreviewTransport: () => ({ preview: async () => ({ kind: 'success' as const, legal: true }) }),
      currentTargetPreviewTransport: () => null,
      currentGate: () => gate,
      nextRequestID: () => `draft-${++request}`,
      validate: () => [],
      serialize: (_name: string, input: unknown) => ({
        Placements: String((input as { Placements: string }).Placements),
      }),
      changed: () => controls.requestUpdate(),
    };
    const controller = new PlacementDraftController<string, number>(host);
    const options = {
      items: ['a', 'b', 'c'], targets: [0, 1], minPlacements: 2, maxPlacements: 2,
      action: (placements: readonly { item: string; target: number }[]) => {
        const key = snapshotKey();
        return createMoveAction('Play', service, {
          snapshotKey: key,
          currentSnapshotKey: snapshotKey,
          snapshotVersion: host.gameVersion,
          currentSnapshotVersion: () => host.gameVersion,
          viewingAsPlayer: 0, proposingAsPlayer: 0, proposingAsAdmin: false,
          currentLegality: () => ({ legalForPlayer: true, legalForAnyone: true }),
          currentAnimating: () => false,
          baselineLegalityApplies: true,
        }).with({ Placements: JSON.stringify(placements) });
      },
    };
    controls.draft = controller.bind(options);
    const rack = document.createElement('div');
    rack.setAttribute('aria-label', 'Letter rack');
    const itemA = document.createElement('boardgame-placement-item');
    itemA.label = 'Letter A'; itemA.item = controls.draft.item('a');
    const itemC = document.createElement('boardgame-placement-item');
    itemC.label = 'Letter C'; itemC.item = controls.draft.item('c');
    placementItems = [itemA, itemC];
    rack.append(itemA, itemC);
    document.body.append(rack, controls);
    await controls.updateComplete;
    await Promise.all(placementItems.map(item => item.updateComplete));
    const root = controls.shadowRoot!;
    const initialStatus = root.querySelector('#status')?.textContent?.trim();
    const initialCommit = root.querySelector('boardgame-action-button')!;
    await (initialCommit as typeof initialCommit & { updateComplete: Promise<unknown> }).updateComplete;
    const initialReason = initialCommit.shadowRoot?.querySelector('#status')?.textContent?.trim();

    const itemAButton = itemA.shadowRoot!.querySelector<HTMLButtonElement>('button')!;
    const itemCButton = itemC.shadowRoot!.querySelector<HTMLButtonElement>('button')!;
    const itemHitTarget = itemAButton.getBoundingClientRect();
    itemAButton.click();
    await itemA.updateComplete;
    const itemSelected = itemAButton.getAttribute('aria-pressed');
    controls.draft.place(0);
    await itemA.updateComplete;
    const itemPlaced = itemA.shadowRoot!.querySelector('#status')?.textContent;
    controls.draft.assign('b', 1);
    await controls.updateComplete;
    await itemC.updateComplete;
    const itemCapacityDisabled = itemCButton.disabled;
    const completeCount = root.querySelector('#count')?.textContent?.replace(/\s+/g, ' ').trim();
    const undo = root.querySelector<HTMLButtonElement>('button[part="undo"]')!;
    undo.click();
    await controls.updateComplete;
    const undoneCount = controls.draft.placements.length;
    controls.draft.assign('b', 1);
    await controls.updateComplete;
    const commit = root.querySelector('boardgame-action-button') as HTMLElement & {
      updateComplete: Promise<unknown>;
    };
    await commit.updateComplete;
    commit.shadowRoot?.querySelector<HTMLButtonElement>('button')?.click();
    await new Promise(resolve => setTimeout(resolve, 0));

    host.state = {};
    host.gameVersion++;
    controls.draft = controller.bind(options);
    await controls.updateComplete;
    const notice = root.querySelector('#notice')?.textContent?.replace(/\s+/g, ' ').trim();
    const clearedCount = controls.draft.placements.length;
    const direction = getComputedStyle(root.querySelector('#buttons')!).flexDirection;

    const invalid = document.createElement('boardgame-draft-controls');
    document.body.append(invalid);
    let invalidError = '';
    try { await invalid.updateComplete; } catch (error) {
      invalidError = error instanceof Error ? error.message : String(error);
    }
    invalid.remove();
    return {
      initialStatus, initialReason, completeCount, undoneCount, submissions,
      notice, clearedCount, direction, invalidError, itemSelected, itemPlaced,
      itemCapacityDisabled, itemWidth: itemHitTarget.width, itemHeight: itemHitTarget.height,
    };
  });
  expect(result).toEqual({
    initialStatus: '0 placements drafted.',
    initialReason: 'Add 2 more before committing',
    completeCount: '2 / 2',
    undoneCount: 1,
    submissions: 1,
    notice: 'The game changed, so the local draft was cleared. Dismiss',
    clearedCount: 0,
    direction: 'column',
    invalidError: expect.stringContaining('.draft must be a placement or selection draft binding'),
    itemSelected: 'true',
    itemPlaced: 'Placed',
    itemCapacityDisabled: true,
    itemWidth: expect.any(Number),
    itemHeight: expect.any(Number),
  });
  expect(result.itemWidth).toBeGreaterThanOrEqual(44);
  expect(result.itemHeight).toBeGreaterThanOrEqual(44);
  const axeResult = await new AxeBuilder({ page })
    .include('[aria-label="Letter rack"]')
    .include('boardgame-draft-controls')
    .analyze();
  expect(axeResult.violations).toEqual([]);
});

test('selection options own pressed state, capacity, hit targets, and content safety', async ({ page }) => {
  await page.goto('/client_config.js');
  const result = await page.evaluate(async () => {
    await import('/src/components/boardgame-selection-option.ts');
    const region = document.createElement('div');
    region.setAttribute('aria-label', 'Resource cards');
    let selected: readonly string[] = [];
    const clay = document.createElement('boardgame-selection-option');
    clay.label = 'Clay card';
    const art = document.createElement('span');
    art.textContent = '🧱 Clay';
    clay.append(art);
    const ore = document.createElement('boardgame-selection-option');
    ore.label = 'Ore card';
    const options = [clay, ore];
    const refresh = () => {
      const bind = (choice: string) => ({
        choice,
        selected: selected.includes(choice),
        capacityBlocked: !selected.includes(choice) && selected.length >= 1,
        toggle: () => {
          selected = selected.includes(choice) ? [] : [choice];
          refresh();
        },
      });
      clay.option = bind('clay');
      ore.option = bind('ore');
    };
    refresh();
    region.append(...options);
    document.body.append(region);
    await Promise.all(options.map(option => option.updateComplete));
    const clayButton = clay.shadowRoot!.querySelector<HTMLButtonElement>('button')!;
    const oreButton = ore.shadowRoot!.querySelector<HTMLButtonElement>('button')!;
    const initial = {
      pressed: clayButton.getAttribute('aria-pressed'),
      label: clayButton.getAttribute('aria-label'),
      width: clayButton.getBoundingClientRect().width,
      height: clayButton.getBoundingClientRect().height,
      fallback: ore.shadowRoot!.querySelector('[part="fallback"]')?.textContent,
    };
    clayButton.click();
    await Promise.all(options.map(option => option.updateComplete));
    const atCapacity = {
      clayPressed: clayButton.getAttribute('aria-pressed'),
      clayDisabled: clayButton.disabled,
      oreDisabled: oreButton.disabled,
    };
    clayButton.click();
    await Promise.all(options.map(option => option.updateComplete));
    const deselected = { oreDisabled: oreButton.disabled, selected };

    const malformed = document.createElement('boardgame-selection-option');
    malformed.label = 'Malformed'; malformed.option = {
      choice: 'clay', selected: true, capacityBlocked: true, toggle: () => {},
    };
    document.body.append(malformed);
    let malformedError = '';
    try { await malformed.updateComplete; } catch (error) {
      malformedError = error instanceof Error ? error.message : String(error);
    }
    malformed.remove();

    const nested = document.createElement('boardgame-selection-option');
    nested.label = 'Nested'; nested.option = clay.option;
    nested.append(document.createElement('button'));
    document.body.append(nested);
    let nestedError = '';
    try { await nested.updateComplete; } catch (error) {
      nestedError = error instanceof Error ? error.message : String(error);
    }
    nested.remove();
    return { initial, atCapacity, deselected, malformedError, nestedError };
  });
  expect(result).toMatchObject({
    initial: { pressed: 'false', label: 'Clay card', fallback: 'Ore card' },
    atCapacity: { clayPressed: 'true', clayDisabled: false, oreDisabled: true },
    deselected: { oreDisabled: false, selected: [] },
    malformedError: expect.stringContaining('.option is malformed'),
    nestedError: expect.stringContaining('cannot contain interactive content'),
  });
  expect(result.initial.width).toBeGreaterThanOrEqual(44);
  expect(result.initial.height).toBeGreaterThanOrEqual(44);
  const axeResult = await new AxeBuilder({ page })
    .include('[aria-label="Resource cards"]')
    .analyze();
  expect(axeResult.violations).toEqual([]);
});

test('inspector supplies modal focus, dismissal, mobile sizing, and loud content contracts', async ({ page }) => {
  await page.goto('/client_config.js');
  await page.setViewportSize({ width: 320, height: 640 });
  const result = await page.evaluate(async () => {
    await import('/src/components/boardgame-inspector.ts');
    const inspector = document.createElement('boardgame-inspector');
    inspector.label = 'Moon vision';
    inspector.description = 'A moonlit path through a forest';
    const thumbnail = document.createElement('span');
    thumbnail.slot = 'thumbnail';
    thumbnail.textContent = '🌙';
    const detail = document.createElement('article');
    detail.slot = 'detail';
    detail.textContent = 'Large moonlit vision';
    inspector.append(thumbnail, detail);
    const events: { open: boolean; reason: string }[] = [];
    inspector.addEventListener('inspector-open-changed', event => {
      events.push((event as CustomEvent<{ open: boolean; reason: string }>).detail);
    });
    document.body.append(inspector);
    await inspector.updateComplete;
    const root = inspector.shadowRoot!;
    const trigger = root.querySelector<HTMLButtonElement>('#trigger')!;
    trigger.click();
    await inspector.updateComplete;
    const dialog = root.querySelector<HTMLDialogElement>('dialog')!;
    const open = dialog.open && dialog.matches(':modal');
    const labelledBy = dialog.getAttribute('aria-labelledby');
    const describedBy = dialog.getAttribute('aria-describedby');
    const activeWhileOpen = root.activeElement?.id;
    const bounds = dialog.getBoundingClientRect();
    const escaped = new Promise<void>(resolve => dialog.addEventListener('close', () => resolve(), { once: true }));
    dialog.dispatchEvent(new Event('cancel', { bubbles: false, cancelable: true }));
    dialog.close();
    await escaped;
    await Promise.resolve();
    const focusedAfterEscape = root.activeElement?.id;

    inspector.show();
    await inspector.updateComplete;
    const backdropClosed = new Promise<void>(resolve => dialog.addEventListener('close', () => resolve(), { once: true }));
    dialog.dispatchEvent(new MouseEvent('click', { bubbles: true, clientX: 1, clientY: 1 }));
    await backdropClosed;

    inspector.dismissible = false;
    inspector.show();
    await inspector.updateComplete;
    const cancel = new Event('cancel', { bubbles: false, cancelable: true });
    dialog.dispatchEvent(cancel);
    const escapePrevented = cancel.defaultPrevented;
    const programmaticallyClosed = new Promise<void>(resolve => dialog.addEventListener('close', () => resolve(), { once: true }));
    inspector.close();
    await programmaticallyClosed;

    const missing = document.createElement('boardgame-inspector');
    missing.label = 'Missing detail';
    document.body.append(missing);
    await missing.updateComplete;
    missing.open = true;
    let missingError = '';
    try { await missing.updateComplete; } catch (error) {
      missingError = error instanceof Error ? error.message : String(error);
    }
    await missing.updateComplete.catch(() => undefined);
    missing.remove();

    const badLabel = document.createElement('boardgame-inspector');
    badLabel.label = '   ';
    document.body.append(badLabel);
    let labelError = '';
    try { await badLabel.updateComplete; } catch (error) {
      labelError = error instanceof Error ? error.message : String(error);
    }
    badLabel.remove();

    const nestedControl = document.createElement('boardgame-inspector');
    nestedControl.label = 'Nested control';
    const nestedButton = document.createElement('button');
    nestedButton.slot = 'thumbnail';
    const nestedDetail = document.createElement('span');
    nestedDetail.slot = 'detail';
    nestedControl.append(nestedButton, nestedDetail);
    document.body.append(nestedControl);
    let thumbnailError = '';
    try { await nestedControl.updateComplete; } catch (error) {
      thumbnailError = error instanceof Error ? error.message : String(error);
    }
    nestedControl.remove();
    inspector.dismissible = true;
    inspector.show();
    await inspector.updateComplete;
    return {
      open, labelledBy, describedBy, activeWhileOpen, focusedAfterEscape,
      mobileWidth: bounds.width, mobileBottom: Math.abs(bounds.bottom - 640),
      escapePrevented, events, missingError, labelError, thumbnailError,
    };
  });
  expect(result).toMatchObject({
    open: true,
    labelledBy: 'title',
    describedBy: 'description',
    activeWhileOpen: 'close',
    focusedAfterEscape: 'trigger',
    mobileWidth: 320,
    mobileBottom: 0,
    escapePrevented: true,
    missingError: expect.stringContaining('provide non-empty slot="detail" content'),
    labelError: expect.stringContaining('label must be a non-empty visible dialog title'),
    thumbnailError: expect.stringContaining('cannot contain interactive content'),
  });
  expect(result.events).toEqual([
    { open: true, reason: 'trigger' },
    { open: false, reason: 'escape' },
    { open: true, reason: 'programmatic' },
    { open: false, reason: 'backdrop' },
    { open: true, reason: 'programmatic' },
    { open: false, reason: 'programmatic' },
    { open: true, reason: 'programmatic' },
  ]);
  const axeResult = await new AxeBuilder({ page })
    .include('boardgame-inspector')
    .analyze();
  expect(axeResult.violations).toEqual([]);
});

test('readiness makes simultaneous public progress accessible and rejects ambiguous state', async ({ page }) => {
  await page.goto('/client_config.js');
  const result = await page.evaluate(async () => {
    await import('/src/components/boardgame-readiness.ts');
    const readiness = document.createElement('boardgame-readiness');
    readiness.label = 'Day votes';
    readiness.completeLabel = 'All votes cast';
    readiness.progressLabel = 'votes cast';
    readiness.readyLabel = 'Voted';
    readiness.waitingLabel = 'Thinking';
    readiness.notRequiredLabel = 'Eliminated';
    readiness.participants = [
      { key: 0, label: 'Ada', state: 'ready' },
      { key: 1, label: 'Grace', state: 'waiting' },
      { key: 2, label: 'Linus', state: 'not-required' },
    ];
    document.body.append(readiness);
    await readiness.updateComplete;
    const root = readiness.shadowRoot!;
    const initial = {
      heading: root.querySelector('#heading')?.textContent,
      summary: root.querySelector('#summary')?.textContent,
      progress: root.querySelector<HTMLProgressElement>('progress')?.value,
      maximum: root.querySelector<HTMLProgressElement>('progress')?.max,
      states: [...root.querySelectorAll<HTMLElement>('li')].map(item => ({
        state: item.dataset.state,
        text: item.textContent?.replace(/\s+/g, ' ').trim(),
      })),
    };
    readiness.participants = [
      { key: 0, label: 'Ada', state: 'ready' },
      { key: 1, label: 'Grace', state: 'ready' },
      { key: 2, label: 'Linus', state: 'not-required' },
    ];
    await readiness.updateComplete;
    const complete = {
      summary: root.querySelector('#summary')?.textContent,
      surface: root.querySelector('#surface')?.getAttribute('data-complete'),
    };
    readiness.view = 'summary';
    await readiness.updateComplete;
    const summaryHidesList = root.querySelector('ul') === null;

    const duplicate = document.createElement('boardgame-readiness');
    duplicate.participants = [
      { key: 'same', label: 'Ada', state: 'ready' },
      { key: 'same', label: 'Grace', state: 'waiting' },
    ];
    document.body.append(duplicate);
    let duplicateError = '';
    try { await duplicate.updateComplete; } catch (error) {
      duplicateError = error instanceof Error ? error.message : String(error);
    }
    duplicate.remove();

    const invalidView = document.createElement('boardgame-readiness');
    invalidView.view = 'dense' as never;
    document.body.append(invalidView);
    let viewError = '';
    try { await invalidView.updateComplete; } catch (error) {
      viewError = error instanceof Error ? error.message : String(error);
    }
    invalidView.remove();
    readiness.view = 'list';
    await readiness.updateComplete;
    return { initial, complete, summaryHidesList, duplicateError, viewError };
  });
  expect(result).toEqual({
    initial: {
      heading: 'Day votes', summary: '1 of 2 votes cast', progress: 1, maximum: 2,
      states: [
        { state: 'ready', text: 'Ada Voted' },
        { state: 'waiting', text: 'Grace Thinking' },
        { state: 'not-required', text: 'Linus Eliminated' },
      ],
    },
    complete: { summary: 'All votes cast', surface: 'true' },
    summaryHidesList: true,
    duplicateError: expect.stringContaining('duplicate participant key'),
    viewError: expect.stringContaining('view must be "list" or "summary"'),
  });
  const axeResult = await new AxeBuilder({ page })
    .include('boardgame-readiness')
    .analyze();
  expect(axeResult.violations).toEqual([]);
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

test('game outcome waits for settled animation and renders public or personal verdicts', async ({ page }) => {
  const diagnostics = await prepareRendererFixturePage(page);
  try {
    await page.emulateMedia({ reducedMotion: 'reduce' });
    const result = await page.evaluate(async () => {
      await import('/src/client.ts');
      const outcome = document.createElement('boardgame-game-outcome');
      outcome.finished = true;
      outcome.animating = true;
      outcome.winners = [0];
      outcome.winnerLabels = ['Ada'];
      document.body.append(outcome);
      await outcome.updateComplete;
      const gated = outcome.shadowRoot?.querySelector('#outcome') === null;

      outcome.animating = false;
      await outcome.updateComplete;
      const section = outcome.shadowRoot?.querySelector('#outcome');
      const publicVerdict = {
        text: section?.textContent?.replace(/\s+/g, ' ').trim(),
        role: section?.getAttribute('role'),
        live: section?.getAttribute('aria-live'),
        atomic: section?.getAttribute('aria-atomic'),
        part: section?.getAttribute('part'),
        animation: section ? getComputedStyle(section).animationName : null,
      };

      outcome.viewer = 0;
      await outcome.updateComplete;
      const personalWin = outcome.shadowRoot?.querySelector('#message')?.textContent?.trim();
      outcome.viewer = 1;
      await outcome.updateComplete;
      const personalLoss = outcome.shadowRoot?.querySelector('#message')?.textContent?.trim();
      outcome.winners = [];
      outcome.winnerLabels = [];
      await outcome.updateComplete;
      const draw = outcome.shadowRoot?.querySelector('#message')?.textContent?.trim();

      const renderError = (configure: (element: HTMLElement & Record<string, unknown>) => void) => {
        const element = document.createElement('boardgame-game-outcome') as HTMLElement &
          Record<string, unknown> & { render(): unknown };
        configure(element);
        try {
          element.render();
          return '<resolved>';
        } catch (error) {
          return error instanceof Error ? error.message : String(error);
        }
      };
      const unfinishedWinner = renderError(element => { element.winners = [0]; });
      const duplicateWinner = renderError(element => { element.finished = true; element.winners = [0, 0]; });
      const invalidViewer = renderError(element => { element.viewer = -1; });
      const mismatchedLabels = renderError(element => {
        element.finished = true;
        element.winners = [0, 1];
        element.winnerLabels = ['Ada'];
      });
      const blankTitle = renderError(element => { element.title = ' '; });
      return {
        gated,
        publicVerdict,
        personalWin,
        personalLoss,
        draw,
        unfinishedWinner,
        duplicateWinner,
        invalidViewer,
        mismatchedLabels,
        blankTitle,
      };
    });

    expect(result.gated).toBe(true);
    expect(result.publicVerdict).toEqual({
      text: 'Game over! Ada wins!',
      role: 'status',
      live: 'polite',
      atomic: 'true',
      part: 'outcome',
      animation: 'none',
    });
    expect(result.personalWin).toBe('You won!');
    expect(result.personalLoss).toBe('You lost.');
    expect(result.draw).toBe("It's a draw.");
    expect(result.unfinishedWinner).toContain('winners cannot be present before finished is true');
    expect(result.duplicateWinner).toContain('duplicate winner index');
    expect(result.invalidViewer).toContain('viewer must be null or a nonnegative safe player index');
    expect(result.mismatchedLabels).toContain('exactly one label per winner');
    expect(result.blankTitle).toContain('title must be non-empty');
    const axeResult = await new AxeBuilder({ page }).include('boardgame-game-outcome').analyze();
    expect(axeResult.violations).toEqual([]);
    diagnostics.assertEmpty();
  } finally {
    diagnostics.stop();
  }
});

test('game surface supplies a semantic responsive shell with optional named regions', async ({ page }) => {
  const diagnostics = await prepareRendererFixturePage(page);
  try {
    const result = await page.evaluate(async () => {
      await import('/src/client.ts');
      const surface = document.createElement('boardgame-game-surface');
      surface.heading = 'Fixture game';
      surface.style.width = '50rem';
      surface.style.setProperty('--boardgame-game-surface-max-width', '34rem');

      const headerAction = document.createElement('button');
      headerAction.slot = 'header';
      headerAction.textContent = 'Rules';
      const content = document.createElement('div');
      content.textContent = 'Board';
      const status = document.createElement('p');
      status.slot = 'status';
      status.textContent = 'Your turn';
      status.setAttribute('role', 'status');
      const actions = document.createElement('div');
      actions.slot = 'actions';
      actions.textContent = 'Actions';
      const footer = document.createElement('small');
      footer.slot = 'footer';
      footer.textContent = 'Round 1';
      surface.append(headerAction, content, status, actions, footer);
      document.body.append(surface);
      await surface.updateComplete;
      await new Promise<void>(resolve => requestAnimationFrame(() => resolve()));
      await surface.updateComplete;

      const root = surface.shadowRoot;
      const shell = root?.querySelector('#surface') as HTMLElement | null;
      const header = root?.querySelector('#header') as HTMLElement | null;
      const heading = root?.querySelector('#heading');
      const wideDirection = header ? getComputedStyle(header).flexDirection : '';
      const initial = {
        labelReference: shell?.getAttribute('aria-labelledby'),
        heading: heading?.textContent?.trim(),
        headingRole: heading?.getAttribute('role'),
        headingLevel: heading?.getAttribute('aria-level'),
        surfacePart: shell?.getAttribute('part'),
        contentPart: root?.querySelector('#content')?.getAttribute('part'),
        statusHidden: (root?.querySelector('#status') as HTMLElement | null)?.hidden,
        actionsHidden: (root?.querySelector('#actions') as HTMLElement | null)?.hidden,
        footerHidden: (root?.querySelector('#footer') as HTMLElement | null)?.hidden,
        shellWidth: shell?.getBoundingClientRect().width ?? 0,
      };

      surface.style.width = '20rem';
      await new Promise<void>(resolve => requestAnimationFrame(() => resolve()));
      const narrowDirection = header ? getComputedStyle(header).flexDirection : '';
      actions.remove();
      footer.remove();
      await new Promise<void>(resolve => requestAnimationFrame(() => resolve()));
      await surface.updateComplete;
      const emptied = {
        actionsHidden: (root?.querySelector('#actions') as HTMLElement | null)?.hidden,
        footerHidden: (root?.querySelector('#footer') as HTMLElement | null)?.hidden,
      };

      surface.hideHeading = true;
      await surface.updateComplete;
      const hiddenHeadingClass = root?.querySelector('#heading')?.getAttribute('class');

      const renderError = (name: string, value: unknown) => {
        const element = document.createElement('boardgame-game-surface') as HTMLElement &
          Record<string, unknown> & { render(): unknown };
        element.heading = 'Valid';
        element[name] = value;
        try {
          element.render();
          return '<resolved>';
        } catch (error) {
          return error instanceof Error ? error.message : String(error);
        }
      };

      return {
        initial,
        wideDirection,
        narrowDirection,
        emptied,
        hiddenHeadingClass,
        blankHeading: renderError('heading', ' '),
        invalidHeadingLevel: renderError('headingLevel', 7),
      };
    });

    expect(result.initial).toMatchObject({
      labelReference: 'heading',
      heading: 'Fixture game',
      headingRole: 'heading',
      headingLevel: '2',
      surfacePart: 'surface',
      contentPart: 'content',
      statusHidden: false,
      actionsHidden: false,
      footerHidden: false,
    });
    expect(result.initial.shellWidth).toBeCloseTo(34 * 16, 0);
    expect(result.wideDirection).toBe('row');
    expect(result.narrowDirection).toBe('column');
    expect(result.emptied).toEqual({ actionsHidden: true, footerHidden: true });
    expect(result.hiddenHeadingClass).toBe('visually-hidden');
    expect(result.blankHeading).toContain('heading must be a non-empty game name');
    expect(result.invalidHeadingLevel).toContain('headingLevel must be a safe integer from 1 through 6');
    const axeResult = await new AxeBuilder({ page }).include('boardgame-game-surface').analyze();
    expect(axeResult.violations).toEqual([]);
    diagnostics.assertEmpty();
  } finally {
    diagnostics.stop();
  }
});

test('turn status names ordinary, observer, admin, and simultaneous perspectives honestly', async ({ page }) => {
  const diagnostics = await prepareRendererFixturePage(page);
  try {
    const result = await page.evaluate(async () => {
      const { AdminPlayerIndex, AnyPlayerIndex, ObserverPlayerIndex } = await import('/src/client.ts');
      const status = document.createElement('boardgame-turn-status');
      document.body.append(status);
      const show = async (currentPlayerIndex: number, viewerPlayerIndex: number, extras = {}) => {
        status.turn = {
          currentPlayerIndex,
          viewerPlayerIndex,
          finished: false,
          animating: false,
          ...extras,
        };
        await status.updateComplete;
        const element = status.shadowRoot?.querySelector('#status');
        return element ? {
          text: element.textContent?.trim(),
          part: element.getAttribute('part'),
          role: element.getAttribute('role'),
          live: element.getAttribute('aria-live'),
          atomic: element.getAttribute('aria-atomic'),
        } : null;
      };

      const active = await show(0, 0);
      status.playerLabels = ['Ada', 'Grace'];
      const waiting = await show(1, 0);
      const observer = await show(0, ObserverPlayerIndex);
      const admin = await show(0, AdminPlayerIndex);
      const simultaneousPlayer = await show(AnyPlayerIndex, 0);
      const simultaneousObserver = await show(AnyPlayerIndex, ObserverPlayerIndex);
      const animating = await show(0, 0, { animating: true });
      const finished = await show(0, 0, { finished: true });

      const renderError = (turn: unknown, playerLabels: unknown = []) => {
        const element = document.createElement('boardgame-turn-status') as HTMLElement & {
          turn: unknown;
          playerLabels: unknown;
          render(): unknown;
        };
        element.turn = turn;
        element.playerLabels = playerLabels;
        try {
          element.render();
          return '<resolved>';
        } catch (error) {
          return error instanceof Error ? error.message : String(error);
        }
      };
      const base = { currentPlayerIndex: 0, viewerPlayerIndex: 0, finished: false, animating: false };
      const invalidActiveLabel = document.createElement('boardgame-turn-status') as HTMLElement & {
        turn: typeof base;
        activeLabel: unknown;
        render(): unknown;
      };
      invalidActiveLabel.turn = base;
      invalidActiveLabel.activeLabel = false;
      let activeLabelError = '<resolved>';
      try {
        invalidActiveLabel.render();
      } catch (error) {
        activeLabelError = error instanceof Error ? error.message : String(error);
      }
      return {
        active,
        waiting,
        observer,
        admin,
        simultaneousPlayer,
        simultaneousObserver,
        animating,
        finished,
        missingField: renderError({ currentPlayerIndex: 0, viewerPlayerIndex: 0 }),
        invalidViewer: renderError({ ...base, viewerPlayerIndex: AnyPlayerIndex }),
        unknownField: renderError({ ...base, phase: 'Playing' }),
        blankLabel: renderError(base, ['']),
        activeLabelError,
      };
    });

    expect(result.active).toEqual({
      text: 'Your turn',
      part: 'status active',
      role: 'status',
      live: 'polite',
      atomic: 'true',
    });
    expect(result.waiting?.text).toBe("Grace's turn");
    expect(result.waiting?.part).toBe('status waiting');
    expect(result.observer?.text).toBe("Ada's turn");
    expect(result.admin?.text).toBe("Ada's turn");
    expect(result.simultaneousPlayer?.text).toBe('Your turn');
    expect(result.simultaneousObserver).toMatchObject({
      text: 'All players may act',
      part: 'status simultaneous',
    });
    expect(result.animating).toBeNull();
    expect(result.finished).toBeNull();
    expect(result.missingField).toContain('must contain exactly');
    expect(result.invalidViewer).toContain('viewerPlayerIndex must be a concrete player');
    expect(result.unknownField).toContain('must contain exactly');
    expect(result.blankLabel).toContain('playerLabels must contain only non-empty strings');
    expect(result.activeLabelError).toContain('activeLabel must be a non-empty string');

    await page.evaluate(() => {
      const status = document.querySelector('boardgame-turn-status');
      if (!status) throw new Error('turn status disappeared');
      status.turn = { currentPlayerIndex: 0, viewerPlayerIndex: 0, finished: false, animating: false };
    });
    const axeResult = await new AxeBuilder({ page }).include('boardgame-turn-status').analyze();
    expect(axeResult.violations).toEqual([]);
    diagnostics.assertEmpty();
  } finally {
    diagnostics.stop();
  }
});

test('player grid supplies named responsive layout and a useful empty state', async ({ page }) => {
  const diagnostics = await prepareRendererFixturePage(page);
  try {
    const result = await page.evaluate(async () => {
      await import('/src/client.ts');
      const grid = document.createElement('boardgame-player-grid');
      grid.style.width = '50rem';
      document.body.append(grid);
      await grid.updateComplete;
      const region = grid.shadowRoot?.querySelector('#region');
      const heading = grid.shadowRoot?.querySelector('#heading');
      const layout = grid.shadowRoot?.querySelector('#grid') as HTMLElement | null;
      const empty = {
        label: heading?.textContent?.trim(),
        regionLabel: region?.getAttribute('aria-labelledby'),
        headingRole: heading?.getAttribute('role'),
        headingLevel: heading?.getAttribute('aria-level'),
        text: grid.shadowRoot?.querySelector('#empty')?.textContent?.trim(),
        part: grid.shadowRoot?.querySelector('#empty')?.getAttribute('part'),
      };

      for (let index = 0; index < 3; index += 1) {
        const player = document.createElement('section');
        player.textContent = `Player ${index + 1}`;
        grid.append(player);
      }
      await new Promise<void>(resolve => requestAnimationFrame(() => resolve()));
      await grid.updateComplete;
      const wideColumns = layout ? getComputedStyle(layout).gridTemplateColumns.split(' ').length : 0;
      const emptyAfterPlayers = grid.shadowRoot?.querySelector('#empty') !== null;
      grid.style.width = '14rem';
      await new Promise<void>(resolve => requestAnimationFrame(() => resolve()));
      const narrowColumns = layout ? getComputedStyle(layout).gridTemplateColumns.split(' ').length : 0;
      grid.hideHeading = true;
      await grid.updateComplete;
      const hiddenHeadingClass = grid.shadowRoot?.querySelector('#heading')?.getAttribute('class');

      const renderError = (name: string, value: unknown) => {
        const element = document.createElement('boardgame-player-grid') as HTMLElement &
          Record<string, unknown> & { render(): unknown };
        element[name] = value;
        try {
          element.render();
          return '<resolved>';
        } catch (error) {
          return error instanceof Error ? error.message : String(error);
        }
      };
      const blankLabel = renderError('label', ' ');
      const invalidHeading = renderError('headingLevel', 0);
      const blankEmpty = renderError('emptyLabel', ' ');
      return {
        empty,
        wideColumns,
        narrowColumns,
        emptyAfterPlayers,
        hiddenHeadingClass,
        blankLabel,
        invalidHeading,
        blankEmpty,
      };
    });

    expect(result.empty).toEqual({
      label: 'Players',
      regionLabel: 'heading',
      headingRole: 'heading',
      headingLevel: '2',
      text: 'No players',
      part: 'empty',
    });
    expect(result.wideColumns).toBe(3);
    expect(result.narrowColumns).toBe(1);
    expect(result.emptyAfterPlayers).toBe(false);
    expect(result.hiddenHeadingClass).toBe('visually-hidden');
    expect(result.blankLabel).toContain('label must be a non-empty player collection name');
    expect(result.invalidHeading).toContain('headingLevel must be a safe integer from 1 through 6');
    expect(result.blankEmpty).toContain('emptyLabel must be non-empty');
    const axeResult = await new AxeBuilder({ page }).include('boardgame-player-grid').analyze();
    expect(axeResult.violations).toEqual([]);
    diagnostics.assertEmpty();
  } finally {
    diagnostics.stop();
  }
});

test('player panel supplies a named responsive area and honest current-player state', async ({ page }) => {
  const diagnostics = await prepareRendererFixturePage(page);
  try {
    const result = await page.evaluate(async () => {
      await import('/src/client.ts');
      const panel = document.createElement('boardgame-player-panel');
      panel.label = 'Ada';
      panel.active = true;
      panel.style.width = '24rem';
      const score = document.createElement('p');
      score.textContent = 'Score 4';
      const status = document.createElement('p');
      status.slot = 'status';
      status.textContent = 'Protected';
      const actions = document.createElement('button');
      actions.slot = 'actions';
      actions.textContent = 'Pass';
      const footer = document.createElement('small');
      footer.slot = 'footer';
      footer.textContent = '2 cards';
      panel.append(score, status, actions, footer);
      document.body.append(panel);
      await panel.updateComplete;
      await new Promise<void>(resolve => requestAnimationFrame(() => resolve()));
      await panel.updateComplete;

      const root = panel.shadowRoot;
      const region = root?.querySelector('#panel');
      const header = root?.querySelector('#header') as HTMLElement | null;
      const heading = root?.querySelector('#heading');
      const wideDirection = header ? getComputedStyle(header).flexDirection : '';
      const active = {
        labelReference: region?.getAttribute('aria-labelledby'),
        current: region?.getAttribute('aria-current'),
        part: region?.getAttribute('part'),
        heading: heading?.textContent?.trim(),
        headingRole: heading?.getAttribute('role'),
        headingLevel: heading?.getAttribute('aria-level'),
        activeLabel: root?.querySelector('#active')?.textContent?.trim(),
        statusHidden: (root?.querySelector('#status') as HTMLElement | null)?.hidden,
        actionsHidden: (root?.querySelector('#actions') as HTMLElement | null)?.hidden,
        footerHidden: (root?.querySelector('#footer') as HTMLElement | null)?.hidden,
      };

      panel.active = false;
      panel.style.width = '15rem';
      await panel.updateComplete;
      await new Promise<void>(resolve => requestAnimationFrame(() => resolve()));
      const inactive = {
        current: region?.getAttribute('aria-current'),
        part: region?.getAttribute('part'),
        activeLabel: root?.querySelector('#active')?.textContent?.trim() ?? null,
        narrowDirection: header ? getComputedStyle(header).flexDirection : '',
      };
      actions.remove();
      footer.remove();
      await new Promise<void>(resolve => requestAnimationFrame(() => resolve()));
      await panel.updateComplete;
      const emptied = {
        actionsHidden: (root?.querySelector('#actions') as HTMLElement | null)?.hidden,
        footerHidden: (root?.querySelector('#footer') as HTMLElement | null)?.hidden,
      };
      panel.hideHeading = true;
      await panel.updateComplete;
      const hiddenHeadingClass = heading?.getAttribute('class');

      const renderError = (name: string, value: unknown) => {
        const element = document.createElement('boardgame-player-panel') as HTMLElement &
          Record<string, unknown> & { render(): unknown };
        element.label = 'Valid';
        element[name] = value;
        try {
          element.render();
          return '<resolved>';
        } catch (error) {
          return error instanceof Error ? error.message : String(error);
        }
      };
      return {
        active,
        inactive,
        wideDirection,
        emptied,
        hiddenHeadingClass,
        blankLabel: renderError('label', ' '),
        invalidHeading: renderError('headingLevel', 8),
        blankActiveLabel: renderError('activeLabel', ''),
      };
    });

    expect(result.active).toEqual({
      labelReference: 'heading',
      current: 'true',
      part: 'panel active',
      heading: 'Ada',
      headingRole: 'heading',
      headingLevel: '3',
      activeLabel: 'Current player',
      statusHidden: false,
      actionsHidden: false,
      footerHidden: false,
    });
    expect(result.wideDirection).toBe('row');
    expect(result.inactive).toEqual({
      current: 'false',
      part: 'panel',
      activeLabel: null,
      narrowDirection: 'column',
    });
    expect(result.emptied).toEqual({ actionsHidden: true, footerHidden: true });
    expect(result.hiddenHeadingClass).toBe('visually-hidden');
    expect(result.blankLabel).toContain('label must be a non-empty player name');
    expect(result.invalidHeading).toContain('headingLevel must be a safe integer from 1 through 6');
    expect(result.blankActiveLabel).toContain('activeLabel must be a non-empty string');
    const axeResult = await new AxeBuilder({ page }).include('boardgame-player-panel').analyze();
    expect(axeResult.violations).toEqual([]);
    diagnostics.assertEmpty();
  } finally {
    diagnostics.stop();
  }
});

test('player info derives typed state and publishes chip presentation without creator events', async ({ page }) => {
  const diagnostics = await prepareRendererFixturePage(page);
  try {
    const result = await page.evaluate(async () => {
      await import('/src/client.ts');
      await import('/src/components/boardgame-player-roster-item.ts');
      await import('/game-src/tictactoe/boardgame-render-player-info-tictactoe.ts');

      const roster = document.createElement('boardgame-player-roster-item');
      roster.gameName = 'tictactoe';
      roster.active = true;
      roster.rendererLoaded = true;
      roster.playerIndex = 0;
      roster.state = { Players: [{ TokenValue: 'X' }] };
      document.body.append(roster);
      await roster.updateComplete;
      const wrapper = roster.shadowRoot?.querySelector('boardgame-render-player-info') as
        (HTMLElement & { updateComplete: Promise<boolean>; renderer?: HTMLElement & { updateComplete?: Promise<boolean> } }) | null;
      if (!wrapper) throw new Error('player-info wrapper was not rendered');
      await wrapper.updateComplete;
      if (!wrapper.renderer) throw new Error('game-specific player-info renderer was not instantiated');
      if (wrapper.renderer.updateComplete) await wrapper.renderer.updateComplete;
      await roster.updateComplete;
      const firstText = roster.shadowRoot?.querySelector('.chip')?.textContent?.trim();

      roster.state = { Players: [{ TokenValue: 'O' }] };
      await roster.updateComplete;
      await wrapper.updateComplete;
      if (wrapper.renderer.updateComplete) await wrapper.renderer.updateComplete;
      await roster.updateComplete;
      const secondText = roster.shadowRoot?.querySelector('.chip')?.textContent?.trim();
      const derivedState = (wrapper.renderer as { playerState?: { TokenValue?: string } }).playerState?.TokenValue;

      roster.active = false;
      await roster.updateComplete;
      await wrapper.updateComplete;
      await roster.updateComplete;
      const clearedWhenInactive = roster.shadowRoot?.querySelector('.chip')?.textContent?.trim();
      const removedWhenInactive = wrapper.renderer == null;
      roster.active = true;
      await roster.updateComplete;
      await wrapper.updateComplete;
      if (!wrapper.renderer) throw new Error('player-info renderer was not restored after reactivation');
      if (wrapper.renderer.updateComplete) await wrapper.renderer.updateComplete;
      await roster.updateComplete;
      const restoredWhenActive = roster.shadowRoot?.querySelector('.chip')?.textContent?.trim();

      const { BoardgameBasePlayerInfoRenderer } = await import('/src/client.ts');
      class Probe extends BoardgameBasePlayerInfoRenderer<
        { readonly Players: readonly unknown[] },
        unknown
      > {}
      if (!customElements.get('boardgame-player-info-probe-test')) {
        customElements.define('boardgame-player-info-probe-test', Probe);
      }
      const configurationError = (chip: unknown, playerIndex = 0, state: { Players: readonly unknown[] } | null = null) => {
        const probe = document.createElement('boardgame-player-info-probe-test') as Probe;
        Object.defineProperty(probe, 'chip', { configurable: true, value: chip });
        probe.playerIndex = playerIndex;
        probe.state = state;
        try {
          (probe as unknown as { updated(changed: Map<PropertyKey, unknown>): void }).updated(new Map());
          return '<resolved>';
        } catch (error) {
          return error instanceof Error ? error.message : String(error);
        }
      };

      return {
        firstText,
        secondText,
        derivedState,
        clearedWhenInactive,
        removedWhenInactive,
        restoredWhenActive,
        invalidIndex: configurationError({}, -1),
        outOfRange: configurationError({}, 2, { Players: [{}] }),
        invalidShape: configurationError('X'),
        invalidText: configurationError({ text: 1 }),
        invalidColor: configurationError({ color: 'red; background: url(https://invalid.example)' }),
        unknownField: configurationError({ label: 'X' }),
      };
    });

    expect(result.firstText).toBe('X');
    expect(result.secondText).toBe('O');
    expect(result.derivedState).toBe('O');
    expect(result.clearedWhenInactive).toBe('0');
    expect(result.removedWhenInactive).toBe(true);
    expect(result.restoredWhenActive).toBe('O');
    expect(result.invalidIndex).toContain('playerIndex must be a non-negative safe integer');
    expect(result.outOfRange).toContain('outside the 1-player state');
    expect(result.invalidShape).toContain('chip must return an object');
    expect(result.invalidText).toContain('chip.text must be a string');
    expect(result.invalidColor).toContain('is not a valid CSS color');
    expect(result.unknownField).toContain('chip contains unknown field label');
    diagnostics.assertEmpty();
  } finally {
    diagnostics.stop();
  }
});

test('timer display consumes a scoped clock without rerendering game state', async ({ page }) => {
  const diagnostics = await prepareRendererFixturePage(page);
  try {
    const result = await page.evaluate(async () => {
      const {
        TIMER_SERVICE_REQUEST_EVENT,
        TimerService,
      } = await import('/src/timers/timer-service.ts');
      await import('/src/client.ts');

      const service = new TimerService();
      service.update({ hide: { TimeLeft: 2500, originalTimeLeft: 5000 } });
      const provider = document.createElement('div');
      provider.addEventListener(TIMER_SERVICE_REQUEST_EVENT, event => {
        const request = event as CustomEvent<{ accept(service: InstanceType<typeof TimerService>): void }>;
        event.stopPropagation();
        request.detail.accept(service);
      });
      const timer = document.createElement('boardgame-timer');
      timer.timer = { ID: 'hide', IsTimer: true };
      timer.label = 'Cards hide in';
      provider.append(timer);
      document.body.append(provider);
      await timer.updateComplete;
      const progress = timer.shadowRoot?.querySelector('progress') as HTMLProgressElement | null;
      const initial = {
        value: timer.shadowRoot?.querySelector('#value')?.textContent?.trim(),
        progress: progress?.value,
        status: timer.shadowRoot?.querySelector('#timer')?.getAttribute('data-status'),
        label: progress?.getAttribute('aria-labelledby'),
      };

      timer.hideProgress = true;
      await timer.updateComplete;
      const originalRequestUpdate = timer.requestUpdate.bind(timer);
      let selectiveUpdates = 0;
      timer.requestUpdate = (...args: Parameters<typeof timer.requestUpdate>) => {
        selectiveUpdates++;
        return originalRequestUpdate(...args);
      };
      service.update({ hide: { TimeLeft: 2400, originalTimeLeft: 5000 } });
      service.update({ hide: { TimeLeft: 1900, originalTimeLeft: 5000 } });
      await timer.updateComplete;
      timer.requestUpdate = originalRequestUpdate;
      timer.hideProgress = false;

      timer.format = 'clock';
      service.update({ hide: { TimeLeft: 61_000, originalTimeLeft: 120_000 } });
      await timer.updateComplete;
      const clock = timer.shadowRoot?.querySelector('#value')?.textContent?.trim();
      service.update({ hide: { TimeLeft: 0, originalTimeLeft: 120_000 } });
      await timer.updateComplete;
      const elapsed = {
        value: timer.shadowRoot?.querySelector('#value')?.textContent?.trim(),
        announcement: timer.shadowRoot?.querySelector('[role="status"]')?.textContent?.trim(),
        status: timer.shadowRoot?.querySelector('#timer')?.getAttribute('data-status'),
      };

      timer.timer = { ID: '', IsTimer: true };
      await timer.updateComplete;
      await timer.updateComplete;
      const hiddenWhenIdle = timer.shadowRoot?.querySelector('#timer') === null;

      const renderError = (name: string, value: unknown) => {
        const element = document.createElement('boardgame-timer') as HTMLElement &
          Record<string, unknown> & { render(): unknown };
        element[name] = value;
        try {
          element.render();
          return '<resolved>';
        } catch (error) {
          return error instanceof Error ? error.message : String(error);
        }
      };
      timer.timer = { ID: 'hide', IsTimer: true };
      service.update({ hide: { TimeLeft: 1000, originalTimeLeft: 2000 } });
      await timer.updateComplete;
      await timer.updateComplete;
      return {
        initial,
        selectiveUpdates,
        clock,
        elapsed,
        hiddenWhenIdle,
        blankLabel: renderError('label', ' '),
        invalidFormat: renderError('format', 'minutes'),
        blankExpired: renderError('expiredLabel', ' '),
      };
    });

    expect(result.initial).toEqual({ value: '3s', progress: 0.5, status: 'running', label: 'label' });
    expect(result.selectiveUpdates).toBe(1);
    expect(result.clock).toBe('1:01');
    expect(result.elapsed).toEqual({ value: 'Time expired', announcement: 'Time expired', status: 'elapsed' });
    expect(result.hiddenWhenIdle).toBe(true);
    expect(result.blankLabel).toContain('label must be non-empty');
    expect(result.invalidFormat).toContain('unknown format');
    expect(result.blankExpired).toContain('expiredLabel must be non-empty');
    const axeResult = await new AxeBuilder({ page }).include('boardgame-timer').analyze();
    expect(axeResult.violations).toEqual([]);
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
    try {
      await mountRendererFixture({
        ...base,
        snapshot: {
          ...base.snapshot,
          playerPresentations: [{ playerIndex: 1, label: 'Wrong slot' }],
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
      playerPresentations: readonly { playerIndex: number; label: string; color?: string }[] = [];
      playerPresentation(playerIndex: number) {
        return this.playerPresentations[playerIndex] ?? { playerIndex, label: `Player ${playerIndex + 1}` };
      }
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
      playerPresentations: readonly { playerIndex: number; label: string; color?: string }[] = [];
      playerPresentation(playerIndex: number) {
        return this.playerPresentations[playerIndex] ?? { playerIndex, label: `Player ${playerIndex + 1}` };
      }
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
    'playerPresentations: presentation 0 must have playerIndex 0, not 1',
    'deliberate render failure',
    'Uncaught Error: Renderer fixture received unknown move proposal: toString',
  ]);
  expect(result).toMatchObject({
    proposalListenersAdded: 1,
    proposalListenersRemoved: 1,
    leakedHosts: 0,
  });
});

test('dynamic host ignores a module load superseded by a newer game route', async ({ page }) => {
  await page.goto('/client_config.js');
  await page.evaluate(async () => {
    await import('/src/components/boardgame-render-game.ts');
    const BaseRenderer = customElements.get('boardgame-base-game-renderer');
    if (!BaseRenderer) throw new Error('Framework renderer base was not registered');
    customElements.define('boardgame-render-game-secondgame', class extends BaseRenderer {});
    customElements.define('boardgame-render-game-secondgame-table', class extends BaseRenderer {});
    type TestHost = HTMLElement & {
      active: boolean;
      gameId: string;
      gameName: string;
      renderer: HTMLElement | null;
      rendererError: string;
      updateComplete: Promise<unknown>;
      _loadRendererModule(modulePath: string): Promise<unknown>;
    };
    const host = document.createElement('boardgame-render-game') as TestHost;
    let releasePig: (() => void) | undefined;
    const requestedPaths: string[] = [];
    host._loadRendererModule = (modulePath: string) => {
      requestedPaths.push(modulePath);
      if (modulePath.includes('/pig/')) {
        return new Promise<void>(resolve => releasePig = resolve);
      }
      return Promise.resolve();
    };
    host.active = true;
    host.gameId = 'FIRST';
    host.gameName = 'pig';
    document.body.append(host);
    await host.updateComplete;
    if (!releasePig) throw new Error('Pig renderer load did not start');

    host.gameId = 'SECOND';
    host.gameName = 'secondgame';
    await host.updateComplete;
    (globalThis as unknown as {
      __dynamicRendererHost: TestHost;
      __releasePigRenderer: () => void;
      __rendererModulePaths: string[];
    }).__dynamicRendererHost = host;
    (globalThis as unknown as { __releasePigRenderer: () => void }).__releasePigRenderer = releasePig;
    (globalThis as unknown as { __rendererModulePaths: string[] }).__rendererModulePaths = requestedPaths;
  });
  await expect.poll(() => page.evaluate(() => {
      const host = (globalThis as unknown as {
        __dynamicRendererHost: { renderer: HTMLElement | null };
      }).__dynamicRendererHost;
      return host.renderer?.tagName.toLowerCase() ?? '';
  })).toBe('boardgame-render-game-secondgame');

  await page.evaluate(() => (
    globalThis as unknown as { __releasePigRenderer: () => void }
  ).__releasePigRenderer());
  await page.waitForTimeout(50);
  expect(await page.evaluate(() => {
      const host = (globalThis as unknown as {
        __dynamicRendererHost: { renderer: HTMLElement | null; rendererError: string };
      }).__dynamicRendererHost;
      return {
        tagName: host.renderer?.tagName.toLowerCase() ?? '',
        error: host.rendererError,
      };
  })).toEqual({ tagName: 'boardgame-render-game-secondgame', error: '' });

  await page.evaluate(async () => {
    const host = (globalThis as unknown as {
      __dynamicRendererHost: HTMLElement & { gameId: string; updateComplete: Promise<unknown> };
    }).__dynamicRendererHost;
    document.cookie = 'surface_THIRD=table; Path=/';
    host.gameId = 'THIRD';
    await host.updateComplete;
  });
  await expect.poll(() => page.evaluate(() => {
    const host = (globalThis as unknown as {
      __dynamicRendererHost: { renderer: HTMLElement | null };
    }).__dynamicRendererHost;
    return host.renderer?.tagName.toLowerCase() ?? '';
  })).toBe('boardgame-render-game-secondgame-table');
  expect(await page.evaluate(() => (
    globalThis as unknown as { __rendererModulePaths: string[] }
  ).__rendererModulePaths)).toEqual([
    '../../game-src/pig/boardgame-render-game-pig.ts',
    '../../game-src/secondgame/boardgame-render-game-secondgame.ts',
    '../../game-src/secondgame/boardgame-render-game-secondgame-table.ts',
  ]);
});

test('player renderer load failures are visible and reject the wrong base', async ({ page }) => {
  await page.route('**/game-src/contractplayer/boardgame-render-player-info-contractplayer.ts', route => (
    route.fulfill({
      contentType: 'text/javascript',
      body: `customElements.define('boardgame-render-player-info-contractplayer', class extends HTMLElement {});`,
    })
  ));
  await page.goto('/client_config.js');
  await page.evaluate(async () => {
    (globalThis as unknown as {
      CONFIG: { offline_dev_mode: boolean; firebase: Record<string, string> };
    }).CONFIG = {
      offline_dev_mode: true,
      firebase: { apiKey: 'fixture', projectId: 'fixture', appId: '1:fixture:web:fixture' },
    };
    await import('/src/components/boardgame-player-roster.ts');
    const roster = document.createElement('boardgame-player-roster') as HTMLElement & {
      active: boolean;
      gameRoute: { name: string; id: string };
      updateComplete: Promise<unknown>;
    };
    roster.active = true;
    roster.gameRoute = { name: 'contractplayer', id: 'GAME' };
    document.body.append(roster);
    await roster.updateComplete;
  });
  const alert = page.locator('boardgame-player-roster').getByRole('alert');
  await expect(alert).toContainText(
    'Player renderer <boardgame-render-player-info-contractplayer> must extend the generated PlayerInfoRenderer base',
  );
  await expect(alert).toContainText('boardgame-util check-client');
});

test('generated registration helpers fail early with game-specific diagnostics', async ({ page }) => {
  await page.goto('/client_config.js');
  const result = await page.evaluate(async () => {
    const contract = await import('/game-src/pig/_game_renderer.ts');
    class FirstGameRenderer extends contract.GameRenderer {}
    class SecondGameRenderer extends contract.GameRenderer {}
    class GameRendererUsedAsTable extends contract.GameRenderer {}

    contract.registerGameRenderer(FirstGameRenderer);
    let duplicateMessage = '';
    try {
      contract.registerGameRenderer(SecondGameRenderer);
    } catch (error) {
      duplicateMessage = error instanceof Error ? error.message : String(error);
    }

    let wrongBaseMessage = '';
    try {
      contract.registerTableRenderer(GameRendererUsedAsTable as never);
    } catch (error) {
      wrongBaseMessage = error instanceof Error ? error.message : String(error);
    }

    return {
      duplicateMessage,
      wrongBaseMessage,
      registeredConstructor: customElements.get('boardgame-render-game-pig')?.name,
    };
  });

  expect(result).toEqual({
    duplicateMessage:
      '[pig] cannot register game renderer SecondGameRenderer as <boardgame-render-game-pig>: ' +
      'that tag is already registered by FirstGameRenderer',
    wrongBaseMessage:
      '[pig] table renderer GameRendererUsedAsTable must extend the generated TableRenderer base',
    registeredConstructor: 'FirstGameRenderer',
  });
});

test('dynamic host rejects missing registrations and renderers that bypass the generated base', async ({ page }) => {
  await page.goto('/client_config.js');
  const result = await page.evaluate(async () => {
    await import('/src/components/boardgame-render-game.ts');
    const tagName = 'boardgame-render-game-wrongbase-contract';
    customElements.define(tagName, class extends HTMLElement {});
    const host = document.createElement('boardgame-render-game') as HTMLElement & {
      rendererLoaded: boolean;
      rendererError: string;
      updateComplete: Promise<boolean>;
      _instantiateRenderer(surfaceSuffix?: string): void;
    };
    document.body.append(host);
    await host.updateComplete;
    let missingMessage = '';
    try {
      host._instantiateRenderer('missing-contract');
    } catch (error) {
      missingMessage = error instanceof Error ? error.message : String(error);
    }
    let message = '';
    try {
      host._instantiateRenderer('wrongbase-contract');
    } catch (error) {
      message = error instanceof Error ? error.message : String(error);
    }
    await host.updateComplete;
    const alert = host.shadowRoot?.querySelector<HTMLElement>('[role="alert"]');
    const result = {
      missingMessage,
      message,
      rendererLoaded: host.rendererLoaded,
      rendererError: host.rendererError,
      alert: alert?.innerText.replace(/\s+/g, ' ').trim() ?? '',
    };
    host.remove();
    return result;
  });
  const message = 'Renderer <boardgame-render-game-wrongbase-contract> must extend the generated renderer base; use GameRenderer, TableRenderer, or HandRenderer with its generated registration decorator';
  expect(result).toEqual({
    missingMessage: 'Renderer module loaded but did not register <boardgame-render-game-missing-contract>; use the generated registration decorator for this exact surface',
    message,
    rendererLoaded: false,
    rendererError: message,
    alert: `Game renderer unavailable ${message} Run boardgame-util check-client and fix every reported diagnostic.`,
  });
});

test('Checkers composes source selection with typed destination actions', async ({ page }) => {
  const diagnostics = await prepareRendererFixturePage(page);
  try {
    await retryRendererEvaluation(page, () => page.evaluate(async () => {
      await import('/game-src/checkers/boardgame-render-game-checkers.ts');
      const { checkersRendererFixture } = await import(
        '/game-src/checkers/boardgame-render-fixtures-checkers.ts'
      );
      const { mountRendererFixture } = await import('/src/testing/renderer-fixture.ts');
      const handle = await mountRendererFixture(checkersRendererFixture);
      (globalThis as unknown as { __checkersFixtureHandle: typeof handle }).__checkersFixtureHandle = handle;
    }));

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

test('game board composes typed placement destinations without parallel click state', async ({ page }) => {
  await page.goto('/client_config.js');
  const result = await page.evaluate(async () => {
    await import('/src/components/boardgame-game-board.ts');
    const board = document.createElement('boardgame-game-board');
    board.rows = 2; board.cols = 2; board.boardLabel = 'Word board';
    let selected: string | null = null;
    let placedAt: number | null = null;
    const binding = () => ({
      targets: [0, 1, 2, 3] as const,
      selectedItem: selected,
      target: (target: number) => ({
        target,
        occupiedBy: placedAt === target ? 'tile-a' : null,
        canPlace: selected !== null && placedAt !== target,
        reason: selected === null ? 'Select an item first'
          : placedAt === target ? 'Destination is occupied' : null,
        place: () => {
          placedAt = target;
          selected = null;
          board.placementDraft = binding();
        },
      }),
    });
    board.placementDraft = binding();
    document.body.append(board);
    await board.updateComplete;
    const cells = board.shadowRoot!.querySelectorAll<HTMLButtonElement>('.cell');
    const initial = {
      disabled: cells[0]?.getAttribute('aria-disabled'),
      reason: cells[0]?.getAttribute('title'),
    };
    selected = 'tile-a';
    board.placementDraft = binding();
    await board.updateComplete;
    const available = cells[2]?.getAttribute('aria-disabled');
    cells[2]?.click();
    await board.updateComplete;
    const after = { placedAt, reason: cells[2]?.getAttribute('title') };

    const failure = async (configure: (candidate: HTMLElementTagNameMap['boardgame-game-board']) => void) => {
      const candidate = document.createElement('boardgame-game-board');
      candidate.rows = 1; candidate.cols = 2;
      configure(candidate);
      document.body.append(candidate);
      try { await candidate.updateComplete; return '<missing error>'; }
      catch (error) { return error instanceof Error ? error.message : String(error); }
      finally { candidate.remove(); }
    };
    const malformed = { ...binding(), targets: [0] };
    const errors = await Promise.all([
      failure(candidate => { candidate.placementDraft = malformed; }),
      failure(candidate => {
        candidate.placementDraft = { ...binding(), targets: [0, 1] };
        candidate.action = {
          candidates: [], preview: { kind: 'ready' }, get: () => undefined,
          ensurePreview: async () => ({ kind: 'ready' }), refreshPreview: async () => ({ kind: 'ready' }),
          subscribe: () => () => undefined,
        };
      }),
      failure(candidate => {
        candidate.placementDraft = { ...binding(), targets: [0, 1] };
        candidate.disabledSpaces = [0];
      }),
    ]);
    return { initial, available, after, errors };
  });
  expect(result).toEqual({
    initial: { disabled: 'true', reason: 'Select an item first' },
    available: 'false',
    after: { placedAt: 2, reason: 'Select an item first' },
    errors: [
      expect.stringContaining('targets must cover exactly 0 through 1'),
      expect.stringContaining('mutually exclusive'),
      expect.stringContaining('placementDraft and disabledSpaces are mutually exclusive'),
    ],
  });
  const axeResult = await new AxeBuilder({ page }).include('boardgame-game-board').analyze();
  expect(axeResult.violations).toEqual([]);
});

test('source-destination grids disable cells that cannot perform an interaction', async ({ page }) => {
  await page.goto('/client_config.js');
  const result = await page.evaluate(async () => {
    await import('/src/components/boardgame-game-board.ts');
    const board = document.createElement('boardgame-game-board') as HTMLElement & {
      rows: number;
      cols: number;
      sourceDestination: unknown;
      updateComplete: Promise<unknown>;
    };
    const base = {
      sources: [1],
      selectedSource: null,
      action: null,
      selectSource: () => undefined,
      clear: () => undefined,
    };
    board.rows = 1;
    board.cols = 3;
    board.sourceDestination = base;
    document.body.append(board);
    await board.updateComplete;
    const disabled = () => [...board.shadowRoot!.querySelectorAll<HTMLButtonElement>('.cell')]
      .map(cell => cell.getAttribute('aria-disabled') === 'true');
    const before = disabled();
    const candidate = {
      key: 2,
      action: { canActivate: true, canPropose: true, reason: null },
    };
    board.sourceDestination = {
      ...base,
      selectedSource: 1,
      action: {
        candidates: [candidate],
        preview: { kind: 'ready' },
        get: (key: number) => key === 2 ? candidate : undefined,
        subscribe: () => () => undefined,
      },
    };
    await board.updateComplete;
    const after = disabled();
    board.remove();
    return { before, after };
  });
  expect(result).toEqual({
    before: [true, false, true],
    after: [true, false, false],
  });
});

test('spatial board sanitizes authored geometry and shares typed target activation', async ({ page }) => {
  await page.goto('/client_config.js');
  const result = await page.evaluate(async () => {
    const source = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="10 20 200 100" preserveAspectRatio="xMaxYMin meet"
      onload="globalThis.__unsafeBoardRoot = true" style="fill:url(https://attacker.invalid/root)">
      <script>globalThis.__unsafeBoardScript = true</script>
      <g transform="translate(20 10)">
        <g data-board-space="room:one/?" data-board-label="Library" data-board-order="3"
          data-board-group="rooms"
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
      const focused = region.classList.contains('focused');

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

      board.action = { ...action, candidates: [] };
      await board.updateComplete;
      await board.updateComplete;
      const invalidActionStatus = root?.querySelector('#status')?.textContent?.replace(/\s+/g, ' ').trim();
      board.action = null;
      await board.updateComplete;
      await board.updateComplete;
      const clearedActionError = root?.querySelector('#status')?.hasAttribute('hidden');
      board.action = action;
      await board.updateComplete;

      const successful = {
        activations,
        label: button.textContent?.trim(),
        focused,
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
        invalidActionStatus,
        clearedActionError,
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
    invalidActionStatus: expect.stringContaining('target keys do not match geometry'),
    clearedActionError: true,
    componentViewForwarded: true,
    inspector: expect.stringContaining('room:one/? — Library; group=rooms'),
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

test('board viewport provides bounded navigation without turning drags into board actions', async ({ page }) => {
  await page.goto('/client_config.js');
  await page.evaluate(async () => {
    await import('/src/components/boardgame-board-viewport.ts');
    const viewport = document.createElement('boardgame-board-viewport') as HTMLElement & {
      label: string;
      maxScale: number;
      zoomStep: number;
      view: { readonly scale: number; readonly x: number; readonly y: number };
      zoomIn(): void;
      resetView(): void;
      setView(view: { readonly scale: number; readonly x: number; readonly y: number }): void;
      updateComplete: Promise<unknown>;
    };
    viewport.id = 'viewport-fixture';
    viewport.label = 'Fixture map navigation';
    viewport.maxScale = 3;
    viewport.zoomStep = 0.5;
    viewport.style.width = '400px';
    const target = document.createElement('button');
    target.type = 'button';
    target.textContent = 'Map target';
    target.style.cssText = 'display:block;width:100%;height:240px;border:0';
    target.addEventListener('click', () => {
      const state = globalThis as unknown as { __viewportClicks?: number };
      state.__viewportClicks = (state.__viewportClicks ?? 0) + 1;
    });
    viewport.append(target);
    document.body.append(viewport);
    await viewport.updateComplete;
    viewport.zoomIn();
    viewport.zoomIn();
    await viewport.updateComplete;
  });

  const viewport = page.locator('#viewport-fixture').locator('#viewport');
  const bounds = await viewport.boundingBox();
  if (!bounds) throw new Error('Viewport fixture had no bounds');
  await page.mouse.move(bounds.x + bounds.width / 2, bounds.y + bounds.height / 2);
  await page.mouse.down();
  await page.mouse.move(bounds.x + bounds.width / 2 + 40, bounds.y + bounds.height / 2 + 20, { steps: 4 });
  await page.mouse.up();
  await page.waitForTimeout(0);
  expect(await page.evaluate(() => (globalThis as unknown as { __viewportClicks?: number }).__viewportClicks ?? 0)).toBe(0);
  await page.evaluate(async () => {
    const viewport = document.querySelector('#viewport-fixture') as HTMLElement & {
      view: { readonly scale: number; readonly x: number; readonly y: number };
      resetView(): void;
      setView(view: { readonly scale: number; readonly x: number; readonly y: number }): void;
      updateComplete: Promise<unknown>;
    };
    (globalThis as unknown as { __afterDragView?: typeof viewport.view }).__afterDragView = viewport.view;
    viewport.resetView();
    await viewport.updateComplete;
  });
  await page.getByRole('button', { name: 'Map target' }).click();
  expect(await page.evaluate(() => (globalThis as unknown as { __viewportClicks?: number }).__viewportClicks ?? 0)).toBe(1);

  const cdp = await page.context().newCDPSession(page);
  const center = { x: bounds.x + bounds.width / 2, y: bounds.y + bounds.height / 2 };
  await cdp.send('Input.dispatchTouchEvent', {
    type: 'touchStart',
    touchPoints: [
      { id: 1, x: center.x - 30, y: center.y },
      { id: 2, x: center.x + 30, y: center.y },
    ],
  });
  await cdp.send('Input.dispatchTouchEvent', {
    type: 'touchMove',
    touchPoints: [
      { id: 1, x: center.x - 60, y: center.y },
      { id: 2, x: center.x + 60, y: center.y },
    ],
  });
  await cdp.send('Input.dispatchTouchEvent', { type: 'touchEnd', touchPoints: [] });
  const pinchScale = await page.evaluate(() => (
    document.querySelector('#viewport-fixture') as HTMLElement & { view: { readonly scale: number } }
  ).view.scale);
  expect(pinchScale).toBeGreaterThan(1);
  await page.evaluate(async () => {
    const viewport = document.querySelector('#viewport-fixture') as HTMLElement & {
      resetView(): void;
      updateComplete: Promise<unknown>;
    };
    viewport.resetView();
    await viewport.updateComplete;
  });

  const result = await page.evaluate(async () => {
    const viewport = document.querySelector('#viewport-fixture') as HTMLElement & {
      label: string;
      maxScale: number;
      zoomStep: number;
      view: { readonly scale: number; readonly x: number; readonly y: number };
      resetView(): void;
      updateComplete: Promise<unknown>;
    };
    const inner = viewport.shadowRoot?.querySelector('#viewport');
    if (!(inner instanceof HTMLDivElement)) throw new Error('Viewport internals unavailable');
    const afterDrag = (globalThis as unknown as { __afterDragView?: typeof viewport.view }).__afterDragView;
    if (!afterDrag) throw new Error('Drag view was not captured');
    viewport.zoomIn();
    viewport.zoomIn();
    await viewport.updateComplete;
    inner.dispatchEvent(new KeyboardEvent('keydown', { key: 'ArrowRight', bubbles: true }));
    await viewport.updateComplete;
    const afterKeyboard = viewport.view;
    inner.dispatchEvent(new WheelEvent('wheel', { deltaY: -10, bubbles: true }));
    const afterPlainWheel = viewport.view;
    inner.dispatchEvent(new WheelEvent('wheel', {
      deltaY: -10, ctrlKey: true, clientX: 200, clientY: 120, bubbles: true, cancelable: true,
    }));
    await viewport.updateComplete;
    const afterCtrlWheel = viewport.view;
    viewport.maxScale = 2.2;
    await viewport.updateComplete;
    const clampedAfterMaxChange = viewport.view;
    viewport.setView({ scale: 2.2, x: -10_000, y: -10_000 });
    viewport.remove();
    viewport.style.width = '300px';
    document.body.append(viewport);
    await viewport.updateComplete;
    await new Promise<void>(resolve => requestAnimationFrame(() => requestAnimationFrame(() => resolve())));
    const reconnected = viewport.view;
    viewport.resetView();
    await viewport.updateComplete;
    const reset = viewport.view;
    const invalid = async (properties: Record<string, unknown>) => {
      const element = document.createElement('boardgame-board-viewport');
      Object.assign(element, properties);
      document.body.append(element);
      try {
        await (element as typeof element & { updateComplete: Promise<unknown> }).updateComplete;
        return '<resolved>';
      } catch (error) {
        return error instanceof Error ? error.message : String(error);
      } finally {
        element.remove();
      }
    };
    return {
      afterDrag,
      afterKeyboard,
      afterPlainWheel,
      afterCtrlWheel,
      clampedAfterMaxChange,
      reconnected,
      reset,
      status: viewport.shadowRoot?.querySelector('#status')?.textContent,
      parts: [...(viewport.shadowRoot?.querySelectorAll('[part]') ?? [])].map(element => element.getAttribute('part')),
      errors: await Promise.all([
        invalid({ label: ' ' }), invalid({ maxScale: 17 }), invalid({ zoomStep: 0 }),
      ]),
    };
  });
  expect(result.afterDrag.scale).toBe(2);
  expect(result.afterDrag.x).toBeGreaterThan(-200);
  expect(result.afterKeyboard.x).toBeLessThan(result.afterDrag.x);
  expect(result.afterPlainWheel).toEqual(result.afterKeyboard);
  expect(result.afterCtrlWheel.scale).toBe(2.5);
  expect(result.clampedAfterMaxChange.scale).toBe(2.2);
  expect(result.reconnected.x).toBeGreaterThanOrEqual(-360);
  expect(result.reset).toEqual({ scale: 1, x: 0, y: 0 });
  expect(result.status).toBe('100% zoom');
  expect(result.parts).toEqual(expect.arrayContaining(['toolbar', 'viewport', 'scene', 'zoom-in', 'zoom-out', 'reset']));
  expect(result.errors).toEqual([
    expect.stringContaining('label must be a non-empty string'),
    expect.stringContaining('maxScale must be from 1 through 16'),
    expect.stringContaining('zoomStep must be greater than 0'),
  ]);
  const axeResult = await new AxeBuilder({ page })
    .include('boardgame-board-viewport')
    .withRules(['button-name', 'aria-allowed-attr', 'aria-valid-attr-value', 'nested-interactive'])
    .analyze();
  expect(axeResult.violations).toEqual([]);
});

test('spatial board keeps raster artwork, hotspots, focus, and pieces in one responsive coordinate system', async ({ page }) => {
  await page.goto('/client_config.js');
  const result = await page.evaluate(async () => {
    await import('/src/components/boardgame-spatial-board.ts');
    const { rasterBoardArtwork } = await import('/src/components/spatial-board-geometry.ts');
    const canvas = document.createElement('canvas');
    canvas.width = 200;
    canvas.height = 100;
    const context = canvas.getContext('2d');
    if (!context) throw new Error('Canvas unavailable');
    context.fillStyle = '#315c3b';
    context.fillRect(0, 0, 100, 100);
    context.fillStyle = '#d7a84a';
    context.fillRect(100, 0, 100, 100);
    const src = canvas.toDataURL('image/png');
    const spaces = [
      {
        key: 'harbor', label: 'Harbor', order: 0,
        group: 'nodes',
        region: { shape: 'circle' as const, center: { x: 0.25, y: 0.5 }, radius: 0.24 },
        focusAnchor: { x: 0.2, y: 0.45 }, pieceAnchor: { x: 0.3, y: 0.55 },
      },
      {
        key: 'market', label: 'Market', order: 1,
        group: 'tiles',
        region: { shape: 'rect' as const, x: 0.55, y: 0.2, width: 0.35, height: 0.6 },
      },
      {
        key: 'road', label: 'Road', order: 2,
        group: 'nodes',
        region: { shape: 'polygon' as const, points: [
          { x: 0.42, y: 0.1 }, { x: 0.58, y: 0.5 }, { x: 0.42, y: 0.9 },
        ] },
      },
    ];
    const board = document.createElement('boardgame-spatial-board') as HTMLElement & {
      artwork: unknown;
      svgUrl: string;
      geometry: unknown;
      action: unknown;
      actionGroup: string;
      placementDraft: unknown;
      disabledSpaces: number[];
      pathOverlays: readonly unknown[];
      pieces: readonly unknown[];
      tokenSize: number;
      panZoom: boolean;
      maxZoom: number;
      svgLoaded: boolean;
      updateComplete: Promise<unknown>;
    };
    board.style.width = '320px';
    board.panZoom = true;
    board.maxZoom = 3;
    board.artwork = rasterBoardArtwork({ src, spaces, fit: 'contain', viewportAspectRatio: 1 });
    const token = { Index: 0, Values: {}, Deck: 'tokens', GameName: 'fixture', ID: 'raster-token' };
    const stack = {
      Deck: 'tokens', Indexes: [0], IDs: [token.ID], IDsLastSeen: {}, ShuffleCount: 0,
      GameName: 'fixture', Components: [token],
    };
    board.tokenSize = 20;
    board.pieces = [{ id: token.ID, space: 'harbor', stack, slot: 0, component: token }];
    board.pathOverlays = [{
      id: 'trade-route', label: 'Trade route from Harbor through Road to Market',
      spaces: ['harbor', 'road', 'market'], tone: 'secondary', width: 6,
    }];
    document.body.append(board);
    const waitForLoad = async () => {
      for (let attempt = 0; attempt < 40 && !board.svgLoaded; attempt++) {
        await new Promise(resolve => setTimeout(resolve, 0));
      }
      await board.updateComplete;
      if (!board.svgLoaded) throw new Error('Raster board did not load');
      await new Promise<void>(resolve => requestAnimationFrame(() => requestAnimationFrame(() => resolve())));
    };
    await waitForLoad();
    const root = board.shadowRoot;
    const sceneInfo = () => {
      const outer = root?.querySelector('#container > svg');
      const scene = outer?.querySelector(':scope > svg');
      const image = scene?.querySelector('image');
      const regions = [...(scene?.querySelectorAll('[data-space]') ?? [])];
      const focus = root?.querySelector('.space-focus');
      const path = root?.querySelector('#path-overlay polyline');
      const stackElement = root?.querySelector('boardgame-component-stack') as HTMLElement & {
        spatialPositions: readonly ({ top: number; left: number } | null)[];
        updateComplete: Promise<unknown>;
      } | null;
      if (!(outer instanceof SVGSVGElement) || !(scene instanceof SVGSVGElement)
        || !(image instanceof SVGImageElement) || regions.length !== 3
        || !(focus instanceof HTMLButtonElement) || !(path instanceof SVGGraphicsElement)
        || path.localName !== 'polyline' || !stackElement) {
        throw new Error('Raster scene was incomplete');
      }
      const harbor = regions[0] as SVGGraphicsElement;
      const focusAnchor = scene.querySelectorAll('circle')[1];
      const pieceAnchor = scene.querySelectorAll('circle')[2];
      if (!(focusAnchor instanceof SVGGraphicsElement) || !(pieceAnchor instanceof SVGGraphicsElement)) {
        throw new Error('Raster anchors were not generated');
      }
      const center = (element: Element) => {
        const bounds = element.getBoundingClientRect();
        return { x: bounds.left + bounds.width / 2, y: bounds.top + bounds.height / 2 };
      };
      return {
        outer, scene, image, harbor, focus, focusAnchor, pieceAnchor,
        path: path as SVGPolylineElement, regions, stackElement,
        focusError: Math.max(
          Math.abs(center(focus).x - center(focusAnchor).x),
          Math.abs(center(focus).y - center(focusAnchor).y),
        ),
      };
    };
    const contain = sceneInfo();
    await contain.stackElement.updateComplete;
    const containPosition = contain.stackElement.spatialPositions[0];
    const containerBounds = root?.querySelector('#container')?.getBoundingClientRect();
    if (!containPosition || !containerBounds) throw new Error('Raster piece was not positioned');
    const pieceBounds = contain.pieceAnchor.getBoundingClientRect();
    const jitter = (axis: number) => {
      let hash = axis * 41;
      hash = ((hash >>> 16) ^ hash) * 0x45d9f3b;
      hash = ((hash >>> 16) ^ hash) * 0x45d9f3b;
      hash = (hash >>> 16) ^ hash;
      return ((hash & 0xFFFF) / 0x7FFF) - 1;
    };
    const containPieceError = Math.max(
      Math.abs(containerBounds.left + containPosition.left + 10
        - (pieceBounds.left + pieceBounds.width / 2 + jitter(0) * 20)),
      Math.abs(containerBounds.top + containPosition.top + 10
        - (pieceBounds.top + pieceBounds.height / 2 + jitter(1) * 20)),
    );

    const viewport = root?.querySelector('boardgame-board-viewport') as HTMLElement & {
      view: { readonly scale: number; readonly x: number; readonly y: number };
      zoomIn(): void;
      updateComplete: Promise<unknown>;
    } | null;
    if (!viewport) throw new Error('Raster board did not compose its pan/zoom viewport');
    viewport.zoomIn();
    viewport.zoomIn();
    await viewport.updateComplete;

    const fits: Record<string, {
      preserve: string | null; focusError: number; pieceError: number; pathError: number;
    }> = {};
    for (const fit of ['contain', 'cover', 'fill'] as const) {
      board.artwork = rasterBoardArtwork({ src, spaces, fit, viewportAspectRatio: 1 });
      await board.updateComplete;
      await waitForLoad();
      const info = sceneInfo();
      await info.stackElement.updateComplete;
      const pieceError = () => {
        const position = info.stackElement.spatialPositions[0];
        const container = root?.querySelector('#container')?.getBoundingClientRect();
        const anchor = info.pieceAnchor.getBoundingClientRect();
        if (!position || !container) throw new Error('Zoomed raster piece disappeared');
        return Math.max(
          Math.abs(position.left + 10 - ((anchor.left + anchor.width / 2 - container.left) / viewport.view.scale + jitter(0) * 20)),
          Math.abs(position.top + 10 - ((anchor.top + anchor.height / 2 - container.top) / viewport.view.scale + jitter(1) * 20)),
        );
      };
      const pathError = () => {
        const container = root?.querySelector('#container')?.getBoundingClientRect();
        if (!container || info.path.points.length !== 3) throw new Error('Route path points disappeared');
        const anchors = [info.pieceAnchor, info.regions[2]!, info.regions[1]!];
        return Math.max(...anchors.flatMap((anchor, index) => {
          const bounds = anchor.getBoundingClientRect();
          const point = info.path.points.getItem(index);
          return [
            Math.abs(point.x - (bounds.left + bounds.width / 2 - container.left) / viewport.view.scale),
            Math.abs(point.y - (bounds.top + bounds.height / 2 - container.top) / viewport.view.scale),
          ];
        }));
      };
      fits[fit] = {
        preserve: info.scene.getAttribute('preserveAspectRatio'),
        focusError: info.focusError,
        pieceError: pieceError(),
        pathError: pathError(),
      };
      for (const width of [260, 640]) {
        board.style.width = `${width}px`;
        await new Promise<void>(resolve => requestAnimationFrame(() => requestAnimationFrame(() => resolve())));
        await board.updateComplete;
        fits[fit]!.focusError = Math.max(fits[fit]!.focusError, sceneInfo().focusError);
        fits[fit]!.pieceError = Math.max(fits[fit]!.pieceError, pieceError());
        fits[fit]!.pathError = Math.max(fits[fit]!.pathError, pathError());
      }
    }

    let groupActivations = 0;
    const groupedCandidates = ['harbor', 'road'].map(key => ({
      key,
      action: {
        canActivate: true,
        reason: null,
        activate: async () => {
          groupActivations++;
          return { kind: 'success', requestID: `group-${key}` };
        },
      },
    }));
    board.actionGroup = 'nodes';
    board.action = {
      candidates: groupedCandidates,
      preview: { kind: 'ready' },
      get: (key: string | number) => groupedCandidates.find(candidate => candidate.key === key),
      ensurePreview: async () => ({ kind: 'ready' }),
      refreshPreview: async () => ({ kind: 'ready' }),
      subscribe: () => () => undefined,
    };
    await board.updateComplete;
    const groupedRegions = [...(root?.querySelectorAll('[data-space]') ?? [])];
    const marketRegion = groupedRegions[1];
    const harborRegion = groupedRegions[0];
    if (!(marketRegion instanceof SVGGraphicsElement) || !(harborRegion instanceof SVGGraphicsElement)) {
      throw new Error('Grouped geometry was unavailable');
    }
    const groupedLabels = [...(root?.querySelectorAll('#space-list button') ?? [])]
      .map(button => button.textContent?.trim());
    const inactiveMarket = marketRegion.classList.contains('inactive')
      && !marketRegion.classList.contains('disabled');
    marketRegion.dispatchEvent(new MouseEvent('click', { bubbles: true, composed: true }));
    harborRegion.dispatchEvent(new MouseEvent('click', { bubbles: true, composed: true }));
    await Promise.resolve();
    board.actionGroup = 'missing';
    await board.updateComplete;
    await board.updateComplete;
    const unknownGroupError = root?.querySelector('#status')?.textContent?.replace(/\s+/g, ' ').trim();
    board.actionGroup = 'nodes';
    await board.updateComplete;
    await board.updateComplete;
    const recoveredGroupError = root?.querySelector('#status')?.textContent?.trim() ?? '';
    const pathDescription = root?.querySelector('#path-descriptions')?.textContent?.replace(/\s+/g, ' ').trim();
    const pathTone = root?.querySelector('#path-overlay polyline')?.getAttribute('class');
    const pathWidth = root?.querySelector('#path-overlay polyline')?.getAttribute('stroke-width');
    board.action = null;
    board.actionGroup = '';
    board.pathOverlays = [];
    await board.updateComplete;

    let selectedDraftItem: string | null = null;
    let placedDraftTarget: string | null = null;
    const placementBinding = () => ({
      targets: ['harbor', 'market', 'road'] as const,
      selectedItem: selectedDraftItem,
      target: (target: string | number) => ({
        target,
        occupiedBy: placedDraftTarget === target ? 'tile-a' : null,
        canPlace: selectedDraftItem !== null && placedDraftTarget !== target,
        reason: selectedDraftItem === null ? 'Select an item first'
          : placedDraftTarget === target ? 'Destination is occupied' : null,
        place: () => {
          if (selectedDraftItem === null) throw new Error('No selected draft item');
          placedDraftTarget = String(target);
          selectedDraftItem = null;
          board.placementDraft = placementBinding();
        },
      }),
    });
    board.placementDraft = placementBinding();
    await board.updateComplete;
    const placementInitialReason = root?.querySelector('#space-list button')?.getAttribute('title');
    selectedDraftItem = 'tile-a';
    board.placementDraft = placementBinding();
    await board.updateComplete;
    const placementButtons = root?.querySelectorAll<HTMLButtonElement>('#space-list button');
    placementButtons?.[1]?.click();
    await board.updateComplete;
    const placementAfterReason = root?.querySelectorAll<HTMLButtonElement>('#space-list button')[1]?.getAttribute('title');
    board.action = {
      candidates: [], preview: { kind: 'ready' }, get: () => undefined,
      ensurePreview: async () => ({ kind: 'ready' }), refreshPreview: async () => ({ kind: 'ready' }),
      subscribe: () => () => undefined,
    };
    await board.updateComplete;
    await board.updateComplete;
    const ambiguousPlacementError = root?.querySelector('#status')?.textContent?.replace(/\s+/g, ' ').trim();
    board.action = null;
    await board.updateComplete;
    board.disabledSpaces = [0];
    await board.updateComplete;
    await board.updateComplete;
    const ambiguousDisabledError = root?.querySelector('#status')?.textContent?.replace(/\s+/g, ' ').trim();
    board.disabledSpaces = [];
    board.placementDraft = null;
    await board.updateComplete;

    // Rapid replacement must leave only the final descriptor installed.
    board.artwork = rasterBoardArtwork({ src, spaces: [{ ...spaces[0]!, label: 'Stale Harbor' }] });
    board.artwork = rasterBoardArtwork({ src, spaces: [{ ...spaces[0]!, label: 'Final Harbor' }] });
    await board.updateComplete;
    await waitForLoad();
    const finalLabel = root?.querySelector('#space-list button')?.textContent?.trim();

    const errorFor = async (properties: {
      artwork: unknown; svgUrl?: string; geometry?: unknown; pathOverlays?: readonly unknown[];
    }) => {
      const invalid = document.createElement('boardgame-spatial-board') as typeof board;
      Object.assign(invalid, properties);
      document.body.append(invalid);
      await invalid.updateComplete;
      for (let attempt = 0; attempt < 40; attempt++) {
        const status = invalid.shadowRoot?.querySelector('#status')?.textContent?.replace(/\s+/g, ' ').trim();
        if (status?.includes('could not be loaded')) {
          invalid.remove();
          return status;
        }
        await new Promise(resolve => setTimeout(resolve, 0));
      }
      invalid.remove();
      return '<missing error>';
    };
    const errors = await Promise.all([
      errorFor({ artwork: rasterBoardArtwork({ src, spaces }), svgUrl: '/also.svg' }),
      errorFor({ artwork: rasterBoardArtwork({ src, spaces }), geometry: () => ({ spaces: [] }) }),
      errorFor({ artwork: rasterBoardArtwork({ src: 'data:image/png;base64,not-an-image', spaces }) }),
      errorFor({
        artwork: rasterBoardArtwork({ src, spaces }),
        pathOverlays: [{ id: 'lost', label: 'Lost route', spaces: ['harbor', 'unknown'] }],
      }),
    ]);
    const lateInvalid = document.createElement('boardgame-spatial-board') as typeof board;
    lateInvalid.artwork = rasterBoardArtwork({ src, spaces });
    document.body.append(lateInvalid);
    for (let attempt = 0; attempt < 40 && !lateInvalid.svgLoaded; attempt++) {
      await new Promise(resolve => setTimeout(resolve, 0));
    }
    lateInvalid.geometry = () => ({ spaces: [] });
    await lateInvalid.updateComplete;
    for (let attempt = 0; attempt < 40
      && !lateInvalid.shadowRoot?.querySelector('#status')?.textContent?.includes('geometry is only valid'); attempt++) {
      await new Promise(resolve => setTimeout(resolve, 0));
    }
    const lateGeometryError = lateInvalid.shadowRoot?.querySelector('#status')?.textContent?.replace(/\s+/g, ' ').trim();
    const lateGeometrySvgCount = lateInvalid.shadowRoot?.querySelectorAll('#container > svg').length;
    lateInvalid.remove();
    return {
      fitPreserves: Object.fromEntries(Object.entries(fits).map(([key, value]) => [key, value.preserve])),
      maxFocusError: Math.max(...Object.values(fits).map(value => value.focusError)),
      maxZoomedPieceError: Math.max(...Object.values(fits).map(value => value.pieceError)),
      maxPathError: Math.max(...Object.values(fits).map(value => value.pathError)),
      viewportScale: viewport.view.scale,
      containPieceError,
      finalLabel,
      groupedLabels,
      inactiveMarket,
      groupActivations,
      unknownGroupError,
      recoveredGroupError,
      generatedGroups: groupedRegions.map(region => region.getAttribute('data-board-group')),
      pathDescription,
      pathTone,
      pathWidth,
      placementInitialReason,
      placedDraftTarget,
      placementAfterReason,
      ambiguousPlacementError,
      ambiguousDisabledError,
      errors,
      lateGeometryError,
      lateGeometrySvgCount,
      rasterImageCount: root?.querySelectorAll('#container > svg image').length,
      fallbackLabels: [...(root?.querySelectorAll('#space-list button') ?? [])].map(button => button.textContent?.trim()),
    };
  });
  expect(result).toMatchObject({
    fitPreserves: { contain: 'xMidYMid meet', cover: 'xMidYMid slice', fill: 'none' },
    finalLabel: 'Final Harbor',
    rasterImageCount: 1,
    fallbackLabels: ['Final Harbor'],
    groupedLabels: ['Harbor', 'Road'],
    inactiveMarket: true,
    groupActivations: 1,
    unknownGroupError: expect.stringContaining('actionGroup "missing" has no geometry'),
    recoveredGroupError: '',
    generatedGroups: ['nodes', 'tiles', 'nodes'],
    pathDescription: 'Trade route from Harbor through Road to Market',
    pathTone: 'secondary',
    pathWidth: '6',
    placementInitialReason: 'Select an item first',
    placedDraftTarget: 'market',
    placementAfterReason: 'Select an item first',
    ambiguousPlacementError: expect.stringContaining('action and placementDraft are mutually exclusive'),
    ambiguousDisabledError: expect.stringContaining('placementDraft and disabledSpaces are mutually exclusive'),
    errors: [
      expect.stringContaining('choose svgUrl or artwork, not both'),
      expect.stringContaining('geometry is only valid with svgUrl'),
      expect.stringContaining('could not be decoded'),
      expect.stringContaining('references unknown space "unknown"'),
    ],
    lateGeometryError: expect.stringContaining('geometry is only valid with svgUrl'),
    lateGeometrySvgCount: 0,
    viewportScale: 2,
  });
  expect(result.maxFocusError).toBeLessThanOrEqual(1);
  expect(result.containPieceError).toBeLessThanOrEqual(1);
  expect(result.maxZoomedPieceError).toBeLessThanOrEqual(1);
  expect(result.maxPathError).toBeLessThanOrEqual(1);
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
      rejects({ pathOverlays: [{ id: 'short', label: 'Short', spaces: ['a'] }] }),
      rejects({ pathOverlays: [{ id: 'odd', label: 'Odd', spaces: ['a', 'b'], tone: 'rainbow' }] }),
      rejects({ pathOverlays: [
        { id: 'same', label: 'First', spaces: ['a', 'b'] },
        { id: 'same', label: 'Second', spaces: ['b', 'c'] },
      ] }),
      rejects({ pathOverlays: [{ id: 'verbose', label: 'x'.repeat(1025), spaces: ['a', 'b'] }] }),
      rejects({ pathOverlays: Array.from({ length: 17 }, (_, index) => ({
        id: `large-${index}`,
        label: `Large route ${index}`,
        spaces: Array.from({ length: 256 }, (_, point) => point % 2 ? 'a' : 'b'),
      })) }),
    ]);
  });
  expect(messages).toEqual([
    expect.stringContaining('choose componentView or componentViews, not both'),
    expect.stringContaining('componentViews has 1 entries for 2 effective stack layers'),
    expect.stringContaining('componentViews has 1 entries for 0 effective stack layers'),
    expect.stringContaining('requires 2 through 256 spaces'),
    expect.stringContaining('has unknown tone "rainbow"'),
    expect.stringContaining('duplicate path overlay id "same"'),
    expect.stringContaining('accessible label of at most 1024 characters'),
    expect.stringContaining('exceeds the 4096-point total limit'),
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
