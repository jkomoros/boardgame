import { html } from '../../src/client.js';
import { PlayerInfoRenderer, registerPlayerInfoRenderer } from './_game_renderer.js';

@registerPlayerInfoRenderer
export class BoardgameRenderPlayerInfoPig extends PlayerInfoRenderer {
  override render() {
    return html`
      <div>Round Score <boardgame-status-text .value=${this.playerState?.RoundScore}></boardgame-status-text></div>
      <div>Total Score <boardgame-status-text .value=${this.playerState?.Score}></boardgame-status-text></div>
    `;
  }
}
