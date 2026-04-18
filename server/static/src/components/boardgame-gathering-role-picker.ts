/**
 * boardgame-gathering-role-picker
 *
 * Shows role selection UI when a move with a SelectedRole field (EnumName:
 * "role") is legal. Same pattern as team picker.
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

  private get _roleMoveForm(): MoveForm | null {
    if (!this.moveForms) return null;
    return this.moveForms.find(f =>
      f.LegalForAnyone &&
      f.Fields?.some((field: MoveFormField) => field.Name === 'TargetPlayerIndex') &&
      f.Fields?.some((field: MoveFormField) => field.Name === 'SelectedRole' && field.EnumName === 'role')
    ) ?? null;
  }

  private get _availableRoles(): EnumValue[] {
    return this.state?.Game?.Computed?.Global?.AvailableRoles || [];
  }

  private get _isVisible(): boolean {
    return !!this._roleMoveForm && this._availableRoles.length > 0;
  }

  private _playerRoleValue(playerIndex: number): string {
    const players = this.state?.Players;
    if (!players || !players[playerIndex]) return '';
    return players[playerIndex]?.Computed?.RoleValue || '';
  }

  private _handleRoleChange(e: Event): void {
    const select = e.target as HTMLSelectElement;
    const selectedName = select.value;
    const moveForm = this._roleMoveForm;
    if (!moveForm || !selectedName) return;

    this.dispatchEvent(new CustomEvent('propose-move', {
      composed: true,
      bubbles: true,
      detail: {
        name: moveForm.Name,
        arguments: {
          TargetPlayerIndex: String(this.viewingAsPlayer),
          SelectedRole: selectedName,
        }
      }
    }));
  }

  render() {
    if (!this._isVisible) return nothing;

    const roles = this._availableRoles;
    const isInteractive = this._roleMoveForm?.LegalForPlayer ?? false;

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
