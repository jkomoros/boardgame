import { expect, test } from '@playwright/test';
import { prepareRendererFixturePage } from './renderer-fixture-helpers.js';

test('render-game preserves animation gate listeners and watchdog across reconnection', async ({ page }) => {
  const diagnostics = await prepareRendererFixturePage(page);
  try {
    const result = await page.evaluate(async () => {
      await import('/src/components/boardgame-render-game.ts');
      const host = document.createElement('boardgame-render-game') as HTMLElement & {
        updateComplete: Promise<unknown>;
        isAnimating: boolean;
        _resetAnimating(): void;
        _gate: { pendingCount: number };
      };
      const participant = document.createElement('div');
      host.append(participant);
      document.body.append(host);
      await host.updateComplete;

      host._resetAnimating();
      participant.dispatchEvent(new CustomEvent('will-animate', {
        bubbles: true,
        composed: true,
        detail: { ele: participant, expectedSettleMs: 100 },
      }));
      const pendingBeforeDetach = host._gate.pendingCount;

      host.remove();
      participant.dispatchEvent(new CustomEvent('animation-done', {
        bubbles: true,
        composed: true,
        detail: { ele: participant },
      }));
      const openAfterDetachedSettlement = host.isAnimating;

      document.body.append(host);
      await host.updateComplete;
      host._resetAnimating();
      participant.dispatchEvent(new CustomEvent('will-animate', {
        bubbles: true,
        composed: true,
        detail: { ele: participant, expectedSettleMs: 100 },
      }));
      const pendingAfterReconnect = host._gate.pendingCount;
      participant.dispatchEvent(new CustomEvent('animation-done', {
        bubbles: true,
        composed: true,
        detail: { ele: participant },
      }));
      const openAfterReconnectSettlement = host.isAnimating;
      host.remove();

      return {
        pendingBeforeDetach,
        openAfterDetachedSettlement,
        pendingAfterReconnect,
        openAfterReconnectSettlement,
      };
    });

    expect(result).toEqual({
      pendingBeforeDetach: 1,
      openAfterDetachedSettlement: false,
      pendingAfterReconnect: 1,
      openAfterReconnectSettlement: false,
    });
    diagnostics.assertEmpty();
  } finally {
    diagnostics.stop();
  }
});

test('an empty structural generation publishes settlement', async ({ page }) => {
  const diagnostics = await prepareRendererFixturePage(page);
  try {
    const result = await page.evaluate(async () => {
      await import('/src/components/boardgame-component-animator.ts');
      const animator = document.createElement('boardgame-component-animator') as HTMLElement & {
        updateComplete: Promise<unknown>;
        prepare(): void;
        animateFlip(): Promise<void>;
        observeStructuralMotionEvents(observer: (event: { kind: string }) => void): () => void;
        _solvedMotionPlan: { phase: string; segments: readonly unknown[] } | null;
      };
      document.body.append(animator);
      await animator.updateComplete;
      const events: string[] = [];
      const stop = animator.observeStructuralMotionEvents(event => events.push(event.kind));

      animator.prepare();
      await animator.animateFlip();
      stop();
      const snapshot = {
        phase: animator._solvedMotionPlan?.phase,
        segmentCount: animator._solvedMotionPlan?.segments.length,
        events,
      };
      animator.remove();
      return snapshot;
    });

    expect(result).toEqual({
      phase: 'settled',
      segmentCount: 0,
      events: ['generation-settled'],
    });
    diagnostics.assertEmpty();
  } finally {
    diagnostics.stop();
  }
});

test('every delayed transfer in a batch reaches active-observed independently', async ({ page }) => {
  await page.emulateMedia({ reducedMotion: 'no-preference' });
  const diagnostics = await prepareRendererFixturePage(page);
  try {
    const result = await page.evaluate(async () => {
      await import('/src/components/boardgame-component-animator.ts');
      await import('/src/components/boardgame-component-stack.ts');
      const { compileMotionTransferDeclarations } = await import('/src/motion/transfer.ts');
      const host = document.createElement('div');
      const root = host.attachShadow({ mode: 'open' });
      root.innerHTML = `
        <boardgame-component-stack id="registry"></boardgame-component-stack>
        <div id="transfer-source"></div>
        <div id="transfer-carrier-a"></div>
        <div id="transfer-carrier-b"></div>
      `;
      const animator = document.createElement('boardgame-component-animator') as HTMLElement & {
        updateComplete: Promise<unknown>;
        prepare(): void;
        installMotionTransfers(declarations: readonly unknown[]): void;
        animateFlip(): Promise<void>;
        _lastExplicitMotionPlan: null | {
          segments: Array<{ execution: { status: string } }>;
        };
      };
      root.append(animator);
      document.body.append(host);
      const source = root.querySelector<HTMLElement>('#transfer-source')!;
      const carrierA = root.querySelector<HTMLElement>('#transfer-carrier-a')!;
      const carrierB = root.querySelector<HTMLElement>('#transfer-carrier-b')!;
      Object.assign(source.style, {
        position: 'fixed', left: '0px', top: '0px', width: '10px', height: '10px',
      });
      Object.assign(carrierA.style, {
        position: 'fixed', left: '100px', top: '0px', width: '10px', height: '10px',
      });
      Object.assign(carrierB.style, {
        position: 'fixed', left: '200px', top: '0px', width: '10px', height: '10px',
      });
      await Promise.all([
        animator.updateComplete,
        (root.querySelector('#registry') as HTMLElement & { updateComplete: Promise<unknown> })
          .updateComplete,
      ]);

      const localStartAtMs = Date.now() + 10_000;
      animator.prepare();
      animator.installMotionTransfers(compileMotionTransferDeclarations([
        {
          key: 'deal:0', subjectId: 'opaque-0', source: 'transfer-source',
          carrier: 'transfer-carrier-a', durationMs: 100,
          timing: { localStartAtMs },
        },
        {
          key: 'deal:1', subjectId: 'opaque-1', source: 'transfer-source',
          carrier: 'transfer-carrier-b', durationMs: 100,
          timing: { localStartAtMs },
        },
      ]));
      const settled = animator.animateFlip();
      for (let frame = 0; frame < 20; frame++) {
        if (animator._lastExplicitMotionPlan?.segments.every(
          segment => segment.execution.status === 'armed',
        )) break;
        await new Promise(requestAnimationFrame);
      }

      const animations = [carrierA, carrierB].map(carrier => carrier.getAnimations()[0]);
      for (const animation of animations) {
        if (!(animation?.effect instanceof KeyframeEffect)) {
          throw new Error('transfer animation was not armed');
        }
        animation.currentTime = Number(animation.effect.getTiming().delay) + 1;
      }
      await new Promise(requestAnimationFrame);
      const statusesAtActivation = animator._lastExplicitMotionPlan?.segments.map(
        segment => segment.execution.status,
      );
      for (const animation of animations) animation.finish();
      await settled;
      host.remove();
      return statusesAtActivation;
    });

    expect(result).toEqual(['active-observed', 'active-observed']);
    diagnostics.assertEmpty();
  } finally {
    diagnostics.stop();
  }
});

test('Hand and Table cancel deferred compatibility flights when ownership changes', async ({ page }) => {
  const diagnostics = await prepareRendererFixturePage(page);
  try {
    const result = await page.evaluate(async () => {
      const { BoardgameHandViewBase } = await import('/src/components/boardgame-hand-view-base.ts');
      const { BoardgameTableViewBase } = await import('/src/components/boardgame-table-view-base.ts');
      const { html } = await import('/src/client.ts');

      class LifecycleHand extends BoardgameHandViewBase<any, any, any, any> {
        calls = 0;
        protected override get animator(): any {
          return { animateBetween: () => { this.calls++; } };
        }
      }
      class LifecycleTable extends BoardgameTableViewBase<any, any, any, any> {
        calls = 0;
        protected override get animator(): any {
          return { animateBetween: () => { this.calls++; } };
        }
        override render() {
          return html`<div id="deal-source"></div><div id="stub:p0:hand"></div>`;
        }
      }
      customElements.define('lifecycle-deferred-hand', LifecycleHand);
      customElements.define('lifecycle-deferred-table', LifecycleTable);
      const hand = document.createElement('lifecycle-deferred-hand') as LifecycleHand;
      const table = document.createElement('lifecycle-deferred-table') as LifecycleTable;
      document.body.append(hand, table);

      hand.viewingAsPlayer = 0;
      hand.state = { Players: [{ Hand: { IDs: ['existing'] } }] } as any;
      table.state = { Players: [{ Hand: { Indexes: [] } }] } as any;
      await Promise.all([hand.updateComplete, table.updateComplete]);

      hand.state = { Players: [{ Hand: { IDs: ['existing', 'incoming'] } }] } as any;
      table.state = { Players: [{ Hand: { Indexes: [-1] } }] } as any;
      await Promise.all([hand.updateComplete, table.updateComplete]);
      await new Promise(requestAnimationFrame);
      hand.remove();
      table.remove();
      await new Promise(requestAnimationFrame);
      const detached = { handCalls: hand.calls, tableCalls: table.calls };

      document.body.append(hand, table);
      hand.state = { Players: [{ Hand: { IDs: ['existing', 'incoming', 'later'] } }] } as any;
      table.state = { Players: [{ Hand: { Indexes: [-1, -1] } }] } as any;
      await Promise.all([hand.updateComplete, table.updateComplete]);
      await new Promise(requestAnimationFrame);
      hand.autoFlyIncoming = false;
      table.autoFlyDeals = false;
      await Promise.all([hand.updateComplete, table.updateComplete]);
      await new Promise(requestAnimationFrame);

      return {
        detached,
        disabled: { handCalls: hand.calls, tableCalls: table.calls },
      };
    });

    expect(result).toEqual({
      detached: { handCalls: 0, tableCalls: 0 },
      disabled: { handCalls: 0, tableCalls: 0 },
    });
    diagnostics.assertEmpty();
  } finally {
    diagnostics.stop();
  }
});
