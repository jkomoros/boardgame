import { html } from '../../src/client.js';
import { PlayerInfoRenderer, registerPlayerInfoRenderer } from './_game_renderer.js';

@registerPlayerInfoRenderer
export class BoardgameRenderPlayerInfoCheckers extends PlayerInfoRenderer {
  override render() {
    return html`
      <boardgame-stat label="Captured" .stack=${this.playerState?.CapturedTokens ?? null}></boardgame-stat>
    `;
  }
}
