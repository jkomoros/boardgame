import { BoardgameComponent } from './boardgame-component.js';
import { html, css, TemplateResult } from 'lit';
import { property, query } from 'lit/decorators.js';

export class BoardgameCard extends BoardgameComponent {
  static override styles = [
    BoardgameComponent.styles,
    css`
      :host {
        /* Shadow elevation styles for rotated cards */
        --shadow-elevation-normal-rotated: 2px 0 2px 0 rgba(0, 0, 0, 0.14),
                                            1px 0 5px 0 rgba(0, 0, 0, 0.12),
                                            3px 0 1px -2px rgba(0, 0, 0, 0.2);

        --shadow-elevation-raised-rotated: 8px 0 10px 1px rgba(0, 0, 0, 0.14),
                                            3px 0 14px 2px rgba(0, 0, 0, 0.12),
                                            5px 0 5px -3px rgba(0, 0, 0, 0.4);

        --alt-shadow-elevation-normal-rotated: drop-shadow(2px 0 2px rgba(0, 0, 0, 0.14))
                                                drop-shadow(1px 0 5px rgba(0, 0, 0, 0.12))
                                                drop-shadow(3px 0 1px rgba(0, 0, 0, 0.2));

        --alt-shadow-elevation-raised-rotated: drop-shadow(8px 0 10px rgba(0, 0, 0, 0.14))
                                                drop-shadow(3px 0 14px rgba(0, 0, 0, 0.12))
                                                drop-shadow(5px 0 5px rgba(0, 0, 0, 0.4));
      }

      #outer {
        --default-component-width: 100px;
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
        font-size: 12px;
        line-height: 12px;
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

      #outer #front {
        background-color: #ccfcfc;
        z-index: 2;
        transform: rotateY(180deg);
      }

      #outer #back {
        background-color: #00cccc;
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
  basicRotated = false;

  @property({ type: Boolean })
  overrideRotated = false;

  @property({ type: Boolean })
  noContent = false;

  @property({ type: Boolean })
  tall = false;

  @property({ type: Number })
  aspectRatio = 0.6666666;

  @query('#front-slot')
  private frontSlot!: HTMLSlotElement;

  connectedCallback() {
    super.connectedCallback();
    this._updateInnerTransform();
  }

  protected override updated(changedProperties: Map<string, any>) {
    super.updated(changedProperties);

    if (changedProperties.has('faceUp') || changedProperties.has('rotated') || changedProperties.has('basicRotated') || changedProperties.has('overrideRotated')) {
      this._updateInnerTransform();
    }

    if (changedProperties.has('rotated')) {
      this._rotatedChanged(this.rotated);
    }

    if (
      changedProperties.has('noContent') ||
      changedProperties.has('rotated') ||
      changedProperties.has('tall')
    ) {
      this._updateClasses();
    }

    if (changedProperties.has('aspectRatio')) {
      this._outerStyle = this._computeOuterStyle(this.aspectRatio);
    }
  }

  override firstUpdated(_changedProperties: Map<PropertyKey, unknown>) {
    super.firstUpdated(_changedProperties);
    this._frontChanged();
    if (this.frontSlot) {
      this.frontSlot.addEventListener('slotchange', () => this._frontChanged());
    }
  }

  private _computeOuterStyle(aspectRatio: number): string {
    return `--component-aspect-ratio: ${aspectRatio};`;
  }

  override prepareForBeingAnimatingComponent(stack: any) {
    this.noContent = true;
    this.rotated = stack.stackDefault('rotated');
  }

  override get animatingProperties(): string[] {
    return super.animatingProperties.concat(['rotated', 'faceUp']);
  }

  override computeAnimationProps(isAfter: boolean, props: Record<string, any>): Record<string, any> {
    // We override these props for performance.
    // All of these set inner rotation on card, so do them all at once

    if (isAfter) {
      return {
        faceUp: props.faceUp,
        overrideRotated: false,
        basicRotated: props.rotated
      };
    }

    return {
      faceUp: props.faceUp,
      overrideRotated: true,
      basicRotated: props.rotated
    };
  }

  override get cloneContent(): boolean {
    return !this.noContent;
  }

  override animationRotates(beforeProps: Record<string, any>, afterProps: Record<string, any>): boolean {
    return beforeProps.rotated !== afterProps.rotated;
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

  private _rotatedChanged(newValue: boolean) {
    // there's a class of bugs where basicRotation isn't set the same as
    // rotation at beginning of rotation. The most recent one was when
    // moving a card that DIDN'T flip faceUp but did change from not
    // rotated to rotated, the first animation wouldn't work. To fix that,
    // we have basicRotated mirror rotated whenever rotated is explicitly
    // set, to verify basicRotated defaults to a reasonable value.
    this.basicRotated = newValue;
    this._updateInnerTransform();
  }

  private _updateInnerTransform() {
    if (!this.innerElement) return;

    const transformPieces: string[] = ['scale(var(--component-effective-scale))'];
    // Chrome Canary used to interpolate fine if you left out the 0deg
    // rotation term, but then broke. Setting it explicitly fixes the bug.
    transformPieces.push(this.faceUp ? 'rotateY(180deg)' : 'rotateY(0deg)');
    transformPieces.push(
      (this.overrideRotated ? this.basicRotated : this.rotated) ? 'rotate(90deg)' : 'rotate(0deg)'
    );
    const transform = transformPieces.join(' ') || 'none';
    this.innerElement.style.transform = transform;
    this._expectTransitionEnd(this.innerElement, 'transform');
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
  protected override _computeClasses(): string {
    const result: string[] = ['card'];
    if (this.rotated) {
      result.push('rotated');
    }
    if (this.noContent) {
      result.push('no-content');
    }
    result.push(this.tall ? 'tall' : 'wide');
    result.push(super._computeClasses());
    return result.join(' ');
  }

  override render(): TemplateResult {
    return html`
      <div id="outer" class="${this._computeClasses()}" @click="${this.handleTap}" style="${this._outerStyle}">
        <div id="inner">
          <div id="front">
            <div class="normal">
              <slot id="front-slot">
                <div id="top-rank">${this.suit}${this.rank}</div>
                <div id="bottom-rank">${this.suit}${this.rank}</div>
              </slot>
            </div>
            <div class="fallback">
              <slot name="fallback"></slot>
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
