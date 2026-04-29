# TODO: Improve type safety of the propose-move pipeline

**Current state:** Move proposals flow from client to server as untyped string dictionaries. Components dispatch `propose-move` CustomEvents with `detail: { name: string, arguments: Record<string, any> }`. The arguments object is assembled by hand at each call site — gathering pickers hardcode field names like `TargetPlayerIndex` and `SelectedTeam` as string keys with string values. The server's `getMoveFromForm` in `server/api/context.go` then parses each string into the correct Go type (`PlayerIndex`, `enum.Val`, `int`, `bool`) via `strconv.Atoi` / `SetStringValue`.

**Why it's fragile:** Nothing at compile time connects a client-side field name to its Go struct field. A typo in `TargetPlayerIndex` or a renamed field silently produces a runtime error. The type coercions (e.g., `String(this.viewingAsPlayer)` to pass a number as a string that the server parses back to `PlayerIndex`) are implicit conventions, not enforced contracts.

**Scope:** This affects all moves, not just gathering. Gathering pickers, game-specific renderers, and the move-form component all have the same gap.

**Proposed improvement:** Generate per-move TypeScript argument interfaces from the server's `formFields` metadata (which already enumerates field names and types). A lighter alternative: a validation layer in `proposeMove` that checks incoming arguments against the `MoveForm.Fields` metadata already available on the client, logging or rejecting mismatches before the server round-trip.

**Risk of inaction:** Silent runtime failures from field name typos or type mismatches, with no safety net until a user hits the broken path. Refactoring move structs on the Go side has no way to surface breakage in client code.

**Key files:**
- `server/static/src/components/boardgame-base-game-renderer.ts` — propose-move dispatch
- `server/static/src/components/boardgame-move-form.ts` — proposeMove + submitForm
- `server/static/src/components/gathering-shared.ts` — gathering move dispatch sites
- `server/api/context.go` — server-side move form deserialization
