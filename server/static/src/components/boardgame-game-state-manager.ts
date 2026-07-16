import { LitElement, html } from 'lit';
import { property } from 'lit/decorators.js';

import { companionTimeline, ingestVersionTiming, usableAnimationContext } from './companion-sync.js';
import type { VersionAnimationContext } from './companion-sync.js';
import { surfaceForGame } from '../utils/companion-surface.js';
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
  clearFetchedVersion
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
import type { MoveForm, ServerStateBundle } from '../types/api';
import { clientMoveFromWire } from '../types/client-move.js';
import type { HostedGameRenderer } from './boardgame-render-game.js';

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

  private _overlapTimerId: ReturnType<typeof setTimeout> | null = null;
  private _scheduledInstallTimerId: ReturnType<typeof setTimeout> | null = null;
  private _waitingForTimingVersion: number | null = null;
  // Invalidates timeout/RAF callbacks when a newer scheduling decision wins.
  private _installScheduleGeneration = 0;

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
    this.readyForNextState();
    e.stopPropagation();
  }

  override firstUpdated(_changedProperties: Map<PropertyKey, unknown>) {
    super.firstUpdated(_changedProperties);
    this.updateData();
  }

  stateChanged(state: RootState) {
    // Sync Redux loading state per-operation (non-duplicated, used for local logic)
    const prevVersionFetching = this._versionFetching;
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
    if (this._fetchedInfo && this._fetchedInfo !== prevFetchedInfo) {
      this._handleInfoData(this._fetchedInfo);
      // Clear after processing to prevent re-processing
      store.dispatch(clearFetchedInfo());
    }

    // Process fetched version when it becomes available
    if (this._fetchedVersion && this._fetchedVersion !== prevFetchedVersion) {
      this._handleVersionData(this._fetchedVersion);
      // Clear after processing to prevent re-processing
      store.dispatch(clearFetchedVersion());
    }

    // Detect changes in properties and trigger handlers
    const currentTargetVersion = this.targetVersion;
    const currentGameVersion = this.gameVersion;
    const currentLastFetchedVersion = this.lastFetchedVersion;

    // Handle targetVersion changes
    if (this._prevTargetVersion !== currentTargetVersion && currentTargetVersion >= 0) {
      this._handleTargetVersionChanged();
    }

    // Handle version fetch completion - retry fetch if we have a pending target
    if (prevVersionFetching && !this._versionFetching && currentTargetVersion > currentGameVersion) {
      this._handleTargetVersionChanged();
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

    // Only update socket URL if active or _infoInstalled changes
    // gameRoute changes are handled via those dependencies
    if (changedProperties.has('active') || changedProperties.has('_infoInstalled')) {
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

    if (changedProperties.has('active')) {
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

    // Dispatch the thunk - data will be processed via stateChanged when it arrives
    requestAnimationFrame(() => {
      store.dispatch(
        fetchGameVersion(
          this.gameRoute!,
          this.targetVersion,
          this.requestedPlayer,
          this.admin,
          this.autoCurrentPlayer,
          this.lastFetchedVersion,
          this.gameVersion
        )
      );
    });
  }

  private _socketUrlChanged(newValue: string) {
    // Don't tear down an existing connection if the new URL is empty
    // This can happen during property update cycles
    if (this._socket && !newValue) {
      return;
    }

    if (this._socket) {
      this._socket.close();
      this._socket = null;
    }

    this._connectSocket();
  }

  private _connectSocket() {
    const theUrl = this._socketUrl;

    // If there's no URL, don't establish a socket.
    if (!theUrl) return;

    this._socket = new WebSocket(theUrl);

    this._socket.onclose = (e) => this._socketClosed(e);
    this._socket.onerror = (e) => this._socketError(e);
    this._socket.onmessage = (e) => this._socketMessage(e);
    this._socket.onopen = (e) => this._socketOpened(e);
  }

  private _socketMessage(e: MessageEvent) {
    const data = e.data as string;

    // Signal that WebSocket is working (chat panel uses this to reduce polling)
    this.dispatchEvent(new CustomEvent('socket-active', {
      composed: true, bubbles: true,
    }));

    // Feature-detect JSON framing vs legacy raw version numbers
    if (data.startsWith('{')) {
      try {
        const msg = JSON.parse(data);
        if (msg.type === 'version') {
          if (this.gameRoute) {
            companionTimeline.announce(this.gameRoute.id, msg.data);
          }
          if (Number.isInteger(msg.data) && msg.data > this.targetVersion) {
            store.dispatch(setTargetVersion(msg.data));
          }
        } else if (msg.type === 'version-timing') {
          // Companion-mode cross-screen animation sync (spec §8.4).
          // Sibling to 'version' — old clients ignore it. Carries
          // serverSentAt + serverPlayAt; we feed them into the
          // version-bound timeline and minimum-wins latency estimator so
          // the shared animator can target a server-anchored instant.
          if (this.gameRoute) {
            ingestVersionTiming(this.gameRoute.id, msg.data);
            // A fetched bundle can beat its sibling timing frame on a very
            // fast local connection. Only reschedule the bundle explicitly
            // waiting in the 200ms grace window; never advance a later bundle
            // while the current animation gate is still open.
            if (this._waitingForTimingVersion === msg.data?.version) {
              this._scheduleNextStateBundle();
            }
          }
        } else if (msg.type === 'clock-sync') {
          companionTimeline.ingestClockSync(msg.data);
        } else if (msg.type === 'mode-changed') {
          // Companion-mode → solo downgrade triggered by host
          // switchToSolo (spec §9.6). Clear THIS game's surface cookie
          // before reloading so the loader picks the solo renderer
          // (server only set the cookie-clear on the host's response;
          // phones need to clear themselves). Scoped to this game only:
          // the browser may hold surface cookies for other in-flight
          // companion games, and the broadcast is per-game.
          this._clearSurfaceCookieForThisGame();
          window.location.reload();
        } else if (msg.type === 'presence-changed') {
          // Companion-mode heartbeat scan flipped a player into/out of
          // the Absent set. Refetch gameInfo so the new Absent list
          // surfaces in state and the Table view re-renders the
          // "Waiting for Alice…" badges (spec §9.1).
          this.fetchInfo();
        } else if (msg.type === 'chat') {
          // Dispatch chat notification event for the chat panel to handle
          this.dispatchEvent(new CustomEvent('chat-notification', {
            composed: true,
            bubbles: true,
            detail: msg.data,
          }));
        } else {
          console.warn('Unknown socket message type:', msg.type);
        }
      } catch (err) {
        console.warn('Failed to parse socket JSON message:', data, err);
      }
      return;
    }

    // Legacy: raw version number
    const version = parseInt(data);
    if (isNaN(version)) {
      console.warn('Socket message was not a valid version number:', data);
      return;
    }
    if (version > this.targetVersion) store.dispatch(setTargetVersion(version));
  }

  private _socketError(e: Event) {
    console.warn('Socket error', e);
    store.dispatch(socketError(e.toString()));
  }

  private _socketOpened(e: Event) {
    store.dispatch(socketConnected());
    this._startHeartbeat();
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

    // TODO: exponential backoff on server connect.
    setTimeout(() => this._connectSocket(), 250);
  }

  private _startHeartbeat() {
    this._stopHeartbeat();
    // 10s cadence: server's absentThreshold is 30s, so one missed
    // heartbeat (e.g. brief network burp) doesn't flap the absent flag.
    this._heartbeatTimer = window.setInterval(() => {
      if (this._socket && this._socket.readyState === WebSocket.OPEN) {
        try {
          this._socket.send(JSON.stringify({ type: 'heartbeat' }));
        } catch (err) {
          // Send failures are non-fatal — the socket will close on its
          // own and _socketClosed re-arms via _connectSocket.
          console.warn('heartbeat send failed:', err);
        }
      }
    }, 10000);
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
  }

  updateData() {
    this.fetchInfo();
  }

  // When we should do a soft reset; that is, when we haven't flipped out and
  // back; it's still the same game we're viewing as before.
  softReset() {
    this._infoInstalled = false;
    window.requestAnimationFrame(() => this.updateData());
  }

  // When everything should be reset
  reset() {
    this._installScheduleGeneration++;
    this._waitingForTimingVersion = null;
    if (this._overlapTimerId !== null) {
      clearTimeout(this._overlapTimerId);
      this._overlapTimerId = null;
    }
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
  ): StateBundle {
    return {
      originalWallClockStartTime: Date.now(),
      game,
      move: clientMoveFromWire(move),
      moveForms,
      viewingAsPlayer,
    };
  }

  // Called when gameView tells us to pass up the next state if we have one
  // (the animations are done).
  readyForNextState() {
    if (this._overlapTimerId !== null) {
      clearTimeout(this._overlapTimerId);
      this._overlapTimerId = null;
    }
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

    // A re-schedule (e.g. from readyForNextState(), the overlap timer, or a
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
        detail: { ...bundle, animationContext },
      }));
      if (this.gameRoute && Number.isInteger(bundle.game?.Version)) {
        companionTimeline.forgetVersion(this.gameRoute.id, bundle.game.Version);
      }

      // Check for overlap: start next animation early if renderer requests it
      if (this._pendingBundles.length > 0 && animationContext === null) {
        const renderer = this.activeRenderer;
        if (renderer?.animationOverlap) {
          const nextBundle = this._pendingBundles[0];
          const fraction = Math.max(0, Math.min(1,
            renderer.animationOverlap(bundle.move, nextBundle?.move ?? null)
          ));
          if (fraction > 0 && fraction < 1) {
            const overlapMs = fraction * effectiveAnimationLength;
            this._overlapTimerId = setTimeout(() => {
              this._overlapTimerId = null;
              this._scheduleNextStateBundle();
            }, overlapMs);
          }
        }
      }
    }
  }

  // Add the next state bundle to the end
  private _enqueueStateBundle(bundle: StateBundle) {
    const wasEmpty = this._pendingBundles.length === 0;
    store.dispatch(enqueueStateBundle(bundle));
    // If that was the first one we added, go ahead and fire it right now.
    if (wasEmpty) this._scheduleNextStateBundle();
  }

  private _handleInfoData(data: FetchedGameInfo) {
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

    const bundle = this._prepareStateBundle(data.Game, data.Forms, data.ViewingAsPlayer, null);
    this._enqueueStateBundle(bundle);

    this._infoInstalled = true;

    // We don't use data.Game.Version, because in some cases the current
    // state we're returning is not actually current state, but an old one to
    // force us to play animations for moves that are made before a player move
    // is. The server ships down this information in a special field.
    store.dispatch(setLastFetchedVersion(data.StateVersion));
    store.dispatch(setTargetVersion(data.Game.Version));
    store.dispatch(setCurrentVersion(data.Game.Version));
  }

  private _handleVersionData(data: FetchedGameVersion) {
    let lastServerBundle: ServerStateBundle | null = null;

    for (let i = 0; i < data.Bundles.length; i++) {
      const serverBundle = data.Bundles[i];
      const bundle = this._prepareStateBundle(serverBundle.Game, serverBundle.Forms, serverBundle.ViewingAsPlayer, serverBundle.Move);
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

export { BoardgameGameStateManager };
