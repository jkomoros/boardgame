# Diagnose Dev Server Issues

Use this skill when encountering unexplained errors during development, especially:
- Module loading failures ("Unexpected token", "SyntaxError")
- 404 errors for files that definitely exist
- Inconsistent behavior across page loads/refreshes
- Dynamic imports failing mysteriously
- Vite HMR not working correctly

## Root Cause: Multiple Dev Server Instances

The most common cause is **multiple Vite/Node dev servers running simultaneously**, causing:
- Port conflicts
- Stale module caches
- Conflicting transformations
- Race conditions in file serving

## Diagnostic Steps

### 1. Check for Multiple Processes
```bash
ps aux | grep -E "vite|node.*serve" | grep -v grep
```

If you see multiple processes, that's likely your problem.

### 2. Clean Restart
```bash
# Kill all Node processes (includes Vite servers)
killall -9 node

# Wait a moment
sleep 2

# Verify they're gone
ps aux | grep vite | grep -v grep

# Start fresh server
./boardgame-util/boardgame-util serve > server.log 2>&1 &
echo $! > server.pid

# Wait for startup
sleep 5

# Verify only ONE Vite is running
ps aux | grep vite | grep -v grep | wc -l
```

Expected output: `1` (exactly one Vite process)

### 3. Clear Browser Cache (if needed)
Hard refresh: Cmd+Shift+R (Mac) or Ctrl+Shift+R (Windows/Linux)

### 4. Verify Server is Responding
```bash
curl -s http://localhost:8080 | head -5
```

Should return HTML, not errors.

## Prevention

**Before debugging code issues, always check the environment first:**
1. How many dev server processes are running?
2. Is the correct port being served?
3. Are there stale processes from previous sessions?

## When NOT to Use This Skill

Don't use this if:
- You're getting actual syntax errors in YOUR code (check the file itself)
- Tests are failing (that's a code issue, not server)
- Build fails (check build config)
- Runtime errors in browser console that point to your code

Use this ONLY for mysterious errors in valid code that suddenly appeared.
