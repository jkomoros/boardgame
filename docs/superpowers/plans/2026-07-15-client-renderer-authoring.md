# Client Renderer Authoring Implementation Plan

> **For agentic workers:** Execute this plan one task at a time. Use failing
> tests or compile fixtures before implementation, commit named files only,
> and run an adversarial review after each architectural seam. Tasks 1–9 are
> the approved first vertical tranche. The later roadmap must be re-planned
> after the Task 10 review gate rather than implemented speculatively.

**Goal:** Make high-quality game renderers straightforward and difficult to
misconfigure: strict generated TypeScript, tiny common cases, composable
headless building blocks, accessible and responsive defaults, loud actionable
errors, and explicit escape hatches for unusual games.

**Architecture:** Build upward in layers: honest generated wire and author-input
contracts; an immutable renderer snapshot plus stable services; one typed move
action currency; small headless interaction controllers; accessible animated
primitives; optional polished compositions; and ordinary Lit/SVG/canvas as the
advanced escape hatch. Do not build a giant declarative renderer schema or a
client-side duplicate of server legality.

**Branch:** `client-renderer-authoring-audit`

**Repositories:**

- Framework: `/Users/jkomoros/Code/go/src/github.com/jkomoros/boardgame`
- External game corpus: `/Users/jkomoros/Code/go/src/github.com/jkomoros/games`

**Two-repository protocol:** Treat framework and games as independent
repositories. Before any external migration, require both worktrees clean except
for explicitly recorded unrelated files, fetch their bases, and create a
dedicated matching branch in `games`. Never edit or commit external-game files
on `master`, never stage across repository boundaries, and report status/tests
and commit SHAs separately. Framework primitives and framework-owned fixtures
must land without depending on uncommitted sibling-repository state. An external
migration records the minimum framework commit it requires.

The unrelated untracked `SECURITY-INCIDENT-injection-canary.md` must remain
untouched. Never use `git add .` or `git add -A`.

## `boardgame-util` modernization scope

The creator-facing CLI is part of the renderer API. This tranche modernizes it
where it directly shapes the authoring loop:

- `serve` owns a supervised, reproducible Vite/API session, supports arbitrary
  allocated ports, exits cleanly, and does not leave child processes behind;
- `stub` emits strict Lit/TypeScript using the public facade and generated game
  contracts, writes atomically, and refuses destructive overwrites by default;
- `check-client` provides one documented local/CI gate with structured
  diagnostics for generation, type, lint, assembly, and focused fixture checks;
- generation reports stable outputs and actionable failures rather than relying
  on incidental compiler or shell behavior.

Database administration, deployment, and other unrelated commands are not part
of this tranche unless they block the renderer workflow. Modernization must
preserve scriptability: stable exit codes, non-interactive defaults, and useful
plain terminal output remain requirements.

## Why this work comes before more components

The framework already has strong animation, server-authoritative legality, and
useful low-level pieces. The creator-facing front door is internally
contradictory:

- `boardgame-util stub` still emits Polymer JavaScript, Polymer imports,
  `dom-repeat`, `@apply`, moustache bindings, and obsolete paths
  (`boardgame-util/lib/stub/templates.go`).
- `TUTORIAL.md` and `server/README.md` still teach substantial portions of the
  old client model. Migration documents recommend an import path that
  `boardgame-util/lib/build/static/lint.go` says will return a development-server
  500.
- `Component<T>` claims values are flattened while real components expose them
  below `.Values`. Real renderers cast around the mismatch.
- Generated move arguments describe raw Go move fields, not the fields a game
  creator supplies after configured defaults and framework context injection.
  Pig's Roll Dice and Tic-tac-toe's Place Token would incorrectly require a
  `TargetPlayerIndex`; Checkers naturally supplies numeric indexes even though
  the wire encoder uses strings.
- All ten surveyed solo renderers omit the fourth renderer generic, silently
  widening proposal arguments to `Record<string, unknown>`. The sibling games
  repository does not currently generate `_move_args.ts`.
- Game actions are authored through three weakly connected paths: the typed-ish
  `proposeMove()` method, legacy `propose-move`/`data-arg-*` attributes, and
  untyped `componentAttrs` forwarding. Legality, pending state, errors, and
  accessibility are separately wired or omitted.
- Game renderer type errors are warnings in production assembly and are not a
  normal development gate.
- Grid, spatial, component, and die interactions are click-centric inside
  framework shadow roots, so a game creator cannot reliably repair keyboard and
  assistive-technology behavior in game code.
- Player info, deck templates, player presentation, turn prompts, action areas,
  responsive sizing, and outcome UI contain repeated bespoke code across the
  real corpus.

The first implementation tranche therefore restores contract trust and proves
one vertical path through a simple action game, a regular grid, and an authored
graphic board before attempting a broad component catalog.

## Non-negotiable design principles

1. **Safe APIs do not silently erase types.** Omitted generics must not degrade
   to `any` or `Record<string, unknown>`. Dynamic data is `unknown` and narrowed.
2. **Generated inputs model creator responsibility.** Move inputs are generated
   after defaults and framework-provided context are accounted for, in native
   ergonomic types. Wire serialization remains internal.
3. **One action currency.** Buttons, cards, stacks, boards, spatial targets, and
   future workflows consume the same typed, legality-aware action object.
4. **Server authority remains authoritative.** The client may render server
   legality, preview bound arguments, and validate the transport contract; it
   does not reimplement game rules.
5. **Headless behavior before opinionated visuals.** A game can use framework
   behavior with completely custom Lit markup.
6. **Accessibility lives in primitives.** Keyboard operation, disabled and
   selected semantics, focus preservation, accessible naming, and live status
   cannot be left to each game.
7. **Reasonable output is the zero-configuration output.** Standard components
   should be attractive, responsive, reduced-motion-aware, and usable on touch.
8. **Errors are loud and actionable.** Development errors identify the game,
   surface, file/line where possible, observed value, required contract, and a
   concrete fix. Deterministically invalid structures fail closed.
9. **Advanced escape hatches are explicit.** Raw proposal, arbitrary attributes,
   DOM animation, SVG, canvas, and dynamic data remain possible through clearly
   named low-level or `unsafe*` APIs.
10. **Do not freeze accidents.** Inventory existing deep imports and behavior,
    but compatibility governance covers only a deliberately curated public
    facade after it has survived the assembled build and real corpus.
11. **Input responsibility is explicit.** Never infer that a move field is
    optional or framework-provided by observing one execution of arbitrary,
    state-dependent `DefaultsForState`. Standard behaviors contribute metadata;
    custom configurations resolve ambiguity explicitly.
12. **Dependency direction is enforced.** Generated contracts depend on public
    facade types; renderers/controllers depend on snapshots and services;
    primitives depend on action/controller interfaces; only host adapters may
    import Redux, raw API/socket modules, or internal selectors.

## Twelve creator stories and their desired minimum path

| Story | Desired pit-of-success |
|---|---|
| Tic-tac-toe | Bind one typed action to a 3x3 board; preview, disabled cells, proposal, keyboard navigation, labels, turn status, and outcome follow automatically. |
| Checkers | Select a source, preview destinations, reselect/cancel, preserve focus, reset stale selection on version change, and submit without a bespoke lifecycle state machine. |
| Memory | Register one typed card view, render a grid, bind reveal action, declaratively hold a result, animate captures, and show scores/timer accessibly. |
| Pig | Render typed zero-input Roll and Bank actions, an interactive die, turn status, rejection feedback, and optional history with little game code. |
| Solo Blackjack | Compose card zones, players, Hit/Stand actions, dealer sequence, responsive layout, and animation-gated outcome. |
| Companion Blackjack | Reuse card views on shared Table and private Hand surfaces; semantic deal zones, presence, privacy, and synchronized cross-screen flights work by default. |
| Werewolf | Compose private role, public phase, player choices, simultaneous readiness, sanitized surfaces, and verdicts without duplicating server phase rules. |
| Murder Mr Monroe | Render custom typed card faces and pieces over authored SVG artwork with named spaces, typed spatial actions, responsive scaling, stable animation anchors, and keyboard/list access. |
| Scrabble-like | Maintain a local rack-to-board draft with touch, pointer, and keyboard alternatives; validate, undo, rebase, and commit exact typed inputs. |
| Catan-like | Use authored map geometry for spaces, edges, and vertices; add pan/zoom, legal placement, resources, trading workflow, and responsive side panels. |
| 7 Wonders-like | Support simultaneous private choices, countdown/readiness, multi-step payment, reconnect, and a synchronized public reveal without leaking choices. |
| Scotland Yard-like | Render asymmetric hidden state, routes over map artwork, public movement history, reveal phases, and privacy assertions for player and observer surfaces. |

Stories 1–8 are grounded in existing renderers and should become real fixtures.
Stories 9–12 are acceptance narratives and small capability probes until a
supporting abstraction is actively being built. Do not create four speculative
fake games.

**Traceability:** Pig proves zero-input action and semantic die in Task 7;
Tic-tac-toe proves generic target state plus square-grid presentation in Task 8;
Monroe proves authored geometry in Task 9 and defers typed cards to roadmap C;
Checkers proves source/destination in roadmap A; Memory proves timers, typed
cards, and result hold in B/C/D; Blackjack proves zones/outcome and then paired
surface identity in C/D; Werewolf proves private simultaneous choices in E.
Scrabble, Catan, 7 Wonders, and Scotland Yard each receive a small compile/API
probe when their seam is designed so early APIs cannot assume numeric cells,
single-player turns, public actions, or one screen; their visual/game-specific
rules remain custom until evidence supports another shared primitive.

## Success metrics

- A freshly scaffolded game strict-compiles, production-builds, loads, and
  successfully proposes one generated typed move.
- Migrated creators import only the curated facade; deep imports are classified
  legacy/internal and are not added to the stable API manifest.
- New safe APIs, compile fixtures, and fully migrated renderers contain no `any`,
  unexplained double casts, `@ts-ignore`, or omitted safety generics.
  `@ts-expect-error` requires a reason and tracking issue outside deliberate
  negative compile-contract fixtures.
- Move submission rejects missing, extra, mistyped, nonintegral, and invalid-enum
  author inputs in development and fails safely in production.
- Canonical interactive fixtures pass keyboard and axe checks and retain focus
  across legality/state refreshes.
- Selected canonical fixtures render at phone, tablet, and desktop sizes without
  overflow or graphic-target misalignment.
- Per-viewer fixtures assert that private data is absent, not merely hidden by
  CSS.
- Standard proposal and target interactions require materially less creator code
  than today's Pig and Tic-tac-toe renderers.
- The first tranche keeps existing animation, solo, companion, and Go tests
  green; no new retry-based flake masking is introduced.
- No author-input disposition is inferred from sample `DefaultsForState`
  execution; ambiguous custom defaults fail with an actionable configuration
  diagnostic.
- Aliased and spread objects with extra move fields fail compile-contract
  fixtures, not only fresh object literals.
- Generated-client/server author-schema skew fails closed before submission.
- No proposal is silently discarded by animation, stale-state, or transport
  gating; the user and creator receive a structured blocked/rejected result.
- Renderer snapshots and generated expanded state are deeply readonly to creator
  code.
- Critical Lit tag/property/event mistakes fail intentional CI fixtures.
- No new `any`, broad double cast, or suppression is introduced. Each touched
  renderer records its remaining legacy-debt count; “migrated” means that count
  is zero. Negative compile fixtures may use `@ts-expect-error` only with the
  exact invariant they prove.
- Pig, Tic-tac-toe, and the spatial proof record before/after imports, casts,
  imperative handlers, manually synchronized properties, and renderer lines.

## Normative execution order after sub-agent review

The detailed task sections below describe work packages. Execute them in this
dependency order; do not assume their original document order is an
implementation dependency:

0. Establish reproducible green gates and a renderer-test Playwright shard.
1. Inventory actual behavior and prove facade resolution.
2A. Define the explicit author-input disposition/codec seam.
2B. Generate the author-input contract, runtime schema, fingerprint, freshness,
   and staged atomic replacement.
3A. Correct expanded renderer/component types.
3B. Prove one unregistered generated bound `GameRenderer`.
5. Establish a new green, package-scoped strict authoring project and
   `check-client` foundation.
6A. Build the minimal fixture host and a snapshot adapter over existing renderer
   properties.
7. Build typed request transport and `MoveAction`; fully migrate Pig.
4. Build the final Lit scaffold, minimal visual compositions, quickstart, and
   end-to-end conformance against the now-proven action API.
8. Build `TargetAction` and the accessible square-grid adapter; migrate
   Tic-tac-toe.
9A. Build the spatial-artwork contract and framework-owned SVG fixture.
9B. On a separate `games` branch, migrate only Monroe's geometry and Move-to-Room
   interaction; record remaining deck/card debt.
10. Run the tranche review and stop.

This order corrects three original contradictions: the scaffold no longer
teaches an action API that does not exist; the fixture host no longer assumes a
snapshot replacement that has not been designed; and Pig is only fully migrated
once `MoveAction` exists.

### Task 0: Establish reproducible green gates

- [x] Record exact currently green commands and classify every existing failure.
  Repair stale harness tests or use only a narrow temporary exclusion tied to a
  tracking issue; never create a broad accepted-failure baseline.
- [x] Create a renderer-fixture Playwright configuration that starts/stops its
  own server on an allocated port, uses `retries: 0`, and does not inherit the
  real-time animation suite's global sequential-worker constraint.
- [x] Keep real animation/companion tests in their existing sequential shard.
- [x] Require each implementation work package to state: failing test/compile
  fixture; smallest implementation; focused green command; regression commands;
  and named-file commit boundary.
- [x] Use focused commands where applicable: generator/static/stub/API Go tests,
  `npm run type-check`, new `type-check:authoring`, unit tests, and explicit
  Playwright spec paths with `--reporter=line`.

## Diagnostic policy

- Preserve native TypeScript `TSxxxx` codes and lint rule IDs.
- Use stable `BGCLIENTxxxx` codes only for framework-specific checks such as
  stale generation, internal imports, filename/tag mismatch, missing companion
  partner, malformed target geometry, duplicate deck registration, or runtime
  client/server schema mismatch.
- Initial `serve` startup fails before launching when configured game clients
  are invalid. After a valid startup, incremental errors keep the last valid page
  visible while displaying the same structured diagnostic in terminal and
  browser overlay.
- Production does not throw merely because an accessible label is game-specific;
  strict fixture checks fail, development warns, and primitives synthesize safe
  coordinate labels where possible.
- Duplicate keys, malformed geometry, out-of-bounds targets, invalid move input,
  and contract/catalog skew are deterministic structural failures and fail
  closed.

Each first-tranche `BGCLIENT` diagnostic must have a table-driven acceptance row
covering trigger, stable code, startup/build/runtime severity, fail-open versus
fail-closed behavior, remediation text, and unit/browser fixture. At minimum,
cover facade resolution, stale generated schema, missing/extra/mistyped action
input, duplicate/unknown targets, preview key/cardinality mismatch, SVG fetch/
parse/sanitization failure, duplicate/missing geometry keys, missing label/order/
anchor, piece-to-unknown-space, ambiguous overlapping rectangular hotspots, and
an action attached to a component that cannot provide interactive semantics.
Diagnostics must not leak private state or inject server/artwork markup.

## First vertical implementation tranche

### Task 1: Baseline actual behavior and facade-resolution feasibility

**Purpose:** Do not turn inaccurate declarations or accidental deep imports into
a compatibility promise.

**Files:**

- Create: `docs/superpowers/design/2026-07-15-client-renderer-baseline.md`
- Create initially: `server/static/src/client.ts` (experimental curated facade)
- Modify as needed for resolution only: `server/static/vite.config.ts`,
  `server/static/tsconfig*.json`, test configuration
- Add focused Go tests under `boardgame-util/lib/build/static/`

- [x] Record actual serialized component/stack/player shapes from representative
  Pig, Tic-tac-toe, Memory, Blackjack, Werewolf, and Monroe states.
- [x] Inventory imports, properties, events, casts, action paths, and deck view
  registration across `examples/` and `../games`; classify each symbol as legacy,
  candidate facade, experimental, or internal.
- [x] Export one deliberately small experimental facade: base renderer, `html`,
  `css`, and only the primitives needed by Pig.
- [x] Prove the same import resolves from framework source, assembled temporary
  static directory with symlinked `game-src`, TypeScript, Vite development,
  Vite production, Node/browser tests, and one external game.
- [x] Decide from evidence whether the public spelling can be an
  `@boardgame/client` alias or must initially be `../../src/client.js`. Do not
  paper over differing resolution behavior with game-specific paths.
- [x] Add tests that fail if the chosen facade cannot resolve in an assembled
  workspace.
- [x] Commit only the baseline, facade spike, configuration, and tests.

**Review gate:** An adversarial reviewer checks that `client.ts` exports no
legacy moustache/template machinery merely for convenience and that no current
declaration was accidentally blessed as stable.

### Task 2: Generate the honest creator move-input contract

**Purpose:** Strictly typing today's raw `_move_args.ts` would create incorrect
boilerplate, so define creator inputs before enforcing them.

**Likely files:**

- `boardgame-util/lib/build/moveargs/typescript.go`
- `boardgame-util/cmd_emit_move_args.go`
- move/config reflection helpers discovered during implementation
- generator tests and fixtures under `boardgame-util/lib/build/moveargs/`
- generated `_move_args.ts` or successor fixtures

- [ ] Trace where move defaults, proposer/current-player injection, enum
  association, and field serialization are represented. Document unsupported or
  ambiguous cases before changing output.
- [ ] Define the authoritative per-`MoveConfig` author-input disposition seam.
  Dispositions are `required`, `server-defaulted` (omit or explicitly override),
  `context-owned` (forbidden to the safe author API), and `unsupported`.
  Standard move behaviors contribute dispositions automatically and
  `auto.MustConfig` collects them so common move structs need no custom methods.
  Custom configuration has an explicit option for unusual/defaulted fields.
  Never infer responsibility by calling `DefaultsForState` on one example state.
- [ ] Treat unclassified supported fields as required. Manager construction or
  generation fails with move and field names on contradictory providers,
  context-owned overrides, ambiguous custom defaults, or unsupported fields.
- [ ] Allow a disposition to name a validated author codec distinct from the Go
  property/wire type. Standard codecs cover integer-as-string, finite numbers,
  booleans, enums, player indexes, and strings. Custom codecs are explicit
  advanced registrations with round-trip tests.
- [ ] Define a generated schema separating author input, expanded creator model,
  and internal form-encoded wire type while recording dispositions and codecs.
- [ ] Preserve supported native author types: integer/number, boolean, string,
  enum union, and optional fields. Serialization to form strings remains
  internal; collection types remain unsupported until the following decision
  gate proves their transport.
- [ ] Add generator fixtures for: zero-input move; one required field; configured
  default; proposer/current-player context field; optional versus zero default;
  enum; boolean; server-defaulted-but-overrideable field; context-owned forbidden
  field; state-dependent custom default; the same Go struct configured under two
  move names with different dispositions; ambiguous custom default failure; and
  explicit target-player override.
- [ ] Treat slice inputs as a decision gate, not an assumed feature. Either prove
  an unambiguous end-to-end transport for empty/repeated/numeric/enum arrays and
  server parsing, or emit an actionable unsupported-field diagnostic and defer
  slices.
- [ ] Prove Pig Roll Dice is zero-input, Tic-tac-toe Place Token requires only
  `Slot: number`, and Checkers Move Token accepts numeric indexes.
- [ ] Generate runtime metadata for exact author-input validation: required
  fields, primitive/integer constraints, enums, and serialization.
- [ ] Emit one deterministic author-schema fingerprint in the generated runtime
  module and expose the expected fingerprint through server static/game info.
  Old-client/new-server and new-client/old-server tests must fail closed before
  proposal with a structured stale-generation diagnostic.
- [ ] Generate the complete set in staging and validate it before touching
  destinations. Replace each destination by same-filesystem atomic rename. A
  pre-replacement failure leaves old files untouched; a crash during the rename
  phase may leave mixed generations, which `--check` plus the shared fingerprint
  must detect deterministically.
- [ ] Add `--check` freshness behavior; stale or partial generated contracts must
  fail and `--check` performs zero writes. Test second-game extraction failure,
  TypeScript validation failure, read-only packages, missing client directories,
  crash/skew simulation, and output determinism independent of map/package order.

**Review gate:** Compare generated author inputs against actual default behavior
for every example move family. A reviewer must specifically look for fields that
are defaulted only in some configurations and fields whose zero value is
semantically distinct from omission.

### Task 3: Correct expanded renderer types and prove one bound renderer

**Likely files:**

- `server/static/src/types/boardgame-types.ts`
- `server/static/src/types/components.d.ts`
- `boardgame-util/lib/build/gametypes/typescript.go`
- `server/static/src/components/boardgame-base-game-renderer.ts`
- generated client contract fixtures

- [ ] Name and separate the sanitized server wire payload, immutable expanded
  renderer snapshot, generated creator move input, and internal form transport.
  Creator APIs never expose form-encoded strings.
- [ ] Model a visible component with required visible metadata/`.Values`, an
  explicit hidden/opaque `{}` variant, and `null` separately for empty fixed-stack
  positions. Provide an `isVisibleComponent()` guard. Keep `.DynamicValues`
  optional where deck configuration or sanitization can omit it.
- [ ] Add compile/runtime fixtures for visible, hidden, null, fixed-stack null,
  dynamic-values, and no-static-values cases.
- [ ] Replace misleading precision with `unknown` or honest optionality when
  computed or sanitized type information is unavailable.
- [ ] Generate framework/game computed fields only from an authoritative declared
  schema. Arbitrary/state-dependent `Computed*Properties` cannot be inferred by
  one execution; absent a declaration, use separate `GameComputed`/
  `PlayerComputed` augmentation points or `unknown`. Test that conditional keys
  do not become falsely required.
- [ ] Document that TypeScript cannot prove a sanitized typed value represents
  unsanitized truth; privacy is enforced by server policy and viewer-matrix tests.
- [ ] Correct public component-stack declarations to match the implementation.
- [ ] Generate one thin, abstract, unregistered `GameRenderer` bound to a single
  `GameClientContract` tying expanded state, computed augmentation, component
  catalog, move inputs, runtime schema, and fingerprint together. It imports one
  shared facade runtime, performs no `customElements.define`, and has no widening
  generic defaults.
- [ ] Make creator snapshots and generated arrays/properties deeply readonly;
  optionally freeze snapshots in development. Service identity remains stable
  while snapshot identity changes once per installed version.
- [ ] Make zero-input moves expose `propose()` with no argument. Exact-input APIs
  infer the actual argument type and reject extra keys from object literals,
  aliased variables, and spread objects, with runtime validation as defense in
  depth.
- [ ] Add negative compile fixtures for missing fields, fresh/aliased/spread extra
  fields, forbidden context-owned fields, wrong enums, fractional integers,
  zero-input arguments, and parameterized proposal without binding.
- [ ] Keep an explicitly named unsafe compatibility base only if a real corpus
  game cannot yet migrate; record the reason and removal condition.
- [ ] Make Pig compile through the bound renderer as the first proof, but do not
  call it fully migrated until Task 7 replaces its action paths.

**Review gate:** Prove class/module identity, Lit reactivity, tree-shaking, and
custom-element registration in source and assembled builds; also prove no
duplicate Lit runtime, no facade/generated-module cycle, HMR, and two games in
one shell before generating Table, Hand, or PlayerInfo bound runtime classes.

### Task 4: Modern Lit/TypeScript scaffold and end-to-end conformance

**Files:**

- `boardgame-util/lib/stub/templates.go`
- `boardgame-util/lib/stub/main_test.go` and golden fixtures
- create a scaffold conformance test under the static build/stub packages
- minimal accurate quickstart sections in `TUTORIAL.md` and `server/README.md`

- [ ] Replace Polymer `.js` renderer and player-info templates with Lit
  TypeScript using generated bound classes and the proven facade.
- [ ] Add a deliberately small experimental composition subset so “reasonable
  zero-configuration output” is tested, not aspirational:
  `boardgame-game-surface` (responsive header/main/status/action slots),
  `boardgame-action-bar` (wrapping, touch spacing, primary/secondary grouping),
  and `boardgame-game-status` (polite live turn/pending/rejection/outcome region).
  Give each semantic CSS custom properties with fallbacks, stable high-value CSS
  parts, container-query defaults, no mandatory Material theme, zero-CSS usable
  output, and a headless/custom-markup escape hatch. Keep them experimental until
  Task 10; do not pull player panels, card zones, timers, dialogs, or history into
  this tranche.
- [ ] Generate a minimal renderer that uses the proven `MoveAction` and these
  compositions, looks intentional without custom CSS, and does not duplicate
  legality/pending/error wiring.
- [ ] Generate current file extensions, imports, custom-element registration,
  and `HTMLElementTagNameMap` support.
- [ ] Build `boardgame-util` from the checked-out source into the test temporary
  directory and invoke that binary's public noninteractive stub command; never
  use a globally installed binary. Assemble the result through the same static
  build path production uses—the generated game has no independent npm project.
- [ ] Add conformance tests that generate all contracts, strict-compile,
  production-build, and load the renderer; then mount it in the fixture host and
  assert the exact proposal. Keep a separate Pig real-server journey proving the
  same action reaches the API.
- [ ] Compile rendered documentation snippets or derive them from tested source.
- [ ] Replace the contradictory quickstart and import guidance. Defer the full
  task cookbook until MoveAction and target APIs stabilize.

### Task 5: Fatal package-scoped strict `check-client`

**Likely files:**

- `boardgame-util/cmd_check_client.go` (create) and command registration
- `boardgame-util/lib/build/static/lint.go`
- shared creator TypeScript configuration
- `server/static/package.json`
- Vite checker/overlay integration selected during implementation

- [ ] Record existing whole-framework advanced-strict debt; the count may only
  shrink. Create a new green authoring project covering the facade, all new or
  materially modified headless modules, generated contracts, fixture host, and
  selected `game-src`. Do not normalize known failures into accepted output and
  do not silently exclude new framework files.
- [ ] Add a fatal package-scoped command that runs generated-file freshness,
  authoring-project strict TypeScript, and Lit-aware checks. Whole-framework
  advanced-strict cleanup remains an explicit migration task.
- [ ] Enable `strict`, `noUncheckedIndexedAccess`,
  `exactOptionalPropertyTypes`, `noPropertyAccessFromIndexSignature`,
  `noImplicitOverride`, `useUnknownInCatchVariables`, `noImplicitReturns`, and
  fallthrough/unused checks where the corpus can support them honestly.
- [ ] Do not allow a game to weaken the shared configuration.
- [ ] Ban creator-code `any`, `@ts-ignore`, unexplained `@ts-expect-error`, and
  broad double casts through typed lint or framework checks.
- [ ] Preserve native TS/lint diagnostics. Add structured `BGCLIENT` diagnostics
  only for framework validations, with JSON output for editor/overlay use.
- [ ] Make `check-client` fatal immediately. Keep ordinary `serve` and production
  rollout changes for after the known corpus is migrated so the branch remains
  incrementally usable.
- [ ] Select and pin command-line lint/Lit analysis tools that exit nonzero in CI;
  a tsserver/editor-only plugin is insufficient. Prove the tools with failing
  fixtures for explicit `any`, unexplained suppression, unknown custom element,
  misspelled property, wrong property value, boolean attribute/property misuse,
  and invalid event-detail handler. If no checker can reliably prove a critical
  binding, provide a typed helper rather than overstating safety.
- [ ] Keep TypeScript versions consistent across framework, scaffold, assembled
  workspace, and external games. Check local/framework declarations even if
  third-party dependency declarations require a separately justified skip.
- [ ] Commit `package.json` and `package-lock.json` together for every tooling
  change; add axe dependencies with one deliberately inaccessible red fixture.
- [ ] Add an initial CI workflow once the authoring project is green: focused Go
  generator/static/stub/API tests, generation freshness, authoring strict/unit
  checks, scaffold conformance, and a Chromium renderer-fixture shard with
  retries disabled and lockfile-based dependency/browser setup.

### Task 6: Minimal deterministic renderer fixture host

**Likely files:**

- Create a small fixture host under `server/static/src/testing/` or
  `server/static/tests/fixtures/` after deciding which parts are public
- Browser component/fixture tests under `server/static/tests/renderers/`
- `server/static/playwright.config.ts`

- [ ] Introduce a tested snapshot adapter over existing renderer properties for
  fixtures; do not replace broad host property plumbing before `MoveAction` is
  green.
- [ ] Mount a renderer with a typed snapshot: state, viewing/current player,
  legality, version, outcome, and surface.
- [ ] Provide a proposal spy and request correlation suitable for later pending
  tests.
- [ ] Capture runtime/console diagnostics as test failures.
- [ ] Support explicit phone, tablet, and desktop viewports.
- [ ] Add initial fixtures for Pig and Tic-tac-toe using hand-authored expanded
  client states built with typed builders plus `satisfies`; carry a fixture schema
  version so drift fails loudly. Do not build a golden importer yet.
- [ ] Add focused axe and keyboard helpers; run them against one representative
  interaction rather than every state.
- [ ] Use reduced motion except where animation is the behavior under test.
- [ ] Keep real-server animation timing tests in their existing sequential shard;
  do not globally fake `requestAnimationFrame`, WAAPI, and companion clocks.
- [ ] Add measurable default-quality assertions at approximately 320, 768, and
  1280 CSS pixels: no horizontal overflow, visible/touchable primary action,
  preserved board aspect ratio, no overlap at 200% text zoom, visible focus in
  normal/forced-colors modes, adequate default contrast, comprehensible reduced
  motion, intentional empty/loading/error/finished states, and accessible
  wrapping/truncation of long names. Screenshot only the scaffold, grid, and
  spatial proof; test geometry numerically.

### Task 7: Exact `MoveAction` semantic core, proven in Pig

**Likely files:**

- New headless move service/action modules under `server/static/src/`
- `server/static/src/components/boardgame-base-game-renderer.ts`
- `server/static/src/components/boardgame-render-game.ts`
- `server/static/src/components/boardgame-game-view.ts`
- `server/static/src/utils/move-validation.ts`
- Pig renderer and fixture tests

- [x] Introduce stable services plus the tested snapshot adapter; do not replace
  broad host property plumbing in the same commit as action semantics. After the
  action path is green, migrate one host seam at a time with identity/reactivity
  tests. Never create one giant mutable context bag.
- [x] Replace the fire-and-forget proposal seam with a typed transport service
  returning a discriminated success, server-rejection, network-failure, blocked,
  or stale-snapshot result and a local request identity. The HTTP promise is
  sufficient for initial correlation; no server echo is required.
- [x] Use a global single-flight submission lock by default so two different
  actions from the same stale snapshot cannot race. A successful POST consumes
  its snapshot until a newer state arrives; rejected/network attempts release
  the lock for an explicit retry.
- [x] Make animation gating an explicit blocked action state, never a silently
  dropped `propose-move` event. Provide an accessible error/status surface and a
  typed telemetry callback; production returns a safe result rather than
  throwing the renderer.
- [x] Expose zero-input `this.move(moves.RollDice).propose()`.
- [x] Design `MoveAction` as separate discriminated facts rather than one
  overloaded status: baseline `availability`; bound preview
  `not-needed|unchecked|checking|legal|illegal|failed`; submission
  `idle|pending|rejected`; derived fail-closed `canPropose`; and structured
  reason. `LegalForAnyone` is structural and never enables a viewer's control.
- [x] Make `with(args)` return an immutable action tied to the snapshot/version
  that created it. `propose()` fails closed if retained across a newer version.
  Preview requests carry version, viewer, candidate-set hash, request identity,
  and cancellation; late results cannot mutate current status. Bind callable
  methods so direct Lit handler usage cannot lose `this`.
- [x] Preserve structured precondition reasons currently dropped when legality
  is derived for renderers.
- [x] Compare the generated/server author-schema fingerprints before preview or
  proposal and fail closed on skew.
- [x] Upgrade runtime validation from unknown-key warnings to exact required,
  extra, primitive, integer, finite, enum, and schema-version validation.
- [x] In development, attempted invalid/illegal proposals produce an actionable
  diagnostic and accessible feedback. Production fails safely and exposes a
  telemetry hook rather than throwing the whole renderer.
- [x] Add `boardgame-action-button` with native semantics, `.action=${action}` on
  framework interactives such as die/card/targets, and one documented binding
  adapter proven against `md-filled-button`. The binding owns activation,
  `aria-disabled`, pending/rejection/reason presentation, and a minimum 44x44 CSS
  pixel pointer target. Noninteractive components remain inert until an action is
  supplied; states are never conveyed by color alone.
- [x] Isolate ambient legacy descendant `propose-move`/`data-arg-*` scraping in a
  compatibility adapter outside the curated facade. Test that an unrelated child
  with a `proposeMove` property cannot accidentally submit. Keep arbitrary visual
  attributes separate from unsafe behavioral forwarding.
- [x] After the full known renderer corpus migrated, remove that compatibility
  adapter and the direct `proposeMove()` shortcut. Legacy renderer attributes
  are inert; `move()` is the sole creator proposal entry point and always keeps
  snapshot, preview, animation, pending, and discriminated-result semantics.
- [x] Remove component-stack `indexAttributes`/proposal forwarding as well.
  Removed behavioral keys fail loudly through `.unsafeComponentAttrs`, and a
  legacy `propose-move` attribute cannot make a die appear interactive.
- [x] Migrate Pig fully and compare creator code size and behavior to the
  baseline. The renderer fell from 54 to 46 lines and from eight direct module/
  type imports to three facade/generated imports; two manually synchronized
  disabled attributes and two ambient proposal attributes became two typed
  `.action` bindings. Casts and imperative handlers remain zero. Fixture,
  Material-adapter, keyboard/axe, transport, and stale/rejection behavior are
  covered by browser contracts.

### Task 8: `TargetAction` and accessible grid, proven in Tic-tac-toe

**Likely files:**

- `server/static/src/legal/previewLegality.ts`
- move service target adapter
- `server/static/src/components/boardgame-game-board.ts`
- `server/static/src/components/boardgame-render-game.ts`
- Tic-tac-toe renderer and fixture tests

- [x] Create a small generic target interaction for independent candidates. A
  target key is not restricted to numeric grids, though this first adapter uses
  numeric cells.
- [x] Keep `TargetAction<Key>` fully headless. Expose a generic target collection,
  a separate square-grid presentation adapter, and ordinary custom Lit markup as
  consumers; unusual boards never have to pretend to be row-major grids.
- [x] Bind exact native move arguments and own preview debouncing, stale response
  rejection, loading/error/unknown/legal states, and version reset.
  Task 7 deliberately deduplicates exact previews only per immutable action;
  collections must use this Task 8 coordinator so a large board cannot create
  an unbounded fan-out of individual preview requests.
- [x] Let the board consume one interaction object rather than separately
  receiving handlers, preview specs, disabled arrays, and refresh requests.
- [x] Choose and document one DOM focus model: roving-focus gridcells or native
  buttons inside gridcells. Illegal targets remain discoverable where grid
  navigation requires it via `aria-disabled` plus guarded activation rather than
  being removed from the focus model.
- [x] Specify initial focus when no target is legal; arrows/Home/End and row
  boundaries; Enter/Space; focus after success, rejection, version refresh, and
  selected-piece removal; selected-source versus destination semantics; and
  `aria-busy` during preview without announcing every intermediate refresh.
- [x] Synthesize occupant-aware labels such as “B3, your token, selected,” with
  author overrides and discoverable illegality reasons. Validate forced colors,
  high contrast, reduced motion, coarse pointers, 200% text zoom, and indicators
  that do not rely on color alone.
- [x] Preserve focus across preview and state refreshes.
- [x] Loudly diagnose rows x columns versus stack-size mismatch, out-of-bounds
  targets, duplicate candidates, mapper exceptions, and preview cardinality
  mismatch.
- [x] Also diagnose nonpositive/nonintegral or excessive dimensions, duplicate
  mapped keys, and nonunique accessible labels. Include a regression for the
  current coordinate-label layout being clipped by an overflow-hidden board.
- [x] Migrate Tic-tac-toe to the intended minimum path and measure the reduction
  in renderer wiring.

Task 8 result: Tic-tac-toe fell from 62 to 48 lines. Its renderer now has zero
imperative proposal handlers, preview specs, disabled-space arrays, refresh
coordination, casts, or direct API imports: one inferred `.targets(...)` value
feeds the board. The target coordinator is covered for batching, opaque result
correlation, cache reuse, abort/stale behavior, retry, invalid mappings, and
bounded work; the grid contract covers native-button keyboard navigation, loud
geometry/label validation, coordinate-label containment, responsive rectangular
geometry, reduced-motion/forced-color CSS, and axe.

**Explicitly deferred:** source/destination, branching choice flows, optimistic
draft/rebase, drag/drop, and modal ownership. Extract them from later real games;
do not make `TargetAction` a workflow DSL.

### Task 9: Authored SVG geometry and accessible interaction, then Monroe proof

**Framework files:**

- `server/static/src/components/boardgame-spatial-board.ts`
- new geometry/hotspot modules as justified by the prototype
- spatial fixture assets and browser tests

**External game files:**

- `../games/murdermrmonroe/client/boardgame-render-game-murdermrmonroe.ts`
- its authored `board.svg` and types only as needed

- [x] Land the framework primitive and a minimal copied/framework-owned SVG
  fixture first. The framework commit must not depend on uncommitted `../games`.
- [x] Define an explicit `BoardGeometry<Key>` contract containing hit region,
  keyboard order, accessible label, focus anchor, and piece/animation anchor.
  Keep those three geometry concepts separate. `TargetAction<Key>` supplies
  legality/activation; `BoardPiece<Key>` supplies stable identity, space, visual
  reference, and placement strategy.
- [x] Make `data-board-space`, `data-board-label`, and optional authored order/
  anchors (or a typed sidecar) the intended SVG model. Retain `spacePrefix` only
  as a migration adapter. Add build/check extraction or validation so literal
  keys can form a `SpaceKey` union and misspellings fail before an empty box.
- [x] Provide `piecesFromSizedStacks(...)` as a convenience adapter rather than
  making implicit stack-slot-to-SVG-index coupling the principal API.
- [x] Harden SVG loading: require `response.ok` and sane content, abort or
  generation-token stale requests when the URL changes, ignore completion after
  disconnect, reject parser errors/non-SVG roots/missing or invalid `viewBox`,
  and show an accessible visible failure state with retry where appropriate.
- [x] Define and enforce the trusted-artwork boundary. If arbitrary SVG remains
  accepted, sanitize scripts, event-handler attributes, `foreignObject`, unsafe
  links/schemes, and external resource loads before insertion. Never interpolate
  unescaped creator keys/prefixes into selectors or reflect diagnostic markup as
  HTML.
- [x] Diagnose zero/duplicate/malformed keys, duplicate order, missing labels,
  invisible/zero-sized regions, `getBBox` failures, nonfinite transforms, unknown
  piece/target keys, and stack/space cardinality mismatch.
- [x] Resolve nested SVG interaction with explicit closest/data semantics rather
  than `event.target.id`, including keys containing punctuation.
- [x] Prototype real-region pointer hit testing, a coordinated native keyboard
  focus control at the focus anchor, a visible focus/highlight layer derived from
  source geometry, and an always-available compact list representation. Use
  rectangular overlay buttons only for nonoverlapping geometry that proves them
  correct; bounding boxes are not the universal model for concave/transformed/
  overlapping regions.
- [x] Keep pieces/markers in a separate pointer-events-safe layer and expose
  stable animation anchors without obscuring component interaction or animation
  measurement.
- [x] Forward one renderer-scoped component view through spatial-board for the
  uniform common case, plus a cardinality-checked `componentViews` array for
  heterogeneous stack layers. Reject mixed or misaligned configuration loudly.
- [x] Transform each element through its own/ancestor CTM, not only the root SVG.
  Test nested transformed groups, nonzero `viewBox`, `preserveAspectRatio`
  letterboxing modes, CSS/container resize, page zoom, polygon/path regions,
  distinct region/focus/piece anchors, and rapid ResizeObserver notifications
  without loops.
- [x] Add a development geometry inspector displaying keys, order, hit regions,
  anchors, labels, unknown pieces, and overlaps.
- [x] Resolve labels from an explicit label map, SVG `aria-label`/`title`, or a
  loud development diagnostic. Supply keyboard order and a compact list
  fallback for screen-reader and small-screen use.
- [x] Connect the same typed target action/preview contract used by the grid.
- [x] After the framework proof is committed, create a dedicated `games` branch
  and migrate only Monroe's board geometry, position projection, and Move-to-Room
  interaction. Record remaining `any`, deck/card template, and legacy action debt;
  full Monroe migration waits for typed component rendering. Prove custom visuals
  remain ordinary game-owned Lit/SVG.

Task 9 proof: framework commits `7ce5833e`, `145fb0a9`, `13446504`,
`bc105a38`, and `05b00e0f`
provide the authored-geometry primitive, the nested move-context generation
fix exposed by `MoveOnGraph`, and explicit sentinel-slot projection. Games
branch `client-authored-spatial-monroe` commits `3f8ddfd` and `35f5d69`
annotate all 24 real
room regions, replaces the proposal handler and closed-room array with one
typed target collection, and projects every player/Monroe location stack with
an explicit unknown sentinel. The isolated strict client checker reports zero
diagnostics, the Monroe Go package passes, and the entire Monroe client now has
zero explicit `any`, `@ts-ignore`, `@ts-expect-error`, or double-unknown casts.
The component-view tranche subsequently removed Monroe's remaining global card/
token templates: the game-owned card front is ordinary typed Lit content, and
heterogeneous token layers use explicit spatial-board views.

**Raster adapter completed (2026-07-16):** `rasterBoardArtwork(...)` adds a
strict immutable descriptor for raster sources and normalized circle, rectangle,
and polygon regions with optional focus/piece anchors. `boardgame-spatial-board`
turns it into one nested SVG scene, so image and interaction geometry share exact
`contain`, `cover`, or `fill` transforms while reusing the existing action,
focus, list, piece, resize, error, and inspector pipeline. Unit, strict facade,
responsive browser, source-race, negative-runtime, and axe proofs cover it.
Pan/zoom, routes, edges, and vertices can follow without changing the action
contract.

### Task 10: First-tranche review and stop gate

- [ ] Record an explicit command/result matrix rather than “relevant tests”:
  focused generator/gametypes/static/stub/API Go packages; full `go test ./...`
  with any unrelated environment exclusions named; `npm run type-check`;
  `type-check:authoring`; unit tests; exact renderer fixture specs; keyboard/axe;
  320/768/1280 and 200%-text checks; scaffold conformance; and the existing
  animation/companion specs touching modified primitives.
- [x] Run a scaffold-from-clean-temporary-directory conformance test.
- [x] Have one adversarial reviewer inspect public API elegance and type escape
  hatches, one inspect real-game migration and creator code, and one inspect
  build/test robustness.
- [x] Compare success metrics and record unresolved design questions.
- [ ] Record framework and games branch names, SHAs, clean/dirty status, tests,
  remaining legacy-debt counts, and the framework minimum required by the games
  migration.
- [ ] Stop and ask for user review before implementing the later roadmap.

Task 10 review pass (2026-07-16): three independent adversarial reviews covered
the public API, Monroe migration, and build/generation lifecycle. Their blocking
findings were fixed: root SVG attributes are sanitized, custom geometry is a
callback over the actual sanitized SVG, action dispatch follows explicit regions
and covered sibling artwork, invalid artwork is removed, duplicate/orphan anchors
fail loudly, canonical keys cannot collide, inspector work is opt-in and bounded,
piece jitter is contained, contradictory rendering modes and invalid token sizes
fail loudly, and the artwork has one deliberate accessibility model. Generator
keys now default to strings with explicit safe numeric opt-in; filename collisions,
orphan outputs, size/root/namespace mismatches, read-only packages, and non-atomic
writes are handled. The starter `-d` path bug was fixed and new games now receive
strict Lit TypeScript renderer stubs rather than Polymer JavaScript.

Validation matrix so far:

- `go test ./...`: pass, including generator, gametypes, static assembly, stub,
  API, examples, animation, and companion packages.
- `npm run type-check`, `type-check:authoring`, `test:unit` (69/69),
  `test:check-client` (12/12), `test:facade-production`, and `build`: pass.
- Full renderer Playwright suite: 11/11 pass, including keyboard/axe, root SVG
  sanitizer attacks, explicit sidecar activation, sibling-overlay hit testing,
  anchor separation, stale-load races, visible failures, and inspector output.
- Clean temporary-module `stub -f -d ... authoringproof`: pass; output contains
  `.ts` Lit renderers and no legacy `.js` renderer.
- Monroe: `go test ./murdermrmonroe` passes; the full games checker reports no
  Monroe diagnostic and accepts the numeric authored-board contract.

The follow-on typed player-info tranche adds a generated `PlayerInfoRenderer`,
migrates every framework example plus Monroe, Pass, and Valentine, removes the
dynamic host's `any` assignments, and refreshes previously missing Pass and
Valentine contracts. The combined framework+games strict checker now reports
zero diagnostics across all nine configured clients. The transform matrix
covers authored `preserveAspectRatio`, nested CTMs, page zoom,
320/768/1280 widths, 200% text, and coalesced rapid resize while checking focus
alignment and contained pieces. The typed component/deck renderer remains a
deliberately deferred roadmap seam. The interaction half of that seam is now
implemented independently: `boardgame-component-stack.componentActions`
accepts one bound action or explicit `null` per logical slot and owns
mouse/keyboard activation, focus and ARIA state, live availability, reasons,
subscriptions, and animated child replacement. Cardinality, invalid actions,
and mixed legacy wiring fail loudly. Memory proves homogeneous per-card
targets; Monroe proves heterogeneous actions within one hand and a whole-pile
action; Valentine proves merged/sanitized hands. The remaining deck template
syntax can therefore evolve later without coupling renderer registration to
proposal behavior.

## Migration and fatal-check rollout

After the first tranche proves the seams, migrate in increasing complexity:

1. Pig: zero-input actions.
2. Tic-tac-toe: independent board targets.
3. Checkers: source/destination selection.
4. Murder Mr Monroe: custom components and authored graphic board.
5. Memory: typed cards, hold/capture animation, timers.
6. Solo and companion Blackjack: cards, zones, Table/Hand, animation identity.
7. Werewolf: privacy, simultaneous phases, player choices.
8. Valentine and Pass, then the remaining examples/player-info renderers.

Use compatibility adapters while a real migration is in flight, but do not
promote the legacy attribute/moustache path through the new facade. Once all
known renderers are migrated and green:

- make configured-game initial `serve` checks fatal;
- make production client checks fatal;
- remove legacy APIs rather than maintaining two co-equal authoring models,
  unless an explicit external compatibility requirement is discovered and
approved.

## Later roadmap: re-plan after Task 10

### A. Source/destination and small move drafts

- [x] Extract a `SourceDestinationController` from Checkers. It auto-resets on
  state, version, route, epoch, and proposal-perspective changes; supports
  reselect, Escape/cancel, accessible selection announcements, successful
  submission clearing, and batched bound destination preview. The controller
  owns only local source state while `TargetAction` and `boardgame-game-board`
  retain transport and presentation responsibilities.
- Use Valentine to determine whether a small partial `MoveDraft` is warranted.
  It may accumulate exact author-input fields and submit only when complete; it
  must not own layout, gestures, or modal sequencing.

### B. Typed player-info and selective services

- [x] Add a generated typed PlayerInfo base with a `playerState` getter derived
  from the exact state/player-index pair, eliminating a separately synchronized
  property. Reject invalid and out-of-range indexes loudly.
- [x] Replace bespoke chip properties/events with one typed declarative `chip`
  presentation getter. The base validates its closed text/color shape and CSS
  color, emits only on semantic changes, and the dynamic host relays it to the
  roster. Migrate Tic-tac-toe and external Murder Mr Monroe, correct the stale
  `rendererLoaded` attribute binding discovered by the lifecycle proof, and use
  Lit's safe style-map path instead of concatenated inline CSS.
- [x] Add a route-scoped timer service plus `TimerController` with closed
  second/frame cadence policies and immutable readings, so only subscribed Lit
  consumers update. Keep generated timer state as stable identity rather than a
  dishonest live-clock snapshot. Add `boardgame-timer` as the accessible
  zero-config countdown/progress composition and migrate Memory. Correct the
  underlying countdown to subtract from its installed baseline and generation-
  scope animation-frame loops so repeated state installs cannot accelerate a
  timer or leave concurrent tick loops behind.
- Add transition/history data only after a real stable contract exists.

### C. Renderer-scoped Lit-native deck/component rendering

Before changing author syntax, codify invariants for hidden components, null
slots, stable logical IDs, component type changes, renderer switches, HMR,
faux/orphan animation elements, flip/rotation/content cloning, simultaneous
renderer instances, companion timing, and identical deck names across games.

Prototype a renderer-scoped view resolver or direct render function that feeds
the existing stack lifecycle. Do not replace the global registry with another
global registry, and do not rewrite stamping and author syntax simultaneously.
Migrate one token deck, then a card deck with dynamic values, then companion
surfaces. Deprecate moustache templates only after every invariant passes.

First component-view tranche (2026-07-16):

- [x] Add renderer-scoped `cardView`, `tokenView`, and custom `componentView`
  recipes with a strict visible/hidden/empty context and typed host properties.
- [x] Feed Lit content through the existing stable stack hosts rather than
  replacing the FLIP/pooling lifecycle; verify host identity across snapshots.
- [x] Fail loudly for invalid, reused, unregistered, or type-changing custom
  component factories, and restore omitted pooled host properties.
- [x] Pass views through `boardgame-game-board`; prove token decks in
  Tic-tac-toe and Checkers and a card deck in Memory.
- [x] Switch the generated tutorial scaffold and tutorial guidance to the typed
  view path while retaining legacy deck templates only for incremental migration.
- [x] Exercise dynamic values and a registered custom component host in the
  browser lifecycle proof, in addition to their strict compile contracts.
- [x] Prove both paths in real games and migrate solo/companion Blackjack; the
  synchronized real Hit test proves renderer-scoped views preserve stable card
  identity across socket delivery, queued bundles, Table/Hand FLIP, and auto-fly.
- [x] Migrate Murder Mr Monroe as the external proof: one typed custom card
  view now serves draw, discard, and action-bound hand zones; typed player and
  Mr Monroe token views flow through the authored-SVG spatial board without a
  global deck registry or manual repeated card hosts.
- [x] Repair the companion Blackjack baseline that blocked that proof: cached
  actions now notify subscribed controls when the live animation gate changes,
  without losing stable identity or preview state; omitted context-owned move
  fields now retain server defaults while omitted required creator input still
  fails loudly. The real socket/bundle/Table/Hand/auto-fly Hit test passes.
- [x] Migrate every remaining framework example and external game renderer off
  the global deck-template registry. Pass and Valentine prove ordinary card and
  token recipes in the external corpus; DebugAnimations proves typed card/token
  views across hidden, sanitized, faux, pooled, grid, pile, fan, and stack
  transitions. Its real slow-animation To Hidden/To Visible gate test passes.
- [x] Replace ordinary untyped `componentAttrs` usage with
  `view.withProperties(...)`, whose host properties are checked against the
  concrete card/token/custom element type while bound views retain the base
  recipe's stable identity. Add `components-disabled` for display-only stacks;
  reject mixing it with per-slot actions. Rename the remaining unusual-property
  escape hatch to `.unsafeComponentAttrs` and migrate the full framework and
  external corpus so ordinary creator code contains no `componentAttrs`.

With no known game-owned `<boardgame-deck-defaults>` registrations remaining,
the legacy registry and moustache-binding fallback are removed rather than
retained as a second authoring model. Stack reconciliation now happens after
Lit commits both `.stack` and `.componentView`, so normal property source order
cannot race host creation; a truly missing view fails with an actionable error.
The registry-free path passes all 15 renderer fixtures and all 26 animation
tests (including a repaired gate-close test that now waits for reflected
renderer state instead of racing the diagnostic counter).

### D. Evidence-driven compositions

Add only pieces repeated by at least two real games:

- game surface/shell;
- zone/card-zone;
- player area/grid/panel;
- action bar;
- turn/phase/status live region;
- animation-gated outcome.

Use slots, CSS parts, theme tokens, container queries, and template methods so a
game can retain a distinctive design. Refactor Solo/Table/Hand bases to compose
these pieces internally while preserving existing helper/template overrides.

First evidence-driven primitive tranche (2026-07-16):

- [x] Add `boardgame-game-surface` after six framework renderers and all three
  external solo renderers independently repeated root headings, width/padding,
  and status/action placement. Require one semantic heading; provide default
  content plus header/status/actions/footer slots; unassigned optional regions
  vanish. Keep responsive bounds, parts, and tokens theme-neutral. Prove strict
  property contracts, wide/narrow layout, runtime diagnostics, and axe; migrate
  Pig, Tic-tac-toe, Checkers, Memory, solo Blackjack, Werewolf, Pass, Valentine,
  Murder Mr Monroe, and generated scaffold output while leaving diagnostic and
  companion-specific surfaces custom.
- [x] Replace five duplicated `isCurrentPlayer`/“Your Turn” fading callouts with
  a sentinel-aware persistent `boardgame-turn-status`. Export named client
  Observer/Admin/AnyPlayerIndex constants matching Go, replace magic numbers in
  the renderer base, and expose one typed `turnStatus` context so the common
  binding is a single property. Distinguish acting, waiting, observer/admin, and
  simultaneous perspectives; gate animation and finished state; keep custom
  phase/readiness text game-owned. Prove the pure matrix, exact runtime shape,
  strict negative contracts, live-region semantics, axe, five real examples,
  and generated scaffold output.
- [x] Complete the player collection layer with `boardgame-player-panel` for the
  arbitrary score/zone/status/action group repeated by Blackjack, Pass, and
  Valentine. Require a semantic label; provide responsive zero-CSS panel
  styling, header/status/actions/footer slots, current-player badge and
  `aria-current`, parts, and tokens. Keep elimination, selection, roles, and
  game-specific status outside the primitive. Prove strict properties,
  wide/narrow layout, optional-region cleanup, runtime diagnostics, axe, nested
  zone composition, external selected styling, and generated scaffold output.
- [x] Modernize `boardgame-status-text`, a primitive repeated by seven framework
  and external games, around a typed `.value` string/number contract instead of
  hidden slotted content plus DOM mutation observation. Register it from the
  client facade, provide polite atomic live-region semantics by default, retain
  the existing configurable fading policies, and fail loudly at runtime for
  legacy attributes/content, non-scalar values, or unknown policies. Migrate the
  full corpus, tutorial, server guide, and scaffold source/golden output; prove
  positive behavior, accessibility wiring, negative TypeScript contracts, and
  runtime diagnostics in the renderer fixture.
- [x] Tighten the underlying `boardgame-fading-text` callout primitive around
  typed scalar triggers and closed message/suppression policy unions. Preserve
  existing common-case attributes, add polite atomic announcement and reduced-
  motion behavior, handle decimal differences without truncation, restart rapid
  successive animations across a real rendered-frame boundary, and reject
  non-finite triggers or unknown policies loudly. Export its contract types and
  prove positive, restart, accessibility, and negative compile/runtime cases.
- [x] Harden the corpus-wide `boardgame-action-button` without adding creator
  ceremony: visible text remains the ordinary accessible name, while a typed
  `label` supports icon-only controls. Reject blank names and non-bound actions
  loudly, surface pending work with a reduced-motion-safe spinner, and expose
  stable button/label/spinner/status CSS parts and theme variables. Prove native
  naming, pending-to-settled state, styling hooks, strict negative contracts,
  and runtime diagnostics against a real generated Pig action.
- [x] Add the first evidence-backed layout composition,
  `boardgame-action-bar`, after Blackjack and Murder Mr Monroe independently
  proved the repeated need. Give the zero-config case a named group, wrapping
  row, consistent gap, and container-query mobile stack; keep horizontal and
  vertical orientations plus four alignments as closed typed policies. Expose a
  stable bar part/token, reject blank labels and unknown policies, prove wide,
  narrow, and vertical computed layouts in-browser, and migrate both games.
- [x] Add `targetList(...)` plus `boardgame-target-list` for the labeled target
  menu repeated by Werewolf voting and Valentine's card guessing. Preserve the
  exact target-key union through one unforgeable binding; own batched preview,
  action buttons, reasons, heading/list/empty semantics, and responsive stack or
  grid presentation. Migrate Werewolf's ordinary choice list; keep Valentine's
  richer name/description rows on direct candidate actions as the intentional
  edge-case path. Prove strict negative contracts, loud runtime failures,
  end-to-end proposal, and axe.
- [x] Make the curated facade the deterministic registration boundary for every
  public renderer element, including cards, tokens, stacks, and grid boards.
  Remove redundant deep side-effect imports from all framework and external
  renderers, and add fatal `BGCLIENT0105` policy diagnostics so future creator
  code cannot depend on undocumented transitive registration order. Keep the
  companion Table/Hand bases explicitly outside this rule until their typed
  host contracts are replaced.
- [x] Replace the Redux-connected, ambient-store `boardgame-player-badge` with
  explicit immutable `PlayerPresentation` values supplied by the renderer host.
  Give creators the one-binding `playerPresentation(index)` common case, useful
  numbered fixture fallbacks, optional compact accessible rendering, public
  facade registration, shared bounded normalization, and loud index/label/color
  failures. Migrate Memory and make deterministic player identities part of the
  renderer-fixture contract.
- [x] Make companion surfaces follow the ordinary generated-renderer pattern:
  generate fully bound `TableRenderer` and `HandRenderer` bases alongside
  `GameRenderer`, export the underlying advanced bases, seat type, and avatar
  helper from the facade, ban their deep implementation imports, and migrate
  Blackjack, Werewolf, and the companion tutorial. This removes four repeated
  generic arguments and duplicated runtime move-schema wiring from the common
  case while retaining the generic bases as the advanced escape hatch.
- [x] Remove hand-typed renderer tag names from the creator path. Generate
  base-restricted `registerGameRenderer`, `registerTableRenderer`,
  `registerHandRenderer`, and `registerPlayerInfoRenderer` decorators with the
  exact game tag embedded; migrate every framework and external renderer plus
  scaffold goldens and tutorials. Auxiliary game-local custom elements retain
  ordinary `customElements.define` as the explicit edge path.
- [x] Preserve that contract through the dynamic host instead of erasing it
  with property-bag `any`: require every loaded surface to be an actual
  `BoardgameBaseGameRenderer`, fail before hiding the fallback when a custom
  element bypasses the generated base, and narrow Table/Hand companion fields
  through their real classes. Keep interrupted-animation cleanup behind the
  animator API rather than reaching through its private stack registry.
- [x] Declare the web toolchain package as ESM so Node runs native TypeScript
  tests without reparsing every file and burying diagnostics in warnings. Move
  the one legacy CommonJS precache config to `.cjs` and make the self-starting
  Playwright config use `import.meta.url` instead of CommonJS globals. Mark the
  deliberately runtime-generated classic `client_config.js` as Vite-ignored so
  production builds preserve it externally without emitting a false bundling
  warning that obscures real diagnostics.
- [x] Re-audit the primary `TUTORIAL.md` rather than relying on newer appended
  sections: remove its stale Polymer-era `.js` location, mutable timer/component
  model, and `propose-move` instructions; teach generated bases, exact
  registration decorators, typed actions, `.Values`/`.DynamicValues`, assembled
  TypeScript clients, and `check-client` consistently from the first renderer
  section through the worked and player-info examples.
- [x] Replace unsafe/stale CLI onboarding: the quick start and offline guide no
  longer hard-code one developer's path, background unsupervised children,
  sleep for readiness, or recommend `kill -9`. They now lead with the fatal
  client gate, supervised foreground `serve`, memory/offline options, safe port
  overrides, clean shutdown, and the current game inventory. Give
  `boardgame-util` itself a concise renderer-authoring command loop.
- [x] Remove the eight accidentally committed `temp_serve_*` build trees,
  including their generated API binary, copied client tree, and Vite caches.
  Keep the existing ignore rule as the forward guard and verify the generated
  API no longer appears as a fake package in `go list ./boardgame-util/...`.
- [x] Make renderer load/registration failures visible, bounded, and
  actionable instead of console-only blank screens. Keep a failed companion
  specialization eligible for the existing solo fallback; if the final surface
  cannot load or does not extend the generated base, retain the diagram area as
  an assertive error panel that names the failure and points to `check-client`.
- [x] Carry exact host and store types through `boardgame-game-view` instead of
  erasing the generated renderer at the application shell. Expose the roster's
  join-dialog operation as an intentional public method and fail loudly if an
  animation event arrives before its renderer host exists. Correct gathering
  accessors to consume the actual expanded `Game.Computed` and
  `Players[i].Computed` shape, share the readiness lookup between panel and
  roster, and validate malformed server values at that untrusted boundary.
- [x] Preserve the distinction between server version bundles and installed
  animation bundles through the API, action, reducer, selector, and state-manager
  layers. Type initial game payloads as `GameFromServer`, make absent success
  payloads explicit failures, replace JSON cloning with `structuredClone`, and
  remove the state manager's duplicate mutable enum expansion. Keep runtime
  custom-element registration separate from type-only imports so compiler
  erasure cannot silently leave the application shell unupgraded.
- [x] Make the shared fetch envelope fail closed before endpoint-specific data
  reaches the store: parse JSON as `unknown`, require an object with an exact
  Success/Failure status, accept only bounded version metadata, and narrow
  request bodies to unknown-valued records. Share the GET/POST unwrapping path
  so diagnostics cannot drift, with unit coverage for malformed envelopes and
  failure metadata.
- [x] Validate and normalize game-info and version payloads before Redux accepts
  them: bound every untrusted collection and JSON default, report the exact
  malformed path, normalize nullable Go slices, and model wire-only distinctions
  such as numeric `PropertyType`, optional raw-state versions, nullable winners,
  and client-installed timer baselines. Reuse the corrected move-form contract
  in the legacy admin form instead of maintaining a second local shape, and
  prove the decoder against both focused unit cases and a real Pig server flow.
  Follow-up the boundary audit by explicitly copying the closed Game,
  CompanionInfo, and timer envelopes instead of spreading their wire objects;
  deep-copy Chest and raw state as opaque bounded JSON until the generated
  game-specific contract takes ownership at the renderer boundary; unify the
  duplicate chest types; and remove ambient string-to-`any` escape hatches so
  core code cannot pretend it knows creator-defined keys and unknown envelope
  fields cannot enter Redux unnoticed.
- [x] Extend the same fail-closed transport to list and create-game flows:
  replace `any[]` manager/game state with shared exact manager, variant, agent,
  player, and game-list contracts; bound and copy the Go payloads; normalize nil
  slices and the intentionally omitted non-admin list; and surface invalid data
  through the existing error UI. Correct the previously hidden PascalCase game
  model and `PhotoURL` wire spelling, make an empty manager catalog safe, and use
  the shared form encoder so variant values cannot corrupt create-game requests.
  Prove the manager decoder against the live Go list page.
- [x] Harden the authentication boundary without making local authoring brittle:
  encode every faux/Firebase identity field through the shared form transport,
  validate and copy the exact public user response before Redux accepts it, and
  distinguish a valid logout from malformed or empty server data with a loud
  sign-in error. Make supervised `boardgame-util serve` temporarily authorize
  both exact loopback origins for its allocated static port, normalize the
  documented comma-delimited allowlist consistently for legacy HTTP CORS and
  WebSockets, and prove special-character identities through the real generated
  API/Vite session. Bound config discovery to an actual absolute-parent walk so
  a config-less nested `check-client -c ...` invocation terminates instead of
  growing a relative path forever and consuming a CPU core.
- [x] Harden the companion phone join boundary: distinguish conventional HTTP
  JSON endpoints from legacy `Status` envelopes in the shared transport, carry
  auth headers and errors without raw component fetches, and validate/copy
  bounded room, seat-option, and seat-result payloads before rendering or
  navigation. Reject cross-game identity, seat-policy, cardinality, and index
  contradictions, remove the join view's ambient `any`, and prove a real guest
  can create/enter a Blackjack room and land on its Hand renderer.
- [x] Make Table host controls one typed mutation surface: serialize host
  actions so rapid clicks cannot trip the shared rate limiter, disable and mark
  controls busy while a mutation is pending, decode exact conventional-HTTP
  success bodies, preserve useful server errors, reject contradictory lock
  acknowledgements, and prove lock/unlock against the real companion server.
  The real journey also exposed and fixed a strict-boundary drift: Go nil move
  field/precondition slices arrive as `null`, so the game-info decoder now
  normalizes that canonical empty representation instead of rejecting every
  freshly created game containing a zero-input move.
- [x] Turn chat into a decoded, lifecycle-safe client service: expose separate
  view/send channel policies so read-only channels cannot become server-rejected
  guesses; validate and bound messages, channels, user maps, configuration, and
  send acknowledgements; normalize Go nil slices; serialize/abort refreshes;
  reject stale-route responses; resume polling after socket loss; deduplicate
  repeated notifications; retain drafts on failed sends; and make transport or
  contract failures visible with retry. Prove rejection, retry, form encoding,
  draft preservation, success clearing, and deduplication in an isolated browser.
- [x] Make the WebSocket stream a strict protocol boundary and an owned browser
  resource: decode bounded exhaustive frame unions instead of trusting parsed
  JSON or permissive integer prefixes; reject contradictory timing policy;
  correlate behavior-changing mode/presence frames to the active game; and only
  advertise a working socket after a valid frame. Bind callbacks and reconnect
  timers to the socket and route that created them, cancel heartbeat/reconnect
  work on removal, and restore the connection on reinsertion. Prove malformed
  and legacy frames in unit tests, preserve unknown-frame forward compatibility,
  assert the server emits mode-change game identity, and exercise replacement,
  stale callbacks, removal, and reattachment in an isolated browser.
- [x] Make dynamic renderer loading route-owned and loudly contractual: bind
  game and player-summary imports to the exact game name/ID that requested
  them; ignore superseded completions; restart interrupted loads on reinsertion;
  and re-evaluate table/hand/solo selection when only the game ID changes.
  Require loaded modules to register the exact generated tag and extend the
  correct generated base, render bounded actionable failures in the game or
  roster surface, and never silently substitute a solo renderer for a requested
  companion surface. Give every bundled example a useful typed player summary,
  and prove delayed-route races, missing registration, wrong-base failures, and
  same-game companion surface changes in an isolated browser.
- [x] Make generated registration decorators the earliest renderer contract
  boundary: validate the constructor against the exact generated surface base
  before touching the registry, replace browser-native duplicate errors with a
  game-, surface-, tag-, and constructor-named diagnostic, and keep the ordinary
  decorator common case unchanged. Prove both duplicate registration and
  cross-surface reuse through the generated Pig contract in a real browser, and
  document the four matching decorators as the only creator-facing registration
  path.
- [x] Make the component-stack contract that future zone compositions build on
  honest: move `StackLayout` onto the real custom-element class, export a
  dynamic-value guard, type stack snapshots and last-seen records, and reject
  unknown layouts or invalid numeric/spatial geometry loudly. Prove the contract
  with negative TypeScript fixtures, a browser diagnostic matrix, and the full
  external-game checker before wrapping it in a higher-level zone.
- [x] Add `boardgame-component-zone` as the evidence-backed card/token
  composition over ordinary stack, fan, grid, pile, and spread layouts. Its
  one required label supplies visible/semantic naming; count, empty state,
  responsive styling, parts, and tokens are automatic; no-action zones fail
  safe as display-only while bound per-slot actions opt into interaction.
  Keep board/spatial layouts on their dedicated lower layer. Prove the facade,
  strict negative contracts, runtime diagnostics, slot/style hooks, and real
  animation identity; migrate Blackjack, Murder Mr Monroe, and generated
  `boardgame-util stub` output as independent framework/corpus/scaffold proofs.
- [x] Extract `boardgame-game-outcome` from the companion Table banner and the
  renderer's existing authoritative finished/winners/animation properties.
  Public and personal perspectives handle named winners, cooperative wins,
  losses, and draws; the verdict remains absent until the final animation gate
  settles, then announces once through a polite atomic status region. Reject
  premature, duplicate, invalid, or mislabelled winners and sentinel viewers;
  provide parts/tokens and reduced-motion behavior. Compose it back into Table,
  solo Blackjack, and generated scaffold output.
- [x] Add `boardgame-player-grid` for the repeated arbitrary-player-panel
  collection in Blackjack, Pass, and generated scaffold output. Keep it free of
  game-state assumptions: ordinary slotted children sit in a named semantic
  region with a visible heading, useful empty state, and container-responsive
  auto-fit columns. Expose stable parts and gap/minimum-width tokens; reject
  blank labels, invalid heading levels, and blank enabled empty states. Prove
  wide/narrow layout, semantics, empty/hidden-heading modes, facade typing, and
  runtime diagnostics.

- [x] Add `boardgame-inspector` for the Mysterium/card-art inspection story:
  ordinary `thumbnail` and `detail` slots get a named trigger, visible title,
  native modal focus containment/restoration, Escape/backdrop policy, mobile
  bottom-sheet sizing, scroll containment, forced-colors/reduced-motion support,
  typed controlled-state events, parts/tokens, and loud empty-content errors.
  Keep trading/configuration workflow state outside this presentation primitive.

Readiness, trading dialogs, history/timeline, drag/drop, and advanced map controls
are separate projects driven by the corresponding stories, not a single design
system phase.

### E. Graphic-board adapters and advanced workflows

- [x] Raster artwork: normalized intrinsic hotspots and explicit object-fit
  mapping, implemented by `rasterBoardArtwork(...)` through the existing
  spatial-board interaction pipeline.
- [x] Large maps: `boardgame-board-viewport` provides a geometry-independent,
  bounded pan/zoom shell with controls, keyboard, wheel, pointer, pinch,
  resize clamping, immutable persistence, and reveal/reset APIs. Spatial boards
  opt in with `pan-zoom`; only the graphic scene transforms while errors and the
  semantic space list remain stable. Transform-aware focus/piece calculations
  prevent double scaling.
- [x] Route/graph/hex games: geometry groups now let one strict `TargetAction` scope
  itself to named tiles, edges, vertices, routes, or another creator-defined
  class without weakening exact-key validation; SVG and raster descriptors share
  the contract, inactive geometry remains available for pieces, and misspelled
  groups fail visibly. Typed `BoardPathOverlay` descriptors add accessible,
  themeable route/supply/patrol lines over either artwork source, aligned through
  responsive layout and pan/zoom using the same piece anchors. Bounded, closed
  validation keeps decorative paths separate from `TargetAction` semantics and
  fails unknown keys, duplicate IDs, malformed routes, and excessive data loudly.
- [x] Scrabble-like draft: `PlacementDraftController` owns an immutable,
  bounded item-to-target overlay with select/place and pointer assignment over
  the same checked mutation, undo/clear, safe-clear or explicit keep-valid
  snapshot rebasing, visible prune notices, and one ordinary typed
  snapshot-bound commit action. Shared `boardgame-draft-controls` makes the
  count, status, reasons, responsive Commit/Undo/Clear, and announcement path
  automatic while leaving game-specific tile and board presentation ordinary.
  Exact `draft.item(key)` and `draft.target(key)` bindings now preserve both key
  unions. `boardgame-placement-item` owns accessible rack selection, while SVG
  and raster spatial boards consume `.placementDraft` directly for keyboard and
  pointer destinations, reasons, occupancy, and exact geometry validation. The
  rectangular game board consumes the same binding with exact numeric coverage,
  so grid and graphic-board creators share one drafting model.
- [x] Multi-select/payment draft: `SelectionDraftController` adds bounded,
  immutable toggle/select/deselect, undo/clear, safe snapshot rebasing, visible
  prune notices, and one exact typed commit for cards/resources without
  pretending local state is private. Placement and selection now share the
  single structural `DraftControlsBinding` and `boardgame-draft-controls`; the
  placement-only component introduced on this branch is removed before release.
  `boardgame-selection-option` supplies the repeated accessible option shell:
  arbitrary game-owned visuals get a named 44px toggle, pressed/capacity state,
  keyboard focus, parts/tokens, and loud malformed/nested-interactive failures.
  Its single `draft.option(key)` binding preserves the controller's exact key
  union, eliminating the weaker independently assigned draft/key property pair.
- [x] Public simultaneous readiness: strict, bounded `ReadinessParticipant`
  state and `boardgame-readiness` make counts, progress, participant status,
  completion, live announcements, responsive presentation, and loud malformed
  state the default. Werewolf day voting proves the sanitized public case.
- Private simultaneous choice/reveal: keep choice submission on ordinary
  snapshot-bound actions; add server-owned request identity, visibility policy,
  and synchronized reveal before exposing any client abstraction that could
  falsely imply secrecy.
- [x] Transition-envelope privacy: animation bundles expose only readonly move
  `Name` and `Version`; raw serialized arguments, proposer, initiator, phase,
  and timestamp never bypass state sanitization through `MoveStorageRecord`.
  Type animation hooks as `ClientMove`, regression-test the exact JSON keys,
  and validate/copy the narrow envelope again at the untrusted client boundary
  so broad older-server objects cannot reach creator renderers.
- Hidden movement remainder: visibility-safe generated surfaces, an explicitly
  public game-owned log, reveal markers, and viewer-matrix privacy assertions.

## Test strategy

### Every pull request

- Affected Go tests.
- Generated-contract freshness.
- Strict TypeScript and typed lint over framework and configured game sources.
- Pure unit tests.
- Scaffold generate/compile/production-build smoke.
- Chromium fixture tests, parallelized where isolated.
- Keyboard and axe checks on one representative fixture per touched primitive.
- Phone/tablet/desktop checks for selected high-information states.
- Curated public-facade manifest diff only after that facade is declared stable.

Target 8–12 minutes. Do not screenshot every state at every viewport.

### Nightly

- Full real-server Playwright suite.
- All canonical fixture states.
- Firefox and WebKit smoke.
- Paired companion timing/reconnect tests.
- Full viewer/privacy matrix.
- SVG/raster alignment, resize, and later pan/zoom matrix.
- Animation transition suite and moderate performance budgets.

### Release gate

- Clean checkout and dependency installation.
- Scaffold a game from scratch.
- Build every example and external corpus game.
- Full production build.
- Canonical journeys in Chromium, Firefox, and WebKit.
- Companion projector/phone journey.
- Privacy exclusions and public API compatibility report.
- Migration documentation check.

### Flake policy

- Retries may collect traces but do not turn a pass-on-retry into green status.
- Use seeded fixtures, deterministic IDs, reduced motion when motion is not under
  test, and framework settlement hooks.
- Keep real-server animation tests in a small sequential shard.
- Maintain visual baselines on one pinned environment.
- Test artwork geometry numerically where possible; screenshots test appearance,
  not timing correctness.

## Explicit non-goals

- No giant declarative renderer JSON/schema.
- No Redux exposure to game creators.
- No general client-side legality engine.
- No one controller that owns every multi-step workflow, drag gesture, or modal.
- No global replacement registry for the current global deck registry.
- No speculative full Scrabble, Catan, 7 Wonders, or Scotland Yard implementation.
- No permanent silent opt-out from strict checks.
- No animation rewrite as part of this audit; new components consume the existing
  WAAPI and companion timing contracts.
