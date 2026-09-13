import { GameRenderer, registerGameRenderer } from './_game_renderer.js';
import { html, css } from 'lit';
import { MoveNames } from './_move_names.js';
import type { MoveName } from './_move_names.js';
import type { GameState, State } from './_types.js';
import {
  diffVisibleComponents,
  fx,
  isVisibleComponent,
  tokenView,
} from '../../src/client.js';
import type { EffectSpec, EffectTransitionContext } from '../../src/client.js';

const WINNING_LINES = [
  [0, 1, 2], [3, 4, 5], [6, 7, 8],
  [0, 3, 6], [1, 4, 7], [2, 5, 8],
  [0, 4, 8], [2, 4, 6],
] as const;

function winningTokenIds(slots: GameState['Slots']): ReadonlySet<string> {
  const result = new Set<string>();
  for (const line of WINNING_LINES) {
    const first = slots.Components[line[0]];
    const second = slots.Components[line[1]];
    const third = slots.Components[line[2]];
    if (!isVisibleComponent(first) || !isVisibleComponent(second)
      || !isVisibleComponent(third)) continue;
    if (first.Values.Value === second.Values.Value && first.Values.Value === third.Values.Value) {
      result.add(first.ID);
      result.add(second.ID);
      result.add(third.ID);
    }
  }
  return result;
}

@registerGameRenderer
export class BoardgameRenderGameTictactoe extends GameRenderer {
  private readonly tokens = tokenView<GameState['Slots']>({
    properties: ({ kind, component }) => ({
      type: kind === 'visible'
        && this.state?.Game?.Slots
        && winningTokenIds(this.state.Game.Slots).has(component.ID)
        ? 'token'
        : 'chip',
    }),
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

  override effectsForTransition(
    context: EffectTransitionContext<State, MoveName>,
  ): readonly EffectSpec[] {
    if (context.kind === 'initial') return [];
    const winners = winningTokenIds(context.after.Game.Slots);
    if (winners.size === 0) return [];

    const diff = diffVisibleComponents(
      context.before.Game.Slots.Components,
      context.after.Game.Slots.Components,
    );
    if (diff.status !== 'exact' || diff.added.length !== 1) return [];
    const placed = diff.added[0];
    if (!placed || !winners.has(placed)) return [];

    return [fx.pulse({
      at: fx.subject(placed),
      tone: 'reward',
      intensity: 'small',
      timing: 'version',
      key: 'winning-line',
    })];
  }

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
