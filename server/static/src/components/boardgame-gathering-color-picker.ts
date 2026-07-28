/**
 * boardgame-gathering-color-picker
 *
 * Shows color selection as accessible swatches when a SelectColor move is legal.
 *
 * @fires propose-move - When the user selects a color
 * @csspart --boardgame-gathering-color-picker-display - Set to 'none' to hide
 */
import { LitElement, html, css, nothing } from 'lit';
import { customElement, property } from 'lit/decorators.js';
import type { MoveForm } from '../types/api';
import { type EnumValue, type PlayerInfo, getAvailableValues, getPlayerComputedValue } from './gathering-shared.js';

/**
 * Get the CSS color for a color enum value. Prefers the CSSColor field
 * sent by the server (from behaviors.CSSColorForKey), which matches
 * the framework's canonical color mapping. Falls back to a neutral gray
 * if the server didn't provide a CSS color for this value.
 */
function colorToCss(colorValue: EnumValue): string {
  return colorValue.CSSColor || '#9e9e9e';
}

/**
 * Find the CSS color for a color name from the available colors list.
 * Used for the read-only player summary.
 */
function cssColorForName(name: string, availableColors: EnumValue[]): string {
  const match = availableColors.find(c => c.Name === name);
  return match?.CSSColor || '#9e9e9e';
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
      color: var(--md-sys-color-on-secondary-container, #271A10);
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
      padding: 0;
      background: none;
      outline: none;
      -webkit-appearance: none;
      appearance: none;
    }
    .swatch:focus-visible {
      outline: 2px solid var(--md-sys-color-primary, #2E6B4F);
      outline-offset: 2px;
    }
    @media (prefers-reduced-motion: reduce) {
      .swatch { transition: none; }
    }
    .swatch:hover:not([aria-disabled="true"]):not([aria-checked="true"]) {
      transform: scale(1.1);
    }
    .swatch[aria-checked="true"] {
      border-color: var(--md-sys-color-on-surface, #1C1810);
      transform: scale(1.15);
    }
    .swatch[aria-checked="true"]::after {
      content: "✓";
      position: absolute;
      top: 50%;
      left: 50%;
      transform: translate(-50%, -50%);
      font-size: 16px;
      font-weight: bold;
      color: white;
      text-shadow: 0 1px 2px rgba(0,0,0,.5);
    }
    .swatch[aria-disabled="true"] {
      opacity: 0.4;
      cursor: not-allowed;
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
      color: var(--md-sys-color-on-surface-variant, #4A4539);
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

  /** The resolved SelectColor move form (passed from gathering panel). */
  @property({ type: Object, attribute: 'move-form' })
  moveForm: MoveForm | null = null;

  @property({ type: Object })
  state: any = null;

  @property({ type: Number, attribute: 'viewing-as-player' })
  viewingAsPlayer = 0;

  @property({ type: Array, attribute: 'players-info' })
  playersInfo: PlayerInfo[] = [];

  private get _availableColors(): EnumValue[] {
    return getAvailableValues(this.state, 'AvailableColors');
  }

  private get _isVisible(): boolean {
    return !!this.moveForm && this._availableColors.length > 0;
  }

  private _playerColorValue(playerIndex: number): string {
    return getPlayerComputedValue(this.state, playerIndex, 'ColorValue');
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

  /** Keyboard navigation for the radiogroup (arrow keys move focus, Enter/Space selects) */
  private _handleSwatchKeydown(e: KeyboardEvent, colors: EnumValue[], currentName: string): void {
    const validColors = colors.filter(c => c.Name);
    const currentIdx = validColors.findIndex(c => c.Name === currentName);
    let nextIdx = -1;

    switch (e.key) {
      case 'ArrowRight':
      case 'ArrowDown':
        nextIdx = (currentIdx + 1) % validColors.length;
        break;
      case 'ArrowLeft':
      case 'ArrowUp':
        nextIdx = (currentIdx - 1 + validColors.length) % validColors.length;
        break;
      case ' ':
      case 'Enter':
        e.preventDefault();
        if (!this._claimedColors.has(currentName)) {
          this._handleColorClick(currentName);
        }
        return;
      default:
        return;
    }

    e.preventDefault();
    // Move focus to the next swatch
    const swatches = this.shadowRoot?.querySelectorAll('.swatch') as NodeListOf<HTMLElement>;
    if (swatches && swatches[nextIdx]) {
      swatches[nextIdx].focus();
    }
  }

  private _handleColorClick(colorName: string): void {
    const moveForm = this.moveForm;
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
    const isInteractive = this.moveForm?.LegalForPlayer ?? false;
    const myColor = this._playerColorValue(this.viewingAsPlayer);
    const claimed = this._claimedColors;

    return html`
      <div>
        <h4 id="color-heading">Color</h4>
        ${isInteractive ? html`
          <div class="swatches" role="radiogroup" aria-labelledby="color-heading">
            ${colors.filter(c => c.Name).map(c => {
              const isClaimed = claimed.has(c.Name);
              const isSelected = c.Name === myColor;
              return html`
                <button
                  class="swatch"
                  role="radio"
                  aria-checked=${isSelected ? 'true' : 'false'}
                  aria-disabled=${isClaimed ? 'true' : 'false'}
                  aria-label="${c.Name}${isClaimed ? ' (taken)' : ''}"
                  tabindex=${isSelected ? 0 : -1}
                  style="background-color: ${colorToCss(c)}"
                  @click=${() => !isClaimed && this._handleColorClick(c.Name)}
                  @keydown=${(e: KeyboardEvent) => this._handleSwatchKeydown(e, colors, c.Name)}>
                </button>
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
                      <span class="color-dot" style="background-color: ${cssColorForName(colorVal, this._availableColors)}"></span>
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
