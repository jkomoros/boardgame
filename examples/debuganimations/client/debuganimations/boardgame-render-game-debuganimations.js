import '@polymer/polymer/polymer-element.js';
import '@polymer/polymer/lib/elements/dom-repeat.js';
import '@polymer/iron-flex-layout/iron-flex-layout.js';
import '@polymer/paper-button/paper-button.js';
import '@polymer/paper-dropdown-menu/paper-dropdown-menu.js';
import '@polymer/paper-listbox/paper-listbox.js';
import '@polymer/paper-item/paper-item.js';
import '@polymer/paper-toggle-button/paper-toggle-button.js';
import '@polymer/paper-slider/paper-slider.js';
import '../../src/boardgame-deck-defaults.js';
import '../../src/boardgame-base-game-renderer.js';
import '../../src/boardgame-card.js';
import '../../src/boardgame-component-stack.js';
import '../../src/boardgame-fading-text.js';
import '../../src/boardgame-token.js';
import { html } from '@polymer/polymer/lib/utils/html-tag.js';

class BoardgameRenderGameDebuganimations extends BoardgameBaseGameRenderer {
  static get template() {
    return html`
    <style>

      .slow {
        --animation-length: 5s;
      }

      #shortstacks {
        @apply --layout-horizontal;
      }

      #draw {
        @apply --layout-horizontal;
      }

      #shortstacks boardgame-card > div {
        @apply --layout-horizontal;
        @apply --layout-center;
        @apply --layout-center-justified;
      }

      #fan {
        @apply --layout-horizontal;
      }

      #fan boardgame-component-stack:first-child {
        --component-scale:1.2;
      }

      .flex {
        @apply --layout-flex;
      }

      .controls {
        @apply --layout-vertical;
      }

      #hidden {
        @apply --layout-horizontal;
      }

      #controls {
        @apply --layout-horizontal;
      }

      #token {
        @apply --layout-horizontal;
      }

      #tokens {
        @apply --layout-horizontal;
      }

      #tokens-sanitized {
        @apply --layout-horizontal;
      }

    </style>
    <boardgame-deck-defaults>
      <template deck="cards">
        <boardgame-card>
          <div tall="">
            {{item.Values.Type}}
          </div>
        </boardgame-card>
      </template>
      <template deck="tokens">
        <boardgame-token></boardgame-token>
      </template>
    </boardgame-deck-defaults>
    <div id="container" class\$="{{_classes(slowAnimations)}}">
      <div id="controls">
        <paper-toggle-button checked="{{fromStackRotated}}">From Rotated</paper-toggle-button>
        <paper-toggle-button checked="{{toStackRotated}}">To Rotated</paper-toggle-button>
        <paper-toggle-button checked="{{messy}}">Messy</paper-toggle-button>
        <paper-toggle-button checked="{{slowAnimations}}">Slow Animation</paper-toggle-button>
        From scale:<paper-slider min="0.5" max="2.0" value="{{fromCardScale}}" pin="" step=".05"></paper-slider>
        To Scale: <paper-slider min="0.5" max="2.0" value="{{toCardScale}}" pin="" step=".05"></paper-slider>
      </div>
      <div id="shortstacks">
        <boardgame-component-stack layout="stack" stack="{{state.Game.FirstShortStack}}" messy="{{messy}}" component-propose-move="Move Card Between Short Stacks">
        </boardgame-component-stack>
        <boardgame-component-stack layout="stack" messy="{{messy}}" stack="{{state.Game.SecondShortStack}}" component-propose-move="Move Card Between Short Stacks">
        </boardgame-component-stack>
        <paper-button propose-move="Move Card Between Short Stacks">Swap</paper-button>
      </div>

      <div id="draw">
        <boardgame-component-stack layout="stack" messy="{{messy}}" stack="{{state.Game.DrawStack}}" component-rotated="{{messy}}" component-index-attributes="my-index,other-index">
        </boardgame-component-stack>
        <boardgame-component-stack layout="stack" messy="{{messy}}" stack="{{state.Game.DiscardStack}}">
        </boardgame-component-stack>
        <paper-button propose-move="Move Card Between Draw and Discard Stacks">Draw</paper-button>
        <boardgame-fading-text trigger="{{state.Game.DrawStack.Components.length}}" auto-message="diff"></boardgame-fading-text>
      </div>

      <div id="draw">
        <boardgame-component-stack layout="stack" messy="{{messy}}" stack="{{state.Game.Card}}"> 
        </boardgame-component-stack>
        <paper-button propose-move="Flip Card Between Hidden and Revealed">Flip</paper-button>
      </div>

      <div id="fan">
        <boardgame-component-stack layout="{{fromStackLayout}}" messy="{{messy}}" stack="{{state.Game.FanStack}}" style\$="--component-scale:{{fromCardScale}}" component-rotated="{{fromStackRotated}}">
        </boardgame-component-stack>
        <div class="flex"></div>
        <boardgame-component-stack layout="stack" messy="{{messy}}" stack="{{state.Game.FanDiscard}}" style\$="--component-scale:{{toCardScale}}" component-rotated="{{toStackRotated}}">
        </boardgame-component-stack>
        <div class="controls">
          <paper-button propose-move="Move Fan Card">Draw</paper-button>
          <paper-button propose-move="Visible Shuffle">Public Shuffle</paper-button>
          <paper-button propose-move="Shuffle">Shuffle</paper-button>
          <paper-dropdown-menu label="Layout">
            <paper-listbox slot="dropdown-content" selected="{{fromStackLayout}}" attr-for-selected="value">
              <paper-item value="fan">fan</paper-item>
              <paper-item value="spread">spread</paper-item>
              <paper-item value="stack">stack</paper-item>
              <paper-item value="grid">grid</paper-item>
              <paper-item value="pile">pile</paper-item>
            </paper-listbox>
          </paper-dropdown-menu>
        </div>
      </div>

      <div id="hidden">
        <boardgame-component-stack layout="fan" messy="{{messy}}" stack="{{state.Game.VisibleStack}}" style\$="--component-scale:{{fromCardScale}}" component-rotated="{{fromStackRotated}}">
        </boardgame-component-stack>
        <boardgame-component-stack layout="stack" messy="{{messy}}" stack="{{state.Game.HiddenStack}}" style\$="--component-scale:{{toCardScale}}" faux-components="5" component-rotated="{{toStackRotated}}">
        </boardgame-component-stack>
        <paper-button propose-move="Move Between Hidden">Draw</paper-button>
      </div>

      <div id="token">
        <boardgame-token color="{{tokenColor}}" highlighted="{{tokenHighlighted}}" active="{{tokenActive}}" type="{{tokenType}}"></boardgame-token>
        <div class="flex"></div>
        <paper-toggle-button checked="{{tokenHighlighted}}">Token Highlighted</paper-toggle-button>
        <paper-toggle-button checked="{{tokenActive}}">Token Active</paper-toggle-button>
        <paper-dropdown-menu label="Type">
          <paper-listbox slot="dropdown-content" selected="{{tokenType}}" attr-for-selected="value">
            <template is="dom-repeat" items="{{legalTokenTypes}}">
              <paper-item value="{{item}}">{{item}}</paper-item>
            </template>
          </paper-listbox>
        </paper-dropdown-menu>
        <paper-dropdown-menu label="Color">
          <paper-listbox slot="dropdown-content" selected="{{tokenColor}}" attr-for-selected="value">
            <template is="dom-repeat" items="{{legalTokenColors}}">
              <paper-item value="{{item}}">{{item}}</paper-item>
            </template>
          </paper-listbox>
        </paper-dropdown-menu>
      </div>

      <div id="tokens">
        <boardgame-component-stack layout="grid" messy="{{messy}}" stack="{{state.Game.TokensFrom}}" component-color="{{tokenColor}}" component-type="{{tokenType}}">
        </boardgame-component-stack>
        <boardgame-component-stack layout="grid" messy="{{messy}}" stack="{{state.Game.TokensTo}}" component-color="{{tokenColor}}" component-type="{{tokenType}}">
        </boardgame-component-stack>
        <paper-button propose-move="Move Token">Swap</paper-button>
      </div>


      <div id="tokens-sanitized">
        <boardgame-component-stack layout="pile" messy="{{messy}}" stack="{{state.Game.SanitizedTokensFrom}}" component-color="{{tokenColor}}" component-type="{{tokenType}}">
        </boardgame-component-stack>
        <boardgame-component-stack layout="pile" messy="{{messy}}" stack="{{state.Game.SanitizedTokensTo}}" faux-components="5" component-color="{{tokenColor}}" component-type="{{tokenType}}">
        </boardgame-component-stack>
        <paper-button propose-move="Move Token Sanitized">Swap</paper-button>
      </div>

    </div>
`;
  }

  static get is() {
    return "boardgame-render-game-debuganimations"
  }

  static get properties() {
    return {
      fromStackLayout: {
        type: String,
        value: "fan"
      },
      fromStackRotated: {
        type: Boolean,
        value: false
      },
      toStackRotated: {
        type: Boolean,
        value: false
      },
      messy: {
        type: Boolean,
        value: true
      },
      slowAnimations: {
        type: Boolean,
        value: false,
      },
      fromCardScale: {
        type: Number,
        value: 1.0
      },
      toCardScale: {
        type: Number,
        value: 1.0
      },
      tokenActive: {
        type: Boolean,
        value: false,
      },
      tokenHighlighted: {
        type: Boolean,
        value: false,
      },
      tokenType: {
        type: String,
        value: "cube",
      },
      tokenColor: {
        type: String,
        value: "red",
      },
      legalTokenTypes: {
        type: Array,
      },
      legalTokenColors: {
        type: Array,
      }
    }
  }

  _classes() {
    if (this.slowAnimations) {
      return "slow"
    }
    return "";
  }

  ready() {
    super.ready();
    let token = this.shadowRoot.querySelector("boardgame-token");
    this.legalTokenTypes = token.legalTypes;
    this.legalTokenColors = token.legalColors;
  }
}

customElements.define(BoardgameRenderGameDebuganimations.is, BoardgameRenderGameDebuganimations);
