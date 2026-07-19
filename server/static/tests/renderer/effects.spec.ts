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

test('motion anchors and sanitized trails decorate real structural lifecycle', async ({ page }) => {
  await page.emulateMedia({ reducedMotion: 'no-preference' });
  const diagnostics = await prepareRendererFixturePage(page);
  try {
    const result = await page.evaluate(async () => {
      await import('/src/components/boardgame-effect-layer.ts');
      const { fx } = await import('/src/effects/effect-spec.ts');
      const eventObservers = new Set<(event: any) => void>();
      const motionSource = {
        observeStructuralMotionEvents(observer: (event: any) => void) {
          eventObservers.add(observer);
          return () => eventObservers.delete(observer);
        },
      };
      const layer = document.createElement('boardgame-effect-layer') as HTMLElement & {
        updateComplete: Promise<unknown>;
        shadowRoot: ShadowRoot;
        configure(config: object): void;
        beginMotionTransition(expected: boolean): void;
        playTransition(effect: object): {
          finished: Promise<{ status: string; reason?: string }>;
        };
      };
      document.body.append(layer);
      await layer.updateComplete;
      layer.configure({
        anchorRoot: document,
        seedScope: 'motion-anchor:2',
        theme: {},
        animationContext: null,
        motionSource,
      });
      layer.beginMotionTransition(true);
      const departure = layer.playTransition(fx.pulse({
        at: fx.motion('card-17', 'departure'),
        tone: 'attention',
        advanced: { durationMs: 120 },
      }));
      const arrival = layer.playTransition(fx.burst({
        at: fx.motion('card-17'),
        tone: 'reward',
        advanced: { durationMs: 120, count: 3 },
      }));
      const trail = layer.playTransition(fx.sequence([
        fx.travel({
          from: fx.point(5, 5),
          to: fx.point(15, 15),
          advanced: { durationMs: 600 },
        }),
        fx.decorateMotion({
          subject: 'card-17',
          trail: {
            tone: 'magic',
            intensity: 'small',
            advanced: { echoes: 3, lagMs: 4, opacity: 0.3 },
          },
        }),
      ]));
      const before = {
        pulses: layer.shadowRoot.querySelectorAll('.pulse').length,
        particles: layer.shadowRoot.querySelectorAll('.particle').length,
        trails: layer.shadowRoot.querySelectorAll('.trail-echo').length,
      };
      const segment = {
        subjectId: 'card-17',
        visualSubject: { kind: 'silhouette', shape: 'rounded-rectangle' },
        path: {
          kind: 'travel',
          from: { space: 'viewport', left: 20, top: 30, width: 40, height: 20 },
          to: { space: 'viewport', left: 200, top: 100, width: 60, height: 40 },
        },
        execution: {
          status: 'started',
          animations: [{
            channel: 'host:transform',
            delayMs: 0,
            durationMs: 240,
            endDelayMs: 0,
            iterations: 1,
            easing: 'ease-in-out',
            fill: 'none',
          }],
        },
      };
      for (const observer of eventObservers) observer({
        source: 'flip', generation: 2, kind: 'started', subjectId: 'card-17', segment,
      });
      const pulse = layer.shadowRoot.querySelector<HTMLElement>('.pulse');
      const trailEcho = layer.shadowRoot.querySelector<HTMLElement>('.trail-echo');
      const trailAnimation = trailEcho?.getAnimations()[0];
      const trailFrames = trailAnimation?.effect instanceof KeyframeEffect
        ? trailAnimation.effect.getKeyframes()
        : [];
      const atDeparture = {
        pulses: layer.shadowRoot.querySelectorAll('.pulse').length,
        particles: layer.shadowRoot.querySelectorAll('.particle').length,
        trails: layer.shadowRoot.querySelectorAll('.trail-echo').length,
        left: pulse?.style.left,
        top: pulse?.style.top,
        trailRadius: trailEcho?.style.borderRadius,
        trailFrom: trailFrames[0]?.transform,
        trailTo: trailFrames.at(-1)?.transform,
      };
      const departureResult = await departure.finished;
      for (const observer of eventObservers) observer({
        source: 'flip', generation: 2, kind: 'finished', subjectId: 'card-17', segment,
      });
      const particle = layer.shadowRoot.querySelector<HTMLElement>('.particle');
      const atArrival = {
        particles: layer.shadowRoot.querySelectorAll('.particle').length,
        left: particle?.style.left,
        top: particle?.style.top,
      };
      const arrivalResult = await arrival.finished;
      const trailResult = await trail.finished;

      layer.beginMotionTransition(true);
      const interruptedTrail = layer.playTransition(fx.trail({
        subject: 'card-interrupted', advanced: { echoes: 2, lagMs: 4 },
      }));
      const interruptedSegment = {
        ...segment,
        subjectId: 'card-interrupted',
        execution: {
          ...segment.execution,
          animations: [{ ...segment.execution.animations[0], durationMs: 1000 }],
        },
      };
      for (const observer of eventObservers) observer({
        source: 'flip', generation: 3, kind: 'started',
        subjectId: 'card-interrupted', segment: interruptedSegment,
      });
      const interruptedBeforeCancel = layer.shadowRoot.querySelectorAll('.trail-echo').length;
      for (const observer of eventObservers) observer({
        source: 'flip', generation: 3, kind: 'cancelled',
        subjectId: 'card-interrupted', segment: interruptedSegment,
      });
      const interruptedResult = await interruptedTrail.finished;
      const interruptedAfterCancel = layer.shadowRoot.querySelectorAll('.trail-echo').length;

      layer.beginMotionTransition(true);
      const unsafeSubject = layer.playTransition(fx.trail({ subject: 'unsafe-card' }));
      for (const observer of eventObservers) observer({
        source: 'flip', generation: 4, kind: 'started', subjectId: 'unsafe-card',
        segment: { ...segment, subjectId: 'unsafe-card', visualSubject: undefined },
      });
      const unsafeSubjectResult = await unsafeSubject.finished;

      layer.beginMotionTransition(true);
      const stationary = layer.playTransition(fx.trail({ subject: 'stationary-card' }));
      for (const observer of eventObservers) observer({
        source: 'flip', generation: 5, kind: 'started', subjectId: 'stationary-card',
        segment: {
          ...segment,
          subjectId: 'stationary-card',
          path: { ...segment.path, kind: 'stationary' },
        },
      });
      const stationaryResult = await stationary.finished;

      layer.beginMotionTransition(true);
      const missing = layer.playTransition(fx.pulse({
        at: fx.motion('missing-card'),
        advanced: { durationMs: 120 },
      }));
      for (const observer of eventObservers) observer({
        source: 'flip', generation: 5, kind: 'generation-settled',
      });
      const missingResult = await missing.finished;

      layer.beginMotionTransition(true);
      const skipped = layer.playTransition(fx.pulse({
        at: fx.motion('skipped-card'),
        advanced: { durationMs: 120 },
      }));
      for (const observer of eventObservers) observer({
        source: 'flip',
        generation: 4,
        kind: 'skipped',
        subjectId: 'skipped-card',
        segment: { subjectId: 'skipped-card' },
      });
      const skippedResult = await skipped.finished;
      return {
        before,
        atDeparture,
        atArrival,
        departureResult,
        arrivalResult,
        trailResult,
        interruptedBeforeCancel,
        interruptedAfterCancel,
        interruptedResult,
        unsafeSubjectResult,
        stationaryResult,
        missingResult,
        skippedResult,
      };
    });

    expect(result.before).toEqual({ pulses: 0, particles: 0, trails: 0 });
    expect(result.atDeparture).toEqual({
      pulses: 1,
      particles: 0,
      trails: 3,
      left: '40px',
      top: '40px',
      trailRadius: '12%',
      trailFrom: 'translate(40px, 40px) translate(-50%, -50%)',
      trailTo: 'translate(230px, 120px) translate(-50%, -50%)',
    });
    expect(result.atArrival).toEqual({ particles: 3, left: '230px', top: '120px' });
    expect(result.departureResult).toEqual({ status: 'finished' });
    expect(result.arrivalResult).toEqual({ status: 'finished' });
    expect(result.trailResult).toEqual({ status: 'finished' });
    expect(result.interruptedBeforeCancel).toBe(2);
    expect(result.interruptedAfterCancel).toBe(0);
    expect(result.interruptedResult).toEqual({ status: 'cancelled' });
    expect(result.unsafeSubjectResult).toEqual({ status: 'skipped', reason: 'missing-subject' });
    expect(result.stationaryResult).toEqual({ status: 'skipped', reason: 'no-motion-path' });
    expect(result.missingResult).toEqual({ status: 'skipped', reason: 'missing-anchor' });
    expect(result.skippedResult).toEqual({ status: 'skipped', reason: 'motion-skipped' });
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
      const eventObservers = new Set<(event: any) => void>();
      const motionSource = {
        observeStructuralMotionEvents(observer: (event: any) => void) {
          eventObservers.add(observer);
          return () => eventObservers.delete(observer);
        },
      };
      const anchor = document.createElement('div');
      document.body.append(anchor);
      const layer = document.createElement('boardgame-effect-layer') as HTMLElement & {
        updateComplete: Promise<unknown>;
        shadowRoot: ShadowRoot;
        configure(config: object): void;
        play(effect: object): { finished: Promise<{ status: string }> };
        beginMotionTransition(expected: boolean): void;
        playTransition(effect: object): { finished: Promise<{ status: string }> };
      };
      document.body.append(layer);
      await layer.updateComplete;
      layer.configure({
        anchorRoot: document,
        seedScope: 'reduced:1',
        theme: {},
        animationContext: null,
        motionSource,
      });
      const handle = layer.play(fx.burst({ at: anchor, tone: 'confirm', intensity: 'large' }));
      const during = {
        particles: layer.shadowRoot.querySelectorAll('.particle').length,
        pulses: layer.shadowRoot.querySelectorAll('.pulse').length,
      };
      const settled = await handle.finished;
      layer.beginMotionTransition(true);
      const trail = layer.playTransition(fx.parallel([
        fx.trail({ subject: 'reduced-card', tone: 'magic' }),
      ], { timing: 'version' }));
      for (const observer of eventObservers) observer({
        source: 'flip', generation: 1, kind: 'finished', subjectId: 'reduced-card',
        segment: {
          subjectId: 'reduced-card',
          path: {
            kind: 'travel',
            from: { space: 'viewport', left: 10, top: 20, width: 40, height: 20 },
            to: { space: 'viewport', left: 100, top: 80, width: 40, height: 20 },
          },
        },
      });
      const reducedTrailDuring = {
        trails: layer.shadowRoot.querySelectorAll('.trail-echo').length,
        pulses: layer.shadowRoot.querySelectorAll('.pulse').length,
      };
      const trailSettled = await trail.finished;
      return {
        during,
        settled,
        reducedTrailDuring,
        trailSettled,
        final: layer.shadowRoot.querySelectorAll('.particle, .pulse').length,
      };
    });
    expect(result.during).toEqual({ particles: 0, pulses: 1 });
    expect(result.settled).toEqual({ status: 'finished' });
    expect(result.reducedTrailDuring).toEqual({ trails: 0, pulses: 1 });
    expect(result.trailSettled).toEqual({ status: 'finished' });
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
