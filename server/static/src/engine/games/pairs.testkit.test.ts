// What games/pairs.test.ts, scripts/pairs-proof.mjs and
// scripts/pairs-mutants.mjs share: a game built through gameFromData from
// its exported definition (conformance/definitions/<game>.json) with a
// bodies table made of the stored pairs' accepted.ts functions
// (boardgame-util/lib/transpile/testdata/pairs/, imported through the
// store's index.ts), and the game's whole recorded corpus: its conformance
// documents (conformance/index.json) and its checked-in fuzz documents
// (conformance/fuzz/<game>/), each replayed by testing/fuzz.ts, which also
// re-proposes every recorded rejection.
//
// It does not import the hand ports (games/<game>.ts): the store is what
// preserves them once they are removed. The one FnRef no pair can serve is
// werewolf's empty SetUp.Distribute (Go's werewolf declares no
// DistributeComponentToStarterStack); it gets STAND_INS' body, Go's base
// delegate's error. A MoveComponentToSlot the registry cannot build from the
// exported configuration gets the registry's parts with Go's discovered slot
// field through the `stock:<name>` hook (resolvedSlotField).
//
// This file declares no test.
import { existsSync, readFileSync, readdirSync } from 'node:fs';
import path from 'node:path';
import { fileURLToPath, pathToFileURL } from 'node:url';
import { gunzipSync } from 'node:zlib';
import { compileGame } from '../rules/compile.ts';
import { type Bodies, type GameData, type MoveData, gameFromData, stockOverrideRef } from '../rules/data.ts';
import type { CompiledGame } from '../rules/state.ts';
import { CONFIG_PREFIX, type StockParts, stockMove } from '../rules/stockRegistry.ts';
import type { GameSpec } from '../rules/types.ts';
import { type FuzzDocument, type FuzzResult, runFuzzDocument } from '../testing/fuzz.ts';

export type AnyBody = (...args: never[]) => unknown;

/** The store's index.ts (generated with the store). */
export interface StoreIndex {
	accepted(game: string, ref: string): AnyBody | undefined;
	pairRefs(game: string): string[];
}

let loadedIndex: StoreIndex | undefined;

/**
 * The store's index, loaded once. The store is outside server/static, the
 * root of the test project's tsconfig, so it is imported by a computed
 * specifier that the type checker does not follow; the store is
 * type-checked by its own tsconfig.json instead (see there).
 */
export async function storeIndex(): Promise<StoreIndex> {
	if (loadedIndex === undefined) {
		const file = path.join(repoRoot(), 'boardgame-util', 'lib', 'transpile', 'testdata', 'pairs', 'index.ts');
		loadedIndex = (await import(pathToFileURL(file).href)) as StoreIndex;
	}
	return loadedIndex;
}

export const PAIR_GAMES = ['memory', 'blackjack', 'tictactoe', 'pig', 'debuganimations', 'werewolf'] as const;
export type PairGame = (typeof PAIR_GAMES)[number];

export function repoRoot(): string {
	let dir = path.dirname(fileURLToPath(import.meta.url));
	for (;;) {
		if (existsSync(path.join(dir, 'conformance', 'index.json'))) return dir;
		const parent = path.dirname(dir);
		if (parent === dir) throw new Error('could not find conformance/ above ' + import.meta.url);
		dir = parent;
	}
}

export function definitionOf(game: string): GameData {
	return JSON.parse(readFileSync(path.join(repoRoot(), 'conformance', 'definitions', game + '.json'), 'utf8')) as GameData;
}

/** Whether rules/stockRegistry.ts builds this move's stock type yet. */
function stockAvailable(move: MoveData): boolean {
	if (move.Stock === null) return true;
	try {
		return stockMove(move.Stock.Type, move.Stock.Config) !== undefined;
	} catch {
		return false;
	}
}

/**
 * The registry's MoveComponentToSlot for a move configured without
 * WithSlotField, with the slot field Go discovers at boot: the move's only
 * int field (moves/move_component_to_slot.go resolveSlotField). The exported
 * configuration does not carry it, so stockMove refuses the move
 * (fabric/transpile/status/integration-1.md); this supplies it through the
 * `stock:` hook. Undefined for any other move.
 */
function resolvedSlotField(move: MoveData): StockParts | undefined {
	if (move.Stock === null || move.Stock.Type !== 'MoveComponentToSlot') return undefined;
	if (Object.prototype.hasOwnProperty.call(move.Stock.Config, CONFIG_PREFIX + 'SlotField')) return undefined;
	const ints = Object.keys(move.Fields).filter(name => move.Fields[name].Type === 'int');
	if (ints.length !== 1) return undefined;
	try {
		return stockMove(move.Stock.Type, { ...move.Stock.Config, [CONFIG_PREFIX + 'SlotField']: ints[0] });
	} catch {
		return undefined;
	}
}

/**
 * Bodies for what no pair can serve, by game and FnRef: werewolf's empty
 * SetUp.Distribute, where Go's base delegate (base/game_delegate.go) returns
 * this error; the engine never calls it (werewolf has no deck).
 */
function standIn(game: string, ref: string): AnyBody | undefined {
	if (game === 'werewolf' && ref === '') {
		return (): never => {
			throw new Error('DistributeComponentToStarterStack was called, but the component was not stored in a stack');
		};
	}
	return undefined;
}

export interface StoreBodies {
	bodies: Bodies;
	/** FnRefs the loader asked for, served by a stored pair. */
	served: string[];
	/** FnRefs (or the empty FnRef) the loader asked for that no pair serves, given standIn's body. */
	fallbacks: string[];
	/** Moves whose stock type the registry cannot build even with the discovered slot field (no parts given). */
	stockStandIns: string[];
	/** Moves given the registry's parts, with Go's discovered slot field, through `stock:`. */
	stockResolved: string[];
	/** FnRefs replaced by `substitutions`. */
	substituted: string[];
}

/**
 * The bodies table of a game made of the store's accepted functions, with
 * `substitutions` (FnRef to a function) in their place. Which FnRefs were
 * served by what is recorded as gameFromData asks for them.
 */
export function storeBodies(index: StoreIndex, game: string, data: GameData, substitutions: ReadonlyMap<string, AnyBody> = new Map()): StoreBodies {
	const out: StoreBodies = { bodies: () => undefined, served: [], fallbacks: [], stockStandIns: [], stockResolved: [], substituted: [] };
	const note = (list: string[], ref: string): void => {
		if (!list.includes(ref)) list.push(ref);
	};
	const standIns = new Map<string, StockParts>();
	const resolvedNames = new Set<string>();
	for (const move of data.Moves) {
		if (stockAvailable(move)) continue;
		const resolved = resolvedSlotField(move);
		if (resolved !== undefined) {
			standIns.set(stockOverrideRef(move.Name), resolved);
			resolvedNames.add(stockOverrideRef(move.Name));
			continue;
		}
		note(out.stockStandIns, move.Name);
	}
	out.bodies = ref => {
		const parts = standIns.get(ref);
		if (parts !== undefined) {
			note(resolvedNames.has(ref) ? out.stockResolved : out.stockStandIns, ref.replace(/^stock:/, ''));
			return (): StockParts => parts;
		}
		const substitute = substitutions.get(ref);
		if (substitute !== undefined) {
			note(out.substituted, ref);
			return substitute;
		}
		const stored = index.accepted(game, ref);
		if (stored !== undefined) {
			note(out.served, ref);
			return stored;
		}
		const fallback = standIn(game, ref);
		if (fallback !== undefined) {
			note(out.fallbacks, ref);
			return fallback;
		}
		return undefined;
	};
	return out;
}

/** The game built from its definition and the store (plus substitutions), compiled. */
export async function storeGame(game: string, substitutions: ReadonlyMap<string, AnyBody> = new Map()): Promise<{ game: CompiledGame; spec: () => GameSpec; table: StoreBodies }> {
	const data = definitionOf(game);
	const table = storeBodies(await storeIndex(), game, data, substitutions);
	const spec = (): GameSpec => gameFromData(data, table.bodies);
	return { game: compileGame(spec()), spec, table };
}

export interface CorpusDocument {
	/** 'conformance/<file>' or 'conformance/fuzz/<game>/<file>', relative to the repository root. */
	file: string;
	kind: 'conformance' | 'fuzz';
}

/** Every recorded document of a game: conformance corpus first, then the fuzz documents, each in name order. */
export function corpusDocuments(game: string): CorpusDocument[] {
	const root = repoRoot();
	const index = JSON.parse(readFileSync(path.join(root, 'conformance', 'index.json'), 'utf8')) as { file: string; game: string }[];
	const out: CorpusDocument[] = index.filter(entry => entry.game === game).map(entry => ({ file: 'conformance/' + entry.file, kind: 'conformance' as const }));
	const fuzzDir = path.join(root, 'conformance', 'fuzz', game);
	if (existsSync(fuzzDir)) {
		for (const file of readdirSync(fuzzDir).sort()) if (/^seed-\d+\.json\.gz$/.test(file)) out.push({ file: 'conformance/fuzz/' + game + '/' + file, kind: 'fuzz' });
	}
	return out;
}

export function readDocument(file: string): FuzzDocument {
	return JSON.parse(gunzipSync(readFileSync(path.join(repoRoot(), file))).toString('utf8')) as FuzzDocument;
}

export interface DocumentCheck {
	file: string;
	passed: boolean;
	versions: number;
	rejections: number;
	/** Why it failed: the first problems, and the divergence report. */
	problems: string[];
	result: FuzzResult;
}

/**
 * Replays one document. It passes when every field of every version and
 * viewer matches, every recorded rejection is refused with Go's text, and
 * every version, viewer and rejection was compared.
 */
export function checkDocument(game: CompiledGame, file: string, doc: FuzzDocument, spec?: () => GameSpec): DocumentCheck {
	const result = runFuzzDocument(game, doc, spec === undefined ? {} : { spec });
	const problems: string[] = [];
	if (result.document.fatal) problems.push('replay stopped: ' + result.document.fatal);
	for (const field of Object.keys(result.document.fields) as (keyof typeof result.document.fields)[]) {
		const f = result.document.fields[field];
		if (f.mismatches > 0) problems.push(field + ': ' + f.mismatches + ' of ' + f.compared + ' differ, e.g. ' + f.examples.slice(0, 2).join('; '));
	}
	if (result.rejections.mismatches > 0) problems.push('rejections: ' + result.rejections.mismatches + ' of ' + result.rejections.compared + ' differ, e.g. ' + result.rejections.examples.slice(0, 2).join('; '));
	if (!result.document.fatal && result.document.fields.state.compared !== doc.versions.length) problems.push('compared ' + result.document.fields.state.compared + ' of ' + doc.versions.length + ' versions');
	if (!result.document.fatal && result.document.fields['view.game'].compared !== doc.versions.length * doc.viewers.length) problems.push('compared ' + result.document.fields['view.game'].compared + ' of ' + doc.versions.length * doc.viewers.length + ' viewer-versions');
	if (result.rejections.compared !== (doc.rejections ?? []).length) problems.push('compared ' + result.rejections.compared + ' of ' + (doc.rejections ?? []).length + ' rejections');
	if (result.report) problems.push('divergence report: ' + JSON.stringify(result.report));
	return { file, passed: problems.length === 0, versions: doc.versions.length, rejections: result.rejections.compared, problems, result };
}

/** The FnRefs the store holds a pair for, of one game. */
export async function storedRefs(game: string): Promise<string[]> {
	return (await storeIndex()).pairRefs(game);
}
