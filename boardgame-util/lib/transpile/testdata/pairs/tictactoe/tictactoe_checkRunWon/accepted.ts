// Stored pair: tictactoe.checkRunWon (producer hand, extraction wrapped).
// From server/static/src/engine/games/tictactoe.ts, runWinner.
// runWinner(run). Go first returns Empty when len(runState) != dim; the hand
// port has no such check (every run it builds has three tokens), so the
// wrapper does not add one.
// Written by server/static/scripts/pairs-extract.mjs; rebuild the store, do not edit.

import type { RulesState } from '../../../../../../../server/static/src/engine/rules/state.ts';

const EMPTY = '';

function runWinner(run: string[]): string {
	const target = run[0];
	if (target === EMPTY) return EMPTY;
	for (const token of run) if (token !== target) return EMPTY;
	return target;
}

export function tictactoe_checkRunWon(state: RulesState, runState: string[] | null): string {
	return runWinner(runState ?? []);
}
