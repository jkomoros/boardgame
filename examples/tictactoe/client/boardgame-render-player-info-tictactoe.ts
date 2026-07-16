import { html } from '../../src/client.js';
import { PlayerInfoRenderer, registerPlayerInfoRenderer } from './_game_renderer.js';

@registerPlayerInfoRenderer
export class BoardgameRenderPlayerInfoTictactoe extends PlayerInfoRenderer {
  override get chip() {
    return { text: this.playerState?.TokenValue ?? '' };
  }

  // chipColor intentionally not overridden - uses framework computed color

  override render() {
    return html``;
  }
}
