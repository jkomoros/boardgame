import { html } from '../../src/client.js';
import { PlayerInfoRenderer } from './_game_renderer.js';

class BoardgameRenderPlayerInfoCheckers extends PlayerInfoRenderer {
  override render() {
    return html`<p>Render player summary information here.</p>`;
  }
}

customElements.define('boardgame-render-player-info-checkers', BoardgameRenderPlayerInfoCheckers);
