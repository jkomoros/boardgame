import { BoardgameAnimatableItem} from './boardgame-animatable-item.js';
import '@polymer/iron-flex-layout/iron-flex-layout.js';
import '@polymer/polymer/lib/elements/dom-repeat.js';
import '@polymer/paper-styles/shadow.js';
import { html } from '@polymer/polymer/lib/utils/html-tag.js';

class BoardgameDie extends BoardgameAnimatableItem {
  static get template() {
    return html`
    <style>

      :host {
        --effective-die-scale: var(--die-scale, 1.0);
        --effective-die-size: 50px;
        --pip-size: 7px; 
      }

      #scaler {
        height: calc(var(--effective-die-size) * var(--effective-die-scale));
        width: calc(var(--effective-die-size) * var(--effective-die-scale));
        position: relative;
        @apply --layout-vertical;
        @apply --layout-center;
        @apply --layout-center-justified;
      }

      #main.disabled {
        cursor:default;
      }

      #main {
        height: var(--effective-die-size);
        width: var(--effective-die-size);
        border-radius: 2px;
        background-color: #CCC;
        overflow: hidden;
        cursor: pointer;
        @apply --shadow-elevation-2dp;
        transform: scale(var(--effective-die-scale));
        /* The second part of this transition is from paper-styles/shadow, because we need to do our own to get rotated */
        transition: transform var(--animation-length) ease-in-out, box-shadow 0.28s cubic-bezier(0.4, 0, 0.2, 1);
      }

      #main.interactive:hover {
        @apply --shadow-elevation-8dp;
      }

      #inner {
        position:relative;
        transform: translateY(calc(-1 * var(--effective-die-size) * var(--selected-face)));
        transition: transform var(--animation-length) ease-in-out;
      }

      .face {
        @apply --layout-vertical;
        @apply --layout-center;
        @apply --layout-center-justified;
        height: var(--effective-die-size);
        width: var(--effective-die-size);
        @apply --paper-font-title;
        position: relative;
      }

      .pip {
        background-color: black;
        height: var(--pip-size);
        width: var(--pip-size);
        border-radius: calc(var(--pip-size) / 2);
        position: absolute;
        display: none;
      }

      .face.one span, .face.two span, .face.three span, .face.four span, .face.five span, .face.six span {
        display: none;
      }

      .pip.mid {
        top: calc(var(--effective-die-size) / 2 - var(--pip-size) / 2);
      }

      .pip.center {
        left: calc(var(--effective-die-size) / 2 - var(--pip-size) / 2);;
      }

      .pip.top {
        top: calc(var(--effective-die-size) / 2 - var(--pip-size) * 1.5 - var(--pip-size) / 2);
      }

      .pip.left {
        left: calc(var(--effective-die-size) / 2 - var(--pip-size) * 1.5 - var(--pip-size) / 2); 
      }

      .pip.bottom {
        top: calc(var(--effective-die-size) / 2 + var(--pip-size) * 1.5 - var(--pip-size) / 2);
      }

      .pip.right {
        left: calc(var(--effective-die-size) / 2 + var(--pip-size) * 1.5 - var(--pip-size) / 2); 
      }

      .face.one .pip.mid.center {
        display: block;
      }

      .face.two .pip.top.right, .face.two .pip.bottom.left {
        display: block;
      }

      .face.three .pip.top.right, .face.three .pip.mid.center, .face.three .pip.bottom.left {
        display: block;
      }

      .face.four .pip.top.right, .face.four .pip.top.left, .face.four .pip.bottom.left, .face.four .pip.bottom.right {
        display: block;
      }

      .face.five .pip.top.right, .face.five .pip.top.left, .face.five .pip.bottom.left, .face.five .pip.bottom.right, .face.five .pip.mid.center {
        display: block;
      }

      .face.six .pip.top.right, .face.six .pip.top.left, .face.six .pip.bottom.left, .face.six .pip.bottom.right, .face.six .pip.mid.left, .face.six .pip.mid.right {
        display: block;
      }


    </style>

    <div id="scaler">
      <div id="main" style\$="--selected-face:[[selectedFace]]" class\$="[[_classes(disabled)]]">
        <div id="inner">
          <template is="dom-repeat" items="[[faces]]">
            <div class\$="[[_classForFace(item)]]">
              <span>[[item]]</span>
              <div class="pip mid center"></div>
              <div class="pip top left"></div>
              <div class="pip top right"></div>
              <div class="pip bottom left"></div>
              <div class="pip bottom right"></div>
              <div class="pip mid left"></div>
              <div class="pip mid right"></div>
            </div>
          </template>
        </div>
      </div>
    </div>
`;
  }

  static get is() {
    return "boardgame-die";
  }

  static get properties() {
    return {
      item: {
        type: Object,
        observer: "_itemChanged"
      },
      value: Number,
      faces: Array,
      selectedFace: {
        type: Number,
        observer: "_selectedFaceChanged"
      },
      disabled: Boolean,
    }
  }

  ready() {
    super.ready();

    this.shadowRoot.addEventListener("tap", (e) => this._handleTap(e))
  }

  _handleTap(e) {
    if (this.disabled) {
      e.stopPropagation();
    }
  }

  _selectedFaceChanged(newValue) {
    this._expectTransitionEnd(this.$.inner, "transform");
  }

  _itemChanged(newValue) {
    if (!newValue) {
      this.setProperties({
        faces: [],
        selectedFace: 0,
        value: 0,
      })
      return;
    }
    this.setProperties({
      faces: newValue.Values.Faces,
      selectedFace: newValue.DynamicValues.SelectedFace,
      value: newValue.DynamicValues.Value,
    })
  }

  _classForFace(face) {
    var str = "";
    switch (face) {
      case 1:
        str = "one";
        break;
      case 2:
        str = "two";
        break;
      case 3:
        str = "three";
        break;
      case 4: 
        str = "four";
        break;
      case 5:
        str = "five";
        break;
      case 6:
        str = "six";
        break;
    }

    return "face " + str;
  }

  _classes(disabled) {
    var pieces = [];
    pieces.push(disabled ? "disabled" : "interactive")
    return pieces.join(" ");
  }
}

customElements.define(BoardgameDie.is, BoardgameDie);
