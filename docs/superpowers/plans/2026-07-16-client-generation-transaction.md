# Transactional Client Contract Maintenance

## Motivation

Game authors should have one dependable maintenance loop:

1. change Go state, moves, enums, constants, or authored board SVG;
2. run `boardgame-util emit-types` (or start `serve`);
3. receive one internally consistent generated TypeScript surface;
4. run `boardgame-util check-client` and get actionable diagnostics.

Today each generated file is replaced atomically, but the complete surface is
not failure-atomic. Move names, move inputs, board spaces, state contracts, and
renderer bases are extracted and installed in separate passes. A later build or
strict-TypeScript failure can therefore leave a mixture of generations.

## User stories

- A move rename and state change land together; the client never observes only
  one half of the change.
- A malformed SVG cannot update `_move_args.ts` before board-space extraction
  fails.
- A generated renderer that fails strict TypeScript validation changes no
  checked-in contract.
- Removing an authored board removes its orphan contract only if every other
  replacement is ready to install.
- `serve` and explicit generation have identical behavior and diagnostics.
- `check-client` remains read-only and reports every stale/orphan path in one
  deterministic result.
- A filesystem failure midway through installation restores every prior file
  and preserves a backup if restoration itself fails.
- Existing focused `emit-move-*` commands remain available, but the documented
  common path is the complete transaction.

## Design

Introduce a generated-client contract set containing all desired replacements
and orphan deletions. Extraction and schema validation populate the set without
touching destinations. Strict TypeScript validation consumes the staged move
contracts rather than reading potentially stale files from disk.

Installation sorts and validates every destination, stages every replacement
beside its destination, renames existing outputs/orphans to backups, then
installs replacements. Any failure rolls all completed mutations back in
reverse order. Backups are removed only after the entire transaction succeeds.

`emit-types`, `serve`, and generated-contract freshness checking all call this
same orchestration seam. Individual generators retain their narrow commands for
advanced use and tests.

## Invariants

- Extraction, validation, and staging failures perform zero destination writes.
- Duplicate replacement/deletion paths fail before mutation.
- Symlinks, directories, and other non-regular destinations are never replaced.
- Check mode performs no writes or deletions.
- Diagnostics and mutation order are deterministic.
- Read-only dependency packages participate in checking but not installation.
- No creator-authored file is inferred to be generated merely from its name;
  orphan deletion still requires the generated header.

## Verification

- Unit tests inject staging, replacement, orphan-removal, and rollback failures.
- Generator tests prove staged dependencies are used for strict validation.
- `emit-types --check` and `check-client` prove non-mutation and complete stale
  reporting.
- Full Go, strict TypeScript, browser renderer, production build, framework
  example, and companion-game checks remain green.
