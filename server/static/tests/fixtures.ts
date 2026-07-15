import { Page } from '@playwright/test';

/**
 * Test fixtures and helper functions for Playwright tests
 */

/**
 * Expose the Redux store to tests via window.__TEST_STORE__
 * This allows tests to inspect application state
 */
export async function exposeStore(page: Page, timeout = 10000): Promise<void> {
  // Wait for the app element to be available (indicating the app is loaded)
  await page.waitForSelector('boardgame-app', { timeout });

  // Import and expose the store module
  await page.evaluate(async () => {
    try {
      // Dynamically import the store module
      const storeModule = await import('/src/store.ts');
      (window as any).__TEST_STORE__ = storeModule.store;
    } catch (error) {
      console.error('Failed to import store:', error);
    }
  });

  // Verify the store was successfully exposed
  await page.waitForFunction(
    () => {
      return (window as any).__TEST_STORE__ !== undefined;
    },
    { timeout: 5000 }
  );
}

/**
 * Get the current Redux store state
 * Must call exposeStore() first
 */
export async function getStoreState(page: Page): Promise<any> {
  return await page.evaluate(() => {
    const store = (window as any).__TEST_STORE__;
    return store ? store.getState() : null;
  });
}

/**
 * Wait for the animation queue to drain
 * This ensures all animations have completed before continuing
 */
export async function waitForAnimationQueue(page: Page, timeout = 5000): Promise<void> {
  await page.waitForFunction(
    () => {
      const store = (window as any).__TEST_STORE__;
      if (!store) return false;

      const state = store.getState();
      const bundleCount = state?.app?.pendingBundleCount || 0;

      return bundleCount === 0;
    },
    { timeout }
  );
}

/**
 * Get the current pending bundle count (animations in progress)
 */
export async function getPendingBundleCount(page: Page): Promise<number> {
  return await page.evaluate(() => {
    const store = (window as any).__TEST_STORE__;
    if (!store) return 0;

    const state = store.getState();
    return state?.app?.pendingBundleCount || 0;
  });
}

/**
 * Login helper - waits for auth state to be loaded
 * Note: This doesn't perform actual login, just waits for auth initialization
 */
export async function waitForAuth(page: Page, timeout = 10000): Promise<void> {
  await page.waitForFunction(
    () => {
      const store = (window as any).__TEST_STORE__;
      if (!store) return false;

      const state = store.getState();
      // Check if user state is initialized (could be null for anonymous)
      return state?.user !== undefined;
    },
    { timeout }
  );
}

/**
 * Get list of active games from API
 */
export async function getActiveGames(page: Page): Promise<any[]> {
  try {
    await authenticateOfflineTestUser(page);
    const response = await page.request.get('/api/list/game');
    if (!response.ok()) {
      console.error(`Failed to fetch games: ${response.status()} ${response.statusText()}`);
      return [];
    }
    const data = await response.json();
    // API returns different arrays of games
    const participating = data.ParticipatingActiveGames || [];
    const visible = data.VisibleActiveGames || [];
    const joinable = data.VisibleJoinableActiveGames || [];

    // Combine all game lists
    return [...participating, ...visible, ...joinable];
  } catch (error) {
    console.error('Error fetching active games:', error);
    return [];
  }
}

/**
 * Create a new game via the API
 * Returns the game ID
 */
export async function createGame(
  page: Page,
  gameName: string,
  numPlayers: number = 2
): Promise<string> {
  try {
    // Offline-dev mode accepts the stable fake identity and returns the same
    // auth cookie used by page requests. Authenticate explicitly instead of
    // falling back to a fake game ID when the harness starts with a clean
    // browser context.
    await authenticateOfflineTestUser(page);

    const response = await page.request.post('/api/new/game', {
      form: {
        manager: gameName,
        numplayers: String(numPlayers),
      },
    });

    if (!response.ok()) {
      const text = await response.text();
      console.error(`Failed to create game: ${response.status()} ${response.statusText()}`, text);
      throw new Error(`Failed to create game: ${response.status()}`);
    }

    const data = await response.json();
    console.log('Created game response:', data);
    const gameId = data.GameID || data.ID || data.gameId || data.id;

    if (!gameId) {
      console.error('No game ID in response:', data);
      throw new Error('No game ID returned from API');
    }

    return gameId;
  } catch (error) {
    console.error('Error creating game:', error);
    throw error;
  }
}

async function authenticateOfflineTestUser(page: Page): Promise<void> {
  const authResponse = await page.request.post('/api/auth', {
    form: {
      uid: 'renderer-test@example.com',
      token: 'offline-test-token',
      email: 'renderer-test@example.com',
      displayname: 'Renderer Test',
    },
  });
  const auth = await authResponse.json();
  if (!authResponse.ok() || auth.Status !== 'Success') {
    throw new Error(`Offline test authentication failed: ${JSON.stringify(auth)}`);
  }
}

/**
 * Find or create a game of the specified type
 * Returns the game ID
 */
export async function getOrCreateGame(
  page: Page,
  gameName: string
): Promise<string> {
  // First, check if there's an existing game
  const games = await getActiveGames(page);
  console.log(`Found ${games.length} existing games`);

  const existingGame = games.find((g: any) => g.Name === gameName);

  if (existingGame) {
    console.log(`Using existing ${gameName} game: ${existingGame.ID}`);
    return existingGame.ID;
  }

  // If not, try to create a new game (may require authentication)
  console.log(`No existing ${gameName} game found, attempting to create one`);
  return await createGame(page, gameName);
}

/**
 * Navigate to a game by name, finding or creating it first
 * This uses the proper URL format: /game/{gameName}/{gameId}
 *
 * Fails loudly if no real game can be found or created. A placeholder ID makes
 * renderer-presence assertions pass while silently skipping actual game state.
 */
export async function navigateToGameByName(
  page: Page,
  gameName: string
): Promise<void> {
  // Each renderer test gets a known initial state. Reusing a visible game is
  // an explicit helper capability, not the default navigation behavior.
  const gameId = await createGame(page, gameName);

  const url = `/game/${gameName}/${gameId}`;
  console.log(`Navigating to game: ${url}`);
  await page.goto(url);
  await page.waitForLoadState('networkidle');
}

/**
 * Navigate to a game by name (deprecated - use navigateToGameByName instead)
 * @deprecated This function uses an invalid URL format. Use navigateToGameByName instead.
 */
export async function navigateToGame(page: Page, gameName: string): Promise<void> {
  await page.goto(`/game/${gameName}`);
  await page.waitForLoadState('networkidle');
}

/**
 * Wait for a custom element to be defined and upgraded
 */
export async function waitForCustomElement(page: Page, tagName: string, timeout = 5000): Promise<void> {
  await page.waitForFunction(
    (tag) => {
      const element = document.querySelector(tag);
      return element && customElements.get(tag);
    },
    tagName,
    { timeout }
  );
}

/**
 * Take a screenshot with a descriptive name
 */
export async function takeScreenshot(page: Page, name: string): Promise<void> {
  await page.screenshot({ path: `screenshots/${name}.png`, fullPage: true });
}
