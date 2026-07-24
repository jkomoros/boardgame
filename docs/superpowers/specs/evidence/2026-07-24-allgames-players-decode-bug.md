# Evidence pack: AllGames entries crash the games-list decoder (pre-existing bug)

**Claim:** With Admin Mode on and at least one game in storage, the client throws
"The server returned an invalid games list" in a blocking modal, breaking every
offline-dev-mode e2e test that proposes moves (they all enable Admin Mode via
`tests/animations/helpers.ts` `createOfflineGame`).

**Server side:** `server/api/main.go:1027` populates `AllGames` directly from
`s.storage.ListGames(100, listing.All, "", gameName)` — raw
`extendedgame.CombinedStorageRecord`s with no `Players` field. The
`Participating*`/`Visible*` lists instead go through `listGamesWithUsers`
(`main.go:1021-1022`), which augments each record with `Players`.

**Client side:** `decodeGamesListResponse` (`server/static/src/types/list-response.ts`,
introduced by commit `d8b60e68` "Type and validate game list transport") required a
`Players` array on **every** entry, including `AllGames`. Absent key → throw →
error dialog intercepts all pointer events.

**Reproduction:** fetch `/api/list/game?admin=1` on an offline-dev server with ≥1
stored game: `AllGames[0]` has no `Players` key; `ParticipatingActiveGames[0]` in the
same payload does. Confirmed the unmodified pre-existing test
`tests/animations/waapi-gate.spec.ts` ("memory: card reveal completes cleanly") fails
identically — this is independent of the unification work.

**Why the fix is correct and minimal:** the server has never sent `Players` for
`AllGames` (admin listing is a raw storage dump); the decoder is the new, stricter
party. Fix: absent `Players` decodes to `[]`; a present-but-malformed `Players`
still throws. Unit-tested in `list-response.test.ts` (reproduces the exact bare
entry shape from the live payload).

**Scope note:** this fix is required for the parity harness (Phase 0) to run at
all; it is a test-blocking bug fix, not a behavior change to animation code.
