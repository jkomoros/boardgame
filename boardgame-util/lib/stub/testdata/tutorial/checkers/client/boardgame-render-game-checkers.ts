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
  //Spreading GameRenderer.styles first is what brings in the shared
  //layout/state vocabulary (.horizontal, .vertical, .center, .flex, .active,
  //.responding, .selected, .targetable, .disabled, .eliminated) -- assigning
  //css`` directly to static styles would REPLACE it.
  static override styles = [
    ...(GameRenderer.styles ? [GameRenderer.styles] : []),
    css`
      :host { display: block; }
    `,
  ];

  override render() {
    return html`
      <boardgame-game-surface heading="Checkers">
        <boardgame-game-outcome
          slot="status"
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
        <boardgame-action-bar slot="actions" label="Turn actions">
          <boardgame-action-button .action=${this.move(MoveNames.DrawCard)}>
            Draw a card
          </boardgame-action-button>
        </boardgame-action-bar>
        <boardgame-player-grid>
          ${this.state?.Players.map((player, index) => html`
            <boardgame-player-panel
                class="player"
                label=${`Player ${index + 1}`}
                .active=${index === this.currentPlayerIndex}>
              <boardgame-component-zone
                  label="Hand"
                  .stack=${player.Hand}
                  .componentView=${this.cards.withProperties({ rotated: true })}
                  layout="fan" messy>
                <boardgame-fading-text
                  .trigger=${player.Computed?.RoundScore ?? 0}
                  auto-message="diff-up">
                </boardgame-fading-text>
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
