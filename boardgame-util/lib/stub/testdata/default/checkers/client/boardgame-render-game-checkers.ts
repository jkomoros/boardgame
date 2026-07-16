import { css, html } from '../../src/client.js';
import { GameRenderer } from './_game_renderer.js';

class BoardgameRenderGameCheckers extends GameRenderer {
  static override styles = css`
    :host { display: block; }
    .players { display: flex; flex-wrap: wrap; gap: 1rem; }
    .player { flex: 1 1 12rem; }
  `;

  override render() {
    return html`
      <p>Build your game renderer here. State and move names are strictly typed.</p>
    `;
  }
}

customElements.define('boardgame-render-game-checkers', BoardgameRenderGameCheckers);
