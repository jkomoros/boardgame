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
  centeredInversionDelta,
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
  StructuralMotionObserver,
  StructuralMotionPlan,
  StructuralProvenance,
} from '../motion/structural-plan.js';

export type { AnimationTimingPolicy } from '../motion/timing.js';

export interface AnimateBetweenOptions {
  /** Defaults to the installed version's companion slot. */
  timing?: AnimationTimingPolicy;
}

export interface ComponentAnimatorAPI {
  animateBetween(
    realId: string | HTMLElement,
    stubId: string | HTMLElement,
    durationMs?: number,
    opts?: AnimateBetweenOptions,
  ): Promise<void>;
  /** Internal observation surface; does not confer animation ownership. */
  observeStructuralMotion(observer: StructuralMotionObserver): () => void;
}

interface ComponentRecord {
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
  motionDraft?: StructuralMotionDraft;
}

interface CollectionRecord {
  stack: any;
  version: number;
  winnerAmbiguous?: boolean;
  runnerUpStack?: any;
  runnerUpVersion?: number;
  runnerUpAmbiguous?: boolean;
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
  motionDraft?: StructuralMotionDraft;
}

export class BoardgameComponentAnimator extends LitElement {
  // Supplied by boardgame-render-game for the currently installing version.
  // Public so the wrapper can set it, but game renderers normally need no
  // timing plumbing: animateBetween consumes it by default.
  animationContext: VersionAnimationContext | null = null;
  // Note: Can't use @query decorator because 'animate' method conflicts with Element.animate()
  private get stackElement(): BoardgameComponentStack {
    return this.shadowRoot!.querySelector('#stack')!;
  }

  private _infoById: { [id: string]: ComponentRecord } = {};
  private _lastSeenNodesById = new Map<string, Node[]>();
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
  private _motionObservers = new Set<StructuralMotionObserver>();

  ancestorOffsetParent: HTMLElement | null = null;

  observeStructuralMotion(observer: StructuralMotionObserver): () => void {
    this._motionObservers.add(observer);
    return () => this._motionObservers.delete(observer);
  }

  private _notifyStructuralMotion(plan: StructuralMotionPlan): void {
    for (const observer of this._motionObservers) {
      try {
        observer(plan);
      } catch (error) {
        console.error('[animator] structural motion observer failed:', error);
      }
    }
  }

  private _setSolvedMotionPlan(plan: StructuralMotionPlan): void {
    this._solvedMotionPlan = plan;
    this._notifyStructuralMotion(plan);
  }

  private _invalidateSolvedMotionPlan(): void {
    const plan = this._solvedMotionPlan;
    if (!plan || plan.phase === 'settled') return;
    const updates = new Map<string, StructuralExecution>();
    for (const segment of plan.segments) {
      if (segment.execution.status === 'started') {
        updates.set(segment.subjectId, {
          status: 'cancelled',
          animations: segment.execution.animations,
        });
      } else if (segment.execution.status === 'planned') {
        updates.set(segment.subjectId, {
          status: 'skipped',
          reason: 'not-started',
        });
      }
    }
    this._notifyStructuralMotion(updateStructuralMotionExecutions(plan, updates));
  }

  override firstUpdated(_changedProperties: Map<PropertyKey, unknown>) {
    super.firstUpdated(_changedProperties);
    this._lastSeenNodesById = new Map();
  }

  /** Clear interrupted faux components without exposing the stack registry. */
  clearAnimatingComponents(): void {
    for (const stack of this.stackElement._sharedStackList) {
      stack.clearAnimatingComponents();
    }
  }

  prepare() {
    this._invalidateSolvedMotionPlan();
    this._generation++;
    this._solvedMotionPlan = null;
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

        this._beforeSeenIds.add(component.id);

        record.offsets = captureOffsetGeometry(component, this.ancestorOffsetParent);
        record.viewportOffsets = captureViewportGeometry(component);

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

        if (component.cloneContent) {
          const newNodes: Node[] = [];
          const children = component.children;
          for (let k = 0; k < children.length; k++) {
            const child = children[k];
            if ((child as HTMLElement).slot) {
              // Skip content that doesn't go in default slot
              continue;
            }
            if ((child as Element).localName === 'dom-bind') {
              continue;
            }
            newNodes.push(child.cloneNode(true));
          }
          if (newNodes.length > 0) {
            this._lastSeenNodesById.set(component.id, newNodes);
          }
        }
        result[component.id] = record;
      }
    }

    this._infoById = result;
  }

  /**
   * _resolveAnimationTarget turns an animateBetween argument into a live
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

  private _executedTiming(animation: Animation): StructuralExecutedTiming {
    const timing = animation.effect instanceof KeyframeEffect
      ? animation.effect.getTiming()
      : {};
    const iterations = timing.iterations === undefined
      ? 1
      : Math.max(0, finiteTimingMs(timing.iterations));
    return Object.freeze({
      delayMs: finiteTimingMs(timing.delay),
      durationMs: finiteTimingMs(timing.duration),
      endDelayMs: finiteTimingMs(timing.endDelay),
      iterations,
      easing: timing.easing ?? 'linear',
      fill: timing.fill ?? 'none',
    });
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
    subjectId: string,
    execution: StructuralExecution,
  ): void {
    const current = this._explicitMotionPlans.get(generation);
    if (!current) return;
    const updated = updateStructuralMotionExecutions(
      current,
      new Map([[subjectId, execution]]),
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
  ): void {
    if (!animation) {
      this._updateExplicitMotion(generation, subjectId, {
        status: 'skipped',
        reason: 'not-started',
      });
      return;
    }
    const animations = Object.freeze([this._executedTiming(animation)]);
    this._updateExplicitMotion(generation, subjectId, {
      status: 'started',
      animations,
    });
    void animation.finished.then(
      () => this._updateExplicitMotion(generation, subjectId, { status: 'finished', animations }),
      () => this._updateExplicitMotion(generation, subjectId, { status: 'cancelled', animations }),
    );
  }

  /**
   * animateBetween runs a one-off FLIP animation moving the element with id
   * `realId` from the on-screen position of `stubId` (or vice versa). Used
   * by the Table view's fake-deck row to fly cards between the visible
   * board and the off-screen per-player stub stacks (spec §8.1).
   *
   * This is intentionally simpler than the prepare/animateFlip pipeline:
   * it doesn't walk _sharedStackList and doesn't depend on the diff
   * machinery — caller supplies both element IDs explicitly. The
   * synthetic-stub-ID approach (e.g. "stub:p3:c17" for player 3's mirror
   * of real card "c17") avoids the flat-map collision that would happen
   * if both elements shared the same component.id.
   *
   * Direction: the `realId` element starts at `stubId`'s on-screen
   * position and flies to its own natural rendered position (i.e. it
   * visually ARRIVES FROM the stub). To make an element visually depart
   * toward a location instead, swap the argument order — the departing
   * element then plays the arrival animation at the other end.
   *
   * Returns a Promise that resolves when the animation completes (or
   * immediately if either element is missing from the DOM).
   *
   * Implementation: WAAPI overlay animation. We measure both elements
   * with getBoundingClientRect, compute the delta, and run a two-keyframe
   * element.animate() from the inverted (stub-aligned) transform back to
   * the element's own resting transform. fill:'none' means the animation
   * never writes a persistent style — no inline transform/transition
   * juggling, no forced reflow, no transitionend/setTimeout race.
   * Settlement (anim.finished, or the cancel rejection) is ground truth
   * for "done".
   *
   * The current version's companion slot is used automatically. A caller may
   * choose immediate timing for a deliberately local animation, or supply an
   * explicit local start through the discriminated timing policy.
   */
  async animateBetween(
    realId: string | HTMLElement,
    stubId: string | HTMLElement,
    durationMs: number = 500,
    opts?: AnimateBetweenOptions,
  ): Promise<void> {
    const explicitGeneration = ++this._explicitMotionSequence;
    const subjectId = typeof realId === 'string'
      ? realId
      : realId.id || `explicit-motion-${explicitGeneration}`;
    const timing = opts?.timing ?? 'version';
    const real = this._resolveAnimationTarget(realId);
    const stub = this._resolveAnimationTarget(stubId);
    if (!real || !stub) {
      const draft = createStructuralMotionDraft({
        subjectId,
        presence: 'retained',
        provenance: {
          kind: 'unresolved',
          endpoint: !stub ? 'source' : 'destination',
        },
      });
      this._installExplicitMotion(publishStructuralMotionPlan(explicitGeneration, [{
        draft,
        timingRequest: { policy: timing, delayMs: 0, durationMs },
      }], 'explicit'));
      this._updateExplicitMotion(explicitGeneration, subjectId, {
        status: 'skipped',
        reason: 'missing-endpoint',
      });
      // Loud on purpose: an unresolvable endpoint silently killed the
      // entire cross-screen animation feature once already (the id
      // property wasn't reflected to the DOM, so no card ever matched).
      console.warn('[animator] animateBetween: could not resolve',
        !real ? realId : stubId, '— animation skipped');
      return;
    }
    // Measure both endpoints in the SAME coordinate space (viewport rects)
    // and align CENTERS. The previous offsetParent-chain math mixed
    // coordinate spaces when the two elements had different fixed/absolute
    // ancestors, and corner-alignment launched flights from the top-left
    // of full-width containers instead of from the visual source.
    const realRect = captureViewportGeometry(real);
    const stubRect = captureViewportGeometry(stub);
    const { x: dx, y: dy } = centeredInversionDelta(realRect, stubRect);
    const draft = createStructuralMotionDraft({
      subjectId,
      presence: 'retained',
      provenance: { kind: 'identity' },
      viewportFrom: stubRect,
      viewportTo: realRect,
      inversion: Object.freeze({
        translateX: dx,
        translateY: dy,
        scale: 1,
        changed: dx !== 0 || dy !== 0,
      }),
    });
    this._installExplicitMotion(publishStructuralMotionPlan(explicitGeneration, [{
      draft,
      timingRequest: { policy: timing, delayMs: 0, durationMs },
    }], 'explicit'));
    if (dx === 0 && dy === 0) {
      this._updateExplicitMotion(explicitGeneration, subjectId, {
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
      },
    );
    if (resolution.kind === 'skip') {
      this._updateExplicitMotion(explicitGeneration, subjectId, {
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
    const keyframes: Keyframe[] = [
      { transform: `translate(${dx}px, ${dy}px) ${real.style.transform || ''}`.trim() },
      { transform: real.style.transform || 'none' },
    ];
    const realTag = real.tagName.toLowerCase() + (real.id ? `#${real.id}` : '');

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
    // Plain elements (e.g. the divs the waapi-play.spec.ts animateBetween
    // test uses) have no play() and keep the raw-element fallback below.
    if (typeof (real as any).play === 'function') {
      const anim = (real as any).play(real, keyframes,
        resolution.timing,
        { timing: 'immediate' }, { recordActive: false });
      this._trackExplicitAnimation(explicitGeneration, subjectId, anim);
      // play() returns null under noAnimate; nothing is in flight then.
      if (anim) {
        this._recordAnimationActive(
          anim,
          delay,
          'fly:' + realTag,
          resolution.activeContext,
        );
        await anim.finished.catch(() => {});
        // The Animation promise and play()'s settlement bookkeeping are
        // separate promise reactions. Do not resolve animateBetween until
        // the latter has closed the render-game completion gate too.
        if (typeof (real as any).settled === 'function') {
          await (real as any).settled();
        }
      }
      return;
    }

    const anim = real.animate(keyframes, resolution.timing);
    this._trackExplicitAnimation(explicitGeneration, subjectId, anim);
    this._recordAnimationActive(
      anim,
      delay,
      'fly:' + realTag,
      resolution.activeContext,
    );
    // Settlement is ground truth: finished resolves on completion, rejects
    // on cancel (element removed mid-flight) — both mean "done" here.
    await anim.finished.catch(() => {});
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
      this._doAnimate(resolve, generation);
    });
  }

  private _doAnimate(resolve: (p: Promise<void>) => void, generation: number) {
    const collections = this.stackElement._sharedStackList;

    // The last seen location of a given card ID
    const idToPossibleCollection = new Map<string, CollectionRecord>();

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

      // Note which Ids were last seen here
      this._ingestStack(idToPossibleCollection, collection);

      const components = collection.Components;
      for (let j = 0; j < components.length; j++) {
        const component = components[j];
        if (component.id === '') continue;
        let record = this._infoById[component.id];
        if (!record) {
          record = {};
          this._infoById[component.id] = record;
        }
        record.newOffsets = captureOffsetGeometry(component, this.ancestorOffsetParent);
        record.newViewportOffsets = captureViewportGeometry(component);
      }
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
        const hadExactBefore = record.offsets !== undefined;
        let presence: 'retained' | 'appearing' = 'retained';
        let provenance: StructuralProvenance = { kind: 'identity' };

        if (!record.offsets) {
          // Hmm, a record who didn't have its offsets set in prepare(),
          // presumably because it didn't exist. This MAY be an element who
          // came from a PolicyNonEmpty stack.

          const collectionRecord = idToPossibleCollection.get(component.id);

          if (!collectionRecord) {
            // Nah, we don't know where it came from. Just skip animating it.
            continue;
          }

          let theStack = collectionRecord.stack;
          // We actually want the runner up, if it exists. the winner is
          // the stack it's now in, and the runner up should be where it
          // just came from.
          if (collectionRecord.runnerUpStack) {
            theStack = collectionRecord.runnerUpStack;
          }

          presence = 'appearing';
          provenance = {
            kind: 'stack-history',
            endpoint: 'source',
            stackId: theStack.id,
            evidence: collectionRecord.runnerUpStack
              ? collectionRecord.runnerUpAmbiguous ? 'ambiguous' : 'runner-up'
              : collectionRecord.winnerAmbiguous ? 'ambiguous' : 'only-candidate',
          };

          record.offsets = this._beforeCollectionOffsets.get(theStack.id);
          record.viewportOffsets = this._beforeCollectionViewportOffsets.get(theStack.id);

          record.before = component.animatingPropDefaults(theStack);

          record.afterOpacity = component.style.opacity;
          record.afterTransform = component.style.transform;

          theStack.setUnknownAnimationState(component);

          record.beforeTransform = component.style.transform;
          record.beforeOpacity = component.style.opacity || '1';
        } else {
          record.afterOpacity = component.style.opacity;
          record.afterTransform = component.style.transform;
        }

        // Mark that we've seen where this one is going.
        this._beforeSeenIds.delete(component.id);

        record.after = component.animatingPropValues();

        // CRITICAL: Transform composition order - invert + external + scale.
        const geometry = solveFlipGeometry(record.offsets!, record.newOffsets!, {
          rotates: component.animationRotates(record.before, record.after),
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

        // Check if any animating properties changed (e.g. faceUp, rotated)
        const beforeProps = record.before || {};
        const afterProps = record.after!;
        let propsChanged = false;
        for (const propName of component.animatingProperties) {
          if (beforeProps[propName] !== afterProps[propName]) {
            propsChanged = true;
            break;
          }
        }

        // Check opacity change
        const beforeOpacity = parseFloat(record.beforeOpacity || '1');
        const afterOpacity = parseFloat(record.afterOpacity || '1');
        const opacityChanged = Math.abs(beforeOpacity - afterOpacity) > 0.01;

        record.needsAnimation = record.needsHostTransition || propsChanged || opacityChanged;

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
          record.invertedTransform = composeFlipTransform(geometry, record.beforeTransform);
          record.motionDraft = createStructuralMotionDraft({
            subjectId: component.id,
            presence: hadExactBefore ? 'retained' : presence,
            provenance,
            from: record.offsets!,
            to: record.newOffsets!,
            viewportFrom: record.viewportOffsets!,
            viewportTo: record.newViewportOffsets!,
            inversion: geometry,
            beforeTransform: hadExactBefore
              ? record.beforeInlineTransform
              : record.beforeTransform,
            afterTransform: record.afterTransform,
            beforeProperties: record.before,
            afterProperties: record.after,
            animatingProperties: component.animatingProperties,
            beforeOpacity: record.beforeOpacity,
            afterOpacity: record.afterOpacity,
          });

          const clonedNodes = this._lastSeenNodesById.get(component.id);

          if (clonedNodes && clonedNodes.length > 0) {
            // Clear out old nodes.
            for (let k = 0; k < component.children.length; k++) {
              const child = component.children[k];
              if ((child as HTMLElement).slot === 'fallback') {
                component.removeChild(child);
              }
            }
            for (let k = 0; k < clonedNodes.length; k++) {
              const node = clonedNodes[k];
              (node as HTMLElement).slot = 'fallback';
              component.appendChild(node);
            }
          }
        }
      }
    }

    this._animatingComponents = [];

    // Any items still in _beforeSeenIds did not have a specific card to
    // animate to. Let's see if we can figure out which collection they
    // went to.
    for (const id of this._beforeSeenIds) {
      // Which stack do we think this is in now?
      const anonRecord = idToPossibleCollection.get(id);

      if (!anonRecord) {
        // Guess it's a mystery. :-(
        continue;
      }

      const component = anonRecord.stack.newAnimatingComponent();

      const record = this._infoById[id];

      record.after = component.animatingPropDefaults(anonRecord.stack);

      const animatingRecord: AnimatingComponentRecord = {
        subjectId: id,
        stack: anonRecord.stack,
        component: component,
        before: record.before || {},
        after: record.after || {},
        afterTransform: component.style.transform,
        afterOpacity: component.style.opacity,
        invertedTransform: '',
        beforeOpacity: record.beforeOpacity || '1.0',
        needsHostTransition: true
      };
      this._animatingComponents.push(animatingRecord);

      const stackLocation = collectionOffsets.get(anonRecord.stack.id);
      const stackViewportLocation = collectionViewportOffsets.get(anonRecord.stack.id);
      const oldLocation = record.offsets;
      const oldViewportLocation = record.viewportOffsets;

      if (!stackLocation || !stackViewportLocation || !oldLocation || !oldViewportLocation) continue;

      const geometry = solveFlipGeometry(oldLocation, stackLocation, {
        rotates: component.animationRotates(record.before, record.after),
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
      animatingRecord.motionDraft = createStructuralMotionDraft({
        subjectId: id,
        presence: 'departing',
        provenance: {
          kind: 'stack-history',
          endpoint: 'destination',
          stackId: anonRecord.stack.id,
          evidence: anonRecord.winnerAmbiguous ? 'ambiguous' : 'latest-seen',
        },
        from: oldLocation,
        to: stackLocation,
        viewportFrom: oldViewportLocation,
        viewportTo: stackViewportLocation,
        inversion: geometry,
        beforeTransform: record.beforeInlineTransform ?? record.beforeTransform,
        afterTransform: animatingRecord.afterTransform,
        beforeProperties: animatingRecord.before,
        afterProperties: animatingRecord.after,
        animatingProperties: component.animatingProperties,
        beforeOpacity: animatingRecord.beforeOpacity,
        afterOpacity: animatingRecord.afterOpacity,
      });

      const clonedNodes = this._lastSeenNodesById.get(id);
      if (clonedNodes) {
        for (let k = 0; k < clonedNodes.length; k++) {
          const node = clonedNodes[k];
          (node as HTMLElement).slot = 'fallback';
          component.appendChild(node);
        }
      }
    }

    // CRITICAL: Wait for styles to be set, then schedule PLAY phase in RAF
    // Polyfill for older browsers
    const raf = window.requestAnimationFrame ||
                (window as any).webkitRequestAnimationFrame ||
                ((cb: FrameRequestCallback) => window.setTimeout(cb, 16));
    raf(() => this._startAnimations(resolve, generation));
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
      };
      motionDraft?: StructuralMotionDraft;
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
        },
        motionDraft: ac.motionDraft,
        durationMs: ac.component.animationLengthMs(),
        delayMs: 0,
      });
    }

    // Publication barrier: everything above is measurement and planning.
    // Only a still-current generation may publish, and publication happens
    // before the first component begins playback.
    if (this._generation !== generation) { resolve(Promise.resolve()); return; }
    this._setSolvedMotionPlan(publishStructuralMotionPlan(
      generation,
      playback.flatMap(item => item.motionDraft ? [{
        draft: item.motionDraft,
        timingRequest: {
          policy: 'version' as const,
          delayMs: item.delayMs,
          durationMs: item.durationMs,
        },
      }] : []),
    ));

    const settledPromises: Promise<void>[] = [];
    const executionUpdates = new Map<string, StructuralExecution>();
    const terminalUpdates: Array<{
      subjectId: string;
      settled: Promise<boolean[]>;
    }> = [];
    for (const item of playback) {
      const animations = item.component.playAnimation(item.config) as readonly Animation[];
      if (item.motionDraft) {
        if (animations.length === 0) {
          executionUpdates.set(item.motionDraft.subjectId, {
            status: 'skipped',
            reason: 'not-started',
          });
        } else {
          executionUpdates.set(item.motionDraft.subjectId, {
            status: 'started',
            animations: Object.freeze(animations.map(animation => this._executedTiming(animation))),
          });
          terminalUpdates.push({
            subjectId: item.motionDraft.subjectId,
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
      void terminal.settled.then(cancelled => {
        if (this._generation !== generation || !this._solvedMotionPlan) return;
        const current = this._solvedMotionPlan.segments.find(
          segment => segment.subjectId === terminal.subjectId,
        );
        const animations = current?.execution.status === 'started'
          ? current.execution.animations
          : Object.freeze([]);
        this._setSolvedMotionPlan(updateStructuralMotionExecutions(
          this._solvedMotionPlan,
          new Map([[terminal.subjectId, {
            status: cancelled.some(Boolean) ? 'cancelled' as const : 'finished' as const,
            animations,
          }]]),
        ));
      });
    }

    // The promise animateFlip() hands out now means "everything SETTLED",
    // not "everything started" — the gate awaits real completion.
    resolve(Promise.all(settledPromises).then(() => {}));
  }

  private _ingestStack(possibleLocations: Map<string, CollectionRecord>, stack: any) {
    const idsLastSeen = stack.idsLastSeen;

    for (const key in idsLastSeen) {
      if (!idsLastSeen.hasOwnProperty(key)) continue;

      if (possibleLocations.has(key)) {
        const record = possibleLocations.get(key)!;
        const seenVersion = idsLastSeen[key];

        if (seenVersion > record.version) {
          // new winner
          const newRecord: CollectionRecord = {
            version: seenVersion,
            stack: stack,
            runnerUpVersion: record.version,
            runnerUpStack: record.stack,
            runnerUpAmbiguous: record.winnerAmbiguous,
          };
          possibleLocations.set(key, newRecord);
        } else if (seenVersion === record.version) {
          possibleLocations.set(key, {
            ...record,
            winnerAmbiguous: true,
            runnerUpVersion: seenVersion,
            runnerUpStack: stack,
            runnerUpAmbiguous: true,
          });
        } else if (!record.runnerUpStack || seenVersion > (record.runnerUpVersion || 0)) {
          // Found a new second!
          possibleLocations.set(key, {
            ...record,
            version: record.version,
            stack: record.stack,
            runnerUpVersion: seenVersion,
            runnerUpStack: stack,
            runnerUpAmbiguous: false,
          });
        } else if (seenVersion === record.runnerUpVersion) {
          possibleLocations.set(key, { ...record, runnerUpAmbiguous: true });
        }
      } else {
        // We're the first one that's been seen; add it.
        possibleLocations.set(key, {
          version: idsLastSeen[key],
          stack: stack
        });
      }
    }
  }

  override render(): TemplateResult {
    return html` <boardgame-component-stack id="stack" no-default-spacer=""></boardgame-component-stack> `;
  }
}

customElements.define('boardgame-component-animator', BoardgameComponentAnimator);
