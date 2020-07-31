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
import './boardgame-ajax.js';
import { html } from '@polymer/polymer/lib/utils/html-tag.js';

import { connect } from 'pwa-helpers/connect-mixin.js';
import { store } from './store.js';

import list from './reducers/list.js';
store.addReducers({
	list
});

import {
  selectManagers
} from './selectors.js';


import {
  fetchManagers
} from './actions/list.js';

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
      <boardgame-create-game logged-in="[[loggedIn]]" managers="[[managers]]"></boardgame-create-game>
    </div>
    <div class="card">
      <paper-dropdown-menu name="manager" label="Game Type Filter">
        <paper-listbox slot="dropdown-content" selected="0" selected-item="{{selectedManager}}">
          <paper-item value="">All Games</paper-item>
          <template is="dom-repeat" items="[[managers]]">
            <paper-item value="[[item.Name]]" data="[[item]]">[[item.DisplayName]]</paper-item> 
          </template>
        </paper-listbox>
      </paper-dropdown-menu>
    </div>
    <template is="dom-if" if="[[data.ParticipatingActiveGames.length]]">
      <h2>Active</h2>
      <template is="dom-repeat" items="[[data.ParticipatingActiveGames]]">
        <boardgame-game-item item="[[item]]" managers="[[managers]]"></boardgame-game-item>
      </template>
    </template>
    <template is="dom-if" if="[[data.PaticipatingFinishedGames.length]]">
      <h2>Finished</h2>
      <template is="dom-repeat" items="[[data.PaticipatingFinishedGames]]">
        <boardgame-game-item item="[[item]]" managers="[[managers]]"></boardgame-game-item>
      </template>
    </template>
    <template is="dom-if" if="[[data.VisibleJoinableActiveGames.length]]">
      <h2>Joinable</h2>
      <template is="dom-repeat" items="[[data.VisibleJoinableActiveGames]]">
        <boardgame-game-item item="[[item]]" managers="[[managers]]"></boardgame-game-item>
      </template>
    </template>
    <template is="dom-if" if="[[data.VisibleActiveGames.length]]">
      <h2>Spectator</h2>
      <template is="dom-repeat" items="[[data.VisibleActiveGames]]">
        <boardgame-game-item item="[[item]]" managers="[[managers]]"></boardgame-game-item>
      </template>
    </template>
    <template is="dom-if" if="[[data.AllGames.length]]">
      <h2>All Games</h2>
      <template is="dom-repeat" items="[[data.AllGames]]">
        <boardgame-game-item item="[[item]]" managers="[[managers]]"></boardgame-game-item>
      </template>
    </template>

    
    <boardgame-ajax auto="" debounce-duration="100" id="games" path="list/game" handle-as="json" params="[[gamesArgs]]" last-response="{{data}}"></boardgame-ajax>
`;
  }

  static get is() {
    return "boardgame-list-games-view"
  }

  static get properties() {
    return {
      data: Object,
      managers: {
        type: Array,
      },
      selectedManager: Object,
      admin: {
        type: Boolean,
        value: false,
      },
      gameType: {
        type: String,
        value: "",
        computed: "_computeGameType(selectedManager)",
      },
      gamesArgs: {
        type: Object,
        computed: "_computeGamesArgs(gameType, admin)"
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

  ready() {
    super.ready();
    store.dispatch(fetchManagers());
  }

  stateChanged(state) {
    this.managers = selectManagers(state);
  }

  _computeGamesArgs(gameType, admin) {
    return {"name": gameType, "admin" : (admin ? 1 : 0)}
  }

  _computeGameType(selectedManager) {
    if (!selectedManager) return "";
    return selectedManager.value || selectedManager.getAttribute("value") || "";
  }

  _loggedInChanged(newValue) {
    //TODO: this is a race. Ideally loggedIn wouldn't change until the
    //user was logged out as far as server was concerned.
    setTimeout(() => this.$.games.generateRequest(), 250);
  }

  _selectedChanged(newValue) {
    if (newValue) {
      if (this.$.games.loading) return;
      this.$.games.generateRequest();
    }
  }
}

customElements.define(BoardgameListGamesView.is, BoardgameListGamesView);
