# Adversarial critique — Sub-project A initial design (pre-spec)

Verdict: **rethink** — two items structurally unsound, one mis-scoped. The
committed spec (`docs/superpowers/specs/2026-07-12-legality-completeness-design.md`)
is the amended design this critique prescribed. Preserved findings:

## Critical

1. **`proposer.X` incompatible with `LegalForAnyone`; doesn't restore darwin.**
   Darwin's reverted checks index by MOVE FIELD (`players[m.TargetPlayerIndex]`,
   ../games/darwin/moves.go:531), not the proposer. And `LegalForAnyone` =
   plan evaluation under AdminPlayerIndex (server/api/main.go:1765);
   `proposer.X` → Unknown under Admin → LegalForAnyone permanently false for
   every proposer-gated move, unlike the old imperative behavior. Fix adopted:
   move-field-indexed player paths (`players[move.Field].X`); proposer-relative
   legality deferred pending a LegalForAnyone redesign.
2. **DealCountComponents partial contribution desyncs the ledger.** Its
   Legal() is round-robin state machine + MayMoveTo (moves/deal_components.go:243-285),
   not a count check; contributing only a count predicate makes the ledger
   report legal while Legal() rejects — the same divergence class the branch
   boot-errors elsewhere. Fix adopted: stays blocked until round-robin
   predicates exist.

## Important

3. **MayMoveTo boot-time `HasConstraints()` narrowing unsound** while
   `Stack.AddConstraint` (stack.go:396) is public runtime API — a mid-game
   value-reading constraint would leak value comparisons to Order-sanitized
   viewers. Fix adopted: deferred; pessimistic FacetValues stands.
4. **StackEmpty on FacetCount is inevaluable under PolicyNonEmpty** — the one
   policy that reveals exactly emptiness (facetSurvives, legal_path.go:249).
   Fix adopted: new `LegalFacetNonEmpty`.
5. **Hardcoded seam allowlist is brittle**; FixUpMulti's
   `AllowMultipleInProgression` (moves/fixup.go:48) interaction with the
   contributed inProgression atom unverified. Fix adopted: go/parser
   source-level invariant test + explicit FixUpMulti equivalence test gate.
6. **catalogVersion skew handling doesn't exist** (nothing consumes
   LegalCatalogVersion; evaluable is server-computed). Fix adopted: bump
   without behavioral claims.

## Mis-scoping corrected

- Valentine's real blockers are enum/PlayerIndex-typed comparisons
  (../games/valentine/player_moves.go:384), not bool negation → spec §2
  PropEquals typed dispatch.
- Darwin's `DoneWithPhase==false` negation need is coupled to the
  current-player-is-Admin problem → needs §3 paths, not just §4 leaves.
- debuganimations: only 4 of the blocked moves are single-threshold counts;
  the disjunction-of-conjunctions pair stays LegalCustom (no `all`
  compositor; `any` depth-1).
- Signature-change blast radius of a BootContext change would have been
  ~15 files, not 1 (moot — narrowing deferred).

## Verified-true (kept for the record)

- FixUp/FixUpMulti/StartPhase declare no Legal() override — legality IS
  Default.Legal; StartPhase's Apply hooks are irrelevant to the seam.
- checkers is the only game-registered constructor; tutorial snippets don't
  touch constructor signatures.
- FacetCount survives Visible/Order/Len and is currently unused — StackCount
  is its first consumer.
