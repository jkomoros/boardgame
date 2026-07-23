import { LitElement, html } from 'lit';
import { property } from 'lit/decorators.js';

import { companionTimeline, ingestVersionTiming, usableAnimationContext } from './companion-sync.js';
import type { VersionAnimationContext } from './companion-sync.js';
import { forgetSurfaceForGame, surfaceForGame } from '../utils/companion-surface.js';
import { animHooks } from '../utils/anim-test-hooks.js';
import { store } from '../store.js';
import {
  fetchGameInfo,
  fetchGameVersion,
  enqueueStateBundle,
  dequeueStateBundle,
  clearStateBundles,
  setCurrentVersion,
  setTargetVersion,
  setLastFetchedVersion,
  socketConnected,
  socketDisconnected,
  socketError,
  clearFetchedInfo,
  clearFetchedVersion,
  cancelGameReadFlights,
} from '../actions/game.js';
import {
  selectPendingBundles,
  selectLastFiredBundle,
  selectNextBundle,
  selectHasPendingBundles,
  selectCurrentVersion,
  selectTargetVersion,
  selectLastFetchedVersion,
  selectSocketConnected,
  selectFetchedInfo,
  selectFetchedVersion,
  selectVersionFetching,
  selectInfoFetching
} from '../selectors.js';

import { connect } from 'pwa-helpers/connect-mixin.js';
import type {
  FetchedGameInfo,
  FetchedGameVersion,
  RootState,
} from '../types/store';
import type { GameFromServer, RawGameState, StateBundle, TimerInfo } from '../types/game-state';
import type { MoveForm, ProjectedMoveChoicesWire, ServerStateBundle } from '../types/api';
import { clientMoveFromWire } from '../types/client-move.js';
import { decodeSocketFrame } from '../types/socket-frame.js';
import type { HostedGameRenderer } from './boardgame-render-game.js';
import { BoardgameBaseGameRenderer } from './boardgame-base-game-renderer.js';
import { retryDelayMs } from '../utils/retry-policy.js';
import {
  compileLegacyAnimationOverlap,
  hasLegacyAnimationOverlap,
} from '../motion/legacy-overlap.js';
import { isCurrentMotionCycleRelease } from '../motion/release.js';

// Matches --animation-length: 0.5s default in boardgame-game-view.ts
const DEFAULT_ANIMATION_LENGTH_MS = 500;
const MAX_COMPANION_WAIT_MS = 10_000;

/**
 * StateManager keeps track of fetching state bundles from the server and
 * figuring out when it makes sense to have the game-view install them.
 *
 * When the game-view becomes active, the info bundle is fetched. This
 * includes information about who the current viewing player is and also
 * includes the initial state bundle.
 *
 * Once the first state bundle has been installed, a socket is connected to
 * receive updates about when the state increases. When the state version
 * increases, that increments TargetVersion, which changes the URL to fetch.
 */
class BoardgameGameStateManager extends connect(store)(LitElement) {
  @property({ type: Object })
  gameRoute: { name: string; id: string } | null = null;

  @property({ type: Boolean })
  gameFinished = false;

  @property({ type: Boolean })
  admin = false;

  @property({ type: Boolean })
  autoCurrentPlayer = false;

  @property({ type: Boolean })
  active = false;

  @property({ type: Boolean })
  loggedIn = false;

  @property({ type: String, attribute: false })
  gameVersionPath = '';

  @property({ type: String, attribute: false })
  gameViewPath = '';

  @property({ type: String })
  gameBasePath = '';

  @property({ type: String, attribute: false })
  effectiveGameVersionPath = '';

  @property({ type: Number })
  viewingAsPlayer = 0;

  @property({ type: Number })
  requestedPlayer = 0;

  @property({ type: Object })
  activeRenderer: HostedGameRenderer | null = null;

  @property({ type: String, attribute: false })
  private _socketUrl = '';

  @property({ type: Boolean, attribute: false })
  private _infoInstalled = false;

  @property({ type: Object, attribute: false })
  private _socket: WebSocket | null = null;
  private _reconnectTimer: number | null = null;
  private _socketRetryAttempt = 0;
  private _versionRetryAttempt = 0;
  private _versionRetryTimer: number | null = null;
  private _versionRequestFrame: number | null = null;
  private _infoRetryAttempt = 0;
  private _infoRetryTimer: number | null = null;
  private _infoRequestFrame: number | null = null;
  private readonly _onlineHandler = () => this.retryConnection();
  private readonly _offlineHandler = () => this._wentOffline();
  private readonly _resumeHandler = () => this._resumeVisibleSession();
  // True from dispatch until a valid response is installed. Unlike
  // _infoInstalled, this also covers authoritative metadata refreshes after
  // reconnect (roster, presence, room lock, and Table ownership).
  private _infoRetryRequired = false;
  // A refresh notification arriving during an in-flight read must not be
  // dropped: that response may have been generated before the notification.
  private _infoRefreshQueued = false;

  // _heartbeatTimer fires every 10 seconds while the socket is open and
  // sends a {"type":"heartbeat"} application-level keepalive. The server's
  // versionNotifier uses these to maintain per-(gameID, playerIndex)
  // presence; without them the absent-player badge + host SkipTurn flow
  // is end-to-end inert (spec §9.1).
  private _heartbeatTimer: number | null = null;

  // Fetched data - synced from Redux
  @property({ type: Object, attribute: false })
  private _fetchedInfo: FetchedGameInfo | null = null;

  @property({ type: Object, attribute: false })
  private _fetchedVersion: FetchedGameVersion | null = null;

  // Loading state - synced from Redux (per-operation)
  @property({ type: Boolean, attribute: false })
  private _versionFetching = false;

  @property({ type: Boolean, attribute: false })
  private _infoFetching = false;

  private _scheduledInstallTimerId: ReturnType<typeof setTimeout> | null = null;
  private _overlapTimerId: ReturnType<typeof setTimeout> | null = null;
  private _waitingForTimingVersion: number | null = null;
  // Invalidates timeout/RAF callbacks when a newer scheduling decision wins.
  private _installScheduleGeneration = 0;
  private _motionCycleSequence = 0;
  private _activeMotionCycleId = 0;
  private _releasedMotionCycleId = 0;

  // Track previous values for change detection
  private _prevTargetVersion = -1;
  private _prevGameVersion = 0;
  private _prevLastFetchedVersion = 0;
  private _prevVersionFetching = false;

  // Reactive properties - synced from Redux in stateChanged()
  @property({ type: Number, attribute: false })
  targetVersion = -1;

  @property({ type: Number, attribute: false })
  gameVersion = 0;

  @property({ type: Number, attribute: false })
  lastFetchedVersion = 0;

  @property({ type: Boolean, attribute: false })
  socketActive = false;

  @property({ type: Array, attribute: false })
  _pendingBundles: StateBundle[] = [];

  @property({ type: Object, attribute: false })
  _lastFiredBundle: StateBundle | null = null;

  constructor() {
    super();
    // Listen for ready-for-next-state event from game-view
    this.addEventListener('ready-for-next-state', (e: Event) => this._handleReadyForNextState(e));
  }

  private _handleReadyForNextState(e: Event) {
    const cycleId = (e as CustomEvent<{ cycleId?: number }>).detail?.cycleId;
    this.readyForNextState(cycleId);
    e.stopPropagation();
  }

  override firstUpdated(_changedProperties: Map<PropertyKey, unknown>) {
    super.firstUpdated(_changedProperties);
    this.updateData();
  }

  stateChanged(state: RootState) {
    const priorLastFetchedVersion = this.lastFetchedVersion;
    // Sync Redux loading state per-operation (non-duplicated, used for local logic)
    const prevVersionFetching = this._versionFetching;
    const prevInfoFetching = this._infoFetching;
    this._versionFetching = selectVersionFetching(state);
    this._infoFetching = selectInfoFetching(state);

    // Sync Redux fetched data (non-duplicated, used for one-time processing)
    const prevFetchedInfo = this._fetchedInfo;
    const prevFetchedVersion = this._fetchedVersion;
    this._fetchedInfo = selectFetchedInfo(state);
    this._fetchedVersion = selectFetchedVersion(state);

    // Sync properties that were previously getters
    this.targetVersion = selectTargetVersion(state);
    this.gameVersion = selectCurrentVersion(state);
    this.lastFetchedVersion = selectLastFetchedVersion(state);
    this.socketActive = selectSocketConnected(state);
    this._pendingBundles = selectPendingBundles(state);
    this._lastFiredBundle = selectLastFiredBundle(state);

    // Process fetched info when it becomes available
    const receivedInfo = Boolean(this._fetchedInfo && this._fetchedInfo !== prevFetchedInfo);
    if (receivedInfo) {
      const refreshAgain = this._infoRefreshQueued;
      this._infoRefreshQueued = false;
      this._infoRetryRequired = false;
      this._handleInfoData(this._fetchedInfo!);
      this._clearInfoRetry(true);
      // Clear after processing to prevent re-processing
      store.dispatch(clearFetchedInfo());
      if (refreshAgain) this.fetchInfo();
    }

    const receivedVersion = Boolean(
      this._fetchedVersion && this._fetchedVersion !== prevFetchedVersion
    );
    // Process fetched version when it becomes available
    if (receivedVersion) {
      this._handleVersionData(this._fetchedVersion!);
      // Clear after processing to prevent re-processing
      store.dispatch(clearFetchedVersion());
    }

    // Detect changes in properties and trigger handlers
    const currentTargetVersion = this.targetVersion;
    const currentGameVersion = this.gameVersion;
    const currentLastFetchedVersion = this.lastFetchedVersion;

    // Handle targetVersion changes
    const targetChanged = this._prevTargetVersion !== currentTargetVersion;
    if (targetChanged && currentTargetVersion >= 0) {
      this._clearVersionRetry(true);
      this._cancelVersionRequestFrame();
      this._handleTargetVersionChanged();
    }

    // A failed read leaves the target pending. Retry with capped backoff rather
    // than recursively fetching at render speed. A successful read is handled
    // above and resets the retry budget.
    if (prevVersionFetching && !this._versionFetching && !receivedVersion &&
        currentTargetVersion > currentGameVersion) {
      this._scheduleVersionRetry();
    }
    if (receivedVersion && currentLastFetchedVersion > priorLastFetchedVersion) {
      this._clearVersionRetry(true);
      if (currentTargetVersion > currentLastFetchedVersion) this._handleTargetVersionChanged();
    } else if (receivedVersion && currentTargetVersion > currentGameVersion) {
      // A nominal success that made no progress must not become a hot loop.
      this._scheduleVersionRetry();
    }
    if (prevInfoFetching && !this._infoFetching && !receivedInfo
      && (this._infoRetryRequired || !this._infoInstalled)) {
      this._scheduleInfoRetry();
    }

    // Trigger requestUpdate if properties changed (for updated() lifecycle)
    if (this._prevTargetVersion !== currentTargetVersion ||
        this._prevGameVersion !== currentGameVersion ||
        this._prevLastFetchedVersion !== currentLastFetchedVersion) {
      this.requestUpdate();
    }

    // Update previous values for next change detection
    this._prevTargetVersion = currentTargetVersion;
    this._prevGameVersion = currentGameVersion;
    this._prevLastFetchedVersion = currentLastFetchedVersion;
    this._prevVersionFetching = this._versionFetching;
  }

  override updated(changedProperties: Map<PropertyKey, unknown>) {
    super.updated(changedProperties);

    const previousRoute = changedProperties.get('gameRoute') as typeof this.gameRoute | undefined;
    const routeChanged = changedProperties.has('gameRoute')
      && !sameGameRoute(previousRoute ?? null, this.gameRoute);
    if (routeChanged) this._gameRouteChanged(previousRoute ?? null);

    // Get current values from computed properties (reading from Redux)
    const currentTargetVersion = this.targetVersion;
    const currentGameVersion = this.gameVersion;
    const currentLastFetchedVersion = this.lastFetchedVersion;

    // Recompute dependent properties when inputs change
    // Note: stateChanged() triggers requestUpdate() when computed properties change
    this.gameVersionPath = this._computeGameVersionPath(
      this.active, this.requestedPlayer, this.admin, currentTargetVersion, this.autoCurrentPlayer
    );

    this.gameViewPath = this._computeGameViewPath(this.requestedPlayer, this.admin, currentLastFetchedVersion);

    this.effectiveGameVersionPath = this._computeEffectiveGameVersionPath(
      this.gameVersionPath, currentLastFetchedVersion, currentGameVersion
    );

    if (changedProperties.has('active') || changedProperties.has('_infoInstalled') || routeChanged) {
      const newSocketUrl = this._computeSocketUrl(this.active, this._infoInstalled);
      // Only update if the URL actually changed
      // This prevents redundant reconnections during property update cycles
      if (newSocketUrl !== this._socketUrl) {
        this._socketUrl = newSocketUrl;
      }
    }

    // Emit event when socketActive changes so parent can update
    if (changedProperties.has('socketActive')) {
      this.dispatchEvent(new CustomEvent('socket-active-changed', {
        composed: true,
        bubbles: true,
        detail: { value: this.socketActive }
      }));
    }

    // Handle observers
    if (changedProperties.has('loggedIn')) {
      this._loggedInChanged(this.loggedIn);
    }

    if (changedProperties.has('active') && !routeChanged) {
      this._activeChanged(this.active);
    }

    if (changedProperties.has('_socketUrl')) {
      this._socketUrlChanged(this._socketUrl);
    }
  }

  private _computeEffectiveGameVersionPath(gameVersionPath: string, lastFetchedVersion: number, version: number): string {
    if (!gameVersionPath) return '';
    // version is already part of gameVersionPath. However, often on first
    // load, version and lastFetchedVersion are the same, and we should skip
    // fetching because we already have that info. However in some cases the
    // info bundle will not have all of the most up to date stuff, and we still
    // do need to fetch.
    if (lastFetchedVersion === version) return '';
    return `${gameVersionPath}&from=${lastFetchedVersion}`;
  }

  private _computeGameVersionPath(active: boolean, requestedPlayer: number, admin: boolean, version: number, autoCurrentPlayer: boolean): string {
    if (!active) return '';
    if (version < 0) return '';
    // TODO: factor this out with computeGameViewUrl a bit
    return `version/${version}?player=${requestedPlayer}&admin=${admin ? 1 : 0}&current=${autoCurrentPlayer ? 1 : 0}`;
  }

  private _computeGameViewPath(requestedPlayer: number, admin: boolean, lastFetchedVersion: number): string {
    return `info?player=${requestedPlayer}&admin=${admin ? 1 : 0}&from=${lastFetchedVersion}`;
  }

  private _computeSocketUrl(active: boolean, infoInstalled: boolean): string {
    if (!active) return '';
    if (!infoInstalled) return '';
    if (!this.gameRoute) return '';

    // Construct the socket URL from gameRoute
    const host = window.API_HOST ?? '';
    let result = `${host}/api/game/${this.gameRoute.name}/${this.gameRoute.id}/socket`;
    result = result.split('http:').join('ws:');
    result = result.split('https:').join('wss:');
    return result;
  }

  private _loggedInChanged(newValue: boolean) {
    this.softReset();
  }

  private _activeChanged(newValue: boolean) {
    if (newValue) {
      this.reset();
    } else {
      // If we don't clear this out when we deactivate then when we become
      // re-active there might be a brief period where our gameRoute is the
      // old one.
      this.gameRoute = null;
    }
  }

  private _gameRouteChanged(previousRoute: typeof this.gameRoute): void {
    cancelGameReadFlights();
    this._infoRetryRequired = false;
    this._infoRefreshQueued = false;
    this._clearVersionRetry(true);
    this._cancelVersionRequestFrame();
    this._clearInfoRetry(true);
    this._cancelInfoRequestFrame();
    this._socketRetryAttempt = 0;
    if (previousRoute) companionTimeline.resetGame(previousRoute.id);
    // Close synchronously. A queued frame from the previous route must not get
    // a chance to mutate the newly selected game's target version.
    this._socketUrl = '';
    this._socketUrlChanged('');
    if (this._reconnectTimer !== null) {
      window.clearTimeout(this._reconnectTimer);
      this._reconnectTimer = null;
    }
    this._stopHeartbeat();
    this.reset();
  }

  private _handleTargetVersionChanged() {
    if (this.targetVersion < 0) {
      return;
    }

    if (this.autoCurrentPlayer && this.requestedPlayer === this.viewingAsPlayer && this.targetVersion === this.gameVersion) {
      return;
    }

    // Skip if already have this version
    if (this.lastFetchedVersion === this.gameVersion && this.targetVersion === this.gameVersion) {
      return;
    }

    // Only block on version fetch, not on move submissions or other operations
    if (this._versionFetching) {
      return;
    }

    if (!this.gameRoute) {
      return;
    }

    if (this._versionRequestFrame !== null) return;
    const route = this.gameRoute;
    const targetVersion = this.targetVersion;
    const requestedPlayer = this.requestedPlayer;
    const admin = this.admin;
    const autoCurrentPlayer = this.autoCurrentPlayer;
    const lastFetchedVersion = this.lastFetchedVersion;
    const gameVersion = this.gameVersion;

    // Capture the request identity. A queued frame must not read a newly
    // selected route with an old route's target or viewer options.
    this._versionRequestFrame = requestAnimationFrame(() => {
      this._versionRequestFrame = null;
      if (!this.isConnected || !sameGameRoute(route, this.gameRoute) ||
          targetVersion !== this.targetVersion || this._versionFetching) return;
      store.dispatch(
        fetchGameVersion(
          route,
          targetVersion,
          requestedPlayer,
          admin,
          autoCurrentPlayer,
          lastFetchedVersion,
          gameVersion
        )
      );
    });
  }

  private _scheduleVersionRetry(): void {
    if (this._versionRetryTimer !== null || !this.active || !this.gameRoute) return;
    if (navigator.onLine === false) return;
    const route = this.gameRoute;
    const target = this.targetVersion;
    const delay = retryDelayMs(this._versionRetryAttempt++);
    this._versionRetryTimer = window.setTimeout(() => {
      this._versionRetryTimer = null;
      if (!sameGameRoute(route, this.gameRoute) || target !== this.targetVersion) return;
      this._handleTargetVersionChanged();
    }, delay);
  }

  private _clearVersionRetry(resetAttempt: boolean): void {
    if (this._versionRetryTimer !== null) {
      window.clearTimeout(this._versionRetryTimer);
      this._versionRetryTimer = null;
    }
    if (resetAttempt) this._versionRetryAttempt = 0;
  }

  private _cancelVersionRequestFrame(): void {
    if (this._versionRequestFrame === null) return;
    cancelAnimationFrame(this._versionRequestFrame);
    this._versionRequestFrame = null;
  }

  private _scheduleInfoRetry(): void {
    if (this._infoRetryTimer !== null || !this.active || !this.gameRoute) return;
    if (navigator.onLine === false) return;
    const route = this.gameRoute;
    const delay = retryDelayMs(this._infoRetryAttempt++);
    this._infoRetryTimer = window.setTimeout(() => {
      this._infoRetryTimer = null;
      if (!sameGameRoute(route, this.gameRoute)
        || (!this._infoRetryRequired && this._infoInstalled)) return;
      this._startInfoFetch();
    }, delay);
  }

  private _clearInfoRetry(resetAttempt: boolean): void {
    if (this._infoRetryTimer !== null) {
      window.clearTimeout(this._infoRetryTimer);
      this._infoRetryTimer = null;
    }
    if (resetAttempt) this._infoRetryAttempt = 0;
  }

  private _cancelInfoRequestFrame(): void {
    if (this._infoRequestFrame === null) return;
    cancelAnimationFrame(this._infoRequestFrame);
    this._infoRequestFrame = null;
  }

  private _socketUrlChanged(newValue: string) {
    if (this._socket) {
      const oldSocket = this._socket;
      this._socket = null;
      oldSocket.close();
    }

    if (newValue) this._connectSocket();
  }

  private _connectSocket() {
    if (this._reconnectTimer !== null) {
      window.clearTimeout(this._reconnectTimer);
      this._reconnectTimer = null;
    }
    const theUrl = this._socketUrl;

    // If there's no URL, don't establish a socket.
    if (!theUrl) return;
    if (navigator.onLine === false) return;

    const socket = new WebSocket(theUrl);
    this._socket = socket;

    // A close/message queued by an old route's socket must not mutate the new
    // route or arm another reconnect loop after this._socket was replaced.
    socket.onclose = (event) => {
      if (this._socket !== socket) return;
      this._socket = null;
      this._socketClosed(event);
    };
    socket.onerror = (event) => {
      if (this._socket === socket) this._socketError(event);
    };
    socket.onmessage = (event) => {
      if (this._socket === socket) this._socketMessage(event);
    };
    socket.onopen = (event) => {
      if (this._socket === socket) this._socketOpened(event);
    };
  }

  private _socketMessage(e: MessageEvent) {
    let frame;
    try {
      frame = decodeSocketFrame(e.data);
    } catch (error) {
      console.warn('Rejected malformed socket frame:', error);
      return;
    }

    // Reset only after protocol-valid traffic. Merely opening and immediately
    // flapping must continue increasing the backoff budget.
    this._socketRetryAttempt = 0;

    // Signal only after a frame satisfies the protocol. A stream of malformed
    // text must not masquerade as working chat delivery and disable polling.
    this.dispatchEvent(new CustomEvent('socket-active', {
      composed: true, bubbles: true,
    }));

    if (frame.type === 'version') {
      // Legacy raw frames have no timing sibling, so don't create a grace wait.
      if (frame.transport === 'json' && this.gameRoute) {
        companionTimeline.announce(this.gameRoute.id, frame.version);
      }
      if (frame.version > this.targetVersion) store.dispatch(setTargetVersion(frame.version));
      return;
    }
    if (frame.type === 'version-timing') {
      // Companion-mode cross-screen animation sync (spec §8.4).
      if (this.gameRoute) {
        ingestVersionTiming(this.gameRoute.id, frame.timing);
        if (this._waitingForTimingVersion === frame.timing.version) this._scheduleNextStateBundle();
      }
      return;
    }
    if (frame.type === 'clock-sync') {
      companionTimeline.ingestClockSync(frame.clock);
      return;
    }
    if (frame.type === 'mode-changed') {
      // Only the decoded one-way transition to solo reaches this branch.
      if (!this.gameRoute || frame.gameID !== this.gameRoute.id) {
        console.warn('Ignored mode-change frame for a different game:', frame.gameID);
        return;
      }
      this._clearSurfaceCookieForThisGame();
      window.location.reload();
      return;
    }
    if (frame.type === 'presence-changed') {
      if (!this.gameRoute || frame.gameID !== this.gameRoute.id) {
        console.warn('Ignored presence frame for a different game:', frame.gameID);
        return;
      }
      this.fetchInfo();
      return;
    }
    if (frame.type === 'table-session-changed' || frame.type === 'table-lease-lost') {
      if (!this.gameRoute || frame.gameID !== this.gameRoute.id) {
        console.warn('Ignored Table-session frame for a different game:', frame.gameID);
        return;
      }
      this.fetchInfo();
      return;
    }
    if (frame.type === 'chat') {
      this.dispatchEvent(new CustomEvent('chat-notification', {
        composed: true,
        bubbles: true,
        detail: { channel: frame.channel, messageId: frame.messageID },
      }));
      return;
    }
    console.warn('Unknown socket message type:', frame.wireType);
  }

  private _socketError(e: Event) {
    console.warn('Socket error', e);
    store.dispatch(socketError(e.toString()));
  }

  private _socketOpened(e: Event) {
    store.dispatch(socketConnected());
    this._startHeartbeat();
    // Socket registration only carries a game version. Re-read authoritative
    // non-version metadata on every open so changes missed while disconnected
    // (presence, roster, lock, Table ownership) converge immediately. This
    // also closes the initial info-GET/socket-registration race.
    this.fetchInfo();
    // Warm the midpoint clock estimator before the first real move. Three
    // quick request/reply samples let it select the lowest-RTT offset instead
    // of treating gameplay versions as clock-sync warmup.
    for (let i = 0; i < 3; i++) {
      this._socket?.send(JSON.stringify({
        type: 'clock-sync',
        data: { clientSentAt: Date.now(), nonce: i },
      }));
    }
  }

  private _socketClosed(e: CloseEvent) {
    store.dispatch(socketDisconnected());
    this._stopHeartbeat();
    // We always want a socket, so connect. Wait a bit so we don't just
    // busy spin if the server is down.

    // If we closed because we no longer have a valid URL, then
    // _connectSocket will just exit, and this loop won't be called.

    if (!this._socketUrl || navigator.onLine === false) return;
    const reconnectUrl = this._socketUrl;
    const delay = retryDelayMs(this._socketRetryAttempt++);
    this._reconnectTimer = window.setTimeout(() => {
      this._reconnectTimer = null;
      // Navigation or an already-open replacement supersedes this timer.
      if (this._socket === null && this._socketUrl === reconnectUrl) this._connectSocket();
    }, delay);
  }

  override connectedCallback(): void {
    super.connectedCallback();
    window.addEventListener('online', this._onlineHandler);
    window.addEventListener('offline', this._offlineHandler);
    window.addEventListener('pageshow', this._resumeHandler);
    document.addEventListener('visibilitychange', this._resumeHandler);
    // A Lit element may be temporarily detached and reinserted without any
    // reactive property changing. Restore the connection that teardown owns.
    if (this._socketUrl && this._socket === null) this._connectSocket();
  }

  override disconnectedCallback(): void {
    super.disconnectedCallback();
    window.removeEventListener('online', this._onlineHandler);
    window.removeEventListener('offline', this._offlineHandler);
    window.removeEventListener('pageshow', this._resumeHandler);
    document.removeEventListener('visibilitychange', this._resumeHandler);
    cancelGameReadFlights();
    this._clearVersionRetry(true);
    this._cancelVersionRequestFrame();
    this._clearInfoRetry(true);
    this._cancelInfoRequestFrame();
    this._stopHeartbeat();
    if (this._reconnectTimer !== null) {
      window.clearTimeout(this._reconnectTimer);
      this._reconnectTimer = null;
    }
    if (this._socket) {
      const socket = this._socket;
      this._socket = null;
      socket.close();
    }
  }

  /** Immediately retries all live-session transports after a user action or
   * the browser's online event. Safe to call repeatedly. */
  retryConnection(): void {
    this._socketRetryAttempt = 0;
    if (this._reconnectTimer !== null) {
      window.clearTimeout(this._reconnectTimer);
      this._reconnectTimer = null;
    }
    this._clearVersionRetry(true);
    this._cancelVersionRequestFrame();
    this._clearInfoRetry(true);
    this._cancelInfoRequestFrame();
    // A fetch may be stuck behind the network transition. Abort both
    // route-scoped reads and synchronously reset their Redux loading flags so
    // the replacements below cannot be suppressed as "already fetching".
    cancelGameReadFlights();
    if (this._socket === null && this._socketUrl) this._connectSocket();
    if (this.targetVersion > this.gameVersion) this._handleTargetVersionChanged();
    this.fetchInfo();
  }

  private _wentOffline(): void {
    if (this._reconnectTimer !== null) {
      window.clearTimeout(this._reconnectTimer);
      this._reconnectTimer = null;
    }
    this._stopHeartbeat();
    this._infoRetryRequired = true;
    cancelGameReadFlights();
    if (this._socket) {
      const socket = this._socket;
      this._socket = null;
      socket.close();
    }
    if (this.socketActive) store.dispatch(socketDisconnected());
  }

  private _resumeVisibleSession(): void {
    if (document.visibilityState === 'hidden' || navigator.onLine === false) return;
    if (this._socket === null) {
      this.retryConnection();
      return;
    }
    // Timers are aggressively throttled in background tabs. Renew immediately
    // and refresh metadata when the page resumes rather than waiting up to one
    // heartbeat/lease interval to discover displacement or expiry.
    this._sendHeartbeat();
    this.fetchInfo();
  }

  private _startHeartbeat() {
    this._stopHeartbeat();
    // Renew a Table lease immediately. Waiting for the first interval would
    // burn a quarter of the grace window during every connection/reload.
    this._sendHeartbeat();
    // 10s cadence: server's absentThreshold is 30s, so one missed
    // heartbeat (e.g. brief network burp) doesn't flap the absent flag.
    this._heartbeatTimer = window.setInterval(() => {
      this._sendHeartbeat();
    }, 10000);
  }

  private _sendHeartbeat(): void {
    if (!this._socket || this._socket.readyState !== WebSocket.OPEN) return;
    try {
      this._socket.send(JSON.stringify({ type: 'heartbeat' }));
    } catch (err) {
      // Send failures are non-fatal — the socket will close on its own and
      // _socketClosed re-arms via _connectSocket.
      console.warn('heartbeat send failed:', err);
    }
  }

  private _stopHeartbeat() {
    if (this._heartbeatTimer !== null) {
      window.clearInterval(this._heartbeatTimer);
      this._heartbeatTimer = null;
    }
  }

  // _clearAllSurfaceCookies expires every surface_<gameID> cookie this
  // browser holds. Used on switchToSolo (mode-changed) so the post-reload
  // loader picks the solo renderer. Iterates document.cookie because we
  // don't track which gameIDs the user has touched.
  private _clearSurfaceCookieForThisGame() {
    if (!this.gameRoute) return;
    document.cookie = 'surface_' + this.gameRoute.id + '=; Path=/; Max-Age=0';
    forgetSurfaceForGame(this.gameRoute.id);
  }

  updateData() {
    this.fetchInfo();
  }

  // When we should do a soft reset; that is, when we haven't flipped out and
  // back; it's still the same game we're viewing as before.
  softReset() {
    this._infoInstalled = false;
    this._clearInfoRetry(true);
    this._cancelInfoRequestFrame();
    this._infoRequestFrame = window.requestAnimationFrame(() => {
      this._infoRequestFrame = null;
      if (this.isConnected) this.updateData();
    });
  }

  // When everything should be reset
  reset() {
    this._clearVersionRetry(true);
    this._cancelVersionRequestFrame();
    this._clearInfoRetry(true);
    this._cancelInfoRequestFrame();
    this._installScheduleGeneration++;
    this._activeMotionCycleId = ++this._motionCycleSequence;
    this._releasedMotionCycleId = this._activeMotionCycleId;
    this._waitingForTimingVersion = null;
    this._clearOverlapTimer();
    if (this._scheduledInstallTimerId !== null) {
      clearTimeout(this._scheduledInstallTimerId);
      this._scheduledInstallTimerId = null;
    }
    if (this.gameRoute) companionTimeline.resetGame(this.gameRoute.id);
    store.dispatch(setLastFetchedVersion(0));
    store.dispatch(setTargetVersion(-1));
    store.dispatch(setCurrentVersion(0));
    store.dispatch(clearStateBundles());
    this.softReset();
  }

  fetchInfo() {
    this._infoRetryRequired = true;
    if (this._infoFetching) {
      this._infoRefreshQueued = true;
      return;
    }
    this._startInfoFetch();
  }

  private _startInfoFetch(): void {
    // Only block on info fetch, not on other operations
    if (this._infoFetching) {
      return;
    }

    if (!this.active) {
      return;
    }

    if (!this.gameRoute) {
      // The URL will be junk
      return;
    }

    this._infoRetryRequired = true;
    this._clearInfoRetry(false);

    // Dispatch the thunk - data will be processed via stateChanged when it arrives
    store.dispatch(
      fetchGameInfo(
        this.gameRoute,
        this.requestedPlayer,
        this.admin,
        this.lastFetchedVersion
      )
    );
  }

  private _prepareStateBundle(
    game: GameFromServer,
    moveForms: MoveForm[] | null,
    viewingAsPlayer: number,
    move: unknown,
    projectedMoveChoices: ProjectedMoveChoicesWire | null,
  ): StateBundle {
    return {
      originalWallClockStartTime: Date.now(),
      game,
      move: clientMoveFromWire(move),
      moveForms,
      viewingAsPlayer,
      projectedMoveChoices,
    };
  }

  // Called when gameView tells us to pass up the next state if we have one
  // (the animations are done).
  readyForNextState(cycleId?: number) {
    if (!isCurrentMotionCycleRelease(
      cycleId,
      this._activeMotionCycleId,
      this._releasedMotionCycleId,
    )) return;
    this._releasedMotionCycleId = cycleId;
    this._clearOverlapTimer();
    if (this._scheduledInstallTimerId !== null) {
      clearTimeout(this._scheduledInstallTimerId);
      this._scheduledInstallTimerId = null;
    }
    this._scheduleNextStateBundle();
  }

  // A new state bundle has been enqueued. Ensure that we're working to fire a
  // state bundle. renderer might be a reference to the underlying renderer, or
  // null.
  private _scheduleNextStateBundle() {
    if (!this._pendingBundles.length) return;

    const generation = ++this._installScheduleGeneration;
    this._waitingForTimingVersion = null;

    // A re-schedule (e.g. from an exact-cycle release or a
    // fresh enqueue) supersedes any previously-armed scheduled-install
    // timer, so cancel it here to avoid arming a second timer that could
    // fire early or install twice within the up-to-2s companion-sync window.
    if (this._scheduledInstallTimerId !== null) {
      clearTimeout(this._scheduledInstallTimerId);
      this._scheduledInstallTimerId = null;
    }

    const renderer = this.activeRenderer;
    let effectiveAnimationLength = DEFAULT_ANIMATION_LENGTH_MS;

    // If we were given a renderer that customizes animation length, consult it.
    if (renderer) {
      const nextBundle = this._pendingBundles[0];
      const lastBundle = this._lastFiredBundle;
      const nextMove = nextBundle ? nextBundle.move : null;
      const lastMove = lastBundle ? lastBundle.move : null;
      if (nextMove || lastMove) {
        if (renderer.animationLength) {
          const length = renderer.animationLength(lastMove, nextMove);
          // If the length is negative, that's the signal to skip binding this one.
          if (length < 0) {
            // We always render the last bundle to install
            if (this._pendingBundles.length > 1) {
              // Skip this bundle by dequeuing it
              store.dispatch(dequeueStateBundle());
              if (this.gameRoute && Number.isInteger(nextBundle.game?.Version)) {
                companionTimeline.forgetVersion(this.gameRoute.id, nextBundle.game.Version);
              }
              this._scheduleNextStateBundle();
              return;
            }
          } else {
            if (length > 0) effectiveAnimationLength = length;
            this.dispatchEvent(new CustomEvent('set-animation-length', { composed: true, bubbles: true, detail: length }));
          }
        }
      }
    }

    // Companion sync (#798): resolve timing for THIS bundle's version. HTTP
    // catch-up responses may contain several bundles; consulting a mutable
    // "latest" timestamp here would assign all of them the final version's
    // slot. Solo games remain immediate.
    const surface = this.gameRoute ? surfaceForGame(this.gameRoute.id) : null;
    if (surface === 'table' || surface === 'hand') {
      const nextBundle = this._pendingBundles[0];
      const version = nextBundle?.game?.Version;
      if (this.gameRoute && Number.isInteger(version)) {
        const schedule = companionTimeline.schedule(this.gameRoute.id, version);
        if (schedule.kind === 'awaiting-timing') {
          this._waitingForTimingVersion = version;
          this._scheduledInstallTimerId = setTimeout(() => {
            if (generation !== this._installScheduleGeneration) return;
            this._scheduledInstallTimerId = null;
            this._scheduleNextStateBundle();
          }, schedule.waitMs);
          return;
        }

        if (schedule.kind === 'scheduled') {
          const context = usableAnimationContext(schedule.context, Date.now(), MAX_COMPANION_WAIT_MS);
          if (context) {
            if (effectiveAnimationLength > context.maxAnimationDurationMs) {
              console.warn(
                `[state-manager] companion animation length ${effectiveAnimationLength}ms exceeds ` +
                `${context.maxAnimationDurationMs}ms version contract; capping synchronized cycle`,
              );
              effectiveAnimationLength = context.maxAnimationDurationMs;
              this.dispatchEvent(new CustomEvent('set-animation-length', {
                composed: true, bubbles: true, detail: effectiveAnimationLength,
              }));
            }
            // Install only inside the protocol's preparation window. This is
            // early enough to render and pre-arm backwards-filled WAAPI
            // animations, without exposing the next logical state seconds
            // before its visible cycle begins.
            const preparationLeadMs = Math.max(0,
              context.slotDurationMs - context.maxAnimationDurationMs);
            const installDelayMs = Math.max(0,
              context.startAtMs - preparationLeadMs - Date.now());
            if (installDelayMs > 0) {
              this._scheduledInstallTimerId = setTimeout(() => {
                if (generation !== this._installScheduleGeneration) return;
                this._scheduledInstallTimerId = null;
                this._asyncFireNextStateBundle(effectiveAnimationLength, context, generation);
              }, installDelayMs);
            } else {
              this._asyncFireNextStateBundle(effectiveAnimationLength, context, generation);
            }
          } else {
            console.warn('[state-manager] version animation target is unusable; installing immediately');
            this._asyncFireNextStateBundle(effectiveAnimationLength, null, generation);
          }
          return;
        }
      }
    }

    this._asyncFireNextStateBundle(effectiveAnimationLength, null, generation);
  }

  private _asyncFireNextStateBundle(
    effectiveAnimationLength = DEFAULT_ANIMATION_LENGTH_MS,
    animationContext: VersionAnimationContext | null = null,
    generation = this._installScheduleGeneration,
  ) {
    // Not entirely sure why this has to be done this way, but it needs to be
    // done outside of the current task, even when fired from a timeout.
    window.requestAnimationFrame(() => this._fireNextStateBundle(effectiveAnimationLength, animationContext, generation));
  }

  private _fireNextStateBundle(
    effectiveAnimationLength = DEFAULT_ANIMATION_LENGTH_MS,
    animationContext: VersionAnimationContext | null = null,
    generation = this._installScheduleGeneration,
  ) {
    if (generation !== this._installScheduleGeneration) return;
    // Called when the next state bundle should be installed NOW.
    // Dequeue from Redux and fire event
    if (this._pendingBundles.length > 0) {
      const bundle = this._pendingBundles[0];
      const motionCycleId = ++this._motionCycleSequence;
      this._activeMotionCycleId = motionCycleId;
      this._clearOverlapTimer();
      const renderer = this.activeRenderer;
      const legacyOverlapConfigured = !!renderer && hasLegacyAnimationOverlap(
        renderer,
        BoardgameBaseGameRenderer.prototype.animationOverlap,
      );
      store.dispatch(dequeueStateBundle());
      if (animationContext) {
        animHooks.record('install', undefined, {
          version: animationContext.version,
          targetAtMs: animationContext.startAtMs,
        });
      }
      this.dispatchEvent(new CustomEvent('install-state-bundle', {
        composed: true,
        bubbles: true,
        detail: {
          ...bundle,
          animationContext,
          motionCycleId,
          legacyAnimationOverlapConfigured: legacyOverlapConfigured,
        },
      }));
      if (this.gameRoute && Number.isInteger(bundle.game?.Version)) {
        companionTimeline.forgetVersion(this.gameRoute.id, bundle.game.Version);
      }
      // Preserve the historical hook exactly: it is a solo-only state-clock
      // delay based on animationLength and the already-buffered successor. The
      // callback still enters through the exact-cycle token gate, so a stale
      // timeout can never advance a newer installation.
      const successor = this._pendingBundles[0] ?? null;
      if (renderer && successor && animationContext === null) {
        const legacyOverlap = compileLegacyAnimationOverlap(
          renderer,
          BoardgameBaseGameRenderer.prototype.animationOverlap,
          bundle.move,
          successor.move,
          effectiveAnimationLength,
        );
        if (legacyOverlap.delayMs !== null) {
          this._overlapTimerId = setTimeout(() => {
            this._overlapTimerId = null;
            this.readyForNextState(motionCycleId);
          }, legacyOverlap.delayMs);
        }
      }
    }
  }

  private _clearOverlapTimer(): void {
    if (this._overlapTimerId === null) return;
    clearTimeout(this._overlapTimerId);
    this._overlapTimerId = null;
  }

  // Add the next state bundle to the end
  private _enqueueStateBundle(bundle: StateBundle) {
    const wasEmpty = this._pendingBundles.length === 0;
    store.dispatch(enqueueStateBundle(bundle));
    // If that was the first one we added, go ahead and fire it right now.
    if (wasEmpty) this._scheduleNextStateBundle();
  }

  private _handleInfoData(data: FetchedGameInfo) {
    const installingInitialState = !this._infoInstalled;
    const gameInfo = {
      chest: data.Chest,
      playersInfo: data.Players,
      hasEmptySlots: data.HasEmptySlots,
      open: data.GameOpen,
      visible: data.GameVisible,
      isOwner: data.IsOwner,
      // Companion-mode bundle from doGameInfo (spec §9.1 + §12). Empty
      // object for solo-mode games (CompanionInfo is always present in
      // the response but its sub-fields are zero-valued).
      companionInfo: data.CompanionInfo || null,
			moveInputSchemaFingerprint: data.MoveInputSchemaFingerprint,
    };

    this.dispatchEvent(new CustomEvent('install-game-static-info', { composed: true, bubbles: true, detail: gameInfo }));

    this._infoInstalled = true;
    if (installingInitialState) {
      const bundle = this._prepareStateBundle(
        data.Game, data.Forms, data.ViewingAsPlayer, null, data.ProjectedMoveChoices ?? null,
      );
      this._enqueueStateBundle(bundle);

      // We don't use data.Game.Version for lastFetched because first load may
      // intentionally return an older state to animate setup moves. The server
      // ships the authoritative downloaded-through version separately.
      store.dispatch(setLastFetchedVersion(data.StateVersion));
      store.dispatch(setTargetVersion(data.Game.Version));
      store.dispatch(setCurrentVersion(data.Game.Version));
      return;
    }

    // Reconnect/presence/Table refreshes install static metadata only. Replaying
    // the current state would duplicate animations. If this GET also discovers
    // a missed game version, feed that target into the ordinary version-bundle
    // path, which preserves every animation and stale-response guard.
    if (data.StateVersion > this.targetVersion) {
      store.dispatch(setTargetVersion(data.StateVersion));
    }
  }

  private _handleVersionData(data: FetchedGameVersion) {
    let lastServerBundle: ServerStateBundle | null = null;

    for (let i = 0; i < data.Bundles.length; i++) {
      const serverBundle = data.Bundles[i];
      const bundle = this._prepareStateBundle(
        serverBundle.Game, serverBundle.Forms, serverBundle.ViewingAsPlayer,
        serverBundle.Move, serverBundle.ProjectedMoveChoices ?? null,
      );
      this._enqueueStateBundle(bundle);
      lastServerBundle = serverBundle;
    }

    if (lastServerBundle) {
      store.dispatch(setLastFetchedVersion(lastServerBundle.Game.Version));
      store.dispatch(setCurrentVersion(lastServerBundle.Game.Version));
    }
  }

  override render() {
    // Component manages fetching via Redux thunks, no template needed
    return html``;
  }
}

customElements.define('boardgame-game-state-manager', BoardgameGameStateManager);

function sameGameRoute(
  first: { name: string; id: string } | null,
  second: { name: string; id: string } | null,
): boolean {
  return first?.name === second?.name && first?.id === second?.id;
}

export { BoardgameGameStateManager };
