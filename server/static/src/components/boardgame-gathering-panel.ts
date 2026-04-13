/**
 * boardgame-gathering-panel
 *
 * A framework-provided component that renders gathering UI (waiting status,
 * share link, team/color/role pickers, start button) between the player roster
 * and the game renderer. It auto-hides when it has nothing to show.
 *
 * Detection is driven by move form legality and game state — no explicit
 * "lobby mode" flag. Each sub-component independently decides whether to
 * render based on its corresponding move's presence in moveForms.
 */
import { LitElement, html, css, nothing } from 'lit';
import { customElement, property } from 'lit/decorators.js';
import type { MoveForm } from '../types/api';

import './boardgame-gathering-status.js';
import './boardgame-gathering-share.js';
import './boardgame-gathering-start.js';

interface PlayerInfo {
  IsEmpty: boolean;
  IsAgent: boolean;
  PhotoUrl: string;
  DisplayName: string;
}

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

    :host([hidden]) {
      display: none !important;
    }

    .panel {
      background: var(--md-sys-color-secondary-container, #e8def8);
      color: var(--md-sys-color-on-secondary-container, #1d192b);
      padding: 16px;
      margin: 8px 0;
      border-radius: 12px;
      box-shadow: var(--md-sys-elevation-1, 0 1px 3px 1px rgba(0,0,0,.15), 0 1px 2px rgba(0,0,0,.3));
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
  `;

  @property({ type: Array })
  moveForms: MoveForm[] | null = null;

  @property({ type: Object })
  state: any = null;

  @property({ type: Number })
  viewingAsPlayer = 0;

  @property({ type: Boolean })
  hasEmptySlots = false;

  @property({ type: Boolean })
  gameOpen = false;

  @property({ type: Boolean })
  finished = false;

  @property({ type: Boolean })
  isOwner = false;

  @property({ type: Boolean })
  loggedIn = false;

  @property({ type: Object })
  gameRoute: GameRoute | null = null;

  @property({ type: Array })
  playersInfo: PlayerInfo[] = [];

  private get _readyToStartError(): string {
    return this.state?.Game?.Computed?.Global?.ReadyToStartError || '';
  }

  private get _isObserver(): boolean {
    return this.viewingAsPlayer === -1;
  }

  private get _showStatus(): boolean {
    return (this.hasEmptySlots && this.gameOpen && !this.finished) ||
           !!this._readyToStartError;
  }

  private get _showShare(): boolean {
    return this.hasEmptySlots && this.gameOpen && !this._isObserver;
  }

  private get _startMoveForm(): MoveForm | null {
    if (!this.moveForms) return null;
    const startNames = new Set([
      'Confirm Players', 'Close All Seats', 'Start Game', 'Finalize Set Up'
    ]);
    return this.moveForms.find(f => startNames.has(f.Name) && f.LegalForAnyone) ?? null;
  }

  private get _hasAnythingToShow(): boolean {
    return this._showStatus || this._showShare || !!this._startMoveForm;
  }

  render() {
    if (!this._hasAnythingToShow) {
      this.setAttribute('hidden', '');
      return nothing;
    }
    this.removeAttribute('hidden');

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

            <span style="flex:1"></span>

            ${this._showShare ? html`
              <boardgame-gathering-share
                .gameRoute=${this.gameRoute}>
              </boardgame-gathering-share>
            ` : nothing}

            ${this._startMoveForm ? html`
              <boardgame-gathering-start
                .moveForm=${this._startMoveForm}
                .viewingAsPlayer=${this.viewingAsPlayer}>
              </boardgame-gathering-start>
            ` : nothing}
          </div>
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
