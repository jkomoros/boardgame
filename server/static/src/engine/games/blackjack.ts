// Blackjack: players draw toward 21 over several rounds. A move-for-move
// port of examples/blackjack. It is the example with hidden state (each
// player's hole card) and with phases whose moves run in a fixed order.

import { playerBool, playerBoolIs } from '../legal/catalogCompare.ts';
import { allActivePlayers, any } from '../legal/catalogPlayers.ts';
import { stackNotEmpty } from '../legal/catalogStack.ts';
import { GENERIC_COMPONENT } from '../ids.ts';
import type { RulesState } from '../rules/state.ts';
import type { GameSpec, MoveSpec } from '../rules/types.ts';
import {
	activateInactivePlayer, add, addForPhase, addOrderedForPhase, combine, currentPlayerProps, dealCountComponents, defaultRoundSetup,
	drawToPlayer, eliminationProps, finishTurn, forceFinishTurn, inactivePlayerProps, inactivateEmptySeat, optional, phaseProps,
	roundRobinProps, scoreGameScore, scoreProps, seatPlayer, seatProps, shuffleDiscardIntoDraw, startPhase, waitForEnoughPlayers,
} from '../rules/stock.ts';

const TARGET_SCORE = 21;
const DEFAULT_MAX_ROUNDS = 5;

const PHASE_GATHERING = 0;
const PHASE_INITIAL_DEAL = 1;
const PHASE_NORMAL_PLAY = 2;
const PHASE_ROUND_CLEANUP = 3;

// components/playingcards: the enum keys and the canonical deck order.
const SUITS: Record<string, string> = { 0: '�', 1: '♠', 2: '♥', 3: '♣', 4: '♦', 5: 'Jokers' };
const RANKS: Record<string, string> = {
	0: 'Unknown', 1: 'Ace', 2: '2', 3: '3', 4: '4', 5: '5', 6: '6', 7: '7', 8: '8', 9: '9', 10: '10', 11: 'Jack', 12: 'Queen', 13: 'King', 14: 'Joker',
};
const RANK_ACE = 1;
const RANK_JACK = 11;
const RANK_KING = 13;

function newDeck(): { Suit: string; Rank: string }[] {
	const cards: { Suit: string; Rank: string }[] = [];
	for (const suit of [1, 2, 3, 4]) for (let rank = RANK_ACE; rank <= RANK_KING; rank++) cards.push({ Suit: SUITS[suit], Rank: RANKS[rank] });
	return cards;
}

/** A rank's enum key, from its name. */
function rankKey(rank: string): number {
	for (const key of Object.keys(RANKS)) {
		if (RANKS[key] === rank) return Number(key);
	}
	return 0;
}

/**
 * playerState.HandValue. Works on whatever state it is given, so on a
 * viewer's sanitized state a hidden card counts as 0.
 *
 * QUIRK reproduced from examples/blackjack/state.go:96: an ace is promoted
 * from 1 to 11 only while the hand is below 11, so ten plus ace scores 11,
 * not 21. It looks like a bug in the Go example; matching it is required.
 */
export function handValue(state: RulesState, player: number): number {
	const hand = state.playerView(player, 'Hand');
	let aces = 0;
	let value = 0;
	for (let i = 0; i < hand.len(); i++) {
		const index = hand.componentAt(i);
		if (index === null) continue;
		const rank = index === GENERIC_COMPONENT ? 0 : rankKey(state.component('cards', index).Rank as string);
		if (rank === RANK_ACE) {
			aces++;
			value++;
		} else if (rank >= RANK_JACK && rank <= RANK_KING) {
			value += 10;
		} else {
			value += rank;
		}
	}
	while (aces > 0) {
		if (value >= TARGET_SCORE - 10) break;
		aces--;
		value += 10;
	}
	return value;
}

const hasCards = (state: RulesState, player: number): boolean =>
	state.playerStack(player, 'HiddenHand').numComponents() > 0 || state.playerStack(player, 'VisibleHand').numComponents() > 0;

function draw() {
	return drawToPlayer({ gameStack: 'DrawStack', playerStack: 'VisibleHand' });
}

function hit(): MoveSpec {
	return {
		name: 'Current Player Hit',
		helpText: 'The current player hits, drawing a card.',
		fixUp: false,
		...draw(),
		preconditions: [
			{ ...playerBoolIs('Eliminated', false), message: 'hit.already_busted' },
			{ ...stackNotEmpty('game.DrawStack'), message: 'hit.no_cards_left' },
		],
		// A hand's value is computed, so no predicate can read it.
		legalCustom: state => (handValue(state, state.currentPlayerIndex()) >= TARGET_SCORE ? 'Current player is already at target scores' : null),
		apply(state, move) {
			draw().apply(state, move);
			const current = state.currentPlayerIndex();
			const value = handValue(state, current);
			if (value > TARGET_SCORE) state.players[current].Eliminated = true;
			if (value === TARGET_SCORE) state.players[current].Stood = true;
		},
	};
}

function stand(): MoveSpec {
	return {
		name: 'Current Player Stand',
		helpText: 'If the current player no longer wants to draw cards, they can stand.',
		fixUp: false,
		base: 'currentPlayer',
		preconditions: [
			{ ...playerBoolIs('Eliminated', false), message: 'stand.already_busted' },
			{ ...playerBoolIs('Stood', false), message: 'stand.already_stood' },
		],
		apply(state) {
			state.players[state.ensureValid(state.g.CurrentPlayer as number)].Stood = true;
		},
	};
}

function revealHiddenCard(): MoveSpec {
	return {
		name: 'Reveal Hidden Card',
		helpText: "Reveals the hidden card in the user's hand",
		fixUp: true,
		base: 'currentPlayer',
		preconditions: [{ ...stackNotEmpty('player.HiddenHand'), message: 'reveal.no_hidden_card' }],
		legalCustom(state, move) {
			const player = move.fields.TargetPlayerIndex as number;
			const hidden = state.playerStack(player, 'HiddenHand');
			// Go dereferences the first component without checking for one; the
			// declared precondition above is what guarantees it.
			if (hidden.numComponents() === 0) throw new Error('runtime error: invalid memory address or nil pointer dereference');
			return hidden.mayMoveComponentTo(hidden.firstComponentIndex(), state.playerStack(player, 'VisibleHand'));
		},
		apply(state, move) {
			const player = move.fields.TargetPlayerIndex as number;
			const hidden = state.playerStack(player, 'HiddenHand');
			hidden.moveComponentToFirstSlot(hidden.firstComponentIndex(), state.playerStack(player, 'VisibleHand'));
		},
	};
}

function accumulateScores(): MoveSpec {
	return {
		name: 'Accumulate Scores',
		helpText: "Adds each player's hand value to their total score.",
		fixUp: true,
		legal(state, _move, _proposer, base) {
			const err = base();
			if (err !== null) return err;
			for (let i = 0; i < state.numPlayers(); i++) {
				if (state.playerIsInactive(i)) continue;
				if (state.playerView(i, 'Hand').numComponents() > 0) return null;
			}
			return 'no active players have cards to score';
		},
		apply(state) {
			for (let i = 0; i < state.numPlayers(); i++) {
				if (state.playerIsInactive(i) || state.players[i].Eliminated === true) continue;
				state.players[i].Score = (state.players[i].Score as number) + handValue(state, i);
			}
		},
	};
}

function collectCards(): MoveSpec {
	return {
		name: 'Collect Cards',
		helpText: 'Collects all cards from players back to the discard pile.',
		fixUp: true,
		legal(state, _move, _proposer, base) {
			const err = base();
			if (err !== null) return err;
			for (let i = 0; i < state.numPlayers(); i++) if (hasCards(state, i)) return null;
			return 'no player has cards to collect';
		},
		apply(state) {
			const discard = state.gameStack('DiscardStack');
			for (let i = 0; i < state.numPlayers(); i++) {
				state.playerStack(i, 'HiddenHand').moveAllTo(discard);
				state.playerStack(i, 'VisibleHand').moveAllTo(discard);
			}
		},
	};
}

function resetPlayerForNewRound(): MoveSpec {
	return {
		name: 'Reset Player For New Round',
		helpText: 'Resets player flags for the next round.',
		fixUp: true,
		legal(state, _move, _proposer, base) {
			const err = base();
			if (err !== null) return err;
			for (let i = 0; i < state.numPlayers(); i++) {
				if (state.playerIsInactive(i)) continue;
				if (state.players[i].Eliminated === true || state.players[i].Stood === true) return null;
			}
			return 'no active players need resetting';
		},
		apply(state) {
			for (const player of state.players) {
				player.Eliminated = false;
				player.Stood = false;
			}
		},
	};
}

function incrementRoundsCompleted(): MoveSpec {
	return {
		name: 'Increment Rounds Completed',
		helpText: 'Increments the round counter.',
		fixUp: true,
		legal(state, _move, _proposer, base) {
			const err = base();
			if (err !== null) return err;
			for (let i = 0; i < state.numPlayers(); i++) if (hasCards(state, i)) return "cards haven't been collected yet";
			if (state.currentPhaseKey() !== PHASE_ROUND_CLEANUP) return 'not in cleanup phase';
			return null;
		},
		apply(state) {
			state.g.RoundsCompleted = (state.g.RoundsCompleted as number) + 1;
		},
	};
}

/** The variant's values: 1 to 10 rounds. */
function maxRoundsValues(): Record<string, { description: string }> {
	const values: Record<string, { description: string }> = {};
	for (let i = 1; i <= 10; i++) values[String(i)] = { description: i + ' rounds' };
	return values;
}

export function blackjack(): GameSpec {
	return {
		name: 'blackjack',
		minNumPlayers: 2,
		maxNumPlayers: 7,
		defaultNumPlayers: 4,

		decks: {
			cards: { components: newDeck(), props: { Suit: { Type: 'enum', Enum: 'Suit' }, Rank: { Type: 'enum', Enum: 'Rank' } } },
		},
		enums: {
			phase: { Values: { 0: 'Gathering', 1: 'Initial Deal', 2: 'Normal Play', 3: 'Round Cleanup' }, DefaultValue: 0 },
		},
		componentEnums: {
			Suit: { Values: SUITS, DefaultValue: 0 },
			Rank: { Values: RANKS, DefaultValue: 0 },
		},

		game: {
			props: {
				...roundRobinProps,
				...currentPlayerProps,
				...phaseProps,
				// Only how many cards are left is public.
				DiscardStack: { Type: 'stack', Deck: 'cards', Sanitize: 'len' },
				DrawStack: { Type: 'stack', Deck: 'cards', Sanitize: 'len' },
				UnusedCards: { Type: 'stack', Deck: 'cards' },
				MaxRounds: { Type: 'int' },
				RoundsCompleted: { Type: 'int' },
			},
			traits: { drawDiscardPair: { draw: 'DrawStack', discard: 'DiscardStack' } },
		},
		player: {
			props: {
				...seatProps,
				...inactivePlayerProps,
				...scoreProps,
				...eliminationProps,
				// The hole card: its owner sees it, everyone else sees only that it is there.
				HiddenHand: { Type: 'stack', Deck: 'cards', MaxSize: 1, Sanitize: 'len' },
				VisibleHand: { Type: 'stack', Deck: 'cards' },
				Hand: { Type: 'mergedStack', Deck: 'cards', Overlap: false, Stacks: ['HiddenHand', 'VisibleHand'] },
				Stood: { Type: 'bool' },
			},
			traits: {
				seat: true,
				inactive: true,
				elimination: true,
				gameScore: scoreGameScore,
				turn: {
					turnDone: (state, player) =>
						state.players[player].Eliminated !== true && state.players[player].Stood !== true ? 'they have neither busted nor decided to stand' : null,
					resetForTurnStart(state, player) {
						state.players[player].Stood = false;
					},
					resetForTurnEnd: () => undefined,
				},
			},
		},

		variants: {
			maxrounds: {
				displayName: 'Max Rounds',
				description: 'Number of rounds to play before the game ends',
				default: String(DEFAULT_MAX_ROUNDS),
				values: maxRoundsValues(),
			},
		},

		// Computed from the state each viewer is shown: another player's hole
		// card counts as 0 until it is revealed.
		computed: [{ name: 'HandValue', scope: 'player', type: 'int', compute: handValue }],

		legalTemplates: {
			'cleanup.players_unfinished': 'not all active players have finished their turn',
			'stand.already_busted': 'the current player has already busted',
			'stand.already_stood': 'the current player already stood',
			'hit.already_busted': 'Current player is busted',
			'hit.no_cards_left': 'No cards left in draw stack',
			'reveal.no_hidden_card': 'Target player has no cards to reveal',
		},

		moves: combine(
			add(
				shuffleDiscardIntoDraw({ helpText: 'When the draw deck is empty, shuffles the discard deck into draw deck.' }),
				// Players may be seated at any time; they join inactive.
				seatPlayer(),
			),
			// Gathering: wait for players, then begin.
			addOrderedForPhase(
				PHASE_GATHERING,
				optional(activateInactivePlayer({ nameSuffix: 'Gathering' })),
				waitForEnoughPlayers({ nameSuffix: 'Gathering' }),
				optional(inactivateEmptySeat({ nameSuffix: 'Gathering' })),
				startPhase({ phase: PHASE_INITIAL_DEAL, phaseName: 'Initial Deal', name: 'Begin Game' }),
			),
			addForPhase(
				PHASE_NORMAL_PLAY,
				startPhase({
					phase: PHASE_ROUND_CLEANUP,
					phaseName: 'Round Cleanup',
					name: 'Start Round Cleanup',
					helpText: 'When all players have finished, transitions to round cleanup.',
					preconditions: [{ ...allActivePlayers(any(playerBool('Eliminated'), playerBool('Stood'))), message: 'cleanup.players_unfinished' }],
				}),
				hit(),
				stand(),
				revealHiddenCard(),
				finishTurn({ helpText: 'When the current player has either busted or decided to stand, we advance to next player.' }),
				// The admin may end a turn that is not done. Not a fix-up: it is
				// always legal for the admin, so as one it would run forever.
				forceFinishTurn({
					name: 'Force Finish Turn',
					helpText: "Admin-only: end the current player's turn even if TurnDone() would refuse. Used by host SkipTurn in Table+Hand mode.",
				}),
			),
			addOrderedForPhase(
				PHASE_INITIAL_DEAL,
				defaultRoundSetup(),
				dealCountComponents({ name: 'Deal Initial Hidden Card', helpText: 'Deals a hidden card to each player', gameStack: 'DrawStack', playerStack: 'HiddenHand' }),
				dealCountComponents({ name: 'Deal Initial Visible Card', helpText: 'Deals a visible card to each player', gameStack: 'DrawStack', playerStack: 'VisibleHand' }),
				startPhase({ phase: PHASE_NORMAL_PLAY, phaseName: 'Normal Play' }),
			),
			addOrderedForPhase(
				PHASE_ROUND_CLEANUP,
				accumulateScores(),
				collectCards(),
				resetPlayerForNewRound(),
				incrementRoundsCompleted(),
				startPhase({ phase: PHASE_INITIAL_DEAL, phaseName: 'Initial Deal', name: 'Start Next Round' }),
			),
		),

		setUp: {
			begin(state, variant) {
				const text = variant.maxrounds ?? '';
				state.g.MaxRounds = /^[+-]?[0-9]+$/.test(text) ? Number(text) : DEFAULT_MAX_ROUNDS;
			},
			distribute: (state, _deck, index) => state.gameStack(state.component('cards', index).Rank === 'Joker' ? 'UnusedCards' : 'DrawStack'),
			finish: state => state.gameStack('DrawStack').shuffle(),
		},

		end: {
			conditionMet: state => (state.g.RoundsCompleted as number) >= (state.g.MaxRounds as number),
		},
	};
}
