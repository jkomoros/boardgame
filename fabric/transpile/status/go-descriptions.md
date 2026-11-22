# Go descriptions (workstream 3): status

Workstream 3 of `fabric/TRANSPILE-PROPOSAL.md`. Additive Go only: two new
source files, two new test files, this note. No existing file was edited and
no existing test changed. Based on `a11c3883`; one local commit (the one
adding this file). Nothing pushed, no PR.

## Exported API added

Package `moves` (`moves/describe.go`):

```go
const (
	ProgressionKindMove     = "move"
	ProgressionKindSerial   = "serial"
	ProgressionKindRepeat   = "repeat"
	ProgressionKindParallel = "parallel"
	ProgressionKindUnknown  = "unknown"
)
const (
	CounterKindAny, CounterKindAll, CounterKindInfinite = "any", "all", "infinite"
	CounterKindAtLeast, CounterKindAtMost, CounterKindExactly = "atLeast", "atMost", "exactly"
	CounterKindBetween, CounterKindCustom = "between", "custom"
)
const CounterGridMax = 64

type ProgressionDescription struct {
	Kind     string                   `json:"Kind"`
	Name     string                   `json:"Name,omitempty"`
	Children []ProgressionDescription `json:"Children,omitempty"`
	Child    *ProgressionDescription  `json:"Child,omitempty"`
	Count    *CounterDescription      `json:"Count,omitempty"`
	Detail   string                   `json:"Detail,omitempty"`
}
type CounterDescription struct {
	Kind   string `json:"Kind"`
	N      *int   `json:"N,omitempty"`
	Min    *int   `json:"Min,omitempty"`
	Max    *int   `json:"Max,omitempty"`
	Detail string `json:"Detail,omitempty"`
}

func InstalledProgression(config boardgame.PropertyCollection) MoveProgressionGroup
func DescribeProgression(group MoveProgressionGroup) ProgressionDescription
func DescribeCounter(counter ValidCounter, length int) CounterDescription
func NewCounterFromDescription(desc CounterDescription) (ValidCounter, error)
```

Package `boardgame` (`legal_plan_export.go`):

```go
func (g *GameManager) LegalPlanSpecs(moveName string) (specs []LegalSpec, declared bool, hasCustom bool, ok bool)
```

- `ok`: a move type of that name is installed (case-insensitive, like
  `ExampleMoveByName`).
- `declared`: a plan was assembled (`GameManager.legalPlans[name] != nil`).
- `specs`: deep copy of `legalPlan.specs` (contributed, minus suppressions,
  then authored). The `LegalCustom` tail is not a `LegalSpec`; it is
  `hasCustom` (`legalPlan.custom != nil`). Ledgers list it last as `custom`.

### What an exporter reads from an installed move

| Need | Source | New? |
|---|---|---|
| Progression group | `moves.InstalledProgression(move.Info().CustomConfiguration())` | Yes: the config key `configPropLegalMoveProgression` is unexported |
| Legal phases | `(*moves.Default).InstalledLegalPhases()` (already exported, returns a copy) via `move.(interface{ InstalledLegalPhases() []enum.EnumKey })` | No. Only `Default` reads that key, so a move without `Default` has no phase restriction |
| Allows multiple in progression | `move.(interfaces.AllowMultipleInProgression).AllowMultipleInProgression()` | No |
| Plan | `manager.LegalPlanSpecs(name)` | Yes |

## How the description works

`DescribeProgression` switches on the concrete type: `*defaultMoveConfig`
(every `AutoConfigurer.Config` / `NewGroupableMoveConfig` leaf) is `move`;
`serial` is `serial`; `repeat` (`Repeat`, and `Optional`, which is
`Repeat(CountBetween(0,1))`) is `repeat` with `Child`; `parallelCount` and
`*parallelCount` (`Parallel`, `ParallelCount`) are `parallel`;
`repeatFromProp` is `unknown` with a Detail naming the path (its child is
still described under `Child`); nil or any foreign `MoveProgressionGroup` is
`unknown` with a Detail. Counters are fitted at the length the group itself
passes them: 1 for repeat (`repeatSatisfied` calls `count(i, 1)`), number of
children for parallel.

`DescribeCounter` evaluates the closure on currentCount 0..64 at that length
and tries stock counters in this fixed order, accepting the first that gives
the same nil/error and the same error text at every grid point: `infinite`,
`all`, `any`, `exactly(N)`, `atLeast(N)`, `atMost(N)`, `between(Min, Max)`;
parameters 0..65 then -1 (between's Min only 0..65). A panic anywhere, or a
nil counter, is `custom`. Consequences, documented in the doc comment and
pinned by `TestDescribeCounterPreference`:

- `any` at length 1 is described as `all` (the same function there); every
  repeat counts at length 1.
- `atMost(-1)` / `between(n, -1)` are described with the literal length.
- `atLeast(N<=0)`, `atMost(N>=64)`, `between(<=0, >=64)` are `infinite`.
- A parameter above 65 is described as 65, with a Detail saying the
  description is exact only for currentCount <= 64.

The description is behaviorally exact at that length and on the grid only.

## Acceptance (real outcomes)

| Command | Outcome |
|---|---|
| `./scripts/go-local build ./...` | passes |
| `./scripts/go-local vet ./moves/ .` | clean |
| `gofmt -l` on the four Go files | no output |
| `./scripts/go-local test ./moves/...` | `ok github.com/jkomoros/boardgame/moves`; `moves/interfaces` has no tests |
| `./scripts/go-local test .` | `ok github.com/jkomoros/boardgame` |
| `./scripts/go-local run ./boardgame-util conformance export --check` | "The conformance corpus is current" |

No failures, so no comparison against `a11c3883` was needed. The full
`./scripts/go-local test ./...` was not run.

New tests (all pass):

- `moves/describe_test.go` (package `moves_test`):
  - `TestDescribeCounterStockRoundTrip`: 2763 fits. Every stock constructor
    (`CountAny`, `CountAll`, `CountInfinite`, and `CountAtLeast`,
    `CountAtMost`, `CountExactly` with 16 parameters from -3 to 100,
    `CountBetween` with all 16x16 pairs) at lengths 0..8: never `custom`,
    and the description rebuilt through the stock constructors (directly and
    after a JSON round trip) equals a fresh original on the grid, results and
    error texts.
  - `TestDescribeCounterPreference`: 13 pinned tie-breaks.
  - `TestDescribeCounterCustom`: five non-stock counters (parity; exactly-2
    with its own text; atLeast-2 with one changed text; panicking; nil) are
    `custom` with a Detail and do not rebuild.
  - `TestDescribeProgressionKinds`: exact JSON for a tree using Serial,
    Optional, ParallelCount, Parallel, Repeat; `unknown` for RepeatFromProp
    (also nested), a foreign group, nil; a custom counter inside a parallel.
  - `TestDescribeProgressionExamples`: all seven example games. Moves with a
    progression: blackjack 15, werewolf 4, checkers 2; the other four have
    none. No `unknown` node, no `custom` counter, every leaf names an
    installed move, each move's progression names that move. Each distinct
    group (blackjack 3, werewolf 1, checkers 1) was rebuilt from its
    description with the stock constructors and compared with the installed
    group's `Satisfied` on every tape of 1..5 moves over its own moves plus a
    foreign name (3905, 19607, 9330, 3905, 363 tapes); all agree.
- `legal_plan_export_test.go` (package `boardgame_test`):
  - `TestLegalPlanSpecsCallable`: every move of all seven games is `ok`;
    undeclared moves return nil specs and no custom; case-insensitive lookup
    agrees; mutating the returned specs (names, args, subs) does not change
    a second call; an unknown name and a nil manager are not `ok`. Declared
    plans: memory 4, blackjack 4, pig 3, tictactoe 1, werewolf 0,
    debuganimations 5, checkers 2.
  - `TestLegalPlanSpecsMatchLedgers`: for the six corpus games, every
    recorded ledger equals the plan's spec names and args in order, plus
    `custom` when `hasCustom`; every move in each document's move table is
    known; a move offered without a ledger has no (or an empty) plan; where
    a spec sets `Message` and the ledger records a template (non-pass
    verdicts only), they are equal. Ledgers read: memory 1376, blackjack
    6236, pig 2264, tictactoe 115, werewolf 0, debuganimations 304. Distinct
    (move, ledger) pairs compared: 2, 2, 2, 1, 0, 5. Message checks: 1000,
    691, 928, 0, 0, 128.

## Unverified

- Plans the ledgers never record. A ledger exists only for a move offered in
  a view, so fix-ups are absent. Declared plans with no recorded ledger:
  memory `Capture Cards`, `Start Hide Cards Timer`; blackjack `Start Round
  Cleanup`, `Reveal Hidden Card`; pig `Count Die`. Werewolf declares no plan
  and records no ledger. Checkers (2 declared plans) has no corpus. For these
  only callability and copying are tested, not content.
- The ledger records top-level `name` and `args` only: no compositor `Sub`,
  no `AdminPolicy`, and `message.template` only for non-pass verdicts. The
  one compositor in the examples (`legal.Any` in blackjack's `Start Round
  Cleanup`) is in an unrecorded fix-up, so `Sub` is compared nowhere.
- No example game uses `Parallel`, `ParallelCount`, `Repeat` with a count
  other than `Optional`'s, or `RepeatFromProp`. The only counter fitted from
  a real game is `CountBetween(0, 1)` at length 1. The other kinds are
  covered by the synthetic tests only.
- Counter fits are exact on currentCount 0..64; behavior above 64 is not
  checked (a parameter above 65 is reported with a Detail).
- No consumer (exporter, loader) exists yet, so the JSON shape has been
  checked only against the proposal's type, not used.

## Where the proposal and the code disagree

1. The proposal's signature is `LegalPlanSpecs(moveName) []LegalSpec`. The
   `Plan` it feeds needs `Declared` and `HasCustom`, which a slice alone
   cannot carry (an opted-in move can have zero specs), so the accessor
   returns four values.
2. `ProgressionData` has no kind for `RepeatFromProp`, which package `moves`
   builds; it is reported as `unknown` with a Detail, a kind the proposal's
   type lacks. No example uses it.
3. `CounterData`'s `custom` carries `Count: FnRef`. Package `moves` cannot
   name the closure's function, so the description carries `Detail` instead;
   the static half must supply the reference.
4. A `move` leaf's matching depends on `AllowMultipleInProgression` (one move
   or a run of them). `ProgressionData`'s leaf has only `Name`; a loader must
   take that flag from the move's `MoveData.AllowMultipleInProgression`.
5. Workstream 3's acceptance says "for every move of every example game,
   LegalPlanSpecs equals ... the recorded ledgers". The corpus cannot show
   that: checkers has no corpus and fix-up plans are never recorded (list
   above).
6. The proposal's caveat that error texts are shared between families: in
   `count.go` only `any` and `all` share texts. Every other family has its
   own, so the fit distinguishes them by text as well as by result; `any`
   and `all` are told apart by results except at length 1.
7. Every move in one `AddOrderedForPhase` call shares one group, the implied
   `Serial` of the groups passed, so each move's `Progression` is the whole
   phase's progression, always with a `serial` root.

Code behavior found while reading, not changed:

- `CountAtMost` and `CountBetween` assign a negative `max` to the closure's
  captured variable on first call, so one counter value evaluated at two
  lengths keeps the first. Groups evaluate a counter at one length, so no
  current progression is affected; a counter value shared by two
  `ParallelCount`s with different child counts would be.
- `CountAll`'s doc says it is equivalent to `CountBetween(0,-1)`. It is not:
  `CountBetween(0,-1)` accepts 0..length, `CountAll` only length, and their
  error texts differ.
- `legalPlan.specs`'s comment calls it "the serializable subset of the
  plan". It is the full assembled spec list; the opaque part is the separate
  `custom` tail.
