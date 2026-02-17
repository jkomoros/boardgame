import '@material/web/button/filled-button.js';
import '../../../server/static/src/components/boardgame-die.js';
import { BoardgameBaseGameRenderer } from '../../../server/static/src/components/boardgame-base-game-renderer.js';
import '../../../server/static/src/components/boardgame-fading-text.js';
import { html, css } from 'lit';
import { MoveNames } from './_move_names.js';

class BoardgameRenderGamePig extends BoardgameBaseGameRenderer {
  static override styles = [
    ...(BoardgameBaseGameRenderer.styles ? [BoardgameBaseGameRenderer.styles] : []),
    css`
      .die {
        height: 100px;
        width: 100px;
        background-color: #ccc;
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
