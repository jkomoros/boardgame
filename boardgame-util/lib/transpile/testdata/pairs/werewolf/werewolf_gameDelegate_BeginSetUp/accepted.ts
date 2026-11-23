// Stored pair: werewolf.(*gameDelegate).BeginSetUp (producer hand, extraction mechanical).
// From server/static/src/engine/games/werewolf.ts, setUp.begin.
// Written by server/static/scripts/pairs-extract.mjs; rebuild the store, do not edit.

import type { RulesState } from '../../../../../../../server/static/src/engine/rules/state.ts';

export function werewolf_gameDelegate_BeginSetUp(state: RulesState, _variant: Readonly<Record<string, string>>): void | string | null {
	for (const player of state.players) {
		player.DayVote = -1;
		player.NightVote = -1;
	}
}
