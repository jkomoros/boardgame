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

The unrelated untracked `SECURITY-INCIDENT-injection-canary.md` must remain
untouched. Never use `git add .` or `git add -A`.

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

## Success metrics

- A freshly scaffolded game strict-compiles, production-builds, loads, and
  successfully proposes one generated typed move.
- Migrated creators import only the curated facade; deep imports are classified
  legacy/internal and are not added to the stable API manifest.
- Safe game code contains no `any`, unexplained double casts, `@ts-ignore`, or
  omitted safety generics. `@ts-expect-error` requires a reason and tracking
  issue.
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

- [ ] Record actual serialized component/stack/player shapes from representative
  Pig, Tic-tac-toe, Memory, Blackjack, Werewolf, and Monroe states.
- [ ] Inventory imports, properties, events, casts, action paths, and deck view
  registration across `examples/` and `../games`; classify each symbol as legacy,
  candidate facade, experimental, or internal.
- [ ] Export one deliberately small experimental facade: base renderer, `html`,
  `css`, and only the primitives needed by Pig.
- [ ] Prove the same import resolves from framework source, assembled temporary
  static directory with symlinked `game-src`, TypeScript, Vite development,
  Vite production, Node/browser tests, and one external game.
- [ ] Decide from evidence whether the public spelling can be an
  `@boardgame/client` alias or must initially be `../../src/client.js`. Do not
  paper over differing resolution behavior with game-specific paths.
- [ ] Add tests that fail if the chosen facade cannot resolve in an assembled
  workspace.
- [ ] Commit only the baseline, facade spike, configuration, and tests.

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
- [ ] Define a generated schema separating author input type from wire type and
  recording required/defaulted/context-provided fields.
- [ ] Preserve native author types: integer/number, boolean, string, enum union,
  slices, and optional fields. Serialization to form strings remains internal.
- [ ] Add generator fixtures for: zero-input move; one required field; configured
  default; proposer/current-player context field; optional versus zero default;
  enum; boolean; slice; and explicit target-player override.
- [ ] Prove Pig Roll Dice is zero-input, Tic-tac-toe Place Token requires only
  `Slot: number`, and Checkers Move Token accepts numeric indexes.
- [ ] Generate runtime metadata for exact author-input validation: required
  fields, primitive/integer constraints, enums, and serialization.
- [ ] Keep generated output atomic: write temporary files, compile/validate the
  complete set, and replace old files only after success.
- [ ] Add `--check` freshness behavior; stale or partial generated contracts must
  fail.

**Review gate:** Compare generated author inputs against actual default behavior
for every example move family. A reviewer must specifically look for fields that
are defaulted only in some configurations and fields whose zero value is
semantically distinct from omission.

### Task 3: Correct wire types and prove one generated bound renderer

**Likely files:**

- `server/static/src/types/boardgame-types.ts`
- `server/static/src/types/components.d.ts`
- `boardgame-util/lib/build/gametypes/typescript.go`
- `server/static/src/components/boardgame-base-game-renderer.ts`
- generated client contract fixtures

- [ ] Correct component values to match `.Values` and `.DynamicValues`, including
  IDs, deck/game metadata, hidden/partial components, and fixed-stack null slots.
- [ ] Replace misleading precision with `unknown` or honest optionality when
  computed or sanitized type information is unavailable.
- [ ] Investigate whether computed global/player fields are reflectable. Generate
  known shapes when provable; provide an explicit game augmentation point rather
  than claiming `Record<string, unknown>` is strongly typed.
- [ ] Correct public component-stack declarations to match the implementation.
- [ ] Generate one bound `GameRenderer` whose state and move service cannot lose
  safety by omitting generics.
- [ ] Make zero-input moves expose `propose()` and exact-input moves reject extra
  fields at compile time as well as runtime.
- [ ] Keep an explicitly named unsafe compatibility base only if a real corpus
  game cannot yet migrate; record the reason and removal condition.
- [ ] Migrate Pig only as the first proof and remove its avoidable casts/legacy
  action attributes.

**Review gate:** Prove class/module identity, Lit reactivity, tree-shaking, and
custom-element registration in source and assembled builds before generating
Table, Hand, or PlayerInfo bound runtime classes.

### Task 4: Modern Lit/TypeScript scaffold and end-to-end conformance

**Files:**

- `boardgame-util/lib/stub/templates.go`
- `boardgame-util/lib/stub/main_test.go` and golden fixtures
- create a scaffold conformance test under the static build/stub packages
- minimal accurate quickstart sections in `TUTORIAL.md` and `server/README.md`

- [ ] Replace Polymer `.js` renderer and player-info templates with Lit
  TypeScript using generated bound classes and the proven facade.
- [ ] Generate a minimal renderer that looks respectable without custom CSS and
  demonstrates one typed action without duplicating legality wiring.
- [ ] Generate current file extensions, imports, custom-element registration,
  and `HTMLElementTagNameMap` support.
- [ ] Add a conformance test that creates a temporary game through the public
  CLI, generates all contracts, strict-compiles, production-builds, loads the
  renderer, and successfully proposes one move.
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

- [ ] Add a fatal package-scoped command that runs generated-file freshness,
  strict TypeScript, and Lit-aware checks over the framework plus selected
  `game-src`.
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

### Task 6: Minimal deterministic renderer fixture host

**Likely files:**

- Create a small fixture host under `server/static/src/testing/` or
  `server/static/tests/fixtures/` after deciding which parts are public
- Browser component/fixture tests under `server/static/tests/renderers/`
- `server/static/playwright.config.ts`

- [ ] Mount a renderer with a typed snapshot: state, viewing/current player,
  legality, version, outcome, and surface.
- [ ] Provide a proposal spy and request correlation suitable for later pending
  tests.
- [ ] Capture runtime/console diagnostics as test failures.
- [ ] Support explicit phone, tablet, and desktop viewports.
- [ ] Add initial fixtures for Pig and Tic-tac-toe using hand-authored expanded
  client states. Do not build a golden importer yet.
- [ ] Add focused axe and keyboard helpers; run them against one representative
  interaction rather than every state.
- [ ] Use reduced motion except where animation is the behavior under test.
- [ ] Keep real-server animation timing tests in their existing sequential shard;
  do not globally fake `requestAnimationFrame`, WAAPI, and companion clocks.

### Task 7: Exact `MoveAction` semantic core, proven in Pig

**Likely files:**

- New headless move service/action modules under `server/static/src/`
- `server/static/src/components/boardgame-base-game-renderer.ts`
- `server/static/src/components/boardgame-render-game.ts`
- `server/static/src/components/boardgame-game-view.ts`
- `server/static/src/utils/move-validation.ts`
- Pig renderer and fixture tests

- [ ] Introduce an immutable replacement `RendererSnapshot` for state/version/
  viewer/outcome and stable services for moves and animation. Do not create one
  giant mutable context bag.
- [ ] Preserve Lit reactivity and version identity with tests before replacing
  loose property plumbing broadly.
- [ ] Expose zero-input `this.move(moves.RollDice).propose()`.
- [ ] For parameterized moves expose a bound action whose status distinguishes
  `unknown`, `checking`, `legal`, and `illegal`; do not claim baseline/default
  legality is argument-specific legality.
- [ ] Preserve structured precondition reasons currently dropped when legality
  is derived for renderers.
- [ ] Scope pending/error state to a request identity or explicitly retain a
  global submission lock. Never expose a precise-looking per-action pending flag
  without a precise source.
- [ ] Prevent duplicate submission only for the same outstanding request; do not
  suppress legitimate consecutive moves with the same name.
- [ ] Upgrade runtime validation from unknown-key warnings to exact required,
  extra, primitive, integer, finite, enum, and schema-version validation.
- [ ] In development, attempted invalid/illegal proposals produce an actionable
  diagnostic and accessible feedback. Production fails safely and exposes a
  telemetry hook rather than throwing the whole renderer.
- [ ] Add one framework action button or styling-neutral consumer. Add a generic
  Lit directive only if a real Material/custom-element consumer demonstrates the
  need.
- [ ] Migrate Pig fully and compare creator code size and behavior to the
  baseline.

### Task 8: `TargetAction` and accessible grid, proven in Tic-tac-toe

**Likely files:**

- `server/static/src/legal/previewLegality.ts`
- move service target adapter
- `server/static/src/components/boardgame-game-board.ts`
- `server/static/src/components/boardgame-render-game.ts`
- Tic-tac-toe renderer and fixture tests

- [ ] Create a small generic target interaction for independent candidates. A
  target key is not restricted to numeric grids, though this first adapter uses
  numeric cells.
- [ ] Bind exact native move arguments and own preview debouncing, stale response
  rejection, loading/error/unknown/legal states, and version reset.
- [ ] Let the board consume one interaction object rather than separately
  receiving handlers, preview specs, disabled arrays, and refresh requests.
- [ ] Implement appropriate grid/gridcell semantics, roving focus, arrow-key
  navigation, Enter/Space activation, disabled/selected state, and synthesized
  coordinate labels with author overrides.
- [ ] Preserve focus across preview and state refreshes.
- [ ] Loudly diagnose rows x columns versus stack-size mismatch, out-of-bounds
  targets, duplicate candidates, mapper exceptions, and preview cardinality
  mismatch.
- [ ] Migrate Tic-tac-toe to the intended minimum path and measure the reduction
  in renderer wiring.

**Explicitly deferred:** source/destination, branching choice flows, optimistic
draft/rebase, drag/drop, and modal ownership. Extract them from later real games;
do not make `TargetAction` a workflow DSL.

### Task 9: Authored SVG board with accessible hotspots, proven in Monroe

**Framework files:**

- `server/static/src/components/boardgame-spatial-board.ts`
- new geometry/hotspot modules as justified by the prototype
- spatial fixture assets and browser tests

**External game files:**

- `../games/murdermrmonroe/client/boardgame-render-game-murdermrmonroe.ts`
- its authored `board.svg` and types only as needed

- [ ] Harden SVG loading: require `response.ok`, display a visible safe error,
  validate parser/root, and diagnose zero, duplicate, or malformed spaces.
- [ ] Resolve nested artwork targets with explicit data attributes/closest
  semantics rather than `event.target.id`.
- [ ] Separate authored geometry from action/legality. Geometry answers where a
  space/region/anchor is; target interaction answers what selecting it means.
- [ ] Support named keys such as `kitchen` alongside numeric stack slots.
- [ ] Overlay stable native hotspot controls positioned from artwork-space
  geometry. Provide selected, legal, disabled, and focus visuals without
  mutating the artwork's semantic source.
- [ ] Keep pieces/markers in a separate pointer-events-safe layer and expose
  stable animation anchors for FLIP/`animateBetween`.
- [ ] Recompute coordinate transforms using SVG CTM/`ResizeObserver`; test at
  two aspect ratios and after resize.
- [ ] Resolve labels from an explicit label map, SVG `aria-label`/`title`, or a
  loud development diagnostic. Supply keyboard order and a compact list
  fallback for screen-reader and small-screen use.
- [ ] Connect the same typed target action/preview contract used by the grid.
- [ ] Migrate Monroe and prove custom visuals remain ordinary game-owned Lit/SVG.

**Next adapter, not part of the first proof:** raster artwork plus normalized
hotspot manifest and explicit `object-fit` coordinate mapping. Pan/zoom, routes,
edges, and vertices follow without changing the action contract.

### Task 10: First-tranche review and stop gate

- [ ] Run strict client checks for the framework, examples, and `../games` paths
  touched by the tranche.
- [ ] Run Go tests, client unit tests, selected renderer fixtures, keyboard/axe,
  phone/tablet/desktop layout checks, and the full existing animation/companion
  regression suite relevant to touched primitives.
- [ ] Run a scaffold-from-clean-temporary-directory conformance test.
- [ ] Have one adversarial reviewer inspect public API elegance and type escape
  hatches, one inspect real-game migration and creator code, and one inspect
  build/test robustness.
- [ ] Compare success metrics and record unresolved design questions.
- [ ] Stop and ask for user review before implementing the later roadmap.

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

- Extract a `SourceDestinationController` from Checkers.
- Auto-reset on state/version change; preserve or cancel only through explicit
  policy.
- Support reselect, Escape/cancel, accessible selection announcement, and bound
  destination preview.
- Use Valentine to determine whether a small partial `MoveDraft` is warranted.
  It may accumulate exact author-input fields and submit only when complete; it
  must not own layout, gestures, or modal sequencing.

### B. Typed player-info and selective services

- Add a generated typed PlayerInfo base with derived `playerState` and reactive
  presentation helpers.
- Replace bespoke chip-change events and `any` assignments.
- Add selective timer signals that update subscribed consumers at an appropriate
  cadence instead of rerendering the entire state at 60Hz.
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

Timer/readiness, dialog, history/timeline, drag/drop, and advanced map controls
are separate projects driven by the corresponding stories, not a single design
system phase.

### E. Graphic-board adapters and advanced workflows

- Raster artwork: normalized intrinsic hotspots and explicit object-fit mapping.
- Large maps: viewport/pan/zoom adapter independent of board geometry.
- Route/graph/hex games: named spaces, edges, vertices, routes, and layered
  overlays without changing `TargetAction`.
- Scrabble-like draft: local overlay, undo, version rebase, exact commit, and
  mandatory non-drag alternative.
- Simultaneous choice/readiness: request/version-scoped private choices and
  synchronized reveal.
- Hidden movement: visibility-safe generated surfaces, public log, reveal
  markers, and viewer-matrix privacy assertions.

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
