# Redux Architecture Patterns

## Data Flow Principles

### Core Rule: Redux for State, Events for DOM

**Redux Actions**: For ALL application state changes
**Custom Events**: ONLY for DOM coordination and component communication

---

## Redux Patterns

### When to Use Redux Actions

✅ **Use Redux for**:
1. **Application State** - Data that persists across components
2. **Server Data** - Fetched game state, user info, game lists
3. **UI State** - View preferences, selected player, dialog open/closed
4. **Derived State** - Version tracking, animation queue, socket status

**Examples**:
```typescript
// Good: User interaction changes application state
private _handleRequestedPlayerChanged(e: CustomEvent) {
  store.dispatch(setRequestedPlayer(e.detail.value));
}

// Good: WebSocket updates application state
private _socketMessage(e: MessageEvent) {
  const version = parseInt(e.data);
  store.dispatch(setTargetVersion(version));
}

// Good: Fetch completion updates state
const response = await store.dispatch(fetchGameInfo(...));
if (response.data) {
  store.dispatch(updateViewState(...));
}
```

### Component Dispatch Guidelines

**Root/Container Components**: Can dispatch directly
- `boardgame-app.ts` - Navigation, error handling
- `boardgame-game-view.ts` - Game state installation
- `boardgame-game-state-manager.ts` - Fetch coordination

**View Components**: Dispatch in response to user actions
- `boardgame-user.ts` - Sign in/out
- `boardgame-create-game.ts` - Form updates, create game
- `boardgame-move-form.ts` - Submit move
- `boardgame-player-roster.ts` - Join game

**Presentation Components**: Should NOT dispatch
- Use events to bubble up to container
- Container dispatches the action

---

## Event Patterns

### When to Use Custom Events

✅ **Use Events for**:
1. **DOM Interactions** - User clicks, taps, hovers
2. **Component Communication** - Parent-child coordination
3. **Animation Lifecycle** - Animation start/end signals
4. **UI Commands** - Show dialog, refresh display

**Examples**:
```typescript
// Good: User interaction event (bubbles to parent)
this.dispatchEvent(new CustomEvent('component-tapped', {
  composed: true,
  detail: { index: this.index }
}));

// Good: Animation lifecycle coordination
this.dispatchEvent(new CustomEvent('will-animate', {
  composed: true,
  detail: { ele: this }
}));

// Good: Parent-child coordination
this.dispatchEvent(new CustomEvent('install-state-bundle', {
  composed: true,
  detail: bundle
}));

// Good: UI command (triggers dialog)
this.dispatchEvent(new CustomEvent('show-error', {
  composed: true,
  detail: { title, message, friendlyMessage }
}));
```

### Event Naming Conventions

- **Past tense** for things that happened: `component-tapped`, `animation-done`
- **Present tense** for commands: `show-error`, `refresh-info`
- **Descriptive** and specific: `requested-player-changed` not just `changed`

---

## Current Event Catalog

### User Interaction Events
- `component-tapped` - Component clicked by user
- `region-tapped` - Board region clicked
- `propose-move` - Renderer wants to propose a move

### Animation Events
- `will-animate` - Animation about to start
- `animation-done` - Single animation completed
- `all-animations-done` - All animations complete

### Coordination Events
- `install-state-bundle` - State manager → game view (install new state)
- `install-game-static-info` - State manager → game view (install static data)
- `set-animation-length` - State manager → renderer (set animation duration)
- `refresh-info` - Trigger → state manager (fetch new data)

### Form/UI Events
- `requested-player-changed` - Requested player value changed
- `auto-current-player-changed` - Auto-current-player toggled

### Dialog Events
- `show-error` - Show error dialog
- `show-login` - Show login dialog

---

## Redux State Structure

```typescript
state = {
  app: {
    // Routing
    page, pageExtra, location,
    // UI state
    offline, snackbarOpened, headerPanelOpen
  },

  user: {
    // Auth state
    loggedIn, admin, adminAllowed, user,
    // Sign-in dialog
    dialogOpen, dialogEmail, dialogPassword, dialogIsCreate
  },

  game: {
    // Game identity
    id, name,

    // Static info
    chest, playersInfo, hasEmptySlots, open, visible, isOwner,

    // Current state (RAW - use selectExpandedGameState)
    currentState, timerInfos, pathsToTick, originalWallClockTime,

    // Animation system
    animation: {
      pendingBundles, lastFiredBundle, activeAnimations
    },

    // Version tracking
    versions: {
      current, target, lastFetched
    },

    // WebSocket
    socket: {
      connected, connectionAttempts, lastError
    },

    // View state
    view: {
      game, viewingAsPlayer, requestedPlayer, autoCurrentPlayer, moveForms
    }
  },

  list: {
    // Game list
    managers, selectedManagerIndex, gameTypeFilter,
    allGames, participatingActiveGames, visibleActiveGames, etc.

    // Create game form
    numPlayers, agents, variantOptions, open, visible
  },

  error: {
    message, friendlyMessage, title, showing
  }
}
```

---

## Component Architecture

### Component Types

**1. Container Components** (connect to Redux)
- Read state via selectors in `stateChanged()`
- Dispatch actions in response to events/user input
- Coordinate child components
- Examples: `boardgame-game-view`, `boardgame-app`

**2. Smart Components** (connect to Redux)
- Read specific state slices
- Dispatch specific actions
- Handle their own logic
- Examples: `boardgame-user`, `boardgame-game-state-manager`

**3. Presentation Components** (no Redux)
- Receive all data via properties
- Fire events for user interactions
- Purely presentational
- Examples: `boardgame-component`, `boardgame-animatable-item`

### Component Connection Pattern

```typescript
import { connect } from 'pwa-helpers/connect-mixin.js';
import { store } from '../store.js';
import { selectFoo, selectBar } from '../selectors.js';

class MyComponent extends connect(store)(LitElement) {
  // Properties synced from Redux
  @property({ type: Object, attribute: false })
  foo: any = null;

  @property({ type: String, attribute: false })
  bar = '';

  // Sync from Redux
  stateChanged(state: RootState) {
    this.foo = selectFoo(state);
    this.bar = selectBar(state);
  }

  // Dispatch on user interaction
  private _handleClick() {
    store.dispatch(updateFoo(newValue));
  }
}
```

---

## Selector Patterns

### Memoized Selectors (reselect)

Use `createSelector` for derived/computed state:

```typescript
import { createSelector } from 'reselect';

// Base selectors (simple property access)
const selectGameCurrentState = (state) => state.game?.currentState;
const selectGameChest = (state) => state.game?.chest;

// Memoized selector (only recomputes when inputs change)
export const selectExpandedGameState = createSelector(
  [selectGameCurrentState, selectGameChest, selectGameName],
  (rawState, chest, gameName) => {
    // Expensive computation here
    return expandState(rawState, chest, gameName);
  }
);
```

**Benefits**:
- Only recompute when dependencies change
- Efficient for expensive operations (state expansion)
- Automatic memoization

### Simple Selectors

For direct property access, use plain functions:

```typescript
export const selectGame = (state: RootState) =>
  state.game?.view.game || null;

export const selectViewingAsPlayer = (state: RootState) =>
  state.game?.view.viewingAsPlayer || 0;
```

---

## Action Patterns

### Synchronous Actions

Simple state updates:

```typescript
export const SET_VIEWING_AS_PLAYER = 'SET_VIEWING_AS_PLAYER';

export const setViewingAsPlayer = (playerIndex: number) => {
  return {
    type: SET_VIEWING_AS_PLAYER,
    playerIndex
  };
};
```

### Async Actions (Thunks)

For API calls and complex logic:

```typescript
export const fetchGameInfo = (
  gameRoute: GameRoute,
  requestedPlayer: number,
  admin: boolean,
  lastFetchedVersion: number
) => async (dispatch: Dispatch): Promise<ApiResponse<any>> => {
  dispatch({ type: FETCH_GAME_INFO_REQUEST });

  const url = buildGameUrl(...);
  const response = await apiGet(url);

  if (response.error) {
    dispatch({
      type: FETCH_GAME_INFO_FAILURE,
      error: response.error
    });
  } else {
    dispatch({
      type: FETCH_GAME_INFO_SUCCESS,
      data: response.data
    });
  }

  return response;
};
```

---

## Reducer Patterns

### Immutable Updates

**Always** return new objects:

```typescript
case UPDATE_VIEW_STATE:
  return {
    ...state,  // Spread state
    view: {
      ...state.view,  // Spread nested object
      game: action.game,  // Update specific properties
      viewingAsPlayer: action.viewingAsPlayer,
      moveForms: action.moveForms
    }
  };
```

### Array Updates

```typescript
case ENQUEUE_STATE_BUNDLE:
  return {
    ...state,
    animation: {
      ...state.animation,
      pendingBundles: [...state.animation.pendingBundles, action.bundle]
    }
  };

case DEQUEUE_STATE_BUNDLE:
  const [first, ...rest] = state.animation.pendingBundles;
  return {
    ...state,
    animation: {
      ...state.animation,
      pendingBundles: rest,
      lastFiredBundle: first || state.animation.lastFiredBundle
    }
  };
```

---

## Anti-Patterns to Avoid

### ❌ Don't: Mutate State

```typescript
// BAD
stack.Components = components;
timer.TimeLeft = 0;

// GOOD
return {
  ...stack,
  Components: components
};
```

### ❌ Don't: Store Duplicate State

```typescript
// BAD: State in both Redux and component
@property({ type: Object })
game: any = null;  // Also in Redux!

// GOOD: Only in Redux, synced to component
@property({ type: Object, attribute: false })
game: any = null;  // Read-only, synced from Redux

stateChanged(state: RootState) {
  this.game = selectGame(state);
}
```

### ❌ Don't: Use Property Watchers for Side Effects

```typescript
// BAD: Automatic side effect
override updated(changedProps) {
  if (changedProps.has('targetVersion')) {
    this._fetchVersion();  // Hidden side effect!
  }
}

// GOOD: Explicit in stateChanged
stateChanged(state: RootState) {
  const prevTarget = this.targetVersion;
  this.targetVersion = selectTargetVersion(state);

  if (prevTarget !== this.targetVersion && this.targetVersion >= 0) {
    this._handleTargetVersionChanged();  // Explicit!
  }
}
```

### ❌ Don't: Store Derived State

```typescript
// BAD: Store both raw and expanded
currentState: rawState,
expandedState: expand(rawState)  // Duplication!

// GOOD: Store raw, expand via selector
currentState: rawState  // Only raw state
// Use selectExpandedGameState selector to get expanded
```

### ❌ Don't: Use Events for State Changes

```typescript
// BAD: Event changes application state
this.dispatchEvent(new CustomEvent('update-player', {
  detail: { playerIndex: 1 }
}));

// GOOD: Redux action for state, event for DOM coordination
store.dispatch(setViewingAsPlayer(1));  // State change
this.dispatchEvent(new CustomEvent('player-changed'));  // Notify
```

---

## Testing Patterns

### Selector Tests

```typescript
import { selectExpandedGameState } from './selectors';

test('expands state correctly', () => {
  const rawState = { ... };
  const chest = { ... };
  const state = { game: { currentState: rawState, chest } };

  const expanded = selectExpandedGameState(state);

  expect(expanded.Game.Stack.Components).toBeDefined();
});
```

### Reducer Tests

```typescript
import reducer from './reducers/game';
import { setViewingAsPlayer } from './actions/game';

test('sets viewing as player', () => {
  const state = { view: { viewingAsPlayer: 0 } };
  const action = setViewingAsPlayer(1);

  const newState = reducer(state, action);

  expect(newState.view.viewingAsPlayer).toBe(1);
  expect(newState).not.toBe(state);  // Immutability
});
```

### Mutation Detection

```typescript
test('does not mutate state', () => {
  const state = Object.freeze({ ... });
  const action = setViewingAsPlayer(1);

  expect(() => reducer(state, action)).not.toThrow();
});
```

---

## Migration Checklist

When adding new features:

1. **State**: Add to Redux if it needs to persist or be shared
2. **Actions**: Create action types and creators
3. **Reducer**: Handle actions immutably
4. **Selectors**: Create selectors for derived data
5. **Components**: Connect via `stateChanged()`, dispatch actions
6. **Events**: Use only for DOM coordination
7. **Tests**: Add selector and reducer tests

---

## Future Considerations

### Redux Toolkit (RTK)

After architecture stabilizes (6+ months):

**Benefits**:
- Less boilerplate (`createSlice`)
- Built-in immutability (Immer)
- Better TypeScript support
- `createAsyncThunk` for async actions

**Migration Path**:
1. Current architecture is RTK-compatible
2. Can migrate slice-by-slice
3. Keep lazy reducer loading (pwa-helpers)

### RTK Query

For server state caching:
- Automatic request deduplication
- Cache invalidation
- Optimistic updates
- WebSocket integration

---

## Loading and Error State Patterns

### REQUEST/SUCCESS/FAILURE Action Pattern

For async operations (API calls, thunks), use a three-action pattern:

```typescript
// Action types
export const FETCH_GAME_INFO_REQUEST = 'FETCH_GAME_INFO_REQUEST';
export const FETCH_GAME_INFO_SUCCESS = 'FETCH_GAME_INFO_SUCCESS';
export const FETCH_GAME_INFO_FAILURE = 'FETCH_GAME_INFO_FAILURE';

// Thunk action creator
export const fetchGameInfo = (params) => async (dispatch: Dispatch) => {
  // Step 1: Dispatch REQUEST (sets loading=true)
  dispatch({ type: FETCH_GAME_INFO_REQUEST });

  try {
    const response = await apiGet(url);

    if (response.error) {
      // Step 2a: Dispatch FAILURE (sets loading=false, error)
      dispatch({
        type: FETCH_GAME_INFO_FAILURE,
        error: response.error
      });
    } else {
      // Step 2b: Dispatch SUCCESS (sets loading=false, data)
      dispatch({
        type: FETCH_GAME_INFO_SUCCESS,
        data: response.data
      });
    }

    return response;
  } catch (error) {
    dispatch({
      type: FETCH_GAME_INFO_FAILURE,
      error: error.message
    });
    throw error;
  }
};
```

### Reducer Handling

```typescript
case FETCH_GAME_INFO_REQUEST:
  return {
    ...state,
    loading: true,
    error: null
  };

case FETCH_GAME_INFO_SUCCESS:
  return {
    ...state,
    loading: false,
    error: null,
    // Update data from response
    ...action.data
  };

case FETCH_GAME_INFO_FAILURE:
  return {
    ...state,
    loading: false,
    error: action.error
  };
```

### Reading Loading/Error State in Components

```typescript
import { selectGameLoading, selectGameError } from '../selectors.js';

class MyComponent extends connect(store)(LitElement) {
  @property({ type: Boolean, attribute: false })
  loading = false;

  @property({ type: String, attribute: false })
  error: string | null = null;

  stateChanged(state: RootState) {
    this.loading = selectGameLoading(state);
    this.error = selectGameError(state);
  }

  render() {
    if (this.loading) {
      return html`<div class="spinner">Loading...</div>`;
    }

    if (this.error) {
      return html`<div class="error">Error: ${this.error}</div>`;
    }

    return html`<div>Content here</div>`;
  }
}
```

### Best Practices

1. **Always dispatch REQUEST first** - Sets loading state immediately
2. **Always dispatch SUCCESS or FAILURE** - Clears loading state
3. **Store errors in Redux** - Makes them debuggable in DevTools
4. **Use selectors for loading/error** - Consistent access pattern
5. **Show loading UI** - Better user experience during async operations

---

## Thunk Best Practices

### Core Principle: Thunks for Side Effects, Not Data

Thunks should coordinate async operations and dispatch actions. They should NOT be used to return data to components.

### ✅ Correct Pattern: Data Flows Through Redux

```typescript
// Thunk action (returns Promise<void> or ApiResponse)
export const fetchGameInfo = (gameId: string) =>
  async (dispatch: Dispatch): Promise<void> => {
    dispatch({ type: FETCH_GAME_INFO_REQUEST });

    const response = await apiGet(`/api/game/${gameId}`);

    if (response.error) {
      dispatch({ type: FETCH_GAME_INFO_FAILURE, error: response.error });
    } else {
      // Store data in Redux state
      dispatch({
        type: FETCH_GAME_INFO_SUCCESS,
        data: response.data
      });
    }
  };

// Component dispatches and reads from Redux
class MyComponent extends connect(store)(LitElement) {
  @property({ type: Object, attribute: false })
  gameInfo: any = null;

  stateChanged(state: RootState) {
    // Read data from Redux state
    this.gameInfo = selectGameInfo(state);
  }

  async _loadGame() {
    // Dispatch thunk (don't use return value)
    await store.dispatch(fetchGameInfo(this.gameId));

    // Data is already in Redux state
    // this.gameInfo will update via stateChanged()
  }
}
```

### ❌ Incorrect Pattern: Returning Data

```typescript
// BAD: Thunk returns data directly
export const fetchGameInfo = (gameId: string) =>
  async (dispatch: Dispatch): Promise<GameInfo> => {
    const response = await apiGet(`/api/game/${gameId}`);
    return response.data;  // ❌ Anti-pattern!
  };

// BAD: Component uses return value instead of Redux
class MyComponent extends connect(store)(LitElement) {
  async _loadGame() {
    // ❌ Using return value instead of Redux state
    const gameInfo = await store.dispatch(fetchGameInfo(this.gameId));
    this.gameInfo = gameInfo;  // ❌ Bypasses Redux!
  }
}
```

### Why This Matters

1. **Redux DevTools** - Data in Redux is visible and debuggable
2. **Time-travel** - Only works if data flows through Redux
3. **Component coordination** - Multiple components can read same data
4. **Testability** - Selectors are easier to test than return values
5. **Consistency** - Single source of truth

### Thunk Return Values

**When to return values**:
- `ApiResponse` objects for error handling
- `Promise<void>` for simple operations
- Never return actual data

```typescript
// Good: Return ApiResponse for error checking
export const submitMove = (move: any) =>
  async (dispatch: Dispatch): Promise<ApiResponse<void>> => {
    dispatch({ type: SUBMIT_MOVE_REQUEST });

    const response = await apiPost('/api/move', move);

    if (response.error) {
      dispatch({ type: SUBMIT_MOVE_FAILURE, error: response.error });
    } else {
      dispatch({ type: SUBMIT_MOVE_SUCCESS });
    }

    return response;  // For error handling, not data
  };

// Component can check for errors
async _handleSubmit() {
  const response = await store.dispatch(submitMove(this.move));

  if (response.error) {
    // Handle error
    console.error('Move failed:', response.error);
  }

  // Success state is in Redux
}
```

---

## Selector Performance Patterns

### Stable Default Objects

When returning default objects/arrays, use stable references to prevent unnecessary re-renders:

```typescript
// ❌ BAD: Creates new object every time
export const selectGameInfo = (state: RootState) =>
  state.game?.info || {};  // New object each call!

// ✅ GOOD: Stable default object
const EMPTY_OBJECT = {};
export const selectGameInfo = (state: RootState) =>
  state.game?.info || EMPTY_OBJECT;  // Same object each call

// ❌ BAD: Creates new array every time
export const selectPlayers = (state: RootState) =>
  state.game?.players || [];  // New array each call!

// ✅ GOOD: Stable default array
const EMPTY_ARRAY: any[] = [];
export const selectPlayers = (state: RootState) =>
  state.game?.players || EMPTY_ARRAY;  // Same array each call
```

### Why Stable Defaults Matter

Lit components use shallow equality checks. New objects cause unnecessary re-renders:

```typescript
// With unstable defaults
const obj1 = selectGameInfo(state);  // Returns {}
const obj2 = selectGameInfo(state);  // Returns {} (different reference!)
obj1 === obj2  // false → component re-renders

// With stable defaults
const obj1 = selectGameInfo(state);  // Returns EMPTY_OBJECT
const obj2 = selectGameInfo(state);  // Returns EMPTY_OBJECT (same reference!)
obj1 === obj2  // true → no unnecessary re-render
```

### Memoize Frequently-Used Selectors

For selectors called often (>10Hz), use memoization:

```typescript
import { createSelector } from 'reselect';

// Base selectors (not memoized, simple property access)
const selectGameCurrentState = (state: RootState) =>
  state.game?.currentState;

const selectGameChest = (state: RootState) =>
  state.game?.chest;

// Memoized selector (only recomputes when inputs change)
export const selectExpandedGameState = createSelector(
  [selectGameCurrentState, selectGameChest, selectGameName],
  (rawState, chest, gameName) => {
    // Expensive computation
    return expandState(rawState, chest, gameName);
  }
);
```

**When to memoize**:
- Expensive computations (state expansion, filtering, sorting)
- Selectors called frequently (animation loops, timers)
- Derived data that combines multiple state slices

**When NOT to memoize**:
- Simple property access (`state.game?.id`)
- Rarely called selectors
- Selectors that always return new data

### Timer Expansion Performance

Special case: Timers update at 60+ Hz. Use separate selector to avoid re-expanding entire game state:

```typescript
// ✅ GOOD: Separate timer expansion (60+ Hz)
export const selectTimerExpandedGameState = createSelector(
  [selectExpandedGameState, selectTimerInfos, selectPathsToTick, selectOriginalWallClockTime],
  (expandedState, timerInfos, pathsToTick, wallClockTime) => {
    if (!expandedState || !timerInfos) return expandedState;

    // Only expand timers, not entire state
    return expandTimersInState(expandedState, timerInfos, pathsToTick, wallClockTime);
  }
);

// ❌ BAD: Would re-expand full state 60+ times per second
// Just use selectExpandedGameState directly
```

**When to use each**:
- `selectExpandedGameState` - For rendering game state (boards, components, etc.)
- `selectTimerExpandedGameState` - For displaying timer values that tick

### Performance Checklist

- [ ] Use stable default objects/arrays for fallbacks
- [ ] Memoize selectors with expensive computations
- [ ] Separate high-frequency updates (timers) from low-frequency (game state)
- [ ] Profile with Redux DevTools to identify selector hotspots
- [ ] Use `createSelector` from reselect for derived data

---

## Single Source of Truth Pattern

### Core Principle: Don't Duplicate Redux State

State should live in ONE place: Redux. Components should read from Redux, not maintain their own copies.

### ❌ Anti-Pattern: Duplicated State

```typescript
// BAD: State exists in both Redux and component
class BoardgameGameView extends connect(store)(LitElement) {
  // ❌ This duplicates Redux state
  @property({ type: Object })
  game: any = null;

  @property({ type: Number })
  viewingAsPlayer = 0;

  stateChanged(state: RootState) {
    // Syncing Redux → component property
    this.game = selectGame(state);
    this.viewingAsPlayer = selectViewingAsPlayer(state);
  }

  _handlePlayerChange(newPlayer: number) {
    // ❌ Updating both component and Redux
    this.viewingAsPlayer = newPlayer;  // Duplication!
    store.dispatch(setViewingAsPlayer(newPlayer));
  }
}
```

**Problems**:
- State can get out of sync
- Unclear which is source of truth
- Harder to debug (which value is correct?)
- More code to maintain

### ✅ Correct Pattern: Component Properties as Read-Only Cache

```typescript
// GOOD: Component properties are read-only cache of Redux state
class BoardgameGameView extends connect(store)(LitElement) {
  // ✅ Read-only cache, synced from Redux
  @property({ type: Object, attribute: false })
  game: any = null;

  @property({ type: Number, attribute: false })
  viewingAsPlayer = 0;

  // Only sync FROM Redux
  stateChanged(state: RootState) {
    this.game = selectGame(state);
    this.viewingAsPlayer = selectViewingAsPlayer(state);
  }

  // Only write TO Redux
  _handlePlayerChange(newPlayer: number) {
    // ✅ Only update Redux (component updates via stateChanged)
    store.dispatch(setViewingAsPlayer(newPlayer));
  }
}
```

### Computed Getters with Selectors

For derived data, use computed getters instead of storing in component state:

```typescript
class BoardgameGameView extends connect(store)(LitElement) {
  // Store minimal state
  @property({ type: Object, attribute: false })
  rawGameState: any = null;

  @property({ type: Object, attribute: false })
  chest: any = null;

  // ✅ Computed getter (no stored state)
  get expandedGameState() {
    if (!this.rawGameState || !this.chest) return null;
    return expandState(this.rawGameState, this.chest);
  }

  stateChanged(state: RootState) {
    this.rawGameState = selectGameCurrentState(state);
    this.chest = selectGameChest(state);
    // expandedGameState automatically computed
  }

  render() {
    const state = this.expandedGameState;
    if (!state) return html`<div>Loading...</div>`;
    return html`<div>Render ${state.Game.Name}</div>`;
  }
}
```

### Benefits of Single Source of Truth

1. **No sync bugs** - Can't get out of sync if there's only one copy
2. **Easier debugging** - Redux DevTools shows the truth
3. **Time-travel works** - State changes replay correctly
4. **Clearer data flow** - Always know where state comes from
5. **Less code** - Don't need sync logic

### Migration Pattern

Converting duplicated state to single source of truth:

```typescript
// Before: Duplicated state
@property({ type: String })
selectedTab = 'overview';  // ❌ Component owns this

_handleTabClick(tab: string) {
  this.selectedTab = tab;  // ❌ Direct update
}

// After: Redux owns state
@property({ type: String, attribute: false })
selectedTab = 'overview';  // ✅ Read-only cache

stateChanged(state: RootState) {
  this.selectedTab = selectSelectedTab(state);  // ✅ Sync from Redux
}

_handleTabClick(tab: string) {
  store.dispatch(setSelectedTab(tab));  // ✅ Update Redux
  // Component updates via stateChanged()
}
```

---

## Unidirectional Data Flow Pattern

### Core Principle: No Imperative Child Component Method Calls

Data should flow DOWN (props), events should flow UP (events). Never call methods on child components to change their state.

### ❌ Anti-Pattern: Imperative Method Calls

```typescript
// BAD: Parent calling child methods directly
class ParentComponent extends LitElement {
  @query('child-component')
  child!: ChildComponent;

  _handleSomething() {
    // ❌ Imperatively calling child method
    this.child.updateState(newData);
    this.child.refresh();
  }
}

class ChildComponent extends LitElement {
  // ❌ Public method for parent to call
  updateState(data: any) {
    this.data = data;
    this.requestUpdate();
  }
}
```

**Problems**:
- Breaks unidirectional data flow
- Hard to track state changes
- Not Redux DevTools visible
- Tight coupling between components
- Can't time-travel debug

### ✅ Correct Pattern: Redux Actions for State, Events for Coordination

```typescript
// GOOD: Parent dispatches Redux actions
class ParentComponent extends connect(store)(LitElement) {
  _handleSomething() {
    // ✅ Update Redux state
    store.dispatch(updateChildData(newData));

    // ✅ If child needs to do something, use custom event
    this.dispatchEvent(new CustomEvent('refresh-requested', {
      composed: true
    }));
  }
}

// GOOD: Child reads from Redux and listens to events
class ChildComponent extends connect(store)(LitElement) {
  @property({ type: Object, attribute: false })
  data: any = null;

  connectedCallback() {
    super.connectedCallback();
    // ✅ Listen for coordination events
    this.addEventListener('refresh-requested', this._handleRefresh);
  }

  stateChanged(state: RootState) {
    // ✅ Read state from Redux
    this.data = selectChildData(state);
  }

  private _handleRefresh = () => {
    // ✅ Dispatch action if needed
    store.dispatch(refreshData());
  };
}
```

### Use Custom Events for DOM Coordination

For UI coordination that doesn't need to be in Redux:

```typescript
// Parent wants child to scroll to top
class ParentComponent extends LitElement {
  _handleResetView() {
    // ✅ Dispatch custom event for DOM operation
    this.dispatchEvent(new CustomEvent('scroll-to-top', {
      composed: true,
      bubbles: true
    }));
  }
}

class ChildComponent extends LitElement {
  connectedCallback() {
    super.connectedCallback();
    this.addEventListener('scroll-to-top', this._handleScrollToTop);
  }

  private _handleScrollToTop = () => {
    // Pure DOM operation (no state change)
    this.scrollTop = 0;
  };
}
```

### When to Use Each Pattern

**Redux Actions** (state changes):
- Update application data
- Change UI state (selected tab, viewing player)
- Trigger async operations (fetch data)
- Anything that should be in Redux DevTools

**Custom Events** (DOM coordination):
- Scroll to position
- Focus an input
- Trigger animation
- Show/hide DOM elements
- Anything that's purely UI coordination

### Real Example: Animation System

```typescript
// State Manager dispatches action to add animation to queue
class StateManager extends connect(store)(LitElement) {
  _handleNewStateBundle(bundle: StateBundle) {
    // ✅ Add to Redux animation queue
    store.dispatch(enqueueStateBundle(bundle));
  }
}

// Animator reads queue from Redux
class Animator extends connect(store)(LitElement) {
  @property({ type: Array, attribute: false })
  pendingBundles: StateBundle[] = [];

  stateChanged(state: RootState) {
    // ✅ Read animation queue from Redux
    this.pendingBundles = selectPendingStateBundles(state);
  }

  async _processNextAnimation() {
    if (this.pendingBundles.length === 0) return;

    // ✅ Remove from queue via action
    store.dispatch(dequeueStateBundle());

    // Perform animation
    await this._animate();

    // ✅ Use event for DOM coordination
    this.dispatchEvent(new CustomEvent('animation-done', {
      composed: true
    }));
  }
}
```

### Benefits of Unidirectional Flow

1. **Predictable** - Data always flows same direction
2. **Debuggable** - State changes visible in Redux DevTools
3. **Testable** - Can test components in isolation
4. **Loosely coupled** - Components don't depend on each other's APIs
5. **Time-travel works** - All state changes go through Redux

---

## Summary

**Redux Architecture** = Pure, Predictable, Debuggable

✅ Single source of truth (Redux state)
✅ Immutable updates (always new objects)
✅ Explicit actions (no hidden side effects)
✅ Memoized selectors (efficient derived data)
✅ Clear boundaries (Redux for state, events for DOM)
✅ Time-travel debugging (Redux DevTools)
✅ Loading/error state infrastructure
✅ Thunks for side effects, not data return
✅ Stable default objects prevent re-renders
✅ No state duplication in components
✅ Unidirectional data flow (no imperative calls)

**Result**: Maintainable, testable, understandable codebase.
