import { LitElement, html, css } from 'lit';
import { customElement, property } from 'lit/decorators.js';

import { connect } from 'pwa-helpers/connect-mixin.js';
import { store } from '../store.js';
import {
  selectGamePlayersInfo,
  selectPlayerColors,
} from '../selectors.js';

import type { RootState, PlayerInfo } from '../types/store';

/**
 * An inline player identity badge for use in game renderers.
 *
 * Full mode (default): colored circle with initial + display name.
 * Compact mode: small colored circle with initial only.
 *
 * Usage:
 *   <boardgame-player-badge player-index="0"></boardgame-player-badge>
 *   <boardgame-player-badge player-index="1" compact></boardgame-player-badge>
 */
@customElement('boardgame-player-badge')
export class BoardgamePlayerBadge extends connect(store)(LitElement) {
  static styles = css`
    :host {
      display: inline-flex;
      align-items: center;
      gap: 4px;
      vertical-align: middle;
    }

    .avatar {
      display: inline-flex;
      align-items: center;
      justify-content: center;
      border-radius: 50%;
      color: white;
      font-weight: 600;
      text-transform: uppercase;
      flex-shrink: 0;
    }

    .avatar.full {
      width: 24px;
      height: 24px;
      font-size: 12px;
    }

    .avatar.compact {
      width: 16px;
      height: 16px;
      font-size: 9px;
    }

    .name {
      font-size: 13px;
      font-weight: 500;
      white-space: nowrap;
      overflow: hidden;
      text-overflow: ellipsis;
      max-width: 120px;
    }
  `;

  @property({ type: Number, attribute: 'player-index' })
  playerIndex = 0;

  @property({ type: Boolean })
  compact = false;

  @property({ type: Array, attribute: false })
  private _playersInfo: PlayerInfo[] = [];

  @property({ type: Array, attribute: false })
  private _playerColors: string[] = [];

  stateChanged(state: RootState): void {
    this._playersInfo = selectGamePlayersInfo(state);
    this._playerColors = selectPlayerColors(state);
  }

  private get _playerInfo(): PlayerInfo | null {
    return this._playersInfo[this.playerIndex] || null;
  }

  private get _color(): string {
    return this._playerColors[this.playerIndex] || '#757575';
  }

  private get _initial(): string {
    const name = this._playerInfo?.DisplayName || '';
    return name ? name[0] : String(this.playerIndex);
  }

  render() {
    const sizeClass = this.compact ? 'compact' : 'full';
    return html`
      <span
        class="avatar ${sizeClass}"
        style="background-color: ${this._color}">
        ${this._initial}
      </span>
      ${this.compact ? '' : html`
        <span class="name">${this._playerInfo?.DisplayName || `Player ${this.playerIndex}`}</span>
      `}
    `;
  }
}

declare global {
  interface HTMLElementTagNameMap {
    'boardgame-player-badge': BoardgamePlayerBadge;
  }
}
