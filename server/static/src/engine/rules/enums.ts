// Enums as the rules layer uses them: the core's EnumDef (decimal key to
// string value) bound to the lookups Go's enum.Enum offers. A tree enum
// (enum.TreeEnum) is an EnumDef with Parents; of the tree API only
// Ancestors and IsLeaf are provided.

import type { EnumDef } from '../definition.ts';
import type { LegalEnum, LegalEnumValue } from '../legal/access.ts';

export interface BoundEnum extends LegalEnum {
	readonly def: EnumDef;
	/** Enum.Values(): the keys, ascending. */
	keys(): number[];
	/** Enum.String(key); '' for a key the enum does not have. */
	stringOf(key: number): string;
	/** The key whose string value is exactly `value`, or null. */
	keyOf(value: unknown): number | null;
	/** Enum.NewVal(key) as the evaluator reads one. */
	value(key: number): LegalEnumValue;
}

/** strings.TrimSpace + strings.ToLower, as Enum.ValueFromString normalizes. */
function normalize(text: string): string {
	return text.toLowerCase().trim();
}

export function bindEnum(name: string, def: EnumDef): BoundEnum {
	const keys = Object.keys(def.Values)
		.map(Number)
		.sort((a, b) => a - b);
	const bound: BoundEnum = {
		name,
		def,
		keys: () => keys.slice(),
		stringOf: key => (Object.prototype.hasOwnProperty.call(def.Values, String(key)) ? def.Values[String(key)] : ''),
		keyOf(value) {
			for (const key of keys) if (def.Values[String(key)] === value) return key;
			return null;
		},
		// Go ranges over a map here, so two values that normalize alike make
		// its answer depend on iteration order. The lowest key is taken.
		valueFromString(s) {
			const wanted = normalize(s);
			for (const key of keys) if (normalize(def.Values[String(key)]) === wanted) return key;
			if (/^[+-]?[0-9]+$/.test(s) && Object.prototype.hasOwnProperty.call(def.Values, String(Number(s)))) return Number(s);
			return null;
		},
		// TreeEnum.Ancestors: the key, then each parent up to the root.
		ancestors(key) {
			const parents = def.Parents;
			if (!parents) return null;
			const chain = [key];
			let current = key;
			while (Object.prototype.hasOwnProperty.call(parents, String(current)) && parents[String(current)] !== current && chain.length <= keys.length) {
				current = parents[String(current)];
				chain.push(current);
			}
			return chain;
		},
		value: key => ({ enum: bound, value: key, string: bound.stringOf(key) }),
	};
	return bound;
}

/** Every enum a game names, chest enums first. Identity is the name. */
export interface EnumRegistry {
	readonly byName: Record<string, BoundEnum>;
	readonly chestNames: string[];
	get(name: string | undefined): BoundEnum | null;
	/** ComponentChest.Enums(): only the enums the chest ships. */
	chestEnums(): BoundEnum[];
}

/** Creates an `EnumRegistry`. */
export function newEnumRegistry(chest: Readonly<Record<string, EnumDef>>, other: Readonly<Record<string, EnumDef>> = {}): EnumRegistry {
	const self: EnumRegistry = {
		byName: {},
		chestNames: [],
		get(name: string | undefined): BoundEnum | null {
			if (name === undefined) return null;
			return this.byName[name] ?? null;
		},
		/** ComponentChest.Enums(): only the enums the chest ships. */
		chestEnums(): BoundEnum[] {
			return this.chestNames.map(name => this.byName[name]);
		},
	};
	for (const name of Object.keys(chest).sort()) {
			self.byName[name] = bindEnum(name, chest[name]);
			self.chestNames.push(name);
		}
	for (const name of Object.keys(other)) {
			if (!self.byName[name]) self.byName[name] = bindEnum(name, other[name]);
		}
	return self;
}
