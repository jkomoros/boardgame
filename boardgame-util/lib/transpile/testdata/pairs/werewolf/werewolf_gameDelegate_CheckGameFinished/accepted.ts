// Stored pair: werewolf.(*gameDelegate).CheckGameFinished (producer hand, extraction mechanical).
// From server/static/src/engine/games/werewolf.ts, end.check.
// Written by server/static/scripts/pairs-extract.mjs; rebuild the store, do not edit.

import type { RulesState } from '../../../../../../../server/static/src/engine/rules/state.ts';

const PHASE_GATHERING = 0;

const WEREWOLF = 'Werewolf';

const isWerewolf = (state: RulesState, player: number): boolean => state.players[player].Role === WEREWOLF;

const eliminated = (state: RulesState, player: number): boolean => state.players[player].Eliminated === true;

export function werewolf_gameDelegate_CheckGameFinished(state: RulesState): [finished: boolean, winners: number[] | null] | { finished: boolean; winners: number[] } {
	if (state.currentPhaseKey() === PHASE_GATHERING) return { finished: false, winners: [] };
	let wolves = 0;
	let villagers = 0;
	for (let i = 0; i < state.numPlayers(); i++) {
		if (state.playerIsInactive(i) || eliminated(state, i)) continue;
		if (isWerewolf(state, i)) wolves++;
		else villagers++;
	}
	const villagersWin = wolves === 0;
	const werewolvesWin = !villagersWin && wolves >= villagers;
	if (!villagersWin && !werewolvesWin) return { finished: false, winners: [] };
	const winners: number[] = [];
	for (let i = 0; i < state.numPlayers(); i++) {
		if (state.playerIsInactive(i)) continue;
		if (villagersWin !== isWerewolf(state, i)) winners.push(i);
	}
	return { finished: true, winners };
}
