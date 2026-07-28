/**
 * boardgame-gathering-status
 *
 * Shows gathering status: "Waiting for N more players", "Ready to start",
 * or the ReadyToStart error message.
 */
import { LitElement, html, css, nothing } from 'lit';
import { customElement, property } from 'lit/decorators.js';
import type { MoveForm } from '../types/api';
import type { PlayerInfo } from '../types/store';

@customElement('boardgame-gathering-status')
export class BoardgameGatheringStatus extends LitElement {
  static styles = css`
    :host {
      display: inline-block;
    }
    .status {
      font-family: var(--md-sys-typescale-body-medium-font, 'Source Sans 3', sans-serif);
      font-size: var(--md-sys-typescale-body-medium-size, 14px);
      color: var(--md-sys-color-on-secondary-container, #271A10);
    }
    .error {
      color: var(--md-sys-color-error, #BA1A1A);
    }
  `;

  @property({ type: Array, attribute: 'players-info' })
  playersInfo: PlayerInfo[] = [];

  @property({ type: Boolean, attribute: 'has-empty-slots' })
  hasEmptySlots = false;

  @property({ type: Boolean, attribute: 'game-open' })
  gameOpen = false;

  @property({ type: Boolean })
  finished = false;

  @property({ type: String, attribute: 'ready-to-start-error' })
  readyToStartError = '';

  @property({ type: Object, attribute: 'start-move-form' })
  startMoveForm: MoveForm | null = null;

  private get _emptyCount(): number {
    return this.playersInfo.filter(p => p.IsEmpty && !p.IsAgent).length;
  }

  private get _statusText(): string {
    if (this.readyToStartError) {
      return this.readyToStartError;
    }
    if (this.hasEmptySlots && this.gameOpen) {
      const empty = this._emptyCount;
      if (empty === 1) {
        return 'Waiting for 1 more player';
      }
      return `Waiting for ${empty} more players`;
    }
    if (this.startMoveForm?.LegalForAnyone) {
      return 'Ready to start';
    }
    return '';
  }

  render() {
    const text = this._statusText;
    if (!text) return nothing;

    const isError = !!this.readyToStartError;
    return html`
      <span class="status ${isError ? 'error' : ''}">${text}</span>
    `;
  }
}

declare global {
  interface HTMLElementTagNameMap {
    'boardgame-gathering-status': BoardgameGatheringStatus;
  }
}
