import {
  BoardgameBaseGameRenderer,
  type EffectTransitionContext,
  type FullGameState,
} from '../client.js';

type State = FullGameState<object, object>;
type Renderer = BoardgameBaseGameRenderer<State, object, 'Play', { Play: object },
  object, object, Record<never, never>, 'Play' | 'Retire' | 'Hidden Action'>;

export function checkTransitionKeys(renderer: Renderer, context: EffectTransitionContext<State, 'Play' | 'Retire' | 'Hidden Action'>): void {
  renderer.move('Play');
  // @ts-expect-error A fixup animation key is not a player-proposable move.
  renderer.move('Retire');
  // @ts-expect-error A sanitized alias is not a player-proposable move.
  renderer.move('Hidden Action');
  renderer.motionCohortsForTransition(context);
}
