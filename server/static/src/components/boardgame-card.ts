import { BoardgameComponent } from './boardgame-component.js';
import { html, css, TemplateResult } from 'lit';
import { property, query } from 'lit/decorators.js';
import { classMap } from 'lit/directives/class-map.js';
import { motionSilhouette } from '../motion/subject.js';
import type { MotionSubjectSnapshot } from '../motion/subject.js';
import type { VisualMotionTrackInput } from '../motion/component-track.js';

export class BoardgameCard extends BoardgameComponent {
  static override styles = [
    BoardgameComponent.styles,
    css`
      :host {
        /* Override component width for cards */
        --default-component-width: 100px;

        /* The DRAWN card's height/width, matching the aspectRatio property's
           default. _computeOuterStyle republishes the property's live value on
           #outer, which overrides this; the declaration is here so the very
           first paint -- before updated() has computed that inline style --
           still has a ratio to draw with. Deliberately NOT
           --component-aspect-ratio: that names the BOX's ratio and only works
           from :host or above. See the #inner rules. */
        --card-aspect-ratio: 0.6666666;

        /* Shadow elevation styles for rotated cards */
        --shadow-elevation-normal-rotated: 2px 0 2px 0 rgba(60, 40, 20, 0.14),
                                            1px 0 5px 0 rgba(60, 40, 20, 0.12),
                                            3px 0 1px -2px rgba(60, 40, 20, 0.2);

        --shadow-elevation-raised-rotated: 8px 0 10px 1px rgba(60, 40, 20, 0.14),
                                            3px 0 14px 2px rgba(60, 40, 20, 0.12),
                                            5px 0 5px -3px rgba(60, 40, 20, 0.4);

        --alt-shadow-elevation-normal-rotated: drop-shadow(2px 0 2px rgba(60, 40, 20, 0.14))
                                                drop-shadow(1px 0 5px rgba(60, 40, 20, 0.12))
                                                drop-shadow(3px 0 1px rgba(60, 40, 20, 0.2));

        --alt-shadow-elevation-raised-rotated: drop-shadow(8px 0 10px rgba(60, 40, 20, 0.14))
                                                drop-shadow(3px 0 14px rgba(60, 40, 20, 0.12))
                                                drop-shadow(5px 0 5px rgba(60, 40, 20, 0.4));
      }

      #outer {
        --card-effective-border-radius: 5px;
      }

      #outer div.fallback {
        display: none;
      }

      #outer.no-content div.normal {
        display: none;
      }

      #outer.no-content div.fallback {
        display: block;
      }

      #front {
        display: flex;
        flex-direction: column;
        align-items: center;
        justify-content: center;
      }

      #outer {
        height: var(--component-effective-height);
        width: var(--component-effective-width);
        display: flex;
        flex-direction: column;
        align-items: center;
        justify-content: center;
        perspective: 1000px;
      }

      #outer.tall {
        height: var(--component-effective-width);
        width: var(--component-effective-height);
      }

      #outer.rotated {
        height: var(--component-effective-width);
        width: var(--component-effective-height);
      }

      #outer.tall.rotated {
        height: var(--component-effective-height);
        width: var(--component-effective-width);
      }

      /*
       * THE CARD'S RATIO IS THE ART'S, NOT THE BOX'S, AND IT HAS ITS OWN NAME.
       *
       * --card-aspect-ratio used to be published as --component-aspect-ratio,
       * written inline on #outer from the aspectRatio property. Two things were
       * wrong with that, and only the second was doing any work.
       *
       * It never shaped the BOX. #outer's height comes from
       * --component-effective-height, and a custom property that REFERENCES
       * another is substituted where it is DECLARED --
       * --component-effective-height is declared at :host
       * (boardgame-component.ts), above #outer, so it was always substituted
       * with the :host ratio of 1.0 no matter what #outer said. Measured at
       * --component-width: 200px, #outer computed
       * --component-aspect-ratio: 0.6666666 and drew a 200x200 box; real
       * blackjack cards measured 105x105 hosts around a 103x71 card. This is
       * the same trap boardgame-token.ts describes at length for its deleted
       * per-shape rules.
       *
       * What it DID shape is this rule, which reads the ratio directly rather
       * than through --component-effective-height and so sees the value #outer
       * inherits down. That is the job worth keeping -- and the reason the
       * declaration was renamed instead of deleted the way the token's were.
       *
       * Under the old name it also SHADOWED the one place
       * --component-aspect-ratio is meant to be set and does work: :host or
       * above. A card handed --component-aspect-ratio: 1.5 on its host drew a
       * 200x300 box around a 100x66.7 card, box and art disagreeing, because
       * the inline write on #outer overwrote the author's value for everything
       * inside it. Under two names each value keeps its own job.
       *
       * The box itself stays square, for the reasons boardgame-token.ts spells
       * out: the board layout puts aspect-ratio: 1 on every component host,
       * boardgame-spatial-board's tokenPosition centres a piece at
       * coords - tokenSize / 2 in BOTH axes, and the stack's spread/fan margins
       * and the FLIP scale ratio all key off that one box. The card is already
       * drawn in true proportion inside it, exactly as a token's SVG is. See
       * tests/animations/parity/card-box.spec.ts.
       */
      #inner {
        width: var(--default-component-width);
        height: calc(var(--default-component-width) * var(--card-aspect-ratio));
        transform: scale(var(--component-effective-scale));
        border-radius: var(--card-effective-border-radius);
        transform-style: preserve-3d;
        position: absolute;
      }

      .tall #inner {
        height: var(--default-component-width);
        width: calc(var(--default-component-width) * var(--card-aspect-ratio));
      }

      #outer.shadow.rotated #inner {
        box-shadow: var(--shadow-elevation-normal-rotated);
      }

      #outer.shadow.interactive.rotated:hover #inner {
        box-shadow: var(--shadow-elevation-raised-rotated);
      }

      /* The rotated alt-shadow elevation, ON #outer -- never on #inner, for
         exactly the reason boardgame-component.ts spells out for the unrotated
         pair and boardgame-token.ts for its throb: motionTrackTarget('visual')
         returns #inner, so #inner is where a component-owned 3D scene mounts
         and where 'transform-style: preserve-3d' has to go, and a 'filter'
         forces 'transform-style: flat' on the element carrying it. A rotated
         card with altShadow set would therefore have been unable to host one.

         Nothing sets 'altShadow' on a card today -- not here, not in ../games
         -- so this pair was unreachable and the flattening was latent rather
         than live. That is the argument for moving it NOW: it costs nothing
         while nothing depends on its stacking, and #inner is exactly where a
         3D card would have to live.

         Visually inert for the same reason as the unrotated pair: #outer paints
         nothing of its own, so the alpha silhouette the drop-shadows derive
         from is the one #inner produced either way.

         .disabled's saturate is restated for the same reason
         boardgame-component.ts restates it: the two now share ONE filter slot
         on #outer, and this selector outranks '#outer.alt-shadow.disabled', so
         a disabled rotated card would otherwise silently stop looking
         disabled. Elevation first, then saturate -- the order the two-element
         version painted them. */
      #outer.alt-shadow.rotated {
        filter: var(--alt-shadow-elevation-normal-rotated);
      }

      #outer.alt-shadow.rotated.disabled {
        filter: var(--alt-shadow-elevation-normal-rotated) saturate(60%);
      }

      #outer.alt-shadow.interactive.rotated:hover {
        filter: var(--alt-shadow-elevation-raised-rotated);
      }

      #front,
      #back {
        height: 100%;
        width: 100%;
        position: absolute;
        top: 0;
        left: 0;
        backface-visibility: hidden;
        -webkit-backface-visibility: hidden;
        overflow: hidden;
        border-radius: var(--card-effective-border-radius);
      }

      #top-rank,
      #bottom-rank {
        position: absolute;
        /* Scale corner indices with the card so they stay readable at any
           size (--component-width comes from the surrounding view). The
           old fixed 12px was illegible on phone-sized cards. */
        font-size: max(11px, calc(var(--component-effective-width) * 0.17));
        line-height: 1;
        font-weight: 600;
        color: var(--card-ink-color, #1c2b22);
      }

      #top-rank {
        bottom: 5px;
        left: 5px;
        transform: rotate(-90deg);
      }

      #bottom-rank {
        right: 5px;
        top: 5px;
        transform: rotate(90deg);
      }

      /* Big center pip so a hand card reads at arm's length. Rotated 90°
         for the same reason the corner indices are: the card's natural
         frame is landscape; views display it upright via the rotated
         attribute. */
      #center-rank {
        position: absolute;
        inset: 0;
        display: flex;
        align-items: center;
        justify-content: center;
        font-size: calc(var(--component-effective-width) * 0.48);
        line-height: 1;
        font-weight: 700;
        color: var(--card-ink-color, #1c2b22);
        transform: rotate(90deg);
      }

      /* Classic red suits: ♥/♦ fronts carry the .red-suit class. */
      .red-suit #top-rank,
      .red-suit #bottom-rank,
      .red-suit #center-rank {
        color: var(--card-red-ink-color, #B3362B);
      }

      #outer #front {
        background-color: var(--card-front-color, #D4E8DA);
        box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.5);
        z-index: 2;
        transform: rotateY(180deg);
      }

      #outer #back {
        background-color: var(--card-back-color, #2E6B4F);
        color: var(--card-back-text-color, rgba(255, 255, 255, 0.6));
        box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.15),
                    inset 0 -1px 0 rgba(0, 0, 0, 0.1);
        transform: rotateY(0deg);
      }

      #default-back {
        height: 100%;
        width: 120%;
        opacity: 0.2;
        font-size: 13.5px;
        line-height: 14px;
        overflow: hidden;
        text-overflow: clip;
        user-select: none;
        display: flex;
        flex-direction: column;
        align-items: center;
        justify-content: center;
      }

      .tall #default-back {
        width: 130%;
      }
    `
  ];

  @property({ type: String })
  suit = '';

  @property({ type: String })
  rank = '';

  @property({ type: Boolean, attribute: 'face-up' })
  faceUp = false;

  @property({ type: Boolean, reflect: true })
  rotated = false;

  @property({ type: Boolean, attribute: 'no-content' })
  noContent = false;

  @property({ type: Boolean })
  tall = false;

  @property({ type: Number, attribute: 'aspect-ratio' })
  aspectRatio = 0.6666666;

  @query('#front-slot')
  private frontSlot!: HTMLSlotElement;

  override motionSubjectSnapshot(): MotionSubjectSnapshot {
    // Shape only: card face/back/content never crosses this boundary.
    return motionSilhouette('rounded-rectangle');
  }

  private _boundFrontChanged?: () => void;

  override connectedCallback() {
    super.connectedCallback();
    this._updateInnerTransform();
  }

  // Optimization opportunity: shouldUpdate() could be added here to prevent
  // unnecessary re-renders during animations. However, cards have complex state
  // with multiple interdependent properties (faceUp, rotated, noContent, tall,
  // aspectRatio) that affect visual output.
  // Conservative approach: Allow all renders to ensure correctness.
  // Future optimization: Skip renders when only non-visual properties change.

  protected override updated(changedProperties: Map<string, any>) {
    super.updated(changedProperties);

    if (changedProperties.has('faceUp') || changedProperties.has('rotated')) {
      this._updateInnerTransform();
    }

    if (changedProperties.has('rotated')) {
      this._rotatedChanged(this.rotated);
    }

    if (changedProperties.has('aspectRatio')) {
      this._outerStyle = this._computeOuterStyle(this.aspectRatio);
    }
  }

  override firstUpdated(_changedProperties: Map<PropertyKey, unknown>) {
    super.firstUpdated(_changedProperties);
    this._frontChanged();
    this._boundFrontChanged = () => this._frontChanged();
    if (this.frontSlot) {
      this.frontSlot.addEventListener('slotchange', this._boundFrontChanged);
    }
  }

  override disconnectedCallback() {
    super.disconnectedCallback();
    if (this._boundFrontChanged && this.frontSlot) {
      this.frontSlot.removeEventListener('slotchange', this._boundFrontChanged);
    }
  }

  /**
   * Publishes the ART's ratio, under a name that is only the art's. See the
   * `#inner` rules for why this is not `--component-aspect-ratio`: that one
   * names the BOX's ratio, is only readable from `:host` or above, and this
   * used to overwrite it.
   */
  private _computeOuterStyle(aspectRatio: number): string {
    return `--card-aspect-ratio: ${aspectRatio};`;
  }

  override prepareMotionCarrier(
    defaults: Readonly<Record<string, unknown>>,
    stack?: any,
  ): void {
    if (this.prepareForBeingAnimatingComponent
      !== BoardgameCard.prototype.prepareForBeingAnimatingComponent) {
      this.prepareForBeingAnimatingComponent(stack);
      return;
    }
    this.noContent = true;
    this.rotated = !!defaults.rotated;
  }

  /** @deprecated Compatibility adapter for pre-motion component callers. */
  override prepareForBeingAnimatingComponent(stack: any): void {
    this.noContent = true;
    this.rotated = !!stack?.stackDefault?.('rotated');
  }

  override get animatingProperties(): string[] {
    return super.animatingProperties.concat(['rotated', 'faceUp']);
  }

  // _innerTransformFor computes the resting inner transform for a given
  // faceUp/rotated combination — the pure function behind what
  // _updateInnerTransform writes as the resting style.
  private _innerTransformFor(faceUp: boolean, rotated: boolean): string {
    return [
      'scale(var(--component-effective-scale))',
      faceUp ? 'rotateY(180deg)' : 'rotateY(0deg)',
      rotated ? 'rotate(90deg)' : 'rotate(0deg)',
    ].join(' ');
  }

  protected override propertyMotionTracks(
    before: Record<string, any>,
    after: Record<string, any>,
  ): readonly VisualMotionTrackInput[] {
    if (before.faceUp === after.faceUp && before.rotated === after.rotated) return [];
    return [{
      target: 'visual',
      property: 'transform',
      from: this._innerTransformFor(!!before.faceUp, !!before.rotated),
      to: this._innerTransformFor(!!after.faceUp, !!after.rotated),
    }];
  }

  /** @deprecated Compatibility adapter; framework playback uses planned tracks. */
  override playPropertyAnimation(
    before: Record<string, any>,
    after: Record<string, any>,
    delayMs: number = 0,
  ): void {
    if (before.faceUp === after.faceUp && before.rotated === after.rotated) return;
    if (!this.innerElement) return;
    this.play(this.innerElement, [
      { transform: this._innerTransformFor(!!before.faceUp, !!before.rotated) },
      { transform: this._innerTransformFor(!!after.faceUp, !!after.rotated) },
    ], { delay: delayMs });
  }

  protected override shouldPlayLegacyPropertyAnimation(): boolean {
    return this.playPropertyAnimation !== BoardgameCard.prototype.playPropertyAnimation;
  }

  override get historicalPresentationPolicy(): 'none' | 'clone-default-slot' {
    return this.cloneContent ? 'clone-default-slot' : 'none';
  }

  /** @deprecated Compatibility adapter for pre-motion component callers. */
  override get cloneContent(): boolean {
    return !this.noContent;
  }

  override motionEndpointOrientation(
    state: Readonly<Record<string, unknown>>,
  ): 'natural' | 'quarter-turned' {
    return state.rotated ? 'quarter-turned' : 'natural';
  }

  /** @deprecated Compatibility adapter for pre-motion geometry callers. */
  override animationRotates(
    beforeProps: Record<string, any>,
    afterProps: Record<string, any>,
  ): boolean {
    return beforeProps.rotated !== afterProps.rotated;
  }

  /** Legacy subclass rotation policy is authoritative when overridden. */
  override legacyAnimationRotationRequested(
    beforeProps: Record<string, any>,
    afterProps: Record<string, any>,
  ): boolean | null {
    if (this.animationRotates === BoardgameCard.prototype.animationRotates) return null;
    return this.animationRotates(beforeProps, afterProps);
  }

  private _frontChanged() {
    if (!this.frontSlot) return;

    const nodes = this.frontSlot.assignedNodes();
    let newValue = false;
    for (let i = 0; i < nodes.length; i++) {
      const node = nodes[i];
      if (node.nodeType !== 1) continue;
      const element = node as Element;
      if (element.hasAttribute('tall')) {
        newValue = true;
      }
      if (element.hasAttribute('aspect-ratio')) {
        this.aspectRatio = parseFloat(element.getAttribute('aspect-ratio') || '0.6666666');
      }
    }
    this.tall = newValue;
  }

  private _rotatedChanged(_newValue: boolean) {
    this._updateInnerTransform();
  }

  private _updateInnerTransform() {
    if (!this.innerElement) return;
    this.innerElement.style.transform =
      this._innerTransformFor(this.faceUp, this.rotated) || 'none';
  }

  protected override _itemChanged(newValue: any) {
    if (newValue === undefined) return;
    if (newValue === null) {
      this.noContent = true;
      this.faceUp = false;
      super._itemChanged(newValue);
      return;
    }
    if (newValue.Values) {
      this.faceUp = true;
      this.noContent = false;
    } else {
      this.faceUp = false;
      this.noContent = true;
    }
    super._itemChanged(newValue);
  }

  // Override _computeClasses and add some more.
  protected override _computeClasses(): Record<string, boolean> {
    return {
      ...super._computeClasses(),
      card: true,
      rotated: this.rotated,
      'no-content': this.noContent,
      tall: this.tall,
      wide: !this.tall
    };
  }

  override render(): TemplateResult {
    return html`
      <div id="outer" class="${classMap(this._computeClasses())}" @click="${this.handleTap}" style="${this._outerStyle}">
        <div id="inner">
          <div id="front">
            <div class="normal ${this.suit === '♥' || this.suit === '♦' ? 'red-suit' : ''}">
              <slot id="front-slot">
                <div id="top-rank">${this.suit}${this.rank}</div>
                <div id="center-rank">${this.suit}</div>
                <div id="bottom-rank">${this.suit}${this.rank}</div>
              </slot>
            </div>
            <div class="fallback">
              <slot name="motion-history"><slot name="fallback"></slot></slot>
            </div>
          </div>
          <div id="back">
            <slot name="back">
              <div id="default-back">
                ★ ☆ ★ ☆ ★ ☆ ★ ☆ ★ ☆ ★ ☆ ★ ☆ ★ ☆ ★ ☆ ★ ☆ ★ ☆ ★ ☆ ★ ☆ ★ ☆ ★ ☆ ★ ☆ ★ ☆ ★ ☆
              </div>
            </slot>
          </div>
        </div>
      </div>
    `;
  }
}

customElements.define('boardgame-card', BoardgameCard);

declare global {
  interface HTMLElementTagNameMap {
    'boardgame-card': BoardgameCard;
  }
}
