import { html, css } from '../../src/client.js';
import { GameRenderer, registerGameRenderer } from './_game_renderer.js';
import { MoveNames } from './_move_names.js';

@registerGameRenderer
export class BoardgameRenderGamePig extends GameRenderer {
  static override styles = [
    ...(GameRenderer.styles ? [GameRenderer.styles] : []),
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
      <boardgame-game-surface heading="Pig">
        <div class="container">
          <boardgame-die
            .item="${this.state?.Game?.Die?.Components?.[0]}"
            .action="${this.move(MoveNames.RollDice)}">
          </boardgame-die>
          <div class="flex"></div>
          <boardgame-action-button .action="${this.move(MoveNames.DoneTurn)}">
            Done
          </boardgame-action-button>
        </div>
        <boardgame-turn-status
          slot="status"
          .turn=${this.turnStatus}>
        </boardgame-turn-status>
      </boardgame-game-surface>
    `;
  }
}
