import { PolymerElement } from '@polymer/polymer/polymer-element.js';
import '@polymer/iron-flex-layout/iron-flex-layout-classes.js';
import '@polymer/paper-styles/typography.js';
import '@polymer/paper-styles/color.js';
import { html } from '@polymer/polymer/lib/utils/html-tag.js';

class BoardgameTictactoeCell extends PolymerElement {
  static get template() {
    return html`
    <style include="iron-flex">
      :host {
        height:100px;
        width:100px;
        cursor:pointer;
        @apply(--paper-font-display2);
      }
      .cell {
        height: 100%;
        width: 100%;
      }
      .cell>div {
        text-align:center;
      }
    </style>
    <div class="cell layout vertical center center-justified" propose-move="Place Token" data-arg-slot\$="{{index}}">
      {{value}}
    </div>
`;
  }

  static get is() {
    return "boardgame-tictactoe-cell"
  }

  static get properties() {
    return {
      token: {
        type: Object,
        observer: "_tokenChanged",
      },
      index: Number,
      value: String,
    }
  }

  _tokenChanged(newValue) {
    if (!newValue) {
      this.value = "";
      return
    }
    this.value = newValue.Values.Value;
  }

  handleTap() {
    this.dispatchEvent(new CustomEvent("propose-move", {composed: true, detail: {name: "Place Token", arguments: {
      "Slot": this.index,
    }}}));
  }
}

customElements.define(BoardgameTictactoeCell.is, BoardgameTictactoeCell);
