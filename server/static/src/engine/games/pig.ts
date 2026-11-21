// Pig: roll a die to build a turn score, bank it before rolling a one. A
// move-for-move port of examples/pig. It is the example of dynamic component
// values (the die's face) and of randomness drawn while a move applies.

import { playerBool, playerBoolIs } from '../legal/catalogCompare.ts';
import type { RulesState } from '../rules/state.ts';
import type { GameSpec, MoveSpec } from '../rules/types.ts';
import {
	activateInactivePlayer, add, currentPlayerProps, eliminationProps, finishTurn, inactivePlayerProps, scoreGameScore, scoreProps, seatPlayer, seatProps,
} from '../rules/stock.ts';

const DEFAULT_TARGET_SCORE = 100;
const DICE = 'dice';

const currentPlayer = (state: RulesState) => state.players[state.ensureValid(state.g.CurrentPlayer as number)];

function resetForTurn(state: RulesState, player: number): void {
	const p = state.players[player];
	p.Done = false;
	p.Eliminated = false;
	p.RoundScore = 0;
	p.DieCounted = true;
}

const rollDice: MoveSpec = {
	name: 'Roll Dice',
	helpText: 'Rolls the dice for the current player',
	fixUp: false,
	base: 'currentPlayer',
	preconditions: [{ ...playerBool('DieCounted'), message: 'pig.roll_not_counted' }],
	apply(state) {
		const player = currentPlayer(state);
		// dice.DynamicValue.Roll: one draw from the state's generator.
		const faces = state.component(DICE, 0).Faces as number[];
		const die = state.dynamic(DICE, 0);
		const face = state.rand().intn(faces.length);
		die.SelectedFace = face;
		die.Value = faces[face];
		die.RollCount = (die.RollCount as number) + 1;
		player.DieCounted = false;
	},
};

const doneTurn: MoveSpec = {
	name: 'Done Turn',
	helpText: 'Played when a player is done with their turn and wants to keep their score.',
	fixUp: false,
	base: 'currentPlayer',
	preconditions: [
		{ ...playerBool('DieCounted'), message: 'pig.done_roll_not_counted' },
		{ ...playerBoolIs('Done', false), message: 'pig.already_done' },
	],
	apply(state) {
		currentPlayer(state).Done = true;
	},
};

const countDie: MoveSpec = {
	name: 'Count Die',
	helpText: 'After a die has been rolled, tabulating its impact',
	fixUp: true,
	base: 'currentPlayer',
	preconditions: [{ ...playerBoolIs('DieCounted', false), message: 'pig.roll_already_counted' }],
	apply(state) {
		const player = currentPlayer(state);
		const value = state.dynamic(DICE, 0).Value as number;
		// A one is a bust: the turn's points are lost.
		if (value === 1) player.Eliminated = true;
		else player.RoundScore = (player.RoundScore as number) + value;
		player.DieCounted = true;
	},
};

export const pig: GameSpec = {
	name: 'pig',
	minNumPlayers: 2,
	maxNumPlayers: 6,
	defaultNumPlayers: 2,

	decks: { [DICE]: { components: [{ Faces: [1, 2, 3, 4, 5, 6] }] } },

	game: {
		props: {
			...currentPlayerProps,
			Die: { Type: 'sizedStack', Deck: DICE, Size: 1 },
			TargetScore: { Type: 'int' },
		},
		initial: () => ({ TargetScore: DEFAULT_TARGET_SCORE }),
	},
	player: {
		props: {
			...seatProps,
			...inactivePlayerProps,
			...scoreProps,
			...eliminationProps,
			Done: { Type: 'bool' },
			DieCounted: { Type: 'bool' },
			RoundScore: { Type: 'int' },
		},
		traits: {
			seat: true,
			inactive: true,
			elimination: true,
			gameScore: scoreGameScore,
			turn: {
				turnDone(state, player) {
					const p = state.players[player];
					if (p.DieCounted !== true) return 'the most recent die roll has not been counted';
					if (p.Eliminated !== true && p.Done !== true) return 'they have not either busted or signaled that they are done';
					return null;
				},
				resetForTurnStart: resetForTurn,
				resetForTurnEnd(state, player) {
					const p = state.players[player];
					if (p.Done === true) p.Score = (p.Score as number) + (p.RoundScore as number);
					resetForTurn(state, player);
				},
			},
		},
	},
	// The die's face is a value of the component, not of the game state.
	dynamicComponentValues: {
		[DICE]: {
			props: { Value: { Type: 'int' }, SelectedFace: { Type: 'int' }, RollCount: { Type: 'int' } },
			initial: () => ({ Value: 1 }),
		},
	},

	legalTemplates: {
		'pig.roll_not_counted': 'Your most recent roll has not yet been counted',
		'pig.done_roll_not_counted': 'your most recent roll has not yet been counted',
		'pig.already_done': 'you already signaled that you are done',
		'pig.roll_already_counted': 'the most recent die roll has already been counted',
	},

	moves: add(
		seatPlayer(),
		activateInactivePlayer(),
		rollDice,
		doneTurn,
		countDie,
		finishTurn({ helpText: 'Advance to the next player when the current player has busted or said they are done.' }),
	),

	setUp: {
		distribute: state => state.gameStack('Die'),
		finish(state) {
			// A random player starts.
			state.g.CurrentPlayer = state.rand().intn(state.numPlayers());
			state.g.TargetScore = DEFAULT_TARGET_SCORE;
		},
	},

	end: {
		conditionMet(state) {
			for (const player of state.players) if ((player.Score as number) >= (state.g.TargetScore as number)) return true;
			return false;
		},
	},
};
