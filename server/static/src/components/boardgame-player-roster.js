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
import '@polymer/paper-dialog/paper-dialog.js';
import '@polymer/paper-button/paper-button.js';
import '@polymer/iron-flex-layout/iron-flex-layout-classes.js';
import '@polymer/paper-styles/typography.js';
import '@polymer/paper-styles/color.js';
import './boardgame-configure-game-properties.js';
import './boardgame-player-roster-item.js';
import './boardgame-ajax.js';
import './shared-styles.js';
import { html } from '@polymer/polymer/lib/utils/html-tag.js';

class BoardgamePlayerRoster extends PolymerElement {
  static get template() {
    return html`
    <style is="custom-style" include="iron-flex shared-styles">
      h3 {
        margin:0;
      }


    </style>
    <div class="layout horizontal center">
      <h3 class="flex">[[_bannerText(finished, winners)]]</h3>
      <boardgame-configure-game-properties game-visible="[[gameVisible]]" game-open="[[gameOpen]]" admin="[[admin]]" is-owner="[[isOwner]]" game-route="[[gameRoute]]" configurable="">
      </boardgame-configure-game-properties>
    </div>
    <div class="layout horizontal justified players">
      <template is="dom-repeat" items="{{playersInfo}}">
        <boardgame-player-roster-item class="flex" state="[[state]]" game-name="[[gameRoute.name]]" is-empty="[[item.IsEmpty]]" finished="[[finished]]" winner="[[_isWinner(index, winners)]]" is-agent="[[item.IsAgent]]" photo-url="[[item.PhotoUrl]]" display-name="[[item.DisplayName]]" player-index="[[index]]" viewing-as-player="[[viewingAsPlayer]]" current-player-index="[[currentPlayerIndex]]" renderer-loaded="[[rendererLoaded]]" active="[[active]]">
        </boardgame-player-roster-item>
      </template>
    </div>
    <div hidden\$="[[!isObserver]]">
      <div class="layout horizontal center">
        <h3 class="flex">
          Observing
        </h3>
        <div hidden\$="[[!showJoin]]">
          <paper-button on-tap="showDialog" raised="" default="">Join game</paper-button>
        </div>
      </div>
    </div>
    <paper-dialog id="join">
      <h2>Join game?</h2>
      <p>We're still looking for players for this game.</p>
      <div class="buttons">
        <paper-button dialog-dismiss="">I'll just watch</paper-button>
        <paper-button dialog-confirm="" default="" autofocus="">I'm in!</paper-button>
      </div>
    </paper-dialog>
    <boardgame-ajax id="request" game-path="join" game-route="[[gameRoute]]" handle-as="json" method="POST" last-response="{{response}}">
  </boardgame-ajax>
`;
  }

  static get is() {
    return "boardgame-player-roster"
  }

  static get properties() {
    return {
      viewingAsPlayer: Number,
      hasEmptySlots: Boolean,
      gameOpen: Boolean,
      gameVisible: Boolean,
      gameRoute: {
        type: Object,
        observer: "_gameRouteChanged"
      },
      active: Boolean,
      admin: Boolean,
      isOwner: Boolean,
      playersInfo: Array,
      currentPlayerIndex: Number,
      state: Object,
      isObserver: {
        type: Number,
        computed: "_computeIsObserver(viewingAsPlayer)"
      },
      showJoin: {
        type: Boolean,
        computed: "_computeShowJoin(viewingAsPlayer, hasEmptySlots, gameOpen)"
      },
      finished: Boolean,
      winners: Array,
      loggedIn: Boolean,
      response: {
        type: Object,
        observer: "_responseChanged",
      },
      //TODO: there must be a better way to do constants...
      OBSERVER_PLAYER_INDEX : {
        type: Number,
        value: -1,
      },
      ADMIN_PLAYER_INDEX: {
        type: Number,
        value: -2,
      },
      rendererLoaded: {
        type: Boolean,
        value: false,
      }
    }
  }

  ready() {
    super.ready();
    this.addEventListener('iron-overlay-closed', e => this.dialogClosed(e));
  }

  _isWinner(index, winners) {
    if (!winners) return false;
    for (var i = 0; i < winners.length; i++) {
      if (winners[i] == index) {
        return true;
      }
    }
    return false;
  }

  _bannerText(finished, winners) {
    if (!finished) {
      return "Playing"
    }
    return "Game Over"
  }

  playerName(viewingAsPlayer) {
    if (viewingAsPlayer == this.ADMIN_PLAYER_INDEX) return "Admin"
    return "player " + viewingAsPlayer;
  }

  _computeIsObserver(viewingAsPlayer) {
    return viewingAsPlayer == this.OBSERVER_PLAYER_INDEX;
  }

  _computeShowJoin(viewingAsPlayer, hasEmptySlots, gameOpen) {
    return viewingAsPlayer == this.OBSERVER_PLAYER_INDEX && hasEmptySlots && gameOpen;
  }

  showDialog() {
    if (this.$.join.opened) return;
    if (this.viewingAsPlayer != this.OBSERVER_PLAYER_INDEX) return;
    this.$.join.open();
  }

  dialogClosed(e) {

    //If it wasn't confirmed, it was effectively canceled.

    if (!e.detail.confirmed) return;

    this.doJoin();
  }

  doJoin() {
    if (!this.loggedIn) {
      this.dispatchEvent(new CustomEvent('show-login', {composed: true, detail: {nextAction:this.doJoin.bind(this)}}));
      return;
    }
    this.$.request.generateRequest();
  }

  _gameRouteChanged(newValue) {
    if (!newValue) return;
    this.rendererLoaded = false;
    import("../game-src/" +newValue.name + "/boardgame-render-player-info-" + newValue.name + ".js").then(this._rendererLoaded.bind(this), null);
  }

  _rendererLoaded(e) {
    this.rendererLoaded = true;
  }

  _responseChanged(newValue) {
    if (!newValue) return;

    if (newValue.Status == "Success") {
      //Tell game-view to fetch data now
      this.dispatchEvent(new CustomEvent("refresh-info", {composed: true}));
    } else {
      this.dispatchEvent(new CustomEvent("show-error", {composed: true, detail: {"message" : newValue.Error,"friendlyMessage": newValue.FriendlyError, "title": "Couldn't Join"}}));
    }
  }
}

customElements.define(BoardgamePlayerRoster.is, BoardgamePlayerRoster);
