import '@polymer/polymer/polymer-element.js';
import '@polymer/polymer/lib/elements/dom-repeat.js';
import { BoardgameBaseGameRenderer } from '../../src/components/boardgame-base-game-renderer.js';
import '../../src/components/boardgame-board.js';
import '../../src/components/boardgame-token.js';
import { html } from '@polymer/polymer/lib/utils/html-tag.js';

class BoardgameRenderGameCheckers extends BoardgameBaseGameRenderer {
  static get template() {
    return html`
    <style>
      boardgame-token {
        --component-scale:1.25;
      }
    </style>

    <boardgame-board rows="{{size}}" cols="{{size}}">
      <dom-repeat items="{{_components}}">
        <!-- note: these don't get the styling from board's slotted rule because they aren't direct children -->
        <boardgame-token color="red"></boardgame-token>
      </dom-repeat>
      <dom-repeat items="{{_components}}">
        <boardgame-token color="black"></boardgame-token>
      </dom-repeat>
    </boardgame-board>
`;
  }

  static get is() {
    return "boardgame-render-game-checkers"
  }

  static get properties() {
    return {
      size: {
        type: Number,
        value: 8,
      },
      _components: {
        computed: "_computeComponents(size)"
      }
    }
  }

  _computeComponents(size) {
    let result = [];
    for (let i = 0; i < size; i++){
      result.push(true);
    }
    return result;
  }
}

customElements.define(BoardgameRenderGameCheckers.is, BoardgameRenderGameCheckers);
