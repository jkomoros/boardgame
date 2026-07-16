import { html } from '../../src/client.js';
import { PlayerInfoRenderer } from './_game_renderer.js';

class BoardgameRenderPlayerInfoCheckers extends PlayerInfoRenderer {
  override render() {
    return html`
      Number of cards:
      <boardgame-status-text .value=${this.playerState?.Hand.Indexes.length ?? 0}></boardgame-status-text>
    `;
  }
}

customElements.define('boardgame-render-player-info-checkers', BoardgameRenderPlayerInfoCheckers);
