import '@polymer/polymer/lib/elements/dom-repeat.js';
import '@polymer/iron-flex-layout/iron-flex-layout.js';
import '@polymer/paper-button/paper-button.js';
import '../../src/boardgame-component-stack.js';
import '../../src/boardgame-card.js';
import '../../src/boardgame-base-game-renderer.js';
import '../../src/boardgame-fading-text.js';
import '../../src/boardgame-deck-defaults.js';
import { html } from '@polymer/polymer/lib/utils/html-tag.js';

class BoardgameRenderGameBlackjack extends BoardgameBaseGameRenderer {
  static get template() {
    return html`
    <style>
      #draw, #players {
        @apply --layout-horizontal;
        @apply --layout-center;
      }
      .flex {
        @apply --layout-flex;
      }
      .player {
        @apply --layout-vertical;
      }

      .busted {
        filter: saturate(0.5) blur(1px);
      }
    </style>
    <boardgame-deck-defaults>
      <template deck="cards">
        <boardgame-card suit="{{item.Values.Suit}}" rank="{{item.Values.Rank}}"></boardgame-card>
      </template>
    </boardgame-deck-defaults>
    <div id="draw">
      <boardgame-component-stack stack="{{state.Game.DrawStack}}" layout="stack" messy="" component-propose-move="Current Player Hit">
      </boardgame-component-stack>
      <div class="flex">
        <paper-button raised="" propose-move="Current Player Hit" disabled="{{!isCurrentPlayer}}">Hit</paper-button>
        <paper-button raised="" propose-move="Current Player Stand" disabled="{{!isCurrentPlayer}}">Stand</paper-button>
      </div>
      <boardgame-component-stack stack="{{state.Game.DiscardStack}}" layout="stack" messy="">
      </boardgame-component-stack>
    </div>
    <div id="players">
      <template is="dom-repeat" items="{{state.Players}}">
        <div class\$="player flex {{_bustedClass(item.Busted)}}">
          <strong>Player {{index}}</strong>
          <boardgame-component-stack stack="{{item.Hand}}" layout="fan" messy="" component-rotated="">
            <boardgame-fading-text trigger="{{item.Busted}}" message="Busted!"></boardgame-fading-text>
            <boardgame-fading-text trigger="{{item.Stood}}" message="Stand!"></boardgame-fading-text>
          </boardgame-component-stack>
        </div>
      </template>
    </div>
    <boardgame-fading-text trigger="{{isCurrentPlayer}}" message="Your Turn" suppress="falsey"></boardgame-fading-text>
`;
  }

  static get is() {
    return "boardgame-render-game-blackjack"
  }

  //We don't need to compute any properties that BoardgameBaseGamErenderer
  //doesn't have.

  _bustedClass(busted) {
    return (busted) ? "busted" : ""
  }
}

customElements.define(BoardgameRenderGameBlackjack.is, BoardgameRenderGameBlackjack);
