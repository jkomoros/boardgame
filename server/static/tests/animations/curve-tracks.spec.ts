import { test, expect } from '@playwright/test';
import type { Page } from '@playwright/test';

// Task 2: the animation kernel honours a SAMPLED motion track at playback
// time. Task 1 taught the compiler to bake a curve into many keyframes and to
// report the easing a sampled timeline needs (componentMotionTrackEasing) plus
// the value the channel must hold afterwards (ComponentMotionTrack.resting);
// nothing consumed either. These tests pin the three consequences:
//
//   1. per-track timing -- a sampled channel plays under an EFFECT-LEVEL
//      'linear', while an endpoint channel in the same batch keeps the
//      kernel's 'ease-in-out'. Effect level, not per-keyframe: keyframe easing
//      already defaults to linear (so a per-keyframe write is a no-op), and
//      boardgame-component-animator publishes effect.getTiming().easing into
//      StructuralExecutedTiming, which would stop meaning anything if every
//      track reported 'linear' there.
//   2. the resting write -- with fill:'none' both finish() and natural
//      completion render the RESTING style, not the last sample, and the
//      framework writes a resting style for the host channel only.
//   3. per-gated-play declaration -- the watchdog must hear about the LONGEST
//      gated play in a cycle, not just the first.
//
// Fixture-mounted like tests/animations/parity/fading-text.spec.ts, with the
// customElements registration guarded so repeated evaluates are idempotent.

// Runs IN THE PAGE. BoardgameAnimatableItem deliberately resolves only the
// 'host' channel (motionTrackTarget returns null for 'visual'); real
// subclasses point 'visual' at their own inner surface. This probe does the
// same against a plain div so the visual channel is reachable from a test.
const installProbe = async () => {
  const { BoardgameAnimatableItem } = await import('/src/components/boardgame-animatable-item.ts');
  // Mounts a probe plus the div its 'visual' channel resolves to.
  (window as any).__mountCurveProbe = () => {
    const el = document.createElement('curve-track-probe') as any;
    el.style.cssText = 'position:fixed;top:60px;left:60px;width:40px;height:40px;';
    document.body.appendChild(el);
    const visual = document.createElement('div');
    visual.style.cssText = 'position:absolute;width:20px;height:20px;';
    document.body.appendChild(visual);
    el.visualEl = visual;
    return { el, visual };
  };
  if (customElements.get('curve-track-probe')) return;
  class CurveTrackProbe extends BoardgameAnimatableItem {
    visualEl: HTMLElement | null = null;

    protected override motionTrackTarget(target: 'host' | 'visual'): HTMLElement | null {
      return target === 'host' ? this : this.visualEl;
    }

    // playMotionTracks is protected; expose it verbatim for the tests.
    playTracks(
      tracks: Parameters<CurveTrackProbe['playMotionTracks']>[0],
      timing?: OptionalEffectTiming,
      opts?: Parameters<CurveTrackProbe['playMotionTracks']>[2],
    ) {
      return this.playMotionTracks(tracks, timing, opts);
    }
  }
  customElements.define('curve-track-probe', CurveTrackProbe);
};

const setup = async (page: Page) => {
  await page.goto('/');
  await page.evaluate(installProbe);
};

// (a) Per-track timing. One batch, two channels, two different effect-level
// easings -- which a single shared `timing` object structurally cannot
// express, so this also pins that timing is derived per binding.
test('a sampled track plays linear while an endpoint track keeps the kernel easing', async ({ page }) => {
  await setup(page);

  const result = await page.evaluate(async () => {
    const { componentMotionTracks } = await import('/src/motion/component-track.ts');
    const { el, visual } = (window as any).__mountCurveProbe();
    await el.updateComplete;

    const tracks = componentMotionTracks([
      {
        target: 'host', property: 'transform',
        from: 'translateX(0px)', to: 'translateX(40px)',
      },
      {
        target: 'visual', property: 'transform',
        curve: (progress: number) => `rotate(${progress * 720}deg)`,
        resolution: 16,
      },
    ]);

    const played = el.playTracks(tracks, { duration: 400 }, { timing: 'immediate' });
    const byChannel: Record<string, { easing: string; frames: number; target: string }> = {};
    for (const playback of played.playbacks ?? []) {
      const effect = playback.animation.effect as KeyframeEffect;
      byChannel[playback.channel] = {
        easing: String(effect.getTiming().easing),
        frames: effect.getKeyframes().length,
        target: effect.target === el ? 'host' : (effect.target === visual ? 'visual' : 'other'),
      };
      playback.animation.cancel();
    }
    return { status: played.status, byChannel };
  });

  expect(result.status).toBe('started');
  expect(result.byChannel['host:transform']).toEqual({
    easing: 'ease-in-out', frames: 2, target: 'host',
  });
  expect(result.byChannel['visual:transform']).toEqual({
    easing: 'linear', frames: 16, target: 'visual',
  });
});

// (b) Two time warps on one channel is the same class of error as two owners.
// The throw happens BEFORE any animation starts, so there is nothing half
// played to unwind; the control call proves the batch is otherwise playable.
// (b) The easing collision. A sampled track carries its own timeline, so a
// caller-supplied easing on the same channel is a producer error and has to be
// refused before anything plays.
//
// Refused, and REPORTED -- not thrown out of playMotionTracks.
// boardgame-component-animator plays a cycle's components in a bare `for` loop
// and each component writes its own final transform only after playMotionTracks
// returns, so an exception escaping here would unwind the loop and strand every
// component after the offending one at its inverted FLIP transform: one bad
// track, a boardful of components frozen where they were animating out of. The
// diagnosis has to stay loud (console.error names the channel and the easing)
// and the result has to say 'playback-error', but it must be survivable.
test('an explicit timing.easing alongside a sampled track is refused loudly, without throwing', async ({ page }) => {
  await setup(page);

  const result = await page.evaluate(async () => {
    const { componentMotionTracks } = await import('/src/motion/component-track.ts');
    const { el } = (window as any).__mountCurveProbe();
    await el.updateComplete;

    const tracks = componentMotionTracks([
      {
        target: 'visual', property: 'transform',
        curve: (progress: number) => `rotate(${progress * 360}deg)`,
        resolution: 8,
      },
    ]);

    const hooks = (window as any).__bgAnimTestHooks;
    const playsBefore = hooks.plays;
    const logged: string[] = [];
    const originalError = console.error;
    console.error = (...args: unknown[]) => { logged.push(args.map(String).join(' ')); };
    let threw: string | null = null;
    let badStatus: string | null = null;
    let badReason: string | null = null;
    try {
      const bad = el.playTracks(tracks, { duration: 400, easing: 'ease-out' }, { timing: 'immediate' });
      badStatus = bad.status;
      badReason = bad.reason ?? null;
    } catch (error) {
      threw = String((error as Error).message);
    }
    console.error = originalError;
    const playsAfterBad = hooks.plays;

    // Control: the very same batch without an explicit easing plays fine, so
    // the throw above is about the easing collision and not about the batch.
    const control = el.playTracks(tracks, { duration: 400 }, { timing: 'immediate' });
    for (const playback of control.playbacks ?? []) playback.animation.cancel();

    // An endpoint-only batch may still carry a caller easing.
    const endpointOnly = componentMotionTracks([
      { target: 'host', property: 'opacity', from: '1', to: '0.5' },
    ]);
    let endpointThrew: string | null = null;
    let endpointStatus: string | null = null;
    try {
      const ok = el.playTracks(endpointOnly, { duration: 400, easing: 'ease-out' }, { timing: 'immediate' });
      endpointStatus = ok.status;
      for (const playback of ok.playbacks ?? []) playback.animation.cancel();
    } catch (error) {
      endpointThrew = String((error as Error).message);
    }

    return {
      threw,
      badStatus,
      badReason,
      logged,
      playsDuringBad: playsAfterBad - playsBefore,
      controlStatus: control.status,
      endpointThrew,
      endpointStatus,
    };
  });

  expect(result.threw,
    'playMotionTracks must not throw: the animator plays a whole cycle in a bare loop')
    .toBe(null);
  expect(result.badStatus).toBe('skipped');
  expect(result.badReason).toBe('playback-error');
  // Loud and diagnosable: the console line has to name the channel and the
  // easing, or a producer gets a silently dead animation and no lead.
  expect(result.logged.join('\n')).toContain('visual:transform');
  expect(result.logged.join('\n')).toContain('ease-out');
  expect(result.playsDuringBad, 'the refusal must precede every play in the batch').toBe(0);
  expect(result.controlStatus, 'the same batch without an easing must still play').toBe('started');
  expect(result.endpointThrew, 'an endpoint-only batch may carry a caller easing').toBe(null);
  expect(result.endpointStatus).toBe('started');
});

// The consequence, stated as the animator sees it: three components played in
// one bare `for` loop, the middle one poisoned. Before the refusal was made
// survivable this test's third component never played at all and never wrote
// its final transform -- it stayed at the inverted transform the FLIP had put
// it at, i.e. visibly parked wherever it had animated from.
test('a track the kernel refuses does not strand the components played after it', async ({ page }) => {
  await setup(page);

  const result = await page.evaluate(async () => {
    const { componentMotionTracks } = await import('/src/motion/component-track.ts');
    const probes = [0, 1, 2].map(() => (window as any).__mountCurveProbe());
    for (const probe of probes) await probe.el.updateComplete;

    const sampled = componentMotionTracks([{
      target: 'visual', property: 'transform',
      curve: (progress: number) => `rotate(${progress * 360}deg)`,
      resolution: 8,
    }]);
    const endpoint = componentMotionTracks([
      { target: 'host', property: 'opacity', from: '1', to: '0.5' },
    ]);

    const originalError = console.error;
    console.error = () => undefined;
    const statuses: string[] = [];
    const finals: string[] = [];
    let escaped: string | null = null;
    try {
      // The animator's loop, in miniature: for each component, play its tracks
      // and then write its own final transform. The write is what a stranded
      // component never reaches.
      for (const [index, probe] of probes.entries()) {
        const poisoned = index === 1;
        const result = probe.el.playTracks(
          poisoned ? sampled : endpoint,
          poisoned ? { duration: 400, easing: 'ease-out' } : { duration: 400 },
          { timing: 'immediate' },
        );
        statuses.push(result.status);
        for (const playback of result.playbacks ?? []) playback.animation.cancel();
        probe.el.style.transform = `translateX(${index * 10}px)`;
        finals.push(probe.el.style.transform);
      }
    } catch (error) {
      escaped = String((error as Error).message);
    }
    console.error = originalError;
    for (const probe of probes) { probe.el.remove(); probe.visual.remove(); }
    return { escaped, statuses, finals };
  });

  expect(result.escaped, 'nothing may escape into the animator\'s loop').toBe(null);
  // The middle component fails, alone.
  expect(result.statuses).toEqual(['started', 'skipped', 'started']);
  // And every component -- including the one played AFTER the failure -- ends
  // up on its own final transform rather than parked at its inverted one.
  expect(result.finals).toEqual([
    'translateX(0px)', 'translateX(10px)', 'translateX(20px)',
  ]);
});

// (c) The resting-pose contract. Animations run with fill:'none', so the
// moment one finishes (or is finished early, which is what reduced motion and
// the cycle sweep both do) the element renders its resting style. The
// framework writes one for 'host' only, so without this the visual channel
// snaps back to whatever it had before the curve ran.
test('starting a sampled track writes its resting value to the resolved target', async ({ page }) => {
  await setup(page);

  const result = await page.evaluate(async () => {
    const { componentMotionTracks } = await import('/src/motion/component-track.ts');
    const { el, visual } = (window as any).__mountCurveProbe();
    await el.updateComplete;

    const transformTracks = componentMotionTracks([
      {
        target: 'visual', property: 'transform',
        curve: (progress: number) => `translateX(${progress * 30}px)`,
        resolution: 8,
        resting: 'translateX(30px) rotate(45deg)',
      },
    ]);
    const transformPlay = el.playTracks(transformTracks, { duration: 400 }, { timing: 'immediate' });
    const inlineTransform = visual.style.transform;
    for (const playback of transformPlay.playbacks ?? []) playback.animation.cancel();
    const afterCancelTransform = visual.style.transform;

    visual.style.opacity = '';
    const opacityTracks = componentMotionTracks([
      {
        target: 'visual', property: 'opacity',
        curve: (progress: number) => String(1 - progress),
        resolution: 8,
        resting: '0.25',
      },
    ]);
    const opacityPlay = el.playTracks(opacityTracks, { duration: 400 }, { timing: 'immediate' });
    const inlineOpacity = visual.style.opacity;
    for (const playback of opacityPlay.playbacks ?? []) playback.animation.cancel();

    // Default resting is curve(1): a curve that does not declare one still
    // holds where it landed.
    const visual2 = document.createElement('div');
    document.body.appendChild(visual2);
    el.visualEl = visual2;
    const defaultTracks = componentMotionTracks([
      {
        target: 'visual', property: 'transform',
        curve: (progress: number) => `translateY(${progress * 12}px)`,
        resolution: 4,
      },
    ]);
    const defaultPlay = el.playTracks(defaultTracks, { duration: 400 }, { timing: 'immediate' });
    const inlineDefault = visual2.style.transform;
    for (const playback of defaultPlay.playbacks ?? []) playback.animation.cancel();

    return {
      inlineTransform,
      afterCancelTransform,
      inlineOpacity,
      inlineDefault,
      declaredResting: transformTracks[0].resting,
      declaredDefaultResting: defaultTracks[0].resting,
    };
  });

  expect(result.declaredResting).toBe('translateX(30px) rotate(45deg)');
  expect(result.inlineTransform,
    'the resolved visual target must hold the track resting transform')
    .toBe('translateX(30px) rotate(45deg)');
  expect(result.afterCancelTransform).toBe('translateX(30px) rotate(45deg)');
  expect(result.inlineOpacity, 'opacity tracks write style.opacity').toBe('0.25');
  expect(result.declaredDefaultResting).toBe('translateY(12px)');
  expect(result.inlineDefault, 'the default resting is curve(1)').toBe('translateY(12px)');
});

// (d) Every gated play declares itself. The gate is wired exactly as
// boardgame-render-game wires it (_componentWillAnimate ->
// gate.willAnimate(ele, label, expectedSettleMs)), so the assertion is about
// the real watchdog arithmetic and not a restatement of the event count.
// Two gated plays on ONE element: a short host FLIP, then the long sampled
// tumble. Before this change only the 0->1 gated transition dispatched
// will-animate, so the gate never heard 6000ms, kept the 4000ms floor, and
// force-closed the cycle mid-tumble.
test('every gated play declares itself so the watchdog tracks the longest', async ({ page }) => {
  test.setTimeout(30_000);
  await setup(page);

  const result = await page.evaluate(async () => {
    const { componentMotionTracks } = await import('/src/motion/component-track.ts');
    const { AnimationGate } = await import('/src/motion/animation-gate.ts');
    const { el } = (window as any).__mountCurveProbe();
    await el.updateComplete;

    const hooks = (window as any).__bgAnimTestHooks;
    const watchdogBudgets: number[] = [];
    const declared: Array<number | undefined> = [];
    const gate = new AnimationGate({
      onOpen: () => {},
      onAllDone: () => {},
      onWatchdog: (_pending: readonly string[], budgetMs: number) => {
        watchdogBudgets.push(budgetMs);
        hooks.record('watchdog', 'curve-track-probe');
      },
      setTimer: (cb: () => void, ms: number) => setTimeout(cb, ms),
      clearTimer: (handle: unknown) => clearTimeout(handle as ReturnType<typeof setTimeout>),
      now: () => Date.now(),
    });
    document.addEventListener('will-animate', (event: Event) => {
      const detail = (event as CustomEvent).detail;
      declared.push(detail.expectedSettleMs);
      gate.willAnimate(detail.ele, 'curve-track-probe', detail.expectedSettleMs);
    });
    document.addEventListener('animation-done', (event: Event) => {
      gate.animationDone((event as CustomEvent).detail.ele);
    });

    const watchdogsBefore = hooks.watchdogFirings;
    gate.open(1);

    // The short structural FLIP the compiler orders FIRST.
    const short = el.play(
      el,
      [{ transform: 'translateX(0px)' }, { transform: 'translateX(20px)' }],
      { duration: 150 },
      { timing: 'immediate' },
    );
    // The long sampled tumble the compiler orders LAST.
    const tumble = componentMotionTracks([
      {
        target: 'visual', property: 'transform',
        curve: (progress: number) => `rotate(${progress * 1440}deg)`,
        resolution: 32,
      },
    ]);
    const long = el.playTracks(tumble, { duration: 6000 }, { timing: 'immediate' });

    // Past the 4000ms watchdog floor, well short of the 6000ms tumble.
    await new Promise<void>((resolve) => setTimeout(resolve, 5200));

    const observed = {
      declared: [...declared],
      watchdogBudgets: [...watchdogBudgets],
      watchdogDelta: hooks.watchdogFirings - watchdogsBefore,
      gateStillOpen: gate.isOpen,
      gatePending: gate.pendingCount,
      longStatus: long.status,
    };
    short?.cancel();
    for (const playback of long.playbacks ?? []) playback.animation.cancel();
    return observed;
  });

  expect(result.longStatus).toBe('started');
  // The consequence first: an undeclared long play leaves the gate on its
  // 4000ms floor and the watchdog force-closes the cycle mid-tumble.
  expect(result.watchdogBudgets,
    'the watchdog must not fire: the longer declaration pushes the deadline past 4000ms')
    .toEqual([]);
  expect(result.watchdogDelta).toBe(0);
  expect(result.gateStillOpen, 'the cycle must still be open 5.2s into a 6s tumble').toBe(true);
  expect(result.gatePending).toBe(1);
  // Then the mechanism.
  expect(result.declared,
    'both gated plays must declare; only the second knows the cycle really runs 6000ms')
    .toEqual([150, 6000]);
});
