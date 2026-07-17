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
- Corrupt stored states with missing or duplicate components fail loudly on load.

These APIs remain source-compatible:

- `FaceUpMarket` keeps `size`, `SetDisplaySize`, and `DisplaySize`. An explicit size remains useful for an unbounded or intentionally partially filled display. If omitted on a bounded display, `DisplaySize` is inferred from `MaxSize`.
- Existing single anonymous `DrawDiscardPair` and `FaceUpMarket` discovery remains zero-config.
- Existing mutable stack getters remain for now. Fixing const-correctness for every `Connectable` behavior is a framework-wide source-breaking project and is out of scope.
- Team, Score, Elimination, and MoveBudget retain their current interfaces. In particular, this tranche will not silently change `ConsumeMove` or `ResetMovesTo` signatures.
- `ConfigureStackProp` and `ConfigureBoardProp` retain their current contract. Reader implementations cannot reliably distinguish construction from runtime. Replacements become unusable immediately because attachment locators re-read the live property, and lifecycle audits reject the resulting topology.

## Phase 0: Baselines and exploit tests

Before production code, add failing tests that prove the real holes rather than the already-blocked fresh-stack move case.

### #751 red tests

- Setup delegate returns a detached growable or sized stack; game creation fails before storage is written.
- Two real games under one manager attempt a cross-state transfer; both source and destination remain unchanged.
- A captured live stack is replaced in its property and then used as a destination; mutation fails.
- Two attached stack fields are swapped; both stacks become stale because neither matches its canonical location.
- A constructor accidentally returns the same captured stack for two independently created states; the second state fails ownership initialization.
- A merged view over declared owner stacks is accepted; a merged view over hidden backing stacks is rejected.
- One physical stack aliased at two properties, and at a property plus board space, is rejected.
- One board object aliased into two board properties is rejected.
- Missing, duplicate, wrong-deck-at-setup, invalid-index, and unsanitized generic components fail conservation.
- Board spaces, player stacks, and dynamic-component-value stacks participate in ownership and conservation.
- Failed initial or move-time validation does not write storage or advance the version.
- Valid loaded states and state copies still work; corrupt stored records fail during inflation; sanitized copies remain readable and are not subjected to unsanitized conservation rules.
- Historical valid storage fixtures still load; forged deck names, nested/repeated merged leaves, and corrupt indexes fail deterministically.

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
- mutable `Stack` and `Board` properties as physical owners;
- immutable stack properties as views only, with `MergedStack` the supported persisted view, never as hidden physical owners;
- merged views separately, without counting the view as another component location.

It must emit stable canonical paths such as:

- `Game.Draw`
- `Players[1].Hand`
- `Game.Board[3]`
- `DynamicComponentValues[cards][12].Tokens`

Property names, deck names, and map-backed groups will be sorted before diagnostics. Discovery is explicitly multi-pass so map/reflection ordering cannot affect validity:

1. collect every declared physical owner and diagnose nil, alias, and state conflicts;
2. build the state-owned canonical owner registry;
3. recursively validate merged views against the completed owner set.

Nested merged views are flattened recursively. Every concrete leaf must be a declared owner in the same state and deck; repeated leaves within one view are rejected; cycles are detected defensively; existing overlap and sized-stack shape validation remains.

Reject a concrete physical stack or board stored only behind an immutable property interface: `copyReader` intentionally does not import immutable container contents, so accepting one would lose components on copy. Reuse traversal for attachment, integrity validation, component-index construction, and copied-stack lookup. Extend `findCorrespondingStack` to pair stacks by canonical owner path, including board spaces.

## Phase 2: State-owned attachment identity

Add a private owner registry to `state`:

- physical concrete stack identity -> canonical locator;
- canonical path -> locator that re-reads the current property or board space.

After state construction has connected all substates, initialize this registry exactly once from the newly constructed shape. Initialization and later validation are separate operations: validation never attaches or legitimizes a replacement stack.

Rules:

- A stack may be registered to one state and one location only.
- Reusing a stack in a different state or location is an error, even if its state pointer was overwritten.
- Merged stacks are views and never enter the owner registry.
- Merged backing stacks must already be attached through declared physical locations.
- Each copied or inflated state builds a fresh registry over its own preconfigured destination stacks before copy/import or JSON payloads are applied. Registry entries are never imported from a source state.
- Sanitized stacks remain attached to their sanitized state but immutable.

Attachment validation re-reads the property/board location and requires it still point to the registered owner. This makes a captured replaced or swapped stack stale immediately; a non-nil state pointer or stale cached membership is not sufficient.

`mergedStack.setState` must not confer mutation authority on backing leaves. State pointers remain context only; only the state registry grants ownership.

Concrete `setState` wiring must reject overwriting a non-nil pointer with a different state before registry initialization. This catches a constructor-captured stack reused across states rather than erasing the evidence by assigning the new pointer.

Stack import/copy must preserve destination topology rather than whole-struct-copying source identity. In particular, preserve destination `statePtr`, `board`, and `boardIndex` while importing only serializable payload/configuration. Test that `copiedBoard.SpaceAt(i).Board() == copiedBoard` and that copied board spaces pass attachment and movement checks.

## Phase 3: Mutation-boundary enforcement

Centralize a private mutation precondition used by every stack mutator:

- stack is non-nil;
- stack has a state;
- state is unsanitized;
- stack has a unique current attachment matching its canonical property/board location.

Movement additionally requires source and destination to belong to the exact same `*state`, before capacity, constraints, or mutation. `MayMoveTo`, slot variants, `MayMoveAllTo`, and actual moves must agree on these ownership rules.

Errors will distinguish detached, stale/replaced, hidden merged backing, aliased, and foreign-state stacks. Where available, include canonical path, deck, game ID, and state version.

Add an explicit nil-destination check before dereferencing the destination.

Before setup calls private `insertComponentAt`, validate that the returned container is mutable, is a registered owner at its current canonical location, belongs to `stateCopy`, has `stack.Deck() == component.Deck()`, and has capacity. The later conservation audit is a backstop, not the primary setup error.

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

The audit reads framework concrete stacks' raw index slices rather than `ImmutableComponents`, which normalizes corrupt values. For growable stacks every index must be in `[0, deck.Len())`; for sized stacks only the empty sentinel `-1` is additionally legal. Generic/hidden sentinel `-2` is forbidden in unsanitized state. Persisted `deckName` must resolve to the exact manager-chest deck pointer.

Diagnostics for a duplicate name both locations. Diagnostics for a missing component name its deck and index.

Run the audit:

1. after `FinishSetUp`, before the initial state is persisted;
2. from `validateBeforeSave` for every committed move;
3. at the end of `stateFromRecord`, after every game/player/dynamic JSON payload is applied. This path must not assume `state.game` or a game ID is available.

Do not apply exact-conservation checks to sanitized states, whose hidden/generic representations intentionally omit identity.

Keep operation-time attachment checks even with this backstop: invalid actions should fail where attempted, while lifecycle validation protects against direct field assignment, malformed readers, corrupt storage, and internal framework bugs.

Full conservation does not run while constructing the intentionally empty manager boot example. Owner-registry initialization does run inside `emptyState`, after all substates are connected and before embedded behavior and move `ValidConfiguration` checks. Existing `copyFrom`/`importFrom` code imports payload into those registered destination containers and cannot replace the destination registry.

Lifecycle audits guarantee component uniqueness before lazy component-index construction. Rebuild the component index from the validated owner registry rather than an independent property scan. The lazy builder may then remain non-error-returning; duplicate diagnosis belongs to the explicit setup/save/load audits and must never again be silently accepted at a persistence boundary.

Add a focused benchmark with a large sparse sized stack/board, where slot count dominates component count. Record the result in the implementation commit and compare audit cost with full apply+marshal/save cost. Prefer clarity until measurement demonstrates the O(components + slots) audit is material.

## Deferred adjacent stack robustness

`MoveAllTo` can partially mutate its in-memory state before a later constraint rejects. The game pipeline discards a failed Apply copy, so conservation is preserved at persistence, but the direct method is not transaction-safe for a caller that retains the failed state.

Do not claim stack-level rollback in this tranche. Creator constraint closures can capture or mutate external state, randomness, callbacks, or other state that a two-stack snapshot cannot restore. Honest atomicity requires a whole-state transaction design and a side-effect contract for constraints. Record that as a separate follow-up issue.

For #793, `Legal` preflights the operation on a copy and leaves the authoritative state untouched; if an unforeseen Apply failure occurs, the normal game pipeline discards the Apply copy. Real `MoveAllTo` does not call `MayMoveAllTo` internally or promise rollback.

The `Board.SpaceAt(Len())` bounds bug is real but unrelated to #751/#793 and is deferred to its own focused fix.

## Phase 5: Harden the stack-backed behaviors

### Shared structural validation

Expose one narrow read-only root-package seam, `boardgame.ValidateStackAttachment(state, stack) error` (final name subject to local naming conventions), so subpackages can ask the core owner registry whether a stack is a current owner without expanding the `Stack` interface or exposing state pointers. Build an internal behavior helper on it that validates a source/destination pair:

- both connected and non-nil;
- distinct physical stacks;
- same deck;
- attached to the example state as declared owners;
- same exact state;
- behavior is connected to game state when the companion move discovers it there.

`DrawDiscardPair.ValidConfiguration` and `FaceUpMarket.ValidConfiguration` will use it. Errors identify the behavior and semantic role (`draw`, `discard`, `source`, `display`).

Boot validation order is pinned by integration test: `emptyState` completes owner-registry initialization before `verifySubStatesConnectedAndValid` and move `ValidConfiguration` run. Static boot checks cover topology, distinctness, deck identity, and target/capacity policy. Content-dependent capacity and constraint compatibility belong in move `Legal`, not behavior boot validation.

### FaceUpMarket target semantics

- When explicit `size`/`SetDisplaySize` is positive, use it.
- Otherwise, if the display is bounded, infer `DisplaySize()` dynamically from the current `display.MaxSize()` on every query. Omitted size means “fill to current capacity,” including later capacity expansion or contraction.
- If the display is unbounded and has no explicit size, fail configuration with exact guidance.
- Reject explicit targets larger than bounded capacity.
- Treat zero as unspecified/infer and reject negative explicit sizes.
- Keep existing API names to avoid gratuitous source churn.
- Use `MoveToNextSlot` so a sized market fills its first empty slot and a growable market appends.
- In `Legal`, obtain the source's first component and compute the same slot used by `MoveToNextSlot`: `display.Len()` for a growable stack, or `display.SizedStack().NextSlot()` for a sized stack. Call `MayMoveToSlot(display, slot)`. `Apply` executes `MoveToNextSlot`; occupied-gap and constraint tests pin parity without adding `NextSlot` to the common `Stack` interface.
- Test explicit partial targets, omitted targets across bounded capacity changes, over-target displays, sized gaps, and source exhaustion.

### DrawDiscardPair companion

- In `Legal`, after `NeedsReshuffle`, call the transfer preflight so predictable capacity/constraint failures are reported before `Apply`.
- Preserve the one-shuffle-after-complete-transfer behavior.
- Wrap underlying errors with `%w`.

### Symmetric named lookup

Add `WithDrawDiscardPairField("TrainCards")`, symmetric with `WithMarketField`, so games with multiple decks do not need custom move structs.

Factor named behavior lookup into one safe helper that:

- validates the option is a non-empty string;
- accepts only a direct value field of the exact behavior struct type, matching what `autoConnectBehaviors` actually connects;
- checks addressability/interfaceability before reflection operations;
- returns contextual errors and never panics;
- is exercised during `ValidConfiguration`, so `auto.MustConfig` fails at boot.

Anonymous single-behavior discovery remains the trivial zero-option path.

An explicit field option always wins and never falls back to an anonymous behavior. Empty, missing, inaccessible, or wrong-type fields fail boot. Register both `WithMarketField` and `WithDrawDiscardPairField` with custom-configuration consumer validation so using either option on an unrelated move fails during `auto.Config` instead of being silently ignored.

Named options also derive stable unique move names without requiring `WithMoveName`: for example `Replenish Merchant Market` and `Shuffle Train Cards Discard Into Draw`. Explicit `WithMoveName` retains normal precedence. Test two named instances boot together and remain distinguishable in history/debug output.

## Phase 6: Real-game adoption and author documentation

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

Update package docs and the tutorial with three progressive examples, consistently calling these “stack-backed” or “tag-configured” behaviors rather than conflating them with declarative legality:

1. one bounded anonymous market and one draw/discard pair—no move methods, no repeated target size;
2. two named markets/pairs selected with config options;
3. exception composition, where a game-specific ordered FixUp sits beside the generic replenisher instead of injecting callbacks into the behavior.

Correct references to nonexistent `interfaces.DrawDiscardProvider` and `interfaces.MarketProvider`. Document the actual `Has*` discovery contracts.

## Scope explicitly deferred

- A framework-wide immutable/mutable redesign for behavior getters.
- Transactional direct `MoveAllTo` and the adjacent board bounds fix.
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
5. `test: cover behavior companions in real states`
6. `fix: harden stack-backed behaviors` (#793)
7. `refactor(metaltrader): adopt named face-up markets`
8. `docs: teach stack-backed behaviors`
9. Final adversarial review, targeted corrections, full validation, careful rebase, and only then landing.

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
- The stack-backed behavior common case remains one embedded behavior plus `auto.MustConfig`.
- Multiple markets and draw/discard pairs are possible without custom move types.
- Companion moves fail predictably in `Legal`, never panic during configuration, and preserve deterministic slot/shuffle behavior.
- Metaltrader proves the named multi-market API in a real game.
- Boardgame and every game in `../games` pass independently and together.
- #751 and #793 receive accurate GitHub summaries; #793 is described as completed earlier and hardened here, not newly invented.
