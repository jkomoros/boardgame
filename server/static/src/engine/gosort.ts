// An exact port of Go's sort.Sort: the pattern-defeating quicksort in
// sort/zsortinterface.go.
//
// sort.Sort is not stable, but it is deterministic: for a given input and
// comparison, which of two equal elements ends up first is fixed by this
// algorithm (including its "random" pattern breaking, which is an xorshift
// seeded by the range length). Stack.SortComponents must therefore run the
// same algorithm, or a stack sorted by a key with ties would come out in a
// different order than on the Go server.

/** sort.Sort over indexes [0, n), in terms of less(i, j) and swap(i, j). */
export function goSort(n: number, less: (i: number, j: number) => boolean, swap: (i: number, j: number) => void): void {
	if (n <= 1) return;
	const UNKNOWN = 0;
	const INCREASING = 1;
	const DECREASING = 2;

	/** bits.Len for a non-negative int below 2^32. */
	const bitsLen = (x: number): number => (x === 0 ? 0 : 32 - Math.clz32(x));

	function insertionSort(a: number, b: number): void {
		for (let i = a + 1; i < b; i++) {
			for (let j = i; j > a && less(j, j - 1); j--) swap(j, j - 1);
		}
	}

	function siftDown(lo: number, hi: number, first: number): void {
		let root = lo;
		for (;;) {
			let child = 2 * root + 1;
			if (child >= hi) break;
			if (child + 1 < hi && less(first + child, first + child + 1)) child++;
			if (!less(first + root, first + child)) return;
			swap(first + root, first + child);
			root = child;
		}
	}

	function heapSort(a: number, b: number): void {
		const first = a;
		const lo = 0;
		const hi = b - a;
		for (let i = Math.trunc((hi - 1) / 2); i >= 0; i--) siftDown(i, hi, first);
		for (let i = hi - 1; i >= 0; i--) {
			swap(first, first + i);
			siftDown(lo, i, first);
		}
	}

	function partition(a: number, b: number, pivot: number): { mid: number; alreadyPartitioned: boolean } {
		swap(a, pivot);
		let i = a + 1;
		let j = b - 1;
		while (i <= j && less(i, a)) i++;
		while (i <= j && !less(j, a)) j--;
		if (i > j) {
			swap(j, a);
			return { mid: j, alreadyPartitioned: true };
		}
		swap(i, j);
		i++;
		j--;
		for (;;) {
			while (i <= j && less(i, a)) i++;
			while (i <= j && !less(j, a)) j--;
			if (i > j) break;
			swap(i, j);
			i++;
			j--;
		}
		swap(j, a);
		return { mid: j, alreadyPartitioned: false };
	}

	function partitionEqual(a: number, b: number, pivot: number): number {
		swap(a, pivot);
		let i = a + 1;
		let j = b - 1;
		for (;;) {
			while (i <= j && !less(a, i)) i++;
			while (i <= j && less(a, j)) j--;
			if (i > j) break;
			swap(i, j);
			i++;
			j--;
		}
		return i;
	}

	function partialInsertionSort(a: number, b: number): boolean {
		const maxSteps = 5;
		const shortestShifting = 50;
		let i = a + 1;
		for (let step = 0; step < maxSteps; step++) {
			while (i < b && !less(i, i - 1)) i++;
			if (i === b) return true;
			if (b - a < shortestShifting) return false;
			swap(i, i - 1);
			if (i - a >= 2) {
				for (let j = i - 1; j >= 1; j--) {
					if (!less(j, j - 1)) break;
					swap(j, j - 1);
				}
			}
			if (b - i >= 2) {
				for (let j = i + 1; j < b; j++) {
					if (!less(j, j - 1)) break;
					swap(j, j - 1);
				}
			}
		}
		return false;
	}

	function breakPatterns(a: number, b: number): void {
		const length = b - a;
		if (length < 8) return;
		// xorshift64 seeded with the length; only the low 32 bits of each
		// output are used (the modulus is a power of two no larger than 2^32).
		let hi = 0;
		let lo = length >>> 0;
		const next = (): number => {
			// r ^= r << 13
			hi = (hi ^ ((hi << 13) | (lo >>> 19))) >>> 0;
			lo = (lo ^ (lo << 13)) >>> 0;
			// r ^= r >> 7
			lo = (lo ^ ((lo >>> 7) | (hi << 25))) >>> 0;
			hi = (hi ^ (hi >>> 7)) >>> 0;
			// r ^= r << 17
			hi = (hi ^ ((hi << 17) | (lo >>> 15))) >>> 0;
			lo = (lo ^ (lo << 17)) >>> 0;
			return lo;
		};
		const shift = bitsLen(length);
		const mask = shift >= 32 ? 0xffffffff : ((1 << shift) >>> 0) - 1;
		const quarter = Math.trunc(length / 4);
		for (let idx = a + quarter * 2 - 1; idx <= a + quarter * 2 + 1; idx++) {
			let other = (next() & mask) >>> 0;
			if (other >= length) other -= length;
			swap(idx, a + other);
		}
	}

	function choosePivot(a: number, b: number): { pivot: number; hint: number } {
		const shortestNinther = 50;
		const maxSwaps = 4 * 3;
		const l = b - a;
		let swaps = 0;
		const quarter = Math.trunc(l / 4);
		let i = a + quarter * 1;
		let j = a + quarter * 2;
		let k = a + quarter * 3;

		const order2 = (x: number, y: number): [number, number] => {
			if (less(y, x)) {
				swaps++;
				return [y, x];
			}
			return [x, y];
		};
		const median = (x: number, y: number, z: number): number => {
			[x, y] = order2(x, y);
			[y, z] = order2(y, z);
			[x, y] = order2(x, y);
			return y;
		};
		const medianAdjacent = (x: number): number => median(x - 1, x, x + 1);

		if (l >= 8) {
			if (l >= shortestNinther) {
				i = medianAdjacent(i);
				j = medianAdjacent(j);
				k = medianAdjacent(k);
			}
			j = median(i, j, k);
		}
		if (swaps === 0) return { pivot: j, hint: INCREASING };
		if (swaps === maxSwaps) return { pivot: j, hint: DECREASING };
		return { pivot: j, hint: UNKNOWN };
	}

	function reverseRange(a: number, b: number): void {
		let i = a;
		let j = b - 1;
		while (i < j) {
			swap(i, j);
			i++;
			j--;
		}
	}

	function pdqsort(a: number, b: number, limit: number): void {
		const maxInsertion = 12;
		let wasBalanced = true;
		let wasPartitioned = true;
		for (;;) {
			const length = b - a;
			if (length <= maxInsertion) {
				insertionSort(a, b);
				return;
			}
			if (limit === 0) {
				heapSort(a, b);
				return;
			}
			if (!wasBalanced) {
				breakPatterns(a, b);
				limit--;
			}
			const chosen = choosePivot(a, b);
			let pivot = chosen.pivot;
			let hint = chosen.hint;
			if (hint === DECREASING) {
				reverseRange(a, b);
				pivot = b - 1 - (pivot - a);
				hint = INCREASING;
			}
			if (wasBalanced && wasPartitioned && hint === INCREASING) {
				if (partialInsertionSort(a, b)) return;
			}
			if (a > 0 && !less(a - 1, pivot)) {
				a = partitionEqual(a, b, pivot);
				continue;
			}
			const part = partition(a, b, pivot);
			const mid = part.mid;
			wasPartitioned = part.alreadyPartitioned;
			const leftLen = mid - a;
			const rightLen = b - mid;
			const balanceThreshold = Math.trunc(length / 8);
			if (leftLen < rightLen) {
				wasBalanced = leftLen >= balanceThreshold;
				pdqsort(a, mid, limit);
				a = mid + 1;
			} else {
				wasBalanced = rightLen >= balanceThreshold;
				pdqsort(mid + 1, b, limit);
				b = mid;
			}
		}
	}

	pdqsort(0, n, bitsLen(n));
}
