// Stored pair: blackjack.(*moveIncrementRoundsCompleted).Legal (producer hand, extraction mechanical).
// From server/static/src/engine/games/blackjack.ts, moves["Increment Rounds Completed"].legal.
// Written by server/static/scripts/pairs-extract.mjs; rebuild the store, do not edit.

import type { RulesState } from '../../../../../../../server/static/src/engine/rules/state.ts';
import type { MoveInstance } from '../../../../../../../server/static/src/engine/rules/types.ts';

const PHASE_ROUND_CLEANUP = 3;

const hasCards = (state: RulesState, player: number): boolean =>
	state.playerStack(player, 'HiddenHand').numComponents() > 0 || state.playerStack(player, 'VisibleHand').numComponents() > 0;

export function blackjack_moveIncrementRoundsCompleted_Legal(state: RulesState, _move: MoveInstance, _proposer: number, base: () => string | null): string | null {
	const err = base();
	if (err !== null) return err;
	for (let i = 0; i < state.numPlayers(); i++) if (hasCards(state, i)) return "cards haven't been collected yet";
	if (state.currentPhaseKey() !== PHASE_ROUND_CLEANUP) return 'not in cleanup phase';
	return null;
}
