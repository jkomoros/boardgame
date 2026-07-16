import AxeBuilder from '@axe-core/playwright';
import { expect, test } from '@playwright/test';
import {
  RENDERER_VIEWPORTS,
  focusWithKeyboard,
  prepareRendererFixturePage,
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
    const axeResult = await new AxeBuilder({ page }).include('[data-renderer-fixture]').analyze();
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
  await page.goto('/');
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

test('Tic-tac-toe fixture proposes native numeric targets and stays bounded at canonical widths', async ({ page }) => {
  const diagnostics = await prepareRendererFixturePage(page);
  try {
    const proposals = await page.evaluate(async () => {
      await import('/game-src/tictactoe/boardgame-render-game-tictactoe.ts');
      const { tictactoeRendererFixture } = await import(
        '/game-src/tictactoe/boardgame-render-fixtures-tictactoe.ts'
      );
      const { mountRendererFixture } = await import('/src/testing/renderer-fixture.ts');
      const handle = await mountRendererFixture(tictactoeRendererFixture);
      const board = handle.renderer.shadowRoot?.querySelector('boardgame-game-board');
      if (!(board instanceof HTMLElement)) throw new Error('Tic-tac-toe fixture did not render its board');
      board.dispatchEvent(new CustomEvent('space-tapped', {
        detail: { index: 1 },
        bubbles: true,
        composed: true,
      }));
      await Promise.resolve();
      return handle.proposals;
    });

    expect(proposals).toEqual([{
      requestID: 'fixture-v4-request-1',
      snapshotVersion: 4,
      name: 'Place Token',
      arguments: { Slot: '1' },
    }]);

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
        updateComplete: Promise<unknown>;
      };
      board.rows = 2;
      board.cols = 3;
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
