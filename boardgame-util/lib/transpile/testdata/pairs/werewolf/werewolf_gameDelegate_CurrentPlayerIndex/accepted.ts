// Stored pair: werewolf.(*gameDelegate).CurrentPlayerIndex (producer hand, extraction mechanical).
// From server/static/src/engine/games/werewolf.ts, currentPlayerIndex.
// Written by server/static/scripts/pairs-extract.mjs; rebuild the store, do not edit.

import { ANY_PLAYER_INDEX, OBSERVER_PLAYER_INDEX } from '../../../../../../../server/static/src/engine/legal/types.ts';
import type { RulesState } from '../../../../../../../server/static/src/engine/rules/state.ts';

const PHASE_DAY = 1;

const PHASE_NIGHT = 2;

export function werewolf_gameDelegate_CurrentPlayerIndex(state: RulesState): number {
	const phase = state.currentPhaseKey();
	return phase === PHASE_DAY || phase === PHASE_NIGHT ? ANY_PLAYER_INDEX : OBSERVER_PLAYER_INDEX;
}
