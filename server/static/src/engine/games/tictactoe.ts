// Tic Tac Toe. A move-for-move port of examples/tictactoe. It is the example
// of a move whose name and argument only its proposer may see.

import type { RulesState } from '../rules/state.ts';
import type { GameSpec, MoveSpec } from '../rules/types.ts';
import { activateInactivePlayer, add, currentPlayerProps, finishTurn, inactivePlayerProps, moveComponentToSlot, phaseProps, seatPlayer, seatProps } from '../rules/stock.ts';

const DIM = 3;
// DIM * DIM, written out: a pattern module may not compute at module scope.
const TOTAL_DIM = 9;
const X = 'X';
const O = 'O';
const EMPTY = '';
const PHASE_AFTER_FIRST_MOVE = 1;
const TOKENS_PER_PLAYER = 5;

function newDeck(): { Value: string }[] {
	const tokens: { Value: string }[] = [];
	for (let i = 0; i < TOKENS_PER_PLAYER; i++) tokens.push({ Value: X });
	for (let i = 0; i < TOKENS_PER_PLAYER; i++) tokens.push({ Value: O });
	return tokens;
}

function runWinner(run: string[]): string {
	const target = run[0];
	if (target === EMPTY) return EMPTY;
	for (const token of run) if (token !== target) return EMPTY;
	return target;
}

/** checkGameFinished: rows, columns, both diagonals, then a full board is a draw. */
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

// "Play the first of my unused tokens into the slot I chose." The source is
// a stack on the player the move names.
function place() {
	return moveComponentToSlot({ source: 'players[move.TargetPlayerIndex].UnusedTokens', destination: 'game.Slots', slotField: 'Slot' });
}

function placeToken(): MoveSpec {
	return {
		name: 'Place Token',
		helpText: "Place a player's token in a specific space.",
		fixUp: false,
		...place(),
		// Only the proposer sees the move's name and where the token went;
		// everyone else gets an opaque "Hidden Action" and reads the result from
		// the state.
		nameSanitization: 'self:visible',
		hiddenAnimationKey: 'Hidden Action',
		fields: { Slot: { type: 'int', sanitize: 'self:visible' } },
		defaults(state, move) {
			const slots = state.gameStack('Slots');
			for (let i = 0; i < slots.len(); i++) {
				if (slots.componentAt(i) !== null) continue;
				move.fields.Slot = i;
				break;
			}
		},
		apply(state, move) {
			place().apply(state, move);
			const player = state.players[state.ensureValid(move.fields.TargetPlayerIndex as number)];
			player.TokensToPlaceThisTurn = (player.TokensToPlaceThisTurn as number) - 1;
			state.setPhase(PHASE_AFTER_FIRST_MOVE);
		},
	};
}

export function tictactoe(): GameSpec {
	return {
		name: 'tictactoe',
		displayName: 'Tic Tac Toe',
		minNumPlayers: 2,
		maxNumPlayers: 2,
		defaultNumPlayers: 2,

		decks: { tokens: { components: newDeck() } },
		enums: { phase: { Values: { 0: 'Before First Move', 1: 'After First Move' }, DefaultValue: 0 } },
		constants: { TOTAL_DIM },

		game: {
			props: {
				...currentPlayerProps,
				Slots: { Type: 'sizedStack', Deck: 'tokens', Size: TOTAL_DIM },
				...phaseProps,
			},
		},
		player: {
			props: {
				...seatProps,
				...inactivePlayerProps,
				TokenValue: { Type: 'string' },
				// Go constrains this stack to tokens of one value (same(Value)).
				// Stack constraints are not ported; nothing ever moves in here
				// after set-up, so the constraint is never consulted.
				UnusedTokens: { Type: 'stack', Deck: 'tokens' },
				TokensToPlaceThisTurn: { Type: 'int' },
			},
			initial: player => ({ TokensToPlaceThisTurn: 1, TokenValue: player === 1 ? O : X }),
			traits: {
				seat: true,
				inactive: true,
				turn: {
					turnDone: (state, player) =>
						(state.players[player].TokensToPlaceThisTurn as number) > 0 ? 'they still have tokens left to place this turn' : null,
					resetForTurnStart(state, player) {
						state.players[player].TokensToPlaceThisTurn = 1;
					},
					resetForTurnEnd: () => undefined,
				},
			},
		},

		// The one failure message this game rewords.
		legalTemplates: { 'legal.no_component_to_move': "there aren't any remaining tokens for the current player to place" },

		moves: add(seatPlayer(), activateInactivePlayer(), placeToken(), finishTurn()),

		setUp: {
			distribute: (state, _deck, index) => state.playerStack(state.component('tokens', index).Value === X ? 0 : 1, 'UnusedTokens'),
		},

		end: {
			check(state) {
				const { finished, winner } = boardResult(boardTokens(state));
				if (!finished) return { finished: false, winners: [] };
				if (winner === EMPTY) return { finished: true, winners: [] };
				let winningPlayer = 0;
				for (let i = 0; i < state.numPlayers(); i++) if (state.players[i].TokenValue === winner) winningPlayer = i;
				return { finished: true, winners: [winningPlayer] };
			},
		},
	};
}
