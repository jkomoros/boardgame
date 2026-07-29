import { html } from '../../src/client.js';
import { PlayerInfoRenderer, registerPlayerInfoRenderer } from './_game_renderer.js';

@registerPlayerInfoRenderer
export class BoardgameRenderPlayerInfoDebuganimations extends PlayerInfoRenderer {
  override render() {
    return html`
      <boardgame-stat label="Cards" .stack=${this.playerState?.Hand ?? null}></boardgame-stat>
    `;
  }
}
