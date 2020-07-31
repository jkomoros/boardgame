import { PolymerElement } from '@polymer/polymer/polymer-element.js';
import './boardgame-fading-text.js';
import { html } from '@polymer/polymer/lib/utils/html-tag.js';

class BoardgameStatusText extends PolymerElement {
  static get template() {
    return html`
    <style>

      :host {
        position: relative;
        display: inline-block;
      }

      .hidden {
        display:none;
      }

    </style>
    <strong>[[message]]</strong>
    <div class="hidden">
      <slot id="content"></slot>
    </div>
    <boardgame-fading-text trigger="[[message]]" auto-message="[[autoMessage]]" suppress="falsey"></boardgame-fading-text>
`;
  }

  static get is() {
    return "boardgame-status-text";
  }

  static get properties() {
    return {
      //The message to show. If autoMessage is set to something other than
      //'fixed' this will be set based on trigger.
      message: String,
      autoMessage: {
        type: String,
        value: "diff-up",
      },
      _textContentNode: Object,
      _observer: Object,
    }
  }

  ready() {
    super.ready();
    this.$.content.addEventListener("slotchanged", (e) => this._slotChanged(e));
    this._slotChanged();
  }

  _textContentChanged(rec) {

    let ele = rec[rec.length - 1].target;
    let message = ele.textContent ? ele.textContent : ele.innerText;
    if (!message) message = "";
    message = message.trim(); 
    this.message = message;
  }

  _slotChanged(e) {
    var nodes = this.$.content.assignedNodes();
    if (!nodes.length) return;
    for (var i = 0; i < nodes.length; i++) {
      let ele = nodes[i];
      let message = ele.textContent ? ele.textContent : ele.innerText;
      //This could happen if it's an empy text node for example.
      if (!message) message = "";
      message = message.trim(); 
      //We used to only register these if the message existed, but in many
      //real cases like the BUSTED line in blackjack it starts off as a
      //nil message.
      if (this._observer) {
        this._observer.disconnect();
        this._observer = null;
        this._node = null;
      }
      this._observer = new MutationObserver(rec => this._textContentChanged(rec));
      this._observer.observe(ele, {characterData: true});
      this.message = message;
      return;
    }
  }
}

customElements.define(BoardgameStatusText.is, BoardgameStatusText);
