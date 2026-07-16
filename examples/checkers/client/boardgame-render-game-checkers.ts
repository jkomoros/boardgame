import { GameRenderer } from './_game_renderer.js';
import '../../src/components/boardgame-token.js';
import '../../src/components/boardgame-game-board.js';
import '../../src/components/boardgame-fading-text.js';
import { html, css, isVisibleComponent, SourceDestinationController } from '../../src/client.js';
import { MoveNames } from './_move_names.js';

class BoardgameRenderGameCheckers extends GameRenderer {
  static override styles = [
    ...(GameRenderer.styles ? [GameRenderer.styles] : []),
    css`
      boardgame-game-board {
        max-width: 500px;
        margin: 0 auto;
      }
    `
  ];

  private readonly moveToken = new SourceDestinationController<number>(this);

  override render() {
    const spaces = this.state?.Game?.Spaces ?? null;
    const components = spaces?.Components ?? [];
    const actingPlayer = this.proposingAsPlayer >= 0 ? this.proposingAsPlayer : this.viewingAsPlayer;
    const playerColor = this.state?.Players[actingPlayer]?.Color;
    const interaction = this.moveToken.bind({
      sources: components.flatMap((component, index) =>
        isVisibleComponent(component)
          && (this.proposingAsAdmin || component.Values.Color === playerColor)
          ? [index]
          : []),
      destinations: TokenIndexToMove => this.move(MoveNames.MoveToken).targets(
        components.flatMap((component, SpaceIndex) => component ? [] : [SpaceIndex]),
        SpaceIndex => ({ TokenIndexToMove, SpaceIndex }),
      ),
    });
    return html`
      <boardgame-deck-defaults>
        <template deck="tokens">
          <boardgame-token type="disc" color="{{item.Values.Color}}"></boardgame-token>
        </template>
      </boardgame-deck-defaults>
      <boardgame-game-board
        rows="8" cols="8" checkerboard
        .stack="${spaces}"
        .sourceDestination=${interaction}>
      </boardgame-game-board>
      <boardgame-fading-text
        .trigger="${this.isCurrentPlayer}"
        message="Your Turn"
        suppress="falsey">
      </boardgame-fading-text>
    `;
  }
}

customElements.define('boardgame-render-game-checkers', BoardgameRenderGameCheckers);
