# Boardgame on Common Fabric: research and CFC design

Status: draft, 2026-11-21. Nothing here is built yet. Findings come from
reading code and draft rulings; nothing was run.
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

## What the research found

### The boardgame frontend (`server/static/src`)

- Lit 3 and TypeScript, built with Vite. Nothing is published; the package is
  the private `boardgame-webapp`.
- The portable leaf set depends only on `lit`: `boardgame-card`,
  `boardgame-token`, `boardgame-die`, `boardgame-component`,
  `boardgame-component-stack`, `boardgame-component-animator`,
  `boardgame-animatable-item`, `boardgame-fading-text`, `component-view.ts`,
  and the `motion/` and `solid/` modules. About 51 files and 20.7k lines. It
  is a whole motion stack, not a small library.
- Coupled to the Go server, Redux and routing: `boardgame-render-game`,
  `boardgame-game-state-manager`, `boardgame-game-view`, the generated
  per-game contracts.
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
- **Execution.** Server-primary execution is on by default; handlers and
  derivations run server-side and clients commit intents.
- **Randomness and time.** `Math.random` and `Date.now` work only in handlers
  (one-second clock). Lifts and computeds throw on entropy. No `setTimeout`.
  No `crypto` in the pattern sandbox.
- **CFC, shipped and usable.** `Confidential<T, [PolicyOf<rules>]>`,
  module-authored exchange rules guarded by a runtime-minted
  `TransformedBy{moduleIdentity, symbol}` atom (the "blessed lift" form,
  labs#8027 and follow-ups), input witnesses, the custody seal,
  `WriteAuthorizedBy`, render boundaries, and the host-read gate that decides
  what an iframe guest is handed. Worked examples:
  `labs/patterns/examples/cfc-exchange-rules/blessed-computation.tsx`,
  `labs/patterns/showcase/inductor/pool/pool-policy.tsx`, and loom's
  `src/patterns/cf-location-findmy.tsx`.
- **Prior exploration.** labs draft PR #3750
  (`origin/boardgame-cfc-sanitization`,
  `docs/proposals/cfc-game-sanitization-memo.md`). Its conclusion stands: no
  new primitive; per-field labels, a trusted reducer per coarser view (count,
  exists, order), and reveals as relabels fired by trusted actions. Its
  reducers were simulated; the blessed-lift workflow now makes them real.

### Gaps that affect games

| Gap | Effect on a card game | Plan |
|---|---|---|
| Release to a principal chosen at runtime ("deal to player p") is not expressible with shipped primitives; Find My generates one module per viewer. | Per-player hands | Seats as audience lists, on specs#59 / labs#8608 (see "Dealing to a player"). Memory needs no per-player release, so it goes first. |
| The memory server ships a space's documents to every reader regardless of label. | A modified client could read hands | Dependency, not a blocker. Track label-gated delivery; state the tier each demo achieves. |
| Lifts have no entropy; no sealed-secret primitive. | Shuffle | Shuffle inside the move handler, which has `Math.random` and runs server-side. Not cryptographic. |
| A released value is sealed again once mapped. | Rendering a hand row by row | Release each seat's view as one whole object and hand it straight to the guest. |
| Input witnesses cover only confidential inputs. | Phase-gated reveals ("showdown") | Keep phase inside the sealed state. |
| Commit metadata and content-addressed ids can leak linkage. | Unlinkability across shuffles | Re-roll opaque ids in the engine; document the residual. |
| No `setTimeout`. | Timers (memory's hide-cards delay) | A spike: deadline in state plus a clock-driven tick. |

## Architecture

```
boardgame repo
  elements, renderer kit   shared web client layers (see PROPOSAL.md)
  engine                   TS engine: stacks, moves, predicates, sanitizer
  server/static            Go webapp, a consumer of all three
  fabric/                  Fabric patterns, guest bundle, example games

Fabric runtime (attested)                     Browser (untrusted)
  game pattern                                 cf-iframe guest
    sealed full state  --blessed lift-->  seat view  -->  shared elements
    proposeMove handler  <------------- move intent ----  propose-move event
```

The shared rendering packages and the TypeScript engine are specified in
[PROPOSAL.md](PROPOSAL.md). The rest of this section is the Fabric side.

### The Fabric host (`fabric/`)

- **Game pattern.** One sealed cell holds the full state, labeled with the
  game's policy. One `proposeMove` handler is the only writer
  (`WriteAuthorizedBy`). It identifies the proposer's seat, runs the engine
  step, draws randomness, and appends the resulting states.
- **Seat views.** An exported blessed lift computes `viewFor(state, seat)`
  for a bounded window of recent versions. An exchange rule keyed to that
  lift's `TransformedBy` atom releases the view. Coarser facts (hand counts)
  are part of the view, so no separate reducers are needed.
- **Guest.** A `cf-iframe` guest bundles the shared client layers, receives
  the seat view, feeds new versions to the renderer host in order, and sends
  `propose-move` events to the pattern's move stream. Card art ships as data
  files.
- **Authoring.** A game is a definition object (decks, state shape,
  policies, moves) plus a guest template. The toolkit supplies the pattern
  wiring.

The patterns import the engine by relative path, so they live in this repo
and deploy from it with `cf`.

### Sealed state: data no person reads

The game state is readable by no person, the host included, while trusted
code computes on it. This is not unusual in CFC. It is the common idiom, and
it has no name yet.

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
  and rewrites state `p` cannot see; the taint propagates and only writes,
  egress and display are checked.
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
  second writer.
- Witness every release: the rule's `TransformedBy` names the lift and an
  input witness naming `move`.
- One flow label per transaction. The move handler writes the sealed state
  and nothing public; turn indicators, logs and counts are projections.
- Keep `User(p)` out. Type the move event with the policy; anything labeled
  to one player that the handler reads seals every projection from everyone
  else.
- Projectors are total: no data-dependent throw, abort or log.
- No steerable public parameter chooses what a projector releases.
- The state names its own seats. `seatView` returns an empty view unless the
  key it is given is one the state itself links to. Without this, the host
  can run a second instance of the module, seat themselves everywhere, point
  it at the real state, and the witness still holds.

**What "nobody" does not yet cover.**

| Route to the secret | Who | Closes with |
|---|---|---|
| Raw replica or storage | Any reader of the table space; the operator | Delivery as a boundary (specs#95), attested runtime |
| Owner release by intent (specs#90, merged in labs#9660) | The host, at a terminal, with a record only they read | A ruling: exempt policy-only clauses, or show the record to the space's readers |
| Successor code via `setsrc` | The host | A ruling on write authority across updates for sealed stores |
| Adding oneself to a seat list | The host, who owns every list | The roster as a projection every player sees; a seat form of specs#59 where the host is not the list's owner |
| Predicting or steering the shuffle | Whoever runs the handler | Policy keys (specs#58) as the seed |

**Charting it for other authors.** The same shape recurs with no shared name,
helper or recipe, and every existing example is append-and-tally. A game is
the first mutable state machine with per-viewer projections, so the first
demo doubles as the missing worked example:

- labs example: `patterns/examples/cfc-exchange-rules/sealed-state-machine.tsx`
  (handler-mutated state, two projections, a multi-user test). Memory is
  exactly this.
- labs helper: export `Sealed<T, Rules>` from `commonfabric/cfc`, and a rule
  builder that always emits the witnessed form.
- docs: one author recipe, "Sealed state: data no person reads", with the
  rules and the table above.
- specs, clarification: name the idiom; state that a policy-only clause
  admits no principal and that compute is not clearance-gated.
- specs, rulings: owner release by intent against sealed clauses, and a
  witness bound to the instance.

### Dealing to a player: seats are audience lists

"Deal to player p" is not a relabel of a card. Cards never leave the sealed
state. What is released is each seat's **view**, and the audience of seat
`i`'s view is "whoever is on seat `i`'s list". This is boardgame's own model
(`SanitizedForPlayer`), with CFC deciding who may hold which sanitized copy.

It rests on the list-release ruling, specs#59 ("releasing to a list named by a
cell", option E) and its draft implementation labs#8608. Neither has landed:
the ruling is unruled and the labs PR is a draft from 2026-10-09. The shapes
below are a reading of those drafts, to be proven by spike 1.

**Two audiences, two mechanisms.**

| View | Audience | Mechanism | Status |
|---|---|---|---|
| Observer view (what everyone at the table sees) | Readers of the table space | `HasRole(reader, THIS_POLICY.subject)` adds `User(reader)` | Shipped |
| Seat view (observer view plus that seat's secrets) | The seat's occupant list | `Members{list, subject}` resolved live at display | specs#59 |

**The pieces.**

- **Table space.** One space per table. The host is its sole `OWNER`;
  players and spectators get `WRITE` or `READ`. The game policy's subject is
  this space.
- **Sealed state.** Labeled with the game policy clause alone, so no person
  reads it, the host included. Written only by `proposeMove`.
- **Seat records, authored once at setup.** In the host's setup run, for each
  seat `i` the pattern writes a small `seatKey` document `{seat: i}` whose
  schema declares `members: "/seats/i/occupants"`. Label capture gives it
  `[User(host) ∨ Members(occupants_i, table)]`. This is the only place a
  `Members` clause is authored, and the host solely owns the subject, so the
  capture check passes.
- **Occupant lists.** One list per seat, owner-bound to the host, with
  declared writers (`seat` and `unseat` handlers behind a gesture). An entry
  pins the player's principal at the gesture. Vacating a seat writes an empty
  list; a new occupant sees the hand from then on, as in boardgame.
- **The blessed lift.** `seatView(state, seatKey)` reads the seat index from
  the key and returns that seat's sanitized view. Its output inherits the
  key's `Members` clause as an ordinary input-derived clause, so recomputing
  after another player's move authors nothing and needs no capture.
- **One module rule.** Under `TransformedBy{this module, "seatView"}` with a
  literal input witness, and a `Members` pattern on `THIS_POLICY.subject`,
  add that `Members` to the policy clause. The view's label becomes
  `[Policy ∨ Members(occupants_i)] ∧ [User(host) ∨ Members(occupants_i)]`:
  the occupant satisfies both clauses, and the host satisfies only the second
  unless seated.

**What this buys.**

- One policy digest per game, whatever the seating.
- The seat index and its audience are bound in one host-authored document, so
  a view for seat 2 cannot be paired with seat 1's list.
- Reveals over time (flop, showdown, a played card) need no CFC step: the
  state changes and every view shows it. The whole information-flow argument
  reduces to "is `seatView` a correct sanitizer", which is one function to
  review.
- Boardgame's groups (self, other, teams, custom) stay inside the sanitizer.

**Questions this case raises for specs#59.** A card table differs from Find
My in three ways worth recording there:

1. The secret belongs to no one. The host owns the subject only as custodian
   and must not read hands. Is a policy clause with no owner alternative an
   intended use, and does `User(Q)` in the authored clause stay harmless?
2. Recomputation is triggered by other players. The design depends on a
   carried `Members` clause satisfying the rule's precondition without a new
   capture, and on the key document counting as a witnessed input.
3. Players seat themselves. May a declared-writer handler run by player `p`
   append to a list owner-bound to the host, or must the host admit each
   player by gesture? The plan assumes host admission until ruled.

Two more, already open on labs#8608: whether the rule and the lift must live
in each game's entry module (a toolkit wants the module-graph relation), and
minting `ListedIn` in the iframe host-read gate, which postdates the draft.

**Residuals.** The table's owner can use an owner release by intent
(specs#90) on documents in the table space; that is in person and recorded.
The roster of each seat is visible to table readers.

**Until the ruling lands.** The toolkit exposes the seat abstraction from the
start. An interim binding releases every seat view to table readers and has
the pattern hand each guest only its own seat's view. That protects against
nothing but an honest pattern, so demos built on it say so.

## Spikes on the Fabric side

The sequence of work is in [PROPOSAL.md](PROPOSAL.md). Two questions here
need a throwaway spike in labs before the design above is relied on:

- **Seats as audience lists**, on the labs#8608 branch. A two-seat table with
  one hidden card each: prove the carried `Members` clause releases a seat
  view to its occupant, that the host cannot read it, and that a move by the
  other player recomputes it.
- **Timers without `setTimeout`:** a deadline in state plus a clock-driven
  tick.

## Security tiers

State plainly which tier a demo reaches:

- **Tier A, available now:** honest runtimes. Labels, the blessed release and
  the host-read gate keep other seats' cards out of the guest.
- **Tier B, needs label-gated delivery:** an untrusted browser cannot obtain
  another seat's cards from the memory server.
- **Accepted residuals:** the operator of the attested runtime is trusted;
  shuffles use a non-cryptographic generator; linkage may leak through commit
  metadata.

## Open decisions

1. Whether the Fabric patterns live in this repository permanently or move to
   labs once the toolkit stabilizes.
2. Whether blackjack waits for the list-release ruling (specs#59) or ships
   first on the interim binding.
3. What to do with labs PR #3750: update the memo with the blessed-lift
   workflow, or close it in favor of these documents.
