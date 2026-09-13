import { GameRenderer, registerGameRenderer } from './_game_renderer.js';
import {
  diffVisibleComponents,
  fx,
  html,
  css,
  isVisibleComponent,
  SourceDestinationController,
  tokenView,
} from '../../src/client.js';
import type { EffectSpec, EffectTransitionContext } from '../../src/client.js';
import { MoveNames } from './_move_names.js';
import type { AnimationKey } from './_move_names.js';
import type { GameState, State } from './_types.js';

type VisibleChecker = Extract<
  NonNullable<GameState['Spaces']['Components'][number]>,
  { readonly ID: string }
>;

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
      type: kind === 'visible' && component.DynamicValues?.Crowned ? 'token' : 'disc',
      color: kind === 'visible' ? component.Values.Color : '',
    }),
  });

  override effectsForTransition(
    context: EffectTransitionContext<State, AnimationKey>,
  ): readonly EffectSpec[] {
    if (context.kind === 'initial') return [];
    const membership = diffVisibleComponents(
      context.before.Game.Spaces.Components,
      context.after.Game.Spaces.Components,
    );
    if (membership.status !== 'exact') return [];

    const beforeById = new Map<string, VisibleChecker>();
    for (const component of context.before.Game.Spaces.Components) {
      if (isVisibleComponent(component)) beforeById.set(component.ID, component);
    }
    const crowned = context.after.Game.Spaces.Components.flatMap(component => {
      if (!isVisibleComponent(component)) return [];
      const before = beforeById.get(component.ID);
      return before && component.DynamicValues?.Crowned === true
        && before.DynamicValues?.Crowned !== true ? [component.ID] : [];
    });
    if (crowned.length !== 1) return [];

    return [fx.pulse({
      at: fx.subject(crowned[0]!),
      tone: 'reward',
      intensity: 'small',
      timing: 'version',
      key: 'crown-token',
    })];
  }

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
