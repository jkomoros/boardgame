// Stored pair: memory.(*playerState).GameScore (producer hand, extraction mechanical).
// From server/static/src/engine/games/memory.ts, player.traits.gameScore.
// Written by server/static/scripts/pairs-extract.mjs; rebuild the store, do not edit.

import type { RulesState } from '../../../../../../../server/static/src/engine/rules/state.ts';

export function memory_playerState_GameScore(state: RulesState, player: number): number {
	return state.playerView(player, 'WonCards').numComponents();
}
