import { html } from '../../src/client.js';
import { PlayerInfoRenderer, registerPlayerInfoRenderer } from './_game_renderer.js';

@registerPlayerInfoRenderer
export class BoardgameRenderPlayerInfoPig extends PlayerInfoRenderer {
  override render() {
    return html`
      <div><boardgame-stat label="Round Score" .value=${this.playerState?.RoundScore}></boardgame-stat></div>
      <div><boardgame-stat label="Total Score" .value=${this.playerState?.Score}></boardgame-stat></div>
    `;
  }
}
