# Playwright MCP Isolated Browser Configuration Fix

## Changes Made

### 1. Created `.playwright-mcp-config.json`

New configuration file at project root that tells Playwright MCP to launch an isolated browser instance:

```json
{
  "browser": {
    "type": "chromium",
    "isolated": true,
    "headless": false,
    "launchArgs": [
      "--user-data-dir=/tmp/playwright-mcp-profile",
      "--no-default-browser-check",
      "--disable-extensions",
      "--disable-infobars"
    ]
  },
  "contextOptions": {
    "viewport": {
      "width": 1280,
      "height": 720
    }
  }
}
```

**Key Settings:**
- `isolated: true` - Enables ephemeral isolated mode
- `headless: false` - Shows the browser window (with automation banner)
- `--user-data-dir=/tmp/playwright-mcp-profile` - Uses dedicated temp directory
- `viewport: 1280x720` - Consistent viewport size for testing

### 2. Updated `.mcp.json`

Added `MCP_CONFIG_FILE` environment variable to reference the new config:

```json
"env": {
  "PLAYWRIGHT_CONFIG": "./server/static/playwright.config.ts",
  "MCP_CONFIG_FILE": "./.playwright-mcp-config.json"
}
```

## Next Steps

### ⚠️ CRITICAL: Restart Required

**You must restart Claude Code to pick up the new MCP configuration:**

1. Exit Claude Code completely (Ctrl+C or Command+Q)
2. Restart Claude Code
3. MCP servers are initialized on startup with the new config

### Verification Steps

After restarting, test that the isolated browser works:

1. **Navigate to test page:**
   ```
   Use mcp__playwright__browser_navigate to http://localhost:8080
   ```

2. **Expected behavior:**
   - ✅ New Chrome window opens (NOT your personal browser)
   - ✅ Window shows "Chrome is being controlled by automated test software" banner
   - ✅ Browser uses profile in `/tmp/playwright-mcp-profile`
   - ✅ Your personal Chrome browser is unaffected

3. **Close the browser:**
   ```
   Use mcp__playwright__browser_close
   ```

4. **Verify isolation:**
   - Isolated browser window closes
   - Your personal Chrome remains unaffected

## What This Fixes

**Before:**
- ❌ Playwright MCP connected to user's personal Chrome browser
- ❌ Could interfere with personal browsing
- ❌ No clear indication of automation control

**After:**
- ✅ Playwright MCP launches its own isolated Chrome instance
- ✅ Uses dedicated temporary profile directory
- ✅ Clear automation banner visible
- ✅ User's personal browser never touched
- ✅ Safe for testing debuganimations and other games

## Troubleshooting

### If the browser still connects to personal Chrome:

1. Verify Claude Code was fully restarted (not just reconnected)
2. Check that `.playwright-mcp-config.json` exists in project root
3. Check that `.mcp.json` has the `MCP_CONFIG_FILE` env var
4. Try explicitly closing all Chrome instances before testing

### If no browser opens:

1. Check Playwright is installed: `npx playwright install chromium`
2. Verify the config file syntax is valid JSON
3. Check Claude Code logs for MCP initialization errors

### To clean up the profile directory:

```bash
rm -rf /tmp/playwright-mcp-profile
```

## Files Modified

- ✅ `.playwright-mcp-config.json` - Created (new file)
- ✅ `.mcp.json` - Updated (added MCP_CONFIG_FILE env var)

## Files Not Modified

- ⚪ `server/static/playwright.config.ts` - Unchanged (used for `npx playwright test`)
- ⚪ Any source code files - No application code changes needed

## Ready to Test

Once you restart Claude Code, you'll be able to safely test the debuganimations game using Playwright MCP tools without any risk to your personal browser.
