import '@polymer/polymer/polymer-element.js';
import '@polymer/iron-flex-layout/iron-flex-layout-classes.js';
import { BoardgameBaseGameRenderer } from '../../src/boardgame-base-game-renderer.js';
import './boardgame-tictactoe-cell.js';
import { html } from '@polymer/polymer/lib/utils/html-tag.js';
class BoardgameRenderGameTictactoe extends BoardgameBaseGameRenderer {
  static get template() {
    return html`
    <style include="iron-flex">
     .row {
      border-bottom:1px solid black;
     }
     .row:last-of-type {
      border-bottom:0;
     }

     boardgame-tictactoe-cell {
      border-right: 1px solid black;
     }

     boardgame-tictactoe-cell:last-of-type {
      border-right:0;
     }
    </style>

    <h2>Tictactoe</h2>
    <div class="horizontal layout">
      <div class="board vertical layout">
        <div class="row layout horizontal">
          <boardgame-tictactoe-cell token="{{state.Game.Slots.Components.0}}" index="0"></boardgame-tictactoe-cell>
          <boardgame-tictactoe-cell token="{{state.Game.Slots.Components.1}}" index="1"></boardgame-tictactoe-cell>
          <boardgame-tictactoe-cell token="{{state.Game.Slots.Components.2}}" index="2"></boardgame-tictactoe-cell>
        </div>
        <div class="row layout horizontal">
          <boardgame-tictactoe-cell token="{{state.Game.Slots.Components.3}}" index="3"></boardgame-tictactoe-cell>
          <boardgame-tictactoe-cell token="{{state.Game.Slots.Components.4}}" index="4"></boardgame-tictactoe-cell>
          <boardgame-tictactoe-cell token="{{state.Game.Slots.Components.5}}" index="5"></boardgame-tictactoe-cell>
        </div>
        <div class="row layout horizontal">
          <boardgame-tictactoe-cell token="{{state.Game.Slots.Components.6}}" index="6"></boardgame-tictactoe-cell>
          <boardgame-tictactoe-cell token="{{state.Game.Slots.Components.7}}" index="7"></boardgame-tictactoe-cell>
          <boardgame-tictactoe-cell token="{{state.Game.Slots.Components.8}}" index="8"></boardgame-tictactoe-cell>
        </div>
      </div>
    </div>
    <boardgame-fading-text trigger="{{isCurrentPlayer}}" message="Your Turn" suppress="falsey"></boardgame-fading-text>
`;
  }

  static get is() {
    return "boardgame-render-game-tictactoe"
  }

  //We don't need to define any properties that BoardgameBaseGameRenderer
  //doesn't already have.
}

customElements.define(BoardgameRenderGameTictactoe.is, BoardgameRenderGameTictactoe);
