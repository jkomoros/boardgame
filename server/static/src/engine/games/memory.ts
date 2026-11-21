// Memory: players take turns revealing two cards and keep any pair that
// matches. A move-for-move port of examples/memory. It is the worked example
// in ../README.md.

import { componentPropsEqual, componentPropsNotEqual, firstOccupied, occupiedOrdinal, revealableCardAt } from '../legal/catalogPurpose.ts';
import { propAtLeast, propCompare } from '../legal/catalogCompare.ts';
import { mayMoveToSameSlot, stackCount, stackNotEmpty } from '../legal/catalogStack.ts';
import type { GameSpec, MoveSpec } from '../rules/types.ts';
import {
	activateInactivePlayer, add, currentPlayerProps, finishTurn, inactivePlayerProps, moveComponentToSlot, seatPlayer, seatProps,
} from '../rules/stock.ts';

const GENERAL = ['🚴', '✋', '💘', '🎓', '🌍', '🏖', '🏛', '⛺', '🚑', '🚕', '⚓', '🕰', '🌈', '🔥', '⛄', '🎄', '🎁', '🏆', '⚽', '🎳'];
const FOODS = ['🍒', '🍔', '🍭', '🍇', '🍉', '🍊', '🍌', '🍍', '🍓', '🌽', '🥕', '🍗', '🍕', '🍩', '🍦', '🍺', '🌮', '🌭', '🧀', '🥐'];
const ANIMALS = ['🐕', '🐄', '🐘', '🐍', '🦀', '🏇', '🦍', '🐈', '🐖', '🐫', '🐁', '🐿', '🦇', '🐓', '🦅', '🦉', '🐋', '🦑', '🐝', '🐡'];

type Card = { Type: string; CardSet: string };

/** Two of each card, set by set: the deck order ids are derived from. */
function newDeck(): Card[] {
	const cards: Card[] = [];
	for (const [cardSet, types] of [['general', GENERAL], ['foods', FOODS], ['animals', ANIMALS]] as const) {
		for (const type of types) for (let i = 0; i < 2; i++) cards.push({ Type: type, CardSet: cardSet });
	}
	return cards;
}

const HIDE_CARDS_DURATION_MS = 4000;
const HIDE_CARDS = 'Hide Cards';

// "Move the card at CardIndex out of HiddenCards into the same slot of
// VisibleCards", plus this game's bookkeeping.
function reveal() {
	return moveComponentToSlot({ source: 'game.HiddenCards', destination: 'game.VisibleCards', slotField: 'CardIndex', sourceSlotField: 'CardIndex' });
}

function revealCard(): MoveSpec {
	return {
		name: 'Reveal Card',
		helpText: 'Reveals the card at the specified location',
		fixUp: false,
		...reveal(),
		fields: { CardIndex: { type: 'int' } },
		// The cards still hidden are the choices offered to the player.
		choices: { field: 'CardIndex', stack: { scope: 'game', property: 'HiddenCards' } },
		defaults(state, move) {
			const hidden = state.gameStack('HiddenCards');
			for (let i = 0; i < hidden.len(); i++) {
				if (hidden.componentAt(i) === null) continue;
				move.fields.CardIndex = i;
				break;
			}
		},
		// The slot check the reusable move contributes would run before this
		// game's own gates and its generic message would win. Suppress it and
		// re-author it last, as Go does.
		suppress: ['mayMoveToSlot'],
		preconditions: [
			{ ...propAtLeast('player.CardsLeftToReveal', 1), message: 'reveal.no_cards_left' },
			revealableCardAt('game.HiddenCards', 'game.VisibleCards', 'move.CardIndex'),
			mayMoveToSameSlot('game.HiddenCards', 'game.VisibleCards', 'move.CardIndex'),
		],
		apply(state, move) {
			reveal().apply(state, move);
			const player = state.players[state.ensureValid(state.g.CurrentPlayer as number)];
			player.CardsLeftToReveal = (player.CardsLeftToReveal as number) - 1;
		},
	};
}

function hideCards(): MoveSpec {
	return {
		name: HIDE_CARDS,
		helpText: 'After the current player has revealed both cards and tried to memorize them, this move hides the cards so that play can continue to next player.',
		fixUp: false,
		base: 'currentPlayer',
		preconditions: [
			{ ...propCompare('player.CardsLeftToReveal', '<=', 0), message: 'hide.cards_still_to_reveal' },
			{ ...stackNotEmpty('game.VisibleCards'), message: 'hide.no_cards_to_hide' },
		],
		apply(state) {
			// The player hid the cards themself; the timer is no longer needed.
			state.timer('HideCardsTimer').cancel();
			const visible = state.gameStack('VisibleCards');
			const hidden = state.gameStack('HiddenCards');
			for (let i = 0; i < visible.len(); i++) {
				if (visible.componentAt(i) !== null) visible.moveComponent(i, hidden, i);
			}
		},
	};
}

function captureCards(): MoveSpec {
	return {
		name: 'Capture Cards',
		helpText: "If two cards are showing and they are the same type, capture them to the current player's hand.",
		fixUp: true,
		preconditions: [
			{ ...stackCount('game.VisibleCards', '==', 2), message: 'memory.two_cards_required' },
			{ ...componentPropsEqual(firstOccupied('game.VisibleCards'), occupiedOrdinal('game.VisibleCards', 1), 'Type'), message: 'memory.cards_must_match' },
		],
		apply(state) {
			const won = state.playerStack(state.g.CurrentPlayer as number, 'WonCards');
			const visible = state.gameStack('VisibleCards');
			for (let i = 0; i < visible.len(); i++) {
				if (visible.componentAt(i) !== null) visible.moveComponentToNextSlot(i, won);
			}
		},
	};
}

function startHideCardsTimer(): MoveSpec {
	return {
		name: 'Start Hide Cards Timer',
		helpText: 'If two cards are showing and they are not the same type and the timer is not active, start a timer to automatically hide them.',
		fixUp: true,
		preconditions: [
			{ ...stackCount('game.VisibleCards', '==', 2), message: 'memory.two_cards_required' },
			{ ...componentPropsNotEqual(firstOccupied('game.VisibleCards'), occupiedOrdinal('game.VisibleCards', 1), 'Type'), message: 'memory.cards_must_differ' },
		],
		// Whether a timer is running is not a property a predicate can read.
		legalCustom: state => (state.timer('HideCardsTimer').active() ? 'the timer is already active' : null),
		apply(state) {
			state.timer('HideCardsTimer').start(HIDE_CARDS_DURATION_MS, state.moveByName(HIDE_CARDS));
		},
	};
}

export function memory(): GameSpec {
	return {
		name: 'memory',
		minNumPlayers: 2,
		maxNumPlayers: 6,
		defaultNumPlayers: 2,

		decks: { cards: { components: newDeck() } },

		game: {
			props: {
				...currentPlayerProps,
				CardSet: { Type: 'string' },
				NumCards: { Type: 'int' },
				// Everyone sees how many cards are face down and when they move,
				// but not which card is which.
				HiddenCards: { Type: 'sizedStack', Deck: 'cards', Size: 40, Sanitize: 'order' },
				VisibleCards: { Type: 'sizedStack', Deck: 'cards', Size: 40 },
				// The table as the players see it: each slot shows the face-up
				// card if there is one, else the face-down one.
				Cards: { Type: 'mergedStack', Deck: 'cards', Overlap: true, Stacks: ['VisibleCards', 'HiddenCards'] },
				HideCardsTimer: { Type: 'timer' },
				UnusedCards: { Type: 'stack', Deck: 'cards' },
			},
		},
		player: {
			props: {
				...seatProps,
				...inactivePlayerProps,
				CardsLeftToReveal: { Type: 'int' },
				WonCards: { Type: 'stack', Deck: 'cards' },
			},
			traits: {
				seat: true,
				inactive: true,
				gameScore: (state, player) => state.playerView(player, 'WonCards').numComponents(),
				turn: {
					turnDone(state, player) {
						if ((state.players[player].CardsLeftToReveal as number) > 0) return 'they still have cards left to reveal';
						if (state.gameStack('VisibleCards').numComponents() > 0) return 'there are still some cards revealed, which they must hide';
						return null;
					},
					resetForTurnStart(state, player) {
						state.players[player].CardsLeftToReveal = 2;
					},
					resetForTurnEnd: () => undefined,
				},
			},
		},

		variants: {
			cardset: {
				displayName: 'Card Set',
				description: 'Which theme of cards to use',
				default: 'all',
				values: {
					all: { displayName: 'All Cards', description: 'All cards mixed together' },
					foods: { description: 'Food cards' },
					animals: { description: 'Animal cards' },
					general: { description: 'Random cards with no particular theme' },
				},
			},
			numcards: {
				displayName: 'Number of Cards',
				description: 'How many cards to use? Larger numbers are more difficult.',
				default: 'medium',
				values: {
					medium: { description: 'A default difficulty game' },
					small: { description: 'An easy game' },
					large: { description: 'A challenging game' },
				},
			},
		},

		computed: [
			{
				name: 'CurrentPlayerHasCardsToReveal',
				scope: 'global',
				type: 'bool',
				compute: state => (state.players[state.g.CurrentPlayer as number].CardsLeftToReveal as number) > 0,
			},
		],

		legalTemplates: {
			'reveal.no_cards_left': 'You have no cards left to reveal this turn',
			'hide.cards_still_to_reveal': 'You still have to reveal more cards before your turn is over',
			'hide.no_cards_to_hide': 'no cards left to hide',
			'memory.two_cards_required': "there aren't two cards showing",
			'memory.cards_must_match': 'The two revealed cards are not of the same type',
			'memory.cards_must_differ': 'The two revealed cards are of the same type',
		},

		moves: add(
			seatPlayer(),
			// A seated player starts inactive. Memory has no rounds, so they are
			// activated at once.
			activateInactivePlayer(),
			revealCard(),
			hideCards(),
			finishTurn(),
			captureCards(),
			startHideCardsTimer(),
		),

		setUp: {
			begin(state, variant) {
				state.g.CardSet = variant.cardset;
				const numCards = variant.numcards === 'small' ? 10 : variant.numcards === 'large' ? 40 : 20;
				state.g.NumCards = numCards;
				state.gameStack('HiddenCards').setSize(numCards);
				state.gameStack('VisibleCards').setSize(numCards);
			},
			// Everything starts in UnusedCards; finish builds the layout.
			distribute: state => state.gameStack('UnusedCards'),
			finish(state) {
				const unused = state.gameStack('UnusedCards');
				const hidden = state.gameStack('HiddenCards');
				const numCards = state.g.NumCards as number;
				const cardSet = state.g.CardSet as string;
				const card = (slot: number): Card => state.component('cards', unused.componentAt(slot) as number) as Card;

				unused.shuffle();
				// Take pairs from the front: into the layout if they are of the
				// chosen set, otherwise out of the way at the back.
				while (hidden.numComponents() < numCards) {
					const first = card(0);
					let pairIndex = 0;
					for (let i = 1; i < unused.len(); i++) {
						if (card(i).Type === first.Type) {
							pairIndex = i;
							break;
						}
					}
					if (pairIndex === 0) throw new Error('unexpectedly unable to find the pair when sorting cards to include');
					if (cardSet === 'all' || cardSet === first.CardSet) {
						unused.moveComponentToNextSlot(pairIndex, hidden);
						unused.moveComponentToNextSlot(0, hidden);
					} else {
						unused.slideToLastSlot(pairIndex);
						unused.slideToLastSlot(0);
					}
				}
				hidden.shuffle();
				state.players[0].CardsLeftToReveal = 2;
			},
		},

		end: {
			// Over when no card is left on the table; most pairs wins.
			conditionMet: state => state.gameView('Cards').numComponents() === 0,
		},
	};
}
