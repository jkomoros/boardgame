import '@polymer/paper-button/paper-button.js';
import '@polymer/paper-dropdown-menu/paper-dropdown-menu.js';
import '@polymer/paper-listbox/paper-listbox.js';
import '@polymer/paper-item/paper-item.js';
import '@polymer/paper-toggle-button/paper-toggle-button.js';
import '@polymer/paper-slider/paper-slider.js';
import '../../../server/static/src/components/boardgame-deck-defaults.js';
import { BoardgameBaseGameRenderer } from '../../../server/static/src/components/boardgame-base-game-renderer.js';
import '../../../server/static/src/components/boardgame-card.js';
import '../../../server/static/src/components/boardgame-component-stack.js';
import '../../../server/static/src/components/boardgame-fading-text.js';
import '../../../server/static/src/components/boardgame-status-text.js';
import '../../../server/static/src/components/boardgame-token.js';
import { html, css } from 'lit';
import { property } from 'lit/decorators.js';
import { repeat } from 'lit/directives/repeat.js';
import { styleMap } from 'lit/directives/style-map.js';

class BoardgameRenderGameDebuganimations extends BoardgameBaseGameRenderer {
  static override styles = [
    ...(BoardgameBaseGameRenderer.styles ? [BoardgameBaseGameRenderer.styles] : []),
    css`
      .slow {
        --animation-length: 5s;
      }

      #shortstacks {
        display: flex;
        flex-direction: row;
      }

      #draw {
        display: flex;
        flex-direction: row;
      }

      #shortstacks boardgame-card > div {
        display: flex;
        flex-direction: row;
        align-items: center;
        justify-content: center;
      }

      #fan {
        display: flex;
        flex-direction: row;
      }

      #fan boardgame-component-stack:first-child {
        --component-scale: 1.2;
      }

      .flex {
        flex: 1;
      }

      .controls {
        display: flex;
        flex-direction: column;
      }

      #hidden {
        display: flex;
        flex-direction: row;
      }

      #controls {
        display: flex;
        flex-direction: row;
      }

      #all {
        display: flex;
        flex-direction: row;
      }

      #token {
        display: flex;
        flex-direction: row;
      }

      #tokens {
        display: flex;
        flex-direction: row;
      }

      #tokens-sanitized {
        display: flex;
        flex-direction: row;
      }
    `
  ];

  @property({ type: String })
  fromStackLayout = 'fan';

  @property({ type: Boolean })
  fromStackRotated = false;

  @property({ type: Boolean })
  toStackRotated = false;

  @property({ type: Boolean })
  messy = true;

  @property({ type: Boolean })
  slowAnimations = false;

  @property({ type: Number })
  fromCardScale = 1.0;

  @property({ type: Number })
  toCardScale = 1.0;

  @property({ type: Boolean })
  tokenActive = false;

  @property({ type: Boolean })
  tokenHighlighted = false;

  @property({ type: String })
  tokenType = 'cube';

  @property({ type: String })
  tokenColor = 'red';

  @property({ type: Array })
  legalTokenTypes: string[] = [];

  @property({ type: Array })
  legalTokenColors: string[] = [];

  private _classes(): string {
    if (this.slowAnimations) {
      return 'slow';
    }
    return '';
  }

  override async firstUpdated(_changedProperties: Map<PropertyKey, unknown>) {
    super.firstUpdated(_changedProperties);
    await this.updateComplete; // CRITICAL: Wait for render
    const token = this.renderRoot.querySelector('boardgame-token') as any;
    if (token) {
      this.legalTokenTypes = token.legalTypes;
      this.legalTokenColors = token.legalColors;
    }
  }

  override render() {
    return html`
      <boardgame-deck-defaults>
        <template deck="cards">
          <boardgame-card>
            <div tall>
              {{item.Values.Type}}
            </div>
          </boardgame-card>
        </template>
        <template deck="tokens">
          <boardgame-token></boardgame-token>
        </template>
      </boardgame-deck-defaults>
      <div id="container" class="${this._classes()}">
        <div id="controls">
          <paper-toggle-button
            ?checked="${this.fromStackRotated}"
            @checked-changed="${(e: CustomEvent) => { this.fromStackRotated = e.detail.value; }}">
            From Rotated
          </paper-toggle-button>
          <paper-toggle-button
            ?checked="${this.toStackRotated}"
            @checked-changed="${(e: CustomEvent) => { this.toStackRotated = e.detail.value; }}">
            To Rotated
          </paper-toggle-button>
          <paper-toggle-button
            ?checked="${this.messy}"
            @checked-changed="${(e: CustomEvent) => { this.messy = e.detail.value; }}">
            Messy
          </paper-toggle-button>
          <paper-toggle-button
            ?checked="${this.slowAnimations}"
            @checked-changed="${(e: CustomEvent) => { this.slowAnimations = e.detail.value; }}">
            Slow Animation
          </paper-toggle-button>
          From scale:
          <paper-slider
            min="0.5"
            max="2.0"
            value="${this.fromCardScale}"
            @value-changed="${(e: CustomEvent) => { this.fromCardScale = e.detail.value; }}"
            pin
            step=".05">
          </paper-slider>
          To Scale:
          <paper-slider
            min="0.5"
            max="2.0"
            value="${this.toCardScale}"
            @value-changed="${(e: CustomEvent) => { this.toCardScale = e.detail.value; }}"
            pin
            step=".05">
          </paper-slider>
        </div>
        <div id="shortstacks">
          <boardgame-component-stack
            layout="stack"
            .stack="${this.state?.Game?.FirstShortStack}"
            ?messy="${this.messy}"
            component-propose-move="Move Card Between Short Stacks">
          </boardgame-component-stack>
          <boardgame-component-stack
            layout="stack"
            ?messy="${this.messy}"
            .stack="${this.state?.Game?.SecondShortStack}"
            component-propose-move="Move Card Between Short Stacks">
          </boardgame-component-stack>
          <paper-button propose-move="Move Card Between Short Stacks">Swap</paper-button>
        </div>

        <div id="draw">
          <boardgame-component-stack
            layout="stack"
            ?messy="${this.messy}"
            .stack="${this.state?.Game?.DrawStack}"
            ?component-rotated="${this.messy}"
            component-index-attributes="my-index,other-index">
          </boardgame-component-stack>
          <boardgame-component-stack
            layout="stack"
            ?messy="${this.messy}"
            .stack="${this.state?.Game?.DiscardStack}">
          </boardgame-component-stack>
          <paper-button propose-move="Move Card Between Draw And Discard Stacks">Draw</paper-button>
          <boardgame-fading-text
            .trigger="${this.state?.Game?.DrawStack?.Components?.length}"
            auto-message="diff">
          </boardgame-fading-text>
        </div>

        <div id="draw">
          <boardgame-component-stack
            layout="stack"
            ?messy="${this.messy}"
            .stack="${this.state?.Game?.Card}">
          </boardgame-component-stack>
          <paper-button propose-move="Flip Card Between Hidden and Revealed">Flip</paper-button>
        </div>

        <div id="fan">
          <boardgame-component-stack
            layout="${this.fromStackLayout}"
            ?messy="${this.messy}"
            .stack="${this.state?.Game?.FanStack}"
            style="${styleMap({ '--component-scale': this.fromCardScale.toString() })}"
            ?component-rotated="${this.fromStackRotated}">
          </boardgame-component-stack>
          <div class="flex"></div>
          <boardgame-component-stack
            layout="stack"
            ?messy="${this.messy}"
            .stack="${this.state?.Game?.FanDiscard}"
            style="${styleMap({ '--component-scale': this.toCardScale.toString() })}"
            ?component-rotated="${this.toStackRotated}">
          </boardgame-component-stack>
          <div class="controls">
            <paper-button propose-move="Move Fan Card">Draw</paper-button>
            <paper-button propose-move="Visible Shuffle">Public Shuffle</paper-button>
            <paper-button propose-move="Shuffle">Shuffle</paper-button>
            <paper-button propose-move="Shuffle Hidden">Shuffle Hidden</paper-button>
            <boardgame-status-text>${this.state?.Game?.FanShuffleCount}</boardgame-status-text>
            <paper-dropdown-menu label="Layout">
              <paper-listbox
                slot="dropdown-content"
                selected="${this.fromStackLayout}"
                @selected-changed="${(e: CustomEvent) => { this.fromStackLayout = e.detail.value; }}"
                attr-for-selected="value">
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
          <boardgame-component-stack
            layout="fan"
            ?messy="${this.messy}"
            .stack="${this.state?.Game?.VisibleStack}"
            style="${styleMap({ '--component-scale': this.fromCardScale.toString() })}"
            ?component-rotated="${this.fromStackRotated}">
          </boardgame-component-stack>
          <boardgame-component-stack
            layout="stack"
            ?messy="${this.messy}"
            .stack="${this.state?.Game?.HiddenStack}"
            style="${styleMap({ '--component-scale': this.toCardScale.toString() })}"
            faux-components="5"
            ?component-rotated="${this.toStackRotated}">
          </boardgame-component-stack>
          <paper-button propose-move="Move Between Hidden">Draw</paper-button>
        </div>

        <div id="all">
          <boardgame-component-stack
            layout="stack"
            ?messy="${this.messy}"
            .stack="${this.state?.Game?.AllVisibleStack}">
          </boardgame-component-stack>
          <boardgame-component-stack
            layout="stack"
            ?messy="${this.messy}"
            .stack="${this.state?.Game?.AllHiddenStack}">
          </boardgame-component-stack>
          <paper-button propose-move="Start Move All Components To Hidden">To Hidden</paper-button>
          <paper-button propose-move="Start Move All Components To Visible">To Visible</paper-button>
        </div>

        <div id="token">
          <boardgame-token
            color="${this.tokenColor}"
            ?highlighted="${this.tokenHighlighted}"
            ?active="${this.tokenActive}"
            type="${this.tokenType}">
          </boardgame-token>
          <div class="flex"></div>
          <paper-toggle-button
            ?checked="${this.tokenHighlighted}"
            @checked-changed="${(e: CustomEvent) => { this.tokenHighlighted = e.detail.value; }}">
            Token Highlighted
          </paper-toggle-button>
          <paper-toggle-button
            ?checked="${this.tokenActive}"
            @checked-changed="${(e: CustomEvent) => { this.tokenActive = e.detail.value; }}">
            Token Active
          </paper-toggle-button>
          <paper-dropdown-menu label="Type">
            <paper-listbox
              slot="dropdown-content"
              selected="${this.tokenType}"
              @selected-changed="${(e: CustomEvent) => { this.tokenType = e.detail.value; }}"
              attr-for-selected="value">
              ${repeat(this.legalTokenTypes, (item) => item, (item) => html`
                <paper-item value="${item}">${item}</paper-item>
              `)}
            </paper-listbox>
          </paper-dropdown-menu>
          <paper-dropdown-menu label="Color">
            <paper-listbox
              slot="dropdown-content"
              selected="${this.tokenColor}"
              @selected-changed="${(e: CustomEvent) => { this.tokenColor = e.detail.value; }}"
              attr-for-selected="value">
              ${repeat(this.legalTokenColors, (item) => item, (item) => html`
                <paper-item value="${item}">${item}</paper-item>
              `)}
            </paper-listbox>
          </paper-dropdown-menu>
        </div>

        <div id="tokens">
          <boardgame-component-stack
            layout="grid"
            ?messy="${this.messy}"
            .stack="${this.state?.Game?.TokensFrom}"
            component-color="${this.tokenColor}"
            component-type="${this.tokenType}">
          </boardgame-component-stack>
          <boardgame-component-stack
            layout="grid"
            ?messy="${this.messy}"
            .stack="${this.state?.Game?.TokensTo}"
            component-color="${this.tokenColor}"
            component-type="${this.tokenType}">
          </boardgame-component-stack>
          <paper-button propose-move="Move Token">Swap</paper-button>
        </div>

        <div id="tokens-sanitized">
          <boardgame-component-stack
            layout="pile"
            ?messy="${this.messy}"
            .stack="${this.state?.Game?.SanitizedTokensFrom}"
            component-color="${this.tokenColor}"
            component-type="${this.tokenType}">
          </boardgame-component-stack>
          <boardgame-component-stack
            layout="pile"
            ?messy="${this.messy}"
            .stack="${this.state?.Game?.SanitizedTokensTo}"
            faux-components="5"
            component-color="${this.tokenColor}"
            component-type="${this.tokenType}">
          </boardgame-component-stack>
          <paper-button propose-move="Move Token Sanitized">Swap</paper-button>
        </div>
      </div>
    `;
  }
}

customElements.define('boardgame-render-game-debuganimations', BoardgameRenderGameDebuganimations);
