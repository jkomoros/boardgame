import { Element } from '@polymer/polymer/polymer-element.js';
import '../../src/boardgame-status-text.js';
import { html } from '@polymer/polymer/lib/utils/html-tag.js';

class BoardgameRenderPlayerInfoMemory extends Element {
  static get template() {
    return html`
    Won Cards <boardgame-status-text>{{playerState.WonCards.Indexes.length}}</boardgame-status-text>
`;
  }

  static get is() {
    return "boardgame-render-player-info-memory"
  }

  static get properties() {
    return {
      state: Object,
      playerIndex: Number,
      playerState: Object,
    }
  }
}

customElements.define(BoardgameRenderPlayerInfoMemory.is, BoardgameRenderPlayerInfoMemory);
