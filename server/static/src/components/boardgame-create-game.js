import '@polymer/paper-button/paper-button.js';
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

import { connect } from 'pwa-helpers/connect-mixin.js';
import { store } from '../store.js';

import { LitElement, html } from '@polymer/lit-element';

import { SharedStyles } from './shared-styles-lit.js';

import {
  createGame,
  updateSelectedMangerIndex,
  updateNumPlayers,
  updateAgentName,
  updateVariantOption,
  updateOpen,
  updateVisible
} from '../actions/list.js';

import {
  selectManagers,
  selectSelectedManagerIndex,
  selectCreateGameNumPlayers,
  selectCreateGameAgents,
  selectCreateGameOpen,
  selectCreateGameVisible,
  selectCreateGameVariantOptions
} from '../selectors.js';

//The templates are a pain to tell to expect an empty manager, so have a blank
//one for use in tempaltes. Every time the templates below rely on a new
//property of selected manager we should change this.
const EMPTY_MANAGER = {
  DefaultNumPlayers: 0,
  MinNumPlayers: 0,
  MaxNumPlayers: 0,
  Variant: [],
  Agents: [],
}

class BoardgameCreateGame extends connect(store)(LitElement) {
  render() {
    return html`
    ${SharedStyles}
    <style>
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

      .layout {
        display: flex;
      }

      .vertical {
        flex-direction: column;
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
        flex-grow: 1;
      }
    </style>
    <div class="vertical layout">
      <div class="horizontal layout center game">
        <paper-dropdown-menu name="manager" label="Game Type" horizontal-align="left">
          <paper-listbox slot="dropdown-content" .selected=${this._selectedManagerIndex} @selected-changed=${this._handleSelectedManagerIndexChanged}>
          ${this._managers.map((item) =>
            html`
              <paper-item .value=${item.Name} .label=${item.DisplayName}>
                <paper-item-body two-line>
                  <div>${item.DisplayName}</div>
                  <div secondary>${item.Description}</div>
                </paper-item-body>
              </paper-item>`
            )}
          </paper-listbox>
        </paper-dropdown-menu>
        <div class="vertical layout">
            ${this._managerHasFixedPlayerCount ? 
             html`<div class="secondary"><strong>${this._selectedManager.MinNumPlayers}</strong> players</div>` :
              html`<div class="secondary">Number of Players</div>
              <paper-slider name="numplayers" label="Number of Players" .min=${this._selectedManager.MinNumPlayers} .max=${this._selectedManager.MaxNumPlayers} .value=${this._numPlayers} @change=${this._handleSliderValueChanged} snaps pin editable max-markers="100"></paper-slider>`
            }
        </div>
        <div class="flex"></div>
        <paper-button @tap=${this.createGame} default raised>Create Game</paper-button>
      </div>

      <div class="horizontal layout justified" ?hidden=${this._managerHasAgents}>
      ${this._agents.map((item, index) => html`
          <div class="flex">
            <div class="vertical layout">
              Player ${index}
              <paper-radio-group .selected=${item} .disabled=${this._managerHasAgents} .name=${"agent-player-" + index} attr-for-selected="value" .index=${index} @selected-changed=${this._handleAgentSelectedChanged}>
                <paper-radio-button .name=${"agent-player-" + index} .disabled=${this._managerHasAgents} .value=${""}>Real Live Human</paper-radio-button>
                ${this._selectedManager.Agents.map((item, index) => html`
                <paper-radio-button .name=${"agent-player-" + index} .value=${item.Name}>${item.DisplayName}</paper-radio-button>
                `)}
              </paper-radio-group>
            </div>
          </div>
      `)}
      </div>
      <div class="horizontal layout variant">
        ${this._variants.map((item, index) => html`
          <div class="vertical layout">
            <paper-dropdown-menu .label=${item.DisplayName} .name=${"variant_" + item.Name} horizontal-align="left">
              <paper-listbox slot="dropdown-content" .selected=${this._variantOptions[index]} @selected-changed=${this._handleVariantOptionChanged} .index=${index}>
                ${item.Values.map(item => html`
                  <paper-item .value=${item.Value} .label=${item.DisplayName}>
                    <paper-item-body two-line>
                      <div>${item.DisplayName}</div>
                      <div secondary>${item.Description}</div>
                    </paper-item-body>
                  </paper-item>
                `)}
              </paper-listbox>
            </paper-dropdown-menu>
            <div class="secondary">${item.Description}</div>
          </div>
        `)}

      </div>
      <div class="horizontal layout">
        <paper-toggle-button name="visible" .checked=${this._visible} @checked-changed=${this._handleVisibleChanged}><iron-icon icon="visibility"></iron-icon> Allow strangers to find the game</paper-toggle-button>
        <paper-toggle-button name="open" .checked=${this._open} @checked-changed=${this._handleOpenChanged}><iron-icon icon="social:people"></iron-icon> Allow anyone who can view the game to join</paper-toggle-button>
      </div>
    </div>
`;
  }

  static get properties() {
    return {
      _selectedManagerIndex: { type: Number },
      _managers: { type: Array },
      _numPlayers: { type: Number },
      _agents: { type: Array },
      _variantOptions: { type: Array },
      _open: { type: Boolean },
      _visible: { type: Boolean}
    }
  }

  constructor() {
    super();

    this._managers = [];
    this._numPlayers = 0;
  }

  stateChanged(state) {
    this._managers = selectManagers(state);
    this._selectedManagerIndex = selectSelectedManagerIndex(state);
    this._numPlayers = selectCreateGameNumPlayers(state);
    this._agents = selectCreateGameAgents(state);
    this._variantOptions = selectCreateGameVariantOptions(state);
    this._open = selectCreateGameOpen(state);
    this._visible = selectCreateGameVisible(state);
  }

  _handleSelectedManagerIndexChanged(e) {
    store.dispatch(updateSelectedMangerIndex(e.detail.value));
  }

  _handleOpenChanged(e) {
    store.dispatch(updateOpen(e.composedPath()[0].checked));
  }

  _handleVisibleChanged(e) {
    store.dispatch(updateVisible(e.composedPath()[0].checked));
  }

  get _selectedManager() {
    //TODO: this and other getters should probably be selectors
    if (!this._managers || this._selectedManagerIndex < 0 || this._selectedManagerIndex >= this._managers.length) return EMPTY_MANAGER;
    return this._managers[this._selectedManagerIndex];
  }

  get _variants() {
    //This is here because games with no variant form server will have null for
    //that field, not an empty array, and the template requires an array.
    return this._selectedManager.Variant || [];
  }

  _handleAgentSelectedChanged(e) {
    const group = e.composedPath()[0];
    store.dispatch(updateAgentName(group.index, group.selected));
  }

  _handleVariantOptionChanged(e) {
    const listbox = e.composedPath()[0];
    store.dispatch(updateVariantOption(listbox.index, listbox.selected));
  }

  _handleSliderValueChanged(e) {
    store.dispatch(updateNumPlayers(e.composedPath()[0].value));
  }

  get _managerHasAgents() {
    return this._selectedManager.Agents.length == 0;
  }

  get _managerHasFixedPlayerCount() {
    return this._selectedManager.MinNumPlayers == this._selectedManager.MaxNumPlayers;
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

  createGame() {
    store.dispatch(createGame(this.serialize()));
  }
}

customElements.define("boardgame-create-game", BoardgameCreateGame);
