/**
 * boardgame-gathering-color-picker
 *
 * Shows color selection UI when a move with a SelectedColor field (EnumName:
 * "color") is legal. Shows color swatches instead of a dropdown.
 */
import { LitElement, html, css, nothing } from 'lit';
import { customElement, property } from 'lit/decorators.js';
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

// Map common color names to CSS colors. The behaviors.CSSColorForPlayer
// system handles the actual display colors, but we need a rough mapping
// for the picker swatches.
const COLOR_MAP: Record<string, string> = {
  red: '#e53935',
  blue: '#1e88e5',
  green: '#43a047',
  yellow: '#fdd835',
  orange: '#fb8c00',
  purple: '#8e24aa',
  pink: '#d81b60',
  brown: '#6d4c41',
  white: '#fafafa',
  black: '#212121',
  cyan: '#00acc1',
  teal: '#00897b',
};

function colorToCss(name: string): string {
  return COLOR_MAP[name.toLowerCase()] || '#9e9e9e';
}

@customElement('boardgame-gathering-color-picker')
export class BoardgameGatheringColorPicker extends LitElement {
  static styles = css`
    :host {
      display: var(--boardgame-gathering-color-picker-display, block);
    }
    h4 {
      margin: 0 0 8px 0;
      font-size: 14px;
      font-weight: 500;
      color: var(--md-sys-color-on-secondary-container, #1d192b);
    }
    .swatches {
      display: flex;
      flex-wrap: wrap;
      gap: 6px;
    }
    .swatch {
      width: 36px;
      height: 36px;
      border-radius: 50%;
      border: 2px solid transparent;
      cursor: pointer;
      transition: border-color 0.2s, transform 0.15s;
      position: relative;
    }
    .swatch:hover:not(.disabled):not(.selected) {
      transform: scale(1.1);
    }
    .swatch.selected {
      border-color: var(--md-sys-color-on-surface, #1c1b1f);
      transform: scale(1.15);
    }
    .swatch.claimed {
      opacity: 0.4;
      cursor: not-allowed;
    }
    .swatch.disabled {
      cursor: default;
    }
    .player-color-row {
      display: flex;
      align-items: center;
      gap: 8px;
      margin-bottom: 4px;
    }
    .player-name {
      font-size: 14px;
      min-width: 80px;
    }
    .color-label {
      font-size: 14px;
      color: var(--md-sys-color-on-surface-variant, #49454f);
      padding: 4px 12px;
      border-radius: 12px;
      display: inline-flex;
      align-items: center;
      gap: 6px;
    }
    .color-dot {
      width: 14px;
      height: 14px;
      border-radius: 50%;
      display: inline-block;
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

  private get _colorMoveForm(): MoveForm | null {
    if (!this.moveForms) return null;
    return this.moveForms.find(f =>
      f.LegalForAnyone &&
      f.Fields?.some((field: MoveFormField) => field.EnumName === 'color')
    ) ?? null;
  }

  private get _availableColors(): EnumValue[] {
    return this.state?.Game?.Computed?.Global?.AvailableColors || [];
  }

  private get _isVisible(): boolean {
    return !!this._colorMoveForm && this._availableColors.length > 0;
  }

  private _playerColorValue(playerIndex: number): string {
    const players = this.state?.Players;
    if (!players || !players[playerIndex]) return '';
    return players[playerIndex]?.Computed?.ColorValue || '';
  }

  /** Get the set of color names claimed by other seated players */
  private get _claimedColors(): Set<string> {
    const claimed = new Set<string>();
    for (let i = 0; i < this.playersInfo.length; i++) {
      if (this.playersInfo[i].IsEmpty) continue;
      if (i === this.viewingAsPlayer) continue;
      const val = this._playerColorValue(i);
      if (val) claimed.add(val);
    }
    return claimed;
  }

  private _handleColorClick(colorName: string): void {
    const moveForm = this._colorMoveForm;
    if (!moveForm || !moveForm.LegalForPlayer) return;

    this.dispatchEvent(new CustomEvent('propose-move', {
      composed: true,
      bubbles: true,
      detail: {
        name: moveForm.Name,
        arguments: {
          TargetPlayerIndex: String(this.viewingAsPlayer),
          SelectedColor: colorName,
        }
      }
    }));
  }

  render() {
    if (!this._isVisible) return nothing;

    const colors = this._availableColors;
    const isInteractive = this._colorMoveForm?.LegalForPlayer ?? false;
    const myColor = this._playerColorValue(this.viewingAsPlayer);
    const claimed = this._claimedColors;

    return html`
      <div>
        <h4>Color</h4>
        ${isInteractive ? html`
          <div class="swatches">
            ${colors.filter(c => c.Name).map(c => {
              const isClaimed = claimed.has(c.Name);
              const isSelected = c.Name === myColor;
              return html`
                <div
                  class="swatch ${isSelected ? 'selected' : ''} ${isClaimed ? 'claimed' : ''}"
                  style="background-color: ${colorToCss(c.Name)}"
                  title="${c.Name}${isClaimed ? ' (taken)' : ''}"
                  @click=${() => !isClaimed && this._handleColorClick(c.Name)}>
                </div>
              `;
            })}
          </div>
        ` : html`
          <div>
            ${this.playersInfo.map((p, i) => {
              if (p.IsEmpty) return nothing;
              const colorVal = this._playerColorValue(i);
              return html`
                <div class="player-color-row">
                  <span class="player-name">${p.DisplayName || `Player ${i}`}</span>
                  <span class="color-label">
                    ${colorVal ? html`
                      <span class="color-dot" style="background-color: ${colorToCss(colorVal)}"></span>
                      ${colorVal}
                    ` : 'Not set'}
                  </span>
                </div>
              `;
            })}
          </div>
        `}
      </div>
    `;
  }
}

declare global {
  interface HTMLElementTagNameMap {
    'boardgame-gathering-color-picker': BoardgameGatheringColorPicker;
  }
}
