/* Making dynamic imports static for modulizer */
/* end faux static imports */
/*
  FIXME(polymer-modulizer): the above comments were extracted
  from HTML and may be out of place here. Review them and
  then delete this comment!
*/
import { Element } from '@polymer/polymer/polymer-element.js';

import './boardgame-component-animator.js';
import '@polymer/paper-spinner/paper-spinner-lite.js';
import '@polymer/iron-flex-layout/iron-flex-layout.js';
import '../game-src/blackjack/boardgame-render-game-blackjack.js';
import '../game-src/checkers/boardgame-render-game-checkers.js';
import '../game-src/debuganimations/boardgame-render-game-debuganimations.js';
import '../game-src/memory/boardgame-render-game-memory.js';
import '../game-src/pig/boardgame-render-game-pig.js';
import '../game-src/tictactoe/boardgame-render-game-tictactoe.js';
import { html } from '@polymer/polymer/lib/utils/html-tag.js';
import { importHref } from '@polymer/polymer/lib/utils/import-href.js';

class BoardgameRenderGame extends Element {
  static get template() {
    return html`
    <style>
      #container {
        position:relative;
      }

      #loading[active] {
        visibility: visible;
        opacity: 1;
        transition: visibility var(--animation-length) step-start, opacity var(--animation-length, 0.25s) linear;
      }

      #loading {
        position:absolute;
        top: 0;
        left: 0;
        height: 100%;
        width: 100%;
        background-color: rgba(255,255,255,0.7);
        z-index: 10;
        visibility: hidden;
        opacity: 0;
        transition: visibility var(--animation-length) step-end, opacity var(--animation-length, 0.25s) linear;
      }

      #loading > div {
        height:100%;
        width:100%;
        @apply --layout-vertical;
        @apply --layout-center;
        @apply --layout-center-justified;
      }

      paper-spinner-lite {
        height: 100px;
        width: 100px;
        --paper-spinner-stroke-width: 10px;
      }
    </style>
    <boardgame-component-animator id="animator" ancestor-offset-parent="{{\$.container}}"></boardgame-component-animator>
    <div hidden\$="{{rendererLoaded}}">
      <h2>Diagram of {{gameName}}</h2>
      <pre>{{diagram}}</pre>
    </div>
    <div id="container">
    <!-- This is where renderer will go -->
    </div>
    <div id="loading" active\$="{{!socketActive}}">
      <div>
        <paper-spinner-lite active="{{!socketActive}}"></paper-spinner-lite>
      </div>
    </div>
`;
  }

  static get is() {
    return "boardgame-render-game"
  }

  static get properties() {
    return {
      state: {
        type: Object,
      },
      chest: {
        type: Object,
        observer: "_chestChanged",
      },
      active: {
        type: Boolean,
        observer: "_activeChanged",
      },
      diagram : {
        type: String,
        observer: "_diagramChanged",
      },
      gameName: {
        type: String,
        observer: "_gameNameChanged",
      },
      renderer: Object,
      rendererLoaded: {
        type: Boolean,
        value: false,
      },
      viewingAsPlayer: {
        type: Number,
        observer: "_viewingAsPlayerChanged",
      },
      currentPlayerIndex: {
        type: Number,
        observer: "_currentPlayerIndexChanged"
      },
      socketActive: {
        type: Boolean,
        value: false,
      }
    }
  }

  static get observers() {
    return [
      "_stateChanged(state.*)"
    ]
  }

  _diagramChanged(newValue) {
    if (!this.renderer) {
      return;
    }
    this.renderer.diagram = newValue;
  }

  _activeChanged(newValue) {
    if (!newValue) {

      //The game view has gone inactive.

      //Clear out state now so by the time we switch back it will be null
      //and we minimize chance of trying to render state with the wrong
      //renderer.

      //We don't throw out the renderer here because if we come back to a
      //game of the same type we should keep it around.
      this.state = null;
      this.diagram = "";
      this.viewingAsPlayer = 0;
      this.currentPlayerIndex = 0;
    }
  }

  _stateChanged(record) {
    if (!this.renderer) return;
    var stateWasNull = (this.renderer.state == null);
    if (record.path == "state" && !stateWasNull) {
      this.$.animator.prepare();
    }
    this.renderer.set(record.path, record.value);
    //This shiouldn't be necessary... set should have already done
    //notifyPath. Bug in Polymer 2?
    this.renderer.notifyPath(record.path);
    if (record.path == "state" && !stateWasNull) {
      this.$.animator.animate();
    }
  }

  _viewingAsPlayerChanged(newValue) {
    if (!this.renderer) return;
    this.renderer.viewingAsPlayer = newValue;
  }

  _currentPlayerIndexChanged(newValue) {
    if (!this.renderer) return;
    this.renderer.currentPlayerIndex = newValue;
  }

  _chestChanged(newValue) {
    if (!this.renderer) return;
    this.renderer.chest = newValue;
  }

  _gameNameChanged(newValue) {

    //If there was a state, it might for a different game type which would
    //cause a render error.
    this.state = null;

    this.rendererLoaded = false

    if (this.renderer) {
      this.$.container.removeChild(this.renderer);
    }
    this.renderer = null;


    var resolvedUrl = this.resolveUrl("../game-src/" +newValue + "/boardgame-render-game-" + newValue + ".html")
    importHref(resolvedUrl, () => this._instantiateRenderer(), null, true);
  }

  _instantiateRenderer(e) {
    //The import loaded! Add it!
    this.rendererLoaded = true;

    var ele = document.createElement("boardgame-render-game-" + this.gameName);


    ele.diagram = this.diagram;
    ele.state = this.state;
    ele.viewingAsPlayer = this.viewingAsPlayer;
    ele.currentPlayerIndex = this.currentPlayerIndex;
    ele.chest = this.chest;

    this.renderer = ele;

    this.$.container.appendChild(ele);

  }
}

customElements.define(BoardgameRenderGame.is, BoardgameRenderGame)
