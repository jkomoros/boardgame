import { css, html, cardView } from '../../src/client.js';
import { GameRenderer } from './_game_renderer.js';
import type { GameState } from './_types.js';
import '../../src/components/boardgame-component-stack.js';
import '../../src/components/boardgame-fading-text.js';
import { MoveNames } from './_move_names.js';

class BoardgameRenderGameCheckers extends GameRenderer {
  private readonly cards = cardView<GameState['DrawStack']>({
    render: ({ kind, component }) => kind === 'visible'
      ? html`<strong>${component.Values.Value}</strong>`
      : null,
  });
  static override styles = css`
    :host { display: block; }
    .players { display: flex; flex-wrap: wrap; gap: 1rem; }
    .player { flex: 1 1 12rem; }
  `;

  override render() {
    return html`
      <boardgame-component-stack
        .stack=${this.state?.Game.DrawStack ?? null}
        .componentView=${this.cards}
        layout="stack" messy>
      </boardgame-component-stack>
      <boardgame-action-button .action=${this.move(MoveNames.DrawCard)}>
        Draw a card
      </boardgame-action-button>
      <div class="players">
        ${this.state?.Players.map((player, index) => html`
          <section class="player">
            <strong>Player ${index + 1}</strong>
            <boardgame-component-stack
              .stack=${player.Hand}
              .componentView=${this.cards}
              layout="fan" messy
              .componentAttrs=${{ rotated: true }}>
            </boardgame-component-stack>
            <boardgame-fading-text
              .trigger=${player.Computed?.GameScore ?? 0}
              auto-message="diff-up">
            </boardgame-fading-text>
          </section>
        `)}
      </div>
      <boardgame-fading-text
        .trigger=${this.isCurrentPlayer}
        message="Your Turn" suppress="falsey">
      </boardgame-fading-text>
    `;
  }
}

customElements.define('boardgame-render-game-checkers', BoardgameRenderGameCheckers);
