// Move progressions: the ordered groups AddOrderedForPhase installs, and the
// match of a phase's move history ("the tape") against them. Ports
// moves/groups.go, group_serial.go, group_repeat.go, group_parallel.go,
// count.go and the per-move matcher in moves/auto_config.go.
//
// A group is plain data naming moves. The tape is the list of move names
// applied since the phase was entered, with the proposed move appended; the
// proposed move is legal at this point when the group consumes all of it.
//
// Not ported: RepeatFromProp (a count read from state).

/** moves.ValidCounter: null when currentCount is acceptable, else why not. */
export type ValidCounter = (currentCount: number, length: number) => string | null;

export const countAny = (): ValidCounter => current =>
	current === 0 ? 'Not enough count have occurred' : current === 1 ? null : 'too many count have occurred';

export const countAll = (): ValidCounter => (current, length) =>
	current < length ? 'Not enough count have occurred' : current === length ? null : 'too many count have occurred';

export const countInfinite = (): ValidCounter => () => null;

export const countAtLeast = (min: number): ValidCounter => current =>
	current >= min ? null : 'currentCount not yet greater than min configuration';

export const countAtMost = (max: number): ValidCounter => (current, length) =>
	current <= (max < 0 ? length : max) ? null : 'currentCount is greater than max configuration';

export const countBetween = (min: number, max: number): ValidCounter => (current, length) => {
	if (current < min) return 'Count below min';
	if (current > (max < 0 ? length : max)) return 'Count above max';
	return null;
};

export const countExactly = (target: number): ValidCounter => current =>
	target === current ? null : target > current ? 'currentCount is not yet targetCount' : 'currentCount has already passed targetCount';

export type ProgressionGroup =
	/** One move type. allowMultiple is interfaces.AllowMultipleInProgression. */
	| { kind: 'move'; name: string; allowMultiple: boolean }
	| { kind: 'serial'; children: readonly ProgressionGroup[] }
	| { kind: 'repeat'; count: ValidCounter; child: ProgressionGroup }
	| { kind: 'parallel'; count: ValidCounter; children: readonly ProgressionGroup[] };

/** moves.Serial. */
export function serialGroup(...children: ProgressionGroup[]): ProgressionGroup {
	return { kind: 'serial', children };
}

/** moves.Repeat. */
export function repeatGroup(count: ValidCounter, child: ProgressionGroup): ProgressionGroup {
	return { kind: 'repeat', count, child };
}

/** moves.Optional: Repeat(CountBetween(0, 1), group). */
export function optionalGroup(child: ProgressionGroup): ProgressionGroup {
	return repeatGroup(countBetween(0, 1), child);
}

/** moves.ParallelCount. */
export function parallelCountGroup(count: ValidCounter, ...children: ProgressionGroup[]): ProgressionGroup {
	return { kind: 'parallel', count, children };
}

/** moves.Parallel. */
export function parallelGroup(...children: ProgressionGroup[]): ProgressionGroup {
	return parallelCountGroup(countAll(), ...children);
}

/** MoveProgressionGroup.MoveConfigs: the move names a group spans, in order. */
export function groupMoveNames(group: ProgressionGroup): string[] {
	switch (group.kind) {
		case 'move':
			return [group.name];
		case 'repeat':
			return groupMoveNames(group.child);
		default: {
			const out: string[] = [];
			for (const child of group.children) out.push(...groupMoveNames(child));
			return out;
		}
	}
}

// The tape is an array and a position; Go's nil tape is position === length.
interface Match {
	rest: number;
	err: string | null;
}

function satisfied(group: ProgressionGroup, tape: readonly string[], head: number): Match {
	const end = tape.length;
	switch (group.kind) {
		case 'move': {
			if (head === end) return { rest: end, err: null };
			if (tape[head] !== group.name) {
				return { rest: head, err: 'Move name does not match: ' + tape[head] + ' is not ' + group.name };
			}
			let rest = head + 1;
			if (group.allowMultiple) while (rest < end && tape[rest] === group.name) rest++;
			return { rest, err: null };
		}
		case 'serial': {
			let tapeHead = head;
			for (const child of group.children) {
				if (tapeHead === end) return { rest: end, err: null };
				const match = satisfied(child, tape, tapeHead);
				if (match.err !== null) return { rest: head, err: match.err };
				tapeHead = match.rest;
			}
			return { rest: tapeHead, err: null };
		}
		case 'repeat': {
			let tapeHead = head;
			let lowerBoundReached = group.count(0, 1) === null;
			let iteration = 1;
			for (;;) {
				if (tapeHead === end) return { rest: end, err: null };
				const match = satisfied(group.child, tape, tapeHead);
				if (match.err !== null) {
					if (lowerBoundReached) return { rest: tapeHead, err: null };
					return { rest: end, err: match.err };
				}
				const boundErr = group.count(iteration, 1);
				if (lowerBoundReached) {
					if (boundErr !== null) break;
				} else if (boundErr === null) {
					lowerBoundReached = true;
				}
				iteration++;
				tapeHead = match.rest;
			}
			return { rest: tapeHead, err: null };
		}
		case 'parallel': {
			let tapeHead = head;
			const matched: boolean[] = group.children.map(() => false);
			let matchedCount = 0;
			for (;;) {
				if (group.count(matchedCount, group.children.length) === null) break;
				if (tapeHead === end) return { rest: end, err: null };
				// Go keeps the longest match (sort.Slice, descending length). A
				// match that consumes the whole tape counts as infinitely long.
				let best = -1;
				let bestRest = 0;
				let bestLength = -1;
				for (let i = 0; i < group.children.length; i++) {
					if (matched[i]) continue;
					const match = satisfied(group.children[i], tape, tapeHead);
					if (match.err !== null) continue;
					const length = match.rest === end ? Infinity : match.rest - tapeHead;
					if (length > bestLength) {
						best = i;
						bestRest = match.rest;
						bestLength = length;
					}
				}
				if (best === -1) return { rest: end, err: 'no more items match, but tape still left and count not yet reached' };
				matched[best] = true;
				matchedCount++;
				tapeHead = bestRest;
			}
			return { rest: tapeHead, err: null };
		}
	}
}

/**
 * moves.matchTape: null when the group consumes the whole tape, otherwise
 * the error text Default.legalMoveInProgression returns.
 */
export function matchTape(group: ProgressionGroup, tape: readonly string[]): string | null {
	const match = satisfied(group, tape, 0);
	if (match.err !== null) return match.err;
	if (match.rest !== tape.length) return 'The progression only matched some of the proposed move history';
	return null;
}
