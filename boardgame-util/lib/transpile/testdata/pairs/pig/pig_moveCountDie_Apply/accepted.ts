// Stored pair: pig.(*moveCountDie).Apply (producer hand, extraction mechanical).
// From server/static/src/engine/games/pig.ts, moves["Count Die"].apply.
// Written by server/static/scripts/pairs-extract.mjs; rebuild the store, do not edit.

import type { RulesState } from '../../../../../../../server/static/src/engine/rules/state.ts';
import type { MoveInstance } from '../../../../../../../server/static/src/engine/rules/types.ts';

const DICE = 'dice';

const currentPlayer = (state: RulesState) => state.players[state.ensureValid(state.g.CurrentPlayer as number)];

export function pig_moveCountDie_Apply(state: RulesState, _m: MoveInstance): void | string | null {
	const player = currentPlayer(state);
	const value = state.dynamic(DICE, 0).Value as number;
	// A one is a bust: the turn's points are lost.
	if (value === 1) player.Eliminated = true;
	else player.RoundScore = (player.RoundScore as number) + value;
	player.DieCounted = true;
}
