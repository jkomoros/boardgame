# Projected Move Choices

**Status:** Implemented V1

**Date:** 2026-07-19

## Summary

Boardgame already has the right semantic unit for an observable player
decision: a move. It also has the right authority for whether a concrete move
is possible: the move's complete `Legal` chain. What it lacked was a safe,
generic way for a client to discover a small finite set of concrete moves that
are worth presenting now.

Projected move choices fill that gap. A game may opt one required move field
into an audited, actor-only projection over a sealed finite source (players or
enum values). At a durable proposal frontier, the server constructs one fresh
move per candidate, binds the candidate through the canonical input codec, and
calls the ordinary `Legal` method exactly once. The client validates the result
against its generated schema and hydrates ordinary `BoundMoveAction` objects.

The projection is deliberately not a move, phase, workflow cursor, modal, or
second legality system. It is an ephemeral read model over true moves:

```text
durable game semantics       ordinary moves, state, phases, Legal, sanitization
safe choice projection       finite actor-visible candidates + exact availability
client interaction           existing bound actions + game-owned presentation
```

For Valentine, playing a Guard, selecting its target, and guessing a card stay
separate committed moves. The framework merely derives the currently relevant
target or guess controls from those moves and their legality.

## The semantic rule

Use a move whenever the decision itself must be observable, replayable,
sanitizable, animated, or independently rejected by the server. A sequence of
such decisions remains a sequence of moves.

Use a client draft when several local selections compose one atomic move. A
checker source and destination, a set of cards to discard together, or a piece
plus orientation should not become several public moves merely to obtain a
generic UI.

Use domain state and phases when the protocol itself is part of the rules: an
auction, reaction window, simultaneous commitment, timer, or participant
obligation. A phase is not necessary merely because the UI temporarily asks a
question.

Move sanitization remains the authority for disclosing committed history. A
secret move may be visible only to its proposer, have a hidden name or fields,
or remain hidden forever. Projected choices are separate, prospective
disclosure and require an explicit actor-exact authorization.

That authorization is a declassification decision. It reveals that a choice
set exists, the identity, count, and order of every candidate, and the exact
`Available` bit for each one. Move-name visibility alone does not authorize any
of those facts. The required audit rationale is a human-review assertion, not
an executable information-flow proof: opaque `Legal` code may inspect hidden
state or external process data. Games must test hidden-equivalent states to
show that projected output does not vary with facts the actor may not learn.
V1 does not hide the timing, occurrence, size, or traffic pattern of a
projection response.

## Why subphases are not the primitive

A subphase can encode Valentine, but it overstates a UI concern as a durable
rules protocol. It adds progression state, transition logic, recovery cases,
and phase-aware render code while still failing to answer the client questions:
which field is being chosen, which values are safe to enumerate, and which
complete bindings are legal?

Subphases remain useful when the rules refer to them. Projected choices handle
the smaller and more common need: present the finite concrete moves that the
existing semantics already imply.

## Authoring contract

The Go declaration contains security and shape, never presentation:

```go
moves.WithChoiceProjection(
    choice.PlayerIndexes("OtherPlayerIndex").
        DiscloseExactAvailabilityToActor(
            "Player identities and target eligibility are public to the actor.",
        ),
)

moves.WithChoiceProjection(
    choice.EnumValues("GuessedCard").
        Excluding("Unknown", "Guard").
        DiscloseExactAvailabilityToActor(
            "The card catalogue is public and legality does not inspect the hidden hand.",
        ),
)
```

V1 validates at manager construction that:

- a move has at most one projection;
- the projected field is the move's only required creator input;
- a player source uses the player-index codec;
- an enum source uses the enum codec and only excludes canonical values;
- the disclosure is actor-exact and has a non-empty audit rationale;
- fix-up moves cannot publish player choices;
- static candidate universes fit the protocol limits.

The audit rationale stays on the server for review. It is excluded from the
wire contract, generated TypeScript, and schema fingerprint. Editing prose
therefore cannot stale a client.

`MoveChoiceProjectionSchema` and its fingerprint are intentionally separate
from `MoveInputSchema`. Choice disclosure can evolve without changing the
creator proposal protocol.

## Server projection

The server projects choices only when all of these conditions hold:

1. The requested state is the durable, recoverable proposal frontier.
2. The authenticated viewer is the same valid seated player who would propose
   the move; observer views, admin impersonation, and auto-current display do
   not inherit that player's disclosure authority.
3. The canonical move name is visible to that actor under move sanitization.
4. The declaration and candidate universe pass their bounds.

For each candidate the server creates a move pinned to the exact immutable
snapshot, preserves server defaults for non-required fields, binds the one
creator value through the ordinary codec, and calls full `Legal(state, actor)`.
The snapshot pin prevents accidental use of a newer game state; it does not
make arbitrary `Legal` code deterministic or remove external dependencies.
`Legal` remains the sole rules authority. Projection does not accept an
author-provided availability predicate and does not reuse the public preview
endpoint.

A set is emitted only if at least one candidate is legal. Once emitted, it
contains the entire safely disclosed finite universe, including unavailable
peers. This distinguishes an applicable choice with disabled alternatives from
an irrelevant move, while avoiding prompts in which nothing can be done.
Hidden, inapplicable, and all-illegal moves are intentionally absent; the
client must not infer which reason caused absence.

The wire model contains only versioned semantics:

```text
ProjectedMoveChoices
  StateVersion
  MoveChoiceProjectionSchemaFingerprint
  ProjectionSchemaVersion
  Status: ready | failed
  Sets[]
    MoveName
    FieldName
    Source: players | enum-values
    Candidates[]
      Value
      Available
```

There are no offer IDs, prompts, titles, layouts, audit rationales, or detailed
legality errors. Projection errors are logged server-side and become a generic
`failed` snapshot so an authoritative failure cannot masquerade as an empty
result.

## Durable proposal frontiers

Choice projection exposed a more fundamental engine requirement: a durable
state head is not necessarily ready for another player proposal. The process
may have died after saving one fix-up but before completing its recursive
chain.

The engine now persists `(ProposalFrontierKnown,
ProposalFrontierVersion)` separately from state commits. Every intermediate
state save carries an old or unknown marker. Only the initiating serialized
operation, after the terminal fix-up check, updates the marker if the supplied
version is still the current durable head. Database-backed stores perform that
comparison atomically. Filesystem storage retains its existing single-process
golden-store semantics and does not claim a cross-process compare-and-set.
Reload trusts durable evidence and never re-runs arbitrary game legality merely
to infer settlement.

The terminal state commit and marker update are deliberately separate writes.
A crash or storage failure between them creates a safe false negative, as does
loading a pre-migration record. On the next projected-choice `/info` load, the
server performs an explicit serialized `ForceFixUp` reconciliation, waits for
the terminal check and marker update, refreshes the read snapshot, and only
then advertises choices. It never treats unknown as settled. Custom storage
without frontier persistence can serve active-process choices but remains
conservatively unknown after reload.

A marker-write failure after the move/state commit is logged and leaves the
frontier unknown; it never reports the already committed move as failed.
Failure to reconcile does not take down ordinary game info: the eligible actor
receives the sanitized game plus a generic versioned `failed` choice snapshot.

This matters beyond projected choices. It provides a recoverable answer to
"may the next proposal be advertised?" for initial loads, reconnects, partial
failures, and externally triggered fix-ups. External `ForceFixUp` invalidates
the frontier before queueing work.

## Resource contract

V1 is intentionally finite:

- at most 8 projected sets;
- at most 64 candidates per set;
- at most 128 total legality evaluations;
- at most 32 KiB of JSON-encoded static enum candidate values;
- at most 64 KiB in the projected wire snapshot.

Static violations fail manager construction. Dynamic player universes and the
complete worst-case visible payload are preflighted before any game-authored
`Legal` call, including sets that might later prove all-illegal. As with every
existing legality endpoint, the count bound assumes each `Legal` implementation
terminates normally; the engine does not create unbounded timeout goroutines
around game code.

## Generated and client contract

The move-argument generator emits:

```ts
type MoveChoiceProjections = {
  'Guess Card': {
    readonly field: 'GuessedCard';
    readonly value: 'Baron' | 'Countess' | 'Handmaid' | 'King' | 'Priest' | 'Prince' | 'Princess';
    readonly input: GuessCardInput;
  };
};
```

It also emits the canonical projection schema and fingerprint, and generated
renderer bases install both. The client treats server JSON as untrusted until
it validates state version, protocol version, fingerprint, move/field/source,
candidate type, exact enum or player universe, uniqueness, bounds, and the
presence of at least one available candidate in every included set.

After validation, game code receives an exact API:

```ts
const guesses = this.choices?.get(MoveNames.GuessCard);
for (const candidate of guesses?.candidates ?? []) {
  candidate.value;      // narrowed card union
  candidate.available;
  candidate.action;     // ordinary BoundMoveAction
}
```

Every disclosed candidate has a bound action. Available and unavailable
candidates therefore share the existing action state, activation, proposal,
staleness, animation, and accessibility behavior. A complete exact projection
supersedes the incomplete default move form's baseline legality; the server
still rechecks `Legal` on proposal.

Prompts and candidate labels are client-owned `MessageDescriptor` values with
stable semantic IDs and required default messages. Framework defaults use
sanitized player presentations and humanized enum values. No locale-dependent
copy affects projection caching or crosses the server boundary.

The framework always renders a fixed, viewport-visible, bounded,
safe-area-aware semantic fallback. It reserves its measured height above
ordinary-flow boards, scrolls internally at constrained mobile/zoom sizes, and
does not block board interaction outside its visible surface. It uses
fieldsets, distinct accessible action names, polite live announcements, and a
visible alert for projection failure. It does not steal focus.

## Valentine

Valentine's existing Guard semantics are modeled correctly:

```text
Play Card (Guard)
  -> Select Player { OtherPlayerIndex }
  -> Guess Card { GuessedCard }
  -> Activate Guard fix-up
```

Each arrow is derived from authoritative state and ordinary move progression.
Each player decision is a committed move. Reconnect therefore needs no
interaction cursor: projection over the durable frontier discovers whichever
move is legal next.

The paired Valentine migration adds player and enum projections to the two
moves, excludes the `Unknown` and `Guard` enum sentinels, and removes
renderer-owned candidate enumeration plus `NeedToSelectPlayer` /
`NeedToGuessCard` UI flags. The generic fallback initially owns presentation.
That migration is a separate commit with its own generation, legality, and
browser gates; it is not implied merely by landing the framework primitive.

## Design-space journeys

| Journey | Correct semantic layer | Projected choices role |
| --- | --- | --- |
| Valentine: select Guard target, then guess | Separate committed moves | Project each one-field move in sequence |
| Choose another player for a public effect | One committed move | Project player indexes and exact target legality |
| Choose one public enum option | One committed move | Project enum values with explicit sentinel exclusions |
| Checkers: select a piece and destination atomically | Client draft of one move | None until the draft has a complete binding |
| Select several cards to discard together | Selection draft of one move | Not a V1 scalar projection |
| Place a piece with position and orientation | Placement draft of one move | Not a V1 scalar projection |
| Secret simultaneous vote | Domain protocol plus sanitized moves | Do not expose exact peer availability or participation |
| Auction or reaction window | Durable domain state/phase | May project bounded response moves inside the window |
| Server seating after external identity changes | External fix-up protocol | Frontier invalidation prevents stale player actions |
| Hidden historical action | Move sanitization | Projection does not alter committed disclosure |

The table is a boundary test for the abstraction. Projected choices are useful
because they do not try to absorb drafts, secrecy, progression, or protocols.

## Alternatives rejected

### Universal `Interaction` state

A persisted interaction object makes a UI grouping authoritative even when the
game rules do not need it. It duplicates progression and creates lifecycle,
recovery, history, and sanitization questions. Durable protocols should use
game-specific state; ordinary move discovery should remain derived.

### Action offers with server presentation

The discarded prototype combined security authorization, candidate discovery,
localization keys, grouping, and native-control ownership in one catalog. It
coupled Go copy edits to fingerprints, created a parallel action abstraction,
and made an empty custom renderer capable of suppressing the safe fallback.
The final design projects only move semantics and lets the existing client
action layer plus client-owned presentation do their jobs.

### Infer choices entirely in the client

The client cannot safely know which identities or enum members may be
enumerated, and duplicating `Legal` drifts from the server. Public preview APIs
also expose a different disclosure surface and may include detailed errors.

### Turn every local step into a move

This damages atomicity and leaks intermediate choices. Existing draft
controllers remain the right abstraction for composing one move.

## Deliberate V1 limits and next extensions

V1 projects exactly one required scalar field. Dependent or multi-field atomic
moves continue to use existing drafts. A future partial-binding extension must
be version-pinned, bounded in depth and total evaluation, validate every prefix
against a generated schema, and submit only one final true move. It must not
silently turn draft steps into history.

V1 also ships only the guaranteed generic fallback. A rich board-native region
would currently duplicate it. A future consumption mechanism must be
render-scoped and snapshot-scoped: the fallback may be suppressed for one set
only after a region proves it contains at least one non-null bound control for
that exact set and snapshot. Empty, stale, hidden, or disconnected regions must
never consume the fallback.

Enum ordering and rich per-candidate descriptions are client presentation
extensions. They should not change the server projection or its fingerprint.

Projection is never part of game correctness. Adding or removing a declaration
does not change legal moves, progression, history, or disclosure of committed
moves; it changes how a client discovers and presents those moves. A game may
choose the framework fallback as its player-facing control, so projection
failure is explicit and recoverable rather than silently interpreted as "no
move." Activating a projected candidate immediately proposes its ordinary
bound move; the projection adds no cancel, rollback, or partially committed
workflow protocol.

## Landing invariants

- Observable decisions remain true moves.
- `Legal` remains the only rules authority.
- Prospective exact disclosure is explicit, actor-only, and auditable.
- Hidden-equivalent-state tests cover every security-sensitive projection.
- Committed disclosure remains move sanitization's responsibility.
- Intermediate or unrecoverable durable heads never advertise actions.
- Projection failure is visible and cannot appear as authoritative emptiness.
- Generated types are exact; runtime validation precedes every cast.
- Generic fallback remains safe and accessible.
- Games without projections retain their client/game-state wire and proposal
  semantics, although built-in storage now records frontier metadata and MySQL
  requires its accompanying migration.
