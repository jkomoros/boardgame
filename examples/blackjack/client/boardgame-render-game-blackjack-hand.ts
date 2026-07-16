import { html, css } from 'lit';
import { cardView } from '../../src/client.js';
import { MoveNames } from './_move_names.js';
import { HandRenderer, registerHandRenderer } from './_game_renderer.js';
import type { GameState } from './_types.js';

/**
 * Blackjack Hand view (the player's phone). Connects as PlayerIndex(n);
 * renders the player's own private hand + Hit/Stand buttons. The base
 * provides the top-edge inbound anchor for cross-screen card animations
 * (deals come in from the Table view).
 *
 * V1 MVP minimal styling.
 */
@registerHandRenderer
export class BlackjackHandView extends HandRenderer {
  private readonly cards = cardView<GameState['DrawStack']>({
    properties: ({ kind, component }) => ({
      suit: kind === 'visible' ? component.Values.Suit : '',
      rank: kind === 'visible' ? component.Values.Rank : '',
    }),
  });
  static override styles = [
    HandRenderer.styles,
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
      boardgame-action-bar boardgame-action-button {
        font-size: 18px;
      }
    `,
  ];

  override render() {
    const player = this.playerState;
    return html`
      ${this.renderTopEdgeAnchor()}
      ${this.renderHandHeader()}
      <h1>Your Hand</h1>
      <div class="hand">
        ${player?.HiddenHand
          ? html`<boardgame-component-stack .stack=${player.HiddenHand} .componentView=${this.cards.withProperties({ rotated: true })} layout="fan"></boardgame-component-stack>`
          : html`<small>waiting…</small>`}
        ${player?.VisibleHand
          ? html`<boardgame-component-stack .stack=${player.VisibleHand} .componentView=${this.cards.withProperties({ rotated: true })} layout="fan"></boardgame-component-stack>`
          : ''}
      </div>
      <boardgame-action-bar label="Blackjack actions">
        <boardgame-action-button .action=${this.move(MoveNames.CurrentPlayerHit)}>Hit</boardgame-action-button>
        <boardgame-action-button .action=${this.move(MoveNames.CurrentPlayerStand)}>Stand</boardgame-action-button>
      </boardgame-action-bar>
    `;
  }
}
