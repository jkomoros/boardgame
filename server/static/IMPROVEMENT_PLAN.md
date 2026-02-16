# Concrete Improvement Plan

## Critical Fix: Implement Proper Redux Pattern

### Problem Statement
The migration replaced `boardgame-ajax` with Redux thunks, but components still handle API responses directly instead of using Redux state. This defeats the purpose of Redux.

### Current (Anti-Pattern)
```typescript
// Component
const response = await store.dispatch(configureGame(...));
if (response.error) {
  this.dispatchEvent(new CustomEvent("show-error", {...}));
}
```

**Issues**:
- Loading/error actions dispatched but never consumed
- Redux state never updates
- Can't use Redux DevTools
- Difficult to test

### Target (Proper Redux)
```typescript
// Component
store.dispatch(configureGame(...));
// No direct response handling

// Render method uses selectors
render() {
  if (this.loading) return html`<md-circular-progress>`;
  if (this.error) return html`<error-message .error="${this.error}">`;
  // ... normal UI
}

// State updates come from Redux
stateChanged(state: RootState) {
  this.loading = selectGameLoading(state);
  this.error = selectGameError(state);
}
```

---

## Step-by-Step Implementation

### Step 1: Add Reducer Cases

**File**: `src/reducers/game.js`

```javascript
// Add to INITIAL_STATE
const INITIAL_STATE = {
  id: '',
  name: '',
  chest: null,
  playersInfo: [],
  hasEmptySlots: false,
  open: false,
  visible: false,
  isOwner: false,
  currentState: null,
  pathsToTick: [],
  originalWallClockTime: 0,
  loading: false,  // ADD
  error: null,     // ADD
};

// Add to reducer switch statement
import {
  UPDATE_GAME_ROUTE,
  UPDATE_GAME_STATIC_INFO,
  UPDATE_GAME_CURRENT_STATE,
  CONFIGURE_GAME_REQUEST,
  CONFIGURE_GAME_SUCCESS,
  CONFIGURE_GAME_FAILURE,
  JOIN_GAME_REQUEST,
  JOIN_GAME_SUCCESS,
  JOIN_GAME_FAILURE,
  SUBMIT_MOVE_REQUEST,
  SUBMIT_MOVE_SUCCESS,
  SUBMIT_MOVE_FAILURE,
  FETCH_GAME_INFO_REQUEST,
  FETCH_GAME_INFO_SUCCESS,
  FETCH_GAME_INFO_FAILURE,
  FETCH_GAME_VERSION_REQUEST,
  FETCH_GAME_VERSION_SUCCESS,
  FETCH_GAME_VERSION_FAILURE
} from '../actions/game.js';

// In the reducer switch:
case CONFIGURE_GAME_REQUEST:
case JOIN_GAME_REQUEST:
case SUBMIT_MOVE_REQUEST:
case FETCH_GAME_INFO_REQUEST:
case FETCH_GAME_VERSION_REQUEST:
  return {
    ...state,
    loading: true,
    error: null
  };

case CONFIGURE_GAME_SUCCESS:
case JOIN_GAME_SUCCESS:
case SUBMIT_MOVE_SUCCESS:
  return {
    ...state,
    loading: false,
    error: null
  };

case CONFIGURE_GAME_FAILURE:
case JOIN_GAME_FAILURE:
case SUBMIT_MOVE_FAILURE:
  return {
    ...state,
    loading: false,
    error: action.friendlyError || action.error || 'An error occurred'
  };

case FETCH_GAME_INFO_SUCCESS:
  return {
    ...state,
    loading: false,
    error: null,
    chest: action.chest,
    playersInfo: action.playersInfo,
    hasEmptySlots: action.hasEmptySlots,
    open: action.open,
    visible: action.visible,
    isOwner: action.isOwner
  };

case FETCH_GAME_INFO_FAILURE:
case FETCH_GAME_VERSION_FAILURE:
  return {
    ...state,
    loading: false,
    error: action.friendlyError || action.error || 'Failed to load game data'
  };

case FETCH_GAME_VERSION_SUCCESS:
  // Version bundles are handled by state manager component
  // Just clear loading state
  return {
    ...state,
    loading: false,
    error: null
  };
```

---

### Step 2: Update Components to Use Redux State

**File**: `src/components/boardgame-configure-game-properties.ts`

```typescript
// BEFORE
private async _submit(open: boolean, visible: boolean): Promise<void> {
  if (!this.gameRoute) return;

  const response = await store.dispatch(
    configureGame(this.gameRoute, open, visible, this.admin)
  );

  if (response.error) {
    this.dispatchEvent(new CustomEvent("show-error", {
      composed: true,
      detail: {
        message: response.error,
        friendlyMessage: response.friendlyError,
        title: "Couldn't toggle"
      }
    }));
  } else {
    this.dispatchEvent(new CustomEvent("refresh-info", { composed: true }));
  }
}

// AFTER
@property({ type: Boolean })
private _loading = false;

@property({ type: String })
private _error: string | null = null;

stateChanged(state: RootState): void {
  this._loading = selectGameLoading(state);
  this._error = selectGameError(state);
}

override updated(changedProperties: Map<string, unknown>): void {
  if (changedProperties.has('_error') && this._error) {
    // Show error via event or built-in UI
    this.dispatchEvent(new CustomEvent("show-error", {
      composed: true,
      detail: {
        message: this._error,
        title: "Couldn't toggle"
      }
    }));
  }

  // Clear error after showing
  if (this._error && changedProperties.get('_error') !== this._error) {
    // Error changed, could auto-clear after timeout
  }
}

private _submit(open: boolean, visible: boolean): void {
  if (!this.gameRoute) return;

  // Just dispatch - state updates will trigger re-render
  store.dispatch(configureGame(this.gameRoute, open, visible, this.admin));

  // Note: No longer dispatching refresh-info event
  // Parent should subscribe to Redux state changes
}

render() {
  return html`
    ${this._loading ? html`<md-circular-progress indeterminate>` : ''}
    <md-icon-button
      ?disabled="${this.disabled || this._loading}"
      @click="${this._handleOpenTapped}"
      title="${this._openAlt(this.gameOpen)}">
      <md-icon>${this._openIcon(this.gameOpen)}</md-icon>
    </md-icon-button>
    <!-- ... -->
  `;
}
```

---

### Step 3: Update Thunks to NOT Return Responses

**File**: `src/actions/game.ts`

```typescript
// BEFORE
export const configureGame = (
  gameRoute: GameRoute,
  open: boolean,
  visible: boolean,
  admin: boolean
) => async (dispatch: Dispatch): Promise<ApiResponse<any>> => {
  dispatch({ type: CONFIGURE_GAME_REQUEST });
  const response = await apiPost(url, {...});
  if (response.error) {
    dispatch({ type: CONFIGURE_GAME_FAILURE, error: response.error });
    return response; // ❌ Component uses this
  }
  dispatch({ type: CONFIGURE_GAME_SUCCESS });
  return response;
};

// AFTER
export const configureGame = (
  gameRoute: GameRoute,
  open: boolean,
  visible: boolean,
  admin: boolean
) => async (dispatch: Dispatch): Promise<void> => {  // Return void
  dispatch({ type: CONFIGURE_GAME_REQUEST });

  const url = buildGameUrl(gameRoute.name, gameRoute.id, 'configure');
  const response = await apiPost(url, {
    open: open ? 1 : 0,
    visible: visible ? 1 : 0,
    admin: admin ? 1 : 0
  }, 'application/x-www-form-urlencoded');

  if (response.error) {
    dispatch({
      type: CONFIGURE_GAME_FAILURE,
      error: response.error,
      friendlyError: response.friendlyError
    });
  } else {
    dispatch({ type: CONFIGURE_GAME_SUCCESS });
    // Could dispatch refresh action here if needed
    dispatch(fetchGameInfo(gameRoute, ...)); // Refresh data
  }
};
```

---

### Step 4: Handle Success Side Effects

**Option A: Thunk Composition**
```typescript
export const configureGameAndRefresh = (
  gameRoute: GameRoute,
  open: boolean,
  visible: boolean,
  admin: boolean,
  requestedPlayer: number,
  lastFetchedVersion: number
) => async (dispatch: Dispatch): Promise<void> => {
  await dispatch(configureGame(gameRoute, open, visible, admin));

  // Check if successful (would need to look at state)
  // Or just always refresh
  await dispatch(fetchGameInfo(gameRoute, requestedPlayer, false, lastFetchedVersion));
};
```

**Option B: Redux Middleware**
```typescript
// middleware/refreshMiddleware.js
const refreshMiddleware = store => next => action => {
  const result = next(action);

  // After successful mutations, trigger refresh
  if (action.type === CONFIGURE_GAME_SUCCESS ||
      action.type === JOIN_GAME_SUCCESS ||
      action.type === SUBMIT_MOVE_SUCCESS) {
    // Dispatch refresh action
    const state = store.getState();
    const gameRoute = selectGameRoute(state);
    if (gameRoute) {
      store.dispatch(fetchGameInfo(gameRoute, ...));
    }
  }

  return result;
};
```

**Option C: Event Bus (Current Pattern - Keep It)**
```typescript
// Component still dispatches events, parent listens
this.dispatchEvent(new CustomEvent("refresh-info", { composed: true }));
// This is actually fine for cross-component communication
```

---

## Testing Strategy

### Unit Tests for Reducers

```typescript
// reducers/game.test.ts
import reducer from './game';
import {
  CONFIGURE_GAME_REQUEST,
  CONFIGURE_GAME_SUCCESS,
  CONFIGURE_GAME_FAILURE
} from '../actions/game';

describe('game reducer', () => {
  const initialState = {
    loading: false,
    error: null,
    // ... other fields
  };

  it('should set loading on REQUEST', () => {
    const state = reducer(initialState, { type: CONFIGURE_GAME_REQUEST });
    expect(state.loading).toBe(true);
    expect(state.error).toBe(null);
  });

  it('should clear loading on SUCCESS', () => {
    const loadingState = { ...initialState, loading: true };
    const state = reducer(loadingState, { type: CONFIGURE_GAME_SUCCESS });
    expect(state.loading).toBe(false);
  });

  it('should set error on FAILURE', () => {
    const state = reducer(initialState, {
      type: CONFIGURE_GAME_FAILURE,
      error: 'Test error',
      friendlyError: 'Something went wrong'
    });
    expect(state.loading).toBe(false);
    expect(state.error).toBe('Something went wrong');
  });
});
```

### Integration Tests for Components

```typescript
// components/boardgame-configure-game-properties.test.ts
import { fixture, html } from '@open-wc/testing';
import { store } from '../store';
import './boardgame-configure-game-properties';

describe('BoardgameConfigureGameProperties', () => {
  it('should show loading state', async () => {
    const el = await fixture(html`
      <boardgame-configure-game-properties
        .gameRoute="${{ name: 'test', id: '1' }}">
      </boardgame-configure-game-properties>
    `);

    // Dispatch loading action
    store.dispatch({ type: 'CONFIGURE_GAME_REQUEST' });
    await el.updateComplete;

    // Should show loading indicator
    expect(el.shadowRoot.querySelector('md-circular-progress')).to.exist;
  });
});
```

---

## Migration Path

### Phase 1: Add Reducer Cases (Low Risk)
- Add all case handlers to `reducers/game.js`
- Deploy and test
- State should update (even if components don't use it yet)

### Phase 2: Update One Component (Test Pattern)
- Pick simplest: `boardgame-configure-game-properties`
- Add loading/error state from Redux
- Test thoroughly
- If works, proceed to others

### Phase 3: Update Remaining Components
- `boardgame-player-roster`
- `boardgame-move-form`
- `boardgame-game-state-manager` (most complex)

### Phase 4: Update Thunks
- Change return type to `Promise<void>`
- Remove response returns
- Add any necessary chaining

---

## Backwards Compatibility

During migration, you can support **both patterns**:

```typescript
export const configureGame = (...) => async (dispatch: Dispatch) => {
  dispatch({ type: CONFIGURE_GAME_REQUEST });
  const response = await apiPost(...);

  if (response.error) {
    dispatch({ type: CONFIGURE_GAME_FAILURE, ... });
  } else {
    dispatch({ type: CONFIGURE_GAME_SUCCESS });
  }

  // Return response for backward compatibility during migration
  // TODO: Remove once all components updated
  return response;
};
```

---

## Verification Checklist

- [ ] Reducer handles all new action types
- [ ] INITIAL_STATE includes loading and error
- [ ] Selectors export loading and error selectors
- [ ] Components use stateChanged to subscribe to Redux
- [ ] Components render loading/error states
- [ ] Thunks dispatch all three action types (REQUEST, SUCCESS, FAILURE)
- [ ] Unit tests for reducer cases
- [ ] Integration tests for component state updates
- [ ] Redux DevTools shows state updates
- [ ] No direct response handling in components

---

## Timeline Estimate

- **Phase 1 (Reducers)**: 2-4 hours
- **Phase 2 (One component)**: 4-6 hours
- **Phase 3 (Remaining)**: 8-12 hours
- **Phase 4 (Thunks)**: 2-4 hours
- **Testing**: 8-16 hours

**Total**: 24-42 hours (3-5 days)

---

## Alternative: Simplify Instead

If full Redux is overkill, consider **simpler approach**:

```typescript
// Just use typed fetch, skip Redux entirely
class MyComponent extends LitElement {
  @property({ type: Boolean })
  private loading = false;

  @property({ type: String })
  private error: string | null = null;

  async handleSubmit() {
    this.loading = true;
    this.error = null;

    const response = await apiPost(...);

    if (response.error) {
      this.error = response.error;
    } else {
      // Success
    }

    this.loading = false;
  }
}
```

**Pros**: Simpler, less boilerplate, easier to understand
**Cons**: No centralized state, harder to share state between components

**Recommendation**: If you don't need shared state across many components, simpler approach is better. If you do need Redux, implement it properly.
