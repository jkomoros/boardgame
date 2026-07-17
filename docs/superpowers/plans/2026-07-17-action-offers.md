# Action Offers — Base-Primitive Implementation Plan

**Status:** Active

**Date:** 2026-07-17

**Design:** `docs/superpowers/specs/2026-07-17-action-offers-and-move-input-domains-design.md`

## Objective

Deliver viewer-specific action offers by extending the framework's existing
legality, creator-input, sanitization, and typed-action primitives. Each commit
must be independently useful and testable before Valentine adopts it. No commit
may introduce a second legality language, action runtime, workflow cursor, or
offer-only workaround for a broken lower-level primitive.

## Implementation order

```text
Exact-bound action legality provenance
    ↓
Canonical creator-input presentation metadata
    ↓
Internal disclosure-aware candidate projection
    ↓
Proven proposal-accepting frontier and bundle delivery
    ↓
Existing-action hydration, typed lookup, accessible UI
    ↓
Valentine adoption in paired worktrees
```

The ordering is binding. In particular, actionable offer wire delivery cannot
precede proof that the source snapshot is a settled proposal-accepting frontier.

## Commit 1: exact-bound action legality provenance

### Problem

`MoveActionImplementation.availability` currently applies the default move
form's `LegalForAnyone` to every action and its `LegalForPlayer` when the
baseline perspective matches. A complete `.with(input)` or `.targets(...)`
candidate can therefore remain blocked even after exact preview proves that
binding legal. The form describes an incomplete default move; it is not a valid
authority for a different complete binding.

### Primitive

Make legality provenance explicit in the existing action implementation:

- unbound builders and zero-input actions use form-baseline legality;
- complete bound actions use exact preview legality;
- all non-legality gates remain shared.

This is not an offer flag. It fixes every current parameterized action.

### Invariants

1. A complete bound action ignores default-form `LegalForAnyone` and
   `LegalForPlayer`.
2. It remains blocked until exact preview succeeds.
3. Exact-preview illegality still blocks activation and proposal.
4. Zero-input actions retain both baseline gates.
5. Invalid input, stale snapshots, schema mismatch, animation, transport,
   submission, and global submission gates are unchanged.
6. Proposal-time server legality remains authoritative.

### Tests

- Direct `.with(...)`: false/false baseline plus exact legal preview enables
  and proposes.
- Direct `.with(...)`: false/false baseline plus exact illegal preview remains
  blocked with `preview-illegal`.
- Zero-input action: false structural baseline remains blocked.
- `TargetAction`: false/false default baseline plus batch-hydrated legal and
  illegal candidates yields the exact per-candidate result.

### Verification

Run the focused TypeScript action and target-action tests, strict typecheck, and
the renderer fixture tests affected by action availability.

## Commit 2: canonical input-presentation metadata

### Primitive

Extend the existing creator-input configuration and
`BuildMoveInputSchema` output with optional presentation metadata on a required
field. Do not add a parallel offer registry or fingerprint.

Version one supports only sealed framework descriptors:

- player-index → all valid player indexes in canonical order;
- enum → the configured enum's canonical values.

Metadata includes:

- structured prompt key;
- optional structured title key;
- source kind;
- trusted `PublicExact` assertion;
- separate disabled-reason disclosure, defaulting to generic.

### Rails

- Explicit opt-in only.
- Exactly one required creator-owned field.
- Only player-index and enum codecs.
- Optional schema fields use `omitempty`; games without presentation metadata
  retain byte-identical schema JSON and fingerprints.
- Boot rejects unknown/context-owned/defaulted/unsupported fields, codec/source
  mismatches, duplicate declarations, and unsupported disclosure modes.
- Metadata does not evaluate legality or produce runtime offers.

### Tests

- Boot/configuration validation matrix.
- Player and enum metadata extraction.
- Clone/deterministic ordering behavior.
- Legacy schema JSON and fingerprint golden unchanged.
- Opted-in fingerprint changes deterministically.

## Commit 3: internal candidate projector

### Primitive

Add a pure, internal server projection function operating on an immutable
version-pinned state, viewer/proposer, configured move, and canonical input
metadata. It returns no public wire contract yet.

For each configured move:

1. Check canonical-name visibility with `MoveNameVisibleToPlayer` for the viewer
   as hypothetical proposer.
2. Enumerate the sealed finite source in canonical order.
3. Bind each value through the canonical creator-input codec.
4. Invoke the complete `move.Legal(state, proposer)` exactly once.
5. Return available/disabled internal candidates with generic disabled reasons
   unless reason disclosure was separately authorized.
6. Produce an offer only if at least one candidate is legal.

### Rails

- Do not reuse the public single or batch preview endpoints. They expose exact
  booleans and raw legal errors without offer disclosure semantics.
- Keep `privateMoveAvailable` unchanged for legacy forms and generic private
  proposal/preview errors. The projector bypasses default-instance discovery;
  it does not call or replace that helper.
- Hard-bound candidates, evaluations, and result bytes.
- No live `CurrentState`, history beyond the pinned version, clock, randomness,
  I/O, custom providers, or map-order output.
- Admin/observer receive no player offers by default.

### Tests

- Player and enum canonical enumeration.
- Full-`Legal` parity including `LegalCustom`.
- Default sentinel illegal while another binding is legal.
- Private canonical name visible/hidden viewer matrix using landed move
  sanitization.
- `PublicExact` status with generic reason; reason opt-in separately.
- Hidden-equivalent-state noninterference fixture.
- Zero legal candidates suppresses the offer while diagnostics distinguish
  projection failure.
- Deterministic order and hard limits.

## Commit 4: proposal-accepting frontier and bundle delivery

### Problem

Each move in a recursive fix-up chain is stored as its own version. Websocket
notification occurs after closure, but initial/info reads may race with an
intermediate stored version. A before/after version equality check does not
prove that a snapshot accepts player proposals.

### Primitive

Prove an existing serialization guarantee or add an explicit server-visible
proposal-frontier/suppression primitive. Offers are computed and attached only
for the current settled state at which the game loop can accept the next player
proposal.

### Rails

- Historical and intermediate bundles contain no actionable offers.
- State version in the existing bundle is the offer version; do not repeat a
  version or fingerprint on every offer.
- Every eventual proposal retains `ExpectedVersion`.

### Tests

- Concurrent bundle request during a multi-fix-up chain never receives offers
  for an intermediate version.
- Final bundle receives offers exactly once for the settled frontier.
- Initial load, refresh, websocket replay, and historical animation paths.

## Commit 5: existing-action hydration and generic UI

### Primitive

Parse the viewer-specific offer projection and hydrate the existing generated
`MoveActionBuilder`, `BoundMoveAction`, and `TargetAction` graph. Do not add a
second preview, submission, cache, animation, or error state machine.

Generate an exact game-specific offer union and typed lookup:

```ts
offers.get(MoveNames.SelectPlayer)?.choices("OtherPlayerIndex")
```

Each disclosed candidate retains a bound action whose exact preview is already
hydrated legal or illegal.

### UI

- Add a generic semantic candidate list using existing action components.
- Board-native UI consumes the same candidate actions.
- Framework-owned player and enum label resolvers provide localized accessible
  labels.
- `ClientMove.AnimationKey` remains animation-only and is never proposal
  identity.

### Tests

- Exact generated move/field/value types.
- Parser validation and stale snapshot rejection.
- Existing action identity, gates, caching, and telemetry preserved.
- Available and disabled candidate parity with `TargetAction`.
- Keyboard and screen-reader semantic fallback.
- Board-native and generic controls submit the same bound action.

## Commit 6+: Valentine adoption

Create paired BOARDGAME/GAMES worktrees only after the generic surface is
complete. Migrate public-exact Valentine choices first:

- `MoveSelectPlayer` player input;
- `MoveGuessCard` enum input;
- public portions of Priest/Baron/return-card behavior where disclosure is
  proven.

Delete renderer-authored candidate collections, `NeedToSelectPlayer`,
`NeedToGuessCard`, and duplicated renderer legality. Preserve committed move
boundaries, progression, activation fix-ups, board-native controls, and full
proposal legality.

Private offer status, public-superset projection, opaque bindings, multiple
same-type instances, dependent multi-input refinement, and presentation
continuity remain later evidence-driven designs.

## Stop conditions

Stop rather than widening scope when:

- an actionable offer would be emitted from an unproven intermediate snapshot;
- exact candidate status is not safely `PublicExact`;
- implementation would duplicate a `Legal` predicate in presentation metadata;
- implementation would add an offer-only action state machine;
- a game requires unbounded or dynamically provided candidates;
- a private move requires traffic-analysis secrecy;
- a required semantic discriminator is absent from durable state or move input.

## Progress

- [x] Commit 1: exact-bound action legality provenance
- [ ] Commit 2: canonical input-presentation metadata
- [ ] Commit 3: internal candidate projector
- [ ] Commit 4: proposal-accepting frontier and bundle delivery
- [ ] Commit 5: existing-action hydration and generic UI
- [ ] Commit 6+: Valentine adoption
