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

## Summary

**Redux Architecture** = Pure, Predictable, Debuggable

✅ Single source of truth (Redux state)
✅ Immutable updates (always new objects)
✅ Explicit actions (no hidden side effects)
✅ Memoized selectors (efficient derived data)
✅ Clear boundaries (Redux for state, events for DOM)
✅ Time-travel debugging (Redux DevTools)

**Result**: Maintainable, testable, understandable codebase.
