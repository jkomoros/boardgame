// The fuzz replay: one document of conformance/fuzz/ (or of a transient
// `boardgame-util fuzz run --out DIR`) on the TypeScript engine.
//
// A fuzz document is a conformance document plus `rejections`: outside
// inputs Go refused, each between two versions. The versions and views are
// replayed and compared by `runDocument` (./conformance.ts), unchanged. This
// module adds the rejections: it feeds the same outside inputs a second
// time and, at each rejection's `atVersion`, makes the recorded proposal
// and requires Go's exact error text, no versions returned, and the
// caller's state and history untouched.
//
// When anything differs, `runFuzzDocument` builds the proposal's divergence
// report (fabric/TRANSPILE-PROPOSAL.md, "The divergence report"): the first
// difference in the runner's order, with the JSON path, the values, the
// move that produced the version, the last outside input, the bodies that
// ran (when the caller supplies an instrumentable spec), and the command
// line that regenerates the minimal document.
import type { CompiledGame, GameRecord, GameSpec, Input, MoveRecord, VersionOutput } from '../rules/index.ts';
import { compileGame, createGame, parseInstant, propose, viewerOutput } from '../rules/index.ts';
import type { StateJSON } from '../state.ts';
import { cloneJSON } from '../state.ts';
import { type ConformanceDocument, type DocumentResult, FIELDS, type Field, firstDifference, runDocument } from './conformance.ts';

type Json = null | boolean | number | string | Json[] | { [key: string]: Json };

/** One refused outside input (conformance/schema.ts `Rejection`). */
export interface Rejection {
	atVersion: number;
	kind: 'propose' | 'timer' | 'seat';
	name: string;
	input: Record<string, Json>;
	proposer: number;
	expectedVersion: number;
	error: string;
}

/** A conformance document as the fuzz driver writes it. */
export interface FuzzDocument extends ConformanceDocument {
	schemaVersion?: number;
	source?: { kind: string; description: string; covers?: string };
	rejections?: Rejection[];
}

/** Called with a body's name each time it runs; see `instrumentBodies`. */
export type BodyCounter = (name: string) => void;

export interface DivergenceReport {
	document: string;
	version: number;
	/** The viewer whose view differed; absent for state, move, status, header and rejections. */
	viewer?: number;
	/** A field of FIELDS, 'rejection', or 'replay' when the replay could not be carried on. */
	field: Field | 'rejection' | 'replay';
	/** The JSON path inside the field ('' for the whole value). */
	path: string;
	want: unknown;
	got: unknown;
	/** The move that produced `version` (or the refused proposal, for a rejection). */
	move?: { name: string; proposer: number; trigger: string };
	/** The last proposal, seat or timer fed before or at `version`. */
	lastOutsideInput?: { atVersion: number; name: string; input: Record<string, Json>; trigger: string };
	/** Bodies run by the step that produced `version`, in first-run order. Present only with `bodies`. */
	bodiesExecuted?: string[];
	/** The command that regenerates the minimal document in Go. */
	repro: string;
}

export interface RejectionResult {
	compared: number;
	mismatches: number;
	/** The first few mismatches, as text. */
	examples: string[];
}

export interface FuzzResult {
	name: string;
	document: DocumentResult;
	rejections: RejectionResult;
	/** The number of rejections the document records. */
	recordedRejections: number;
	passed: boolean;
	/** Set when anything differed. */
	report?: DivergenceReport;
}

export interface FuzzOptions {
	/**
	 * The game's spec factory. When given and the document fails, the
	 * replay is repeated with every body wrapped by `instrumentBodies` to
	 * fill the report's `bodiesExecuted`.
	 */
	spec?: () => GameSpec;
}

// ---------------------------------------------------------------------------
// Bodies

function isPlainObject(value: unknown): value is Record<string, unknown> {
	if (value === null || typeof value !== 'object') return false;
	const proto = Object.getPrototypeOf(value);
	return proto === Object.prototype || proto === null;
}

function wrapValue(value: unknown, name: string, count: BodyCounter): unknown {
	if (typeof value === 'function') {
		const body = value as (...args: unknown[]) => unknown;
		return function (this: unknown, ...args: unknown[]): unknown {
			count(name);
			return body.apply(this, args);
		};
	}
	if (Array.isArray(value)) {
		return value.map((item, i) => {
			const label = isPlainObject(item) && typeof item.name === 'string' ? item.name : String(i);
			return wrapValue(item, name + '[' + label + ']', count);
		});
	}
	if (isPlainObject(value)) {
		const out: Record<string, unknown> = {};
		for (const key of Object.keys(value)) out[key] = wrapValue(value[key], name === '' ? key : name + '.' + key, count);
		return out;
	}
	return value;
}

/**
 * A copy of `spec` in which every function (a move's `legal` and `apply`,
 * set-up, end, computed properties, anything else the spec holds) calls
 * `count` with its path, such as "moves[Reveal Card].apply", before it runs.
 * The copy behaves exactly as the original.
 */
export function instrumentBodies(spec: GameSpec, count: BodyCounter): GameSpec {
	return wrapValue(spec, '', count) as GameSpec;
}

// ---------------------------------------------------------------------------
// Replay of the outside inputs

interface Produced {
	output: VersionOutput;
	settled: boolean;
	/** Length of the history at this version. */
	historyLength: number;
	previousState: StateJSON | undefined;
}

interface Replay {
	record: GameRecord | null;
	history: MoveRecord[];
	produced: Produced[];
	now: number;
	/** Why the replay stopped early. */
	fatal?: string;
	fatalVersion?: number;
}

/** What the replay feeds for a recorded move, as runDocument feeds it. */
function inputFor(game: CompiledGame, move: NonNullable<ConformanceDocument['versions'][number]['move']>, expectedVersion: number): Input | null {
	if (move.trigger === 'propose') {
		const compiled = game.movesByName[move.name];
		const supplied: Record<string, unknown> = {};
		if (compiled) {
			for (const field of compiled.inputFields) {
				if (field.disposition !== 'required' && field.disposition !== 'server-defaulted') continue;
				if (Object.prototype.hasOwnProperty.call(move.arguments, field.name)) supplied[field.name] = move.arguments[field.name];
			}
		}
		return { kind: 'propose', name: move.name, input: supplied, proposer: move.proposer, expectedVersion };
	}
	if (move.trigger === 'seat') return { kind: 'seat', player: move.arguments.TargetPlayerIndex as number };
	if (move.trigger === 'timer') return { kind: 'timer', force: true };
	return null;
}

/** The input a rejection records. */
export function rejectionInput(rejection: Rejection): Input {
	if (rejection.kind === 'timer') return { kind: 'timer', force: true };
	if (rejection.kind === 'seat') {
		const player = rejection.input.TargetPlayerIndex;
		return { kind: 'seat', player: typeof player === 'number' ? player : -1 };
	}
	return { kind: 'propose', name: rejection.name, input: rejection.input, proposer: rejection.proposer, expectedVersion: rejection.expectedVersion };
}

/**
 * Feeds the document's outside inputs. After every settled version,
 * `atSettled` is called with that version and the snapshot; it may make
 * proposals of its own but must not change the replay's state or history.
 * With `stopAfter`, the replay ends once that version is produced.
 */
function replay(
	game: CompiledGame,
	doc: ConformanceDocument,
	atSettled: (version: number, record: GameRecord, history: readonly MoveRecord[], state: StateJSON, now: number) => void,
	stopAfter?: number,
	beforeStep?: (firstVersion: number) => void,
): Replay {
	const now = parseInstant(doc.engine.clock);
	const out: Replay = { record: null, history: [], produced: [], now: now ?? 0 };
	if (now === null) return { ...out, fatal: 'unreadable clock ' + doc.engine.clock, fatalVersion: 0 };
	beforeStep?.(0);
	const created = createGame(game, {
		id: doc.game.id, secretSalt: doc.game.secretSalt, numPlayers: doc.game.numPlayers, variant: doc.game.variant,
		agents: doc.game.agents, seating: doc.game.seating, now,
	});
	if (created.record === null) return { ...out, fatal: 'createGame refused: ' + created.error, fatalVersion: 0 };
	if (created.error !== undefined) return { ...out, fatal: 'set-up failed: ' + created.error, fatalVersion: 0 };
	out.record = created.record;
	let current: StateJSON | undefined;
	const take = (versions: VersionOutput[]): void => {
		versions.forEach((output, i) => {
			if (output.move !== null) out.history.push(output.move);
			out.produced.push({ output, settled: i === versions.length - 1, historyLength: out.history.length, previousState: current });
			current = output.state;
		});
	};
	take(created.versions);
	const done = (): boolean => stopAfter !== undefined && out.produced.length > stopAfter;
	if (done()) return out;
	atSettled(out.history.length, out.record, out.history, current as StateJSON, now);
	while (out.history.length + 1 < doc.versions.length) {
		const next = doc.versions[out.history.length + 1];
		if (next.move === null) return { ...out, fatal: 'version ' + next.version + ' records no move', fatalVersion: next.version };
		const input = inputFor(game, next.move, out.history.length);
		if (input === null) return { ...out, fatal: 'version ' + next.version + ' (' + next.move.name + ', ' + next.move.trigger + ') should have followed version ' + out.history.length + ', but the engine stopped there', fatalVersion: next.version };
		beforeStep?.(next.version);
		const step = propose(game, { record: out.record, history: out.history, state: current as StateJSON }, input, { now });
		if (step.error !== undefined) return { ...out, fatal: 'version ' + next.version + ' (' + next.move.name + '): the engine refused the recorded ' + next.move.trigger + ': ' + step.error, fatalVersion: next.version };
		if (step.versions.length === 0) return { ...out, fatal: 'version ' + next.version + ' (' + next.move.name + '): the engine applied nothing', fatalVersion: next.version };
		take(step.versions);
		if (done()) return out;
		atSettled(out.history.length, out.record, out.history, current as StateJSON, now);
	}
	return out;
}

// ---------------------------------------------------------------------------
// The divergence report

interface Located {
	version: number;
	viewer?: number;
	field: Field | 'rejection' | 'replay';
	path: string;
	/** The text runDocument kept, used when a value cannot be recomputed. */
	text: string;
	/** Order key: version, then the runner's comparison order within it. */
	order: [number, number, number, number];
}

const EXAMPLE = /^(header|v(\d+)(?: viewer (-?\d+))?): (\(root\)|[^:]*): (.*)$/;

function locate(doc: ConformanceDocument, field: Field, example: string): Located | null {
	const match = EXAMPLE.exec(example);
	if (match === null) return null;
	const isHeader = match[1] === 'header';
	const version = isHeader ? -1 : Number(match[2]);
	const viewer = match[3] === undefined ? undefined : Number(match[3]);
	const path = match[4] === '(root)' ? '' : match[4];
	const fieldIndex = FIELDS.indexOf(field);
	const viewerPos = viewer === undefined ? -1 : doc.viewers.indexOf(viewer);
	const stage = field.startsWith('view.') ? 1 : 0;
	const located: Located = { version, field, path, text: match[5], order: [version, stage, viewerPos, fieldIndex] };
	if (viewer !== undefined) located.viewer = viewer;
	return located;
}

function before(a: Located, b: Located): boolean {
	for (let i = 0; i < a.order.length; i++) if (a.order[i] !== b.order[i]) return a.order[i] < b.order[i];
	return false;
}

/** The value at a path such as ".CurrentState.Players[0].Hand"; undefined when it cannot be followed. */
export function valueAt(value: unknown, path: string): unknown {
	const tokens = /\.([^.[\]]+)|\[(\d+)\]/g;
	let at = value;
	let consumed = 0;
	for (let m = tokens.exec(path); m !== null; m = tokens.exec(path)) {
		if (m.index !== consumed) return undefined;
		consumed = m.index + m[0].length;
		if (at === null || typeof at !== 'object') return undefined;
		at = m[1] !== undefined ? (at as Record<string, unknown>)[m[1]] : (at as unknown[])[Number(m[2])];
	}
	return consumed === path.length ? at : undefined;
}

function winnersOrEmpty<T>(value: T | null): T | never[] {
	return value === null ? [] : value;
}

/** The engine's and the recording's value of one field at one version, as runDocument compares them. */
function fieldValues(game: CompiledGame, doc: ConformanceDocument, run: Replay, located: Located): { got: unknown; want: unknown } | null {
	const produced = run.produced[located.version];
	const want = doc.versions[located.version];
	if (produced === undefined || want === undefined || run.record === null) return null;
	const move = produced.output.move;
	switch (located.field) {
		case 'state':
			return { got: produced.output.state, want: want.state };
		case 'move': {
			const compiled = move === null ? null : game.movesByName[move.Name];
			const trigger = want.move?.trigger ?? '';
			const got = move === null ? null : {
				name: move.Name, version: move.Version, initiator: move.Initiator, phase: move.Phase,
				phaseName: game.phaseEnum ? game.phaseEnum.stringOf(move.Phase) : '', proposer: move.Proposer, arguments: move.Blob,
				isFixUp: compiled ? compiled.spec.fixUp : false, trigger: move.Initiator === move.Version ? trigger : 'fixup',
			};
			return { got, want: want.move };
		}
		case 'status':
			return {
				got: { settled: produced.settled, finished: produced.output.finished, winners: produced.output.finished ? produced.output.winners : [] },
				want: { settled: want.settled, finished: want.finished, winners: winnersOrEmpty(want.winners) },
			};
		case 'view.game':
		case 'view.move':
		case 'view.legality':
		case 'view.choices': {
			const recorded = want.views.find(view => view.viewer === located.viewer);
			if (recorded === undefined || located.viewer === undefined) return null;
			const snapshot = {
				record: run.record, history: run.history.slice(0, produced.historyLength), state: produced.output.state,
				previousState: produced.previousState, settled: produced.settled, now: run.now,
			};
			const view = viewerOutput(game, snapshot, located.viewer);
			if (located.field === 'view.game') {
				const g = recorded.game;
				const normalized = g !== null && typeof g === 'object' && !Array.isArray(g) ? { ...g, Winners: winnersOrEmpty(g.Winners) } : g;
				return { got: view.game, want: normalized };
			}
			if (located.field === 'view.move') return { got: view.move, want: recorded.move };
			if (located.field === 'view.legality') return { got: view.legality, want: recorded.legality };
			return { got: view.projectedMoveChoices, want: recorded.projectedMoveChoices };
		}
		default:
			return null;
	}
}

/** The bounds a fuzz document was generated with, read from its source description. */
function sourceBounds(doc: FuzzDocument): { game: string; seed: string; policy: string; maxProposals?: string; refusalRate?: string } {
	const description = doc.source?.description ?? '';
	const pick = (re: RegExp): string | undefined => re.exec(description)?.[1];
	const nameSeed = /\/seed-0*(\d+)$/.exec(doc.name)?.[1];
	return {
		game: pick(/game (\S+?),/) ?? doc.game.name,
		seed: pick(/seed (\d+)/) ?? nameSeed ?? '?',
		policy: pick(/policy (\S+?),/) ?? 'random-legal',
		maxProposals: pick(/max proposals (\d+)/),
		refusalRate: pick(/refusal rate ([0-9.eE+-]+)/),
	};
}

/** `boardgame-util fuzz run` for the minimal document that still reaches `version`. */
export function reproCommand(doc: FuzzDocument, version: number): string {
	const b = sourceBounds(doc);
	const parts = ['boardgame-util fuzz run', '--game ' + b.game, '--seed ' + b.seed];
	if (b.policy !== 'random-legal') parts.push('--policy ' + b.policy);
	if (b.maxProposals !== undefined) parts.push('--max-proposals ' + b.maxProposals);
	parts.push('--max-versions ' + Math.max(1, version));
	if (b.refusalRate !== undefined && b.refusalRate !== '0.125') parts.push('--refusal-rate ' + b.refusalRate);
	parts.push('--out DIR');
	return parts.join(' ');
}

function lastOutsideInput(doc: ConformanceDocument, version: number): DivergenceReport['lastOutsideInput'] {
	for (let v = Math.min(version, doc.versions.length - 1); v >= 1; v--) {
		const move = doc.versions[v]?.move;
		if (move && (move.trigger === 'propose' || move.trigger === 'seat' || move.trigger === 'timer')) {
			return { atVersion: v - 1, name: move.name, input: move.arguments, trigger: move.trigger };
		}
	}
	return undefined;
}

function bodiesAt(spec: () => GameSpec, doc: ConformanceDocument, version: number, rejection?: Rejection): string[] {
	const seen: string[] = [];
	const names = new Set<string>();
	const game = compileGame(instrumentBodies(spec(), name => {
		if (names.has(name)) return;
		names.add(name);
		seen.push(name);
	}));
	// Only the step whose versions include `version` is kept: each step
	// clears what the one before it ran, and the replay stops after it.
	const run = replay(game, doc, () => undefined, version, () => {
		names.clear();
		seen.length = 0;
	});
	const produced = run.produced[version];
	if (rejection !== undefined) {
		// The refused proposal alone.
		names.clear();
		seen.length = 0;
		if (produced !== undefined && run.record !== null) {
			propose(game, { record: run.record, history: run.history.slice(0, produced.historyLength), state: produced.output.state }, rejectionInput(rejection), { now: run.now });
		}
		return seen;
	}
	// The views at that version run bodies too (computed properties, legality).
	if (produced !== undefined && run.record !== null && doc.versions[version]) {
		for (const view of doc.versions[version].views) {
			viewerOutput(game, {
				record: run.record, history: run.history.slice(0, produced.historyLength), state: produced.output.state,
				previousState: produced.previousState, settled: produced.settled, now: run.now,
			}, view.viewer);
		}
	}
	return seen;
}

// ---------------------------------------------------------------------------
// The document

/** Replays a fuzz document: versions and views by runDocument, then every rejection. */
export function runFuzzDocument(game: CompiledGame, doc: FuzzDocument, options: FuzzOptions = {}): FuzzResult {
	const document = runDocument(game, doc);
	const recorded = doc.rejections ?? [];
	const rejections: RejectionResult = { compared: 0, mismatches: 0, examples: [] };
	let firstRejection: { rejection: Rejection; got: unknown; want: unknown } | undefined;

	const pending = new Map<number, Rejection[]>();
	for (const r of recorded) pending.set(r.atVersion, [...(pending.get(r.atVersion) ?? []), r]);
	const fail = (rejection: Rejection, text: string, got: unknown, want: unknown): void => {
		rejections.mismatches++;
		if (rejections.examples.length < 3) rejections.examples.push('at v' + rejection.atVersion + ' ' + JSON.stringify(rejection.name) + ' by ' + rejection.proposer + ': ' + text);
		if (firstRejection === undefined) firstRejection = { rejection, got, want };
	};

	const run = replay(game, doc, (version, record, history, state, now) => {
		const here = pending.get(version);
		if (here === undefined) return;
		pending.delete(version);
		for (const rejection of here) {
			rejections.compared++;
			const before = cloneJSON({ state, history });
			const step = propose(game, { record, history, state }, rejectionInput(rejection), { now });
			const got = { error: step.error ?? null, versions: step.versions.length };
			const want = { error: rejection.error, versions: 0 };
			const untouched = firstDifference({ state, history }, before);
			if (step.error !== rejection.error) fail(rejection, 'error ' + JSON.stringify(step.error ?? null) + ', want ' + JSON.stringify(rejection.error), got, want);
			else if (step.versions.length !== 0) fail(rejection, 'returned ' + step.versions.length + ' versions, want none', got, want);
			else if (untouched !== null) fail(rejection, 'the engine modified its input: ' + untouched, { modified: untouched }, { modified: null });
		}
	});
	// A rejection at a version the replay never settled at was not compared.
	for (const [, list] of pending) {
		for (const rejection of list) {
			rejections.compared++;
			fail(rejection, 'the replay never settled at version ' + rejection.atVersion, { reached: false }, { reached: true });
		}
	}

	const passed = document.passed && rejections.mismatches === 0;
	const result: FuzzResult = { name: doc.name, document, rejections, recordedRejections: recorded.length, passed };
	if (passed) return result;

	// The first difference in the runner's order.
	let first: Located | null = null;
	for (const field of FIELDS) {
		const example = document.fields[field].examples[0];
		if (example === undefined) continue;
		const located = locate(doc, field, example);
		if (located !== null && (first === null || before(located, first))) first = located;
	}
	const fatalVersion = document.fatal !== undefined ? (run.fatalVersion ?? Number(/version (\d+)/.exec(document.fatal)?.[1] ?? 0)) : undefined;
	if (fatalVersion !== undefined && (first === null || fatalVersion < first.version)) {
		first = { version: fatalVersion, field: 'replay', path: '', text: document.fatal ?? '', order: [fatalVersion, 0, 0, 0] };
	}
	if (firstRejection !== undefined && (first === null || firstRejection.rejection.atVersion < first.version)) {
		first = { version: firstRejection.rejection.atVersion, field: 'rejection', path: '', text: '', order: [firstRejection.rejection.atVersion, 2, 0, 0] };
	}
	if (first === null) {
		// Nothing located (an unparsable example); report the document as a whole.
		first = { version: 0, field: 'replay', path: '', text: document.fatal ?? 'unlocated difference', order: [0, 0, 0, 0] };
	}

	const version = Math.max(first.version, 0);
	const report: DivergenceReport = { document: doc.name, version, field: first.field, path: first.path.replace(/^\./, ''), want: undefined, got: undefined, repro: reproCommand(doc, version) };
	if (first.viewer !== undefined) report.viewer = first.viewer;
	if (first.field === 'rejection' && firstRejection !== undefined) {
		const r = firstRejection.rejection;
		report.want = firstRejection.want;
		report.got = firstRejection.got;
		report.move = { name: r.name, proposer: r.proposer, trigger: r.kind };
		report.lastOutsideInput = { atVersion: r.atVersion, name: r.name, input: r.input, trigger: r.kind };
		report.repro = reproCommand(doc, r.atVersion + 1);
	} else if (first.field === 'replay') {
		report.want = doc.versions[version]?.move ?? null;
		report.got = first.text;
	} else {
		const values = fieldValues(game, doc, run, first);
		const want = values === null ? undefined : valueAt(values.want, first.path);
		const got = values === null ? undefined : valueAt(values.got, first.path);
		if (values !== null && (want !== undefined || got !== undefined)) {
			report.want = want ?? null;
			report.got = got ?? null;
		} else {
			// The header, or a path that cannot be followed: the runner's own text.
			report.want = first.text;
			report.got = first.text;
		}
	}
	if (report.move === undefined && version > 0) {
		const move = doc.versions[version]?.move;
		if (move) report.move = { name: move.name, proposer: move.proposer, trigger: move.trigger };
	}
	if (report.lastOutsideInput === undefined) {
		const last = lastOutsideInput(doc, version);
		if (last !== undefined) report.lastOutsideInput = last;
	}
	if (options.spec !== undefined) report.bodiesExecuted = bodiesAt(options.spec, doc, version, first.field === 'rejection' ? firstRejection?.rejection : undefined);
	result.report = report;
	return result;
}

/** One row per document: versions, viewers, rejections, each field, and the result. */
export function fuzzPassTable(results: readonly FuzzResult[]): string {
	const header = ['document', 'versions', 'viewers', 'rejections', ...FIELDS, 'result'];
	const rows = results.map(result => [
		result.name,
		String(result.document.versions),
		String(result.document.viewers),
		result.rejections.mismatches === 0 ? 'ok ' + result.rejections.compared : result.rejections.mismatches + '/' + result.rejections.compared + ' FAIL',
		...FIELDS.map(field => (result.document.fields[field].mismatches === 0 ? 'ok ' + result.document.fields[field].compared : result.document.fields[field].mismatches + '/' + result.document.fields[field].compared + ' FAIL')),
		result.passed ? 'PASS' : 'FAIL' + (result.document.fatal ? ': ' + result.document.fatal : ''),
	]);
	return [header, ...rows].map(row => '| ' + row.join(' | ') + ' |').join('\n');
}
