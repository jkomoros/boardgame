import { PolymerElement } from '@polymer/polymer/polymer-element.js';
import '@polymer/iron-flex-layout/iron-flex-layout.js';
import { html } from '@polymer/polymer/lib/utils/html-tag.js';

class BoardgameFadingText extends PolymerElement {
  static get template() {
    return html`
    <style>

      #container {
        position: absolute;
        top: 0;
        left: 0;
        height: 100%;
        width: 100%;
        @apply --layout-vertical;
        @apply --layout-center;
        @apply --layout-center-justified;
        visibility: hidden;
        pointer-events: none;
      }

      #container.animating {
        visibility: visible;
      }

      #message {
        font-size: var(--message-font-size, 16px);
      }

      .animating #message {
        animation-name: fadetext;
        animation-duration: var(--animation-length, 0.25s);
        animation-timing-function: ease-out;
      }

      @keyframes fadetext {
        from {
          opacity: 1.0;
          transform: scale(1.0);
        }
        to {
          opacity: 0.0;
          transform: scale(6.0);
        }
      }

    </style>

    <div id="container" class\$="[[_classes(_animating)]]">
      <div id="message">
        [[message]]
      </div>
    </div>
`;
  }

  static get is() {
    return "boardgame-fading-text";
  }

  static get properties() {
    return {
      //The message to show. If autoMessage is set to something other than
      //'fixed' this will be set based on trigger.
      message: {
        type: String,
        value: "Point Scored",
      },
      //When Trigger changes, the element will animate.
      trigger: {
        type: Object,
        observer: "_triggerChanged",
      },
      //Policy of when to suppress animation triggering. 'none' (default):
      //trigger always, 'falsey' : suppress triggering when value is
      //false-y, 'truthy' : suppress when value is truthy.
      suppress: {
        type: String,
        value: "none"
      },
      //Values: 'fixed' - no change, 'new' - autoamtically set to
      //newvalue, 'diff' - if both new and old are numbers, set to the
      //difference, and 'diff-up' which is like 'diff' but if the
      //difference is a number and it's less than 0 don't animate.
      autoMessage: {
        type: String,
        value: "fixed",
      },
      _animating: Boolean,
    }
  }

  ready() {
    super.ready()
    this.$.message.addEventListener("animationend", () => this._animationEnded());
  }

  _animationEnded() {
    this._animating = false;
  }

  animate() {
    this._animating = true;
  }

  _triggerChanged(newValue, oldValue) {
    if (oldValue === undefined) return;

    //If people use us directly newValue and oldValue might be a number...
    //but for example boardgame-status-text will pass us strings.
    var newValueAsNumber = parseInt(newValue);
    var oldValueAsNumber = parseInt(oldValue);

    switch (this.autoMessage) {
      case "diff":
      case "diff-up":
        if (!isNaN(newValueAsNumber) && !isNaN(oldValueAsNumber)) {
          let diff = newValueAsNumber - oldValueAsNumber;
          if (this.autoMessage == "diff-up" && diff < 0) {
            //Skip animating
            return;
          }
          this.message = (diff > 0) ? "+" + diff : diff;
        } else {
          this.message = newValue;
        }
        break;
      case "new":
        this.message = newValue;
        break;
    }

    switch(this.suppress) {
      case "falsey":
        if (!newValue) return;
        break;
      case "truthy":
        if (newValue) return;
        break;
    }

    this.animate();
  }

  _classes(_animating) {
    let classes = [];
    if (_animating) {
      classes.push("animating");
    }
    return classes.join(" ");
  }
}

customElements.define(BoardgameFadingText.is, BoardgameFadingText);
