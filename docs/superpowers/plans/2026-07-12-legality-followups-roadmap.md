# Declarative-Legality Follow-ups Roadmap

> **For agentic workers:** This is a ROADMAP, not a single implementation plan.
> It sequences every remaining workstream after the footgun batch, states for
> each one whether it is **plan-ready** (write a plan per
> superpowers:writing-plans, then execute per
> superpowers:subagent-driven-development) or **spec-first** (run
> superpowers:brainstorming → spec → sub-agent critique round → plan, matching
> the process used for the two shipped specs), and points at the checked-in
> exploration report section containing the full design detail. Do not invent
> designs that the cited reports already contain.

**Read these first (in order):**
1. `docs/superpowers/specs/2026-07-10-declarative-legality-design.md` — the normative v1 spec (prime guarantee rules 1–4).
2. `docs/superpowers/specs/2026-07-12-legality-completeness-design.md` — the A-round spec.
3. `docs/superpowers/plans/2026-07-11-declarative-legality-runbook.md` — conventions: gates, exclusions, ledger, golden-fence patterns, end-game protocol.
4. The four exploration reports in `docs/superpowers/design/2026-07-12-legality-footgun-followups/` — the design inputs cited per workstream below.

## Global conventions (apply to EVERY workstream)

- **Branch:** work stacks on `declarative-legality-design` unless the user directs otherwise. Do NOT merge to master without explicit user confirmation (runbook end-game protocol).
- **Purely-sugar guarantee:** the frozen imperative chain must stay byte-for-byte. `go test ./moves/ -run TestLegalChainStringFreeze -count=1` must pass unmodified after every task. Never change behavior for non-opted-in moves.
- **Gates:** `go test ./...` in the boardgame repo, excluding the three known pre-existing failures (`boardgame-util/lib/build/api`, `boardgame-util/lib/gamepkg` — GOPATH sandbox; `storage/mysql` — no local DB). Plus `go test ./...` in `/Users/jkomoros/Code/go/src/github.com/jkomoros/games` (excluding valentine `TestGolden` — testdata deleted upstream). The games repo has a `replace` directive pointing at `../boardgame`, so framework changes are live there; that replace must never ride to a published release.
- **Git discipline:** stage NAMED FILES ONLY (`git add <file>...` or `git commit -- <file>...`). NEVER `git add -A`/`git add .` — the working tree carries session noise (.database files, temp dirs, game-src symlinks). If dispatching parallel agents on one worktree, serialize their commits or require explicit-pathspec commits (a shared-index sweep incident is on record).
- **TDD:** failing test first, prove red, implement, prove green. Adversarial review agent per task; fix Criticals/Importants before proceeding.
- **Ledger:** append a row per task to `.superpowers/sdd/progress-legality.md` (git-ignored working ledger; do not force-add).
- **Migrations:** every move migration gets a golden-equivalence fence (legacy `Legal()` oracle copy × states × proposers, message-string assertions, named divergence maps) — copy the pattern from `../games/darwin/legal_golden_test.go` or `examples/memory`.
- **New catalog predicates:** each needs a conformance corpus file (`legal/testdata/conformance/<name>.json`, ≥3 cases, "template" pinned on fails — the completeness meta-test enumerates the registry and will fail until present), a `Template*` constant with default body, `EmittedTemplates` AND `EmittedBindings` metadata (boot-validated since the footgun batch), Reads with honest facets, and a `legal/doc.go` catalog-table row.
- **Boot errors:** name the offending move type; match existing gauntlet error style in `legal_plan.go`.

---

## Workstream 0 — Finish the footgun batch (✅ COMPLETE)

**DONE** (commits `1e2dfce4`, `42102032`, `7ee43c7a`; do NOT merge — awaits explicit user confirmation). Task-4 review minors landed (`1e2dfce4`); Task-5 whole-batch verification passed (full gates green both repos, `TestLegalChainStringFreeze` unmodified, F1–F10 disposition audit done). **The audit turned up a real gap: F5 (LegalCustom-without-opt-in) was mis-annotated as fixed — its `[BATCH: FIXED in 72b981b4]` note actually described F2. F5 was documentation-only (fail-open) and was genuinely boot-enforced this session in `42102032` (TDD + adversarial review, both no-plan paths, zero false positives). Annotation drift for F5/F9 corrected in `7ee43c7a`.** See `.superpowers/sdd/progress-legality.md` for the full disposition audit.

Plan: `docs/superpowers/plans/2026-07-12-legality-footgun-batch.md`. Status at time of writing: Tasks 1–4 implemented, reviewed (all APPROVE), commits `72b981b4`, `3586f278`, `1d82dfba`, `63519c73`, `e195c947`, `b96e13b6`, `77d36a45`, `06419dbf`, `3f11d57c`, `5f1ab043`, plus a Task 3 minors polish commit (in flight at writing time). Remaining (now all done):

1. **Task 4 review minors** (all Minor, none blocking):
   - Soften the surviving overclaim in `moves/legal_plan_test.go` (~line 603 doc comment: "the frozen imperative chain still enforces the check the suppression names") to match the corrected M1 message wording.
   - Delete (preferred, per reviewer) or test the runtime TypeInt tolerance branch in `legal_path.go` (~268-277) — it is untested, unreachable from boot-validated paths, and preserves a silent wrong-player read where an error would fail closed.
   - Update the stale coverage comment above `validateLegalTemplates` in `legal_error.go` (~316-334) claiming EmittedTemplates is "not implemented" — check the polish commit's edits to that file first to avoid double-fixing.
2. **Task 5 whole-batch verification** exactly as written in the batch plan: full gates both repos, explicit `TestLegalChainStringFreeze`, F1–F9 disposition audit (each fixed/documented/deferred — the checked-in audit report has [BATCH: ...] status annotations to verify against), ledger summary row. Do NOT merge.

## Workstream 1 — Codegen path constants, Phase 1 (plan-ready)

**Input:** `report-compile-time-safety.md`, Proposal 1 Phase 1 (full generated-code sketch, magic-comment design, cost estimate, constraints section).
**Goal:** `boardgame-util codegen` emits `auto_paths.go` with untyped `const` path strings for game/player/move structs; path typos become compile errors. Zero changes to the `legal` package; raw strings keep working.
**Mode:** plan-ready — the report contains the design; write the implementation plan directly.
**Key facts the plan must honor:** codegen's `fieldsInfo` map in `boardgame-util/lib/codegen/reader.go` already has field→PropertyType including cross-package embeds; game-vs-player kind needs a NEW magic comment (`//boardgame:legalpaths game|player`) because codegen is syntactic and `structConfig` silently ignores unknown args on the existing comment; move structs are auto-detected via their embedded `moves.*` type; emit `players[*].X` variants for player kind; constants are package-private (fine — `ConfigureMoves` is same-package).
**Done bar:** memory (flagship) regenerated and its `main.go` preconditions migrated to constants; a deliberate typo in a constant name fails `go build`; codegen tests cover both annotated and unannotated packages; docs: TUTORIAL.md section + codegen README mention.
**Size:** ~300–500 LOC in codegen + tests. Follow-on (separate, later): Phase 2 typed paths — do NOT start it until Phase 1 has user feedback.

## Workstream 2 — `WithMessagef(key, body)` colocated templates (plan-ready)

**Input:** `report-compile-time-safety.md`, Proposal 3 Tier 1 (and Tier 0). The rejected auto-derived-key variant is documented there — do not implement it.
**Goal:** declare the template body next to its key: `legal.PropAtLeast(...).WithMessagef("reveal.no_cards_left", "You have no cards left to reveal this turn")`; `ConfigureLegalTemplates` becomes unnecessary for one-off messages.
**Mode:** plan-ready.
**Key facts:** `LegalSpec` gains `MessageBody string` excluded from wire JSON (frozen-wire test must stay green); boot overlays collected bodies onto `DefaultTemplates()` BEFORE `ConfigureLegalTemplates` (delegate keeps last word); the F4 placeholder/binding boot validation applies to these bodies too (test it); two same-key `WithMessagef` declarations with different bodies = boot error (ambiguity); Tier 0 (tutorial teaches shared constants for key pairing) ships in the same change.
**Done bar:** memory's `reveal.no_cards_left` migrated as the demonstration (delete its `ConfigureLegalTemplates` entry); TUTORIAL.md updated; frozen-wire + string-freeze green.
**Size:** ~100–150 LOC + docs.

## Workstream 3 — Behavior predicates v1: `legal/catalog_behaviors.go` (plan-ready)

**Input:** `report-behavior-preconditions.md` §3(a) + §4 + §6 "Minimal v1" (predicate list, Reads/facet guidance, template names, migration targets). Also `report-idiom-survey.md` item 2.
**Goal:** `legal.SeatFilled(sel)`, `legal.PlayerActive(sel)`, `legal.PlayerNotEliminated(sel)`, `legal.PlayerEliminated(sel)`, `legal.ProposerIsAdmin()`, `legal.HasMoveBudget(sel)` — thin, Reads-honest predicates over the behaviors' canonical property names, with default templates.
**Mode:** plan-ready.
**Key facts:** evaluate via raw property paths (client-evaluable), not interface assertions; boot path-validation already errors when the behavior isn't embedded (conservative, correct); polarity is ALWAYS the author's choice — no auto-attachment (the report's §3(b) rejection is binding); include a test with a sanitization-hidden behavior field (Role-style `other:hidden`) proving the ledger's evaluable computation handles it.
**Done bar:** all six predicates with conformance rows + templates + EmittedBindings; migrate blackjack Stand/Hit gates (report §4 has the exact before/after) and valentine/murdermrmonroe MoveBudget gates behind golden fences; stale "no negation primitive" comment in blackjack discharged; legal/doc.go table updated.
**Size:** day-scale.

## Workstream 4 — Behavior contributions v2: widen seam to AnyPlayer/AdminPlayer (SPEC-FIRST)

**Input:** `report-behavior-preconditions.md` §3(d) + §5 — read it in full; it contains the seam-invariant change, the conditional-contribution plumbing decision (omit-if-behavior-absent at plan assembly vs. `ContributedPreconditionsForState`), and the suppression-divergence guard requirement.
**Goal:** `moves.AnyPlayer`/`moves.AdminPlayer` contribute `proposerIsTargetPlayer`/`targetSeatFilled`/`targetIsAdmin` atoms, unblocking werewolf-shaped games.
**Mode:** SPEC-FIRST — this changes the seam allowlist rule ("no Legal() override") and needs the CurrentPlayer-style equivalence treatment; run a short brainstorm → spec → critique round. Depends on Workstream 3 (the atoms are its predicates).
**Blocking design questions for the spec:** (1) omit-at-assembly vs. state-aware interface for conditional contribution; (2) exact golden-equivalence obligation vs. the frozen `AnyPlayer.Legal` override (byte-for-byte message match required — the seam test must be extended, not weakened); (3) suppression boot-guard generalization per widened type (the `legalMoveEmbedsCurrentPlayer` precedent).
**Done bar (for the eventual plan):** werewolf `moveCastVote` migrated per report §4's sketch (six of eight checks declarative, disjunction in LegalCustom), golden-fenced; seam source test updated; `LegalSupportedMovesBaseTypeNames()` grows.

## Workstream 5 — Component-values predicates (SPEC-FIRST, biggest migration unlock)

**Input:** `report-idiom-survey.md` §1 Cluster B + Top-5 item 1 (recurring shapes, before/after sketch, static-vs-dynamic scoping).
**Goal:** ~3 predicates over STATIC chest component values: `ComponentPropEquals/NotEquals(stackPath, indexOrField, prop, want)`, `ComponentsPropMatch` (memory's same-type pair), one cross-path count-vs-component-prop compare (metaltrader). Unblocks memory, metaltrader, valentine (nine Activate moves), parts of darwin.
**Mode:** SPEC-FIRST (short) — design questions: arg shapes (literal index vs move-field vs key), enum-value validation at boot (the predicate-constructor example-state hook from Workstream 7d would close the typo→Unknown gap — decide ordering), facet honesty for chest reads (chest is static/visible: FacetValues on the stack, chest values are not sanitized — verify), TS-evaluator implications (chest ships client-side already). Dynamic component values stay OUT of scope (documented durable blocker).
**Done bar:** predicates + corpus + templates; migrate memory's card-type comparison and at least valentine's `baseActivateMove` family behind golden fences (report Cluster F explains why that family is high-value).

## Workstream 6 — Structured messages on the API error envelope (SPEC-FIRST)

**Input:** `report-idiom-survey.md` §2 items 1–4 + Top-5 item 3 (exact seams: `game.go:966`, `renderer.Error()` at `server/api/main.go:321-351`, host/join/seat endpoints list; the `errors.Friendly` triple).
**Goal:** ProposeMove rejections (then host/join/seat endpoints) carry `Message:{template,bindings}` alongside the existing string fields; client can render/localize from the chest's template table.
**Mode:** SPEC-FIRST (short) — wire-contract addition; coordinate with sub-project B (the TS renderer is shared work). Old fields stay byte-identical (additive only).
**Also in scope:** delete the vestigial `lastErrorMessage` (`server/api/main.go:46` — only ever cleared, never set; verify before deleting).

## Workstream 7 — Boot gauntlet round 2 (plan-ready, itemized)

**Input:** `report-idiom-survey.md` §3 items 1–5 + Top-5 item 4. Ordered by cost/benefit; each is independently shippable — one plan with one task per item:
- (a) Unconditional `WithLegalPhases` key validation against `PhaseEnum()` in `Default.ValidConfiguration` (`moves/default.go:113-153` currently only validates when a progression coexists). Hours.
- (b) Recursive behavior connect/validate walkers (`state.go:975-995`, `game_manager.go:437-463` don't recurse into nested embeds → nil panic at first move). Small.
- (c) Agent-name collision boot check (`game_manager.go:400-403` silently last-write-wins). Trivial.
- (d) Example-state validation hook for predicate constructors — closes the documented enum-typo→Unknown gap (`legal/doc.go` limits section). Design note: constructors currently lack exampleState by design (lazy dispatch); the hook is a separate optional interface consulted at boot, NOT a constructor signature change.
- (e) Throwaway-game setup-hook probe (run `DistributeComponentToStarterStack`/`BeginSetUp`/`FinishSetUp` at boot via a `DefaultNumPlayers` game). Verify no side effects on storage (use a discard storage manager).
- (f) Progression reachability probe (the `probeLegalReachable` idiom applied to progressions; most speculative — spike first).
**Mode:** plan-ready for (a)(b)(c); (d)(e)(f) each need a one-page design note before their task.

## Workstream 8 — Quantifier v2 (SPEC-FIRST)

**Input:** `report-idiom-survey.md` Cluster D + Top-5 item 5.
**Goal:** `AnyActivePlayer` (EXISTS dual), inner-leaf support for enum/PlayerIndex/stack-count reads, failing-player capture into message bindings.
**Mode:** SPEC-FIRST — Kleene EXISTS semantics need the same exhaustive conformance treatment `any` got; binding capture changes the message model slightly.
**Unblocks:** blackjack cleanup moves, werewolf `moveResolveVotes`, darwin's name-the-player FixUp errors.

## Workstream 9 — Stale-comment re-migration sweep (plan-ready, mechanical)

**Input:** `report-idiom-survey.md` Cluster A (exact file:line list).
**Goal:** migrate the five moves whose blockers no longer exist (pig `moveCountDie`, blackjack `moveCurrentPlayerStand` — or fold into Workstream 3, memory `moveHideCards`, checkers `movePlaceToken`, darwin `moveReplaceHot/ColdClimateCard`) and update/discharge their now-false survey comments.
**Mode:** plan-ready. Each migration behind a golden fence; each is an independent task. Low risk; good warm-up workstream for a new agent.

## Workstream 10 — Sanitization silent-visible default (SEPARATE DESIGN PASS; privacy)

**Input:** `report-idiom-survey.md` §3 item 1 (`base/game_delegate.go:363-365`).
**Goal:** an unmatched sanitization group must not silently resolve to `PolicyVisible`.
**Mode:** its own brainstorm → spec → critique cycle, independent of the legality campaign (it predates this branch and is a privacy bug class). Do not fold into another workstream. Candidate directions to explore in brainstorming (not decided): boot error on unresolvable groups; explicit `PolicyVisible` opt-in; boot warning when observer-view of example state exposes fields whose tags mention any policy.

## Deferred / explicitly out of scope (do not pick up without a new design round)

Carried from `legal/doc.go`'s limits section and the specs' deferred lists:
- `proposer.X` paths (requires LegalForAnyone per-player-existential redesign — not sugar-compatible).
- General `not`/`all` compositors (anti-tarpit rule; leaf-level negation via op/arg completeness has proven sufficient).
- Dynamic (per-game) component values paths.
- MayMoveTo facet narrowing (needs constraint immutability).
- Round-robin predicates → DealCountComponents/FinishTurn contributions (ledger desync risk documented in the A-round critique).
- Full constraints/predicate execution merger (message-layer unification only — see `report-idiom-survey.md` §4).
- go vet analyzer (rejected — dominated by Workstream 1; see `report-compile-time-safety.md` Proposal 4).
- F10 ledger `provisional`/`evaluable` client rendering semantics — belongs to sub-project B.

## Queued sub-projects (pre-existing, unchanged by this roadmap)

- **Sub-project B:** TS client evaluator + live move-graying. Contract: the conformance corpus. Must handle catalogVersion skew and the F10 semantics (see the audit report). Needs its own critique-hardened design round on user go-ahead. Workstream 6 shares its client renderer.
- **Sub-project C:** dirty-tracking write-set audit.

## Suggested execution order

W0 (finish batch) → W9 (mechanical warm-up) + W1 (codegen consts) in parallel → W2 + W3 → W7(a-c) → then user check-in on priorities among W4/W5/W6/W8/W10/B — they are independent and each needs either a spec round or explicit prioritization.
