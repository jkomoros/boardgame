import type { ExpandedStack } from '../types/boardgame-types.js';
import {
  projectedPlayerChoices,
  projectedStackChoices,
  type ProjectedMoveChoiceSet,
} from './projected-choices.js';

type CardProjection = { readonly field: 'TargetCard'; readonly value: number; readonly input: { TargetCard: number } };
type PlayerProjection = { readonly field: 'TargetPlayer'; readonly value: number; readonly input: { TargetPlayer: number } };
declare const cards: ProjectedMoveChoiceSet<'Choose Card', CardProjection>;
declare const players: ProjectedMoveChoiceSet<'Choose Player', PlayerProjection>;
declare const hand: ExpandedStack;

const stackBinding = projectedStackChoices(cards, hand);
const playerBinding = projectedPlayerChoices(players, playerIndex => `Player ${playerIndex + 1}`);
stackBinding.actions[0]?.activate();
playerBinding.choices[0]?.action.activate();

// @ts-expect-error player labels are indexed numerically
projectedPlayerChoices(players, (playerIndex: string) => playerIndex);
// @ts-expect-error stack projections have numeric candidate values
projectedStackChoices({} as ProjectedMoveChoiceSet<'Choose', {
  readonly field: 'Value'; readonly value: string; readonly input: { Value: string };
}>, hand);
