# Batteries included: components and renderers — Design

Date: 2026-07-29
Issue: #802. Follows #800 (3D dice) and #801 (3D components).
Method: three independent surveys — an outside view of the tabletop medium, a demand
inventory across 21 built and planned games, and a line-by-line audit of every renderer in
both repos.

## The thesis

**The Go side already has the vocabulary. The client side does not.**

This is not a framework that lacks abstractions. It ships `DrawDiscardPair`, `FaceUpMarket`,
`LocationBehavior` with Dijkstra pathfinding, `PlayerSubmission`, Phase/RoundRobin/
PlayerOrder, MoveBudget, Elimination, `Board`, sized stacks, `same(Type)` constraints, and
seven client stack layouts including a spatial board that takes authored SVG or raster with
pan/zoom.

The single clearest piece of evidence for where the gap actually is:

> **`metaltrader` models the richest component vocabulary in either repo — two face-up
> markets, four typed supply bowls, cost-escalation tracks, player mats, tokens riding on
> market cards — entirely in first-class framework behaviors. And it ships no `client/`
> directory at all.**

It has no renderer because there is nothing to render it *with*. `behaviors.FaceUpMarket`
has no client counterpart. Neither does a supply pool, a resource counter, or a player mat.

The second clearest:

> **`darwin`, the most component-rich game that does ship a renderer, uses zero framework
> component primitives.** No `boardgame-card`, no `boardgame-token`, no
> `boardgame-component-stack`, no zone. 498 lines, ~170 of bespoke CSS, everything divs —
> and therefore no FLIP animation anywhere in the game.

## What the games actually rebuilt

Ranked by how many of five renderers independently reimplemented it:

| Rebuilt | Games | Evidence |
|---|---|---|
| Flex layout utilities | **5/5** | `.horizontal{display:flex}` verbatim in four files; debuganimations has **eight identical copies** |
| Labeled counter / stat row | **5/5** | `${name} · ${p.Score} pts` by string concat; `Supply<br>${n}`; `Food ${n}/${m}` |
| "This is selected/active" | **4/5** | Four unrelated visual languages, incl. `::part(panel)`, `outline: 3px`, and a background class |
| Card face layout | **3 built, 2 skipped** | Three from-scratch implementations; two games gave up and render a bare `<h2>` |
| Player-panel sizing | 3/5 | `min-width: min(100%, 27rem)` vs `20rem` — different magic numbers, same missing API |
| Track / progress meter | 2/5 | darwin's 9-cell CSS grid with `color-mix` arithmetic on a `--zone` property |
| Market row with fixed slots | 2/5 | darwin hand-builds in CSS what metaltrader models as `behaviors.FaceUpMarket` |
| Supply pool | 2/5 | |
| Capacity meter over a sized stack | 2/5 | `Food N/M` where the stack's own `maxSize` is the M |
| Pip / glyph row | 2/5 | A `for` loop concatenating `"☘"`; `☀ ${n} · ❄ ${m}` |
| Destination picker | 2/5 | Eight hand-rolled buttons where `boardgame-target-list` exists |
| Rules-reference legend | 2/5 | An **8-way CSS selector cross-product**, one clause per enum value |

**No client-side `TODO`, `FIXME` or `HACK` exists anywhere in either repo.** The workarounds
are invisible in comments and visible only in the CSS. Nobody experienced this as friction —
they experienced it as how you build a renderer. That is the strongest argument that the
defaults, not the docs, are what need to change.

## The art gap, stated precisely

`darwin` ships 46 MB of assets. Three of the four non-illustration files are **fully generic
component art** — a card back, a round chip, a board mat — and **not one of them can be
attached to `boardgame-card`, `boardgame-token`, or any board component.** They are all
applied as CSS `background:` on hand-rolled divs.

A shared house-style catalog already exists (`art/house-styles/tactile-naturalist`, reachable
via `boardgame-util imagegen`). So the framework has an art *pipeline* and art *style*, and
no art *slots*. Issue #600 has been open on this since 2018.

## Design principles

1. **Match the Go vocabulary.** Every server-side behavior that describes a physical
   arrangement should have a client component named for the same concept. `FaceUpMarket` →
   `<boardgame-market>`. This is the fastest route from "modelled" to "playable" and it is
   why `metaltrader` is the test case.
2. **Compose, don't replace.** These are assemblies of existing primitives —
   `boardgame-component-stack` in a particular layout with labels and a count — not new
   rendering machinery. Anything that cannot be built from what exists is a signal to fix
   the primitive instead.
3. **Trivial by default, escapable by slot.** Every component ships a good default *and* a
   named slot to replace the part a game will want to own. A default that cannot be
   overridden is worse than no default.
4. **Take the values from the state, not from a literal.** Every renderer that hand-copies
   an enum's values into TypeScript did so because the generated types give a union but no
   ordered list, labels, or per-value art. Fixing that erases four separate duplications.

## Priority 1 — the primitives every renderer rebuilt

These are small, they are universal, and together they are most of the ~170 lines of CSS in
the worst renderer.

- **Layout utilities.** A row/column/center/flex vocabulary the framework owns. 5/5 games.
  Deliberately boring; the point is that it stops being copied.
- **`<boardgame-stat>`** — icon, label, value, optional `of` capacity, optional delta
  animation. Replaces `boardgame-status-text`'s bare-value form, which every game wraps in a
  hand-built label. Sized-stack capacity (`Food 3/6`) comes free from the stack's own
  `maxSize`.
- **A selection/active visual language.** One `active` / `selected` / `targetable` /
  `eliminated` vocabulary applied consistently, replacing four incompatible ones. Note there
  are legitimately **two** kinds of "active player" in the wild (whose turn it is, and who is
  currently responding) — the vocabulary must express both.
- **Card face layout.** A corner / center / footer slot structure, plus the "fill the card
  slot" box model that three games each rediscovered by hand. `boardgame-card` provides the
  chrome and nothing for the face.
- **Art slots** on card, token and board — so the generic art a game already generates can
  be attached without bypassing the component.
- **Enum metadata to the client.** Ordered values, display labels, and a place to hang
  per-value color/art, so no renderer needs a hand-copied array. Erases darwin's 9 climate
  names, valentine's 8 card names plus its 8-way selector cross-product, and darwin's 36
  art-path entries keyed on display strings (which break silently when a label changes).

## Priority 2 — the arrangements the Go side already models

Each of these has a server-side behavior or idiom shipping today and no client counterpart.

- **`<boardgame-market>`** ← `behaviors.FaceUpMarket`. Fixed slots, a face-down source with
  a remaining count, refill animation. Wanted by metaltrader (×2), darwin, Splendor,
  Ticket to Ride, Evolution.
- **`<boardgame-deck>`** — face-down pile with a remaining count. 16 of 21 games; and for
  Evolution and Love Letter the count is a *rules-visible* quantity because deck exhaustion
  ends the game.
- **`<boardgame-supply>`** — a typed pool with a cap, over a stack with a `same(Type)`
  constraint. metaltrader's four bowls, Splendor's gems, Catan's bank.
- **`<boardgame-track>`** — N labeled slots, markers from multiple players, optional
  wraparound. Converged on by all three surveys. The server half exists
  (`LocationBehavior` + a ranged enum, or `Board`); the client half is zero.
- **`<boardgame-mat>`** — a player board of named, typed slots with per-slot capacity. One
  primitive that is simultaneously worker placement, personal engine mats, and area control.

## Priority 3 — the state-model gaps worth closing

Deeper than any component, and each blocks a specific planned game.

- **Indexed stack families.** `sequenceforge` declares `BuildPile0..3` and `DiscardPile0..3`
  as separate fields with switch-statement accessors, and reassembles them into arrays in
  four places in the client. `metaltrader` does the same with four bowls. `Board` exists and
  would fit — **no game reaches for it**, which is itself the finding.
- **Sanitized enums read as their zero value, silently.** A shipped bug: in `werewolf` every
  hidden `Role` read as `"Villager"`, so the client declared "Villagers Win!" from the first
  Day. Catan's hidden VP cards, Codenames' key card and Mysterium's ghost screen all sit on
  this. This is a correctness bug, not a convenience gap.
- **Merged stacks cannot span decks** (`stack.go:617`). Blocks Codenames' core
  representation — 25 word cards with cover cards overlaid — and Mysterium's paired sets.
  Asked directly in games issue #36.
- **`DynamicComponentValues` has no path grammar** in the declarative-legality catalog, so
  every legality check about a card's attached state falls to custom code. Cards that carry
  state are a large slice of modern games.

## Explicitly not building

From the outside-view survey, and worth writing down so they are not re-proposed:

- **Money with denominations.** Digital play deleted making change for free.
- **Trick-taking as a battery.** Hearts, Bridge and The Crew disagree irreconcilably. Ship
  the zone — one card per player, winner takes the pile — not the rules.
- **An area-majority scorer.** Counting per player per region generalizes; awarding does
  not. Build `CountByOwner` and stop.
- **A generic "player board".** What generalizes is labeled slots with capacity. The board
  itself is art, and a creator will override it completely.
- **Player screens.** Sanitization *is* the screen, and it is better.
- **A card-effect DSL.** Every framework in this space attempts one; it always loses to
  writing Go.

## How this gets validated

Not by unit tests alone. The test is **`metaltrader` gets a renderer** — the game whose
component vocabulary is richest and whose client is empty precisely because the batteries
did not exist. If the palette is right, that renderer should be mostly assembly. If it still
needs hand-rolled CSS, the palette is not done, and *what* it hand-rolls names the next
battery.

Secondary: `darwin` should be migratable off its bespoke divs and onto components, which
would also give it the FLIP animation it currently forfeits entirely.
