# Proposal: factoring the web client for Common Fabric

Status: draft for critique, 2026-11-21. Nothing here is built.
Companion: [RESEARCH.md](RESEARCH.md) holds the findings this rests on and the
CFC design for hidden state (sealed state, seats as audience lists).

## Summary

Split the web client into layers with one-way imports, so the same card,
stack and animation code serves three hosts: the Go-backed webapp, a
`cf-iframe` guest inside a Common Fabric pattern, and, later, a set of trusted
web components that any pattern can use directly. Every step is a refactor of
the webapp that leaves it working. The Go core is not changed; a TypeScript
engine interprets the same semantics and is held to the Go engine by the
scenario replays the repository already records.

## Goals

**The ultimate goal: a trusted component set.** The boardgame elements become
a set of web components that a Common Fabric user's policy can admit as
trusted, and that any pattern can then use from its own JSX. A pattern is
untrusted code running in a worker; it can send attributes, serializable
properties and children to the main thread and nothing else. The components
are the trusted part, and what a user trusts them for is narrow:

1. They render exactly what they are given.
2. They report only what the person actually did.
3. They touch nothing outside their own subtree: no network, no storage, no
   reads of the surrounding page.

Common Fabric cannot admit third-party trusted components today. This goal
shapes the design now and is built last.

**The near-term goal: a guest set.** The same elements, bundled into a
`cf-iframe` guest. A guest is a full browser document that receives JSON state
from its pattern and sends events back. Nothing in Common Fabric needs to
change for this.

**Two constraints.**

- **One codebase.** The webapp is the first consumer of every extracted
  piece. No fork, no vendored copy, no second implementation of an element.
- **The Go core is the reference and does not change.** Fabric has no Go
  runtime. The game semantics are interpreted there by a TypeScript engine
  that reads and writes the same wire formats.

**Not goals.** Gathering, chat, lobbies, seats and accounts (the host provides
these). Porting every example game. Cryptographic protocols between players.
Changing how Go games are authored.

## The layers

Today `server/static/src` is one tree. The proposal draws three layers in it.
A file imports from its own layer or a lower one.

| Layer | Holds | May depend on | Today |
|---|---|---|---|
| 1. **Elements** | Presentational custom elements and the motion system: card, token, die, component, stack, animator, animatable item, fading text, and the layout and status elements; `motion/`, `solid/`, the state and visibility types | `lit` only | The "leaf set": about 51 files, 20.7k lines, already free of Redux, routing and `fetch` |
| 2. **Renderer kit** | What a game's renderer is written against: `client.ts`, the base renderer classes, move actions and drafts, effects, and a host interface | Layer 1 | Mixed into `components/`, `moves/`, `effects/`; reaches the server through `api.js` in a few places |
| 3. **Hosts** | Whatever supplies state and carries moves | Layers 1 and 2 | The webapp: Redux, the HTTP and websocket API, routing, Material, Firebase |

A game's own renderer (`examples/*/client`) sits on layer 2, as now.

**Which layer serves which goal.**

| Host | Layers it loads | Who writes the game's UI |
|---|---|---|
| Webapp | 1, 2, 3 | A renderer class, as now |
| Fabric guest (near term) | 1, 2, and a small guest host | The same renderer class, unchanged |
| Trusted set (ultimate) | 1 only | The pattern's JSX, composing layer-1 elements |

The last row is the one that drives the design. A renderer class is a Lit
component with arbitrary code, so it can never be the trusted part; under the
ultimate goal it does not exist, and everything it does today must be
expressible to layer 1 as data. So the long-run direction of the refactor is
to **push capability down from layer 2 into declarative layer-1 elements**.
That also makes webapp renderers thinner, which is the sense in which the
webapp "gets more and more reusable".

## The seams

Each is a specific change, listed with the goal that needs it. "Guest" means
the near-term goal cannot ship without it.

| # | Seam | Today | Change | Needed for |
|---|---|---|---|---|
| S1 | Package boundary | Game clients and generated code import `../../src/client.js` | Layers 1 and 2 become packages; `boardgame-util` emits package specifiers | Guest |
| S2 | Renderer host | `boardgame-render-game` owns the animation loop and also the server API, companion sync and dynamic imports | Extract the loop (prepare, set state, animate, release the next state) behind a host interface; the webapp, the test fixture and the guest each implement it | Guest |
| S3 | State expansion | In `selectors.ts`, behind `reselect` | A pure function in layer 1 | Guest |
| S4 | Server-only helpers | Legality preview and projected choices call the server | Already transport-shaped (`MoveTransport`, `MovePreviewTransport`); finish the job so layer 2 never imports `api.js` | Guest |
| S5 | Assets | Token art defaults to a document-relative path | Art is always passed in, or resolved by a host-supplied function | Guest |
| S6 | Element registration | Defined as a side effect of import | An explicit `define()` entry point, one Lit copy | Guest (bundling), trusted set |
| S7 | Animation scope | A module-global list of every stack, and `document.getElementById` | Scoped to one animation root element | Trusted set |
| S8 | Input mutation | The stack writes `ID` onto the items it is given | Never mutate input; accept frozen state | Trusted set; also Fabric cells |
| S9 | Function-valued properties | `cardView({render})`, action builders | A declarative equivalent for each: data, or slotted children | Trusted set |
| S10 | Identity across re-renders | Animation follows DOM nodes the renderer keeps alive | The animation root tracks components by id from its own measurements, so a host that recreates nodes still animates | Trusted set |
| S11 | Geometry | `offsetParent` sums | Hold inside foreign shadow roots, transformed and scrolling ancestors | Trusted set |
| S12 | Ambient access | A `window` test hook; reduced-motion via `matchMedia` | Test hooks opt in; an audited, written list of every global layer 1 reads | Trusted set |

`testing/renderer-fixture.ts` already mounts a renderer from a JSON snapshot
with fake transports, and `examples/dice-keep-demo.ts` already drives the
animator with no server. S2 generalizes what those two do.

## The engine: interpreting the core in TypeScript

**What "interpret" means here.** The Go core stays as it is and stays the
reference. A TypeScript package implements the same semantics over the same
wire formats: the state JSON, the per-viewer sanitized JSON
(`Indexes`, `IDs`, `IDsLastSeen`), and move records. It is a second
implementation of the model, not a translation of Go source.

- **Scope.** Decks and stacks with the one-slot-per-component invariant; the
  five visibility policies over groups; moves as `legal` plus `apply`; the
  fix-up loop; phases; the legal-predicate vocabulary with its facet tags;
  per-viewer sanitization.
- **Left out.** Reflection, struct tags, generated readers, the `moves`
  embedding hierarchy, storage, agents.
- **Held to Go by replay.** `scenario.Run` already records, for a scripted
  game, the authoritative `History` and per-viewer `Replay` snapshots. The
  conformance suite feeds each recorded move to the TypeScript engine and
  compares its per-viewer output with the recording. A game is "ported" when
  its scenarios pass.
- **Randomness is injected.** The engine takes its draws from the caller. In
  conformance runs they come from the recording; on Fabric, from the move
  handler. The engine does not reproduce Go's generator.
- **Ids are opaque, not hashed.** They keep the shape and the re-roll
  behavior of the Go ids, not their values, so conformance compares ids up to
  a consistent renaming.

**Game definitions.** A Go game's `Apply` and `LegalCustom` bodies are Go, and
nothing here translates them. A Fabric game is authored in TypeScript against
the engine. Much of a Go game is already data (decks, state shape, policies,
stock moves configured by `auto.MustConfig`, legal preconditions, phases), so
a later step can have `boardgame-util` emit that part as a manifest and leave
only the custom bodies to port by hand. That step is optional and not
promised here.

**Why not compile the Go core to WebAssembly.** The engine must run where
Fabric trusts it, which is the pattern's move handler, not the guest. I found
no WebAssembly in the pattern sandbox's globals. The check was a search, not
a test, so it is worth confirming before this is treated as closed.

**The webapp uses it too.** The same package can give the webapp client-side
legality, and an engine for offline development, without a second
implementation.

## Sequence

Each step lands as its own change with the webapp's tests passing.

1. **Draw the boundary in place.** Assign every file under `src/` to a layer
   in a checked-in table and add a test that fails on an upward import. List
   the existing violations and burn them down (S3, S4, S5). No files move, so
   this is cheap to revise during critique.
2. **Prove the guest early.** A throwaway `cf-iframe` guest that bundles
   layers 1 and 2 as they are and animates a recorded memory replay. It
   surfaces bundling, asset and bridge problems before the refactor commits
   to a shape.
3. **Make the packages.** Layers 1 and 2 become workspace packages; the
   webapp and the code generator import them by name (S1, S6).
4. **Extract the renderer host** (S2). The webapp's `render-game` becomes one
   implementation; the test fixture becomes another.
5. **The guest host.** A third implementation, in `fabric/`, with memory's
   existing renderer running unchanged inside it against recorded state.
6. **The engine**, with memory's scenarios as its first conformance suite,
   then blackjack's.
7. **Memory on Fabric end to end:** engine in the move handler, sealed state,
   one released view, the guest. Then blackjack, which needs per-seat release.
8. **Trusted-set hardening** (S7 to S12), each landing as a webapp
   improvement.
9. **The declarative proof.** A plain HTML page that plays a recorded game
   using layer 1 alone: no renderer class, no script beyond feeding
   attributes. When that page works, a pattern's JSX can do the same.
10. **Propose admission** of the set to Common Fabric.

Steps 1 to 5 serve the near-term goal and need no engine. Steps 6 and 7 can
proceed beside them. Steps 8 to 10 serve the ultimate goal.

## How we will know

| Goal | Evidence |
|---|---|
| One codebase | The webapp imports layers 1 and 2 as packages; the import test passes; no element source exists in two places |
| Guest set | Memory's unmodified renderer plays a recorded game inside a `cf-iframe` guest, with the same animation fixtures the webapp uses |
| Core unchanged | No diff under the Go packages; memory and blackjack scenarios pass against the TypeScript engine |
| Trusted set | The declarative page (step 9); a written list of every global layer 1 touches; two games animating on one page |

## Open for critique

1. **Is layer 2 worth carrying to Fabric?** The guest could skip renderer
   classes and go straight to declarative layer 1. That is more work before
   the first demo, but it avoids building a guest host that the ultimate goal
   discards.
2. **How large is the trusted core?** 20.7k lines is a lot to ask a user to
   trust. The dice physics and the spatial board could be a second, optional
   set.
3. **Where the packages live.** `packages/` at the repository root, or under
   `server/static`. The first is clearer; the second moves less.
4. **The separate games repository.** Generated clients there import the same
   paths, so S1 is a paired change.
5. **Distribution.** The guest bundle can be built in this repository, so
   nothing needs publishing to npm yet. The trusted set eventually needs a
   published, content-addressed artifact.
6. **Conformance up to id renaming and injected draws.** Is that strict
   enough to call the engine faithful?
7. **Names.** "Elements", "renderer kit" and "renderer host" are
   placeholders. `boardgame-game-surface` is already a layout element, so the
   animation root needs a different name.
