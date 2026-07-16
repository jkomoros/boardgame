import { html } from '../../src/client.js';
import { PlayerInfoRenderer } from './_game_renderer.js';

class BoardgameRenderPlayerInfoTictactoe extends PlayerInfoRenderer {
  override get chip() {
    return { text: this.playerState?.TokenValue ?? '' };
  }

  // chipColor intentionally not overridden - uses framework computed color

  override render() {
    return html``;
  }
}

customElements.define('boardgame-render-player-info-tictactoe', BoardgameRenderPlayerInfoTictactoe);
