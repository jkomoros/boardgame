// Stored pair: pig.(*moveRollDice).Apply (producer hand, extraction mechanical).
// From server/static/src/engine/games/pig.ts, moves["Roll Dice"].apply.
// Written by server/static/scripts/pairs-extract.mjs; rebuild the store, do not edit.

import type { RulesState } from '../../../../../../../server/static/src/engine/rules/state.ts';
import type { MoveInstance } from '../../../../../../../server/static/src/engine/rules/types.ts';

const DICE = 'dice';

const currentPlayer = (state: RulesState) => state.players[state.ensureValid(state.g.CurrentPlayer as number)];

export function pig_moveRollDice_Apply(state: RulesState, _m: MoveInstance): void | string | null {
	const player = currentPlayer(state);
	// dice.DynamicValue.Roll: one draw from the state's generator.
	const faces = state.component(DICE, 0).Faces as number[];
	const die = state.dynamic(DICE, 0);
	const face = state.rand().intn(faces.length);
	die.SelectedFace = face;
	die.Value = faces[face];
	die.RollCount = (die.RollCount as number) + 1;
	player.DieCounted = false;
}
