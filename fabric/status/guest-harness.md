# Guest harness: status

Stream B, step 2 of [`../PROPOSAL.md`](../PROPOSAL.md). Branch
`claude/fabric-b-guest-harness-c2bd17`, cut from `34bfd655`. Nothing pushed,
no PR, `master` untouched. Everything is under `fabric/guest/` and
`fabric/status/`; `server/static/src` and `examples/` are unmodified.

## Outcome

A single file, `fabric/guest/dist/guest.html`, plays Memory inside
`<iframe sandbox="allow-scripts">` from JSON messages alone. It mounts the
**real `boardgame-render-game`** with Memory's **unmodified** renderer; the
fixture fallback was not needed. A complete recorded game (30 versions, to
game over) plays through with the real animator and gate.

- Bundle: **589,701 bytes** (576 KiB, 14% of 4 MiB); 162 KB gzipped.
- Tests: **21 of 21 pass** (Playwright, Chromium).
- **Results are from Node v24.22.3**, matching the repository's CI: a clean
  `npm ci`, `npm run type-check`, and `npm test` twice. The work was first
  developed and passed on Node v20.20.0 with the same results and the same
  bundle size, byte for byte.
- Start with [`../guest/README.md`](../guest/README.md); the output stream A
  needs is [`../guest/FINDINGS.md`](../guest/FINDINGS.md).

## Commits

| Commit | What |
|---|---|
| `24c05323` Add a self-contained Fabric guest that plays Memory from JSON messages | Protocol types, guest host, version queue, API shim, Vite single-file build, harness page, static server, Go replay exporter and its output, Playwright suite |
| `e75e29ca` Document the guest protocol and findings; harden the guest | `PROTOCOL.md`, `FINDINGS.md`, `README.md`, `tsconfig.json`; the guest ignores message types it does not handle; API shim trimmed to the three names actually imported; `file:` test |
| `7ca6f801` Record the guest harness status | This note |
| `8930fdd2` Note the clean-checkout verification in the guest status | Adds the clean-checkout result to this note |
| (the commit after it) | Records the rerun on Node 24 |

## What was verified, and how

All by `npm test` in `fabric/guest` (build, then Playwright on port 8719)
unless stated.

| Claim | Evidence |
|---|---|
| Runs in an opaque origin | Test asserts `window.origin === "null"`, that reading `sessionStorage` throws, and that the iframe's `sandbox` is exactly `allow-scripts` |
| The real host and renderer are mounted | Test reads the element tree: `boardgame-render-game` holding `boardgame-render-game-memory`, with the animator and effect layer in its shadow root |
| Cards render; a revealed card shows its face | 10 cards in the grid; after version 1 the card in the revealed slot has the emoji the chest says, and no other card has text |
| The reveal really animates | During the cycle `isAnimating` is true and cards have running WAAPI animations; afterwards the gate opened once and the watchdog never fired |
| `animations-done` arrives for each version | "Play all" over the 30-version game yields versions 0 to 29 in order, with no `error` message and no watchdog firing |
| A window plays every version in order | One message with versions 0 to 6 yields 1 to 6 in order with `remaining` counting down; repeated and overlapping windows add nothing; a window overlapping and extending adds only the new versions |
| A late joiner cuts to the present | First window 8 to 12 shows 12 as `initial`; then 13 and 14 play |
| Clicking a hidden card emits the right `propose-move` | Click on grid card 3 yields `Reveal Card`, `{ CardIndex: "3" }`, `snapshotVersion: 0`, `viewer: 0`. A second click while unanswered proposes nothing. The seat not to move proposes nothing |
| Move results | A rejected move frees the board for another; an accepted one advances |
| Seat switch | Switching to player 2 and to the observer cuts to that view as `initial`; "Your turn" disappears |
| Timers | The four-second hide timer at version 3 produces exactly one `tick` |
| Zero network | Over the whole game the page requested exactly `harness.html`, `dist/guest.html` and the replay JSON; no request came from the guest frame; no WebSocket. Separately, the guest plays under `default-src 'none'` with no violation |
| Works from `srcdoc` and from `file:` | One test each. The `file:` test loads `dist/guest.html` top-level with nothing beside it |
| CSP | Under a hash-only script policy with no inline style: no script violation; style violations only (details in FINDINGS) |
| The scenario replay plays | `examples/memory/client/scenario-replay.json` plays its four frames, reporting `jumpedFrom` for the versions it does not contain |
| Version queue logic | Seven unit tests in `tests/version-queue.spec.ts` |
| Types | `npm run type-check` passes over the guest and the client source it imports |
| Clean checkout | On Node 24, with `node_modules` and `dist` deleted: `npm ci`, `npm run type-check` and `npm test` pass (21 of 21, run twice), and the build reproduces 589,701 bytes |
| Go tool | `./scripts/go-local vet ./fabric/guest/tools/...` and `gofmt -l` are clean; the tool ran and produced `replays/memory-full.json` |

## Only eyeballed, or not checked

- **Visual fidelity.** Looked at screenshots of version 0, a mid-game
  mismatch with the timer bar, and the finished game. They look like the
  webapp's Memory. Not compared pixel for pixel with the webapp, and no
  screenshot test exists.
- **The action tray.** A "Choose Card Index: Slot 1 to 10" tray appears below
  the board for the seat to move, because projected move choices are sent.
  The same code and data would produce it in the webapp, but that was not
  checked against a running webapp.
- **Animation quality.** That animations run and settle is tested. That they
  look right (flip, travel to the pile, the one-second hold on a match) was
  seen only in two still screenshots, which cannot show motion.
- **npm's install-script gate.** On Node 24, `npm ci` warns that install
  scripts were not run and need approving (`npm install-scripts ls`). The
  build and tests pass regardless. `server/static/package.json` handles the
  same gate with an `allowScripts` list; the guest has none.
- **Other browsers.** Chromium only. No Firefox or Safari run.
- **The in-app browser pane.** On one early attempt the harness page loaded
  there but the guest posted nothing; headless Chromium worked from the
  first run. Not investigated.
- **Inside Common Fabric.** Nothing was run in `cf-iframe`. The policy it
  applies, how it delivers `$view`, and whether it relays guest events as
  the protocol assumes are all untested.
- **A second game.** Only Memory is bundled. Art, tokens and dice are in the
  bundle but nothing exercised them.
- **Fonts.** Omitted with fallbacks; the result was looked at, not measured.

## Open questions

1. **Does "cut to newest on first window" suit the pattern?** A guest that
   loads mid-game shows the newest entry and does not replay the window. The
   webapp instead may start from an older state to animate setup. If Fabric
   wants the replay, it is a one-line change in `version-queue.ts`.
2. **How does the host learn a move failed?** The protocol allows
   `move-result` but does not require it, and falls back to a ten-second
   timeout. If a pattern cannot answer a specific event, the timeout is the
   only failure signal, and ten seconds is a guess.
3. **Who supplies `moveInputSchemaFingerprint`?** The Go engine computes it.
   A TypeScript engine has to produce the same value or the guest must stop
   checking it. Today the guest falls back to the renderer's own.
4. **`gameId` is unused** (FINDINGS F2, F3). Once render-game stops calling
   the server and reading cookies for a non-empty id, the guest should pass
   it, so effect seeds differ per table.
5. **One replay exporter or two?** `tools/export-replay` overlaps with stream
   C's conformance exporter, which is the better source: it rebuilds timers
   per version and covers more games. After integration the harness should
   read `conformance/memory/scenario-full-game` and this tool, with its
   1.9 MB `replays/memory-full.json`, should be deleted.
6. **Should the guest report the initial entrance animation?** See F17. A
   host pacing on `animations-done` gets `initial` about half a second before
   the first render has finished animating.
7. **Two Go observations** (F16) that belong to whoever owns the engine:
   `Game.JSONForPlayer` reports the head's version, outcome and timers for
   any state, and `Game.MoveRecords` was missing the final record after a
   run.

## Not done

- Table actions, companion surfaces, synchronized timing, admin perspective
  (listed under "Not in version 1" in `PROTOCOL.md`).
- Inlined fonts or art.
- The legacy `animationOverlap` hook (F18).
- No change to `server/static/src`: every item in FINDINGS' first table is
  stream A's to make.
