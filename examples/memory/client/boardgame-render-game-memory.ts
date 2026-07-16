import { GameRenderer, registerGameRenderer } from './_game_renderer.js';
import { html, css } from 'lit';
import { MoveNames } from './_move_names.js';
import type { CardsComponentValues, GameState } from './_types.js';
import { cardView, isVisibleComponent } from '../../src/client.js';

@registerGameRenderer
export class BoardgameRenderGameMemory extends GameRenderer {
  private readonly cards = cardView<GameState['Cards']>({
    render: ({ kind, component }) => kind === 'visible'
      ? html`<div>${component.Values.Type}</div>`
      : null,
  });

  static override styles = [
    ...(GameRenderer.styles ? [GameRenderer.styles] : []),
    css`
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
  // nulls at unrevealed slots. Component field values live under `.Values`,
  // not directly on the component. Opaque occupied slots are `{}`
  // and must be narrowed with the shared guard before reading card values.
  private _revealHoldMs(): number {
    const components = this.state?.Game?.VisibleCards?.Components;
    if (!components) return 0;
    const revealed = components.filter(isVisibleComponent);
    if (revealed.length !== 2) return 0;
    const [first, second] = revealed;
    const firstType: CardsComponentValues['Type'] = first!.Values.Type;
    const secondType: CardsComponentValues['Type'] = second!.Values.Type;
    return firstType === secondType ? 1000 : 0;
  }

  override render() {
    const cardStack = this.state?.Game?.Cards ?? null;
    const cardSlots = cardStack?.Components.map((_component, index) => index) ?? [];
    const reveals = this.move(MoveNames.RevealCard).targets(
      cardSlots, cardIndex => ({ CardIndex: cardIndex }),
    );
    return html`
      <boardgame-game-surface heading="Memory">
        <div>
          <boardgame-component-stack
            layout="grid"
            messy
            post-animation-delay="${this._revealHoldMs()}"
            .stack="${cardStack}"
            .componentView=${this.cards}
            .componentActions=${reveals.candidates.map(candidate => candidate.action)}>
          </boardgame-component-stack>
          <boardgame-fading-text
            message="Match"
            .trigger="${this.state?.Game?.Cards?.Components?.length}">
          </boardgame-fading-text>
        </div>
        <div class="discards">
          <div class="discard-pile">
            <boardgame-player-badge .player=${this.playerPresentation(0)} compact></boardgame-player-badge>
            <boardgame-component-stack
              layout="stack"
              .stack="${this.state?.Players?.[0]?.WonCards}"
              .componentView=${this.cards}
              messy
              components-disabled>
            </boardgame-component-stack>
          </div>
          <!-- have a boardgame-card spacer just to keep that row height sane even with no cards -->
          <boardgame-card spacer></boardgame-card>
          <div class="discard-pile">
            <boardgame-player-badge .player=${this.playerPresentation(1)} compact></boardgame-player-badge>
            <boardgame-component-stack
              layout="stack"
              messy
              .stack="${this.state?.Players?.[1]?.WonCards}"
              .componentView=${this.cards}
              components-disabled>
            </boardgame-component-stack>
          </div>
        </div>
        <boardgame-action-bar slot="actions" label="Memory actions">
          <boardgame-action-button
            id="hide"
            .action=${this.move(MoveNames.HideCards)}>
            Hide Cards
          </boardgame-action-button>
        </boardgame-action-bar>
        <boardgame-timer
          slot="status"
          id="timeleft"
          label="Cards hide in"
          .timer=${this.state?.Game?.HideCardsTimer ?? null}>
        </boardgame-timer>
        <boardgame-turn-status
          slot="status"
          .turn=${this.turnStatus}>
        </boardgame-turn-status>
      </boardgame-game-surface>
    `;
  }
}
