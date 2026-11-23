// Stored pair: blackjack.(*moveAccumulateScores).Legal (producer hand, extraction mechanical).
// From server/static/src/engine/games/blackjack.ts, moves["Accumulate Scores"].legal.
// Written by server/static/scripts/pairs-extract.mjs; rebuild the store, do not edit.

import type { RulesState } from '../../../../../../../server/static/src/engine/rules/state.ts';
import type { MoveInstance } from '../../../../../../../server/static/src/engine/rules/types.ts';

export function blackjack_moveAccumulateScores_Legal(state: RulesState, _move: MoveInstance, _proposer: number, base: () => string | null): string | null {
	const err = base();
	if (err !== null) return err;
	for (let i = 0; i < state.numPlayers(); i++) {
		if (state.playerIsInactive(i)) continue;
		if (state.playerView(i, 'Hand').numComponents() > 0) return null;
	}
	return 'no active players have cards to score';
}
