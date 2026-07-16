import { html } from '../../src/client.js';
import { PlayerInfoRenderer, registerPlayerInfoRenderer } from './_game_renderer.js';

@registerPlayerInfoRenderer
export class BoardgameRenderPlayerInfoCheckers extends PlayerInfoRenderer {
  override render() {
    return html`
      Number of cards:
      <boardgame-status-text .value=${this.playerState?.Hand.Indexes.length ?? 0}></boardgame-status-text>
    `;
  }
}
