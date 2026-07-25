import { BoardgameAnimatableItem } from './boardgame-animatable-item.js';
import { html, css, CSSResult, TemplateResult } from 'lit';
import { property, query } from 'lit/decorators.js';
import { classMap } from 'lit/directives/class-map.js';
import { motionSilhouette } from '../motion/subject.js';
import type { MotionSubjectSnapshot } from '../motion/subject.js';
import { compileComponentMotionTracks } from '../motion/component-track.js';
import type {
  ComponentMotionTrack,
  ComponentMotionTarget,
  VisualMotionTrackInput,
} from '../motion/component-track.js';
import type { HistoricalPresentationPolicy } from '../motion/historical-presentation.js';
import type { MotionEndpointOrientation } from '../motion/endpoint-pose.js';
import type { AnimationTimingPolicy } from '../motion/timing.js';

// FlipRecord is the bundle the animator computes for each animating
// component and hands to playAnimation(). before/after are the
// animatingPropValues() snapshots on either side of the databinding;
// invertedTransform/finalTransform/opacity describe the host FLIP delta.
export interface FlipRecord {
  before: Record<string, any>;       // animatingPropValues() before
  after: Record<string, any>;        // animatingPropValues() after
  invertedTransform: string;         // FLIP inverted transform (animator-computed, includes beforeTransform + scale)
  finalTransform: string;            // final inline/messy transform ('' if none)
  beforeOpacity: string;             // '' treated as '1'
  finalOpacity: string;
  needsHostTransition: boolean;      // host transform keyframes worth playing
  delayMs?: number;                  // per-component start delay, from a stack's stagger attribute (#728)
  durationMs?: number;
  timingPolicy?: AnimationTimingPolicy;
  tracks?: readonly ComponentMotionTrack[]; // planned once; executor consumes exactly these channels
}

export class BoardgameComponent extends BoardgameAnimatableItem {
  static override styles: any = css`
    :host {
      --default-component-scale: 1.0;
      --component-aspect-ratio: 1.0;
      --default-component-width: 30px;
      --component-effective-scale: var(--component-scale, var(--default-component-scale));
      --component-effective-width: calc(var(--component-effective-scale) * var(--component-width, var(--default-component-width)));
      --component-effective-height: calc(var(--component-effective-width) * var(--component-aspect-ratio));
    }

    /* Shadow elevation styles - copied from paper-styles */
    :host {
      --shadow-elevation-normal: 0 2px 2px 0 rgba(60, 40, 20, 0.14),
                                  0 1px 5px 0 rgba(60, 40, 20, 0.12),
                                  0 3px 1px -2px rgba(60, 40, 20, 0.2);

      --shadow-elevation-raised: 0 8px 10px 1px rgba(60, 40, 20, 0.14),
                                  0 3px 14px 2px rgba(60, 40, 20, 0.12),
                                  0 5px 5px -3px rgba(60, 40, 20, 0.4);

      --alt-shadow-elevation-normal: drop-shadow(0 2px 2px rgba(60, 40, 20, 0.14))
                                      drop-shadow(0 1px 5px rgba(60, 40, 20, 0.12))
                                      drop-shadow(0 3px 1px rgba(60, 40, 20, 0.2));

      --alt-shadow-elevation-raised: drop-shadow(0 8px 10px rgba(60, 40, 20, 0.14))
                                      drop-shadow(0 3px 14px rgba(60, 40, 20, 0.12))
                                      drop-shadow(0 5px 5px rgba(60, 40, 20, 0.4));
    }

    .spacer {
      visibility: hidden;
    }

    #outer.interactive {
      cursor: pointer;
    }

    .disabled {
      filter: saturate(60%);
    }

    #outer {
      cursor: default;
    }

    #outer.shadow #inner {
      box-shadow: var(--shadow-elevation-normal);
    }

    #outer.alt-shadow #inner {
      filter: var(--alt-shadow-elevation-normal);
    }

    #outer {
      transition: transform 0.1s ease-in-out;
    }

    #outer.interactive:hover {
      transform: translateY(-0.25em);
    }

    #outer.shadow.interactive:hover #inner {
      box-shadow: var(--shadow-elevation-raised);
    }

    #outer.alt-shadow.interactive:hover #inner {
      filter: var(--alt-shadow-elevation-raised);
    }

    #inner {
      /* box-shadow/filter transitions are from paper-styles/shadow; the
         transform term is gone — flips are WAAPI-driven now. */
      transition: box-shadow 0.28s cubic-bezier(0.4, 0, 0.2, 1),
                  filter 0.28s cubic-bezier(0.4, 0, 0.2, 1);
    }
  `;

  @property({ type: Number })
  index = 0;

  @property({ type: Object })
  item: any = null;

  // reflect: true is LOAD-BEARING. This Lit property shadows the native
  // Element.id accessor; without reflection, `ele.id = x` stores the value
  // in Lit state but never writes the DOM id attribute — so
  // getElementById / #id selectors could never match any component, which
  // silently no-op'ed every id-based animateBetween lookup.
  @property({ type: String, reflect: true })
  id = '';

  @property({ type: Boolean })
  disabled = false;

  @property({ type: Boolean, attribute: 'boardgame-component', reflect: true })
  boardgameComponent = true;

  @property({ type: Boolean })
  spacer = false;

  @property({ type: Boolean })
  noShadow = false;

  @property({ type: Boolean })
  altShadow = false;

  @query('#inner')
  protected innerElement!: HTMLElement;

  protected _outerStyle = '';

  private _memoizedComposedPropertyDefinition: any = null;

  private _layoutTransform = '';

  // Tracks the setter's OWN in-flight layout animation so a mid-flight
  // retarget cancels it before starting the next one (mirrors
  // boardgame-token.ts's `_throb` pattern). Without this, play() does not
  // auto-cancel a prior play, so an interrupted retarget would leave two
  // live animations racing the host's transform instead of exactly one --
  // the retired CSS `transition: transform ...` never had that problem
  // because a CSS transition retarget always replaces itself in place.
  private _layoutTransformAnimation: Animation | null = null;

  // get/set layoutTransform is the self-animating replacement for the
  // stack's retired CSS `transition: transform var(--animation-length,
  // 0.25s) ease-in-out` (boardgame-component-stack.ts). NOTHING calls this
  // setter yet (Task 12 wires the stack's layout writes through it) -- this
  // is purely the mechanism.
  //
  // Semantics, chosen to match the CSS transition it replaces:
  //  - Setting the SAME value is a no-op: no style write, no play(). (A CSS
  //    transition never restarts on an unchanged authored value either.)
  //  - Setting a DIFFERENT value snaps `this.style.transform` to it
  //    immediately (so layout/hit-testing/etc. always see the true final
  //    value, exactly like an authored CSS property does under a
  //    transition), then -- unless suppressed -- plays a gated host
  //    animation from the PRE-snap *computed* transform (not the previous
  //    setter argument) to the new computed transform. Capturing the
  //    computed value is what gives CSS-transition-style retargeting parity
  //    when interrupted mid-flight: a CSS transition that gets a new target
  //    while running continues from wherever the box actually is on screen,
  //    never from the stale authored target of the animation it interrupts.
  //  - Suppressed (matches `.no-animate`'s effect on the CSS transition):
  //    `noAnimate`, disconnected, or the computed transform did not actually
  //    change (e.g. an equivalent-but-differently-spelled value) all skip
  //    the play and leave only the snap.
  //  - `{ timing: 'immediate' }`: a layout tweak is a local, one-off
  //    visual correction -- the CSS transition it replaces had no notion of
  //    a shared render-game version slot either. Same reasoning as the
  //    fading-text / game-outcome fixes (see boardgame-fading-text.ts /
  //    boardgame-game-outcome.ts).
  //
  // Note for future readers: the animator's OWN direct `style.transform`
  // writes during FLIP (boardgame-component-animator.ts) intentionally
  // bypass this setter -- they write the property directly, not through
  // `layoutTransform`. This setter mediates only the stack's per-layout
  // writes; it is not a general interceptor of every transform mutation.
  get layoutTransform(): string {
    return this._layoutTransform;
  }

  set layoutTransform(value: string) {
    if (value === this._layoutTransform) return;
    const before = this.isConnected ? getComputedStyle(this).transform : 'none';
    this._layoutTransform = value;
    // Cancel our own prior layout animation (if any) BEFORE probing the
    // final computed value below: while it is still running, a fill:'none'
    // WAAPI animation overrides getComputedStyle() with its own live
    // sample, so reading "after" first would just re-observe the outgoing
    // animation's current frame (indistinguishable from `before`, since no
    // wall-clock time passes between the two synchronous reads) instead of
    // the true resting value the new authored style resolves to -- making
    // every mid-flight retarget silently no-op and leaving the ORIGINAL
    // animation running past its own retarget. play() does not auto-cancel
    // a prior play, so this is also what keeps a retarget to exactly one
    // live animation (mirrors boardgame-token.ts's `_throb` pattern).
    this._layoutTransformAnimation?.cancel();
    this._layoutTransformAnimation = null;
    this.style.transform = value;
    // noAnimate suppresses the self-play (snap only). During an animator
    // cycle the stack's relayout write lands from Lit's slotchange/updated
    // pass, which runs microtasks BEFORE the animator raises its
    // component-level noAnimate barrier -- so noAnimate is FALSE here and the
    // setter self-plays concurrently with the same cycle's FLIP. That is
    // deliberate parity: the retired CSS `transition: transform
    // var(--animation-length) ease-in-out` this setter replaces fired at the
    // same pre-barrier slotchange moment with identical easing/duration.
    // Verified by geometry golden geometry-debuganimations-fan-draw (evidence
    // 2026-07-26-stack-transition-cutover.md). noAnimate only snaps writes
    // issued WHILE the barrier is up (the animator's own measurement-time
    // style mutations).
    if (this.noAnimate || !this.isConnected) return;
    const after = getComputedStyle(this).transform;
    if (before === after) return;
    this._layoutTransformAnimation = this.play(
      this,
      [{ transform: before }, { transform: after }],
      { easing: 'ease-in-out' },
      { timing: 'immediate' },
    );
  }

  get interactive(): boolean {
    return !this.spacer && !this.disabled;
  }

  // animatingProperties should return an array of strings of property
  // names that change during animations. animatingPropValues() and
  // animatingPropDefaults() will use this.
  get animatingProperties(): string[] {
    return [];
  }

  /**
   * Privacy-safe capability for overlay decoration that follows this component.
   * Override with null to opt out; never return DOM or hidden game content.
   */
  motionSubjectSnapshot(): MotionSubjectSnapshot | null {
    return motionSilhouette('rectangle');
  }

  // Returns the bundle of properties, as configured by
  // animatingProperties(), at their current value.
  animatingPropValues(): Record<string, any> {
    const result: Record<string, any> = {};
    for (const propName of this.animatingProperties) {
      result[propName] = (this as any)[propName];
    }
    return result;
  }

  // Returns the bundle of animating properties, as defined by
  // animatingProperties(), set to the defaults for the given stack. Used
  // when there isn't an element analog before or after the animation to
  // compare to.
  animatingPropDefaults(stack: any): Record<string, any> {
    const result: Record<string, any> = {};
    for (const propName of this.animatingProperties) {
      result[propName] = stack.stackDefault(propName);
    }
    return result;
  }

  /** Purely describe every host and component-owned channel that will play. */
  planMotionTracks(rec: FlipRecord): readonly ComponentMotionTrack[] {
    // An overridden imperative hook was the complete property-motion owner on
    // the legacy Card path. Do not layer a modern default visual track beneath
    // it: a no-op override historically suppressed the default flip entirely.
    const visualTracks = this.shouldPlayLegacyPropertyAnimation()
      ? []
      : this.propertyMotionTracks(rec.before, rec.after);
    const tracks = compileComponentMotionTracks({
      needsHostTransition: rec.needsHostTransition,
      invertedTransform: rec.invertedTransform,
      finalTransform: rec.finalTransform,
      beforeOpacity: rec.beforeOpacity,
      finalOpacity: rec.finalOpacity,
      visualTracks,
    });
    // Legacy playback treated property targets independently: a temporarily
    // unavailable inner surface skipped its effect but never cancelled valid
    // host travel. Filter unavailable component-owned channels at planning so
    // the published track list remains the exact executable set.
    return Object.freeze(tracks.filter(track => this.motionTrackTarget(track.target)));
  }

  /** Subclasses describe visual consequences without starting WAAPI. */
  protected propertyMotionTracks(
    _before: Record<string, any>,
    _after: Record<string, any>,
  ): readonly VisualMotionTrackInput[] {
    return [];
  }

  /**
   * @deprecated Override propertyMotionTracks() so motion can be planned and
   * observed. Kept as an opaque playback adapter for existing components.
   */
  playPropertyAnimation(
    _before: Record<string, any>,
    _after: Record<string, any>,
    _delayMs: number = 0,
  ): void {
    // Legacy subclasses may start their own gated animations here.
  }

  /** Whether a legacy property hook owns additional opaque visual work. */
  protected shouldPlayLegacyPropertyAnimation(): boolean {
    return this.propertyMotionTracks === BoardgameComponent.prototype.propertyMotionTracks
      && this.playPropertyAnimation !== BoardgameComponent.prototype.playPropertyAnimation;
  }

  /** Internal bridge: opaque legacy property work cannot become a fake track. */
  legacyPropertyMotionRequested(
    before: Record<string, any>,
    after: Record<string, any>,
  ): boolean {
    const ownsLegacyPlayback = this.shouldPlayLegacyPropertyAnimation()
      || this.playAnimation !== BoardgameComponent.prototype.playAnimation;
    return ownsLegacyPlayback
      && this.animatingProperties.some(property => before[property] !== after[property]);
  }

  protected motionTrackTarget(target: ComponentMotionTarget): HTMLElement | null {
    return target === 'host' ? this : this.innerElement ?? null;
  }

  // Execute the already-planned track descriptions through the one shared
  // timing/gating kernel. Planning and playback no longer independently guess
  // which component property transitions exist.
  playAnimation(rec: FlipRecord): readonly Animation[] {
    const delayMs = rec.delayMs ?? 0;
    const result = this.playMotionTracks(
      rec.tracks ?? this.planMotionTracks(rec),
      { delay: delayMs, ...(rec.durationMs === undefined ? {} : { duration: rec.durationMs }) },
      { timing: rec.timingPolicy ?? 'version' },
    );
    // Host tracks are overlays (fill:'none'); authored resting styles remain
    // the final source of truth after WAAPI settles.
    if (rec.needsHostTransition) this.style.transform = rec.finalTransform;
    this.style.opacity = rec.finalOpacity;
    // A component that adopted propertyMotionTracks owns its visual channels
    // declaratively. Otherwise preserve the old imperative customization
    // point; animations it starts still join settled() through play().
    if (this.shouldPlayLegacyPropertyAnimation()) {
      this.playPropertyAnimation(rec.before, rec.after, delayMs);
    }
    return result.status === 'started'
      ? Object.freeze(result.playbacks.map(playback => playback.animation))
      : Object.freeze([]);
  }

  /** Prepare a fresh, inert component host to carry departing motion. */
  prepareMotionCarrier(
    _defaults: Readonly<Record<string, unknown>>,
    stack?: any,
  ): void {
    this.prepareForBeingAnimatingComponent(stack);
  }

  /** @deprecated Override prepareMotionCarrier(). */
  prepareForBeingAnimatingComponent(_stack: any): void {
    // Legacy subclasses may configure a faux component.
  }

  /** Opt in only to cloning already-rendered default-slot presentation. */
  get historicalPresentationPolicy(): HistoricalPresentationPolicy {
    return this.cloneContent ? 'clone-default-slot' : 'none';
  }

  /** @deprecated Override historicalPresentationPolicy. */
  get cloneContent(): boolean {
    return false;
  }

  /** Finite box-axis fact consumed by structural geometry. */
  motionEndpointOrientation(_state: Readonly<Record<string, unknown>>): MotionEndpointOrientation {
    return 'natural';
  }

  /**
   * @deprecated Override motionEndpointOrientation(). The animator still
   * consults this pairwise hook for components that have not migrated.
   */
  animationRotates(
    _beforeProps: Record<string, any>,
    _afterProps: Record<string, any>,
  ): boolean {
    return false;
  }

  /** Null means endpoint orientation owns the modern policy. */
  legacyAnimationRotationRequested(
    beforeProps: Record<string, any>,
    afterProps: Record<string, any>,
  ): boolean | null {
    if (this.animationRotates === BoardgameComponent.prototype.animationRotates) return null;
    return this.animationRotates(beforeProps, afterProps);
  }

  handleTap(e: Event) {
    if (!this.interactive) {
      return;
    }
    this.dispatchEvent(new CustomEvent('component-tapped', { composed: true, bubbles: true, detail: { index: this.index } }));
  }

  protected override updated(changedProperties: Map<string, any>) {
    super.updated(changedProperties);

    if (changedProperties.has('item')) {
      this._itemChanged(this.item);
    }
  }

  protected _itemChanged(newValue: any) {
    if (newValue === undefined) return;
    if (newValue === null) {
      this.spacer = true;
      return;
    }
    this.spacer = false;
    this.id = newValue.ID || '';
  }

  protected _computeClasses(): Record<string, boolean> {
    return {
      spacer: this.spacer,
      shadow: !this.noShadow && !this.altShadow,
      'alt-shadow': !this.noShadow && this.altShadow,
      interactive: this.interactive,
      disabled: this.disabled,
      'no-animate': this.noAnimate
    };
  }

  // obj.properties, smooshed down all the way to the upper.
  get _composedPropertyDefinition(): any {
    // TODO: can we get rid of this? Doesn't seem to be used, and I believe
    // Lit does this for us now.
    if (!this._memoizedComposedPropertyDefinition) {
      const result: any = {};
      let obj: any = this;
      while (obj) {
        const props = obj.constructor.properties;
        if (!props) break;
        for (const key of Object.keys(props)) {
          result[key] = props[key];
        }
        obj = Object.getPrototypeOf(obj);
      }
      this._memoizedComposedPropertyDefinition = result;
    }
    return this._memoizedComposedPropertyDefinition;
  }

  override render(): TemplateResult {
    return html`
      <div id="outer" class="${classMap(this._computeClasses())}" @click="${this.handleTap}" style="${this._outerStyle}">
        <div id="inner">
          <slot></slot>
        </div>
      </div>
    `;
  }
}

customElements.define('boardgame-component', BoardgameComponent);
