import { BoardgameBaseGameRenderer } from '../../../server/static/src/components/boardgame-base-game-renderer.js';
import '../../../server/static/src/components/boardgame-token.js';
import '../../../server/static/src/components/boardgame-game-board.js';
import '../../../server/static/src/components/boardgame-fading-text.js';
import { html, css } from 'lit';
import { MoveNames } from './_move_names.js';
import type { MoveName } from './_move_names.js';
import type { GameState, PlayerState } from './_types.js';

class BoardgameRenderGameTictactoe extends BoardgameBaseGameRenderer<GameState, PlayerState, MoveName> {
  static override styles = [
    ...(BoardgameBaseGameRenderer.styles ? [BoardgameBaseGameRenderer.styles] : []),
    css`
      boardgame-game-board {
        max-width: 320px;
        margin: 0 auto;
      }
    `
  ];

  private _onSpaceTapped(e: CustomEvent) {
    const { index } = e.detail;
    this.proposeMove(MoveNames.PlaceToken, { Slot: index });
  }

  override render() {
    return html`
      <h2>Tictactoe</h2>
      <boardgame-deck-defaults>
        <template deck="tokens">
          <boardgame-token type="chip"></boardgame-token>
        </template>
      </boardgame-deck-defaults>
      <boardgame-game-board
        rows="3" cols="3"
        .stack="${this.state?.Game?.Slots}"
        @space-tapped="${this._onSpaceTapped}">
      </boardgame-game-board>
      <boardgame-fading-text
        .trigger="${this.isCurrentPlayer}"
        message="Your Turn"
        suppress="falsey">
      </boardgame-fading-text>
    `;
  }
}

customElements.define('boardgame-render-game-tictactoe', BoardgameRenderGameTictactoe);
