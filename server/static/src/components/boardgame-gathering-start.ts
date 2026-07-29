/**
 * boardgame-gathering-start
 *
 * Shows a "Start Game" button that proposes the CloseAllSeats / "Confirm
 * Players" move. Enabled when the move is LegalForPlayer. Shows the
 * LegalForPlayerError as a tooltip/subtitle when disabled.
 */
import { LitElement, html, css, nothing } from 'lit';
import { customElement, property } from 'lit/decorators.js';
import '@material/web/button/filled-button.js';
import type { MoveForm } from '../types/api';
import { OBSERVER_PLAYER_INDEX } from './gathering-shared.js';

@customElement('boardgame-gathering-start')
export class BoardgameGatheringStart extends LitElement {
  static styles = css`
    :host {
      display: inline-block;
    }
    .error-hint {
      font-size: 12px;
      color: var(--md-sys-color-error, #BA1A1A);
      margin-top: 4px;
      max-width: 250px;
    }
  `;

  @property({ type: Object, attribute: 'move-form' })
  moveForm: MoveForm | null = null;

  @property({ type: Number, attribute: 'viewing-as-player' })
  viewingAsPlayer = 0;

  private get _isLegal(): boolean {
    return this.moveForm?.LegalForPlayer ?? false;
  }

  private get _errorHint(): string {
    if (this._isLegal) return '';
    return this.moveForm?.LegalForPlayerError || '';
  }

  private get _isObserver(): boolean {
    return this.viewingAsPlayer === OBSERVER_PLAYER_INDEX;
  }

  private _handleClick(): void {
    if (!this.moveForm || !this._isLegal) return;
    this.dispatchEvent(new CustomEvent('propose-move', {
      composed: true,
      bubbles: true,
      detail: {
        name: this.moveForm.Name,
        arguments: {}
      }
    }));
  }

  render() {
    if (!this.moveForm) return nothing;
    // Don't show the button to observers
    if (this._isObserver) return nothing;

    return html`
      <md-filled-button
        @click=${this._handleClick}
        ?disabled=${!this._isLegal}
        title=${this._errorHint}>
        Start Game
      </md-filled-button>
      ${this._errorHint ? html`
        <div class="error-hint">${this._errorHint}</div>
      ` : nothing}
    `;
  }
}

declare global {
  interface HTMLElementTagNameMap {
    'boardgame-gathering-start': BoardgameGatheringStart;
  }
}
