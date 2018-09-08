import { BoardgameBaseGameRenderer } from '../../src/boardgame-base-game-renderer.js';
import { html } from '@polymer/polymer/lib/utils/html-tag.js';
import '@polymer/polymer/lib/elements/dom-repeat.js';
import '@polymer/iron-flex-layout/iron-flex-layout.js';
import '../../src/boardgame-component-stack.js';
import '../../src/boardgame-card.js';
import '../../src/boardgame-deck-defaults.js';
import '../../src/boardgame-fading-text.js';

class BoardgameRenderGameCheckers extends BoardgameBaseGameRenderer {

  static get template() {
  	
  	return html`<style>
      #players {
        @apply --layout-horizontal;
        @apply --layout-center;
      }
      .flex {
        @apply --layout-flex;
      }
      .player {
        @apply --layout-vertical;
      }
    </style>
    <boardgame-deck-defaults>
      <template deck="examplecards">
        <boardgame-card rank="{{item.Values.Value}}"></boardgame-card>
      </template>
    </boardgame-deck-defaults>
    <boardgame-component-stack stack="{{state.Game.DrawStack}}" layout="stack" messy  component-propose-move="Draw Card"></boardgame-component-stack>
    <div id="players">
      <template is="dom-repeat" items="{{state.Players}}">
      	<div class="player flex">
		    <strong>Player {{index}}</strong>
		    <boardgame-component-stack stack="{{item.Hand}}" layout="fan" messy component-rotated>
		    	<boardgame-fading-text trigger="{{item.Computed.GameScore}}" auto-message="diff-up"></boardgame-fading-text>
		    </boardgame-component-stack>
	    </div>
      </template>
    </div>
    <boardgame-fading-text trigger="{{isCurrentPlayer}}" message="Your Turn" suppress="falsey"></boardgame-fading-text>
`;

  }

  static get is() {
    return "boardgame-render-game-checkers"
  }

  //We don't need to compute any properties that BoardgameBaseGamErenderer
  //doesn't have.

}

customElements.define(BoardgameRenderGameCheckers.is, BoardgameRenderGameCheckers);

