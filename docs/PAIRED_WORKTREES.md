# Paired BOARDGAME and GAMES worktrees

Use ordinary single-repository Git worktrees for ordinary changes. Use a paired
bundle only when a feature needs unpublished changes in both BOARDGAME and
GAMES:

```text
<worktree-root>/<task>/
├── .cache/
├── go.work
├── boardgame/
└── games/
```

The worktrees provide independent Git indexes and branches. The task-local Go
workspace makes each module resolve the other module from the same bundle.

## Create a pair

From a directory containing the canonical `boardgame/` and `games/` checkouts:

```sh
task=<unique-task-name>
root=worktrees/$task

mkdir -p "$root"
git -C boardgame worktree add -b "$task" "$PWD/$root/boardgame" origin/master
git -C games worktree add -b "$task" "$PWD/$root/games" origin/master

cd "$root"
GOWORK=off go work init ./boardgame ./games
```

Use a unique branch/task name. The same name can be used in both repositories
because their branch namespaces are independent. Keep the bundle outside both
repositories so generated workspace files cannot be committed accidentally.

If the task begins from local branches instead of `origin/master`, substitute
the intended starting ref in each `git worktree add` command.

## Verify the workspace

From either worktree:

```sh
go env GOWORK
go list -m -f '{{.Path}} {{.Dir}}' \
  github.com/jkomoros/boardgame github.com/jkomoros/games
```

`go env GOWORK` should print the bundle's `go.work`, and both module directories
should point inside that bundle. Go uses the nearest applicable `go.work`; this
check catches accidental selection of a different workspace.

In agent or sandboxed sessions, use `boardgame/scripts/go-local` for Boardgame
commands. It reuses `<bundle>/.cache/go-build` rather than allocating a new Go
cache in `/tmp` for every test or review. For commands run directly in Games,
point at that same concurrency-safe cache:

```sh
cd <bundle>/games
GOCACHE="$PWD/../.cache/go-build" go test ./...
```

For a single Boardgame worktree, the wrapper uses `<worktree>/.cache/go-build`.
Both locations are ignored and deliberately live inside the directory removed
during normal teardown. To reclaim the cache without removing the worktree,
run `./scripts/go-local clean -cache` from Boardgame.

Do not persist a task path with `go env -w GOWORK=...`, commit the workspace
files, or run `go work sync` unless intentionally updating module requirements.

## Test and land

Run workspace tests from each module directory:

```sh
cd <bundle>/boardgame
./scripts/go-local test ./...

cd <bundle>/games
GOCACHE="$PWD/../.cache/go-build" go test ./...
```

Also disable the workspace before landing to verify each repository's committed
module definition:

```sh
cd <bundle>/boardgame
GOWORK=off ./scripts/go-local test ./...

cd <bundle>/games
GOWORK=off GOCACHE="$PWD/../.cache/go-build" go test ./...
```

The repositories have independent histories and cannot land atomically. If
GAMES consumes a new BOARDGAME API, publish BOARDGAME first, then update GAMES'
pinned dependency and test again:

```sh
cd <bundle>/games
GOWORK=off GOCACHE="$PWD/../.cache/go-build" \
  go get github.com/jkomoros/boardgame@<boardgame-commit>
GOWORK=off GOCACHE="$PWD/../.cache/go-build" go test ./...
```

Other cross-repository changes may use a different landing order when no new
BOARDGAME API is required.

## Clean up

After the work is integrated, confirm both worktrees are clean and remove them
through Git:

```sh
git -C <canonical-boardgame> worktree remove <bundle>/boardgame
git -C <canonical-games> worktree remove <bundle>/games
```

Before removal, preserve any uncommitted work as a commit or patch outside the
bundle. Do not keep a second recovery checkout in `/tmp`; register a real
worktree under the managed worktree root instead. If ignored artifacts prevent
worktree removal, delete `<bundle>/.cache` and retry without `--force`.

Delete local task branches with `git branch -d <task>` when Git recognizes them
as merged. A squash merge may require separate verification before branch
deletion. Remove the task-root `go.work`, `go.work.sum`, and empty directories
last. Use `git worktree prune --dry-run` only to inspect stale metadata left by
an accidentally deleted worktree.
