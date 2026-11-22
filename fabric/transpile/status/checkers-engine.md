# Checkers engine gaps: status

Phase 3 preparation of `fabric/TRANSPILE-PROPOSAL.md`: close what the
TypeScript engine lacked to load, compile and run checkers from
`conformance/definitions/checkers.json`, and hold it to real Go. Based on
`435af872`; one local commit. Nothing pushed, no PR.

## What was built

| File | What |
|---|---|
| `rules/types.ts` | `GameSpec.predicates?: { name, construct }[]` (Go's `ConfigurePredicateConstructors`); `PredicateSpec`, `PredicateContext`, `GamePredicate`, `GamePredicateConstructor` |
| `rules/compile.ts` | `gamePredicateConstructor(name, construct)`: a game constructor as a registry constructor; `compileGame` overlays them on the built-in catalog (`defaultLegalRegistry(...game)`), so a plan naming one builds and evaluates through the normal plan, ledger and evaluability code |
| `rules/data.ts` | `Predicates[i].Construct` resolved through `bodies` (role `predicate`); new `GameData.ComponentLegal` (deck to FnRef, role `component`) resolved into `DeckSpec.componentLegal`; `Traits.Player.Color` now sets `PlayerTraits.color` (was ignored); the non-empty-Predicates GameSpecError is gone; calling-convention table rows and details for both roles |
| `boardgame-util/lib/gamedef/schema.go` | `GameData.ComponentLegal map[string]FnRef` (`omitempty`, so every checked-in definition is byte-identical) |
| `rules/checkers.test.ts` | The proof (below), with checkers' functions transliterated onto the facade as clearly marked stand-in bodies |
| `rules/testdata/checkers/seed-000{0,1,6}.json.gz` | Go fuzz-driver documents (unseated; seated; one refusal), 16 to 19 KB each |
| `testdata/gen/checkersvec/main.go` | Generator of `testdata/checkers.vectors.json.gz` (156 KB): constructed positions on a live Go checkers state |

No `legal/` change was needed: `defaultLegalRegistry(...gameConstructors)`
and `legalRegistry`'s by-name overlay already existed.

## The predicate body convention

Role `predicate`, `Predicates[i].Construct`:
`(spec: PredicateSpec) => GamePredicate | string`, THROWS.

- `spec` is `{ name, args, message }`: Go's `legal.Spec` (`args` `[]` for
  nil, `message` `''` for none).
- It returns Go's `*legal.Predicate`: `{ name, args?, reads, requiredReadTypes?,
  allowedReadTypes?, cost?, usesProposer?, emittedTemplates?, emittedBindings?,
  evaluate }`, absent fields being Go's zero values, or Go's constructor
  error (thrown, or returned as a string).
- `evaluate(ctx: PredicateContext): LegalVerdict`. The context is all it can
  read: `ctx.state` (RulesState, for the facade's readers), `ctx.move`
  (MoveInstance, or null where Go's `ctx.Move` is nil), `ctx.proposer`, and
  `ctx.resolvePath(path)`, Go's `ctx.ResolvePath` as `[value, PropertyType,
  err]` (enum value `{ enum: { name }, value: key, string }` or null). It
  returns `passVerdict()`, `failT(template, bindings?)` or
  `unknownVerdict(reason)` and must not write. A throw is Go's panic (an
  unknown verdict, `predicate "<name>" panicked: <text>`).
- Go's `chest` and `resolve` constructor arguments are not passed (no game
  uses them).

Role `component`, `ComponentLegal[deck]`:
`(state, index, legalType: number | null) => string | null`, Go's
`(*T).Legal(state, legalType)` on that deck's values type, which
`moves.DefaultComponent` calls.

## Results, as run

| Command | Outcome |
|---|---|
| `node --test src/engine/rules/checkers.test.ts` | 5 of 5 pass |
| `npm run test:unit` | 1,258 tests: 1,252 pass, 6 skipped, 0 fail (base 1,253 / 1,247 / 6, plus 5) |
| `npm run type-check` | passes |
| `npx tsc --noEmit -p src/engine/tsconfig.sandbox.json` | passes |
| `npm run test:layers` | 11 of 11 |
| `./scripts/go-local build ./...` | passes |
| `./scripts/go-local test -count=1 ./boardgame-util/lib/gamedef/... ./boardgame-util/lib/conformance/...` | ok (gamedef, conformance, conformance/corpus) |
| `gamedef export --check` | "The game definitions are current" |
| `conformance export --check` | "The conformance corpus is current" |
| `go vet`, `gofmt -l` on `checkersvec` and `lib/gamedef` | clean |
| `checkersvec` run twice | byte-identical output (same SHA-1) |

### 1. Loads, compiles, header and plans equal Go's

`checkers.json` with the stand-in FnRefs patched in goes through
`gameFromData` and `compileGame` with no `stock:` or `move:` hook: all seven
moves get their stock parts from the real `stockMove` (SeatPlayer,
ActivateInactivePlayer, FixUpMulti, StartPhase, DefaultComponent, FinishTurn,
CurrentPlayer). Every declared move's compiled plan steps equal
`MoveData.Plan.Specs` (Go's `LegalPlanSpecs`) name for name and arg for arg,
custom tail included; `checkers.spaceIsBlack`'s step reads
`move.SpaceIndex` (values), is field-dependent and not client-evaluable,
emits `checkers.black_spaces_only`. The runner's header check (move table,
help texts, fields and enum names, input schema with the two range-enum
fields at codec `integer`, both fingerprints, choice schema, computed
descriptors, variant config, chest decks, enums with `Dimensions`,
constants, merged templates) passes for all three documents.

### 2. Replay of Go's documents: everything matches

| document | versions | viewers | header | state | move | status | view.game | view.move | view.legality | view.choices | result |
|---|---|---|---|---|---|---|---|---|---|---|---|
| checkers/seed-0000 | 26 | 4 | ok 1 | ok 26 | ok 26 | ok 26 | ok 104 | ok 104 | ok 104 | ok 104 | PASS |
| checkers/seed-0001 | 30 | 4 | ok 1 | ok 30 | ok 30 | ok 30 | ok 120 | ok 120 | ok 120 | ok 120 | PASS |
| checkers/seed-0006 | 26 | 4 | ok 1 | ok 26 | ok 26 | ok 26 | ok 104 | ok 104 | ok 104 | ok 104 | PASS |

Set-up, seating (seed 1: Seat Player, Activate Inactive Player), the 24
Place Token fix-ups (Place Token's DefaultsForState, LegalCustom and Apply,
the StackNotEmpty precondition), Start Phase Playing, and the stuck
position: every viewer's view including `AvailableColors`, `ColorValue`,
`GameScore`, and Move Token's full ledger with the game predicate and the
custom tail. Seed 6's one refusal (Move Token with integer input
`TokenIndexToMove: 48`) is refused again by `propose` at version 25 with
Go's text, `that token isn't your token to move`. No first difference.
The base has no fuzz replay helper (`testing/fuzz.ts`); the refusal check
is a small loop in the test that proposes from the document's own state and
move records.

The first real difference found along the way was in the runner, not the
engine: `testing/conformance.ts` `headerOf` wrote no `Dimensions` for a range
enum, which Go's chest carries (`spaces`: `[8, 8]`). Fixed there (two lines,
additive).

### 3. Constructed positions (`checkersvec`), all match

44 positions on a real Go checkers state (the post-set-up position as Go
leaves it, 13 designed, 30 random clustered from a fixed seed), built with
Go's own stack operations and property setters; checkers' code is not
changed and its Legal and Apply are always the ones called. Replayed
against the stand-in bodies, zero mismatches:

| What | Count |
|---|---|
| Move Token `Legal` texts (token: every occupied space and two empty; every space; both proposers) | 35,456, of which 109 legal |
| Move Token full ledgers (verdict and every entry: name, args, verdict with bindings and reasons, serializable, clientEvaluable, fieldDependent, reads) | 4,702 |
| Crown Token defaults (44) and `Legal` for the admin at slots -1 .. 64 | 2,904, 8 legal |
| Finish Turn `Legal` | 44, 6 legal |
| Applies (each on a fresh copy; error text, panic, or the whole state after) | 1,459: 567 nil-error with state after, 11 captures |
| `GameEndConditionMet`, `PlayerScore` per player | 44 each |

What these show about Go's checkers, all observed in Go:

- A move can be legal in a constructed position. `spaceIsBlack` is
  `index % 2 == 0`, so a token on an odd square steps legally to an empty
  diagonal (even) square, and a token on an even square captures legally
  (landing parity is the start's). The post-set-up position has none:
  all 3,328 (token, space, proposer) pairs there are illegal, as the fuzz
  driver found.
- **Legal then Apply fails.** Of the 109 legal Move Tokens, Apply returns
  `Invalid result from range to value` for 86, succeeds for 22, and for 1
  returns `Couldn't capture token: property Color has mixed values: Black
  and Red`. The "middle" square is finish minus start (proposal's risk 1),
  so any move to a lower row has a negative row and no square; only moves
  down and to the right name a real square, (1,1) = `1_1` for a step and
  (2,2) = `2_2` for a jump, and a jump's `2_2` is never between the
  squares. A capture happens only when an opposing token sits on `1_1`
  (11 of the 567 successful applies; one refused by the CapturedTokens
  `same(Color)` constraint).
- Crowning: a red token in row 7 or a black one in row 0, not yet crowned,
  is legal for Crown Token (red moves up and is crowned at the bottom).
- Nine mutations of the stand-in bodies, one at a time (parity of
  `spaceIsBlack`, the middle-square arithmetic, the crown target row, the
  `tokenCaptured` branch, the graph choice, the capture landing test, the
  `Already crowned` check, Crown Token's write, the distribute stack), each
  fail at least one test; the doc replay alone catches three of them.

## Gaps hit, and where each was closed

| Gap | Where | Status |
|---|---|---|
| A game-registered predicate (`checkers.spaceIsBlack`) | `rules/types.ts`, `compile.ts`, `data.ts` | Done |
| A component's `Legal` (interfaces.LegalComponent) had no GameData slot | `GameData.ComponentLegal` in `schema.go` and `data.ts` | Done (schema); the exporter cannot fill it at run time, see below |
| `Traits.Player.Color` ignored by the loader | `data.ts` | Done |
| Range enum `Dimensions` missing from the runner's header | `testing/conformance.ts` | Done (outside my files) |
| Range-enum move fields (`RangeVal`, codec `integer`) | nothing to do | Worked: schema, fingerprint, defaults, ledger reads, integer input |
| `PropDef.Constraints` (`same(Color)`) | nothing to do | Worked: the refused capture's text matches |
| Dynamic values (`Crowned`), stock types, phases, progressions | nothing to do | Worked |

Belonging elsewhere, named precisely:

- **The static half** (`boardgame-util/lib/gamedef/static.go`, another
  agent's file) must fill, for checkers: `SetUp.Distribute`,
  `End.ConditionMet`, `End.PlayerScore`, `Traits.Player.GameScore`,
  `Traits.Player.Turn` (three), `Predicates[0].Construct`,
  `ComponentLegal.Tokens` (new field; needs a `StaticInfo` field and a
  line in `Export`'s join), and the Bodies of Place Token (Defaults,
  LegalCustom, Apply), Move Token (LegalCustom, Apply) and Crown Token
  (Apply). The spellings the test expects are `STAND_IN_REFS` in
  `rules/checkers.test.ts` (`checkers.(*movePlaceToken).DefaultsForState`,
  `checkers.(*gameDelegate).ConfigurePredicateConstructors.func1`, ...); a
  different spelling fails the test with both names.
- **The facade** (`go/api.ts`): no `Owner` for a move's fields, so a
  `RangeVal` move field's `Value()`, `SetValue(k)` and `RangeValue()` have
  no facade call (the stand-in reads `m.fields.X` through the enum); and no
  way to get a `Comp` for a deck index (Go's
  `t.ContainingComponent().ImmutableInstance(state)`, which a component
  `Legal` body needs; the stand-in builds the handle by hand).
- **The translator**: checkers' package variables (`spacesEnum`,
  `graphDownward`, `graphUpward`) have to become a `GraphRef` literal and
  the enum's name; there is no GameData slot for package-level values.

## Judgment calls

1. Game predicates are adapted in `compile.ts` rather than `legal/`: the
   adapter needs the RulesState and MoveInstance the plan's context was
   bound from (the tags `rules/state.ts` sets), and `legal/` must not
   import `rules/`.
2. The constructor gets `{ name, args, message }` only. Go also passes the
   chest and a sub-spec resolver; no game uses them, and leaving them out
   keeps the body API small.
3. `ComponentLegal` is a top-level map by deck, not a field of
   `Definition.Decks` (which is the shared `GameDefinition` the corpus
   header and `testdata/defs` also use).
4. Stand-in bodies live in the test, not in `games/`, as the brief says;
   each FnRef slot is patched only where the data leaves it empty.
5. Three fuzz seeds rather than one: seed 0 is unseated, seed 1 adds the
   seating fix-ups, seed 6 is the only one of 0 to 7 with a refusal.
6. Vectors apply moves directly (`move.Apply(copy)`), not through
   `ProposeMove`, so no fix-up runs after them and no history is needed;
   the state after is compared whole.

## Unverified

- The exact boot-error text for a failing game predicate constructor: the
  test checks only that compileGame fails with the constructor's own text
  after Go's `boardgame: legal spec "<name>": constructor failed: ` (the
  engine prefixes `move "Move Token": `); no Go vector pins the whole text.
- Go's boot smoke probe (a field-independent game predicate evaluated with
  a panicking move reader) is not ported; checkers' predicate reads a move
  field, so it would not be probed.
- `ctx.move === null` (a predicate evaluated with no move) never happens in
  the engine today; it is typed and passed through but not exercised.
- `resolvePath`'s stack and board values are the legal layer's handles, not
  facade `StackRef`s; checkers does not read one.
- Move Token through `propose` (with Finish Turn and Crown Token fixing up
  after it) is not vectored; each move's Legal and Apply are, separately.

## Where the proposal was wrong about the code

- The custom predicate is not agent work in any meaningful sense: it is
  ten lines of `ResolvePath`, a type check and `spaceIsBlack`, and
  transliterates directly against `PredicateContext`.
- "A custom predicate is a fifth kind of imperative body" is right, but it
  is a constructor returning metadata (reads, templates, bindings) as well
  as an evaluate function; the metadata decides the ledger's provisional
  and evaluable flags, so it cannot be dropped.
- Risk 1 (finish-minus-start) is observable after all, but only from a
  constructed position: 86 of 109 legal moves then fail in Apply.
- The engine's `legal/` already accepted game constructors; the gap was
  only in `GameSpec`, `compileGame` and the loader.

## Edits outside my files

- `server/static/src/engine/testing/conformance.ts`: `headerOf` writes a
  range enum's `Dimensions` (two lines). Without it no game with a range
  enum can pass the header check.
- `server/static/src/engine/testdata/gen/regen.sh`: `checkersvec` added to
  the loop and the output list.
- `boardgame-util/lib/gamedef/schema.go`: the `ComponentLegal` field
  (additive, omitempty; the exporter's tests and `--check` pass unchanged).
