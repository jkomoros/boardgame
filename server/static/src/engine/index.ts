// The engine's public surface. See README.md.

export type { GoRand } from './rand.ts';
export { newGoRand, stateRand, stateRandSeed, stateRandSeedInput } from './rand.ts';
export { fnv1_64, fnv1_64Bytes } from './fnv.ts';
export { sha1Hex, sha1HexBytes } from './sha1.ts';
export { goSort } from './gosort.ts';
export type { U64 } from './u64.ts';

export { componentId, secretMoveCountFor, EMPTY_SLOT, GENERIC_COMPONENT } from './ids.ts';
export type { GameIdentity } from './ids.ts';

export { POLICY_ORDER, GROUP_ALL, GROUP_OTHER, GROUP_SELF, parseSanitizeTag, policyFromString, policyMapFor } from './definition.ts';
export type { GameDefinition, SubStateDef, PropDef, PropType, DeckDef, EnumDef, PolicyName } from './definition.ts';

export { loadEngineState, stateError, isStateError, cloneJSON } from './state.ts';
export type { EngineState, StateError } from './state.ts';
export type { StateJSON, SubStateJSON, BoardJSON, VisibilityJSON, StackRef, SubStateRef, SubStateGroup, LoadOptions } from './state.ts';

export { isStack, isMergedStackView, stackError, isStackError, UNLIMITED_SLOTS, indexesOf } from './stack.ts';
export type { Stack, MergedStackView, StackError } from './stack.ts';
export type { StackJSON, StackView, StackKind } from './stack.ts';

export {
	viewFor, sanitizedForPlayer, sanitizationTransformation, resolveSanitizationPolicy, visibleFacets, facetSurvives,
	sanitizeError, isSanitizeError, OBSERVER_PLAYER_INDEX, ADMIN_PLAYER_INDEX, ANY_PLAYER_INDEX,
} from './sanitize.ts';
export type { SanitizationTransformation, SanitizeOptions, ViewOptions } from './sanitize.ts';
export type { SanitizeError } from './sanitize.ts';
