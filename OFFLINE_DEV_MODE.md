# Offline Dev Mode Guide

This guide explains how to set up and use offline dev mode for testing animations and game functionality without Firebase authentication.

## What is Offline Dev Mode?

Offline dev mode allows you to develop and test the boardgame app without requiring Firebase authentication. It's useful for:
- Testing animations in games like `debuganimations`
- Working without internet
- Avoiding Firebase API rate limits during development
- Faster iteration cycles

## Quick Start

```bash
# From the boardgame project root:
go run ./boardgame-util serve --storage memory --offline-dev-mode
```

Keep the command in the foreground. It supervises both child servers and waits
for readiness; press `Ctrl+C` once to stop them cleanly. If the default ports
are occupied, pass another explicit `--port` and `--static-port` pair instead
of killing an unknown process.

In another terminal, verify offline mode:

```bash
curl -s http://localhost:8080/client_config.js | grep offline_dev_mode
# Should output: "offline_dev_mode": true
```

The `--offline-dev-mode` flag is handled automatically: the serve command generates a `client_config.js` with `offline_dev_mode: true` in the temp directory, and Vite runs from that same temp directory so the correct config is served.

## How `boardgame-util serve` Works

1. Creates an isolated temporary assembly directory.
2. Generates fresh client contracts and assembles every configured game client.
3. Writes `client_config.js` with `offline_dev_mode: true` and matching ports.
4. Starts and supervises both the API server and Vite.
5. Keeps framework and game sources connected for HMR.
6. Terminates both children and removes temporary output when `serve` exits.

## Creating a Game in Offline Dev Mode

### URL Patterns

- **DON'T** navigate to `/game/debuganimations/new` (treats "new" as a game ID)
- **DO** use the proper game creation flow via `/list-games`

### Game Creation Flow

1. Navigate to `http://localhost:8080/list-games`
2. Select game type from the dropdown (e.g. "Animations Debugger")
3. Click "Create Game"
4. An "Offline Dev Mode" dialog appears - click any sign-in button
5. Enter a fake email (e.g. "test@example.com")
6. The game is created and you're redirected to the game page

## Troubleshooting

### "Please sign in" dialog appears (Firebase login)

Offline mode is not enabled. Verify with:
```bash
curl -s http://localhost:8080/client_config.js | grep offline_dev_mode
```
If missing, restart the server with `--offline-dev-mode`.

### "Couldn't find game with id new"

Use the game creation flow via `/list-games` instead of navigating directly.

### WebSocket not connecting

Check the API server is running:
```bash
curl --fail http://localhost:8888/api/list/manager
```
