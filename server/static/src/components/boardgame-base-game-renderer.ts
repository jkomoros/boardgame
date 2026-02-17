import { LitElement, html } from 'lit';
import { property } from 'lit/decorators.js';
import type { MoveLegalityInfo } from '../selectors.js';

export class BoardgameBaseGameRenderer extends LitElement {
  @property({ type: Object })
  state: any = null;

  @property({ type: Object })
  chest: any = null;

  @property({ type: String })
  diagram = '';

  @property({ type: Number })
  viewingAsPlayer = 0;

  @property({ type: Number })
  currentPlayerIndex = 0;

  /**
   * Map of move name → legality info, set by boardgame-render-game from the
   * Redux store. Renderers should use the convenience helpers
   * isMoveCurrentlyLegal() and isMovePossible() instead of reading this
   * directly.
   */
  @property({ type: Object })
  moveLegality: Record<string, MoveLegalityInfo> = {};

  get isCurrentPlayer(): boolean {
    if (this.viewingAsPlayer === -2) return true;
    return this.currentPlayerIndex === this.viewingAsPlayer;
  }

  /**
   * Returns true if the named move is legal for the viewing player right now.
   * Use this to disable buttons when a move can't be made (e.g. not your turn).
   */
  isMoveCurrentlyLegal(moveName: string): boolean {
    return this.moveLegality[moveName]?.legalForPlayer ?? false;
  }

  /**
   * Returns true if the named move is structurally possible right now (legal
   * for any player / admin). Use this to hide buttons entirely when a move
   * isn't applicable in the current game phase.
   */
  isMovePossible(moveName: string): boolean {
    return this.moveLegality[moveName]?.legalForAnyone ?? false;
  }

  private _boundHandleButtonTapped?: (e: Event) => void;

  override firstUpdated(_changedProperties: Map<PropertyKey, unknown>) {
    super.firstUpdated(_changedProperties);
    this._boundHandleButtonTapped = (e: Event) => this._handleButtonTapped(e);

    // CHANGED: tap → click (Polymer event → standard event)
    this.addEventListener('click', this._boundHandleButtonTapped);
    this.addEventListener('component-tapped', this._boundHandleButtonTapped);
  }

  override disconnectedCallback() {
    super.disconnectedCallback();
    if (this._boundHandleButtonTapped) {
      this.removeEventListener('click', this._boundHandleButtonTapped);
      this.removeEventListener('component-tapped', this._boundHandleButtonTapped);
    }
  }

  // animationLength is consulted when applying an animation to configure the
  // animation length (in milliseconds) by setting `--animation-length` on the
  // renderer. Zero will specify default animation length (that is, unset an
  // override style). A negative return value will skip the animation entirely.
  // The default one returns 0 for all combinations. See also delayAnimation.
  animationLength(fromMove: any, toMove: any): number {
    return 0;
  }

  // delayAnimation will be consulted when applying an animation. It will delay
  // by the returned number of milliseconds. The default one returns 0 for all
  // combinations. See also animationLength.
  delayAnimation(fromMove: any, toMove: any): number {
    return 0;
  }

  private _handleButtonTapped(e: Event): void {
    const composedPath = e.composedPath();
    let ele: HTMLElement | null = null;

    for (const tempEle of composedPath) {
      // Runtime type check (no unsafe casts)
      if (!(tempEle instanceof Element)) continue;
      if (!tempEle.hasAttribute) continue;

      const proposeMove = (tempEle as any).proposeMove || tempEle.getAttribute('propose-move');
      if (proposeMove) {
        // found it!
        ele = tempEle as HTMLElement;
        break;
      }
    }

    if (!ele) {
      return;
    }

    if (ele.hasAttribute('boardgame-component') && e.type === 'click') {
      // Cards we'll fire on the component-tapped, not the click.
      return;
    }

    const moveName = (ele as any).proposeMove || ele.getAttribute('propose-move');
    if (!moveName) return;

    const data = ele.dataset;
    const args: Record<string, any> = {};

    for (const key in data) {
      if (!Object.prototype.hasOwnProperty.call(data, key)) continue;
      if (!key.startsWith('arg')) continue;
      let effectiveKey = key.replace('arg', '');
      // Handle the case where the attribute was literally just data-arg
      if (!effectiveKey) continue;
      // The first character is now upperCase, which is desired as per Move field convention
      args[effectiveKey] = data[key];
    }

    this.dispatchEvent(new CustomEvent('propose-move', {
      composed: true,
      bubbles: true,
      detail: { name: moveName, arguments: args }
    }));
  }

  override render() {
    return html``;
  }
}

customElements.define('boardgame-base-game-renderer', BoardgameBaseGameRenderer);
