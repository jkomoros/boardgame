import '@material/web/button/filled-button.js';
import '@material/web/button/outlined-button.js';
import '@material/web/progress/linear-progress.js';
import { BoardgameBaseGameRenderer } from '../../src/components/boardgame-base-game-renderer.js';
import '../../src/components/boardgame-card.js';
import '../../src/components/boardgame-component-stack.js';
import '../../src/components/boardgame-fading-text.js';
import '../../src/components/boardgame-deck-defaults.js';
import '../../src/components/boardgame-player-badge.js';
import { html, css } from 'lit';
import { MoveNames } from './_move_names.js';
import type { MoveName } from './_move_names.js';
import { moveInputSchema as generatedMoveInputSchema, moveInputSchemaFingerprint as generatedMoveInputSchemaFingerprint, type MoveInputs } from './_move_args.js';
import type { CardsComponentValues, GameState, PlayerState } from './_types.js';

class BoardgameRenderGameMemory extends BoardgameBaseGameRenderer<GameState, PlayerState, MoveName, MoveInputs> {
	protected override readonly moveInputSchema = generatedMoveInputSchema;
	protected override readonly moveInputSchemaFingerprint = generatedMoveInputSchemaFingerprint;
  static override styles = [
    ...(BoardgameBaseGameRenderer.styles ? [BoardgameBaseGameRenderer.styles] : []),
    css`
      md-linear-progress {
        width: 100%;
      }

      .current {
        font-weight: bold;
      }

      boardgame-card > div {
        font-family: var(--md-sys-typescale-display-small-font, 'Crimson Text', serif);
        font-size: 34px;
        font-weight: 400;
        letter-spacing: -.01em;
        line-height: 40px;
      }

      .discards {
        --component-scale: 0.7;
        display: flex;
        flex-direction: row;
        justify-content: space-around;
      }

      .discard-pile {
        display: flex;
        flex-direction: column;
        align-items: center;
        gap: 4px;
      }
    `
  ];

  get maxTimeLeft(): number {
    return this.computeMaxTimeLeft(this.state?.Game?.HideCardsTimer?.originalTimeLeft ?? 0);
  }

  private computeMaxTimeLeft(timeLeft: number): number {
    return Math.max(timeLeft, 100);
  }

  // _revealHoldMs replaces this renderer's old imperative delay-animation
  // hook, which delayed installing the next state by 1000ms whenever the move
  // about to be installed was the engine's "Capture Cards" FixUp move (i.e.
  // the two currently-revealed cards match), so players could see the
  // matched pair for a beat before they animate away to the winner's pile.
  // That hook was told about the upcoming move directly; the declarative
  // replacement (post-animation-delay, #715) instead infers the same
  // condition from currently-rendered state: two visible cards of the same
  // Type is exactly the situation in which the engine's next queued bundle
  // will be Capture Cards (its Legal() requires exactly two matching
  // VisibleCards; otherwise a different fixup move applies, and the hold
  // does not apply).
  //
  // VisibleCards.Components is a fixed-size (SizedStack) array padded with
  // nulls at unrevealed slots, matching the template's own
  // {{item.Values.Type}} access -- component field values live under
  // `.Values`, not directly on the component (that nesting isn't reflected
  // in the shared Component<T> TS type, so read defensively).
  private _revealHoldMs(): number {
    const components = this.state?.Game?.VisibleCards?.Components;
    if (!components) return 0;
    const revealed = components.filter((c): c is NonNullable<typeof c> => !!c);
    if (revealed.length !== 2) return 0;
    const [first, second] = revealed as unknown as { Values?: CardsComponentValues }[];
    const firstType = first.Values?.Type;
    const secondType = second.Values?.Type;
    if (firstType === undefined || secondType === undefined) return 0;
    return firstType === secondType ? 1000 : 0;
  }

  override render() {
    return html`
      <boardgame-deck-defaults>
        <template deck="cards">
          <boardgame-card>
            <div>
              {{item.Values.Type}}
            </div>
          </boardgame-card>
        </template>
      </boardgame-deck-defaults>
      <h2>Memory</h2>
      <div>
        <boardgame-component-stack
          layout="grid"
          messy
          post-animation-delay="${this._revealHoldMs()}"
          .stack="${this.state?.Game?.Cards}"
          .componentAttrs=${{ proposeMove: MoveNames.RevealCard, indexAttributes: 'data-arg-card-index' }}>
        </boardgame-component-stack>
        <boardgame-fading-text
          message="Match"
          .trigger="${this.state?.Game?.Cards?.Components?.length}">
        </boardgame-fading-text>
      </div>
      <div class="discards">
        <div class="discard-pile">
          <boardgame-player-badge player-index="0" compact></boardgame-player-badge>
          <boardgame-component-stack
            layout="stack"
            .stack="${this.state?.Players?.[0]?.WonCards}"
            messy
            .componentAttrs=${{ disabled: true }}>
          </boardgame-component-stack>
        </div>
        <!-- have a boardgame-card spacer just to keep that row height sane even with no cards -->
        <boardgame-card spacer></boardgame-card>
        <div class="discard-pile">
          <boardgame-player-badge player-index="1" compact></boardgame-player-badge>
          <boardgame-component-stack
            layout="stack"
            messy
            .stack="${this.state?.Players?.[1]?.WonCards}"
            .componentAttrs=${{ disabled: true }}>
          </boardgame-component-stack>
        </div>
      </div>
      <md-outlined-button
        id="hide"
        propose-move="${MoveNames.HideCards}"
        ?disabled="${!this.isMoveCurrentlyLegal(MoveNames.HideCards)}">
        Hide Cards
      </md-outlined-button>
      <md-linear-progress
        id="timeleft"
        .value="${(this.state?.Game?.HideCardsTimer?.TimeLeft || 0) / (this.maxTimeLeft || 1)}"
        .max="${1}">
      </md-linear-progress>
      <boardgame-fading-text
        .trigger="${this.isCurrentPlayer}"
        message="Your Turn"
        suppress="falsey">
      </boardgame-fading-text>
    `;
  }
}

customElements.define('boardgame-render-game-memory', BoardgameRenderGameMemory);
