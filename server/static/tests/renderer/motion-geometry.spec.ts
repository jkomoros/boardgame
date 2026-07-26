import { expect, test } from '@playwright/test';
import { prepareRendererFixturePage } from './renderer-fixture-helpers.js';

test('card face motion is a planned component-owned visual track', async ({ page }) => {
  await page.emulateMedia({ reducedMotion: 'no-preference' });
  const diagnostics = await prepareRendererFixturePage(page);
  try {
    const result = await page.evaluate(async () => {
      await import('/src/components/boardgame-card.ts');
      const card = document.createElement('boardgame-card') as HTMLElement & {
        updateComplete: Promise<unknown>;
        planMotionTracks(record: object): readonly object[];
        playAnimation(record: object): readonly Animation[];
      };
      card.style.setProperty('--animation-length', '80ms');
      document.body.append(card);
      await card.updateComplete;
      const record = {
        before: { faceUp: false, rotated: false },
        after: { faceUp: true, rotated: false },
        invertedTransform: 'none',
        finalTransform: '',
        beforeOpacity: '1',
        finalOpacity: '',
        needsHostTransition: false,
      };
      const tracks = card.planMotionTracks(record) as Array<{
        target: string;
        property: string;
        timeline: string;
        samples: Array<{ offset: number; value: string }>;
      }>;
      const animations = card.playAnimation({ ...record, tracks });
      const inner = card.shadowRoot?.querySelector<HTMLElement>('#inner');
      const animation = animations[0];
      const frames = animation?.effect instanceof KeyframeEffect
        ? animation.effect.getKeyframes()
        : [];
      const during = {
        tracks,
        count: animations.length,
        targetIsVisual: animation?.effect instanceof KeyframeEffect
          && animation.effect.target === inner,
        from: frames[0]?.transform,
        to: frames.at(-1)?.transform,
      };
      await Promise.all(animations.map(item => item.finished));
      return during;
    });

    expect(result).toEqual({
      tracks: [{
        target: 'visual',
        property: 'transform',
        timeline: 'eased',
        samples: [
          {
            offset: 0,
            value: 'scale(var(--component-effective-scale)) rotateY(0deg) rotate(0deg)',
          },
          {
            offset: 1,
            value: 'scale(var(--component-effective-scale)) rotateY(180deg) rotate(0deg)',
          },
        ],
      }],
      count: 1,
      targetIsVisual: true,
      from: 'scale(var(--component-effective-scale)) rotateY(0deg) rotate(0deg)',
      to: 'scale(var(--component-effective-scale)) rotateY(180deg) rotate(0deg)',
    });
    diagnostics.assertEmpty();
  } finally {
    diagnostics.stop();
  }
});

test('an unavailable visual target does not suppress valid host motion', async ({ page }) => {
  await page.emulateMedia({ reducedMotion: 'no-preference' });
  const diagnostics = await prepareRendererFixturePage(page);
  try {
    const result = await page.evaluate(async () => {
      const { BoardgameComponent } = await import('/src/components/boardgame-component.ts');

      class MissingVisualTarget extends BoardgameComponent {
        protected override propertyMotionTracks() {
          return [{
            target: 'visual' as const,
            property: 'transform' as const,
            from: 'rotate(0deg)',
            to: 'rotate(90deg)',
          }];
        }

        protected override motionTrackTarget(target: 'host' | 'visual'): HTMLElement | null {
          return target === 'host' ? this : null;
        }
      }
      customElements.define('missing-visual-target', MissingVisualTarget);
      const component = document.createElement('missing-visual-target') as MissingVisualTarget;
      component.style.setProperty('--animation-length', '80ms');
      document.body.append(component);
      await component.updateComplete;
      const record = {
        before: {},
        after: {},
        invertedTransform: 'translateX(-40px)',
        finalTransform: '',
        beforeOpacity: '1',
        finalOpacity: '',
        needsHostTransition: true,
      };
      const tracks = component.planMotionTracks(record);
      const animations = component.playAnimation({ ...record, tracks });
      const result = {
        channels: tracks.map(track => `${track.target}:${track.property}`),
        returnedAnimations: animations.length,
        hostAnimations: component.getAnimations().length,
      };
      for (const animation of animations) animation.finish();
      return result;
    });

    expect(result).toEqual({
      channels: ['host:transform'],
      returnedAnimations: 1,
      hostAnimations: 1,
    });
    diagnostics.assertEmpty();
  } finally {
    diagnostics.stop();
  }
});

test('a visual target disappearing at the playback barrier still preserves host travel', async ({ page }) => {
  await page.emulateMedia({ reducedMotion: 'no-preference' });
  const diagnostics = await prepareRendererFixturePage(page);
  try {
    const result = await page.evaluate(async () => {
      const { BoardgameComponent } = await import('/src/components/boardgame-component.ts');
      await import('/src/components/boardgame-component-animator.ts');
      await import('/src/components/boardgame-component-stack.ts');
      const { componentView } = await import('/src/components/component-view.ts');
      class VanishingVisual extends BoardgameComponent {
        turned = false;
        visualResolutions = 0;
        override get animatingProperties() { return ['turned']; }
        protected override propertyMotionTracks(before: any, after: any) {
          return before.turned === after.turned ? [] : [{
            target: 'visual' as const, property: 'transform' as const,
            from: 'rotate(0deg)', to: 'rotate(90deg)',
          }];
        }
        protected override motionTrackTarget(target: 'host' | 'visual'): HTMLElement | null {
          if (target === 'host') return this;
          return this.visualResolutions++ === 0 ? this : null;
        }
      }
      customElements.define('vanishing-motion-visual', VanishingVisual);
      const animator = document.createElement('boardgame-component-animator') as any;
      const stack = document.createElement('boardgame-component-stack') as any;
      Object.assign(stack.style, { position: 'fixed', left: '20px', top: '20px' });
      stack.style.setProperty('--animation-length', '20ms');
      stack.componentView = componentView(
        () => document.createElement('vanishing-motion-visual'),
        { properties: ({ component }: any) => ({ turned: !!component.Values.turned }) },
      );
      const data = (turned: boolean) => ({
        Deck: 'pieces', Indexes: [0], IDs: ['vanishing-visual'], IDsLastSeen: {},
        ShuffleCount: 0, Size: 1, GameName: 'target-race', Components: [{
          ID: 'vanishing-visual', Index: 0, Deck: 'pieces', GameName: 'target-race',
          Values: { turned },
        }],
      });
      stack.stack = data(false);
      document.body.append(animator, stack);
      await Promise.all([animator.updateComplete, stack.updateComplete]);
      animator.prepare();
      stack.style.left = '220px';
      stack.stack = data(true);
      await stack.updateComplete;
      await animator.animateFlip();
      const segment = animator._solvedMotionPlan?.segments[0];
      return {
        channels: segment?.channels,
        status: segment?.execution.status,
        resolutions: stack.Components[0].visualResolutions,
      };
    });

    expect(result).toEqual({
      channels: ['host:transform'],
      status: 'finished',
      resolutions: 2,
    });
    diagnostics.assertEmpty();
  } finally {
    diagnostics.stop();
  }
});

test('retained travel preserves the historical final-opacity snap', async ({ page }) => {
  await page.emulateMedia({ reducedMotion: 'no-preference' });
  const diagnostics = await prepareRendererFixturePage(page);
  try {
    const result = await page.evaluate(async () => {
      await import('/src/components/boardgame-component-animator.ts');
      await import('/src/components/boardgame-component-stack.ts');
      const { cardView } = await import('/src/client.ts');
      const animator = document.createElement('boardgame-component-animator') as any;
      const stack = document.createElement('boardgame-component-stack') as any;
      Object.assign(stack.style, { position: 'fixed', left: '20px', top: '20px' });
      stack.style.setProperty('--animation-length', '20ms');
      stack.componentView = cardView({});
      stack.stack = {
        Deck: 'cards', Indexes: [0], IDs: ['opacity-card'], IDsLastSeen: {},
        ShuffleCount: 0, Size: 1, GameName: 'opacity-compatibility',
        Components: [{
          ID: 'opacity-card', Index: 0, Deck: 'cards', GameName: 'opacity-compatibility',
          Values: {},
        }],
      };
      document.body.append(animator, stack);
      await Promise.all([animator.updateComplete, stack.updateComplete]);
      const card = stack.Components[0] as HTMLElement;
      card.style.opacity = '0.2';
      const prototype = customElements.get('boardgame-card')!.prototype as any;
      const original = prototype.playAnimation;
      let observed: any = null;
      prototype.playAnimation = function(record: any) {
        if (this.id === 'opacity-card') {
          observed = {
            beforeOpacity: record.beforeOpacity,
            finalOpacity: record.finalOpacity,
            channels: record.tracks.map((track: any) => `${track.target}:${track.property}`),
          };
        }
        return original.call(this, record);
      };
      try {
        animator.prepare();
        stack.style.left = '220px';
        card.style.opacity = '0.8';
        await new Promise(requestAnimationFrame);
        await animator.animateFlip();
        return observed;
      } finally {
        prototype.playAnimation = original;
      }
    });

    expect(result).toEqual({
      beforeOpacity: '0.8',
      finalOpacity: '0.8',
      channels: ['host:transform'],
    });
    diagnostics.assertEmpty();
  } finally {
    diagnostics.stop();
  }
});

test('historical presentation preserves legacy artwork identity with an opt-in safe mode', async ({ page }) => {
  const diagnostics = await prepareRendererFixturePage(page);
  try {
    const result = await page.evaluate(async () => {
      await import('/src/components/boardgame-card.ts');
      await import('/src/components/boardgame-token.ts');
      const {
        captureHistoricalPresentation,
        installHistoricalPresentation,
      } = await import('/src/motion/historical-presentation.ts');
      const source = document.createElement('boardgame-card');
      source.noContent = false;
      source.append(document.createTextNode('bare text must remain absent'));
      const visible = document.createElement('span');
      visible.id = 'must-not-duplicate';
      visible.tabIndex = 0;
      visible.textContent = 'visible historical art';
      const named = document.createElement('span');
      named.slot = 'back';
      named.textContent = 'named private content';
      const svg = document.createElementNS('http://www.w3.org/2000/svg', 'svg');
      svg.append(document.createElementNS('http://www.w3.org/2000/svg', 'circle'));
      const namedSvg = document.createElementNS('http://www.w3.org/2000/svg', 'svg');
      namedSvg.setAttribute('slot', 'back');
      namedSvg.append(document.createElementNS('http://www.w3.org/2000/svg', 'rect'));
      source.append(visible, named, svg, namedSvg);
      const presentation = captureHistoricalPresentation(source)!;
      const target = document.createElement('boardgame-card');
      const safeSource = source.cloneNode(true) as HTMLElement & {
        historicalPresentationPolicy: string;
      };
      Object.defineProperty(safeSource, 'historicalPresentationPolicy', {
        value: 'clone-default-slot-safe',
      });
      const safeTarget = document.createElement('boardgame-card');
      const safePresentation = captureHistoricalPresentation(safeSource)!;
      installHistoricalPresentation(safeTarget, safePresentation);
      const mismatch = document.createElement('boardgame-token');
      const safeMismatch = document.createElement('boardgame-token');
      const installed = installHistoricalPresentation(target, presentation);
      return {
        installed,
        mismatch: installHistoricalPresentation(mismatch, presentation),
        safeMismatch: installHistoricalPresentation(safeMismatch, safePresentation),
        forged: installHistoricalPresentation(target, {
          kind: 'cloned-default-slot', identity: 'preserve',
        }),
        serialized: JSON.stringify(presentation),
        text: target.textContent,
        historyCount: target.querySelectorAll('[slot="motion-history"]').length,
        svgHistoryCount: target.querySelectorAll('svg[slot="motion-history"]').length,
        fallbackCount: target.querySelectorAll('[slot="fallback"]').length,
        svgFallbackCount: target.querySelectorAll('svg[slot="fallback"]').length,
        safeHistoryCount: safeTarget.querySelectorAll('[slot="motion-history"]').length,
        identityCount: target.querySelectorAll('[id], [tabindex], [autofocus]').length,
        safeIdentityCount: safeTarget.querySelectorAll('[id], [tabindex], [autofocus]').length,
      };
    });
    expect(result).toEqual({
      installed: true,
      mismatch: true,
      safeMismatch: false,
      forged: false,
      serialized: '{"kind":"cloned-default-slot","identity":"preserve"}',
      text: 'visible historical art',
      historyCount: 0,
      svgHistoryCount: 0,
      fallbackCount: 2,
      svgFallbackCount: 1,
      safeHistoryCount: 2,
      identityCount: 1,
      safeIdentityCount: 0,
    });
    diagnostics.assertEmpty();
  } finally {
    diagnostics.stop();
  }
});

test('a throwing component planner degrades to framework-owned structural channels', async ({ page }) => {
  await page.emulateMedia({ reducedMotion: 'no-preference' });
  const diagnostics = await prepareRendererFixturePage(page);
  try {
    const result = await page.evaluate(async () => {
      await import('/src/components/boardgame-component-animator.ts');
      const animator = document.createElement('boardgame-component-animator') as HTMLElement & {
        updateComplete: Promise<unknown>;
        _planMotionTracks(component: object, input: object): ReadonlyArray<{
          target: string;
          property: string;
          timeline: string;
          samples: ReadonlyArray<{ offset: number; value: string }>;
        }>;
      };
      document.body.append(animator);
      await animator.updateComplete;

      const reported: string[] = [];
      const originalError = console.error;
      console.error = (...values: unknown[]) => { reported.push(values.map(String).join(' ')); };
      try {
        const tracks = animator._planMotionTracks({
          planMotionTracks() {
            throw new Error('fixture visual planner failed');
          },
        }, {
          before: {},
          after: {},
          needsHostTransition: true,
          invertedTransform: 'translateX(-20px)',
          finalTransform: 'none',
          beforeOpacity: '0.5',
          finalOpacity: '1',
          visualTracks: [{
            target: 'visual',
            property: 'transform',
            from: 'rotate(0deg)',
            to: 'rotate(90deg)',
          }],
        });
        return {
          reported: reported.some(message => message.includes('fixture visual planner failed')),
          channels: tracks.map(track => `${track.target}:${track.property}`),
        };
      } finally {
        console.error = originalError;
      }
    });

    expect(result).toEqual({
      reported: true,
      channels: ['host:transform', 'host:opacity'],
    });
    diagnostics.assertEmpty();
  } finally {
    diagnostics.stop();
  }
});

test('standalone die spin uses the shared visual-track executor', async ({ page }) => {
  await page.emulateMedia({ reducedMotion: 'no-preference' });
  const diagnostics = await prepareRendererFixturePage(page);
  try {
    const result = await page.evaluate(async () => {
      await import('/src/components/boardgame-die.ts');
      const die = document.createElement('boardgame-die') as HTMLElement & {
        faces: number[];
        selectedFace: number;
        updateComplete: Promise<unknown>;
      };
      die.faces = [1, 2, 3, 4, 5, 6];
      die.style.setProperty('--animation-length', '80ms');
      document.body.append(die);
      await die.updateComplete;
      die.selectedFace = 4;
      await die.updateComplete;
      const inner = die.shadowRoot?.querySelector<HTMLElement>('#inner');
      const animations = inner?.getAnimations() ?? [];
      const animation = animations[0];
      const frames = animation?.effect instanceof KeyframeEffect
        ? animation.effect.getKeyframes()
        : [];
      const during = {
        count: animations.length,
        from: frames[0]?.transform,
        to: frames.at(-1)?.transform,
      };
      await Promise.all(animations.map(item => item.finished));
      return during;
    });

    // --reel-step is one die-size per reel face, and zero on a solid (which
    // has no reel to scroll): a variable of its own so that neutralizing the
    // reel scroll cannot zero anything else measured against the die's size.
    expect(result).toEqual({
      count: 1,
      from: 'translateY(calc(-1 * var(--reel-step) * 0))',
      to: 'translateY(calc(-1 * var(--reel-step) * 4))',
    });
    diagnostics.assertEmpty();
  } finally {
    diagnostics.stop();
  }
});

test('animateBetween aligns differently-sized endpoints by viewport center', async ({ page }) => {
  await page.emulateMedia({ reducedMotion: 'no-preference' });
  const diagnostics = await prepareRendererFixturePage(page);
  try {
    const result = await page.evaluate(async () => {
      await import('/src/components/boardgame-component-animator.ts');

      const animator = document.createElement('boardgame-component-animator') as HTMLElement & {
        updateComplete: Promise<unknown>;
        _lastExplicitMotionPlan: null | {
          source: string;
          phase: string;
          segments: Array<{
            execution: { status: string };
          viewport?: {
            from: { space: string; left: number; top: number };
            to: { space: string; left: number; top: number };
            };
          }>;
        };
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
      await Promise.resolve();
      const plan = animator._lastExplicitMotionPlan;
      await animator.animateBetween(real, 'missing-explicit-source', 500, { timing: 'immediate' });
      return {
        keyframes: result,
        plan,
        unresolved: animator._lastExplicitMotionPlan,
      };
    });

    // real center = (210, 105), stub center = (40, 45).
    expect(result.keyframes[0]).toBe('translate(-170px, -60px)');
    expect(result.keyframes[1]).toBe('none');
    // Compatibility calls do not claim declarative lifecycle ownership.
    expect(result.plan).toBeNull();
    expect(result.unresolved).toBeNull();
    diagnostics.assertEmpty();
  } finally {
    diagnostics.stop();
  }
});

test('fly names carrier direction and preserves its computed resting transform', async ({ page }) => {
  await page.emulateMedia({ reducedMotion: 'no-preference' });
  const diagnostics = await prepareRendererFixturePage(page);
  try {
    const result = await page.evaluate(async () => {
      await import('/src/components/boardgame-component-animator.ts');
      const animator = document.createElement('boardgame-component-animator') as HTMLElement & {
        fly(request: {
          subjectId: string;
          source: string;
          carrier: string;
          durationMs: number;
          timing: 'immediate';
        }): Promise<void>;
      };
      const style = document.createElement('style');
      style.textContent = '#flight-carrier { transform: scale(0.8); }';
      const source = document.createElement('div');
      source.id = 'flight-source';
      const carrier = document.createElement('div');
      carrier.id = 'flight-carrier';
      Object.assign(source.style, {
        position: 'fixed', left: '10px', top: '20px', width: '20px', height: '20px',
      });
      Object.assign(carrier.style, {
        position: 'fixed', left: '110px', top: '220px', width: '20px', height: '20px',
      });
      document.head.append(style);
      document.body.append(animator, source, carrier);
      let passedFrames: Keyframe[] = [];
      const nativeAnimate = carrier.animate.bind(carrier);
      carrier.animate = ((frames: Keyframe[] | PropertyIndexedKeyframes, timing?: number | KeyframeAnimationOptions) => {
        passedFrames = frames as Keyframe[];
        const animation = nativeAnimate(frames, timing);
        queueMicrotask(() => animation.finish());
        return animation;
      }) as typeof carrier.animate;
      // String endpoints deliberately exercise the registered-update barrier;
      // callers do not count animation frames before requesting the flight.
      await animator.fly({
        subjectId: 'semantic-card',
        source: source.id,
        carrier: carrier.id,
        durationMs: 10_000,
        timing: 'immediate',
      });
      return {
        from: passedFrames[0]?.transform,
        to: passedFrames[1]?.transform,
        resting: getComputedStyle(carrier).transform,
      };
    });

    expect(result).toEqual({
      from: 'translate(-100px, -200px) matrix(0.8, 0, 0, 0.8, 0, 0)',
      to: 'matrix(0.8, 0, 0, 0.8, 0, 0)',
      resting: 'matrix(0.8, 0, 0, 0.8, 0, 0)',
    });
    diagnostics.assertEmpty();
  } finally {
    diagnostics.stop();
  }
});

test('a newer flight owns one carrier without replaying older generation events', async ({ page }) => {
  await page.emulateMedia({ reducedMotion: 'no-preference' });
  const diagnostics = await prepareRendererFixturePage(page);
  try {
    const result = await page.evaluate(async () => {
      await import('/src/components/boardgame-component-animator.ts');
      const animator = document.createElement('boardgame-component-animator') as HTMLElement & {
        fly(request: {
          subjectId: string;
          source: HTMLElement;
          carrier: HTMLElement;
          durationMs: number;
          timing: 'immediate';
        }): Promise<void>;
        observeStructuralMotionEvents(observer: (event: {
          id: string;
          source: string;
          generation: number;
          kind: string;
        }) => void): () => void;
      };
      const sourceA = document.createElement('div');
      const sourceB = document.createElement('div');
      const carrier = document.createElement('div');
      Object.assign(sourceA.style, { position: 'fixed', left: '0px', top: '0px', width: '10px', height: '10px' });
      Object.assign(sourceB.style, { position: 'fixed', left: '40px', top: '0px', width: '10px', height: '10px' });
      Object.assign(carrier.style, { position: 'fixed', left: '100px', top: '100px', width: '10px', height: '10px' });
      document.body.append(animator, sourceA, sourceB, carrier);
      const first = animator.fly({
        subjectId: 'first', source: sourceA, carrier, durationMs: 10_000, timing: 'immediate',
      });
      const second = animator.fly({
        subjectId: 'second', source: sourceB, carrier, durationMs: 10_000, timing: 'immediate',
      });
      const active = carrier.getAnimations().at(-1);
      active?.finish();
      await Promise.all([first, second]);
      await Promise.resolve();
      const events: Array<{ id: string; generation: number; kind: string }> = [];
      const stop = animator.observeStructuralMotionEvents(event => {
        if (event.source === 'explicit') events.push(event);
      });
      stop();
      return events;
    });

    expect(new Set(result.map(event => event.id)).size).toBe(result.length);
    const firstKinds = result.filter(event => event.generation === 1).map(event => event.kind);
    expect(firstKinds.slice(0, 2)).toEqual(['planned', 'armed']);
    expect(firstKinds.slice(-2)).toEqual(['cancelled', 'generation-settled']);
    expect(result.filter(event => event.generation === 2).map(event => event.kind)).toEqual([
      'planned', 'armed', 'active-observed', 'finished', 'generation-settled',
    ]);
    diagnostics.assertEmpty();
  } finally {
    diagnostics.stop();
  }
});

test('declarative transfers publish and settle as one scoped explicit batch', async ({ page }) => {
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
          generation: number;
          phase: string;
          segments: Array<{
            declarationKey?: string;
            execution: { status: string; reason?: string };
          }>;
        };
      };
      root.append(animator);
      document.body.append(host);
      const source = root.querySelector<HTMLElement>('#transfer-source')!;
      const carrierA = root.querySelector<HTMLElement>('#transfer-carrier-a')!;
      const carrierB = root.querySelector<HTMLElement>('#transfer-carrier-b')!;
      Object.assign(source.style, { position: 'fixed', left: '0px', top: '0px', width: '10px', height: '10px' });
      Object.assign(carrierA.style, { position: 'fixed', left: '100px', top: '0px', width: '10px', height: '10px' });
      Object.assign(carrierB.style, { position: 'fixed', left: '200px', top: '0px', width: '10px', height: '10px' });
      await Promise.all([
        animator.updateComplete,
        (root.querySelector('#registry') as any).updateComplete,
      ]);
      for (const carrier of [carrierA, carrierB]) {
        const nativeAnimate = carrier.animate.bind(carrier);
        carrier.animate = ((frames: Keyframe[] | PropertyIndexedKeyframes, timing?: number | KeyframeAnimationOptions) => {
          const animation = nativeAnimate(frames, timing);
          queueMicrotask(() => animation.finish());
          return animation;
        }) as typeof carrier.animate;
      }
      animator.prepare();
      animator.installMotionTransfers(compileMotionTransferDeclarations([
        { key: 'deal:0', subjectId: 'opaque-0', source: 'transfer-source', carrier: 'transfer-carrier-a', durationMs: 10_000, timing: 'immediate' },
        { key: 'deal:1', subjectId: 'opaque-1', source: 'transfer-source', carrier: 'transfer-carrier-b', durationMs: 10_000, timing: 'immediate' },
        { key: 'deal:2', subjectId: 'opaque-2', source: 'transfer-source', carrier: 'missing-carrier', durationMs: 10_000, timing: 'immediate' },
      ]));
      await animator.animateFlip();
      await Promise.resolve();
      return animator._lastExplicitMotionPlan;
    });

    expect(result).toMatchObject({
      generation: 1,
      phase: 'settled',
      segments: [
        { declarationKey: 'deal:0', execution: { status: 'finished' } },
        { declarationKey: 'deal:1', execution: { status: 'finished' } },
        { declarationKey: 'deal:2', execution: { status: 'skipped', reason: 'missing-endpoint' } },
      ],
    });
    diagnostics.assertEmpty();
  } finally {
    diagnostics.stop();
  }
});

test('explicit motion cannot override reduced-motion scheduling', async ({ page }) => {
  await page.emulateMedia({ reducedMotion: 'reduce' });
  const diagnostics = await prepareRendererFixturePage(page);
  try {
    const result = await page.evaluate(async () => {
      await import('/src/components/boardgame-component-animator.ts');
      const animator = document.createElement('boardgame-component-animator') as HTMLElement & {
        updateComplete: Promise<unknown>;
        animationContext: object | null;
        fly(request: {
          subjectId: string;
          source: HTMLElement;
          carrier: HTMLElement;
          durationMs: number;
        }): Promise<void>;
        observeStructuralMotionEvents(observer: (event: { source: string; kind: string }) => void): () => void;
        _lastExplicitMotionPlan: null | {
          segments: Array<{
            execution: {
              status: string;
              animations?: Array<{ delayMs: number; durationMs: number }>;
            };
          }>;
        };
      };
      const real = document.createElement('div');
      const stub = document.createElement('div');
      Object.assign(real.style, { position: 'fixed', left: '200px', top: '100px', width: '20px', height: '20px' });
      Object.assign(stub.style, { position: 'fixed', left: '20px', top: '30px', width: '20px', height: '20px' });
      document.body.append(animator, real, stub);
      await animator.updateComplete;
      animator.animationContext = {
        version: 4,
        startAtMs: Date.now() + 1000,
        slotDurationMs: 1200,
        maxAnimationDurationMs: 1000,
      };
      const lifecycle: string[] = [];
      const unobserve = animator.observeStructuralMotionEvents(event => {
        if (event.source === 'explicit') lifecycle.push(event.kind);
      });
      const startedAt = performance.now();
      await animator.fly({
        subjectId: 'reduced-explicit', source: stub, carrier: real, durationMs: 900,
      });
      unobserve();
      return {
        elapsedMs: performance.now() - startedAt,
        execution: animator._lastExplicitMotionPlan?.segments[0]?.execution,
        lifecycle,
      };
    });

    expect(result.elapsedMs).toBeLessThan(250);
    expect(result.execution).toEqual({
      status: 'finished',
      animations: [expect.objectContaining({ delayMs: 0, durationMs: 0 })],
    });
    expect(result.lifecycle).toEqual([
      'planned', 'armed', 'active-observed', 'finished', 'generation-settled',
    ]);
    diagnostics.assertEmpty();
  } finally {
    diagnostics.stop();
  }
});

test('reduced motion removes movement but preserves semantic post-animation holds', async ({ page }) => {
  await page.emulateMedia({ reducedMotion: 'reduce' });
  const diagnostics = await prepareRendererFixturePage(page);
  try {
    const result = await page.evaluate(async () => {
      const { BoardgameAnimatableItem } = await import('/src/components/boardgame-animatable-item.ts');
      class SemanticHoldItem extends BoardgameAnimatableItem {}
      customElements.define('semantic-hold-item', SemanticHoldItem);
      const item = document.createElement('semantic-hold-item') as SemanticHoldItem;
      item.postAnimationDelay = 1000;
      item.waitForAnimation = true;
      let expectedSettleMs: number | null = null;
      item.addEventListener('will-animate', (event: Event) => {
        expectedSettleMs = (event as CustomEvent).detail.expectedSettleMs;
      });
      document.body.append(item);
      await item.updateComplete;
      const animation = item.play(
        item,
        [{ transform: 'translateX(40px)' }, { transform: 'translateX(0)' }],
        { duration: 500 },
        { timing: 'immediate' },
      )!;
      const timing = (animation.effect as KeyframeEffect).getTiming();
      const observed = {
        duration: timing.duration,
        endDelay: timing.endDelay,
        expectedSettleMs,
      };
      animation.finish();
      await animation.finished;
      return observed;
    });

    expect(result).toEqual({ duration: 0, endDelay: 1000, expectedSettleMs: 1000 });
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
        installMotionTransfers(declarations: readonly unknown[]): void;
        animateFlip(): Promise<void>;
        observeStructuralMotionEvents(observer: (event: {
          id: string;
          generation: number;
          kind: string;
          subjectId?: string;
        }) => void): () => void;
        _solvedMotionPlan: null | {
          generation: number;
          phase: string;
          segments: Array<{
            subjectId: string;
            presence: string;
            provenance: { kind: string };
            channels: string[];
            path?: {
              kind: string;
              from: { space: string; left: number; top: number };
              to: { space: string; left: number; top: number };
            };
            timingRequest: { policy: string; delayMs: number; durationMs: number };
            execution: {
              status: string;
              animations?: Array<{ durationMs: number; fill: string }>;
            };
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
      const observedEvents: Array<{
        id: string;
        generation: number;
        kind: string;
        subjectId: string;
      }> = [];
      const unobserveEvents = animator.observeStructuralMotionEvents(event => {
        observedEvents.push({
          id: event.id,
          generation: event.generation,
          kind: event.kind,
          subjectId: event.subjectId ?? '',
        });
      });

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
      const replayedKinds: string[] = [];
      const unobserveReplay = animator.observeStructuralMotionEvents(event => {
        if (event.generation === second.generation && event.id.startsWith('flip:')) {
          replayedKinds.push(event.kind);
        }
      });
      unobserveReplay();
      unobserveEvents();

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
          phase: second.phase,
          segment: second.segments[0],
        },
        observedEvents,
        replayedKinds,
      };
    });

    expect(result.nullDuringMeasurement).toBe(true);
    expect(result.playingAfterPublish).toBe(true);
    expect(result.invalidatedImmediately).toBe(true);
    expect(result.first.frozen).toBe(true);
    expect(result.first.phase).toBe('executing');
    expect(result.first.segment).toMatchObject({
      subjectId: 'card-plan',
      presence: 'retained',
      provenance: { kind: 'identity' },
      path: { kind: 'stationary' },
      channels: ['host:transform'],
      timingRequest: { policy: 'version', delayMs: 0, durationMs: 80 },
      execution: {
        status: 'active-observed',
        animations: [expect.objectContaining({ durationMs: 80 })],
      },
    });
    expect(result.second.generation).toBe(result.first.generation + 1);
    expect(result.second.phase).toBe('settled');
    expect(result.second.segment.execution.status).toBe('finished');
    expect(result.second.segment.path.kind).toBe('stationary');
    expect(result.observedEvents).toEqual([
      { id: 'flip:1:0:planned', generation: 1, kind: 'planned', subjectId: 'card-plan' },
      { id: 'flip:1:0:armed', generation: 1, kind: 'armed', subjectId: 'card-plan' },
      { id: 'flip:1:0:active-observed', generation: 1, kind: 'active-observed', subjectId: 'card-plan' },
      { id: 'flip:1:0:cancelled', generation: 1, kind: 'cancelled', subjectId: 'card-plan' },
      { id: 'flip:1:generation-settled', generation: 1, kind: 'generation-settled', subjectId: '' },
      { id: 'flip:2:0:planned', generation: 2, kind: 'planned', subjectId: 'card-plan' },
      { id: 'flip:2:0:armed', generation: 2, kind: 'armed', subjectId: 'card-plan' },
      { id: 'flip:2:0:active-observed', generation: 2, kind: 'active-observed', subjectId: 'card-plan' },
      { id: 'flip:2:0:finished', generation: 2, kind: 'finished', subjectId: 'card-plan' },
      { id: 'flip:2:generation-settled', generation: 2, kind: 'generation-settled', subjectId: '' },
    ]);
    expect(result.replayedKinds).toEqual([
      'planned', 'armed', 'active-observed', 'finished', 'generation-settled',
    ]);
    diagnostics.assertEmpty();
  } finally {
    diagnostics.stop();
  }
});

test('structural continuity preserves legacy ordered ties and supports declared history', async ({ page }) => {
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
          viewport?: {
            from: { space: string };
            to: { space: string };
          };
          execution: { status: string };
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
      const ambiguousDestination = makeStackElement(500);
      const visible = {
        Index: 0,
        Values: { rank: 'Q' },
        Deck: 'cards',
        GameName: 'motion-provenance-test',
        ID: 'inferred-card',
      };
      const anchor = document.createElement('div');
      anchor.id = 'declared-hand-edge';
      Object.assign(anchor.style, {
        position: 'fixed', left: '50px', top: '0px', width: '1px', height: '1px',
      });
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
      ambiguousDestination.stack = stackData([], [], {});
      document.body.append(animator, source, destination, ambiguousDestination, anchor);
      await Promise.all([
        animator.updateComplete, source.updateComplete, destination.updateComplete,
        ambiguousDestination.updateComplete,
      ]);

      // The card materializes without an exact prior host. Stack history says
      // destination is the latest sighting and source is the runner-up origin.
      animator.prepare();
      source.stack = stackData([], [], { 'inferred-card': 1 });
      destination.stack = stackData([visible], ['inferred-card'], { 'inferred-card': 2 });
      ambiguousDestination.stack = stackData([], [], {});
      await Promise.all([
        source.updateComplete, destination.updateComplete, ambiguousDestination.updateComplete,
      ]);
      await animator.animateFlip();
      const appearing = animator._solvedMotionPlan?.segments[0];

      // The real host disappears and equally recent stacks retain the old
      // registration-order winner/runner behavior for visual compatibility.
      animator.prepare();
      source.stack = stackData([], [], { 'inferred-card': 4 });
      destination.stack = stackData([], [], { 'inferred-card': 4 });
      ambiguousDestination.stack = stackData([], [], { 'inferred-card': 4 });
      await Promise.all([
        source.updateComplete, destination.updateComplete, ambiguousDestination.updateComplete,
      ]);
      await animator.animateFlip();
      const departing = animator._solvedMotionPlan?.segments[0];
      animator.clearAnimatingComponents();

      // A Hand-only surface has no rendered source stack or history. The
      // declaration establishes after-only presence and supplies geometry;
      // destination defaults supply the safe presentation pose.
      const declared = { ...visible, ID: 'declared-card' };
      animator.prepare();
      animator.installMotionTransfers([{
        key: 'auto-hand:0',
        subjectId: 'declared-card',
        source: anchor.id,
        carrier: 'declared-card',
        durationMs: 80,
        timing: 'immediate',
      }]);
      source.stack = stackData([], [], {});
      destination.stack = stackData([declared], ['declared-card'], {});
      ambiguousDestination.stack = stackData([], [], {});
      await Promise.all([
        source.updateComplete, destination.updateComplete, ambiguousDestination.updateComplete,
      ]);
      await animator.animateFlip();
      const declaredAppearing = animator._solvedMotionPlan?.segments[0];

      return {
        sourceId: source.id,
        appearing,
        departing,
        declaredAppearing,
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
      path: {
        kind: 'travel',
        from: { space: 'viewport' },
        to: { space: 'viewport' },
      },
      execution: { status: 'finished' },
    });
    expect(result.declaredAppearing).toMatchObject({
      subjectId: 'declared-card',
      presence: 'appearing',
      declarationKey: 'auto-hand:0',
      pathOrigin: 'declared-anchor',
      provenance: {
        kind: 'declaration',
        declarationKey: 'auto-hand:0',
        pose: 'destination-defaults',
      },
      path: { kind: 'travel', from: { space: 'viewport' }, to: { space: 'viewport' } },
      execution: { status: 'finished' },
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
      path: { kind: 'travel' },
      execution: { status: 'finished' },
    });
    diagnostics.assertEmpty();
  } finally {
    diagnostics.stop();
  }
});

test('departing motion uses a fresh inert carrier without publishing presentation', async ({ page }) => {
  await page.emulateMedia({ reducedMotion: 'no-preference' });
  const diagnostics = await prepareRendererFixturePage(page);
  try {
    const result = await page.evaluate(async () => {
      await import('/src/components/boardgame-component-animator.ts');
      await import('/src/components/boardgame-component-stack.ts');
      const { cardView } = await import('/src/client.ts');
      const animator = document.createElement('boardgame-component-animator') as any;
      const makeStack = () => {
        const stack = document.createElement('boardgame-component-stack') as any;
        stack.style.setProperty('--animation-length', '40ms');
        stack.componentView = cardView({});
        return stack;
      };
      const source = makeStack();
      const destination = makeStack();
      const cardPrototype = customElements.get('boardgame-card')!.prototype as any;
      const originalPlayAnimation = cardPrototype.playAnimation;
      let carrierAtPlayback: unknown = null;
      cardPrototype.playAnimation = function(record: any) {
        if (this.inert) {
          carrierAtPlayback = {
            inlineTransform: this.style.transform,
            inlineOpacity: this.style.opacity,
            finalTransform: record.finalTransform,
            finalOpacity: record.finalOpacity,
            beforeOpacity: record.beforeOpacity,
          };
        }
        return originalPlayAnimation.call(this, record);
      };
      const visible = {
        Index: 0, Values: { rank: 'A' }, Deck: 'cards', GameName: 'carrier-test', ID: 'carrier-card',
      };
      const data = (components: readonly unknown[], ids: readonly string[], lastSeen: Record<string, number>) => ({
        Deck: 'cards', Indexes: components.map((_item, index) => index), IDs: ids,
        IDsLastSeen: lastSeen, ShuffleCount: 0, Size: components.length,
        GameName: 'carrier-test', Components: components,
      });
      source.stack = data([visible], ['carrier-card'], { 'carrier-card': 1 });
      destination.stack = data([], [], {});
      document.body.append(animator, source, destination);
      await Promise.all([animator.updateComplete, source.updateComplete, destination.updateComplete]);
      const live = source.Components[0] as HTMLElement;
      live.style.opacity = '0.35';
      const art = document.createElement('span');
      art.id = 'visible-art-id';
      art.textContent = 'VISIBLE SOURCE ART';
      live.append(art);

      animator.prepare();
      source.stack = data([], [], { 'carrier-card': 1 });
      destination.stack = data([], [], { 'carrier-card': 2 });
      await Promise.all([source.updateComplete, destination.updateComplete]);
      await animator.animateFlip();
      const carrier = animator._animatingComponents[0]?.component as HTMLElement | undefined;
      const planJSON = JSON.stringify(animator._solvedMotionPlan);
      const observed = {
        segmentStatus: animator._solvedMotionPlan?.segments[0]?.execution.status,
        carrierTag: carrier?.localName,
        carrierId: carrier?.id,
        inert: carrier?.inert,
        ariaHidden: carrier?.getAttribute('aria-hidden'),
        pointerEvents: carrier?.style.pointerEvents,
        text: carrier?.textContent,
        duplicateIds: carrier?.querySelectorAll('[id]').length,
        presentationPublished: planJSON.includes('VISIBLE SOURCE ART') || planJSON.includes('motion-history'),
        carrierAtPlayback,
      };
      animator.clearAnimatingComponents();
      cardPrototype.playAnimation = originalPlayAnimation;
      return { ...observed, carriersAfterClear: destination.shadowRoot?.querySelectorAll('#animating-components > *').length };
    });
    expect(result).toEqual({
      segmentStatus: 'finished',
      carrierTag: 'boardgame-card',
      carrierId: '',
      inert: true,
      ariaHidden: 'true',
      pointerEvents: 'none',
      text: 'VISIBLE SOURCE ART',
      // Master-compatible artwork cloning preserves SVG/CSS identity refs.
      duplicateIds: 1,
      presentationPublished: false,
      carrierAtPlayback: {
        inlineTransform: '',
        inlineOpacity: '',
        finalTransform: 'scale(0.6)',
        finalOpacity: '0',
        beforeOpacity: '1.0',
      },
      carriersAfterClear: 0,
    });
    diagnostics.assertEmpty();
  } finally {
    diagnostics.stop();
  }
});

test('last visible card artwork survives an intervening hidden generation', async ({ page }) => {
  await page.emulateMedia({ reducedMotion: 'no-preference' });
  const diagnostics = await prepareRendererFixturePage(page);
  try {
    const result = await page.evaluate(async () => {
      await import('/src/components/boardgame-component-animator.ts');
      await import('/src/components/boardgame-component-stack.ts');
      const { cardView } = await import('/src/client.ts');
      const animator = document.createElement('boardgame-component-animator') as any;
      const makeStack = (left: number) => {
        const stack = document.createElement('boardgame-component-stack') as any;
        Object.assign(stack.style, { position: 'fixed', left: `${left}px`, top: '20px' });
        stack.style.setProperty('--animation-length', '20ms');
        stack.componentView = cardView({});
        return stack;
      };
      const source = makeStack(20);
      const destination = makeStack(220);
      const visible = {
        ID: 'remembered-card', Index: 0, Deck: 'cards', GameName: 'history-cache', Values: {},
      };
      const data = (
        components: readonly unknown[],
        ids: readonly string[],
        seen: number | null,
      ) => ({
        Deck: 'cards', Indexes: components.map((_item, index) => index), IDs: ids,
        IDsLastSeen: seen === null ? {} : { 'remembered-card': seen },
        ShuffleCount: 0, Size: components.length, GameName: 'history-cache',
        Components: components,
      });
      source.stack = data([visible], ['remembered-card'], 1);
      destination.stack = data([], [], null);
      document.body.append(animator, source, destination);
      await Promise.all([animator.updateComplete, source.updateComplete, destination.updateComplete]);
      const art = document.createElement('span');
      art.textContent = 'remembered visible face';
      source.Components[0].append(art);

      animator.prepare();
      source.stack = data([], [], 1);
      destination.stack = data([], [], 2);
      await Promise.all([source.updateComplete, destination.updateComplete]);
      await animator.animateFlip();
      animator.clearAnimatingComponents();

      // No exact card is visible during this prepare. Reappearance must use
      // the bounded viewer-local presentation cache, not transition-local DOM.
      animator.prepare();
      source.stack = data([], [], 2);
      destination.stack = data([visible], ['remembered-card'], 3);
      await Promise.all([source.updateComplete, destination.updateComplete]);
      await animator.animateFlip();
      const returned = destination.Components[0] as HTMLElement;
      return {
        text: returned.textContent,
        fallbackCount: returned.querySelectorAll('[slot="fallback"]').length,
        cacheSize: animator._historicalPresentationById.size,
      };
    });

    expect(result).toEqual({
      text: 'remembered visible face',
      fallbackCount: 1,
      cacheSize: 1,
    });
    diagnostics.assertEmpty();
  } finally {
    diagnostics.stop();
  }
});

test('structural outcomes report compiled version timing and exhausted stagger skips', async ({ page }) => {
  await page.emulateMedia({ reducedMotion: 'no-preference' });
  const diagnostics = await prepareRendererFixturePage(page);
  try {
    const result = await page.evaluate(async () => {
      await import('/src/components/boardgame-component-animator.ts');
      await import('/src/components/boardgame-component-stack.ts');
      const { cardView } = await import('/src/client.ts');

      const animator = document.createElement('boardgame-component-animator') as HTMLElement & {
        updateComplete: Promise<unknown>;
        animationContext: object;
        prepare(): void;
        animateFlip(): Promise<void>;
        _solvedMotionPlan: {
          phase: string;
          segments: Array<{
            subjectId: string;
            timingRequest: { delayMs: number; durationMs: number };
            execution: {
              status: string;
              reason?: string;
              animations?: Array<{ delayMs: number; durationMs: number; fill: string }>;
            };
          }>;
        } | null;
      };
      const stack = document.createElement('boardgame-component-stack') as HTMLElement & {
        stack: unknown;
        componentView: unknown;
        stagger: number;
        updateComplete: Promise<unknown>;
      };
      stack.style.setProperty('--animation-length', '80ms');
      stack.stagger = 1;
      stack.componentView = cardView({});
      const cards = ['timed-a', 'timed-b'].map((id, index) => ({
        Index: index,
        Values: { rank: String(index) },
        Deck: 'cards',
        GameName: 'motion-timing-test',
        ID: id,
      }));
      stack.stack = {
        Deck: 'cards', Indexes: [0, 1], IDs: ['timed-a', 'timed-b'],
        IDsLastSeen: {}, ShuffleCount: 0, Size: 2,
        GameName: 'motion-timing-test', Components: cards,
      };
      document.body.append(animator, stack);
      await Promise.all([animator.updateComplete, stack.updateComplete]);
      const elements = [...stack.querySelectorAll<HTMLElement>('boardgame-card')];
      await Promise.all(elements.map(element => (
        element as HTMLElement & { updateComplete: Promise<unknown> }
      ).updateComplete));
      animator.animationContext = {
        version: 21,
        startAtMs: Date.now() + 100,
        slotDurationMs: 200,
        maxAnimationDurationMs: 80,
      };

      animator.prepare();
      elements.forEach((element, index) => {
        element.style.transform = `translateX(${20 + index * 10}px)`;
      });
      await animator.animateFlip();
      return animator._solvedMotionPlan;
    });

    expect(result?.phase).toBe('settled');
    const first = result?.segments.find(segment => segment.subjectId === 'timed-a');
    const second = result?.segments.find(segment => segment.subjectId === 'timed-b');
    expect(first).toMatchObject({
      timingRequest: { delayMs: 0, durationMs: 80 },
      execution: {
        status: 'finished',
        animations: [expect.objectContaining({ durationMs: 80, fill: 'backwards' })],
      },
    });
    expect(first?.execution.animations?.[0].delayMs).toBeGreaterThan(0);
    expect(second).toMatchObject({
      timingRequest: { delayMs: 80, durationMs: 80 },
      execution: { status: 'skipped', reason: 'not-started' },
    });
    diagnostics.assertEmpty();
  } finally {
    diagnostics.stop();
  }
});

test('explicit motion cohorts schedule a deterministic order across stacks', async ({ page }) => {
  await page.emulateMedia({ reducedMotion: 'no-preference' });
  const diagnostics = await prepareRendererFixturePage(page);
  try {
    const result = await page.evaluate(async () => {
      await import('/src/components/boardgame-component-animator.ts');
      await import('/src/components/boardgame-component-stack.ts');
      const { cardView, motion } = await import('/src/client.ts');

      const animator = document.createElement('boardgame-component-animator') as HTMLElement & {
        updateComplete: Promise<unknown>;
        prepare(): void;
        installMotionCohorts(specs: unknown[]): void;
        animateFlip(): Promise<void>;
        observeStructuralMotionEvents(observer: (event: {
          generation: number; subjectId?: string; kind: string;
        }) => void): () => void;
        _solvedMotionPlan: {
          segments: Array<{
            subjectId: string;
            timingRequest: { delayMs: number };
            execution: { status: string };
            ref: { generation: number };
          }>;
        } | null;
      };
      const makeStack = (name: string, ids: string[]) => {
        const stack = document.createElement('boardgame-component-stack') as HTMLElement & {
          stack: unknown;
          componentView: unknown;
          updateComplete: Promise<unknown>;
        };
        stack.style.setProperty('--animation-length', '80ms');
        stack.componentView = cardView({});
        stack.stack = {
          Deck: 'cards',
          Indexes: ids.map((_id, index) => index),
          IDs: ids,
          IDsLastSeen: {},
          ShuffleCount: 0,
          Size: ids.length,
          GameName: `cohort-${name}`,
          Components: ids.map((id, index) => ({
            Index: index,
            Values: { rank: `${name}-${index}` },
            Deck: 'cards',
            GameName: `cohort-${name}`,
            ID: id,
          })),
        };
        return stack;
      };
      const left = makeStack('left', ['cohort-a', 'cohort-b']);
      const right = makeStack('right', ['cohort-c']);
      document.body.append(animator, left, right);
      await Promise.all([animator.updateComplete, left.updateComplete, right.updateComplete]);
      const elements = [...document.querySelectorAll<HTMLElement>('boardgame-card')];
      await Promise.all(elements.map(element => (
        element as HTMLElement & { updateComplete: Promise<unknown> }
      ).updateComplete));
      const lifecycle: Array<{ generation: number; subjectId: string; kind: string }> = [];
      const unobserve = animator.observeStructuralMotionEvents(event => lifecycle.push({
        generation: event.generation,
        subjectId: event.subjectId ?? '',
        kind: event.kind,
      }));

      animator.prepare();
      animator.installMotionCohorts([
        motion.stagger({
          subjects: ['cohort-c', 'cohort-a', 'cohort-b'],
          intervalMs: 25,
        }),
      ]);
      elements.forEach((element, index) => {
        element.style.transform = `translateX(${25 + index * 5}px)`;
      });
      await animator.animateFlip();
      const scheduled = animator._solvedMotionPlan?.segments.map(segment => ({
        subjectId: segment.subjectId,
        delayMs: segment.timingRequest.delayMs,
      }));

      // A declaration belongs only to the generation opened by prepare(). A
      // later prepare must discard it before the next playback plan exists.
      animator.prepare();
      animator.installMotionCohorts([
        motion.stagger({ subjects: ['cohort-a', 'cohort-b'], intervalMs: 70 }),
      ]);
      animator.prepare();
      elements.forEach((element, index) => {
        element.style.transform = `translateX(${70 + index * 5}px)`;
      });
      await animator.animateFlip();
      const afterInterruption = animator._solvedMotionPlan?.segments.map(segment => ({
        subjectId: segment.subjectId,
        delayMs: segment.timingRequest.delayMs,
      }));

      animator.prepare();
      animator.installMotionCohorts([
        motion.stagger({ subjects: ['cohort-a', 'cohort-b'], intervalMs: 300 }),
      ]);
      elements.forEach((element, index) => {
        element.style.transform = `translateX(${120 + index * 5}px)`;
      });
      const interrupted = animator.animateFlip();
      for (let frame = 0; frame < 20; frame++) {
        const segment = animator._solvedMotionPlan?.segments.find(
          candidate => candidate.subjectId === 'cohort-b',
        );
        if (segment?.execution.status === 'armed') break;
        await new Promise(requestAnimationFrame);
      }
      const interruptedGeneration = animator._solvedMotionPlan?.segments[0]?.ref.generation;
      animator.prepare();
      await interrupted;
      unobserve();
      const delayedCancellation = lifecycle.filter(event => (
        event.generation === interruptedGeneration && event.subjectId === 'cohort-b'
      )).map(event => event.kind);
      return { scheduled, afterInterruption, delayedCancellation };
    });

    expect(Object.fromEntries(result.scheduled?.map(entry => [entry.subjectId, entry.delayMs]) ?? [])).toEqual({
      'cohort-a': 25,
      'cohort-b': 50,
      'cohort-c': 0,
    });
    expect(result.afterInterruption?.every(entry => entry.delayMs === 0)).toBe(true);
    expect(result.delayedCancellation).toEqual(['planned', 'armed', 'cancelled']);
    diagnostics.assertEmpty();
  } finally {
    diagnostics.stop();
  }
});

test('buffered queue release follows every real staggered primary, not nominal wall time', async ({ page }) => {
  await page.emulateMedia({ reducedMotion: 'no-preference' });
  const diagnostics = await prepareRendererFixturePage(page);
  try {
    const result = await page.evaluate(async () => {
      await import('/src/components/boardgame-component-animator.ts');
      await import('/src/components/boardgame-component-stack.ts');
      const { cardView, motion } = await import('/src/client.ts');
      const animator = document.createElement('boardgame-component-animator') as HTMLElement & {
        updateComplete: Promise<unknown>;
        prepare(): void;
        installMotionCohorts(specs: unknown[]): void;
        installMotionRelease(spec: unknown, cycleId: number): void;
        animateFlip(): Promise<void>;
      };
      const stack = document.createElement('boardgame-component-stack') as HTMLElement & {
        stack: unknown;
        componentView: unknown;
        updateComplete: Promise<unknown>;
      };
      stack.style.setProperty('--animation-length', '500ms');
      stack.componentView = cardView({});
      stack.stack = {
        Deck: 'cards', Indexes: [0, 1], IDs: ['release-a', 'release-b'],
        IDsLastSeen: {}, ShuffleCount: 0, Size: 2, GameName: 'release',
        Components: ['release-a', 'release-b'].map((ID, Index) => ({
          ID, Index, Deck: 'cards', GameName: 'release', Values: { rank: Index },
        })),
      };
      document.body.append(animator, stack);
      await Promise.all([animator.updateComplete, stack.updateComplete]);
      const cards = [...document.querySelectorAll<HTMLElement>('boardgame-card')];
      await Promise.all(cards.map(card => (
        card as HTMLElement & { updateComplete: Promise<unknown> }
      ).updateComplete));
      animator.prepare();
      animator.installMotionCohorts([
        motion.stagger({ subjects: ['release-a', 'release-b'], intervalMs: 250 }),
      ]);
      animator.installMotionRelease(
        motion.release({ key: 'deal-cutover', progress: 0.25 }),
        77,
      );
      cards.forEach((card, index) => { card.style.transform = `translateX(${60 + index * 10}px)`; });
      let settled = false;
      const startedAt = performance.now();
      const released = new Promise<{ cycleId: number; key: string; elapsed: number; settled: boolean }>(resolve => {
        animator.addEventListener('motion-cycle-release', event => {
          const detail = (event as CustomEvent).detail;
          resolve({ ...detail, elapsed: performance.now() - startedAt, settled });
        }, { once: true });
      });
      const settlement = animator.animateFlip().then(() => { settled = true; });
      const milestone = await released;
      await settlement;
      return { milestone, total: performance.now() - startedAt };
    });

    expect(result.milestone.cycleId).toBe(77);
    expect(result.milestone.key).toBe('deal-cutover');
    expect(result.milestone.settled).toBe(false);
    // The second primary starts 250ms late and must itself cross 25% of 500ms.
    expect(result.milestone.elapsed).toBeGreaterThanOrEqual(340);
    expect(result.milestone.elapsed).toBeLessThan(result.total);
    expect(result.total).toBeGreaterThanOrEqual(680);
    diagnostics.assertEmpty();
  } finally {
    diagnostics.stop();
  }
});

test('Table deal defaults preserve decorative compatibility flights and baselines', async ({ page }) => {
  const diagnostics = await prepareRendererFixturePage(page);
  try {
    const result = await page.evaluate(async () => {
      const { BoardgameTableViewBase } = await import('/src/components/boardgame-table-view-base.ts');
      const { html } = await import('/src/client.ts');
      class TestTableDefaults extends BoardgameTableViewBase<any, any, any, any> {
        calls: Array<{ carrier: string; source: string; duration: number }> = [];
        protected override get animator(): any {
          return {
            animateBetween: (carrier: HTMLElement, source: HTMLElement, duration: number) => {
              this.calls.push({ carrier: carrier.id, source: source.id, duration });
            },
          };
        }
        override render() {
          return html`
            <div id="deal-source"></div>
            <div id="stub:p0:hand"></div>
            <div id="stub:p1:hand"></div>
            <div id="stub:p2:hand"></div>
          `;
        }
      }
      customElements.define('test-table-motion-defaults', TestTableDefaults);
      const table = document.createElement('test-table-motion-defaults') as TestTableDefaults;
      const stack = (indexes: number[]) => ({ Indexes: indexes });
      const state = (players: unknown[]) => ({ Players: players });
      const before = state([
        { A: stack([-1]), B: stack([-2]), ignored: { IDs: ['private'] } },
        { A: stack([0, 1]) },
        { A: stack([]) },
      ]);
      const after = state([
        { A: stack([-1, -1]), B: stack([-2]), ignored: { Indexes: 'not-an-array' } },
        { A: stack([0]), B: stack([1]) },
        { A: stack([-1, -1]) },
      ]);
      document.body.append(table);
      table.state = before as any;
      await table.updateComplete;
      table.state = after as any;
      await table.updateComplete;
      await new Promise(requestAnimationFrame);
      await new Promise(requestAnimationFrame);
      const first = [...table.calls];
      table.autoFlyDeals = false;
      table.state = state([
        { A: stack([-1, -1, -1]) },
        { A: stack([0, 1, 2]) },
        { A: stack([-1, -1, -1]) },
      ]) as any;
      await table.updateComplete;
      await new Promise(requestAnimationFrame);
      await new Promise(requestAnimationFrame);
      const disabled = [...table.calls];
      table.autoFlyDeals = true;
      table.state = state([
        { A: stack([-1, -1, -1]) },
        { A: stack([0, 1, 2, 3]) },
        { A: stack([-1, -1, -1]) },
      ]) as any;
      await table.updateComplete;
      await new Promise(requestAnimationFrame);
      await new Promise(requestAnimationFrame);
      return { first, disabled, final: table.calls };
    });

    expect(result.first).toEqual([
      { carrier: 'stub:p0:hand', source: 'deal-source', duration: 600 },
      { carrier: 'stub:p2:hand', source: 'deal-source', duration: 600 },
    ]);
    expect(result.disabled).toEqual(result.first);
    expect(result.final).toEqual([
      ...result.first,
      { carrier: 'stub:p1:hand', source: 'deal-source', duration: 600 },
    ]);
    diagnostics.assertEmpty();
  } finally {
    diagnostics.stop();
  }
});
