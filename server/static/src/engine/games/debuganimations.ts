// The animations test bed: a table of stacks and one move per kind of stack
// change. A move-for-move port of examples/debuganimations. It is not a game
// anyone wins; it is here because it exercises what the others do not:
// shuffles in mid-game (secret and public), every visibility policy on a
// stack, a move whose name some viewers may not learn at all, a move
// argument everyone may see, and a phase enum that is a tree.

import { stackCount, stackEmpty, stackNotEmpty } from '../legal/catalogStack.ts';
import type { Stack } from '../stack.ts';
import type { RulesState } from '../rules/state.ts';
import type { GameSpec, LegalResult, MoveSpec } from '../rules/types.ts';
import { add, currentPlayerProps, inactivePlayerProps, phaseProps, seatPlayer, seatProps } from '../rules/stock.ts';

const CARD_NAMES = ['🏇', '🚴', '✋', '💘', '🎓', '🐕', '🐄', '🐘', '🐍', '🦀', '🍒', '🍔', '🍭', '🍺', '🌮', '🌭', '🧀', '🥐'];
const NUM_TOKENS = 38;

function cards(): { Type: string }[] {
	const deck: { Type: string }[] = [];
	for (const name of CARD_NAMES) for (let i = 0; i < 3; i++) deck.push({ Type: name });
	return deck;
}

const count = (state: RulesState, prop: string): number => state.gameStack(prop).numComponents();

/** stack.First().MoveToFirstSlot(to). */
function firstToFirstSlot(from: Stack, to: Stack): void {
	from.moveComponentToFirstSlot(from.firstComponentIndex(), to);
}

/** "A card from one of two stacks to the other", the direction given by a bool field everyone may see. */
function moveBetween(options: { name: string; helpText: string; field: string; whenTrue: [string, string]; defaultTrue: (state: RulesState) => boolean }): MoveSpec {
	const ends = (state: RulesState, flag: unknown): [Stack, Stack] => {
		const [from, to] = flag === true ? options.whenTrue : [options.whenTrue[1], options.whenTrue[0]];
		return [state.gameStack(from), state.gameStack(to)];
	};
	return {
		name: options.name,
		helpText: options.helpText,
		fixUp: false,
		fields: { [options.field]: { type: 'bool', sanitize: 'all:visible' } },
		defaults(state, move) {
			move.fields[options.field] = options.defaultTrue(state);
		},
		legal(state, move, _proposer, base) {
			const err = base();
			if (err !== null) return err;
			const [from, to] = ends(state, move.fields[options.field]);
			if (from.numComponents() === 0) return 'source stack has no cards to move';
			return from.mayMoveComponentTo(from.firstComponentIndex(), to);
		},
		apply(state, move) {
			const [from, to] = ends(state, move.fields[options.field]);
			firstToFirstSlot(from, to);
		},
	};
}

/** "The third of the full stack to the other; back again into the third slot", legal only in those two arrangements. */
function toggleThird(options: { name: string; helpText: string; full: string; other: string; fullCount: number; otherCount: number; error: string }): MoveSpec {
	return {
		name: options.name,
		helpText: options.helpText,
		fixUp: false,
		legal(state, _move, _proposer, base): LegalResult {
			const err = base();
			if (err !== null) return err;
			const full = count(state, options.full);
			const other = count(state, options.other);
			if (full === options.fullCount && other === options.otherCount) return null;
			if (full === options.fullCount - 1 && other === options.otherCount + 1) return null;
			return options.error;
		},
		apply(state) {
			const full = state.gameStack(options.full);
			const other = state.gameStack(options.other);
			if (full.numComponents() < options.fullCount) other.moveComponent(other.firstComponentIndex(), full, 2);
			else full.moveComponentToFirstSlot(2, other);
		},
	};
}

function moveAll(options: { name: string; helpText: string; from: string; to: string; emptyMessage: string; occupiedMessage: string }): MoveSpec {
	return {
		name: options.name,
		helpText: options.helpText,
		fixUp: false,
		preconditions: [
			{ ...stackNotEmpty('game.' + options.from), message: options.emptyMessage },
			{ ...stackEmpty('game.' + options.to), message: options.occupiedMessage },
		],
		apply(state) {
			const from = state.gameStack(options.from);
			const to = state.gameStack(options.to);
			while (from.numComponents() > 0) from.moveComponentToNextSlot(from.firstComponentIndex(), to);
		},
	};
}

const flipHiddenCard: MoveSpec = {
	name: 'Flip Card Between Hidden and Revealed',
	helpText: 'Flips the card between hidden and revealed',
	fixUp: false,
	legal(state, _move, _proposer, base) {
		const err = base();
		if (err !== null) return err;
		const hidden = count(state, 'HiddenCard');
		const visible = count(state, 'VisibleCard');
		if (hidden < 1 && visible < 1) return 'Neither the HiddenCard nor RevealedCard is set';
		if (hidden > 0 && visible > 0) return 'both hidden and revealed are full';
		return null;
	},
	apply(state) {
		const hidden = state.gameStack('HiddenCard');
		const visible = state.gameStack('VisibleCard');
		if (hidden.numComponents() > 0) firstToFirstSlot(hidden, visible);
		else firstToFirstSlot(visible, hidden);
	},
};

const fanTooSmall = [{ ...stackCount('game.FanStack', '>', 1), message: 'debuganimations.fan_stack_too_small' }];

export const debuganimations: GameSpec = {
	name: 'debuganimations',
	displayName: 'Animations Debugger',
	minNumPlayers: 2,
	maxNumPlayers: 2,
	defaultNumPlayers: 2,

	decks: {
		cards: { components: cards() },
		// Tokens have no values of their own.
		tokens: { components: Array.from({ length: NUM_TOKENS }, () => null) },
	},
	// A tree enum: "Normal" is a child of the unnamed root. The default of a
	// tree enum is its first leaf.
	enums: { phase: { Values: { 0: '', 1: 'Normal' }, DefaultValue: 1, Parents: { 0: 0, 1: 0 } } },

	game: {
		props: {
			...currentPlayerProps,
			...phaseProps,
			DiscardStack: { Type: 'stack', Deck: 'cards' },
			DrawStack: { Type: 'stack', Deck: 'cards', Sanitize: 'order' },
			FirstShortStack: { Type: 'stack', Deck: 'cards', Sanitize: 'order' },
			SecondShortStack: { Type: 'stack', Deck: 'cards', Sanitize: 'order' },
			HiddenCard: { Type: 'sizedStack', Deck: 'cards', Size: 1, Sanitize: 'order' },
			VisibleCard: { Type: 'sizedStack', Deck: 'cards', Size: 1 },
			Card: { Type: 'mergedStack', Deck: 'cards', Overlap: true, Stacks: ['VisibleCard', 'HiddenCard'] },
			FanStack: { Type: 'stack', Deck: 'cards' },
			FanDiscard: { Type: 'stack', Deck: 'cards', Sanitize: 'order' },
			FanShuffleCount: { Type: 'int' },
			VisibleStack: { Type: 'stack', Deck: 'cards' },
			HiddenStack: { Type: 'stack', Deck: 'cards', Sanitize: 'nonempty' },
			AllVisibleStack: { Type: 'stack', Deck: 'cards' },
			AllHiddenStack: { Type: 'stack', Deck: 'cards', Sanitize: 'order' },
			TokensFrom: { Type: 'stack', Deck: 'tokens' },
			TokensTo: { Type: 'stack', Deck: 'tokens' },
			SanitizedTokensFrom: { Type: 'stack', Deck: 'tokens' },
			SanitizedTokensTo: { Type: 'stack', Deck: 'tokens', Sanitize: 'nonempty' },
		},
	},
	player: {
		props: { ...seatProps, ...inactivePlayerProps, Hand: { Type: 'stack', Deck: 'cards', MaxSize: 1 } },
		traits: { seat: true, inactive: true },
	},

	legalTemplates: {
		'debuganimations.fan_stack_too_small': "Aren't enough cards to shuffle",
		'debuganimations.fan_discard_empty': 'FanDiscard has no cards to shuffle',
		'debuganimations.all_visible_stack_empty': 'No components in visible stack to move',
		'debuganimations.all_hidden_stack_occupied': "The hidden stack already has items. Use the 'To Visible' move",
		'debuganimations.all_hidden_stack_empty': 'No components in hidden stack to move',
		'debuganimations.all_visible_stack_occupied': "The visible stack already has items. Use the 'To Hidden' move",
	},

	// Every move is open to any player at any time.
	moves: add(
		seatPlayer(),
		moveBetween({
			name: 'Move Card Between Short Stacks',
			helpText: 'Moves a card between two short stacks',
			field: 'FromFirst',
			whenTrue: ['FirstShortStack', 'SecondShortStack'],
			defaultTrue: state => count(state, 'FirstShortStack') >= 1,
		}),
		moveBetween({
			name: 'Move Card Between Draw And Discard Stacks',
			helpText: 'Moves a card between draw and discard stacks',
			field: 'FromDraw',
			whenTrue: ['DrawStack', 'DiscardStack'],
			defaultTrue: state => count(state, 'DiscardStack') < 3,
		}),
		flipHiddenCard,
		toggleThird({
			name: 'Move Fan Card', helpText: 'Moves a card from or to Fan and Fan Discard',
			full: 'FanStack', other: 'FanDiscard', fullCount: 6, otherCount: 3, error: "Fan stacks aren't in known toggle state",
		}),
		{
			name: 'Visible Shuffle',
			helpText: 'Performs a visible shuffle',
			fixUp: false,
			preconditions: fanTooSmall,
			// Everyone can follow where each card went.
			apply: state => state.gameStack('FanStack').publicShuffle(),
		},
		{
			name: 'Shuffle',
			helpText: 'Performs a secret shuffle',
			fixUp: false,
			preconditions: fanTooSmall,
			// Every card's id changes, so nobody can follow.
			apply: state => state.gameStack('FanStack').shuffle(),
		},
		toggleThird({
			name: 'Move Between Hidden', helpText: 'Moves between hidden and visible stacks',
			full: 'VisibleStack', other: 'HiddenStack', fullCount: 5, otherCount: 4, error: "Cards aren't in known position",
		}),
		toggleThird({
			name: 'Move Token', helpText: 'Moves tokens',
			full: 'TokensFrom', other: 'TokensTo', fullCount: 10, otherCount: 9, error: "tokens aren't in known position",
		}),
		toggleThird({
			name: 'Move Token Sanitized', helpText: 'Moves tokens',
			full: 'SanitizedTokensFrom', other: 'SanitizedTokensTo', fullCount: 10, otherCount: 9, error: "tokens aren't in known position",
		}),
		moveAll({
			name: 'Start Move All Components To Hidden', helpText: 'Moves all components from visible to hidden',
			from: 'AllVisibleStack', to: 'AllHiddenStack',
			emptyMessage: 'debuganimations.all_visible_stack_empty', occupiedMessage: 'debuganimations.all_hidden_stack_occupied',
		}),
		moveAll({
			name: 'Start Move All Components To Visible', helpText: 'Moves all components from hidden to visible',
			from: 'AllHiddenStack', to: 'AllVisibleStack',
			emptyMessage: 'debuganimations.all_hidden_stack_empty', occupiedMessage: 'debuganimations.all_visible_stack_occupied',
		}),
		{
			name: 'Shuffle Hidden',
			helpText: 'Shuffles the fan discard pile and increments the shuffle count',
			fixUp: false,
			// Only the proposer learns this move happened by name. With no
			// alias, everyone else gets no move record, only the new state.
			nameSanitization: 'self:visible',
			preconditions: [{ ...stackNotEmpty('game.FanDiscard'), message: 'debuganimations.fan_discard_empty' }],
			apply(state) {
				state.gameStack('FanDiscard').shuffle();
				state.g.FanShuffleCount = (state.g.FanShuffleCount as number) + 1;
			},
		},
	),

	setUp: {
		distribute(state, deck) {
			const fill = (prop: string, target: number): Stack | null => (count(state, prop) < target ? state.gameStack(prop) : null);
			if (deck === 'tokens') {
				return fill('TokensTo', 9) ?? fill('SanitizedTokensFrom', 10) ?? fill('SanitizedTokensTo', 9) ?? state.gameStack('TokensFrom');
			}
			return (
				fill('FirstShortStack', 1) ?? fill('SecondShortStack', 1) ?? fill('DiscardStack', 2) ?? fill('HiddenCard', 1) ?? fill('FanStack', 6) ??
				fill('FanDiscard', 3) ?? fill('VisibleStack', 5) ?? fill('HiddenStack', 4) ?? fill('AllVisibleStack', 4) ?? state.gameStack('DrawStack')
			);
		},
		finish: state => state.gameStack('DrawStack').shuffle(),
	},
	// Never ends.
};
