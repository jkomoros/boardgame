// Component ids, reproducing componentInstance.ID() in component.go.
//
// An id is the lowercase hex SHA-1 of the plain concatenation
//
//     gameId + secretSalt + deckName + deckIndex + secretMoveCount
//
// with the two integers in decimal and no separators. secretMoveCount is the
// per-component counter in state.SecretMoveCount[deckName][deckIndex]; it
// rises by one each time the component's id is scrambled (Shuffle, and
// SecretMoveTo for every component of the destination).

import { sha1Hex } from './sha1.ts';

/** Deck index of an empty slot in a sized stack (stack.go emptyIndexSentinel). */
export const EMPTY_SLOT = -1;
/** Deck index of the generic component a sanitized stack shows (deck.go). */
export const GENERIC_COMPONENT = -2;

/** The identity of one game: what ids and the generator are derived from. */
export interface GameIdentity {
	readonly gameId: string;
	readonly secretSalt: string;
}

/**
 * The id of the component at deckIndex in deckName. Returns "" for the
 * generic component, as Go does. An empty slot has no component; callers
 * render it as "" too (stackIdsImpl leaves the zero string).
 */
export function componentId(
	identity: GameIdentity,
	deckName: string,
	deckIndex: number,
	secretMoveCount: number,
): string {
	if (deckIndex < 0) return '';
	return sha1Hex(identity.gameId + identity.secretSalt + deckName + String(deckIndex) + String(secretMoveCount));
}

/**
 * The secret-move count Go would use for a component: 0 when the deck has no
 * counts yet or the index is out of range (componentInstance.secretMoveCount).
 */
export function secretMoveCountFor(
	counts: Readonly<Record<string, readonly number[]>> | undefined | null,
	deckName: string,
	deckIndex: number,
): number {
	if (!counts) return 0;
	const deckCounts = counts[deckName];
	if (!deckCounts) return 0;
	if (deckIndex < 0 || deckIndex >= deckCounts.length) return 0;
	return deckCounts[deckIndex];
}
