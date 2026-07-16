# Quick start

The repository includes a development configuration and seven example games.
From the repository root, verify every game client before starting the server:

```sh
go run ./boardgame-util check-client
```

Then start the supervised API and Vite development servers:

```sh
go run ./boardgame-util serve --storage memory --offline-dev-mode
```

Open <http://localhost:8080>. Offline development mode uses local faux
authentication, so it does not require a Firebase connection or real account.
Create a game from the game list rather than constructing a `/game/.../new`
URL.

`serve` owns both child processes. Keep it in the foreground while developing;
press `Ctrl+C` once to stop the API and Vite cleanly. It assembles configured
game clients, regenerates their contracts, and provides frontend hot reload.

## What is available

The development configuration contains:

- Blackjack
- Checkers
- Debug Animations
- Memory
- Pig
- Tic-tac-toe
- Werewolf

Memory is a useful first animation smoke test. Pig is the smallest typed-action
example. Tic-tac-toe demonstrates typed targets, and Checkers demonstrates a
source/destination interaction.

## Useful checks

The frontend should return HTML:

```sh
curl --fail http://localhost:8080/
```

The API should return manager metadata:

```sh
curl --fail http://localhost:8888/api/list/manager
```

For a file-backed development database instead of ephemeral memory storage,
omit `--storage memory`; the repository configuration uses Bolt and creates
`dev.db` on first use.

## Port conflicts

Choose another explicit pair of ports; do not kill unrelated processes:

```sh
go run ./boardgame-util serve \
  --storage memory \
  --offline-dev-mode \
  --port 8889 \
  --static-port 8081
```

Open the static port you selected. The generated client configuration points
Vite at the matching API port automatically.

## Create a game package

From a games repository with a boardgame config:

```sh
boardgame-util stub examplegame
boardgame-util config add games github.com/USERNAME/REPONAME/examplegame
boardgame-util check-client
boardgame-util serve --offline-dev-mode
```

`stub` emits Lit TypeScript, generated game/state/move contracts, typed renderer
bases, exact registration decorators, and accessible responsive starter
compositions. `check-client` is the fatal local/CI gate for stale generation,
strict TypeScript, Lit bindings, unsafe escape hatches, and deep imports.

See [TUTORIAL.md](TUTORIAL.md) for game authoring and
[OFFLINE_DEV_MODE.md](OFFLINE_DEV_MODE.md) for the offline security model.
