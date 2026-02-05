import '@polymer/paper-button/paper-button.js';
import '@polymer/paper-progress/paper-progress.js';
import { BoardgameBaseGameRenderer } from '../../../server/static/src/components/boardgame-base-game-renderer.js';
import '../../../server/static/src/components/boardgame-card.js';
import '../../../server/static/src/components/boardgame-component-stack.js';
import '../../../server/static/src/components/boardgame-fading-text.js';
import '../../../server/static/src/components/boardgame-deck-defaults.js';
import { html, css } from 'lit';

class BoardgameRenderGameMemory extends BoardgameBaseGameRenderer {
  static override styles = [
    ...(BoardgameBaseGameRenderer.styles ? [BoardgameBaseGameRenderer.styles] : []),
    css`
      paper-progress {
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
      <paper-button
        id="hide"
        propose-move="Hide Cards"
        raised
        ?disabled="${this.state?.Computed?.Global?.CurrentPlayerHasCardsToReveal}">
        Hide Cards
      </paper-button>
      <paper-progress
        id="timeleft"
        value="${this.state?.Game?.HideCardsTimer?.TimeLeft}"
        max="${this.maxTimeLeft}">
      </paper-progress>
      <boardgame-fading-text
        .trigger="${this.isCurrentPlayer}"
        message="Your Turn"
        suppress="falsey">
      </boardgame-fading-text>
    `;
  }
}

customElements.define('boardgame-render-game-memory', BoardgameRenderGameMemory);
