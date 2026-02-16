# Backend Server Setup Issue

## Problem

The `boardgame-util serve` command fails with an npm build error:

```
npm error gyp ERR! build error
node-pre-gyp ERR! build error [grpc package]
```

**Root Cause:** The old `grpc@1.20.0` package (deprecated) doesn't compile on modern Node.js (v20+) on Apple Silicon Macs.

## Why This Happens

The `boardgame-util serve` command tries to:
1. Build the Go API server
2. Build the static frontend with npm
3. The frontend has old Firebase dependencies that include `grpc`
4. `grpc` won't compile on Node v20 + macOS ARM64

## Solutions

### Option 1: Use Older Node Version (Quickest)

Use Node v16 or v14 which works with the old grpc package:

```bash
# Install nvm if you don't have it
brew install nvm

# Install Node 16
nvm install 16
nvm use 16

# Try again
cd /Users/jkomoros/Code/boardgame
boardgame-util serve
```

### Option 2: Skip grpc Installation (Workaround)

The grpc package isn't actually needed for local development. You can try:

```bash
# Install without optional dependencies
cd /Users/jkomoros/Code/boardgame/server/static
npm install --no-optional --legacy-peer-deps

# Then run serve
cd /Users/jkomoros/Code/boardgame
boardgame-util serve
```

### Option 3: Run Frontend & Backend Separately (Current State)

**Frontend (already working):**
```bash
cd /Users/jkomoros/Code/boardgame/server/static
npm run dev  # Runs on http://localhost:3000
```

**Backend (needs solution):**
The Go API server needs to be started separately, but `boardgame-util serve` is the standard way to do this.

### Option 4: Update Firebase Dependencies (Best Long-term)

Update the Firebase SDK in `server/static/package.json` to remove grpc:

```json
{
  "dependencies": {
    "firebase": "^10.0.0"  // Already on v10, should use @grpc/grpc-js
  }
}
```

The issue is that something in the dependency tree is pulling in the old grpc package.

## Current Status

✅ **Frontend working**: Modern Material Design 3 UI on http://localhost:3000
❌ **Backend not running**: Can't test full authentication and game creation
⚠️ **Blocked by**: npm build failure for old dependencies

## What We Tested So Far

Without the backend, we successfully tested:
- ✅ Frontend renders with beautiful MD3 styling
- ✅ UI components work (buttons, switches, dropdowns)
- ✅ Navigation works
- ✅ Sign In dialog opens
- ❌ Can't actually authenticate (no backend)
- ❌ Can't create games (no backend)
- ❌ Can't test animations with real game state (no backend)

## Recommended Next Step

**Try Option 1 (Node 16)** - This is the quickest solution:

```bash
# Install Node 16
nvm install 16
nvm use 16

# Verify version
node --version  # Should show v16.x.x

# Clean cache and try again
boardgame-util clean cache
cd /Users/jkomoros/Code/boardgame
boardgame-util serve
```

This should build successfully and start both servers on:
- Frontend: http://localhost:8080
- API: http://localhost:8888

Then we can test the full end-to-end experience with Playwright!

## Alternative: Docker

If you have Docker, you could run the backend in a container with the right Node version:

```dockerfile
FROM golang:1.21
RUN curl -fsSL https://deb.nodesource.com/setup_16.x | bash -
RUN apt-get install -y nodejs
# ... rest of setup
```

But this is more complex for local development.

## Files Involved

- `/Users/jkomoros/Code/boardgame/config.json` - Backend config (✅ ready)
- `/Users/jkomoros/Code/boardgame/server/static/package.json` - Has old dependencies
- `/Users/jkomoros/Code/boardgame/server/static/client_config.js` - Frontend config (✅ ready)

## Error Log

Full error in: `/private/tmp/claude-501/-Users-jkomoros-Code-boardgame/tasks/ba31862.output`

Key error:
```
npm warn deprecated grpc@1.20.0: This library will not receive further updates
gyp ERR! build error
node-pre-gyp ERR! build error
```
