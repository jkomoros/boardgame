import { html, css } from 'lit';
import { customElement } from 'lit/decorators.js';
import { BoardgameTableViewBase } from '../../src/components/boardgame-table-view-base.js';
import '../../src/components/boardgame-component-stack.js';
import '../../src/components/boardgame-card.js';
import '../../src/components/boardgame-deck-defaults.js';
import type { MoveName } from './_move_names.js';
import type { GameState, PlayerState } from './_types.js';

/**
 * Blackjack Table view (the shared projector). Connects as
 * ObserverPlayerIndex; renders the dealer area + the avatar strip up
 * top + the fake-deck row along the bottom for cross-screen card
 * animations.
 *
 * V1 MVP: minimal visual polish. The goal is to prove the
 * Table+Hand convention works end-to-end with a real game registered
 * in config.json; the boardgame-util filesystem walk picks up this
 * file (paired with -hand.ts) and emits "blackjack" into the
 * companion_capable_games list so the create-game form gates the
 * toggle on.
 */
@customElement('boardgame-render-game-blackjack-table')
export class BlackjackTableView extends BoardgameTableViewBase<GameState, PlayerState, MoveName> {
  static override styles = [
    BoardgameTableViewBase.styles,
    css`
      :host {
        display: block;
        min-height: 100vh;
        padding: 24px;
        background: #1a4d2e;
        color: white;
        font-family: system-ui, sans-serif;
      }
      h1 {
        text-align: center;
        margin: 0 0 16px 0;
      }
      .draw {
        text-align: center;
        margin: 32px auto;
      }
      .seats {
        display: flex;
        flex-wrap: wrap;
        gap: 24px;
        justify-content: center;
        margin: 24px 0;
      }
      .seat {
        text-align: center;
        padding: 12px 16px;
        border-radius: 12px;
        border: 2px solid transparent;
      }
      .seat.current {
        border-color: gold;
      }
      .seat-name {
        font-weight: 700;
        margin-bottom: 8px;
      }
      .seat-cards {
        display: flex;
        gap: 8px;
        justify-content: center;
      }
    `,
  ];

  override render() {
    const players = this.state?.Players ?? [];
    const nameFor = (i: number): string => {
      const seat = this.seatPresentations.find((s) => s.playerIndex === i);
      return seat ? seat.displayName : `Player ${i}`;
    };
    return html`
      <h1>Blackjack — Table</h1>
      ${this.renderRoomCodeBanner()}
      ${this.renderGameOverBanner()}
      ${this.renderAvatarStrip()}
      ${this.renderHostControls()}
      <boardgame-deck-defaults>
        <template deck="cards">
          <boardgame-card suit="{{item.Values.Suit}}" rank="{{item.Values.Rank}}"></boardgame-card>
        </template>
      </boardgame-deck-defaults>
      <div class="draw" id="deal-source">
        ${this.state?.Game?.DrawStack
          ? html`<boardgame-component-stack .stack=${(this.state.Game as any).DrawStack}></boardgame-component-stack>`
          : html`<small>waiting for state…</small>`}
      </div>
      <div class="seats">
        ${players.map((p, i) => html`
          <div class="seat ${i === this.currentPlayerIndex ? 'current' : ''}">
            <div class="seat-name">${nameFor(i)} · ${(p as any).Score ?? 0} pts</div>
            <div class="seat-cards">
              ${(p as any).VisibleHand
                ? html`<boardgame-component-stack .stack=${(p as any).VisibleHand} layout="fan" messy></boardgame-component-stack>`
                : ''}
              ${(p as any).HiddenHand
                ? html`<boardgame-component-stack .stack=${(p as any).HiddenHand} layout="fan" messy></boardgame-component-stack>`
                : ''}
            </div>
            ${(p as any).Stood ? html`<small>Standing</small>` : ''}
            ${(p as any).Eliminated ? html`<small>Busted!</small>` : ''}
          </div>
        `)}
      </div>
      ${this.renderFakeDeckRow()}
    `;
  }
}
