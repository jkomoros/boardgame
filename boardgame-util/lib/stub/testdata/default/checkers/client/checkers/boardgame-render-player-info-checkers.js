import { PolymerElement } from '@polymer/polymer/polymer-element.js';
import { html } from '@polymer/polymer/lib/utils/html-tag.js';


class BoardgameRenderPlayerInfoCheckers extends PolymerElement {

  static get template() {
  	
		return html`This is where you render info on player, typically using &lt;boardgame-status-text&gt;.`;
	

  }

  static get is() {
    return "boardgame-render-player-info-checkers"
  }

  

}

customElements.define(BoardgameRenderPlayerInfoCheckers.is, BoardgameRenderPlayerInfoCheckers);
