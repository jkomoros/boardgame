import '@polymer/polymer/polymer-element.js';
import { BoardgameComponent } from './boardgame-component.js';
import { html } from '@polymer/polymer/lib/utils/html-tag.js';

let memoizedTemplate;

class BoardgameToken extends BoardgameComponent {
  static get templateContents() {
    return html`
    <style>

    #inner {
      height: var(--component-effective-height);
      width: var(--component-effective-width);
    }

    #inner img {
      height: 100%;
      width: 100%;
    }

    #outer.pawn {
      --component-aspect-ratio:2.0;
    }

    #outer.meeple {
      --component-aspect-ratio:1.25;
    }

    #outer.active #inner, #outer.highlighted #inner {
      animation-name: throb;
      animation-duration: 1s;
      animation-timing-function: ease-in-out;
      animation-direction: alternate;
      animation-iteration-count: infinite;
    }

    #outer.active #inner {

      --throb-color-from: rgba(136,136,38,1.0);
      --throb-color-to: rgba(136,136,38,0.5);
    }

    #outer.highlighted #inner{
      --throb-color-from: rgba(0,0,0,1.0);
      --throb-color-to: rgba(0,0,0,0.5);
    }

    #outer.active.highlighted #inner {
      --throb-color-from: rgba(255,255,0,1.0);
      --throb-color-to: rgba(255,255,0,0.0);
    }

    @keyframes throb {
      from {
        filter: drop-shadow(0 0 0.25em var(--throb-color-to)) drop-shadow(0 0 0.25em var(--throb-color-to));
      }
      to {
         /* double the effect so it's darker */
        filter: drop-shadow(0 0 0.25em var(--throb-color-from)) drop-shadow(0 0 0.25em var(--throb-color-from));
      }
    }

    #outer.gray img {
      filter: saturate(0.0) brightness(3.0);
    }

    #outer.green img {
      filter: hue-rotate(130deg) brightness(2.0);
    }

    #outer.teal img {
      filter: hue-rotate(185deg) brightness(2.4);
    }

    #outer.purple img {
      filter: hue-rotate(300deg) brightness(1.0);
    }

    #outer.pink img {
      filter: hue-rotate(-93deg) brightness(4) saturate(0.8);
    }

    /* red is the default color, no need for shifting */

    #outer.blue img {
      filter: hue-rotate(220deg) brightness(2.0) saturate(1.5);
    }

    #outer.orange img {
      filter: hue-rotate(50deg) brightness(2.5);
    }

    #outer.yellow img {
      filter: hue-rotate(70deg) brightness(4);
    }

    #outer.black img {
      filter: saturate(0.0) brightness(1.7);
    }

    </style>
    <div id="import">
      <img src="[[_asset]]">
    </div>
`;
  }

  static get properties() {
    return {
      //Color to set. One of the colors returned by legalColors.
      color: {
        type: String,
        value: "red",
      },
      //Active changes the styling to make it clear the thing is selcted
      active: Boolean,
      //highlighted has a different visual style than active. different
      //games will use it for different things.
      highlighted: Boolean,
      //The type of token. Supported values: "token" (default), "chip",
      //"cube", "pawn", "meeple"
      type: {
        type: String,
        value: "token",
      },
      _asset: {
        type: String,
        computed: "_computeAsset(type)"
      }
    }
  }

  static get is() {
    return "boardgame-token"
  }

  get legalTypes() {
    return [
      "token",
      "chip",
      "cube",
      "pawn",
      "meeple",
    ]
  }

  get legalColors() {
    return [
      "gray",
      "green",
      "teal",
      "purple",
      "pink",
      "red",
      "blue",
      "yellow",
      "orange",
      "black",
    ]
  }

  static get template() {
    if (!memoizedTemplate) {
      memoizedTemplate = BoardgameComponent.combinedTemplate(this.templateContents);
    }
    return memoizedTemplate;
  }

  ready() {
    super.ready();
    this.altShadow = true;
  }

  static get observers() {
    return [
      //Update
      "_updateClasses(spacer, noShadow, interactive, disabled, noAnimate, altShadow, color, highlighted, active, type)"
    ]
  }

  _computeAsset(type) {
    return "src/assets/token_" + type + ".svg";  
  }

  //Override _computeClasses and add some more.
  _computeClasses() {
    let result = [this.color];
    if (this.active) {
      result.push("active");
    }
    if (this.highlighted) {
      result.push("highlighted");
    }
    result.push(this.type);
    result.push(super._computeClasses());
    return result.join(" ");
  }
}

customElements.define(BoardgameToken.is, BoardgameToken);
