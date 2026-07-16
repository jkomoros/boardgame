# UX Audit: Game Authors (Companion / Table+Hand Mode)

*Written 2026-07-02. Grounded in: reading every companion-facing base class
and build path, fixing the blackjack and werewolf renderers hands-on, and
walking the "add companion mode to an existing game" journey step by step.
✅ FIXED = addressed during the audit passes.*

## The promise, measured

"Ship two TypeScript files, get the whole experience" is **real**:

- Blackjack's table view: **66 lines**. Hand view: **84 lines.**
- For those lines an author gets: capability detection + the create-form
  toggle, room codes + QR, the entire join/identity/avatar/seat flow, rate
  limiting, room lock, host controls (SkipTurn / lock / switch-to-solo),
  presence + absent handling, reconnection, per-seat avatar+name data
  delivered to the renderer, hidden-by-default roles for
  `PlayerRole`/`PlayerTeam` games with a one-tag opt-out, **and** the
  cross-screen deal animation (auto-wired: hand side by default, table side
  by rendering one `id="deal-source"` element).
- The base helpers are à-la-carte: `renderRoomCodeBanner()`,
  `renderAvatarStrip()`, `renderHostControls()`, `renderFakeDeckRow()`,
  `renderHandHeader()`, `renderTopEdgeAnchor()`, plus `playerState`,
  `viewingAs`, `isMoveCurrentlyLegal()`, typed snapshot-bound `move()` actions,
  `this.animator`.

## The five traps (ranked by how fast a new author hits them)

*Update 2026-07-02: traps 1, 2, 4, and 5 below are addressed — an
authoring guide now exists at docs/companion-mode-authoring.md; pig's
imports are fixed; the generic constraints were relaxed (`extends object`,
`CSSResultGroup` styles, `Record<string, object>` move-args) so all
in-repo companion renderers type-check with ZERO errors and honest typed
code needs no casts; and spec §5.4 now truthfully documents the restart
requirement. game-src is now type-checked in the prod build
(boardgame-util build static, non-fatal warnings). Trap 3 (zero-value masquerade) is documented in the
authoring guide and its cautionary tale is fixed in werewolf.*

1. **There is no authoring guide.** The only documentation is a design spec
   in `docs/superpowers/specs/` (spec-speak, includes unshipped features) and
   base-class doc comments. The journey "I have a game, I want companion
   mode" has no README. **Highest-leverage author fix in the repo**: a
   `docs/companion-mode-authoring.md` with the two-file recipe, the helper
   catalog, the sanitization rules, and a copy-paste skeleton. (Blackjack is
   now a good exemplar to point at.)
2. **The import-path landmine.** `../../src/components/…` works;
   `../../../server/static/src/…` 500s in dev serve with an opaque vite
   overlay. ✅ FIXED (partially): the build now prints a warning naming the
   file and the correct path. Remaining: pig still trips it (task filed);
   consider making the lint an error once existing games are clean.
3. **The zero-value masquerade.** Sanitized enum properties read as the
   enum's *zero value*, not as "hidden" — on the table, every werewolf is a
   "Villager". Werewolf's own table renderer fell for it and declared
   "Villagers Win!" from Day 1 (✅ FIXED by removing the impossible
   computation; server-side game-end filed). The rule for the authoring
   guide: **never branch on another player's sanitized property; verdicts
   must be server-computed public fields.**
4. **The types fight honest authors.** Generated `GameState`/`PlayerState`
   don't satisfy the bases' `Record<string, unknown>` constraints, so a
   type-checking author sees errors on their own class declaration and
   reaches for casts — which is exactly how the shipped
   `'Hit' as MoveName` no-op bug happened. Fix the generator to emit
   compatible types (or relax the constraints) and add game-src to a
   type-check step (task filed). Until then the honest path punishes people.
5. **Spec drift on dev hot-add.** §5.4 promises adding `-table.ts`/`-hand.ts`
   mid-serve refreshes the toggle with no restart. There is no watcher; the
   capability walk runs once at startup and is baked into the api binary.
   Either build the watcher or fix the spec — an author following the doc
   will conclude their opt-in silently failed.

## Sharp edges worth documenting (not bugs)

- **`animateBetween` direction**: the first argument *arrives from* the
  second argument's position (✅ doc fixed). Auto-fly covers the common
  case; bespoke wiring needs this.
- **Move buttons should use `isMoveCurrentlyLegal()`** — blackjack now
  demonstrates the pattern (disabled buttons + `renderHandHeader()` turn
  status). Without it, out-of-turn taps are silently dropped by client-side
  validation.
- **Label people, not seats.** `seatPresentations` is on both bases; use it
  for anything player-facing ("🐺 WolfBot2", not "Player 1"). Werewolf's
  vote buttons now demonstrate this.
- **`autoFlyIncoming` diffing** treats an id new to *any* of your stacks as
  "incoming"; games that materialize ids locally (create tokens client-side)
  should opt out.
- **Sanitization policies are static.** You cannot "reveal roles at game
  end" by flipping a policy — ship a separate public field computed
  server-side.

## Structural suggestions (future work)

1. ✅ **Template-method default for the table — DONE.** Both bases now
   ship a default `render()` composed of the standard chrome around an
   overridable `renderBoard()` / `renderHand()`. A minimal table or hand
   view is now ~10 lines (see the skeletons in the authoring guide).
   Existing games that override `render()` are unaffected.
2. **A `boardgame-util stub companion <game>` generator.** The codegen
   machinery exists (`lib/stub`); emitting the two files with correct
   imports, typed generics, and TODO markers would erase traps #1 and #2 in
   one stroke.
3. **Companion-mode "works with sanitization" checklist at boot.** The
   server already crash-validates ForceFinishTurn for companion games; it
   could also warn when a companion game's playerState has no
   `sanitize:` tags at all (a hand view with zero private state is usually a
   mistake).
4. ✅ **Game-over surface contract — DONE.** `gameFinished`, `gameWinners`,
   and `gameVersion` are plumbed to every renderer;
   `renderGameOverBanner()` (table) and the hand header verdict consume
   them. Verified with a real finished game.
5. **Second exemplar beyond cards.** Blackjack (cards) and werewolf
   (hidden roles) are both "private hand" shaped. The first author of a
   board-centric game (memory, checkers) will discover whether the hand
   view has anything to say when private state is thin — worth doing
   in-repo before an external author does.

## Verification status

Everything asserted above was exercised against the running dev server
(`boardgame-util serve --offline-dev-mode`, GOPATH set) except MySQL
storage (no server in the dev environment; migrations unexercised).
Absent-player SkipTurn was subsequently driven end-to-end with real
timing (no-show claim → absent at ~35s → host SkipTurn advances play).
