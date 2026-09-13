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

export class TransitionRenderer extends BoardgameBaseGameRenderer<State, object, 'Play', { Play: object },
  object, object, Record<never, never>, 'Play' | 'Retire' | 'Hidden Action'> {
  override effectsForTransition(context: EffectTransitionContext<State, 'Play' | 'Retire' | 'Hidden Action'>) {
    if (context.move?.AnimationKey === 'Retire') {
      // @ts-expect-error Transition-only vocabulary never becomes proposable inside a hook.
      this.move(context.move.AnimationKey);
    }
    return [];
  }
}
