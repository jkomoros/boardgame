import { Element } from '@polymer/polymer/polymer-element.js';
import { html } from '@polymer/polymer/lib/utils/html-tag.js';

class BoardgameRenderPlayerInfoTictactoe extends Element {
  static get template() {
    return html`

`;
  }

  static get is() {
    return "boardgame-render-player-info-tictactoe"
  }

  static get properties() {
    return {
      state: Object,
      playerIndex: Number,
      chipText: {
        type: String,
        notify: true,
        computed: "_computeChipText(state, playerIndex)"
      },
      chipColor: {
        type: String,
        notify: true,
        computed: "_computeChipColor(state, playerIndex)"
      }
    }
  }

  _computeChipColor(state, playerIndex) {
    if (state.Players[playerIndex].TokenValue == "X") {
      return "blue";
    }
    return "red";
  }

  _computeChipText(state, playerIndex) {
    return state.Players[playerIndex].TokenValue
  }
}

customElements.define(BoardgameRenderPlayerInfoTictactoe.is, BoardgameRenderPlayerInfoTictactoe);
