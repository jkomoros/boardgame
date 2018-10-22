import '@polymer/polymer/polymer-element.js';
import '@polymer/paper-styles/typography.js';
import '@polymer/iron-flex-layout/iron-flex-layout.js';
import '@polymer/polymer/lib/elements/dom-bind.js';
import { BoardgameComponent } from './boardgame-component.js';
import { html } from '@polymer/polymer/lib/utils/html-tag.js';
let memoizedTemplate;

class BoardgameCard extends BoardgameComponent {
  static get templateContents() {
    return html`
    <style include="iron-flex">

      :root {
        /* These are copied and lightly modified from paper-styles/shadow, because we need rotated versions, too. */
        --shadow-elevation-normal-rotated: {
          box-shadow: 2px 0 2px 0 rgba(0, 0, 0, 0.14),
                      1px 0 5px 0 rgba(0, 0, 0, 0.12),
                      3px 0 1px -2px rgba(0, 0, 0, 0.2);
        };
        --shadow-elevation-raised-rotated: {
          box-shadow: 8px 0 10px 1px rgba(0, 0, 0, 0.14),
                      3px 0 14px 2px rgba(0, 0, 0, 0.12),
                      5px 0 5px -3px rgba(0, 0, 0, 0.4);
        };
        --alt-shadow-elevation-normal-rotated: {
          filter: drop-shadow(2px 0 2px rgba(0, 0, 0, 0.14)),
                  drop-shadow(1px 0 5px rgba(0, 0, 0, 0.12)),
                  drop-shadow(3px 0 1px rgba(0, 0, 0, 0.2));
        };
        --alt-shadow-elevation-raised-rotated: {
          filter: drop-shadow(8px 0 10px rgba(0, 0, 0, 0.14)),
                  drop-shadow(3px 0 14px rgba(0, 0, 0, 0.12)),
                  drop-shadow(5px 0 5px rgba(0, 0, 0, 0.4));
        };
      }

      #outer {
        /* note that boardgame-card-stack has a couple of hard-coded margins based on this value in its stylesheet */

        --default-component-width: 100px;
        --card-effective-border-radius: 5px;
      }

      #outer div.fallback {
        display:none;
      }

      #outer.no-content div.normal {
        display:none;
      }

      #outer.no-content div.fallback {
        display:block;
      }

      #front {
        @apply --layout-vertical;
        @apply --layout-center;
        @apply --layout-center-justified;
      }

      #outer {
        height: var(--component-effective-height);
        width: var(--component-effective-width);
        @apply --layout-vertical;
        @apply --layout-center;
        @apply --layout-center-justified;
        perspective: 1000px;
      }

      #outer.tall {
        height: var(--component-effective-width);
        width: var(--component-effective-height);
      }

      #outer.rotated {
        height: var(--component-effective-width);
        width: var(--component-effective-height);
      }

      #outer.tall.rotated {
        height: var(--component-effective-height);
        width: var(--component-effective-width);
      }

      #inner {

        width: var(--default-component-width);
        height: calc(var(--default-component-width) * var(--component-aspect-ratio));
        transform: scale(var(--component-effective-scale));

        border-radius: var(--card-effective-border-radius);

        transform-style: preserve-3d;
        position: absolute;

      }

      .tall #inner {
        height: var(--default-component-width);
        width: calc(var(--default-component-width) * var(--component-aspect-ratio)); 
      }

      #outer.shadow.rotated #inner {
        @apply --shadow-elevation-normal-rotated;
      }

      #outer.shadow.interactive.rotated:hover #inner {
        @apply --shadow-elevation-raised-rotated;
      }

      #outer.alt-shadow.rotated #inner {
        @apply --alt-shadow-elevation-normal-rotated;
      }

      #outer.alt-shadow.interactive.rotated:hover #inner {
        @apply --alt-shadow-elevation-raised-rotated;
      }

      #front, #back {
        height:100%;
        width:100%;
        position:absolute;
        top:0;
        left: 0;
        backface-visibility: hidden;
        -webkit-backface-visibility: hidden;

        overflow: hidden;
        border-radius:var(--card-effective-border-radius);
      }

      #top-rank, #bottom-rank {
        position:absolute;
        @apply --paper-font-caption;
      }

      #top-rank {
        bottom: 5px;
        left: 5px;
        transform: rotate(-90deg);
      }

      #bottom-rank {
        right: 5px;
        top: 5px;
        transform: rotate(90deg);
      }

      #outer #front {
        background-color: #CCFCFC;
        z-index: 2;
        transform: rotateY(180deg);
      }

      #outer #back {
        background-color: #00CCCC;
        transform: rotateY(0deg);
      }

      #default-back {
        height:100%;
        width:120%;
        opacity: 0.2;
        font-size: 13.5px;
        line-height: 14px;
        overflow: hidden;
        text-overflow:clip;
        user-select:none;
        @apply --layout-vertical;
        @apply --layout-center;
        @apply --layout-center-justified;
      }

      .tall #default-back {
        width:130%;
      } 

    </style>

    <div id="import">
      <div id="front">
        <div class="normal">
          <slot id="front-slot">
            <div id="top-rank">
              {{suit}}{{rank}}
            </div>
            <div id="bottom-rank">
              {{suit}}{{rank}}
            </div>
          </slot>
        </div>
        <div class="fallback">
          <slot name="fallback"></slot>
        </div>
      </div>
      <div id="back">
        <slot name="back">
          <div id="default-back">
                ★ ☆ ★  ☆ ★  ☆ ★  ☆ ★  ☆ ★  ☆ ★  ☆ ★  ☆ ★  ☆ ★  ☆ ★  ☆ ★  ☆ ★  ☆ ★  ☆ ★  ☆ ★  ☆ ★  ☆ ★  ☆ ★  ☆
          </div>
        </slot>
      </div>
    </div>
`;
  }

  static get is() {
    return "boardgame-card"
  }

  static get properties() {
    return {
      suit: String,
      rank: String,
      //If true, the card will be rendered with its face showing, not its
      //back. If noAnimate is false, will animate the card flip.
      faceUp: {
        type: Boolean,
        observer: "_updateInnerTransform"
      },
      //If true, the card is rotated 90 degrees.
      rotated: {
        type: Boolean,
        observer: "_rotatedChanged",
        value: false,
        reflectToAttribute: true
      },
      //basicRotated is similar to rotation, but it doesn't affect
      //containing layout. The inner tranform will be applied though.
      //Designed for being used in component-animator where we don't want
      //to affect layout in thie main loop. Only active when
      //overrideRotated is true. When in doubt, use rotated instead.
      basicRotated: {
        type: Boolean,
        observer: "_updateInnerTransform",
        value: false
      },
      //If true, rotated's value for transform will come from basicRotate,
      //not normal rotate. Designed to switch between rotated and
      //basicRotated, which is mainly necessary for during component-
      //animator.
      overrideRotated : {
        type: Boolean,
        obserer: "_updateInnerTransform",
        value: false
      },
      //If true, content with a slot of "fallback" will be rendered
      //instead of the normal front content. boardgame-component-animator
      //uses this functionality to inject old content that technically has
      //disappeared when a card flips, so that visually the content
      //doesn't disappear before the flip.
      noContent: {
        type: Boolean,
        value: false,
      },
      //tall will be true IFF the content for the front of the card has a
      //`tall` attribute. This reflects the *natural* orientation of the
      //card, and should be the same for all cards of a given type without
      //changing. If a card is tall, its size for layout will have its
      //major axis be the vertical axis.
      tall: {
        type: Boolean,
        value: false,
        readOnly: true,
      },
      //aspectRati is the ratio of width to height. It will be set based
      //on the asepct-ratio attribute on the card-inner.
      aspectRatio: {
        type: Number,
        value: 0.6666666,
        readOnly: true,
      },

      _outerStyle: {
        type: String,
        computed: "_computeOuterStyle(aspectRatio)"
      },
      _animating: {
        type: Boolean,
        value: false
      },
    }
  }

  static get observers() {
    return [
      //Update
      "_updateClasses(spacer, noShadow, interactive, disabled, noAnimate, altShadow, noContent, rotated, tall)"
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
    this.$['front-slot'].addEventListener("slotchange", e => this._frontChanged());
    this._frontChanged();
  }

  _computeOuterStyle(aspectRatio) {
    return "--component-aspect-ratio: " + aspectRatio + ";"
  }

  prepareForBeingAnimatingComponent(stack) {
    this.noContent = true;
    this.rotated = stack.stackDefault('rotated');
  }

  get animatingProperties() {
    return super.animatingProperties.concat(["rotated", "faceUp"]);
  }

  computeAnimationProps(isAfter, props) {

    //We override these props for performance.

    //All of these set inner rotation on card, so do them all at once

    if (isAfter) {
      return {
        faceUp: props.faceUp,
        overrideRotated: false,
        basicRotated: props.rotated
      }
    }

    return {
      faceUp: props.faceUp,
      overrideRotated: true,
      basicRotated: props.rotated
    }
  }

  get cloneContent() {
    return !this.noContent;
  }

  animationRotates(beforeProps, afterProps) {
    return beforeProps.rotated != afterProps.rotated;
  }

  _frontChanged() {
    var nodes = this.$["front-slot"].assignedNodes();
    var newValue = false;
    for (var i = 0; i < nodes.length; i++) {
      var node = nodes[i];
      if (node.nodeType != 1) continue;
      if (node.hasAttribute("tall")) {
        newValue = true;
      }
      if (node.hasAttribute("aspect-ratio")) {
        this._setAspectRatio(parseFloat(node.getAttribute("aspect-ratio")));
      }
    }
    this._setTall(newValue);
  }

  _rotatedChanged(newValue) {
    //there's a class of bugs where basicRotation isn't set the same as
    //rotation at beginning of rotation. The most recent one was when
    //moving a card that DIDN'T flip faceUp but did change from not
    //rotated to rotated, the first animation wouldn't work. To fix that,
    //we have basicRotated mirror rotated whenever rotated is explicitly
    //set, to verify basicRotated defaults to a reasonable value.
    this.basicRotated = newValue;
    this._updateInnerTransform();
  }

  _updateInnerTransform() {
    var transformPieces = ["scale(var(--component-effective-scale))"];
    //Chrome Canary used to interprolate fine if you left out the 0deg
    //rotation term, but then broke. Setting it explicitly fixes the bug.
    transformPieces.push(this.faceUp ? "rotateY(180deg)" : "rotateY(0deg)");
    transformPieces.push(((this.overrideRotated) ? this.basicRotated : this.rotated) ? "rotate(90deg)" : "rotate(0deg)");
    var transform = transformPieces.join(" ")
    if (!transform) {
      transform = "none";
    }
    this.$.inner.style.transform = transform;
    this._startingAnimation(this.$.inner, "transform");
  }

  _itemChanged(newValue) {
    if (newValue === undefined) return;
    if (newValue === null) {
      this.noContent = true;
      this.faceUp = false;
      super._itemChanged(newValue);
      return;
    }
    if (newValue.Values) {
      this.faceUp = true;
      this.noContent = false;
    } else {
      this.faceUp = false;
      this.noContent = true;
    }
    if (this._binder) this._binder.item = newValue;
    super._itemChanged(newValue);
  }

  //Override _computeClasses and add some more.
  _computeClasses() {
    let result = ["card"];
    if (this.rotated) {
      result.push("rotated");
    }
    if (this.noContent) {
      result.push("no-content")
    }
    result.push(this.tall ? "tall" : "wide");
    result.push(super._computeClasses());
    return result.join(" ");
  }
}

customElements.define(BoardgameCard.is, BoardgameCard);
