import { BoardgameAnimatableItem } from './boardgame-animatable-item.js';
import { html, css } from 'lit';
import { property } from 'lit/decorators.js';
import { query } from 'lit/decorators.js';
import { repeat } from 'lit/directives/repeat.js';

class BoardgameDie extends BoardgameAnimatableItem {
  static override styles = [
    ...(BoardgameAnimatableItem.styles ? [BoardgameAnimatableItem.styles] : []),
    css`
      :host {
        --effective-die-scale: var(--die-scale, 1.0);
        --effective-die-size: 50px;
        --pip-size: 7px;
      }

      #scaler {
        height: calc(var(--effective-die-size) * var(--effective-die-scale));
        width: calc(var(--effective-die-size) * var(--effective-die-scale));
        position: relative;
        display: flex;
        flex-direction: column;
        align-items: center;
        justify-content: center;
      }

      #main.disabled {
        cursor: default;
      }

      #main {
        height: var(--effective-die-size);
        width: var(--effective-die-size);
        border-radius: 2px;
        background-color: #CCC;
        overflow: hidden;
        cursor: pointer;
        box-shadow: 0 2px 2px 0 rgba(0, 0, 0, 0.14),
                    0 1px 5px 0 rgba(0, 0, 0, 0.12),
                    0 3px 1px -2px rgba(0, 0, 0, 0.2);
        transform: scale(var(--effective-die-scale));
        transition: transform var(--animation-length) ease-in-out, box-shadow 0.28s cubic-bezier(0.4, 0, 0.2, 1);
      }

      #main.interactive:hover {
        box-shadow: 0 8px 10px 1px rgba(0, 0, 0, 0.14),
                    0 3px 14px 2px rgba(0, 0, 0, 0.12),
                    0 5px 5px -3px rgba(0, 0, 0, 0.4);
      }

      #inner {
        position: relative;
        transform: translateY(calc(-1 * var(--effective-die-size) * var(--selected-face)));
        transition: transform var(--animation-length) ease-in-out;
      }

      .face {
        display: flex;
        flex-direction: column;
        align-items: center;
        justify-content: center;
        height: var(--effective-die-size);
        width: var(--effective-die-size);
        position: relative;
        font-family: 'Roboto', 'Noto', sans-serif;
        font-size: 20px;
        font-weight: 500;
        line-height: 28px;
      }

      .pip {
        background-color: black;
        height: var(--pip-size);
        width: var(--pip-size);
        border-radius: calc(var(--pip-size) / 2);
        position: absolute;
        display: none;
      }

      .face.one span, .face.two span, .face.three span, .face.four span, .face.five span, .face.six span {
        display: none;
      }

      .pip.mid {
        top: calc(var(--effective-die-size) / 2 - var(--pip-size) / 2);
      }

      .pip.center {
        left: calc(var(--effective-die-size) / 2 - var(--pip-size) / 2);
      }

      .pip.top {
        top: calc(var(--effective-die-size) / 2 - var(--pip-size) * 1.5 - var(--pip-size) / 2);
      }

      .pip.left {
        left: calc(var(--effective-die-size) / 2 - var(--pip-size) * 1.5 - var(--pip-size) / 2);
      }

      .pip.bottom {
        top: calc(var(--effective-die-size) / 2 + var(--pip-size) * 1.5 - var(--pip-size) / 2);
      }

      .pip.right {
        left: calc(var(--effective-die-size) / 2 + var(--pip-size) * 1.5 - var(--pip-size) / 2);
      }

      .face.one .pip.mid.center {
        display: block;
      }

      .face.two .pip.top.right, .face.two .pip.bottom.left {
        display: block;
      }

      .face.three .pip.top.right, .face.three .pip.mid.center, .face.three .pip.bottom.left {
        display: block;
      }

      .face.four .pip.top.right, .face.four .pip.top.left, .face.four .pip.bottom.left, .face.four .pip.bottom.right {
        display: block;
      }

      .face.five .pip.top.right, .face.five .pip.top.left, .face.five .pip.bottom.left, .face.five .pip.bottom.right, .face.five .pip.mid.center {
        display: block;
      }

      .face.six .pip.top.right, .face.six .pip.top.left, .face.six .pip.bottom.left, .face.six .pip.bottom.right, .face.six .pip.mid.left, .face.six .pip.mid.right {
        display: block;
      }
    `
  ];

  @property({ type: Object })
  item: any = null;

  @property({ type: Number })
  value = 0;

  @property({ type: Array })
  faces: number[] = [];

  @property({ type: Number })
  selectedFace = 0;

  @property({ type: Boolean })
  disabled = false;

  @query('#inner')
  private _innerElement?: HTMLElement;

  private _boundHandleClick?: (e: Event) => void;

  override firstUpdated(_changedProperties: Map<PropertyKey, unknown>) {
    super.firstUpdated(_changedProperties);

    this._boundHandleClick = (e: Event) => this._handleClick(e);
    this.renderRoot.addEventListener('click', this._boundHandleClick);
  }

  override disconnectedCallback() {
    super.disconnectedCallback();
    if (this._boundHandleClick) {
      this.renderRoot.removeEventListener('click', this._boundHandleClick);
    }
  }

  override updated(changedProperties: Map<PropertyKey, unknown>) {
    super.updated(changedProperties);

    if (changedProperties.has('selectedFace')) {
      this._selectedFaceChanged(this.selectedFace);
    }

    if (changedProperties.has('item')) {
      this._itemChanged(this.item);
    }
  }

  private _handleClick(e: Event) {
    if (this.disabled) {
      e.stopPropagation();
    }
  }

  private _selectedFaceChanged(newValue: number) {
    if (this._innerElement) {
      this._expectTransitionEnd(this._innerElement, 'transform');
    }
  }

  private _itemChanged(newValue: any) {
    if (!newValue) {
      this.faces = [];
      this.selectedFace = 0;
      this.value = 0;
      return;
    }
    this.faces = newValue.Values.Faces;
    this.selectedFace = newValue.DynamicValues.SelectedFace;
    this.value = newValue.DynamicValues.Value;
  }

  private _classForFace(face: number): string {
    let str = '';
    switch (face) {
      case 1:
        str = 'one';
        break;
      case 2:
        str = 'two';
        break;
      case 3:
        str = 'three';
        break;
      case 4:
        str = 'four';
        break;
      case 5:
        str = 'five';
        break;
      case 6:
        str = 'six';
        break;
    }

    return 'face ' + str;
  }

  private _classes(disabled: boolean): string {
    const pieces = [];
    pieces.push(disabled ? 'disabled' : 'interactive');
    return pieces.join(' ');
  }

  override render() {
    return html`
      <div id="scaler">
        <div id="main" style="--selected-face:${this.selectedFace}" class="${this._classes(this.disabled)}">
          <div id="inner">
            ${repeat(this.faces, (face) => face, (face) => html`
              <div class="${this._classForFace(face)}">
                <span>${face}</span>
                <div class="pip mid center"></div>
                <div class="pip top left"></div>
                <div class="pip top right"></div>
                <div class="pip bottom left"></div>
                <div class="pip bottom right"></div>
                <div class="pip mid left"></div>
                <div class="pip mid right"></div>
              </div>
            `)}
          </div>
        </div>
      </div>
    `;
  }
}

customElements.define('boardgame-die', BoardgameDie);
