import { html } from '../../src/client.js';
import { PlayerInfoRenderer } from './_game_renderer.js';
import '../../src/components/boardgame-status-text.js';

class BoardgameRenderPlayerInfoCheckers extends PlayerInfoRenderer {
  override render() {
    return html`
      Number of cards:
      <boardgame-status-text>${this.playerState?.Hand.Indexes.length ?? 0}</boardgame-status-text>
    `;
  }
}

customElements.define('boardgame-render-player-info-checkers', BoardgameRenderPlayerInfoCheckers);
