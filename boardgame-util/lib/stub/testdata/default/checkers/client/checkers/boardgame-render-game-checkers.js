import { BoardgameBaseGameRenderer } from '../../src/boardgame-base-game-renderer.js';
import { html } from '@polymer/polymer/lib/utils/html-tag.js';

class BoardgameRenderGameCheckers extends BoardgameBaseGameRenderer {

  static get template() {
  	
return html`This is where you game should render itself. See boardgame/server/README.md for more on the components you can use, or check out the examples in boardgame/examples.`;

  }

  static get is() {
    return "boardgame-render-game-checkers"
  }

  //We don't need to compute any properties that BoardgameBaseGamErenderer
  //doesn't have.

}

customElements.define(BoardgameRenderGameCheckers.is, BoardgameRenderGameCheckers);

