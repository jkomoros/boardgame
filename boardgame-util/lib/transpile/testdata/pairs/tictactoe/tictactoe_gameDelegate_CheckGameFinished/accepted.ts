// Stored pair: tictactoe.(*gameDelegate).CheckGameFinished (producer hand, extraction mechanical).
// From server/static/src/engine/games/tictactoe.ts, end.check.
// Written by server/static/scripts/pairs-extract.mjs; rebuild the store, do not edit.

import type { RulesState } from '../../../../../../../server/static/src/engine/rules/state.ts';

const DIM = 3;

const TOTAL_DIM = 9;

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

function boardTokens(state: RulesState): string[] {
	const slots = state.gameStack('Slots');
	const tokens: string[] = [];
	for (let i = 0; i < TOTAL_DIM; i++) {
		const index = slots.componentAt(i);
		tokens.push(index === null || index < 0 ? EMPTY : (state.component('tokens', index).Value as string));
	}
	return tokens;
}

export function tictactoe_gameDelegate_CheckGameFinished(state: RulesState): [finished: boolean, winners: number[] | null] | { finished: boolean; winners: number[] } {
	const { finished, winner } = boardResult(boardTokens(state));
	if (!finished) return { finished: false, winners: [] };
	if (winner === EMPTY) return { finished: true, winners: [] };
	let winningPlayer = 0;
	for (let i = 0; i < state.numPlayers(); i++) if (state.players[i].TokenValue === winner) winningPlayer = i;
	return { finished: true, winners: [winningPlayer] };
}
