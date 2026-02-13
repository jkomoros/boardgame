import '@material/web/button/filled-button.js';
import '@material/web/progress/linear-progress.js';
import { BoardgameBaseGameRenderer } from '../../src/components/boardgame-base-game-renderer.js';
import '../../src/components/boardgame-card.js';
import '../../src/components/boardgame-component-stack.js';
import '../../src/components/boardgame-fading-text.js';
import '../../src/components/boardgame-deck-defaults.js';
import { html, css } from 'lit';

class BoardgameRenderGameMemory extends BoardgameBaseGameRenderer {
  static override styles = [
    ...(BoardgameBaseGameRenderer.styles ? [BoardgameBaseGameRenderer.styles] : []),
    css`
      md-linear-progress {
        width: 100%;
      }

      .current {
        font-weight: bold;
      }

      boardgame-card > div {
        font-family: 'Roboto', 'Noto', sans-serif;
        font-size: 34px;
        font-weight: 400;
        letter-spacing: -.01em;
        line-height: 40px;
      }

      .discards {
        --component-scale: 0.7;
        display: flex;
        flex-direction: row;
        justify-content: space-around;
      }
    `
  ];

  get maxTimeLeft(): number {
    return this.computeMaxTimeLeft(this.state?.Game?.HideCardsTimer?.originalTimeLeft);
  }

  override delayAnimation(fromMove: any, toMove: any): number {
    if (toMove && toMove.Name === 'Capture Cards') {
      // Show the cards for a second before capturing them.
      return 1000;
    }
    return 0;
  }

  private computeMaxTimeLeft(timeLeft: number): number {
    return Math.max(timeLeft, 100);
  }

  override render() {
    return html`
      <boardgame-deck-defaults>
        <template deck="cards">
          <boardgame-card>
            <div>
              {{item.Values.Type}}
            </div>
          </boardgame-card>
        </template>
      </boardgame-deck-defaults>
      <h2>Memory</h2>
      <div>
        <boardgame-component-stack
          layout="grid"
          messy
          .stack="${this.state?.Game?.Cards}"
          component-propose-move="Reveal Card"
          component-index-attributes="data-arg-card-index">
        </boardgame-component-stack>
        <boardgame-fading-text
          message="Match"
          .trigger="${this.state?.Game?.Cards?.NumComponents}">
        </boardgame-fading-text>
      </div>
      <div class="discards">
        <boardgame-component-stack
          layout="stack"
          .stack="${this.state?.Players?.[0]?.WonCards}"
          messy
          component-disabled>
        </boardgame-component-stack>
        <!-- have a boardgame-card spacer just to keep that row height sane even with no cards -->
        <boardgame-card spacer></boardgame-card>
        <boardgame-component-stack
          layout="stack"
          messy
          .stack="${this.state?.Players?.[1]?.WonCards}"
          component-disabled>
        </boardgame-component-stack>
      </div>
      <md-filled-button
        id="hide"
        propose-move="Hide Cards"
        ?disabled="${this.state?.Computed?.Global?.CurrentPlayerHasCardsToReveal}">
        Hide Cards
      </md-filled-button>
      <md-linear-progress
        id="timeleft"
        .value="${(this.state?.Game?.HideCardsTimer?.TimeLeft || 0) / (this.maxTimeLeft || 1)}"
        .max="${1}">
      </md-linear-progress>
      <boardgame-fading-text
        .trigger="${this.isCurrentPlayer}"
        message="Your Turn"
        suppress="falsey">
      </boardgame-fading-text>
    `;
  }
}

customElements.define('boardgame-render-game-memory', BoardgameRenderGameMemory);
