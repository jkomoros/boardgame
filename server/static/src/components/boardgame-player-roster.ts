/**
@license
Copyright (c) 2016 The Polymer Project Authors. All rights reserved.
This code may only be used under the BSD style license found at http://polymer.github.io/LICENSE.txt
The complete set of authors may be found at http://polymer.github.io/AUTHORS.txt
The complete set of contributors may be found at http://polymer.github.io/CONTRIBUTORS.txt
Code distributed by Google as part of the polymer project is also
subject to an additional IP rights grant found at http://polymer.github.io/PATENTS.txt
*/
import { LitElement, html, css } from 'lit';
import { customElement, property, query } from 'lit/decorators.js';
import { repeat } from 'lit/directives/repeat.js';
import { when } from 'lit/directives/when.js';
import '@material/web/dialog/dialog.js';
import '@material/web/button/filled-button.js';
import '@material/web/button/outlined-button.js';
import './boardgame-configure-game-properties.js';
import './boardgame-player-roster-item.js';
import { BoardgameBasePlayerInfoRenderer } from './boardgame-base-player-info-renderer.js';

import { connect } from 'pwa-helpers/connect-mixin.js';
import { store } from '../store.js';
import { joinGame } from '../actions/game.js';
import { selectGameError } from '../selectors.js';
import type { PlayerInfo, RootState } from '../types/store';

import type { MdDialog } from '@material/web/dialog/dialog.js';
import { getReadyToStartError } from './gathering-shared.js';

interface GameRoute {
  name: string;
  id: string;
}

@customElement('boardgame-player-roster')
export class BoardgamePlayerRoster extends connect(store)(LitElement) {
  static styles = css`
    :host {
      display: block;
    }

    h3 {
      margin: 0;
      font-family: var(--md-sys-typescale-title-medium-font, 'Source Sans 3', sans-serif);
      font-size: var(--md-sys-typescale-title-medium-size, 16px);
      font-weight: var(--md-sys-typescale-title-medium-weight, 500);
      color: var(--md-sys-color-on-surface, #1C1810);
    }

    .layout {
      display: flex;
    }

    .horizontal {
      flex-direction: row;
    }

    .center {
      align-items: center;
    }

    .justified {
      justify-content: space-between;
    }

    .flex {
      flex: 1;
    }

    .players {
      gap: 8px;
    }

    .card {
      background: linear-gradient(180deg, var(--md-sys-color-surface-container-low, #F5F0E8) 0%, var(--md-sys-color-surface-container, #F0EBE3) 100%);
      padding: 16px;
      margin: 8px 0;
      border-radius: 12px;
      box-shadow: 0 1px 3px 0 rgba(60, 40, 20, 0.10),
                  0 1px 2px 0 rgba(60, 40, 20, 0.06),
                  inset 0 1px 0 rgba(255, 255, 255, 0.5);
      color: var(--md-sys-color-on-surface, #1C1810);
    }

    .renderer-error {
      margin: 8px 0;
      padding: 12px;
      border: 1px solid var(--md-sys-color-error, #BA1A1A);
      border-radius: 8px;
      color: var(--md-sys-color-on-error-container, #410002);
      background: var(--md-sys-color-error-container, #FFDAD6);
    }
  `;

  @property({ type: Number })
  viewingAsPlayer = 0;

  @property({ type: Boolean })
  hasEmptySlots = false;

  @property({ type: Boolean })
  gameOpen = false;

  @property({ type: Boolean })
  gameVisible = false;

  @property({ type: Object })
  gameRoute: GameRoute | null = null;

  @property({ type: Boolean })
  active = false;

  @property({ type: Boolean })
  admin = false;

  @property({ type: Boolean })
  isOwner = false;

  @property({ type: Array })
  playersInfo: PlayerInfo[] = [];

  @property({ type: Number })
  currentPlayerIndex = 0;

  @property({ type: Object })
  state: unknown = null;

  @property({ type: Boolean })
  finished = false;

  @property({ type: Array })
  winners: number[] = [];

  @property({ type: Boolean })
  loggedIn = false;

  @property({ type: Boolean })
  rendererLoaded = false;

  @property({ type: String, attribute: false })
  rendererError = '';

  private _rendererLoadGeneration = 0;

  // Framework-computed CSS colors per player (from selectPlayerColors).
  @property({ type: Array })
  playerColors: string[] = [];

  // Framework-computed activity per player (from selectPlayerActivity).
  @property({ type: Array })
  playerActivity: boolean[] = [];

  // Custom player display order (from selectPlayerOrder), or null for default.
  @property({ type: Array })
  playerOrder: number[] | null = null;

  @query('#join')
  private joinDialog!: MdDialog;

  private readonly OBSERVER_PLAYER_INDEX = -1;
  private readonly ADMIN_PLAYER_INDEX = -2;

  // Memoized ordered indices, recomputed only when inputs change.
  @property({ type: Array, attribute: false })
  private _orderedIndices: number[] = [];

  protected willUpdate(changedProperties: Map<string, unknown>): void {
    if (changedProperties.has('playersInfo') || changedProperties.has('playerOrder')) {
      this._orderedIndices = this._computeOrderedIndices();
    }
  }

  // Validates that playerOrder contains in-range, unique indices; falls back to
  // default sequential order on any invalid input.
  private _computeOrderedIndices(): number[] {
    const n = this.playersInfo.length;
    if (this.playerOrder && this.playerOrder.length === n) {
      const seen = new Set<number>();
      let valid = true;
      for (const idx of this.playerOrder) {
        if (idx < 0 || idx >= n || seen.has(idx)) {
          valid = false;
          break;
        }
        seen.add(idx);
      }
      if (valid) {
        return this.playerOrder;
      }
      console.warn('boardgame-player-roster: invalid playerOrder (out-of-range or duplicate indices), falling back to default order', this.playerOrder);
    }
    return Array.from({ length: n }, (_, i) => i);
  }

  private _lastError: string | null = null;

  stateChanged(state: RootState): void {
    const error = selectGameError(state);
    // Show error if it changed and is new
    if (error && error !== this._lastError) {
      this._lastError = error;
      this.dispatchEvent(new CustomEvent("show-error", {
        composed: true,
        bubbles: true,
        detail: {
          message: error,
          friendlyMessage: error,
          title: "Couldn't Join"
        }
      }));
    } else if (!error) {
      this._lastError = null;
    }
  }

  get isObserver(): boolean {
    return this.viewingAsPlayer === this.OBSERVER_PLAYER_INDEX;
  }

  get showJoin(): boolean {
    return this.viewingAsPlayer === this.OBSERVER_PLAYER_INDEX &&
           this.hasEmptySlots &&
           this.gameOpen;
  }

  protected firstUpdated(): void {
    this.joinDialog.addEventListener('close', () => this._dialogClosed());
  }

  protected updated(changedProperties: Map<string, unknown>): void {
    if (changedProperties.has('gameRoute')) {
      this._gameRouteChanged(this.gameRoute);
    }
  }

  override disconnectedCallback(): void {
    this._rendererLoadGeneration++;
    super.disconnectedCallback();
  }

  override connectedCallback(): void {
    super.connectedCallback();
    if (this.hasUpdated && this.gameRoute && !this.rendererLoaded) {
      void this._gameRouteChanged(this.gameRoute);
    }
  }

  private _isWinner(index: number, winners: number[]): boolean {
    if (!winners) return false;
    for (let i = 0; i < winners.length; i++) {
      if (winners[i] === index) {
        return true;
      }
    }
    return false;
  }

  private get _readyToStartError(): string {
    return getReadyToStartError(this.state);
  }

  private _bannerText(finished: boolean, winners: number[]): string {
    if (finished) {
      return "Game Over";
    }
    // Only show "Setting Up" when ReadyToStart reports a configuration error
    // (e.g., teams not balanced). Don't change the banner based on
    // hasEmptySlots — that would affect drop-in/drop-out games that keep
    // slots open during active play. The gathering panel's status component
    // handles "Waiting for N more players" messaging separately.
    if (this._readyToStartError) {
      return "Setting Up";
    }
    return "Playing";
  }

  private playerName(viewingAsPlayer: number): string {
    if (viewingAsPlayer === this.ADMIN_PLAYER_INDEX) return "Admin";
    return "player " + viewingAsPlayer;
  }

  /** Open the join flow when this observer is currently eligible to join. */
  openJoinDialog(): void {
    if (this.joinDialog.open) return;
    if (this.viewingAsPlayer !== this.OBSERVER_PLAYER_INDEX) return;
    this.joinDialog.show();
  }

  private _dialogClosed(): void {
    // Check returnValue instead of e.detail.confirmed
    if (this.joinDialog.returnValue !== 'confirm') return;
    this.doJoin();
  }

  private doJoin(): void {
    if (!this.loggedIn) {
      this.dispatchEvent(new CustomEvent('show-login', {
        composed: true,
        bubbles: true,
        detail: { nextAction: this.doJoin.bind(this) }
      }));
      return;
    }

    if (!this.gameRoute) return;

    // Dispatch action - errors will be handled via Redux state in stateChanged()
    store.dispatch(joinGame(this.gameRoute));

    // Tell game-view to fetch data now
    this.dispatchEvent(new CustomEvent("refresh-info", { composed: true, bubbles: true }));
  }

  private async _gameRouteChanged(newValue: GameRoute | null, retry = false): Promise<void> {
    const generation = ++this._rendererLoadGeneration;
    this.rendererLoaded = false;
    this.rendererError = '';
    if (!newValue) return;
    if (!/^[a-z][a-z0-9]*$/.test(newValue.name)) {
      this.rendererError = `Invalid player renderer game name ${JSON.stringify(newValue.name)}; expected lowercase letters and digits`;
      console.error(this.rendererError);
      return;
    }

    const tagName = `boardgame-render-player-info-${newValue.name}`;
    const existing = customElements.get(tagName);
    if (existing) {
      try {
        this._validateRenderer(existing, tagName);
        this._rendererLoaded();
        return;
      } catch (error) {
        this.rendererError = this._errorMessage(error);
        return;
      }
    }

    try {
      // Use /* @vite-ignore */ to allow fully dynamic imports in dev mode
      const baseModulePath = `../../game-src/${newValue.name}/boardgame-render-player-info-${newValue.name}.ts`;
      const modulePath = retry ? `${baseModulePath}?retry=${generation}` : baseModulePath;
      await import(/* @vite-ignore */ modulePath);
      if (!this._rendererLoadIsCurrent(generation, newValue)) return;
      const constructor = customElements.get(tagName);
      if (!constructor) {
        throw new Error(
          `Player renderer module loaded but did not register <${tagName}>; ` +
          'use @registerPlayerInfoRenderer',
        );
      }
      this._validateRenderer(constructor, tagName);
      this._rendererLoaded();
    } catch (error) {
      if (!this._rendererLoadIsCurrent(generation, newValue)) return;
      this.rendererError = `Failed to load player renderer for ${newValue.name}: ${this._errorMessage(error)}`;
      console.error(`Failed to load player info renderer for ${newValue.name}:`, error);
    }
  }

  private _validateRenderer(constructor: CustomElementConstructor, tagName: string): void {
    if (!(constructor.prototype instanceof BoardgameBasePlayerInfoRenderer)) {
      throw new Error(
        `Player renderer <${tagName}> must extend the generated PlayerInfoRenderer base`,
      );
    }
  }

  private _rendererLoadIsCurrent(generation: number, route: GameRoute): boolean {
    return generation === this._rendererLoadGeneration
      && this.gameRoute?.name === route.name
      && this.gameRoute.id === route.id
      && this.isConnected;
  }

  private _errorMessage(error: unknown): string {
    const message = error instanceof Error ? error.message : String(error);
    const normalized = message.trim() || 'unknown renderer error';
    return normalized.length <= 500 ? normalized : `${normalized.slice(0, 497)}...`;
  }

  private _rendererLoaded(): void {
    this.rendererLoaded = true;
  }

  private retryRenderer(): void {
    if (this.gameRoute) void this._gameRouteChanged(this.gameRoute, true);
  }

  render() {
    return html`
      <div class="layout horizontal center">
        <h3 class="flex">${this._bannerText(this.finished, this.winners)}</h3>
        <boardgame-configure-game-properties
          ?game-visible="${this.gameVisible}"
          ?game-open="${this.gameOpen}"
          ?admin="${this.admin}"
          ?is-owner="${this.isOwner}"
          .gameRoute="${this.gameRoute}"
          configurable>
        </boardgame-configure-game-properties>
      </div>
      ${this.rendererError ? html`
        <section class="renderer-error" role="alert" aria-live="assertive">
          ${this.rendererError}. Run <code>boardgame-util check-client</code> and fix every diagnostic.
          <md-outlined-button @click=${this.retryRenderer}>Retry renderer</md-outlined-button>
        </section>
      ` : null}
      <div class="layout horizontal justified players">
        ${repeat(this._orderedIndices, (idx) => idx, (idx) => {
          const item = this.playersInfo[idx];
          if (!item) return html``;
          return html`
          <boardgame-player-roster-item
            class="flex"
            .state="${this.state}"
            .gameName="${this.gameRoute?.name}"
            ?is-empty="${item.IsEmpty}"
            ?finished="${this.finished}"
            ?winner="${this._isWinner(idx, this.winners)}"
            ?is-agent="${item.IsAgent}"
            .photoUrl="${item.PhotoURL || ''}"
            .displayName="${item.DisplayName}"
            .playerIndex="${idx}"
            .viewingAsPlayer="${this.viewingAsPlayer}"
            .currentPlayerIndex="${this.currentPlayerIndex}"
            .computedColor="${this.playerColors[idx] || ''}"
            .mayBeActive="${this.playerActivity[idx] !== false}"
            ?renderer-loaded="${this.rendererLoaded}"
            ?active="${this.active}">
          </boardgame-player-roster-item>
        `})}
      </div>
      ${when(this.isObserver, () => html`
        <div>
          <div class="layout horizontal center">
            <h3 class="flex">Observing</h3>
            ${when(this.showJoin, () => html`
              <div>
                <md-filled-button @click="${this.openJoinDialog}" raised>
                  Join game
                </md-filled-button>
              </div>
            `)}
          </div>
        </div>
      `)}
      <md-dialog id="join">
        <div slot="headline">Join game?</div>
        <form id="join-form" slot="content" method="dialog">
          <p>We're still looking for players for this game.</p>
        </form>
        <div slot="actions">
          <md-outlined-button value="dismiss" form="join-form">
            I'll just watch
          </md-outlined-button>
          <md-filled-button value="confirm" form="join-form" autofocus>
            I'm in!
          </md-filled-button>
        </div>
      </md-dialog>
    `;
  }
}

declare global {
  interface HTMLElementTagNameMap {
    'boardgame-player-roster': BoardgamePlayerRoster;
  }
}
