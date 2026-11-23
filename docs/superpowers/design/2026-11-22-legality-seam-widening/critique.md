# Adversarial critique of the seam-widening draft (2026-11-22)

Reviewer: an independent model pass (Fable), read-only against the live branch,
asked to interrogate the draft's assumptions rather than confirm them. Nothing
was executed; findings are from reading code. Verdict: **sound with
amendments**. The spec (`docs/superpowers/specs/2026-11-22-legality-seam-widening-design.md`)
records how each finding was disposed.

1. **Uniform suppression silently deletes security checks.** Existing
   doctrine (moves/preconditions.go `requireDefiningPreconditionReauthored`)
   is that suppressing a defining atom is for *reordering*, not deleting.
   Framework-check atoms had no exported constructor, so suppressing one
   could only delete it — and the deleted checks include ForceFinishTurn's
   admin-only gate, SeatPlayer's admin gate, AnyPlayer's proposer≡target and
   its target auto-correct (whose absence makes Apply index
   `PlayerStates()[Observer]`). Amendment: every framework-check atom is
   defining; suppression requires re-authoring.
2. **"No Reads ⇒ unknown ⇒ never memoized" is the wrong layer.** It does not
   compose through `any` (reads are unioned), it weakens the F3
   undeclared-move-read probe (which only walks the field-independent
   bucket), and an empty read-set is vacuously "evaluable". Core already has
   the concept — the unexported `opaque` flag on the LegalCustom wrapper.
   Amendment: a core constructor for named opaque atoms, bucketed
   field-dependent because they are opaque. Also: be honest that these are
   *contributed LegalCustom*, not rule 2.
3. **The plan is determined by type; the frozen chain by call path.** A game
   override that super-calls a *non-immediate* ancestor (documented as
   legitimate in moves/move_components.go) would gain the skipped level's
   check on opt-in. The boot probe only proves Default was reached.
   Amendment: each level's guard reports itself during the probe; fail boot
   if a contributed level was never entered.
4. **Full-ledger mode runs every atom after a failure,** but residues assume
   their predecessors passed (AdminPlayer indexes an unvalidated target,
   DealCountComponents the next player, SeatPlayer hits storage). Amendment:
   opaque atoms after the first non-Pass report Unknown "not evaluated" in
   ledger mode. (Unsure whether the existing custom tail's identical
   behavior is intentional.)
5. **The survey is too pessimistic exactly where client evaluability
   matters.** Fix-up types are never rendered as forms, so opacity costs
   nothing there. The player-facing family (AnyPlayer, AdminPlayer, Select*)
   is mostly expressible with today's catalog (`ProposerIsPlayerFromMove`,
   `PlayerSeatIsFilled(PlayerFromMove(...))`, `PlayerIsAdmin`); the real
   blocker is AnyPlayer mutating `TargetPlayerIndex` inside Legal, which
   also makes authored `players[move.TargetPlayerIndex]` leaves see the
   sentinel on a client. AllPlayersSubmitted is
   `AllActivePlayers(PlayerHasSubmitted)` modulo message.
6. **Enforcement and golden audit are weaker than they look.** A
   source-parse presence check does not prove correspondence; a shadow-plan
   audit shares the residue code with the frozen chain and never exercises
   the real opted-in path (short-circuit, RoundRobin recursion,
   ForceFinishTurn entry). Amendment: differential comparison on the real
   opted-in manager — frozen vs plan on the same state — over proposers
   (Observer, Admin, each player), move-field variants, and two-failure
   states; tighten the AST test to a strict body template.
7. **Mechanism details mostly check out.** Dropping RoundRobin's second
   Default.Legal is unobservable (pure apart from idempotent memos) but does
   edit the frozen path — say so. Plans are stored before the probe. Per-type
   private residue names are unshadowable. Corrections: FinishTurn
   super-calls `f.Default.Legal`; ForceFinishTurn embeds FinishTurn, so its
   contribution must drop inherited inPhase/inProgression/finishTurn.
8. **Missed alternative: always-plan as the destination.** DrawToPlayer
   already forces `LegalPlanEnabled()` and declares no Legal(). Doing that
   everywhere deletes the dual path, but costs a ledger wire change for
   un-migrated games, panic→Unknown, new boot validations on existing games,
   and memo overhead in the fix-up loop. Recommend naming it as the endpoint
   and treating the guard as scaffolding.
