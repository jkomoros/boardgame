// The data a game hands the engine to describe its state: decks, enums, and
// the property schema of each sub-state with its sanitization policy.
//
// This is what Go discovers by reflection over struct tags. Here it is plain
// JSON. testdata/gen/defs derives one from any Go game's live manager, in
// exactly this shape, so a ported game can start from an exported definition.
// Keys are capitalized to match that export and the state JSON beside it.

/** A sanitization policy, least restrictive first (sanitization.go Policy). */
export type PolicyName = 'visible' | 'order' | 'len' | 'nonempty' | 'hidden' | 'invalid';

/** Policies in Go's iota order: the numeric value is the index. */
export const POLICY_ORDER: readonly PolicyName[] = ['visible', 'order', 'len', 'nonempty', 'hidden', 'invalid'];

export type PropType =
	| 'int'
	| 'bool'
	| 'string'
	| 'playerIndex'
	| 'enum'
	| 'intSlice'
	| 'boolSlice'
	| 'stringSlice'
	| 'playerIndexSlice'
	| 'enumSlice'
	| 'timer'
	/** A growable stack; MaxSize 0 means uncapped. */
	| 'stack'
	/** A fixed number of slots, each empty or holding one component. */
	| 'sizedStack'
	/** A read-only view over sibling stack properties. */
	| 'mergedStack'
	/** An array of growable stacks ("spaces"). */
	| 'board';

export interface PropDef {
	Type: PropType;
	/**
	 * Group name to policy, or the Go struct-tag string ("other:len" or just
	 * "hidden"). Absent means visible to the sub-state's default group.
	 */
	Sanitize?: Readonly<Record<string, PolicyName>> | string;
	/** For enum and enumSlice: the enum's name in GameDefinition.Enums. */
	Enum?: string;
	/** For stacks and boards: the deck the stack draws from. */
	Deck?: string;
	/** Slot count of a sizedStack, or space count of a board. */
	Size?: number;
	/** Cap of a growable stack (or of each board space). 0 or absent: none. */
	MaxSize?: number;
	/** mergedStack: true for an overlap view, false for a concatenation. */
	Overlap?: boolean;
	/** mergedStack: the sibling properties it views, in order. */
	Stacks?: readonly string[];
}

export interface SubStateDef {
	Props: Readonly<Record<string, PropDef>>;
}

export interface DeckDef {
	/** Number of components. Their deck indexes are 0 .. Size-1. */
	Size: number;
	/** The components' static values, by deck index. Not used by the core. */
	Components?: readonly unknown[];
}

export interface EnumDef {
	/** Decimal enum key to the string a value serializes as. */
	Values: Readonly<Record<string, string>>;
	/** The key a hidden enum value collapses to (Enum.DefaultValue()). */
	DefaultValue: number;
}

export interface GameDefinition {
	Name: string;
	Decks: Readonly<Record<string, DeckDef>>;
	Enums: Readonly<Record<string, EnumDef>>;
	/**
	 * The enum whose values are player groups (delegate.GroupEnum()); the
	 * base delegate uses the enum named "group". Empty or absent: none.
	 */
	GroupEnum?: string;
	Game: SubStateDef;
	Player: SubStateDef;
	/** Deck name to the schema of that deck's per-component dynamic values. */
	DynamicComponentValues: Readonly<Record<string, SubStateDef>>;
}

/** sanitization.go SanitizationDefaultGroup. */
export const GROUP_ALL = 'all';
/** sanitization.go SanitizationDefaultPlayerGroup. */
export const GROUP_OTHER = 'other';
export const GROUP_SELF = 'self';

/** policyFromString: trims, lowercases, and yields 'invalid' when unknown. */
export function policyFromString(name: string): PolicyName {
	switch (name.toLowerCase().trim()) {
		case 'visible':
			return 'visible';
		case 'order':
			return 'order';
		case 'len':
			return 'len';
		case 'nonempty':
			return 'nonempty';
		case 'hidden':
			return 'hidden';
	}
	return 'invalid';
}

/**
 * policyFromStructTagWithDefault (struct_inflater.go): parses a `sanitize`
 * struct tag. A config with no group applies to defaultGroup. Group names are
 * used verbatim (not trimmed), as in Go.
 */
export function parseSanitizeTag(
	tag: string,
	defaultGroup: string,
	missingPolicy: PolicyName = 'visible',
): Record<string, PolicyName> {
	if (tag === '') return { [defaultGroup]: missingPolicy };
	const result: Record<string, PolicyName> = {};
	for (const piece of tag.split(',')) {
		const split = piece.split(':');
		if (split.length > 2) return { [GROUP_ALL]: 'invalid' };
		if (split.length === 1) {
			result[defaultGroup] = policyFromString(split[0]);
		} else {
			result[split[0]] = policyFromString(split[1]);
		}
	}
	return result;
}

/** The policy map for one property, resolved from either accepted form. */
export function policyMapFor(prop: PropDef, defaultGroup: string): Readonly<Record<string, PolicyName>> {
	if (prop.Sanitize === undefined) return { [defaultGroup]: 'visible' };
	if (typeof prop.Sanitize === 'string') return parseSanitizeTag(prop.Sanitize, defaultGroup);
	return prop.Sanitize;
}

export function isStackType(type: PropType): boolean {
	return type === 'stack' || type === 'sizedStack' || type === 'mergedStack';
}
