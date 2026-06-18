import { BoardgameBaseGameRenderer } from '../../src/components/boardgame-base-game-renderer.js';
import { html, css } from 'lit';
import type { MoveName } from './_move_names.js';
import type { GameState, PlayerState } from './_types.js';

/**
 * Werewolf solo renderer (non-companion mode). Minimal fallback that shows
 * game state as text. Werewolf is designed for companion mode, so this is
 * deliberately simple.
 */
class BoardgameRenderGameWerewolf extends BoardgameBaseGameRenderer<GameState, PlayerState, MoveName> {
  static override styles = [
    ...(BoardgameBaseGameRenderer.styles ? [BoardgameBaseGameRenderer.styles] : []),
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
