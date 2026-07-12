import { BoardgameBaseGameRenderer } from '../../src/components/boardgame-base-game-renderer.js';
import '../../src/components/boardgame-token.js';
import '../../src/components/boardgame-game-board.js';
import '../../src/components/boardgame-fading-text.js';
import { html, css } from 'lit';
import { MoveNames } from './_move_names.js';
import type { MoveName } from './_move_names.js';
import type { GameState, PlayerState } from './_types.js';
import type { MovePreviewSpec } from '../../src/legal/previewLegality.js';

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

  // Opt into per-cell legality preview: ask the server which of the 9 slots can
  // legally receive a token right now (occupied ones, and every slot when it's
  // not our turn, come back illegal). boardgame-render-game batches these and
  // feeds the illegal spaces back via previewDisabledSpaces, which the board
  // below binds to .disabledSpaces (graying + blocking them).
  override previewSpec(): MovePreviewSpec | null {
    const slots = this.state?.Game?.Slots?.Components;
    if (!slots) return null;
    return {
      moveName: MoveNames.PlaceToken,
      candidates: slots.map((_, index) => ({ space: index, args: { Slot: String(index) } })),
    };
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
        .disabledSpaces="${this.previewDisabledSpaces}"
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
