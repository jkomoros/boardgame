import '@material/web/select/filled-select.js';
import '@material/web/select/select-option.js';
import '@material/web/switch/switch.js';
import '@material/web/slider/slider.js';
import type { MdSwitch } from '@material/web/switch/switch.js';
import type { MdSlider } from '@material/web/slider/slider.js';
import type { MdFilledSelect } from '@material/web/select/filled-select.js';
import { GameRenderer, registerGameRenderer } from './_game_renderer.js';
import { html, css } from 'lit';
import { property } from 'lit/decorators.js';
import { MoveNames } from './_move_names.js';
import { repeat } from 'lit/directives/repeat.js';
import { styleMap } from 'lit/directives/style-map.js';
import { cardView, fx, isStackLayout, motion, tokenView } from '../../src/client.js';
import type { ClientMove, StackLayout } from '../../src/client.js';
import type {
  EffectSpec,
  EffectTransitionContext,
  MotionReleaseDeclaration,
  MotionStaggerCohortSpec,
} from '../../src/client.js';
import type { GameState, State } from './_types.js';
import type { MoveName } from './_move_names.js';

@registerGameRenderer
export class BoardgameRenderGameDebuganimations extends GameRenderer {
  private readonly cards = cardView<GameState['DrawStack']>({
    render: ({ kind, component }) => kind === 'visible'
      ? html`<div tall>${component.Values.Type}</div>`
      : null,
  });

  private readonly tokens = tokenView<GameState['TokensFrom']>({});

  static override styles = [
    ...(GameRenderer.styles ? [GameRenderer.styles] : []),
    css`
      .slow {
        --animation-length: 5s;
      }

      #container {
        display: flex;
        flex-direction: column;
        gap: 16px;
        padding: 16px;
      }

      #shortstacks {
        display: flex;
        flex-direction: row;
        gap: 16px;
        align-items: center;
      }

      #draw {
        display: flex;
        flex-direction: row;
        gap: 16px;
        align-items: center;
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
        gap: 16px;
        align-items: center;
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
        gap: 8px;
      }

      #hidden {
        display: flex;
        flex-direction: row;
        gap: 16px;
        align-items: center;
      }

      #controls {
        display: flex;
        flex-direction: row;
        gap: 16px;
        align-items: center;
        flex-wrap: wrap;
        padding: 12px 16px;
        background: var(--md-sys-color-surface-container-low, #f7f2fa);
        border-radius: 12px;
      }

      #all {
        display: flex;
        flex-direction: row;
        gap: 16px;
        align-items: center;
      }

      #token {
        display: flex;
        flex-direction: row;
        gap: 16px;
        align-items: center;
      }

      #tokens {
        display: flex;
        flex-direction: row;
        gap: 16px;
        align-items: center;
      }

      #tokens-sanitized {
        display: flex;
        flex-direction: row;
        gap: 16px;
        align-items: center;
      }

      md-filled-button,
      md-outlined-button {
        align-self: center;
      }

      label {
        display: inline-flex;
        align-items: center;
        gap: 8px;
        font-family: var(--md-sys-typescale-body-medium-font, 'Roboto', sans-serif);
        font-size: var(--md-sys-typescale-body-medium-size, 14px);
        color: var(--md-sys-color-on-surface, #1c1b1f);
      }
    `
  ];

  @property({ type: String })
  fromStackLayout: StackLayout = 'fan';

  private _setStackLayout(event: Event): void {
    const value = (event.target as MdFilledSelect).value;
    if (!isStackLayout(value)) throw new Error(`Unexpected stack layout option: ${value}`);
    this.fromStackLayout = value;
  }

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

  /**
   * What the demo token in `#token` stands for.
   *
   * `BoardgameComponent.item` defaults to null, and a null item means
   * `spacer = true`, which means `visibility: hidden`. The demo bound `color`,
   * `type`, `active` and `highlighted` and no item -- so the one widget in the
   * repo where a person can switch a token through all six shapes and ten
   * colours by hand has been invisible the whole time.
   *
   * A frozen constant rather than an inline object literal, because a fresh
   * object every render is a fresh value to Lit: it would re-run
   * `_itemChanged` -- and so rewrite `spacer` and the reflected `id` -- on
   * every keystroke of the two selects. It carries an ID because
   * `_itemChanged` copies `item.ID` onto the host's `id`, and a stable one
   * keeps this widget addressable. See
   * `tests/animations/parity/debuganimations-token-demo.spec.ts`.
   */
  private static readonly demoTokenItem = Object.freeze({ ID: 'demo-token' });

  override animationLength(_fromMove: ClientMove | null, _toMove: ClientMove | null): number {
    if (this.slowAnimations) return 5000;
    return 0;
  }

  override motionReleaseForTransition(
    context: EffectTransitionContext<State, MoveName>,
  ): MotionReleaseDeclaration | null {
    if (!this.slowAnimations || context.kind === 'initial') return null;
    return motion.release({ key: 'slow-animation-cutover', progress: 0.3 });
  }

  override effectsForTransition(
    context: EffectTransitionContext<State, MoveName>,
  ): readonly EffectSpec[] {
    if (context.kind === 'initial') return [];
    if (context.move?.AnimationKey === MoveNames.VisibleShuffle) {
      const priorIndex = new Map(
        context.before.Game.FanStack.IDs.map((id, index) => [id, index]),
      );
      const moved = context.after.Game.FanStack.IDs.filter(
        (id, index) => priorIndex.get(id) !== index,
      );
      return moved.length === 0 ? [] : [fx.afterMotion({
        key: 'visible-shuffle-complete',
        subjects: moved,
        effect: fx.burst({
          at: fx.anchor('visible-shuffle'),
          tone: 'magic',
          intensity: 'small',
          timing: 'immediate',
        }),
      })];
    }
    if (context.move?.AnimationKey !== MoveNames.MoveToken) return [];
    const beforeFrom = new Set(context.before.Game.TokensFrom.IDs);
    const movedTokenId = context.after.Game.TokensFrom.IDs.find(id => !beforeFrom.has(id))
      ?? context.before.Game.TokensFrom.IDs.find(
        id => !context.after.Game.TokensFrom.IDs.includes(id),
      );
    if (!movedTokenId) return [];
    return [fx.decorateMotion({
      subject: movedTokenId,
      trail: {
        tone: 'magic',
        intensity: 'small',
      },
      arrival: fx.burst({
        at: fx.motion(movedTokenId),
        tone: 'reward',
        intensity: 'small',
        timing: 'immediate',
      }),
      key: 'token-transfer',
    })];
  }

  override motionCohortsForTransition(
    context: EffectTransitionContext<State, MoveName>,
  ): readonly MotionStaggerCohortSpec[] {
    if (context.kind === 'initial' || context.move?.AnimationKey !== MoveNames.VisibleShuffle) return [];
    return [motion.stagger({
      key: 'visible-shuffle-cascade',
      subjects: context.after.Game.FanStack.IDs,
      intervalMs: 45,
    })];
  }

  private _classes(): string {
    if (this.slowAnimations) {
      return 'slow';
    }
    return '';
  }

  override async firstUpdated(_changedProperties: Map<PropertyKey, unknown>) {
    super.firstUpdated(_changedProperties);
    await this.updateComplete; // CRITICAL: Wait for render
    const token = this.renderRoot.querySelector('boardgame-token') as (HTMLElement & { legalTypes: string[]; legalColors: string[] }) | null;
    if (token) {
      this.legalTokenTypes = token.legalTypes;
      this.legalTokenColors = token.legalColors;
    }
  }

  override render() {
    const game = this.state?.Game;
    const fromFirstShortStack = (game?.FirstShortStack.Components.length ?? 0) > 0;
    const fromDrawStack = (game?.DiscardStack.Components.length ?? 0) < 3;
    return html`
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
            @change="${(e: Event) => { this.fromCardScale = (e.target as MdSlider).value ?? 1.0; }}"
            labeled
            step="0.05">
          </md-slider>
          To Scale:
          <md-slider
            min="0.5"
            max="2.0"
            .value="${this.toCardScale}"
            @change="${(e: Event) => { this.toCardScale = (e.target as MdSlider).value ?? 1.0; }}"
            labeled
            step="0.05">
          </md-slider>
          <button
            id="effect-demo"
            type="button"
            @click=${(event: Event) => this.effects?.play(fx.parallel([
              fx.burst({
                at: event.currentTarget as HTMLElement,
                tone: 'magic',
                intensity: 'large',
                key: 'debug-celebration',
              }),
              fx.pulse({
                at: event.currentTarget as HTMLElement,
                tone: 'reward',
                intensity: 'medium',
              }),
            ]))}>
            Celebrate
          </button>
        </div>
        <div id="shortstacks">
          <boardgame-component-stack
            layout="stack"
            .stack="${this.state?.Game?.FirstShortStack}"
            .componentView=${this.cards}
            ?messy="${this.messy}"
            components-disabled>
          </boardgame-component-stack>
          <boardgame-component-stack
            layout="stack"
            ?messy="${this.messy}"
            .stack="${this.state?.Game?.SecondShortStack}"
            .componentView=${this.cards}
            components-disabled>
          </boardgame-component-stack>
          <boardgame-action-button
            .action=${this.move(MoveNames.MoveCardBetweenShortStacks).with({ FromFirst: fromFirstShortStack })}>
            Swap
          </boardgame-action-button>
        </div>

        <div id="draw">
          <boardgame-component-stack
            layout="stack"
            ?messy="${this.messy}"
            .stack="${this.state?.Game?.DrawStack}"
            .componentView=${this.cards.withProperties({ rotated: this.messy })}>
          </boardgame-component-stack>
          <boardgame-component-stack
            layout="stack"
            ?messy="${this.messy}"
            .stack="${this.state?.Game?.DiscardStack}"
            .componentView=${this.cards}>
          </boardgame-component-stack>
          <boardgame-action-button
            .action=${this.move(MoveNames.MoveCardBetweenDrawAndDiscardStacks).with({ FromDraw: fromDrawStack })}>
            Draw
          </boardgame-action-button>
          <boardgame-fading-text
            .trigger="${this.state?.Game?.DrawStack?.Components?.length}"
            auto-message="diff">
          </boardgame-fading-text>
        </div>

        <div id="draw">
          <boardgame-component-stack
            layout="stack"
            ?messy="${this.messy}"
            .stack="${this.state?.Game?.Card}"
            .componentView=${this.cards}>
          </boardgame-component-stack>
          <boardgame-action-button .action=${this.move(MoveNames.FlipCardBetweenHiddenAndRevealed)}>Flip</boardgame-action-button>
        </div>

        <div id="fan" data-effect-anchor="visible-shuffle">
          <boardgame-component-stack
            layout="${this.fromStackLayout}"
            ?messy="${this.messy}"
            .stack="${this.state?.Game?.FanStack}"
            .componentView=${this.cards.withProperties({ rotated: this.fromStackRotated })}
            style="${styleMap({ '--component-scale': this.fromCardScale.toString() })}"
            >
          </boardgame-component-stack>
          <div class="flex"></div>
          <boardgame-component-stack
            layout="stack"
            ?messy="${this.messy}"
            .stack="${this.state?.Game?.FanDiscard}"
            .componentView=${this.cards.withProperties({ rotated: this.toStackRotated })}
            style="${styleMap({ '--component-scale': this.toCardScale.toString() })}"
            >
          </boardgame-component-stack>
          <div class="controls">
            <boardgame-action-button .action=${this.move(MoveNames.MoveFanCard)}>Draw</boardgame-action-button>
            <boardgame-action-button .action=${this.move(MoveNames.VisibleShuffle)}>Public Shuffle</boardgame-action-button>
            <boardgame-action-button .action=${this.move(MoveNames.Shuffle)}>Shuffle</boardgame-action-button>
            <boardgame-action-button .action=${this.move(MoveNames.ShuffleHidden)}>Shuffle Hidden</boardgame-action-button>
            <boardgame-status-text .value=${this.state?.Game?.FanShuffleCount}></boardgame-status-text>
            <md-filled-select
              label="Layout"
              .value="${this.fromStackLayout}"
              @change="${this._setStackLayout}">
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
            .componentView=${this.cards.withProperties({ rotated: this.fromStackRotated })}
            style="${styleMap({ '--component-scale': this.fromCardScale.toString() })}"
            >
          </boardgame-component-stack>
          <boardgame-component-stack
            layout="stack"
            ?messy="${this.messy}"
            .stack="${this.state?.Game?.HiddenStack}"
            .componentView=${this.cards.withProperties({ rotated: this.toStackRotated })}
            style="${styleMap({ '--component-scale': this.toCardScale.toString() })}"
            faux-components="5"
            >
          </boardgame-component-stack>
          <boardgame-action-button .action=${this.move(MoveNames.MoveBetweenHidden)}>Draw</boardgame-action-button>
        </div>

        <div id="all">
          <boardgame-component-stack
            layout="stack"
            ?messy="${this.messy}"
            .stack="${this.state?.Game?.AllVisibleStack}"
            .componentView=${this.cards}>
          </boardgame-component-stack>
          <boardgame-component-stack
            layout="stack"
            ?messy="${this.messy}"
            .stack="${this.state?.Game?.AllHiddenStack}"
            .componentView=${this.cards}>
          </boardgame-component-stack>
          <boardgame-action-button .action=${this.move(MoveNames.StartMoveAllComponentsToHidden)}>To Hidden</boardgame-action-button>
          <boardgame-action-button .action=${this.move(MoveNames.StartMoveAllComponentsToVisible)}>To Visible</boardgame-action-button>
        </div>

        <div id="token">
          <boardgame-token
            .item=${BoardgameRenderGameDebuganimations.demoTokenItem}
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
            data-effect-anchor="token-source"
            layout="grid"
            ?messy="${this.messy}"
            .stack="${this.state?.Game?.TokensFrom}"
            .componentView=${this.tokens.withProperties({ color: this.tokenColor, type: this.tokenType })}>
          </boardgame-component-stack>
          <boardgame-component-stack
            data-effect-anchor="token-destination"
            layout="grid"
            ?messy="${this.messy}"
            .stack="${this.state?.Game?.TokensTo}"
            .componentView=${this.tokens.withProperties({ color: this.tokenColor, type: this.tokenType })}>
          </boardgame-component-stack>
          <boardgame-action-button .action=${this.move(MoveNames.MoveToken)}>Swap</boardgame-action-button>
        </div>

        <div id="tokens-sanitized">
          <boardgame-component-stack
            layout="pile"
            ?messy="${this.messy}"
            .stack="${this.state?.Game?.SanitizedTokensFrom}"
            .componentView=${this.tokens.withProperties({ color: this.tokenColor, type: this.tokenType })}>
          </boardgame-component-stack>
          <boardgame-component-stack
            layout="pile"
            ?messy="${this.messy}"
            .stack="${this.state?.Game?.SanitizedTokensTo}"
            .componentView=${this.tokens.withProperties({ color: this.tokenColor, type: this.tokenType })}
            faux-components="5"
            >
          </boardgame-component-stack>
          <boardgame-action-button .action=${this.move(MoveNames.MoveTokenSanitized)}>Swap</boardgame-action-button>
        </div>
      </div>
    `;
  }
}
