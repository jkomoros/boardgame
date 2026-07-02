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
    `,
  ];

  override render() {
    return html`
      <h1>Blackjack — Table</h1>
      ${this.renderRoomCodeBanner()}
      ${this.renderAvatarStrip()}
      ${this.renderHostControls()}
      <div class="draw">
        <boardgame-deck-defaults>
          <template deck="cards">
            <boardgame-card suit="{{item.Values.Suit}}" rank="{{item.Values.Rank}}"></boardgame-card>
          </template>
        </boardgame-deck-defaults>
        ${this.state?.Game?.DrawStack
          ? html`<boardgame-component-stack .stack=${(this.state.Game as any).DrawStack}></boardgame-component-stack>`
          : html`<small>waiting for state…</small>`}
      </div>
      ${this.renderFakeDeckRow()}
    `;
  }
}
