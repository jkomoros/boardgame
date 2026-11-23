# Go bugs: the register

Every behavior of the Go engine, its stock moves or the example games that
looks like a bug, with what carries it through the transpile pipeline, so
that a later fix in Go is matched by a regeneration instead of being
rediscovered as a divergence (`fabric/TRANSPILE-PROPOSAL.md`, "Go bugs the
hand ports reproduced, and how the pipeline treats them"). The rule is the
same for all: a port does what Go does. **None is fixed.** Fixing one is a
separate change, approved by the owner, that regenerates the conformance
corpus, the fuzz corpus, the affected definitions, the generated games and
the pair store together. Exception: the owner had checkers fixed in Go at
commit `2256c0d3` (coordinator branch); rows G6-G10 are marked fixed there.
That commit is not in this stream's base (`17168752`), so nothing here was
re-run against it.

Sources: the proposal's table; `fabric/status/engine-rules.md` ("Things in
the Go that look like bugs", R1-R12); `engine-core.md` (divergences C1-C14);
`engine-legal.md` ("Findings in the Go", L1-L10); `conformance-export.md`
("Engine findings", E1-E12); and this pipeline's own reports under
`fabric/transpile/status/` (checkers-engine, fuzz-driver, go-descriptions,
engine-stock). "Read, not run" means no test here demonstrates it.

**Carried by** names the part of the pipeline that reproduces the
behavior:

- **translator**: the game's own Go, translated as written
  (`boardgame-util/lib/transpile`), so the generated game has it.
- **engine stock move**: a stock move or behavior of package `moves`,
  ported by hand in `server/static/src/engine/rules/` (`stock.ts`,
  `stockRegistry.ts`).
- **engine core**: state, stacks, sanitization, timers, the step
  (`server/static/src/engine/*.ts`, `rules/engine.ts`).
- **legal evaluator**: `server/static/src/engine/legal/`.
- **exporter**: `boardgame-util/lib/gamedef` or `lib/conformance`, which
  reads the value Go actually uses or records what it does.
- **nothing can**: Go has no single answer, or the behavior is a Go
  process property the TypeScript engine does not have.

**Pinned by** says whether a checked-in vector or document fails if the
behavior changes on one side only.

## Example games

| # | Bug | Source | Carried by | Pinned by | Status |
|---|---|---|---|---|---|
| G1 | Blackjack scores ten plus ace as 11, not 21: an ace is promoted only while the hand is below 11 (`currentValue >= targetScore - 10` stops it exactly when promotion would make 21) | `examples/blackjack/state.go:96` (`handValue`); R3, E11 | translator (`blackjack.handValue`), hand port `games/blackjack.ts handValue` | `conformance/blackjack/scenario-bust-and-finish` v8 player 1; the fuzz corpus (historical mutation, fuzz-replay); `TestBrokenRuleFailsProof` (this stream): the translation with `>= 11` turned into `> 11` fails all 4 blackjack pairs whose translation calls it | open; reproduced |
| G2 | Pig's randomly chosen starting player never gets the first turn: nothing calls `ResetForTurn` at set-up, so `DieCounted` is false, Count Die fires at once on the die's initial 1 (a bust) and Finish Turn passes play on | `examples/pig/main.go:64-75` (`FinishSetUp`); R1 | translator | all three pig conformance documents (v1 Count Die, v2 Finish Turn); every pig fuzz document | open; reproduced |
| G3 | Blackjack's turn order runs backwards by one each round: the player whose stand ends a round is still current when the next round is dealt (proposers 0123456, 6012345, 5601234, ...) | `examples/blackjack` round progression with `moves.FinishTurn` / `StartPhase`; R12 (no single line) | translator and engine stock moves together | `conformance/blackjack/scenario-reshuffle` | open; may be intended |
| G4 | Memory's Capture Cards discards the error of `MoveToNextSlot` and carries on | `examples/memory/moves.go:141` | translator (the facade's error-return shape reproduces it); the hand port `games/memory.ts` lets the stack operation throw instead (a hand-port difference no recording shows) | nothing: the destination (`WonCards`) is uncapped, so the error never occurs | open; read, not run |
| G5 | Memory and debuganimations declare `DefaultNumPlayeres` (misspelled), so the base delegate's default of 2 applies | `examples/memory/main.go:48`, `examples/debuganimations/main.go:48` | exporter (reads `DefaultNumPlayers` from the live manager) | exporter test `TestDefaultNumPlayersAsGoUsesIt`; every document's header | open; changes nothing observable |
| G6 | Checkers: `spaceIsBlack` is `spaceIndex % 2 == 0`, which on an eight-wide board marks every even column, not a checkerboard | `examples/checkers/components.go:52-54` | translator (the function is in the subset; its caller, the predicate constructor, is refused `predicate-constructor` and is hand/agent work) | checkers-engine vectors (`rules/checkers.test.ts`, `testdata/gen/checkersvec`): a mutation of the parity fails them | **fixed in Go at `2256c0d3`** (coordinator branch, the owner's checkers fix); not re-verified by this stream, and the checkers vectors and definition must be regenerated against it |
| G7 | Checkers: no legal move from the starting position. With G6, tokens sit in even columns and every diagonal neighbor is odd; each color's graph points toward its own back edge (red `graphUpward`, black `graphDownward`), so every straight neighbor is occupied. In 200 seeds no Move Token, Crown Token or Finish Turn is ever applied; all 4,096 inputs for player 0 are illegal, all 3,328 (token, space, proposer) pairs at that position are illegal | `examples/checkers/components.go:44-45`, `:143-148` (the graph choice), `:52` | translator (graph variables rebuilt from their constructors by the framework-variable rule) and the engine's `lib/graph.ts` | checkers-engine vectors (the post-set-up position); the fuzz driver's 200-seed run is transient (fuzz-driver.md), not checked in | **fixed in Go at `2256c0d3`** (coordinator branch, the owner's checkers fix); not re-verified by this stream, and the checkers vectors and definition must be regenerated against it |
| G8 | Checkers: both players' `Color` is Red in the stored state, so the black tokens belong to no player | `examples/checkers/state.go:25` (`behaviors.PlayerColor` embedded; nothing assigns a color per player) | engine stock behavior (`behaviors.PlayerColor`) and the game's set-up | checkers-engine vectors | **fixed in Go at `2256c0d3`** (coordinator branch, the owner's checkers fix); not re-verified by this stream, and the checkers vectors and definition must be regenerated against it |
| G9 | Checkers: the "middle" square of a jump is finish minus start, not the midpoint, passed to `RangeToValue`; for a move toward a lower row or column an index is negative and `Apply` fails after `Legal` passed. In constructed positions, 86 of 109 legal Move Tokens fail in `Apply` ("Invalid result from range to value"), 22 succeed, 1 is refused by the `same(Color)` constraint | `examples/checkers/moves.go:247-262` (`(*moveMoveToken).Apply`); proposal risk 1 | translator | checkers-engine vectors on constructed positions (`rules/checkers.test.ts`) | **fixed in Go at `2256c0d3`** (coordinator branch, the owner's checkers fix); not re-verified by this stream, and the checkers vectors and definition must be regenerated against it |
| G10 | Checkers: a capture happens only when an opposing token sits on `1_1`: a step names (1,1) and a jump (2,2), which is never between the squares (11 of 567 successful applies capture) | `examples/checkers/moves.go:253-276` (same arithmetic as G9) | translator | checkers-engine vectors | **fixed in Go at `2256c0d3`** (coordinator branch, the owner's checkers fix); not re-verified by this stream, and the checkers vectors and definition must be regenerated against it |
| G11 | Blackjack's Reveal Hidden Card `LegalCustom` dereferences `ImmutableFirst()` without a nil check; the full ledger runs it after an earlier check failed (L5), so the panic text is recorded as an unknown entry. Likewise checkers' Place Token (`index out of range [-1]`) | `examples/blackjack/moves.go:320-326`; `examples/checkers/moves.go` Place Token `LegalCustom` | translator (`nonNil` reproduces the panic text), legal evaluator (the recover) | blackjack conformance documents' legality ledgers | open; reproduced |

## Stock moves and the step (package `moves`, `game.go`)

| # | Bug | Source | Carried by | Pinned by | Status |
|---|---|---|---|---|---|
| S1 | `ForceFinishTurn` ignores the phase it is installed in (overrides `Legal` without `Default.Legal`) | `moves/force_finish_turn.go:102`; R2 | engine stock move | `conformance/blackjack/scenario-bust-and-finish` v1, v3, v21 | open; reproduced |
| S2 | `AnyPlayer.Legal` writes the proposer into `TargetPlayerIndex`, so `legalForAnyone` depends on who asks first | `moves/any_player.go:53`; R4 | engine stock move | `conformance/werewolf/scenario-timed-vote` v11; a mutation fails it | open; reproduced |
| S3 | `DealCountComponents.Legal` can index the player list with -1 when there is no current player | `moves/deal_components.go:249`; R7 | engine stock move (the panic text is returned as a refusal) | nothing | open; read, not run |
| S4 | A default-score game whose scores are all negative has no winner (the maximum starts at 0) | `base/game_delegate.go:563`; R6 | engine stock (default scoring in `rules/engine.ts`) | nothing | open; read, not run |
| S5 | The fix-up cap commits 258 versions before returning `ErrTooManyFixUps` (`recurseCount > 256` from 0); during creation `NewGame` errors while the game sits in storage | `game.go:1194`; R5 | engine core (the step; `createGame` returns the versions and the error together) | engine unit tests only; no Go document (the fuzz driver never reached the cap) | open; reproduced as far as the engine can |
| S6 | `MoveWithInput` names a move two ways in its errors (its own name in one, the caller's spelling in others); move names are matched without case | `move_input_bind.go`; R10 | engine core (the step's input binding) | the rejection vectors | open; reproduced |
| S7 | A timer that fired keeps its deadline; one canceled while inactive clears only the id, so memory's Hide Cards ends `{fired, deadline}` or `{canceled, zero}` by who proposed it | `timer.go:160`; R11 | engine core (`rules/timers.ts`) | memory conformance documents | open; harmless; reproduced |
| S8 | The "moves with no progression" cache is keyed by game name in a package variable: two managers whose delegates share a name share one answer | `moves/default.go:19`, `:570`; R8 | nothing can (a Go process property; the engine has no such cache) | nothing; the stock-move Go games are split into packages to avoid it (engine-stock.md) | open; read, not run |
| S9 | Example delegates memoize their name in a package variable (`memoizedDelegateName`), so managers are not safe to run concurrently: parallel fuzz runs crashed in `NewGameManager` with a torn string read | `examples/pig/main.go:27-39` and the same pattern in each example's `main.go`; fuzz-driver.md call 8 | nothing can (Go concurrency; the TypeScript engine is single-threaded and the exporter reads the name once) | nothing; the fuzz driver runs sequentially | open |
| S10 | `MoveRecords` slices past the end for the state under construction (`cachedHistoricalMoves[:upToVersion]`) | `game.go:496`; R9 | nothing today (reached only from `Legal` with a committed state) | nothing | open; read, not run |
| S11 | `CountAtMost` and `CountBetween` assign a negative `max` to the closure's captured variable on the first call, so one counter value evaluated at two lengths keeps the first length | `moves/count.go:83-90`, `:98-110`; go-descriptions.md | nothing: the exporter fits a counter at the one length its group passes it (decision 4), and the engine evaluates a counter description fresh each time, so the stickiness is not reproduced; no current progression evaluates one counter at two lengths | nothing | open; latent (a counter shared by two `ParallelCount`s with different child counts would differ) |
| S12 | `CountAll`'s doc comment says it is equivalent to `CountBetween(0,-1)`; it is not (`CountBetween(0,-1)` accepts 0..length, `CountAll` only length, and their error texts differ) | `moves/count.go:55` and its comment | nothing (documentation) | `moves` description tests (`DescribeCounter`) fit each separately | open; doc only |
| S13 | `legalPlan.specs`'s comment calls it "the serializable subset of the plan"; it is the full assembled spec list (the opaque part is the separate `custom` tail) | `legal_plan.go`; go-descriptions.md | nothing (documentation) | nothing | open; doc only |

## Engine core (`engine-core.md` divergences)

| # | Bug | Source | Carried by | Pinned by | Status |
|---|---|---|---|---|---|
| C1 | Sanitizing dynamic values through a visible merged stack depends on map order (`Props()` ranged) and can panic `index out of range [-2]` | `sanitization.go:383`, `:398-421`, `:467` | nothing can (several outcomes); the engine never reveals by default, `mergedStacksRevealDynamicValues` gives the other answer; the lint warns on the shape | sanitize vectors (45 of 176 cases) | open; no example game has the shape |
| C2 | A non-visible policy on a merged stack panics (`*mergedStack` is not a `Stack`) | `sanitization.go:579` | engine core (throws `SanitizeError`) | sanitize vectors | open |
| C3 | `sizedStack.ContractSize` leaves the component index stale: a shifted component cannot be moved in the same state | `stack.go:2561` | engine core (index ported literally) | stack vectors | open; reproduced |
| C4 | Resizing one part of an overlap makes state serialization panic (`mergedStack.IDs` indexes by the first part's length) | `stack.go:1274` | not reproduced: the engine serializes and reports a conservation error | stack vectors avoid it (Go has no answer) | open |
| C5 | An empty concatenation serializes `"IDs": null` | `stack.go:1283` | engine core | blackjack documents (`Hand` at the start of each round) | open; reproduced |
| C6 | Sorting an empty stack leaves `"Indexes": null` | `stack.go:2122`, `:2134` | engine core | stack vectors | open; reproduced |
| C7 | `SortComponents` is unstable but deterministic (`sort.Sort`, pdqsort) | `stack.go` `sortComponentsImpl` | engine core (`gosort.ts` ports `sort.Sort`) | gosort and stack vectors | open; reproduced; a ported `less` must answer every pair as Go's does |
| C8 | A merged stack's `IDsLastSeen` drops ids last seen at version 0 | `stack.go:1332` | engine core | every golden history's version 0 | open; reproduced |
| C9 | Under `len`, a sized stack's ids and indexes stop lining up, so `len` still reveals which slots are occupied; `len`-policy stacks keep their (sorted) component ids (E6) | `sanitization.go:766`, `:827` | engine core | sanitize vectors; blackjack documents | open; reproduced |
| C10 | Two components that hold each other in their dynamic stacks would hang sanitization (no visited set) | `sanitization.go:485` | not reproduced: the engine visits each component once | nothing (Go does not return) | open |
| C11 | A hidden `PlayerIndex` becomes 0, a real player | `sanitization.go:548` | engine core | sanitize vectors | open; reproduced |
| C12 | A player state is judged by its own player's groups, not the viewer's (`Red:visible` on a player property means "visible to all if this player is Red") | `sanitization.go:268` | engine core | sanitize vectors; werewolf documents | open; reproduced |

## Legal evaluator (`engine-legal.md` findings)

| # | Bug | Source | Carried by | Pinned by | Status |
|---|---|---|---|---|---|
| L3 | Go's hot path and full ledger can disagree on a hand-mutated state (a verdict memoized by move, version, proposer, predicate) | `legal_plan.go:231`, `legal_memo.go:52` | not reproduced: the engine has no memo and always gives the ledger's answer | `plan.test.ts` pins the count (4 of 1,848 scenarios) | open; a hazard for tools, not games |
| L4 | A read of an indexed board (`game.Board[move.Zone]`) is not treated as a move read, so its verdict is memoized without the move's fields | `legal_predicate.go:815` | legal evaluator (reports the same `fieldDependent: false`) | nothing | open; read, not run |
| L5 | The full ledger runs `LegalCustom` even after an earlier check failed, so a `LegalCustom` written on that assumption panics (G11) | `legal_plan.go:172` | legal evaluator (recover into unknown) and translator (`nonNil`, panic texts) | blackjack documents' ledgers | open; reproduced |
| L6 | `any` silently ignores an admin policy (`legal.Any(...).WithAdminBypass()` does nothing) | `legal_predicate.go:256`, `:279` | legal evaluator | a legal test | open; mirrored |
| L7 | Two definitions of "this viewer can see it" disagree about a property with no policy (`LegalReadEvaluable` true, `PropertyFacetAvailable` fails closed) | `legal_evaluable.go:133`, `state_visibility.go:86` | legal evaluator (follows the first) | Go legal vectors | open; harmless |
| L8 | An out-of-range negative player index panics in `PlayerIndex.Next` | `state.go:485` | legal evaluator (same panic text) | a Go vector (`index out of range [-6]`) | open; reproduced |
| L9 | `enum.ValueFromString` is nondeterministic for two values that normalize alike (map iteration) | `enum/main.go:971` | nothing can; the translator refuses it and the lint flags it | nothing | open |
| L10 | `playerBool` binds the bare property name, `playerBoolAt` the whole path, under one template | `legal/catalog_compare.go:524`, `legal/catalog_behaviors.go:158` | legal evaluator | legal vectors | open; mirrored |
| L1, L2 | The legal corpus names fixtures that exist only as Go test code, and its harness is not the production path | `legal/conformance_test.go:270`, `:623` | nothing (test infrastructure) | the plan vectors cover what the corpus cannot | open; not engine behavior |

## Server, exporter and wire (`conformance-export.md` engine findings)

| # | Bug | Source | Carried by | Pinned by | Status |
|---|---|---|---|---|---|
| E1 | `JSONForPlayer` mixes a historical state with head-of-game `Finished`, `Winners`, `Version`, `ActiveTimers` | `game.go:280-292` | exporter (rewrites the four fields to the viewed version); the engine defines them per version | every document | open; deliberately not copied |
| E2 | First-load `/info` evaluates legality on state 0 with moves bound to the head | `server/api/main.go:1639-1651`, `:1761`; `game.go:858` | nothing (server only); the export binds moves to the evaluated state | nothing | open |
| E3 | Form `Fields` order is not deterministic on the wire (a map is ranged); input-field error order likewise | `server/api/main.go:2425`; `move_input_bind.go` | nothing can | nothing | open |
| E10 | `MoveJSONForPlayer` inflates the stored move against the current state while taking groups from the pre-move state | `move_sanitization.go:123` | nothing observable (the stored blob overrides the defaults) | nothing | open; read, not run |
| E-w | The exporter writes `winners: null` at unsettled versions; the choice fingerprint of an empty schema hashes `null` | `boardgame-util/lib/conformance/export.go`; engine fingerprint | the runner's one undeclared normalization (`winnersOrEmpty` in `testing/conformance.ts`), and the engine | every document | open; the one-line exporter fix is decision 8 |

## Not bugs, listed because they look like ones

E4 (computed properties run on the sanitized state), E5 (`SecretMoveCount`
in the admin view), E7 (seating is an input held outside state), E8
(version 0 is not settled with a set-up chain), E9 (a timer's stored move
has the starting version), C13 (timer ids come from `state.Rand()`) and C14
(`randPermForStack` is dead code) are behaviors a port must reproduce or
know, not defects. Werewolf's `SetUp.Distribute` is `""` because Go's
werewolf declares no `DistributeComponentToStarterStack`; the base
delegate's default errors if called, and nothing calls it (werewolf has no
deck).
