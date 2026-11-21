// TEST SUPPORT ONLY. Uses Node's fs; never imported by engine code.
//
// A port of internal/patchtree: the repository's JSON fixtures are stored as
// a base.json plus a chain of jd-format patches, one modification.patch per
// directory. patchTreeJSON composes them, as patchtree.JSON does for the Go
// tests.
import { existsSync, readFileSync } from 'node:fs';
import { dirname, join } from 'node:path';

type JSONValue = null | boolean | number | string | JSONValue[] | { [key: string]: JSONValue };

interface Hunk {
	path: (string | number)[];
	remove: JSONValue[];
	add: JSONValue[];
}

/** Parses jd's native diff format: `@ path`, then `- old` and `+ new` lines. */
export function parseJdPatch(text: string): Hunk[] {
	const hunks: Hunk[] = [];
	for (const line of text.split('\n')) {
		if (line.trim() === '') continue;
		const marker = line.charAt(0);
		const body = line.slice(2);
		if (marker === '@') {
			hunks.push({ path: JSON.parse(body), remove: [], add: [] });
			continue;
		}
		const hunk = hunks[hunks.length - 1];
		if (!hunk) throw new Error('jd patch: value line before any path: ' + line);
		if (marker === '-') hunk.remove.push(JSON.parse(body));
		else if (marker === '+') hunk.add.push(JSON.parse(body));
		else throw new Error('jd patch: unrecognized line: ' + line);
	}
	return hunks;
}

function sameJSON(a: JSONValue, b: JSONValue): boolean {
	if (a === b) return true;
	if (a === null || b === null || typeof a !== 'object' || typeof b !== 'object') return false;
	if (Array.isArray(a) !== Array.isArray(b)) return false;
	if (Array.isArray(a) && Array.isArray(b)) {
		return a.length === b.length && a.every((v, i) => sameJSON(v, b[i]));
	}
	const ao = a as { [key: string]: JSONValue };
	const bo = b as { [key: string]: JSONValue };
	const keys = Object.keys(ao);
	return keys.length === Object.keys(bo).length && keys.every(k => k in bo && sameJSON(ao[k], bo[k]));
}

/** Applies a parsed jd patch in order, checking every removed value. */
export function applyJdPatch(root: JSONValue, hunks: Hunk[]): JSONValue {
	let doc = root;
	for (const hunk of hunks) {
		const where = JSON.stringify(hunk.path);
		if (hunk.path.length === 0) {
			if (hunk.remove.length && !sameJSON(doc, hunk.remove[0])) throw new Error('jd patch: root mismatch');
			doc = hunk.add[0];
			continue;
		}
		let parent = doc as JSONValue;
		for (const step of hunk.path.slice(0, -1)) {
			parent = (parent as Record<string, JSONValue>)[step as string];
			if (parent === undefined) throw new Error('jd patch: no such path ' + where);
		}
		const last = hunk.path[hunk.path.length - 1];
		if (Array.isArray(parent)) {
			const index = last as number;
			for (const expected of hunk.remove) {
				if (!sameJSON(parent[index], expected)) {
					throw new Error(`jd patch: at ${where} expected to remove ${JSON.stringify(expected)}, found ${JSON.stringify(parent[index])}`);
				}
				parent.splice(index, 1);
			}
			parent.splice(index, 0, ...hunk.add);
		} else {
			const obj = parent as { [key: string]: JSONValue };
			const key = last as string;
			if (hunk.remove.length) {
				if (!sameJSON(obj[key], hunk.remove[0])) {
					throw new Error(`jd patch: at ${where} expected to remove ${JSON.stringify(hunk.remove[0])}, found ${JSON.stringify(obj[key])}`);
				}
				delete obj[key];
			}
			if (hunk.add.length) obj[key] = hunk.add[0];
		}
	}
	return doc;
}

/**
 * patchtree.JSON: walks up from `directory` to the directory holding
 * base.json, then applies each modification.patch on the way back down.
 */
export function patchTreeJSON(directory: string): JSONValue {
	const chain: string[] = [];
	let current = directory;
	while (!existsSync(join(current, 'base.json'))) {
		const patch = join(current, 'modification.patch');
		if (!existsSync(patch)) throw new Error('patchtree: no modification.patch in ' + current);
		chain.unshift(patch);
		const parent = dirname(current);
		if (parent === current) throw new Error('patchtree: no base.json above ' + directory);
		current = parent;
	}
	let doc: JSONValue = JSON.parse(readFileSync(join(current, 'base.json'), 'utf8'));
	for (const patch of chain) doc = applyJdPatch(doc, parseJdPatch(readFileSync(patch, 'utf8')));
	return doc;
}
