import { LitElement, html, css } from 'lit';
import { customElement, property } from 'lit/decorators.js';
import { repeat } from 'lit/directives/repeat.js';
import { when } from 'lit/directives/when.js';
import '@material/web/select/filled-select.js';
import '@material/web/select/select-option.js';
import './boardgame-create-game.ts';
import './boardgame-game-item.ts';

import { connect } from 'pwa-helpers/connect-mixin.js';
import { store } from '../store.js';

import list from '../reducers/list.js';
store.addReducers({
  list
});

import {
  selectManagers,
  selectGameTypeFilter,
  selectParticipatingActiveGames,
  selectParticipatingFinishedGames,
  selectVisibleActiveGames,
  selectVisibleJoinableGames,
  selectAllGames,
  selectLoggedIn,
  selectAdmin
} from '../selectors.js';

import {
  fetchManagers,
  updateGameTypeFilter,
  fetchGamesList
} from '../actions/list.js';

import type { MdFilledSelect } from '@material/web/select/filled-select.js';

interface GameItem {
  Name: string;
  ID: string;
}

interface Manager {
  Name: string;
  DisplayName: string;
}

@customElement('boardgame-list-games-view')
export class BoardgameListGamesView extends connect(store)(LitElement) {
  static styles = css`
    :host {
      display: block;
      padding: 16px;
      max-width: 1200px;
      margin: 0 auto;
    }

    h2 {
      margin: 24px 16px 12px 16px;
      font-family: var(--md-sys-typescale-headline-small-font);
      font-size: var(--md-sys-typescale-headline-small-size);
      line-height: var(--md-sys-typescale-headline-small-line-height);
      font-weight: var(--md-sys-typescale-headline-small-weight);
      color: var(--md-sys-color-on-background);
      letter-spacing: 0;
    }

    .card {
      background: var(--md-sys-color-surface-container-low);
      padding: 20px;
      margin: 12px;
      border-radius: 12px;
      box-shadow: var(--md-sys-elevation-1);
      transition: box-shadow 0.2s ease;
    }

    .card:hover {
      box-shadow: var(--md-sys-elevation-2);
    }

    md-filled-select {
      width: 100%;
    }
  `;

  @property({ type: Array })
  private _participatingActiveGames: GameItem[] = [];

  @property({ type: Array })
  private _participatingFinishedGames: GameItem[] = [];

  @property({ type: Array })
  private _visibleActiveGames: GameItem[] = [];

  @property({ type: Array })
  private _visibleJoinableGames: GameItem[] = [];

  @property({ type: Array })
  private _allGames: GameItem[] = [];

  @property({ type: Array })
  private _managers: Manager[] = [];

  @property({ type: String })
  private _gameTypeFilter = '';

  @property({ type: Boolean })
  private _loggedIn = false;

  @property({ type: Boolean })
  private _admin = false;

  @property({ type: Boolean })
  selected = false;

  stateChanged(state: any): void {
    this._managers = selectManagers(state);
    this._gameTypeFilter = selectGameTypeFilter(state);
    this._participatingActiveGames = selectParticipatingActiveGames(state);
    this._participatingFinishedGames = selectParticipatingFinishedGames(state);
    this._visibleActiveGames = selectVisibleActiveGames(state);
    this._visibleJoinableGames = selectVisibleJoinableGames(state);
    this._allGames = selectAllGames(state);
    this._loggedIn = selectLoggedIn(state);
    this._admin = selectAdmin(state);
  }

  protected updated(changedProps: Map<string, unknown>): void {
    if (changedProps.has('selected') && this.selected) {
      store.dispatch(fetchManagers());
      this._fetchGamesList();
      return;
    }
    if (changedProps.has('_loggedIn')) {
      // TODO: this is a race. Ideally loggedIn wouldn't change until the
      // user was logged out as far as server was concerned.
      setTimeout(() => this._fetchGamesList(), 250);
    }
    // TODO: the only reason we keep track of admin is to know to call fetchGamesList again...
    if (changedProps.has('_admin') || changedProps.has('_gameTypeFilter')) {
      this._fetchGamesList();
    }
  }

  private _fetchGamesList(): void {
    store.dispatch(fetchGamesList());
  }

  private _handleSelectedChanged(e: Event): void {
    const select = e.target as MdFilledSelect;
    store.dispatch(updateGameTypeFilter(select.value));
  }

  render() {
    return html`
      <div class="card">
        <boardgame-create-game></boardgame-create-game>
      </div>
      <div class="card">
        <md-filled-select
          name="manager"
          label="Game Type Filter"
          .value="${this._gameTypeFilter}"
          @change="${this._handleSelectedChanged}">
          <md-select-option value="">
            <div slot="headline">All Games</div>
          </md-select-option>
          ${repeat(
            this._managers,
            (manager) => manager.Name,
            (manager) => html`
              <md-select-option value="${manager.Name}">
                <div slot="headline">${manager.DisplayName}</div>
              </md-select-option>
            `
          )}
        </md-filled-select>
      </div>
      ${when(
        this._participatingActiveGames.length > 0,
        () => html`
          <h2>Active</h2>
          ${repeat(
            this._participatingActiveGames,
            (game) => game.ID,
            (game) => html`
              <boardgame-game-item .item="${game}" .managers="${this._managers}">
              </boardgame-game-item>
            `
          )}
        `
      )}
      ${when(
        this._participatingFinishedGames.length > 0,
        () => html`
          <h2>Finished</h2>
          ${repeat(
            this._participatingFinishedGames,
            (game) => game.ID,
            (game) => html`
              <boardgame-game-item .item="${game}" .managers="${this._managers}">
              </boardgame-game-item>
            `
          )}
        `
      )}
      ${when(
        this._visibleJoinableGames.length > 0,
        () => html`
          <h2>Joinable</h2>
          ${repeat(
            this._visibleJoinableGames,
            (game) => game.ID,
            (game) => html`
              <boardgame-game-item .item="${game}" .managers="${this._managers}">
              </boardgame-game-item>
            `
          )}
        `
      )}
      ${when(
        this._visibleActiveGames.length > 0,
        () => html`
          <h2>Spectator</h2>
          ${repeat(
            this._visibleActiveGames,
            (game) => game.ID,
            (game) => html`
              <boardgame-game-item .item="${game}" .managers="${this._managers}">
              </boardgame-game-item>
            `
          )}
        `
      )}
      ${when(
        this._allGames.length > 0,
        () => html`
          <h2>All Games</h2>
          ${repeat(
            this._allGames,
            (game) => game.ID,
            (game) => html`
              <boardgame-game-item .item="${game}" .managers="${this._managers}">
              </boardgame-game-item>
            `
          )}
        `
      )}
    `;
  }
}

declare global {
  interface HTMLElementTagNameMap {
    'boardgame-list-games-view': BoardgameListGamesView;
  }
}
