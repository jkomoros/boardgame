// Range enums (enum.RangeEnum, enum/range.go): an enum whose values are the
// cells of a multi-dimensional grid, each value's string being its indexes
// joined by "_" ("0_0", "0_1", ... for AddRange("spaces", 8, 8)). The core
// and the rules layer store such an enum as an ordinary EnumDef with
// `Range` set to its dimensions; these functions are the Go methods over it.
//
// Every function is a port of the Go method of the same name and keeps its
// odd corners: RangeToValue goes through ValueFromString (so it also accepts
// the decimal key of a one-dimensional enum), ValueToRange of a key the enum
// lacks is nil (null here), and an index outside the grid is IllegalValue.

import type { EnumDef } from '../definition.ts';

/**
 * enum.IllegalValue, math.MaxInt64. A JavaScript number cannot hold it
 * exactly; this literal is the nearest double (2^63), which is also what a
 * Go vector's 9223372036854775807 parses to, so comparisons agree.
 */
export const ILLEGAL_VALUE = 9223372036854775807;

/** enum.rangedValueSeparator. */
const SEPARATOR = '_';

/**
 * enum.Set.AddRange: the EnumDef of a range enum with these dimensions.
 * Throws Go's error text for no dimensions or a dimension below 1.
 */
export function rangeEnumDef(dimensions: readonly number[]): EnumDef {
	if (dimensions.length === 0) throw new Error('No dimensions passed');
	let count = 1;
	for (let i = 0; i < dimensions.length; i++) {
		if (dimensions[i] <= 0) throw new Error('Dimension ' + i + ' is less than or equal to 0, which is illegal');
		count *= dimensions[i];
	}
	const values: Record<string, string> = {};
	const indexes = dimensions.map(() => 0);
	for (let key = 0; key < count; key++) {
		values[String(key)] = keyForIndexes(indexes);
		for (let j = indexes.length - 1; j >= 0; j--) {
			indexes[j]++;
			if (indexes[j] < dimensions[j]) break;
			indexes[j] = 0;
		}
	}
	return { Values: values, DefaultValue: 0, Range: dimensions.slice() };
}

/** enum.keyForIndexes. */
function keyForIndexes(indexes: readonly number[]): string {
	return indexes.map(index => String(index)).join(SEPARATOR);
}

/** strconv.Atoi: the int, or null where Go returns an error. */
function atoi(text: string): number | null {
	if (!/^[+-]?[0-9]+$/.test(text)) return null;
	const value = Number(text);
	// Go's int is 64 bits; past that Atoi reports a range error.
	if (Math.abs(value) > ILLEGAL_VALUE) return null;
	return value === 0 ? 0 : value;
}

/** enum.indexesForKey: null where any piece is not an integer. */
function indexesForKey(text: string): number[] | null {
	const pieces = text.split(SEPARATOR);
	const result: number[] = [];
	for (const piece of pieces) {
		const value = atoi(piece);
		if (value === null) return null;
		result.push(value);
	}
	return result;
}

/** strings.TrimSpace(strings.ToLower(s)), enum.normalizeStringKey. */
function normalize(text: string): string {
	return text.toLowerCase().trim();
}

/** Whether the enum has this key (Enum.Valid). */
export function enumValid(def: EnumDef, key: number): boolean {
	return Object.prototype.hasOwnProperty.call(def.Values, String(key));
}

/** Enum.String(key): '' for a key the enum lacks. */
export function enumString(def: EnumDef, key: number): string {
	return enumValid(def, key) ? def.Values[String(key)] : '';
}

/**
 * Enum.ValueFromString: the key whose string matches after trimming and
 * lower-casing, else the key whose decimal form `text` is, else
 * IllegalValue. Go ranges over a map for the first rule, so two strings
 * that normalize alike have no fixed answer there; the lowest key is taken.
 */
export function enumValueFromString(def: EnumDef, text: string): number {
	const wanted = normalize(text);
	const keys = Object.keys(def.Values)
		.map(Number)
		.sort((a, b) => a - b);
	for (const key of keys) if (normalize(def.Values[String(key)]) === wanted) return key;
	const number = atoi(text);
	if (number !== null && enumValid(def, number)) return number;
	return ILLEGAL_VALUE;
}

/** RangeEnum.RangeDimensions(); null for an enum that is not a range enum. */
export function rangeDimensions(def: EnumDef): number[] | null {
	return def.Range ? def.Range.slice() : null;
}

/**
 * RangeEnum.ValueToRange(key): the indexes of the cell, one per dimension.
 * Null (Go's nil) for a key the enum lacks.
 */
export function valueToRange(def: EnumDef, key: number): number[] | null {
	return indexesForKey(enumString(def, key));
}

/** RangeEnum.RangeToValue(indexes...): the key of that cell, or ILLEGAL_VALUE. */
export function rangeToValue(def: EnumDef, indexes: readonly number[]): number {
	return enumValueFromString(def, keyForIndexes(indexes));
}

/**
 * RangeVal.SetRangeValue(indexes...)'s outcome: Go's error text, or null.
 * The key it sets on success is rangeToValue(def, indexes).
 */
export function setRangeValueError(def: EnumDef, indexes: readonly number[]): string | null {
	return enumValid(def, rangeToValue(def, indexes)) ? null : 'That value is not valid for this enum';
}

/** ImmutableRangeVal.RangeValue() of a value set to `key`. */
export function rangeValue(def: EnumDef, key: number): number[] | null {
	return valueToRange(def, key);
}
