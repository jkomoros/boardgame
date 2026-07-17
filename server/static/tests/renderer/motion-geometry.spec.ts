import { expect, test } from '@playwright/test';
import { prepareRendererFixturePage } from './renderer-fixture-helpers.js';

test('animateBetween aligns differently-sized endpoints by viewport center', async ({ page }) => {
  await page.emulateMedia({ reducedMotion: 'no-preference' });
  const diagnostics = await prepareRendererFixturePage(page);
  try {
    const keyframes = await page.evaluate(async () => {
      await import('/src/components/boardgame-component-animator.ts');

      const animator = document.createElement('boardgame-component-animator') as HTMLElement & {
        updateComplete: Promise<unknown>;
        animateBetween(
          real: HTMLElement,
          stub: HTMLElement,
          durationMs: number,
          options: { timing: 'immediate' },
        ): Promise<void>;
      };
      const real = document.createElement('div');
      const stub = document.createElement('div');
      Object.assign(real.style, {
        position: 'fixed', left: '200px', top: '100px', width: '20px', height: '10px',
      });
      Object.assign(stub.style, {
        position: 'fixed', left: '20px', top: '30px', width: '40px', height: '30px',
      });
      document.body.append(animator, real, stub);
      await animator.updateComplete;

      const finished = animator.animateBetween(real, stub, 10_000, { timing: 'immediate' });
      await new Promise(requestAnimationFrame);
      const animation = real.getAnimations()[0];
      if (!animation || !(animation.effect instanceof KeyframeEffect)) {
        throw new Error('animateBetween did not create a WAAPI keyframe animation');
      }
      const result = animation.effect.getKeyframes().map(frame => frame.transform);
      animation.finish();
      await finished;
      return result;
    });

    // real center = (210, 105), stub center = (40, 45).
    expect(keyframes[0]).toBe('translate(-170px, -60px)');
    expect(keyframes[1]).toBe('none');
    diagnostics.assertEmpty();
  } finally {
    diagnostics.stop();
  }
});

test('shared timing keeps raw and gated card flights in the same version window', async ({ page }) => {
  await page.emulateMedia({ reducedMotion: 'no-preference' });
  const diagnostics = await prepareRendererFixturePage(page);
  try {
    const result = await page.evaluate(async () => {
      await import('/src/components/boardgame-component-animator.ts');
      await import('/src/components/boardgame-animatable-item.ts');

      type FlightAnimator = HTMLElement & {
        updateComplete: Promise<unknown>;
        animationContext: {
          version: number;
          startAtMs: number;
          slotDurationMs: number;
          maxAnimationDurationMs: number;
        };
        animateBetween(
          real: HTMLElement,
          stub: HTMLElement,
          durationMs: number,
        ): Promise<void>;
      };
      type Animatable = HTMLElement & {
        updateComplete: Promise<unknown>;
        isAnimating: boolean;
        settled(): Promise<void>;
      };

      const animator = document.createElement('boardgame-component-animator') as FlightAnimator;
      const item = document.createElement('boardgame-animatable-item') as Animatable;
      const raw = document.createElement('div');
      const stub = document.createElement('div');
      for (const target of [item, raw]) {
        Object.assign(target.style, {
          position: 'fixed', left: '200px', top: '100px', width: '20px', height: '10px',
        });
      }
      Object.assign(stub.style, {
        position: 'fixed', left: '20px', top: '30px', width: '40px', height: '30px',
      });
      document.body.append(animator, item, raw, stub);
      await Promise.all([animator.updateComplete, item.updateComplete]);
      animator.animationContext = {
        version: 9,
        startAtMs: Date.now() + 300,
        slotDurationMs: 1_000,
        maxAnimationDurationMs: 600,
      };

      let willAnimate = 0;
      let animationDone = 0;
      document.addEventListener('will-animate', () => { willAnimate += 1; });
      document.addEventListener('animation-done', () => { animationDone += 1; });
      const itemFinished = animator.animateBetween(item, stub, 400);
      const rawFinished = animator.animateBetween(raw, stub, 400);
      await new Promise(requestAnimationFrame);

      const itemAnimation = item.getAnimations()[0];
      const rawAnimation = raw.getAnimations()[0];
      if (!(itemAnimation?.effect instanceof KeyframeEffect)
        || !(rawAnimation?.effect instanceof KeyframeEffect)) {
        throw new Error('both flight paths must create WAAPI animations');
      }
      const itemTiming = itemAnimation.effect.getTiming();
      const rawTiming = rawAnimation.effect.getTiming();
      const gatedDuring = item.isAnimating;
      itemAnimation.cancel();
      rawAnimation.cancel();
      await Promise.all([itemFinished, rawFinished, item.settled()]);

      return {
        itemTiming: {
          delay: itemTiming.delay,
          duration: itemTiming.duration,
          fill: itemTiming.fill,
        },
        rawTiming: {
          delay: rawTiming.delay,
          duration: rawTiming.duration,
          fill: rawTiming.fill,
        },
        gatedDuring,
        gatedAfter: item.isAnimating,
        willAnimate,
        animationDone,
      };
    });

    expect(result.itemTiming.duration).toBe(400);
    expect(result.rawTiming.duration).toBe(400);
    expect(result.itemTiming.fill).toBe('backwards');
    expect(result.rawTiming.fill).toBe('backwards');
    expect(Math.abs(result.itemTiming.delay - result.rawTiming.delay)).toBeLessThan(30);
    expect(result.gatedDuring).toBe(true);
    expect(result.gatedAfter).toBe(false);
    expect(result.willAnimate).toBe(1);
    expect(result.animationDone).toBe(1);
    diagnostics.assertEmpty();
  } finally {
    diagnostics.stop();
  }
});

test('structural plans publish before playback and invalidate on interruption', async ({ page }) => {
  await page.emulateMedia({ reducedMotion: 'no-preference' });
  const diagnostics = await prepareRendererFixturePage(page);
  try {
    const result = await page.evaluate(async () => {
      await import('/src/components/boardgame-component-animator.ts');
      await import('/src/components/boardgame-component-stack.ts');
      const { cardView } = await import('/src/client.ts');

      const animator = document.createElement('boardgame-component-animator') as HTMLElement & {
        updateComplete: Promise<unknown>;
        prepare(): void;
        animateFlip(): Promise<void>;
        _solvedMotionPlan: null | {
          generation: number;
          phase: string;
          segments: Array<{
            subjectId: string;
            presence: string;
            provenance: { kind: string };
            transform?: { before: string; after: string };
            timingRequest: { policy: string; delayMs: number; durationMs: number };
          }>;
        };
      };
      const stack = document.createElement('boardgame-component-stack') as HTMLElement & {
        stack: unknown;
        componentView: unknown;
        updateComplete: Promise<unknown>;
      };
      stack.style.setProperty('--animation-length', '80ms');
      stack.componentView = cardView({});
      stack.stack = {
        Deck: 'cards',
        Indexes: [0],
        IDs: ['card-plan'],
        IDsLastSeen: {},
        ShuffleCount: 0,
        Size: 1,
        GameName: 'motion-plan-test',
        Components: [{
          Index: 0,
          Values: { rank: 'A' },
          Deck: 'cards',
          GameName: 'motion-plan-test',
          ID: 'card-plan',
        }],
      };
      document.body.append(animator, stack);
      await Promise.all([animator.updateComplete, stack.updateComplete]);
      const card = stack.querySelector<HTMLElement>('#card-plan');
      if (!card) throw new Error('fixture card was not materialized');
      await (card as HTMLElement & { updateComplete: Promise<unknown> }).updateComplete;

      const waitForPlan = async () => {
        for (let frame = 0; frame < 20; frame += 1) {
          if (animator._solvedMotionPlan) return animator._solvedMotionPlan;
          await new Promise(requestAnimationFrame);
        }
        throw new Error('structural motion plan was not published');
      };

      animator.prepare();
      const nullDuringMeasurement = animator._solvedMotionPlan === null;
      card.style.transform = 'translateX(40px)';
      const firstFinished = animator.animateFlip();
      const first = await waitForPlan();
      const playingAfterPublish = card.getAnimations().length > 0;

      animator.prepare();
      const invalidatedImmediately = animator._solvedMotionPlan === null;
      await firstFinished;

      card.style.transform = 'translateX(80px)';
      await animator.animateFlip();
      const second = animator._solvedMotionPlan;
      if (!second) throw new Error('second structural motion plan was not published');

      return {
        nullDuringMeasurement,
        playingAfterPublish,
        invalidatedImmediately,
        first: {
          frozen: Object.isFrozen(first) && Object.isFrozen(first.segments),
          generation: first.generation,
          phase: first.phase,
          segment: first.segments[0],
        },
        second: {
          generation: second.generation,
          segment: second.segments[0],
        },
      };
    });

    expect(result.nullDuringMeasurement).toBe(true);
    expect(result.playingAfterPublish).toBe(true);
    expect(result.invalidatedImmediately).toBe(true);
    expect(result.first.frozen).toBe(true);
    expect(result.first.phase).toBe('ready-to-play');
    expect(result.first.segment).toMatchObject({
      subjectId: 'card-plan',
      presence: 'retained',
      provenance: { kind: 'identity' },
      transform: { before: '', after: 'translateX(40px)' },
      timingRequest: { policy: 'version', delayMs: 0, durationMs: 80 },
    });
    expect(result.second.generation).toBe(result.first.generation + 1);
    expect(result.second.segment.transform).toEqual({
      before: 'translateX(40px)',
      after: 'translateX(80px)',
    });
    diagnostics.assertEmpty();
  } finally {
    diagnostics.stop();
  }
});

test('structural plans preserve uncertainty for inferred appearance and departure', async ({ page }) => {
  await page.emulateMedia({ reducedMotion: 'no-preference' });
  const diagnostics = await prepareRendererFixturePage(page);
  try {
    const result = await page.evaluate(async () => {
      await import('/src/components/boardgame-component-animator.ts');
      await import('/src/components/boardgame-component-stack.ts');
      const { cardView } = await import('/src/client.ts');

      type Plan = {
        segments: Array<{
          subjectId: string;
          presence: string;
          provenance: {
            kind: string;
            endpoint?: string;
            stackId?: string;
            evidence?: string;
          };
        }>;
      };
      const animator = document.createElement('boardgame-component-animator') as HTMLElement & {
        updateComplete: Promise<unknown>;
        prepare(): void;
        animateFlip(): Promise<void>;
        clearAnimatingComponents(): void;
        _solvedMotionPlan: Plan | null;
      };
      const makeStackElement = (left: number) => {
        const stack = document.createElement('boardgame-component-stack') as HTMLElement & {
          readonly id: string;
          stack: unknown;
          componentView: unknown;
          updateComplete: Promise<unknown>;
        };
        Object.assign(stack.style, {
          position: 'fixed', left: `${left}px`, top: '100px', width: '120px',
        });
        stack.style.setProperty('--animation-length', '60ms');
        stack.componentView = cardView({});
        return stack;
      };
      const source = makeStackElement(20);
      const destination = makeStackElement(260);
      const visible = {
        Index: 0,
        Values: { rank: 'Q' },
        Deck: 'cards',
        GameName: 'motion-provenance-test',
        ID: 'inferred-card',
      };
      const stackData = (
        components: readonly unknown[],
        ids: readonly string[],
        idsLastSeen: Readonly<Record<string, number>>,
      ) => ({
        Deck: 'cards',
        Indexes: components.map((_item, index) => index),
        IDs: ids,
        IDsLastSeen: idsLastSeen,
        ShuffleCount: 0,
        Size: components.length,
        GameName: 'motion-provenance-test',
        Components: components,
      });
      source.stack = stackData([], [], {});
      destination.stack = stackData([], [], {});
      document.body.append(animator, source, destination);
      await Promise.all([
        animator.updateComplete, source.updateComplete, destination.updateComplete,
      ]);

      // The card materializes without an exact prior host. Stack history says
      // destination is the latest sighting and source is the runner-up origin.
      animator.prepare();
      source.stack = stackData([], [], { 'inferred-card': 1 });
      destination.stack = stackData([visible], ['inferred-card'], { 'inferred-card': 2 });
      await Promise.all([source.updateComplete, destination.updateComplete]);
      await animator.animateFlip();
      const appearing = animator._solvedMotionPlan?.segments[0];

      // The real host disappears. The latest stack-history sighting supplies
      // only an inferred destination for the faux departing component.
      animator.prepare();
      source.stack = stackData([], [], { 'inferred-card': 4 });
      destination.stack = stackData([], [], { 'inferred-card': 3 });
      await Promise.all([source.updateComplete, destination.updateComplete]);
      await animator.animateFlip();
      const departing = animator._solvedMotionPlan?.segments[0];
      animator.clearAnimatingComponents();

      return {
        sourceId: source.id,
        appearing,
        departing,
      };
    });

    expect(result.appearing).toMatchObject({
      subjectId: 'inferred-card',
      presence: 'appearing',
      provenance: {
        kind: 'stack-history',
        endpoint: 'source',
        stackId: result.sourceId,
        evidence: 'runner-up',
      },
    });
    expect(result.departing).toMatchObject({
      subjectId: 'inferred-card',
      presence: 'departing',
      provenance: {
        kind: 'stack-history',
        endpoint: 'destination',
        stackId: result.sourceId,
        evidence: 'latest-seen',
      },
    });
    diagnostics.assertEmpty();
  } finally {
    diagnostics.stop();
  }
});
