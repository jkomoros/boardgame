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

/** A chest enum as it ships on /info: int-string key -> value NAME. */
export interface ChestEnum {
  DefaultValue: number;
  /** int-as-string -> value name, e.g. {"0":"Red","4":"Black"}. */
  Values: Record<string, string>;
  /** Present only for range enums (e.g. checkers' "spaces"). */
  Dimensions?: number[];
  /** Present only for a TreeEnum: child int-string key -> parent int key. */
  Parents?: Record<string, number>;
}
/** One deck component's IMMUTABLE values (state carries only dynamic values). */
export interface ChestDeckComponent {
  Index: number;
  Values: Record<string, unknown>; // e.g. { Color: "Red" }
}
/** The ComponentChest as it ships on /info (server GameChest). */
export interface GameChest {
  Decks: Record<string, ChestDeckComponent[]>;
  Enums: Record<string, ChestEnum>;
  Constants?: Record<string, unknown> | null;
  LegalTemplates?: Record<string, string>;
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
  /**
   * The chest from /info (GameChest), or null if not yet loaded. Enum/component
   * predicates fail-closed to Unknown when this is null: the state serializes an
   * enum prop as its value NAME with no enum identity, and a deck component's
   * immutable values (e.g. a token's Color) live only in the chest.
   */
  chest: GameChest | null;
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

// resolveBoolPath resolves a path and requires a boolean value (Go
// resolveBoolPath). Missing/non-bool -> ok:false -> Unknown.
function resolveBoolPath(path: string, ctx: EvalContext): { value: boolean; ok: boolean } {
  const r = resolvePath(path, ctx);
  if (!r.ok || typeof r.value !== 'boolean') {
    return { value: false, ok: false };
  }
  return { value: r.value, ok: true };
}

// A stack in RawGameState is { Deck, Indexes, ... } where Indexes is the
// component index in each slot and -1 marks an empty slot (sized stacks) —
// growable stacks carry no -1 padding, so `Indexes !== -1` count is universally
// NumComponents().
interface RawStack {
  Deck: string;
  Indexes: number[];
}
function isRawStack(v: unknown): v is RawStack {
  return (
    typeof v === 'object' &&
    v !== null &&
    Array.isArray((v as RawStack).Indexes) &&
    typeof (v as RawStack).Deck === 'string'
  );
}

// resolveStackPath resolves a path to a stack (Go resolveStackPath). Missing /
// non-stack -> ok:false -> Unknown.
function resolveStackPath(path: string, ctx: EvalContext): { stack: RawStack; ok: boolean } {
  const r = resolvePath(path, ctx);
  if (!r.ok || !isRawStack(r.value)) {
    return { stack: { Deck: '', Indexes: [] }, ok: false };
  }
  return { stack: r.value, ok: true };
}

// stackNumComponents mirrors Go's Stack.NumComponents(): the number of occupied
// slots (a slot is occupied iff its index is not the -1 empty sentinel).
function stackNumComponents(stack: RawStack): number {
  let n = 0;
  for (const idx of stack.Indexes) if (idx !== -1) n++;
  return n;
}

// stackComponentAt mirrors Go's Stack.ImmutableComponentAt(idx) != nil: whether
// slot idx of the stack is occupied (in-bounds and not the -1 sentinel).
function stackComponentPresent(stack: RawStack, idx: number): boolean {
  return idx >= 0 && idx < stack.Indexes.length && stack.Indexes[idx] !== -1;
}

// --- chest-backed enum / component resolution -----------------------------
//
// The RawGameState serializes an enum property as its value NAME ("Red",
// "0_0") and a PlayerIndex as a bare int, carrying NO property type. The Go
// catalog switches on the resolved PropertyType; the client recovers the enum
// by MEMBERSHIP against the /info chest — find the enum whose value-name set
// contains the actual name. Anything unreproducible (no chest, name in no
// enum) is fail-closed to Unknown, matching Go's UnknownVerdict paths.

interface EnumMatch {
  /** The integer key of the value within its enum (== the stack slot index). */
  key: number;
  /** The full value-name set of the matching enum. */
  names: Set<string>;
}

// enumForValue finds the chest enum whose Values contains valueName and returns
// that value's integer key plus the enum's full value-name set. The enum
// identity comes from the resolved value (which serializes only as a name).
// Returns null if no enum contains valueName — the value is then a plain string
// / the chest is absent.
function enumForValue(chest: GameChest | null, valueName: string): EnumMatch | null {
  if (!chest || !chest.Enums) return null;
  for (const enumName of Object.keys(chest.Enums)) {
    const e = chest.Enums[enumName];
    if (!e || !e.Values) continue;
    for (const intStr of Object.keys(e.Values)) {
      if (e.Values[intStr] === valueName) {
        return { key: Number.parseInt(intStr, 10), names: new Set(Object.values(e.Values)) };
      }
    }
  }
  return null;
}

// resolveEnumKey resolves an enum-valued path (Go resolveEnumPath, used for the
// stack KEY of componentPresentAtKey / componentPropEqualsCurrentPlayer). The
// path resolves to a value NAME (string); its integer key comes from the chest
// enum containing it. Missing/non-string/unknown-name -> ok:false -> Unknown.
function resolveEnumKey(path: string, ctx: EvalContext): { key: number; name: string; ok: boolean } {
  const r = resolvePath(path, ctx);
  if (!r.ok || typeof r.value !== 'string') return { key: 0, name: '', ok: false };
  const found = enumForValue(ctx.chest, r.value);
  if (!found) return { key: 0, name: '', ok: false };
  return { key: found.key, name: r.value, ok: true };
}

// lookupComponentValue reads a deck component's IMMUTABLE property from the
// chest (Go: comp.Values().Reader().ImmutableEnumProp(prop)). The state's
// Components carry only dynamic values, so the immutable value (e.g. a token's
// Color) is reachable only here. Returns undefined if unreachable -> Unknown.
function lookupComponentValue(
  chest: GameChest | null,
  deckName: string,
  compIndex: number,
  prop: string,
): unknown {
  if (!chest || !chest.Decks) return undefined;
  const deck = chest.Decks[deckName];
  if (!Array.isArray(deck) || compIndex < 0 || compIndex >= deck.length) return undefined;
  const comp = deck[compIndex];
  if (!comp || !comp.Values) return undefined;
  return comp.Values[prop];
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

// propAtLeast mirrors legal/catalog_compare.go propAtLeastConstructor: resolve
// an int at args[0], Pass iff value >= n (args[1]); unresolvable/non-int ->
// Unknown; else FailT("legal.prop_at_least", {value, min}).
function evalPropAtLeast(spec: LegalSpec, ctx: EvalContext): LegalVerdict {
  const args = spec.args;
  if (!args || args.length !== 2) return unknown();
  const [path, nStr] = args;
  const n = Number.parseInt(nStr, 10);
  if (Number.isNaN(n)) return unknown();
  const r = resolveIntPath(path, ctx);
  if (!r.ok) return unknown();
  if (r.value >= n) return pass();
  const template = spec.message && spec.message.length > 0 ? spec.message : 'legal.prop_at_least';
  return failT(template, { value: r.value, min: n });
}

// playerBool mirrors legal/catalog_compare.go playerBoolConstructor: resolve
// the CURRENT player's bool prop args[0], Pass iff it equals want (args[1],
// "true"/"false", default true); unresolvable/non-bool -> Unknown; else
// FailT("legal.player_bool", {prop, want}). want is bound as its "true"/"false"
// string (Go strconv.FormatBool), matching the Go bindings byte-for-byte.
function evalPlayerBool(spec: LegalSpec, ctx: EvalContext): LegalVerdict {
  const args = spec.args;
  if (!args || (args.length !== 1 && args.length !== 2)) return unknown();
  const prop = args[0];
  let want = true;
  if (args.length === 2) {
    if (args[1] === 'true') want = true;
    else if (args[1] === 'false') want = false;
    else return unknown(); // Go: playerBoolWant errors -> construction failure
  }
  const r = resolveBoolPath(`player.${prop}`, ctx);
  if (!r.ok) return unknown();
  if (r.value === want) return pass();
  const template = spec.message && spec.message.length > 0 ? spec.message : 'legal.player_bool';
  return failT(template, { prop, want: String(want) });
}

// stackCount mirrors legal/catalog_count.go stackCountConstructor: resolve the
// stack at args[0], compare NumComponents() to n (args[2]) via op (args[1]);
// unresolvable -> Unknown; else Pass or FailT("legal.stack_count",{value,op,n}).
function evalStackCount(spec: LegalSpec, ctx: EvalContext): LegalVerdict {
  const args = spec.args;
  if (!args || args.length !== 3) return unknown();
  const [path, op, nStr] = args;
  const cmp = COMPARE_OPS[op];
  if (!cmp) return unknown();
  const n = Number.parseInt(nStr, 10);
  if (Number.isNaN(n)) return unknown();
  const r = resolveStackPath(path, ctx);
  if (!r.ok) return unknown();
  const count = stackNumComponents(r.stack);
  if (cmp(count, n)) return pass();
  const template = spec.message && spec.message.length > 0 ? spec.message : 'legal.stack_count';
  return failT(template, { value: count, op, n });
}

// stackEmpty / stackNotEmpty mirror stackEmptinessConstructor: Pass iff
// NumComponents()==0 equals wantEmpty; bindingless Fail template.
function evalStackEmptiness(wantEmpty: boolean, defaultTemplate: string) {
  return (spec: LegalSpec, ctx: EvalContext): LegalVerdict => {
    const args = spec.args;
    if (!args || args.length !== 1) return unknown();
    const r = resolveStackPath(args[0], ctx);
    if (!r.ok) return unknown();
    const empty = stackNumComponents(r.stack) === 0;
    if (empty === wantEmpty) return pass();
    const template = spec.message && spec.message.length > 0 ? spec.message : defaultTemplate;
    return failT(template, {});
  };
}
const evalStackEmpty = evalStackEmptiness(true, 'legal.stack_empty');
const evalStackNotEmpty = evalStackEmptiness(false, 'legal.stack_not_empty');

// componentPresentAt / componentAbsentAt mirror catalog_stack.go: resolve int
// idxField (args[1]) + stack (args[0]); Pass iff the slot's occupancy matches
// wantPresent; else FailT(template, {index}).
function evalComponentPresence(wantPresent: boolean, defaultTemplate: string) {
  return (spec: LegalSpec, ctx: EvalContext): LegalVerdict => {
    const args = spec.args;
    if (!args || args.length !== 2) return unknown();
    const [stackPath, idxField] = args;
    const idxR = resolveIntPath(idxField, ctx);
    if (!idxR.ok) return unknown();
    const r = resolveStackPath(stackPath, ctx);
    if (!r.ok) return unknown();
    const present = stackComponentPresent(r.stack, idxR.value);
    if (present === wantPresent) return pass();
    const template = spec.message && spec.message.length > 0 ? spec.message : defaultTemplate;
    return failT(template, { index: idxR.value });
  };
}
const evalComponentPresentAt = evalComponentPresence(true, 'legal.component_missing');
const evalComponentAbsentAt = evalComponentPresence(false, 'legal.component_present_unexpected');

// propEquals / propNotEquals mirror catalog_compare.go propEqualsFamilyConstructor:
// resolve args[0], then dispatch on the RESOLVED value's shape (Go dispatches on
// PropertyType, which the state does not carry — see the chest helpers above).
// Negation flips only a DEFINITE match; an Unknown (unresolvable path,
// unparseable target for the resolved type, unknown enum name) is NEVER flipped
// to Pass, exactly as Go's `if negate { match = !match }` sits after every
// UnknownVerdict early-return.
function evalPropEqualsFamily(negate: boolean, defaultTemplate: string) {
  return (spec: LegalSpec, ctx: EvalContext): LegalVerdict => {
    const args = spec.args;
    if (!args || args.length !== 2) return unknown();
    const [path, value] = args;
    const r = resolvePath(path, ctx);
    if (!r.ok) return unknown();

    let actual: string;
    let matched: boolean;
    const v = r.value;

    if (typeof v === 'boolean') {
      // Go TypeBool arm: target must be exactly "true"/"false".
      if (value !== 'true' && value !== 'false') return unknown();
      actual = String(v);
      matched = v === (value === 'true');
    } else if (typeof v === 'number' && Number.isInteger(v)) {
      // Go TypeInt / TypePlayerIndex arms. The state serializes BOTH as a bare
      // int, so route by the TARGET spelling: "observer" -> -1, "admin" -> -2
      // (the real Go constants — ObserverPlayerIndex=-1, AdminPlayerIndex=-2),
      // otherwise an int literal (integer equality is identical for either type).
      let want: number;
      if (value === 'observer') want = -1;
      else if (value === 'admin') want = -2;
      else {
        const n = Number.parseInt(value, 10);
        if (Number.isNaN(n)) return unknown();
        want = n;
      }
      actual = String(v);
      matched = v === want;
    } else if (typeof v === 'string') {
      // Go TypeEnum arm: recover the enum by membership, then require the TARGET
      // to be a valid value NAME in that SAME enum (Go: an unknown value name ->
      // Unknown). A genuine string prop (value in no enum) -> Unknown, matching
      // Go's TypeString default branch. This is the KEY case: player.Color
      // "Bogus" -> Unknown, "Black" -> fail, "Red" -> pass.
      const found = enumForValue(ctx.chest, v);
      if (!found) return unknown();
      if (!found.names.has(value)) return unknown();
      actual = v;
      matched = v === value;
    } else {
      // Stacks/objects/null (Go: TypeStack etc. -> default -> Unknown).
      return unknown();
    }

    if (negate) matched = !matched;
    if (matched) return pass();
    const template = spec.message && spec.message.length > 0 ? spec.message : defaultTemplate;
    return failT(template, { value: actual, want: value });
  };
}
const evalPropEquals = evalPropEqualsFamily(false, 'legal.prop_equals');
const evalPropNotEquals = evalPropEqualsFamily(true, 'legal.prop_not_equals');

// componentPresentAtKey mirrors catalog_stack.go componentPresentAtKeyConstructor:
// resolve the enum-keyed slot (args[1]) and the stack (args[0]); Pass iff
// ImmutableComponentAtKey(key) != nil (== occupancy of slot int(key)).
function evalComponentPresentAtKey(spec: LegalSpec, ctx: EvalContext): LegalVerdict {
  const args = spec.args;
  if (!args || args.length !== 2) return unknown();
  const [stackPath, keyField] = args;
  const keyR = resolveEnumKey(keyField, ctx);
  if (!keyR.ok) return unknown();
  const r = resolveStackPath(stackPath, ctx);
  if (!r.ok) return unknown();
  if (stackComponentPresent(r.stack, keyR.key)) return pass();
  const template = spec.message && spec.message.length > 0 ? spec.message : 'legal.component_missing_key';
  return failT(template, { key: keyR.name });
}

// componentPropEqualsCurrentPlayer mirrors catalog_purpose.go: the component at
// the enum-keyed slot (args[1]) of stack (args[0]) has a prop (args[2]) whose
// value equals the CURRENT player's own prop of the same name. The component's
// value comes from the CHEST deck (immutable), the player's from the state. No
// component at the slot, or any unresolvable side, -> Unknown.
function evalComponentPropEqualsCurrentPlayer(spec: LegalSpec, ctx: EvalContext): LegalVerdict {
  const args = spec.args;
  if (!args || args.length !== 3) return unknown();
  const [stackPath, keyField, prop] = args;
  const keyR = resolveEnumKey(keyField, ctx);
  if (!keyR.ok) return unknown();
  const r = resolveStackPath(stackPath, ctx);
  if (!r.ok) return unknown();
  // Component at slot int(key): stack.Indexes[key] is the deck component index;
  // -1 / out-of-range means the slot is empty (Go: comp == nil -> Unknown).
  const compIndex =
    keyR.key >= 0 && keyR.key < r.stack.Indexes.length ? r.stack.Indexes[keyR.key] : -1;
  if (compIndex < 0) return unknown();
  const compVal = lookupComponentValue(ctx.chest, r.stack.Deck, compIndex, prop);
  if (typeof compVal !== 'string') return unknown();
  const playerR = resolvePath(`player.${prop}`, ctx);
  if (!playerR.ok || typeof playerR.value !== 'string') return unknown();
  if (compVal === playerR.value) return pass();
  const template =
    spec.message && spec.message.length > 0 ? spec.message : 'legal.component_prop_not_current_player';
  return failT(template, { prop });
}

// revealableCardAt mirrors legal/catalog_purpose.go revealableCardAtConstructor:
// resolve int idxField (args[2]) + hidden (args[0]) + visible (args[1]) stacks;
// Pass iff hidden has a component at idx; else Fail "legal.no_card_here" when
// visible ALSO has none there, else Fail "legal.already_revealed". Occupancy
// only (never reads component values) — client-evaluable under sanitize:"order".
function evalRevealableCardAt(spec: LegalSpec, ctx: EvalContext): LegalVerdict {
  const args = spec.args;
  if (!args || args.length !== 3) return unknown();
  const [hiddenPath, visiblePath, idxField] = args;
  const idxR = resolveIntPath(idxField, ctx);
  if (!idxR.ok) return unknown();
  const hidden = resolveStackPath(hiddenPath, ctx);
  if (!hidden.ok) return unknown();
  const visible = resolveStackPath(visiblePath, ctx);
  if (!visible.ok) return unknown();
  if (stackComponentPresent(hidden.stack, idxR.value)) return pass();
  const override = spec.message && spec.message.length > 0 ? spec.message : undefined;
  if (!stackComponentPresent(visible.stack, idxR.value)) {
    return failT(override ?? 'legal.no_card_here', {});
  }
  return failT(override ?? 'legal.already_revealed', {});
}

// PlayerIndex sentinels (Go state.go): the special negative indices.
const OBSERVER_PLAYER_INDEX = -1;
const ADMIN_PLAYER_INDEX = -2;
const ANY_PLAYER_INDEX = -3;

// playerMayBeActive mirrors base.GameDelegate.PlayerMayBeActive for the DEFAULT
// delegate: active unless behaviors.InactivePlayer's PlayerInactive bool is true
// (absent -> active). A game overriding PlayerMayBeActive can't be reproduced
// client-side (no delegate) — an undetectable divergence, the same "by
// convention" class the Go Read on game.CurrentPlayer already documents.
function playerMayBeActive(player: Record<string, unknown> | undefined): boolean {
  if (!player) return false;
  return player.PlayerInactive !== true;
}

// playerIndexValid mirrors PlayerIndex.Valid (state.go) for the default
// delegate: specials always valid; a concrete index must be in-bounds AND active.
function playerIndexValid(p: number, ctx: EvalContext): boolean {
  if (p === ADMIN_PLAYER_INDEX || p === OBSERVER_PLAYER_INDEX || p === ANY_PLAYER_INDEX) {
    return true;
  }
  const players = ctx.state?.Players;
  if (!Array.isArray(players) || p < 0 || p >= players.length) return false;
  return playerMayBeActive(players[p] as Record<string, unknown>);
}

// playerIndexEquivalent mirrors PlayerIndex.Equivalent (state.go) byte-for-byte.
function playerIndexEquivalent(p: number, other: number): boolean {
  if (p < ANY_PLAYER_INDEX || other < ANY_PLAYER_INDEX) return false;
  if (p === OBSERVER_PLAYER_INDEX || other === OBSERVER_PLAYER_INDEX) return false;
  if (p === ADMIN_PLAYER_INDEX || other === ADMIN_PLAYER_INDEX) return true;
  if (p === ANY_PLAYER_INDEX || other === ANY_PLAYER_INDEX) return true;
  return p === other;
}

// proposerIsCurrentPlayer mirrors legal/catalog_players.go: replicates
// moves.CurrentPlayer's TargetPlayerIndex checks. Field-dependent (reads
// move.TargetPlayerIndex; no move -> Unknown). EnsureValid's invalid-raw branch
// calls PlayerIndex.Next, which is delegate/CustomPlayerOrder-dependent and not
// shipped, so fail-close (Unknown) when raw is not already Valid — the valid
// path (all in-repo games / the whole corpus) is faithful; Next never fires.
function evalProposerIsCurrentPlayer(spec: LegalSpec, ctx: EvalContext): LegalVerdict {
  if (spec.args && spec.args.length !== 0) return unknown();
  const raw = resolveIntPath('move.TargetPlayerIndex', ctx);
  if (!raw.ok) return unknown();
  if (!playerIndexValid(raw.value, ctx)) return unknown();
  const target = raw.value; // Valid -> EnsureValid returns raw unchanged.
  const override = spec.message && spec.message.length > 0 ? spec.message : undefined;
  if (target < 0) {
    return failT(override ?? 'legal.proposer_target_invalid', {
      detail: 'The specified target player is not valid',
    });
  }
  if (!playerIndexEquivalent(target, ctx.currentPlayerIndex)) {
    return failT(override ?? 'legal.proposer_not_your_turn', { detail: "it's not your turn" });
  }
  if (!playerIndexEquivalent(target, ctx.proposer)) {
    return failT(override ?? 'legal.proposer_not_your_turn', { detail: "it's not your turn" });
  }
  return pass();
}

// phaseAncestors mirrors enum TreeEnum.Ancestors (enum/tree.go): root+self-
// inclusive up the Parents chain; a flat enum (no Parents) yields just [key].
function phaseAncestors(key: number, parents?: Record<string, number>): number[] {
  if (!parents) return [key];
  const out: number[] = [];
  const seen = new Set<number>();
  let cur = key;
  while (true) {
    out.unshift(cur);
    if (cur === 0 || seen.has(cur)) break; // Go base case val==0 -> [0]
    seen.add(cur);
    const p = parents[String(cur)];
    cur = typeof p === 'number' ? p : 0;
  }
  return out;
}

// inPhase mirrors legal/catalog_framework.go -> boardgame.LegalInPhaseCheck.
// Reproducible only for the conventional PhaseBehavior case (phase == the
// game.Phase enum property) AND only with the chest's phase enum. The state
// ships Phase as its value NAME; the enum (and its key + tree Parents) is
// recovered from the chest by membership. A non-conventional CurrentPhase
// delegate or a missing phase enum fail-closes to Unknown.
function evalInPhase(spec: LegalSpec, ctx: EvalContext): LegalVerdict {
  const args = spec.args ?? [];
  if (args.length === 0) return pass(); // zero phases -> legal in every phase
  const phaseR = resolvePath('game.Phase', ctx);
  if (!phaseR.ok || typeof phaseR.value !== 'string') return unknown();
  const phaseName = phaseR.value;
  if (!ctx.chest || !ctx.chest.Enums) return unknown();
  // Recover the phase enum by membership of the current phase NAME.
  let currentKey: number | undefined;
  let parents: Record<string, number> | undefined;
  for (const enumName of Object.keys(ctx.chest.Enums)) {
    const e = ctx.chest.Enums[enumName];
    if (!e || !e.Values) continue;
    for (const [k, name] of Object.entries(e.Values)) {
      if (name === phaseName) {
        currentKey = Number.parseInt(k, 10);
        parents = e.Parents;
        break;
      }
    }
    if (currentKey !== undefined) break;
  }
  if (currentKey === undefined || Number.isNaN(currentKey)) return unknown();
  const ancestors = phaseAncestors(currentKey, parents);
  for (const a of args) {
    const k = Number.parseInt(a, 10);
    if (!Number.isNaN(k) && ancestors.includes(k)) return pass();
  }
  const template = spec.message && spec.message.length > 0 ? spec.message : 'legal.in_phase';
  return failT(template, { detail: phaseName });
}

type PredicateFn = (spec: LegalSpec, ctx: EvalContext) => LegalVerdict;

const PREDICATES: Record<string, PredicateFn> = {
  propCompare: evalPropCompare,
  propAtLeast: evalPropAtLeast,
  playerBool: evalPlayerBool,
  stackCount: evalStackCount,
  stackEmpty: evalStackEmpty,
  stackNotEmpty: evalStackNotEmpty,
  componentPresentAt: evalComponentPresentAt,
  componentAbsentAt: evalComponentAbsentAt,
  propEquals: evalPropEquals,
  propNotEquals: evalPropNotEquals,
  componentPresentAtKey: evalComponentPresentAtKey,
  componentPropEqualsCurrentPlayer: evalComponentPropEqualsCurrentPlayer,
  revealableCardAt: evalRevealableCardAt,
  proposerIsCurrentPlayer: evalProposerIsCurrentPlayer,
  inPhase: evalInPhase,
  // NOTE: mayMoveTo / mayMoveToSlot are deliberately absent (fail-closed to
  // Unknown): a destination stack's constraints are Go closures that are never
  // serialized to the client, so their legality cannot be soundly reproduced.
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
