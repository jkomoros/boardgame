/**
 * boardgame-gathering-team-picker
 *
 * Shows team selection UI when a SelectTeam move is legal. The parent panel
 * passes the resolved moveForm as a prop (detection centralized in the panel).
 *
 * Interactive for the viewing player (if LegalForPlayer), read-only for others.
 *
 * @fires propose-move - When the user selects a team
 * @csspart --boardgame-gathering-team-picker-display - Set to 'none' to hide
 */
import { LitElement, html, css, nothing } from 'lit';
import { customElement, property } from 'lit/decorators.js';
import '@material/web/select/filled-select.js';
import '@material/web/select/select-option.js';
import type { MoveForm } from '../types/api';
import { type EnumValue, type PlayerInfo, getAvailableValues, getPlayerComputedValue } from './gathering-shared.js';

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
      color: var(--md-sys-color-on-surface-variant, #4A4539);
      padding: 8px 12px;
      background: var(--md-sys-color-surface-container, #F0EBE3);
      border-radius: 8px;
    }
    h4 {
      margin: 0 0 8px 0;
      font-size: 14px;
      font-weight: 500;
    }
  `;

  /** The resolved SelectTeam move form (passed from gathering panel). */
  @property({ type: Object, attribute: 'move-form' })
  moveForm: MoveForm | null = null;

  @property({ type: Object })
  state: any = null;

  @property({ type: Number, attribute: 'viewing-as-player' })
  viewingAsPlayer = 0;

  @property({ type: Array, attribute: 'players-info' })
  playersInfo: PlayerInfo[] = [];

  private get _availableTeams(): EnumValue[] {
    return getAvailableValues(this.state, 'AvailableTeams');
  }

  private _playerTeamValue(playerIndex: number): string {
    return getPlayerComputedValue(this.state, playerIndex, 'TeamValue');
  }

  private _handleTeamChange(e: Event): void {
    const select = e.target as HTMLElement & { value: string };
    const selectedName = select.value;
    if (!this.moveForm || !selectedName) return;

    this.dispatchEvent(new CustomEvent('propose-move', {
      composed: true,
      bubbles: true,
      detail: {
        name: this.moveForm.Name,
        arguments: {
          TargetPlayerIndex: String(this.viewingAsPlayer),
          SelectedTeam: selectedName,
        }
      }
    }));
  }

  render() {
    if (!this.moveForm || this._availableTeams.length === 0) return nothing;

    const teams = this._availableTeams;
    const isInteractive = this.moveForm.LegalForPlayer ?? false;

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
                    label="Team"
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
