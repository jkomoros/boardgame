# Client primitives Implementation Plan

**Goal:** Make the client powerful enough that a game renderer is assembly, not CSS
archaeology. Build up from the primitives every renderer independently rebuilt, then the
arrangements the Go side already models.

**Spec:** `docs/superpowers/specs/2026-07-29-batteries-included-design.md`. Read it first —
its rankings come from measured evidence across 21 games, not from taste.

## Global Constraints

- Work in `/Users/jkomoros/Code/go/src/github.com/jkomoros/boardgame/.claude/worktrees/three-d-dice`
  on `worktree-three-d-dice`. npm/npx from `server/static/`.
- **Compose, don't replace.** Every one of these should be an assembly of existing
  primitives. If something can't be built from what exists, that is a signal to fix the
  primitive, not to add machinery beside it. Say so rather than working around it.
- **Every default needs a named escape hatch.** A slot, a part, or a custom property. A
  default that cannot be overridden is worse than no default — the audit found games
  bypassing components entirely because one detail was unreachable.
- **Take values from state, never from a literal.** Four renderers hand-copied enum values
  into TypeScript. Anything you build that needs an ordered list, a label, or per-value art
  must read it from the server, not from a constant.
- Run Playwright in the FOREGROUND. Parity goldens must never be regenerated to make a test
  pass; `git status --short -- server/static/tests/animations/parity/goldens/` must be EMPTY
  unless a change is declared, in which case scope `PARITY_RECORD` to the affected spec —
  never the directory, which rewrites trace goldens from a fresh randomized deal.
- The four trace goldens must stay byte-identical: `blackjack-deal bb56cb90…`,
  `debuganimations-card-move cbf62c9b…`, `memory-reveal-one e75291f0…`, `pig-roll c1de9534…`.
- `*.test.ts` and specs are outside `tsconfig.json`, so the type checker will NOT catch a
  signature error there. Hand-check them; this has already hidden three on this branch.
- Known load flakes, acceptable as a SOLE failure after an isolated re-run: `waapi-gate:120`,
  `waapi-gate:147`, `waapi-buttons:30`, `waapi-companion:81`, `trace: memory reveal one
  card`, `waapi-attrs`, `fading-text`, `die-shape` legibility/pip tests.
- Commits: imperative subject, trailer `Co-Authored-By: Claude Fable 5 <noreply@anthropic.com>`.
  Never stage `.database` or `boardgame-util/boardgame-util`.
- **Every new component gets a `TUTORIAL.md` entry in the same commit.** The dice work found
  a fully-featured component with no tutorial section, and the audit found the resulting
  workarounds invisible in comments. Undiscoverable is indistinguishable from missing.

---

## Task 1: Layout and state vocabulary

**Files:** new shared styles module; `TUTORIAL.md`.

Five of five renderers declare their own `.horizontal{display:flex;flex-direction:row}` —
debuganimations has **eight identical copies**. Four of five invent their own "this is
selected" language: `::part(panel)` background, `outline: 3px solid`, a wrapper class, a
tinted border.

Build the row/column/center/flex vocabulary and one **state vocabulary**: `active`,
`selected`, `targetable`, `disabled`, `eliminated`. Note there are legitimately **two** kinds
of active player in the wild — whose turn it is, and who is currently responding to something
(murdermrmonroe's `current` vs `current-luck`) — so the vocabulary must express both.

Consume it in at least three existing renderers, deleting their local copies. That deletion
is the deliverable; a vocabulary nothing adopts is decoration.

## Task 2: `<boardgame-stat>`

**Files:** new component + test; `TUTORIAL.md`.

Five of five renderers hand-build labeled values: `${name} · ${p.Score} pts` by string
concatenation, `Supply<br>${n}`, `Food ${n}/${m}`, four `<div><strong>` rows,
`&nbsp;` appended to keep a row height stable.

Icon, label, value, optional capacity, optional delta animation. **Capacity comes free from
a sized stack's own `maxSize`** — `Food 3/6` should require passing the stack, not two
numbers. `boardgame-status-text` already animates a changing number; build on it rather than
beside it.

Escape hatches: slots for icon and label, a part for the value.

## Task 3: Enum metadata reaches the client

**Files:** Go enum/server serialization; generated client types; `TUTORIAL.md`.

Four renderers hand-copied enum values into TypeScript because the generated types give a
union but no ordered list, no display labels, and nowhere to hang per-value color or art:
darwin's 9 climate names, valentine's 8 card names **plus an 8-way CSS selector
cross-product**, and darwin's 36 art paths keyed on *display strings* — which break silently
when a label changes.

Ship ordered values with labels to the client, and delete every duplication. Deleting
valentine's selector cross-product is the proof.

## Task 4: Art slots on card, token and board

**Files:** `boardgame-card.ts`, `boardgame-token.ts`, board components; `TUTORIAL.md`.

`darwin` ships 46 MB of assets. Three of the four non-illustration files are **generic**
component art — a card back, a round chip, a board mat — and none can attach to any
component; all three are applied as CSS `background:` on hand-rolled divs. A shared house
style already exists (`art/house-styles/tactile-naturalist`, via `boardgame-util imagegen`).
The framework has an art pipeline and no art slots. Issue #600, open since 2018.

Card back, card face art, token art, board/mat art. Respect what already exists: the card
has a `<slot name="back">`, and the token's colours come from a filter table over red-family
SVG — art must compose with recolouring, not fight it.

## Task 5: Card face layout

**Files:** `boardgame-card.ts`; `TUTORIAL.md`.

Three games built a card face from scratch, three different ways — a `<button>` with a
`radial-gradient` corner pip and negative-margin art bleed; a separate `LitElement` with
`position:absolute; inset 0`; a grid with an `inset 0 0 0 4px` faux frame. Two more gave up
and render a bare `<h2>`. All three rediscovered the same "fill the card slot" box model by
hand.

Corner / center / footer slots, plus that box model. `boardgame-card` gives the chrome
(`face-up`, `rotated`, aspect ratio, back slot) and nothing for the face — this is the
missing half.

Also worth owning, since two games each wrote one: a pip/glyph row (a count rendered as
repeated symbols) and a "hide this stat when zero" helper.

## Task 6: `<boardgame-deck>`, `<boardgame-market>`, `<boardgame-supply>`

**Files:** three new components + tests; `TUTORIAL.md`.

Each has a Go-side counterpart shipping today and no client expression.

- **Deck** — face-down pile with a remaining count. 16 of 21 games; for Evolution and Love
  Letter the count is *rules-visible* because exhaustion ends the game.
- **Market** ← `behaviors.FaceUpMarket`. Fixed slots, face-down source with a count, refill.
  metaltrader declares two; darwin hand-builds the same concept in CSS.
- **Supply** — a typed pool with a cap, over a stack with a `same(Type)` constraint.
  metaltrader's four bowls; Splendor's gems; Catan's bank.

Build these as assemblies of `boardgame-component-stack` in the right layout. If one needs
anything the stack can't do, fix the stack.

## Task 7: `<boardgame-track>` and `<boardgame-mat>`

**Files:** two new components + tests; `TUTORIAL.md`.

- **Track** — N labeled slots, markers from multiple players on a shared rail, optional
  wraparound. All three surveys converged on this as the most-wanted missing renderer.
  Server half exists (`LocationBehavior` + ranged enum, or `Board`). darwin hand-built a
  9-cell version with `color-mix` arithmetic on a per-cell custom property.
- **Mat** — a player board of named, typed slots with per-slot capacity (default 1). One
  primitive that is simultaneously worker placement, personal engine mats, and area control.

## Task 8: Prove it — give `metaltrader` a renderer

**Files:** `/Users/jkomoros/Code/go/src/github.com/jkomoros/games/metaltrader/client/` (the
games repo).

This is the acceptance test for the whole plan, not a nice-to-have. `metaltrader` has the
richest component vocabulary in either repo and no client, *because* these batteries did not
exist. If the palette is right, its renderer is mostly assembly.

**Whatever it still needs hand-rolled CSS for names the next battery — record that, don't
quietly write the CSS.**

Note the games repo is separate and must stay in sync; check whether client API changes here
require migrating the other five renderers.

## Task 9: Migrate `darwin` off bespoke divs

**Files:** the games repo's `darwin/client/`.

498 lines, ~170 of bespoke CSS, zero component primitives, and therefore **no FLIP animation
anywhere in the game**. Migrating it onto components is the second proof — and it is the
game that most visibly gains, because its cards, tokens and mats are real components in Go
that currently animate not at all.

Its generic art (card back, chip, mat) should attach via Task 4's slots.

## Task 10: Close-out

Docs, an evidence pack recording which battery replaced which hand-rolled block with
file:line, a full verification sweep, and a final whole-branch review.
