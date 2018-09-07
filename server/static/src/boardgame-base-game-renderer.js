import { Element } from '@polymer/polymer/polymer-element.js';
import { html } from '@polymer/polymer/lib/utils/html-tag.js';

class BoardgameBaseGameRenderer extends Element {
  static get template() {
    return html`

`;
  }

  static get is() {
    return "boardgame-base-game-renderer"
  }

  static get properties() {
    return {
      state: Object,
      chest: Object,
      diagram : String,
      viewingAsPlayer: Number,
      currentPlayerIndex: Number,
      isCurrentPlayer: {
        type: Boolean,
        computed: "_computeIsCurrentPlayer(currentPlayerIndex, viewingAsPlayer)"
      }
    }
  }

  ready() {
    super.ready();
    this.addEventListener("tap", e => this._handleButtonTapped(e));
    this.addEventListener("component-tapped", e => this._handleButtonTapped(e));
  }

  _computeIsCurrentPlayer(currentPlayerIndex, viewingAsPlayer) {
    if (viewingAsPlayer == -2) return true;
    return currentPlayerIndex == viewingAsPlayer;
  }

  _handleButtonTapped(e) {
    var composedPath = e.composedPath();
    var ele = null;
    for (var i = 0; i < composedPath.length; i++) {
      var tempEle = composedPath[i];
      //Skip things like shadow roots
      if (!tempEle.getAttribute) continue;
      if (tempEle.proposeMove || tempEle.getAttribute("propose-move")) {
        //found it!
        ele = tempEle;
        break;
      }
    }
    if (!ele) {
      return;
    }
    if (ele.hasAttribute("boardgame-component") && e.type == "tap") {
      //Cards we'll fire on the component-tapped, not the tap.
      return;
    }
    var moveName = ele.proposeMove || ele.getAttribute("propose-move");
    if (!moveName) return;
    var data = ele.dataset;
    var args = {};
    for (var key in data) {
      if (!data.hasOwnProperty(key)) continue;
      if (!key.startsWith("arg")) continue;
      var effectiveKey = key.replace("arg", "");
      //Handle the case where the attribute was literally just data-arg
      if (!effectiveKey) continue;
      //The first character is now upperCase, which is desired as per Move field convention
      args[effectiveKey] = data[key];
    }
    this.dispatchEvent(new CustomEvent("propose-move", {composed: true, detail: {name: moveName, arguments:args}}));
  }
}

customElements.define(BoardgameBaseGameRenderer.is, BoardgameBaseGameRenderer);
