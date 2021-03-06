/* Making dynamic imports static for modulizer */
/* end faux static imports */
/*
  FIXME(polymer-modulizer): the above comments were extracted
  from HTML and may be out of place here. Review them and
  then delete this comment!
*/
import { PolymerElement } from '@polymer/polymer/polymer-element.js';

import { html } from '@polymer/polymer/lib/utils/html-tag.js';

class BoardgameRenderPlayerInfo extends PolymerElement {
  static get template() {
    return html`
    <div id="container">
    <!-- This is where renderer will go -->
    </div>
`;
  }

  static get is() {
    return "boardgame-render-player-info"
  }

  static get properties() {
    return {
      state: {
        type: Object,
      },
      active: {
        type: Boolean,
        observer: "_activeChanged"
      },
      gameName: {
        type: String,
        observer: "_gameNameChanged"
      },
      renderer: Object,
      rendererGameName: String,
      rendererLoaded: {
        type: Boolean,
        value: false,
        observer: "_rendererLoadedChanged",
      },
      chipText: {
        type: String,
        notify: true,
      },
      chipColor: {
        type: String,
        notify: true,
      },
      playerIndex: Number,
      playerState: {
        type: Object,
        computed: "_computePlayerState(state, playerIndex)",
        observer: "_playerStateChanged",
      }
    }
  }

  static get observers() {
    return [
      "_stateChanged(state.*)"
    ]
  }

  ready() {
    super.ready();

    if (this.instantiateWhenReady) {
      this.instantiateRenderer();
    }
  }

  _activeChanged(newValue) {
    if (!newValue) {
      if (!this.renderer) return;
      this.renderer.parentElement.removeChild(this.renderer);
      this.renderer = null;
    }
  }

  _chipTextChanged(e) {
    this.chipText = e.detail.value;
  }

  _chipColorChanged(e) {
    this.chipColor = e.detail.value;
  }

  _gameNameChanged(newValue) {
    if (!newValue) return;
    if (newValue != this.rendererGameName) {
      this.resetRenderer();
    }
    if (!this.instantiateWhenGameNameSet) return;
    this.instantiateRenderer();
  }

  _rendererLoadedChanged(newValue) {
    if (!newValue) return;
    if (!this.renderer) {
      this.instantiateRenderer();
    }
  }

  _stateChanged(record) {

    if (!this.renderer) return;

    if (record.path == "state" && !record.value) {
      //skip it
      return;
    }

    this.renderer.set(record.path, record.value);
    //This shiouldn't be necessary... set should have already done
    //notifyPath. Bug in Polymer 2?
    this.renderer.notifyPath(record.path);

  }

  _computePlayerState(state, playerIndex) {
    if (!state) return;
    return state.Players[playerIndex];
  }

  _playerStateChanged(newValue) {
    if (!this.renderer) return;
    this.renderer.playerState = newValue;
  }

  resetRenderer() {
    if (!this.$ || !this.$.container) return;
    if (this.renderer) {
      var container = this.$.container;
      container.removeChild(this.renderer);
    }
    this.renderer = null;
    this.rendererGameName = "";
  }

  instantiateRenderer() {

    if (!this.rendererLoaded) return;
    if (!this.$) {
      this.instantiateWhenReady = true;
      return
    }

    if (!this.gameName) {
      this.instantiateWhenGameNameSet = true;
      return;
    }

    var ele = document.createElement("boardgame-render-player-info-" + this.gameName);


    ele.state = this.state;
    ele.playerIndex = this.playerIndex;
    ele.playerState = this.playerState;

    this.chipText = ele.chipText || "";
    this.chipColor = ele.chipColor || "";

    this.renderer = ele;
    this.rendererGameName = this.gameName;

    //I believe in Polymer 2 (or in native shadow dom), the change events
    //don't have composed set to true, so we have to listen for them
    //directly on the renderer.
    this.renderer.addEventListener("chip-text-changed", e => this._chipTextChanged(e));
    this.renderer.addEventListener("chip-color-changed", e => this._chipColorChanged(e));


    this.$.container.appendChild(ele);

  }
}

customElements.define(BoardgameRenderPlayerInfo.is, BoardgameRenderPlayerInfo);
