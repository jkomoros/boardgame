import { GameRenderer, registerGameRenderer } from './_game_renderer.js';
import { html, css } from 'lit';
import { classMap } from 'lit/directives/class-map.js';
import { repeat } from 'lit/directives/repeat.js';
import { MoveNames } from './_move_names.js';
import { cardView } from '../../src/client.js';
import type { GameState } from './_types.js';

@registerGameRenderer
export class BoardgameRenderGameBlackjack extends GameRenderer {
  private readonly cards = cardView<GameState['DrawStack']>({
    properties: ({ kind, component }) => ({
      suit: kind === 'visible' ? component.Values.Suit : '',
      rank: kind === 'visible' ? component.Values.Rank : '',
    }),
  });

  static override styles = [
    ...(GameRenderer.styles ? [GameRenderer.styles] : []),
    css`
      #players {
        --boardgame-player-grid-min-width: 14rem;
      }

      /*
       * A busted seat is the framework's "eliminated" state. Blackjack wants
       * more than the default drain -- it blurs, so a busted hand reads as out
       * of play at a glance -- and says so by replacing the one filter token
       * rather than declaring a private .busted class, which is how this repo
       * ended up with four unrelated "eliminated" treatments.
       */
      .player {
        --boardgame-state-eliminated-filter: saturate(0.5) blur(1px);
      }
    `
  ];

  override render() {
    return html`
      <boardgame-game-surface heading="Blackjack">
        <boardgame-game-outcome
          slot="status"
          .finished=${this.gameFinished}
          .animating=${this.animating}
          .winners=${this.gameWinners}
          .viewer=${this.viewingAsPlayer >= 0 ? this.viewingAsPlayer : null}>
        </boardgame-game-outcome>
        <div id="draw" class="horizontal center">
          <boardgame-component-zone
            label="Draw pile"
            .stack="${this.state?.Game?.DrawStack}"
            .componentView=${this.cards}
            layout="stack"
            messy>
          </boardgame-component-zone>
          <boardgame-action-bar class="flex" label="Blackjack actions">
            <boardgame-action-button .action=${this.move(MoveNames.CurrentPlayerHit)}>Hit</boardgame-action-button>
            <boardgame-action-button .action=${this.move(MoveNames.CurrentPlayerStand)}>Stand</boardgame-action-button>
          </boardgame-action-bar>
          <boardgame-component-zone
            label="Discard pile"
            .stack="${this.state?.Game?.DiscardStack}"
            .componentView=${this.cards}
            layout="stack"
            messy>
          </boardgame-component-zone>
        </div>
        <boardgame-player-grid id="players">
          ${repeat(this.state?.Players || [], (_player, index) => index, (player, index) => html`
            <boardgame-player-panel
                class=${classMap({ player: true, vertical: true, eliminated: player.Eliminated })}
                label=${`Player ${index + 1}`}
                .active=${index === this.currentPlayerIndex}>
              <boardgame-component-zone
                label="Hand"
                .stack="${player.Hand}"
                .componentView=${this.cards.withProperties({ rotated: true })}
                layout="fan"
                messy>
                <boardgame-fading-text .trigger="${player.Eliminated}" message="Busted!"></boardgame-fading-text>
                <boardgame-fading-text .trigger="${player.Stood}" message="Stand!"></boardgame-fading-text>
              </boardgame-component-zone>
            </boardgame-player-panel>
          `)}
        </boardgame-player-grid>
        <boardgame-turn-status
          slot="status"
          .turn=${this.turnStatus}>
        </boardgame-turn-status>
      </boardgame-game-surface>
    `;
  }
}
