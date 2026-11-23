// Stored pair: tictactoe.checkGameFinished (producer hand, extraction wrapped).
// From server/static/src/engine/games/tictactoe.ts, boardResult.
// boardResult(tokens), which returns an object where Go returns (finished,
// winner). Go also returns (false, Empty) when len(state) != totalDim; the
// hand port always passes nine tokens and does not check.
// Written by server/static/scripts/pairs-extract.mjs; rebuild the store, do not edit.

import type { RulesState } from '../../../../../../../server/static/src/engine/rules/state.ts';

const DIM = 3;

const EMPTY = '';

function runWinner(run: string[]): string {
	const target = run[0];
	if (target === EMPTY) return EMPTY;
	for (const token of run) if (token !== target) return EMPTY;
	return target;
}

function boardResult(tokens: string[]): { finished: boolean; winner: string } {
	const at = (row: number, col: number): string => tokens[row * DIM + col];
	const runs: string[][] = [];
	for (let r = 0; r < DIM; r++) runs.push([0, 1, 2].map(c => at(r, c)));
	for (let c = 0; c < DIM; c++) runs.push([0, 1, 2].map(r => at(r, c)));
	runs.push([0, 1, 2].map(i => at(i, i)));
	runs.push([0, 1, 2].map(i => at(DIM - i - 1, i)));
	for (const run of runs) {
		const winner = runWinner(run);
		if (winner !== EMPTY) return { finished: true, winner };
	}
	for (const token of tokens) if (token === EMPTY) return { finished: false, winner: EMPTY };
	return { finished: true, winner: EMPTY };
}

export function tictactoe_checkGameFinished(state: RulesState, state_: string[] | null): [boolean, string] {
	const result = boardResult(state_ ?? []);
	return [result.finished, result.winner];
}
