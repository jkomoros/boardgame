import { PolymerElement } from '@polymer/polymer/polymer-element.js';
import '../../src/components/boardgame-status-text.js';
import { html } from '@polymer/polymer/lib/utils/html-tag.js';

class BoardgameRenderPlayerInfoBlackjack extends PolymerElement {
  static get template() {
    return html`
    <div>Score <strong>{{playerState.Computed.HandValue}}</strong></div>
    <div><boardgame-status-text>{{_calculateStatus(playerState)}}</boardgame-status-text></div>
`;
  }

  static get is() {
    return "boardgame-render-player-info-blackjack"
  }

  static get properties() {
    return {
      state: Object,
      playerIndex: Number,
      playerState: Object,
    }
  }

  _calculateStatus(playerState) {
    if (playerState.Busted) {
      return "Busted"
    }
    if (playerState.Stood) {
      return "Stood"
    }
    //Non breakable space so when the first player busts the layout doesn't jump
    return "\xa0"
  }
}

customElements.define(BoardgameRenderPlayerInfoBlackjack.is, BoardgameRenderPlayerInfoBlackjack);
