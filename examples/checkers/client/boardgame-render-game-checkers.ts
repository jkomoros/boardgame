import { BoardgameBaseGameRenderer } from '../../src/components/boardgame-base-game-renderer.js';
import '../../src/components/boardgame-token.js';
import '../../src/components/boardgame-game-board.js';
import '../../src/components/boardgame-fading-text.js';
import { html, css } from 'lit';
import { property } from 'lit/decorators.js';
import { MoveNames } from './_move_names.js';
import type { MoveName } from './_move_names.js';
import type { GameState, PlayerState } from './_types.js';

class BoardgameRenderGameCheckers extends BoardgameBaseGameRenderer<GameState, PlayerState, MoveName> {
  static override styles = [
    ...(BoardgameBaseGameRenderer.styles ? [BoardgameBaseGameRenderer.styles] : []),
    css`
      boardgame-game-board {
        max-width: 500px;
        margin: 0 auto;
      }
    `
  ];

  @property({ type: Number, attribute: false })
  private selectedSpace = -1;

  // Reset selection when game state changes (e.g., opponent moved)
  protected override updated(changedProperties: Map<string, unknown>): void {
    super.updated(changedProperties);
    if (changedProperties.has('state') && changedProperties.get('state') !== undefined) {
      this.selectedSpace = -1;
    }
  }

  private _onSpaceTapped(e: CustomEvent) {
    const { index } = e.detail;
    const component = this.state?.Game?.Spaces?.Components?.[index];

    if (this.selectedSpace < 0) {
      // No piece selected yet — select this one if occupied
      if (component) {
        this.selectedSpace = index;
      }
    } else {
      // A piece is selected — try to move to this cell
      if (index !== this.selectedSpace) {
        this.proposeMove(MoveNames.MoveToken, {
          TokenIndexToMove: this.selectedSpace,
          SpaceIndex: index,
        });
      }
      this.selectedSpace = -1;
    }
  }

  override render() {
    return html`
      <boardgame-deck-defaults>
        <template deck="tokens">
          <boardgame-token type="chip" color="{{item.Values.Color}}"></boardgame-token>
        </template>
      </boardgame-deck-defaults>
      <boardgame-game-board
        rows="8" cols="8" checkerboard
        .stack="${this.state?.Game?.Spaces}"
        .selectedSpace="${this.selectedSpace}"
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

customElements.define('boardgame-render-game-checkers', BoardgameRenderGameCheckers);
