# Transpile pipeline: rules for every workstream agent

The plan is `fabric/TRANSPILE-PROPOSAL.md`. A coordinator session assigns
workstreams, reviews each diff, re-runs its tests and merges it into
`claude/transpile-pipeline-e23924`. A report is a claim until the coordinator
has checked it, so report only what you ran, with real counts.

## Setup

- You work in your own Git worktree, on your own local branch, based on the
  commit your brief names. Run `git log --oneline -1` first; if HEAD is not
  that commit or a descendant, `git reset --hard <that commit>` before
  changing anything. Commit locally. Never push, never open a pull request,
  never touch `master` or `fabric-factoring`.
- Go 1.27.4: always `./scripts/go-local ...`, never bare `go`, no
  `GOTOOLCHAIN`. To share a warm build cache, in every shell command that
  runs Go first run
  `export BOARDGAME_GO_CACHE_DIR=/Users/jkomoros/Code/go/src/github.com/jkomoros/boardgame/.claude/worktrees/transpile-pipeline-e23924/.cache/go-build`.
- Node: the default `node` is v18 and cannot run the tests. In every shell
  command that runs Node or npm first run
  `export PATH=$HOME/.nvm/versions/node/v24.22.3/bin:$PATH`.
- `server/static/node_modules`: symlink the coordinator's,
  `ln -s /Users/jkomoros/Code/go/src/github.com/jkomoros/boardgame/.claude/worktrees/transpile-pipeline-e23924/server/static/node_modules server/static/node_modules`.
  Never commit it.
- Do not install toolchains or global packages. If a tool is missing, stop
  and say so.
- Commands likely to run longer than two minutes go in the background.

## Constraints

- Own only the files your brief lists. Other agents are writing other files
  at the same time; an edit outside your list will collide at merge. If you
  need a change in a file you do not own, make the smallest additive edit,
  and list it under "Edits outside my files" in your report.
- Existing Go: additive only. No existing function changes behavior, no
  existing test changes, the checked-in conformance corpus stays byte for
  byte (`./scripts/go-local run ./boardgame-util conformance export --check`).
  Do not fix Go game bugs; a port reproduces Go, bugs included.
- Engine TypeScript (everything under `server/static/src/engine/` outside
  `testing/` and `*.test.ts`): no classes, no module-scope `let`, no record
  of functions at module scope, no `x as unknown as T`, no DOM, Node APIs,
  `crypto`, timers, clock, `Math.random`, `BigInt` or asynchrony; imports
  use the `.ts` extension and stay inside the engine. `sandbox.test.ts`,
  `purity.test.ts` and `rules/purity.test.ts` enforce this.
- Expected values in tests come from running Go, never from hand derivation.
  Generators live under a `testdata/gen/` directory.
- No wall-clock timeouts, sleeps, retries or polling in anything you build.
  Bound runs by counts and seeds.
- Nothing is sent to any external service or person.
- The six hand ports in `engine/games/*.ts` are not deleted or replaced.

## Baseline on the freeze base

`./scripts/go-local build ./...` passes; Go tests of
`./boardgame-util/lib/conformance/...`, `./boardgame-util/lib/lint/...`,
`./boardgame-util/lib/gamepkg/...`, `./moves/...` and `.` pass. In
`server/static`: `npm run type-check` passes, `npm run test:unit` is 1,168
of 1,168, `npm run test:layers` is 11 of 11.

## Before you report

Run, and give the real outcome of each: `./scripts/go-local build ./...`;
the Go tests of every package you touched plus
`./boardgame-util/lib/conformance/...`; in `server/static`,
`npm run type-check`, `npx tsc --noEmit -p src/engine/tsconfig.sandbox.json`,
`npm run test:unit` (with the count), `npm run test:layers`. Skip the Node
commands only if you changed no TypeScript, and the Go ones only if you
changed no Go; say which you skipped.

Commit with a message that ends with the line
`Co-Authored-By: Claude Fable 5.2 <noreply@anthropic.com>`.

Write `fabric/transpile/status/<your stream>.md`: what you built, the
exported interfaces, test results with counts, judgment calls, what is
unverified, where the proposal was wrong about the code, and edits outside
your files. Your final message repeats the branch, the commit, the
`git diff --stat` against your base, and the same facts in brief.
