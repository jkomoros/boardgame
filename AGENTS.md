# Agent instructions

- Work in a dedicated Git worktree. Never edit a checkout owned by another
  agent.
- A BOARDGAME-only change needs only a BOARDGAME worktree.
- If a task changes both BOARDGAME and GAMES, create paired `boardgame/` and
  `games/` worktrees under one external task directory and add a task-local
  `go.work`. Follow `docs/PAIRED_WORKTREES.md`.
- Keep `go.work` and `go.work.sum` outside both repositories and do not commit
  them. Test each repository independently before landing it.
