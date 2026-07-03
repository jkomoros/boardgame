import { html, css } from 'lit';
import { customElement } from 'lit/decorators.js';
import { BoardgameHandViewBase } from '../../src/components/boardgame-hand-view-base.js';
import '../../src/components/boardgame-component-stack.js';
import '../../src/components/boardgame-card.js';
import '../../src/components/boardgame-deck-defaults.js';
import { MoveNames, type MoveName } from './_move_names.js';
import type { GameState, PlayerState } from './_types.js';

/**
 * Blackjack Hand view (the player's phone). Connects as PlayerIndex(n);
 * renders the player's own private hand + Hit/Stand buttons. The base
 * provides the top-edge inbound anchor for cross-screen card animations
 * (deals come in from the Table view).
 *
 * V1 MVP minimal styling.
 */
@customElement('boardgame-render-game-blackjack-hand')
export class BlackjackHandView extends BoardgameHandViewBase<GameState, PlayerState, MoveName> {
  static override styles = [
    BoardgameHandViewBase.styles,
    css`
      :host {
        display: block;
        min-height: 100vh;
        padding: 16px;
        background: #1a4d2e;
        color: white;
        font-family: system-ui, sans-serif;
      }
      h1 {
        text-align: center;
        margin: 0 0 16px 0;
        font-size: 20px;
      }
      .hand {
        display: flex;
        flex-wrap: wrap;
        padding: 0 14px;
        justify-content: center;
        align-items: center;
        gap: 12px;
        margin: 24px 0;
        min-height: 170px;
        /* Long hands (repeated Hits) wrap instead of clipping off the
           right edge of the phone; card width also shrinks a little on
           narrow screens. */
        /* Big readable cards on a phone. boardgame-card's sizing model:
           width var × aspect; the rotated attr (set below, matching the
           solo renderer's convention) swaps the axes so cards stand
           upright/portrait. */
        --component-width: clamp(76px, 24vw, 104px);
      }
      .hand boardgame-component-stack {
        /* The stack's own :host defaults to width:100%, which makes two
           sibling stacks fight for the row and pile their cards on top
           of each other. Size to content instead. */
        width: auto;
      }
      .actions {
        display: flex;
        justify-content: center;
        gap: 12px;
      }
      .actions button {
        padding: 16px 32px;
        font-size: 18px;
        border-radius: 8px;
        border: 2px solid white;
        background: transparent;
        color: white;
        font-weight: 600;
        cursor: pointer;
      }
      .actions button[disabled] {
        opacity: 0.35;
        cursor: default;
      }
    `,
  ];

  override render() {
    const player = this.playerState as any;
    const canAct = this.isMoveCurrentlyLegal(MoveNames.CurrentPlayerHit);
    return html`
      ${this.renderTopEdgeAnchor()}
      ${this.renderHandHeader()}
      <h1>Your Hand</h1>
      <div class="hand">
        <boardgame-deck-defaults>
          <template deck="cards">
            <boardgame-card suit="{{item.Values.Suit}}" rank="{{item.Values.Rank}}"></boardgame-card>
          </template>
        </boardgame-deck-defaults>
        ${player?.HiddenHand
          ? html`<boardgame-component-stack .stack=${player.HiddenHand} layout="fan" .componentAttrs=${{ rotated: true }}></boardgame-component-stack>`
          : html`<small>waiting…</small>`}
        ${player?.VisibleHand
          ? html`<boardgame-component-stack .stack=${player.VisibleHand} layout="fan" .componentAttrs=${{ rotated: true }}></boardgame-component-stack>`
          : ''}
      </div>
      <div class="actions">
        <button ?disabled=${!canAct} @click=${() => this.proposeMove(MoveNames.CurrentPlayerHit, { TargetPlayerIndex: this.viewingAs })}>Hit</button>
        <button ?disabled=${!canAct} @click=${() => this.proposeMove(MoveNames.CurrentPlayerStand, { TargetPlayerIndex: this.viewingAs })}>Stand</button>
      </div>
    `;
  }
}
