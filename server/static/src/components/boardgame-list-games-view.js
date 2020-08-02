
import { LitElement, html } from '@polymer/lit-element';
import { repeat } from 'lit-html/directives/repeat';

import '@polymer/polymer/lib/elements/dom-repeat.js';
import '@polymer/paper-styles/typography.js';
import '@polymer/paper-dropdown-menu/paper-dropdown-menu.js';
import '@polymer/paper-listbox/paper-listbox.js';
import './boardgame-create-game.js';
import './boardgame-game-item.js';

import { SharedStyles } from './shared-styles-lit.js';

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

class BoardgameListGamesView extends connect(store)(LitElement) {
  render() {
    return html`
    ${SharedStyles}
    <style>
      :host {
        display: block;

        padding: 10px;
      }
      h2 {
        margin: 0 24px;
        @apply --paper-font-title;
      }
    </style>
    <div class="card">
      <boardgame-create-game .loggedIn=${this._loggedIn} .managers=${this._managers}></boardgame-create-game>
    </div>
    <div class="card">
      <paper-dropdown-menu name="manager" label="Game Type Filter">
        <paper-listbox slot="dropdown-content" .selected=${this._gameTypeFilter} @selected-changed=${this._handleSelectedChanged}>
          <paper-item value="">All Games</paper-item>
          ${repeat(this._managers, (i) => html`<paper-item .value=${i.Name} .data=${i}>${i.DisplayName}</paper-item>`)}
        </paper-listbox>
      </paper-dropdown-menu>
    </div>
    ${
      this._participatingActiveGames.length ? 
      html`
        <h2>Active</h2>
        ${repeat(this._participatingActiveGames, (i) => html`<boardgame-game-item .item=${i} .managers=${this._managers}></boardgame-game-item>`)}
      ` :
      html``
    }
    ${
      this._participatingFinishedGames.length ? 
      html`
        <h2>Finished</h2>
        ${repeat(this._participatingFinishedGames, (i) => html`<boardgame-game-item .item=${i} .managers=${this._managers}></boardgame-game-item>`)}
      ` :
      html``
    }
    ${
      this._visibleJoinableGames.length ? 
      html`
        <h2>Joinable</h2>
        ${repeat(this._visibleJoinableGames, (i) => html`<boardgame-game-item .item=${i} .managers=${this._managers}></boardgame-game-item>`)}
      ` :
      html``
    }
    ${
      this._visibleActiveGames.length ? 
      html`
        <h2>Spectator</h2>
        ${repeat(this._visibleActiveGames, (i) => html`<boardgame-game-item .item=${i} .managers=${this._managers}></boardgame-game-item>`)}
      ` :
      html``
    }
    ${
      this._allGames.length ? 
      html`
        <h2>All Games</h2>
        ${repeat(this._allGames, (i) => html`<boardgame-game-item .item=${i} .managers=${this._managers}></boardgame-game-item>`)}
      ` :
      html``
    }
`;
  }

  static get properties() {
    return {
      _participatingActiveGames: { type: Array },
      _participatingFinishedGames: { type: Array },
      _visibleActiveGames: { type: Array },
      _visibleJoinableGames: { type: Array },
      _allGames: { type: Array },
      _managers: { type: Array },
      _gameTypeFilter: { type: String },
      _loggedIn: { type: Boolean},
      _admin: { type: Boolean },
      selected: { type: Boolean },
    }
  }

  stateChanged(state) {
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

  updated(changedProps) {
    if (changedProps.has('selected') && this.selected) {
      store.dispatch(fetchManagers());
      this._fetchGamesList();
      return
    }
    if (changedProps.has('_loggedIn')) {
      //TODO: this is a race. Ideally loggedIn wouldn't change until the
      //user was logged out as far as server was concerned.
      setTimeout(() =>  this._fetchGamesList(), 250);
    }
    if (changedProps.has('_admin') || changedProps.has('_gameTypeFilter')) {
      this._fetchGamesList();
    }
  }

  _fetchGamesList() {
    store.dispatch(fetchGamesList(this._gameTypeFilter, this._admin));
  }

  _handleSelectedChanged(e) {
    const item = e.path[0].selectedItem;
    if (!item) return;
    store.dispatch(updateGameTypeFilter(item.value))
  }
}

customElements.define('boardgame-list-games-view', BoardgameListGamesView);
