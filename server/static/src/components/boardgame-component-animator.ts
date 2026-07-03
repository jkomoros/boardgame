import { LitElement, html, TemplateResult } from 'lit';
import { query } from 'lit/decorators.js';
import './boardgame-component-stack.js';
import type { BoardgameComponentStack } from './boardgame-component-stack.js';

interface ComponentRecord {
  offsets?: OffsetRect;
  newOffsets?: OffsetRect;
  before?: Record<string, any>;
  after?: Record<string, any>;
  beforeTransform?: string;
  beforeInlineTransform?: string;
  afterTransform?: string;
  afterOpacity?: string;
  needsHostTransition?: boolean;
  needsAnimation?: boolean;
}

interface OffsetRect {
  top: number;
  left: number;
  width: number;
  height: number;
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
  after: Record<string, any>;
  afterTransform: string;
  afterOpacity: string;
  needsHostTransition: boolean;
}

export class BoardgameComponentAnimator extends LitElement {
  // Note: Can't use @query decorator because 'animate' method conflicts with Element.animate()
  private get stackElement(): BoardgameComponentStack {
    return this.shadowRoot!.querySelector('#stack')!;
  }

  private _infoById: { [id: string]: ComponentRecord } = {};
  private _lastSeenNodesById = new Map<string, Node[]>();
  private _beforeSeenIds = new Set<string>();
  private _animatingComponents: AnimatingComponentRecord[] = [];
  private _beforeCollectionOffsets = new Map<string, OffsetRect>();
  private _generation = 0;

  ancestorOffsetParent: any = null;

  private _calculateOffsets(ele: HTMLElement): OffsetRect {
    let top = 0;
    let left = 0;
    const width = ele.offsetWidth;
    const height = ele.offsetHeight;

    let offsetEle: HTMLElement | null = ele;
    while (offsetEle) {
      top += offsetEle.offsetTop;
      left += offsetEle.offsetLeft;

      if (offsetEle === this.ancestorOffsetParent) {
        offsetEle = null;
      } else {
        offsetEle = offsetEle.offsetParent as HTMLElement | null;
      }
    }

    return {
      top: top,
      left: left,
      width: width,
      height: height
    };
  }

  override firstUpdated(_changedProperties: Map<PropertyKey, unknown>) {
    super.firstUpdated(_changedProperties);
    this._lastSeenNodesById = new Map();
  }

  prepare() {
    this._generation++;
    const collections = this.stackElement._sharedStackList;

    this._beforeCollectionOffsets = new Map();

    const result: { [id: string]: ComponentRecord } = {};

    // keep track of all of the ids we've seen this round to make sure we
    // found a home for all of them in the end.
    this._beforeSeenIds = new Set();

    for (let i = 0; i < collections.length; i++) {
      const collection = collections[i];

      const offsetComponent = collection.offsetComponent;
      this._beforeCollectionOffsets.set(collection.id, this._calculateOffsets(offsetComponent));

      const components = collection.Components;
      for (let j = 0; j < components.length; j++) {
        const component = components[j];

        // Skip components without ids (e.g. faux-components, spacer components).
        if (component.id === '') continue;

        const record = result[component.id] || {};

        this._beforeSeenIds.add(component.id);

        record.offsets = this._calculateOffsets(component);

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
   * Implementation: pure CSS transform animation. We measure both
   * elements with _calculateOffsets, compute the delta, set
   * transition:'none', apply the delta as a translate, force layout,
   * then re-enable transitions and translate back to 0. The CSS
   * transition handles the actual frame interpolation.
   */
  async animateBetween(realId: string | HTMLElement, stubId: string | HTMLElement, durationMs: number = 500): Promise<void> {
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
    const realRect = real.getBoundingClientRect();
    const stubRect = stub.getBoundingClientRect();
    const dx = (stubRect.left + stubRect.width / 2) - (realRect.left + realRect.width / 2);
    const dy = (stubRect.top + stubRect.height / 2) - (realRect.top + realRect.height / 2);
    if (dx === 0 && dy === 0) {
      return;
    }
    const prevTransform = real.style.transform;
    // First/Invert: jump immediately to the stub position (no animation).
    real.style.transition = 'none';
    real.style.transform = `translate(${dx}px, ${dy}px) ${prevTransform || ''}`.trim();
    // Force layout to lock in the inverted position before transitioning.
    real.getBoundingClientRect();
    // Last/Play: animate back to original transform over durationMs.
    real.style.transition = `transform ${durationMs}ms ease-out`;
    real.style.transform = prevTransform || '';
    await new Promise<void>((resolve) => {
      let done = false;
      const cleanup = () => {
        if (done) return;
        done = true;
        // Restore to '' (stylesheet transition), NOT the captured inline
        // value: if another pipeline (the FLIP state-install path) had an
        // inline transition mid-flight when we started, blindly restoring
        // it later could stomp a NEWER animation on this element.
        real.style.transition = '';
        real.removeEventListener('transitionend', cleanup);
        resolve();
      };
      real.addEventListener('transitionend', cleanup);
      // Defensive timeout in case transitionend doesn't fire (which can
      // happen if the element is removed mid-animation).
      setTimeout(cleanup, durationMs + 100);
    });
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

    const generation = this._generation;

    return new Promise((resolve) => {
      // CRITICAL: First microtask - Let Polymer dispatch change events
      Promise.resolve().then(() => {
        if (this._generation !== generation) { resolve(); return; }
        this._scheduleAnimate(resolve, generation);
      });
    });
  }

  private _scheduleAnimate(resolve: () => void, generation: number) {
    // CRITICAL: Second microtask - Ensure ALL databinding cascades complete
    // This bizarre indirection is necessary because by the time the first
    // microtask resolves some databinding won't have been done, so we need to
    // one more time wait until the end of the microtask. See #722 for more.
    Promise.resolve().then(() => {
      if (this._generation !== generation) { resolve(); return; }
      this._doAnimate(resolve, generation);
    });
  }

  private _doAnimate(resolve: () => void, generation: number) {
    const collections = this.stackElement._sharedStackList;

    // The last seen location of a given card ID
    const idToPossibleCollection = new Map<string, CollectionRecord>();

    const collectionOffsets = new Map<string, OffsetRect>();

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
        // We reset this here, and not in prepare(), because we only want to
        // animate properties we set from here on out, and also all of the
        // physical components might not yet be created during prepare, for
        // example if a new component is added to a stack and a new one is
        // stamped. Calling this here makes sure they can all be
        // resetAnimating.
        component.resetAnimating();
      }
    }

    // This layout readback is the most important thing to do quickly
    // because if we thrash the DOM there will be a lot of recalc style. So
    // do it in its own pass.
    for (let i = 0; i < collections.length; i++) {
      const collection = collections[i];

      const offsetComponent = collection.offsetComponent;
      collectionOffsets.set(collection.id, this._calculateOffsets(offsetComponent));

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
        record.newOffsets = this._calculateOffsets(component);
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

        // CRITICAL: Transform composition order - invert + external + scale
        const invertTop = record.offsets!.top - record.newOffsets!.top;
        const invertLeft = record.offsets!.left - record.newOffsets!.left;
        let scaleFactor = record.offsets!.width / record.newOffsets!.width;

        // Defensive check: prevent crashes from zero-width/height calculations
        if (!isFinite(scaleFactor) || scaleFactor === 0) {
          scaleFactor = 1.0;
        }

        // If the before and after are rotated differently then the scale
        // factor will need to compare height vs width to get the right
        // scale factor.
        if (component.animationRotates(record.before, record.after)) {
          scaleFactor = record.offsets!.height / record.newOffsets!.width;
          // Defensive check: prevent crashes from zero-width/height calculations
          if (!isFinite(scaleFactor) || scaleFactor === 0) {
            scaleFactor = 1.0;
          }
        }

        // The containing box has physically shrunk (or grown), and the
        // transform will make its apparent edge be that much smaller or
        // bigger, so correct for that.
        let adjustedInvertTop = invertTop - (record.newOffsets!.height - record.offsets!.height) / 2;
        let adjustedInvertLeft = invertLeft - (record.newOffsets!.width - record.offsets!.width) / 2;

        // Determine whether the host element's CSS transform will actually
        // change during the FLIP animation. The browser only fires
        // transitionend when the computed value differs between the inverted
        // and final states. For components that didn't move position and whose
        // inline transform (e.g. messy stack rotation) is unchanged, the
        // inversion is effectively identity and the target matches — so no
        // transition fires and we must not expect one.
        const hasPositionChange = Math.abs(adjustedInvertTop) > 0.5 ||
          Math.abs(adjustedInvertLeft) > 0.5 || Math.abs(scaleFactor - 1) > 0.01;
        const hasInlineTransformChange =
          (record.beforeInlineTransform || '') !== (record.afterTransform || '');
        record.needsHostTransition = hasPositionChange || hasInlineTransformChange;

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
        const transform = `translateY(${adjustedInvertTop}px) translateX(${adjustedInvertLeft}px)`;
        const scaleTransform = `scale(${scaleFactor})`;
        const beforeInvertedTransform = `${transform} ${record.beforeTransform} ${scaleTransform}`;

        // Only prepare animation (set inverted transform, clone content) for
        // components that actually need animation. Non-animating components
        // skip the entire FLIP pipeline, avoiding spurious will-animate events.
        if (record.needsAnimation) {
          // TODO: what should opacity be?
          component.prepareAnimation(record.before, beforeInvertedTransform, '1.0');

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

      this._animatingComponents.push({
        stack: anonRecord.stack,
        component: component,
        after: record.after || {},
        afterTransform: component.style.transform,
        afterOpacity: component.style.opacity,
        needsHostTransition: true
      });

      const stackLocation = collectionOffsets.get(anonRecord.stack.id);
      const oldLocation = record.offsets;

      if (!stackLocation || !oldLocation) continue;

      let invertTop = oldLocation.top - stackLocation.top;
      let invertLeft = oldLocation.left - stackLocation.left;

      invertTop -= (stackLocation.height - oldLocation.height) / 2;
      invertLeft -= (stackLocation.width - oldLocation.width) / 2;

      let scaleFactor = oldLocation.width / stackLocation.width;

      // Defensive check: prevent crashes from zero-width/height calculations
      if (!isFinite(scaleFactor) || scaleFactor === 0) {
        scaleFactor = 1.0;
      }

      if (component.animationRotates(record.before, record.after)) {
        // The before and after are different rotations which means the
        // invert top and left have to be tweaked.
        scaleFactor = oldLocation.height / stackLocation.width;
        // Defensive check: prevent crashes from zero-width/height calculations
        if (!isFinite(scaleFactor) || scaleFactor === 0) {
          scaleFactor = 1.0;
        }
      }

      // We used to only bother setting transforms for items that had
      // physically moved. However, the browser is smart enough to ignore
      // transforms that are basically no ops. And if we don't set it
      // then cards that don't physically move but do have transform
      // changes won't animate because the transform was set during
      // noAnimate and is never set to anything different. In testing
      // this didn't appear to have any appreciable performance difference.
      const transform = `translateY(${invertTop}px) translateX(${invertLeft}px)`;
      const scaleTransform = `scale(${scaleFactor})`;

      const beforeInvertedTransform = `${transform} ${record.beforeTransform} ${scaleTransform}`;
      const beforeOpacity = '1.0';

      component.style.transform = beforeInvertedTransform;
      component.style.opacity = beforeOpacity;

      component.prepareAnimation(record.before, beforeInvertedTransform, beforeOpacity);

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

  private async _startAnimations(resolve: () => void, generation: number) {
    if (this._generation !== generation) { resolve(); return; }

    const collections = this.stackElement._sharedStackList;

    // Phase 1: Restore noAnimate on ALL components (required — was set during measurement)
    const allComponents: any[] = [];
    for (let i = 0; i < collections.length; i++) {
      const collection = collections[i];
      collection.noAnimate = false;
      const components = collection.Components;
      for (let j = 0; j < components.length; j++) {
        const component = components[j];
        if (component.id === '') continue;
        component.noAnimate = false;
        allComponents.push(component);
      }
    }

    // Also restore noAnimate on animating components (cross-stack overlays)
    for (const ac of this._animatingComponents) {
      ac.component.noAnimate = false;
      allComponents.push(ac.component);
    }

    // Phase 2: Wait for Lit to process noAnimate changes
    await Promise.all(allComponents.map(c => c.updateComplete));

    // Check generation after await — overlap timer may have started a new cycle
    if (this._generation !== generation) { resolve(); return; }

    // Phase 3: Build filtered list of components that actually need animation
    const componentsToAnimate: any[] = [];
    for (let i = 0; i < collections.length; i++) {
      const components = collections[i].Components;
      for (let j = 0; j < components.length; j++) {
        const component = components[j];
        if (component.id === '') continue;
        const record = this._infoById[component.id];
        if (!record || !record.needsAnimation) continue;
        componentsToAnimate.push({ component, record });
      }
    }

    // Animating components (cross-stack) always animate
    for (const ac of this._animatingComponents) {
      componentsToAnimate.push({ component: ac.component, record: ac });
    }

    // Phase 4: FLIP Play phase
    // CRITICAL ORDERING for CSS transitions to fire:
    //
    // prepareAnimation() set `style.transition = 'none'` and the inverted
    // transform on each component's host element. For the browser to animate
    // from the inverted position to the final position, it must:
    //   1. Commit the inverted position while transitions are disabled
    //   2. Re-enable transitions
    //   3. Set the final transform — browser sees the change and animates
    //
    // A reflow between steps 1-2 and 2-3 ensures the browser doesn't batch
    // the writes and skip the animation.

    // Step 1: Force layout to commit inverted transforms (transition: none still in effect)
    this.offsetHeight;

    // Step 2: Re-enable CSS transitions on the host element
    for (const item of componentsToAnimate) {
      item.component.style.transition = '';
    }

    // Step 3: Force layout so browser registers transition is now active
    this.offsetHeight;

    // Step 4: Set final transforms — browser sees change and animates
    for (const item of componentsToAnimate) {
      item.component.startAnimation(item.record.after, item.record.afterTransform, item.record.afterOpacity, item.record.needsHostTransition ?? true);
    }

    resolve();
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
