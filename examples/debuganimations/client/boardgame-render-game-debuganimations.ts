import '@material/web/button/filled-button.js';
import '@material/web/select/filled-select.js';
import '@material/web/select/select-option.js';
import '@material/web/switch/switch.js';
import '@material/web/slider/slider.js';
import type { MdSwitch } from '@material/web/switch/switch.js';
import type { MdSlider } from '@material/web/slider/slider.js';
import type { MdFilledSelect } from '@material/web/select/filled-select.js';
import '../../src/components/boardgame-deck-defaults.js';
import { BoardgameBaseGameRenderer } from '../../src/components/boardgame-base-game-renderer.js';
import '../../src/components/boardgame-card.js';
import '../../src/components/boardgame-component-stack.js';
import '../../src/components/boardgame-fading-text.js';
import '../../src/components/boardgame-status-text.js';
import '../../src/components/boardgame-token.js';
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
          <label><md-switch
            ?selected="${this.fromStackRotated}"
            @change="${(e: Event) => { this.fromStackRotated = (e.target as MdSwitch).selected; }}">
          </md-switch> From Rotated</label>
          <label><md-switch
            ?selected="${this.toStackRotated}"
            @change="${(e: Event) => { this.toStackRotated = (e.target as MdSwitch).selected; }}">
          </md-switch> To Rotated</label>
          <label><md-switch
            ?selected="${this.messy}"
            @change="${(e: Event) => { this.messy = (e.target as MdSwitch).selected; }}">
          </md-switch> Messy</label>
          <label><md-switch
            ?selected="${this.slowAnimations}"
            @change="${(e: Event) => { this.slowAnimations = (e.target as MdSwitch).selected; }}">
          </md-switch> Slow Animation</label>
          From scale:
          <md-slider
            min="0.5"
            max="2.0"
            .value="${this.fromCardScale}"
            @change="${(e: Event) => { this.fromCardScale = (e.target as MdSlider).value; }}"
            labeled
            step="0.05">
          </md-slider>
          To Scale:
          <md-slider
            min="0.5"
            max="2.0"
            .value="${this.toCardScale}"
            @change="${(e: Event) => { this.toCardScale = (e.target as MdSlider).value; }}"
            labeled
            step="0.05">
          </md-slider>
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
          <md-filled-button propose-move="Move Card Between Short Stacks">Swap</md-filled-button>
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
          <md-filled-button propose-move="Move Card Between Draw And Discard Stacks">Draw</md-filled-button>
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
          <md-filled-button propose-move="Flip Card Between Hidden and Revealed">Flip</md-filled-button>
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
            <md-filled-button propose-move="Move Fan Card">Draw</md-filled-button>
            <md-filled-button propose-move="Visible Shuffle">Public Shuffle</md-filled-button>
            <md-filled-button propose-move="Shuffle">Shuffle</md-filled-button>
            <md-filled-button propose-move="Shuffle Hidden">Shuffle Hidden</md-filled-button>
            <boardgame-status-text>${this.state?.Game?.FanShuffleCount}</boardgame-status-text>
            <md-filled-select
              label="Layout"
              .value="${this.fromStackLayout}"
              @change="${(e: Event) => { this.fromStackLayout = (e.target as MdFilledSelect).value; }}">
              <md-select-option value="fan">
                <div slot="headline">fan</div>
              </md-select-option>
              <md-select-option value="spread">
                <div slot="headline">spread</div>
              </md-select-option>
              <md-select-option value="stack">
                <div slot="headline">stack</div>
              </md-select-option>
              <md-select-option value="grid">
                <div slot="headline">grid</div>
              </md-select-option>
              <md-select-option value="pile">
                <div slot="headline">pile</div>
              </md-select-option>
            </md-filled-select>
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
          <md-filled-button propose-move="Move Between Hidden">Draw</md-filled-button>
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
          <md-filled-button propose-move="Start Move All Components To Hidden">To Hidden</md-filled-button>
          <md-filled-button propose-move="Start Move All Components To Visible">To Visible</md-filled-button>
        </div>

        <div id="token">
          <boardgame-token
            color="${this.tokenColor}"
            ?highlighted="${this.tokenHighlighted}"
            ?active="${this.tokenActive}"
            type="${this.tokenType}">
          </boardgame-token>
          <div class="flex"></div>
          <label><md-switch
            ?selected="${this.tokenHighlighted}"
            @change="${(e: Event) => { this.tokenHighlighted = (e.target as MdSwitch).selected; }}">
          </md-switch> Token Highlighted</label>
          <label><md-switch
            ?selected="${this.tokenActive}"
            @change="${(e: Event) => { this.tokenActive = (e.target as MdSwitch).selected; }}">
          </md-switch> Token Active</label>
          <md-filled-select
            label="Type"
            .value="${this.tokenType}"
            @change="${(e: Event) => { this.tokenType = (e.target as MdFilledSelect).value; }}">
            ${repeat(this.legalTokenTypes, (item) => item, (item) => html`
              <md-select-option value="${item}">
                <div slot="headline">${item}</div>
              </md-select-option>
            `)}
          </md-filled-select>
          <md-filled-select
            label="Color"
            .value="${this.tokenColor}"
            @change="${(e: Event) => { this.tokenColor = (e.target as MdFilledSelect).value; }}">
            ${repeat(this.legalTokenColors, (item) => item, (item) => html`
              <md-select-option value="${item}">
                <div slot="headline">${item}</div>
              </md-select-option>
            `)}
          </md-filled-select>
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
          <md-filled-button propose-move="Move Token">Swap</md-filled-button>
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
          <md-filled-button propose-move="Move Token Sanitized">Swap</md-filled-button>
        </div>
      </div>
    `;
  }
}

customElements.define('boardgame-render-game-debuganimations', BoardgameRenderGameDebuganimations);
