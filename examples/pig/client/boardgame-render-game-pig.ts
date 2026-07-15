import '@material/web/button/filled-button.js';
import { BoardgameBaseGameRenderer, html, css } from '../../src/client.js';
import { MoveNames } from './_move_names.js';
import type { MoveName } from './_move_names.js';
import { moveInputSchema as generatedMoveInputSchema, moveInputSchemaFingerprint as generatedMoveInputSchemaFingerprint, type MoveInputs } from './_move_args.js';
import type { GameState, PlayerState } from './_types.js';

class BoardgameRenderGamePig extends BoardgameBaseGameRenderer<GameState, PlayerState, MoveName, MoveInputs> {
	protected override readonly moveInputSchema = generatedMoveInputSchema;
	protected override readonly moveInputSchemaFingerprint = generatedMoveInputSchemaFingerprint;
  static override styles = [
    ...(BoardgameBaseGameRenderer.styles ? [BoardgameBaseGameRenderer.styles] : []),
    css`
      .die {
        height: 100px;
        width: 100px;
      }

      .container {
        display: flex;
        flex-direction: row;
      }

      .flex {
        flex: 1;
      }
    `
  ];

  override render() {
    return html`
      <div class="container">
        <boardgame-die
          propose-move="${MoveNames.RollDice}"
          .item="${this.state?.Game?.Die?.Components?.[0]}"
          ?disabled="${!this.isMoveCurrentlyLegal(MoveNames.RollDice)}">
        </boardgame-die>
        <div class="flex"></div>
        <md-filled-button
          propose-move="${MoveNames.DoneTurn}"
          ?disabled="${!this.isMoveCurrentlyLegal(MoveNames.DoneTurn)}">
          Done
        </md-filled-button>
      </div>
      <boardgame-fading-text
        .trigger="${this.isCurrentPlayer}"
        message="Your Turn"
        suppress="falsey">
      </boardgame-fading-text>
    `;
  }
}

customElements.define('boardgame-render-game-pig', BoardgameRenderGamePig);
