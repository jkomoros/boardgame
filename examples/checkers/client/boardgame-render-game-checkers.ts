import { GameRenderer } from './_game_renderer.js';
import '../../src/components/boardgame-token.js';
import '../../src/components/boardgame-game-board.js';
import '../../src/components/boardgame-fading-text.js';
import { html, css } from 'lit';
import { property } from 'lit/decorators.js';
import { MoveNames } from './_move_names.js';
import type { MovePreviewSpec } from '../../src/legal/previewLegality.js';

class BoardgameRenderGameCheckers extends GameRenderer {
  static override styles = [
    ...(GameRenderer.styles ? [GameRenderer.styles] : []),
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

    // The selected source drives previewSpec(); the host can't see this local
    // state, so ask it to re-run the destination preview.
    this.requestPreviewRefresh();
  }

  // Once a source piece is selected, gray the EMPTY squares it cannot legally
  // move to (batch-checked on the server, which is the only place mayMoveToSlot
  // + the diagonal/capture rules live). Occupied squares are deliberately NOT
  // candidates, so they stay tappable for re-selecting a different piece; legal
  // empty destinations stay enabled so the second tap completes the move. With
  // no source selected there's nothing to preview.
  override previewSpec(): MovePreviewSpec | null {
    if (this.selectedSpace < 0) return null;
    const spaces = this.state?.Game?.Spaces?.Components;
    if (!spaces) return null;
    const candidates: { space: number; args: Record<string, string> }[] = [];
    for (let dest = 0; dest < spaces.length; dest++) {
      if (spaces[dest]) continue; // occupied — not a destination; keep tappable
      candidates.push({
        space: dest,
        args: { TokenIndexToMove: String(this.selectedSpace), SpaceIndex: String(dest) },
      });
    }
    return { moveName: MoveNames.MoveToken, candidates };
  }

  override render() {
    return html`
      <boardgame-deck-defaults>
        <template deck="tokens">
          <boardgame-token type="disc" color="{{item.Values.Color}}"></boardgame-token>
        </template>
      </boardgame-deck-defaults>
      <boardgame-game-board
        rows="8" cols="8" checkerboard
        .stack="${this.state?.Game?.Spaces}"
        .selectedSpace="${this.selectedSpace}"
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

customElements.define('boardgame-render-game-checkers', BoardgameRenderGameCheckers);
