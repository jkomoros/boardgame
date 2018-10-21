import '@polymer/polymer/polymer-element.js';
import '@polymer/polymer/lib/elements/dom-repeat.js';
import '@polymer/iron-flex-layout/iron-flex-layout-classes.js';
import '@polymer/paper-button/paper-button.js';
import '@polymer/paper-progress/paper-progress.js';
import '@polymer/paper-styles/typography.js';
import { BoardgameBaseGameRenderer } from '../../src/boardgame-base-game-renderer.js';
import '../../src/boardgame-card.js';
import '../../src/boardgame-component-stack.js';
import '../../src/boardgame-fading-text.js';
import '../../src/boardgame-deck-defaults.js';
import { html } from '@polymer/polymer/lib/utils/html-tag.js';

class BoardgameRenderGameMemory extends BoardgameBaseGameRenderer {
  static get template() {
    return html`
    <style include="iron-flex iron-flex-alignment">
      paper-progress {
        width:100%;
      }
      .current {
        font-weight:bold;
      }
      boardgame-card>div{
        @apply(--paper-font-display2);
      }
      .discards {
        --component-scale: 0.7;
      }
    </style>
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
      <boardgame-component-stack layout="grid" messy="" stack="{{state.Game.Cards}}" component-propose-move="Reveal Card" component-index-attributes="data-arg-card-index">
      </boardgame-component-stack>
       <boardgame-fading-text message="Match" trigger="{{state.Game.Cards.NumComponents}}"></boardgame-fading-text>
    </div>
    <div class="layout horizontal around-justified discards">
      <boardgame-component-stack layout="stack" stack="{{state.Players.0.WonCards}}" messy="" component-disabled="">
      </boardgame-component-stack>
      <!-- have a boardgame-card spacer just to keep that row height sane even with no cards -->
      <boardgame-card spacer=""></boardgame-card>
      <boardgame-component-stack layout="stack" messy="" stack="{{state.Players.1.WonCards}}" component-disabled="">
      </boardgame-component-stack>
    </div>
    <paper-button id="hide" propose-move="Hide Cards" raised="" disabled="{{state.Computed.Global.CurrentPlayerHasCardsToReveal}}">Hide Cards</paper-button>
    <paper-progress id="timeleft" value="{{state.Game.HideCardsTimer.TimeLeft}}" max="{{maxTimeLeft}}"></paper-progress>
    <boardgame-fading-text trigger="{{isCurrentPlayer}}" message="Your Turn" suppress="falsey"></boardgame-fading-text>
`;
  }

  static get is() {
    return "boardgame-render-game-memory"
  }

  static get properties() {
    return {
      maxTimeLeft: {
        type: Number,
        computed: 'computeMaxTimeLeft(state.Game.HideCardsTimer.originalTimeLeft)'
      }
    }
  }

  delayAnimation(fromMove, toMove) {
    if (toMove && toMove.Name == "Capture Cards") {
      //Show the cards for a second before capturing them.
      return 1000;
    }
    return 0;
  }

  computeMaxTimeLeft(timeLeft) {
    return Math.max(timeLeft, 100);
  }
}

customElements.define(BoardgameRenderGameMemory.is, BoardgameRenderGameMemory);
