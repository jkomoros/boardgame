import { BoardgameBaseGameRenderer } from '../../../server/static/src/components/boardgame-base-game-renderer.js';
import '../../../server/static/src/components/boardgame-fading-text.js';
import './boardgame-tictactoe-cell.js';
import { html, css } from 'lit';

class BoardgameRenderGameTictactoe extends BoardgameBaseGameRenderer {
  static override styles = [
    ...(BoardgameBaseGameRenderer.styles ? [BoardgameBaseGameRenderer.styles] : []),
    css`
      .row {
        border-bottom: 1px solid black;
        display: flex;
        flex-direction: row;
      }

      .row:last-of-type {
        border-bottom: 0;
      }

      boardgame-tictactoe-cell {
        border-right: 1px solid black;
      }

      boardgame-tictactoe-cell:last-of-type {
        border-right: 0;
      }

      .container {
        display: flex;
        flex-direction: row;
      }

      .board {
        display: flex;
        flex-direction: column;
      }
    `
  ];

  override render() {
    return html`
      <h2>Tictactoe</h2>
      <div class="container">
        <div class="board">
          <div class="row">
            <boardgame-tictactoe-cell .token="${this.state?.Game?.Slots?.Components?.[0]}" index="0"></boardgame-tictactoe-cell>
            <boardgame-tictactoe-cell .token="${this.state?.Game?.Slots?.Components?.[1]}" index="1"></boardgame-tictactoe-cell>
            <boardgame-tictactoe-cell .token="${this.state?.Game?.Slots?.Components?.[2]}" index="2"></boardgame-tictactoe-cell>
          </div>
          <div class="row">
            <boardgame-tictactoe-cell .token="${this.state?.Game?.Slots?.Components?.[3]}" index="3"></boardgame-tictactoe-cell>
            <boardgame-tictactoe-cell .token="${this.state?.Game?.Slots?.Components?.[4]}" index="4"></boardgame-tictactoe-cell>
            <boardgame-tictactoe-cell .token="${this.state?.Game?.Slots?.Components?.[5]}" index="5"></boardgame-tictactoe-cell>
          </div>
          <div class="row">
            <boardgame-tictactoe-cell .token="${this.state?.Game?.Slots?.Components?.[6]}" index="6"></boardgame-tictactoe-cell>
            <boardgame-tictactoe-cell .token="${this.state?.Game?.Slots?.Components?.[7]}" index="7"></boardgame-tictactoe-cell>
            <boardgame-tictactoe-cell .token="${this.state?.Game?.Slots?.Components?.[8]}" index="8"></boardgame-tictactoe-cell>
          </div>
        </div>
      </div>
      <boardgame-fading-text
        .trigger="${this.isCurrentPlayer}"
        message="Your Turn"
        suppress="falsey">
      </boardgame-fading-text>
    `;
  }
}

customElements.define('boardgame-render-game-tictactoe', BoardgameRenderGameTictactoe);
