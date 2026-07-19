import { LitElement, html, TemplateResult } from 'lit';
import { query } from 'lit/decorators.js';
import './boardgame-component-stack.js';
import type { BoardgameComponentStack } from './boardgame-component-stack.js';
import { animHooks } from '../utils/anim-test-hooks.js';
import {
  finiteTimingMs,
  resolveMotionTiming,
} from '../motion/timing.js';
import type {
  AnimationTimingPolicy,
  VersionAnimationContext,
} from '../motion/timing.js';
import {
  captureOffsetGeometry,
  captureViewportGeometry,
  composeFlipTransform,
  solveFlipGeometry,
} from '../motion/geometry.js';
import type { OffsetGeometry, ViewportGeometry } from '../motion/geometry.js';
import {
  createStructuralMotionDraft,
  publishStructuralMotionPlan,
  updateStructuralMotionExecutions,
} from '../motion/structural-plan.js';
import type {
  StructuralExecutedTiming,
  StructuralExecution,
  StructuralMotionDraft,
  StructuralMotionPlan,
  StructuralMotionSegmentRef,
  StructuralProvenance,
} from '../motion/structural-plan.js';
import { compileStructuralMotionEvents } from '../motion/structural-events.js';
import type { StructuralMotionEvent } from '../motion/structural-events.js';
import { sanitizeMotionSubjectSnapshot } from '../motion/subject.js';
import type { MotionSubjectSnapshot } from '../motion/subject.js';
import type {
  ComponentMotionChannel,
  ComponentMotionTrack,
  BaseComponentMotionInput,
} from '../motion/component-track.js';
import { compileMotionCohortSchedule } from '../motion/cohort.js';
import type { MotionStaggerCohortSpec } from '../motion/cohort.js';
import {
  MotionActivationMonitor,
  primaryStructuralAnimationIndex,
} from '../motion/activation.js';
import {
  compileComponentMotionTracks,
  componentMotionChannel,
  componentMotionKeyframes,
  componentMotionTracks,
} from '../motion/component-track.js';
import { compileViewportFlight } from '../motion/flight.js';
import { resolveStructuralContinuity } from '../motion/continuity.js';
import type {
  MotionExactSighting,
  MotionCollectionHistory,
  MotionContinuityResolution,
} from '../motion/continuity.js';
import {
  captureHistoricalPresentation,
  installHistoricalPresentation,
} from '../motion/historical-presentation.js';
import type { HistoricalPresentation } from '../motion/historical-presentation.js';
import { motionPresenceHostStyle } from '../motion/presence.js';
import type { CompiledMotionTransferDeclaration } from '../motion/transfer.js';

export type { AnimationTimingPolicy } from '../motion/timing.js';

export interface AnimateBetweenOptions {
  /** Defaults to the installed version's companion slot. */
  timing?: AnimationTimingPolicy;
}

export interface MotionFlightRequest {
  /** Semantic presentation subject; never inferred from a DOM id. */
  subjectId: string;
  /** Geometry-only visual origin. */
  source: string | HTMLElement;
  /** Animated host, already rendered at its natural destination. */
  carrier: string | HTMLElement;
  durationMs?: number;
  timing?: AnimationTimingPolicy;
}

export interface ComponentAnimatorAPI {
  installMotionTransfers(transfers: readonly CompiledMotionTransferDeclaration[]): void;
  fly(request: MotionFlightRequest): Promise<void>;
  animateBetween(
    realId: string | HTMLElement,
    stubId: string | HTMLElement,
    durationMs?: number,
    opts?: AnimateBetweenOptions,
  ): Promise<void>;
  /** Ordered, replayable observation surface; does not confer animation ownership. */
  observeStructuralMotionEvents(observer: (event: StructuralMotionEvent) => void): () => void;
}

interface ComponentRecord {
  beforeCollectionIds?: string[];
  sourceTagName?: string;
  historicalPresentation?: HistoricalPresentation;
  offsets?: OffsetGeometry;
  newOffsets?: OffsetGeometry;
  viewportOffsets?: ViewportGeometry;
  newViewportOffsets?: ViewportGeometry;
  before?: Record<string, any>;
  after?: Record<string, any>;
  beforeTransform?: string;
  beforeInlineTransform?: string;
  afterTransform?: string;
  afterOpacity?: string;
  invertedTransform?: string;
  beforeOpacity?: string;
  needsHostTransition?: boolean;
  needsAnimation?: boolean;
  motionTracks?: readonly ComponentMotionTrack[];
  motionDraft?: StructuralMotionDraft;
  visualSubject?: MotionSubjectSnapshot;
}

interface AnimatingComponentRecord {
  subjectId: string;
  stack: any;
  component: any;
  before: Record<string, any>;
  after: Record<string, any>;
  afterTransform: string;
  afterOpacity: string;
  invertedTransform: string;
  beforeOpacity: string;
  needsHostTransition: boolean;
  motionTracks?: readonly ComponentMotionTrack[];
  motionDraft?: StructuralMotionDraft;
}

export class BoardgameComponentAnimator extends LitElement {
  // Supplied by boardgame-render-game for the currently installing version.
  // Public so the wrapper can set it, but game renderers normally need no
  // timing plumbing: fly consumes it by default.
  animationContext: VersionAnimationContext | null = null;
  // Note: Can't use @query decorator because 'animate' method conflicts with Element.animate()
  private get stackElement(): BoardgameComponentStack {
    return this.shadowRoot!.querySelector('#stack')!;
  }

  private _infoById: { [id: string]: ComponentRecord } = {};
  private _beforeSeenIds = new Set<string>();
  private _animatingComponents: AnimatingComponentRecord[] = [];
  private _beforeCollectionOffsets = new Map<string, OffsetGeometry>();
  private _beforeCollectionViewportOffsets = new Map<string, ViewportGeometry>();
  private _generation = 0;
  private _explicitMotionSequence = 0;
  // Published only after the generation's final updateComplete check and
  // immediately before playback. A new prepare() invalidates it synchronously.
  private _solvedMotionPlan: StructuralMotionPlan | null = null;
  private _explicitMotionPlans = new Map<number, StructuralMotionPlan>();
  private _lastExplicitMotionPlan: StructuralMotionPlan | null = null;
  private _motionEventObservers = new Set<(event: StructuralMotionEvent) => void>();
  private _motionEventRevisions = new Map<string, StructuralMotionPlan>();
  private _motionEventHistory = new Map<string, readonly StructuralMotionEvent[]>();
  private _motionCohorts: Readonly<{
    generation: number;
    specs: readonly MotionStaggerCohortSpec[];
  }> | null = null;
  private _motionTransfers: Readonly<{
    generation: number;
    declarations: readonly CompiledMotionTransferDeclaration[];
  }> | null = null;
  private readonly _activationMonitor = new MotionActivationMonitor();
  private readonly _explicitAnimations = new Map<Animation, HTMLElement>();
  private readonly _carrierFlights = new WeakMap<HTMLElement, Animation>();

  ancestorOffsetParent: HTMLElement | null = null;

  override disconnectedCallback(): void {
    this._interruptExplicitMotion();
    this._activationMonitor.clear();
    super.disconnectedCallback();
  }

  observeStructuralMotionEvents(observer: (event: StructuralMotionEvent) => void): () => void {
    this._motionEventObservers.add(observer);
    for (const history of this._motionEventHistory.values()) {
      for (const event of history) {
        try {
          observer(event);
        } catch (error) {
          console.error('[animator] structural motion event observer failed:', error);
        }
      }
    }
    return () => this._motionEventObservers.delete(observer);
  }

  private _notifyStructuralMotion(plan: StructuralMotionPlan): void {
    const historyKey = `${plan.source}:${plan.generation}`;
    const previous = this._motionEventRevisions.get(historyKey) ?? null;
    const events = compileStructuralMotionEvents(previous, plan);
    this._motionEventRevisions.set(historyKey, plan);
    const continuing = previous?.generation === plan.generation;
    const history = continuing
      ? [...(this._motionEventHistory.get(historyKey) ?? []), ...events]
      : [...events];
    this._motionEventHistory.set(historyKey, Object.freeze(history));
    while (this._motionEventHistory.size > 64) {
      const oldest = this._motionEventHistory.keys().next().value as string | undefined;
      if (oldest === undefined) break;
      this._motionEventHistory.delete(oldest);
      this._motionEventRevisions.delete(oldest);
    }
    for (const event of events) {
      for (const observer of this._motionEventObservers) {
        try {
          observer(event);
        } catch (error) {
          console.error('[animator] structural motion event observer failed:', error);
        }
      }
    }
  }

  private _setSolvedMotionPlan(plan: StructuralMotionPlan): void {
    this._solvedMotionPlan = plan;
    this._notifyStructuralMotion(plan);
  }

  private _segmentId(ref: StructuralMotionSegmentRef): string {
    return `${ref.source}:${ref.generation}:${ref.segmentIndex}`;
  }

  private _updateFlipMotion(
    generation: number,
    segmentIndex: number,
    execution: StructuralExecution,
  ): void {
    if (this._generation !== generation || !this._solvedMotionPlan
      || this._solvedMotionPlan.generation !== generation) return;
    this._setSolvedMotionPlan(updateStructuralMotionExecutions(
      this._solvedMotionPlan,
      new Map([[segmentIndex, execution]]),
    ));
  }

  private _invalidateSolvedMotionPlan(): void {
    const plan = this._solvedMotionPlan;
    if (!plan || plan.phase === 'settled') return;
    const updates = new Map<number, StructuralExecution>();
    for (const [segmentIndex, segment] of plan.segments.entries()) {
      if (segment.execution.status === 'armed'
        || segment.execution.status === 'active-observed') {
        updates.set(segmentIndex, {
          status: 'cancelled',
          animations: segment.execution.animations,
        });
      } else if (segment.execution.status === 'planned') {
        updates.set(segmentIndex, {
          status: 'skipped',
          reason: 'not-started',
        });
      }
    }
    this._notifyStructuralMotion(updateStructuralMotionExecutions(plan, updates));
  }

  private _interruptExplicitMotion(): void {
    for (const animation of this._explicitAnimations.keys()) {
      try { animation.cancel(); } catch { /* already terminal */ }
    }
    this._explicitAnimations.clear();
    for (const [generation, plan] of this._explicitMotionPlans) {
      if (plan.phase === 'settled') continue;
      const updates = new Map<number, StructuralExecution>();
      for (const [segmentIndex, segment] of plan.segments.entries()) {
        if (segment.execution.status === 'armed'
          || segment.execution.status === 'active-observed') {
          updates.set(segmentIndex, {
            status: 'cancelled',
            animations: segment.execution.animations,
          });
        } else if (segment.execution.status === 'planned') {
          updates.set(segmentIndex, { status: 'skipped', reason: 'superseded' });
        }
      }
      const updated = updateStructuralMotionExecutions(plan, updates);
      this._explicitMotionPlans.set(generation, updated);
      if (this._lastExplicitMotionPlan?.generation === generation) {
        this._lastExplicitMotionPlan = updated;
      }
      this._notifyStructuralMotion(updated);
    }
  }

  /** Clear interrupted faux components without exposing the stack registry. */
  clearAnimatingComponents(): void {
    for (const stack of this.stackElement._sharedStackList) {
      stack.clearAnimatingComponents();
    }
  }

  /** Install author timing for the generation opened by the latest prepare(). */
  installMotionCohorts(specs: readonly MotionStaggerCohortSpec[]): void {
    this._motionCohorts = Object.freeze({
      generation: this._generation,
      specs: Object.freeze([...specs]),
    });
  }

  /** Install an already atomically validated transfer batch for this cycle. */
  installMotionTransfers(declarations: readonly CompiledMotionTransferDeclaration[]): void {
    this._motionTransfers = Object.freeze({
      generation: this._generation,
      declarations: Object.freeze([...declarations]),
    });
  }

  prepare() {
    this._invalidateSolvedMotionPlan();
    this._interruptExplicitMotion();
    this._activationMonitor.clear();
    this._generation++;
    this._solvedMotionPlan = null;
    this._motionCohorts = null;
    this._motionTransfers = null;
    const collections = this.stackElement._sharedStackList;

    this._beforeCollectionOffsets = new Map();
    this._beforeCollectionViewportOffsets = new Map();

    const result: { [id: string]: ComponentRecord } = {};

    // keep track of all of the ids we've seen this round to make sure we
    // found a home for all of them in the end.
    this._beforeSeenIds = new Set();

    // Interruption semantics (spec): a new cycle must measure resting
    // positions, so jump any still-live animations to their end state.
    for (let i = 0; i < collections.length; i++) {
      const components = collections[i].Components;
      for (let j = 0; j < components.length; j++) {
        const c = components[j];
        c.animationContext = this.animationContext;
        if (typeof c.finishAllAnimations === 'function') c.finishAllAnimations();
      }
    }

    for (let i = 0; i < collections.length; i++) {
      const collection = collections[i];

      const offsetComponent = collection.offsetComponent;
      this._beforeCollectionOffsets.set(
        collection.id,
        captureOffsetGeometry(offsetComponent, this.ancestorOffsetParent),
      );
      this._beforeCollectionViewportOffsets.set(
        collection.id,
        captureViewportGeometry(offsetComponent),
      );

      const components = collection.Components;
      for (let j = 0; j < components.length; j++) {
        const component = components[j];

        // Skip components without ids (e.g. faux-components, spacer components).
        if (component.id === '') continue;

        const record = result[component.id] || {};

        record.beforeCollectionIds ??= [];
        record.beforeCollectionIds.push(collection.id);
        record.sourceTagName = component.localName;

        this._beforeSeenIds.add(component.id);

        record.offsets = captureOffsetGeometry(component, this.ancestorOffsetParent);
        record.viewportOffsets = captureViewportGeometry(component);
        record.visualSubject = this._captureMotionSubject(component);

        // We use getComputedStyle instead of just card.style.transform,
        // because if the card is in the middle of transforming, we want
        // the exact value at that second, not what the logical final value
        // is.

        const computedStyle = getComputedStyle(component);

        record.beforeTransform = computedStyle.transform;

        if (record.beforeTransform === 'none') {
          record.beforeTransform = '';
        }

        record.before = component.animatingPropValues();
        record.beforeInlineTransform = component.style.transform;
        record.beforeOpacity = component.style.opacity || '1';

        try {
          record.historicalPresentation = captureHistoricalPresentation(component) ?? undefined;
        } catch (error) {
          console.error('[animator] historical presentation capture failed:', error);
          record.historicalPresentation = undefined;
        }
        result[component.id] = record;
      }
    }

    this._infoById = result;
  }

  /**
   * _resolveAnimationTarget turns a flight endpoint into a live
   * element. Elements pass through. String ids are resolved against the
   * document first, then against the root node (usually a renderer's
   * shadow root) of every registered component stack — renderers are Lit
   * elements, so their cards, fake-deck stubs, and edge anchors all live
   * inside shadow DOM where document.getElementById can't see them. The
   * shared stack registry is the one thing that already spans those
   * boundaries, so we borrow its roots.
   */
  private _resolveAnimationTarget(target: string | HTMLElement): HTMLElement | null {
    if (typeof target !== 'string') return target;
    const direct = document.getElementById(target);
    if (direct) return direct;
    const stacks = this.stackElement?._sharedStackList ?? [];
    const seenRoots = new Set<Node>();
    for (const stack of stacks) {
      const root = stack.getRootNode();
      if (seenRoots.has(root)) continue;
      seenRoots.add(root);
      const found = (root as Document | ShadowRoot).getElementById?.(target)
        ?? (root as Document | ShadowRoot).querySelector?.(`#${CSS.escape(target)}`);
      if (found) return found as HTMLElement;
    }
    return null;
  }

  /** Resolve exactly one endpoint inside roots registered to this animator. */
  private _resolveScopedTransferTarget(id: string): HTMLElement | null {
    const roots = new Set<Node>();
    const matches: HTMLElement[] = [];
    for (const stack of this.stackElement?._sharedStackList ?? []) {
      const root = stack.getRootNode();
      if (roots.has(root)) continue;
      roots.add(root);
      for (const match of (root as Document | ShadowRoot).querySelectorAll?.(
        `#${CSS.escape(id)}`,
      ) ?? []) {
        if (match instanceof HTMLElement) matches.push(match);
      }
    }
    return matches.length === 1 ? matches[0] : null;
  }

  private _recordAnimationActive(
    anim: Animation,
    delay: number,
    detail: string,
    context: VersionAnimationContext | null,
  ) {
    const observe = () => {
      const currentTime = anim.currentTime;
      if (typeof currentTime === 'number' && currentTime + 0.5 >= delay) {
        animHooks.record('active', detail, context ? {
          version: context.version,
          targetAtMs: context.startAtMs,
        } : undefined);
        return;
      }
      if (anim.playState === 'idle' || anim.playState === 'finished') return;
      requestAnimationFrame(observe);
    };
    requestAnimationFrame(observe);
  }

  private _executedTiming(
    animation: Animation,
    channel: ComponentMotionChannel,
  ): StructuralExecutedTiming {
    const timing = animation.effect instanceof KeyframeEffect
      ? animation.effect.getTiming()
      : {};
    const iterations = timing.iterations === undefined
      ? 1
      : Math.max(0, finiteTimingMs(timing.iterations));
    return Object.freeze({
      channel,
      delayMs: finiteTimingMs(timing.delay),
      durationMs: finiteTimingMs(timing.duration),
      endDelayMs: finiteTimingMs(timing.endDelay),
      iterations,
      easing: timing.easing ?? 'linear',
      fill: timing.fill ?? 'none',
    });
  }

  private _captureMotionSubject(component: unknown): MotionSubjectSnapshot | undefined {
    try {
      const provider = component as { motionSubjectSnapshot?: () => unknown };
      if (typeof provider.motionSubjectSnapshot !== 'function') return undefined;
      return sanitizeMotionSubjectSnapshot(provider.motionSubjectSnapshot()) ?? undefined;
    } catch (error) {
      // Decoration capability failures are isolated from queue-critical motion.
      console.error('[animator] motion subject snapshot failed:', error);
      return undefined;
    }
  }

  private _planMotionTracks(
    component: { planMotionTracks?: (input: BaseComponentMotionInput & {
      before: Record<string, unknown>;
      after: Record<string, unknown>;
    }) => readonly ComponentMotionTrack[] },
    input: BaseComponentMotionInput & {
      before: Record<string, unknown>;
      after: Record<string, unknown>;
    },
  ): readonly ComponentMotionTrack[] {
    try {
      if (typeof component.planMotionTracks !== 'function') {
        return compileComponentMotionTracks(input);
      }
      return componentMotionTracks(component.planMotionTracks(input));
    } catch (error) {
      // A component-owned visual hook must never strand queue-critical host
      // continuity. Preserve safe structural transform/opacity tracks only.
      console.error('[animator] component motion planning failed:', error);
      return compileComponentMotionTracks({
        needsHostTransition: input.needsHostTransition,
        invertedTransform: input.invertedTransform,
        finalTransform: input.finalTransform,
        beforeOpacity: input.beforeOpacity,
        finalOpacity: input.finalOpacity,
      });
    }
  }

  private _restoreNoAnimateBarrier(): void {
    for (const collection of this.stackElement._sharedStackList) {
      collection.noAnimate = false;
      for (const component of collection.Components) component.noAnimate = false;
    }
    for (const record of this._animatingComponents) record.component.noAnimate = false;
  }

  private _abortAnimationCycle(
    error: unknown,
    resolve: (settled: Promise<void>) => void,
  ): void {
    console.error('[animator] structural motion cycle failed:', error);
    try {
      this._restoreNoAnimateBarrier();
    } catch (cleanupError) {
      console.error('[animator] failed to restore animation barrier:', cleanupError);
    }
    try {
      this._invalidateSolvedMotionPlan();
      this._solvedMotionPlan = null;
    } catch (cleanupError) {
      console.error('[animator] failed to invalidate motion plan:', cleanupError);
    }
    try {
      this.clearAnimatingComponents();
    } catch (cleanupError) {
      console.error('[animator] failed to clear animation clones:', cleanupError);
    } finally {
      resolve(Promise.resolve());
    }
  }

  private _installExplicitMotion(plan: StructuralMotionPlan): void {
    this._explicitMotionPlans.set(plan.generation, plan);
    this._lastExplicitMotionPlan = plan;
    this._notifyStructuralMotion(plan);
    while (this._explicitMotionPlans.size > 32) {
      const oldest = this._explicitMotionPlans.keys().next().value as number | undefined;
      if (oldest === undefined) break;
      this._explicitMotionPlans.delete(oldest);
    }
  }

  private _updateExplicitMotion(
    generation: number,
    segmentIndex: number,
    execution: StructuralExecution,
  ): void {
    const current = this._explicitMotionPlans.get(generation);
    if (!current) return;
    const updated = updateStructuralMotionExecutions(
      current,
      new Map([[segmentIndex, execution]]),
    );
    this._explicitMotionPlans.set(generation, updated);
    if (this._lastExplicitMotionPlan?.generation === generation) {
      this._lastExplicitMotionPlan = updated;
    }
    this._notifyStructuralMotion(updated);
  }

  private _trackExplicitAnimation(
    generation: number,
    subjectId: string,
    animation: Animation | null,
    carrier?: HTMLElement,
    segmentIndex = 0,
  ): void {
    if (!animation) {
      this._updateExplicitMotion(generation, segmentIndex, {
        status: 'skipped',
        reason: 'not-started',
      });
      return;
    }
    const animations = Object.freeze([
      this._executedTiming(animation, 'host:transform'),
    ]);
    if (carrier) {
      const previous = this._carrierFlights.get(carrier);
      if (previous && previous !== animation) {
        try { previous.cancel(); } catch { /* already terminal */ }
      }
      this._carrierFlights.set(carrier, animation);
      this._explicitAnimations.set(animation, carrier);
      void animation.finished.catch(() => {}).finally(() => {
        this._explicitAnimations.delete(animation);
        if (this._carrierFlights.get(carrier) === animation) {
          this._carrierFlights.delete(carrier);
        }
      });
    }
    this._updateExplicitMotion(generation, segmentIndex, {
      status: 'armed',
      animations,
    });
    this._activationMonitor.observe(
      `explicit:${generation}:0`,
      animation,
      animations[0].delayMs,
      () => this._updateExplicitMotion(generation, segmentIndex, {
        status: 'active-observed', animations,
      }),
    );
    void animation.finished.then(
      () => {
        this._activationMonitor.cancel(`explicit:${generation}:0`);
        this._updateExplicitMotion(generation, segmentIndex, {
          status: 'active-observed', animations,
        });
        this._updateExplicitMotion(generation, segmentIndex, { status: 'finished', animations });
      },
      () => {
        this._activationMonitor.cancel(`explicit:${generation}:0`);
        this._updateExplicitMotion(generation, segmentIndex, { status: 'cancelled', animations });
      },
    );
  }

  private async _awaitRegisteredTargets(generation: number): Promise<boolean> {
    await this.updateComplete;
    if (this._generation !== generation) return false;
    const stacks = [...(this.stackElement?._sharedStackList ?? [])];
    await Promise.all(stacks.map(stack => stack.updateComplete));
    await Promise.resolve();
    return this._generation === generation;
  }

  /** Fly one retained carrier from source geometry to its natural position. */
  async fly(request: MotionFlightRequest): Promise<void> {
    const subjectId = request.subjectId.trim();
    if (!subjectId) throw new Error('motion flight subjectId must be nonempty');
    const durationMs = request.durationMs ?? 500;
    if (!Number.isFinite(durationMs) || durationMs < 0) {
      throw new Error('motion flight durationMs must be finite and nonnegative');
    }
    const explicitGeneration = ++this._explicitMotionSequence;
    const structuralGeneration = this._generation;
    const timing = request.timing ?? 'version';
    if (typeof request.source === 'string' || typeof request.carrier === 'string') {
      if (!await this._awaitRegisteredTargets(structuralGeneration)) {
        const draft = createStructuralMotionDraft({
          subjectId,
          presence: 'retained',
          provenance: { kind: 'unresolved', endpoint: 'destination' },
        });
        this._installExplicitMotion(publishStructuralMotionPlan(explicitGeneration, [{
          draft,
          timingRequest: { policy: timing, delayMs: 0, durationMs },
        }], 'explicit'));
        this._updateExplicitMotion(explicitGeneration, 0, {
          status: 'skipped',
          reason: 'superseded',
        });
        return;
      }
    }
    const carrier = this._resolveAnimationTarget(request.carrier);
    const source = this._resolveAnimationTarget(request.source);
    if (!carrier || !source || !carrier.isConnected || !source.isConnected) {
      const draft = createStructuralMotionDraft({
        subjectId,
        presence: 'retained',
        provenance: {
          kind: 'unresolved',
          endpoint: !source || !source.isConnected ? 'source' : 'destination',
        },
      });
      this._installExplicitMotion(publishStructuralMotionPlan(explicitGeneration, [{
        draft,
        timingRequest: { policy: timing, delayMs: 0, durationMs },
      }], 'explicit'));
      this._updateExplicitMotion(explicitGeneration, 0, {
        status: 'skipped',
        reason: 'missing-endpoint',
      });
      // Loud on purpose: an unresolvable endpoint silently killed the
      // entire cross-screen animation feature once already (the id
      // property wasn't reflected to the DOM, so no card ever matched).
      console.warn('[animator] fly: could not resolve',
        !source || !source.isConnected ? request.source : request.carrier,
        '— animation skipped');
      return;
    }
    // Measure both endpoints in the SAME coordinate space (viewport rects)
    // and align CENTERS. The previous offsetParent-chain math mixed
    // coordinate spaces when the two elements had different fixed/absolute
    // ancestors, and corner-alignment launched flights from the top-left
    // of full-width containers instead of from the visual source.
    const carrierRect = captureViewportGeometry(carrier);
    const sourceRect = captureViewportGeometry(source);
    const restingTransform = getComputedStyle(carrier).transform || 'none';
    const compiled = compileViewportFlight(sourceRect, carrierRect, restingTransform);
    const { translateX: dx, translateY: dy } = compiled.inversion;
    const draft = createStructuralMotionDraft({
      subjectId,
      presence: 'retained',
      provenance: { kind: 'identity' },
      viewportFrom: sourceRect,
      viewportTo: carrierRect,
      inversion: compiled.inversion,
      channels: [{ target: 'host', property: 'transform' }],
    });
    this._installExplicitMotion(publishStructuralMotionPlan(explicitGeneration, [{
      draft,
      timingRequest: { policy: timing, delayMs: 0, durationMs },
    }], 'explicit'));
    if (!compiled.inversion.changed) {
      this._updateExplicitMotion(explicitGeneration, 0, {
        status: 'skipped',
        reason: 'no-spatial-change',
      });
      return;
    }
    const resolution = resolveMotionTiming(
      { duration: durationMs, easing: 'ease-out', fill: 'none' },
      {
        policy: timing,
        context: timing === 'version' ? this.animationContext : null,
        reducedMotion: window.matchMedia('(prefers-reduced-motion: reduce)').matches,
      },
    );
    if (resolution.kind === 'skip') {
      this._updateExplicitMotion(explicitGeneration, 0, {
        status: 'skipped',
        reason: 'timing',
      });
      return;
    }
    if (resolution.activeContext
      && finiteTimingMs(resolution.timing.duration) < durationMs) {
      console.warn(
        `[animator] synchronized animation requested ${durationMs}ms; ` +
        `capping to the version slot's ` +
        `${resolution.activeContext.maxAnimationDurationMs}ms contract. ` +
        `Use { timing: 'immediate' } for a longer local-only effect.`,
      );
    }
    const delay = finiteTimingMs(resolution.timing.delay);
    const keyframes = [...componentMotionKeyframes(compiled.tracks[0])] as Keyframe[];
    const carrierTag = carrier.tagName.toLowerCase() + (carrier.id ? `#${carrier.id}` : '');

    // When the flight target is a real animatable item (a boardgame-card /
    // -component), route through its play() so the flight is GATED: it
    // registers a will-animate/animation-done pair and joins the item's
    // live set. This closes two holes that raw real.animate() left open
    // (#798): (a) the completion gate could close — and the game-over
    // verdict banner appear — while a delayed synced deal flight was still
    // mid-air; (b) prepare()'s interruption
    // pass (finishAllAnimations) couldn't reach a raw flight to settle it
    // before measuring a new cycle. play() supplies its own default timing
    // (duration = --animation-length), so we override with the caller's
    // durationMs + the sync delay and match the raw path's ease-out/none.
    // Plain elements (e.g. the divs in explicit-flight browser tests
    // test uses) have no play() and keep the raw-element fallback below.
    if (typeof (carrier as any).play === 'function') {
      const anim = (carrier as any).play(carrier, keyframes,
        resolution.timing,
        { timing: 'immediate' }, { recordActive: false });
      this._trackExplicitAnimation(explicitGeneration, subjectId, anim, carrier);
      // play() returns null under noAnimate; nothing is in flight then.
      if (anim) {
        this._recordAnimationActive(
          anim,
          delay,
          'fly:' + carrierTag,
          resolution.activeContext,
        );
        await anim.finished.catch(() => {});
        // The Animation promise and play()'s settlement bookkeeping are
        // separate promise reactions. Do not resolve fly until
        // the latter has closed the render-game completion gate too.
        if (typeof (carrier as any).settled === 'function') {
          await (carrier as any).settled();
        }
      }
      return;
    }

    const anim = carrier.animate(keyframes, resolution.timing);
    this._trackExplicitAnimation(explicitGeneration, subjectId, anim, carrier);
    this._recordAnimationActive(
      anim,
      delay,
      'fly:' + carrierTag,
      resolution.activeContext,
    );
    // Settlement is ground truth: finished resolves on completion, rejects
    // on cancel (element removed mid-flight) — both mean "done" here.
    await anim.finished.catch(() => {});
  }

  /** @deprecated Prefer fly(), whose source/carrier direction is explicit. */
  animateBetween(
    realId: string | HTMLElement,
    stubId: string | HTMLElement,
    durationMs: number = 500,
    opts?: AnimateBetweenOptions,
  ): Promise<void> {
    const subjectId = typeof realId === 'string'
      ? realId
      : realId.id || `explicit-motion-${this._explicitMotionSequence + 1}`;
    return this.fly({
      subjectId,
      source: stubId,
      carrier: realId,
      durationMs,
      timing: opts?.timing,
    });
  }

  private _startInstalledMotionTransfers(generation: number): Promise<void> {
    const installed = this._motionTransfers;
    if (!installed || installed.generation !== generation
      || installed.declarations.length === 0) return Promise.resolve();
    const explicitGeneration = ++this._explicitMotionSequence;
    const collections = this.stackElement._sharedStackList;
    const structuralCarriers = new Set<HTMLElement>();
    for (const collection of collections) {
      for (const component of collection.Components) structuralCarriers.add(component);
    }
    const compiled = installed.declarations.map(declaration => {
      const source = this._resolveScopedTransferTarget(declaration.source);
      const carrier = this._resolveScopedTransferTarget(declaration.carrier);
      const ownershipConflict = !!carrier && structuralCarriers.has(carrier);
      if (!source || !carrier || !source.isConnected || !carrier.isConnected || ownershipConflict) {
        return { declaration, source, carrier, ownershipConflict, flight: null };
      }
      const sourceRect = captureViewportGeometry(source);
      const carrierRect = captureViewportGeometry(carrier);
      const restingTransform = getComputedStyle(carrier).transform || 'none';
      return {
        declaration,
        source,
        carrier,
        ownershipConflict,
        flight: compileViewportFlight(sourceRect, carrierRect, restingTransform),
      };
    });
    const entries = compiled.map(({ declaration, source, carrier, flight }) => ({
      draft: createStructuralMotionDraft({
        subjectId: declaration.subjectId,
        declarationKey: declaration.key,
        presence: 'retained',
        provenance: source && carrier
          ? { kind: 'identity' }
          : { kind: 'unresolved', endpoint: !source ? 'source' : 'destination' },
        ...(flight ? {
          viewportFrom: flight.source,
          viewportTo: flight.destination,
          inversion: flight.inversion,
          channels: flight.tracks,
        } : {}),
      }),
      timingRequest: {
        policy: declaration.timing,
        delayMs: 0,
        durationMs: declaration.durationMs,
      },
    }));
    this._installExplicitMotion(publishStructuralMotionPlan(
      explicitGeneration,
      entries,
      'explicit',
    ));

    const settled: Promise<unknown>[] = [];
    for (const [segmentIndex, item] of compiled.entries()) {
      const { declaration, carrier, flight, ownershipConflict } = item;
      if (ownershipConflict) {
        this._updateExplicitMotion(explicitGeneration, segmentIndex, {
          status: 'skipped', reason: 'ownership-conflict',
        });
        continue;
      }
      if (!carrier || !flight) {
        this._updateExplicitMotion(explicitGeneration, segmentIndex, {
          status: 'skipped', reason: 'missing-endpoint',
        });
        continue;
      }
      if (!flight.inversion.changed) {
        this._updateExplicitMotion(explicitGeneration, segmentIndex, {
          status: 'skipped', reason: 'no-spatial-change',
        });
        continue;
      }
      const resolution = resolveMotionTiming(
        { duration: declaration.durationMs, easing: 'ease-out', fill: 'none' },
        {
          policy: declaration.timing,
          context: declaration.timing === 'version' ? this.animationContext : null,
          reducedMotion: window.matchMedia('(prefers-reduced-motion: reduce)').matches,
        },
      );
      if (resolution.kind === 'skip') {
        this._updateExplicitMotion(explicitGeneration, segmentIndex, {
          status: 'skipped', reason: 'timing',
        });
        continue;
      }
      const animation = carrier.animate(
        [...componentMotionKeyframes(flight.tracks[0])] as Keyframe[],
        resolution.timing,
      );
      this._trackExplicitAnimation(
        explicitGeneration,
        declaration.subjectId,
        animation,
        carrier,
        segmentIndex,
      );
      this._recordAnimationActive(
        animation,
        finiteTimingMs(resolution.timing.delay),
        `transfer:${declaration.key}`,
        resolution.activeContext,
      );
      settled.push(animation.finished.catch(() => undefined));
    }
    return Promise.all(settled).then(() => undefined);
  }

  // CRITICAL: Double microtask delay for Polymer databinding completion
  // animateFlip returns a promise that is resolved once all animations that will
  // be started are started.
  // Note: Can't use 'animate' as method name due to conflict with Element.animate()
  animateFlip(): Promise<void> {
    // Wait for the style to be set--but BEFORE a frame is rendered!
    // Originally, on Chrome, requestAnimationFrame happens right before this--
    // but microTask timing isn't sufficiently late.

    // On Safari, requestAnimationFrame is already too late, and you'll see a
    // visual glitch if you wait until then. As of October 18, Chrome seems to
    // now have the Safari behavior, so just doing that.

    // The inner promise resolves with a Promise<void> that means "everything
    // SETTLED" — the WAAPI PLAY phase hands that up through _startAnimations.
    // Flattening it means the promise animateFlip() returns now completes at
    // real animation settlement, not merely when animations were started.
    const generation = this._generation;

    return new Promise<Promise<void>>((resolve) => {
      // CRITICAL: First microtask - Let Polymer dispatch change events
      Promise.resolve().then(() => {
        if (this._generation !== generation) { resolve(Promise.resolve()); return; }
        this._scheduleAnimate(resolve, generation);
      });
    }).then((settled) => settled);
  }

  private _scheduleAnimate(resolve: (p: Promise<void>) => void, generation: number) {
    // CRITICAL: Second microtask - Ensure ALL databinding cascades complete
    // This bizarre indirection is necessary because by the time the first
    // microtask resolves some databinding won't have been done, so we need to
    // one more time wait until the end of the microtask. See #722 for more.
    Promise.resolve().then(() => {
      if (this._generation !== generation) { resolve(Promise.resolve()); return; }
      try {
        this._doAnimate(resolve, generation);
      } catch (error) {
        this._abortAnimationCycle(error, resolve);
      }
    });
  }

  private _doAnimate(resolve: (p: Promise<void>) => void, generation: number) {
    const collections = this.stackElement._sharedStackList;

    const beforeExact: MotionExactSighting[] = [];
    for (const [subjectId, record] of Object.entries(this._infoById)) {
      for (const collectionId of record.beforeCollectionIds ?? []) {
        beforeExact.push({ subjectId, collectionId });
      }
    }
    const afterExact: MotionExactSighting[] = [];
    const histories: MotionCollectionHistory[] = collections.map(collection => ({
      collectionId: collection.id,
      lastSeen: Object.freeze({ ...(collection.idsLastSeen ?? {}) }),
    }));
    const stackById = new Map(collections.map(collection => [collection.id, collection]));

    const collectionOffsets = new Map<string, OffsetGeometry>();
    const collectionViewportOffsets = new Map<string, ViewportGeometry>();

    // CRITICAL: noAnimate barrier during measurement phase
    // Turning off animations and setting card flip all require recalcing
    // style so do them once before readback in the second loop.

    for (let i = 0; i < collections.length; i++) {
      const collection = collections[i];
      collection.noAnimate = true;
      const components = collection.Components;
      for (let j = 0; j < components.length; j++) {
        const component = components[j];
        if (component.id === '') continue;
        component.noAnimate = true;
      }
    }

    // This layout readback is the most important thing to do quickly
    // because if we thrash the DOM there will be a lot of recalc style. So
    // do it in its own pass.
    for (let i = 0; i < collections.length; i++) {
      const collection = collections[i];

      const offsetComponent = collection.offsetComponent;
      collectionOffsets.set(
        collection.id,
        captureOffsetGeometry(offsetComponent, this.ancestorOffsetParent),
      );
      collectionViewportOffsets.set(
        collection.id,
        captureViewportGeometry(offsetComponent),
      );

      const components = collection.Components;
      for (let j = 0; j < components.length; j++) {
        const component = components[j];
        if (component.id === '') continue;
        afterExact.push({ subjectId: component.id, collectionId: collection.id });
        let record = this._infoById[component.id];
        if (!record) {
          record = {};
          this._infoById[component.id] = record;
        }
        record.newOffsets = captureOffsetGeometry(component, this.ancestorOffsetParent);
        record.newViewportOffsets = captureViewportGeometry(component);
        record.visualSubject ??= this._captureMotionSubject(component);
      }
    }

    const continuityBySubject = new Map<string, MotionContinuityResolution>();
    const subjectIds = new Set([
      ...beforeExact.map(sighting => sighting.subjectId),
      ...afterExact.map(sighting => sighting.subjectId),
    ]);
    for (const subjectId of subjectIds) {
      continuityBySubject.set(subjectId, resolveStructuralContinuity(
        subjectId,
        beforeExact,
        afterExact,
        histories,
      ));
    }

    // This is the meat of the method, where we set all layout-affecting
    // properties, append fake dom, etc.
    for (let i = 0; i < collections.length; i++) {
      const collection = collections[i];

      const components = collection.Components;
      for (let j = 0; j < components.length; j++) {
        const component = components[j];

        if (component.id === '') continue;

        const record = this._infoById[component.id];
        const continuity = continuityBySubject.get(component.id);
        if (!continuity || continuity.status !== 'resolved'
          || continuity.presence === 'departing') continue;
        const hadExactBefore = continuity.presence === 'retained';
        let presence: 'retained' | 'appearing' = 'retained';
        let provenance: StructuralProvenance = { kind: 'identity' };

        if (continuity.presence === 'appearing') {
          // Hmm, a record who didn't have its offsets set in prepare(),
          // presumably because it didn't exist. This MAY be an element who
          // came from a PolicyNonEmpty stack.

          if (continuity.from.kind !== 'collection') continue;
          const theStack = stackById.get(continuity.from.collectionId);
          if (!theStack) continue;

          presence = 'appearing';
          provenance = {
            kind: 'stack-history',
            endpoint: 'source',
            stackId: theStack.id,
            evidence: 'runner-up',
          };

          record.offsets = this._beforeCollectionOffsets.get(theStack.id);
          record.viewportOffsets = this._beforeCollectionViewportOffsets.get(theStack.id);

          record.before = component.animatingPropDefaults(theStack);

          record.afterOpacity = component.style.opacity;
          record.afterTransform = component.style.transform;

          const presenceStyle = motionPresenceHostStyle(theStack.motionPresenceFacts());
          record.beforeTransform = presenceStyle.transform;
          record.beforeOpacity = presenceStyle.opacity;
        } else {
          record.afterOpacity = component.style.opacity;
          record.afterTransform = component.style.transform;
        }

        // Mark that we've seen where this one is going.
        this._beforeSeenIds.delete(component.id);

        record.after = component.animatingPropValues();

        // CRITICAL: Transform composition order - invert + external + scale.
        const geometry = solveFlipGeometry(record.offsets!, record.newOffsets!, {
          beforeOrientation: component.motionEndpointOrientation(record.before),
          afterOrientation: component.motionEndpointOrientation(record.after),
        });

        // Determine whether the host element's CSS transform will actually
        // change during the FLIP animation. The browser only fires
        // transitionend when the computed value differs between the inverted
        // and final states. For components that didn't move position and whose
        // inline transform (e.g. messy stack rotation) is unchanged, the
        // inversion is effectively identity and the target matches — so no
        // transition fires and we must not expect one.
        const hasInlineTransformChange =
          (record.beforeInlineTransform || '') !== (record.afterTransform || '');
        record.needsHostTransition = geometry.changed || hasInlineTransformChange;

        // Plan every owned animation channel once. The same immutable tracks
        // decide whether work exists and later drive WAAPI playback.
        record.invertedTransform = composeFlipTransform(geometry, record.beforeTransform);
        const motionTracks = this._planMotionTracks(component, {
          before: record.before || {},
          after: record.after!,
          invertedTransform: record.invertedTransform,
          finalTransform: record.afterTransform || '',
          beforeOpacity: record.beforeOpacity || '1',
          finalOpacity: record.afterOpacity || '',
          needsHostTransition: record.needsHostTransition,
        });
        record.motionTracks = motionTracks;
        record.needsAnimation = motionTracks.length > 0;

        // We used to only bother setting transforms for items that had
        // physically moved. However, the browser is smart enough to ignore
        // transforms that are basically no ops. And if we don't set it
        // then cards that don't physically move but do have transform
        // changes won't animate because the transform was set during
        // noAnimate and is never set to anything different. In testing
        // this didn't appear to have any appreciable performance difference.
        // Only prepare animation (clone content) for components that
        // actually need animation. Non-animating components skip the entire
        // FLIP pipeline, avoiding spurious will-animate events. The inverted
        // transform and before-opacity are stashed on the record; the WAAPI
        // PLAY phase in _startAnimations turns them into keyframes.
        if (record.needsAnimation) {
          record.motionDraft = createStructuralMotionDraft({
            subjectId: component.id,
            presence: hadExactBefore ? 'retained' : presence,
            provenance,
            visualSubject: record.visualSubject,
            viewportFrom: record.viewportOffsets!,
            viewportTo: record.newViewportOffsets!,
            inversion: geometry,
            channels: record.motionTracks,
          });

          if (record.historicalPresentation) {
            installHistoricalPresentation(component, record.historicalPresentation);
          }
        }
      }
    }

    this._animatingComponents = [];

    // Any items still in _beforeSeenIds did not have a specific card to
    // animate to. Let's see if we can figure out which collection they
    // went to.
    for (const id of this._beforeSeenIds) {
      const continuity = continuityBySubject.get(id);
      if (!continuity || continuity.status !== 'resolved'
        || continuity.presence !== 'departing'
        || continuity.to.kind !== 'collection') continue;
      const destinationStack = stackById.get(continuity.to.collectionId);
      if (!destinationStack) continue;

      const record = this._infoById[id];
      const carrier = destinationStack.newMotionCarrier();
      const component = carrier.component;
      if (record.sourceTagName !== component.localName
        || (record.historicalPresentation
          && !installHistoricalPresentation(component, record.historicalPresentation))) {
        if (typeof component.beforeOrphaned === 'function') component.beforeOrphaned();
        component.remove();
        continue;
      }
      const carrierPresenceStyle = motionPresenceHostStyle(carrier.presence);

      record.after = carrier.defaults;

      const animatingRecord: AnimatingComponentRecord = {
        subjectId: id,
        stack: destinationStack,
        component: component,
        before: record.before || {},
        after: record.after || {},
        afterTransform: carrierPresenceStyle.transform,
        afterOpacity: carrierPresenceStyle.opacity,
        invertedTransform: '',
        beforeOpacity: record.beforeOpacity || '1.0',
        needsHostTransition: true
      };
      this._animatingComponents.push(animatingRecord);

      const stackLocation = collectionOffsets.get(destinationStack.id);
      const stackViewportLocation = collectionViewportOffsets.get(destinationStack.id);
      const oldLocation = record.offsets;
      const oldViewportLocation = record.viewportOffsets;

      if (!stackLocation || !stackViewportLocation || !oldLocation || !oldViewportLocation) continue;

      const geometry = solveFlipGeometry(oldLocation, stackLocation, {
        beforeOrientation: component.motionEndpointOrientation(record.before),
        afterOrientation: component.motionEndpointOrientation(record.after),
      });

      // We used to only bother setting transforms for items that had
      // physically moved. However, the browser is smart enough to ignore
      // transforms that are basically no ops. And if we don't set it
      // then cards that don't physically move but do have transform
      // changes won't animate because the transform was set during
      // noAnimate and is never set to anything different. In testing
      // this didn't appear to have any appreciable performance difference.
      // Stash the inverted transform / before-opacity for the WAAPI PLAY
      // phase. We deliberately do NOT write component.style.transform/opacity
      // here: the resting inline transform stays put, and playAnimation()
      // supplies the inverted state as the animation's opening keyframe.
      animatingRecord.invertedTransform = composeFlipTransform(geometry, record.beforeTransform);
      animatingRecord.motionTracks = this._planMotionTracks(component, {
        before: animatingRecord.before,
        after: animatingRecord.after,
        invertedTransform: animatingRecord.invertedTransform,
        finalTransform: animatingRecord.afterTransform,
        beforeOpacity: animatingRecord.beforeOpacity,
        finalOpacity: animatingRecord.afterOpacity,
        needsHostTransition: true,
      });
      animatingRecord.motionDraft = createStructuralMotionDraft({
        subjectId: id,
        presence: 'departing',
        provenance: {
          kind: 'stack-history',
          endpoint: 'destination',
          stackId: destinationStack.id,
          evidence: 'latest-seen',
        },
        visualSubject: record.visualSubject,
        viewportFrom: oldViewportLocation,
        viewportTo: stackViewportLocation,
        inversion: geometry,
        channels: animatingRecord.motionTracks,
      });

    }

    // CRITICAL: Wait for styles to be set, then schedule PLAY phase in RAF
    // Polyfill for older browsers
    const raf = window.requestAnimationFrame ||
                (window as any).webkitRequestAnimationFrame ||
                ((cb: FrameRequestCallback) => window.setTimeout(cb, 16));
    raf(() => {
      void this._startAnimations(resolve, generation).catch(error => {
        this._abortAnimationCycle(error, resolve);
      });
    });
  }

  private async _startAnimations(resolve: (p: Promise<void>) => void, generation: number) {
    if (this._generation !== generation) { resolve(Promise.resolve()); return; }

    const collections = this.stackElement._sharedStackList;

    // Restore noAnimate (was the measurement barrier; still gates play()).
    const allComponents: any[] = [];
    for (let i = 0; i < collections.length; i++) {
      const collection = collections[i];
      collection.noAnimate = false;
      const components = collection.Components;
      for (let j = 0; j < components.length; j++) {
        const component = components[j];
        if (component.id === '') continue;
        component.noAnimate = false;
        component.animationContext = this.animationContext;
        allComponents.push(component);
      }
    }
    for (const ac of this._animatingComponents) {
      ac.component.noAnimate = false;
      ac.component.animationContext = this.animationContext;
      allComponents.push(ac.component);
    }

    await Promise.all(allComponents.map(c => c.updateComplete));
    if (this._generation !== generation) { resolve(Promise.resolve()); return; }

    const playback: Array<{
      component: any;
      config: {
        before: Record<string, any>;
        after: Record<string, any>;
        invertedTransform: string;
        finalTransform: string;
        beforeOpacity: string;
        finalOpacity: string;
        needsHostTransition: boolean;
        delayMs?: number;
        tracks?: readonly ComponentMotionTrack[];
      };
      motionDraft?: StructuralMotionDraft;
      motionSegmentIndex?: number;
      durationMs: number;
      delayMs: number;
    }> = [];

    for (let i = 0; i < collections.length; i++) {
      const collection = collections[i];
      const staggerFraction = collection.stagger || 0;
      let animIndex = 0;
      const components = collection.Components;
      for (let j = 0; j < components.length; j++) {
        const component = components[j];
        if (component.id === '') continue;
        const record = this._infoById[component.id];
        if (!record || !record.needsAnimation) continue;
        const delayMs = staggerFraction > 0
          ? animIndex * staggerFraction * component.animationLengthMs()
          : 0;
        animIndex++;
        playback.push({
          component,
          config: {
            before: record.before || {},
            after: record.after || {},
            invertedTransform: record.invertedTransform || '',
            finalTransform: record.afterTransform || '',
            beforeOpacity: record.beforeOpacity || '1',
            finalOpacity: record.afterOpacity || '',
            needsHostTransition: record.needsHostTransition ?? true,
            delayMs,
            tracks: record.motionTracks,
          },
          motionDraft: record.motionDraft,
          durationMs: component.animationLengthMs(),
          delayMs,
        });
      }
    }

    for (const ac of this._animatingComponents) {
      playback.push({
        component: ac.component,
        config: {
          before: ac.before || {},
          after: ac.after,
          invertedTransform: ac.invertedTransform || '',
          finalTransform: ac.afterTransform,
          beforeOpacity: ac.beforeOpacity || '1',
          finalOpacity: ac.afterOpacity,
          needsHostTransition: true,
          tracks: ac.motionTracks,
        },
        motionDraft: ac.motionDraft,
        durationMs: ac.component.animationLengthMs(),
        delayMs: 0,
      });
    }

    const installedCohorts = this._motionCohorts?.generation === generation
      ? this._motionCohorts.specs
      : [];
    const schedule = compileMotionCohortSchedule(
      playback.map(item => ({
        subjectId: item.component.id,
        legacyDelayMs: item.delayMs,
      })),
      installedCohorts,
    );
    if (schedule.status === 'fallback' && installedCohorts.length > 0) {
      console.error(`[motion] cohort scheduling fell back to stack timing: ${schedule.reason}`);
    }
    const delayBySubject = new Map(schedule.entries.map(entry => [entry.subjectId, entry.delayMs]));
    for (const item of playback) {
      const delayMs = delayBySubject.get(item.component.id) ?? item.delayMs;
      item.delayMs = delayMs;
      item.config.delayMs = delayMs;
    }

    // Publication barrier: everything above is measurement and planning.
    // Only a still-current generation may publish, and publication happens
    // before the first component begins playback.
    if (this._generation !== generation) { resolve(Promise.resolve()); return; }
    const planEntries: Array<{
      draft: StructuralMotionDraft;
      timingRequest: { policy: 'version'; delayMs: number; durationMs: number };
    }> = [];
    for (const item of playback) {
      if (!item.motionDraft) continue;
      item.motionSegmentIndex = planEntries.length;
      planEntries.push({
        draft: item.motionDraft,
        timingRequest: {
          policy: 'version' as const,
          delayMs: item.delayMs,
          durationMs: item.durationMs,
        },
      });
    }
    this._setSolvedMotionPlan(publishStructuralMotionPlan(
      generation,
      planEntries,
    ));

    // Declarative transfers are one separately-owned explicit batch. Resolve,
    // publish, and arm the complete batch before automatic FLIP playback.
    const settledPromises: Promise<void>[] = [
      this._startInstalledMotionTransfers(generation),
    ];
    const executionUpdates = new Map<number, StructuralExecution>();
    const terminalUpdates: Array<{
      segmentIndex: number;
      primaryAnimation: Animation | null;
      primaryDelayMs: number;
      settled: Promise<boolean[]>;
    }> = [];
    for (const item of playback) {
      const animations = item.component.playAnimation(item.config) as readonly Animation[];
      const tracks = item.config.tracks ?? [];
      if (item.motionDraft) {
        const segmentIndex = item.motionSegmentIndex!;
        if (animations.length === 0 || animations.length !== tracks.length) {
          for (const animation of animations) {
            void animation.finished.catch(() => undefined);
            animation.cancel();
          }
          executionUpdates.set(segmentIndex, {
            status: 'skipped',
            reason: animations.length === 0 ? 'not-started' : 'playback-error',
          });
        } else {
          const executedTimings = Object.freeze(animations.map((animation, index) => (
            this._executedTiming(animation, componentMotionChannel(tracks[index]))
          )));
          executionUpdates.set(segmentIndex, {
            status: 'armed',
            animations: executedTimings,
          });
          const primary = primaryStructuralAnimationIndex(
            item.motionDraft.path?.kind,
            executedTimings,
          );
          terminalUpdates.push({
            segmentIndex,
            primaryAnimation: primary === null ? null : animations[primary],
            primaryDelayMs: primary === null ? 0 : executedTimings[primary].delayMs,
            settled: Promise.all(animations.map(animation => animation.finished.then(
              () => false,
              () => true,
            ))),
          });
        }
      }
      settledPromises.push(item.component.settled());
    }
    const plannedPlan = this._solvedMotionPlan;
    if (!plannedPlan) { resolve(Promise.resolve()); return; }
    this._setSolvedMotionPlan(updateStructuralMotionExecutions(
      plannedPlan,
      executionUpdates,
    ));
    for (const terminal of terminalUpdates) {
      if (!terminal.primaryAnimation) continue;
      const ref = this._solvedMotionPlan?.segments[terminal.segmentIndex]?.ref;
      if (!ref) continue;
      const current = this._solvedMotionPlan?.segments[terminal.segmentIndex];
      if (current?.execution.status !== 'armed') continue;
      const animations = current.execution.animations;
      this._activationMonitor.observe(
        this._segmentId(ref),
        terminal.primaryAnimation,
        terminal.primaryDelayMs,
        () => this._updateFlipMotion(generation, terminal.segmentIndex, {
          status: 'active-observed', animations,
        }),
      );
    }
    for (const terminal of terminalUpdates) {
      void terminal.settled.then(cancelled => {
        if (this._generation !== generation || !this._solvedMotionPlan) return;
        const current = this._solvedMotionPlan.segments[terminal.segmentIndex];
        const animations = current?.execution.status === 'armed'
          || current?.execution.status === 'active-observed'
          ? current.execution.animations
          : Object.freeze([]);
        if (current) this._activationMonitor.cancel(this._segmentId(current.ref));
        const wasCancelled = cancelled.some(Boolean);
        if (!wasCancelled && current?.execution.status === 'armed') {
          this._setSolvedMotionPlan(updateStructuralMotionExecutions(
            this._solvedMotionPlan,
            new Map([[terminal.segmentIndex, {
              status: 'active-observed' as const,
              animations,
            }]]),
          ));
        }
        this._setSolvedMotionPlan(updateStructuralMotionExecutions(
          this._solvedMotionPlan,
          new Map([[terminal.segmentIndex, {
            status: wasCancelled ? 'cancelled' as const : 'finished' as const,
            animations,
          }]]),
        ));
      });
    }

    // The promise animateFlip() hands out now means "everything SETTLED",
    // not "everything started" — the gate awaits real completion.
    resolve(Promise.all(settledPromises).then(() => {}));
  }

  override render(): TemplateResult {
    return html` <boardgame-component-stack id="stack" no-default-spacer=""></boardgame-component-stack> `;
  }
}

customElements.define('boardgame-component-animator', BoardgameComponentAnimator);
