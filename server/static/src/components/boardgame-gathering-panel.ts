/**
 * boardgame-gathering-panel
 *
 * A framework-provided component that renders gathering UI (waiting status,
 * share link, team/color/role pickers, start button) between the player roster
 * and the game renderer. It auto-hides when it has nothing to show.
 *
 * Detection is driven by move form legality and game state — no explicit
 * "lobby mode" flag. The panel detects gathering moves by their field
 * signatures (TargetPlayerIndex + Selected{Team,Role,Color} fields) and
 * passes resolved move forms to each sub-component as props.
 *
 * Override system:
 * - CSS: set --boardgame-gathering-{team,role,color}-picker-display: none
 * - Full: register boardgame-render-gathering-GAMENAME as a custom element
 */
import { LitElement, html, css, nothing } from 'lit';
import { customElement, property } from 'lit/decorators.js';
import type { MoveForm } from '../types/api';
import {
  type PlayerInfo,
  OBSERVER_PLAYER_INDEX,
  findStartMoveForm,
  findTeamMoveForm,
  findRoleMoveForm,
  findColorMoveForm,
  getReadyToStartError,
} from './gathering-shared.js';

import './boardgame-gathering-status.js';
import './boardgame-gathering-share.js';
import './boardgame-gathering-start.js';
import './boardgame-gathering-team-picker.js';
import './boardgame-gathering-role-picker.js';
import './boardgame-gathering-color-picker.js';

interface GameRoute {
  name: string;
  id: string;
}

@customElement('boardgame-gathering-panel')
export class BoardgameGatheringPanel extends LitElement {
  static styles = css`
    :host {
      display: block;
    }

    .join-code {
      font-size: 16px;
      letter-spacing: 1px;
    }
    .join-code strong {
      font-size: 20px;
      letter-spacing: 4px;
    }

    .panel {
      background: linear-gradient(180deg, var(--md-sys-color-primary-container, #D4E8DA) 0%, #C8DECC 100%);
      color: var(--md-sys-color-on-primary-container, #0A2818);
      padding: 16px;
      margin: 8px 0;
      border-radius: 12px;
      box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.4),
                  0 1px 3px 0 rgba(60, 40, 20, 0.08);
    }

    .panel-content {
      display: flex;
      flex-direction: column;
      gap: 12px;
    }

    .row {
      display: flex;
      flex-wrap: wrap;
      align-items: center;
      gap: 12px;
    }

    .spacer {
      flex: 1;
    }
  `;

  @property({ type: Array, attribute: 'move-forms' })
  moveForms: MoveForm[] | null = null;

  @property({ type: Object })
  state: any = null;

  @property({ type: Number, attribute: 'viewing-as-player' })
  viewingAsPlayer = 0;

  @property({ type: Boolean, attribute: 'has-empty-slots' })
  hasEmptySlots = false;

  @property({ type: Boolean, attribute: 'game-open' })
  gameOpen = false;

  @property({ type: Boolean })
  finished = false;

  @property({ type: Object, attribute: 'game-route' })
  gameRoute: GameRoute | null = null;

  @property({ type: Array, attribute: 'players-info' })
  playersInfo: PlayerInfo[] = [];

  // ---- Derived state (centralized detection) ----

  private get _readyToStartError(): string {
    return getReadyToStartError(this.state);
  }

  private get _isObserver(): boolean {
    return this.viewingAsPlayer === OBSERVER_PLAYER_INDEX;
  }

  private get _showStatus(): boolean {
    return (this.hasEmptySlots && this.gameOpen && !this.finished) ||
           !!this._readyToStartError;
  }

  /**
   * The room code for companion (Table+Hand) games, empty otherwise. When
   * set, the share affordance becomes "Join code: XXXX" instead of the
   * solo-flow "Copy invite link" — companion players join by code on
   * their phones, not by account-bound URL.
   */
  @property({ type: String, attribute: 'companion-room-code' })
  companionRoomCode = '';

  private get _showShare(): boolean {
    return this.hasEmptySlots && this.gameOpen && !this._isObserver && !this.companionRoomCode;
  }

  private get _showJoinCode(): boolean {
    return this.hasEmptySlots && !this.finished && !!this.companionRoomCode;
  }

  // Centralized move detection — resolve once, pass to children as props
  private get _startMoveForm(): MoveForm | null {
    return findStartMoveForm(this.moveForms);
  }

  private get _teamMoveForm(): MoveForm | null {
    return findTeamMoveForm(this.moveForms);
  }

  private get _roleMoveForm(): MoveForm | null {
    return findRoleMoveForm(this.moveForms);
  }

  private get _colorMoveForm(): MoveForm | null {
    return findColorMoveForm(this.moveForms);
  }

  private get _hasAnythingToShow(): boolean {
    return this._showStatus || this._showShare ||
           !!this._startMoveForm ||
           !!this._teamMoveForm || !!this._roleMoveForm || !!this._colorMoveForm;
  }

  render() {
    if (!this._hasAnythingToShow) {
      return nothing;
    }

    return html`
      <div class="panel">
        <div class="panel-content">
          <div class="row">
            ${this._showStatus ? html`
              <boardgame-gathering-status
                .playersInfo=${this.playersInfo}
                .hasEmptySlots=${this.hasEmptySlots}
                .gameOpen=${this.gameOpen}
                .finished=${this.finished}
                .readyToStartError=${this._readyToStartError}
                .startMoveForm=${this._startMoveForm}>
              </boardgame-gathering-status>
            ` : nothing}

            <span class="spacer"></span>

            ${this._showShare ? html`
              <boardgame-gathering-share
                .gameRoute=${this.gameRoute}>
              </boardgame-gathering-share>
            ` : nothing}

            ${this._showJoinCode ? html`
              <span class="join-code">Join code: <strong>${this.companionRoomCode}</strong></span>
            ` : nothing}

            ${this._startMoveForm ? html`
              <boardgame-gathering-start
                .moveForm=${this._startMoveForm}
                .viewingAsPlayer=${this.viewingAsPlayer}>
              </boardgame-gathering-start>
            ` : nothing}
          </div>

          ${this._teamMoveForm ? html`
            <boardgame-gathering-team-picker
              .moveForm=${this._teamMoveForm}
              .state=${this.state}
              .viewingAsPlayer=${this.viewingAsPlayer}
              .playersInfo=${this.playersInfo}>
            </boardgame-gathering-team-picker>
          ` : nothing}

          ${this._roleMoveForm ? html`
            <boardgame-gathering-role-picker
              .moveForm=${this._roleMoveForm}
              .state=${this.state}
              .viewingAsPlayer=${this.viewingAsPlayer}
              .playersInfo=${this.playersInfo}>
            </boardgame-gathering-role-picker>
          ` : nothing}

          ${this._colorMoveForm ? html`
            <boardgame-gathering-color-picker
              .moveForm=${this._colorMoveForm}
              .state=${this.state}
              .viewingAsPlayer=${this.viewingAsPlayer}
              .playersInfo=${this.playersInfo}>
            </boardgame-gathering-color-picker>
          ` : nothing}
        </div>
      </div>
    `;
  }
}

declare global {
  interface HTMLElementTagNameMap {
    'boardgame-gathering-panel': BoardgameGatheringPanel;
  }
}
