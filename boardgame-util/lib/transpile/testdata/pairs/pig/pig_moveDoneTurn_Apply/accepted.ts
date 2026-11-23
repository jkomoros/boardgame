// Stored pair: pig.(*moveDoneTurn).Apply (producer hand, extraction mechanical).
// From server/static/src/engine/games/pig.ts, moves["Done Turn"].apply.
// Written by server/static/scripts/pairs-extract.mjs; rebuild the store, do not edit.

import type { RulesState } from '../../../../../../../server/static/src/engine/rules/state.ts';
import type { MoveInstance } from '../../../../../../../server/static/src/engine/rules/types.ts';

const currentPlayer = (state: RulesState) => state.players[state.ensureValid(state.g.CurrentPlayer as number)];

export function pig_moveDoneTurn_Apply(state: RulesState, _m: MoveInstance): void | string | null {
	currentPlayer(state).Done = true;
}
