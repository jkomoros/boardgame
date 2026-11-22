// components/playingcards: a deck of American playing cards. A card's static
// values are `{ Suit, Rank }`, each an enum value stored as its string, with
// the enums "Suit" and "Rank" (playingcards' own enum set, auto_enum.go).
// Deck order is Go's deckCanonicalOrder: suits spades, hearts, clubs,
// diamonds; within each, ace through king; then two jokers if asked for.

import type { EnumDef } from '../definition.ts';

export const SUIT_UNKNOWN = 0;
export const SUIT_SPADES = 1;
export const SUIT_HEARTS = 2;
export const SUIT_CLUBS = 3;
export const SUIT_DIAMONDS = 4;
export const SUIT_JOKERS = 5;

export const RANK_UNKNOWN = 0;
export const RANK_ACE = 1;
export const RANK_2 = 2;
export const RANK_3 = 3;
export const RANK_4 = 4;
export const RANK_5 = 5;
export const RANK_6 = 6;
export const RANK_7 = 7;
export const RANK_8 = 8;
export const RANK_9 = 9;
export const RANK_10 = 10;
export const RANK_JACK = 11;
export const RANK_QUEEN = 12;
export const RANK_KING = 13;
export const RANK_JOKER = 14;

/** playingcards.SuitEnum, named "Suit". */
export function suitEnum(): EnumDef {
	return {
		Values: { 0: '�', 1: '♠', 2: '♥', 3: '♣', 4: '♦', 5: 'Jokers' },
		DefaultValue: SUIT_UNKNOWN,
	};
}

/** playingcards.RankEnum, named "Rank". */
export function rankEnum(): EnumDef {
	return {
		Values: {
			0: 'Unknown', 1: 'Ace', 2: '2', 3: '3', 4: '4', 5: '5', 6: '6', 7: '7', 8: '8', 9: '9', 10: '10', 11: 'Jack', 12: 'Queen', 13: 'King', 14: 'Joker',
		},
		DefaultValue: RANK_UNKNOWN,
	};
}

/** Both enums under their Go names, for a GameSpec's componentEnums. */
export function cardEnums(): Record<string, EnumDef> {
	return { Suit: suitEnum(), Rank: rankEnum() };
}

/** A card's static values: the enums' strings. */
export interface Card {
	Suit: string;
	Rank: string;
}

/** The static values for a card of this suit and rank key. */
export function card(suit: number, rank: number): Card {
	return { Suit: suitEnum().Values[String(suit)], Rank: rankEnum().Values[String(rank)] };
}

/** deckCanonicalOrder, appended to `cards`. */
function canonicalOrder(cards: Card[], withJokers: boolean): void {
	for (const suit of [SUIT_SPADES, SUIT_HEARTS, SUIT_CLUBS, SUIT_DIAMONDS]) {
		for (let rank = RANK_ACE; rank <= RANK_KING; rank++) cards.push(card(suit, rank));
	}
	if (withJokers) for (let i = 0; i < 2; i++) cards.push(card(SUIT_JOKERS, RANK_JOKER));
}

/** playingcards.NewDeck(withJokers): the components, in deck order. */
export function newDeck(withJokers: boolean): Card[] {
	const cards: Card[] = [];
	canonicalOrder(cards, withJokers);
	return cards;
}

/** playingcards.NewDeckMulti(count, withJokers): `count` decks (at least one) end to end. */
export function newDeckMulti(count: number, withJokers: boolean): Card[] {
	const cards: Card[] = [];
	for (let i = 0; i < Math.max(count, 1); i++) canonicalOrder(cards, withJokers);
	return cards;
}

/** The deck's generic values (Deck.SetGenericValues): unknown suit and rank. */
export function genericCard(): Card {
	return card(SUIT_UNKNOWN, RANK_UNKNOWN);
}

/** The key of a rank string, or RANK_UNKNOWN for one the enum lacks. */
export function rankKey(rank: string): number {
	const values = rankEnum().Values;
	for (const key of Object.keys(values)) if (values[key] === rank) return Number(key);
	return RANK_UNKNOWN;
}

/** The key of a suit string, or SUIT_UNKNOWN for one the enum lacks. */
export function suitKey(suit: string): number {
	const values = suitEnum().Values;
	for (const key of Object.keys(values)) if (values[key] === suit) return Number(key);
	return SUIT_UNKNOWN;
}

/** (*playingcards.Card).String(): "<suit> <rank>". */
export function cardString(c: Card): string {
	return c.Suit + ' ' + c.Rank;
}
