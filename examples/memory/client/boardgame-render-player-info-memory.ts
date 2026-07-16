import { html } from '../../src/client.js';
import { PlayerInfoRenderer } from './_game_renderer.js';

class BoardgameRenderPlayerInfoMemory extends PlayerInfoRenderer {
  override render() {
    return html`
      Won Cards <boardgame-status-text .value=${this.playerState?.WonCards?.Indexes?.length}></boardgame-status-text>
    `;
  }
}

customElements.define('boardgame-render-player-info-memory', BoardgameRenderPlayerInfoMemory);
