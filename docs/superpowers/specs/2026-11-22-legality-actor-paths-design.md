# Legality: `player.X` names the move's player (actor paths)

Status: implemented (boardgame `claude/legal-proposer-paths-a74521`; games
`legal-actor-paths`). Closes the "current player and proposer are deliberately
distinct" limit in `legal/doc.go`.

## Problem

`player.X` (and `legal.CurrentPlayer()`, now `legal.Player()`, which spells it) resolved against
`state.CurrentPlayerIndex()`. In a simultaneous phase the delegate returns
`AnyPlayerIndex`, so every `player.X` read failed closed. Darwin's
Play Cards phase was the first real case, and its golden fence reverted the
migration.

Two workarounds have landed since then:

- **`proposer.X` / `legal.Proposer()`** resolves against the proposing
  player.
- **`players[move.TargetPlayerIndex].X` / `legal.PlayerFromMove("TargetPlayerIndex")`**
  resolves against a move field. Darwin's six Play Cards moves use this today.

### Current state, reproduced (step 1)

In a paired GAMES worktree I swapped Darwin's six Play Cards plans from
`PlayerFromMove("TargetPlayerIndex")` to `Proposer()`, and then to
`CurrentPlayer()`. Both fail every one of the six golden tests:

- `Proposer()`/`proposer.X` fails for **admin proposers**: `proposer.X`
  cannot resolve `AdminPlayerIndex`. The imperative code it replaces reads
  `players[m.TargetPlayerIndex]`, i.e. the player the admin is acting for.
  `AdminBypass`, which the semantic behavior helpers add, would make admin
  moves legal that the legacy code rejects, so it diverges too.
- `CurrentPlayer()`/`player.X` fails for every proposer, because the current
  player is `AnyPlayerIndex`.

So `proposer.X` does not cover Darwin. The only correct spelling was the
`PlayerFromMove` workaround, which has three costs:

1. it leaks a framework base type's field name (`TargetPlayerIndex`) into
   game rules;
2. it makes every such read field-dependent, which loses memoization, marks
   ledger entries provisional, and makes client evaluability conservative
   across all players, even in turn-based phases;
3. a move that runs in both a turn-based and a simultaneous phase has to be
   written in the simultaneous spelling everywhere.

## What authors mean

A rule like "you must have a card in hand" is about **the player whose move
this is**. Whether the phase is turn-based or simultaneous is the delegate's
call (`CurrentPlayerIndex`), not something the move decides. An author should
not have to pick between "current player" and "proposer" to say "you".

The framework already has this concept, spelled "current player" with a
wildcard. `moves.CurrentPlayer` is documented as "a move made by the current
player", and the delegate is told to return `AnyPlayerIndex` for
simultaneous phases. `PlayerIndex.Equivalent` treats `AnyPlayerIndex` and
`AdminPlayerIndex` as wildcards. `CurrentPlayer.Legal` accepts a move when
its `TargetPlayerIndex` is Equivalent to both the current player and the
proposer. Under that algebra, the concrete player a move belongs to is what
remains after narrowing the wildcards. Moves built on this base already mean
"current player" this way. `player.X` was the one place that read the
wildcard literally.

## Decision: `player.X` is the player the move is made for

`player.X` (selector `legal.Player()`; `legal.CurrentPlayer()` stays as a
deprecated alias) resolves to **the player the move is made for**:

- **A move that names its own player.** Moves built on `moves.CurrentPlayer`
  or `moves.AnyPlayer` implement the new `boardgame.LegalOnBehalfOfer` hook
  as their `TargetPlayerIndex`. For them, `player.X` is the proposer, or, for
  an `AdminPlayerIndex` proposer, the player the admin is acting for. This
  holds in every phase, and the current player is never consulted.
- **Any other move** (`moves.Default`, fix-ups): the game's current player
  when it is concrete. This is unchanged from before, for every proposer.
  When the current player is a wildcard (`AnyPlayerIndex` during simultaneous
  play, or `AdminPlayerIndex`), it is the proposer.

Anything non-concrete fails closed to Unknown, as before. That covers an
Observer proposer, an `ObserverPlayerIndex` current player for a move with no
player of its own, and an admin with nobody to act for.

Implemented in `legalMovePlayerIndex` (`legal_path.go`). It is the only
resolver for `player.X` paths, so every catalog predicate, every
game-registered predicate that uses `ctx.ResolvePath`, and every
`Player()` selector picks it up automatically. The moves package's
`player.` stack paths and `DrawToPlayer.Apply` use the same rule
(`stackPathMovePlayerState`). `Apply` has no proposer, so they read the
move's own `TargetPlayerIndex`. Legality has already tied that player to the
proposer, and `AnyPlayer`'s Legal fills an unset target in from the
proposer.

This rule came out of the adversarial review below. The first version
consulted a concrete current player first, for every move. That left an
`AnyPlayer` move made out of turn reading the current player's state, and an
`AnyPlayer` move in an Observer gathering phase reading nobody's.

### Why this is behavior-preserving

- **Moves without a player of their own.** They are byte-for-byte unchanged
  in every turn-based phase.
- **`CurrentPlayer` moves.** The contributed `proposerIsCurrentPlayer` atom
  runs first and requires target ≡ current ≡ proposer. So wherever it
  passes, "the proposer, or the admin's on-behalf target" is exactly the
  current player in turn-based phases, and exactly the legacy
  `players[m.TargetPlayerIndex]` in simultaneous ones. Verdicts can differ
  only in full-ledger entries for a move that already failed that turn
  check. The golden fence `moves/legal_move_player_golden_test.go` covers
  current ∈ {each player, Any, Observer} × target ∈ {each player, Admin,
  Observer, Any} × proposer ∈ {each player, Admin, Observer}, and it fails
  on the original semantics.
- **`AnyPlayer` moves.** These are not on the seam yet, so no declarative
  plan for them exists to change.
  `TestGoldenAnyPlayerMovePlayerIgnoresCurrentPlayer` fences the resolution
  (both `player.X` and the `player.` stack path) against the imperative
  `players[m.TargetPlayerIndex]` across turn-based, out-of-turn, simultaneous
  and Observer phases. It fails 72 times under the current-player-first rule.
- **Before this change**, every wildcard-phase read was Unknown, i.e.
  illegal. Nothing shipped could depend on that, because a move that is
  always illegal in its own phase is a bug.

### Interactions

- **`proposerIsCurrentPlayer` (contributed by `CurrentPlayer`).** Unchanged.
  It remains the actor gate. Path resolution never re-checks identity: it
  only names a player. If a game suppresses that atom, an author reading
  `player.X` gets the proposer (or the admin's on-behalf player), which is
  still "the player whose move this is". It is not an arbitrary target.
- **Admin.** Admin acting for a player is evaluated *as that player*, which
  matches `Equivalent` and the legacy code. `AdminBypass` (skip the check for
  admin) is still a separate, explicit policy for proposer-scoped behavior
  conditions. The two answer different questions, and on-behalf is the
  default because it preserves legacy semantics.
- **`AllActivePlayers`.** Its inner `player.X` already means the quantified
  player. Both uses now read the same way: "player" is the player in scope,
  meaning the move's player at top level and each player inside the
  quantifier.
- **`players[move.X]` / `PlayerFromMove`.** Still the spelling for a player
  a move *names*, such as a deal recipient, a vote or an attack target
  (Darwin's `moveDealCard` keeps it). It is no longer needed to name the
  actor.
- **`proposer.X` / `Proposer()`.** It is kept for checks that should exempt
  admin entirely (its semantic helpers carry AdminBypass). It is not needed
  for simultaneous play or for `AnyPlayer` moves. A plain `Default` move made
  out of turn names its player with `PlayerFromMove` +
  `ProposerIsPlayerFromMove`.
- **Memoization.** `player.X` stays in the field-independent bucket. That is
  correct whenever the move's player comes from the current player or the
  proposer, both of which the memo key already covers. For the admin
  on-behalf case the memo key gains a `movePlayer` component, the resolved
  index (`legal_memo.go`). Plans whose field-independent steps never read
  `player.*` keep a constant there, so their keys are unchanged. The
  alternative, demoting every `player.X` read to field-dependent, would cost
  memoization and mark ledger entries provisional in every turn-based game.
- **Ledger `FieldDependent`.** A `player.X` entry is reported field-dependent
  exactly when the admin's on-behalf `TargetPlayerIndex` supplied its player
  in that evaluation, which is when editing that field really would change
  the verdict.
- **Per-viewer evaluability.** `LegalReadEvaluable` sees a read, not a move,
  and takes a viewer, not a trusted proposer. So each ledger entry now
  carries `MovePlayerIsProposer`. It is true when the move names its own
  player and its plan has no `proposerIsCurrentPlayer` check tying that
  player to the current one (an `AnyPlayer` move). The server calls
  `LegalVerdictEntry.ReadsEvaluable`, which judges such reads, and every
  `player.X` read under a wildcard current player, like `proposer.X`:
  conservatively, across every player. With a concrete current player and a
  `CurrentPlayer` or plain move, evaluability is judged against the current
  player as before. Admin viewers remain omniscient, and an Observer current
  player is not evaluable for moves without a player of their own.
- **Resolved once per evaluation.** When a plan reads `player.*` anywhere,
  `legalPlan.evaluate` resolves the move's player once, while the whole move
  is still in hand, and pins it on `LegalContext`. `ctx.ResolvePath` reads
  the pinned index, so a predicate never needs `ctx.Move` to resolve
  `player.X`. This keeps the change compatible with withholding the move
  from predicates that have no move-dependent Reads (the enforced-Reads
  work). Plans that never read `player.*` do not call `CurrentPlayerIndex`
  at all, so they are unchanged. If the delegate panics during resolution,
  the context stays unresolved and the affected predicates fail closed
  inside `evalLegalPredicate`'s recover, as before.
- **Moves stack paths.** `moves/stack_path.go` has its own `player.` kind
  (`WithSourceProperty("player.Hand")`). It used to resolve literally to the
  current player, so after the legality change a `MoveComponentToSlot` in a
  simultaneous phase passed legality and then failed in `Apply`. It now uses
  the same rule: the concrete current player, else the move's
  `LegalOnBehalfOf()`. `Apply` receives no proposer, so step 2 is absent
  there. That is sound for `CurrentPlayer`/`AnyPlayer` moves, whose
  legality already tied the target to the proposer. A move with no
  on-behalf player still finds no `player.` during simultaneous play, and
  should name one with `players[move.Field]`.
  `TestStackPathPlayerAgreesWithLegalityDuringSimultaneousPlay` fails
  without this fix.
- **Boot probe.** The undeclared-move-read probe proposes as Observer and
  wraps the move in a sentinel type that does not forward
  `LegalOnBehalfOf`, so the new resolution never touches the sentinel's
  properties.

## Alternatives considered

1. **Status quo: authors pick `CurrentPlayer()`, `Proposer()` or
   `PlayerFromMove`.** Rejected. It exposes a distinction authors do not have
   in their heads. The right choice depends on a delegate method rather than
   the move, and `Proposer()` gives the wrong answer for admin. Reproduced
   above.
2. **A new `actor.X` path / `legal.Actor()` selector.** Rejected as a new
   concept. It would be what `player.X` should have meant, and it would leave
   `player.X` as a trap that only fails in simultaneous phases.
3. **Always resolve `player.X` to the move's `TargetPlayerIndex` when the
   base type has one.** Close to what shipped, but not the same. The
   shipped rule uses the *proposer* when it is a concrete player and reads
   `TargetPlayerIndex` only for an admin. So `player.X` stays
   field-independent (memoized, not provisional) except in the
   admin-on-behalf case. Moves without a player of their own keep the
   current-player rule.
3a. **Consult a concrete current player first, for every move** (the first
   shipped version). Rejected after adversarial review: an `AnyPlayer` move
   made out of turn, or in an Observer gathering phase, read the wrong
   player or nobody.
4. **Rewrite `player.X` to the proposer (dropping the current player).**
   Rejected. It changes turn-based semantics for `moves.Default` moves read
   by a non-current proposer, and still fails for admin.
5. **Resolve by convention from any `TargetPlayerIndex` field.** Rejected in
   favor of an explicit hook. Fix-ups such as Darwin's `moveDealCard` use a
   field with that name for a *recipient*, not an actor. Only base types
   whose contributed atom ties the target to the proposer implement
   `LegalOnBehalfOfer`.

## Layering

- **Core (`legal_path.go`).** `legalMovePlayerIndex` and `LegalOnBehalfOfer`
  (exported because `moves` implements it, but authors never touch it). The
  wire grammar is unchanged: no new path kind and no serialization change.
- **`moves`.** `CurrentPlayer` and `AnyPlayer` implement `LegalOnBehalfOf()`.
  Embedders inherit it, so `DrawToPlayer`, `MoveComponentToSlot` and any game
  move get it for free.
- **`legal`.** No catalog change. The selector is `Player()`, with `CurrentPlayer()` kept as a deprecated alias; its doc now describes the
  move's player.

## GAMES migrations (step 3)

- **Darwin's six Play Cards moves** (`moveIncreasePopulation`,
  `moveIncreaseBodySize`, `moveAddTrait`, `moveRemoveTrait`,
  `moveDonePlayingCards`, `moveTradeCards`) move from
  `PlayerFromMove("TargetPlayerIndex")`/`players[move.TargetPlayerIndex].X`
  to `Player()`/`player.X`. They are fenced by Darwin's existing
  golden tests, which include the admin proposer and an AnyPlayerIndex
  current player.
- **`moveDealCard`** keeps `players[move.TargetPlayerIndex]`: its target is a
  recipient, not an actor.
- **`moveSubmitFoodCard`** (Select Food, `moves.AnyPlayer`). The seam
  session put `AnyPlayer` on the composition seam and migrated it with
  `players[move.TargetPlayerIndex].X`. After rebasing onto that, it reads
  `StackEmpty("player.SubmittedFoodCard")` and
  `ComponentPresentAt("player.Hand", "move.TargetCard")`, fenced by
  `darwin/legal_golden_submit_food_test.go`.
- **The species component path** shared by the Play Cards moves
  (`targetSpecies`, added by the dynamic-values session) is now
  `legal.MoveIndex("player.Species", "move.TargetSpecies")`.
- **Other games:** none have a wildcard phase; see the survey below.

## Related issue history (jkomoros/boardgame)

I searched both issues and discussions. Discussions hold only the welcome
post (#776).

- **#761, Allow declarative legal methods (closed).** This is the umbrella
  for this work. It *advances* it by removing the last reason a
  simultaneous-phase actor check had to stay imperative or use a
  workaround.
- **#189 (reject obviously illegal moves client-side) and #213 (client
  proposeMove bails early).** *Advances.* Simultaneous-phase `player.X`
  checks used to be Unknown everywhere. They now produce real verdicts, and
  they are client-evaluable whenever the read survives every player's
  sanitization, under the conservative rule above.
- **#266 (in admin mode, moves are made on behalf of the viewing-as player),
  #214 and #224 (moves can't be made on behalf of another player; a move is
  made for its `TargetPlayerIndex`).** These established that
  `TargetPlayerIndex` is the player a move is *for* and that admin acts *on
  behalf of* that player. Step 3 (`LegalOnBehalfOfer`) is exactly this
  established semantic, which is the main argument for on-behalf over
  `AdminBypass`.
- **#56, Should CurrentPlayerMove be a base struct (closed).** This is where
  `moves.CurrentPlayer` came from: "a move that can only be applied when the
  TargetPlayerIndex is current player". Together with the documented
  `AnyPlayerIndex` wildcard, it is why "current player, narrowed" is the
  existing concept rather than a new one.
- **#345, "not your turn" messages.** Unaffected. The actor gate is still
  `proposerIsCurrentPlayer`, with its existing messages.
- **#732, Consider a moves.CurrentPlayerFixUp.** *Related, not closed.* A
  CurrentPlayer-shaped fix-up is proposed by admin on a player's behalf.
  If it is added, it should implement `LegalOnBehalfOfer`, so that its
  `player.X` reads resolve to that player during simultaneous play.
- **#473 and #640 (Legal performance).** Not regressed. `player.X` stays in
  the memoized bucket. Plans that never read `player.*` make no extra
  delegate call. Plans that do make one `CurrentPlayerIndex` call per
  evaluation, where previously they made one per predicate.
- **#65, #624, #736, #644, #295.** Checked; not related to this limit.

## Survey of consumers (step 3)

The only phases with a wildcard current player (`AnyPlayerIndex`) are
Darwin's Select Food and Play Cards in GAMES, and werewolf's Day and Night
in boardgame. Every other game in GAMES (metaltrader, murdermrmonroe, pass,
secretgroups, sequenceforge, terrainweave, valentine) always has a concrete
current player. That includes murdermrmonroe's `CurrentLuckPlayer` during
attacks. For them the change is a no-op.

The remaining `PlayerFromMove` and `players[move.X]` uses all name a
*target* rather than the actor, so they are correct as written:

- Valentine's `OtherPlayerIndex`;
- Valentine's out-of-turn card return, where the current player is concrete
  but is not the actor;
- Darwin's `moveDealCard` recipient;
- sequenceforge's discard destination, which is a moves stack path, not a
  legality path.

Every move the change makes expressible therefore falls into one of two
groups:

- **Migrated:** Darwin's six Play Cards moves.
- **Simplified after the seam landed:** the seam session migrated Darwin's
  `moveSubmitFoodCard` and werewolf's `moveCastVote` (both `moves.AnyPlayer`)
  with `players[move.TargetPlayerIndex].X`. After rebasing, both read plain
  `player.X` / `legal.Player()`. Their snapshot golden tests
  (`darwin/legal_golden_submit_food_test.go`,
  `examples/werewolf/legal_golden_test.go`) sweep every proposer, admin
  included, and both fail if the admin on-behalf step is removed.

## Adversarial review (pre-landing)

An adversarial reviewer was asked to make the strongest case that these
semantics confuse authors. Outcomes:

- **Fixed: legal-but-Apply-fails.** `moves.DrawToPlayer` contributes a
  `player.` precondition, but its `Apply` read `state.CurrentPlayer()`, which
  is nil during simultaneous play. It now resolves through the same
  move's-player rule as stack paths (`stackPathMovePlayerState`).
  `TestDrawToPlayerAgreesWithLegalityDuringSimultaneousPlay` fails without
  the fix. `moves.Increment` also reads `state.CurrentPlayer()`, but it is a
  fix-up with no on-behalf player, so its legality and its `Apply` both fail
  closed together.
- **Fixed, docs: admin and out-of-turn moves.** `Proposer()` helpers *exempt*
  admin (AdminBypass), whereas `player.X` checks the player admin acts for.
  The docs now point out-of-turn moves at
  `PlayerFromMove("TargetPlayerIndex")` + `ProposerIsPlayerFromMove`
  (Valentine's existing shape) instead of `Proposer()`, and they no longer
  claim `player.X` covers every phase.
- **Resolved: `AnyPlayer` in a turn-based phase.** The first rule started
  from a concrete current player, so an `AnyPlayer` move proposed by a
  non-current player read the current player's state. The rule is now "the
  player the move is made for" (see Decision). For `CurrentPlayer` moves it
  gives the same answer wherever the turn check passes. Ledger entries mark
  `player.X` reads of such moves `MovePlayerIsProposer` when nothing ties
  their player to the current one, and
  `LegalVerdictEntry.ReadsEvaluable` judges those reads like `proposer.X`.
- **Resolved: naming.** `legal.Player()` is the selector, matching the
  `player.X` wire spelling, so no new path kind is needed.
  `legal.CurrentPlayer()` remains as a deprecated alias.
- **Held up under attack:** memo and ledger keying; `AnyPlayer`'s
  target auto-correct (Legal and Apply share one move instance).

## Open questions

- Should `Proposer()` be documented as "rarely needed" or folded away
  entirely? Today it has no production users. I kept it, because out-of-turn
  turn-based moves are a real shape.
- Should the client (TypeScript) evaluator, once it exists, mirror
  `legalMovePlayerIndex`? It needs the proposer, which the viewer-keyed
  evaluability API does not trust today. The conservative evaluability rule
  above is what keeps this safe.
- `AnyPlayer` on the seam: once it is allowed, `moveSubmitFoodCard` and other
  `AnyPlayer` moves become fully declarative with plain `player.X`.
