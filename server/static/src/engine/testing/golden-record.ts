// TEST SUPPORT ONLY. Uses Node's fs; never imported by engine code.
//
// Reads a golden history as storage/filesystem/record does. A golden file
// holds the game record, the moves, and one entry per version in
// StatePatches. Entries are either full states, or deltas in the
// jsondiffpatch format that github.com/yudai/gojsondiff writes (the first one
// a delta against the empty object).
import { readFileSync } from 'node:fs';
import type { StateJSON } from '../state.ts';

type JSONValue = null | boolean | number | string | JSONValue[] | { [key: string]: JSONValue };
type JSONObject = { [key: string]: JSONValue };

export interface GoldenGame {
	Name: string;
	ID: string;
	SecretSalt: string;
	Version: number;
	NumPlayers: number;
	Finished: boolean;
	Winners: number[] | null;
	Variant: Record<string, string> | null;
}

export interface GoldenMove {
	Name: string;
	Version: number;
	Initiator: number;
	Phase: number;
	Proposer: number;
	Blob: JSONObject;
}

export interface GoldenHistory {
	Game: GoldenGame;
	Moves: GoldenMove[];
	/** States[v] is the full, unsanitized state at version v. */
	States: StateJSON[];
}

function clone<T extends JSONValue>(value: T): T {
	return JSON.parse(JSON.stringify(value)) as T;
}

/**
 * Applies a diff-match-patch text patch (the `[patch, 0, 2]` delta
 * jsondiffpatch uses for long strings). Golden patches are exact, so each
 * hunk is applied at its recorded position and its context is verified.
 */
export function applyTextPatch(text: string, patch: string): string {
	const lines = patch.split('\n');
	let i = 0;
	while (i < lines.length) {
		if (lines[i] === '') {
			i++;
			continue;
		}
		const header = /^@@ -(\d+),?(\d*) \+(\d+),?(\d*) @@$/.exec(lines[i]);
		if (!header) throw new Error('golden text patch: bad header ' + lines[i]);
		// diff-match-patch prints a zero-length range with its 0-based start
		// and any other range 1-based.
		let start = Number(header[3]);
		if (header[4] !== '0') start -= 1;
		i++;
		let before = '';
		let after = '';
		for (; i < lines.length && lines[i].charAt(0) !== '@'; i++) {
			if (lines[i] === '') continue;
			const sign = lines[i].charAt(0);
			const body = decodeURI(lines[i].slice(1));
			if (sign === ' ') {
				before += body;
				after += body;
			} else if (sign === '-') before += body;
			else if (sign === '+') after += body;
			else throw new Error('golden text patch: bad line ' + lines[i]);
		}
		if (text.slice(start, start + before.length) !== before) {
			throw new Error('golden text patch: context mismatch at ' + start);
		}
		text = text.slice(0, start) + after + text.slice(start + before.length);
	}
	return text;
}

/** Applies one jsondiffpatch delta, returning the patched value. */
export function applyDelta(value: JSONValue | undefined, delta: JSONValue): JSONValue {
	if (Array.isArray(delta)) {
		if (delta.length === 1) return clone(delta[0]); // added
		if (delta.length === 2) return clone(delta[1]); // replaced
		if (delta.length === 3 && delta[2] === 2) return applyTextPatch(value as string, delta[0] as string);
		throw new Error('golden delta: unsupported array delta ' + JSON.stringify(delta).slice(0, 80));
	}
	const d = delta as JSONObject;
	if (d._t === 'a') {
		const array = value as JSONValue[];
		const removals: number[] = [];
		const insertions: { index: number; value: JSONValue }[] = [];
		const modifications: { index: number; delta: JSONValue }[] = [];
		for (const key of Object.keys(d)) {
			if (key === '_t') continue;
			const entry = d[key];
			if (key.charAt(0) === '_') {
				const e = entry as JSONValue[];
				if (e.length !== 3 || e[2] !== 0) throw new Error('golden delta: array moves are not supported');
				removals.push(Number(key.slice(1)));
			} else if (Array.isArray(entry) && entry.length === 1) {
				insertions.push({ index: Number(key), value: clone(entry[0]) });
			} else {
				modifications.push({ index: Number(key), delta: entry });
			}
		}
		removals.sort((a, b) => b - a);
		for (const index of removals) array.splice(index, 1);
		insertions.sort((a, b) => a.index - b.index);
		for (const ins of insertions) array.splice(ins.index, 0, ins.value);
		for (const mod of modifications) array[mod.index] = applyDelta(array[mod.index], mod.delta);
		return array;
	}
	const obj = (value ?? {}) as JSONObject;
	for (const key of Object.keys(d)) {
		const entry = d[key];
		if (Array.isArray(entry) && entry.length === 3 && entry[2] === 0) delete obj[key];
		else obj[key] = applyDelta(obj[key], entry);
	}
	return obj;
}

/** Reads a golden file and expands every version's state. */
export function readGoldenHistory(path: string): GoldenHistory {
	const raw = JSON.parse(readFileSync(path, 'utf8')) as { Game: GoldenGame; Moves: GoldenMove[]; StatePatches: JSONObject[] };
	const states: StateJSON[] = [];
	// record.diffEncoderImpl.Matches: a delta against {} has Game as a list.
	const deltas = raw.StatePatches.length > 0 && Array.isArray(raw.StatePatches[0].Game);
	let current: JSONValue = {};
	for (const patch of raw.StatePatches) {
		current = deltas ? applyDelta(clone(current), patch) : clone(patch);
		states.push(clone(current) as unknown as StateJSON);
	}
	return { Game: raw.Game, Moves: raw.Moves, States: states };
}
