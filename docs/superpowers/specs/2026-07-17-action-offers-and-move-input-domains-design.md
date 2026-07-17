# Action Offers and Executable Move-Input Domains

**Status:** Design draft for continued critique

**Date:** 2026-07-17

**Scope:** Boardgame engine move authoring, viewer-specific server projection, and
client interaction behavior. This document does not authorize implementation.

## Summary

Boardgame already has authoritative representations for game state, committed
moves, move ordering, and concrete move legality. It does not have a generic
way for a client to discover how a player should supply the creator-owned inputs
of a currently available move. Games therefore repeat rules in their renderers:
they rediscover that a target must be selected, determine which targets look
valid, and manually keep several committed decisions visually connected.

This design adds a derived, viewer-specific `ActionOffer[]` read model. An action
offer explains how one viewer can construct a move at one game version. It is
derived from existing authoritative state plus two additions to move authoring:

1. **Availability rules** describe when an unbound move should be offered to an
   actor. Every availability rule is also an authoritative legal precondition.
2. **Executable input domains** describe and validate the possible values of a
   creator-owned move field. Bounded domains can also enumerate choices for a
   generic client.

The same executable rule objects generate the offer and validate the eventual
proposal. `ActionOffer` is never game state, never a workflow cursor, and never
an alternate legality engine. It can be discarded and regenerated from the
authoritative version at any time.

The intended semantic layers are:

```text
Durable rule-level protocol
    Domain state and phases, only when the protocol matters to the game

Committed decisions
    Ordinary moves, progression, Legal(), history, and disclosure

Uncommitted composition
    A client-side draft of one move

Presentation read model
    Viewer-specific ActionOffer[] derived from all of the above
```

For Valentine, playing a Guard, selecting a player, and guessing a card remain
separate committed moves. `MoveSelectPlayer` advertises a player-index domain;
`MoveGuessCard` advertises an enum domain. The client presents successive offers
as one continuous card-resolution experience without persisting an
`Interaction` object or creating card-specific sub-phases.

## Problem

### The engine and renderer answer different questions today

The engine can answer whether this fully populated move is legal:

```go
MoveSelectPlayer{OtherPlayerIndex: 3}
```

The renderer additionally needs to know:

- whether selecting a player is an action the viewer can begin now;
- that `OtherPlayerIndex` denotes a player rather than an arbitrary integer;
- which players are candidates in the current sanitized state;
- how to label and contextualize the choice;
- whether it should replace the contents of an existing action surface;
- how to bind the selected value into a complete move proposal.

Without a framework contract, renderers answer those questions with game-specific
computed values and event handlers. Valentine exposes `NeedToSelectPlayer` and
`NeedToGuessCard`, while the corresponding moves independently check active
card type, selected-player state, target validity, protection, elimination, and
self-selection rules. The UI and server can drift.

### Default move instances conflate unavailable and incomplete

A default `MoveSelectPlayer` commonly contains an unset sentinel such as
`AdminPlayerIndex`. That is not a legal concrete move, but it does not imply
that selecting a player is unavailable. It means that the move is available
and still requires creator input.

The framework needs to distinguish:

```text
move type is unavailable
move type is available but has unbound required inputs
partially bound move has more inputs
fully bound move is legal
```

Calling full `Legal()` on a default instance cannot make these distinctions.

### Multi-step UI has three different underlying semantics

Experiences that look like multi-step dialogs are not one rules concept:

1. **Local composition of one move.** Choosing a checker and then its
   destination is normally one uncommitted `MovePiece{Source, Destination}`.
   Backtracking and cancellation are local editing.
2. **A sequence of committed decisions.** Playing a Guard, selecting its
   target, and guessing a card may need separate history, observation,
   disclosure, reactions, or animation boundaries. Each decision is a move.
3. **A durable game protocol.** An auction, reaction window, or simultaneous
   secret vote has rule-level identity, participants, completion conditions,
   and perhaps a timer. It belongs in domain state and sometimes a phase.

The design must support all three without forcing them into one universal
`Interaction` state machine.

## Goals

1. Make ordinary follow-up choices easy to author and render generically.
2. Keep committed, observable decisions as ordinary moves.
3. Use one executable definition for candidate generation and validation.
4. Separate move availability from missing creator input.
5. Reconstruct current actions after reconnect from authoritative state.
6. Support multiple simultaneous or alternative offers; do not assume one
   global active prompt.
7. Make offers viewer-specific and subject to the same disclosure discipline
   as moves and state.
8. Let game-specific board UI and a generic accessible UI bind the same action.
9. Preserve `Legal()` and proposal-time validation as the final authority.
10. Build on the existing creator move-input schema and declarative legality
    systems rather than adding a second workflow language.

## Non-goals

- Inferring an operational workflow cursor from `MoveProgressionGroup`.
- Making presentation continuity authoritative or persistent.
- Replacing domain state for auctions, reactions, simultaneous obligations, or
  other real game protocols.
- Making every possible domain generically enumerable.
- Allowing an offer to guarantee that a later proposal will succeed after the
  state version changes.
- Adding rollback semantics for already committed moves.
- Solving traffic-analysis secrecy. Viewer sanitization and concealment of game
  version activity are separate concerns.
- Requiring all games to use a generic modal or any particular layout.

## Design principles

### One authority per concern

| Concern | Authority |
| --- | --- |
| What happened | Committed moves |
| Durable facts and obligations | Game and player state |
| Permitted move ordering | Move progression |
| Whether a concrete proposal is legal | `Legal()` |
| What a viewer may know | Sanitization and move disclosure |
| Who may begin an unbound move now | Availability rules |
| What values an input may take | Executable input domains |
| How the viewer can act now | Derived `ActionOffer[]` |
| Where controls appear | Renderer |

### Derived, not persisted

An action offer is a cacheable read model for a viewer and state version. It has
no independent lifecycle. The framework does not commit “open interaction,”
“advance interaction,” or “close interaction” events merely to manage UI.

### Plural offers, not a singular current step

At one version a player may be able to draw a card, pass, or react. Several
players may each owe a secret choice. Optional and parallel progression also do
not imply a unique current node. The wire and client APIs therefore expose
`ActionOffer[]`.

### Presentation continuity is weak

An optional `continuityGroup` may tell the client that a new offer can reuse an
existing action surface. It is a presentation category, not an instance ID. It
must never determine legality, correlate responses, cancel work, or distinguish
two concurrent resolutions. The first implementation may omit it until a
concrete UI needs it.

## Authoritative rule model

### Four stages of move legality

Move authoring is decomposed into four stages:

1. **Availability:** field-independent rules for whether the actor may begin
   constructing this move now.
2. **Input domains:** rules for each creator-owned field, evaluated against the
   current state and any declared prerequisite input bindings.
3. **Cross-input constraints:** relationships among multiple populated inputs
   that are not naturally owned by one field domain.
4. **Final legality:** the complete existing legal chain, including
   `LegalCustom`, evaluated again when the proposal is submitted.

Every availability rule and input-domain validator contributes to final
legality automatically. Authors must not restate them in `LegalCustom`.

### Availability

Availability answers:

> Ignoring creator-owned fields that have not been supplied yet, may this actor
> begin constructing this move in this state?

Proposed authoring shape:

```go
auto.MustConfig(
    new(MoveSelectPlayer),
    moves.WithAvailability(
        valentine.ActiveCardRequiresTarget(),
        legal.PropEquals(
            "player.SelectedPlayer",
            "admin",
        ).WithMessage("valentine.already_selected"),
    ),
)
```

`WithAvailability` is not presentation metadata. Its rules are authoritative
legal preconditions and are included in the normal legal plan. The distinction
is that they promise not to depend on creator-owned fields that are still
unbound.

Existing behavior contributions such as current-player authorization and
in-progression checks should be classified as availability-capable where sound.
The engine must reject an availability rule whose declared read set includes a
creator-owned move field.

An opaque game-specific availability predicate is allowed when declarative
paths cannot express the rule, for example dynamic component values. It must be
a named, reusable server predicate and must still contribute to final legality.

### Executable input domains

An input domain owns the accepted values of one creator-controlled input. It is
both a validator and, when possible, a candidate or constraint provider.

Conceptually:

```go
type Domain interface {
    Field() string
    Dependencies() []string
    Describe(ctx DomainContext) (InputDescription, error)
    Validate(ctx DomainContext, value any) error
}
```

`Describe` returns one of three domain modes:

- **choices:** a bounded set of candidates such as players or enum values;
- **constraints:** a constructible range or schema such as an integer interval
  or bounded string;
- **binding:** the value is supplied by a specialized board surface and the
  generic client can only explain that requirement.

`Validate` is always required, including for non-enumerable and specialized
domains. Proposal-time validation never trusts the offered candidate list.

Candidate generation and validation must be two views of the same domain
object. It is invalid to configure one function for visible choices and an
unrelated function for authoritative membership.

### Input dependencies and atomic local wizards

A move may contain multiple creator inputs. A later input domain may depend on
earlier bindings, for example destination depends on source. Dependencies are
explicit and must form a directed acyclic graph validated at configuration
time:

```go
moves.WithInputs(
    input.Component("Source").From("player.Pieces"),
    input.Space("Destination").
        DependsOn("Source").
        Domain(game.LegalDestinationsFrom("Source")),
)
```

The client maintains a local partial binding and asks for or derives the next
input descriptions. Nothing commits until the complete move is proposed. This
is the framework seam for source/destination, card/rotation, and other atomic
move drafts. It is not a committed interaction sequence.

Version one should support only independent fields or a simple declared order
unless a concrete game requires general DAG recomputation.

### Cross-input constraints

Rules involving several populated fields can be attached to an input group:

```go
moves.WithInputConstraint(
    input.Fields("Space", "Orientation", "Resources"),
    game.ValidPlacement(),
)
```

These rules contribute to full legality and may participate in exact preview
once all dependencies are populated. They do not make an unbound move
unavailable.

### `LegalCustom` remains the escape hatch

`LegalCustom` remains authoritative and runs on the fully bound move. It is
appropriate for irreducible domain rules but should not normally contain:

- missing-value checks for declared required inputs;
- enum membership or integer-range checks;
- context-owned proposer checks;
- candidate membership already owned by an input domain;
- availability rules already declared through `WithAvailability`.

An action offer cannot generally project an opaque `LegalCustom` rule. For
bounded domains the server may exact-preview candidate bindings and filter any
that fail final legality. For unbounded domains, authors must expose any rule
needed to determine whether the move should be offered as an availability rule;
the proposal endpoint still reports final rejection.

## Proposed game-author API

The exact Go spelling remains open, but the semantic ownership should resemble:

```go
auto.MustConfig(
    new(MoveSelectPlayer),

    moves.WithAvailability(
        valentine.ActiveCardRequiresTarget(),
        legal.PropEquals("player.SelectedPlayer", "admin").
            WithMessage("valentine.already_selected"),
    ),

    moves.WithInputs(
        input.PlayerIndex("OtherPlayerIndex").
            Label("valentine.choose_player").
            Domain(valentine.ValidTargets()),
    ),

    moves.WithPresentation(
        presentation.Title("valentine.resolve_card"),
        presentation.Context(valentine.ActiveCardContext()),
        presentation.ContinuityGroup("card-resolution"),
    ),
)
```

`ValidTargets` expresses one cohesive game rule:

```go
func ValidTargets() input.PlayerDomain {
    return input.Players().
        Where(player.NotEliminated()).
        Where(player.NotProtected()).
        PreferOthers().
        AllowSelfOnlyWhenEmpty()
}
```

The Guard guess becomes:

```go
auto.MustConfig(
    new(MoveGuessCard),

    moves.WithAvailability(
        valentine.ActiveCardIs(cardGuard),
        valentine.SelectedTargetIsNotSelf(),
    ),

    moves.WithInputs(
        input.Enum("GuessedCard", cardEnum).
            Label("valentine.guess_card").
            Except(cardUnknown, cardGuard),
    ),

    moves.WithPresentation(
        presentation.Title("valentine.resolve_guard"),
        presentation.Context(valentine.SelectedPlayerContext()),
        presentation.ContinuityGroup("card-resolution"),
    ),
)
```

All labels and context must use localizable templates with viewer-safe
bindings. Literal strings above illustrate intent only.

### Relationship to the creator move-input contract

`boardgame.BuildMoveInputSchema` remains the authority for which persisted move
fields are creator-owned, server-defaulted, context-owned, or unsupported and
for how values are encoded. A domain may only attach to a creator-owned field
in that canonical schema. Configuration fails for:

- an unknown or context-owned field;
- a domain whose kind is incompatible with the field codec;
- two domains claiming the same field;
- a required creator field lacking an input description on an offerable move;
- cyclic or unknown input dependencies;
- an availability predicate that reads an unbound creator field.

Domains describe legal values; codecs describe representation and transport.
Neither replaces the other.

## Runtime `ActionOffer` projection

### Conceptual wire model

```ts
interface ActionOffer {
  readonly moveName: string;
  readonly basedOnStateVersion: number;
  readonly moveInputSchemaFingerprint: string;
  readonly presentation?: OfferPresentation;
  readonly inputs: readonly InputOffer[];
}

interface OfferPresentation {
  readonly title?: LocalizedText;
  readonly context?: LocalizedText;
  readonly continuityGroup?: string;
  readonly urgency?: 'normal' | 'required' | 'reaction';
}

type InputOffer =
  | ChoiceInputOffer
  | ConstraintInputOffer
  | BindingInputOffer;

interface ChoiceInputOffer {
  readonly field: string;
  readonly kind: 'player' | 'enum' | 'component' | 'space' | 'custom';
  readonly mode: 'choices';
  readonly choices: readonly Choice[];
}

interface Choice {
  readonly id: string;
  readonly label: LocalizedText;
  readonly binding: Readonly<Record<string, unknown>>;
}
```

`binding` is expressed in the generated creator-input representation, not raw
form strings. A choice may eventually bind several fields, although version one
should keep one field per choice. Privacy-sensitive or computationally complex
choices may later use opaque, state-version-bound binding tokens; tokens are
deferred until there is a demonstrated need because they complicate typed move
proposal and debugging.

The offer does not need a durable interaction or offer ID. `moveName`, the
binding, and `basedOnStateVersion` are sufficient for proposal. If future
transport requires an ephemeral correlation ID, it must not acquire game
semantics.

### Projection algorithm

For one authenticated viewer at one sanitized game version:

1. Determine move types structurally admitted by existing progression.
2. Apply actor/context rules and declared availability rules for the viewer.
3. Read the canonical creator-input schema.
4. Ask each input domain for a viewer-safe description or candidates.
5. For bounded candidates, optionally bind complete candidates and evaluate
   exact legality in a server batch, filtering failures.
6. Apply offer and candidate disclosure policies.
7. Emit zero or more offers tied to the source state version and schema
   fingerprint.

The exact proposal endpoint always reconstructs the move, applies context and
server defaults, checks the fingerprint and state version policy, and runs the
complete authoritative legal chain. The offer is advisory and can become stale.

The projection should be server-owned initially. Client-side evaluation can
later optimize latency for predicates proven safe on sanitized state, but it
must produce the same model and cannot become the sole authority.

### Candidate cardinality and cost

Domains declare their cost and cardinality behavior:

- small bounded domains can be enumerated and exact-previewed eagerly;
- large bounded domains can be paginated or board-bound;
- combinatorial and unbounded domains expose constraints or specialized
  bindings instead of enumerating every complete move;
- expensive candidate providers may require game-authored board binding.

The server enforces candidate and payload limits. A domain may not accidentally
materialize every path, card subset, or word in a large search space.

## Client behavior

### Headless controller

The framework client should expose a headless action controller rather than
owning DOM or focus. It maintains:

- current plural offers;
- a local partial binding for the selected offer;
- exact-preview state for a complete candidate;
- stale/rejected/resolving status;
- actions to bind, clear, submit, and minimize.

Specialized renderers and the generic fallback use the same controller.

### Generic accessible surface

The default UI is a non-modal prompt region or side sheet with a semantic list
of choices. A focus-trapped modal is valid only when all possible choices are
inside it. If valid targets are board elements outside the prompt, keyboard and
screen-reader users must still be able to reach an equivalent semantic target
list.

Before a move is committed, Escape may cancel or clear a local draft. After a
step in a committed sequence, Escape means minimize; it cannot imply that the
committed move was undone.

### Board-native binding

A renderer can bind an offer field to players, components, or spaces already on
the board. The offer remains the semantic source for highlighting and proposal:

```ts
const offer = actions.firstForMove('Select Player');

renderPlayer(player, {
  selectable: offer?.accepts('OtherPlayerIndex', player.index),
  onSelect: () => offer?.bind('OtherPlayerIndex', player.index),
});
```

The generic surface remains available as a fallback and accessibility path.

### Continuity and animation

When a committed proposal is accepted, the client may keep the action surface
in a resolving state while the accepted move and fix-ups animate. On the next
snapshot it replaces the old offer with newly derived offers. A matching
`continuityGroup` permits visual reuse but conveys no rules relationship.

Animation state is client-local. It is not part of a server workflow cursor.

## Valentine walkthrough

The authoritative sequence remains:

```text
MovePlayCard
→ MoveSelectPlayer, when the active card requires a target
→ MoveGuessCard, when the active card is Guard and the target is not self
→ MoveActivateGuard or another activation fix-up
```

1. The player locally chooses a card index and commits `MovePlayCard`.
2. The resulting state and existing progression admit `MoveSelectPlayer`.
3. Its availability rules pass for the current actor.
4. `ValidTargets` produces unprotected, non-eliminated candidates and applies
   Valentine's self-only-when-no-alternative policy.
5. The viewer receives a `SelectPlayer` offer and the client displays or
   highlights those candidates.
6. Selecting one binds and commits a real `MoveSelectPlayer`.
7. The next version admits `MoveGuessCard`; its enum domain excludes Unknown
   and Guard.
8. The client may retain the same card-resolution surface.
9. Refresh at either point regenerates the current offer from authoritative
   state. No interaction cursor is reconstructed.

Valentine should move card targeting policy toward named data or helpers, for
example `NoTarget`, `TargetOther`, and `TargetOtherOrSelfWhenNoAlternative`, so
`NeedToSelectPlayer`, selection legality, activation legality, and renderer
behavior no longer contain independent card-type switches.

## Hard journey: simultaneous secret choices

Suppose every player secretly selects a card and all choices are revealed once
everyone has submitted.

The game records the durable obligation and private choice in sanitized player
or game state. Each submission is a real, private move:

```go
type MoveSubmitChoice struct {
    moves.CurrentPlayer
    CardIndex int
}
```

For a player who still owes a choice, the server projects one private
`SubmitChoice` offer whose component domain contains only choices safe for that
viewer. After submission, that player's offer disappears. Other players receive
neither the offer nor candidate metadata unless the rules intentionally expose
waiting status.

When all obligations are satisfied, a fix-up commits a public resolution move.
The original submission moves may remain permanently secret, including their
proposer and occurrence, subject to the move-disclosure design tracked in issue
#693. The offer system does not require retrospective mutation of move history.

This journey needs durable “has submitted” state because reconnect and
completion depend on it. It does not need a generic persisted interaction
cursor; offers are projections of the durable domain state.

## Design-space validation

| Journey | Durable semantics | Offer/input treatment |
| --- | --- | --- |
| Valentine Guard | Several committed moves | Player then enum offers with cosmetic continuity |
| Valentine Priest/Baron | Committed choice plus private information | Viewer-specific player/confirm offers and disclosure |
| Checkers source/destination | One atomic move | Local draft with dependent component/space domains |
| Darwin card/species | One atomic move | Local draft with dependent component choices |
| Memory repeated reveals | Repeated committed move | New offer derived after each reveal; no occurrence cursor |
| Werewolf voting | Durable per-player obligation | Private plural offers projected from player state |
| Scrabble placement | One combinatorial atomic move | Specialized board binding plus final exact validation |
| Ticket to Ride two draws | State changes after first committed draw | Newly derived second-draw offers |
| Catan robber | Durable discards, then serial committed choices | Per-player offers followed by space/player offers |
| Coup reaction window | Durable timed/priority protocol | Domain state owns window; offers project legal responses |
| Secret choose/reveal | Durable private submissions | Per-viewer private offers, then public fix-up |

No one abstraction owns all of these journeys. The common UI boundary is the
viewer-specific action offer; the authoritative representation remains
appropriate to each journey.

## Privacy and disclosure

Action offers are generated after authenticating the viewer and must be treated
as potentially secret. Sensitive data includes:

- whether an offer exists;
- the actor who received it;
- title and contextual text;
- candidate identities and counts;
- exclusion or failure reasons;
- continuity metadata;
- waiting and completion status.

An offer must never be generated globally and merely hidden in the DOM. Server
projection and serialization are viewer-specific. Candidate providers operate
on authoritative state but return only viewer-safe descriptions.

Move disclosure and offer disclosure should share policy vocabulary where
possible, but they are not identical. A player may see an offer before a move
exists; a committed move may have a different audience from the offer that
created it. Issue #693 remains the prerequisite for fully private move history.

Even perfect payload sanitization does not hide websocket activity or global
version advancement. Games requiring traffic-analysis resistance need a
separate transport design.

## Why not the alternatives?

### Universal persisted `Interaction`

Wrapping progression in an interaction cursor duplicates ordering and legality.
The current progression API recognizes move-name histories; optional, repeated,
and parallel groups do not yield a canonical current node, actor, or instance.
Auctions and reaction windows should use explicit domain state, but ordinary
follow-up prompts should not create another state machine.

### Sub-phases for every prompt

Phases are appropriate for low-cardinality, rule-significant modes. Encoding
every card target and guess as a phase creates a phase taxonomy of screen
choreography, duplicates progression, and still does not explain which move
field represents a player or which players are legal candidates.

### One atomic move for every interaction

Atomic local drafts are correct when nothing may observe or react between
fields. They are incorrect when intermediate decisions must be durable,
visible, interruptible, animated separately, or disclosed to different
audiences.

### Game-authored `OffersFor`

A manual `OffersFor(viewer, state)` API is flexible but invites each game to
duplicate progression and legality. Games may author domains and presentation
context; the framework must derive the offer itself from those executable
descriptions and the authoritative legal chain.

### Static prompt tags alone

A tag can say that a field represents a player. It cannot safely determine
current candidates, exact legality, viewer disclosure, stale-state behavior, or
multiple simultaneous actions. Static metadata is the authoring source;
`ActionOffer[]` is the necessary runtime projection.

## Invariants

1. Deleting all action offers cannot change game correctness.
2. An offer never makes an illegal move legal.
3. Every offered binding is validated again at proposal time.
4. Every availability rule and domain validator participates in full legality.
5. No candidate-generation rule has an independent authoritative membership
   implementation.
6. Offers are tied to a state version and move-input schema fingerprint.
7. Offers are generated and sanitized per viewer.
8. Continuity metadata has no game semantics.
9. Committed decisions remain moves; partial bindings are not moves.
10. Durable protocols remain domain state or phases, not inferred UI state.

## Failure behavior

- Configuration fails for invalid field ownership, incompatible domain types,
  duplicate domains, cyclic dependencies, or field-dependent availability.
- Offer projection fails closed for stale schema fingerprints and unsafe domain
  output.
- A candidate provider error suppresses the affected offer and produces a
  diagnosable server error; it must not emit an unvalidated candidate set.
- A zero-candidate bounded domain normally suppresses the offer. Required
  protocols that reach zero candidates must encode their game-specific fallback
  or resolution in rules/state rather than trapping the client in an empty UI.
- Proposal against a changed state receives the normal stale or illegal result
  and refreshes offers from the newest version.
- A generic client encountering a binding-only domain renders an explanatory
  fallback and lets the renderer supply the specialized binding.

## Testing strategy

### Configuration tests

- creator-input ownership and codec compatibility;
- missing, duplicate, and unknown input domains;
- dependency DAG validation;
- rejection of field-dependent availability predicates;
- presentation template registration and safe bindings.

### Rule equivalence tests

- every domain candidate passes that domain's validator;
- representative rejected values fail both offer membership and proposal;
- availability rules contribute identically to offer projection and full legal
  evaluation;
- `LegalCustom` can reject an exact-previewed candidate without being bypassed.

### Per-viewer matrices

- current actor, other player, observer, and admin receive only allowed offers;
- hidden offer existence and candidate counts do not leak;
- sanitized context never references hidden component values;
- simultaneous secret choices expose only each viewer's own obligation.

### Client tests

- plural offers and alternatives;
- local partial binding, clearing, and dependent-domain recomputation;
- exact candidate preview and stale rejection;
- refresh reconstruction;
- continuity across committed snapshots;
- Escape cancels a local draft but only minimizes after commit;
- keyboard and screen-reader access to board-bound choices;
- animation/fix-up snapshots do not flash transient prompts.

### Golden game journeys

Valentine Guard is the first migration target. Checkers or Darwin validates a
multi-field local draft. A simultaneous secret-choice example validates
viewer-specific projection and disclosure. A reaction-window example remains a
later proof that durable domain protocols compose with offers without becoming
generic interactions.

## Incremental delivery

### Stage 0: prerequisite audit

- Verify that exact bound-move preview bypasses no legal checks and that a
  default illegal field value cannot permanently disable a legal non-default
  candidate.
- Confirm how current-player and in-progression atoms can be classified for
  availability without changing the frozen legal chain.
- Coordinate offer disclosure with the move-sanitization work in issue #693.

### Stage 1: bounded single-input offers

- Add availability classification whose rules also contribute to legality.
- Add executable player-index and enum domains.
- Add boot validation against `BuildMoveInputSchema`.
- Project server-owned, viewer-specific `ActionOffer[]` for bounded inputs.
- Exact-preview complete candidates before emitting them.

### Stage 2: headless client and generic surface

- Generate exact TypeScript offer and binding types.
- Add the headless action controller.
- Add a non-modal generic prompt surface and semantic candidate lists.
- Add board-native binding hooks.
- Migrate Valentine Guard, Priest/Baron, and return-card choices.

### Stage 3: atomic multi-input drafts

- Add ordered/dependent input domains.
- Support range/schema input descriptions.
- Validate with Checkers, Darwin, or a placement-heavy game.

### Stage 4: advanced domains and protocols

- Add component/space domains, pagination, and specialized binding contracts.
- Consider opaque state-version-bound binding tokens if privacy or computation
  demonstrates the need.
- Validate simultaneous secret choices and a durable reaction window.

## Open questions for critique

1. Should availability be a new explicit `WithAvailability` category, or a
   flagged subset of the existing `WithLegalPreconditions` plan?
2. Can existing behavior atoms such as current-player and in-progression be
   classified as availability-capable without creating a second evaluation
   order?
3. Should v1 enumerate candidates on the server, or emit domains and use the
   existing exact batch-preview endpoint from the client for safe cases?
4. What is the smallest domain interface that guarantees candidate/validation
   single-sourcing without forcing every domain to enumerate?
5. Does a choice bind exactly one field in v1, or should multi-field opaque
   bindings be part of the initial contract?
6. Should `continuityGroup` ship in v1, be inferred by the client, or wait for a
   demonstrated UX discontinuity?
7. How should offer context use history when the relevant committed move is
   private, sanitized, or no longer in the immediately available client bundle?
8. How should required-versus-optional urgency be derived without inventing a
   singular progression cursor?
9. What server cost and payload budgets should domains declare and enforce?
10. Which parts of issue #693 must land before a private offer API is safe to
    expose beyond games whose moves are already public?

## Decision

Adopt `ActionOffer[]` as the generic client boundary and executable move-input
domains as its authoring foundation. Do not introduce a universal persisted
`Interaction`, derive a workflow cursor from progression, or encode ordinary
prompts as sub-phases.

Use:

- local drafts for uncommitted composition of one move;
- ordinary committed moves plus offers for sequences of player decisions;
- explicit game state and phases for genuine rule-level protocols;
- per-viewer offers as the replaceable presentation projection across all
  three.

This is the smallest design that improves authoring and UI together without
duplicating the engine's existing sources of truth.
