import { BoardgameAnimatableItem } from './boardgame-animatable-item.js';
import { html, css } from 'lit';
import { property } from 'lit/decorators.js';
import { query } from 'lit/decorators.js';
import { repeat } from 'lit/directives/repeat.js';
import { isBoundMoveAction, type BoundMoveAction } from '../moves/action.js';
import { componentMotionTracks } from '../motion/component-track.js';
import type { ComponentMotionTarget } from '../motion/component-track.js';

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
        border-radius: 6px;
        background: linear-gradient(135deg, #F5F0E8 0%, #E0D9CE 100%);
        overflow: hidden;
        cursor: pointer;
        box-shadow: 0 2px 4px 0 rgba(60, 40, 20, 0.18),
                    0 1px 5px 0 rgba(60, 40, 20, 0.12),
                    0 3px 1px -2px rgba(60, 40, 20, 0.2),
                    inset 0 1px 0 rgba(255, 255, 255, 0.4);
        transform: scale(var(--effective-die-scale));
        transition: box-shadow 0.28s cubic-bezier(0.4, 0, 0.2, 1);
        border: 0;
        padding: 0;
      }

      #main.interactive:hover {
        box-shadow: 0 8px 10px 1px rgba(60, 40, 20, 0.14),
                    0 3px 14px 2px rgba(60, 40, 20, 0.12),
                    0 5px 5px -3px rgba(60, 40, 20, 0.4);
      }

      #action-status {
        position: absolute;
        top: 100%;
        width: max-content;
        max-width: 16rem;
        margin-top: 0.25rem;
        color: var(--md-sys-color-error, #ba1a1a);
        font-size: 0.75rem;
      }

      #inner {
        position: relative;
        transform: translateY(calc(-1 * var(--effective-die-size) * var(--selected-face)));
        /* The spin is WAAPI-driven now; no CSS transform transition. */
      }

      .face {
        display: flex;
        flex-direction: column;
        align-items: center;
        justify-content: center;
        height: var(--effective-die-size);
        width: var(--effective-die-size);
        position: relative;
        font-family: var(--md-sys-typescale-body-large-font, 'Source Sans 3', sans-serif);
        font-size: 20px;
        font-weight: 500;
        line-height: 28px;
      }

      .pip {
        background-color: var(--md-sys-color-on-surface, #1C1810);
        height: var(--pip-size);
        width: var(--pip-size);
        border-radius: calc(var(--pip-size) / 2);
        position: absolute;
        display: none;
        box-shadow: inset 0 1px 2px rgba(0, 0, 0, 0.4),
                    0 1px 0 rgba(255, 255, 255, 0.2);
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

  @property({ attribute: false })
  action: BoundMoveAction<string, object> | null = null;

  @query('#inner')
  private _innerElement?: HTMLElement;

  private _boundHandleClick?: (e: Event) => void;
  private _unsubscribeAction: (() => void) | null = null;

  override connectedCallback() {
    super.connectedCallback();
    this._boundHandleClick ??= (e: Event) => this._handleClick(e);
    this.renderRoot.addEventListener('click', this._boundHandleClick);
    this._subscribeAction();
  }

  override disconnectedCallback() {
    if (this._boundHandleClick) {
      this.renderRoot.removeEventListener('click', this._boundHandleClick);
    }
    this._unsubscribeAction?.();
    this._unsubscribeAction = null;
    super.disconnectedCallback();
  }

  override updated(changedProperties: Map<PropertyKey, unknown>) {
    super.updated(changedProperties);

    if (changedProperties.has('selectedFace')) {
      this._selectedFaceChanged(
        this.selectedFace,
        changedProperties.get('selectedFace') as number | undefined
      );
    }

    if (changedProperties.has('item')) {
      this._itemChanged(this.item);
    }

    if (changedProperties.has('action')) {
      this._subscribeAction();
    }
  }

  private _handleClick(e: Event) {
    if (!isBoundMoveAction(this.action)) {
      if (this.action !== null) e.stopPropagation();
      return;
    }
    if (this.disabled || !this.action.canActivate) {
      e.stopPropagation();
      return;
    }
    e.stopPropagation();
    void this.action.activate();
  }

  private _subscribeAction(): void {
    this._unsubscribeAction?.();
    this._unsubscribeAction = isBoundMoveAction(this.action)
      ? this.action.subscribe(() => this.requestUpdate())
      : null;
  }

  // _innerTransformForFace mirrors the CSS resting transform on #inner for
  // a given selectedFace: translateY(-1 * effective-die-size * face). The
  // #main element carries the --selected-face var that drives the CSS
  // resting position; here we build an explicit transform so WAAPI can
  // interpolate the spin instead of relying on a CSS transition.
  private _innerTransformForFace(face: number): string {
    return `translateY(calc(-1 * var(--effective-die-size) * ${face}))`;
  }

  protected override motionTrackTarget(target: ComponentMotionTarget): HTMLElement | null {
    return target === 'host' ? this : this._innerElement ?? null;
  }

  private _selectedFaceChanged(newValue: number, oldValue: number | undefined) {
    if (!this._innerElement) return;
    // On first render there's no meaningful spin to animate from.
    if (oldValue === undefined || oldValue === newValue) return;
    this.playMotionTracks(componentMotionTracks([{
      target: 'visual',
      property: 'transform',
      from: this._innerTransformForFace(oldValue),
      to: this._innerTransformForFace(newValue),
    }]));
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
    const pieces: string[] = [];
    pieces.push(disabled ? 'disabled' : 'interactive');
    return pieces.join(' ');
  }

  override render() {
    const action = this.action;
    const bound = isBoundMoveAction(action);
    const interactive = bound;
    const effectiveDisabled = this.disabled || !interactive || (bound && !action.canActivate);
    const baseReason = bound
      ? action.reason?.message
      : action ? 'Bind required move input with .with(...)' : null;
    const reason = bound && action.preview.kind === 'failed' && action.preview.retryable
      ? `${baseReason ?? 'Move legality check failed'}. Activate to retry.`
      : baseReason;
    return html`
      <div id="scaler">
        <button
          id="main"
          type="button"
          aria-label=${interactive ? 'Roll die' : 'Die'}
          aria-describedby=${reason ? 'action-status' : ''}
          aria-busy=${String(bound && action.submission.kind === 'pending')}
          ?disabled=${effectiveDisabled}
          style="--selected-face:${this.selectedFace}"
          class="${this._classes(effectiveDisabled)}">
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
        </button>
        ${reason ? html`<span id="action-status" role="status">${reason}</span>` : ''}
      </div>
    `;
  }
}

customElements.define('boardgame-die', BoardgameDie);
