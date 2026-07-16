import { GameRenderer, registerGameRenderer } from './_game_renderer.js';
import { html, css, isVisibleComponent, SourceDestinationController, tokenView } from '../../src/client.js';
import { MoveNames } from './_move_names.js';
import type { GameState } from './_types.js';

@registerGameRenderer
export class BoardgameRenderGameCheckers extends GameRenderer {
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
  private readonly tokens = tokenView<GameState['Spaces']>({
    properties: ({ kind, component }) => ({
      type: 'disc',
      color: kind === 'visible' ? component.Values.Color : '',
    }),
  });

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
      <boardgame-game-surface heading="Checkers">
        <boardgame-game-board
          rows="8" cols="8" checkerboard
          .stack="${spaces}"
          .componentView=${this.tokens}
          .sourceDestination=${interaction}>
        </boardgame-game-board>
        <boardgame-turn-status
          slot="status"
          .turn=${this.turnStatus}>
        </boardgame-turn-status>
      </boardgame-game-surface>
    `;
  }
}
