// The conformance runner: replays one document of conformance/ on the
// TypeScript engine and compares everything the document records.
//
// The engine is given only what an outside party supplied to the Go run: the
// game's identity (id, salt, players, variant, seating) and, in order, each
// player proposal, seat injection and timer firing. Every fix-up move and
// every set-up move has to come out of the engine itself.
//
// At every version the engine's stored state, move record, and each viewer's
// game object (with sanitized state), sanitized move, legality and projected
// choices must equal the recording, structurally. Nothing in the comparison
// knows which game it is looking at.
//
// Normalization. The document's own `normalization` list is honored by
// construction: fields it marks omitted (timestamps, Diagram) are never
// produced, and the fixed clock it names is the `now` the engine is given.
// ONE further rule is applied that the corpus does not declare, flagged in
// fabric/status/engine-rules.md: a `winners` / `Winners` list that is null
// in the recording equals an empty list (the exporter writes Go's nil slice
// as null at versions that were not the head when captured).

import type { CompiledGame, GameRecord, Input, MoveRecord, VersionOutput } from '../rules/index.ts';
import { animationKeys, createGame, gameDisplayName, moveNamePublic, parseInstant, propose, variantConfig, viewerOutput } from '../rules/index.ts';
import { PROPERTY_TYPE_NAMES } from '../rules/binding.ts';
import type { StateJSON } from '../state.ts';

type Json = null | boolean | number | string | Json[] | { [key: string]: Json };

/** The parts of a document the runner reads (conformance/schema.ts). */
export interface ConformanceDocument {
	name: string;
	normalization: { field: string; action: string; reason: string }[];
	game: {
		name: string; displayName: string; id: string; secretSalt: string; numPlayers: number;
		minNumPlayers: number; maxNumPlayers: number; defaultNumPlayers: number;
		variant: Record<string, string>; agents: string[]; seating: boolean;
		finished: boolean; winners: number[]; finalVersion: number;
	};
	engine: {
		clock: string; moveInputSchemaFingerprint: string; moveChoiceProjectionSchemaFingerprint: string;
		moveChoiceProjectionSchemaVersion: number; phaseEnum: string; groupEnum: string;
	};
	viewers: number[];
	moves: { name: string; helpText: string; isFixUp: boolean; namePublic: boolean; animationKeys: string[]; isGatheringStart: boolean; fields: { name: string; type: string; enumName?: string }[] }[];
	moveInputSchema: Json;
	moveChoiceProjectionSchema: Json;
	computedProperties: Json;
	variantConfig: Json;
	chest: { Decks: Record<string, { Index: number; Values: Json }[]>; Enums: Json; Constants: Json; LegalTemplates?: Record<string, string> };
	versions: DocumentVersion[];
}

interface DocumentVersion {
	version: number;
	settled: boolean;
	finished: boolean;
	winners: number[] | null;
	move: null | {
		name: string; version: number; initiator: number; phase: number; phaseName: string; proposer: number;
		arguments: Record<string, Json>; isFixUp: boolean; trigger: 'setup' | 'propose' | 'timer' | 'seat' | 'fixup';
	};
	state: Json;
	views: { viewer: number; game: Json; move: Json; legality: Json; projectedMoveChoices?: Json }[];
}

/** The fields compared, in report order. */
export const FIELDS = ['header', 'state', 'move', 'status', 'view.game', 'view.move', 'view.legality', 'view.choices'] as const;
export type Field = (typeof FIELDS)[number];

export interface FieldResult {
	compared: number;
	mismatches: number;
	/** The first few mismatches, each as "where: path: got X, want Y". */
	examples: string[];
}

export interface DocumentResult {
	name: string;
	versions: number;
	viewers: number;
	fields: Record<Field, FieldResult>;
	/** Set when the replay could not be carried through. */
	fatal?: string;
	passed: boolean;
}

function show(value: unknown): string {
	const text = JSON.stringify(value);
	if (text === undefined) return 'undefined';
	return text.length > 160 ? text.slice(0, 160) + '…' : text;
}

/** The path and values of the first structural difference, or null. */
export function firstDifference(got: unknown, want: unknown, path = ''): string | null {
	if (got === want) return null;
	if (got === null || want === null || typeof got !== 'object' || typeof want !== 'object') {
		return (path || '(root)') + ': got ' + show(got) + ', want ' + show(want);
	}
	if (Array.isArray(got) !== Array.isArray(want)) return (path || '(root)') + ': got ' + show(got) + ', want ' + show(want);
	if (Array.isArray(got) && Array.isArray(want)) {
		if (got.length !== want.length) return (path || '(root)') + ': length ' + got.length + ', want ' + want.length + ' (got ' + show(got) + ', want ' + show(want) + ')';
		for (let i = 0; i < got.length; i++) {
			const diff = firstDifference(got[i], want[i], path + '[' + i + ']');
			if (diff !== null) return diff;
		}
		return null;
	}
	const g = got as Record<string, unknown>;
	const w = want as Record<string, unknown>;
	const keys = Object.keys(g).filter(key => g[key] !== undefined);
	const wantKeys = Object.keys(w).filter(key => w[key] !== undefined);
	for (const key of wantKeys) if (!Object.prototype.hasOwnProperty.call(g, key) || g[key] === undefined) return path + '.' + key + ': missing, want ' + show(w[key]);
	for (const key of keys) if (!Object.prototype.hasOwnProperty.call(w, key) || w[key] === undefined) return path + '.' + key + ': unexpected ' + show(g[key]);
	for (const key of keys.sort()) {
		const diff = firstDifference(g[key], w[key], path + '.' + key);
		if (diff !== null) return diff;
	}
	return null;
}

/** The one undeclared normalization: a recorded null winners list is an empty one. */
function winnersOrEmpty<T>(value: T | null): T | never[] {
	return value === null ? [] : value;
}

function normalizeRecordedGame(game: Json): Json {
	if (game === null || typeof game !== 'object' || Array.isArray(game)) return game;
	return { ...game, Winners: winnersOrEmpty(game.Winners) };
}

function emptyResult(): Record<Field, FieldResult> {
	const fields = {} as Record<Field, FieldResult>;
	for (const field of FIELDS) fields[field] = { compared: 0, mismatches: 0, examples: [] };
	return fields;
}

/** The static tables a document's header records, derived from the game definition. */
function headerOf(game: CompiledGame, record: GameRecord): Record<string, unknown> {
	const spec = game.spec;
	const decks: Record<string, unknown> = {};
	for (const deckName of Object.keys(spec.decks)) {
		decks[deckName] = spec.decks[deckName].components.map((values, index) => ({ Index: index, Values: values }));
	}
	const enums: Record<string, unknown> = {};
	for (const e of game.enums.chestEnums()) enums[e.name] = { DefaultValue: e.def.DefaultValue, Values: e.def.Values };
	const computedTypes: Record<string, number> = { int: 1, bool: 2, string: 3, playerIndex: 4, intSlice: 6, boolSlice: 7, stringSlice: 8, playerIndexSlice: 9 };
	const computed = (spec.computed ?? [])
		.map(item => ({ EnumName: '', Name: item.name, Scope: item.scope, Type: computedTypes[item.type] }))
		.sort((a, b) => (a.Scope !== b.Scope ? (a.Scope < b.Scope ? -1 : 1) : a.Name < b.Name ? -1 : a.Name > b.Name ? 1 : 0));
	return {
		game: {
			name: spec.name, displayName: gameDisplayName(spec), minNumPlayers: spec.minNumPlayers, maxNumPlayers: spec.maxNumPlayers,
			defaultNumPlayers: spec.defaultNumPlayers, numPlayers: record.numPlayers, variant: record.variant ?? {}, agents: record.agents,
		},
		engine: {
			moveInputSchemaFingerprint: game.moveInputSchemaFingerprint,
			moveChoiceProjectionSchemaFingerprint: game.choiceSchemaFingerprint,
			moveChoiceProjectionSchemaVersion: 1,
			phaseEnum: game.phaseEnum ? game.phaseEnum.name : '',
			groupEnum: game.def.GroupEnum ?? '',
		},
		moves: game.moves.map(move => ({
			name: move.spec.name,
			helpText: move.spec.helpText,
			isFixUp: move.spec.fixUp,
			namePublic: moveNamePublic(move),
			animationKeys: animationKeys(move),
			isGatheringStart: move.spec.gatheringStart === true,
			fields: move.fieldNames.map(name => {
				const field = move.fields[name];
				const out: Record<string, unknown> = { name, type: PROPERTY_TYPE_NAMES[field.type] };
				if (field.type === 'enum' && field.enum) out.enumName = field.enum;
				return out;
			}),
		})),
		moveInputSchema: game.moveInputSchema,
		moveChoiceProjectionSchema: game.choiceSchema.length === 0 ? null : game.choiceSchema,
		computedProperties: computed,
		variantConfig: variantConfig(spec),
		chest: { Decks: decks, Enums: enums, Constants: spec.constants ?? null, LegalTemplates: game.templates },
	};
}

function recordedHeader(doc: ConformanceDocument): Record<string, unknown> {
	const { name, displayName, minNumPlayers, maxNumPlayers, defaultNumPlayers, numPlayers, variant, agents } = doc.game;
	const { moveInputSchemaFingerprint, moveChoiceProjectionSchemaFingerprint, moveChoiceProjectionSchemaVersion, phaseEnum, groupEnum } = doc.engine;
	return {
		game: { name, displayName, minNumPlayers, maxNumPlayers, defaultNumPlayers, numPlayers, variant, agents },
		engine: { moveInputSchemaFingerprint, moveChoiceProjectionSchemaFingerprint, moveChoiceProjectionSchemaVersion, phaseEnum, groupEnum },
		moves: doc.moves,
		moveInputSchema: doc.moveInputSchema,
		moveChoiceProjectionSchema: doc.moveChoiceProjectionSchema,
		computedProperties: doc.computedProperties,
		variantConfig: doc.variantConfig,
		chest: doc.chest,
	};
}

/** Replays one document on the engine and compares every recorded field. */
export function runDocument(game: CompiledGame, doc: ConformanceDocument): DocumentResult {
	const fields = emptyResult();
	const result: DocumentResult = { name: doc.name, versions: doc.versions.length, viewers: doc.viewers.length, fields, passed: false };
	const check = (field: Field, where: string, got: unknown, want: unknown): void => {
		fields[field].compared++;
		const diff = firstDifference(got, want);
		if (diff === null) return;
		fields[field].mismatches++;
		if (fields[field].examples.length < 3) fields[field].examples.push(where + ': ' + diff);
	};
	const finish = (fatal?: string): DocumentResult => {
		if (fatal !== undefined) result.fatal = fatal;
		result.passed = fatal === undefined && FIELDS.every(field => fields[field].mismatches === 0);
		return result;
	};

	const now = parseInstant(doc.engine.clock);
	if (now === null) return finish('unreadable clock ' + doc.engine.clock);

	const created = createGame(game, {
		id: doc.game.id,
		secretSalt: doc.game.secretSalt,
		numPlayers: doc.game.numPlayers,
		variant: doc.game.variant,
		agents: doc.game.agents,
		seating: doc.game.seating,
		now,
	});
	if (created.record === null) return finish('createGame refused: ' + created.error);
	if (created.error !== undefined) return finish('set-up failed: ' + created.error);
	const record = created.record;

	check('header', 'header', headerOf(game, record), recordedHeader(doc));

	const history: MoveRecord[] = [];
	let previousState: StateJSON | undefined;
	let currentState: StateJSON | undefined;

	// Compares the versions one step produced. `trigger` is what the first of them answers to.
	const compareStep = (versions: VersionOutput[], trigger: string): string | undefined => {
		for (let i = 0; i < versions.length; i++) {
			const produced = versions[i];
			const want = doc.versions[produced.version];
			const where = 'v' + produced.version;
			if (!want) return 'the engine produced version ' + produced.version + ', past the end of the recording';
			const settled = i === versions.length - 1;
			if (produced.move !== null) history.push(produced.move);

			check('state', where, produced.state, want.state);
			const move = produced.move;
			const compiled = move === null ? null : game.movesByName[move.Name];
			check(
				'move',
				where,
				move === null
					? null
					: {
							name: move.Name,
							version: move.Version,
							initiator: move.Initiator,
							phase: move.Phase,
							phaseName: game.phaseEnum ? game.phaseEnum.stringOf(move.Phase) : '',
							proposer: move.Proposer,
							arguments: move.Blob,
							isFixUp: compiled ? compiled.spec.fixUp : false,
							trigger: move.Initiator === move.Version ? trigger : 'fixup',
						},
				want.move,
			);
			check(
				'status',
				where,
				{ settled, finished: produced.finished, winners: produced.finished ? produced.winners : [] },
				{ settled: want.settled, finished: want.finished, winners: winnersOrEmpty(want.winners) },
			);

			const snapshot = { record, history: history.slice(), state: produced.state, previousState: currentState, settled, now };
			for (const recorded of want.views) {
				const view = viewerOutput(game, snapshot, recorded.viewer);
				const at = where + ' viewer ' + recorded.viewer;
				check('view.game', at, view.game, normalizeRecordedGame(recorded.game));
				check('view.move', at, view.move, recorded.move);
				check('view.legality', at, view.legality, recorded.legality);
				check('view.choices', at, view.projectedMoveChoices, recorded.projectedMoveChoices);
			}
			previousState = currentState;
			currentState = produced.state;
		}
		return undefined;
	};

	let fatal = compareStep(created.versions, 'setup');
	if (fatal !== undefined) return finish(fatal);

	while (history.length + 1 < doc.versions.length) {
		const next = doc.versions[history.length + 1];
		const move = next.move;
		if (move === null) return finish('version ' + next.version + ' records no move');
		let input: Input;
		if (move.trigger === 'propose') {
			// Only the fields a proposer supplies; the engine must derive the rest.
			const compiled = game.movesByName[move.name];
			const supplied: Record<string, unknown> = {};
			if (compiled) {
				for (const field of compiled.inputFields) {
					if (field.disposition !== 'required' && field.disposition !== 'server-defaulted') continue;
					if (Object.prototype.hasOwnProperty.call(move.arguments, field.name)) supplied[field.name] = move.arguments[field.name];
				}
			}
			input = { kind: 'propose', name: move.name, input: supplied, proposer: move.proposer, expectedVersion: history.length };
		} else if (move.trigger === 'seat') {
			input = { kind: 'seat', player: move.arguments.TargetPlayerIndex as number };
		} else if (move.trigger === 'timer') {
			// The recording's clock never advances; its timers are forced.
			input = { kind: 'timer', force: true };
		} else {
			return finish('version ' + next.version + ' (' + move.name + ', ' + move.trigger + ') should have followed version ' + history.length + ', but the engine stopped there');
		}
		const step = propose(game, { record, history, state: currentState as StateJSON }, input, { now });
		if (step.error !== undefined) return finish('version ' + next.version + ' (' + move.name + '): the engine refused the recorded ' + move.trigger + ': ' + step.error);
		if (step.versions.length === 0) return finish('version ' + next.version + ' (' + move.name + '): the engine applied nothing');
		fatal = compareStep(step.versions, move.trigger);
		if (fatal !== undefined) return finish(fatal);
	}
	void previousState;
	return finish();
}

/** One row per document, one column per field: "ok" or "mismatches/compared". */
export function passTable(results: readonly DocumentResult[]): string {
	const header = ['document', 'versions', 'viewers', ...FIELDS, 'result'];
	const rows = results.map(result => [
		result.name,
		String(result.versions),
		String(result.viewers),
		...FIELDS.map(field => (result.fields[field].mismatches === 0 ? 'ok ' + result.fields[field].compared : result.fields[field].mismatches + '/' + result.fields[field].compared + ' FAIL')),
		result.passed ? 'PASS' : 'FAIL' + (result.fatal ? ': ' + result.fatal : ''),
	]);
	return [header, ...rows].map(row => '| ' + row.join(' | ') + ' |').join('\n');
}
