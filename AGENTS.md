# Agent instructions

- Work in a dedicated Git worktree. Never edit a checkout owned by another
  agent.
- A BOARDGAME-only change needs only a BOARDGAME worktree.
- If a task changes both BOARDGAME and GAMES, create paired `boardgame/` and
  `games/` worktrees under one external task directory and add a task-local
  `go.work`. Follow `docs/PAIRED_WORKTREES.md`.
- Keep `go.work` and `go.work.sum` outside both repositories and do not commit
  them. Test each repository independently before landing it.
- In agent or sandboxed sessions, run Go through `./scripts/go-local` instead
  of creating a task-named `GOCACHE` under `/tmp`. The wrapper keeps one cache
  with the worktree (or paired-worktree bundle), where normal teardown removes
  it. Go build caches are concurrency-safe; do not make a new cache per test or
  review pass.
- Put diagnostic/recovery checkouts under the managed worktree root, not in
  `/tmp`. Before teardown, inspect `git status`, preserve any needed patch or
  commit, then remove the checkout with `git worktree remove`.
