/**
@license
Copyright (c) 2016 The Polymer Project Authors. All rights reserved.
This code may only be used under the BSD style license found at http://polymer.github.io/LICENSE.txt
The complete set of authors may be found at http://polymer.github.io/AUTHORS.txt
The complete set of contributors may be found at http://polymer.github.io/CONTRIBUTORS.txt
Code distributed by Google as part of the polymer project is also
subject to an additional IP rights grant found at http://polymer.github.io/PATENTS.txt
*/
import { PolymerElement } from '@polymer/polymer/polymer-element.js';

import '@polymer/polymer/lib/elements/dom-repeat.js';
import '@polymer/paper-styles/typography.js';
import '@polymer/paper-dropdown-menu/paper-dropdown-menu.js';
import '@polymer/paper-listbox/paper-listbox.js';
import './shared-styles.js';
import './boardgame-create-game.js';
import './boardgame-game-item.js';
import { html } from '@polymer/polymer/lib/utils/html-tag.js';

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
} from '../selectors.js';


import {
  fetchManagers,
  updateGameTypeFilter,
  fetchGamesList
} from '../actions/list.js';

class BoardgameListGamesView extends connect(store)(PolymerElement) {
  static get template() {
    return html`
    <style include="shared-styles">
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
      <boardgame-create-game logged-in="[[loggedIn]]" managers="[[_managers]]"></boardgame-create-game>
    </div>
    <div class="card">
      <paper-dropdown-menu name="manager" label="Game Type Filter">
        <paper-listbox slot="dropdown-content" selected="[[_gameTypeFilter]]" on-selected-changed="_handleSelectedChanged">
          <paper-item value="">All Games</paper-item>
          <template is="dom-repeat" items="[[_managers]]">
            <paper-item value="[[item.Name]]" data="[[item]]">[[item.DisplayName]]</paper-item> 
          </template>
        </paper-listbox>
      </paper-dropdown-menu>
    </div>
    <template is="dom-if" if="[[_participatingActiveGames.length]]">
      <h2>Active</h2>
      <template is="dom-repeat" items="[[_participatingActiveGames]]">
        <boardgame-game-item item="[[item]]" managers="[[_managers]]"></boardgame-game-item>
      </template>
    </template>
    <template is="dom-if" if="[[_participatingFinishedGames.length]]">
      <h2>Finished</h2>
      <template is="dom-repeat" items="[[_participatingFinishedGames]]">
        <boardgame-game-item item="[[item]]" managers="[[_managers]]"></boardgame-game-item>
      </template>
    </template>
    <template is="dom-if" if="[[_visibleJoinableActiveGames.length]]">
      <h2>Joinable</h2>
      <template is="dom-repeat" items="[[_visibleJoinableActiveGames]]">
        <boardgame-game-item item="[[item]]" managers="[[_managers]]"></boardgame-game-item>
      </template>
    </template>
    <template is="dom-if" if="[[_visibleActiveGames.length]]">
      <h2>Spectator</h2>
      <template is="dom-repeat" items="[[_visibleActiveGames]]">
        <boardgame-game-item item="[[item]]" managers="[[_managers]]"></boardgame-game-item>
      </template>
    </template>
    <template is="dom-if" if="[[_allGames.length]]">
      <h2>All Games</h2>
      <template is="dom-repeat" items="[[_allGames]]">
        <boardgame-game-item item="[[item]]" managers="[[_managers]]"></boardgame-game-item>
      </template>
    </template>
`;
  }

  static get is() {
    return "boardgame-list-games-view"
  }

  static get properties() {
    return {
      _participatingActiveGames: { type: Array },
      _participatingFinishedGames: { type: Array },
      _visibleActiveGames: { type: Array },
      _visibleJoinableGames: { type: Array },
      _allGames: { type: Array },
      _managers: {
        type: Array,
      },
      _gameTypeFilter: { 
        type: String,
        observer: "_gameTypeChanged",
      },
      admin: {
        type: Boolean,
        value: false,
        observer: "_adminChanged",
      },
      selected: {
        type: Boolean,
        observer: '_selectedChanged',
      },
      loggedIn: {
        type: Boolean,
        observer: "_loggedInChanged",
      }
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
  }

  _handleSelectedChanged(e) {
    store.dispatch(updateGameTypeFilter(e.path[0].selectedItem.value))
  }

  _adminChanged() {
    this._fetchGames();
  }

  _gameTypeChanged() {
    this._fetchGames();
  }

  _loggedInChanged(newValue) {
    //TODO: this is a race. Ideally loggedIn wouldn't change until the
    //user was logged out as far as server was concerned.
    setTimeout(() => this._fetchGames(), 250);
  }

  _fetchGames() {
    store.dispatch(fetchGamesList(this._gameTypeFilter, this.admin));
  }

  _selectedChanged(newValue) {
    if (newValue) {
      store.dispatch(fetchManagers());
      this._fetchGames();
    }
  }
}

customElements.define(BoardgameListGamesView.is, BoardgameListGamesView);
