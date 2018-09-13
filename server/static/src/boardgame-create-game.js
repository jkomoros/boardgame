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
import '@polymer/paper-button/paper-button.js';
import '@polymer/iron-flex-layout/iron-flex-layout-classes.js';
import '@polymer/paper-dropdown-menu/paper-dropdown-menu.js';
import '@polymer/paper-listbox/paper-listbox.js';
import '@polymer/paper-toggle-button/paper-toggle-button.js';
import '@polymer/paper-item/paper-item.js';
import '@polymer/paper-item/paper-item-body.js';
import '@polymer/paper-slider/paper-slider.js';
import '@polymer/paper-radio-button/paper-radio-button.js';
import '@polymer/paper-radio-group/paper-radio-group.js';
import '@polymer/iron-icons/iron-icons.js';
import '@polymer/iron-icons/social-icons.js';
import '@polymer/iron-icon/iron-icon.js';
import '@polymer/paper-styles/typography.js';
import '@polymer/paper-styles/default-theme.js';
import './boardgame-ajax.js';
import './shared-styles.js';
import {GamePathMixin} from './boardgame-game-path.js';
import { html } from '@polymer/polymer/lib/utils/html-tag.js';

class BoardgameCreateGame extends GamePathMixin(PolymerElement) {
  static get template() {
    return html`
    <style is="custom-style" include="iron-flex iron-flex-alignment shared-styles">
      paper-toggle-button {
        margin-right: 1em;
      }

      .secondary {
        @apply(--paper-font-caption);
        color: var(--light-theme-secondary-color);
      }

      .game .secondary {
        margin-bottom: -1em;
      }

      .variant>div {
        margin-right: 1em;
      }

      [hidden] {
        display:none !important;
      }

      .game paper-item-body [secondary] {
        max-width: 40em;
        white-space: normal;
      }
    </style>
    <div class="vertical layout">
      <div class="horizontal layout center game">
        <paper-dropdown-menu name="manager" label="Game Type" horizontal-align="left">
          <paper-listbox slot="dropdown-content" selected="0" selected-item="{{selectedManager}}">
            <template is="dom-repeat" items="{{managers}}">
              <paper-item value="{{item.Name}}" data="{{item}}" label="{{item.DisplayName}}">
                <paper-item-body two-line="">
                  <div>{{item.DisplayName}}</div>
                  <div secondary="">{{item.Description}}</div>
                </paper-item-body>
              </paper-item> 
            </template>
          </paper-listbox>
        </paper-dropdown-menu>
        <div class="vertical layout">

          <div hidden\$="{{managerFixedPlayerCount}}">
            <div class="secondary">Number of Players</div>
            <paper-slider name="numplayers" label="Number of Players" min="{{selectedManager.data.MinNumPlayers}}" max="{{selectedManager.data.MaxNumPlayers}}" value="{{numPlayers}}" snaps="" pin="" editable="" max-markers="100"></paper-slider>
          </div>
          <div hidden\$="{{!managerFixedPlayerCount}}">
            <div class="secondary"><strong>{{selectedManager.data.MinNumPlayers}}</strong> players</div>
          </div>
        </div>
        <div class="flex"></div>
        <paper-button on-tap="createGame" default="" raised="">Create Game</paper-button>
      </div>

      <div class="horizontal layout justified" hidden\$="{{managerHasAgents}}">
        <template is="dom-repeat" items="{{players}}">
          <div class="flex">
            <div class="vertical layout">
              Player {{index}}
              <paper-radio-group selected="" disabled="{{managerHasAgents}}" name="agent-player-{{index}}" attr-for-selected="value">
                <paper-radio-button name="agent-player-{{index}}" value="" disabled="{{managerHasAgents}}">Real Live Human</paper-radio-button>
                <template is="dom-repeat" items="{{selectedManager.data.Agents}}" index-as="agentIndex">
                  <paper-radio-button name="agent-player-{{index}}" value="{{item.Name}}">{{item.DisplayName}}</paper-radio-button>
                </template>
              </paper-radio-group>
            </div>
          </div>
        </template>
      </div>
      <div class="horizontal layout variant">
        <template is="dom-repeat" items="{{selectedManager.data.Variant}}">
          <div class="vertical layout">
            <paper-dropdown-menu label="{{item.DisplayName}}" name="variant_{{item.Name}}" horizontal-align="left">
              <paper-listbox slot="dropdown-content" selected="0">
                <template is="dom-repeat" items="{{item.Values}}">
                  <paper-item value="{{item.Value}}" label="{{item.DisplayName}}">
                    <paper-item-body two-line="">
                      <div>{{item.DisplayName}}</div>
                      <div secondary="">{{item.Description}}</div>
                    </paper-item-body>
                  </paper-item>
                </template>
              </paper-listbox>
            </paper-dropdown-menu>
            <div class="secondary">{{item.Description}}</div>
          </div>
        </template>
      </div>
      <div class="horizontal layout">
        <paper-toggle-button name="visible" checked=""><iron-icon icon="visibility"></iron-icon> Allow strangers to find the game</paper-toggle-button>
        <paper-toggle-button name="open" checked=""><iron-icon icon="social:people"></iron-icon> Allow anyone who can view the game to join</paper-toggle-button>
      </div>
    </div>
    <boardgame-ajax id="create" path="new/game" method="POST" content-type="application/x-www-form-urlencoded" last-response="{{createGameResponse}}"></boardgame-ajax>
`;
  }

  static get is() {
    return "boardgame-create-game";
  }

  static get properties() {
    return {
      loggedIn: Boolean,
      selectedManager: {
        type: Object,
        observer: "_selectedManagerChanged",
      },
      managers: Array,
      managerHasAgents: {
        type: Boolean,
        computed: "_computeManagerHasAgents(selectedManager)"
      },
      managerFixedPlayerCount: {
        type: Boolean,
        computed: "_computeManagerHasFixedPlayerCount(selectedManager)"
      },
      numPlayers: {
        type: Number,
        value: 0,
      },
      players: {
        type: Object,
        computed: "_computePlayers(selectedManager, numPlayers)",
      },
      createGameResponse: {
        type: Object,
        observer: "_createGameResponseChanged"
      }
    }
  }

  _computePlayers(selectedManager, numPlayers) {
    if (!selectedManager) return [];
    var data = selectedManager.data;
    if (numPlayers == 0) {
      numPlayers = data.DefaultNumPlayers;
    }
    var result = [];
    for (var i = 0; i < numPlayers; i++) {
      result.push("");
    }
    return result;
  }



  _selectedManagerChanged(newValue) {
    if (!newValue) return;
    this.numPlayers = newValue.data.DefaultNumPlayers;
  }

  _computeManagerHasAgents(selectedManager) {
    if (!selectedManager) return false;
    return selectedManager.data.Agents.length == 0;
  }

  _computeManagerHasFixedPlayerCount(selectedManager) {
    if (!selectedManager) return false;
    let data = selectedManager.data;
    return data.MinNumPlayers == data.MaxNumPlayers;
  }

  serialize() {
    var body = {};

    var eles = [...this.shadowRoot.querySelectorAll("paper-radio-group")];

    eles = eles.concat([...this.shadowRoot.querySelectorAll("paper-slider")]);
    eles = eles.concat([...this.shadowRoot.querySelectorAll("paper-dropdown-menu")]);
    eles = eles.concat([...this.shadowRoot.querySelectorAll("paper-toggle-button")]);

    for (var i = 0; i < eles.length; i++) {
      var ele = eles[i];
      let name = ele.name;
      if (ele.disabled) continue;
      if (ele.localName == "paper-radio-group" || ele.localName == "paper-dropdown-menu") {
        ele = ele.selectedItem;
      }
      if (ele.localName == "paper-toggle-button") {
        body[ele.name] = ele.checked ? "1" : "0";
        continue;
      }
      body[name] = ele.value;
    }

    return body;
  }

  _createGameResponseChanged(newValue) {
    if (newValue.Status == "Success") {
      this.dispatchEvent(new CustomEvent("navigate-to", {composed: true, detail: this.GamePath(newValue.GameName, newValue.GameId)}));
    } else {
      this.dispatchEvent(new CustomEvent("show-error", {composed: true, detail:{message:newValue.Error, friendlyMessage: newValue.FriendlyError}}));
    }
  }

  createGame() {
    if (!this.loggedIn) {
      this.dispatchEvent(new CustomEvent('show-login', {composed: true, detail: {nextAction:this.createGame.bind(this)}}));
      return;
    }
    this.$.create.body = this.serialize();
    this.$.create.generateRequest();
  }
}

customElements.define(BoardgameCreateGame.is, BoardgameCreateGame);
