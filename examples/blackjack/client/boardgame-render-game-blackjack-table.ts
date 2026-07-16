import { html, css } from 'lit';
import { cardView } from '../../src/client.js';
import { TableRenderer, registerTableRenderer } from './_game_renderer.js';
import type { GameState } from './_types.js';

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
@registerTableRenderer
export class BlackjackTableView extends TableRenderer {
  private readonly cards = cardView<GameState['DrawStack']>({
    properties: ({ kind, component }) => ({
      suit: kind === 'visible' ? component.Values.Suit : '',
      rank: kind === 'visible' ? component.Values.Rank : '',
    }),
  });

  static override styles = [
    TableRenderer.styles,
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
        display: flex;
        justify-content: center;
        margin: 24px auto;
      }
      .draw boardgame-component-stack {
        width: auto;
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
        align-items: center;
        min-height: 110px;
        --component-width: 64px;
      }
      .seat-cards boardgame-component-stack {
        width: auto;
      }
      .draw {
        --component-width: 64px;
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
      <div class="draw">
        ${this.state?.Game?.DrawStack
          ? html`<boardgame-component-stack id="deal-source" .stack=${this.state.Game.DrawStack} .componentView=${this.cards.withProperties({ rotated: true })}></boardgame-component-stack>`
          : html`<small>waiting for state…</small>`}
      </div>
      <div class="seats">
        ${players.map((p, i) => html`
          <div class="seat ${i === this.currentPlayerIndex ? 'current' : ''}">
            <div class="seat-name">${nameFor(i)} · ${p.Score} pts</div>
            <div class="seat-cards">
              ${p.VisibleHand
                ? html`<boardgame-component-stack .stack=${p.VisibleHand} .componentView=${this.cards.withProperties({ rotated: true })} layout="fan" messy></boardgame-component-stack>`
                : ''}
              ${p.HiddenHand
                ? html`<boardgame-component-stack .stack=${p.HiddenHand} .componentView=${this.cards.withProperties({ rotated: true })} layout="fan" messy></boardgame-component-stack>`
                : ''}
            </div>
            ${p.Stood ? html`<small>Standing</small>` : ''}
            ${p.Eliminated && !this.animating ? html`<small>Busted!</small>` : ''}
          </div>
        `)}
      </div>
      ${this.renderFakeDeckRow()}
    `;
  }
}
