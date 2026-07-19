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

      #inner {
        width: var(--default-component-width);
        height: calc(var(--default-component-width) * var(--component-aspect-ratio));
        transform: scale(var(--component-effective-scale));
        border-radius: var(--card-effective-border-radius);
        transform-style: preserve-3d;
        position: absolute;
      }

      .tall #inner {
        height: var(--default-component-width);
        width: calc(var(--default-component-width) * var(--component-aspect-ratio));
      }

      #outer.shadow.rotated #inner {
        box-shadow: var(--shadow-elevation-normal-rotated);
      }

      #outer.shadow.interactive.rotated:hover #inner {
        box-shadow: var(--shadow-elevation-raised-rotated);
      }

      #outer.alt-shadow.rotated #inner {
        filter: var(--alt-shadow-elevation-normal-rotated);
      }

      #outer.alt-shadow.interactive.rotated:hover #inner {
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

  @property({ type: Boolean })
  faceUp = false;

  @property({ type: Boolean, reflect: true })
  rotated = false;

  @property({ type: Boolean })
  noContent = false;

  @property({ type: Boolean })
  tall = false;

  @property({ type: Number })
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

  private _computeOuterStyle(aspectRatio: number): string {
    return `--component-aspect-ratio: ${aspectRatio};`;
  }

  override prepareMotionCarrier(defaults: Readonly<Record<string, unknown>>): void {
    this.noContent = true;
    this.rotated = !!defaults.rotated;
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

  override get historicalPresentationPolicy(): 'none' | 'clone-default-slot' {
    return this.noContent ? 'none' : 'clone-default-slot';
  }

  override motionEndpointOrientation(
    state: Readonly<Record<string, unknown>>,
  ): 'natural' | 'quarter-turned' {
    return state.rotated ? 'quarter-turned' : 'natural';
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
