// Resolved predicates and how they are run: the TypeScript mirror of Go's
// legal_predicate.go. A LegalSpec (serializable) is resolved through a
// registry of constructors into a LegalPredicate (a closure plus its
// metadata); evalLegalPredicate is the one sanctioned way to run it.

import type { LegalChest, LegalContext, LegalPropertyReader, LegalState } from './access.ts';
import { goQuote } from './goformat.ts';
import { legalPathReadsMove, parseLegalPath, validateLegalPath, validateLegalPathTypes } from './path.ts';
import {
  ADMIN_PLAYER_INDEX,
  type LegalAdminPolicy,
  type LegalCost,
  type LegalRead,
  type LegalSpec,
  type LegalVerdict,
  type PropertyType,
} from './types.ts';

/** Go: legalAnyFailedTemplate. The `any` compositor's default template key. */
export const LEGAL_ANY_FAILED_TEMPLATE = 'legal.any_failed';

/** Go: legalAnyCompositorName. */
export const LEGAL_ANY_NAME = 'any';

/** Go: LegalComponentFieldRequirement. Boot-validation metadata for a static component-value read. */
export interface LegalComponentFieldRequirement {
  stackPath: string;
  field: string;
  allowedTypes: readonly PropertyType[];
  /** When set, enum-typed fields must define this value. */
  requiredEnumValue?: string;
  comparableTypeGroup?: string;
}

/**
 * Go: LegalPredicate. One resolved legality question: a leaf built by a
 * constructor, or the `any` compositor with its children in `sub`.
 */
export interface LegalPredicate {
  /** Registry name ("any" for the compositor, "custom" for the escape hatch). */
  name: string;
  /** The string args it was built with. Empty where Go leaves Args nil. */
  args: readonly string[];
  /** Declared read-set: a conservative over-approximation of what evaluate touches. */
  reads: readonly LegalRead[];
  cost: LegalCost;
  /** The verdict function. Run it through evalLegalPredicate, never directly. */
  evaluate: (ctx: LegalContext) => LegalVerdict;
  /** Boot contract: the property at each path must have exactly this type. */
  requiredReadTypes?: Readonly<Record<string, PropertyType>>;
  /** Boot contract: the property at each path must have one of these types. */
  allowedReadTypes?: Readonly<Record<string, readonly PropertyType[]>>;
  requiredComponentFields?: readonly LegalComponentFieldRequirement[];
  /** Copied from the spec by resolveLegalSpecs. */
  adminPolicy: LegalAdminPolicy;
  /** True when evaluate depends on the proposer other than through a proposer.* read. */
  usesProposer: boolean;
  /** True when the generic client catalog knows this predicate's semantics. */
  clientEvaluable: boolean;
  /** Every template key evaluate may emit, after any message override. */
  emittedTemplates: readonly string[];
  /** Per emitted key, the bindings guaranteed on every emission. null means "no metadata". */
  emittedBindings: Readonly<Record<string, readonly string[]>> | null;
  /** Children; only the `any` compositor has any. */
  sub: readonly LegalPredicate[];
  /** True for the LegalCustom wrapper, which has no serialized form. */
  opaque: boolean;
}

/** Go: (*LegalPredicate).Serializable. */
export function legalPredicateSerializable(p: LegalPredicate | null | undefined): boolean {
  if (!p || p.opaque) {
    return false;
  }
  return p.sub.every(legalPredicateSerializable);
}

/**
 * A spec that cannot be resolved: an unknown name, wrong arguments, a nested
 * `any`, an invalid path. Go reports these as NewGameManager errors; here
 * they are thrown by resolveLegalSpecs, at the equivalent "boot" moment.
 * `message` is Go's error text.
 */
export class LegalSpecError extends Error {
  constructor(message: string) {
    super(message);
    this.name = 'LegalSpecError';
  }
}

/** Go: LegalPredicateConstructor. */
export interface LegalPredicateConstructor {
  /** The registry key; matches the name of every spec it constructs from. */
  name: string;
  /**
   * Builds a predicate from `spec`, throwing LegalSpecError (with Go's text)
   * for invalid arguments. `resolve` resolves a sub-spec through the same
   * registry and rules. `chest` may be null.
   */
  construct(
    spec: LegalSpec,
    chest: LegalChest | null,
    resolve: (spec: LegalSpec) => LegalPredicate,
  ): LegalPredicate;
}

/** A name -> constructor registry. Later entries with the same name win, as in Go. */
export type LegalRegistry = ReadonlyMap<string, LegalPredicateConstructor>;

/** Builds a registry from constructor lists; later lists override earlier ones by name. */
export function legalRegistry(...lists: readonly (readonly LegalPredicateConstructor[])[]): Map<string, LegalPredicateConstructor> {
  const registry = new Map<string, LegalPredicateConstructor>();
  for (const list of lists) {
    for (const ctor of list) {
      if (ctor && ctor.name !== '') {
        registry.set(ctor.name, ctor);
      }
    }
  }
  return registry;
}

/** Fills in the fields most predicates leave at Go's zero value. */
export function newLegalPredicate(
  fields: Pick<LegalPredicate, 'name' | 'reads' | 'cost' | 'evaluate'> & Partial<LegalPredicate>,
): LegalPredicate {
  return {
    args: [],
    adminPolicy: '',
    usesProposer: false,
    clientEvaluable: false,
    emittedTemplates: [],
    emittedBindings: null,
    sub: [],
    opaque: false,
    ...fields,
  };
}

/** Go: legalReadsIncludeMovePath. */
export function legalReadsIncludeMovePath(reads: readonly LegalRead[]): boolean {
  return reads.some(read => legalPathReadsMove(read.path));
}

function thrownText(err: unknown): string {
  return err instanceof Error ? err.message : String(err);
}

/**
 * Go: evalLegalPredicate. The only sanctioned way to run a predicate. It
 * fails closed: a missing predicate, a throwing evaluate (Go: a panic), or a
 * verdict that is not pass/fail/unknown all become `unknown`, never `pass`.
 * A predicate marked AdminBypass passes outright for the admin proposer.
 */
export function evalLegalPredicate(p: LegalPredicate | null | undefined, ctx: LegalContext): LegalVerdict {
  if (!p) {
    return { outcome: 'unknown', reason: 'predicate was nil' };
  }
  if (typeof p.evaluate !== 'function') {
    return { outcome: 'unknown', reason: `predicate ${goQuote(p.name)} has no Evaluate function` };
  }
  if (ctx.proposer === ADMIN_PLAYER_INDEX && p.adminPolicy === 'bypass') {
    return { outcome: 'pass' };
  }
  let verdict: LegalVerdict;
  try {
    verdict = p.evaluate(ctx);
  } catch (err) {
    if (ctx.move === null && !legalReadsIncludeMovePath(p.reads)) {
      return { outcome: 'unknown', reason: 'undeclared move read' };
    }
    return { outcome: 'unknown', reason: `predicate ${goQuote(p.name)} panicked: ${thrownText(err)}` };
  }
  if (!verdict || (verdict.outcome !== 'pass' && verdict.outcome !== 'fail' && verdict.outcome !== 'unknown')) {
    return { outcome: 'unknown', reason: 'predicate returned invalid verdict' };
  }
  return verdict;
}

/**
 * Go: evalLegalAnyKleene. Any child pass -> pass; else any child unknown ->
 * unknown; else fail. Both non-pass outcomes carry the compositor's template
 * and no bindings; the unknown also names the first unknown child.
 */
function evalAnyKleene(subs: readonly LegalPredicate[], ctx: LegalContext, template: string): LegalVerdict {
  let firstUnknown: LegalPredicate | null = null;
  for (const sub of subs) {
    const v = evalLegalPredicate(sub, ctx);
    if (v.outcome === 'pass') {
      return { outcome: 'pass' };
    }
    if (v.outcome === 'unknown' && firstUnknown === null) {
      firstUnknown = sub;
    }
  }
  if (firstUnknown !== null) {
    return {
      outcome: 'unknown',
      message: { template },
      reason: `sub-predicate ${goQuote(firstUnknown.name)} of ${goQuote(LEGAL_ANY_NAME)} was unknown`,
    };
  }
  return { outcome: 'fail', message: { template } };
}

function readKey(read: LegalRead): string {
  return read.facet + '\u0000' + read.path;
}

/** Go: unionLegalReads. De-duplicated union in first-seen order. */
export function unionLegalReads(lists: readonly (readonly LegalRead[])[]): LegalRead[] {
  const seen = new Set<string>();
  const out: LegalRead[] = [];
  for (const reads of lists) {
    for (const read of reads) {
      const key = readKey(read);
      if (!seen.has(key)) {
        seen.add(key);
        out.push(read);
      }
    }
  }
  return out;
}

function typeList(types: readonly PropertyType[]): string {
  return '[' + types.join(' ') + ']';
}

/**
 * Go: unionLegalReadTypes. Keeps the strongest honest type contract per read
 * across a compositor's children and rejects contradictory ones.
 */
function unionReadTypes(subs: readonly LegalPredicate[]): {
  required?: Record<string, PropertyType>;
  allowed?: Record<string, readonly PropertyType[]>;
} {
  const required = new Map<string, PropertyType>();
  const allowed = new Map<string, readonly PropertyType[]>();
  for (const sub of subs) {
    for (const [path, expected] of Object.entries(sub.requiredReadTypes ?? {})) {
      const prior = required.get(path);
      if (prior !== undefined && prior !== expected) {
        throw new LegalSpecError(`composed predicates require conflicting types for ${goQuote(path)} (${prior} and ${expected})`);
      }
      const choices = allowed.get(path);
      if (choices !== undefined && !choices.includes(expected)) {
        throw new LegalSpecError(
          `composed predicates require conflicting types for ${goQuote(path)} (${expected} and one of ${typeList(choices)})`,
        );
      }
      allowed.delete(path);
      required.set(path, expected);
    }
    for (const [path, rawChoices] of Object.entries(sub.allowedReadTypes ?? {})) {
      let choices = rawChoices;
      const expected = required.get(path);
      if (expected !== undefined) {
        if (!choices.includes(expected)) {
          throw new LegalSpecError(
            `composed predicates require conflicting types for ${goQuote(path)} (${expected} and one of ${typeList(choices)})`,
          );
        }
        continue;
      }
      const prior = allowed.get(path);
      if (prior !== undefined) {
        const narrowed = choices;
        choices = prior.filter((candidate, i) => narrowed.includes(candidate) && prior.indexOf(candidate) === i);
        if (choices.length === 0) {
          throw new LegalSpecError(`composed predicates have disjoint allowed types for ${goQuote(path)}`);
        }
      }
      allowed.set(path, [...choices]);
    }
  }
  return {
    required: required.size > 0 ? Object.fromEntries(required) : undefined,
    allowed: allowed.size > 0 ? Object.fromEntries(allowed) : undefined,
  };
}

/** Go: resolveLegalAnySpec. */
function resolveAnySpec(spec: LegalSpec, resolve: (spec: LegalSpec) => LegalPredicate): LegalPredicate {
  const subSpecs = spec.sub ?? [];
  if (subSpecs.length < 2) {
    throw new LegalSpecError(
      `boardgame: legal spec ${goQuote(spec.name)}: ${goQuote(LEGAL_ANY_NAME)} compositor requires at least 2 sub-specs, got ${subSpecs.length}`,
    );
  }
  const subs = subSpecs.map(resolve);
  const template = spec.message ? spec.message : LEGAL_ANY_FAILED_TEMPLATE;
  let types: ReturnType<typeof unionReadTypes>;
  try {
    types = unionReadTypes(subs);
  } catch (err) {
    throw new LegalSpecError(`boardgame: legal spec ${goQuote(spec.name)}: ${thrownText(err)}`);
  }
  return newLegalPredicate({
    name: LEGAL_ANY_NAME,
    reads: unionLegalReads(subs.map(sub => sub.reads)),
    requiredReadTypes: types.required,
    allowedReadTypes: types.allowed,
    cost: subs.reduce<LegalCost>((max, sub) => (sub.cost > max ? sub.cost : max), subs[0].cost),
    sub: subs,
    emittedTemplates: [template],
    // The compositor's own message never carries bindings.
    emittedBindings: { [template]: [] },
    evaluate: ctx => evalAnyKleene(subs, ctx, template),
  });
}

/** Go: checkNoNestedAny. A backstop over the resolved tree, independent of how it was built. */
function checkNoNestedAny(pred: LegalPredicate, insideAny: boolean): string | null {
  const isAny = pred.name === LEGAL_ANY_NAME;
  if (isAny && insideAny) {
    return `resolved predicate ${goQuote(pred.name)}: nested ${goQuote(LEGAL_ANY_NAME)} compositor is not allowed (any is depth-1 only, no nested any anywhere beneath another any)`;
  }
  for (const sub of pred.sub) {
    const nested = checkNoNestedAny(sub, insideAny || isAny);
    if (nested !== null) {
      return nested;
    }
  }
  return null;
}

/**
 * Go: validateLegalPredicateForBoot, minus the component-field deck walk
 * (RequiredComponentFields needs every component of a stack's deck, which
 * the state-access interface deliberately does not expose). Returns Go's
 * error text or null.
 */
export function validateLegalPredicateForBoot(
  pred: LegalPredicate,
  exampleState: LegalState | null,
  moveReader: LegalPropertyReader | null,
): string | null {
  if (pred.adminPolicy === 'bypass') {
    if (!pred.usesProposer) {
      return `predicate ${goQuote(pred.name)} uses AdminBypass without declaring proposer dependence`;
    }
    for (const read of pred.reads) {
      const parsed = parseLegalPath(read.path);
      if (!parsed.ok || (parsed.value.kind !== 'proposer' && parsed.value.kind !== 'move')) {
        return `predicate ${goQuote(pred.name)} uses AdminBypass but read ${goQuote(read.path)} is neither proposer-scoped nor a move field`;
      }
    }
  } else if (pred.adminPolicy !== '') {
    return `predicate ${goQuote(pred.name)} has unknown admin policy ${goQuote(pred.adminPolicy)}`;
  }
  // Go: validateLegalReadsForBoot.
  for (const read of pred.reads) {
    const parsed = parseLegalPath(read.path);
    if (!parsed.ok) {
      return parsed.error;
    }
    if (parsed.value.kind === 'move') {
      if (moveReader === null) {
        continue;
      }
    } else if (exampleState === null) {
      continue;
    }
    const invalid = validateLegalPath(read.path, exampleState, moveReader);
    if (invalid !== null) {
      return invalid;
    }
  }
  const declared = new Set(pred.reads.map(read => read.path));
  const required = pred.requiredReadTypes ?? {};
  const allowed = pred.allowedReadTypes ?? {};
  for (const path of Object.keys(required).sort()) {
    if (!declared.has(path)) {
      return `predicate ${goQuote(pred.name)} requires type metadata for undeclared read ${goQuote(path)}`;
    }
    if (Object.prototype.hasOwnProperty.call(allowed, path)) {
      return `predicate ${goQuote(pred.name)} declares both RequiredReadTypes and AllowedReadTypes for ${goQuote(path)}`;
    }
  }
  for (const path of Object.keys(required).sort()) {
    if (required[path] === 'TypeIllegal') {
      return `predicate ${goQuote(pred.name)} has an illegal RequiredReadTypes contract for ${goQuote(path)}`;
    }
    const invalid = validateLegalPathTypes(path, [required[path]], exampleState, moveReader);
    if (invalid !== null) {
      return invalid;
    }
  }
  for (const path of Object.keys(allowed).sort()) {
    if (!declared.has(path)) {
      return `predicate ${goQuote(pred.name)} allows type metadata for undeclared read ${goQuote(path)}`;
    }
  }
  for (const path of Object.keys(allowed).sort()) {
    const types = allowed[path];
    if (types.length === 0) {
      return `predicate ${goQuote(pred.name)} has an empty AllowedReadTypes contract for ${goQuote(path)}`;
    }
    const seen = new Set<PropertyType>();
    for (const candidate of types) {
      if (candidate === 'TypeIllegal') {
        return `predicate ${goQuote(pred.name)} allows TypeIllegal for ${goQuote(path)}`;
      }
      if (seen.has(candidate)) {
        return `predicate ${goQuote(pred.name)} repeats PropertyType ${candidate} in AllowedReadTypes for ${goQuote(path)}`;
      }
      seen.add(candidate);
    }
    const invalid = validateLegalPathTypes(path, types, exampleState, moveReader);
    if (invalid !== null) {
      return invalid;
    }
  }
  for (const sub of pred.sub) {
    const invalid = validateLegalPredicateForBoot(sub, exampleState, moveReader);
    if (invalid !== null) {
      return invalid;
    }
  }
  return null;
}

/** What resolveLegalSpecs validates declared paths against. Both default to null (skip). */
export interface LegalResolveOptions {
  /** The component chest, for constructors that read it (propEquals' typo guard). */
  chest?: LegalChest | null;
  /** An example state: validates game/player/proposer/players[*] paths. */
  exampleState?: LegalState | null;
  /** An example move's reader: validates move.* paths. */
  moveReader?: LegalPropertyReader | null;
}

/**
 * Go: resolveLegalSpecs. Resolves specs, in order, into predicates.
 *
 * Enforces the anti-tarpit rules: an unknown name is an error; `any` is the
 * only compositor, needs at least two subs, and may not appear anywhere
 * beneath another `any` (enforced at resolution time, so a constructor that
 * resolves an `any` internally is caught too). Every resolved predicate's
 * declared paths are validated against the example state and move reader
 * when those are supplied. Throws LegalSpecError carrying Go's error text.
 */
export function resolveLegalSpecs(
  specs: readonly LegalSpec[],
  registry: LegalRegistry,
  options: LegalResolveOptions = {},
): LegalPredicate[] {
  const chest = options.chest ?? null;
  const exampleState = options.exampleState ?? null;
  const moveReader = options.moveReader ?? null;

  const resolve = (spec: LegalSpec, insideAny: boolean): LegalPredicate => {
    if (spec.name === LEGAL_ANY_NAME) {
      if (insideAny) {
        throw new LegalSpecError(
          `boardgame: legal spec ${goQuote(spec.name)}: nested ${goQuote(LEGAL_ANY_NAME)} compositor is not allowed (any is depth-1 only, no nested any anywhere beneath another any)`,
        );
      }
      // Go: resolveLegalAnySpec's result does not get pred.AdminPolicy set.
      return resolveAnySpec(spec, sub => resolve(sub, true));
    }
    const ctor = registry.get(spec.name);
    if (ctor === undefined) {
      throw new LegalSpecError(
        `boardgame: legal spec ${goQuote(spec.name)}: unknown predicate name ${goQuote(spec.name)}`,
      );
    }
    let pred: LegalPredicate;
    try {
      pred = ctor.construct(spec, chest, sub => resolve(sub, insideAny));
    } catch (err) {
      throw new LegalSpecError(`boardgame: legal spec ${goQuote(spec.name)}: constructor failed: ${thrownText(err)}`);
    }
    if (!pred) {
      throw new LegalSpecError(`boardgame: legal spec ${goQuote(spec.name)}: constructor returned a nil predicate`);
    }
    pred.adminPolicy = spec.adminPolicy ?? '';
    return pred;
  };

  return specs.map(spec => {
    const pred = resolve(spec, false);
    const nested = checkNoNestedAny(pred, false);
    if (nested !== null) {
      throw new LegalSpecError(`boardgame: legal spec ${goQuote(spec.name)}: ${nested}`);
    }
    const invalid = validateLegalPredicateForBoot(pred, exampleState, moveReader);
    if (invalid !== null) {
      throw new LegalSpecError(`boardgame: legal spec ${goQuote(spec.name)}: ${invalid}`);
    }
    return pred;
  });
}
