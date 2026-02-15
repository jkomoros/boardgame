import { LitElement, html, css } from 'lit';
import { customElement, property, query } from 'lit/decorators.js';
import { when } from 'lit/directives/when.js';
import '@material/web/radio/radio.js';
import '@material/web/checkbox/checkbox.js';
import './boardgame-move-form.ts';

import type { MdRadio } from '@material/web/radio/radio.js';
import type { MdCheckbox } from '@material/web/checkbox/checkbox.js';
import type { BoardgameMoveForm } from './boardgame-move-form.ts';

interface GameRoute {
  name?: string;
  id?: string;
}

interface Game {
  NumPlayers: number;
  CurrentPlayerIndex: number;
  Finished: boolean;
}

interface MoveForm {
  Name: string;
  HelpText: string;
  Fields?: unknown[];
}

@customElement('boardgame-admin-controls')
export class BoardgameAdminControls extends LitElement {
  static styles = css`
    :host {
      display: block;
    }

    .horizontal {
      display: flex;
      flex-direction: row;
    }

    .layout {
      display: flex;
    }

    .center {
      align-items: center;
    }

    .flex {
      flex: 1;
    }

    .card {
      background: var(--md-sys-color-surface-container-low, #f7f2fa);
      padding: 16px;
      margin: 8px 0;
      border-radius: 12px;
      box-shadow: var(--md-sys-elevation-1, 0 1px 3px 1px rgba(0,0,0,.15), 0 1px 2px rgba(0,0,0,.3));
      color: var(--md-sys-color-on-surface, #1c1b1f);
    }

    .admin {
      gap: 16px;
    }

    md-radio {
      margin-right: 4px;
    }

    label.radio-label {
      display: inline-flex;
      align-items: center;
      gap: 4px;
      font-family: var(--md-sys-typescale-body-medium-font, 'Roboto', sans-serif);
      font-size: var(--md-sys-typescale-body-medium-size, 14px);
      color: var(--md-sys-color-on-surface, #1c1b1f);
      cursor: pointer;
    }

    input[type="number"] {
      width: 60px;
      margin-left: 8px;
      padding: 8px 12px;
      border: 1px solid var(--md-sys-color-outline, #79747e);
      border-radius: 4px;
      font-family: var(--md-sys-typescale-body-medium-font, 'Roboto', sans-serif);
      color: var(--md-sys-color-on-surface, #1c1b1f);
      background: var(--md-sys-color-surface, #fffbfe);
    }

    [role="radiogroup"] {
      display: inline-flex;
      gap: 12px;
      margin-left: 8px;
    }

    .view-as-label {
      font-family: var(--md-sys-typescale-label-large-font, 'Roboto', sans-serif);
      font-size: var(--md-sys-typescale-label-large-size, 14px);
      font-weight: var(--md-sys-typescale-label-large-weight, 500);
      color: var(--md-sys-color-on-surface-variant, #49454f);
    }

    details {
      margin-bottom: 4px;
    }

    summary {
      cursor: pointer;
      padding: 8px 12px;
      border-radius: 8px;
      font-family: var(--md-sys-typescale-body-large-font, 'Roboto', sans-serif);
      font-size: var(--md-sys-typescale-body-large-size, 16px);
      color: var(--md-sys-color-on-surface, #1c1b1f);
      background: var(--md-sys-color-surface-container, #f0edf1);
    }

    summary:hover {
      background: var(--md-sys-color-surface-container-high, #e6e0e9);
    }

    pre {
      padding: 12px;
      border-radius: 8px;
      background: var(--md-sys-color-surface-container, #f0edf1);
      color: var(--md-sys-color-on-surface, #1c1b1f);
      font-size: 12px;
      overflow: auto;
      max-height: 400px;
    }
  `;

  @property({ type: Boolean })
  active = false;

  @property({ type: Object })
  gameRoute: GameRoute | null = null;

  @property({ type: String })
  viewAs: 'custom' | 'admin' | 'current' | 'observer' = 'current';

  @property({ type: Number })
  customRequestedPlayer = 0;

  @property({ type: Boolean })
  makeMovesAsViewingAsPlayer = true;

  @property({ type: Number })
  viewingAsPlayer = 0;

  @property({ type: Object })
  chest: unknown = null;

  @property({ type: Object })
  currentState: unknown = null;

  @property({ type: Array })
  moveForms: MoveForm[] = [];

  @property({ type: Object })
  game: Game | null = null;

  @query('#moves')
  private movesElement!: BoardgameMoveForm;

  private readonly OBSERVER_PLAYER_INDEX = -1;
  private readonly ADMIN_PLAYER_INDEX = -2;

  constructor() {
    super();
    // Listen for propose-move events to forward to the move form
    this.addEventListener('propose-move', (e: Event) => this._handleProposeMove(e as CustomEvent));
  }

  get requestedPlayer(): number {
    if (!this.active) return this.viewingAsPlayer;
    switch (this.viewAs) {
      case 'admin':
        return this.ADMIN_PLAYER_INDEX;
      case 'observer':
        return this.OBSERVER_PLAYER_INDEX;
      case 'custom':
        return this.customRequestedPlayer;
      case 'current':
        return this.game?.CurrentPlayerIndex || 0;
      default:
        return 0;
    }
  }

  get maxRequestedPlayerIndex(): number {
    if (!this.game) {
      return 0;
    }
    return this.game.NumPlayers - 1;
  }

  get moveAsPlayer(): number {
    if (this.makeMovesAsViewingAsPlayer) return this.viewingAsPlayer;
    return this.ADMIN_PLAYER_INDEX;
  }

  get autoCurrentPlayer(): boolean {
    if (!this.active) return false;
    return this.viewAs === 'current';
  }

  private get _gameStateBlob(): string {
    return JSON.stringify(this.currentState, null, 2);
  }

  private get _chestAsString(): string {
    return JSON.stringify(this.chest, null, 2);
  }

  private _handleProposeMove(e: CustomEvent): void {
    const { name, arguments: moveArguments } = e.detail;
    if (!this.movesElement) {
      console.warn("propose-move fired, but no moves element to forward to.");
      return;
    }
    this.movesElement.proposeMove(name, moveArguments);
    // Stop propagation since we've handled it
    e.stopPropagation();
  }

  private _handleViewAsChange(e: Event): void {
    const radio = e.target as MdRadio;
    if (!radio.checked) return; // Ignore deselection events
    this.viewAs = radio.value as 'custom' | 'admin' | 'current' | 'observer';
  }

  private _handleCustomPlayerInput(e: Event): void {
    const input = e.target as HTMLInputElement;
    this.customRequestedPlayer = input.valueAsNumber || 0;
  }

  private _handleMakeMovesCheckboxChange(e: Event): void {
    const checkbox = e.target as MdCheckbox;
    this.makeMovesAsViewingAsPlayer = checkbox.checked;
  }

  protected updated(changedProperties: Map<string, unknown>): void {
    // Notify parent of property changes that need to be observed
    if (changedProperties.has('viewAs') ||
        changedProperties.has('customRequestedPlayer') ||
        changedProperties.has('game')) {
      this.dispatchEvent(new CustomEvent('requested-player-changed', {
        detail: { value: this.requestedPlayer },
        bubbles: true,
        composed: true
      }));
    }

    if (changedProperties.has('viewAs')) {
      this.dispatchEvent(new CustomEvent('auto-current-player-changed', {
        detail: { value: this.autoCurrentPlayer },
        bubbles: true,
        composed: true
      }));
    }
  }

  render() {
    return html`
      <div ?hidden="${!this.active}">
        <div class="card horizontal layout admin center">
          <div class="flex">
            <span class="view-as-label">View as</span>
            <div role="radiogroup" aria-label="View as" @change="${this._handleViewAsChange}">
              <label class="radio-label">
                <md-radio name="viewAs" value="admin" ?checked="${this.viewAs === 'admin'}"></md-radio>
                Admin
              </label>
              <label class="radio-label">
                <md-radio name="viewAs" value="observer" ?checked="${this.viewAs === 'observer'}"></md-radio>
                Observer
              </label>
              <label class="radio-label">
                <md-radio name="viewAs" value="current" ?checked="${this.viewAs === 'current'}"></md-radio>
                Current Player
              </label>
              <label class="radio-label">
                <md-radio name="viewAs" value="custom" ?checked="${this.viewAs === 'custom'}"></md-radio>
                Custom
              </label>
            </div>
            <input
              type="number"
              .value="${this.customRequestedPlayer}"
              @input="${this._handleCustomPlayerInput}"
              min="0"
              max="${this.maxRequestedPlayerIndex}">
          </div>
          <div>
            <md-checkbox
              id="move-as-player"
              ?checked="${this.makeMovesAsViewingAsPlayer}"
              @change="${this._handleMakeMovesCheckboxChange}">
              Make Moves As ViewingAsPlayer
            </md-checkbox>
          </div>
        </div>
        ${when(!this.game?.Finished, () => html`
          <div class="card">
            <boardgame-move-form
              ?admin="${this.active}"
              .moveAsPlayer="${this.moveAsPlayer}"
              id="moves"
              .config="${this.moveForms}"
              .gameRoute="${this.gameRoute}">
            </boardgame-move-form>
          </div>
        `)}
        <div class="card">
          <details>
            <summary>State</summary>
            <pre>${this._gameStateBlob}</pre>
          </details>
        </div>
        <div class="card">
          <details>
            <summary>Chest</summary>
            <pre>${this._chestAsString}</pre>
          </details>
        </div>
      </div>
    `;
  }
}

declare global {
  interface HTMLElementTagNameMap {
    'boardgame-admin-controls': BoardgameAdminControls;
  }
}
