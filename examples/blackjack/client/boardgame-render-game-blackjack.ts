import '../../src/components/boardgame-component-stack.js';
import '../../src/components/boardgame-card.js';
import { GameRenderer } from './_game_renderer.js';
import '../../src/components/boardgame-fading-text.js';
import '../../src/components/boardgame-deck-defaults.js';
import { html, css } from 'lit';
import { repeat } from 'lit/directives/repeat.js';
import { MoveNames } from './_move_names.js';

class BoardgameRenderGameBlackjack extends GameRenderer {
  static override styles = [
    ...(GameRenderer.styles ? [GameRenderer.styles] : []),
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
          .componentAttrs=${{ disabled: true }}>
        </boardgame-component-stack>
        <div class="flex">
          <boardgame-action-button .action=${this.move(MoveNames.CurrentPlayerHit)}>Hit</boardgame-action-button>
          <boardgame-action-button .action=${this.move(MoveNames.CurrentPlayerStand)}>Stand</boardgame-action-button>
        </div>
        <boardgame-component-stack
          .stack="${this.state?.Game?.DiscardStack}"
          layout="stack"
          messy>
        </boardgame-component-stack>
      </div>
      <div id="players">
        ${repeat(this.state?.Players || [], (_player, index) => index, (player, index) => html`
          <div class="player flex ${this._bustedClass(player.Eliminated)}">
            <strong>Player ${index}</strong>
            <boardgame-component-stack
              .stack="${player.Hand}"
              layout="fan"
              messy
              .componentAttrs=${{ rotated: true }}>
              <boardgame-fading-text .trigger="${player.Eliminated}" message="Busted!"></boardgame-fading-text>
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
