// Golden replay, as far as it goes without game rules. Loads the real golden
// histories under examples/memory/testdata/golden and
// examples/blackjack/testdata/golden and, at every version, checks the engine
// against Go:
//
//   - the engine's expansion of the golden file equals Go's (state digests);
//   - every recorded component id recomputes from the recorded salt and
//     secret-move counts, and every merged-stack view recomputes from its
//     parts (a loaded state serializes back to the recorded JSON);
//   - component conservation holds;
//   - the generator reproduces every recorded shuffle (all of them happen in
//     set-up, at version 0) and every recorded timer id, from the recorded
//     salt and version;
//   - viewFor for every viewer equals what Go's Game.JSONForPlayer produced.
//
// Expected values come from testdata/golden/*.golden.json, written by real Go
// (testdata/gen/goldenvec).
import { test } from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import { join } from 'node:path';
import { EngineState, cloneJSON, compareGoStrings } from './state.ts';
import type { StateJSON } from './state.ts';
import type { StackJSON } from './stack.ts';
import type { GameDefinition } from './definition.ts';
import { viewFor } from './sanitize.ts';
import { componentId, secretMoveCountFor } from './ids.ts';
import { stateRand } from './rand.ts';
import { sha1HexBytes } from './sha1.ts';
import { utf8Bytes } from './u64.ts';
import { readGoldenHistory } from './testing/golden-record.ts';
import type { GoldenHistory } from './testing/golden-record.ts';

const HERE = fileURLToPath(new URL('.', import.meta.url));
const REPO = join(HERE, '..', '..', '..', '..');
const read = (name: string) => JSON.parse(readFileSync(join(HERE, 'testdata', name), 'utf8'));

interface GoHistory {
	File: string;
	GameID: string;
	SecretSalt: string;
	NumPlayers: number;
	Version: number;
	StateDigests: string[];
	ViewDigests: Record<string, string>[];
	Views: Record<string, Record<string, StateJSON>>;
}

/** The canonical encoding of testdata/gen/goldenvec's canon, then SHA-1. */
function digest(value: unknown): string {
	const out: number[] = [];
	const text = (s: string) => { for (let i = 0; i < s.length; i++) out.push(s.charCodeAt(i)); };
	const walk = (v: unknown): void => {
		if (v === null) return text('n');
		if (v === true) return text('t');
		if (v === false) return text('f');
		if (typeof v === 'number') {
			assert.ok(Number.isInteger(v), 'non-integer number in state');
			return text('i' + String(v) + ';');
		}
		if (typeof v === 'string') {
			const bytes = utf8Bytes(v);
			text('s' + bytes.length + ':');
			for (const b of bytes) out.push(b);
			return;
		}
		if (Array.isArray(v)) {
			text('[');
			for (const item of v) walk(item);
			return text(']');
		}
		const obj = v as Record<string, unknown>;
		text('{');
		for (const key of Object.keys(obj).sort(compareGoStrings)) {
			walk(key);
			walk(obj[key]);
		}
		text('}');
	};
	walk(value);
	return sha1HexBytes(out);
}

interface Loaded {
	go: GoHistory;
	history: GoldenHistory;
}

function loadGame(name: string): { def: GameDefinition; histories: Loaded[] } {
	const def = read('defs/' + name + '.def.json') as GameDefinition;
	const { Histories } = read('golden/' + name + '.golden.json') as { Histories: GoHistory[] };
	return { def, histories: Histories.map(go => ({ go, history: readGoldenHistory(join(REPO, go.File)) })) };
}

const viewersOf = (numPlayers: number) => [-2, -1, ...Array.from({ length: numPlayers }, (_, i) => i)];

/** Every stack-shaped object in a state blob, with its JSON path. */
function stackNodes(value: unknown, path = ''): { path: string; node: StackJSON }[] {
	if (Array.isArray(value)) return value.flatMap((v, i) => stackNodes(v, `${path}[${i}]`));
	if (value === null || typeof value !== 'object') return [];
	const obj = value as Record<string, unknown>;
	if (typeof obj.Deck === 'string' && 'Indexes' in obj) return [{ path, node: obj as unknown as StackJSON }];
	return Object.keys(obj).flatMap(key => stackNodes(obj[key], path ? `${path}.${key}` : key));
}

const games: Record<string, ReturnType<typeof loadGame>> = {
	memory: loadGame('memory'),
	blackjack: loadGame('blackjack'),
};

test('the golden corpus is the one expected', () => {
	assert.equal(games.memory.histories.length, 4);
	assert.equal(games.blackjack.histories.length, 5);
	const versions = (g: string) => games[g].histories.reduce((n, h) => n + h.go.Version + 1, 0);
	assert.equal(versions('memory'), 136);
	assert.equal(versions('blackjack'), 201);
});

for (const name of ['memory', 'blackjack']) {
	const { def, histories } = games[name];

	test(`${name}: the engine expands each golden file to the states Go expands it to`, () => {
		for (const { go, history } of histories) {
			assert.equal(history.Game.ID, go.GameID);
			assert.equal(history.States.length, go.Version + 1, go.File);
			history.States.forEach((state, v) => {
				assert.equal(digest(state), go.StateDigests[v], `${go.File} version ${v}`);
			});
		}
	});

	test(`${name}: every recorded id recomputes from the recorded salt and counts`, () => {
		let ids = 0;
		for (const { go, history } of histories) {
			const identity = { gameId: go.GameID, secretSalt: go.SecretSalt };
			history.States.forEach((state, v) => {
				// Directly: each physical stack's ids from its indexes. (Merged
				// stacks are views; the round trip below covers them.)
				for (const { path, node } of stackNodes(state)) {
					const merged = /\.(Cards|Hand)$/.test(path);
					if (merged) continue;
					(node.Indexes ?? []).forEach((deckIndex, slot) => {
						const expected = deckIndex < 0
							? ''
							: componentId(identity, node.Deck, deckIndex, secretMoveCountFor(state.SecretMoveCount, node.Deck, deckIndex));
						assert.equal((node.IDs ?? [])[slot], expected, `${go.File} v${v} ${path}[${slot}]`);
						ids++;
					});
				}
			});
		}
		assert.ok(ids > 5000, String(ids));
	});

	test(`${name}: a loaded state serializes back to the recorded JSON, and conserves components`, () => {
		for (const { go, history } of histories) {
			const identity = { gameId: go.GameID, secretSalt: go.SecretSalt };
			history.States.forEach((recorded, v) => {
				// load() discards recorded ids and merged views and checks
				// conservation; toJSON() recomputes both.
				const state = EngineState.load(def, identity, v, recorded);
				assert.equal(state.conservationError(), null, `${go.File} v${v}`);
				assert.deepEqual(state.toJSON(), recorded, `${go.File} v${v}`);
			});
		}
	});

	test(`${name}: viewFor matches Go for every viewer at every sampled version (full JSON)`, () => {
		let compared = 0;
		for (const { go, history } of histories) {
			const identity = { gameId: go.GameID, secretSalt: go.SecretSalt };
			for (const versionKey of Object.keys(go.Views)) {
				const v = Number(versionKey);
				const state = EngineState.load(def, identity, v, history.States[v]);
				for (const viewer of viewersOf(go.NumPlayers)) {
					assert.deepEqual(viewFor(state, viewer), go.Views[versionKey][String(viewer)], `${go.File} v${v} viewer ${viewer}`);
					compared++;
				}
			}
		}
		assert.ok(compared >= 20, String(compared));
	});

	test(`${name}: viewFor matches Go for every viewer at every version (digests)`, () => {
		let compared = 0;
		for (const { go, history } of histories) {
			const identity = { gameId: go.GameID, secretSalt: go.SecretSalt };
			history.States.forEach((recorded, v) => {
				const state = EngineState.load(def, identity, v, recorded, { skipConservation: true });
				for (const viewer of viewersOf(go.NumPlayers)) {
					assert.equal(digest(viewFor(state, viewer)), go.ViewDigests[v][String(viewer)], `${go.File} v${v} viewer ${viewer}`);
					compared++;
				}
			});
		}
		assert.ok(compared >= 136 * 4, String(compared));
	});

	test(`${name}: the only shuffles in the corpus are in set-up`, () => {
		// The set-up tests below reproduce version 0. If a history ever
		// records a shuffle at a later version, this fails so that coverage
		// is added rather than silently missing.
		for (const { go, history } of histories) {
			for (let v = 1; v < history.States.length; v++) {
				const before = stackNodes(history.States[v - 1]);
				const after = stackNodes(history.States[v]);
				after.forEach((s, i) => assert.equal(s.node.ShuffleCount, before[i].node.ShuffleCount, `${go.File} v${v} ${s.path}`));
			}
		}
	});
}

/** A stack as DistributeComponentToStarterStack leaves it: deck order, each id seen at version 0. */
function starterStack(identity: { gameId: string; secretSalt: string }, deck: string, size: number, maxSize?: number): StackJSON {
	const indexes = Array.from({ length: size }, (_, i) => i);
	const ids = indexes.map(i => componentId(identity, deck, i, 0));
	const seen: Record<string, number> = {};
	for (const id of ids) seen[id] = 0;
	const out: StackJSON = { Deck: deck, Indexes: indexes, IDs: ids, IDsLastSeen: seen, ShuffleCount: 0 };
	if (maxSize) out.MaxSize = maxSize;
	return out;
}

test('blackjack: the generator reproduces the recorded set-up shuffle from salt and version', () => {
	// blackjack's FinishSetUp is `game.DrawStack.Shuffle()` on a deck that
	// set-up distributed in deck order.
	const { def, histories } = games.blackjack;
	for (const { go, history } of histories) {
		const identity = { gameId: go.GameID, secretSalt: go.SecretSalt };
		const recorded = history.States[0];
		const before = cloneJSON(recorded);
		delete before.SecretMoveCount;
		before.Game.DrawStack = starterStack(identity, 'cards', def.Decks.cards.Size);
		const state = EngineState.load(def, identity, 0, before);
		state.gameStack('DrawStack').shuffle();
		assert.deepEqual(state.toJSON(), recorded, go.File);
		// And directly: the recorded order IS Perm(52) of the version-0 generator.
		assert.deepEqual((recorded.Game.DrawStack as StackJSON).Indexes, stateRand(go.GameID, go.SecretSalt, 0).perm(52), go.File);
	}
});

test('memory: the generator reproduces both recorded set-up shuffles from salt and version', () => {
	// A replay of memory's FinishSetUp (examples/memory/main.go): shuffle the
	// unused cards, deal pairs from the configured card set into the hidden
	// row, shuffle the hidden row. Two shuffles drawn from one generator,
	// with SlideToLastSlot and MoveToNextSlot between them.
	const { def, histories } = games.memory;
	const cards = def.Decks.cards.Components as { Type: string; CardSet: string }[];
	for (const { go, history } of histories) {
		const identity = { gameId: go.GameID, secretSalt: go.SecretSalt };
		const recorded = history.States[0];
		const before = cloneJSON(recorded);
		delete before.SecretMoveCount;
		before.Game.UnusedCards = starterStack(identity, 'cards', def.Decks.cards.Size);
		const hiddenSize = (recorded.Game.HiddenCards as StackJSON).Size ?? 0;
		before.Game.HiddenCards = {
			Deck: 'cards', Indexes: Array.from({ length: hiddenSize }, () => -1), IDs: [], IDsLastSeen: {}, ShuffleCount: 0, Size: hiddenSize,
		} as StackJSON;

		const state = EngineState.load(def, identity, 0, before);
		const unused = state.gameStack('UnusedCards');
		const hidden = state.gameStack('HiddenCards');
		const numCards = state.game.NumCards as number;
		const cardSet = state.game.CardSet as string;

		unused.shuffle();
		while (hidden.numComponents() < numCards) {
			const first = cards[unused.componentAt(0)!];
			let pairIndex = 0;
			for (let i = 1; i < unused.len(); i++) {
				if (cards[unused.componentAt(i)!].Type === first.Type) {
					pairIndex = i;
					break;
				}
			}
			assert.notEqual(pairIndex, 0, 'pair found');
			if (cardSet === 'all' || cardSet === first.CardSet) {
				unused.moveComponentToNextSlot(pairIndex, hidden);
				unused.moveComponentToNextSlot(0, hidden);
			} else {
				unused.slideToLastSlot(pairIndex);
				unused.slideToLastSlot(0);
			}
		}
		hidden.shuffle();

		assert.deepEqual(state.toJSON(), recorded, go.File);
	}
});

test('memory: the generator reproduces every recorded timer id', () => {
	// timer.Start draws its 16-character id from state.Rand() (timer.go,
	// game.go randomString): one Intn(16) per character, at the version of
	// the move that started it.
	const chars = 'ABCDEF0123456789';
	let checked = 0;
	for (const { go, history } of games.memory.histories) {
		for (let v = 1; v < history.States.length; v++) {
			const id = (history.States[v].Game.HideCardsTimer as { ID: string }).ID;
			const previous = (history.States[v - 1].Game.HideCardsTimer as { ID: string }).ID;
			if (id === '' || id === previous) continue;
			const rand = stateRand(go.GameID, go.SecretSalt, v);
			let expected = '';
			while (expected.length < 16) expected += chars.charAt(rand.intn(chars.length));
			assert.equal(id, expected, `${go.File} v${v}`);
			checked++;
		}
	}
	assert.ok(checked >= 10, String(checked));
});
