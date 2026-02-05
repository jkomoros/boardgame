import { BoardgameBaseGameRenderer } from '../../../server/static/src/components/boardgame-base-game-renderer.js';
import '../../../server/static/src/components/boardgame-board.js';
import '../../../server/static/src/components/boardgame-token.js';
import { html, css } from 'lit';
import { property } from 'lit/decorators.js';
import { repeat } from 'lit/directives/repeat.js';

class BoardgameRenderGameCheckers extends BoardgameBaseGameRenderer {
  static override styles = [
    ...(BoardgameBaseGameRenderer.styles ? [BoardgameBaseGameRenderer.styles] : []),
    css`
      boardgame-token {
        --component-scale: 1.25;
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
          <boardgame-token color="red"></boardgame-token>
        `)}
        ${repeat(this._components, (item, index) => index, () => html`
          <boardgame-token color="black"></boardgame-token>
        `)}
      </boardgame-board>
    `;
  }
}

customElements.define('boardgame-render-game-checkers', BoardgameRenderGameCheckers);
