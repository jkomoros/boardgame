import { BoardgameBaseGameRenderer } from '../../../server/static/src/components/boardgame-base-game-renderer.js';
import '../../../server/static/src/components/boardgame-board.js';
import '../../../server/static/src/components/boardgame-token.js';
import { html, css } from 'lit';
import { property } from 'lit/decorators.js';
import { repeat } from 'lit/directives/repeat.js';
import type { MoveName } from './_move_names.js';
import type { GameState, PlayerState } from './_types.js';

class BoardgameRenderGameCheckers extends BoardgameBaseGameRenderer<GameState, PlayerState, MoveName> {
  static override styles = [
    ...(BoardgameBaseGameRenderer.styles ? [BoardgameBaseGameRenderer.styles] : []),
    css`
      boardgame-token {
        --component-scale: 1.25;
      }
      boardgame-token.player-0 {
        color: var(--player-0-color, #424242);
      }
      boardgame-token.player-1 {
        color: var(--player-1-color, #D32F2F);
      }
    `
  ];

  @property({ type: Number })
  size = 8;

  get _components(): boolean[] {
    return this._computeComponents(this.size);
  }

  private _computeComponents(size: number): boolean[] {
    const result: boolean[] = [];
    for (let i = 0; i < size; i++) {
      result.push(true);
    }
    return result;
  }

  override render() {
    return html`
      <boardgame-board .rows="${this.size}" .cols="${this.size}">
        ${repeat(this._components, (item, index) => index, () => html`
          <boardgame-token class="player-0"></boardgame-token>
        `)}
        ${repeat(this._components, (item, index) => index, () => html`
          <boardgame-token class="player-1"></boardgame-token>
        `)}
      </boardgame-board>
    `;
  }
}

customElements.define('boardgame-render-game-checkers', BoardgameRenderGameCheckers);
