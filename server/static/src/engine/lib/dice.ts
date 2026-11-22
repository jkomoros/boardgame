// components/dice: die components with any faces, and the roll that draws
// from the state's generator. A die is a component whose static values are
// `{ Faces }` and whose dynamic values are `{ Value, SelectedFace,
// RollCount }` (dice.Value and dice.DynamicValue).

import type { PropDef } from '../definition.ts';
import type { GoRand } from '../rand.ts';
import type { RulesState } from '../rules/state.ts';

/** dice.Value: a die's static values. */
export interface DieValues {
	Faces: number[];
}

/** The dynamic value properties dice.DynamicValue declares, to spread into a deck's dynamicComponentValues props. */
export const dieDynamicProps: Readonly<Record<string, PropDef>> = {
	Value: { Type: 'int' },
	SelectedFace: { Type: 'int' },
	RollCount: { Type: 'int' },
};

/** dice.BasicDie(min, max): a face for each of min through max; null (Go's nil) unless min < max. */
export function basicDie(min: number, max: number): DieValues | null {
	if (min >= max) return null;
	const faces: number[] = [];
	for (let i = min; i <= max; i++) faces.push(i);
	return { Faces: faces };
}

/** dice.DefaultDie(): a six-sided die. */
export function defaultDie(): DieValues {
	return { Faces: [1, 2, 3, 4, 5, 6] };
}

/** math.MaxInt64 and math.MinInt64, as the nearest doubles. */
const MAX_INT64 = 9223372036854775807;
const MIN_INT64 = -9223372036854775808;

/** (*dice.Value).Min(): the lowest face; math.MaxInt64 for a die with none. */
export function dieMin(values: { readonly Faces: readonly number[] }): number {
	let min = MAX_INT64;
	for (const face of values.Faces) if (face < min) min = face;
	return min;
}

/** (*dice.Value).Max(): the highest face; math.MinInt64 for a die with none. */
export function dieMax(values: { readonly Faces: readonly number[] }): number {
	let max = MIN_INT64;
	for (const face of values.Faces) if (face > max) max = face;
	return max;
}

/**
 * (*dice.DynamicValue).Roll(r): one draw of r.Intn(len(Faces)), which
 * becomes SelectedFace; Value becomes that face; RollCount goes up by one
 * even when the face is the one already showing.
 */
export function rollWith(rand: GoRand, values: { readonly Faces: readonly number[] }, dynamic: Record<string, unknown>): void {
	const face = rand.intn(values.Faces.length);
	dynamic.SelectedFace = face;
	dynamic.Value = values.Faces[face];
	dynamic.RollCount = (dynamic.RollCount as number) + 1;
}

/**
 * `c.DynamicValues().(*dice.DynamicValue).Roll(state.Rand())` for the die
 * at `index` of `deck`: draws from the state's generator, as Go does.
 */
export function roll(state: RulesState, deck: string, index: number): void {
	const values = state.component(deck, index) as { readonly Faces?: readonly number[] };
	// Go: a die whose static values are not dice.Value does not roll.
	if (!Array.isArray(values.Faces)) return;
	rollWith(state.rand(), { Faces: values.Faces }, state.dynamic(deck, index));
}
