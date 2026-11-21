// sort.Sort's exact output order, ties included. testdata/sort.vectors.json
// comes from real Go (testdata/gen/sortvec).
import { test } from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { goSort } from './gosort.ts';

const { Vectors } = JSON.parse(readFileSync(new URL('./testdata/sort.vectors.json', import.meta.url), 'utf8')) as {
	Vectors: { Label: string; Keys: number[]; Order: number[] }[];
};

function sortTags(keys: number[]): number[] {
	const k = keys.slice();
	const tags = keys.map((_, i) => i);
	goSort(
		k.length,
		(i, j) => k[i] < k[j],
		(i, j) => {
			[k[i], k[j]] = [k[j], k[i]];
			[tags[i], tags[j]] = [tags[j], tags[i]];
		},
	);
	for (let i = 1; i < k.length; i++) assert.ok(k[i - 1] <= k[i], 'output is sorted');
	return tags;
}

test('goSort places equal keys exactly where sort.Sort places them', () => {
	assert.ok(Vectors.length > 250, String(Vectors.length));
	let unstable = 0;
	for (const v of Vectors) {
		const tags = sortTags(v.Keys);
		assert.deepEqual(tags, v.Order ?? [], `${v.Label} n=${v.Keys.length}`);
		// Count the inputs where Go's order differs from a stable sort, to
		// show the vectors really do pin down the unstable behaviour.
		const stable = v.Keys.map((_, i) => i).sort((a, b) => v.Keys[a] - v.Keys[b] || a - b);
		if (JSON.stringify(stable) !== JSON.stringify(tags)) unstable++;
	}
	assert.ok(unstable > 60, `${unstable} vectors differ from a stable sort`);
});

test('goSort is stable up to 12 elements, where sort.Sort is an insertion sort', () => {
	for (const v of Vectors) {
		if (v.Keys.length > 12) continue;
		const stable = v.Keys.map((_, i) => i).sort((a, b) => v.Keys[a] - v.Keys[b] || a - b);
		assert.deepEqual(sortTags(v.Keys), stable, `${v.Label} n=${v.Keys.length}`);
	}
});
