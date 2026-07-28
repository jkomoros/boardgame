/**
 * boardgame-gathering-role-picker
 *
 * Shows role selection UI when a SelectRole move is legal.
 * Same pattern as team picker.
 *
 * @fires propose-move - When the user selects a role
 * @csspart --boardgame-gathering-role-picker-display - Set to 'none' to hide
 */
import { LitElement, html, css, nothing } from 'lit';
import { customElement, property } from 'lit/decorators.js';
import '@material/web/select/filled-select.js';
import '@material/web/select/select-option.js';
import type { MoveForm } from '../types/api';
import { type EnumValue, type PlayerInfo, getAvailableValues, getPlayerComputedValue } from './gathering-shared.js';

@customElement('boardgame-gathering-role-picker')
export class BoardgameGatheringRolePicker extends LitElement {
  static styles = css`
    :host {
      display: var(--boardgame-gathering-role-picker-display, block);
    }
    .picker-row {
      display: flex;
      flex-wrap: wrap;
      gap: 8px;
      align-items: center;
    }
    .player-role {
      display: flex;
      align-items: center;
      gap: 8px;
    }
    .player-name {
      font-size: 14px;
      min-width: 80px;
    }
    .role-label {
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

  @property({ type: Object, attribute: 'move-form' })
  moveForm: MoveForm | null = null;

  @property({ type: Object })
  state: any = null;

  @property({ type: Number, attribute: 'viewing-as-player' })
  viewingAsPlayer = 0;

  @property({ type: Array, attribute: 'players-info' })
  playersInfo: PlayerInfo[] = [];

  private get _availableRoles(): EnumValue[] {
    return getAvailableValues(this.state, 'AvailableRoles');
  }

  private _playerRoleValue(playerIndex: number): string {
    return getPlayerComputedValue(this.state, playerIndex, 'RoleValue');
  }

  private _handleRoleChange(e: Event): void {
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
          SelectedRole: selectedName,
        }
      }
    }));
  }

  render() {
    if (!this.moveForm || this._availableRoles.length === 0) return nothing;

    const roles = this._availableRoles;
    const isInteractive = this.moveForm.LegalForPlayer ?? false;

    return html`
      <div>
        <h4>Role</h4>
        <div class="picker-row">
          ${this.playersInfo.map((p, i) => {
            if (p.IsEmpty) return nothing;
            const currentRole = this._playerRoleValue(i);
            const isMe = i === this.viewingAsPlayer;

            if (isMe && isInteractive) {
              return html`
                <div class="player-role">
                  <span class="player-name">${p.DisplayName || `Player ${i}`}</span>
                  <md-filled-select
                    label="Role"
                    @change=${this._handleRoleChange}
                    .value=${currentRole}>
                    <md-select-option value="">
                      <div slot="headline">Choose...</div>
                    </md-select-option>
                    ${roles.filter(r => r.Name).map(r => html`
                      <md-select-option
                        value=${r.Name}
                        ?selected=${r.Name === currentRole}>
                        <div slot="headline">${r.Name}</div>
                      </md-select-option>
                    `)}
                  </md-filled-select>
                </div>
              `;
            }

            return html`
              <div class="player-role">
                <span class="player-name">${p.DisplayName || `Player ${i}`}</span>
                <span class="role-label">${currentRole || 'Not set'}</span>
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
    'boardgame-gathering-role-picker': BoardgameGatheringRolePicker;
  }
}
