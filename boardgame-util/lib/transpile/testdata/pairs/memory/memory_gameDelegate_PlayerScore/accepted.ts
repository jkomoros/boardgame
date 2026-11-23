// Stored pair: memory.(*gameDelegate).PlayerScore (producer hand, extraction wrapped).
// From server/static/src/engine/games/memory.ts, player.traits.gameScore.
// the hand port leaves end.playerScore unset, so the engine scores with the
// gameScore trait (rules/engine.ts checkGameFinished); Go's PlayerScore
// returns pState.GameScore(). The trait's function is copied (the same text
// as (*playerState).GameScore).
// Written by server/static/scripts/pairs-extract.mjs; rebuild the store, do not edit.

import type { RulesState } from '../../../../../../../server/static/src/engine/rules/state.ts';

export function memory_gameDelegate_PlayerScore(state: RulesState, player: number): number {
	return state.playerView(player, 'WonCards').numComponents();
}
