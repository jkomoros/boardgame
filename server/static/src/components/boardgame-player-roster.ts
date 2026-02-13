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

import { connect } from 'pwa-helpers/connect-mixin.js';
import { store } from '../store.js';
import { joinGame } from '../actions/game.js';
import { selectGameError } from '../selectors.js';
import type { RootState } from '../types/store';

import type { MdDialog } from '@material/web/dialog/dialog.js';

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

@customElement('boardgame-player-roster')
export class BoardgamePlayerRoster extends connect(store)(LitElement) {
  static styles = css`
    :host {
      display: block;
    }

    h3 {
      margin: 0;
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
      background: white;
      padding: 16px;
      margin: 8px 0;
      border-radius: 4px;
      box-shadow: 0 2px 2px 0 rgba(0, 0, 0, 0.14),
                  0 1px 5px 0 rgba(0, 0, 0, 0.12),
                  0 3px 1px -2px rgba(0, 0, 0, 0.2);
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

  @query('#join')
  private joinDialog!: MdDialog;

  private readonly OBSERVER_PLAYER_INDEX = -1;
  private readonly ADMIN_PLAYER_INDEX = -2;

  private _lastError: string | null = null;

  stateChanged(state: RootState): void {
    const error = selectGameError(state);
    // Show error if it changed and is new
    if (error && error !== this._lastError) {
      this._lastError = error;
      this.dispatchEvent(new CustomEvent("show-error", {
        composed: true,
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

  private _isWinner(index: number, winners: number[]): boolean {
    if (!winners) return false;
    for (let i = 0; i < winners.length; i++) {
      if (winners[i] === index) {
        return true;
      }
    }
    return false;
  }

  private _bannerText(finished: boolean, winners: number[]): string {
    if (!finished) {
      return "Playing";
    }
    return "Game Over";
  }

  private playerName(viewingAsPlayer: number): string {
    if (viewingAsPlayer === this.ADMIN_PLAYER_INDEX) return "Admin";
    return "player " + viewingAsPlayer;
  }

  private showDialog(): void {
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
        detail: { nextAction: this.doJoin.bind(this) }
      }));
      return;
    }

    if (!this.gameRoute) return;

    // Dispatch action - errors will be handled via Redux state in stateChanged()
    store.dispatch(joinGame(this.gameRoute));

    // Tell game-view to fetch data now
    this.dispatchEvent(new CustomEvent("refresh-info", { composed: true }));
  }

  private async _gameRouteChanged(newValue: GameRoute | null): Promise<void> {
    if (!newValue) return;
    this.rendererLoaded = false;

    try {
      // Use /* @vite-ignore */ to allow fully dynamic imports in dev mode
      await import(/* @vite-ignore */ `../../game-src/${newValue.name}/boardgame-render-player-info-${newValue.name}.ts`);
      this._rendererLoaded();
    } catch (error) {
      console.error(`Failed to load player info renderer for ${newValue.name}:`, error);
    }
  }

  private _rendererLoaded(): void {
    this.rendererLoaded = true;
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
      <div class="layout horizontal justified players">
        ${repeat(this.playersInfo, (_, index) => index, (item, index) => html`
          <boardgame-player-roster-item
            class="flex"
            .state="${this.state}"
            .gameName="${this.gameRoute?.name}"
            ?is-empty="${item.IsEmpty}"
            ?finished="${this.finished}"
            ?winner="${this._isWinner(index, this.winners)}"
            ?is-agent="${item.IsAgent}"
            .photoUrl="${item.PhotoUrl}"
            .displayName="${item.DisplayName}"
            .playerIndex="${index}"
            .viewingAsPlayer="${this.viewingAsPlayer}"
            .currentPlayerIndex="${this.currentPlayerIndex}"
            ?renderer-loaded="${this.rendererLoaded}"
            ?active="${this.active}">
          </boardgame-player-roster-item>
        `)}
      </div>
      ${when(this.isObserver, () => html`
        <div>
          <div class="layout horizontal center">
            <h3 class="flex">Observing</h3>
            ${when(this.showJoin, () => html`
              <div>
                <md-filled-button @click="${this.showDialog}" raised>
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
