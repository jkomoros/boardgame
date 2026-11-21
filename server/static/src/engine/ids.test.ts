// SHA-1 and component-id known answers. testdata/ids.vectors.json comes from
// real Go (testdata/gen/idsvec): crypto/sha1 digests, and ids read off live
// boardgame states after shuffles and secret moves.
import { test } from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { sha1Hex, sha1HexBytes } from './sha1.ts';
import { componentId, secretMoveCountFor, GENERIC_COMPONENT, EMPTY_SLOT } from './ids.ts';

const vectors = JSON.parse(readFileSync(new URL('./testdata/ids.vectors.json', import.meta.url), 'utf8')) as {
	SHA1: { InputHex: string; Digest: string }[];
	IDs: { GameID: string; SecretSalt: string; Deck: string; DeckIndex: number; SecretMoveCount: number; ID: string }[];
};

function fromHex(hex: string): number[] {
	const out: number[] = [];
	for (let i = 0; i < hex.length; i += 2) out.push(parseInt(hex.slice(i, i + 2), 16));
	return out;
}

test('SHA-1 standard vectors (FIPS 180 examples)', () => {
	assert.equal(sha1Hex(''), 'da39a3ee5e6b4b0d3255bfef95601890afd80709');
	assert.equal(sha1Hex('abc'), 'a9993e364706816aba3e25717850c26c9cd0d89d');
	assert.equal(
		sha1Hex('abcdbcdecdefdefgefghfghighijhijkijkljklmklmnlmnomnopnopq'),
		'84983e441c3bd26ebaae4aa1f95129e5e54670f1',
	);
	assert.equal(
		sha1Hex('abcdefghbcdefghicdefghijdefghijkefghijklfghijklmghijklmnhijklmnoijklmnopjklmnopqklmnopqrlmnopqrsmnopqrstnopqrstu'),
		'a49b2446a02c645bf419f995b67091253a04a259',
	);
	assert.equal(sha1Hex('a'.repeat(1000000)), '34aa973cd4c4daa4f61eeb2bdbad27316534016f');
});

test('SHA-1 matches crypto/sha1 at every length across three block boundaries', () => {
	assert.ok(vectors.SHA1.length >= 205);
	for (const v of vectors.SHA1) {
		assert.equal(sha1HexBytes(fromHex(v.InputHex)), v.Digest, `length ${v.InputHex.length / 2}`);
	}
});

test('component ids match componentInstance.ID() for live Go states', () => {
	assert.ok(vectors.IDs.length >= 126);
	let nonZero = 0;
	for (const v of vectors.IDs) {
		if (v.SecretMoveCount > 0) nonZero++;
		assert.equal(
			componentId({ gameId: v.GameID, secretSalt: v.SecretSalt }, v.Deck, v.DeckIndex, v.SecretMoveCount),
			v.ID,
			`${v.GameID} ${v.Deck}[${v.DeckIndex}] count ${v.SecretMoveCount}`,
		);
	}
	assert.ok(nonZero >= 20, 'vectors exercise non-zero secret-move counts');
});

test('the generic component and empty slots have no id', () => {
	const identity = { gameId: 'G', secretSalt: 'S' };
	assert.equal(componentId(identity, 'cards', GENERIC_COMPONENT, 0), '');
	assert.equal(componentId(identity, 'cards', EMPTY_SLOT, 3), '');
});

test('secretMoveCountFor defaults to zero like Go', () => {
	assert.equal(secretMoveCountFor(undefined, 'cards', 0), 0);
	assert.equal(secretMoveCountFor({}, 'cards', 0), 0);
	assert.equal(secretMoveCountFor({ cards: [0, 4] }, 'cards', 1), 4);
	assert.equal(secretMoveCountFor({ cards: [0, 4] }, 'cards', 2), 0);
	assert.equal(secretMoveCountFor({ cards: [0, 4] }, 'cards', -2), 0);
});
