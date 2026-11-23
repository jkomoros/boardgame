# Enforced reads for legality predicates

Status: implemented on `claude/legal-enforced-reads-b52323`.
Builds on: `2026-07-10-declarative-legality-design.md` (§1 Context, §5 memo),
`2026-07-12-legality-completeness-design.md`.

## Problem

`LegalPredicate.Reads` was an honor system. Go closures can't be
introspected, so the engine trusted each predicate's declared read-set. One
thing depended on that trust for correctness: plan assembly. A predicate
that declares no `move.*` (or `players[move.*]`) read goes into the
field-independent bucket, and that bucket's verdicts are memoized per
`(move type, state version, proposer)`, with no move field values in the key.
A predicate that under-declared a move read therefore let **the server
itself** serve a verdict computed for another move's fields.

The only guard was a boot smoke probe. It evaluated field-independent
game-registered predicates once, using a sentinel move whose reader panics.
It missed:

- reads that depend on context the example state doesn't exhibit (for example,
  a read gated on a real proposer, since the probe runs as Observer);
- concrete type assertions on `ctx.Move` (the sentinel is not the game's type);
- reads of the move made any way other than `Reader()`/`ResolvePath`.

During this work a second instance turned up, in core rather than in a game
predicate. The board-selector grammar (`game.Board[move.Field]`) resolves
through the move, but `legalReadsIncludeMovePath` only recognized `move.X` and
`players[move.F].X`. A catalog-shaped predicate declaring only a
board-selector path was misclassified as field-independent and memoized
without the field.

## Goal

The safe thing should be the only easy thing. Ideally an author never writes
`Reads` for the common case, and a wrong declaration cannot produce a wrong
cached verdict. The worst it can do is produce a predicate that visibly
doesn't work, identically in tests and in production.

Constraints:

- Catalog and game predicates share `LegalPredicate`. They should get one set
  of rules, not two tiers.
- Keep the memo's performance win.
- Facets and per-viewer evaluability, `RequiredReadTypes`/`AllowedReadTypes`,
  and the ledger keep working unchanged.
- Existing registered predicates keep booting, and authors learn no new
  concept.

## Key insight

Staleness happens only when a memoized computation sees an input its key
doesn't cover. The key is `(move type, version, proposer)`. The chest is
constant per manager, and state is a function of version. **The move is the
only uncovered input.** So we don't need to *verify* that a field-independent
predicate ignores the move. We can **withhold the move from it**. Then the
memo key and the predicate's inputs are the same thing by construction.

`LegalContext.Move`'s doc already promised this ("nil during
field-independent evaluation"), but the engine never delivered it. It passed
the real move to every predicate.

## Design

### 1. Move withholding (the soundness mechanism)

`evalLegalPredicate` evaluates every non-opaque predicate whose Reads include
no move-dependent path with `ctx.Move = nil`. This holds always: in the
memoized hot path, in full-ledger mode, and with no `*Game`. The predicate is
treated the same whether it sits at top level or inside a field-dependent
`any`.

Consequences for each under-declaration shape:

| Shape | Before | After |
|---|---|---|
| `ResolvePath("move.X")`, undeclared | stale memo (probe catches the unconditional case) | scope error, forced `Unknown` |
| `ctx.Move.Reader()` behind a runtime-only condition | stale memo, probe blind | nil-move panic, recovered to `Unknown` |
| `ctx.Move.(*myMove)` unchecked | stale memo, probe blind | panic, `Unknown` |
| `m, ok := ctx.Move.(*myMove)` | stale memo, probe blind | `ok == false`: whatever the author does with no move, deterministically |

The last row is the honest residue. Go can't intercept a comma-ok assertion,
so a predicate that returns `Pass` when the assertion fails will pass every
time. That's an ordinary bug, though, not a cache heisenbug. Its behavior
doesn't depend on evaluation order or memo state, so the author's first test
of an illegal move catches it.

"Move-dependent" is now one function, `legalPathReadsMove`, used by bucketing,
withholding, and the panic guard. It covers `move.X`, `players[move.F].X`, and
any `[move.F]` board selector, which fixes the misclassification above.

### 2. Read scoping on `ctx.ResolvePath`

While the engine evaluates a predicate, the context carries an unexported
per-evaluation enforcement record. `ResolvePath` rejects any path not in the
predicate's declared Reads, returns an error, and records the violation.
`evalLegalPredicate` then returns
`Unknown{Reason: "undeclared read: predicate P resolved X, ..."}`, *whatever
Evaluate returned*. Swallowing the error and returning Pass doesn't work.

Scoping isn't needed for memo soundness (withholding is enough). It earns its
place in three ways:

- it makes undeclared *state* reads through the sanctioned surface loud. That
  matters for per-viewer evaluability, and for the future client evaluator and
  dirty-tracking, which consume Reads;
- it gives a precise, named diagnostic for undeclared move reads instead of a
  generic nil panic;
- it gives the boot probe a precise way to detect undeclared state reads.

The scope is computed once at boot (`sealLegalPredicate`, during
single-threaded resolution and plan assembly) and cached on the predicate as
`readScope`. Per evaluation the cost is one small allocation and a map lookup
per `ResolvePath`. Hand-built predicates evaluated outside a plan get the same
scope computed on the fly, without mutating the shared predicate.

The opaque CustomLegaler wrapper is exempt. It has no read-set, is never
memoized, and needs the move.

### 3. Derived Reads (the authoring win)

If a constructor returns a predicate with `Reads == nil`, the engine derives
one `FacetValues` read per arg that parses as a legal path. Literals such as
`"3"`, `">="`, and phase names fail the path grammar and are skipped. A
non-nil empty slice is an explicit "reads nothing" and is left alone.
Derivation runs inside the resolve closure immediately after each
constructor, so it precedes boot validation (derived paths are type- and
existence-checked like declared ones) and precedes any parent unioning its
children's Reads.

The common game-registered shape (args are paths, Evaluate resolves them)
therefore needs no `Reads`. checkers' `spaceIsBlack` dropped its `Reads` and
kept a `RequiredReadTypes` entry for the arg, which validates against the
derived read. If the derivation is wrong for some predicate, for example
because it reads a path that isn't an arg, enforcement (§2) makes that loud.

`FacetValues` is the most conservative facet, so derivation can only
over-approximate what a client needs. A predicate that wants a narrower facet
for evaluability (for example `FacetCount`) declares Reads explicitly.

### 4. Boot probe, demoted to diagnostic

The probe now evaluates *every* game-registered predicate under its runtime
scope:

- field-independent predicates use the sentinel move, so raw `Reader()` access
  is distinguishable from unrelated panics;
- field-dependent predicates use the real example move.

It reports a move read or an out-of-scope `ResolvePath` as a boot error naming
the move, the predicate, and the path. It still can't see conditional reads,
and it doesn't need to. It exists to turn the unconditional case into an
early, well-worded error.

### 5. `legal.Evaluate` / `boardgame.EvaluateLegalPredicate`

This is the enforced entry point for unit-testing a predicate. The catalog
conformance corpus now runs through it, so every built-in predicate's
`ResolvePath` calls are checked against its declared Reads, case by case. All
of them were already honest.

## Alternatives considered

1. **Capability-scoped context only, with no withholding.** Scoping
   `ResolvePath` alone can't cover `ctx.Move.Reader()`, type assertions, or
   `Info().ConcreteMove()`, because the raw move would still be in the context.
   We adopted it as a complement (§2), not as the mechanism.

2. **Remove `ctx.Move` / `ctx.State` from game predicates and allow only
   `ResolvePath`.** This is the strongest form, but it breaks every existing
   predicate. Catalog predicates (`inPhase`, `inProgression`,
   `AllActivePlayers`, `proposerIsCurrentPlayer`) legitimately need raw state,
   and the change would split the shared `Predicate` type into tiers.
   Withholding gets the soundness benefit with no API removal: `ctx.Move`
   still exists and is non-nil exactly when declared.

3. **Derive Reads from args and nothing else.** Derivation helps authoring but
   can't stop a predicate from reading a path that isn't an arg, so on its own
   it is still an honor system. We kept it as sugar on top of enforcement.

4. **Runtime read-tracking in debug/test mode, then compare against Reads.**
   This finds bugs only on paths the tests exercise. It adds a mode authors
   must know to enable, and it leaves production unprotected. Withholding
   protects production at about the same cost.

5. **Include move fields in the memo key when reads are "uncertain".** Every
   game-registered predicate would be uncertain, so this either drops the
   memo for them or hashes every move's fields on every evaluation. It also
   treats the symptom: the declared read-set stays a lie that evaluability
   still consumes. Rejected.

6. **Stronger static probes (several example states, fuzzed proposers).**
   These only shrink the blind spots, at boot-time cost. Rejected in favor of
   making the blind spots harmless.

## Compatibility

- Existing predicates whose Reads are honest see no change. A field-dependent
  predicate still gets the real move, and a field-independent one never needed
  it.
- Existing predicates with *dishonest* Reads change from "sometimes stale" to
  "deterministically Unknown". That is the intended break, and the boot probe
  names the common case.
- `Reads: nil` used to mean "reads nothing" and now means "derive from args".
  A predicate whose args are paths but that really reads nothing now gets
  over-declared reads. That's harmless (over-approximation), except that boot
  validation will now check those paths exist. Use `Reads: []legal.Read{}` to
  opt out.
- `LegalContext` gained one unexported field. Keyed literals are unaffected.
  The repository has no unkeyed literals.
- The full-ledger `Unknown` reason string for the nil-move guard now starts
  with `"undeclared move read"` and continues with detail. It was previously
  exactly that string.

## Open questions

- **Raw `ctx.State` reads stay unscoped.** They can't cause staleness, but they
  can under-report Reads to a future client evaluator. Closing this would mean
  a read-scoped `ImmutableState` wrapper, which is a large surface, or
  migrating catalog predicates like `inPhase` to `ResolvePath`. It's worth
  doing before game predicates can become `ClientEvaluable`.
- **Comma-ok assertion on a withheld move.** Should the engine pass a
  non-nil "withheld" sentinel move instead of nil? Then `Reader()`/`Info()`
  could record a violation and force `Unknown`, as `ResolvePath` does. A
  comma-ok assertion would still just fail, so the residue in §1 wouldn't
  shrink, and nil matches the contract the docs already promised. Kept nil.
- **Derived facet.** Should derivation pick the facet from
  `RequiredReadTypes`, for example `FacetCount` for a stack whose predicate
  only counts? It can't know what Evaluate does with a stack, so it stays
  `FacetValues` for now.
