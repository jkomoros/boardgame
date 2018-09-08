import { PolymerElement } from '@polymer/polymer/polymer-element.js';
import { html } from '@polymer/polymer/lib/utils/html-tag.js';

import '../../src/boardgame-status-text.js';


class BoardgameRenderPlayerInfoCheckers extends Polymer.Element {

  static get template() {
  	
		return html`Number of Cards <boardgame-status-text>{{playerState.Hand.Indexes.length}}</boardgame-status-text>`;
	

  }

  static get is() {
    return "boardgame-render-player-info-checkers"
  }

  
  static get properties() {
    return {
      state: Object,
      playerIndex: Number,
      playerState: Object,
    }
  }
  

}

customElements.define(BoardgameRenderPlayerInfoCheckers.is, BoardgameRenderPlayerInfoCheckers);
