import { PolymerElement } from '@polymer/polymer/polymer-element.js';
import '../../src/components/boardgame-status-text.js';
import { html } from '@polymer/polymer/lib/utils/html-tag.js';

class BoardgameRenderPlayerInfoPig extends PolymerElement {
  static get template() {
    return html`
    <div>Round Score <boardgame-status-text>{{playerState.RoundScore}}</boardgame-status-text></div>
    <div>Total Score <boardgame-status-text>{{playerState.TotalScore}}</boardgame-status-text></div>
`;
  }

  static get is() {
    return "boardgame-render-player-info-pig"
  }

  static get properties() {
    return {
      state: Object,
      playerIndex: Number,
      playerState: Object,
    }
  }
}

customElements.define(BoardgameRenderPlayerInfoPig.is, BoardgameRenderPlayerInfoPig);
