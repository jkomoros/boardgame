# Quick Start - Get Testing Now! 🚀

## Status: ✅ Ready to Go!

I've already created your `config.json` with:
- ✅ Bolt DB (file-based, no setup needed)
- ✅ Firebase authentication (demo project)
- ✅ All 6 example games enabled
- ✅ Admin checking disabled for easy dev

## Start the Server (One Command!)

```bash
cd /Users/jkomoros/Code/boardgame
boardgame-util serve
```

That's it! The server will:
- Build the backend (Go API)
- Build the frontend (TypeScript/Lit)
- Start both on ports 8080 (frontend) and 8888 (API)

## Open and Test

1. **Open browser**: http://localhost:8080

2. **Create account**:
   - Click "Sign In"
   - Click "Create an account"
   - Enter: `test@example.com` / `password123`

3. **Create a game**:
   - Select game type (try "Memory" - it's simple)
   - Click "Create Game"
   - Play and watch the beautiful animations! ✨

## What's Configured

### Backend Config (`config.json`)
```json
{
  "defaultStorageType": "bolt",           // File-based DB
  "disableAdminChecking": true,          // No admin restrictions
  "allowedOrigins": "http://localhost:8080",
  "storage": { "bolt": "./dev.db" }      // Creates local file
}
```

### Frontend Config (`server/static/client_config.js`)
Already configured with:
- ✅ Firebase auth pointing to demo project
- ✅ API host: `http://localhost:8888`
- ✅ Dev mode enabled

### Games Available
All 6 example games are enabled:
- 🃏 **Blackjack** - Classic card game
- ♟️ **Checkers** - Board game
- 🎬 **Debug Animations** - Animation testing
- 🧠 **Memory** - Card matching (great for testing!)
- 🐷 **Pig** - Dice game
- ❌⭕ **Tic-tac-toe** - Simple strategy game

## Verify It's Working

### Frontend Health Check
```bash
curl http://localhost:8080
# Should return HTML
```

### API Health Check
```bash
curl http://localhost:8888/api/list/manager
# Should return JSON with game managers
```

### Check Database
```bash
ls -lh dev.db
# Should show the database file after first use
```

## Stop the Server

Press `Ctrl+C` in the terminal where `boardgame-util serve` is running.

## Testing the New Material Design 3 UI

Once logged in, you should see:
- 💜 **Purple theme** with indigo accents
- 🎨 **Rounded cards** with proper shadows
- ✨ **Smooth animations** when making moves
- 🎯 **Clean typography** using Roboto font
- 📱 **Responsive layout** that works on mobile too

## Common Issues

### Port already in use?
```bash
# Kill existing processes
lsof -ti:8080 | xargs kill -9
lsof -ti:8888 | xargs kill -9
```

### Can't create account?
- Check browser console for errors
- Firebase might be rate-limiting (wait a minute and try again)
- Try a different email address

### Games not loading?
```bash
# Make sure you're in the right directory
cd /Users/jkomoros/Code/boardgame
pwd  # Should show: /Users/jkomoros/Code/boardgame

# Try again
boardgame-util serve
```

## What Gets Created

When you run the server for the first time:

```
/Users/jkomoros/Code/boardgame/
├── config.json           ← Server configuration (already created)
├── dev.db               ← Bolt database (created on first run)
├── server/
│   └── static/
│       └── client_config.js  ← Frontend config (already exists)
```

## Development Tips

### Hot Reload Frontend
```bash
# Terminal 1: Run backend only
cd /Users/jkomoros/Code/boardgame
boardgame-util serve

# Terminal 2: Run frontend with hot reload
cd /Users/jkomoros/Code/boardgame/server/static
npm run dev
```

Then open: http://localhost:3000 (Vite dev server with instant updates)

### Watch Backend Logs
The `boardgame-util serve` terminal will show:
- API requests
- Authentication events
- Game state changes
- Any errors

### Test Animations
1. Create a Memory game
2. Click cards to flip them
3. Watch the smooth FLIP animations
4. Check for stuck animations or visual glitches

## Next Steps

After you've tested the basic setup:

1. **Run E2E Tests** (from earlier):
   ```bash
   cd server/static
   npm run test:e2e
   ```

2. **Build Your Own Game**:
   - See [TUTORIAL.md](TUTORIAL.md) for a complete guide
   - Example games in `examples/` for reference

3. **Customize the Theme**:
   - Edit `server/static/src/styles/md3-theme.css`
   - Change colors, fonts, spacing

4. **Deploy to Production**:
   - Create your own Firebase project
   - Update `config.json` → `prod` section
   - Set up hosting (Firebase, GCP, AWS, etc.)

## Need More Help?

- 📖 **Full Setup Guide**: [SETUP_GUIDE.md](SETUP_GUIDE.md)
- 🎓 **Tutorial**: [TUTORIAL.md](TUTORIAL.md)
- 🎨 **Styling Changes**: [MD3_STYLING_IMPROVEMENTS.md](MD3_STYLING_IMPROVEMENTS.md)
- 🧪 **E2E Test Results**: [E2E_TESTING_RESULTS.md](E2E_TESTING_RESULTS.md)
- 🏗️ **Architecture**: [ARCHITECTURE.md](ARCHITECTURE.md)

---

**You're all set!** Just run `boardgame-util serve` and start testing! 🎮✨
