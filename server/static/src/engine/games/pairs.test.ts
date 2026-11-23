// The stored pairs (boardgame-util/lib/transpile/testdata/pairs/,
// fabric/TRANSPILE-PROPOSAL.md "Stored pairs") are faithful: each of the
// six games, built by gameFromData from its exported definition
// (conformance/definitions/<game>.json) with a bodies table made of the
// stored accepted.ts functions (pairs.testkit.test.ts), passes the game's
// whole conformance corpus and every checked-in fuzz document, rejections
// included.
//
// Where no pair serves a FnRef the loader asks for, a stand-in
// (pairs.testkit.test.ts) is used; the only such FnRef allowed is werewolf's empty
// SetUp.Distribute (Go's werewolf declares no
// DistributeComponentToStarterStack, so there is no Go function to pair).
// memory's Reveal Card and tictactoe's Place Token are MoveComponentToSlot
// configured without WithSlotField; rules/stockRegistry.ts cannot build them
// from the exported configuration yet (fabric/transpile/status/integration-1.md).
// Through the same `stock:` hook rules/data.test.ts uses, they get the
// registry's own parts with the slot field Go discovers at boot (the move's
// only int field) rather than the hand port's parts, so a translated body's
// stockApply has the stock Apply to call. When the registry learns to
// discover the field, the hook goes unused and the test still passes.
//
// The first test checks that the store is what
// scripts/pairs-extract.mjs cuts from the hand ports today, so a hand-port
// edit that is not carried into the store fails here, by pair. The
// extraction is a one-time step: once the hand ports (games/<game>.ts) are
// removed, that test skips and the store is the record of them. Nothing else
// here, in pairs.testkit.test.ts or in the store, imports a hand port.
import assert from 'node:assert/strict';
import { existsSync, readFileSync } from 'node:fs';
import path from 'node:path';
import { test } from 'node:test';
import { extractPairs, STORE_DIR } from '../../../scripts/pairs-extract.mjs';
import type { CompiledGame, GameSpec } from '../rules/index.ts';
import { type StoreBodies, PAIR_GAMES, checkDocument, corpusDocuments, readDocument, storeGame, storedRefs } from './pairs.testkit.test.ts';

/** FnRefs no pair may serve, with why; anything else falling back fails. */
const ALLOWED_FALLBACKS: Readonly<Record<string, readonly string[]>> = {
	// werewolf's empty SetUp.Distribute is no fallback any more: the loader itself supplies Go's base-delegate stub.
	werewolf: [],
};
/** Moves the registry builds only with the discovered slot field (the registry gap); at most these. */
const SLOT_FIELD_RESOLVED: Readonly<Record<string, readonly string[]>> = {
	memory: ['Reveal Card'],
	tictactoe: ['Place Token'],
};

interface ExtractedPair {
	game: string;
	func: string;
	name: string;
	extraction: string;
	note: string;
	accepted: string;
}

test('the store is what pairs-extract.mjs cuts from the hand ports today', async t => {
	const handPorts = PAIR_GAMES.filter(game => existsSync(new URL('./' + game + '.ts', import.meta.url)));
	if (handPorts.length < PAIR_GAMES.length) {
		t.skip('the hand ports are gone (' + handPorts.length + ' of ' + PAIR_GAMES.length + ' present): the extraction was a one-time step and the store is their record');
		return;
	}
	const { pairs, none } = (await extractPairs()) as { pairs: ExtractedPair[]; none: { func: string; reason: string }[] };
	const differing: string[] = [];
	for (const pair of pairs) {
		const dir = path.join(STORE_DIR, pair.game, pair.name);
		let text = '';
		let meta: { func?: string; extraction?: string; note?: string; producer?: string } = {};
		try {
			text = readFileSync(path.join(dir, 'accepted.ts'), 'utf8');
			meta = JSON.parse(readFileSync(path.join(dir, 'meta.json'), 'utf8')) as typeof meta;
		} catch {
			differing.push(pair.game + '/' + pair.name + ': missing');
			continue;
		}
		if (text !== pair.accepted) differing.push(pair.game + '/' + pair.name + ': accepted.ts differs');
		if (meta.func !== pair.func || meta.extraction !== pair.extraction || (meta.note ?? '') !== pair.note || meta.producer !== 'hand') differing.push(pair.game + '/' + pair.name + ': meta.json differs');
	}
	for (const game of PAIR_GAMES) {
		const extracted = pairs.filter(pair => pair.game === game).map(pair => pair.func).sort();
		assert.deepEqual(await storedRefs(game), extracted, game + ': the store index lists exactly the extracted pairs');
	}
	assert.deepEqual(differing, [], 'rebuild the store: BOARDGAME_NODE=<node> ./scripts/go-local run ./boardgame-util/lib/transpile/pairs/testdata/gen');
	const counts = PAIR_GAMES.map(game => game + ' ' + pairs.filter(pair => pair.game === game).length).join(', ');
	t.diagnostic('pairs: ' + pairs.length + ' (' + counts + '); no pair: ' + none.map(item => item.func).join(', '));
});

for (const game of PAIR_GAMES) {
	let built: { game: CompiledGame; spec: () => GameSpec; table: StoreBodies } | undefined;
	const documents = corpusDocuments(game);

	test(game + ' from its definition and the stored pairs: compiles', async () => {
		built = await storeGame(game);
	});

	for (const entry of documents) {
		test(game + ' from the stored pairs: ' + entry.file, () => {
			assert.ok(built !== undefined, 'the game did not compile');
			const check = checkDocument(built.game, entry.file, readDocument(entry.file), built.spec);
			assert.equal(check.problems.length, 0, '\n  ' + check.problems.join('\n  '));
		});
	}

	test(game + ' from the stored pairs: every FnRef was served by a pair', async t => {
		assert.ok(built !== undefined, 'the game did not compile');
		const { served, fallbacks, stockStandIns, stockResolved } = built.table;
		assert.deepEqual([...fallbacks].sort(), [...(ALLOWED_FALLBACKS[game] ?? [])].sort(), game + ': FnRefs no pair served');
		assert.deepEqual(stockStandIns, [], game + ': moves whose stock parts nothing could build');
		for (const name of stockResolved) assert.ok((SLOT_FIELD_RESOLVED[game] ?? []).includes(name), game + ': ' + name + ' needed the discovered slot field');
		const stored = await storedRefs(game);
		for (const ref of served) assert.ok(stored.includes(ref), ref);
		t.diagnostic(
			game + ': ' + documents.length + ' documents; ' + stored.length + ' pairs stored, ' + served.length + ' FnRefs served by pairs' +
				(fallbacks.length ? ', stand-in for ' + JSON.stringify(fallbacks) : '') + (stockResolved.length ? ', stock: registry parts with the discovered slot field for ' + stockResolved.join(', ') : '') +
				'; pairs not resolved by the loader (helpers): ' + stored.filter(ref => !served.includes(ref)).length,
		);
	});
}
