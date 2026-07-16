# Client Renderer Implementation Baseline

Recorded on 2026-07-15 from branch `client-renderer-authoring-audit` before the
public authoring facade and primitives were introduced.

## Required green gates

Run Go commands from the repository root and npm commands from `server/static`.
Use a writable cache path in restricted environments.

```bash
GOCACHE=/tmp/boardgame-go-cache go test ./boardgame-util ./boardgame-util/lib/build/api ./boardgame-util/lib/build/static ./server/api
npm run type-check
npm run test:unit
npm run test:renderer
```

At this baseline, the ordinary TypeScript check and all 35 Node unit tests pass.
The self-contained renderer smoke test passes with one attempt and verifies the
same-origin API/Vite configuration produced by arbitrary allocated ports.

The real-time shard is also green with zero retries when run against its
documented manually started server:

```bash
# Terminal 1, repository root
boardgame-util serve --storage memory --offline-dev-mode

# Terminal 2, server/static
npx playwright test --config playwright.config.ts --reporter=line
```

All 36 tests pass. The baseline repair removed hard-coded nonexistent game IDs,
made admin harness setup deterministic, and made the Memory fixtures
authenticate, post the server's real form contract, require a real game state,
and use shadow-piercing locators intentionally. No test is excluded or retried.

## Classified existing debt

`npm run type-check:strict` is not green and is not an accepted-failure gate.
It exposes broad pre-existing advanced-strictness debt across the application,
including unchecked indexed access, optional-property exactness, nullable DOM
lookups, legacy event typing, and library declaration boundaries. The first
tranche therefore adds a narrow `type-check:authoring` project that is fully
strict from its first commit; it does not weaken or silently allowlist the
existing strict configuration.

The existing `test:e2e` shard is intentionally not used as the renderer fixture
gate even though it is green. It requires a manually started shared server, one
worker, and contains real-time animation/companion scenarios whose state and
timing constraints are different from isolated renderer contract tests. Those
tests remain in their own shard; new renderer behavior must not depend on their
retries.

In a restricted filesystem sandbox, `go test ./...` is not a meaningful green
gate: legacy codegen tests write golden output beside their package, and Bolt
and filesystem storage tests create local database directories. Those packages
fail with `operation not permitted` before testing behavior. The focused Go
command above covers every package changed by this tranche without treating an
environmental write denial as accepted product debt; the unrestricted full Go
suite remains the pre-merge gate.

Node currently emits `MODULE_TYPELESS_PACKAGE_JSON` warnings while running the
unit suite. They do not hide failures, but package/module metadata should be
rationalized as part of the TypeScript tooling modernization rather than by an
unrelated baseline-only change.
