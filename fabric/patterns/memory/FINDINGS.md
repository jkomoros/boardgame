# Memory on Common Fabric: findings

> **Read this first.** These findings were written against the pattern as it
> was before it moved to labs' viewer-bound release, and the labs commits
> named below. The pattern has changed since: `seatView` is now `viewFor`,
> its rule releases to the principal the view was computed for and carries
> the input witness, and the table starts as `null`.
> [`../../status/viewer-bound-port.md`](../../status/viewer-bound-port.md)
> has what ran on the newer labs branch. Against that branch:
>
> - **F1** (served handlers unattributed) and **F2** (a record its writer
>   reads and sets keeps no witness) are fixed there. The seam they left in
>   `table.tsx` is gone.
> - **F3** no longer applies: the rule adds `User(viewer)`, not a role or a
>   space.
> - **F5** is narrower: the first move still cannot be preceded by a
>   released view, but the table no longer returns nothing before it.
> - **F4**, **F6** and **F7** were not rerun.
> - The latency figures here were measured on the older labs commit. The
>   same pattern ran about five times faster on the newer one, before any
>   change to it.

What building `fabric/patterns/memory` showed about the Common Fabric runtime,
on Deno 2.9.8 and local dev servers (`start-local-dev.sh --port-offset=340`)
with server execution at its default, on. Two labs commits were used:

- **main**, `8c6f09c5fc` (2026-11-21). F1 to F6 were first met here.
- **the spike's fix branch**, `6e12b01bf7`, the tip of
  `worktree-agent-ad697e53f88bab905`, which holds `5b1c700b4a`, "Attribute a
  served handler's transaction to the handler, not to nobody". It is cut
  from an older main (`7be4b18d9b`). The three-way comparison of state shape
  and rule under "The short version" is from here.
- **the fix on current main**, `4815052651`, branch
  `fix/served-handler-write-identity` (draft PR commonfabric/labs#10111). The
  engine-backed pattern, its screenshots, F7, "The engine in the sandbox"
  and the latency figures are from here.

**The pattern depends on that fix** for the release it is designed to have.
The rule it ships today does not (see F1).

Each finding says how it is known:

- **Browser** means it was seen in Chromium against the local toolshed, with
  the piece deployed by `cf` from this repository. Most of those runs were
  driven by `tools/play-two-users.mjs` or by throwaway variants of the
  pattern that are not kept.
- **cf test** means `multi-user.test.tsx` pins it.
- **CLI** means a `cf` command showed it.
- **Read** means it comes from the cited labs file and nothing here ran it.

Line numbers are in labs at that commit.

## The short version

The idiomatic shape works as far as the pattern goes: the table is a sealed
**record**, `move` is its one writer and takes the proposer from the runtime,
and `seatView` returns a structured view in the boardgame client's wire shape.
A full game was played in two browsers under two identities.

The release rule is one condition short of the one designed. `table.tsx`
ships it in one marked place, "The policy, and the interim seam".

| | Rule as designed | Rule as shipped |
|---|---|---|
| Names the lift `seatView` | yes | yes |
| Releases to the table's space, whose readers the display then admits | yes | yes |
| Requires every confidential input to have been written by `move` | yes | **no** (F2) |

Of the findings, F2 is the one behind the seam. F7 shaped how the table
stores the engine's state. F1 is a bug with a fix
on a labs branch, which was run here and works. F3 has an idiomatic answer,
which the pattern uses. F4 and F5 shaped the code without weakening it.

What was run on the fix branch, each a fresh piece played to the end by two
browser identities (**Browser**):

| State | Rule | Result |
|---|---|---|
| record | lift + witness, adds `Space` | nothing released (F2) |
| one JSON string | lift + witness, adds `Space` | **whole game played**: the witnessed release fires on served runs, at the display boundary, for a 32 KB view of nested records, lists and maps |
| record | lift, adds `Space` (**shipped**) | whole game played |

## Where the runtime refused the idiomatic shape

### F1. On labs main, a witnessed release never fires when the handler runs on the server

- **Attempted.** The rule in `table.tsx` under "AS DESIGNED": `TransformedBy`
  naming `seatView`, with an `inputWitness` naming `move`.
- **Minimal repro.** Any pattern whose rule has an `inputWitness`, deployed to
  a toolshed with server execution on (the default), with the handler sent an
  event by any route (`cf piece call`, or a click). Open the piece as a
  principal with access to the space. Labs' own
  `patterns/examples/cfc-exchange-rules/witnessed-chain.test.tsx` shows the
  same at a write gate when run with the flag on (reported by the earlier
  spike; not rerun here).
- **Expected.** The view is released. Every confidential location `seatView`
  read was written by `move`; that is what
  `docs/specs/cfc-transformed-by-input-witnesses.md` says the witness means.
- **Actual (Browser).** Every field of the view renders as "Content hidden by
  policy". The same pattern on servers started with
  `EXPERIMENTAL_SERVER_EXECUTION=false` releases it (with the state shape of
  F2). The four cases, each a fresh piece, one `sit` sent by `cf piece call`,
  then opened in a browser as a second identity granted `WRITE`. Each piece
  showed three fields of the view in plain JSX (`view.phase`,
  `view.revision`, `view.seats[0].label`), with no `cf-iframe`:

  | State | Rule | Server execution on | off |
  |---|---|---|---|
  | record | lift + witness + role | hidden | hidden |
  | one JSON string | lift + witness + role | hidden | **released** |
  | record | lift + role | released | not run |
  | record | lift only, `dropClause` | released | not run |

- **Where.** The refusal is the display gate's:
  `runtime/html/src/worker/reconciler.ts:2169` reports
  `render-confidentiality-ceiling`. The default is
  `runtime/memory/v2/server-execution-default.ts:46`. CI runs the pattern
  tests with it off only: `tasks/integration.ts:71` and `:797`. The cause was
  not looked for here.
- **Fixed on the fix branch (Browser).** With `5b1c700b4a`, the string-state
  pattern with the witnessed rule played a whole two-user game on served
  runs. The labs spike's report names the cause: the serving loop's
  processed-claim write is counted as the transaction's first authored write,
  so the handler's transaction has two writers and is attributed to neither.
  The spike had shown the fix at a write gate only; this is the same fix at
  the display boundary, in a browser.
- **Seam.** None for this finding. The pattern assumes the fix.

### F2. A record that its writer reads and then sets keeps no witness

- **Attempted.** The state as a record,
  `WriteAuthorizedBy<Sealed<TableState<MemoryGame>>, typeof move>`, which
  `move` reads with `table.get()` and replaces with `table.set(next)`.
- **Minimal repro.** The first two rows of the table in F1, with server
  execution off: the same pattern, the same rule, the state as a record or as
  its JSON form in one string.
- **Expected.** For the design, a state machine's writer keeps its witness: a
  writer that reads only what it wrote itself has added nothing foreign.
- **Actual (Browser).** Hidden for the record, released for the string.
- **Where (Read).** This is the specification as written, not a bug in the
  code: a `set` destination is stamped only when "the transaction created the
  destination, or read no content at or beneath it"
  (`docs/specs/cfc-transformed-by-input-witnesses.md:189`), and "a container a
  `Default` put in place keeps no writer at all" (`:169`). A state machine
  reads its state by definition.
- **What would fix it.** A ruling that a destination keeps its stamp when
  everything the transaction read at or beneath it already carried the same
  writer's stamp. Without that, every sealed state machine has to store its
  state as one opaque scalar, which is the contortion this pattern was asked
  not to make.
- **Still so on the fix branch (Browser).** Record state with the witnessed
  rule released nothing; the same rule over a string state played a game.
- **Seam.** The shipped rule has no `inputWitness`. This is the whole of the
  interim seam.
- **Not held today.** The policy releases `seatView`'s view of whatever is in
  the table's cell, not only of what `move` wrote. `WriteAuthorizedBy` on the
  cell is still enforced (see "What a table protects"), so the exposure is
  where that check does not reach: a modified client with write access to the
  space, and a second instance of this module pointed at this table's cell.

### F3. A rule that asks for a role releases each scalar and refuses the whole view

- **Attempted.** The second condition of the rule as designed:
  `cfcPattern.hasRole(v("reader"), THIS_POLICY.subject, "reader")` with
  `post: { addAlternatives: [cfcPattern.user(v("reader"))] }`, and the view
  handed whole to the guest (`$context={{ view }}`), which is the only way a
  `cf-iframe` guest can take a structured value.
- **Minimal repro.** A lift returning `{ a: 1, list: [{ b: 2 }] }` over a
  `Confidential` input, with a rule of `TransformedBy` plus `hasRole`. In JSX,
  `{view.a}` and `{view.list[0].b}` render. In a `cf-iframe`,
  `fabric.cell("view").pull()` is refused.
- **Expected.** A released object is released with its fields.
  `patterns/examples/cfc-exchange-rules/README.md` says of `blessed-object.tsx`
  that the "object node is released along with its fields"; that example's
  rule has no role in it.
- **Actual (Browser).** With the role: `{view?.phase}`, `{view?.revision}` and
  `{view?.seats[0]?.label}` render in JSX, and the guest's read of the whole
  view is refused, before the deal, when the view is seven scalars, one list
  of one small object and an empty list. Without the role (`dropClause`): the
  guest reads the whole view, 32 KB of nested records, lists and maps.
- **Where (Read).** A rule with a role in it is not value-intrinsic
  (`runtime/runner/src/cfc/exchange-eval.ts:233`), so it is evaluated over the
  join of everything a read consumed rather than at each location. The `HasRole`
  fact is minted for a display in `runtime/runner/src/cfc/render-ceiling.ts:68`
  to `:78`. Which location of the join failed was not identified; the gate's
  report said only `integrity: []` (see F6).
- **The idiomatic answer (Browser, fix branch).** Release in two steps. The
  module's rule is guarded by `TransformedBy` alone and adds
  `Space(THIS_POLICY.subject)` to the clause; the runtime's standard display
  rule then admits a reader of that space. Under it the guest reads the whole
  view in one read, lists and lists of records included, which the labs
  spike had left unrun. `cfcPattern` has no `space()` constructor, so the
  atom is written raw:
  `{ type: "https://commonfabric.org/cfc/atom/Space", id: THIS_POLICY.subject }`.
- **Seam.** None. The shipped rule is the two-step form.
- **Not held.** Nothing by this finding. Two things to weigh: `Space(table)`
  is a wider release than `User(reader)`, since a value so labeled also fits
  any store whose ceiling admits the table's space; and nothing warns an
  author that putting a role in the same rule changes where it is evaluated.

### F4. A projection that returns part of its input unchanged is withheld whole

- **Attempted.** `viewOf` built each entry as
  `{ ..., timers: recorded.timers, ... }`, passing the table's own `timers`
  record through.
- **Minimal repro.** A blessed lift `(input) => ({ a: input.count, inner:
  input.inner })` where `inner` is an object. Read the result whole.
- **Expected.** Either the value is copied, since a lift returns data, or the
  refusal names the path.
- **Actual (Browser).** After the deal the guest's read of the whole view was
  refused under the lift-only rule that had released it before the deal. A
  probe added to the guest (`__boardgameFabricProbe`) read every sub-path:
  all were handed over except `entries/0/timers`, which was `{}`.
  Copying the record in `viewOf` fixed it.
- **Where (Read).** The returned object reaches the store as a reference to
  where it already lives, in the sealed table, and a reference "stored any way
  other than by anchoring ... earns its slot no value stamp"
  (`docs/specs/cfc-transformed-by-input-witnesses.md:337`). So the runtime is
  doing what the specification says. The cost is that one aliased object
  anywhere in a projection withholds all of it, and nothing says which.
- **Seam.** None. `viewOf` returns only new objects and says why; a game's
  `sanitize` and `viewerPart` must do the same.
- **Not held.** Nothing.

### F5. Nothing can be released before the first move, so a guest starts blind

- **Attempted.** Draw the lobby from the view from the first paint.
- **Actual (Browser).** On a fresh table the guest's first read of `view` is
  refused, under the shipped rule as under the others, on main and on the
  fix branch. `seatView`
  returns `undefined` until the table's first accepted proposal, by design:
  under the rule as designed, a view of the default state carries no witness,
  and a field the first move leaves unchanged keeps that unwitnessed stamp
  (`docs/specs/cfc-transformed-by-input-witnesses.md:417`). Why an `undefined`
  result is refused under a rule with no witness was not looked for.
- **Consequence.** The guest must let a person sit down without having seen
  the table. `fabric-bridge.ts` draws the lobby on a refusal, enables "Sit
  down", and shows that the view was withheld. After the first `sit` the view
  arrives.
- **Seam.** None.

### F6. A refusal at the display boundary gives a pattern author nothing to go on

Not a refusal of the shape, but it cost more time than any of them.

- The page shows "Content hidden by policy" and the console one line,
  `render-confidentiality-ceiling`, once per page load.
- With the rule as designed and the view in a `cf-iframe`'s `$context`, the
  **whole piece** was replaced, its name in the shell's header included, not
  only the frame holding the view. With the same rule and the view's fields
  in plain JSX, the piece rendered and each field was replaced. With the
  role but no witness, the piece rendered and only the guest's read was
  refused. The difference was not explained.
- The reasons are logged at debug level in the runtime worker
  (`runtime/runner/src/cfc/denial-report.ts`). They can be had with
  `commonfabric.forwardWorkerConsole(true)` and
  `await commonfabric.rt.setLoggerLevel("debug", "cfc")`, then causing a
  re-render; `tools/play-two-users.mjs` does this under `CFC_DEBUG=1`. They
  name the clause, the ceiling and the integrity found (`[]`), and no path.
- `cf test` has no display boundary: `HasRole` is minted only for a render
  (`runtime/runner/src/cfc/render-ceiling.ts`), and a test reads `table.view`
  whatever its label. So `multi-user.test.tsx` passes identically under the
  rule as designed and the rule as shipped, and none of F1 to F5 can be
  pinned by a test in this repository. Each needed a browser.

**Asked for:** the denial's inputs should name the consumed path whose label
or integrity decided it, and `cf test` should be able to read a cell the way
a display would.

### F7. A cell position typed `unknown` reads back as an empty object

- **Attempted.** Store the engine's own state in the sealed cell under the
  engine's own types: `StateJSON`, whose sub-states are
  `Record<string, unknown>`, and `MoveRecord`, whose `Blob` is too.
- **Minimal repro.** A cell of type `{ blob: Record<string, unknown>; any:
  unknown; list: unknown[] }`. A handler sets
  `{ blob: { n: 3, deep: { k: [1] } }, any: { x: 1 }, list: [1, "two"] }`. A
  lift over the cell returns `JSON.stringify` of what it is given.
- **Expected.** The data, or a refusal at compile time.
- **Actual (cf test).** `{"blob":{"n":{},"deep":{}},"any":{},"list":[{},{}]}`.
  The write is accepted and every `unknown` position reads back as `{}`,
  with no warning.
- **Where (Read).** It is what the schema generator says `unknown` means: "a
  field that holds a reference to another piece. A reader of such a field
  gets an opaque reference carrying no properties rather than the value"
  (the compiler's own message, when a pattern's root output declares one).
  That message is given only for a pattern's reserved root keys. Nested, it
  is silent.
- **What works.** A recursive JSON type,
  `type Json = null | boolean | number | string | Json[] | { [key: string]: Json }`,
  round-trips everything (`toolkit/json.ts`). The table stores engine data
  under it and converts at the boundary. Two things to know: a value read
  back has its **object keys in a different order** (sorted), so nothing may
  depend on key order, and `parity.test.ts` replays every recording through
  such a copy for that reason.
- **Seam.** None.
- **Asked for:** the nested case should be a compile error or a warning. As
  it is, a pattern that stores any loosely typed JSON loses it silently.

## The engine in the sandbox

The pattern imports the engine by relative path
(`../../../server/static/src/engine/...`) and deploys with
`cf ... --root <repository root>`. The whole engine now loads and runs in a
pattern: the core, the rules layer, the legal evaluator and all six ported
games (`cf check` on a pattern importing each; Memory and blackjack also run
whole recorded games inside a lift, `parity.test.tsx`).

**What the sandbox refuses, and why.** It is deliberate. A pattern module's
top level is verified statement by statement before the module may run, and
only four things are admitted there: a trusted builder call, a direct
function, data the runtime can freeze, and type and import declarations
(labs `docs/specs/sandboxing/SES_SANDBOXING_SPEC.md`, section 4.2). The goal
is that no invocation of pattern code can reach state another invocation
left behind (section 1, "Invocation Isolation"; section 7.2.1). A class is
named outright: "Authored top-level class declarations are rejected in v1.
They are neither plain data nor trusted builders, and they create mutable
prototype/static state that violates the invocation-isolation goal."

Each form was probed with one small pattern (**cf check**):

| At module scope | Result | Message, and where it comes from in labs |
|---|---|---|
| `class C {}` | refused | `Top-level class declarations are not allowed in SES mode` (`runtime/runner/src/sandbox/compiled-bundle-verifier.ts:210`) |
| `let n = 0` | refused | `` `let` declarations are not allowed at module scope `` (the transformer) |
| `const T = { double: x => x * 2 }` | refused | `PlainDataValidationError: <root>.double: Unsupported value type 'function'` (`runtime/runner/src/sandbox/plain-data.ts:237`) |
| `const f = make()` where `make` returns a function | refused | the same, at `<root>` |
| `const D = make()` where `make` returns data | accepted | |
| `for (...) TABLE[k] = v` | refused | `Compiled module contains unsupported top-level executable code` |
| `const N = DIM * DIM` | refused | `Top-level value is not allowed in SES mode` (`:805`) |
| `const R = { size: DIM }`, `const A = B` | accepted | |
| `function* gen() {}` | refused | `Compiled module contains unsupported top-level executable code` |
| `const K = Symbol("k")` | refused | `Only trusted builder calls, schema(), canonical function hardening, and canonical binding annotation are allowed at module scope` (`:674`) |
| `const T = new Uint32Array(4)` | refused | `Mutable top-level data must be wrapped in __cf_data()` (`:690`) |
| `const M = new Map([...])`, a `bigint` literal | accepted | |
| `x as unknown as T`, anywhere | refused | `Double-casting via 'as unknown as' is not allowed` (`base/ts-transformers/src/transformers/cast-validation.ts:85`) |

Inside a function, none of this applies: a class declared in a function, an
object whose methods use `this`, getters and closures over local state were
all accepted. `Proxy` does not exist. The pattern compiler also type-checks
with `exactOptionalPropertyTypes` (`base/js-compiler/typescript/options.ts:21`),
which this repository's `tsconfig.json` does not set.

**What was chosen.** The engine is rewritten to the rule, in place: one
engine, no copy.

- Its ten classes are interfaces with factory functions returning an object
  that carries its own methods (`newStack`, `loadEngineState`, `newGoRand`,
  `newRulesState`, ...). Calls read the same. `instanceof` became a test
  function over a field.
- Its six `Error` subclasses are ordinary `Error`s with a `name`, a maker
  and a test (`stackError(m)`, `isStackError(e)`).
- A game and each of its moves is a function returning its spec, since a
  spec holds functions.
- One table of functions became a `switch`; three double casts went; two
  module-scope loops and one computed constant became functions and a
  literal; five optional properties were typed precisely.
- The legal-message renderer moved into the engine, because the client's
  module it imported keeps a stateful regular expression at module scope and
  imports a declaration file by an extensionless path, neither of which a
  pattern admits. A test holds the two renderers to the same text.

Nothing behaves differently: 1,168 of 1,168 unit tests pass, with all 22
conformance documents. `engine/sandbox.test.ts` enforces the shape by
reading the sources, and `engine/tsconfig.sandbox.json` type-checks under
the extra option; neither needs labs.

**What was rejected.**

- *Wrapping each class in a function* (`function stackClass() { return class
  ... }`). The sandbox accepts it, but with no module-scope place to keep
  the class, every call makes a new one: `instanceof` stops meaning anything
  and the cost is paid per object. It keeps the word `class` and loses what
  a class is for.
- *Keeping classes and asking labs to allow them.* The refusal is a stated
  design decision with a stated reason, not a gap.
- *A second, class-free surface beside the classes.* Two engines.
- *Bundling the engine to one script and loading it as data.* String
  encoding; the sandbox would evaluate nothing of it.

**The cost.** TypeScript's `private` is gone: an interface lists every
member. And a game compiles on every run that needs it, since nothing may be
kept at module scope; compiling Memory takes under a millisecond under Node.

## The guest bundle

`fabric/patterns/guest/guest.js`, built by `npm run build:fabric` in
`fabric/guest`: **1,084,888 bytes** (289 KB gzipped), plus a 211-byte
`guest.html`. A server-execution upload takes at most 4 MiB
(`docs/common/ai/iframe-pattern-guide.md`, "Images and other assets"), so the
bundle is **26%** of the limit. `cf piece new` uploaded it in about two
seconds.

- The standalone guest is 589,520 bytes. The other 495 KB is Common Fabric's
  own guest client (`base/iframe-sandbox/src/guest.ts` and what it imports, 143
  modules and 402 KB before minifying).
- It is larger than it need be. The client awaits at module top level, so the
  bundle has to be an ES module, and Vite does not strip whitespace from a
  library built as one. A second minifying pass was not added.
- The client is bundled out of a labs checkout by
  `tools/vendor-fabric-client.sh` and is not in this repository. A guest built
  against one labs commit may not match a runtime at another.

`cf-iframe`'s content policy (`base/iframe-sandbox/src/csp.ts`) allows inline
script and style, so the guest ran without change. One violation is reported
on every load, `script-src blocked wasm-eval`: the client tries a WebAssembly
SHA-256 and falls back.

## Latency

From `tools/play-two-users.mjs`: two headless Chromium contexts and the
toolshed on one machine, nothing else running, one game of 16 reveals each
(**Browser**). From a proposal leaving the mover's guest to the new view
reaching a guest:

| Rules | Labs | State a version | View | To the mover, median (min to max) | To the other player, median |
|---|---|---|---|---|---|
| Hand-written, lean state | main | 3 KB | 32 KB | 2,482 ms (985 to 3,277) | 2,406 ms |
| Hand-written, lean state | spike's fix branch | 3 KB | 32 KB | 1,760 ms (1,083 to 2,173) | 1,755 ms |
| Engine, two inputs held | fix on main | 19 KB | 106 KB | 5,930 ms (3,131 to 7,666) | 5,936 ms |
| **Engine, one input held (shipped)** | fix on main | 19 KB | 71 KB | **3,684 ms** (2,595 to 6,144) | 3,673 ms |

`send()` resolved (the handler's run committed) after a median of 1,897 ms
in the shipped form.

- **Going onto the engine roughly doubled the time a move takes**, from 1.8 s
  to 3.7 s, and it is the size of the data, not the engine's work. Under
  Node the engine applies a reveal in 0.7 ms and computes a whole view in
  16 ms. What grew is the state: Go's Memory keeps the 110 cards not in play
  in an `UnusedCards` stack, with an id and a last-seen entry for each, in
  every version. The hand-written rules had left that stack empty.
- **The runtime's cost follows the size of what is stored and released.**
  Holding two inputs' versions instead of one added 2.2 s a move. The same
  shows in `cf test`: the multi-user test took 76 s on the hand-written
  rules, 266 s on the engine with a six-version window, and 166 s as
  shipped; it logs hundreds of `get() took >50ms` warnings on the table and
  the view.
- **So a table holds one input's versions** (`INPUTS_HELD` in
  `toolkit/table.ts`), keyed by version so a held version keeps its place in
  the store. A guest that receives every value of the view still animates
  every step. One that misses a value jumps.
- The hide-cards timer is four seconds. The cards were hidden 6.9 to 8.7 s
  after the second card showed: the guest's tick is sent when four seconds
  have passed on its clock, and then travels like any move.

This is too slow for a card game to feel right. What is left to try is on
the release's side: a view whose versions are separate values, so that a
move adds one and rewrites none, and telling the guest only what changed.
Neither can be done from the pattern alone: the guest's bridge reads one
value.

## Did a move run speculatively in the browser?

Nothing a player could see suggested it (**Browser**).

- The mover's own guest received the new view no sooner than the other
  player's did (medians 3,684 ms and 3,673 ms as shipped; the same held in
  every earlier configuration). A speculative run whose
  result reached the view would have shown first for the mover.
- The dealer's bridge logs every distinct view it is given. In each of eleven
  recorded games it saw exactly one deal (one epoch at version 0). A local
  run of `move` would have drawn its own salt, so its deal would have had a
  different epoch.

So with server execution on, either the client did not run `move`, or what it
wrote never reached the released view. Which was not determined. The guest
is ready for the other case: the view's `epoch` is derived from the game id,
so a speculative deal and the server's are different epochs, and the guest
cuts to the newer one instead of mixing their card ids.

## What a table protects, as observed

Filled in from `fabric/RESEARCH.md`, "What a table protects, and from whom".

| Property | Observed | How |
|---|---|---|
| A player cannot see a hidden card's face in what the pattern shows them | Holds for an unmodified client. The view never holds a hidden face | cf test (every step, every entry of the window); `table.test.ts` |
| A person with access to the space cannot read the sealed table | **Does not hold.** A key granted only `READ` read `table/secretSalt` and the whole layout with `cf cell get --input` | CLI. The specs#95 non-conformance RESEARCH.md expects |
| A player cannot move as someone else | Holds. A payload naming another player as `proposer`, `acting` and `seat` changed nothing; each seat in the view holds the browser's own DID, which the script never sent | cf test; Browser |
| A person with no seat cannot move | Holds. `cf piece call ... move` as an unseated key left the revision unchanged | cf test; CLI |
| A player cannot make an illegal move | Holds for the run that counts: out of turn, off the board, a card already up, a fix-up move, a hide too soon | cf test; `table.test.ts` |
| A timer cannot be hurried | Holds. A tick before the deadline changes nothing; the table reads its own clock | `table.test.ts`; Browser (ticks arrived and were applied only after 4 s) |
| A player cannot forge the table | Refused through the stock client: `cf cell set --input table/phase` as a `WRITE` key fails with `writeAuthorizedBy requires a trusted verified binding identity at /table`. A modified client was not tried | CLI |
| A player cannot forge the view | A `cf cell set view/phase` as a `WRITE` key failed and the value did not change. A modified client was not tried | CLI |
| The release holds only for a state `move` wrote | **Not held**: F2 | Browser |
| The release reaches only the space's readers | Held by the label as far as was seen: the view carries `Space(table)` and the display admits its readers. A principal outside the space cannot load the piece at all, so the label was not what stopped them | Browser |
| The table owner cannot see hands | Not examined. The owner is a reader, so the second row applies | |

## Smaller things

- **The guest's mount could be cancelled by its first state.** A `state` with
  an epoch other than the one on screen resets playback, and that reset
  shared a counter with the mount that `init` starts, so a first state with
  a non-zero epoch arriving before the renderer mounted left the guest with
  no board. The harness always used epoch 0. Fixed in `guest-host.ts`.
- **The animation counters are gone.** The client's test hook is now opt-in,
  so the script's `motion` reads `{}`. It still samples running animations
  in the mover's guest: up to ten on every move of the last game.
- **`identify` is sent on load, not from a gesture.** A lift cannot ask who
  is looking, so the pattern has a second handler that writes
  `currentPrincipal()` into the viewer's own `PerUser` cell, and the guest
  finds its seat by comparing. It is the one event the guest sends unasked.
- **A refused proposal is silent.** `move` writes nothing and does not throw,
  so `send()` resolves either way. The bridge pulls the view after each
  send and tells the renderer the move failed if the table did not advance.
- **`info` is a `computed()` of a constant** so that it is a cell `$context`
  will take. The shell logs `EMPTY SCHEMA SUBSCRIPTION` for it on every load.
- **Each run of the script needs a fresh piece.** A browser identity is new
  each time, and a finished table has seats held by the last run's.
