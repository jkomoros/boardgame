import { html } from '../../src/client.js';
import { PlayerInfoRenderer, registerPlayerInfoRenderer } from './_game_renderer.js';

@registerPlayerInfoRenderer
export class BoardgameRenderPlayerInfoCheckers extends PlayerInfoRenderer {
  override render() {
    return html`<p>Render player summary information here.</p>`;
  }
}
