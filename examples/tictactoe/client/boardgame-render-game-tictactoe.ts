import { GameRenderer, registerGameRenderer } from './_game_renderer.js';
import { html, css } from 'lit';
import { MoveNames } from './_move_names.js';
import type { GameState } from './_types.js';
import { tokenView } from '../../src/client.js';

@registerGameRenderer
export class BoardgameRenderGameTictactoe extends GameRenderer {
  private readonly tokens = tokenView<GameState['Slots']>({
    properties: () => ({ type: 'chip' }),
  });

  static override styles = [
    ...(GameRenderer.styles ? [GameRenderer.styles] : []),
    css`
      boardgame-game-board {
        max-width: 320px;
        margin: 0 auto;
      }
    `
  ];

  override render() {
    const slots = this.state?.Game?.Slots;
    const places = slots
      ? this.move(MoveNames.PlaceToken).targets(
        slots.Components.map((_, slot) => slot),
        slot => ({ Slot: slot }),
      )
      : null;
    return html`
      <boardgame-game-surface heading="Tic-tac-toe">
        <boardgame-game-board
          rows="3" cols="3"
          .stack=${slots ?? null}
          .componentView=${this.tokens}
          .action=${places}>
        </boardgame-game-board>
        <boardgame-turn-status
          slot="status"
          .turn=${this.turnStatus}>
        </boardgame-turn-status>
      </boardgame-game-surface>
    `;
  }
}
