# Engine: the legal evaluator

Branch `claude/fabric-c-engine-legal-52fa6d`. A TypeScript evaluator for
boardgame's declarative move legality, in `server/static/src/engine/legal/`.
It covers the predicate catalog, the `any` compositor, the plan with its base
checks and `LegalCustom`, failure messages, and client evaluability.

Nothing was pushed, no PR was opened, `master` is untouched, and no existing
Go source changed. `./scripts/go-local build ./...` and
`./scripts/go-local test ./legal/` pass.

## Results

| Check | Result |
|---|---|
| Conformance corpus (`legal/testdata/conformance`, read in place) | 32 of 32 files, 130 of 130 cases pass |
| Go predicate vectors | 3,965 of 3,965 full verdicts match |
| Go plan scenarios | 1,848 of 1,848 match |
| Go boot errors | 30 of 30 plan-assembly errors and 82 of 82 constructor errors match, text for text |
| Go evaluability | 25 of 25 policy and facet cells, 2,880 of 2,880 read cases match |
| `npm run type-check` (Node 24.22.3) | passes, exit 0 |
| `npm run test:unit` (Node 24.22.3) | 1,029 of 1,029 pass, exit 0 |
| Tests in this directory | 112 top-level, 242 counting conformance subtests; all pass |

### Which commands ran under which Node

Final results are from Node v24.22.3, with the repository's real commands, run
in `server/static`:

| Command | Node | Result |
|---|---|---|
| `npm run type-check` | 24.22.3 | exit 0 |
| `npm run test:unit` | 24.22.3 | 1,029 tests, 1,029 pass, 0 fail, exit 0 |
| `node --test --test-timeout=20000 --import=./scripts/test-budget.mjs 'src/engine/legal/**/*.test.ts'` | 24.22.3 | 242 tests, 242 pass |

The 1,029 are this directory's 242 plus 787 that were there before; the
client-layers session counted 787 at baseline.

History, for the record. Node 24 was installed partway through. Before that
the machine had only Node 18, 20 and 21:

- `npm run test:unit` could not run as written on any of them. It failed at
  baseline, before reaching a test: Node 20 does not expand the script's glob
  (`Could not find '.../src/**/*.test.ts'`), and none of the three strips
  TypeScript types.
- `npm ci` and `npm run type-check` ran as written on Node 20 and 21.
- As a stopgap, this directory's tests were compiled with the repository's
  own `tsc` to a throwaway directory and run with `node --test` on Node 20,
  and once on Node 21.7.1. 242 passed each time. That stopgap is no longer
  used and left nothing in the tree.

## Commits

| Commit | What |
|---|---|
| `d5dc7f04` Add a Go generator for TypeScript legal-evaluator vectors | `testdata/gen` (fixtures, predicate vectors) and its output |
| `fea538ce` Add the TypeScript evaluator for the legal predicate catalog | Types, access interface, paths, predicates, catalog, the conformance harness and the vector tests |
| `4306be11` Generate plan and evaluability vectors from the real Go engine | Synthetic and example-game plans, boot errors, facet survival |
| `40af3b77` Add the legal plan, failure rendering and client evaluability | `plan.ts`, `message.ts`, `evaluable.ts`, `index.ts`, and their tests |
| `449c0847` Document the legal evaluator and pin the out-of-range target panic | `README.md`, one more Go vector |
| (the commit adding this file) | This report |

## What each test proves

| Test file | Proves |
|---|---|
| `conformance.test.ts` | Each corpus case gives the expected outcome and, where pinned, template key. A file has at least 3 cases and names one predicate. Every catalog predicate has a corpus file and the reverse. |
| `vectors.test.ts` | For 3,965 (spec, fixture, proposer) triples, the whole verdict equals Go's: outcome, template key, each binding and its type, and the Unknown reason text. Each of 341 resolved specs has Go's reads, facets, cost, client evaluability, proposer dependence, emitted templates and bindings, and read-type contracts. Each of 82 rejected specs fails with Go's error. Every predicate is seen passing, failing and unknown, except the two that cannot be unknown. |
| `plan.test.ts` | For 1,848 real plan evaluations: the assembled spec list, every ledger entry and verdict in order, the overall verdict, the hot path's verdict, the rendered text, and the per-viewer wire ledger all equal Go's. 30 configurations `NewGameManager` rejects fail here with the same text. Hand-written: the hot path evaluates nothing after the first non-pass; an unknown stops it; `LegalCustom` runs last and only after everything passed; base checks run in the fixed order before authored specs; `inPhase` and `inProgression` take the framework as an input and are unknown without it. |
| `evaluable.test.ts` | `facetSurvives` equals Go for all 25 cells. `legalReadEvaluable` equals `LegalReadEvaluable` for 24 paths, 5 facets, 4 viewers, 3 policy mixes and 2 current players. Bindings are withheld from a ledger entry the viewer could not have computed. |
| `path.test.ts` | Indexed board paths (`game.Board[move.Zone]`) by int and by enum, on game and player boards, with Go's error text. Boot validation of path existence and type. No example game has a board, so these have no Go vectors. |
| `predicate.test.ts` | Evaluation fails closed on a missing predicate, a throw, and a malformed verdict. Admin bypass. `any`'s full Kleene table, depth-one rule by three routes, read union. Player-index validity, next, custom order and equivalence. `Atoi`, `%q`. Rendering. Every catalog predicate's default templates exist and reference only bindings it supplies. |

A mutation check was run by hand: seven semantic changes were made one at a
time (`>=` to `>`, a facet dropped from a policy, unknown no longer stopping
the plan, bindings leaking, `player.` reading the proposer, `any` naming the
last unknown child, observer matching admin). Five failed tests immediately.
The last two survived, and a Go vector and a hand-written table were added
that now catch them.

## Catalog coverage

Corpus cases are from `legal/testdata/conformance`. Go vectors are generated
single-predicate evaluations. Plan entries count how often the predicate
appears in a recorded Go plan ledger.

| Predicate | Corpus cases | Go vectors | Go plan entries | Corpus passes |
|---|---|---|---|---|
| `allActivePlayers` | 4 | 86 | 64 | yes |
| `componentAbsentAt` | 4 | 43 | 224 | yes |
| `componentPresentAt` | 4 | 87 | 0 | yes |
| `componentPresentAtKey` | 3 | 99 | 16 | yes |
| `componentPropEquals` | 3 | 169 | 72 | yes |
| `componentPropEqualsCurrentPlayer` | 4 | 68 | 16 | yes |
| `componentPropNotEquals` | 3 | 59 | 0 | yes |
| `componentPropsEqual` | 3 | 91 | 88 | yes |
| `componentPropsNotEqual` | 3 | 65 | 32 | yes |
| `inPhase` | 3 | 27 | 136 | yes |
| `mayMoveAllTo` | 3 | 102 | 56 | yes |
| `mayMoveCountTo` | 3 | 58 | 56 | yes |
| `mayMoveFirstTo` | 4 | 103 | 128 | yes |
| `mayMoveFirstToSlot` | 4 | 37 | 16 | yes |
| `mayMoveFixedCountTo` | 3 | 91 | 56 | yes |
| `mayMoveTo` | 4 | 98 | 0 | yes |
| `mayMoveToSlot` | 4 | 76 | 64 | yes |
| `maySwapComponents` | 3 | 75 | 0 | yes |
| `maySwapComponentsByKey` | 3 | 83 | 16 | yes |
| `playerBool` | 6 | 32 | 112 | yes |
| `playerBoolAt` | 3 | 201 | 504 | yes |
| `propAtLeast` | 4 | 843 | 736 | yes |
| `propCompare` | 4 | 45 | 64 | yes |
| `propEquals` | 12 | 342 | 264 | yes |
| `propNotEquals` | 10 | 84 | 8 | yes |
| `proposerIsCurrentPlayer` | 5 | 74 | 1,088 | yes |
| `proposerIsPlayerFromMove` | 4 | 199 | 112 | yes |
| `revealableCardAt` | 4 | 43 | 64 | yes |
| `stackConstraints` | 3 | 129 | 120 | yes |
| `stackCount` | 4 | 274 | 304 | yes |
| `stackEmpty` | 3 | 99 | 408 | yes |
| `stackNotEmpty` | 3 | 83 | 776 | yes |
| `any` (compositor) | none | none | 672 | no corpus file |
| `inProgression` (package `moves`) | none | none | 8 | no corpus file |
| `custom` (`LegalCustom` wrapper) | none | none | 664 | no corpus file |

`any` has no corpus file because the corpus harness cannot resolve it (see
finding 2). It is covered through real plans, where it passes, fails and is
unknown, and by a hand-written truth table.

`inProgression` appears in one real plan (checkers' Place Token) and passes in
every recorded scenario. Its fail and unknown paths are covered only by a
hand-written test, because the decision is the host's.

Two things have no Go vectors at all, only hand-written tests: indexed board
paths, since no example game declares a board, and a stack with constraints as
a `mayMove*` destination, since the test stack does not model constraints
(checkers' `CapturedTokens` is the only constrained stack in the fixtures and
no vector moves into it).

## Findings in the Go

None of these changed what the TypeScript does: it mirrors Go in each case.
"Read, not run" marks a conclusion from reading the source that no test here
demonstrates.

1. **The corpus is not self-contained.** Its cases name fixtures that exist
   only as Go test code (`legal/conformance_test.go:270`,
   `legalFixtureBuilders`). The README says a TypeScript suite "is expected to
   consume it verbatim" (`legal/testdata/conformance/README.md:7`), which is
   not possible without them. The generator restates the builders and refuses
   to write output unless the real catalog agrees with every corpus case. An
   earlier export of these fixtures existed and was removed in `d1650668`.
   `fabric/PROPOSAL.md:228` says the corpus has 32 cases; it has 32 files and
   130 cases.

2. **The corpus harness is not the production path.** It calls each
   constructor directly with a nil chest (`legal/conformance_test.go:623`,
   `resolveSpecViaRegistry`) and then `Evaluate` directly. So the corpus
   cannot hold an `any` case (core resolves `any`, outside the registry), does
   no boot validation (several corpus specs name properties that do not exist
   and would be rejected at `NewGameManager`), never applies admin bypass, and
   never exercises `propEquals`' typo guard. The corpus still passes here
   through the production path. The plan vectors cover what it cannot.

3. **Go's hot path and full ledger can disagree on a hand-mutated state.** In
   4 of 1,848 scenarios `LegalEvaluatePlan` and `LegalEvaluateLedger` return
   different verdicts for the same state, move and proposer. All four are an
   admin proposer on memory's Capture Cards or Start Hide Cards Timer. The hot
   path serves a verdict memoized by (move, state version, proposer,
   predicate) (`legal_plan.go:231`, `legal_memo.go:52`), stored when fix-ups
   ran during setup; the generator then rearranges the state without changing
   its version. Real play never produces two states with one version, so this
   is a hazard for tests and tools, not for games. The corpus fixtures mutate
   state the same way. The TypeScript has no memo and always gives the
   ledger's answer; `plan.test.ts` pins the count at four.

4. **A read of an indexed board is not treated as a move read.** Read, not
   run. `legalReadsIncludeMovePath` (`legal_predicate.go:815`) classifies a
   path by its kind only, and `game.Board[move.Zone]` has kind `game`. A
   predicate whose only reads are such paths, for example
   `stackEmpty("game.Piles[move.Pile]")`, is therefore field-independent: its
   verdict is memoized without the move's fields, so two proposals at one
   version that differ only in `Pile` would share a verdict, and the ledger
   would not mark it provisional. No example game has a board, so nothing
   exercises it. The TypeScript reports the same `fieldDependent: false`.

5. **The full ledger runs `LegalCustom` even after an earlier check failed.**
   `legal_plan.go:172` keeps evaluating in full-ledger mode. Example games
   write `LegalCustom` assuming the earlier checks passed
   (`examples/checkers/moves.go:112` says the first dereference "is guaranteed
   non-nil"). In the recorded ledgers checkers' Place Token and blackjack's
   Reveal Hidden Card panic inside `LegalCustom` (`invalid memory address or
   nil pointer dereference`, `index out of range [-1]`); the recover turns
   each into an unknown entry. The move's verdict is unaffected, because the
   earlier failure is already latched. A TypeScript `LegalCustom` has the same
   exposure and the same protection.

6. **`any` silently ignores an admin policy.** `legal_predicate.go:256`
   returns the compositor before the line that copies `spec.AdminPolicy`
   (`:279`), and boot validation checks the predicate's policy, not the
   spec's. `legal.Any(...).WithAdminBypass()` builds and does nothing.
   Mirrored, with a test.

7. **Two definitions of "this viewer can see it" disagree about an unknown
   property.** `legalPolicyForProp` (`legal_evaluable.go:133`) treats a
   property with no policy as visible, so `LegalReadEvaluable` is true for a
   property that does not exist. `PropertyFacetAvailable`
   (`state_visibility.go:86`) fails closed for the same input. The Go vectors
   pin the first behavior and the TypeScript follows it. It is harmless in
   practice: the predicate reading such a property is unknown anyway.

8. **An out-of-range negative player index panics.** `PlayerIndex.Next`
   (`state.go:485`) indexes the player slice at `p+1` without a lower bound,
   so a move whose `TargetPlayerIndex` is below `AnyPlayerIndex` panics inside
   `proposerIsCurrentPlayer` and `proposerIsPlayerFromMove`. Evaluation
   recovers it into an unknown. The TypeScript reproduces the panic text
   (`runtime error: index out of range [-6]`), checked against a Go vector.

9. **`Enum.ValueFromString` can be nondeterministic.** `enum/main.go:971`
   iterates a map comparing trimmed, lower-cased names, so an enum with two
   values that normalize alike returns either key. A binding has to choose;
   the test binding takes the lowest key.

10. **`playerBool` and `playerBoolAt` bind `prop` differently.** The first
    binds the bare property name (`legal/catalog_compare.go:524`), the second
    the whole path (`legal/catalog_behaviors.go:158`), though they share the
    default template `legal.player_bool`. Mirrored. It reads as an oversight
    rather than a decision.

## Where the TypeScript differs from Go

1. **Integers beyond 2^53.** Go's `strconv.Atoi` accepts any int64. A spec
   argument outside the safe-integer range is rejected here as "value out of
   range" rather than rounded.
2. **`inPhase` and `inProgression` without a framework.** Go always has a
   delegate to ask. Here both return unknown with their own reason text when
   `LegalContext.framework` is absent.
3. **No memo.** See finding 3.
4. **Component-field boot validation is not ported.**
   `validateLegalComponentField` (`legal_predicate.go:401`) checks every
   component of a stack's deck for a field and its type. The access interface
   does not expose a deck's components. The requirement is carried on the
   predicate as metadata and nothing checks it, so a misspelled component
   field is an unknown at evaluation time, not an error at assembly.
5. **Read-failure text.** When a property reader throws, the reason carries
   the thrown message. Go's generated readers have their own wording for a
   wrong-type read (`No such Enum prop: X`, `No such Board prop: X`), and
   those two are reproduced because the vectors reach them. Any other reader
   error text depends on the binding.
6. **`%q` of unprintable non-ASCII text** is passed through unescaped. No path
   or argument contains any.

## Open questions

1. **What does a client rebuild a plan from?** The wire ledger ships each
   predicate's name and args only (`server/api/main.go:133`). It omits `sub`,
   `message` and `adminPolicy`, so an `any`'s children, an
   `allActivePlayers`' inner spec, a `WithMessage` key and an admin bypass are
   all invisible to a client. Client-side legality needs the spec list
   itself, per move type, shipped once (the chest or `/info` is the natural
   place). `LegalPlan.specs` is that list. On Common Fabric the game is
   defined in TypeScript and already has it.

2. **Who supplies `currentPlayerIndex` and the two player predicates?**
   `LegalState` asks for `currentPlayerIndex()`, `playerMayBeActive(i)`,
   `customPlayerOrder()` and `playerIsInactive(i)`. In Go these are delegate
   and behavior methods, not state. The engine core has to decide where they
   live; the interface only fixes what they must answer.

3. **Should player-index arithmetic move?** `playerIndex.ts` implements
   `Valid`, `Next`, `EnsureValid` and `Equivalent` because the proposer
   predicates need them and nothing else on this branch has them. If the
   engine core has its own, this copy should be deleted in its favor.

4. **Should `renderLegalMessage` move into the engine?** The engine imports it
   from `src/legal`, which is right while the webapp is the only consumer of
   both. A Common Fabric build that takes `src/engine` alone would have to
   take that one file too. It also pulls a type from `src/types/api.d.ts`.

5. **Enums the chest does not ship.** Blackjack's card `Rank` and `Suit` come
   from package `playingcards` and are not in the chest's enum set, so the
   chest JSON does not carry them. A client cannot resolve those values by
   name, and `propEquals`' typo guard does not know them. The fixtures carry
   them out of band (`otherEnums`).

6. **`ClientEvaluable` for the component-value predicates.**
   `legal/doc.go:379` says they stay `ClientEvaluable=false` "until a
   TypeScript evaluator implements the same closed selector semantics". This
   one now does, and matches Go on 384 vectors. Flipping the flag is a Go
   catalog change with a `LegalCatalogVersion` bump, so it was left alone.
   `mayMoveAllTo`, `mayMoveCountTo`, `mayMoveFixedCountTo` and
   `allActivePlayers` are also `false` in Go and also implemented here.

7. **Stack constraints.** Every `mayMove*` verdict depends on the bound
   stack's `mayMoveTo`. Commit `d1650668` retired an earlier client evaluator
   on the grounds that constraints are Go closures that do not serialize. This
   package assumes the stack implementation answers correctly and has no test
   with a constrained destination. That coverage has to come from the stack
   task.

8. **Vector staleness.** Nothing fails when the Go changes and
   `testdata/*.json` is not regenerated. A Go test that runs the generator
   and compares would close that, but it would be a new file outside this
   task's directory. The conformance-export session added such a test for its
   own corpus; the same pattern would fit.

9. **The per-version exports.** The conformance-export session records each
   viewer's move forms, ledger included, at every version of real games. This
   evaluator has not been run against them. That is the natural end-to-end
   check once the stack and state types are bound, and it would cover real
   mid-game states that the hand-arranged fixtures here do not.

## For the orchestrator

- Bind `access.ts` to the real state, stack and enum types. The contract is in
  the README under "Binding it". `testing/fixtures.ts` is a working binding
  over wire JSON, and `bindWireFixture` shows the whole shape.
- Supply `LegalContext.framework` from the phase and progression task.
- `testdata/gen` is a `main` package inside a `testdata` directory, so
  `./...` patterns skip it. It builds against the module at the repository
  root. Its wrapped delegate has to forward the optional interfaces core
  finds by type assertion (templates, predicate constructors, sanitization,
  game end); a new one in core would need forwarding there too.
- The four JSON files total about 1.1 MB. Stack ids are stripped from the
  exported state records, since legality never reads them, which is most of
  the saving.
