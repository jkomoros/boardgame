import '@polymer/polymer/polymer-element.js';
import '@polymer/polymer/lib/elements/dom-repeat.js';
import '@polymer/iron-flex-layout/iron-flex-layout.js';
import '@polymer/paper-button/paper-button.js';
import '../../src/boardgame-die.js';
import { BoardgameBaseGameRenderer } from '../../src/boardgame-base-game-renderer.js';
import '../../src/boardgame-fading-text.js';
import { html } from '@polymer/polymer/lib/utils/html-tag.js';

class BoardgameRenderGamePig extends BoardgameBaseGameRenderer {
  static get template() {
    return html`
    <style include="iron-flex">

      .die {
        height: 100px;
        width: 100px;
        background-color: #ccc;
      }

    </style>

    <div class="horizontal layout">
      <boardgame-die propose-move="Roll Dice" item="{{state.Game.Die.Components.0}}" disabled="{{!isCurrentPlayer}}"></boardgame-die>
      <div class="flex"></div>
      <paper-button propose-move="Done Turn" disabled="{{!isCurrentPlayer}}" raised="">Done</paper-button>
    </div>
    <boardgame-fading-text trigger="{{isCurrentPlayer}}" message="Your Turn" suppress="falsey"></boardgame-fading-text>
`;
  }

  static get is() {
    return "boardgame-render-game-pig"
  }

  //We don't define our own properties, so the properties on
  //BoardgameBaseGameRenderer are fine.
}

customElements.define(BoardgameRenderGamePig.is, BoardgameRenderGamePig);
