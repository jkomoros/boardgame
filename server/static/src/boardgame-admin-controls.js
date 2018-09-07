import { Element } from '@polymer/polymer/polymer-element.js';
import '@polymer/iron-flex-layout/iron-flex-layout-classes.js';
import '@polymer/paper-radio-button/paper-radio-button.js';
import '@polymer/paper-radio-group/paper-radio-group.js';
import '@polymer/iron-input/iron-input.js';
import '@polymer/paper-checkbox/paper-checkbox.js';
import '@polymer/polymer/lib/elements/dom-if.js';
import './boardgame-move-form.js';
import './shared-styles.js';
import { html } from '@polymer/polymer/lib/utils/html-tag.js';

class BoardgameAdminControls extends Element {
  static get template() {
    return html`
    <style include="iron-flex shared-styles">

    </style>
    <div hidden="{{!active}}">
      <div class="card horizontal layout admin center">
        <div class="flex">
          View as
          <paper-radio-group selected="{{viewAs}}">
            <paper-radio-button name="admin">Admin</paper-radio-button>
            <paper-radio-button name="observer">Observer</paper-radio-button>
            <paper-radio-button name="current">Current Player</paper-radio-button>
            <paper-radio-button name="custom">Custom</paper-radio-button>
          </paper-radio-group>
          <input is="iron-input" type="number" value-as-number="{{customRequestedPlayer::input}}" min="0" max="{{maxRequestedPlayerIndex}}">
        </div>
        <div>
          <paper-checkbox id="move-as-player" checked="{{makeMovesAsViewingAsPlayer}}">Make Moves As ViewingAsPlayer</paper-checkbox>
        </div>
      </div>
      <template is="dom-if" if="{{!game.Finished}}">
        <div class="card">
          <boardgame-move-form admin="{{active}}" move-as-player="{{moveAsPlayer}}" id="moves" config="{{moveForms}}" game-route="[[gameRoute]]"></boardgame-move-form>
        </div>
      </template>
      <div class="card">
        <details>
          <summary>State</summary>
          <pre>{{gameState}}</pre>
        </details>
      </div>
      <div class="card">
        <details>
          <summary>Chest</summary>
          <pre>{{_chestAsString}}</pre>
        </details>
      </div>
    </div>
`;
  }

  static get is() {
    return "boardgame-admin-controls";
  }

  static get properties() {
    return {
      active: {
        type: Boolean,
        value: false,
      },
      gameRoute: Object,
      //Valid values: 'custom', 'admin', 'current', 'observer'
      viewAs: {
        type: String,
        value: "current"
      },
      requestedPlayer: {
        type: Number,
        value: 0,
        notify: true,
        computed: "_computeRequestedPlayer(active, viewingAsPlayer, viewAs, customRequestedPlayer, game.CurrentPlayerIndex)"
      },
      customRequestedPlayer: {
        type: Number,
        value: 0
      },
      maxRequestedPlayerIndex: {
        type: Number,
        computed: "_computeMaxRequestedPlayerIndex(game)"
      },
      makeMovesAsViewingAsPlayer: {
        type:Boolean,
        value: true,
      },
      moveAsPlayer: {
        type:Number,
        computed: "_computeMoveAsPlayer(viewingAsPlayer, makeMovesAsViewingAsPlayer)"
      },
      viewingAsPlayer: {
        type: Number,
        value: 0,
      },
      autoCurrentPlayer: {
        type: Boolean,
        computed: "_computeAutoCurrentPlayer(active, viewAs)",
        notify: true
      },
      chest : Object,
      _chestAsString: {
        type: String,
        computed: "_computeChestAsString(chest)"
      },
      gameState : String,
      //TODO: there must be a better way to do constants...
      OBSERVER_PLAYER_INDEX : {
        type: Number,
        value: -1,
      },
      ADMIN_PLAYER_INDEX: {
        type: Number,
        value: -2,
      },
      moveForms: Object,
      game: Object
    }
  }

  _computeChestAsString(chest) {
    return JSON.stringify(chest, null, 2);
  }

  proposeMove(moveName, moveArguments) {
    let movesEle = this.shadowRoot.querySelector("#moves");

    if (!movesEle) {
      console.warn("propose-move fired, but no moves element to forward to.");
      return;
    }

    movesEle.proposeMove(moveName, moveArguments);
  }

  _computeMoveAsPlayer(viewingAsPlayer, moveAsViewingAsPlayer) {
    if (moveAsViewingAsPlayer) return viewingAsPlayer;
    return this.ADMIN_PLAYER_INDEX;
  }

  _computeMaxRequestedPlayerIndex(game) {
    if (!game) {
      return 0;
    }
    return game.NumPlayers - 1;
  }

  _computeAutoCurrentPlayer(active, viewAs) {
    if (!active) return false;
    return viewAs == "current";
  }

  _computeRequestedPlayer(admin, viewingAsPlayer, viewAs, customRequestedPlayer, currentPlayer) {
    if (!admin) return viewingAsPlayer;
    switch (viewAs) {
      case "admin":
        return this.ADMIN_PLAYER_INDEX;
      case "observer":
        return this.OBSERVER_PLAYER_INDEX;
      case "custom":
        return customRequestedPlayer;
      case "current":
        //When we first come back to a view with admin mode already on
        //currentPlayer will be undefined because game is not yet set. In
        //that case just default to player 0; the info result will have
        //the CurrentPlayerIndex which will then recalc this and set to
        //its proper value.
        return currentPlayer || 0;
    }
  }
}

customElements.define(BoardgameAdminControls.is, BoardgameAdminControls);
