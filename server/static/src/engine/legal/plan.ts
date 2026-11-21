// The legal plan: the TypeScript mirror of Go's legal_plan.go. One ordered
// list of predicates per move type, assembled once ("boot") and evaluated on
// every legality question.
//
// Order is the whole point. The base checks a move type contributes come
// first (inPhase, inProgression, stackConstraints, proposerIsCurrentPlayer),
// then the author's specs in declaration order, then the LegalCustom
// function, always last. Evaluation stops at the first verdict that is not a
// pass, so a move keeps its historical first-failure message. Nothing is
// reordered by cost, and nothing is cached here (Go's field-independent memo
// is a performance device with no observable effect; see the README).

import type { LegalChest, LegalContext, LegalPropertyReader, LegalState } from './access.ts';
import { inPhase, inProgression, stackConstraints } from './catalogFramework.ts';
import { proposerIsCurrentPlayer } from './catalogPlayers.ts';
import { goQuote } from './goformat.ts';
import { legalTemplatePlaceholders, renderLegalVerdict } from './message.ts';
import {
  LEGAL_ANY_FAILED_TEMPLATE,
  LEGAL_ANY_NAME,
  type LegalPredicate,
  type LegalRegistry,
  LegalSpecError,
  evalLegalPredicate,
  legalPredicateSerializable,
  legalReadsIncludeMovePath,
  newLegalPredicate,
  resolveLegalSpecs,
  unionLegalReads,
} from './predicate.ts';
import { LEGAL_COST_EXPENSIVE, type LegalRead, type LegalSpec, type LegalVerdict } from './types.ts';

/**
 * What a LegalCustom function reports.
 *   null / undefined  the move is legal (Go: a nil error)
 *   a LegalVerdict    a structured failure; build it with legalErrorf (Go: legal.Errorf)
 *   a string          a plain error; its text becomes a one-off template key (Go: any other error)
 */
export type LegalCustomResult = null | undefined | LegalVerdict | string;

/**
 * Go: CustomLegaler.LegalCustom. The escape hatch: imperative residue the
 * catalog cannot express. It runs after every declared precondition has
 * passed, is never cached, has no serialized form, and a client sees its
 * verdict as "unknown".
 */
export type LegalCustom = (ctx: LegalContext) => LegalCustomResult;

/** Go: legalPlanStep. */
export interface LegalPlanStep {
  predicate: LegalPredicate;
  /** True when the predicate reads move.* or players[move.*]: its verdict depends on the move's fields. */
  fieldDependent: boolean;
}

/** Go: legalPlan. */
export interface LegalPlan {
  moveName: string;
  /** Contributed then authored predicates, in evaluation order. */
  steps: readonly LegalPlanStep[];
  /** The LegalCustom wrapper, evaluated last, or null. */
  custom: LegalPredicate | null;
  /** The de-duplicated union of every step's reads, in first-seen order. */
  allReads: readonly LegalRead[];
  /** The assembled spec list the steps were resolved from. */
  specs: readonly LegalSpec[];
}

/** Go: LegalVerdictEntry. One predicate's line in a full-ledger evaluation. */
export interface LegalVerdictEntry {
  /** Registry name ("any" for the compositor, "custom" for the escape hatch). */
  name: string;
  args: readonly string[];
  verdict: LegalVerdict;
  serializable: boolean;
  clientEvaluable: boolean;
  /** True when the verdict is provisional on the move's field values. */
  fieldDependent: boolean;
  reads: readonly LegalRead[];
}

/**
 * The base checks moves.Default and moves.CurrentPlayer derive from a move's
 * configuration, as inputs. Each is contributed only when configured.
 */
export interface LegalBaseChecks {
  /** The owning move type's name; the inProgression predicate's argument. */
  moveName: string;
  /** Go: WithLegalPhases. Enum keys; empty or absent contributes nothing. */
  legalPhases?: readonly number[];
  /** Go: WithLegalMoveProgression was given a non-nil group. */
  hasMoveProgression?: boolean;
  /**
   * Go: WithSourceProperty + WithDestinationProperty, when both name plain
   * game-state stacks. Property names, not paths.
   */
  stackConstraints?: { source: string; destination: string } | null;
  /** Go: the move embeds moves.CurrentPlayer, which appends proposerIsCurrentPlayer last. */
  currentPlayer?: boolean;
}

/**
 * Go: Default.ContributedPreconditions and CurrentPlayer.
 * ContributedPreconditions. The contributed specs, base-first, in the fixed
 * order: phase, progression, stack constraints, proposer.
 */
export function contributedPreconditions(base: LegalBaseChecks): LegalSpec[] {
  const specs: LegalSpec[] = [];
  if (base.legalPhases && base.legalPhases.length > 0) {
    specs.push(inPhase(...base.legalPhases));
  }
  if (base.hasMoveProgression) {
    specs.push(inProgression(base.moveName));
  }
  if (base.stackConstraints) {
    specs.push(stackConstraints(base.stackConstraints.source, base.stackConstraints.destination));
  }
  if (base.currentPlayer) {
    specs.push(proposerIsCurrentPlayer());
  }
  return specs;
}

/**
 * Go: validateLegalSuppressions. Every suppression must name a spec the move
 * actually contributes; an unmatched name is a typo or an opt-out of a check
 * the move never had. Throws LegalSpecError with Go's text.
 */
export function validateLegalSuppressions(
  moveName: string,
  contributed: readonly LegalSpec[],
  suppressions: readonly string[],
): void {
  if (suppressions.length === 0) {
    return;
  }
  const names = new Set<string>();
  const quoted: string[] = [];
  for (const spec of contributed) {
    if (!names.has(spec.name)) {
      quoted.push(goQuote(spec.name));
    }
    names.add(spec.name);
  }
  const contributedDesc =
    quoted.length > 0
      ? quoted.join(', ')
      : 'none at all — this move contributes no specs, so no suppression could ever match';
  for (const name of suppressions) {
    if (!names.has(name)) {
      throw new LegalSpecError(
        `move ${goQuote(moveName)} suppresses ${goQuote(name)} via WithoutLegalPrecondition but contributes no spec with that name: suppression only removes a CONTRIBUTED atom from the plan, so an unmatched name is either a typo or an opt-out of a check this move never had (this move's contributed spec names: ${contributedDesc})`,
      );
    }
  }
}

/**
 * Go: assembleLegalSpecList. Contributed specs first, minus any whose name is
 * suppressed, then authored specs in declaration order. Authored specs are
 * never suppressed.
 */
export function assembleLegalSpecList(
  contributed: readonly LegalSpec[],
  authored: readonly LegalSpec[],
  suppressions: readonly string[] = [],
): LegalSpec[] {
  const suppressed = new Set(suppressions);
  return [...contributed.filter(spec => !suppressed.has(spec.name)), ...authored];
}

/** Go: newLegalCustomWrapper. */
function customWrapper(custom: LegalCustom): LegalPredicate {
  return newLegalPredicate({
    name: 'custom',
    cost: LEGAL_COST_EXPENSIVE,
    reads: [],
    opaque: true,
    evaluate(ctx) {
      if (ctx.move === null) {
        // Go: a nil Move does not implement CustomLegaler.
        return { outcome: 'unknown', reason: 'custom: move does not implement CustomLegaler' };
      }
      const result = custom(ctx);
      if (result === null || result === undefined) {
        return { outcome: 'pass' };
      }
      if (typeof result === 'string') {
        return { outcome: 'fail', message: { template: result } };
      }
      return result;
    },
  });
}

function hasKey(table: Readonly<Record<string, string>>, key: string): boolean {
  return Object.prototype.hasOwnProperty.call(table, key);
}

/** Go: validateLegalTemplateTree. */
function validateTemplateTree(spec: LegalSpec, pred: LegalPredicate, table: Readonly<Record<string, string>>): void {
  if (spec.message) {
    if (!hasKey(table, spec.message)) {
      throw new LegalSpecError(
        `boardgame: legal spec ${goQuote(spec.name)}: template key ${goQuote(spec.message)} (Spec.Message override) not found in template table`,
      );
    }
  } else if (pred.name === LEGAL_ANY_NAME) {
    if (!hasKey(table, LEGAL_ANY_FAILED_TEMPLATE)) {
      throw new LegalSpecError(
        `boardgame: legal spec ${goQuote(spec.name)}: default template key ${goQuote(LEGAL_ANY_FAILED_TEMPLATE)} not found in template table`,
      );
    }
  }
  const subSpecs = spec.sub ?? [];
  if (subSpecs.length === pred.sub.length) {
    subSpecs.forEach((subSpec, i) => validateTemplateTree(subSpec, pred.sub[i], table));
  }
}

function formatNames(names: readonly string[]): string {
  return names.length === 0 ? 'none' : names.map(goQuote).join(', ');
}

/** Go: validateLegalEmittedTemplatesTree. */
function validateEmittedTemplates(pred: LegalPredicate, table: Readonly<Record<string, string>>): void {
  for (const key of pred.emittedTemplates) {
    if (!hasKey(table, key)) {
      throw new LegalSpecError(
        `boardgame: predicate ${goQuote(pred.name)} may emit template key ${goQuote(key)}, which is not found in the game's template table`,
      );
    }
  }
  pred.sub.forEach(sub => validateEmittedTemplates(sub, table));
}

/** Go: validateLegalEmittedBindingsTree. */
function validateEmittedBindings(pred: LegalPredicate, table: Readonly<Record<string, string>>): void {
  if (pred.emittedBindings !== null) {
    const strays = Object.keys(pred.emittedBindings)
      .filter(key => !pred.emittedTemplates.includes(key))
      .sort();
    if (strays.length > 0) {
      throw new LegalSpecError(
        `boardgame: predicate ${goQuote(pred.name)} declares EmittedBindings metadata for template key(s) ${formatNames(strays)} not listed in its EmittedTemplates (${formatNames(pred.emittedTemplates)}) — a bindings entry for an unlisted key is never validated against any template body, so this is almost certainly a typo'd or stale metadata key; list the key in EmittedTemplates or fix the EmittedBindings entry`,
      );
    }
    for (const key of pred.emittedTemplates) {
      if (!Object.prototype.hasOwnProperty.call(pred.emittedBindings, key) || !hasKey(table, key)) {
        continue;
      }
      const bindings = pred.emittedBindings[key];
      const body = table[key];
      for (const placeholder of legalTemplatePlaceholders(body)) {
        if (!bindings.includes(placeholder)) {
          throw new LegalSpecError(
            `boardgame: predicate ${goQuote(pred.name)} emits template key ${goQuote(key)} whose body (${goQuote(body)}) references placeholder {${placeholder}}, but the predicate does not guarantee a binding named ${goQuote(placeholder)} on every emission of that key (guaranteed bindings: ${formatNames(bindings)}) — the placeholder would render as its bare name mid-game; fix the template body or point the spec at a template whose placeholders the predicate always fills`,
          );
        }
      }
    }
  }
  pred.sub.forEach(sub => validateEmittedBindings(sub, table));
}

/**
 * Go: validateLegalTemplates, validateLegalEmittedTemplates and
 * validateLegalEmittedBindings, in the order NewGameManager runs them. Every
 * template key a spec names or a predicate may emit must be in `table`, and
 * every placeholder in a resolved template body must be a binding the
 * predicate guarantees. Throws LegalSpecError with Go's text.
 */
export function validateLegalPlanTemplates(
  specs: readonly LegalSpec[],
  predicates: readonly LegalPredicate[],
  table: Readonly<Record<string, string>>,
): void {
  if (specs.length !== predicates.length) {
    throw new LegalSpecError(
      `boardgame: validateLegalTemplates: specs and predicates must be parallel slices of equal length, got ${specs.length} specs and ${predicates.length} predicates`,
    );
  }
  specs.forEach((spec, i) => validateTemplateTree(spec, predicates[i], table));
  predicates.forEach(pred => validateEmittedTemplates(pred, table));
  predicates.forEach(pred => validateEmittedBindings(pred, table));
}

/** Everything a plan is assembled from. */
export interface LegalPlanInput {
  /** The move type's name; prefixes every assembly error, as Go's does. */
  moveName: string;
  /**
   * The base checks, already as specs. Use contributedPreconditions() for a
   * Default/CurrentPlayer-shaped move, or pass any list a move type
   * contributes. Evaluated first, in this order.
   */
  contributed?: readonly LegalSpec[];
  /** The author's specs (Go: WithLegalPreconditions), evaluated after the base checks, in this order. */
  authored?: readonly LegalSpec[];
  /** Names of contributed specs to drop (Go: WithoutLegalPrecondition). */
  suppressions?: readonly string[];
  /** The LegalCustom function, run last, or null. */
  custom?: LegalCustom | null;
  /** The constructor registry; see defaultLegalRegistry. */
  registry: LegalRegistry;
  chest?: LegalChest | null;
  /** Validates game/player paths at assembly when given. */
  exampleState?: LegalState | null;
  /** Validates move.* paths at assembly when given. */
  moveReader?: LegalPropertyReader | null;
  /** The game's merged template table; validates template keys and placeholders when given. */
  templates?: Readonly<Record<string, string>> | null;
}

/**
 * Go: the per-move body of assembleLegalPlans followed by
 * buildLegalPlanFromPredicates. Validates suppressions, assembles the spec
 * list, resolves it, validates templates, and records the predicates in
 * evaluation order. Throws LegalSpecError; errors from resolution and
 * template validation are prefixed `move "<name>": ` as in Go.
 */
export function buildLegalPlan(input: LegalPlanInput): LegalPlan {
  const contributed = input.contributed ?? [];
  const authored = input.authored ?? [];
  const suppressions = input.suppressions ?? [];
  validateLegalSuppressions(input.moveName, contributed, suppressions);
  const specs = assembleLegalSpecList(contributed, authored, suppressions);
  let predicates: LegalPredicate[];
  try {
    predicates = resolveLegalSpecs(specs, input.registry, {
      chest: input.chest,
      exampleState: input.exampleState,
      moveReader: input.moveReader,
    });
    if (input.templates) {
      validateLegalPlanTemplates(specs, predicates, input.templates);
    }
  } catch (err) {
    throw new LegalSpecError(`move ${goQuote(input.moveName)}: ${err instanceof Error ? err.message : String(err)}`);
  }
  const steps = predicates.map(predicate => ({
    predicate,
    fieldDependent: legalReadsIncludeMovePath(predicate.reads),
  }));
  return {
    moveName: input.moveName,
    steps,
    custom: input.custom ? customWrapper(input.custom) : null,
    allReads: unionLegalReads(predicates.map(predicate => predicate.reads)),
    specs,
  };
}

/** The result of evaluating a plan. */
export interface LegalPlanResult {
  /** The first verdict that was not a pass, or a pass. This is the move's legality. */
  verdict: LegalVerdict;
  /** Full-ledger mode only: one entry per predicate, in plan order. null in hot-path mode. */
  entries: LegalVerdictEntry[] | null;
}

/**
 * Go: legalPlan.evaluate.
 *
 * Hot-path mode (fullLedger false) stops at the first verdict that is not a
 * pass - a fail OR an unknown, since an unknown means legality could not be
 * confirmed - and returns it.
 *
 * Full-ledger mode evaluates every predicate, returns an entry for each, and
 * reports as the overall verdict the same first non-pass the hot path would
 * have stopped at.
 *
 * Either way every predicate runs through evalLegalPredicate, so a throwing
 * predicate degrades to unknown, never to pass.
 */
export function evaluateLegalPlan(plan: LegalPlan | null | undefined, ctx: LegalContext, fullLedger = false): LegalPlanResult {
  if (!plan) {
    return { verdict: { outcome: 'unknown', reason: 'nil legal plan' }, entries: null };
  }
  let overall: LegalVerdict = { outcome: 'pass' };
  let haveOverall = false;
  const entries: LegalVerdictEntry[] = [];
  const steps: LegalPlanStep[] = [...plan.steps];
  if (plan.custom !== null) {
    steps.push({ predicate: plan.custom, fieldDependent: true });
  }
  for (const step of steps) {
    const verdict = evalLegalPredicate(step.predicate, ctx);
    if (fullLedger) {
      entries.push({
        name: step.predicate.name,
        args: step.predicate.args,
        verdict,
        serializable: legalPredicateSerializable(step.predicate),
        clientEvaluable: step.predicate.clientEvaluable,
        fieldDependent: step.fieldDependent,
        reads: step.predicate.reads,
      });
    }
    if (verdict.outcome !== 'pass' && !haveOverall) {
      overall = verdict;
      haveOverall = true;
      if (!fullLedger) {
        return { verdict: overall, entries: null };
      }
    }
  }
  return { verdict: overall, entries: fullLedger ? entries : null };
}

/**
 * Go: move.Legal(state, proposer) for a move whose legality is its plan
 * (GameManager.LegalEvaluatePlan). '' when the move is legal, otherwise the
 * rendered error text of the first non-pass verdict.
 */
export function legalPlanError(
  plan: LegalPlan,
  ctx: LegalContext,
  table: Readonly<Record<string, string>> | undefined,
): string {
  return renderLegalVerdict(evaluateLegalPlan(plan, ctx).verdict, table);
}
