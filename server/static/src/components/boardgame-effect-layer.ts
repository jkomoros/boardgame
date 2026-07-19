import { css, html, LitElement } from 'lit';
import { BoardgameAnimatableItem } from './boardgame-animatable-item.js';
import { reserveEffectBudget } from '../effects/effect-budget.js';
import { burstParticles, MAX_BURST_PARTICLES } from '../effects/particle-burst.js';
import { DEFAULT_EFFECT_THEME } from '../effects/effect-spec.js';
import type {
  EffectAnchor,
  EffectHandle,
  EffectHostAPI,
  EffectIntensity,
  EffectPointAnchor,
  EffectResult,
  EffectSpec,
  EffectTheme,
  EffectTone,
  MotionEffectAnchor,
  PulseEffectSpec,
  TrailEffectSpec,
} from '../effects/effect-spec.js';
import type { AnimationTimingPolicy } from './boardgame-animatable-item.js';
import { captureViewportGeometry, geometryCenter } from '../motion/geometry.js';
import type {
  StructuralMotionEvent,
  StructuralMotionSegmentEvent,
} from '../motion/structural-events.js';

interface StructuralMotionSource {
  observeStructuralMotionEvents(observer: (event: StructuralMotionEvent) => void): () => void;
}

export interface EffectLayerConfiguration {
  anchorRoot: ParentNode | null;
  seedScope: string;
  theme: EffectTheme;
  animationContext: import('./companion-sync.js').VersionAnimationContext | null;
  /** Internal source for automatic component-motion endpoint events. */
  motionSource?: StructuralMotionSource | null;
}

export type EffectAnchorSnapshot = ReadonlyMap<string, Readonly<{ x: number; y: number }>>;

interface ResolvedPolicy {
  tone: EffectTone;
  intensity: EffectIntensity;
  timing: AnimationTimingPolicy;
  seedKey: string;
}

interface InternalHandle extends EffectHandle {}

interface EffectExecutionContext {
  readonly preparedMotion: Map<string, InternalHandle>;
}

type MotionResolution =
  | Readonly<{ kind: 'point'; point: Readonly<{ x: number; y: number }> }>
  | Readonly<{ kind: 'result'; result: EffectResult }>;

interface MotionWaiter {
  readonly epoch: number;
  readonly key: string;
  settle(resolution: MotionResolution): void;
}

type TrailResolution =
  | Readonly<{ kind: 'event'; event: StructuralMotionSegmentEvent }>
  | Readonly<{ kind: 'result'; result: EffectResult }>;

interface TrailWaiter {
  readonly epoch: number;
  readonly subjectId: string;
  settle(resolution: TrailResolution): void;
}

const FINISHED: EffectResult = Object.freeze({ status: 'finished' });
const CANCELLED: EffectResult = Object.freeze({ status: 'cancelled' });

const INTENSITY = Object.freeze({
  subtle: Object.freeze({ count: 5, spread: 28, duration: 240, pulseScale: 1.12, travelSize: 7, arc: 24 }),
  small: Object.freeze({ count: 8, spread: 42, duration: 340, pulseScale: 1.2, travelSize: 9, arc: 38 }),
  medium: Object.freeze({ count: 14, spread: 72, duration: 520, pulseScale: 1.32, travelSize: 11, arc: 56 }),
  large: Object.freeze({ count: 22, spread: 110, duration: 720, pulseScale: 1.48, travelSize: 14, arc: 82 }),
}) satisfies Record<EffectIntensity, Readonly<{
  count: number;
  spread: number;
  duration: number;
  pulseScale: number;
  travelSize: number;
  arc: number;
}>>;

const TRAIL_POLICY = Object.freeze({
  subtle: Object.freeze({ echoes: 2, lagMs: 18, opacity: 0.22 }),
  small: Object.freeze({ echoes: 3, lagMs: 20, opacity: 0.3 }),
  medium: Object.freeze({ echoes: 5, lagMs: 22, opacity: 0.38 }),
  large: Object.freeze({ echoes: 7, lagMs: 24, opacity: 0.48 }),
}) satisfies Record<EffectIntensity, Readonly<{
  echoes: number;
  lagMs: number;
  opacity: number;
}>>;

const TONE_PALETTES = Object.freeze({
  neutral: Object.freeze([
    'var(--boardgame-effect-neutral, var(--md-sys-color-outline, #79747e))',
    'var(--boardgame-effect-neutral-glow, var(--md-sys-color-surface-variant, #e7e0ec))',
  ]),
  reward: Object.freeze([
    'var(--boardgame-effect-reward, #f6bf26)',
    'var(--boardgame-effect-reward-accent, var(--md-sys-color-tertiary, #7d5260))',
  ]),
  confirm: Object.freeze([
    'var(--boardgame-effect-confirm, var(--md-sys-color-primary, #2e6b4f))',
    'var(--boardgame-effect-confirm-glow, var(--md-sys-color-primary-container, #d8eadf))',
  ]),
  attention: Object.freeze([
    'var(--boardgame-effect-attention, var(--md-sys-color-secondary, #625b71))',
    'var(--boardgame-effect-attention-glow, var(--md-sys-color-secondary-container, #e8def8))',
  ]),
  warning: Object.freeze([
    'var(--boardgame-effect-warning, #b26a00)',
    'var(--boardgame-effect-warning-glow, #ffddb3)',
  ]),
  magic: Object.freeze([
    'var(--boardgame-effect-magic, #7c4dff)',
    'var(--boardgame-effect-magic-accent, #ff4081)',
    'var(--boardgame-effect-magic-glow, #80d8ff)',
  ]),
}) satisfies Record<EffectTone, readonly string[]>;

function finite(value: number | undefined, fallback: number, min: number, max: number): number {
  const resolved = typeof value === 'number' && Number.isFinite(value) ? value : fallback;
  return Math.max(min, Math.min(max, resolved));
}

function skipped(reason: Extract<EffectResult, { status: 'skipped' }>['reason']): EffectHandle {
  return { finished: Promise.resolve(Object.freeze({ status: 'skipped', reason })), cancel: () => {} };
}

export class BoardgameEffectLayer extends LitElement implements EffectHostAPI {
  static override styles = css`
    :host {
      position: fixed;
      inset: 0;
      z-index: 1000;
      overflow: hidden;
      pointer-events: none;
      contain: strict;
    }

    #effects {
      position: absolute;
      inset: 0;
      overflow: hidden;
      pointer-events: none;
    }

    boardgame-animatable-item {
      display: none;
    }

    .particle,
    .traveler,
    .pulse,
    .trail-echo {
      position: absolute;
      left: 0;
      top: 0;
      box-sizing: border-box;
      pointer-events: none;
      will-change: transform, opacity;
    }

    .particle,
    .traveler {
      width: var(--effect-size);
      height: var(--effect-size);
      border-radius: 999px;
      background: var(--effect-color);
      box-shadow: 0 0 calc(var(--effect-size) * 1.4) var(--effect-color);
    }

    .trail-echo {
      background: var(--effect-color);
      box-shadow: 0 0 16px color-mix(in srgb, var(--effect-color) 55%, transparent);
      filter: blur(1px);
    }

    .particle:nth-child(3n) {
      border-radius: 2px;
    }

    .pulse {
      width: var(--effect-size);
      height: var(--effect-size);
      border: 3px solid var(--effect-color);
      border-radius: 999px;
      box-shadow: 0 0 18px color-mix(in srgb, var(--effect-color) 65%, transparent);
    }
  `;

  private _configuration: EffectLayerConfiguration = {
    anchorRoot: null,
    seedScope: 'unscoped-effect',
    theme: DEFAULT_EFFECT_THEME,
    animationContext: null,
  };
  private readonly _activeCancels = new Set<() => void>();
  private readonly _transitionCancels = new Set<() => void>();
  private _beforeAnchors: EffectAnchorSnapshot = new Map();
  private _motionSource: StructuralMotionSource | null = null;
  private _unobserveMotionEvents: (() => void) | null = null;
  private _motionEpoch = 0;
  private _expectsMotionPlan = false;
  private _motionPlanSettled = false;
  private _motionResolutions = new Map<string, MotionResolution>();
  private _motionWaiters = new Set<MotionWaiter>();
  private _trailResolutions = new Map<string, TrailResolution>();
  private _trailWaiters = new Set<TrailWaiter>();
  private _activeTrailCancels = new Map<string, Set<() => void>>();

  override render() {
    return html`
      <boardgame-animatable-item id="runner"></boardgame-animatable-item>
      <div id="effects" aria-hidden="true"></div>
    `;
  }

  configure(configuration: EffectLayerConfiguration): void {
    if (configuration.anchorRoot !== this._configuration.anchorRoot) {
      this._beforeAnchors = new Map();
    }
    this._configuration = {
      anchorRoot: configuration.anchorRoot,
      seedScope: configuration.seedScope.trim() || 'unscoped-effect',
      theme: configuration.theme,
      animationContext: configuration.animationContext,
    };
    this._setMotionSource(configuration.motionSource ?? null);
    const runner = this._runner();
    if (runner) runner.animationContext = configuration.animationContext;
  }

  captureNamedAnchors(): EffectAnchorSnapshot {
    const result = new Map<string, Readonly<{ x: number; y: number }>>();
    const root = this._configuration.anchorRoot;
    if (!root) return result;
    for (const element of root.querySelectorAll<HTMLElement>('[data-effect-anchor]')) {
      const name = element.dataset.effectAnchor?.trim();
      if (!name || result.has(name) || !element.isConnected) continue;
      const { x, y } = geometryCenter(captureViewportGeometry(element));
      result.set(name, Object.freeze({ x, y }));
    }
    return result;
  }

  installBeforeAnchors(snapshot: EffectAnchorSnapshot): void {
    this._beforeAnchors = new Map(snapshot);
  }

  override disconnectedCallback(): void {
    this.cancelAll();
    this._setMotionSource(null);
    super.disconnectedCallback();
  }

  /** Start a renderer transition scope before its descriptors are installed. */
  beginMotionTransition(expectsStructuralMotion: boolean): void {
    const abandoned = Object.freeze({ kind: 'result', result: CANCELLED }) as MotionResolution;
    const abandonedTrail = Object.freeze({ kind: 'result', result: CANCELLED }) as TrailResolution;
    for (const waiter of [...this._motionWaiters]) waiter.settle(abandoned);
    for (const waiter of [...this._trailWaiters]) waiter.settle(abandonedTrail);
    this._motionEpoch++;
    this._expectsMotionPlan = expectsStructuralMotion;
    this._motionPlanSettled = !expectsStructuralMotion;
    this._motionResolutions.clear();
    this._trailResolutions.clear();
  }

  play(effect: EffectSpec): EffectHandle {
    return this._start(effect, false);
  }

  playTransition(effect: EffectSpec): EffectHandle {
    return this._start(effect, true);
  }

  private _start(effect: EffectSpec, transitionOwned: boolean): EffectHandle {
    if (!this.isConnected) return skipped('not-connected');
    const context: EffectExecutionContext = { preparedMotion: new Map() };
    const rootPolicy: ResolvedPolicy = {
      tone: 'neutral',
      intensity: 'medium',
      timing: 'immediate',
      seedKey: '',
    };
    this._prepareMotionDecorations(effect, '0', rootPolicy, context);
    const running = this._execute(effect, '0', rootPolicy, context);
    const cancel = () => {
      running.cancel();
      for (const prepared of context.preparedMotion.values()) prepared.cancel();
    };
    this._activeCancels.add(cancel);
    if (transitionOwned) this._transitionCancels.add(cancel);
    void running.finished.finally(() => {
      this._activeCancels.delete(cancel);
      this._transitionCancels.delete(cancel);
    });
    return { finished: running.finished, cancel };
  }

  cancelTransitionEffects(): void {
    for (const cancel of [...this._transitionCancels]) cancel();
  }

  cancelAll(): void {
    for (const cancel of [...this._activeCancels]) cancel();
  }

  private _resolvedPolicy(effect: EffectSpec, inherited: ResolvedPolicy): ResolvedPolicy {
    return {
      tone: effect.tone ?? inherited.tone,
      intensity: effect.intensity ?? inherited.intensity,
      timing: effect.timing ?? inherited.timing,
      seedKey: effect.seedKey ?? effect.key ?? inherited.seedKey,
    };
  }

  private _prepareMotionDecorations(
    effect: EffectSpec,
    path: string,
    inherited: ResolvedPolicy,
    context: EffectExecutionContext,
  ): void {
    const policy = this._resolvedPolicy(effect, inherited);
    if (effect.kind === 'decorate-motion') {
      if (!context.preparedMotion.has(path)) {
        context.preparedMotion.set(
          path,
          this._parallel(effect.effects, path, { ...policy, timing: 'immediate' }, context),
        );
      }
      return;
    }
    if (effect.kind === 'sequence' || effect.kind === 'parallel') {
      effect.effects.forEach((child, index) => {
        this._prepareMotionDecorations(child, `${path}.${index}`, policy, context);
      });
    }
  }

  private _execute(
    effect: EffectSpec,
    path: string,
    inherited: ResolvedPolicy,
    context: EffectExecutionContext,
  ): InternalHandle {
    const policy = this._resolvedPolicy(effect, inherited);
    switch (effect.kind) {
      case 'burst':
        return this._burst(effect, path, policy);
      case 'pulse':
        return this._pulse(
          effect,
          path,
          policy,
          window.matchMedia('(prefers-reduced-motion: reduce)').matches,
        );
      case 'travel':
        return this._travel(effect, path, policy);
      case 'trail':
        return this._trail(effect, path, policy);
      case 'decorate-motion':
        return context.preparedMotion.get(path)
          ?? this._parallel(effect.effects, path, { ...policy, timing: 'immediate' }, context);
      case 'sequence':
        return this._sequence(effect.effects, effect.gapMs, path, policy, context);
      case 'parallel':
        return this._parallel(effect.effects, path, policy, context);
    }
  }

  private _burst(effect: Extract<EffectSpec, { kind: 'burst' }>, path: string, policy: ResolvedPolicy): InternalHandle {
    return this._atPoint(effect.at, point => this._burstAt(effect, path, policy, point));
  }

  private _burstAt(
    effect: Extract<EffectSpec, { kind: 'burst' }>,
    path: string,
    policy: ResolvedPolicy,
    point: Readonly<{ x: number; y: number }>,
  ): InternalHandle {
    if (window.matchMedia('(prefers-reduced-motion: reduce)').matches) {
      const substitute: PulseEffectSpec = { ...effect, kind: 'pulse', at: effect.at };
      return this._pulseAt(substitute, path, policy, true, point);
    }
    const defaults = INTENSITY[policy.intensity];
    const requestedCount = finite(effect.advanced?.count, defaults.count, 1, MAX_BURST_PARTICLES);
    const reservation = reserveEffectBudget(this.ownerDocument, requestedCount);
    if (!reservation) return skipped('budget');

    const spread = finite(effect.advanced?.spreadPx, defaults.spread, 8, 240);
    const duration = finite(effect.advanced?.durationMs, defaults.duration, 120, 1200);
    const palette = this._palette(policy.tone, effect.advanced?.palette);
    const layout = burstParticles(
      reservation.particles,
      spread,
      `${this._configuration.seedScope}:${policy.seedKey || effect.kind}:${path}`,
    );
    const elements: HTMLElement[] = [];
    const animations: Animation[] = [];
    const container = this._container();
    if (!container) {
      reservation.release();
      return skipped('not-connected');
    }

    layout.forEach((particle, index) => {
      const element = this.ownerDocument.createElement('i');
      element.className = 'particle';
      element.style.left = `${point.x}px`;
      element.style.top = `${point.y}px`;
      element.style.setProperty('--effect-size', `${particle.sizePx}px`);
      element.style.setProperty('--effect-color', palette[index % palette.length]);
      container.appendChild(element);
      elements.push(element);

      const radians = particle.angleDeg * Math.PI / 180;
      const dx = Math.cos(radians) * particle.distancePx;
      const dy = Math.sin(radians) * particle.distancePx;
      const animation = this.playAnimation(element, [
        { transform: 'translate(-50%, -50%) scale(0.35)', opacity: 0 },
        { transform: 'translate(-50%, -50%) scale(1)', opacity: 1, offset: 0.16 },
        { transform: `translate(calc(-50% + ${dx}px), calc(-50% + ${dy}px)) rotate(${particle.rotationDeg}deg) scale(0)`, opacity: 0 },
      ], { duration, delay: particle.delayMs, easing: 'cubic-bezier(0.2, 0.8, 0.2, 1)', fill: 'both' }, policy.timing);
      if (animation) animations.push(animation);
    });
    return this._settlingHandle(animations, elements, reservation.release);
  }

  private _pulse(
    effect: PulseEffectSpec,
    _path: string,
    policy: ResolvedPolicy,
    reduced: boolean,
  ): InternalHandle {
    return this._atPoint(effect.at, point => this._pulseAt(effect, _path, policy, reduced, point));
  }

  private _pulseAt(
    effect: PulseEffectSpec,
    _path: string,
    policy: ResolvedPolicy,
    reduced: boolean,
    point: Readonly<{ x: number; y: number }>,
  ): InternalHandle {
    const reservation = reserveEffectBudget(this.ownerDocument, 1);
    if (!reservation) return skipped('budget');
    const container = this._container();
    if (!container) {
      reservation.release();
      return skipped('not-connected');
    }
    const defaults = INTENSITY[policy.intensity];
    const duration = reduced
      ? 160
      : finite(effect.advanced?.durationMs, defaults.duration * 0.72, 120, 900);
    const scale = finite(effect.advanced?.scale, defaults.pulseScale, 1.02, 2);
    const palette = this._palette(policy.tone, effect.advanced?.palette);
    const element = document.createElement('i');
    element.className = 'pulse';
    element.style.left = `${point.x}px`;
    element.style.top = `${point.y}px`;
    element.style.setProperty('--effect-size', `${reduced ? 34 : 42}px`);
    element.style.setProperty('--effect-color', palette[0]);
    container.appendChild(element);

    // Reduced motion deliberately uses a short opacity emphasis rather than
    // calling the shared play primitive, whose policy correctly zeros motion.
    const animation = reduced
      ? element.animate([
        { transform: 'translate(-50%, -50%)', opacity: 0.75 },
        { transform: 'translate(-50%, -50%)', opacity: 0 },
      ], { duration, easing: 'ease-out', fill: 'both' })
      : this.playAnimation(element, [
        { transform: 'translate(-50%, -50%) scale(0.55)', opacity: 0.9 },
        { transform: `translate(-50%, -50%) scale(${scale})`, opacity: 0 },
      ], { duration, easing: 'cubic-bezier(0.2, 0.8, 0.2, 1)', fill: 'both' }, policy.timing);
    if (!animation) {
      element.remove();
      reservation.release();
      return skipped('timing');
    }
    return this._settlingHandle([animation], [element], reservation.release);
  }

  private _travel(effect: Extract<EffectSpec, { kind: 'travel' }>, path: string, policy: ResolvedPolicy): InternalHandle {
    if (window.matchMedia('(prefers-reduced-motion: reduce)').matches) {
      return this._pulse({ ...effect, kind: 'pulse', at: effect.to }, path, policy, true);
    }
    const from = this._anchorPoint(effect.from);
    const to = this._anchorPoint(effect.to);
    if (!from || !to) return skipped('missing-anchor');
    const reservation = reserveEffectBudget(this.ownerDocument, 1);
    if (!reservation) return skipped('budget');
    const container = this._container();
    if (!container) {
      reservation.release();
      return skipped('not-connected');
    }
    const defaults = INTENSITY[policy.intensity];
    const duration = finite(effect.advanced?.durationMs, defaults.duration, 160, 1200);
    const size = finite(effect.advanced?.sizePx, defaults.travelSize, 4, 32);
    const arc = finite(effect.advanced?.arcPx, defaults.arc, -240, 240);
    const midX = (from.x + to.x) / 2;
    const midY = (from.y + to.y) / 2 - arc;
    const palette = this._palette(policy.tone, effect.advanced?.palette);
    const element = document.createElement('i');
    element.className = 'traveler';
    element.style.setProperty('--effect-size', `${size}px`);
    element.style.setProperty('--effect-color', palette[0]);
    container.appendChild(element);
    const animation = this.playAnimation(element, [
      { transform: `translate(${from.x}px, ${from.y}px) translate(-50%, -50%) scale(0.6)`, opacity: 0 },
      { transform: `translate(${from.x}px, ${from.y}px) translate(-50%, -50%) scale(1)`, opacity: 1, offset: 0.12 },
      { transform: `translate(${midX}px, ${midY}px) translate(-50%, -50%) scale(1.15)`, opacity: 1, offset: 0.55 },
      { transform: `translate(${to.x}px, ${to.y}px) translate(-50%, -50%) scale(0.65)`, opacity: 0 },
    ], { duration, easing: 'cubic-bezier(0.3, 0, 0.2, 1)', fill: 'both' }, policy.timing);
    if (!animation) {
      element.remove();
      reservation.release();
      return skipped('timing');
    }
    return this._settlingHandle([animation], [element], reservation.release);
  }

  private _trail(effect: TrailEffectSpec, path: string, policy: ResolvedPolicy): InternalHandle {
    // Structural execution is the trail's clock, even when a composition has
    // an inherited timing policy. Reduced motion likewise decorates arrival
    // immediately instead of scheduling a second version slot.
    const structuralPolicy = { ...policy, timing: 'immediate' as const };
    if (window.matchMedia('(prefers-reduced-motion: reduce)').matches) {
      return this._pulse({
        kind: 'pulse',
        at: { kind: 'motion', subjectId: effect.subject, moment: 'arrival' },
      }, path, structuralPolicy, true);
    }
    return this._deferTrail(effect, path, structuralPolicy);
  }

  private _sequence(
    effects: readonly EffectSpec[],
    gapMs: number,
    path: string,
    policy: ResolvedPolicy,
    context: EffectExecutionContext,
  ): InternalHandle {
    let current: InternalHandle | null = null;
    let timer: ReturnType<typeof setTimeout> | null = null;
    let finishWait: (() => void) | null = null;
    let cancelled = false;
    let resolveFinished!: (result: EffectResult) => void;
    const finished = new Promise<EffectResult>(resolve => { resolveFinished = resolve; });
    const wait = () => new Promise<void>(resolve => {
      finishWait = resolve;
      timer = setTimeout(() => {
        timer = null;
        finishWait = null;
        resolve();
      }, gapMs);
    });
    void (async () => {
      let completed = false;
      let lastSkipped: EffectResult | null = null;
      for (let index = 0; index < effects.length && !cancelled; index++) {
        current = this._execute(effects[index], `${path}.${index}`, policy, context);
        const result = await current.finished;
        current = null;
        if (result.status === 'finished') completed = true;
        else if (result.status === 'skipped') lastSkipped = result;
        if (gapMs > 0 && index < effects.length - 1 && !cancelled) await wait();
      }
      if (cancelled) resolveFinished(CANCELLED);
      else resolveFinished(completed ? FINISHED : lastSkipped ?? FINISHED);
    })();
    return {
      finished,
      cancel(): void {
        if (cancelled) return;
        cancelled = true;
        if (timer !== null) { clearTimeout(timer); timer = null; }
        finishWait?.();
        finishWait = null;
        current?.cancel();
      },
    };
  }

  private _parallel(
    effects: readonly EffectSpec[],
    path: string,
    policy: ResolvedPolicy,
    context: EffectExecutionContext,
  ): InternalHandle {
    const children = effects.map((effect, index) => (
      this._execute(effect, `${path}.${index}`, policy, context)
    ));
    let cancelled = false;
    const finished = Promise.all(children.map(child => child.finished)).then(results => {
      if (cancelled) return CANCELLED;
      if (results.some(result => result.status === 'finished')) return FINISHED;
      return results.find(result => result.status === 'skipped') ?? FINISHED;
    });
    return {
      finished,
      cancel(): void {
        if (cancelled) return;
        cancelled = true;
        for (const child of children) child.cancel();
      },
    };
  }

  private _settlingHandle(
    animations: Animation[],
    elements: HTMLElement[],
    release: () => void,
  ): InternalHandle {
    let settled = false;
    let resolveFinished!: (result: EffectResult) => void;
    const finished = new Promise<EffectResult>(resolve => { resolveFinished = resolve; });
    const settle = (result: EffectResult) => {
      if (settled) return;
      settled = true;
      for (const element of elements) element.remove();
      release();
      resolveFinished(result);
    };
    void Promise.all(animations.map(animation => animation.finished.catch(() => {})))
      .then(() => settle(FINISHED));
    return {
      finished,
      cancel(): void {
        if (settled) return;
        for (const animation of animations) animation.cancel();
        settle(CANCELLED);
      },
    };
  }

  private playAnimation(
    element: HTMLElement,
    keyframes: Keyframe[],
    timing: OptionalEffectTiming,
    policy: AnimationTimingPolicy,
  ): Animation | null {
    const runner = this._runner();
    if (!runner) return null;
    runner.animationContext = this._configuration.animationContext;
    return runner.play(element, keyframes, timing, { gated: false, timing: policy });
  }

  private _palette(tone: EffectTone, advanced?: readonly string[]): readonly string[] {
    const validAdvanced = advanced?.filter(color => typeof color === 'string' && color.trim()) ?? [];
    if (validAdvanced.length) return validAdvanced;
    const themed = this._configuration.theme.tones?.[tone]?.filter(color => color.trim()) ?? [];
    return themed.length ? themed : TONE_PALETTES[tone];
  }

  private _container(): HTMLElement | null {
    return this.shadowRoot?.querySelector<HTMLElement>('#effects') ?? null;
  }

  private _runner(): BoardgameAnimatableItem | null {
    const runner = this.shadowRoot?.querySelector('#runner');
    return runner instanceof BoardgameAnimatableItem ? runner : null;
  }

  private _setMotionSource(source: StructuralMotionSource | null): void {
    if (source === this._motionSource) return;
    this._unobserveMotionEvents?.();
    this._unobserveMotionEvents = null;
    this._motionSource = source;
    if (!source) return;
    this._unobserveMotionEvents = source.observeStructuralMotionEvents(
      event => this._motionEvent(event),
    );
  }

  private _motionEvent(event: StructuralMotionEvent): void {
    if (event.source !== 'flip') return;
    if (event.kind === 'generation-settled') {
      this._motionPlanSettled = true;
      const missing = Object.freeze({
        kind: 'result',
        result: Object.freeze({ status: 'skipped', reason: 'missing-anchor' }),
      }) as MotionResolution;
      for (const waiter of [...this._motionWaiters]) {
        if (waiter.epoch === this._motionEpoch) waiter.settle(missing);
      }
      const missingSubject = Object.freeze({
        kind: 'result',
        result: Object.freeze({ status: 'skipped', reason: 'missing-subject' }),
      }) as TrailResolution;
      for (const waiter of [...this._trailWaiters]) {
        if (waiter.epoch === this._motionEpoch) waiter.settle(missingSubject);
      }
      return;
    }
    const endpoints = event.segment.path;
    if ((event.kind === 'started' || event.kind === 'finished') && !endpoints) {
      const missing = Object.freeze({
        kind: 'result',
        result: Object.freeze({ status: 'skipped', reason: 'missing-anchor' }),
      }) as MotionResolution;
      this._resolveMotion(event.subjectId, 'departure', missing);
      this._resolveMotion(event.subjectId, 'arrival', missing);
      if (event.kind === 'started') {
        this._resolveTrail(event.subjectId, {
          kind: 'result',
          result: Object.freeze({ status: 'skipped', reason: 'no-motion-path' }),
        });
      }
      return;
    }
    if (event.kind === 'started' && endpoints) {
      this._resolveMotion(event.subjectId, 'departure', {
        kind: 'point', point: geometryCenter(endpoints.from),
      });
      this._resolveTrail(event.subjectId, { kind: 'event', event });
    } else if (event.kind === 'finished' && endpoints) {
      this._resolveMotion(event.subjectId, 'arrival', {
        kind: 'point', point: geometryCenter(endpoints.to),
      });
    } else if (event.kind === 'skipped' || event.kind === 'cancelled') {
      const unavailableResult = Object.freeze({
        status: 'skipped', reason: 'motion-skipped',
      }) as EffectResult;
      const unavailable = Object.freeze({
        kind: 'result',
        result: unavailableResult,
      }) as MotionResolution;
      this._resolveMotion(event.subjectId, 'departure', unavailable);
      this._resolveMotion(event.subjectId, 'arrival', unavailable);
      this._resolveTrail(event.subjectId, {
        kind: 'result', result: unavailableResult,
      });
      if (event.kind === 'cancelled') {
        const key = this._activeTrailKey(event);
        for (const cancel of [...(this._activeTrailCancels.get(key) ?? [])]) cancel();
      }
    }
  }

  private _resolveMotion(
    subjectId: string,
    moment: MotionEffectAnchor['moment'],
    input: MotionResolution,
  ): void {
    const resolution: MotionResolution = input.kind === 'point'
      ? Object.freeze({
        kind: 'point',
        point: Object.freeze({ x: input.point.x, y: input.point.y }),
      })
      : Object.freeze({ kind: 'result', result: input.result });
    const key = this._motionKey(subjectId, moment);
    this._motionResolutions.set(key, resolution);
    for (const waiter of [...this._motionWaiters]) {
      if (waiter.epoch === this._motionEpoch && waiter.key === key) waiter.settle(resolution);
    }
  }

  private _motionKey(subjectId: string, moment: MotionEffectAnchor['moment']): string {
    return `${moment}\u0000${subjectId}`;
  }

  private _resolveTrail(subjectId: string, resolution: TrailResolution): void {
    const frozen: TrailResolution = resolution.kind === 'event'
      ? Object.freeze({ kind: 'event', event: resolution.event })
      : Object.freeze({ kind: 'result', result: resolution.result });
    this._trailResolutions.set(subjectId, frozen);
    for (const waiter of [...this._trailWaiters]) {
      if (waiter.epoch === this._motionEpoch && waiter.subjectId === subjectId) {
        waiter.settle(frozen);
      }
    }
  }

  private _deferTrail(
    effect: TrailEffectSpec,
    path: string,
    policy: ResolvedPolicy,
  ): InternalHandle {
    const existing = this._trailResolutions.get(effect.subject);
    if (existing?.kind === 'event') return this._startTrail(effect, path, policy, existing.event);
    if (existing?.kind === 'result') return {
      finished: Promise.resolve(existing.result),
      cancel: () => {},
    };
    if (!this._expectsMotionPlan || this._motionPlanSettled) return skipped('missing-subject');

    const epoch = this._motionEpoch;
    let child: InternalHandle | null = null;
    let settled = false;
    let resolveFinished!: (result: EffectResult) => void;
    const finished = new Promise<EffectResult>(resolve => { resolveFinished = resolve; });
    const finish = (result: EffectResult) => {
      if (settled) return;
      settled = true;
      this._trailWaiters.delete(waiter);
      resolveFinished(result);
    };
    const waiter: TrailWaiter = {
      epoch,
      subjectId: effect.subject,
      settle: resolution => {
        if (settled) return;
        this._trailWaiters.delete(waiter);
        if (resolution.kind === 'result') {
          finish(resolution.result);
          return;
        }
        child = this._startTrail(effect, path, policy, resolution.event);
        void child.finished.then(finish);
      },
    };
    this._trailWaiters.add(waiter);
    return {
      finished,
      cancel: () => {
        if (settled) return;
        child?.cancel();
        finish(CANCELLED);
      },
    };
  }

  private _startTrail(
    effect: TrailEffectSpec,
    _path: string,
    policy: ResolvedPolicy,
    event: StructuralMotionSegmentEvent,
  ): InternalHandle {
    const { segment } = event;
    if (!segment.visualSubject) return skipped('missing-subject');
    if (!segment.path || segment.path.kind !== 'travel') return skipped('no-motion-path');
    if (segment.execution.status !== 'started') return skipped('motion-skipped');

    const spatialTiming = segment.execution.animations.find(
      timing => timing.channel === 'host:transform',
    );
    if (!spatialTiming) return skipped('motion-skipped');
    const startDelay = spatialTiming.delayMs;
    const visualEnd = spatialTiming.delayMs
      + spatialTiming.durationMs * spatialTiming.iterations;
    if (!Number.isFinite(startDelay) || !Number.isFinite(visualEnd) || visualEnd <= startDelay) {
      return skipped('no-motion-path');
    }

    const defaults = TRAIL_POLICY[policy.intensity];
    const requestedEchoes = Math.round(finite(effect.advanced?.echoes, defaults.echoes, 1, 10));
    const reservation = reserveEffectBudget(this.ownerDocument, requestedEchoes);
    if (!reservation) return skipped('budget');
    const container = this._container();
    if (!container) {
      reservation.release();
      return skipped('not-connected');
    }

    const lagMs = finite(effect.advanced?.lagMs, defaults.lagMs, 4, 80);
    const opacity = finite(effect.advanced?.opacity, defaults.opacity, 0.05, 0.7);
    const palette = this._palette(policy.tone, effect.advanced?.palette);
    const from = segment.path.from;
    const to = segment.path.to;
    const fromCenter = geometryCenter(from);
    const toCenter = geometryCenter(to);
    const shape = segment.visualSubject.shape;
    const borderRadius = shape === 'circle' ? '999px'
      : shape === 'rounded-rectangle' ? '12%' : '0';
    const primaryEasing = spatialTiming.easing || 'ease-in-out';
    const elements: HTMLElement[] = [];
    const animations: Animation[] = [];

    for (let index = 0; index < reservation.particles; index++) {
      const echoDelay = startDelay + (index + 1) * lagMs;
      const duration = visualEnd - echoDelay;
      if (duration <= 1) continue;
      const element = this.ownerDocument.createElement('i');
      element.className = 'trail-echo';
      element.style.setProperty('--effect-color', palette[index % palette.length]);
      element.style.borderRadius = borderRadius;
      const baseWidth = Math.max(1, Math.abs(from.width));
      const baseHeight = Math.max(1, Math.abs(from.height));
      const scaleX = Math.max(0.01, Math.abs(to.width) / baseWidth);
      const scaleY = Math.max(0.01, Math.abs(to.height) / baseHeight);
      element.style.width = `${baseWidth}px`;
      element.style.height = `${baseHeight}px`;
      container.appendChild(element);
      elements.push(element);
      const echoOpacity = opacity * (1 - index / (reservation.particles + 1));
      animations.push(element.animate([
        {
          transform: `translate(${fromCenter.x}px, ${fromCenter.y}px) translate(-50%, -50%) scale(1, 1)`,
          opacity: 0,
        },
        {
          transform: `translate(${fromCenter.x}px, ${fromCenter.y}px) translate(-50%, -50%) scale(1, 1)`,
          opacity: echoOpacity,
          offset: 0.08,
        },
        {
          transform: `translate(${toCenter.x}px, ${toCenter.y}px) translate(-50%, -50%) scale(${scaleX}, ${scaleY})`,
          opacity: 0,
        },
      ], {
        delay: echoDelay,
        duration,
        easing: primaryEasing,
        fill: 'none',
      }));
    }
    if (animations.length === 0) {
      for (const element of elements) element.remove();
      reservation.release();
      return skipped('no-motion-path');
    }

    const running = this._settlingHandle(animations, elements, reservation.release);
    const key = this._activeTrailKey(event);
    const cancels = this._activeTrailCancels.get(key) ?? new Set<() => void>();
    const cancel = () => running.cancel();
    cancels.add(cancel);
    this._activeTrailCancels.set(key, cancels);
    void running.finished.finally(() => {
      cancels.delete(cancel);
      if (cancels.size === 0) this._activeTrailCancels.delete(key);
    });
    return running;
  }

  private _activeTrailKey(event: StructuralMotionSegmentEvent): string {
    return `${event.source}:${event.generation}:${event.subjectId}`;
  }

  private _atPoint(
    anchor: EffectPointAnchor,
    start: (point: Readonly<{ x: number; y: number }>) => InternalHandle,
  ): InternalHandle {
    if (!(anchor instanceof HTMLElement) && anchor.kind === 'motion') {
      return this._deferMotionPoint(anchor, start);
    }
    const point = this._anchorPoint(anchor);
    return point ? start(point) : skipped('missing-anchor');
  }

  private _deferMotionPoint(
    anchor: MotionEffectAnchor,
    start: (point: Readonly<{ x: number; y: number }>) => InternalHandle,
  ): InternalHandle {
    const key = this._motionKey(anchor.subjectId, anchor.moment);
    const existing = this._motionResolutions.get(key);
    if (existing?.kind === 'point') return start(existing.point);
    if (existing?.kind === 'result') return {
      finished: Promise.resolve(existing.result),
      cancel: () => {},
    };
    if (!this._expectsMotionPlan || this._motionPlanSettled) return skipped('missing-anchor');

    const epoch = this._motionEpoch;
    let child: InternalHandle | null = null;
    let settled = false;
    let resolveFinished!: (result: EffectResult) => void;
    const finished = new Promise<EffectResult>(resolve => { resolveFinished = resolve; });
    const finish = (result: EffectResult) => {
      if (settled) return;
      settled = true;
      this._motionWaiters.delete(waiter);
      resolveFinished(result);
    };
    const waiter: MotionWaiter = {
      epoch,
      key,
      settle: resolution => {
        if (settled) return;
        this._motionWaiters.delete(waiter);
        if (resolution.kind === 'result') {
          finish(resolution.result);
          return;
        }
        child = start(resolution.point);
        void child.finished.then(finish);
      },
    };
    this._motionWaiters.add(waiter);
    return {
      finished,
      cancel: () => {
        if (settled) return;
        child?.cancel();
        finish(CANCELLED);
      },
    };
  }

  private _anchorPoint(anchor: EffectAnchor): { x: number; y: number } | null {
    if (anchor instanceof HTMLElement) {
      if (!anchor.isConnected) return null;
      const { x, y } = geometryCenter(captureViewportGeometry(anchor));
      return { x, y };
    }
    if (anchor.kind === 'point') return { x: anchor.x, y: anchor.y };
    const root = this._configuration.anchorRoot;
    if (!root) return null;
    const escaped = CSS.escape(anchor.name);
    const element = root.querySelector<HTMLElement>(
      `[data-effect-anchor="${escaped}"], #${escaped}`,
    );
    if (element?.isConnected) {
      const { x, y } = geometryCenter(captureViewportGeometry(element));
      return { x, y };
    }
    return this._beforeAnchors.get(anchor.name) ?? null;
  }
}

customElements.define('boardgame-effect-layer', BoardgameEffectLayer);
