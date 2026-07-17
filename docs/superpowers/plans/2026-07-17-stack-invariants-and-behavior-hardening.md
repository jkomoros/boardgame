# Stack Ownership Invariants and Behavior Hardening

**Issues:** [#751](https://github.com/jkomoros/boardgame/issues/751), [#793](https://github.com/jkomoros/boardgame/issues/793)

**Branches:** paired `stack-invariants-behaviors` worktrees for `boardgame` and `games`

## Outcome

This work has two deliberately separate outcomes:

1. Make the documented component-conservation invariant true at every persistence boundary and reject mutations through stacks that are not unique physical owners in the current state.
2. Finish issue #793 as a hardening, integration, and adoption pass. The six behaviors and two companion moves already landed in commit `086d44c6`; they must not be reimplemented or inaccurately presented as new work.

The implementation will preserve the creator-friendly common case:

```go
type gameState struct {
	base.SubState
	behaviors.DrawDiscardPair `draw:"Draw" discard:"Discard"`
	behaviors.FaceUpMarket     `source:"Deck" display:"Market"`

	Draw    boardgame.Stack `stack:"cards" sanitize:"len"`
	Discard boardgame.Stack `stack:"cards"`
	Deck    boardgame.Stack `stack:"market-cards" sanitize:"len"`
	Market  boardgame.Stack `sizedstack:"market-cards,4"`
}

auto.MustConfig(new(moves.ShuffleDiscardIntoDraw))
auto.MustConfig(new(moves.ReplenishMarket))
```

No custom methods are required. Bounded market capacity is the default target, so the common case does not repeat the number `4` in two places.

## Findings that drive the design

The obvious form of #751—moving into a fresh `Deck.NewStack`—is already rejected because detached stacks have no state pointer. The actual holes are broader:

- Setup accepts any mutable stack returned by `DistributeComponentToStarterStack` and inserts directly. A throwaway stack can swallow a component before the first save.
- Stacks from two games managed by one manager share deck pointers. Cross-state moves therefore pass the existing same-deck check.
- A live stack can be captured, replaced through `ConfigureStackProp` or direct field assignment, retain its state pointer, and remain mutable after it is no longer persisted.
- A merged property currently state-enables its private concrete backing stacks. Undeclared backing stacks can be mutated, but merged-stack unmarshal does not restore their contents.
- `validateBeforeSave` checks player indexes and merged shape, but not component uniqueness or completeness. The component index silently overwrites duplicate locations.
- Initial setup and state inflation do not run a conservation audit.

Issue #793 has the opposite shape: the main types exist, but important quality work is missing:

- `DrawDiscardPair` and `FaceUpMarket` accept structurally impossible configurations.
- Their companion moves have no focused real-state integration tests.
- `ReplenishMarket` fills sized stacks from the last slot instead of the next/first empty slot.
- A bounded display repeats its capacity in a mandatory `size` tag, which can drift.
- Named multiple markets work, but named multiple draw/discard pairs do not.
- The companion-move docs link to provider interfaces that do not exist.
- `metaltrader` still spells out the exact two-market replenishment boilerplate the behavior is intended to replace.

## Compatibility decisions

These are intentional correctness breaks:

- A component may only be inserted into or moved between stacks that are unique attached owners in the exact same state.
- Returning an undeclared stack from `DistributeComponentToStarterStack` fails game creation.
- A physical stack cannot own one persisted collection from two property/board locations.
- A merged view is not an owner. Every concrete backing stack must be independently declared as an owning property or board space.
- Stack and board `Configure*` calls become construction-only and reject replacement of a non-nil value. Direct Go field replacement is still possible, but is caught by attachment checks and the save audit.
- Corrupt stored states with missing or duplicate components fail loudly on load.

These APIs remain source-compatible:

- `FaceUpMarket` keeps `size`, `SetDisplaySize`, and `DisplaySize`. An explicit size remains useful for an unbounded or intentionally partially filled display. If omitted on a bounded display, `DisplaySize` is inferred from `MaxSize`.
- Existing single anonymous `DrawDiscardPair` and `FaceUpMarket` discovery remains zero-config.
- Existing mutable stack getters remain for now. Fixing const-correctness for every `Connectable` behavior is a framework-wide source-breaking project and is out of scope.
- Team, Score, Elimination, and MoveBudget retain their current interfaces. In particular, this tranche will not silently change `ConsumeMove` or `ResetMovesTo` signatures.

## Phase 0: Baselines and exploit tests

Before production code, add failing tests that prove the real holes rather than the already-blocked fresh-stack move case.

### #751 red tests

- Setup delegate returns a detached growable or sized stack; game creation fails before storage is written.
- Two real games under one manager attempt a cross-state transfer; both source and destination remain unchanged.
- A captured live stack is replaced in its property and then used as a destination; mutation fails.
- A merged view over declared owner stacks is accepted; a merged view over hidden backing stacks is rejected.
- One physical stack aliased at two properties, and at a property plus board space, is rejected.
- Missing, duplicate, wrong-deck-at-setup, invalid-index, and unsanitized generic components fail conservation.
- Board spaces, player stacks, and dynamic-component-value stacks participate in ownership and conservation.
- Failed initial or move-time validation does not write storage or advance the version.
- Valid loaded states and state copies still work; corrupt stored records fail during inflation; sanitized copies remain readable and are not subjected to unsanitized conservation rules.

### #793 red tests

- Same-stack, different-deck, detached, foreign-state, over-capacity, and missing/unconnected behavior configurations fail with actionable errors.
- A sized market fills left-to-right/first-hole order; a growable market appends.
- Replenishment applies one component per FixUp, repeats until full, and settles cleanly when its source is exhausted.
- Reshuffle preflight catches capacity and constraint rejection without mutation; success transfers every card and increments the draw stack's shuffle count exactly once.
- Anonymous and named behavior lookup survive copy, storage round-trip, and sanitized-state reads.
- Bad named fields, wrong field types, inaccessible reflection values, and non-string custom configuration return boot errors rather than panic.

## Phase 1: Central physical-stack traversal

Implement one deterministic internal walker over a state. It will be the single definition of persisted physical ownership.

It must visit:

- mutable concrete stack properties on game state, every player state, and every dynamic component value;
- every concrete board space in those same groups;
- concrete stacks exposed through an immutable property interface when they persist contents;
- merged views separately, without counting the view as another component location.

It must emit stable canonical paths such as:

- `Game.Draw`
- `Players[1].Hand`
- `Game.Board[3]`
- `DynamicComponentValues[cards][12].Tokens`

Property names, deck names, and map-backed groups will be sorted before diagnostics. The walker will diagnose nil properties, duplicate physical owners, invalid merged topology, and hidden merged backing stacks.

Reuse this traversal for attachment, integrity validation, component-index construction, and copied-stack lookup. Extend `findCorrespondingStack` to support board spaces.

## Phase 2: One-time attachment identity

Add private attachment metadata to concrete growable and sized stacks:

- exact `*state` identity;
- canonical owning location;
- an attached/owner marker that cannot be set through the public API.

After state construction has connected all substates, attach every physical owner discovered by the walker exactly once.

Rules:

- A stack may attach to one state and one location only.
- Reattaching a stack to a different state or location is an error.
- Merged stacks are views and never receive owner identity.
- Merged backing stacks must already be attached through declared physical locations.
- Each copied or inflated state receives fresh attachments to its own stacks. Attachment metadata is never copied from the source state.
- Sanitized stacks remain attached to their sanitized state but immutable.

Attachment validation will re-check that the current property at the canonical location still points to the same stack. This makes a captured replaced stack stale immediately; a non-nil state pointer or stale cached membership is not sufficient.

`ConfigureStackProp` and `ConfigureBoardProp` in generated, default, and generic readers will reject replacing a non-nil property. The error will say that configuration is construction-only and that callers should mutate the existing stack's contents instead.

## Phase 3: Mutation-boundary enforcement

Centralize a private mutation precondition used by every stack mutator:

- stack is non-nil;
- stack has a state;
- state is modifiable and unsanitized;
- stack has a unique current attachment matching its canonical property/board location.

Movement additionally requires source and destination to belong to the exact same `*state`, before capacity, constraints, or mutation. `MayMoveTo`, slot variants, `MayMoveAllTo`, and actual moves must agree on these ownership rules.

Errors will distinguish detached, stale/replaced, hidden merged backing, aliased, and foreign-state stacks. Where available, include canonical path, deck, game ID, and state version.

Add an explicit nil-destination check before dereferencing the destination.

Rewrite slide-to-first/last in place rather than creating a privately state-enabled scratch stack:

- growable stacks splice the selected index to the requested extreme;
- sized stacks clear the source slot and fill the correct extreme empty slot;
- the component index is updated once;
- failures leave the stack unchanged.

Do not add a public transient-stack escape hatch.

## Phase 4: Conservation at lifecycle boundaries

For every unsanitized state, audit all physical owners and require:

- each owner belongs to this exact state and one canonical location;
- every owner deck belongs to the manager chest;
- every stored component index is in range and is not a generic/hidden sentinel;
- each `(deck, deckIndex)` appears exactly once;
- every non-generic component in every chest deck appears somewhere;
- merged views add no duplicate locations and every backing stack is a declared owner of the same state and deck.

Diagnostics for a duplicate name both locations. Diagnostics for a missing component name its deck and index.

Run the audit:

1. after `FinishSetUp`, before the initial state is persisted;
2. from `validateBeforeSave` for every committed move;
3. after `stateFromRecord` inflates an unsanitized stored state.

Do not apply exact-conservation checks to sanitized states, whose hidden/generic representations intentionally omit identity.

Keep operation-time attachment checks even with this backstop: invalid actions should fail where attempted, while lifecycle validation protects against direct field assignment, malformed readers, corrupt storage, and internal framework bugs.

Add a focused benchmark with a large deck and sized board. Record the result in the implementation commit. Prefer clarity until measurement demonstrates the save-time O(components + slots) audit is material compared with serialization/storage.

## Phase 5: Bulk-operation atomicity and adjacent stack robustness

`MoveAllTo` currently mutates incrementally. A later constraint failure can leave an in-memory caller with a partial transfer even though the game pipeline normally discards a failed move copy.

Make the direct stack method atomic:

- snapshot both concrete stack payloads and mutation metadata needed to restore their exact observable state;
- evaluate each constraint only once while performing the transfer;
- on any error, restore source, destination, component-index locations, IDs/last-seen data, and shuffle-related metadata;
- return the original cause with context.

Test rejection on the first, middle, and final component, including sized gaps and board locations. `MayMoveAllTo` stays copy-based and must locate board-space stacks correctly.

Also fix the adjacent board bounds bug: `SpaceAt(Len())` and `ImmutableSpaceAt(Len())` return nil rather than panic.

This phase is a separate commit from the core ownership/conservation work so it can be reviewed and bisected independently.

## Phase 6: Harden the stack-backed behaviors

### Shared structural validation

Add an internal behavior helper that validates a source/destination pair:

- both connected and non-nil;
- distinct physical stacks;
- same deck;
- attached to the example state as declared owners;
- same exact state;
- behavior is connected to game state when the companion move discovers it there.

`DrawDiscardPair.ValidConfiguration` and `FaceUpMarket.ValidConfiguration` will use it. Errors identify the behavior and semantic role (`draw`, `discard`, `source`, `display`).

### FaceUpMarket target semantics

- When explicit `size`/`SetDisplaySize` is positive, use it.
- Otherwise, if the display is bounded, infer `DisplaySize()` from `display.MaxSize()`.
- If the display is unbounded and has no explicit size, fail configuration with exact guidance.
- Reject explicit targets larger than bounded capacity.
- Keep existing API names to avoid gratuitous source churn.
- Use `MoveToNextSlot` so a sized market fills its first empty slot and a growable market appends.
- In `Legal`, preflight the exact next source component against the display, including its target slot and constraints.

### DrawDiscardPair companion

- In `Legal`, after `NeedsReshuffle`, call the transfer preflight so predictable capacity/constraint failures are reported before `Apply`.
- Preserve the one-shuffle-after-complete-transfer behavior.
- Wrap underlying errors with `%w`.

### Symmetric named lookup

Add `WithDrawDiscardPairField("TrainCards")`, symmetric with `WithMarketField`, so games with multiple decks do not need custom move structs.

Factor named behavior lookup into one safe helper that:

- validates the option is a non-empty string;
- checks pointer/element/addressability/interfaceability before reflection operations;
- verifies the exact behavior type;
- returns contextual errors and never panics;
- is exercised during `ValidConfiguration`, so `auto.MustConfig` fails at boot.

Anonymous single-behavior discovery remains the trivial zero-option path.

## Phase 7: Real-game adoption and author documentation

In the paired `games` worktree, migrate `metaltrader`:

- declare `MerchantMarket` and `PointMarket` named `FaceUpMarket` fields;
- remove the two hand-configured `MoveComponentsUntilCountReached` replenishment blocks;
- add two `ReplenishMarket` configs with `WithMarketField`;
- infer each target from its bounded display stack;
- regenerate checked-in readers/goldens with the repository tools;
- prove gameplay/legal golden parity.

Do not force the abstraction into games whose lifecycle differs:

- Valentine and Darwin also recycle `UnusedCards` and should retain composed custom logic.
- Pass performs between-round setup/scoring, not ordinary draw/discard reshuffling.
- Murder Mr. Monroe embeds a pair but does not currently consume the companion move; audit the embedding and either document its accessor value or remove cargo-cult use without changing rules.

Update package docs and the tutorial with three progressive examples:

1. one bounded anonymous market and one draw/discard pair—no move methods, no repeated target size;
2. two named markets/pairs selected with config options;
3. exception composition, where a game-specific ordered FixUp sits beside the generic replenisher instead of injecting callbacks into the behavior.

Correct references to nonexistent `interfaces.DrawDiscardProvider` and `interfaces.MarketProvider`. Document the actual `Has*` discovery contracts.

## Scope explicitly deferred

- A framework-wide immutable/mutable redesign for behavior getters.
- Generic behavior callbacks, closures, or arbitrary extra recycle stacks.
- Automatic elimination-to-inactive coupling.
- Round-score or score-threshold abstractions.
- Breaking redesigns of Team, Score, Elimination, or MoveBudget.
- Declarative legality helpers for those four behaviors; useful, but independently reviewable.

Record MoveBudget's unchecked negative-value API and PlayerTeam's unset-enum semantics as follow-up issues rather than burying source breaks in this tranche.

## Commit and critique sequence

1. `docs: plan stack invariants and behavior hardening` — this document only.
2. Sub-agent review of this exact committed plan; amend the plan for accepted findings before implementation.
3. `test: reproduce detached and cross-state stack invariant holes`
4. `fix: enforce physical stack ownership and component conservation` (#751)
5. `fix: make bulk stack operations atomic`
6. `test: cover behavior companions in real states`
7. `fix: harden declarative stack behaviors` (#793)
8. `refactor(metaltrader): adopt named face-up markets`
9. `docs: teach declarative stack behaviors`
10. Final adversarial review, targeted corrections, full validation, careful rebase, and only then landing.

## Validation matrix

Run at each relevant boundary, not only at the end:

### Boardgame, independently

```sh
GOWORK=off go test ./...
GOWORK=off go vet ./...
```

Run repository-specific lint/codegen/golden checks discovered from `AGENTS.md`, scripts, and CI configuration.

### Games, independently

```sh
GOWORK=off go test ./...
GOWORK=off go vet ./...
```

Regenerate and verify all affected generated readers and legal goldens.

### Paired workspace

```sh
go test ./boardgame/... ./games/...
go vet ./boardgame/... ./games/...
```

### Focused reliability checks

- Repeat ownership, persistence, behavior FixUp, and metaltrader tests uncached.
- Run the race detector on affected boardgame and game packages where practical.
- Run the conservation benchmark and compare it with state marshal/save cost.
- Confirm both worktrees contain only intended changes and no files from the concurrently active agent.

## Done means

- Every persisted non-generic component has exactly one reachable physical owner in an unsanitized state.
- Detached, stale, hidden, aliased, and foreign-state stacks fail before mutation or persistence.
- Initial setup and corrupt storage cannot bypass the invariant.
- Valid boards, merged views, dynamic values, copies, loads, and sanitized projections retain existing semantics.
- The declarative behavior common case remains one embedded behavior plus `auto.MustConfig`.
- Multiple markets and draw/discard pairs are possible without custom move types.
- Companion moves fail predictably in `Legal`, never panic during configuration, and preserve deterministic slot/shuffle behavior.
- Metaltrader proves the named multi-market API in a real game.
- Boardgame and every game in `../games` pass independently and together.
- #751 and #793 receive accurate GitHub summaries; #793 is described as completed earlier and hardened here, not newly invented.
