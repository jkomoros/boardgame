// Stored pair: blackjack.(*moveResetPlayerForNewRound).Legal (producer hand, extraction mechanical).
// From server/static/src/engine/games/blackjack.ts, moves["Reset Player For New Round"].legal.
// Written by server/static/scripts/pairs-extract.mjs; rebuild the store, do not edit.

import type { RulesState } from '../../../../../../../server/static/src/engine/rules/state.ts';
import type { MoveInstance } from '../../../../../../../server/static/src/engine/rules/types.ts';

export function blackjack_moveResetPlayerForNewRound_Legal(state: RulesState, _move: MoveInstance, _proposer: number, base: () => string | null): string | null {
	const err = base();
	if (err !== null) return err;
	for (let i = 0; i < state.numPlayers(); i++) {
		if (state.playerIsInactive(i)) continue;
		if (state.players[i].Eliminated === true || state.players[i].Stood === true) return null;
	}
	return 'no active players need resetting';
}
