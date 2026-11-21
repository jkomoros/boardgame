# Proposal: factoring the web client for Common Fabric

Status: second draft, 2026-11-21, revised after four independent critiques.
Nothing here is built.
Companion: [RESEARCH.md](RESEARCH.md) holds the findings this rests on and the
design for hidden state on Common Fabric.

## Summary

Split the web client into layers with one-way imports, so the same card,
stack and animation code serves three hosts: the Go-backed webapp, a
`cf-iframe` guest inside a Common Fabric pattern, and, later, a set of trusted
web components that any pattern can use directly. The work is a refactor of
the webapp, and the webapp keeps working throughout. The Go core's game
semantics do not change; a TypeScript engine reproduces them exactly and is
held to the Go engine by fixtures the repository already has.

The work runs as four parallel streams, each with an automated oracle, and a
playable game on Common Fabric is an early deliverable of one of them, not
the end of the sequence.

## Terms

- **Pattern:** an untrusted TypeScript program on Common Fabric. Its code
  runs in a worker, never on the page's main thread.
- **Guest:** the document inside a `cf-iframe`. A full browser page that
  receives JSON from its pattern and sends events back.
- **CFC:** Contextual Flow Control, Common Fabric's information-flow model.
  Values carry labels saying who may see them.
- **Table owner:** the person who created a game's table. Never called "the
  host" here, because "host" means the software that mounts a renderer.
- **Renderer:** a game's own UI code, today a Lit class per game under
  `examples/*/client`.
- **Renderer host:** whatever mounts a renderer, feeds it state, and carries
  its moves. The webapp is one.
- **Fix-up move:** a move the engine applies by itself after a player's move,
  such as dealing or advancing the turn.

## Goals

**The ultimate goal: a trusted component set.** The boardgame elements become
web components that a Common Fabric user's policy can admit as trusted, and
that any pattern can then use from its own JSX.

- **Who it protects, and from whom.** An honest person using an untrusted
  pattern. The pattern cannot run code on the main thread; the trusted
  components can. Trusting them does nothing against another player's
  modified browser. That protection comes from CFC on the server
  ([RESEARCH.md](RESEARCH.md)).
- **What "trusted" has to mean.** A component set that:
  1. renders exactly what it is given, and cannot be dressed by a pattern to
     imitate trusted chrome;
  2. reports only what the person actually did, one event per real gesture,
     tied to what was on screen at the time;
  3. touches nothing outside its own subtree: no network, no storage, no
     reads of the surrounding page, and a written list of every global it
     uses;
  4. carries no information between children with different labels: a stack
     holding a hidden card and a visible one must not report layout or
     timing to the pattern that distinguishes them;
  5. gives the pattern no clock finer than the runtime already allows, which
     rules out raw animation-end events.
- **What it depends on.** Common Fabric has no mechanism for admitting
  third-party trusted components, and none is proposed. This goal shapes the
  design now, is built last, and has a decision gate (see "Dependencies").

**The near-term goal: a guest set.** The same elements, bundled into a
`cf-iframe` guest. The research found this feasible on today's runtime: a
whole pattern may be 4 MiB, and existing guests reach 1.6 MB.

**Constraints.**

- **One codebase.** The webapp is the first consumer of every extracted
  piece. No fork and no copied element source.
- **The Go core is the reference.** Its game semantics do not change. Tooling
  beside it may grow (see "The engine").
- **Design for CFC as ruled, not as shipped.** The design assumes the
  information-flow work that is ruled or in progress is complete, and names
  each assumption.

**Not goals.** Gathering, chat, lobbies and accounts (the pattern's own
runtime provides these). Machine translation of Go game bodies. Cryptographic
protocols between players. Agents and bots on Common Fabric.

## The layers

`server/static/src` is one tree of about 65k lines. The proposal draws three
layers in it. A file imports from its own layer or a lower one.

| Layer | Holds | Today |
|---|---|---|
| 1. **Elements** | Presentational custom elements and the motion system; the state, visibility and move-intent types they take | Card, token, die, component, stack, animator, animatable item, and the layout and status elements. The import closure of the core set is 60 files and 22.1k lines. At runtime it depends on `lit` and pure modules only |
| 2. **Renderer kit** | What a renderer is written against, and the element that mounts one | `client.ts`, the base renderer classes, move actions and drafts, effects and the effect layer, and `boardgame-render-game` |
| 3. **Hosts** | Whatever supplies state and carries moves | `boardgame-game-view`, `boardgame-game-state-manager`, the player roster, Redux, the HTTP and websocket API, routing, Material, Firebase |

Two corrections to the first draft, both from reading the import graph:

- **Layer 1 is not `lit`-only today.** The stack imports about 2k lines of
  move machinery (`moves/action`, `target-action`, `projected-choices`,
  `projected-choice-consumption`) because it invokes move actions directly.
  Those four files move into layer 1 as its move-intent part. Making the
  stack purely declarative is seam S9.
- **`boardgame-render-game` belongs in layer 2.** It reaches the server in
  three places only. With those injected, it is the reusable renderer host,
  and a new host replaces only the state queue.

**Which host loads what.**

| Host | Layers | Who writes the game's UI |
|---|---|---|
| Webapp | 1, 2, 3 | A renderer class, as now |
| Fabric guest (near term) | 1, 2, and a guest host | The same renderer class, unchanged |
| Trusted set (ultimate) | 1 only | The pattern's JSX, composing layer-1 elements |

A renderer class is arbitrary Lit code, so it can never be the trusted part.
Under the ultimate goal it does not exist, and what it does today must be
expressible to layer 1 as data. The long-run direction is therefore to **push
capability down from layer 2 into declarative layer-1 elements**, which also
makes webapp renderers thinner.

Carrying layer 2 to the guest is still worth it. The guest host is small
(about 200 lines of queue logic), and every existing renderer then runs on
Common Fabric unchanged. The declarative path is tested separately and early
(stream D), so the ultimate goal is not left as an untested assumption.

### What a renderer host provides

Written out, because the first draft understated it:

- an ordered queue of state versions, released one at a time as animations
  settle;
- the animation gate, including its watchdog timer;
- the registry and timing context that animatable items find by walking up
  their ancestors;
- a timer service;
- a mount point for player-info renderers (today only the Redux-bound roster
  mounts them);
- move and legality-preview transports;
- a loader for a game's renderer module;
- the table actions (lock room, skip turn, switch to solo).

### What a pattern author writes

Illustrative sketches, not compiled.

Near term, the pattern hands the viewer's released view to a guest and
receives moves:

```tsx
// In the game's pattern. `view` is this viewer's released seat view.
<cf-iframe src={guest} $view={view} onProposeMove={proposeMove} />
```

Ultimately, the pattern composes trusted elements directly and no renderer
class exists:

```tsx
<bg-table>
  <bg-stack layout="grid" stack={view.Game.Cards}
            face="Type" onPick={revealCard} />
  <bg-stack layout="stack" stack={view.Players[0].WonCards} face="Type" />
</bg-table>
```

Everything in the second sketch is serializable. `face="Type"` stands for
whatever replaces today's `cardView({render})` callback, and settling that
shape is the purpose of the declarative experiment in stream D.

## The seams

"Guest" means the near-term goal needs it. Sizes are small, medium, large.

| # | Seam | Today | Change | Size | Needed for |
|---|---|---|---|---|---|
| S1 | Package boundary | Clients and generated code import `../../src/client.js`, `src/types/boardgame-types.js` and `src/testing/renderer-fixture.js`. The path is hard-coded in the code generator, the stub templates, the client linter and checker, and their goldens | Packages under `server/static`, resolved first by alias with files in place. Specifiers change only after the build cache (`lib/build/static/node.go`, which copies `package.json` and runs `npm ci`) and `check_client.go` handle packages | L | Guest |
| S2 | Renderer host | `render-game` calls `movePreviewBatch`, reads location and storage through `companion-surface`, and imports renderer modules by path. The state queue is in the Redux-bound state manager | Inject those three into `render-game`. Write the host interface above. The webapp and the guest each implement the queue | L | Guest |
| S3 | State expansion | A pure function that lives in `selectors.ts` beside `reselect` | Move it to layer 1 | S | Guest |
| S4 | Server calls in the kit | `BoardgameTableViewBase` calls the HTTP API for table actions, and every generated renderer imports it | Table actions and legality preview become host-provided transports | M | Guest |
| S5 | Assets | Token art defaults to a document-relative path; art modules use `import.meta.url` | Art is passed in or resolved by the host. Fonts and theme tokens are documented inputs with fallbacks | S | Guest |
| S6 | Build and registration | Elements register as an import side effect. The source needs `experimentalDecorators` and `useDefineForClassFields: false`; compiled with other flags, Lit reactivity breaks silently. `reselect` is imported but undeclared | Ship compiled JavaScript. An explicit `define()` entry point. One Lit copy. Declare the dependency | M | Guest |
| S7 | Player-info mount | Only the roster mounts a game's player-info renderer | A mount path the host interface owns | M | Guest |
| S8 | Types | Layer-1 type files import Redux store and API types | Split wire types from store types | M | Guest (packages) |
| S9 | Animation scope | A module-global list of every stack, `document.getElementById`, and context found by ancestor walk | Scoped to one animation root element | M | Trusted set |
| S10 | Input mutation | The stack writes `ID` onto the items it is given | Never mutate input; accept frozen state | S | Trusted set; Fabric cells |
| S11 | Function-valued properties and subclassing | `cardView({render})`, action builders, game-authored subclasses of `BoardgameComponent` | A declarative equivalent for each. No subclassing in the trusted set | L | Trusted set |
| S12 | Identity across re-renders | Animation follows DOM nodes the renderer keeps alive | The animation root tracks components by id from its own measurements. A pattern's renderer recreates a child on any property change, not only a move | L | Trusted set |
| S13 | Geometry | `offsetParent` sums | Hold inside foreign shadow roots and under transformed or scrolling ancestors | M | Trusted set |
| S14 | Ambient access | A `window` test hook set at import; `getComputedStyle`, `matchMedia`, `requestAnimationFrame`, a `setTimeout` fallback | Test hooks opt in. An audited list of every global layer 1 reads | M | Trusted set |
| S15 | Events to the pattern | Layout, settle and animation-done events | Only gesture events cross to the pattern; nothing that carries timing or layout | M | Trusted set |

**Test indirection comes before any file moves.** 58 Playwright files hold
371 `'/src/...'` import strings evaluated in the page. The type checker does
not see them, so they fail only at run time. An alias that keeps those URLs
resolving is a precondition of S1.

## The engine: reproducing the core in TypeScript

**The approach.** The Go core stays the reference. A TypeScript package
implements the same semantics over the same wire formats and is required to
match Go exactly, not approximately.

**Exact, because the wire format demands it.** The first draft proposed
injected randomness and ids compared up to renaming. Both were wrong:

- **Randomness.** Go draws are not recorded; they are deterministic from the
  game id, a secret salt and the state version. The engine ports Go's
  generator (about 200 lines) and seeding, and reproduces every shuffle and
  roll from the salt in a recorded history. On Common Fabric, the only fresh
  randomness is the salt.
- **Ids.** A component id is a SHA-1 over the game id, salt, deck, index and
  a per-component counter. Under the length-only policy ids are sorted, so
  their order depends on the hash values. The engine computes the same ids,
  with a pure-JavaScript SHA-1 (about 100 lines).
- **Versions and fix-ups.** `IDsLastSeen` holds state versions, and the
  client's animation keys name fix-up moves. A ported game must produce the
  same sequence of moves and fix-ups as its Go original.

**So a ported game mirrors its Go original move for move.** Nothing
translates Go automatically, but the TypeScript definition of memory has the
same moves, in the same order, with the same names. A game written fresh for
Common Fabric has no Go original and no such obligation.

**Oracles, in the order they are used.**

| Fixture | Covers |
|---|---|
| `testdata/sanitize` | Each visibility policy's exact output |
| `legal/testdata/conformance` (32 cases; its README calls it the future Go-to-TypeScript contract) | Every legal predicate |
| Golden histories (`examples/*/testdata/golden`) | The unsanitized state, move name, proposer and phase at every version |
| Scenario replays | Per-viewer snapshots at decision points |
| A new exporter | Every viewer's view and sanitized move record at every version |

The last row is new Go code. Nothing today emits per-viewer output at every
version, and the scenario corpus is two scenarios. The exporter is additive
tooling in `boardgame-util`; it changes no game semantics. "The Go core does
not change" means the semantics, not that the repository's Go is frozen.

**Comparison is structural, after parsing,** with a named list of fields
ignored or normalized: timestamps, the Go-rendered `Diagram`, and
`ActiveTimers`, which comes from the manager and not the state.

**A game is ported** when its existing golden histories and its exported
per-viewer views match at every version.

**In scope.** Decks and stacks (growable, sized, merged) with conservation;
the five policies over groups; visibility facets; move and move-name
sanitization; the fix-up loop and its cap; phases and move progressions; the
legal plan with its base checks and message templates; projected move
choices; input binding; computed properties; game end and winners; variants.

**Out of scope.** Reflection, struct tags, generated readers, the `moves`
embedding hierarchy (helpers are functions), storage, agents.

**What the engine needs from whatever runs it.**

- The move history beside the state, because move progressions read it.
- A wake-up for timers. A timer is a deadline in state; a move is proposed
  when it passes. Common Fabric has no `setTimeout`, so the guest sends a
  tick that the engine checks against the clock.
- Who the proposer is, authenticated by the runtime.
- Seating.

**Alternatives considered.**

- **Compile the Go core to WebAssembly.** The core runs a goroutine loop with
  timers, and Go's WebAssembly runtime schedules with `setTimeout`, which the
  pattern sandbox lacks. A search also found no WebAssembly in the sandbox's
  globals. Neither point was tested by running code.
- **Keep a Go server as the authority.** It runs the core unchanged, but it
  sits outside the attested runtime, which breaks the trust model.
- **Per-game rules with no shared engine.** Cheaper to start, but the wire
  format still forces each game to match Go, so the shared part is needed
  anyway.

**The webapp uses it too,** for client-side legality and offline
development.

## Common Fabric side, in brief

One space per table. The full game state is sealed by label, so it is
delivered to no browser, and the engine runs only in server-side execution.
A browser contributes one thing, an authenticated move, and receives one
thing, its own seat's view. The design and its assumptions are in
[RESEARCH.md](RESEARCH.md).

## Work streams

Four streams run in parallel. Each has an oracle that does not need a person
to judge it.

**A. Webapp refactor.** Oracle: the existing unit, parity and Playwright
suites.

1. A checked-in table assigning every source file to a layer, and a test
   that fails on an upward import. No files move.
2. The two smallest fixes: state expansion out of `selectors.ts` (S3) and
   the test hook made opt-in.
3. Table actions and preview as transports (S4); the three injections into
   `render-game` (S2); the player-info mount (S7).
4. Test path indirection, then packages by alias (S1, S6, S8).
5. Build-cache and checker support for packages; the paired change in the
   games repository.

**B. First game on Common Fabric.** Oracle: a multi-user test that plays a
game through.

1. Memory with hand-written rules, sealed state, one released observer
   view, and a guest built with Vite from the client in place. No engine and
   no packages. This is also the missing worked example of a sealed state
   machine.
2. The guest host, replacing the state queue, with memory's unmodified
   renderer.
3. The hand-written rules replaced by the engine.
4. Blackjack, which needs per-seat views.

**C. Engine.** Oracle: the fixtures above.

1. The generator, SHA-1, stacks and id bookkeeping; the five policies
   against `testdata/sanitize`.
2. The legal predicates against the conformance corpus.
3. The Go exporter.
4. Memory against its golden histories and exported views, then blackjack.

**D. Trust and CFC.** Oracle: a yes or no on each dependency below.

1. Ask the Common Fabric team whether a third-party component set could ever
   be admitted, and what review it would face.
2. Memory's board written with layer 1 alone, as plain markup with no
   renderer class, recording every callback that had to become data.
3. The same board driven from a pattern's JSX, which is the real test: it
   crosses the worker boundary and the pattern renderer.
4. The seat spike on the list-release draft.

Streams A to C serve the near-term goal. Seams S9 to S15 wait on stream D.

## Dependencies

| Depends on | Status | Tested by | If it fails |
|---|---|---|---|
| Delivery checks labels (specs#95) | Ruled 2026-11-16; not yet implemented in labs | — | Demos state that hidden state is safe only from unmodified clients |
| Release to a list (specs#59, labs#8608) | Unruled; draft | D4 | Blackjack waits; memory is unaffected |
| A sealed state machine releases at all | Unproven; the nearest example lists open snags | B1 | Fix in labs before anything else |
| Trusted-component admission | No mechanism, none proposed | D1 | S9 to S15 land only where they improve the webapp by themselves |
| Declarative layer 1 is workable | Untested | D2, D3 | The ultimate goal is redesigned before any hardening |
| Packages in the build cache | Not supported today | A5 | Packages stay as aliases |

## How we will know

| Goal | Evidence |
|---|---|
| One codebase | The import test passes; no element source exists in two places; the webapp builds from the packages |
| Guest set | Memory's unmodified renderer plays a live game inside a `cf-iframe` guest, for two users |
| Semantics unchanged | Memory and blackjack match Go at every version; no diff to the Go game packages |
| Trusted set | Memory's board driven from pattern JSX; the list of globals; two games animating on one page |

## Open decisions

1. **The Go exporter.** It is additive tooling, but it is new Go code in a
   proposal that says the core does not change.
2. **Ported and native games.** Whether the toolkit should also serve games
   written fresh for Common Fabric, which need no Go parity.
3. **How large the trusted core is.** 22k lines is a lot to ask a person to
   trust. Dice physics and the spatial board could be a second, optional set.
4. **Names.** "Elements", "renderer kit" and "renderer host" are
   placeholders, as are the `bg-` tags in the sketch.
