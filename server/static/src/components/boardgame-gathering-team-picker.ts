/**
 * boardgame-gathering-team-picker
 *
 * Shows team selection UI when a move with a SelectedTeam field (EnumName:
 * "team") is legal. Detects via field signature, not move name.
 *
 * Interactive for the viewing player (if LegalForPlayer), read-only for others.
 */
import { LitElement, html, css, nothing } from 'lit';
import { customElement, property } from 'lit/decorators.js';
import '@material/web/select/filled-select.js';
import '@material/web/select/select-option.js';
import type { MoveForm, MoveFormField } from '../types/api';

interface EnumValue {
  Key: number;
  Name: string;
}

interface PlayerInfo {
  IsEmpty: boolean;
  IsAgent: boolean;
  DisplayName: string;
}

@customElement('boardgame-gathering-team-picker')
export class BoardgameGatheringTeamPicker extends LitElement {
  static styles = css`
    :host {
      display: var(--boardgame-gathering-team-picker-display, block);
    }
    .picker-row {
      display: flex;
      flex-wrap: wrap;
      gap: 8px;
      align-items: center;
    }
    .player-team {
      display: flex;
      align-items: center;
      gap: 8px;
    }
    .player-name {
      font-size: 14px;
      min-width: 80px;
    }
    .team-label {
      font-size: 14px;
      color: var(--md-sys-color-on-surface-variant, #49454f);
      padding: 8px 12px;
      background: var(--md-sys-color-surface-container, #f3edf7);
      border-radius: 8px;
    }
    h4 {
      margin: 0 0 8px 0;
      font-size: 14px;
      font-weight: 500;
      color: var(--md-sys-color-on-secondary-container, #1d192b);
    }
  `;

  @property({ type: Array })
  moveForms: MoveForm[] | null = null;

  @property({ type: Object })
  state: any = null;

  @property({ type: Number })
  viewingAsPlayer = 0;

  @property({ type: Array })
  playersInfo: PlayerInfo[] = [];

  /** Find the team selection move by field signature */
  private get _teamMoveForm(): MoveForm | null {
    if (!this.moveForms) return null;
    return this.moveForms.find(f =>
      f.LegalForAnyone &&
      f.Fields?.some((field: MoveFormField) => field.EnumName === 'team')
    ) ?? null;
  }

  private get _availableTeams(): EnumValue[] {
    return this.state?.Game?.Computed?.Global?.AvailableTeams || [];
  }

  private get _isVisible(): boolean {
    return !!this._teamMoveForm && this._availableTeams.length > 0;
  }

  private _playerTeamValue(playerIndex: number): string {
    const players = this.state?.Players;
    if (!players || !players[playerIndex]) return '';
    return players[playerIndex]?.Computed?.TeamValue || '';
  }

  private _handleTeamChange(e: Event): void {
    const select = e.target as HTMLSelectElement;
    const selectedName = select.value;
    const moveForm = this._teamMoveForm;
    if (!moveForm || !selectedName) return;

    this.dispatchEvent(new CustomEvent('propose-move', {
      composed: true,
      bubbles: true,
      detail: {
        name: moveForm.Name,
        arguments: {
          TargetPlayerIndex: String(this.viewingAsPlayer),
          SelectedTeam: selectedName,
        }
      }
    }));
  }

  render() {
    if (!this._isVisible) return nothing;

    const teams = this._availableTeams;
    const isInteractive = this._teamMoveForm?.LegalForPlayer ?? false;

    return html`
      <div>
        <h4>Team</h4>
        <div class="picker-row">
          ${this.playersInfo.map((p, i) => {
            if (p.IsEmpty) return nothing;
            const currentTeam = this._playerTeamValue(i);
            const isMe = i === this.viewingAsPlayer;

            if (isMe && isInteractive) {
              return html`
                <div class="player-team">
                  <span class="player-name">${p.DisplayName || `Player ${i}`}</span>
                  <md-filled-select
                    @change=${this._handleTeamChange}
                    .value=${currentTeam}>
                    <md-select-option value="">
                      <div slot="headline">Choose...</div>
                    </md-select-option>
                    ${teams.filter(t => t.Name).map(t => html`
                      <md-select-option
                        value=${t.Name}
                        ?selected=${t.Name === currentTeam}>
                        <div slot="headline">${t.Name}</div>
                      </md-select-option>
                    `)}
                  </md-filled-select>
                </div>
              `;
            }

            return html`
              <div class="player-team">
                <span class="player-name">${p.DisplayName || `Player ${i}`}</span>
                <span class="team-label">${currentTeam || 'Not set'}</span>
              </div>
            `;
          })}
        </div>
      </div>
    `;
  }
}

declare global {
  interface HTMLElementTagNameMap {
    'boardgame-gathering-team-picker': BoardgameGatheringTeamPicker;
  }
}
