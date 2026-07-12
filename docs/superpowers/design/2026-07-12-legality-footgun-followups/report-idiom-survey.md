# Exploration report: where the declarative-legality idioms should spread next

> Provenance: sub-agent exploration report, 2026-07-12, session on branch
> `declarative-legality-design`. Checked in verbatim as a design input for the
> follow-up workstreams in
> `docs/superpowers/plans/2026-07-12-legality-followups-roadmap.md`.
> File:line citations reference the branch state at commit d8716722.

Survey of the boardgame framework (branch `declarative-legality-design`) and `../games` after this branch's migrations. Method: read every remaining imperative `Legal()` body in `examples/*` and `../games/*` (the migration-survey doc comments left behind are themselves a high-quality dataset), the `legal/` catalog surface and its `doc.go` "v2 limits" self-assessment, `legal_plan.go`'s boot gauntlet, `stack_constraint.go`/`constraints/`, `behaviors/`, plus two deep traces: error-string flow from Go to the browser, and boot-vs-late validation timing across `NewGameManager`.

## 1. What's left imperative, clustered

Catalog state first: the "completeness round" (spec 2026-07-12, documented in `legal/doc.go:236-317`) already closed four gaps beyond the original spec — `StackCount/StackEmpty/StackNotEmpty`, `PropEquals/PropNotEquals`, the `players[move.<Field>].<Prop>` path kind, `PlayerBoolIs/ComponentAbsentAt` — and widened the seam to Default/CurrentPlayer/FixUp/FixUpMulti/StartPhase. Adoption is broad: every example game except werewolf and every `../games` game has `WithPreconditions` in `main.go`; `pass` (new game) has zero imperative Legal bodies.

The remaining imperative bodies cluster as:

**Cluster A — stale-survey residue (cheap wins, no new machinery).** Several moves' "no catalog primitive exists" comments predate the completeness round and cite now-false reasons:
- `examples/pig/moves.go:139` `moveCountDie` — "DieCounted must be false"; comment says no negation exists, but `PlayerBoolIs(prop, false)` now does.
- `examples/blackjack/moves.go:288` `moveCurrentPlayerStand` — two negated bools; fully expressible via `PlayerBoolIs` x2 today.
- `examples/memory/moves.go:221` `moveHideCards` — `PropCompare("player.CardsLeftToReveal","==",0)` + `StackNotEmpty("game.VisibleCards")`; fully expressible.
- `examples/checkers/moves.go:60` `movePlaceToken` — was seam-blocked on FixUpMulti, which is now supported; `StackNotEmpty` + `MayMoveToSlot` + the already-registered `checkers.spaceIsBlack` cover it.
- `../games/darwin/moves.go:13-38` explicitly flags `moveReplaceHot/ColdClimateCard` as "now fully expressible, not attempted, real low-effort follow-up."

A re-migration sweep is nearly free given the golden-equivalence fence machinery already in every game.

**Cluster B — component-values reads (the #1 durable blocker).** `legal/doc.go:272-277` already names dynamic component values "the single most common reason a real move stays partially opaque," and the survey confirms *static* chest values are equally blocking: memory's card-type comparison (`moves.go:105,167`), metaltrader's `card.Type != merchantCardMetal` and per-metal payout thresholds (`activate_moves.go:24`), valentine's `ActiveCardType().Equals(cardType)` (nine Activate moves, `activate_moves.go:1-160`), blackjack's `HandValue()` (derived from card values). Recurrent shapes: "stack's sole/indexed component has enum prop == constant," "two components in a stack have equal/unequal prop," and cross-path compare "stack count >= component's int prop" (metaltrader).

**Cluster C — target-player validation.** werewolf `moveCastVote` (`moves.go:127`): target valid / not eliminated / not inactive / not self. valentine `otherPlayerActivateMove` (`activate_moves.go:95-160`): selected player valid / not eliminated / not protected / self-only-if-no-others (a count over eligible others). murdermrmonroe touches the same shapes. This is a genuine cross-game vocabulary — "the player this move points at is a legal target" — and it is built almost entirely from *behavior-owned* properties (Eliminated, Protected, PlayerIsInactive).

**Cluster D — quantifier gaps.** blackjack's cleanup moves (`moves.go:83-221`) need EXISTS ("at least one active player has cards"), not just `AllActivePlayers` (FORALL); werewolf `moveResolveVotes` needs per-player *exemption* conditions over enum/PlayerIndex-typed props (documented in-file as "a genuine, reportable catalog gap"); darwin's FixUp moves need a quantifier that *names the failing player in the error* — i.e., quantifiers that capture a binding into the LegalMessage.

**Cluster E — delegate-computed methods.** `HandValue()`, `LuckPower()`, `CanBeSeen()`, `DoneFeeding()`, `Timer.Active()` — genuinely game-specific residue. The right answer already exists (game-registered predicates, per checkers), but the friction is real: no `Predicate1`-style helper ever shipped (spec Implementation notes), so authors hand-roll a ~20-line `PredicateConstructor`. Lowering that boilerplate matters more than growing the catalog here.

**Cluster F — structural anti-pattern: parameterized shared base-move Legal.** valentine's `baseActivateMove.Legal(state, proposer, cardType)` — nine concrete moves share one imperative helper with call-site parameters, defeating the per-move-type seam entirely (documented in `activate_moves.go:11-48` as "exactly where declarative migration is valuable but costly to retrofit"). Truly game-specific in content, but the *shape* (game-local abstract move types) is common; the Cluster B predicates would dissolve most of its body.

Also durable (per `doc.go`): `proposer.X` paths with no move field naming the proposer (needs a LegalForAnyone redesign — **not** sugar-compatible, correctly deferred), no general `not`/`all` compositors (anti-tarpit rule, keep), and the ~24 opaque framework move base types (mechanical seam expansions, sequencing not design).

## 2. Error surfaces beyond Legal — where template+bindings pays

Trace results (the `errors.Friendly` triple msg/friendlyMsg/secureMsg from `errors/main.go` underlies everything; the API always ships both `Error` and `FriendlyError`, and the client shows `friendlyError || error`):

1. **ProposeMove submit rejection is the highest-value target.** `game.go:966` does `errors.NewFriendly(err.Error())` — the Legal failure string (template-rendered server-side for opted-in moves, raw Go string otherwise) becomes the literal dialog text the player reads. The structure is *flattened at exactly the seam where the client could use it*. Promotion: a Friendly variant carrying the `LegalMessage` so `renderer.Error()` (`server/api/main.go:321-351`) serializes `{Error, FriendlyError, Message:{template,bindings}}`. Purely additive wire change.
2. **Host/join/seat REST endpoints bypass Friendly entirely** — `server/api/join.go:42,48`, `join_seat.go:107-257`, `host_actions.go:140-414` dump raw `err.Error()` into `gin.H{"error":...}` for player-triggered actions ("game full", "not your turn to seat"). Worst raw-string leak in the system today.
3. **`LegalForPlayerError` is a dead wire field client-side** — opted-in moves already render via `LegalRenderVerdict` (`main.go:1761`), opaque moves ship raw `err.Error()` (`main.go:1701`), and *no client component reads it at all* (only the booleans drive button state). The ledger + a TS `LegalMessage` renderer is the designed fix; the conformance corpus anticipated it. [NOTE: post-footgun-batch, opted-in booleans/errors also come from ground-truth `move.Legal()` — commit 1d82dfba.]
4. **Apply() and setup-hook errors are already shielded** by generic friendly text ("The move could not be made"; `game.go:983`, `game.go:522-553`) — the raw string travels only in the unused fallback field. Structuring these buys specificity, not leak-plugging: low priority. `boardgame-util` codegen/serve errors are developer-terminal-only; raw strings are fine there. Vestigial: `lastErrorMessage` in the state bundle (`server/api/main.go:46`) is only ever cleared, never set — deletable.

## 3. Boot-gauntlet candidates (late-failure audit)

`NewGameManager` already validates a lot (struct tags incl. deck/enum/policy-name typos, move name collisions, every move's `ValidConfiguration`, sanitization group names for game/DCV states, the new legal-plan assembly + reachability probe). The worst remaining traps, ranked:

1. **Silent privacy leak: unmatched sanitization group → `PolicyVisible`** (`base/game_delegate.go:363-365`). A syntactically-valid-but-never-matching group renders hidden state fully visible, forever, with no error. Hardest to fully boot-check, but even a boot warning for policies that resolve to visible-for-observer on the example state would be a big deal.
2. **`WithLegalPhases` with a bogus/wrong-enum phase key and no progression is never validated** (`moves/default.go:113-153` only checks phases when a progression coexists; `legal_framework.go:92-121` just never matches) → move silently never legal. **Cheapest, highest-value fix**: validate keys against `PhaseEnum()` unconditionally in `Default.ValidConfiguration` — slots straight into the existing gauntlet with the move name attached.
3. **Behaviors nested >1 embed level are never connected nor validated** (`state.go:975-995` and `game_manager.go:437-463` don't recurse) → nil container panic at first move. Fix is recursion in both walkers.
4. **Setup hooks (`DistributeComponentToStarterStack`/`BeginSetUp`/`FinishSetUp`) first run at first `NewGame`, not boot** (`game.go:522-553`). A throwaway `DefaultNumPlayers` game inside `NewGameManager` would move them into the gauntlet.
5. **Unsatisfiable move progressions deadlock silently at runtime** — a reachability dry-run analogous to `probeLegalReachable` (`legal_plan.go:612`) is the same idiom applied to progressions. Also: agent name collisions silently last-write-win (`game_manager.go:400-403`), and the legal catalog's own admitted gap — enum-name typos in `PropEquals` degrade to `LegalUnknown` at evaluate time instead of boot error (`doc.go:292-299`) because predicate constructors lack an example-state validation hook; adding one hook closes both this and future cases.

## 4. Constraints vs. the predicate registry

They already deliberately rhyme (`legal_predicate.go:94` says "Constructors mirror constraints.StackConstraintConstructor"), and the Legal-time bridge already exists: the `stackConstraints` catalog predicate (`legal/catalog_framework.go`) wraps the same `boardgame.LegalStackConstraintsCheck` the frozen chain calls. **Do not merge the execution models**: a `StackConstraint` is a guard on *mutation* (checked in `moveComponentImpl` during Apply, on every insertion path), a predicate is a *precondition* on a move — collapsing them would change Apply-time semantics and break the purely-sugar ethos. What should unify over time:
- **The message layer.** Constraint failures surface only as a verbatim `detail` binding inside `legal.stack_constraints`. Let constraints optionally return template-key+bindings (an optional `interface{ LegalMessage() ... }` on the returned error), declare `EmittedTemplates` on `StackConstraintConstructor`, and validate them in the same boot template check.
- **The registry/validation pattern** is already shared (delegate-configured named constructors, boot-time tag parsing via `StructInflater`). The remaining asymmetry — constraints are anonymous funcs at Legal-time, so the `stackConstraints` predicate can never be client-evaluable — is acceptable; per-constraint TS implementations could ride the same game-registered-predicate extension path later.

## 5. Behaviors package gaps

- `seat.go` and `color_palette.go` have no `ValidConfiguration` at all (partially mitigated by the Seater/SeatPlayerMove boot pairing check).
- The nested-embed connection bug (§3.3) is the structural gap.
- Bigger observation: **behaviors define the legality vocabulary but ship no predicates.** `MoveBudget`'s own doc comment (`behaviors/move_budget.go`) instructs authors to write `if !p.HasMovesLeft() { return errors.New(...) }` in Legal; `Eliminated`, `PlayerIsInactive`, `Role` (with its `other:hidden` sanitization default, whose self-vs-other evaluability nuance the ledger already handles) all recur across games as hand-rolled checks. `AllActivePlayers` already imports `behaviors.PlayerIsInactive` — the dependency direction is settled.

## Top 5 next idioms to promote (ranked)

**1. Component-values predicates (catalog growth, Cluster B).**
Before (metaltrader): 40-line imperative `Legal()` comparing `ActiveCard.ComponentAt(0).Values().(*merchantCard).Type` and four bowl counts. After:
```go
moves.WithPreconditions(
    legal.ComponentPropEquals("game.ActiveCard", 0, "Type", "metal"),
    legal.StackCountAtLeastComponentProp("game.IronBowl", "game.ActiveCard", 0, "MetalIron"),
    ...)
```
Three purpose-built predicates — `ComponentPropEquals/NotEquals` (const compare), `ComponentsPropMatch` (memory's same-type pair), and one cross-path count-vs-prop compare — unblock memory, metaltrader, valentine, and parts of darwin. Static `Values()` first (chest-resolvable at boot, `FacetValues` reads); dynamic values later. Cost: moderate (predicates + conformance rows + templates); purely additive. This is the single biggest migration unlock by the framework's own admission.

**2. Behavior-paired predicates + the target-player family (Clusters C + §5).**
Each behavior exports its legality vocabulary with correct Reads/facets baked in: `legal.HasMovesLeft()`, `legal.PlayerActive("players[move.VoteTarget]")`, `legal.TargetIsNotSelf("move.VoteTarget")`, `legal.RoleIs(...)` (self-scoped read, sanitization-aware). Before (werewolf): six hand-rolled target checks. After: three specs. Cost: low-moderate; additive; also gives the future TS evaluator a stable, small vocabulary. Natural home: a `behaviors`-adjacent catalog file, mirroring how `moves/catalog_framework.go` hosts `inProgression`.

**3. Structured messages on the API error envelope (§2).**
Carry `LegalMessage` through `errors.Friendly` → `renderer.Error()` as an additive `Message:{template,bindings}` field; adopt on the ProposeMove submit path first (`game.go:966`), then convert host/join/seat endpoints to Friendly-with-templates. Before: player sees `err.Error()` verbatim. After: client renders/localizes the same template table the ledger ships in the chest JSON. Cost: low server-side; client renderer is shared work with the (currently dead) `LegalForPlayerError`/ledger display. No back-compat risk — old fields stay byte-identical.

**4. Boot gauntlet round 2 (§3).**
In order of cost/benefit: (a) unconditional `WithLegalPhases` key validation (hours); (b) recursive behavior connect/validate walkers (small, fixes a real panic class); (c) agent-name collision check (trivial); (d) example-state validation hook for predicate constructors (closes the admitted enum-typo-to-Unknown gap); (e) throwaway-game setup-hook probe; (f) progression reachability probe (the `probeLegalReachable` idiom re-applied — most speculative, biggest payoff being deadlock elimination). The sanitization silent-visible-default (§3.1) deserves its own design pass — it's a privacy bug class, not just DX.

**5. Quantifier v2 (Cluster D).**
`AnyActivePlayer` (EXISTS dual of `AllActivePlayers`), inner-leaf support for enum/PlayerIndex/stack-count reads, and failing-player capture into message bindings (darwin's "must name the failing player" requirement — which also makes quantifier failures *better* than the imperative originals, not just equal). Cost: moderate — Kleene semantics for EXISTS are already proven by `any`, but conformance coverage must be exhaustive per the existing standard.

**Ethos flags:** everything above is additive sugar except (i) `proposer.X` paths — genuinely requires the LegalForAnyone per-player-existential redesign, keep deferred and honest; (ii) full constraints/predicate *execution* merger — would change Apply-time guard semantics, recommend message-layer unification only; (iii) general `not`/`all` compositors — keep excluded per the anti-tarpit rule; explicit-leaf negation has proven sufficient. One hygiene note: the stale Cluster A survey comments will actively mislead future authors ("no negation exists") — the re-migration sweep should update or discharge them.
