// Stored pair: blackjack.(*moveCurrentPlayerStand).Apply (producer hand, extraction mechanical).
// From server/static/src/engine/games/blackjack.ts, moves["Current Player Stand"].apply.
// Written by server/static/scripts/pairs-extract.mjs; rebuild the store, do not edit.

import type { RulesState } from '../../../../../../../server/static/src/engine/rules/state.ts';
import type { MoveInstance } from '../../../../../../../server/static/src/engine/rules/types.ts';

export function blackjack_moveCurrentPlayerStand_Apply(state: RulesState, _m: MoveInstance): void | string | null {
	state.players[state.ensureValid(state.g.CurrentPlayer as number)].Stood = true;
}
