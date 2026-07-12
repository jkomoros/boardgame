// evaluator.ts — the narrow, on-the-client legality evaluator (Sub-project B,
// design tier v3(b)). It re-evaluates a declarative-legality SPEC (predicate
// name + args) against local game state so a click-to-propose renderer can gray
// illegal candidate targets BEFORE the player commits, with no round-trip.
//
// It is a PREVIEW engine, never the authority: the server's LegalForPlayer is
// the submit gate (an imperative Legal() residue can disagree with the plan —
// footgun F1). Anything this evaluator cannot faithfully reproduce
// (unimplemented predicate, unresolvable/hidden path, a `custom`/compositor
// entry, a non-concrete current player) is fail-closed to `unknown`, and the
// caller must defer to the server verdict — never optimistically "legal".
//
// Correctness is pinned to the Go catalog by the shared conformance corpus
// (legal/testdata/conformance/*.json) via evaluator.conformance.test.ts. Every
// predicate here MUST match its Go Evaluate byte-for-byte on outcome and Fail
// template.
import type { RawGameState } from '../types/game-state';

export type LegalOutcome = 'pass' | 'fail' | 'unknown';

export interface LegalVerdict {
  outcome: LegalOutcome;
  /** Fail (or compositor Unknown) template key, matching the Go Verdict.Message.Template. */
  template?: string;
  /** Bindings the Fail template would substitute (advisory; not needed for the outcome). */
  bindings?: Record<string, string | number | boolean>;
}

/** A legal.Spec as it appears on the wire / in the conformance corpus. */
export interface LegalSpec {
  name: string;
  args?: string[];
  sub?: LegalSpec[];
  message?: string;
}

export interface EvalContext {
  /** The RawGameState the client already holds (server StorageRecord shape). */
  state: RawGameState;
  /** The move's field values (move.* paths), or null (no move -> move.* is Unknown). */
  move: Record<string, unknown> | null;
  /** The delegate-resolved current player index (NOT a state field; must be supplied). */
  currentPlayerIndex: number;
  /** The proposer of this evaluation. */
  proposer: number;
}

const pass = (): LegalVerdict => ({ outcome: 'pass' });
const unknown = (): LegalVerdict => ({ outcome: 'unknown' });
const failT = (
  template: string,
  bindings: Record<string, string | number | boolean>,
): LegalVerdict => ({ outcome: 'fail', template, bindings });

// --- path resolution -------------------------------------------------------

interface Resolved {
  value: unknown;
  ok: boolean;
}
const UNRESOLVED: Resolved = { value: undefined, ok: false };

// resolvePath resolves a legal path against the context. It currently supports
// the scalar single-level kinds the narrow evaluator needs: "game.X",
// "player.X" (the current player), and "move.X". Anything else — a
// quantifier/stack path, a non-concrete current player, a missing field —
// resolves ok:false, which the callers turn into Unknown (fail-closed).
export function resolvePath(path: string, ctx: EvalContext): Resolved {
  const dot = path.indexOf('.');
  if (dot < 0) return UNRESOLVED;
  const kind = path.slice(0, dot);
  const prop = path.slice(dot + 1);
  if (prop.length === 0 || prop.includes('.')) {
    // Nested/stack paths not supported by the narrow evaluator yet.
    return UNRESOLVED;
  }
  switch (kind) {
    case 'game': {
      const game = ctx.state?.Game as Record<string, unknown> | undefined;
      return readProp(game, prop);
    }
    case 'player': {
      const players = ctx.state?.Players;
      const idx = ctx.currentPlayerIndex;
      // player.X requires a concrete, in-bounds current player (Go: player.*
      // resolves against ImmutableCurrentPlayer(); Admin/Observer/out-of-bounds
      // -> Unknown).
      if (!Array.isArray(players) || !Number.isInteger(idx) || idx < 0 || idx >= players.length) {
        return UNRESOLVED;
      }
      return readProp(players[idx] as Record<string, unknown>, prop);
    }
    case 'move': {
      if (ctx.move == null) return UNRESOLVED;
      return readProp(ctx.move, prop);
    }
    default:
      return UNRESOLVED;
  }
}

function readProp(obj: Record<string, unknown> | undefined, prop: string): Resolved {
  if (!obj || !Object.prototype.hasOwnProperty.call(obj, prop)) return UNRESOLVED;
  return { value: obj[prop], ok: true };
}

// resolveIntPath resolves a path and requires an integer value (Go
// resolveIntPath). A missing path or a non-integer value is ok:false -> Unknown.
function resolveIntPath(path: string, ctx: EvalContext): { value: number; ok: boolean } {
  const r = resolvePath(path, ctx);
  if (!r.ok || typeof r.value !== 'number' || !Number.isInteger(r.value)) {
    return { value: 0, ok: false };
  }
  return { value: r.value, ok: true };
}

// --- predicates ------------------------------------------------------------

const COMPARE_OPS: Record<string, (value: number, n: number) => boolean> = {
  '==': (v, n) => v === n,
  '!=': (v, n) => v !== n,
  '<': (v, n) => v < n,
  '<=': (v, n) => v <= n,
  '>': (v, n) => v > n,
  '>=': (v, n) => v >= n,
};

// propCompare mirrors legal/catalog_compare.go propCompareConstructor's
// Evaluate: resolve an int at args[0], compare with op args[1] to n args[2];
// unresolvable/non-int -> Unknown; else Pass or FailT("legal.prop_compare",
// {value, op, n}).
function evalPropCompare(spec: LegalSpec, ctx: EvalContext): LegalVerdict {
  const args = spec.args;
  if (!args || args.length !== 3) return unknown();
  const [path, op, nStr] = args;
  const cmp = COMPARE_OPS[op];
  if (!cmp) return unknown();
  const n = Number.parseInt(nStr, 10);
  if (Number.isNaN(n)) return unknown();
  const r = resolveIntPath(path, ctx);
  if (!r.ok) return unknown();
  if (cmp(r.value, n)) return pass();
  const template = spec.message && spec.message.length > 0 ? spec.message : 'legal.prop_compare';
  return failT(template, { value: r.value, op, n });
}

type PredicateFn = (spec: LegalSpec, ctx: EvalContext) => LegalVerdict;

const PREDICATES: Record<string, PredicateFn> = {
  propCompare: evalPropCompare,
};

/** The predicate registry names the narrow evaluator can currently reproduce. */
export const IMPLEMENTED_PREDICATES: readonly string[] = Object.keys(PREDICATES);

// evaluateSpec evaluates a single (non-compositor) spec against the context.
// An unimplemented predicate name -> Unknown (fail-closed): the caller must
// defer to the server verdict. `custom` and compositors (`any`,
// `allActivePlayers`) whose children are not on the production wire are
// intentionally absent and thus fail-closed here.
export function evaluateSpec(spec: LegalSpec, ctx: EvalContext): LegalVerdict {
  const fn = PREDICATES[spec.name];
  if (!fn) return unknown();
  return fn(spec, ctx);
}
