import { css, html } from '../../src/client.js';
import { GameRenderer } from './_game_renderer.js';
import '../../src/components/boardgame-component-stack.js';
import '../../src/components/boardgame-fading-text.js';
import { MoveNames } from './_move_names.js';

class BoardgameRenderGameCheckers extends GameRenderer {
  static override styles = css`
    :host { display: block; }
    .players { display: flex; flex-wrap: wrap; gap: 1rem; }
    .player { flex: 1 1 12rem; }
  `;

  override render() {
    return html`
      <boardgame-component-stack
        .stack=${this.state?.Game.DrawStack}
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
              layout="fan" messy component-rotated>
            </boardgame-component-stack>
            <boardgame-fading-text
              .trigger=${player.Computed.GameScore}
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
