import '@polymer/paper-button/paper-button.js';
import '../../../server/static/src/components/boardgame-die.js';
import { BoardgameBaseGameRenderer } from '../../../server/static/src/components/boardgame-base-game-renderer.js';
import '../../../server/static/src/components/boardgame-fading-text.js';
import { html, css } from 'lit';

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
          propose-move="Roll Dice"
          .item="${this.state?.Game?.Die?.Components?.[0]}"
          ?disabled="${!this.isCurrentPlayer}">
        </boardgame-die>
        <div class="flex"></div>
        <paper-button
          propose-move="Done Turn"
          ?disabled="${!this.isCurrentPlayer}"
          raised>
          Done
        </paper-button>
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
