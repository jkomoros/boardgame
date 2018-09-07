import { Element } from '@polymer/polymer/polymer-element.js';
import '@polymer/polymer/lib/elements/dom-repeat.js';
import '@polymer/iron-flex-layout/iron-flex-layout-classes.js';
import { html } from '@polymer/polymer/lib/utils/html-tag.js';

class BoardgameBoard extends Element {
  static get template() {
    return html`
    <style include="iron-flex">
      .cell {
        height:50px;
        width:50px;
        position:relative;
        cursor:pointer;
      }

      .cell.even {
        background-color: black;
      }

      .container {
        position:relative;
      }

      ::slotted([boardgame-component]) {
        position:absolute;
        top:0;
        left:0;
      }
    </style>
    <div class="container">
      <template is="dom-repeat" items="{{_cellItems}}" as="row" index-as="r">
        <div class="row layout horizontal">
          <template is="dom-repeat" items="{{row}}" as="col" index-as="c">
            <div class\$="{{col}} cell layout vertical" r="{{r}}" c="{{c}}" on-tap="_regionTapped">
            </div>
          </template>
        </div>
      </template>
      <slot></slot>
    </div>
`;
  }

  static get is() {
    return "boardgame-board"
  }

  static get properties() {
    return {
      rows: Number,
      cols: Number,
      _cellItems: {
        type: Array,
        computed: "_computeCellItems(rows, cols)"
      }
    }
  }

  _computeCellItems(rows, cols) {
    let isOdd = false;
    let result = [];
    for (let r = 0; r < rows; r++) {
      let row = [];
      for (let c = 0; c < cols; c++) {
        row.push(isOdd ? "odd" : "even");
        isOdd = !isOdd;
      }
      //each row should alternate which way it starts
      isOdd = !isOdd;
      result.push(row);
    }
    return result;
  }

  _regionTapped(e) {
    let r = e.target.r;
    let c = e.target.c;
    this.dispatchEvent(new CustomEvent("region-tapped", {composed: true, detail: {index: [r, c]}}));
  }
}

customElements.define(BoardgameBoard.is, BoardgameBoard);
