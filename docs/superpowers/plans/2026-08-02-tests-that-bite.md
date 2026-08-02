# Tests That Bite — Implementation Plan

**Goal:** Find every check in this codebase that has silently stopped checking, fix them, and
leave behind mechanism so the class cannot quietly return.

**Why:** Across the 3D dice and components work, at least eight vacuous checks were found in
unrelated corners — and almost every one was found by *sabotage*, not by review. Reading a
test does not reveal that it cannot fail. The honest position today is that **we do not know
what fraction of this suite bites.**

## The evidence this is a class, not a run of bad luck

| What | How it failed to check |
|---|---|
| `die-geometry.test.ts` "is centred on its centroid" | Verbatim duplicate of a different test; asserted nothing about centring |
| `die-shape.spec.ts` facet assertions | Compared only multisets of clip-path vertex counts — forcing every facet square left all 8 green while a d7 rendered as shattered slabs |
| `dice-roll.test.ts` × 2 call sites | Passed obsolete signatures and still passed; `*.test.ts` is outside `tsconfig.json` |
| A die-roll fixture | Its first die never rolled, so a roll-identity test passed vacuously |
| `TestGolden` (stub) | Bare `return` before the assert whenever codegen fails — silently skips the comparison |
| `--pile-scale` | Plain field where Lit needs `@state`; a whole compute chain wrote a value nothing read, for the component's entire life |
| 8 Lit `@property` declarations | Observed the lowercased attribute name, so markup writing `faux-components` was never read |
| The tear-verification method | Paused-frame sampling reports 0 defects before *and* after — the prescribed check could not see the bug |
| `boardgame-stat` empty-value test | Deleted by its own author after finding the property was emergent from any inline box |

Second, narrower class: **`EnsureValid` silently retargeting** — four instances, three found
this session, one shipped in the games repo. A helper whose failure mode is "quietly answer
about a different player" rather than "reject."

## Global Constraints

- Work in `/Users/jkomoros/Code/go/src/github.com/jkomoros/boardgame/.claude/worktrees/three-d-dice`
  on `worktree-three-d-dice`. Go from the repo root with `GOWORK=off`; npm from `server/static/`.
- **Phase 1 is measurement only.** Do not fix anything you find; the list is the deliverable,
  and fixing while measuring corrupts the count.
- **A mutant that survives is a finding, not a failure.** Some survivors are legitimately
  equivalent mutants — record them as such with reasoning rather than forcing a test.
- Dependency policy is deliberately minimal (`allowScripts` allowlist, `audit:prod` in CI).
  Check before proposing any mutation-testing dependency; hand-rolled mutation is what found
  every bug listed above and is a legitimate answer.
- Parity goldens must never be regenerated. `git status --short -- server/static/tests/animations/parity/goldens/` EMPTY.
- Commits: imperative subject, trailer `Co-Authored-By: Claude Fable 5 <noreply@anthropic.com>`.
  Never stage `.database` or `boardgame-util/boardgame-util`.

---

## Phase 1 — Measure (no fixes)

### Task 1: Mutate the TypeScript unit suite

`npm run test:unit` is 751 tests in ~9 seconds, so exhaustive mutation is affordable here and
this is where to start.

Target `server/static/src/**` — especially `motion/`, `solid/`, `components/die-face-marks.ts`,
and the animation kernel. Generate mutants mechanically (flip comparisons, negate conditions,
swap operands, drop a term, return a constant, skip a loop body, delete a guard), run the
suite per mutant, and record every survivor.

**Deliverable:** a table of surviving mutants — file, line, mutation, and which test *should*
have caught it. Plus a headline number: mutation score per module.

### Task 2: Mutate the Go framework core

Slower but higher stakes: `moves/`, `legal/`, `behaviors/`, `base/`, and the state model.
`go test ./...` is the oracle.

Prioritize the code paths this session found bugs in — seat/activation reasoning, legality
spec assembly and ordering, sanitization, and stack constraint checks.

**Deliverable:** same table, plus specific attention to whether the *golden* tests (which
compare recorded output) actually discriminate, since a golden that records whatever shipped
is the easiest thing in the world to fool.

### Task 3: Sabotage the load-bearing browser assertions

The Playwright suite is ~10 minutes per run, so exhaustive mutation is infeasible. Instead
sabotage the checks the whole branch leans on, one at a time:

- the render-truth harness's self-validation (a d20 at zero wrong pixels; the positive control)
- the layer-count tripwire
- the parity trace and geometry goldens
- the `preserve-3d` context spec
- the die's legibility floor and upright-content assertions

**One finding is already known and must be quantified**: for the debuganimations scenarios the
geometry comparison is existential in both directions under a 0.08 tolerance, so *cardinality
churn is invisible to pass/fail*. Two consecutive re-recordings of one scenario gave 13 curves
and then 9. Determine what those goldens actually pin, and what they do not.

**Deliverable:** for each, whether it bites, and what specifically it would miss.

---

## Phase 2 — Fix what the measurement found

Ordered by what the survivors turn out to be. Two rules:

- **Prefer strengthening the assertion over adding a test.** A vacuous test plus a real test
  is worse than one real test — the vacuous one still reads as coverage.
- **Every fix is proven by the mutant that motivated it**: apply, confirm red, revert.

---

## Phase 3 — Make it standing

Mutation testing that happens once is a snapshot; the class returns the moment attention moves.

- A repeatable harness (a script, a make target) so the pass can be re-run and the score
  tracked. Cheap enough to run on the unit suites routinely.
- Close the `*.test.ts` type-checking hole, which has now hidden three signature errors.
- Consider a guard against the specific shapes found: a test whose body is a duplicate of
  another's, an assert that follows an unconditional early return, a spec with no assertion.

---

## Phase 4 — Convert prose invariants into enforcement

The other half of the class: invariants that are documented, correct, and ignored.

The pattern that worked, three times this session:
- `validateBehaviorPairings` at construction turned a documented coupling into a boot error
  and immediately found the one live instance repo-wide.
- The attribute-name lint closed a whole class permanently.
- The `FallbackName`-vs-type-name AST test found another bug the moment it existed.

Each replaced a comment that had been correct and ignored for years. The seating deadlock was
documented in **three places** and shipped in **four games** anyway.

**Task:** audit the framework for "you must also…" comments — required companion moves,
required options, required call ordering, required overrides — and for each, decide whether it
can become a boot error, a lint, or a test. Report the ones that genuinely cannot, and why.

**The rule to write down:** prose about a required invariant is not enforcement.
