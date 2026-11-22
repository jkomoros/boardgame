// Graphs over enum values (enum/graph): the grid connectedness a board game
// builds once at package level (graph.MustNewEnumGridConnectedness) and the
// adjacency queries on it. A graph here is data: each node's neighbors,
// sorted, so it can be built when asked for and never needs to be kept.
//
// Go stores a node's edges in a map and Neighbors ranges over it, so Go's
// order of neighbors is unspecified and changes from run to run. The port
// returns them ascending; any Go caller whose result depends on the order
// has no fixed answer to reproduce. Go returns nil for a node with no edges;
// the port returns [].

import type { EnumDef } from '../definition.ts';
import { ILLEGAL_VALUE, enumValid, rangeDimensions, rangeToValue, valueToRange } from './enumrange.ts';

/** A finished directed graph whose nodes are an enum's keys. */
export interface EnumGraph {
	/** Node key (decimal) to its neighbors, ascending. A node with no edges is absent. */
	readonly edges: Readonly<Record<string, readonly number[]>>;
}

/** The Direction* edge filters of enum/graph/grid.go. */
export type Direction = 'up' | 'down' | 'left' | 'right' | 'perpendicular' | 'diagonal';

/**
 * A direction by any of the names a caller may carry: the Go identifier
 * ("graph.DirectionDown", "DirectionDown") or the short form ("down").
 * Throws for anything else.
 */
export function parseDirection(name: string): Direction {
	let short = name.startsWith('graph.') ? name.slice('graph.'.length) : name;
	if (short.startsWith('Direction')) short = short.slice('Direction'.length);
	const lower = short.toLowerCase();
	switch (lower) {
		case 'up':
		case 'down':
		case 'left':
		case 'right':
		case 'perpendicular':
		case 'diagonal':
			return lower;
	}
	throw new Error('graph: unknown direction ' + JSON.stringify(name));
}

function indexesOf(def: EnumDef, key: number): number[] {
	const indexes = valueToRange(def, key);
	// Go indexes into a nil slice here and panics; a grid's own keys never do.
	if (indexes === null) throw new Error('runtime error: index out of range [0] with length 0');
	return indexes;
}

/** The EdgeFilter named by `direction`, applied to one candidate edge. */
function passes(def: EnumDef, direction: Direction, from: number, to: number): boolean {
	const a = indexesOf(def, from);
	const b = indexesOf(def, to);
	switch (direction) {
		case 'up':
			return a[0] > b[0];
		case 'down':
			return a[0] < b[0];
		case 'left':
			return a[1] > b[1];
		case 'right':
			return a[1] < b[1];
		case 'perpendicular':
			return a[0] === b[0] || a[1] === b[1];
		case 'diagonal':
			return !(a[0] === b[0] || a[1] === b[1]);
	}
}

/**
 * graph.NewGridConnectedness(e, filters...): every cell connected to each of
 * its up to 8 neighbors that passes every filter, in order. `def` must be a
 * two-dimensional range enum; otherwise Go's error is thrown.
 */
export function gridConnectedness(def: EnumDef, directions: readonly (Direction | string)[] = []): EnumGraph {
	const dims = rangeDimensions(def);
	if (dims === null || dims.length !== 2) throw new Error('The enum did not have two dimensions');
	const filters = directions.map(parseDirection);
	const edges: Record<string, number[]> = {};
	const keys = Object.keys(def.Values)
		.map(Number)
		.sort((a, b) => a - b);
	for (const key of keys) {
		const start = indexesOf(def, key);
		let candidates: number[] = [];
		for (let r = -1; r < 2; r++) {
			for (let c = -1; c < 2; c++) {
				if (r === 0 && c === 0) continue;
				const value = rangeToValue(def, [start[0] + r, start[1] + c]);
				if (value !== ILLEGAL_VALUE) candidates.push(value);
			}
		}
		for (const filter of filters) candidates = candidates.filter(to => passes(def, filter, key, to));
		if (candidates.length > 0) edges[String(key)] = [...new Set(candidates)].sort((a, b) => a - b);
	}
	return { edges };
}

/** EnumGraph.NeighborsByKey(start): ascending; [] where Go returns nil. */
export function neighborsByKey(graph: EnumGraph, start: number): number[] {
	const found = graph.edges[String(start)];
	return found ? found.slice() : [];
}

/** EnumGraph.ConnectedByKey(from, to). */
export function connectedByKey(graph: EnumGraph, from: number, to: number): boolean {
	const found = graph.edges[String(from)];
	return found !== undefined && found.includes(to);
}

/** Whether every key of a graph's edges is a value of `def` (a sanity check for loaded data). */
export function graphMatchesEnum(graph: EnumGraph, def: EnumDef): boolean {
	for (const from of Object.keys(graph.edges)) {
		if (!enumValid(def, Number(from))) return false;
		for (const to of graph.edges[from]) if (!enumValid(def, to)) return false;
	}
	return true;
}
