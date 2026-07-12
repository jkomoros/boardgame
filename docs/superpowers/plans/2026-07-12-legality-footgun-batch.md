# Legality Footgun Batch Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Close the five verified footguns (plus doc drift) found by the adversarial audit of the declarative-legality system, on branch `declarative-legality-design`, before the branch lands.

**Architecture:** All fixes extend existing seams: the boot gauntlet in `assembleLegalPlans` (legal_plan.go), the server ledger assembly (`server/api/main.go`), the verdict/message layer (`legal_predicate.go`, `legal_error.go`), and package docs. No new packages. No wire-format breaks (additive only).

**Tech Stack:** Go (package boardgame + legal + moves), existing test harnesses (golden-equivalence fences, conformance corpus, `TestLegalChainStringFreeze`).

## Global Constraints

- **Purely-sugar guarantee is inviolable:** the frozen imperative chain must stay byte-for-byte (`TestLegalChainStringFreeze` must pass unmodified). Fixes may only ADD boot errors and server-side checks; never change behavior of non-opted-in moves.
- **Commit discipline:** stage NAMED FILES ONLY. NEVER `git add -A` (session working tree contains unrelated noise: .database, temp dirs, symlinks).
- **Gates per task:** `go test ./...` in boardgame repo excluding known pre-existing failures (boardgame-util/lib/build/api, boardgame-util/lib/gamepkg — GOPATH sandbox; storage/mysql — no local DB). When a task can affect game-visible legality output, also run `go test ./...` in /Users/jkomoros/Code/go/src/github.com/jkomoros/games (excluding valentine TestGolden — pre-existing, testdata deleted upstream).
- **Ledger:** append a row per completed task to `.superpowers/sdd/progress-legality.md`.
- New boot errors must name the offending move type (match existing gauntlet error style in legal_plan.go).
- Audit findings referenced below (F1–F9) are from the adversarial audit report; each finding's file:line citations were verified against this branch.

---

### Task 1: Suppression validation + exported precondition-name constants (F2)

**Files:**
- Modify: `legal_plan.go` (assembleLegalSpecList / assembleLegalPlans area)
- Modify: `moves/with.go` (or `moves/preconditions.go`) — exported constants
- Modify: `legal/doc.go` — WithoutPrecondition section documents the opt-in dependency and the new boot errors
- Test: `legal_plan_test.go` (or wherever existing boot-gauntlet tests live), `moves/preconditions_test.go`

**Interfaces:**
- Produces: `moves.PreconditionInPhase = "inPhase"`, `moves.PreconditionInProgression = "inProgression"`, `moves.PreconditionStackConstraints = "stackConstraints"`, `moves.PreconditionProposerIsCurrentPlayer = "proposerIsCurrentPlayer"` (untyped string consts; `WithoutPrecondition(string)` signature unchanged).

**Behavior to implement (all three flavors from F2, verified latent at legal_plan.go:507-522):**
1. A `WithoutPrecondition` name that matches NO contributed spec name for that move → boot error naming the move and the unmatched name, listing the move's actual contributed names. This subsumes both the typo flavor (`"inphase"`) and the suppressing-a-check-the-move-doesn't-contribute flavor (`"inProgression"` on a move with no progression).
2. `WithoutPrecondition` present but NO authored `WithPreconditions` specs (move not opted in; suppression is dead) → boot error: "WithoutPrecondition requires opting in via WithPreconditions".
3. Existing CurrentPlayer/proposer guard (legal_plan.go:422-440) stays; do not weaken it.

- [ ] **Step 1:** Write failing boot-gauntlet tests for all three flavors (typo'd name; valid name not contributed by this move; suppression without opt-in). Use existing test-game scaffolding in the boardgame package tests.
- [ ] **Step 2:** Run tests, confirm they fail (games boot cleanly today).
- [ ] **Step 3:** Implement validation at plan assembly. Collect the move's contributed spec names before filtering; after filtering, any suppression that matched nothing is an error. The suppression-without-opt-in check runs even on the early-return path (legal_plan.go:408-413) — that path currently skips before suppressions are examined.
- [ ] **Step 4:** Add the four exported constants with doc comments; migrate in-repo call sites (`examples/memory/tutorial_snippets_test.go:91` and any others `grep -rn "WithoutPrecondition("` finds) to use them. Update TUTORIAL.md's WithoutPrecondition example to use the constant.
- [ ] **Step 5:** Update `legal/doc.go` WithoutPrecondition section: the opt-in dependency (currently only an internal comment at legal_plan.go:53-56) and the new boot errors.
- [ ] **Step 6:** Full gate (`go test ./...` both repos), then commit named files.

**Review risks:** ensure no existing game/test uses a no-op suppression that would now fail boot (grep both repos first — if one exists, that's a latent bug to fix in that game, not a reason to weaken validation). Ensure error text style matches gauntlet conventions.

---

### Task 2: Ledger booleans become ground truth (F1)

**Files:**
- Modify: `server/api/main.go` (~1662-1790, generateFormsWithLegality / legalFormFromLedger)
- Test: `server/api/` existing legality-form tests

**Problem (verified):** for opted-in moves, `legalFormFromLedger` computes `LegalForPlayer`/`LegalForPlayerError`/`LegalForAnyone` from the plan alone; a super-calling `Legal()` override with imperative residue (explicitly blessed by spec rule 4) is invisible → button enabled, ProposeMove rejected. Inverse direction too: an override early-`return nil` path makes a plan-Fail move actually legal → button wrongly disabled.

**Fix (preferred approach):** derive the booleans from the real ground truth — `move.Legal(state, playerIndex)` for `LegalForPlayer`/`LegalForPlayerError`, and `move.Legal(state, AdminPlayerIndex)` for `LegalForAnyone` — keeping the per-predicate ledger as advisory explanation detail. For opted-in moves without an override, `move.Legal` IS the plan evaluation and hits the memo (legal_memo.go), so the pass-case cost is a memo lookup, not a re-evaluation. This restores the exact pre-branch boolean semantics for ALL moves while keeping the ledger.

**Fallback approach (only if the preferred approach measurably regresses the move-forms path or breaks golden tests in a way that can't be reconciled):** keep plan-derived booleans but, when the plan verdict is Pass, confirm with `move.Legal()` and demote on error. Document in the commit message why the fallback was chosen.

- [ ] **Step 1:** Write a failing regression test: a test move that opts in via WithPreconditions AND keeps a super-calling `Legal()` override whose residue rejects (e.g. checks a game-state bool). Assert the served move form reports LegalForPlayer=false with the residue's error, while today it reports true.
- [ ] **Step 2:** Write the inverse-direction test: override with a conditional early `return nil` while the plan would Fail → LegalForPlayer must be true.
- [ ] **Step 3:** Confirm both fail today.
- [ ] **Step 4:** Implement. Preserve the existing #693 evaluable/bindings stripping and the ledger contents unchanged — only the boolean/error derivation changes.
- [ ] **Step 5:** Verify the existing LegalForAnyone regression test (ObserverPlayerIndex current player) still passes, plus full gates in both repos (golden fences).
- [ ] **Step 6:** Commit named files.

**Review risks:** double-check proposer semantics (`LegalForPlayer` uses the viewing player as proposer — match ProposeMove's semantics exactly); confirm memoization actually absorbs the extra calls (the memo is keyed moveType/version/proposer — per-player calls each hit their own key, same as the ledger eval already does); ensure fixup-move handling (admin-only moves) is unchanged.

---

### Task 3: Message-layer fixes — `any` Unknown message (F6) + template placeholder/binding boot validation (F4)

**Files:**
- Modify: `legal_predicate.go` (evalLegalAnyKleene, ~309-327)
- Modify: `legal_error.go` (validation, ~145-264; placeholder pattern already exists as legalPlaceholderPattern)
- Modify: `legal_types.go` if a metadata field is needed (e.g. EmittedBindings)
- Modify: `legal/` catalog files — bindings metadata per template (names are already documented prose on each Template* constant, e.g. catalog_compare.go:15-36)
- Test: existing predicate/template test files + conformance corpus additions if fail-message shape changes

**F6 behavior:** `any(Fail, Unknown)` currently returns Unknown with a bare internal Reason and IGNORES the spec's `WithMessage`/`legal.any_failed` template (used only when all children Fail). Fix: attach the spec's Message (author override or default template) to the Unknown verdict too (LegalVerdict explicitly permits Message on Unknown — legal_types.go:130-132), and include the identity of the unknown sub-predicate in the Reason. The frozen chain is untouched (any-composition only exists declaratively).

**F4 behavior:** boot-error when a template body's `{placeholders}` are not a subset of the bindings the predicate emits for that template. Mechanism: promote the documented binding names to metadata (parallel to EmittedTemplates — e.g. `EmittedBindings map[string][]string` on the predicate or registry entry); at boot, for every spec (including WithMessage overrides), extract placeholders from the resolved template body via legalPlaceholderPattern and check subset. For game-registered predicates the metadata is optional: absent metadata skips validation (can't fail-closed without breaking existing registrations) but is documented as recommended. Also validate DefaultTemplates against their canonical catalog predicates' bindings as a package test.

- [ ] **Step 1 (F6):** failing test: any(Fail, Unknown) verdict carries the spec's message template; Reason names the unknown child.
- [ ] **Step 2 (F6):** implement; run conformance corpus (fail-message pins may need updating ONLY for the any-unknown case — justify each corpus edit in the commit message).
- [ ] **Step 3 (F4):** failing boot test: catalog predicate + WithMessage pointing at a game template whose body references a placeholder the predicate never emits → boot error naming move, template key, and the missing binding.
- [ ] **Step 4 (F4):** implement metadata + boot check; add the package test pinning DefaultTemplates' placeholders ⊆ their predicates' emitted bindings.
- [ ] **Step 5:** Full gates both repos; commit named files.

**Review risks:** metadata for ~20 catalog predicates must match what Evaluate actually emits (read each Evaluate body, don't trust the prose); ensure the boot check handles the multi-template case (a predicate emitting different templates on different paths); no wire-format change (metadata is server-side only, json:"-" or unexported).

---

### Task 4: Reads honor-system hardening + doc corrections + players[move.Field] int restriction (F3, F7, F8, F9)

**Files:**
- Modify: `legal/doc.go` (F3 memo-poisoning docs; F8 overclaim correction)
- Modify: `moves/doc.go` (F7 bucket-ordering caveat in the WithPreconditions section)
- Modify: `legal_plan.go` (F7: correct the "in the order you wrote it" comment at ~83-84; F3: optional boot smoke probe)
- Modify: `legal_path.go` (F9: ~168-174)
- Modify: `docs/superpowers/specs/2026-07-10-declarative-legality-design.md` (F7: §4's ordering claim gets the bucket caveat)
- Test: boot-gauntlet tests for the smoke probe and F9

**F3 (docs, mandatory):** document — in legal/doc.go where game-registered predicates are taught — that an under-declared `move.*` read doesn't just skew client evaluability: it lands the predicate in the field-independent bucket whose verdict is memoized WITHOUT move fields in the key (legal_plan.go:543-557, legal_memo.go:52-90), so the server itself returns stale verdicts. State purity (no time/randomness) as a hard requirement in the same place.

**F3 (smoke probe, attempt it; drop with justification if it proves fragile):** at boot, for each field-independent (per declared Reads) game-registered predicate, evaluate once against the example state with a sentinel Move whose PropertyReader panics on every property access (the sentinel is non-nil so defensive `ctx.Move == nil` checks pass). A recovered panic → boot error: "predicate X reads move properties but declares no move.* Reads". Call Evaluate directly (not through evalLegalPredicate's panic-to-Unknown wrapper). Catalog predicates can be skipped (already audited); scope to game-registered ones to bound risk.

**F7 (docs):** moves/doc.go and the spec currently say declarations run "in the order you wrote it"; the field-independent → field-dependent → custom bucket split (legal_plan.go:192-216) means a field-independent check declared after a field-dependent one still reports first. One-sentence caveat in each place; correct the legal_plan.go:83-84 comment.

**F8 (docs):** legal/doc.go:181-184 overclaims that `Errorf`/`FailT` keys inside LegalCustom bodies are boot-validated — they cannot be (closures). Correct the sentence to scope the guarantee to declared Specs and EmittedTemplates.

**F9:** `players[move.Field]` boot validation accepts TypeInt fields (legal_path.go:168-174); an int field silently indexes a wrong-but-valid player. First grep BOTH repos for `players[move.` usages and check each field's type. If no in-repo usage depends on TypeInt: restrict to TypePlayerIndex (boot error for int fields, message suggesting PlayerIndex type). If a legitimate int usage exists: keep the allowance, add an explicit doc warning instead, and note the finding in legal/doc.go's limits section.

- [ ] **Step 1:** F9 grep + decision; failing boot test for whichever behavior chosen; implement.
- [ ] **Step 2:** F3 smoke probe: failing test (game-registered predicate that reads a move field but declares no move.* Read → boot error); implement sentinel reader + probe.
- [ ] **Step 3:** All doc corrections (F3 docs, F7 ×3 sites, F8).
- [ ] **Step 4:** Full gates both repos; commit named files (docs and code may be separate commits).

**Review risks:** the sentinel PropertyReader must implement the full PropertyReader interface (~30 methods — mechanical; consider embedding an existing reader and overriding, or generate); ensure the probe cannot fire for predicates that legitimately never touch the move; F9 restriction must not break darwin/tictactoe's existing `players[move.X]` usages (they use PlayerIndex-typed fields — verify, don't assume).

---

### Task 5 (final): Whole-batch verification

- [ ] Run full gates in both repos (standard exclusions).
- [ ] Run `TestLegalChainStringFreeze` explicitly and confirm untouched.
- [ ] Re-read the audit's F1–F9 list and confirm each is either fixed, documented, or explicitly deferred with a note in legal/doc.go's limits section.
- [ ] Update `.superpowers/sdd/progress-legality.md` with the batch summary.
- [ ] Do NOT merge — branch-finishing still awaits explicit user confirmation per the runbook.
