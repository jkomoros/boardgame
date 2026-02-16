import '@material/web/button/filled-button.js';
import '@material/web/button/outlined-button.js';
import '../../../server/static/src/components/boardgame-component-stack.js';
import '../../../server/static/src/components/boardgame-card.js';
import { BoardgameBaseGameRenderer } from '../../../server/static/src/components/boardgame-base-game-renderer.js';
import '../../../server/static/src/components/boardgame-fading-text.js';
import '../../../server/static/src/components/boardgame-deck-defaults.js';
import { html, css } from 'lit';
import { repeat } from 'lit/directives/repeat.js';

class BoardgameRenderGameBlackjack extends BoardgameBaseGameRenderer {
  static override styles = [
    ...(BoardgameBaseGameRenderer.styles ? [BoardgameBaseGameRenderer.styles] : []),
    css`
      #draw, #players {
        display: flex;
        flex-direction: row;
        align-items: center;
      }

      .flex {
        flex: 1;
      }

      .player {
        display: flex;
        flex-direction: column;
      }

      .busted {
        filter: saturate(0.5) blur(1px);
      }
    `
  ];

  private _bustedClass(busted: boolean): string {
    return busted ? 'busted' : '';
  }

  override render() {
    return html`
      <boardgame-deck-defaults>
        <template deck="cards">
          <boardgame-card suit="{{item.Values.Suit}}" rank="{{item.Values.Rank}}"></boardgame-card>
        </template>
      </boardgame-deck-defaults>
      <div id="draw">
        <boardgame-component-stack
          .stack="${this.state?.Game?.DrawStack}"
          layout="stack"
          messy
          .componentAttrs=${{ proposeMove: 'Current Player Hit' }}>
        </boardgame-component-stack>
        <div class="flex">
          <md-filled-button propose-move="Current Player Hit" ?disabled="${!this.isCurrentPlayer}">Hit</md-filled-button>
          <md-outlined-button propose-move="Current Player Stand" ?disabled="${!this.isCurrentPlayer}">Stand</md-outlined-button>
        </div>
        <boardgame-component-stack
          .stack="${this.state?.Game?.DiscardStack}"
          layout="stack"
          messy>
        </boardgame-component-stack>
      </div>
      <div id="players">
        ${repeat(this.state?.Players || [], (player, index) => index, (player, index) => html`
          <div class="player flex ${this._bustedClass(player.Busted)}">
            <strong>Player ${index}</strong>
            <boardgame-component-stack
              .stack="${player.Hand}"
              layout="fan"
              messy
              .componentAttrs=${{ rotated: true }}>
              <boardgame-fading-text .trigger="${player.Busted}" message="Busted!"></boardgame-fading-text>
              <boardgame-fading-text .trigger="${player.Stood}" message="Stand!"></boardgame-fading-text>
            </boardgame-component-stack>
          </div>
        `)}
      </div>
      <boardgame-fading-text
        .trigger="${this.isCurrentPlayer}"
        message="Your Turn"
        suppress="falsey">
      </boardgame-fading-text>
    `;
  }
}

customElements.define('boardgame-render-game-blackjack', BoardgameRenderGameBlackjack);
