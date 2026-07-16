import { html } from '../../src/client.js';
import { PlayerInfoRenderer, registerPlayerInfoRenderer } from './_game_renderer.js';

@registerPlayerInfoRenderer
export class BoardgameRenderPlayerInfoMemory extends PlayerInfoRenderer {
  override render() {
    return html`
      Won Cards <boardgame-status-text .value=${this.playerState?.WonCards?.Indexes?.length}></boardgame-status-text>
    `;
  }
}
