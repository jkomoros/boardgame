# Boardgame on Common Fabric: research and CFC design

Status: second draft, 2026-11-21. Nothing here is built yet. Findings come
from reading code, specifications and draft rulings; almost nothing was run.
Companion: [PROPOSAL.md](PROPOSAL.md) is the proposal for factoring the web
client and for the TypeScript engine. This document holds the research behind
it and the design for hidden state on Common Fabric.

## Goal

Make boardgame's card, stack and animation UX usable from Common Fabric
patterns, with most of boardgame's game semantics (stacks, moves, fix-up moves,
phases, per-viewer hidden state) available as an interpreted TypeScript toolkit.
No Go machinery runs on the Fabric side.

Three constraints shape everything:

1. **Share code, do not fork.** The extraction is a refactor of the existing
   webapp. One copy of the rendering and animation code serves the Go-backed
   webapp and the Fabric host, and each step leaves the webapp more modular.
2. **Trust the attested runtime, not the browser.** Fabric's server-side
   execution (in TEEs with remote attestation) plays the role the Go server
   plays today. Hidden state is protected by CFC labels and blessed functions,
   not by cryptographic protocols between players.
3. **Two UI hosts over time.** First a `cf-iframe` guest. Later, a trusted
   component set that any pattern's JSX can use, once user policy can admit
   additional trusted web components.
4. **Design for CFC as ruled, not as shipped.** Where the information-flow
   work is ruled or in progress, the design assumes it is complete and names
   the assumption. "What the design assumes" lists each one.

## What the research found

### The boardgame frontend (`server/static/src`)

- Lit 3 and TypeScript, built with Vite. Nothing is published; the package is
  the private `boardgame-webapp`.
- The portable leaf set depends only on `lit`: `boardgame-card`,
  `boardgame-token`, `boardgame-die`, `boardgame-component`,
  `boardgame-component-stack`, `boardgame-component-animator`,
  `boardgame-animatable-item`, `boardgame-fading-text`, `component-view.ts`,
  and the `motion/` and `solid/` modules. Its import closure is 60 files and
  22.1k lines, within a client of about 65k. At runtime it depends on `lit`
  and pure modules; the stack also imports about 2k lines of move machinery.
- Coupled to Redux and routing: `boardgame-game-state-manager`, which holds
  the queue of pending states, and `boardgame-game-view`.
  `boardgame-render-game` reaches the server in three places only.
- The animation loop is: `animator.prepare()`, set the new state, await
  render, `animateFlip()`, then an `all-animations-done` signal releases the
  next state bundle. `src/examples/dice-keep-demo.ts` already drives this with
  no server.
- Data contract per stack: `{Deck, Indexes, IDs, IDsLastSeen, ShuffleCount,
  Size}`. The client expands it (`selectors.ts`): index `-1` is an empty slot,
  `-2` is a hidden card, anything else resolves against the chest. A card
  renders face-up exactly when it has `Values`.
- Hazards for other hosts: a module-global stack registry and
  `document.getElementById` lookups, globally registered element names,
  `offsetParent`-based geometry, a document-relative token asset path, the
  stack mutating `item.ID` on the state it is given, and function-valued
  props (`cardView({render})`).

### The boardgame semantics (Go core)

- Deck contents are public. The only secret is which card sits where.
- Every component is in exactly one slot of one stack; the engine validates
  conservation before every save.
- A move is `Legal(state, proposer)` plus `Apply(state)`. After a player move
  the engine repeatedly applies the first legal fix-up move, as admin.
- Sanitization is per viewer and server-side: five policies (visible, order,
  len, nonempty, hidden) resolved over groups (all, self, other, custom), with
  the least restrictive applicable policy winning.
- Card ids are salted hashes, stable across ordinary moves so a face-down card
  can be followed and animated, and re-rolled for a whole stack on a shuffle
  or secret move.
- Newer subsystems worth keeping: the legal framework (serializable
  predicates, each tagged with the facet it reads, explicitly aimed at a
  TypeScript evaluator that does not exist yet), move choices projected per
  actor, and per-viewer visibility facets.
- The design assumes one omniscient authority in ten places: it shuffles,
  holds the salt, runs `Legal`/`Apply`/fix-ups/scoring on unsanitized state,
  sanitizes per viewer, assigns proposer identity, and fires timers.

### Common Fabric (labs, loom)

- **UI.** Pattern code runs in a worker and cannot define custom elements,
  measure the DOM or run animation code. The reconciler destroys and recreates
  a node that moves between parents. So today the ported components can only
  run in a `cf-iframe` guest, which is a full document whose script is a
  bundle that may include pinned npm modules. Assets arrive as `data:` URLs.
- **Code reuse.** Pattern code imports only relative paths and a few
  `commonfabric` specifiers. No npm in pattern code. Guests can bundle npm.
- **Execution.** Server-primary execution is on by default. The server's
  run is authoritative. A client also runs a handler speculatively when it
  holds the handler's inputs; when it does not, the speculation is simply
  withdrawn.
- **Moves can be authenticated.** An event's author is stamped by the memory
  server from the signed commit, and event streams are append-only for
  clients. (From the protocol document; not checked in code.)
- **Scopes are enforced at the server.** A session cannot read another
  user's scoped instance.
- **Randomness and time.** `Math.random` and `Date.now` work only in handlers
  (one-second clock). Lifts and computeds throw on entropy. No `setTimeout`.
  No `crypto` in the pattern sandbox.
- **CFC primitives that exist.** `Confidential<T, [PolicyOf<rules>]>`,
  module-authored exchange rules guarded by a runtime-minted
  `TransformedBy{moduleIdentity, symbol}` atom (the "blessed lift" form,
  labs#8027 and follow-ups), input witnesses, the custody seal,
  `WriteAuthorizedBy`, render boundaries, and the host-read gate that decides
  what an iframe guest is handed. Worked examples:
  `labs/patterns/examples/cfc-exchange-rules/blessed-computation.tsx`,
  `labs/patterns/showcase/inductor/pool/pool-policy.tsx`, and loom's
  `src/patterns/cf-location-findmy.tsx`. The pool's README lists open
  snags, including a witnessed release that is refused when the witness
  handler read nothing confidential, which is the case of a game's first
  deal.
- **Prior exploration.** labs draft PR #3750
  (`origin/boardgame-cfc-sanitization`,
  `docs/proposals/cfc-game-sanitization-memo.md`). Its conclusion stands: no
  new primitive; per-field labels, a trusted reducer per coarser view (count,
  exists, order), and reveals as relabels fired by trusted actions. Its
  reducers were simulated; the blessed-lift workflow now makes them real.

### Gaps that affect games

| Gap | Effect on a card game | Plan |
|---|---|---|
| Release to a principal chosen at runtime ("deal to player p") is not expressible with shipped primitives. | Per-player hands | Seats as audience lists (see "Dealing to a player"). Memory needs no per-player release, so it goes first. |
| Lifts have no entropy. | Shuffle | Draw only a salt in the move handler; everything else is deterministic from it, as in the Go core. |
| A released value is sealed again once mapped. | Rendering a hand row by row | Release each seat's view as one whole object and hand it straight to the guest. |
| Cells deliver the latest value, with no ordered log. | Animating each fix-up move | The view carries a window of recent versions; the guest plays them in order and drops ones it has shown. |
| Input witnesses cover only confidential inputs. | Phase-gated reveals; seat keys | Keep phase, and anything else that steers a release, inside the sealed state. |
| One flow label per transaction. | The move handler cannot also write public cells | Everything public is a projection. |
| No `setTimeout`. | Timers (memory's hide-cards delay) | A deadline in state; the guest sends a tick that the engine checks against the clock. |
| A module's identity covers everything it imports. | An engine edit changes every game's policy identity and strands sealed state in flight | Pin the engine version per game; plan a migration path. |

## Architecture

```
boardgame repo
  elements, renderer kit   shared web client layers (see PROPOSAL.md)
  engine                   TypeScript engine, held to the Go core
  server/static            Go webapp, a consumer of all three
  fabric/                  Fabric patterns, guest bundle, example games

Fabric server-side execution (trusted)        Browser (untrusted)
  one table space                               cf-iframe guest
    sealed state  --blessed lift-->  seat view  ---->  shared elements
    move handler  <--------------  authenticated move  <--  a gesture
```

The shared client layers and the TypeScript engine are specified in
[PROPOSAL.md](PROPOSAL.md). The rest of this document is the Fabric side.

### The shape: one space, labels do the work

The browser is not trusted. Server-side execution, running in an attested
environment, is. So the Fabric design is the Go server's design, with CFC
labels in the place of the server's per-viewer sanitization:

- **A browser contributes one thing: an authenticated move.** "Player p
  proposes move M with these arguments." A modified client can send any move
  it likes, and it can only ever count as that player's own proposal.
- **Everything else runs server-side only:** legality, apply, the shuffle,
  the fix-up loop, and per-seat sanitization.
- **A browser receives one thing: its own seat's view,** plus what everyone
  at the table sees.

All of it lives in one space per table. Nothing is hidden by placing it in a
space players cannot read; it is hidden by its label.

- **Game pattern.** One sealed cell holds the full state and the move
  history. One `move` handler is the only writer. It takes the proposer from
  the runtime, never from the event's payload, runs the engine step, and
  writes the new state.
- **Views.** An exported blessed lift computes a seat's view for a bounded
  window of recent versions, and an exchange rule releases it. Counts and
  other coarse facts are part of the view.
- **Guest.** A `cf-iframe` guest receives the viewer's view, plays new
  versions in order, and sends moves.
- **Authoring.** A game is a definition (decks, state shape, policies,
  moves) plus a renderer. The toolkit supplies the pattern wiring.

The patterns import the engine by relative path, so they live in this
repository and deploy from it with `cf`.

### What the design assumes

The design is written for CFC as ruled or planned. Each assumption, and what
is true in labs today:

| Assumption | Status | Until then |
|---|---|---|
| A stored value is delivered only to principals in its label's audience (specs#95) | Ruled 2026-11-16. Labs delivers a space's documents to every reader of the space, and documents that as a non-conformance | A modified client that can read the table space can read the sealed state |
| A client does not run the engine | Follows from the row above: a client that is not sent the state cannot speculate on it | Every player's client holds the state and runs the engine on it; a speculative shuffle is replaced when the server's arrives |
| A release can name a list held in a cell (specs#59, labs#8608) | Unruled; draft | Per-seat views use the interim binding below |
| Server-side execution is attested | On the security roadmap | The deployment's operator is trusted |
| A client cannot commit a computation's result | Derived commits from a user session are refused; authored writes into the same documents are admitted | Views are kept where players cannot write |

With all five in place, a modified browser learns its own seat's view and
nothing else, and can do nothing but propose its own moves.

A table that needs that guarantee before specs#95 is implemented would have
to keep the state on a separate deployment under a dedicated dealer
identity, as loom's Find My does. That is a workaround for a gap already
ruled closed, and this design does not build it.

### Sealed state: data no person reads

The game state is readable by no person, the table owner included, while
trusted code computes on it. This is not unusual in CFC. It is the common
shape of the existing trusted patterns, and it has no name yet.

**What it is.** A store whose only confidentiality clause is a module-policy
reference:

```ts
type Sealed<T> = Confidential<T, readonly [PolicyOf<typeof rules>]>;
type State = WriteAuthorizedBy<Sealed<GameState>, typeof move>;
```

- No principal satisfies a policy-only clause by default. There is no owner
  rule and no writer rule: a player whose move rewrites the state cannot read
  it back.
- Compute is not gated on clearance. A handler invoked by player `p` reads
  and rewrites state `p` cannot see; the label propagates, and writes,
  egress, display and (as ruled) delivery are checked.
- The only way out is the policy's own exchange rules, guarded by
  `TransformedBy` for an exported lift of the same module.

**Who already does this.** The inductor pool
(`labs/patterns/showcase/inductor/pool/pool-policy.tsx`) is the closest
match: one write-authorized handler puts a random secret into policy-only
state and a witnessed lift releases. Conclave's custody seal, loom's
questions and Chance Calls policies, and Find My's raw slot carry the same
label. The custody seal itself is not the primitive for a deck: it takes one
person's reviewed value into custody, once and finally. Its review checklist
does apply.

**Rules for the game state**, taken from those cases:

- One writer. `move` is the only handler; bind it more than once, never add a
  second writer. Table setup is a move, so that everything the projections
  read was written by `move`.
- Witness every release: the rule's `TransformedBy` names the lift and an
  input witness naming `move`. A witness survives only when every
  confidential input carries it.
- One flow label per transaction. The move handler writes the sealed state
  and nothing public; turn indicators, logs and counts are projections.
- Keep `User(p)` out. Type the move event with the policy; anything labeled
  to one player that the handler reads seals every projection from everyone
  else.
- Projectors are total: no data-dependent throw, abort or log.
- No steerable public parameter chooses what a projector releases.
- The state names its own seats. A projector returns an empty view unless
  the seat it is asked for is one the state itself records. Without this, a
  table owner can run a second instance of the module, seat themselves
  everywhere, and point it at the real state.

**What a label does not settle.** These are about authority, not delivery:

| Route to the secret | Who | Closes with |
|---|---|---|
| Owner release by intent (specs#90, merged in labs#9660) | The table owner, at a terminal, with a record only they read | A ruling: exempt policy-only clauses, or show the record to the space's readers |
| Successor code via `setsrc` | The table owner | A ruling on write authority across updates for sealed stores |
| Adding oneself to a seat list | The table owner, who owns every list | The roster as a projection every player sees |
| Predicting the salt | Anyone who can observe `Math.random` in the same runtime | Policy keys (specs#58) as the seed |
| A forged write | A modified client with write access | Views and state where players cannot write; see the assumptions table |

**Charting it for other authors.** The same shape recurs with no shared name,
helper or recipe, and every existing example is append-and-tally. A game is
the first mutable state machine with per-viewer projections, so the first
demo doubles as the missing worked example:

- labs example: `patterns/examples/cfc-exchange-rules/sealed-state-machine.tsx`
  (handler-mutated state, two projections, a multi-user test). Memory is
  close to exactly this.
- labs helper: export `Sealed<T, Rules>` from `commonfabric/cfc`, and a rule
  builder that always emits the witnessed form.
- docs: one author recipe, "Sealed state: data no person reads", with the
  rules and the table above.
- specs, clarification: name the shape; state that a policy-only clause
  admits no principal and that compute is not clearance-gated.
- specs, rulings: owner release by intent against sealed clauses, and a
  witness bound to the instance.

### Dealing to a player: seats are audience lists

**This section is a hypothesis.** It is a reading of the list-release ruling
(specs#59, option E) and its draft implementation (labs#8608). The ruling is
unruled, the labs change is a draft from 2026-10-09, and nothing here has
been run. An earlier version of this design, posted as a comment on
specs#59, had a flaw described below.

"Deal to player p" is not a relabel of a card. Cards never leave the sealed
state. What is released is each seat's **view**, and the audience of seat
`i`'s view is "whoever is on seat `i`'s list". This is boardgame's own model
(`SanitizedForPlayer`), with CFC deciding who may hold which sanitized copy.

**Two audiences, two mechanisms.**

| View | Audience | Mechanism | Status |
|---|---|---|---|
| Observer view (what everyone at the table sees) | Readers of the table space | `HasRole(reader, THIS_POLICY.subject)` adds `User(reader)` | Shipped |
| Seat view (observer view plus that seat's secrets) | The seat's occupant list | `Members{list, subject}` resolved live | specs#59 |

**The pieces.**

- **Table space.** One per table. The table owner is its sole `OWNER`. The
  game policy's subject is this space.
- **Occupant lists.** One list per seat, with declared writers (seat and
  unseat handlers behind a gesture). An entry pins the player's principal at
  the gesture. Vacating a seat clears the list; a new occupant sees the hand
  from then on, as in boardgame.
- **A seat's audience, bound inside the sealed state.** The seat's index and
  its occupant list must be bound together by something `move` wrote, so
  that a view for seat 2 cannot be paired with seat 1's list and the release
  keeps its witness.
- **One module rule** adding the seat's `Members` to the policy clause, under
  `TransformedBy` for the projector.

**The flaw in the first version, and what is unresolved.** The first version
had the table owner's setup run write a small `seatKey` document per seat,
outside `move`. Two things are wrong with it:

- A witness survives only when every confidential input carries it. A
  confidential key not written by `move` drops the witness, so the rule
  never fires. A public key is a steerable parameter.
- The draft's capture check requires the output to carry exactly one
  module-policy clause, and requires the writing run's actor to be the
  subject's sole owner. A key written by `move` on another player's turn
  fails the second.

Making setup a move fixes the first. Whether the capture check can then be
satisfied, with the table owner making the setup move, is the question the
seat spike has to answer.

**What this buys, if it holds.**

- One policy digest per game, whatever the seating.
- Reveals over time (flop, showdown, a played card) need no CFC step: the
  state changes and every view shows it. The information-flow argument
  reduces to "is the projector a correct sanitizer".
- Boardgame's groups (self, other, teams, custom) stay inside the sanitizer.

**Questions for specs#59.**

1. Other players trigger the recomputation. Does a `Members` clause carried
   in from an input satisfy the rule's precondition without a new capture?
2. Players seat themselves. May a declared-writer handler run by player `p`
   append to a list owned by the table owner?
3. Must the rule and the lift live in each game's entry module? A toolkit
   wants the module-graph relation.
4. A second run of the same module, pointed at the real state, passes
   capture. Is a witness bound to the instance the intended fix?

**The interim binding.** Until list release exists, the toolkit releases
every seat view to the table's readers and has the pattern hand each guest
only its own. That protects against nothing but an honest pattern, so demos
built on it say so. The toolkit exposes the seat abstraction from the start,
so games do not change when the binding does.

## Spikes on the Fabric side

The order of work is in [PROPOSAL.md](PROPOSAL.md). These need a throwaway
spike in labs before the design above is relied on, most fundamental first:

- **A sealed state machine releases at all.** Hand-written memory: sealed
  state, one blessed observer view, a `cf-iframe` guest, two users. The
  nearest existing example lists open snags, so this comes before anything
  else.
- **The move arrives authenticated.** Confirm in code that the proposer the
  handler sees is stamped by the server.
- **Seats as audience lists**, on the labs#8608 branch: a two-seat table
  with one hidden card each. The occupant reads their view, the table owner
  does not, and a move by the other player recomputes it.
- **Timers:** a deadline in state and a guest tick.

## What a table protects, and from whom

| Property | With the assumptions above | In labs today |
|---|---|---|
| A player cannot see another seat's cards | Holds against a modified browser | Holds only against an unmodified client |
| A player cannot make a move as someone else | Holds: the server stamps the author | Same, per the protocol document |
| A player cannot make an illegal move | Holds: the engine runs server-side | Same for the authoritative run |
| A player cannot forge the state or a view | Holds where players cannot write | A modified client with write access can author into those documents |
| The table owner cannot see hands | Open: see "What a label does not settle" | Open |

**Accepted residuals.** The operator of the attested runtime is trusted. The
salt comes from a non-cryptographic generator until policy keys exist.
Linkage may leak through commit metadata.

## Open decisions

1. Whether the Fabric patterns live in this repository permanently or move to
   labs once the toolkit stabilizes.
2. Whether blackjack waits for the list-release ruling or ships first on the
   interim binding.
3. What to do with labs PR #3750: update the memo with the blessed-lift
   workflow, or close it in favor of these documents.
