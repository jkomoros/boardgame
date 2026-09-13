import { test, expect } from '@playwright/test';
import { prepareRendererFixturePage } from './renderer-fixture-helpers.js';

test('animation and legality changes keep action layout stable while errors remain visible', async ({ page }) => {
  const diagnostics = await prepareRendererFixturePage(page);
  try {
    const result = await page.evaluate(async () => {
      await import('/src/components/boardgame-action-button.ts');
      const { createMoveAction, MoveSubmissionGate, notifyMoveActionLiveStateChanged } = await import('/src/moves/action.ts');
      let animating = false;
      let legal = true;
      let connected = true;
      const gate = new MoveSubmissionGate();
      const transport = { submit: async () => ({ kind: 'success' as const }) };
      const action = createMoveAction<'Take', 'Take', { Take: Record<string, never> }>('Take', {
        currentClientSchemaFingerprint: () => 'schema',
        currentServerSchemaFingerprint: () => 'schema',
        currentTransport: () => connected ? transport : null,
        currentPreviewTransport: () => null,
        currentTargetPreviewTransport: () => null,
        currentGate: () => gate,
        nextRequestID: () => 'fixture-request',
        validate: () => [], serialize: () => ({}), actionCache: new Map(),
      }, {
        snapshotKey: 'fixture', currentSnapshotKey: () => 'fixture',
        snapshotVersion: 1, currentSnapshotVersion: () => 1,
        viewingAsPlayer: 0, proposingAsPlayer: 0, proposingAsAdmin: false,
        currentLegality: () => ({ legalForPlayer: legal, legalForAnyone: legal }),
        currentAnimating: () => animating, baselineLegalityApplies: true,
      });
      const button = document.createElement('boardgame-action-button');
      button.textContent = 'Take card'; button.action = action;
      document.body.append(button);
      const read = async () => {
        notifyMoveActionLiveStateChanged(action);
        await button.updateComplete;
        const native = button.shadowRoot!.querySelector('button')!;
        const status = button.shadowRoot!.querySelector<HTMLElement>('#status');
        return {
          height: button.getBoundingClientRect().height,
          disabled: native.disabled,
          title: native.title,
          description: native.getAttribute('aria-describedby'),
          reason: action.reason?.code ?? null,
          text: status?.textContent ?? '',
          statusHeight: status?.getBoundingClientRect().height ?? 0,
          live: status?.getAttribute('aria-live') ?? null,
        };
      };
      const initial = await read();
      animating = true;
      const busy = await read();
      animating = false; legal = false;
      const unavailable = await read();
      legal = true; connected = false;
      const error = await read();
      button.remove();
      return { initial, busy, unavailable, error };
    });
    expect(result.initial.disabled).toBe(false);
    for (const state of [result.busy, result.unavailable]) {
      expect(state.disabled).toBe(true);
      expect(state.height).toBe(result.initial.height);
      expect(state.description).toBe('status');
      expect(state.title).toBe(state.text);
      expect(state.text.length).toBeGreaterThan(0);
      expect(state.statusHeight).toBe(1);
      expect(state.live).toBe('off');
    }
    expect(result.busy.reason).toBe('animation-running');
    expect(result.unavailable.reason).toBe('move-not-possible');
    expect(result.error.reason).toBe('transport-unavailable');
    expect(result.error.statusHeight).toBeGreaterThan(1);
    expect(result.error.live).toBe('polite');
    diagnostics.assertEmpty();
  } finally { diagnostics.stop(); }
});
