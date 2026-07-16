import { css, html, cardView } from '../../src/client.js';
import { GameRenderer, registerGameRenderer } from './_game_renderer.js';
import type { GameState } from './_types.js';
import { MoveNames } from './_move_names.js';

@registerGameRenderer
export class BoardgameRenderGameCheckers extends GameRenderer {
  private readonly cards = cardView<GameState['DrawStack']>({
    render: ({ kind, component }) => kind === 'visible'
      ? html`<strong>${component.Values.Value}</strong>`
      : null,
  });
  static override styles = css`
    :host { display: block; }
  `;

  override render() {
    return html`
      <boardgame-game-outcome
        .finished=${this.gameFinished}
        .animating=${this.animating}
        .winners=${this.gameWinners}
        .viewer=${this.viewingAsPlayer >= 0 ? this.viewingAsPlayer : null}>
      </boardgame-game-outcome>
      <boardgame-component-zone
        label="Draw pile"
        .stack=${this.state?.Game.DrawStack ?? null}
        .componentView=${this.cards}
        layout="stack" messy>
      </boardgame-component-zone>
      <boardgame-action-bar label="Turn actions">
        <boardgame-action-button .action=${this.move(MoveNames.DrawCard)}>
          Draw a card
        </boardgame-action-button>
      </boardgame-action-bar>
      <boardgame-player-grid>
        ${this.state?.Players.map((player, index) => html`
          <boardgame-component-zone
              class="player"
              label="Player ${index + 1} hand"
              .stack=${player.Hand}
              .componentView=${this.cards.withProperties({ rotated: true })}
              layout="fan" messy>
            <boardgame-fading-text
              .trigger=${player.Computed?.RoundScore ?? 0}
              auto-message="diff-up">
            </boardgame-fading-text>
          </boardgame-component-zone>
        `)}
      </boardgame-player-grid>
      <boardgame-fading-text
        .trigger=${this.isCurrentPlayer}
        message="Your Turn" suppress="falsey">
      </boardgame-fading-text>
    `;
  }
}
