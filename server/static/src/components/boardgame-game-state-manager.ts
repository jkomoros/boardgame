import { LitElement, html } from 'lit';
import { property, query } from 'lit/decorators.js';
import './boardgame-ajax.js';

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
class BoardgameGameStateManager extends LitElement {
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

  @property({ type: Number })
  targetVersion = -1;

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

  @property({ type: Number })
  gameVersion = 0;

  @property({ type: Number })
  lastFetchedVersion = 0;

  @property({ type: Object, attribute: false })
  infoData: any = null;

  @property({ type: Object, attribute: false })
  versionData: any = null;

  @property({ type: Boolean, attribute: true })
  socketActive = false;

  @property({ type: Object })
  activeRenderer: any = null;

  @property({ type: String, attribute: false })
  private _socketUrl = '';

  @property({ type: Boolean, attribute: false })
  private _infoInstalled = false;

  @property({ type: Object, attribute: false })
  private _socket: WebSocket | null = null;

  private _pendingStateBundles: any[] = [];
  private _lastFiredBundle: any = null;

  @query('#version')
  private _versionAjax?: any;

  @query('#info')
  private _infoAjax?: any;

  override firstUpdated(_changedProperties: Map<PropertyKey, unknown>) {
    super.firstUpdated(_changedProperties);
    this._pendingStateBundles = [];
    this._lastFiredBundle = null;
    this.updateData();
  }

  override updated(changedProperties: Map<PropertyKey, unknown>) {
    super.updated(changedProperties);

    // Handle computed properties
    if (changedProperties.has('active') ||
        changedProperties.has('requestedPlayer') ||
        changedProperties.has('admin') ||
        changedProperties.has('targetVersion') ||
        changedProperties.has('autoCurrentPlayer')) {
      this.gameVersionPath = this._computeGameVersionPath(
        this.active, this.requestedPlayer, this.admin, this.targetVersion, this.autoCurrentPlayer
      );
    }

    if (changedProperties.has('requestedPlayer') ||
        changedProperties.has('admin') ||
        changedProperties.has('lastFetchedVersion')) {
      this.gameViewPath = this._computeGameViewPath(this.requestedPlayer, this.admin, this.lastFetchedVersion);
    }

    if (changedProperties.has('gameVersionPath') ||
        changedProperties.has('lastFetchedVersion') ||
        changedProperties.has('gameVersion')) {
      this.effectiveGameVersionPath = this._computeEffectiveGameVersionPath(
        this.gameVersionPath, this.lastFetchedVersion, this.gameVersion
      );
    }

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

    if (changedProperties.has('gameVersionPath')) {
      this._gameVersionPathChanged(this.gameVersionPath, changedProperties.get('gameVersionPath') as string);
    }

    if (changedProperties.has('_socketUrl')) {
      this._socketUrlChanged(this._socketUrl);
    }

    if (changedProperties.has('infoData')) {
      this._infoDataChanged(this.infoData, changedProperties.get('infoData') as any);
    }

    if (changedProperties.has('versionData')) {
      this._versionDataChanged(this.versionData);
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
    if (!this._versionAjax) return '';
    let result = this._versionAjax.gameBasePath + 'socket';
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

  private _gameVersionPathChanged(newValue: string, oldValue: string) {
    if (!newValue) return;

    if (this.autoCurrentPlayer && this.requestedPlayer === this.viewingAsPlayer && this.targetVersion === this.gameVersion) {
      return;
    }

    // TODO: the autoCurrent player stuff has to be done here...
    requestAnimationFrame(() => this._versionAjax?.generateRequest());
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
    this.targetVersion = version;
  }

  private _socketError(e: Event) {
    // TODO: do something more substantive
    console.warn('Socket error', e);
  }

  private _socketOpened(e: Event) {
    this.socketActive = true;
  }

  private _socketClosed(e: CloseEvent) {
    console.warn('Socket closed', e);
    this.socketActive = false;
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
    this.infoData = null;
    this._infoInstalled = false;
    window.requestAnimationFrame(() => this.updateData());
  }

  // When everything should be reset
  reset() {
    this.lastFetchedVersion = 0;
    this.targetVersion = -1;
    this._resetPendingStateBundles();
    this.softReset();
  }

  fetchInfo() {
    if (this._infoAjax?.loading) {
      return;
    }

    if (!this.active) {
      return;
    }

    if (!this.gameRoute) {
      // The URL will be junk
      return;
    }
    this._infoAjax?.generateRequest();
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
    if (!this._pendingStateBundles.length) return;

    const renderer = this.activeRenderer;

    // If we were given a renderer that knows how to delay animations, consult it.
    if (renderer) {
      const nextBundle = this._pendingStateBundles[0];
      const lastBundle = this._lastFiredBundle;
      const nextMove = nextBundle ? nextBundle.move : null;
      const lastMove = lastBundle ? lastBundle.move : null;
      if (nextMove || lastMove) {
        if (renderer.animationLength) {
          const length = renderer.animationLength(lastMove, nextMove);
          // If the length is negative, that's the signal to skip binding this one.
          if (length < 0) {
            // We always render the last bundle to install
            if (this._pendingStateBundles.length > 1) {
              // Skip this bundle.
              this._lastFiredBundle = this._pendingStateBundles.shift();
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

  private _resetPendingStateBundles() {
    this._pendingStateBundles = [];
  }

  private _fireNextStateBundle() {
    // Called when the next state bundle should be installed NOW.
    const bundle = this._pendingStateBundles.shift();
    if (bundle) {
      this._lastFiredBundle = bundle;
      this.dispatchEvent(new CustomEvent('install-state-bundle', { composed: true, detail: bundle }));
    }
  }

  // Add the next state bundle to the end
  private _enqueueStateBundle(bundle: any) {
    this._pendingStateBundles.push(bundle);
    // If that was the first one we added, go ahead and fire it right now.
    if (this._pendingStateBundles.length === 1) this._scheduleNextStateBundle();
  }

  private _infoDataChanged(newValue: any, oldValue: any) {
    if (!newValue) {
      // Sometimes we set null, like when we select the view.
      return;
    }

    this.chest = newValue.Chest;

    const gameInfo = {
      chest: newValue.Chest,
      playersInfo: newValue.Players,
      hasEmptySlots: newValue.HasEmptySlots,
      open: newValue.GameOpen,
      visible: newValue.GameVisible,
      isOwner: newValue.IsOwner,
    };

    this.dispatchEvent(new CustomEvent('install-game-static-info', { composed: true, detail: gameInfo }));

    const bundle = this._prepareStateBundle(newValue.Game, newValue.Forms, newValue.ViewingAsPlayer, null);
    this._enqueueStateBundle(bundle);

    this._infoInstalled = true;

    // We don't use newValue.Game.Version, because in some cases the current
    // state we're returning is not actually current state, but an old one to
    // force us to play animations for moves that are made before a player move
    // is. The server ships down this information in a special field.
    this.lastFetchedVersion = newValue.StateVersion;
    this.targetVersion = newValue.Game.Version;
  }

  private _versionDataChanged(newValue: any) {
    if (!newValue) return;
    if (newValue.Error) {
      console.log('Version getter returned error: ' + newValue.Error);
      return;
    }

    let lastServerBundle: any = {};

    for (let i = 0; i < newValue.Bundles.length; i++) {
      const serverBundle = newValue.Bundles[i];
      const bundle = this._prepareStateBundle(serverBundle.Game, serverBundle.Forms, serverBundle.ViewingAsPlayer, serverBundle.Move);
      this._enqueueStateBundle(bundle);
      lastServerBundle = serverBundle;
    }

    this.lastFetchedVersion = lastServerBundle.Game.Version;
  }

  override render() {
    return html`
      <boardgame-ajax
        id="version"
        .gamePath="${this.effectiveGameVersionPath}"
        .gameRoute="${this.gameRoute}"
        handle-as="json"
        .lastResponse="${this.versionData}"
        @last-response-changed="${(e: CustomEvent) => { this.versionData = e.detail.value; }}">
      </boardgame-ajax>
      <boardgame-ajax
        id="info"
        .gamePath="${this.gameViewPath}"
        .gameRoute="${this.gameRoute}"
        handle-as="json"
        .lastResponse="${this.infoData}"
        @last-response-changed="${(e: CustomEvent) => { this.infoData = e.detail.value; }}">
      </boardgame-ajax>
    `;
  }
}

customElements.define('boardgame-game-state-manager', BoardgameGameStateManager);

export { BoardgameGameStateManager };
