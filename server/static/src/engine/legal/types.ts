// Value types of the declarative-legality framework: the TypeScript mirror of
// Go's legal_types.go. Everything here is plain data, so a verdict or a spec
// can cross any boundary (wire, worker, storage) unchanged.
//
// The Go core is the reference. Where a comment says "Go:", it names the Go
// symbol the TypeScript is required to match.

/**
 * Go: LegalCatalogVersion. Stamps the predicate vocabulary this evaluator
 * implements. A client holding a ledger stamped with a HIGHER version must
 * treat predicate names it does not know as not evaluable and defer to the
 * server.
 */
export const LEGAL_CATALOG_VERSION = 6;

/**
 * Go: LegalOutcome. Spelled the way the wire ledger and the conformance
 * corpus spell it. There is no fourth "invalid" member: a predicate that
 * returns anything else is converted to `unknown` by evalLegalPredicate.
 */
export type LegalOutcome = 'pass' | 'fail' | 'unknown';

/** Go: LegalBindingValue. The closed scalar union a template binding may hold. */
export type LegalBindingValue = string | number | boolean;

/**
 * Go: LegalMessage. A template key plus named bindings, never a pre-baked
 * string. Structurally identical to the wire's PreconditionMessage, so it can
 * be handed straight to renderLegalMessage.
 */
export interface LegalMessage {
  template: string;
  bindings?: Record<string, LegalBindingValue>;
}

/** Go: LegalVerdict. */
export interface LegalVerdict {
  outcome: LegalOutcome;
  /** Set on fail, and optionally on unknown (the `any` compositor). */
  message?: LegalMessage;
  /** Why the outcome is unknown. Diagnostic text; matches Go byte for byte where Go produces it. */
  reason?: string;
}

/** Go: LegalCost. Metadata only; never affects evaluation order. */
export type LegalCost = 0 | 1 | 2 | 3;
export const LEGAL_COST_TRIVIAL: LegalCost = 0;
export const LEGAL_COST_CHEAP: LegalCost = 1;
export const LEGAL_COST_MODERATE: LegalCost = 2;
export const LEGAL_COST_EXPENSIVE: LegalCost = 3;

/**
 * Go: LegalFacet. Which aspect of a property a predicate reads. The names are
 * the wire spelling (Go visibilityFacetName), the same strings a viewer's
 * state carries in its Visibility block.
 */
export type LegalFacet = 'values' | 'count' | 'occupancy' | 'order' | 'nonempty';

/** Every facet, in Go's declaration order (stateVisibilityFacets). */
export const LEGAL_FACETS: readonly LegalFacet[] = ['values', 'count', 'occupancy', 'order', 'nonempty'];

/** Go: LegalRead. One property a predicate depends on, and the facet it needs. */
export interface LegalRead {
  path: string;
  facet: LegalFacet;
}

/** Go: LegalAdminPolicy. */
export type LegalAdminPolicy = '' | 'bypass';

/**
 * Go: LegalSpec. The serializable form of a predicate or compositor, in the
 * exact JSON shape Go marshals and the conformance corpus stores.
 */
export interface LegalSpec {
  name: string;
  args?: readonly string[];
  sub?: readonly LegalSpec[];
  /** Overrides the predicate's default template key. */
  message?: string;
  adminPolicy?: LegalAdminPolicy;
}

/**
 * Go: PropertyType, spelled as its String() form so error text and vectors
 * compare directly.
 */
export type PropertyType =
  | 'TypeIllegal'
  | 'TypeInt'
  | 'TypeBool'
  | 'TypeString'
  | 'TypePlayerIndex'
  | 'TypeEnum'
  | 'TypeIntSlice'
  | 'TypeBoolSlice'
  | 'TypeStringSlice'
  | 'TypePlayerIndexSlice'
  | 'TypeStack'
  | 'TypeBoard'
  | 'TypeTimer'
  | 'TypeEnumSlice';

/** Go: ObserverPlayerIndex. */
export const OBSERVER_PLAYER_INDEX = -1;
/** Go: AdminPlayerIndex. */
export const ADMIN_PLAYER_INDEX = -2;
/** Go: AnyPlayerIndex. */
export const ANY_PLAYER_INDEX = -3;

/** Go: legal.PassVerdict. */
export function passVerdict(): LegalVerdict {
  return { outcome: 'pass' };
}

/**
 * Go: legal.FailT. Bindings are omitted from the message entirely when none
 * are given, exactly as Go leaves Bindings nil.
 */
export function failT(template: string, bindings?: Record<string, LegalBindingValue>): LegalVerdict {
  return bindings === undefined
    ? { outcome: 'fail', message: { template } }
    : { outcome: 'fail', message: { template, bindings } };
}

/** Go: legal.UnknownVerdict. */
export function unknownVerdict(reason: string): LegalVerdict {
  return { outcome: 'unknown', reason };
}
