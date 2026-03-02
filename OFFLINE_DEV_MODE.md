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
cd /Users/jkomoros/Code/go/src/github.com/jkomoros/boardgame

# Kill any existing servers
ps aux | grep "api/api" | grep -v grep | awk '{print $2}' | xargs kill 2>/dev/null
lsof -ti:8080 | xargs kill -9 2>/dev/null
lsof -ti:8888 | xargs kill -9 2>/dev/null

# Start server with offline dev mode
nohup ./boardgame-util/boardgame-util serve --offline-dev-mode > server.log 2>&1 &
sleep 5

# Verify offline mode is working
curl -s http://localhost:8080/client_config.js | grep offline_dev_mode
# Should output: "offline_dev_mode": true
```

The `--offline-dev-mode` flag is handled automatically: the serve command generates a `client_config.js` with `offline_dev_mode: true` in the temp directory, and Vite runs from that same temp directory so the correct config is served.

## How `boardgame-util serve` Works

1. Creates a temporary directory: `temp_serve_XXXXXXXXX/` (random number)
2. Copies the API binary and creates symlinks to source files in the temp directory
3. Writes a `client_config.js` with the correct settings (including `offline_dev_mode: true`) to the temp directory
4. Starts both the API server and Vite dev server from the temp directory
5. Symlinks to `server/static/src/` enable HMR (edits are picked up automatically)

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
curl -s http://localhost:8888/api/version
```
