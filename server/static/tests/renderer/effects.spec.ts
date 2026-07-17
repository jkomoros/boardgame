import { expect, test } from '@playwright/test';
import { prepareRendererFixturePage } from './renderer-fixture-helpers.js';

test('descriptor execution shares a document budget and settles composition', async ({ page }) => {
  await page.emulateMedia({ reducedMotion: 'no-preference' });
  const diagnostics = await prepareRendererFixturePage(page);
  try {
    const result = await page.evaluate(async () => {
      await import('/src/components/boardgame-effect-layer.ts');
      const { fx } = await import('/src/effects/effect-spec.ts');
      const { effectBudgetSnapshot } = await import('/src/effects/effect-budget.ts');
      const anchor = document.createElement('button');
      anchor.id = 'score-anchor';
      anchor.textContent = 'Score';
      document.body.append(anchor);

      const layers = await Promise.all(['a', 'b', 'c'].map(async suffix => {
        const layer = document.createElement('boardgame-effect-layer') as HTMLElement & {
          updateComplete: Promise<unknown>;
          shadowRoot: ShadowRoot;
          configure(config: object): void;
          play(effect: object): { finished: Promise<{ status: string }>; cancel(): void };
          cancelAll(): void;
        };
        document.body.append(layer);
        await layer.updateComplete;
        layer.configure({
          anchorRoot: document,
          seedScope: `fixture:7:${suffix}`,
          theme: {},
          animationContext: null,
        });
        return layer;
      }));

      const handles = layers.map((layer, index) => layer.play(fx.parallel([
        fx.burst({
          at: fx.anchor('score-anchor'),
          tone: 'reward',
          intensity: 'large',
          key: `score-${index}`,
          advanced: { count: 100, durationMs: 120 },
        }),
      ])));
      const initialCounts = layers.map(layer => layer.shadowRoot.querySelectorAll('.particle').length);
      const duringBudget = effectBudgetSnapshot(document);
      const results = await Promise.all(handles.map(handle => handle.finished));
      const afterBudget = effectBudgetSnapshot(document);

      const travel = layers[0].play(fx.travel({
        from: anchor,
        to: fx.point(320, 180),
        tone: 'magic',
        advanced: { durationMs: 1200 },
      }));
      const travelersBeforeCancel = layers[0].shadowRoot.querySelectorAll('.traveler').length;
      layers[0].cancelAll();
      const cancelled = await travel.finished;
      const gapSequence = layers[0].play(fx.sequence([
        fx.pulse({ at: anchor, advanced: { durationMs: 120 } }),
        fx.pulse({ at: anchor, advanced: { durationMs: 120 } }),
      ], { gapMs: 1000 }));
      await new Promise(resolve => setTimeout(resolve, 150));
      gapSequence.cancel();
      const gapCancelled = await gapSequence.finished;
      return {
        initialCounts,
        duringBudget,
        afterBudget,
        results,
        finalParticles: layers.reduce(
          (sum, layer) => sum + layer.shadowRoot.querySelectorAll('.particle').length,
          0,
        ),
        travelersBeforeCancel,
        cancelled,
        gapCancelled,
      };
    });

    expect(result.initialCounts).toEqual([24, 24, 12]);
    expect(result.duringBudget).toEqual({ effects: 3, particles: 60 });
    expect(result.afterBudget).toEqual({ effects: 0, particles: 0 });
    expect(result.results).toEqual([{ status: 'finished' }, { status: 'finished' }, { status: 'finished' }]);
    expect(result.finalParticles).toBe(0);
    expect(result.travelersBeforeCancel).toBe(1);
    expect(result.cancelled).toEqual({ status: 'cancelled' });
    expect(result.gapCancelled).toEqual({ status: 'cancelled' });
    diagnostics.assertEmpty();
  } finally {
    diagnostics.stop();
  }
});

test('named anchors stay renderer-scoped', async ({ page }) => {
  await page.emulateMedia({ reducedMotion: 'no-preference' });
  const diagnostics = await prepareRendererFixturePage(page);
  try {
    const result = await page.evaluate(async () => {
      await import('/src/components/boardgame-effect-layer.ts');
      const { fx } = await import('/src/effects/effect-spec.ts');
      const makeSurface = async (left: number) => {
        const host = document.createElement('section');
        document.body.append(host);
        const root = host.attachShadow({ mode: 'open' });
        const anchor = document.createElement('button');
        anchor.dataset.effectAnchor = 'score';
        anchor.style.position = 'fixed';
        anchor.style.left = `${left}px`;
        anchor.style.top = '100px';
        anchor.style.width = '40px';
        anchor.style.height = '20px';
        root.append(anchor);
        const layer = document.createElement('boardgame-effect-layer') as HTMLElement & {
          updateComplete: Promise<unknown>;
          shadowRoot: ShadowRoot;
          configure(config: object): void;
          play(effect: object): { finished: Promise<unknown> };
        };
        root.append(layer);
        await layer.updateComplete;
        layer.configure({ anchorRoot: root, seedScope: `surface:${left}`, theme: {}, animationContext: null });
        return { layer, anchor };
      };
      const first = await makeSurface(40);
      await makeSurface(440);
      const handle = first.layer.play(fx.pulse({ at: fx.anchor('score'), intensity: 'small' }));
      const pulse = first.layer.shadowRoot.querySelector<HTMLElement>('.pulse');
      const left = Number.parseFloat(pulse?.style.left ?? 'NaN');
      const expected = first.anchor.getBoundingClientRect().left + first.anchor.getBoundingClientRect().width / 2;
      await handle.finished;
      return { left, expected };
    });
    expect(result.left).toBeCloseTo(result.expected, 2);
    diagnostics.assertEmpty();
  } finally {
    diagnostics.stop();
  }
});

test('reduced motion substitutes a stationary emphasis', async ({ page }) => {
  await page.emulateMedia({ reducedMotion: 'reduce' });
  const diagnostics = await prepareRendererFixturePage(page);
  try {
    const result = await page.evaluate(async () => {
      await import('/src/components/boardgame-effect-layer.ts');
      const { fx } = await import('/src/effects/effect-spec.ts');
      const anchor = document.createElement('div');
      document.body.append(anchor);
      const layer = document.createElement('boardgame-effect-layer') as HTMLElement & {
        updateComplete: Promise<unknown>;
        shadowRoot: ShadowRoot;
        configure(config: object): void;
        play(effect: object): { finished: Promise<{ status: string }> };
      };
      document.body.append(layer);
      await layer.updateComplete;
      layer.configure({ anchorRoot: document, seedScope: 'reduced:1', theme: {}, animationContext: null });
      const handle = layer.play(fx.burst({ at: anchor, tone: 'confirm', intensity: 'large' }));
      const during = {
        particles: layer.shadowRoot.querySelectorAll('.particle').length,
        pulses: layer.shadowRoot.querySelectorAll('.pulse').length,
      };
      const settled = await handle.finished;
      return {
        during,
        settled,
        final: layer.shadowRoot.querySelectorAll('.particle, .pulse').length,
      };
    });
    expect(result.during).toEqual({ particles: 0, pulses: 1 });
    expect(result.settled).toEqual({ status: 'finished' });
    expect(result.final).toBe(0);
    diagnostics.assertEmpty();
  } finally {
    diagnostics.stop();
  }
});

test('render host plans authoritative effects exactly once per installed snapshot', async ({ page }) => {
  await page.emulateMedia({ reducedMotion: 'no-preference' });
  const diagnostics = await prepareRendererFixturePage(page);
  try {
    const result = await page.evaluate(async () => {
      const { html } = await import('/src/client.ts');
      const { BoardgameBaseGameRenderer } = await import('/src/components/boardgame-base-game-renderer.ts');
      const { BoardgameRenderGame } = await import('/src/components/boardgame-render-game.ts');
      const { fx } = await import('/src/effects/effect-spec.ts');
      const calls: Array<{ kind: string; move: string | null; before: number | null; after: number }> = [];

      class EffectFixtureRenderer extends BoardgameBaseGameRenderer<any, object, string, Record<string, object>> {
        override effectsForTransition(context: any) {
          calls.push({
            kind: context.kind,
            move: context.move?.Name ?? null,
            before: context.before?.Game.Score ?? null,
            after: context.after.Game.Score,
          });
          if (context.kind === 'initial') return [];
          return [fx.pulse({
            at: fx.anchor('score'),
            tone: 'reward',
            key: 'score-change',
            advanced: { durationMs: 120 },
          })];
        }

        override render() {
          return html`<div data-effect-anchor="score">${this.state?.Game.Score ?? 0}</div>`;
        }
      }
      if (!customElements.get('effect-fixture-renderer')) {
        customElements.define('effect-fixture-renderer', EffectFixtureRenderer);
      }

      const host = new BoardgameRenderGame() as BoardgameRenderGame & Record<string, any>;
      host.active = true;
      host.gameId = 'fixture';
      document.body.append(host);
      await host.updateComplete;
      const renderer = document.createElement('effect-fixture-renderer') as EffectFixtureRenderer;
      host.renderer = renderer as any;
      host.shadowRoot?.querySelector('#container')?.append(renderer);
      host.gameVersion = 1;
      host.snapshotEpoch = 1;
      host.transitionMove = null;
      host.state = { Game: { Score: 0 } } as any;
      await host.updateComplete;
      await renderer.updateComplete;
      await new Promise(resolve => setTimeout(resolve, 0));

      host.gameVersion = 2;
      host.snapshotEpoch = 2;
      host.transitionMove = { Name: 'Claim Point', Version: 2 };
      host.state = { Game: { Score: 1 } } as any;
      await host.updateComplete;
      await renderer.updateComplete;
      await new Promise(resolve => setTimeout(resolve, 0));

      // An unrelated renderer update must not replay the authoritative effect.
      renderer.requestUpdate();
      await renderer.updateComplete;
      const effectLayer = host.shadowRoot?.querySelector('boardgame-effect-layer');
      const during = effectLayer?.shadowRoot?.querySelectorAll('.pulse').length ?? 0;
      await new Promise(resolve => setTimeout(resolve, 180));
      const after = effectLayer?.shadowRoot?.querySelectorAll('.pulse').length ?? 0;
      return { calls, during, after };
    });

    expect(result.calls).toEqual([
      { kind: 'initial', move: null, before: null, after: 0 },
      { kind: 'transition', move: 'Claim Point', before: 0, after: 1 },
    ]);
    expect(result.during).toBe(1);
    expect(result.after).toBe(0);
    diagnostics.assertEmpty();
  } finally {
    diagnostics.stop();
  }
});
