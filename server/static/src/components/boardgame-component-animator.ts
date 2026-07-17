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
  solveFlipGeometry,
} from '../motion/geometry.js';
import type { OffsetGeometry } from '../motion/geometry.js';

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
}

interface ComponentRecord {
  offsets?: OffsetGeometry;
  newOffsets?: OffsetGeometry;
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
}

interface CollectionRecord {
  stack: any;
  version: number;
  runnerUpStack?: any;
  runnerUpVersion?: number;
}

interface AnimatingComponentRecord {
  stack: any;
  component: any;
  before: Record<string, any>;
  after: Record<string, any>;
  afterTransform: string;
  afterOpacity: string;
  invertedTransform: string;
  beforeOpacity: string;
  needsHostTransition: boolean;
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
  private _generation = 0;

  ancestorOffsetParent: HTMLElement | null = null;

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
    this._generation++;
    const collections = this.stackElement._sharedStackList;

    this._beforeCollectionOffsets = new Map();

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

      const components = collection.Components;
      for (let j = 0; j < components.length; j++) {
        const component = components[j];

        // Skip components without ids (e.g. faux-components, spacer components).
        if (component.id === '') continue;

        const record = result[component.id] || {};

        this._beforeSeenIds.add(component.id);

        record.offsets = captureOffsetGeometry(component, this.ancestorOffsetParent);

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
    const real = this._resolveAnimationTarget(realId);
    const stub = this._resolveAnimationTarget(stubId);
    if (!real || !stub) {
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
    if (dx === 0 && dy === 0) {
      return;
    }
    const timing = opts?.timing ?? 'version';
    const resolution = resolveMotionTiming(
      { duration: durationMs, easing: 'ease-out', fill: 'none' },
      {
        policy: timing,
        context: timing === 'version' ? this.animationContext : null,
      },
    );
    if (resolution.kind === 'skip') return;
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

          record.offsets = this._beforeCollectionOffsets.get(theStack.id);

          record.before = component.animatingPropDefaults(theStack);

          record.afterOpacity = component.style.opacity;
          record.afterTransform = component.style.transform;

          theStack.setUnknownAnimationState(component);

          record.beforeTransform = component.style.transform;
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
          beforeTransform: record.beforeTransform,
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
        const beforeOpacity = parseFloat(component.style.opacity || '1');
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
          record.invertedTransform = geometry.invertedTransform;
          record.beforeOpacity = component.style.opacity;

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
        stack: anonRecord.stack,
        component: component,
        before: record.before || {},
        after: record.after || {},
        afterTransform: component.style.transform,
        afterOpacity: component.style.opacity,
        invertedTransform: '',
        beforeOpacity: '1.0',
        needsHostTransition: true
      };
      this._animatingComponents.push(animatingRecord);

      const stackLocation = collectionOffsets.get(anonRecord.stack.id);
      const oldLocation = record.offsets;

      if (!stackLocation || !oldLocation) continue;

      const geometry = solveFlipGeometry(oldLocation, stackLocation, {
        rotates: component.animationRotates(record.before, record.after),
        beforeTransform: record.beforeTransform,
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
      animatingRecord.invertedTransform = geometry.invertedTransform;
      animatingRecord.beforeOpacity = '1.0';

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

    const settledPromises: Promise<void>[] = [];

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
        component.playAnimation({
          before: record.before || {},
          after: record.after || {},
          invertedTransform: record.invertedTransform || '',
          finalTransform: record.afterTransform || '',
          beforeOpacity: record.beforeOpacity || '1',
          finalOpacity: record.afterOpacity || '',
          needsHostTransition: record.needsHostTransition ?? true,
          delayMs,
        });
        settledPromises.push(component.settled());
      }
    }

    for (const ac of this._animatingComponents) {
      ac.component.playAnimation({
        before: ac.before || {},
        after: ac.after,
        invertedTransform: ac.invertedTransform || '',
        finalTransform: ac.afterTransform,
        beforeOpacity: ac.beforeOpacity || '1',
        finalOpacity: ac.afterOpacity,
        needsHostTransition: true,
      });
      settledPromises.push(ac.component.settled());
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

        if (idsLastSeen[key] > record.version) {
          // new winner
          const newRecord: CollectionRecord = {
            version: idsLastSeen[key],
            stack: stack,
            runnerUpVersion: record.version,
            runnerUpStack: record.stack
          };
          possibleLocations.set(key, newRecord);
        } else if (!record.runnerUpStack || idsLastSeen[key] > (record.runnerUpVersion || 0)) {
          // Found a new second!
          possibleLocations.set(key, {
            version: record.version,
            stack: record.stack,
            runnerUpVersion: idsLastSeen[key],
            runnerUpStack: stack
          });
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
