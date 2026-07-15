import { GameRenderer } from './_game_renderer.js';
import { html, css } from 'lit';

/**
 * Werewolf solo renderer (non-companion mode). Minimal fallback that shows
 * game state as text. Werewolf is designed for companion mode, so this is
 * deliberately simple.
 */
class BoardgameRenderGameWerewolf extends GameRenderer {
  static override styles = [
    ...(GameRenderer.styles ? [GameRenderer.styles] : []),
    css`
      :host {
        display: block;
        padding: 24px;
        font-family: system-ui, sans-serif;
      }
      pre {
        white-space: pre-wrap;
        font-size: 14px;
      }
    `
  ];

  override render() {
    return html`
      <h2>Werewolf</h2>
      <pre>${this.diagram || 'Waiting for state...'}</pre>
    `;
  }
}

customElements.define('boardgame-render-game-werewolf', BoardgameRenderGameWerewolf);
