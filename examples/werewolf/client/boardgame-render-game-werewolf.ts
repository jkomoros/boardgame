import { GameRenderer, registerGameRenderer } from './_game_renderer.js';
import { html, css } from 'lit';

/**
 * Werewolf solo renderer (non-companion mode). Minimal fallback that shows
 * game state as text. Werewolf is designed for companion mode, so this is
 * deliberately simple.
 */
@registerGameRenderer
export class BoardgameRenderGameWerewolf extends GameRenderer {
  static override styles = [
    ...(GameRenderer.styles ? [GameRenderer.styles] : []),
    css`
      :host {
        display: block;
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
      <boardgame-game-surface heading="Werewolf">
        <pre>${this.diagram || 'Waiting for state...'}</pre>
      </boardgame-game-surface>
    `;
  }
}
