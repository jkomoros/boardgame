# Exploration report: should behaviors contribute preconditions?

> Provenance: sub-agent exploration report, 2026-07-12, session on branch
> `declarative-legality-design`. Checked in verbatim as a design input for the
> follow-up workstreams in
> `docs/superpowers/plans/2026-07-12-legality-followups-roadmap.md`.
> File:line citations reference the branch state at commit d8716722.

## TL;DR

Behaviors already ARE the vocabulary the declarative-legality system reads — `inPhase` declares a Read on `game.Phase` (PhaseBehavior's field), `proposerIsCurrentPlayer` declares `game.CurrentPlayer` (CurrentPlayerBehavior's field), `AllActivePlayers` hard-codes `behaviors.PlayerIsInactive`, and a migrated game (valentine) already writes `legal.PropAtLeast("player.MovesLeft", 1)` directly against MoveBudget's field. What's missing is not a new attachment mechanism but **named, behavior-aware catalog predicates** (option a), plus a follow-up that widens the move-type seam to `AnyPlayer`/`AdminPlayer` so *those move types* contribute behavior atoms (option d, realized through the existing `ContributedPreconditions` channel). Auto-weaving behavior checks into "every player-targeting move" (option b) is actively wrong — the framework's own moves (`ActivateInactivePlayer`, `SeatPlayer`) deliberately target inactive/unfilled players.

---

## 1. The behaviors package as it stands

Files: `behaviors/` — actual behavior names and their persisted properties:

| Behavior | Properties (codegen'd into PropertyReader) | Companion interface (`moves/interfaces`) |
|---|---|---|
| `Seat` | `SeatFilled`, `SeatClosed` (bools) | `interfaces.Seater` |
| `InactivePlayer` | `PlayerInactive` (bool) | `interfaces.PlayerInactiver` |
| `PlayerElimination` | `Eliminated` (bool) | `interfaces.PlayerEliminator` |
| `GameAdministrator` | `IsAdmin` (bool) | `behaviors.HasGameAdministrator` |
| `MoveBudget` | `MovesLeft` (int) | (HasMovesLeft method) |
| `PlayerSubmission` | `PlayerSubmitted` (bool) | `interfaces.PlayerSubmitter` |
| `CurrentPlayerBehavior` | `CurrentPlayer` (PlayerIndex, on gameState) | `interfaces.CurrentPlayerSetter` |
| `PhaseBehavior` | `Phase` (enum, on gameState) | — |
| plus `PlayerColor/Role/Team`, `ScoreBehavior`, `LocationBehavior`, `RoundRobin`, `DrawDiscardPair`, `FaceUpMarket`, `PlayerOrderBehavior` | | |

Key structural facts (from `behaviors/main.go` doc):
- Behaviors are anonymously embedded in `playerState`/`gameState`; **codegen includes their fields in the generated PropertyReader**, so behavior fields are ordinary state properties — the legal path grammar (`player.X`, `players[move.Field].X`, `game.X`) already resolves them with zero new machinery.
- Discovery is by **type assertion on companion interfaces** (`behaviors.PlayerIsInactive(p)`, `player.(interfaces.Seater)`), with "skip check if behavior absent" as the universal imperative convention.
- `Connectable` behaviors get `ConnectBehavior` called automatically; irrelevant to legality but shows the framework already has a boot-time behavior-discovery pass.
- The package doc's own "Schelling point" framing (main.go:47-52) is exactly the property this design needs: behaviors standardize *names* multiple systems coordinate around.

Real-game embedding (all verified): werewolf playerState embeds `Seat + InactivePlayer + PlayerElimination + PlayerRole`; blackjack embeds `Seat + InactivePlayer + ScoreBehavior + PlayerElimination`; ../games pass, valentine, murdermrmonroe all embed `Seat + InactivePlayer` (+ ScoreBehavior/MoveBudget/PlayerElimination variously).

## 2. Legality semantics behaviors already imply — and where that logic lives today

All imperative, scattered across three layers:

**Framework move types** (`moves/` package):
- `moves.AnyPlayer.Legal` (any_player.go:70-74): target's seat must be filled — `if seater, ok := player.(interfaces.Seater); ok { if !seater.SeatIsFilled() { return errors.New("your seat is not yet filled") } }`. Skip-if-absent semantics.
- `moves.AdminPlayer.Legal` (admin_player.go:45-48): target must be admin, skip-if-absent via `HasGameAdministrator`.
- `moves.SeatPlayer` (seat_player.go): a dozen `SeatIsFilled`/`SeatIsClosed`/`PlayerIsInactive` checks — but these *invert* the usual polarity (a seat must be *unfilled and open* to seat someone).
- `moves.ActivateInactivePlayer.Legal` (inactive_player.go:56): target must *be* inactive — also inverted polarity.
- `moves.SelectTeam/SelectRole/SelectColor`: uniqueness scans skip unfilled seats.
- `moves.AllPlayersSubmitted`, gathering.go's `WithRequireAdmin` path: skip-inactive quantifiers, admin gates.

**Core engine**: `PlayerIndex.Next/Prev` skip inactive players — an *implicit* legality effect (CurrentPlayer moves never land on an inactive player, so nobody writes that check).

**Per-game code**: werewolf's `moveCastVote.Legal` (examples/werewolf/moves.go:127-174) is the canonical specimen — voter not `Eliminated`, target not `Eliminated`, target not `behaviors.PlayerIsInactive`, plus role/phase logic. Blackjack's `moveCurrentPlayerStand.Legal` checks `Eliminated`/`Stood` bools (one is PlayerElimination's field, one a plain game field — instructive: the boundary is invisible at the check site).

**Already-declarative touchpoints** (this is the strongest evidence for option a): `legal/catalog_framework.go:117` declares `Read{Path: "game.Phase"}` *by convention on PhaseBehavior's field*; `legal/catalog_players.go:423` likewise `game.CurrentPlayer`; `AllActivePlayers`' Evaluate calls `behaviors.PlayerIsInactive` directly (catalog_players.go:336). The legal package already imports behaviors.

## 3. Design options

### (a) Behavior-aware catalog predicates — recommended v1

New builders in `legal/` (a `catalog_behaviors.go`), pure sugar over the *existing* path grammar + interface helpers:

```go
legal.SeatFilled(playerSel string)          // "player", "players[move.VoteTarget]"
legal.PlayerActive(playerSel string)        // PlayerInactive == false
legal.PlayerNotEliminated(playerSel string) // Eliminated == false
legal.PlayerEliminated(playerSel string)    // for moves that require it
legal.ProposerIsAdmin()                     // IsAdmin on the proposer-named target field
legal.HasMoveBudget(playerSel string)       // MovesLeft >= 1
```

Mechanics: each declares Reads on the behavior's **canonical property names** (`players[move.VoteTarget].PlayerInactive`, Facet Values), CostTrivial/Cheap, and ships a default template (`legal.seat_not_filled`: "that seat is empty"). The `players[move.<Field>].<Prop>` path kind (legal_path.go:30-39, added in the completeness round and already used by darwin and tictactoe) is exactly the selector shape needed. Evaluation can go through the raw property (client-evaluable) rather than the interface.

- **Suppression**: N/A — these are authored specs; you simply don't write them. `WithoutPrecondition` continues to apply only to contributed atoms.
- **Purely-sugar guarantee**: trivially preserved — nothing changes for any move that doesn't author them.
- **Boot validation**: free. `validateLegalPath` (legal_path.go:138) already checks the property exists on the example player reader; a game whose playerState lacks `Seat` gets a boot error naming the move and path. (A game that implements `interfaces.Seater` with custom field names but no behavior embed fails boot path-validation — the conservative direction; it keeps `LegalCustom`. Same class of "by-convention Read" limitation already documented for `game.Phase`/`game.CurrentPlayer`.)
- **Ledger benefit**: large. Seat/inactive/eliminated bools are visible-by-default properties, so `evaluable: true` for essentially all viewers; a client can gray out "Vote for Bob" with a rendered "Bob has been eliminated" and disable a whole move panel with "your seat is not yet filled" — per-target, pre-round-trip.

Cost: ~6 small predicates + templates + conformance-corpus rows. No core changes. Note most are *already expressible* today via `PropEquals(sel+".SeatFilled", "true")` / `PlayerBoolIs("Eliminated", false)`; the sugar buys stable registry names (a TS evaluator implements `seatFilled` once), default templates, and greppable intent.

### (b) Behavior interface method `ContributedPreconditions(scope)`, auto-woven — rejected

The wrinkle is fatal, twice over. First, "which moves does a playerState behavior's precondition attach to?" has no good answer: moves name their target via arbitrary fields (`TargetPlayerIndex` by convention, but werewolf's `VoteTarget`, and some moves have two player-referencing fields). Second — decisive — polarity varies per move: `ActivateInactivePlayer` *requires* the target inactive; `SeatPlayer` requires the seat *unfilled*; `moveCastVote` requires the voter's `Vote` field to still be unset. Auto-weaving "target must be active/seated" into every player-targeting opted-in move breaks the framework's own moves, forces pervasive `WithoutPrecondition` boilerplate, and would flip existing migrated games' golden-equivalence tests (a plan silently growing atoms changes ledger shape and first-failure messages). It also inverts the system's design grain: everywhere else, *the move knows itself* (spec §2's contribution model) and boot fails loudly on anything implicit.

### (c) `moves.WithBehaviorPreconditions` per-move opt-in — subsumed

Once (a) exists this is just a grouping macro. A defensible middle form is one composite predicate, `legal.TargetablePlayer("move.VoteTarget")` ≙ seated ∧ active ∧ not-eliminated *for whichever of those behaviors the game embeds*, resolved at boot. But conditional-on-embed composition makes the ledger entry's meaning game-dependent and complicates the TS evaluator; explicit atoms in (a) are more honest. Skip in v1; revisit if authoring feels verbose in practice.

### (d) Behavior contributions mediated by the move types that already encode them — recommended follow-up

The observation that dissolves the state-struct/move-type wrinkle: **the framework already has move types whose entire imperative `Legal()` is "consult a behavior"** — `AnyPlayer` (seat filled) and `AdminPlayer` (is admin). These are the natural contribution points, exactly parallel to `CurrentPlayer` contributing `proposerIsCurrentPlayer` today. The behavior never needs to know which moves it affects; the moves that consult it already exist.

Concretely: widen `legalSupportedMovesBaseTypes` (legal_plan.go:670) to `AnyPlayer`/`AdminPlayer`, giving them:

```go
func (a *AnyPlayer) ContributedPreconditions() []legal.Spec {
    return append(a.Default.ContributedPreconditions(),
        legal.ProposerIsTargetPlayer(),                    // the >=0 / Equivalent / bounds checks
        legal.SeatFilled("players[move.TargetPlayerIndex]"))
}
```

What this requires beyond (a):
1. **The seam invariant changes shape.** Today's allowlist rule is "the type must declare no `Legal()` override" (enforced by `moves/seam_source_test.go`); `AnyPlayer`/`AdminPlayer` *do* override. They'd need the `CurrentPlayer` treatment: contributed atoms proven byte-equivalent to the imperative override (golden tests), plus extending the `legalMoveEmbedsCurrentPlayer`-style boot guard (legal_plan.go:434) — suppressing `"targetSeatFilled"` on an `AnyPlayer`-embedding move must be a boot error, since the imperative twin still runs and the ledger would desynchronize.
2. **Conditional contribution needs example-state access.** `AnyPlayer`'s seat check is skip-if-absent; the contributed atom must be *omitted entirely* when the game's playerState lacks `Seat` (a runtime vacuous-pass would lie to the ledger; an unconditional atom would fail boot path-validation for seat-less games). But `ContributedPreconditions()` takes no arguments and is called on an example move with no state (legal_plan.go:443). Cleanest fix: `assembleLegalPlans` filters contributed (never authored) specs that are marked behavior-gated and whose canonical property is absent from the example player reader — core already holds `exampleState` right there. Alternative: a second optional interface `ContributedPreconditionsForState(exampleState)`. Either is small; this is the one genuine plumbing decision in the whole design.
3. Payoff: werewolf's `moveCastVote` (and every `AnyPlayer` gathering-phase move in ../games) becomes seam-eligible, and *every* AnyPlayer move in every game gets a free, client-evaluable "your seat is not yet filled" ledger row with zero authoring.

## 4. Grounding: before/after

### blackjack `moveCurrentPlayerStand` (examples/blackjack/moves.go:287-306) — works with v1 (a) alone; embeds CurrentPlayer, already seam-supported

Before (today, fully imperative):
```go
func (m *moveCurrentPlayerStand) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {
    if err := m.CurrentPlayer.Legal(state, proposer); err != nil { return err }
    game, players := concreteStates(state)
    currentPlayer := players[game.CurrentPlayer.EnsureValid(state)]
    if currentPlayer.Eliminated { return errors.New("the current player has already busted") }
    if currentPlayer.Stood { return errors.New("the current player already stood") }
    return nil
}
```
After (`Legal()` deleted):
```go
auto.MustConfig(new(moveCurrentPlayerStand),
    moves.WithPreconditions(
        legal.PlayerNotEliminated("player").WithMessage("blackjack.already_busted"),
        legal.PlayerBoolIs("Stood", false).WithMessage("blackjack.already_stood"),
    ))
```
Instructive contrast in one plan: `Eliminated` is a behavior field → behavior-aware predicate with a sensible default template; `Stood` is a plain game field → generic `PlayerBoolIs`. (The in-file comment claiming "the catalog has no negation primitive" predates `PlayerBoolIs` and is stale.) Both atoms field-independent → memoized across the move-forms double pass; both client-evaluable.

### werewolf `moveCastVote` (examples/werewolf/moves.go:121-174) — needs follow-up (d)

Before: 45-line imperative `Legal()`, plus the file comment documenting the `AnyPlayer` seam block as "the dispositive blocker."
After (with AnyPlayer seam-widened; `Legal()` shrinks to a `LegalCustom` holding only the phase/role disjunction, which stays imperative per the no-nested-compositor rule):
```go
auto.MustConfig(new(moveCastVote),
    moves.WithPreconditions(
        // contributed free by AnyPlayer: proposer-is-target + targetSeatFilled
        legal.PlayerNotEliminated("players[move.TargetPlayerIndex]").WithMessage("werewolf.eliminated_cannot_vote"),
        legal.PropEquals("players[move.TargetPlayerIndex].Vote", "observer"), // "not yet voted" sentinel
        legal.PlayerNotEliminated("players[move.VoteTarget]").WithMessage("werewolf.target_eliminated"),
        legal.PlayerActive("players[move.VoteTarget]").WithMessage("werewolf.target_inactive"),
        // self-vote guard needs a field-vs-field compare — residue in v1
    ))
func (m *moveCastVote) LegalCustom(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {
    // night=>werewolf-only / day disjunction, verbatim
}
```
Six of eight checks become ledger rows a client can render per-vote-target ("cannot vote for an eliminated player" grayed out on Bob's portrait). Valentine's shipped `legal.PropAtLeast("player.MovesLeft", 1)` (../games/valentine/main.go:148) is the existing proof that games already reach for behavior fields declaratively — `legal.HasMoveBudget("player")` just names it.

## 5. Risks

- **Double-checking** (behavior atom contributed + move authors the same check, or contributed atom + frozen imperative twin on CurrentPlayer/AnyPlayer embedders): correctness-harmless (pure reads, idempotent) but ledger-noisy and, for suppression, dangerous — the existing `legalMoveEmbedsCurrentPlayer` boot guard is the exact precedent and must be generalized per widened type. A dedupe lint (same name+args twice in one plan) is cheap.
- **Ordering**: behavior atoms with `players[move.X]` selectors are field-dependent by construction (legal_path.go:36-38), so they evaluate after all field-independent atoms regardless of declaration order — first-failure-message shifts, already a documented and golden-test-asserted divergence class (`knownMessageOrderingDivergence`).
- **Boot-validation story**: property-missing → existing `validateLegalPath` boot error naming move+path (nothing new to build). Behavior-absent-but-property-coincidentally-present → silently reads the impostor field; acceptable (same convention risk as `game.Phase`) but worth a doc note. For contributed atoms, absence must mean *omit at assembly*, decided against the example state (the one new mechanism, §3d.2).
- **Sanitization edge**: role/team-style behavior fields carry `sanitize:"other:hidden"` tags; the ledger's per-viewer facet-based `evaluable` computation already handles this — behavior predicates need no special casing, but tests should cover a hidden behavior field.
- **Polarity trap**: any temptation to make behavior predicates "smart" (auto-attach, default-on) re-imports option (b)'s failure. Behaviors *supply vocabulary*; moves *choose polarity*.

## 6. Recommendation

**Yes — worth doing, in form (a) now and (d) next; never (b).**

**Minimal v1** (no core changes, no new interfaces): add `legal/catalog_behaviors.go` with `SeatFilled`, `PlayerActive`, `PlayerNotEliminated` (+ inverses where a real move needs them: `PlayerEliminated` exists in SeatPlayer-land), `ProposerIsAdmin`, `HasMoveBudget` — each a thin, Reads-honest predicate over the behavior's canonical property names with default templates and conformance-corpus rows; migrate blackjack's Stand/Hit gates and valentine/murdermrmonroe's MoveBudget gates to the sugar as the golden-test bed. This is a day-scale change that follows the "prefer existing primitives" principle: the path grammar, boot validation, ledger, and suppression machinery all already do the right thing.

**Follow-up v2** (the actual "behaviors contribute preconditions" milestone): widen the seam to `AnyPlayer`/`AdminPlayer` with contributed atoms (`proposerIsTargetPlayer`, `targetSeatFilled`, `targetIsAdmin`), omit-if-behavior-absent at plan assembly, suppression-divergence boot guards, and golden equivalence against the frozen overrides — unblocking werewolf-shaped games and giving every seated-play game free client-side seat/admin gray-outs. This is the same playbook `CurrentPlayer` already validated, so it's sequencing risk, not design risk.
