# Declarative Legality — Execution Runbook (Tasks 5–14 + close-out)

**Audience:** an executing agent with NO memory of this session. Follow it
literally. Companion documents (read in this order when lost):
1. This runbook — HOW to execute.
2. `docs/superpowers/plans/2026-07-11-declarative-legality.md` — WHAT each task builds (per-task Files/Interfaces/Steps).
3. `docs/superpowers/specs/2026-07-10-declarative-legality-design.md` (rev 2.1) — WHY; normative on all design questions.
4. `.superpowers/sdd/progress-legality.md` — the ledger: which tasks are DONE (never re-execute those; if the ledger is missing, reconstruct from `git log --oneline` commit subjects, which match the plan's per-task commit messages).

## 0. Standing facts (verify, don't assume)

- Repo: `/Users/jkomoros/Code/go/src/github.com/jkomoros/boardgame`, branch `declarative-legality-design`. Sibling repo: `/Users/jkomoros/Code/go/src/github.com/jkomoros/games` (Task 13 creates the same-named branch there).
- Verification gate for every commit: `go build ./... && go vet ./...` plus `go test ./legal/... . -count=1` for legal/core tasks, widening to the packages a task touches; full `go test ./...` before a task's final commit.
- **Known pre-existing failures to EXCLUDE from all gates** (verified pre-existing on the base commit): `boardgame-util/lib/build/api`, `boardgame-util/lib/gamepkg` (sandbox GOPATH path issues), `storage/mysql` (no local MySQL). If any OTHER package fails, that's real.
- Frozen-chain rule (inviolable, spec "prime guarantee"): no observable behavior change — checks, order, error STRINGS — for any move that doesn't declare `WithPreconditions`.
- Subagent-driven process scripts (superpowers plugin, version-dated path — glob for it if 6.0.3 is gone):
  - `~/.claude/plugins/cache/claude-plugins-official/superpowers/6.0.3/skills/subagent-driven-development/scripts/task-brief PLAN_FILE N` → writes `.superpowers/sdd/task-N-brief.md`
  - `.../scripts/review-package BASE HEAD` → writes `.superpowers/sdd/review-BASE..HEAD.diff`
- Commit trailer for controller-authored commits: `Co-Authored-By: Claude Fable 5 <noreply@anthropic.com>`. Implementer commits use the plan's per-task commit message verbatim.
- Network flakiness: agent dispatches have died on ConnectionRefused / certificate errors. Protocol: check `git log`/`git status` for partial work (there has never been any — agents die in discovery), re-dispatch the SAME prompt once; after a third consecutive network failure, stop and tell the user connectivity is down.

## 1. The per-task loop (repeat for every remaining task)

1. `BASE=$(git rev-parse --short HEAD)` — record before dispatching.
2. Extract the brief: `task-brief docs/superpowers/plans/2026-07-11-declarative-legality.md N`.
3. Dispatch ONE implementer subagent (model per §3 table). The dispatch prompt contains, in order: (a) one line locating the task in the campaign; (b) "Read your brief first: <path>" + "Normative spec: <path> §refs"; (c) the task's **Forward Context block from §4 below, verbatim**; (d) the TDD job list from the plan task (RED → implement → GREEN → gates → commit message); (e) the report contract: full report to `.superpowers/sdd/task-N-report.md`; reply ONLY status/commits/one-line-tests/concerns/report-path, under 15 lines. Statuses: DONE / DONE_WITH_CONCERNS (read concerns; correctness concerns must be addressed before review) / NEEDS_CONTEXT (answer, re-dispatch) / BLOCKED (more context → retry; else stronger model; else split task; else ask user).
4. On DONE: `review-package $BASE <head>`; dispatch ONE reviewer subagent (sonnet unless §3 says otherwise) with: brief path, report path, diff path, the task's binding constraints (copy from the plan task + Forward Context), and the task's **Named Review Risks from §4**. Reviewer output format: Spec Compliance / Strengths / Issues (Critical/Important/Minor) / Assessment (Approved | Needs fixes).
5. Critical/Important findings → ONE fix subagent with the complete list, each fix naming its covering tests; then a focused re-review of the fix commits (✅/❌ format). Repeat until clean. Never proceed with open Critical/Important. Minors: record in ledger, forward to the final whole-branch review.
6. Append a ledger line: `Task N: complete (commits BASE..HEAD, review <verdict>, <key decisions/minors>)`. Decisions that later tasks consume MUST go in this line.
7. Never run two implementation agents in parallel. Reviewers are read-only and may overlap a controller bookkeeping step, nothing else.

## 2. Ledgered decisions that bind everything downstream

- **`ctx.ResolvePath(p LegalPropPath) (interface{}, PropertyType, error)`** (method on `boardgame.LegalContext`) is THE path-resolution API. The `legal` package contains ZERO grammar/parsing — its `legal/path_resolve.go` holds only typed wrappers. Any new predicate resolves through it. Reviewers should grep-check no parsing creeps back in.
- **Facet honesty:** every predicate's `Reads` declares the MINIMAL facet its Evaluate touches. `MayMoveTo`/`MayMoveToSlot` declare `LegalFacetValues` on dstPath (pessimistic — stack constraints can read values; no constraints accessor exists on stacks; narrowing is documented future work in `legal/catalog_stack.go`).
- **Templates:** every Fail carries a template KEY + bindings. Keys are constants in catalog files, collected in package-level `defaultTemplateKeys` (in `legal/catalog_stack.go`), which **Task 6 consumes** to build `DefaultTemplates()`.
- **Conformance corpus** (`legal/testdata/conformance/*.json`, runner `legal/conformance_test.go`): one file per predicate, ≥3 cases, optional `"template"` field REQUIRED on fail cases. Fixtures build on `examples/memory` + `examples/checkers` (both carry back-reference comments; Tasks 11–12 must not rename state properties those fixtures read without updating the corpus).
- **Naming:** `legal` exports bare `Pass`/`Fail`/`Unknown` outcome consts — no catalog builder may take those names.
- **Zero-value discipline:** `LegalVerdict{}`/`LegalOutcome(0)` is invalid and fails closed; `evalLegalPredicate` converts invalid verdicts and panics to Unknown (never Pass). `any` is depth-1, enforced through the resolve closure AND a post-resolution tree walk.
- **Proposer atom is FIELD-DEPENDENT** (reads `move.TargetPlayerIndex`) — plan §4; it belongs in the field-dependent bucket.

## 3. Remaining tasks: model + one-line shape

| Task | Implementer model | Shape |
|---|---|---|
| 5 (in flight at time of writing) | sonnet | catalog part 2: Any/AllActivePlayers/ProposerIsCurrentPlayer/RevealableCardAt/ComponentPropEqualsCurrentPlayer |
| 6 | sonnet | templates: TemplateConfigurer, DefaultTemplates, RenderMessage, Errorf, LegalError, boot key-validation contract |
| 7 | sonnet | moves wiring: WithPreconditions/WithoutPrecondition, ContributedPreconditions (Default+CurrentPlayer ONLY), framework wrapper predicates (inPhase/inProgression/stackConstraints) via extract-and-share, string-freeze test FIRST, + RepeatFromProp (#644) |
| 8 | **opus** | plan assembly, probe, opt-in Legal path, CustomLegaler, purely-sugar property tests — the campaign's riskiest task |
| 9 | sonnet | phase index + agnostic bucket (superset test), field-indep memo, tape memo, #65 logging |
| 10 | sonnet | server ledger + chest templates + TS types (touches server/api/main.go:1590-1657 + server/static/src) |
| 11 | sonnet | golden harness + migrate memory & blackjack |
| 12 | sonnet | migrate checkers (game-registered predicate!) + survey remaining examples |
| 13 | sonnet | ../games: create branch `declarative-legality-design` there, survey/migrate all five games, `replace` directive in its go.mod if it pins a release |
| 14 | sonnet impl + **the tutorial part reviewed hard** | corpus completeness meta-test, docs, **USER-MANDATED TUTORIAL INTEGRATION (blocking — see plan Task 14's 6-point list; every snippet compile-checked against real branch code)**, full verification, close-out |

Reviewers: sonnet for 5–7, 9–14; **opus for Task 8's review** and the final whole-branch review (opus or better).

## 4. Per-task Forward Context + Named Review Risks

### Task 5 — catalog part 2
Forward Context: §2 bullets 1–6 verbatim; plus: ProposerIsCurrentPlayer must replicate `moves/current_player.go:37-65` EXACTLY (read it; both error strings preserved — pin via template constants whose Task-6 default text is the verbatim legacy strings); AllActivePlayers v1 restriction: inner limited to `playerBool` + player-path `propCompare`/`propAtLeast`, boot error otherwise; inactive-player check per `behaviors.PlayerIsInactive` (type-assert if import cycle; document); quantifier resolves `player.X` against EACH iterated player's reader directly (not ctx.ResolvePath current-player semantics), Reads declared `players[*].X`; RevealableCardAt is occupancy-only two-branch (templates `legal.no_card_here`, `legal.already_revealed`); ComponentPropEqualsCurrentPlayer needs checkers' color↔player-index mapping from `examples/checkers/moves.go:93-138`.
Named Review Risks: proposer string parity + field-dependence declared; per-player resolution correctness (no current-player leakage); v1 inner-restriction enforced at construction; RevealableCardAt reads nothing beyond occupancy; checkers color mapping fidelity; corpus templates pinned.

### Task 6 — templates
Forward Context: consume `defaultTemplateKeys` (legal/catalog_stack.go) — `DefaultTemplates()` must cover every key or the coverage test fails; the proposer templates' default text must be the VERBATIM strings from moves/current_player.go (see Task 5 report for the approach chosen); `LegalError` must never produce a typed-nil error interface (Pass → literal nil); `RenderMessage` missing-binding renders the placeholder name, never panics; boot validation contract signature per plan Task 6 (consumed by Task 8).
Named Review Risks: DefaultTemplates coverage proven by evaluating corpus fail cases (not a hand-list); Errorf round-trips errors.As; nil-interface test present; template rendering deterministic.

### Task 7 — moves wiring
Forward Context: **write the string-freeze test FIRST against today's chain and keep it green through every refactor** (capture exact Legal() error strings for a table of memory/blackjack fixture states — reuse moves package test fixtures); extraction refactors move shared helpers to core, and `moves/default.go`'s frozen chain must call the SAME helpers in the SAME order with byte-identical strings; contributions come ONLY from Default (+CurrentPlayer appending proposer atom); config-bag keys namespaced like moves/with.go:8-31; existing WithLegalPhases/WithLegalMoveProgression/WithSourceProperty stay working AND become readable as specs; RepeatFromProp: `MoveProgressionGroup.Satisfied` gains access to state (plumbing named in spec §7) with its own test.
Named Review Risks: string-freeze test integrity (written before refactor? asserts strings not just nil-ness?); frozen chain byte-identical; contribution order deterministic base-first; no other moves-package type gained contributions; import cycles (moves→legal→core is the only allowed direction — check what the wrapper predicates in legal/catalog_framework.go import).

### Task 8 — plan/probe/opt-in (OPUS, highest risk)
Forward Context: spec "prime guarantee" section is normative — re-read it; probe mechanism: manager-scoped `probing/probeReached` flags checked-and-set first thing in Default.Legal (no public signature change, boot-time only); opt-in = declared specs present AND probe reaches Default.Legal; declared-but-unreachable → boot error naming move; declared-on-unsupported-base → boot error naming base type; plan buckets: field-independent vs field-dependent split by `move.*` reads (proposer atom is field-DEPENDENT), custom (CustomLegaler wrapper — the package-internal `opaque` predicate field from Task 3 gets its exported construction path here) always last; evaluation order = plan order, NO Cost sort; purely-sugar property tests per plan Task 8 list (a/b/c) — the full existing `go test ./...` green IS test (a)'s floor.
Named Review Risks: probe false-positives (super-calling override must PASS and its plan must actually evaluate); probe thread-safety/boot-only-ness; double-evaluation of contributed atoms when an override super-calls (must not — the plan path replaces the frozen chain inside Default.Legal, so super-call = plan evaluation exactly once); zero-verdict fail-closed through the whole plan path; boot errors name the move; LegalVerdictEntry populated for ledger consumers (Task 10 shape).

### Task 9 — engine wins
Forward Context: candidateMoves = phaseIndex[current ∪ TreeEnum ancestors] ∪ phaseAgnostic; phaseAgnostic contains EVERY opaque move and every opted-in move without an inPhase atom — the SUPERSET PROPERTY test is the deliverable that matters; integration point is `base/game_delegate.go:86` (ProposeFixUpMove iteration), NOT the core loop; memo keyed (moveTypeName, stateVersion, proposer), evicted on version advance; tape memo shared by frozen chain AND inProgression predicate (retires moves/default.go:475 TODO); #65 logging via manager.Logger().Debugln with structured fields.
Named Review Risks: superset property test genuinely mixed (opaque + opted-in + no-phase moves); memo never crosses versions/proposers; no behavior change for opaque moves (they're always candidates); eviction bounded (no unbounded growth per game).

### Task 10 — server ledger
Forward Context: wire shape per plan Task 10 verbatim (preconditionEntry, LegalCatalogVersion, chest LegalTemplates); evaluable = Serializable ∧ every Read's facet survives the viewer's sanitization policy (use core facetSurvives + the per-property policy from the sanitization transformation — grep sanitization.go); **#693 guard: evaluable=false ⇒ strip Message.Bindings**; provisional = FieldDependent; ledger evaluation replaces the SECOND admin Legal() call only for opted-in moves; opaque moves' legacy JSON byte-identical (assert against a recorded pre-change fixture); TS types + `cd server/static && npm run type-check` green; dev-server note: `GOPATH=$HOME/Code/go ./boardgame-util/boardgame-util serve` if a live check is needed (see OFFLINE_DEV_MODE.md).
Named Review Risks: bindings-stripping actually applied at the JSON boundary (not just in-memory); byte-identical legacy fixture test real; per-viewer evaluability (player vs admin vs observer differ); no second Legal() call remains for opted-in moves; catalogVersion bump story documented.

### Task 11 — golden harness + memory & blackjack
Forward Context: migrations per spec §8 EXACTLY (memory moveRevealCard: PropAtLeast + RevealableCardAt + MayMoveToSlot, Legal() deleted; blackjack moveStartRoundCleanup: AllActivePlayers(Any(PlayerBool,PlayerBool))); hard-custom stays in LegalCustom with legal.Errorf; golden harness compares kept `legacyLegal` private copies vs plan across recorded states (reuse each game's golden JSON loading — grep `golden` usage); delegates gain ConfigureLegalTemplates with the verbatim legacy strings; **conformance fixtures read these games' state shapes — do not rename properties (see back-reference comments in the games)**; templates: `reveal.*` keys per spec §8.
Named Review Risks: golden equivalence covers ILLEGAL cases + message strings (not just legal/illegal boolean); every deleted Legal() has a golden; template table completeness (boot would fail otherwise — confirm a boot in tests); memory's two card-compare moves correctly LegalCustom not force-declarativized.

### Task 12 — checkers + survey
Forward Context: checkers per spec §8 (game-registered `checkers.spaceIsBlack` via ConfigurePredicateConstructors/ExtendDefaults — FIRST use of the game-registration path, the review should scrutinize the delegate wiring; capture-graph walk stays LegalCustom returning `legal.Errorf("checkers.illegal_dest", nil)`); then survey tictactoe/pig/werewolf/debuganimations: migrate only catalog-covered moves, goldens each, one commit per game, un-migratable moves documented in commit messages (embedding unsupported base types stays untouched — v1 seam is Default+CurrentPlayer only).
Named Review Risks: delegate registration consumed via type-assertion (no GameDelegate interface change); survey honesty (commit messages name what was NOT migrated and why); goldens per migrated move.

### Task 13 — ../games
Forward Context: `git -C /Users/jkomoros/Code/go/src/github.com/jkomoros/games checkout -b declarative-legality-design`; check its go.mod — if it pins a released boardgame version, add `replace github.com/jkomoros/boardgame => ../boardgame` and note in commit; survey ALL FIVE games (murdermrmonroe, pass, valentine, darwin, metaltrader — the last two are Go-only, they still have moves); same migration rules as Task 12; `go test ./...` in that repo green; commit per game there.
Named Review Risks: replace-directive handling (not accidentally committed if already present); migrations use only the public API (legal.* + moves.*) — external repo is the first true out-of-tree consumer, any reach into boardgame internals is a design failure worth flagging loudly.

### Task 14 — close-out (tutorial is BLOCKING)
Forward Context: corpus completeness meta-test (every DefaultConstructors name has a file with ≥3 cases); legal/doc.go full authoring guide; moves/doc.go Legal section; **the user-mandated tutorial work per plan Task 14's six-point list — treat it as the task's main deliverable, not an afterthought**: TUTORIAL.md teaches declarative as primary with the REAL migrated moveRevealCard before/after, templates, LegalCustom, WithoutPrecondition, client ledger payoff, purely-sugar framing; every snippet compile-checked (scratch test file or verbatim-from-source verification); full gates: go build/vet/test ./... (both repos), `cd server/static && npm run type-check`, one Playwright smoke (`npx playwright test tests/animations/waapi-gate.spec.ts -g blackjack` with the dev server up) proving move-forms still work in the real client.
Named Review Risks: tutorial snippets match branch code EXACTLY (reviewer diffs them against source); tutorial teaches the imperative path as still-first-class (purely-sugar framing, not deprecation); meta-test can't be satisfied by empty corpus files.

## 5. End-game (after Task 14)

1. `review-package $(git merge-base master HEAD) HEAD` → dispatch final whole-branch reviewer (opus or the most capable available; template: superpowers requesting-code-review/code-reviewer.md) with: spec + plan + runbook paths, the ledger's accumulated Minors list for triage, and the campaign framing (purely-sugar guarantee, the three user goals: developer ergonomics / cheap client proactivity / correctness-no-footguns).
2. Findings → ONE fix subagent with the complete list → re-verify.
3. Run superpowers:finishing-a-development-branch. **Do NOT merge without explicit user confirmation** — present the standard options; the user's past preference was land-on-master `--no-ff` (see animation campaign, commit 1e86dcf2), but this campaign is larger: ASK.
4. ../games branch: merge/PR decision follows the boardgame repo's (ask in the same message).
5. Final summary must include: per-issue outcomes (#761 #189 #790 #644 #65 #640, #213-enabling), migration stats (N declarative / M LegalCustom / K untouched per repo), deferred items (TS evaluator, dirty-tracking, MayMoveTo facet narrowing, seam expansion beyond Default/CurrentPlayer), and a pointer to this runbook + ledger for archaeology.
