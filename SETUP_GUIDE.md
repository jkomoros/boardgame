# Development Environment Setup Guide

This guide will help you set up your local development environment to run and test the boardgame app with full backend support.

## Prerequisites

✅ You already have:
- Go installed (with `boardgame-util` available at `/Users/jkomoros/Code/go/bin/boardgame-util`)
- Node.js and npm
- Frontend dependencies installed

## What You Need to Set Up

### 1. Firebase Project (for Authentication)

**Option A: Use Existing Demo Project (Quickest)**

The sample config includes a demo Firebase project that you can use for testing:
- Project ID: `example-boardgame`
- Auth Domain: `example-boardgame.firebaseapp.com`

This is fine for local testing and won't affect production data.

**Option B: Create Your Own Firebase Project**

1. Go to [Firebase Console](https://console.firebase.google.com/)
2. Click "Add Project"
3. Follow the wizard to create a new project
4. Enable **Authentication** → **Email/Password** and/or **Google** sign-in
5. Copy your Firebase config values from Project Settings → General

### 2. Database Setup

**Option A: Use Bolt DB (Easiest - File-based, No Setup)**

Bolt DB is a simple file-based database perfect for development. No installation needed!

**Option B: MySQL (If you need it)**

If you want to use MySQL:
```bash
# Install MySQL
brew install mysql

# Start MySQL
brew services start mysql

# Create database
mysql -u root -e "CREATE DATABASE boardgame;"
```

### 3. Create Your Config File

Create `config.json` in `/Users/jkomoros/Code/boardgame/`:

```json
{
  "base": {
    "defaultStorageType": "bolt",
    "googleAnalytics": "",
    "firebase": {
      "apiKey": "AIzaSyDi0hhBgLPbpJgprVCDzDkk8zuFpb9XadM",
      "authDomain": "example-boardgame.firebaseapp.com",
      "databaseURL": "https://example-boardgame.firebaseio.com",
      "projectId": "example-boardgame",
      "storageBucket": "example-boardgame.appspot.com",
      "messagingSenderId": "138149526364"
    },
    "games": {
      "github.com/jkomoros/boardgame/examples": [
        "blackjack",
        "checkers",
        "debuganimations",
        "memory",
        "pig",
        "tictactoe"
      ]
    }
  },
  "dev": {
    "allowedOrigins": "http://localhost:8080",
    "disableAdminChecking": true,
    "storage": {
      "bolt": "./dev.db"
    }
  },
  "prod": {
    "allowedOrigins": "https://www.yourdomain.com",
    "adminUserIds": [],
    "storage": {
      "bolt": "./prod.db"
    }
  }
}
```

### 4. Add Local Games (Optional)

If you have custom games in a `games` directory at the root of the project:

```json
{
  "base": {
    // ... other config ...
    "games": {
      "github.com/jkomoros/boardgame/examples": [
        "blackjack",
        "checkers",
        "debuganimations",
        "memory",
        "pig",
        "tictactoe"
      ],
      "github.com/jkomoros/boardgame/games": [
        "yourgame1",
        "yourgame2"
      ]
    }
  }
}
```

**Note:** The games must be in a directory with the proper Go module structure. Each game should be a Go package within that directory.

## Quick Start with Bolt DB (Recommended)

Here's the fastest way to get started:

### Step 1: Create Minimal Config

```bash
cd /Users/jkomoros/Code/boardgame
cat > config.json << 'EOF'
{
  "base": {
    "defaultStorageType": "bolt",
    "firebase": {
      "apiKey": "AIzaSyDi0hhBgLPbpJgprVCDzDkk8zuFpb9XadM",
      "authDomain": "example-boardgame.firebaseapp.com",
      "databaseURL": "https://example-boardgame.firebaseio.com",
      "projectId": "example-boardgame",
      "storageBucket": "example-boardgame.appspot.com",
      "messagingSenderId": "138149526364"
    },
    "games": {
      "github.com/jkomoros/boardgame/examples": [
        "blackjack",
        "memory",
        "tictactoe"
      ]
    }
  },
  "dev": {
    "allowedOrigins": "http://localhost:8080",
    "disableAdminChecking": true,
    "storage": {
      "bolt": "./dev.db"
    }
  }
}
EOF
```

### Step 2: Start the Server

```bash
# Make sure you're in the boardgame root directory
cd /Users/jkomoros/Code/boardgame

# Start the server
boardgame-util serve
```

The server will:
- Build the API server (Go backend)
- Build the static frontend (from `server/static/`)
- Start both servers
- Frontend will be at `http://localhost:8080`
- API will be at `http://localhost:8888`

### Step 3: Create a Test User

Open `http://localhost:8080` in your browser and:

1. Click **"Sign In"**
2. Click **"Create an account"**
3. Enter email and password (e.g., `test@example.com` / `password123`)
4. Sign up!

Now you can test the full app with authentication!

## Config File Explained

### Base Section
```json
"base": {
  "defaultStorageType": "bolt",  // or "mysql"
  "firebase": { ... },            // Firebase auth config
  "games": { ... }                // Game packages to load
}
```

### Dev Section (for development)
```json
"dev": {
  "allowedOrigins": "http://localhost:8080",  // CORS setting
  "disableAdminChecking": true,               // Skip admin verification
  "storage": {
    "bolt": "./dev.db"                        // File-based database
  }
}
```

### Storage Options

**Bolt DB (Recommended for Dev):**
```json
"storage": {
  "bolt": "./dev.db"
}
```

**MySQL:**
```json
"storage": {
  "mysql": "username:password@tcp(localhost:3306)/boardgame"
}
```

## Client Config

The frontend also needs a config. Create/check `server/static/client_config.js`:

```javascript
var CONFIG = {
  "host": "http://localhost:8888/",
  "dev_host": "http://localhost:8888/",
  "offline_dev_mode": false,
  "google_analytics": ""
};
```

This should already exist and work with the default setup.

## Troubleshooting

### "Config not found" error
- Make sure `config.json` is in `/Users/jkomoros/Code/boardgame/` (project root)
- Or run `boardgame-util serve` from a directory containing the config

### Firebase auth errors
- The demo Firebase project should work for testing
- If it doesn't, create your own Firebase project (see above)
- Make sure Email/Password auth is enabled in Firebase Console

### Port already in use
```bash
# Kill process on port 8080 or 8888
lsof -ti:8080 | xargs kill -9
lsof -ti:8888 | xargs kill -9
```

### Database connection errors
- With Bolt DB, make sure the directory is writable
- With MySQL, make sure the service is running: `brew services list`

### Can't create users
- Check Firebase Console → Authentication → Sign-in method
- Make sure Email/Password is enabled
- Check browser console for specific errors

## Testing the Setup

Once the server is running:

1. **Frontend**: Open `http://localhost:8080`
2. **Create User**: Sign up with email/password
3. **Create Game**: Select a game type (e.g., "Memory") and click "Create Game"
4. **Play**: You should see the game interface
5. **Test Animations**: Make moves and watch the FLIP animations work!

## Development Workflow

### Making Changes

**Frontend changes:**
```bash
# In server/static/
npm run dev  # Vite dev server with hot reload
```

**Backend changes:**
```bash
# In project root
boardgame-util serve  # Rebuilds and restarts automatically
```

### Running Tests

```bash
# Go tests
go test ./...

# Frontend linting
cd server/static
npm run lint

# E2E tests (with Playwright)
npm run test:e2e
```

## Next Steps

- ✅ Create `config.json` with Bolt DB
- ✅ Run `boardgame-util serve`
- ✅ Create a test user
- ✅ Create and play a game
- ✅ Test the Material Design 3 UI
- ✅ Test animations with state changes
- 📝 Build your own game (see [TUTORIAL.md](TUTORIAL.md))

## Configuration Reference

Full config file template with all options:

```json
{
  "base": {
    "defaultStorageType": "bolt",
    "googleAnalytics": "UA-XXXXXXXXX-X",
    "firebase": {
      "apiKey": "YOUR_API_KEY",
      "authDomain": "your-project.firebaseapp.com",
      "databaseURL": "https://your-project.firebaseio.com",
      "projectId": "your-project",
      "storageBucket": "your-project.appspot.com",
      "messagingSenderId": "123456789"
    },
    "games": {
      "github.com/jkomoros/boardgame/examples": [
        "blackjack",
        "checkers",
        "debuganimations",
        "memory",
        "pig",
        "tictactoe"
      ],
      "github.com/jkomoros/boardgame/games": [
        "yourgame"
      ]
    }
  },
  "dev": {
    "allowedOrigins": "http://localhost:8080",
    "disableAdminChecking": true,
    "storage": {
      "bolt": "./dev.db"
    }
  },
  "prod": {
    "allowedOrigins": "https://www.yourdomain.com",
    "adminUserIds": [
      "firebase-user-id-1",
      "firebase-user-id-2"
    ],
    "storage": {
      "bolt": "./prod.db"
    }
  }
}
```

## Need Help?

- Check the [README.md](README.md) for more info
- See the [TUTORIAL.md](TUTORIAL.md) for a complete walkthrough
- File issues at the GitHub repo if you encounter problems
