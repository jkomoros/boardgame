# Action Offers and Move-Input Presentation

**Status:** Revised design draft after engine, client, and adversarial critique

**Date:** 2026-07-17

**Scope:** Boardgame move authoring, viewer-specific server projection, and
client interaction behavior. This document does not authorize implementation.

## Summary

Boardgame already has authoritative representations for game state, committed
moves, move ordering, and concrete move legality. It also has a typed client
action stack: generated creator inputs, `MoveActionBuilder`, `BoundMoveAction`,
`TargetAction`, exact legality preview, submission gates, draft controllers,
and accessible target lists.

What is missing is generic discovery and presentation. A client cannot learn
that a currently relevant move needs a player, which player values are safe to
present, how to label the choice, or that the resulting actions should be shown
as one coherent game experience. Games consequently rebuild that knowledge in
their renderers and can drift from server legality.

This design adds a derived, viewer-specific `ActionOffer[]` read model. An offer
decorates an existing typed move action with:

- a safe candidate source for one creator-owned input;
- structured, localizable presentation metadata;
- disclosure-aware available or disabled candidate actions.

Candidate sources provide a presentable universe, not legal truth. The server
binds candidates using the canonical creator-input codec and evaluates the
existing complete `move.Legal()` chain. There is no domain validator,
availability rule language, workflow cursor, or second client action runtime.

The semantic layers are:

```text
Durable rule-level protocol
    Domain state and phases, only when the protocol matters to the game

Committed decisions
    Ordinary moves, progression, Legal(), history, and disclosure

Uncommitted composition
    Existing specialized client-side drafts of one move

Presentation read model
    Viewer-specific offers decorating existing typed move actions
```

For Valentine, playing a Guard, selecting a player, and guessing a card remain
separate committed moves. `MoveSelectPlayer` presents the safely disclosable
player universe and the normal legal chain determines which players are valid.
`MoveGuessCard` presents the card enum and the normal legal chain excludes
Unknown and Guard. The renderer no longer implements `NeedToSelectPlayer`,
`NeedToGuessCard`, or its own target-legality logic.

## Decision

Adopt typed, viewer-specific action offers as a presentation projection over
existing move forms and actions.

For version one:

- moves explicitly opt into offers;
- an offer has exactly one required creator-owned input;
- player-index and enum candidate sources are supported;
- presentation is limited to structured prompt and title keys;
- `move.Legal()` is the sole semantic authority;
- every selectable candidate is exact-evaluated by the server;
- candidate disclosure is explicitly `PublicExact`;
- offers hydrate existing `BoundMoveAction` and `TargetAction` objects;
- offers are emitted only for the current settled game version;
- legacy games and non-offer moves keep their current behavior.

Do not introduce in version one:

- a universal persisted `Interaction`;
- a progression cursor or structural-frontier traversal;
- `WithAvailability`;
- executable input-domain validators;
- a second client action/controller stack;
- dependency DAGs or a universal local wizard;
- arbitrary cross-field input-group rules;
- pagination, urgency, or presentation continuity groups;
- binding-only inaccessible inputs;
- multi-field choice bindings;
- unbounded input offers;
- opaque tokens unless a privacy-reviewed use case makes them mandatory.

## Problem

### The engine knows legality, but the UI must rediscover intent

The engine can authoritatively answer whether this complete move is legal:

```go
MoveSelectPlayer{OtherPlayerIndex: 3}
```

The renderer also needs to know:

- whether selecting a player is relevant for this viewer and snapshot;
- that `OtherPlayerIndex` should be presented as a player choice;
- which player identities are safe to disclose as candidates;
- which candidates are available or disabled;
- how to label and contextualize the choice;
- how to bind a selected value into the existing typed action.

Without a framework contract, Valentine exposes computed values such as
`NeedToSelectPlayer` and `NeedToGuessCard`, constructs candidate collections in
the renderer, and independently decides which controls to show. The moves still
repeat the authoritative checks. That is a UI-shaped second implementation of
the flow.

### Default move instances conflate unavailable and incomplete

A default `MoveSelectPlayer` commonly has an unset sentinel such as
`AdminPlayerIndex`. That concrete instance is illegal, but selecting a player
may still be the correct action. The move is incomplete, not necessarily
unavailable.

The existing client baseline gate can currently reject a bound action because
the default instance has `LegalForAnyone == false` before exact preview can
validate a legal non-default candidate. Correct offer behavior requires:

```text
candidate source
    → complete candidate binding
    → exact legality
    → existing BoundMoveAction
```

Default-instance legality must not permanently disable a move with required
creator input.

### Similar-looking multi-step experiences have different semantics

1. **Local composition of one move.** Selecting a checker and its destination
   is normally one uncommitted `MovePiece{Source, Destination}`. Existing draft
   controllers own cancellation, undo, selection policy, and editing behavior.
2. **A sequence of committed decisions.** Playing a Guard, selecting its
   target, and guessing a card can require separate history, disclosure,
   observation, and animation boundaries. Each decision is a move.
3. **A durable game protocol.** An auction, reaction window, or simultaneous
   secret vote has rule-level identity, participants, completion rules, and
   perhaps a timer. It belongs in domain state and sometimes a phase.

Offers are the common UI projection across these cases. They do not replace the
appropriate authoritative representation underneath.

## Goals

1. Make ordinary bounded move choices discoverable and generically renderable.
2. Keep observable decisions as ordinary committed moves.
3. Keep the existing legal chain as the only rules authority.
4. Reuse the existing typed action, exact-preview, submission, and draft stack.
5. Reconstruct current choices after reconnect from authoritative state.
6. Expose plural offers rather than inventing one active workflow step.
7. Make offer and candidate disclosure explicit and testable.
8. Support both board-native controls and an accessible semantic fallback.
9. Preserve exact generated creator-input types end to end.
10. Make the Valentine authoring and renderer substantially smaller.

## Non-goals

- Inferring a workflow cursor from `MoveProgressionGroup`.
- Treating presentation grouping as game state.
- Replacing domain state for auctions, reactions, or simultaneous obligations.
- Generically enumerating combinatorial or unbounded move spaces.
- Guaranteeing that an offer remains legal after its source version changes.
- Adding rollback for committed moves.
- Hiding websocket timing or global version advancement.
- Replacing existing specialized client draft controllers.
- Making every existing move automatically offerable.

## One authority per concern

| Concern | Authority |
| --- | --- |
| What happened | Committed moves |
| Durable facts and obligations | Game and player state |
| Permitted move ordering | Existing move progression |
| Concrete proposal legality | Existing `move.Legal()` chain |
| Viewer knowledge | Sanitization and explicit disclosure policy |
| Creator field ownership and encoding | `BuildMoveInputSchema` |
| Presentable input universe | Candidate source |
| Current UI opportunity | Derived action offer |
| Binding, preview, and submission | Existing typed move-action stack |
| Editing behavior for one atomic move | Existing specialized draft controller |
| Layout and visual treatment | Renderer |

Deleting all offers must not change game correctness. An offer can improve
discovery and presentation but cannot create a new legal move or obligation.

## Authoring model

### Candidate sources, not executable domains

A candidate source owns the viewer-safe universe that may be presented for one
creator input. It does not decide authoritative legality. Version-one sources
are sealed framework descriptors, not arbitrary game callbacks:

Conceptually:

```go
type ChoiceSource interface {
    isFrameworkChoiceSource()
}

input.Players()
input.EnumValues()
```

The unexported marker prevents a game from implementing a source that reads the
live `Game`, globals, time, randomness, I/O, or an unbounded allocation before
the framework can enforce a budget. Custom and dynamically filtered sources are
deferred. Game rules filter the finite framework universe through `move.Legal()`.

The framework owns:

- field and codec compatibility checks;
- canonical binding through `BuildMoveInputSchema`;
- exact `move.Legal()` evaluation for every bounded candidate;
- available/disabled/undisclosed projection;
- payload, count, and work limits;
- deterministic ordering checks where possible.

The source does not implement `Validate`. It is impossible to guarantee that
two arbitrary `Choices` and `Validate` methods express the same rule. Candidate
membership and legal membership are intentionally different concepts: a public
player may be safe to show but currently illegal to target.

### Existing legality remains the sole rule language

All game rules remain ordinary legal predicates or `LegalCustom`:

```go
auto.MustConfig(
    new(MoveSelectPlayer),

    moves.WithLegalPreconditions(
        valentine.ActiveCardRequiresTarget(),
        legal.PropEquals("player.SelectedPlayer", "admin").
            WithMessage("valentine.already_selected"),
        valentine.ValidTarget("move.OtherPlayerIndex"),
    ),

    moves.WithInputPresentation(
        input.PlayerIndex("OtherPlayerIndex").
            Prompt("valentine.choose_player").
            Choices(input.Players()).
            Disclosure(input.PublicExact),
    ),

    moves.WithPresentation(
        presentation.Title("valentine.resolve_guard"),
    ),
)
```

The server binds every safely disclosable player and evaluates the same legal
chain that proposal submission uses. `LegalCustom` works automatically during
migration because complete candidates reach the normal `move.Legal()` path.

Guard guessing follows the same pattern:

```go
auto.MustConfig(
    new(MoveGuessCard),

    moves.WithLegalPreconditions(
        valentine.ActiveCardIs(cardGuard),
        legal.PropNotEquals("move.GuessedCard", "Unknown").
            WithMessage("valentine.guess_not_legal_type"),
        legal.PropNotEquals("move.GuessedCard", "Guard").
            WithMessage("valentine.guess_cant_be_guard"),
    ),

    moves.WithInputPresentation(
        input.Enum("GuessedCard", cardEnum).
            Prompt("valentine.guess_card").
            Choices(input.EnumValues()).
            Disclosure(input.PublicExact),
    ),

    moves.WithPresentation(
        presentation.Title("valentine.resolve_guard"),
    ),
)
```

Unknown and Guard can be shown as safely disabled choices or omitted according
to authored presentation policy. The `PropNotEquals` predicates—not the enum
source—own their illegality.

### No separate availability category in version one

For a bounded single-input offer, the framework can bind every candidate and
run complete legality. It emits the offer only when at least one candidate is
legal, then may include safely disclosed disabled sibling candidates.

The existing declarative plan may later support partial evaluation by deferring
predicates whose declared `move.*` reads intersect unbound required creator
fields. Context-owned and server-defaulted fields must be classified using the
canonical input schema rather than a coarse “field-dependent” flag. Opaque
`LegalCustom` remains deferred until binding is complete.

That optimization does not require new author vocabulary. `WithAvailability`
should be introduced only if an unbounded real game cannot be expressed through
candidate enumeration or dependency-aware projection.

### Explicit offer opt-in

Offer configuration is independent from declarative-legality adoption. Only a
move with `WithInputPresentation` participates. Existing required creator fields
without input presentation remain valid legacy moves and retain their current
forms and renderer behavior.

Version one configuration fails only for an opted-in move when:

- the field is unknown, context-owned, server-defaulted, or unsupported;
- the candidate kind is incompatible with the creator codec;
- the move has zero or more than one required creator input;
- two presentation declarations claim the same field;
- prompt or title keys are invalid;
- the disclosure mode is absent or unsupported.

### Structured localization

The server sends message keys and viewer-safe primitive arguments, never
already-localized prose:

```ts
interface LocalizedMessage {
  readonly key: string;
  readonly args?: Readonly<Record<string, string | number | boolean>>;
}
```

Version one supports a static prompt key and title key. Arbitrary context
callbacks, pluralization helpers, urgency, and continuity metadata are deferred
until their localization and disclosure behavior is proven.

Candidate labels are framework-owned in version one. `Players()` uses the
viewer-safe player display-name projection, falling back to a localized
“Player {number}” key. `EnumValues()` uses the generated enum metadata and its
localization-key convention, falling back to the enum's declared display name.
The server sends only semantic keys/values; the client selects the locale.
Custom label resolvers are deferred because they introduce another disclosure
surface.

## Disclosure model

### Exact legality is itself information

The server evaluates candidates against authoritative state. Offer existence,
candidate identity, count, ordering, enabled state, and rejection reason can all
reveal hidden facts. Applying sanitization after exact filtering is too late.

Every candidate source therefore declares how its results may be disclosed. The
long-term policy vocabulary is:

```text
PublicExact
    Candidate identity and exact legal status are public to this viewer.

PublicSuperset
    Candidate identity is public, but hidden-dependent legal status is not.
    Show a safe superset and validate only on submission with a safe response.

Opaque
    Raw creator values are not viewer-visible. Use authenticated, viewer- and
    version-bound tokens or a privacy-reviewed specialized adapter.

Hidden
    Do not emit the offer to this viewer.
```

Version one implements only `PublicExact`. It is an explicit, trusted
declassification by the game author for the offer actor. It authorizes candidate
identity, candidate membership, and enabled/disabled status as public to that
audience. It does not authorize detailed rejection reasons; reason disclosure
is a separate opt-in and defaults to a generic disabled result.

Declared legal read sets can warn about obviously hidden dependencies but do not
prove safety. In particular, opaque `LegalCustom` code is not mechanically
inspectable. The author assertion, review, and hidden-equivalent-state tests are
all required. `PublicSuperset`, `Opaque`, and hidden-dependent preview require a
later privacy design.

### Offer noninterference

The critical privacy invariant is:

> Two authoritative states that are identical under a viewer's allowed
> disclosure must produce byte-equivalent offer snapshots, except for facts
> explicitly authorized by the offer's disclosure policy.

Tests must compare hidden-equivalent state pairs, not only different viewer
roles on one state.

Exact-preview errors are also disclosure-sensitive. The current batch-preview
path returns the complete `move.Legal()` error text. Version one may expose a
candidate reason only when all facts behind that reason are public under
`PublicExact`. Server diagnostics remain detailed; unsafe client responses use
stable generic codes.

### Private moves and issue #693

The private simultaneous-choice journey is aspirational. Full support is
blocked on the relevant move-disclosure work in issue #693.

The design distinguishes:

- move payload secrecy;
- proposer secrecy;
- move-name secrecy;
- version, timing, and occurrence secrecy.

Viewer-specific offers can help with the first three once move disclosure is
implemented. The current transport still broadcasts global version activity
and cannot conceal that some event occurred. This design makes no
traffic-analysis-secrecy promise.

## Runtime offer projection

### Extend existing move forms and actions

`ActionOffer` is the conceptual read model. Its implementation should extend or
associate with the existing viewer-specific move-form bundle and hydrate the
existing typed action system. It must not introduce independent preview,
submission, caching, animation, or stale-state machinery.

The client-facing snapshot is an outer envelope:

```ts
interface ActionOfferSnapshot {
  readonly stateVersion: number;
  readonly moveInputSchemaFingerprint: string;
  readonly offerSchemaVersion: number;
  readonly offers: readonly GameActionOffer[];
}
```

The shared version and fingerprint are not repeated on every offer.

### Preserve exact generated types

Generated offers form a game-specific discriminated union rather than falling
back to `string` and `Record<string, unknown>`:

```ts
type GameActionOffer =
  | ActionOfferFor<
      typeof MoveNames.SelectPlayer,
      {
        OtherPlayerIndex: ChoiceInputOffer<number>;
      }
    >
  | ActionOfferFor<
      typeof MoveNames.GuessCard,
      {
        GuessedCard: ChoiceInputOffer<Card>;
      }
    >;
```

The exact generated facade exposes:

```ts
offers.get(MoveNames.SelectPlayer)
offers.get(MoveNames.GuessCard)
```

A field lookup is also exact:

```ts
offer.choices("OtherPlayerIndex")
```

Misspelled move names, invalid fields, and incorrectly shaped values remain
compile-time errors.

### Candidate result model

For a safely disclosed universe, a candidate is:

```ts
interface OfferedChoice<Key, MoveName extends string, Input extends object> {
  readonly key: Key;
  readonly label: LocalizedMessage;
  readonly action: BoundMoveAction<MoveName, Input>;
}
```

Every disclosed candidate retains its existing bound action. Exact projection
hydrates that action's existing preview as legal or illegal, so `TargetAction`,
generic controls, and board-native controls share one disabled/reason path.
Absence means outside the disclosed universe. Detailed disabled reasons are
optional and must independently satisfy reason-disclosure policy.

Version one emits an offer only when at least one candidate passes complete
`move.Legal()`. Safely disclosed disabled sibling actions may accompany it.
Zero legal candidates suppress the offer; required-protocol deadlock diagnostics
remain a separate server health invariant and cannot silently depend on UI.

### Deterministic ephemeral identity

Version one emits at most one offer per configured move name for a viewer and
snapshot. Each offer still has a deterministic ephemeral `offerKey` for keyed
rendering and local reconciliation. It has no game semantics and is not
submitted as authority.

If two same-type obligations differ semantically, their discriminator must be a
durable state fact and appear as a context-owned or creator-owned move field. If
that cannot produce an unambiguous offer, the game must use a different move
type or defer until multiple-instance offers are designed.

### Projection algorithm

For one authenticated seated player at the current settled version:

1. Enumerate only move types explicitly configured for offers.
2. Use the move's existing hypothetical-next-move `inProgression` legality
   behavior. Do not traverse or interpret the progression tree.
3. Resolve the candidate source against an immutable, version-pinned offer
   context.
4. Bind each candidate through the canonical creator-input schema and codec.
5. Call the complete existing `move.Legal(state, proposer)` chain.
6. Apply the pre-authorized disclosure policy to candidate identity, status,
   and safe reason.
7. If at least one candidate is legal, emit one typed offer associated with
   existing move-form/action data and optionally include safely disabled peers.
   Otherwise suppress the offer and retain separate deadlock diagnostics.

Observers and admin receive no offers by default. `AdminPlayerIndex` is an
evaluation sentinel, not an offer actor. Any exception requires explicit policy.

### Settled snapshots only

Every source and legality evaluation receives only an immutable context pinned
to `stateVersion`. History access, if later supported, is capped at that
version. Providers cannot access the live game, `CurrentState()`, wall clock,
randomness, network, mutation, or process globals.

Actionable offers are emitted only after automatic fix-up closure for the
current proposal-accepting frontier. Historical snapshots and transient
animation versions carry no actionable offers. Every preview and proposal
includes `ExpectedVersion`.

This requires a proven server primitive, not a before/after version check. Stage
0 must either identify an existing lock/serialization guarantee or introduce an
atomic “proposal-accepting frontier” snapshot API. Offer implementation cannot
begin until intermediate fix-up versions are unobservable as actionable offer
sources by construction.

### Determinism, cost, and failure

Projection is deterministic for:

```text
(game, version, viewer identity/disclosure class)
```

Candidate order and ephemeral keys are canonical. A timer affects offers only
through a durable state/version transition, never by silently changing the same
version.

The server enforces mechanical limits rather than trusting provider cost hints:

- total candidates across all offers;
- full legality evaluations per projection;
- response bytes;
- deadline and cancellation;
- panic recovery around projection code;
- authenticated rate limits;
- viewer- and version-scoped caching only.

Cache keys include every projection input. Localization happens on the client,
so locale is not a server projection input. Completed offers are never cached
globally across viewers.

A projection failure is a server diagnostic, not “no legal action.” Development
and tests fail loudly. Production returns a stable generic projection error and
does not emit an unvalidated candidate set. A zero-result required protocol is
diagnosed independently of the UI so an engine deadlock cannot masquerade as an
empty prompt.

## Existing client integration

### Offers decorate `BoundMoveAction`

An available offer choice contains the existing exact action:

```ts
const offer = this.offers.get(MoveNames.SelectPlayer);
const choice = offer?.choices("OtherPlayerIndex").get(player.index);
const action = choice?.action ?? null;
```

Current buttons, stacks, target lists, preview state, submission gates,
snapshot identity, transport errors, animation gates, and telemetry continue to
work unchanged.

Offer-derived complete bindings use their hydrated exact preview instead of the
incomplete default move form's baseline legality. Concretely, the action layer
must mark an offer-hydrated binding as `baselineLegalityApplies: false` (or an
equivalent explicit mode) for both `LegalForAnyone` and `LegalForPlayer` gates.
The existing schema, snapshot, animation, submission, and transport gates still
apply. A hydrated legal preview enables the candidate; a hydrated illegal
preview disables it. This is a required semantic change, not an optimization.

Offers should eliminate this existing renderer responsibility:

```ts
this.move(MoveNames.SelectPlayer).targets(players, player => ({
  OtherPlayerIndex: player.index,
}))
```

They should produce the same `TargetAction`-compatible result rather than
replace the machinery behind it.

### Generic accessible surface

A generic component consumes the same typed offer:

```ts
html`
  <boardgame-action-offer .offer=${offer}></boardgame-action-offer>
`
```

It renders a semantic candidate list, disabled states, safe reasons, pending
state, and errors by using existing action components. A board-native renderer
can bind the same candidate action to player panels, cards, or spaces.

A visual-only “binding” with a generic message saying “use the board” is not an
accessible fallback. Every offerable v1 input has a semantic list representation.

### Commit and cancellation behavior

Version-one offers are single-field committed moves, so activating an available
choice proposes the move immediately through its existing `BoundMoveAction`.
There is no new confirmation or cancellation protocol.

Existing local drafts retain their own clear, undo, and confirmation behavior.
Once an offer choice commits a move, Escape may minimize subsequent UI but
cannot imply rollback.

### Existing draft controllers remain specialized

Source/destination, selection, and placement controllers encode interaction
behavior beyond data dependency: toggle policy, undo, capacity, drag/drop,
rotation, pruning, and snapshot reconciliation. Offers may eventually supply
their candidate actions, but no universal dependency-DAG wizard replaces them.

## Valentine walkthrough

The authoritative sequence remains:

```text
MovePlayCard
→ MoveSelectPlayer, when required
→ MoveGuessCard, for Guard when required
→ MoveActivateGuard or another activation fix-up
```

1. The player chooses and commits `MovePlayCard` using the existing action API.
2. The resulting settled snapshot evaluates configured `MoveSelectPlayer`
   candidates against existing progression and complete legality.
3. The `Players()` source supplies the public player universe.
4. Exact legality marks protected, eliminated, and disallowed self-targets as
   disabled or omits them according to `PublicExact` presentation policy.
5. The client displays the offer generically and/or binds available
   `BoundMoveAction`s to player panels.
6. Activating one commits a real `MoveSelectPlayer`.
7. The next settled snapshot evaluates `MoveGuessCard` candidates.
8. The enum source supplies card values; existing legal predicates reject
   Unknown and Guard.
9. Refresh at either point regenerates the offer from authoritative state.

A successful migration removes:

- renderer-authored player candidate construction;
- renderer-authored guess-card candidate construction;
- `NeedToSelectPlayer`;
- `NeedToGuessCard`;
- duplicated targetability conditions in renderer code.

It preserves:

- `MoveSelectPlayer` and `MoveGuessCard` as committed moves;
- ordinary progression and activation fix-ups;
- proposal-time complete legality;
- current typed actions and target lists;
- current snapshot, submission, and animation behavior;
- board-native player panels and card UI;
- an accessible semantic fallback.

Valentine should also move card targeting policy toward named game data or
reusable legal predicates such as `NoTarget`, `TargetOther`, and
`TargetOtherOrSelfWhenNoAlternative`. That removes repeated card-type switches
without making the offer layer authoritative.

## Design-space validation

| Journey | Authoritative semantics | Offer treatment |
| --- | --- | --- |
| Valentine Guard | Several committed moves | Player then enum offers |
| Valentine Priest/Baron | Committed choice plus private information | Public parts in v1; private projection waits for disclosure design |
| Checkers source/destination | One atomic move | Existing local draft; future offer-supplied candidates |
| Darwin card/species | One atomic move | Existing local draft; future offer-supplied candidates |
| Memory repeated reveals | Repeated committed move | New offer after each settled reveal; no occurrence cursor |
| Werewolf voting | Durable per-player obligation | Domain state plus later private offers |
| Scrabble placement | One combinatorial atomic move | Specialized draft; not a v1 offer domain |
| Ticket to Ride two draws | State changes after first committed draw | Newly derived offer after first draw |
| Catan robber | Durable discards, then serial decisions | Per-player protocol state plus later offers |
| Coup reaction window | Durable timed/priority protocol | Domain state owns window; privacy-reviewed offers project responses |
| Secret choose/reveal | Durable private submissions | Aspirational; blocked on move disclosure and transport limits |

No one workflow abstraction owns these journeys. Offers are the replaceable UI
boundary; each game retains the correct authoritative representation.

## Why not the alternatives?

### Universal persisted `Interaction`

The current progression system recognizes move-name histories. Optional,
repeated, parallel, and custom groups do not expose a canonical node, actor, or
instance. A generic cursor would duplicate progression and legality. Real
auctions or reactions should use explicit domain state.

### Sub-phases for every prompt

Phases are appropriate for low-cardinality, rule-significant modes. Encoding
every card target or guess as a phase creates a taxonomy of screen choreography,
duplicates progression, and still does not describe move-field presentation.

### One atomic move for every interaction

Atomic drafts are correct when no observer or rule can act between fields. They
are incorrect when intermediate decisions must be durable, visible,
interruptible, animated independently, or disclosed differently.

### Executable input-domain validators

Giving input domains an independent `Validate` method creates another ordered
rules system and cannot mechanically guarantee agreement with candidate
description. Candidate sources provide presentation universes; existing legal
predicates remain authoritative.

### Manual game-authored `OffersFor`

A manual projection is maximally flexible but asks each game to duplicate
progression, exact legality, type binding, staleness, and disclosure. Games
select sealed candidate sources and author presentation; the framework derives
offers.

### Static prompt tags alone

A tag can say a field represents a player. It cannot provide current candidates,
exact status, disclosure, type-safe actions, or stale-state behavior. Static
configuration is the authoring source; the runtime offer is the needed
viewer-specific projection.

## Invariants

1. Deleting every offer cannot change game correctness.
2. `move.Legal()` remains the only authority for concrete proposal legality.
3. Candidate sources never make a candidate legal.
4. Every available candidate has passed complete legality at the offer version.
5. Proposal reruns complete legality against the accepted current version.
6. Offers hydrate existing typed actions rather than introducing another action
   runtime.
7. Move and field names remain exact generated types on the client.
8. Offers are deterministic and tied to one immutable settled version.
9. Offer projection is viewer-specific and disclosure-reviewed before exact
   status is exposed.
10. Hidden-equivalent states produce equivalent offers except for authorized
    disclosures.
11. Progression is queried through existing hypothetical-move legality, never
    independently traversed.
12. Committed decisions remain moves; local partial drafts are not moves.
13. Durable protocols remain explicit game state or phases.
14. Admin and observers receive no player offers by default.
15. Provider failure is distinguishable from an empty legal choice set.

## Testing strategy

### Configuration tests

- explicit opt-in and legacy byte-for-byte fallback;
- creator-input ownership and codec compatibility;
- exactly one required creator field in v1;
- player-index and enum source compatibility;
- required disclosure mode;
- prompt/title localization keys.

### Projection and legality tests

- every available candidate equals a complete successful `move.Legal()` result;
- disabled candidates preserve safe reasons only when authorized;
- `LegalCustom` filters complete candidates during migration;
- existing in-progression behavior controls offer presence without tree
  traversal;
- default sentinel illegality cannot block a legal bound candidate;
- proposal reruns full legality and rejects stale offers.

### Privacy tests

- current player, other player, observer, and admin recipient matrices;
- hidden-equivalent authoritative states produce byte-equivalent offers;
- candidate identities, counts, ordering, statuses, and reasons are each tested;
- trusted `PublicExact` covers identity/membership/status for the named audience
  but does not implicitly expose reasons;
- raw detailed projection and legality errors remain server-only when unsafe;
- private-choice journeys remain disabled until their disclosure prerequisites
  land.

### Determinism and resource tests

- canonical ordering and stable ephemeral keys;
- no map-order, wall-clock, randomness, live-state, or I/O dependence;
- per-projection total candidate/evaluation/byte limits;
- deadline, cancellation, panic recovery, and rate limiting;
- cache isolation across viewer and version;
- no actionable offers for historical or transient fix-up snapshots.

### Client tests

- generated exact offer union and typed move/field lookup;
- available and disabled choices map to existing action behavior;
- exact preview, stale rejection, and submission gates remain unchanged;
- generic semantic list and board-native controls share the same action;
- keyboard and screen-reader access;
- refresh reconstruction;
- committed actions cannot be locally “cancelled” after acceptance.

### Golden game journeys

- Valentine Guard is the first migration target.
- Valentine Priest/Baron tests the boundary where v1 public projection stops.
- Checkers or Darwin confirms existing local drafts remain independent.
- A hidden-equivalent fixture proves noninterference.
- Simultaneous secret choice is retained as a blocked future journey.

## Incremental delivery

### Stage 0: prerequisites

- Fix `BoundMoveAction` so default-instance `LegalForAnyone` cannot block a
  legal complete binding with creator input. Offer-hydrated actions explicitly
  defer both baseline legality gates to hydrated exact preview.
- Specify how offer projection associates with existing move forms and typed
  actions; do not create a second controller.
- Audit exact batch preview as a hidden-state oracle. Add disclosure-safe error
  behavior and noninterference tests before reusing it for offers.
- Confirm projection occurs only on settled proposal-accepting snapshots.
- Prove or add the atomic proposal-frontier primitive; this blocks Stage 1.
- Use existing hypothetical-next-move progression semantics without adding a
  structural-frontier API.

### Stage 1: public bounded offers

- Add explicit `WithInputPresentation` opt-in.
- Add only sealed framework-owned player-index and enum candidate sources.
- Validate configuration through `BuildMoveInputSchema`.
- Add structured prompt/title keys, framework candidate-label conventions, and
  trusted `PublicExact` disclosure assertion with separate reason disclosure.
- Derive typed viewer-specific offers through complete legality evaluation.
- Generate the exact TypeScript offer union and outer snapshot envelope.

### Stage 2: existing client integration

- Hydrate offer choices with existing `BoundMoveAction`/`TargetAction` objects.
- Add typed offer lookup to the generated renderer facade.
- Add a generic accessible offer surface using existing action components.
- Add board-native binding examples.
- Migrate Valentine Guard and public portions of Priest/Baron.

### Stage 3: evidence-driven extensions

Only after concrete games and privacy tests should later designs consider:

- public-superset projection;
- opaque viewer/version-bound bindings;
- multiple same-type offer instances;
- dependent multi-input candidate refinement;
- component and board-space sources;
- large-domain pagination;
- carefully scoped presentation continuity.

Each extension must preserve the one-legality-system and one-client-action-stack
invariants.

## Remaining open questions

1. Should offers be additional fields on the existing move-form bundle or a
   separately versioned sibling inside the same state response?
2. What exact API spelling records the trusted `PublicExact` declassification,
   its audience, and the independent reason-disclosure choice? What read-set
   evidence should warn reviewers without pretending to prove opaque code safe?
3. Should safely illegal public candidates default to disabled or omitted?
4. What stable generic error replaces raw `move.Legal()` text when a reason is
   not disclosure-safe?
5. What are the total candidate, evaluation, payload, and deadline limits for
   one projection?
6. How does the server prove or identify the settled fix-up frontier in the
   existing version-delivery pipeline?
7. Is an offer schema fingerprint distinct from the creator-input fingerprint,
   or is a small offer schema version sufficient?
8. Which public Valentine choices can migrate in v1 without depending on issue
   #693?
9. What is the narrowest future contract for `PublicSuperset` that does not turn
   proposal failures into another hidden-state oracle?

## Final rationale

The framework does not need another workflow engine, legality language, or
client action controller. It needs a typed discovery layer over the systems it
already has.

The resulting formulation is deliberately narrow:

> An action offer is a viewer-specific projection that decorates an existing
> typed move action with a safely presentable candidate universe and localized
> UI intent. Existing progression determines order, existing `move.Legal()`
> determines legality, existing disclosure determines what may be revealed,
> and existing client actions perform preview and submission.

That is enough to make Valentine-style committed choice sequences generic and
coherent without duplicating rules or elevating presentation into game state.
