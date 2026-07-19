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
    return compileComponentMotionTracks({
      needsHostTransition: rec.needsHostTransition,
      invertedTransform: rec.invertedTransform,
      finalTransform: rec.finalTransform,
      beforeOpacity: rec.beforeOpacity,
      finalOpacity: rec.finalOpacity,
      visualTracks: this.propertyMotionTracks(rec.before, rec.after),
    });
  }

  /** Subclasses describe visual consequences without starting WAAPI. */
  protected propertyMotionTracks(
    _before: Record<string, any>,
    _after: Record<string, any>,
  ): readonly VisualMotionTrackInput[] {
    return [];
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
    return result.status === 'started'
      ? Object.freeze(result.playbacks.map(playback => playback.animation))
      : Object.freeze([]);
  }

  /** Prepare a fresh, inert component host to carry departing motion. */
  prepareMotionCarrier(_defaults: Readonly<Record<string, unknown>>): void {
    // Do nothing; subclasses might do something.
  }

  /** Opt in only to cloning already-rendered default-slot presentation. */
  get historicalPresentationPolicy(): HistoricalPresentationPolicy {
    return 'none';
  }

  /** Finite box-axis fact consumed by structural geometry. */
  motionEndpointOrientation(_state: Readonly<Record<string, unknown>>): MotionEndpointOrientation {
    return 'natural';
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
