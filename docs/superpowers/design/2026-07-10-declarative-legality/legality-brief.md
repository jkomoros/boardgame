# Design Brief: Declarative Move Legality for jkomoros/boardgame

You are designing a declarative move-legality system for a Go boardgame framework
(github.com/jkomoros/boardgame). This brief is self-contained; the repo is at
/Users/jkomoros/Code/go/src/github.com/jkomoros/boardgame if you need to read code.

## The problem

`Move.Legal(state ImmutableState, proposer PlayerIndex) error` is imperative Go.
The engine can't optimize it (the fixup loop and the server's per-request
move-forms computation call every move's Legal() after every state change — issue
#640), the client can't evaluate it (moves can't be grayed out or rejected
without a round trip — #189/#213), and nothing can explain it (error strings are
ad-hoc — #65). Issues #761 and #189 are the same 9-year-old design thread.

## Quality bar (the user's words)

"Super elegant and delightful and well layered and explainable."

## Locked decisions

- **PURELY SUGAR (added mid-panel, binding at synthesis)**: the declarative layer
  must be fully skippable. `Legal(state, proposer) error` remains the ground-truth
  contract; a game author can ignore preconditions entirely and write imperative
  Legal() as today with zero new required concepts. When declared, preconditions
  are consumed by the default Legal() implementation (declare = implement, never
  both), and every engine capability (bucketing, caching, client evaluation,
  structured errors) degrades gracefully to "opaque move — call Legal()" when
  declarations are absent.

- **Design for client evaluation, ship server-first**: the representation must be
  serializable and sanitization-aware from day one; a TypeScript evaluator is a
  designed-for follow-up, not in this campaign.
- **Break the Go API freely**: all consumers (in-repo examples + 3 client games
  in ../games) get migrated. Legal() can change shape or become optional.
- Related issues folded in or constrained by this design: #790 (move-level stack
  preconditions), #644 (state-dependent Repeat counts in progressions), #693
  (move structs can carry hidden info — sanitization of moves), #44 (propose-time
  vs apply-time move fields), #65 (log/explain when fixups aren't legal).

## What exists today (verified code map)

### The Legal() chain
- `moves.Default.Legal()` (moves/default.go:339) already dispatches THREE
  declarative-ish checks stored in a config property bag:
  1. `legalInPhase` (:425-469) — phases from `WithLegalPhases(...)` config;
     TreeEnum phases match leaf OR ancestors.
  2. `legalMoveInProgression` (:561-631) — the "move tape": moves since the last
     phase transition matched against a combinator tree of
     `Serial/Parallel/ParallelCount/Repeat/Optional` groups (moves/groups.go),
     with `ValidCounter` matchers (CountAll/CountExactly/CountBetween/...).
  3. `legalStackConstraints` (:359-388) — if `WithSourceProperty`/
     `WithDestinationProperty` configured, checks `first.MayMoveTo(dstStack)`.
- `moves.CurrentPlayer.Legal()` adds proposer≡TargetPlayerIndex≡CurrentPlayer.
- Game moves embed these and chain: `if err := m.CurrentPlayer.Legal(state,
  proposer); err != nil { return err }` then custom checks.

### Auto-config idiom (the established DX)
`auto.Config(new(moveType), WithMoveName(...), WithLegalPhases(...),
WithLegalMoveProgression(Serial(...)), WithSourceProperty("DrawStack"), ...)`.
With* options are `func(boardgame.PropertyCollection)` writing namespaced keys.
Move reads config at runtime via `CustomConfiguration()`.

### The constraints package (pattern to rhyme with, landed recently)
`StackConstraint func(dest ImmutableStack, proposed []ImmutableComponentInstance,
state ImmutableState) error`; a `StackConstraintConstructor` registry with
`Name string; Constructor func(args []string, chest *ComponentChest)` enabling
struct-tag syntax like `stack:"cards,max(3),unique(color)"`. Serializable-by-name
constraints with string args.

### Legal() invocation paths (the perf surface)
- game.ProposeMove → applyMove → Legal before Apply.
- Fixup loop: after EVERY move, delegate.ProposeFixUpMove(state) polls candidate
  fixups' Legal() repeatedly (up to 256 recursions).
- Server move-forms (server/api/main.go:1590-1629): for every /info request and
  every state version, every non-fixup move runs Legal() TWICE (as player, as
  admin) to compute LegalForPlayer/LegalForAnyone/LegalForPlayerError. Not cached.
- Agents propose like players.

### Client pipeline today
MoveForm{Name, HelpText, Fields, LegalForPlayer, LegalForPlayerError,
LegalForAnyone} → Redux → memoized selectMoveLegality → renderers. Refreshes on
every state version. Client has: sanitized state JSON, chest (enums/constants/
decks), move form fields with defaults. Client lacks: any logic.

### Sanitization (constraint on client evaluability)
Per-property policies: Visible/Hidden/Len/Order/NonEmpty, driven by delegate
SanitizationPolicy + group membership. A hidden property arrives zeroed — a
client-side evaluator cannot distinguish "0" from "hidden 7". Any client story
must handle unknown-ness honestly (three-valued logic or visibility metadata).

### Move structs
Move properties are Int/Bool/String/PlayerIndex/Enum + slices (no stacks/timers).
`DefaultsForState(state)` sets smart defaults server-side before form render.
Note #761's observation: preconditions split into two classes — those independent
of move field values (phase, turn) and those reading move fields (CardIndex in
range); the former can be checked before DefaultsForState/field-binding, the
latter after.

### Real Legal() bodies to migrate (the acid test — design must show before/after)

1. memory/moveRevealCard (after CurrentPlayer chain): `p.CardsLeftToReveal < 1` →
   "You have no cards left to reveal this turn"; component-at-index nil checks →
   "there is no card at that index" / "that card has already been revealed";
   then `c.MayMoveToSlot(game.VisibleCards, m.CardIndex)`.
2. blackjack/moveStartRoundCleanup (after StartPhase chain): loop over players:
   all active players must be Eliminated || Stood → "not all active players have
   finished their turn".
3. checkers/moveMoveToken: token ownership (`p.Color.Equals(t.Color)`) → "that
   token isn't your token to move"; `spaceIsBlack(m.SpaceIndex)` → "you can only
   move to spaces that are black"; then capture-space graph logic (genuinely
   gnarly — likely stays imperative).

Survey across examples: ~5 phase checks, ~8 current-player, ~6 stack size/
presence, ~4 property comparisons, ~3 MayMoveTo pre-checks, ~2 genuinely custom
(blackjack hand value, checkers capture graph). The design must make the first
five categories declarative AND keep a first-class imperative escape hatch.

## Design questions your proposal MUST answer

1. **Representation**: What is a precondition? Go type(s), how constructed, how
   serialized (name+args like constraints? an AST? both?). What can it reference
   (state property paths, move fields, proposer, phase, chest constants, computed
   properties)? Where's the line before it becomes a Turing tarpit?
2. **Attachment & composition**: How do preconditions attach to a move type —
   With* options? methods? both? How do they compose down the embedding chain
   (Default → CurrentPlayer → game move)? Can a subclass remove/override an
   inherited precondition? How does this interact with auto.Config?
3. **Layering**: What lives in core `boardgame` package vs `moves` vs a new
   package? (#761's instinct: core gets only minimal interfaces.) Draw the
   dependency arrows. The existing three buried checks (phase/progression/
   constraints) should land somewhere principled.
4. **Explainability**: The error model. Structured failures (which precondition,
   with what bindings), player-facing message templates with placeholders,
   fixup-loop logging (#65). Show what the client receives.
5. **Evaluation semantics & the escape hatch**: Ordering (cheap→expensive?
   field-independent → field-dependent per #761?), short-circuit, determinism.
   Imperative residue: Legal() (or equivalent) stays for the checkers capture
   graph — how does it coexist, and what does the client see for it ("unknown")?
6. **Sanitization-aware client story** (designed-for, not shipped): three-valued
   evaluation? per-precondition visibility computation? How does the server tell
   the client which preconditions it can evaluate locally?
7. **Engine wins**: How does the fixup loop / move-forms computation get faster?
   Indexing by phase? Dirty-tracking by referenced property paths? Be concrete
   about what becomes O(cheap) and what stays O(Legal).
8. **Progression**: Does move-tape matching become "just another precondition"
   or stay separate? What about #644 (Repeat counts from state)?
9. **Migration**: Show the three acid-test moves rewritten in your design,
   verbatim, with their error messages preserved or improved. This is where
   "delightful" is proven or disproven.

## Output contract

Write your complete design to the file path given in your prompt. Structure:
Executive summary (≤15 lines), then sections answering the 9 questions in order,
then "Risks & open questions" (honest). Go code for every API you propose; the
three migrated moves in full. Target 300-500 lines. Your final chat message:
10 lines max — file path + the 3 boldest choices you made.
