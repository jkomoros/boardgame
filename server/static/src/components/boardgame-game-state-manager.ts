import { LitElement, html } from 'lit';
import { property } from 'lit/decorators.js';

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
  selectGameLoading
} from '../selectors.js';

import { connect } from 'pwa-helpers/connect-mixin.js';
import type { RootState } from '../types/store';

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

  @property({ type: Object })
  chest: any = null;

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
  activeRenderer: any = null;

  @property({ type: String, attribute: false })
  private _socketUrl = '';

  @property({ type: Boolean, attribute: false })
  private _infoInstalled = false;

  @property({ type: Object, attribute: false })
  private _socket: WebSocket | null = null;

  // Fetched data - synced from Redux
  @property({ type: Object, attribute: false })
  private _fetchedInfo: any = null;

  @property({ type: Object, attribute: false })
  private _fetchedVersion: any = null;

  // Loading state - synced from Redux
  @property({ type: Boolean, attribute: false })
  private _loading = false;

  // Track previous values for change detection
  private _prevTargetVersion = -1;
  private _prevGameVersion = 0;
  private _prevLastFetchedVersion = 0;

  // Computed properties - read directly from Redux selectors
  private get targetVersion(): number {
    return selectTargetVersion(store.getState());
  }

  private get gameVersion(): number {
    return selectCurrentVersion(store.getState());
  }

  private get lastFetchedVersion(): number {
    return selectLastFetchedVersion(store.getState());
  }

  private get socketActive(): boolean {
    return selectSocketConnected(store.getState());
  }

  private get _pendingBundles(): any[] {
    return selectPendingBundles(store.getState());
  }

  private get _lastFiredBundle(): any {
    return selectLastFiredBundle(store.getState());
  }

  override firstUpdated(_changedProperties: Map<PropertyKey, unknown>) {
    super.firstUpdated(_changedProperties);
    this.updateData();
  }

  stateChanged(state: RootState) {
    // Sync Redux loading state (non-duplicated, used for local logic)
    this._loading = selectGameLoading(state);

    // Sync Redux fetched data (non-duplicated, used for one-time processing)
    const prevFetchedInfo = this._fetchedInfo;
    const prevFetchedVersion = this._fetchedVersion;
    this._fetchedInfo = selectFetchedInfo(state);
    this._fetchedVersion = selectFetchedVersion(state);

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

    // Detect changes in computed properties and trigger handlers
    const currentTargetVersion = selectTargetVersion(state);
    const currentGameVersion = selectCurrentVersion(state);
    const currentLastFetchedVersion = selectLastFetchedVersion(state);

    // Handle targetVersion changes
    if (this._prevTargetVersion !== currentTargetVersion && currentTargetVersion >= 0) {
      this._handleTargetVersionChanged();
    }

    // Trigger requestUpdate if computed properties changed (for updated() lifecycle)
    if (this._prevTargetVersion !== currentTargetVersion ||
        this._prevGameVersion !== currentGameVersion ||
        this._prevLastFetchedVersion !== currentLastFetchedVersion) {
      this.requestUpdate();
    }

    // Update previous values for next change detection
    this._prevTargetVersion = currentTargetVersion;
    this._prevGameVersion = currentGameVersion;
    this._prevLastFetchedVersion = currentLastFetchedVersion;
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

    if (changedProperties.has('active') || changedProperties.has('_infoInstalled')) {
      this._socketUrl = this._computeSocketUrl(this.active, this._infoInstalled);
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
    const host = typeof (window as any).API_HOST !== 'undefined' ? (window as any).API_HOST : '';
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
    // Replaces _gameVersionPathChanged property watcher
    // Called explicitly from stateChanged when targetVersion changes

    if (this.targetVersion < 0) return;

    if (this.autoCurrentPlayer && this.requestedPlayer === this.viewingAsPlayer && this.targetVersion === this.gameVersion) {
      return;
    }

    // Skip if already have this version
    if (this.lastFetchedVersion === this.gameVersion && this.targetVersion === this.gameVersion) {
      return;
    }

    // Use Redux loading state instead of local flag
    if (this._loading) {
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
    const version = parseInt(e.data);
    if (isNaN(version)) {
      return;
    }
    store.dispatch(setTargetVersion(version));
  }

  private _socketError(e: Event) {
    console.warn('Socket error', e);
    store.dispatch(socketError(e.toString()));
  }

  private _socketOpened(e: Event) {
    store.dispatch(socketConnected());
  }

  private _socketClosed(e: CloseEvent) {
    console.warn('Socket closed', e);
    store.dispatch(socketDisconnected());
    // We always want a socket, so connect. Wait a bit so we don't just
    // busy spin if the server is down.

    // If we closed because we no longer have a valid URL, then
    // _connectSocket will just exit, and this loop won't be called.

    // TODO: exponential backoff on server connect.
    setTimeout(() => this._connectSocket(), 250);
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
    store.dispatch(setLastFetchedVersion(0));
    store.dispatch(setTargetVersion(-1));
    store.dispatch(setCurrentVersion(0));
    store.dispatch(clearStateBundles());
    this.softReset();
  }

  fetchInfo() {
    // Use Redux loading state instead of local flag
    if (this._loading) {
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

  private _prepareStateBundle(game: any, moveForms: any, viewingAsPlayer: number, move: any): any {
    const bundle: any = {};

    bundle.originalWallClockStartTime = Date.now();
    bundle.game = game;
    bundle.move = move;
    bundle.moveForms = this._expandMoveForms(moveForms);
    bundle.viewingAsPlayer = viewingAsPlayer;

    return bundle;
  }

  private _expandMoveForms(moveForms: any): any {
    if (!moveForms) return null;
    for (let i = 0; i < moveForms.length; i++) {
      const form = moveForms[i];
      // Some forms don't have fields and that's OK.
      if (!form.Fields) continue;
      for (let j = 0; j < form.Fields.length; j++) {
        const field = form.Fields[j];
        if (field.EnumName) {
          field.Enum = this.chest.Enums[field.EnumName];
        }
      }
    }
    return moveForms;
  }

  // Called when gameView tells us to pass up the next state if we have one
  // (the animations are done).
  readyForNextState() {
    this._scheduleNextStateBundle();
  }

  // A new state bundle has been enqueued. Ensure that we're working to fire a
  // state bundle. renderer might be a reference to the underlying renderer, or
  // null.
  private _scheduleNextStateBundle() {
    if (!this._pendingBundles.length) return;

    const renderer = this.activeRenderer;

    // If we were given a renderer that knows how to delay animations, consult it.
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
              this._scheduleNextStateBundle();
              return;
            }
          } else {
            this.dispatchEvent(new CustomEvent('set-animation-length', { composed: true, detail: length }));
          }
        }
        if (renderer.delayAnimation) {
          const delay = renderer.delayAnimation(lastMove, nextMove);
          if (delay < 0) {
            console.warn('Negative value for delayAnimation. Did you mean to use animationLength instead?', lastMove, nextMove);
          }
          // If delay is greater than 0, wait that long before firing
          if (delay > 0) {
            window.setTimeout(() => this._asyncFireNextStateBundle(), delay);
            return;
          }
        }
      }
    }

    this._asyncFireNextStateBundle();
  }

  private _asyncFireNextStateBundle() {
    // Not entirely sure why this has to be done this way, but it needs to be
    // done outside of the current task, even when fired from a timeout.
    window.requestAnimationFrame(() => this._fireNextStateBundle());
  }

  private _fireNextStateBundle() {
    // Called when the next state bundle should be installed NOW.
    // Dequeue from Redux and fire event
    if (this._pendingBundles.length > 0) {
      const bundle = this._pendingBundles[0];
      store.dispatch(dequeueStateBundle());
      this.dispatchEvent(new CustomEvent('install-state-bundle', { composed: true, detail: bundle }));
    }
  }

  // Add the next state bundle to the end
  private _enqueueStateBundle(bundle: any) {
    const wasEmpty = this._pendingBundles.length === 0;
    store.dispatch(enqueueStateBundle(bundle));
    // If that was the first one we added, go ahead and fire it right now.
    if (wasEmpty) this._scheduleNextStateBundle();
  }

  private _handleInfoData(data: any) {
    if (!data) {
      return;
    }

    this.chest = data.Chest;

    const gameInfo = {
      chest: data.Chest,
      playersInfo: data.Players,
      hasEmptySlots: data.HasEmptySlots,
      open: data.GameOpen,
      visible: data.GameVisible,
      isOwner: data.IsOwner,
    };

    this.dispatchEvent(new CustomEvent('install-game-static-info', { composed: true, detail: gameInfo }));

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

  private _handleVersionData(data: any) {
    if (!data) return;
    if (data.Error) {
      console.log('Version getter returned error: ' + data.Error);
      return;
    }

    let lastServerBundle: any = {};

    for (let i = 0; i < data.Bundles.length; i++) {
      const serverBundle = data.Bundles[i];
      const bundle = this._prepareStateBundle(serverBundle.Game, serverBundle.Forms, serverBundle.ViewingAsPlayer, serverBundle.Move);
      this._enqueueStateBundle(bundle);
      lastServerBundle = serverBundle;
    }

    if (lastServerBundle.Game) {
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
