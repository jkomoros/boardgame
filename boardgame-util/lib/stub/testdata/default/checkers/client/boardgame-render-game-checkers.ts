import { css, html } from '../../src/client.js';
import { GameRenderer, registerGameRenderer } from './_game_renderer.js';

@registerGameRenderer
export class BoardgameRenderGameCheckers extends GameRenderer {
  static override styles = css`
    :host { display: block; }
  `;

  override render() {
    return html`
      <p>Build your game renderer here. State and move names are strictly typed.</p>
    `;
  }
}
