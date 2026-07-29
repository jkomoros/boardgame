import { html } from '../../src/client.js';
import { PlayerInfoRenderer, registerPlayerInfoRenderer } from './_game_renderer.js';

@registerPlayerInfoRenderer
export class BoardgameRenderPlayerInfoMemory extends PlayerInfoRenderer {
  override render() {
    return html`
      <boardgame-stat label="Won Cards" .stack=${this.playerState?.WonCards ?? null}></boardgame-stat>
    `;
  }
}
