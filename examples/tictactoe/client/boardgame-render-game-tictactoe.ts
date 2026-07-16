import { GameRenderer } from './_game_renderer.js';
import '../../src/components/boardgame-token.js';
import '../../src/components/boardgame-game-board.js';
import '../../src/components/boardgame-fading-text.js';
import { html, css } from 'lit';
import { MoveNames } from './_move_names.js';

class BoardgameRenderGameTictactoe extends GameRenderer {
  static override styles = [
    ...(GameRenderer.styles ? [GameRenderer.styles] : []),
    css`
      boardgame-game-board {
        max-width: 320px;
        margin: 0 auto;
      }
    `
  ];

  override render() {
    const slots = this.state?.Game?.Slots;
    const places = slots
      ? this.move(MoveNames.PlaceToken).targets(
        slots.Components.map((_, slot) => slot),
        slot => ({ Slot: slot }),
      )
      : null;
    return html`
      <h2>Tictactoe</h2>
      <boardgame-deck-defaults>
        <template deck="tokens">
          <boardgame-token type="chip"></boardgame-token>
        </template>
      </boardgame-deck-defaults>
      <boardgame-game-board
        rows="3" cols="3"
        .stack=${slots}
        .action=${places}>
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
