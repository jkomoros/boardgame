# Status: declarative Memory (stream D, step 2)

Branch: `claude/fabric-d-declarative-99437f`. Date: 2026-11-21.
Main deliverable: [`fabric/declarative/FINDINGS.md`](../declarative/FINDINGS.md).

## Verdict

The ultimate goal is workable for Memory and Blackjack. Both boards run with
no renderer class, from plain markup and one JSON property per state, on the
existing animator. Animation survives element recreation and two tables on
one page, but only with two mechanisms the existing elements lack: a scoped
animation root, and geometry remembered by component id.

## Commits

| Commit | What |
|---|---|
| `b95c90e5` | The `bg-` elements, pages A, B, C and Blackjack, the replay exporter and replays, the first 18 tests |
| `fcce7b74` | Characterization tests for recreation, scope, cohorts and geometry; paint-time sampling |
| the commit after it | `FINDINGS.md` and this note |

## What was verified, and how

All by `npx playwright test` in `fabric/declarative` (Chromium, port 5199).
34 tests, passing on six consecutive full runs.

| Claim | Evidence |
|---|---|
| Cards render with the right faces | Faces compared with the replay's chest, per slot, on pages A, B (three modes) and both tables of C |
| A reveal turns a hidden card face up | `faceUp` and text before and after; a running transform animation observed |
| A captured pair lands in the right won-cards stack | Component ids in the winner's stack; the other stack empty; 18 cards left in the grid |
| An animation ran, not a jump cut | Per-frame positions sampled at paint time: first frame within 35% of the path, at least three frames mid-path, never backwards, ending at the destination; plus a running Web Animations transform animation |
| A click emits the expected `propose-move` | Real mouse click; detail matched and JSON round-tripped; a scripted click and a click on a revealed card emit nothing |
| Without remembered geometry, recreation is a jump cut | Same sampling with `measure=live`: zero mid-path frames, all three modes |
| In `mode=all` every card node is replaced | Set intersection of card nodes before and after: empty |
| Unscoped, one table cuts the other's animation | Running animation count on the left table before and after the right table installs: 2 then 0 unscoped, unchanged scoped |
| Unscoped, the same game on two tables does not animate | Animated cycles `[0, 0]` unscoped, both positive scoped |
| No layout or timing event reaches the document | Listeners for five event types on `document`: none fired |
| Blackjack per-seat hidden card | Face up for `viewer=1`, face down with empty suit and rank for `viewer=0` |

Also run: `go vet ./fabric/declarative/...` and `tsc --noEmit`, both clean.
The checked-in `scenario-replay.json` was played through page A by hand in
the browser pane (`?replay=scenario`): reveal, mismatch, hide all correct.

## Not verified

- Nothing ran under Common Fabric or across a worker boundary (step D3).
- Recreation was simulated in one synchronous task.
- Effects, transfers, motion release, timers, dice, tokens, `board` and
  `spatial` layouts, and multi-step input were not exercised.
- Only Chromium. Reduced-motion was not tested.
- The claim that Memory's "Match" callout never plays in the webapp is
  inferred from replay data, not observed.

## Things the next person should know

- `@playwright/test` is pinned to 1.58.2 so it uses the Chromium build
  already in the local cache. A newer version needs `npx playwright install`.
- The pages import `server/static/src` in place. `vite.config.ts` dedupes
  `lit` to this package, because that tree has no `node_modules` in a fresh
  worktree.
- `replays/*.json` are generated. Regenerate with
  `./scripts/go-local run ./fabric/declarative/replays/cmd/export`.
- `bg-table` writes four private fields of the animator and overrides one
  getter on its inner stack. Any change to
  `boardgame-component-animator.ts`'s `prepare()` or `_doAnimate()` can break
  it; the tests will say so.

## Open questions

1. Does the pattern renderer recreate an element when its own properties
   change? If so the table is recreated every state and the queue must live
   in module scope.
2. Are removal and insertion one task under the real renderer?
3. Can a pattern set CSS custom properties?
4. How should multi-step input (selection, placement) be expressed?
5. Should the table take the raw wire view or a pre-expanded one?
