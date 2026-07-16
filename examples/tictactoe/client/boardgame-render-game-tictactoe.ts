import { GameRenderer } from './_game_renderer.js';
import '../../src/components/boardgame-token.js';
import '../../src/components/boardgame-game-board.js';
import '../../src/components/boardgame-fading-text.js';
import { html, css } from 'lit';
import { MoveNames } from './_move_names.js';
import type { GameState } from './_types.js';
import { tokenView } from '../../src/client.js';

class BoardgameRenderGameTictactoe extends GameRenderer {
  private readonly tokens = tokenView<GameState['Slots']>({
    properties: () => ({ type: 'chip' }),
  });

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
      <boardgame-game-surface heading="Tic-tac-toe">
        <boardgame-game-board
          rows="3" cols="3"
          .stack=${slots ?? null}
          .componentView=${this.tokens}
          .action=${places}>
        </boardgame-game-board>
        <boardgame-fading-text
          slot="status"
          .trigger="${this.isCurrentPlayer}"
          message="Your Turn"
          suppress="falsey">
        </boardgame-fading-text>
      </boardgame-game-surface>
    `;
  }
}

customElements.define('boardgame-render-game-tictactoe', BoardgameRenderGameTictactoe);
