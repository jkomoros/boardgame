import { LitElement, html, css } from 'lit';
import { customElement, property, query } from 'lit/decorators.js';
import './boardgame-player-roster.js';
import './boardgame-gathering-panel.js';
import './boardgame-chat-panel.js';
import './boardgame-render-game.js';
import './boardgame-admin-controls.js';
import './boardgame-game-state-manager.js';
import type { BoardgamePlayerRoster } from './boardgame-player-roster.js';
import type { BoardgameRenderGame, HostedGameRenderer } from './boardgame-render-game.js';
import type { BoardgameAdminControls } from './boardgame-admin-controls.js';
import type { BoardgameGameStateManager } from './boardgame-game-state-manager.js';
import { sharedStyles } from './shared-styles-lit.js';
import { warnOnInvalidMoveArgs } from '../utils/move-validation.js';
import { surfaceForGame } from '../utils/companion-surface.js';
import {
  playerPresentations,
  type PlayerPresentation,
} from '../status/player-presentation.js';
import type {
  ExpandedGameState,
  GameChest,
  PlayerInfo,
  RootState,
} from '../types/store';

import { connect } from 'pwa-helpers/connect-mixin.js';
import { store } from '../store.js';

import {
  selectPage,
  selectPageExtra,
  selectGameRoute,
  selectLoggedIn,
  selectAdmin,
  selectGameChest,
  selectGamePlayersInfo,
  selectGameHasEmptySlots,
  selectGameOpen,
  selectGameVisible,
  selectGameIsOwner,
  selectGameCompanionInfo,
  selectExpandedGameState,
  selectGame,
  selectViewingAsPlayer,
  selectRequestedPlayer,
  selectAutoCurrentPlayer,
  selectMoveForms,
  selectLastFetchedVersion,
  selectPlayerColors,
  selectPlayerActivity,
  selectPlayerOrder,
  selectGameTimerInfos,
  selectSocketConnectionAttempts,
  selectSocketError,
} from '../selectors.js';

import {
  PAGE_GAME
} from '../actions/app.js';

import {
  updateGameRoute,
  updateGameStaticInfo,
  installGameState,
  updateViewState,
  setRequestedPlayer,
  setAutoCurrentPlayer,
  fetchGameInfo,
  submitMove,
} from '../actions/game.js';
import {
  MoveSubmissionGate,
  type MovePreviewTransport,
  type MoveTransport,
} from '../moves/action.js';
import type { TargetPreviewTransport } from '../moves/target-action.js';
import { movePreview, movePreviewBatch } from '../api.js';
import {
  TIMER_SERVICE_REQUEST_EVENT,
  TimerService,
  type TimerServiceRequestDetail,
} from '../timers/timer-service.js';

import type { GameFromServer, StateBundle } from '../types/game-state';
import type { MoveForm } from '../types/api';

import game from '../reducers/game.js';
store.addReducers({
  game
});

@customElement('boardgame-game-view')
export class BoardgameGameView extends connect(store)(LitElement) {
  static override styles = [
    sharedStyles,
    css`
      :host {
        display: block;
        --animation-length: 0.5s;
      }

      [hidden] {
        display: none !important;
      }

      /* Hide-my-hand privacy shield (hand surface only). Fixed positioning
         so the shield covers the entire viewport — app chrome included —
         because the threat is a glance at the whole phone screen, not just
         the renderer area. */
      .privacy-toggle {
        position: fixed;
        top: 12px;
        right: 12px;
        z-index: 1000;
        padding: 8px 14px;
        font-size: 14px;
        font-weight: 600;
        border-radius: 20px;
        border: none;
        background: rgba(0, 0, 0, 0.55);
        color: white;
        cursor: pointer;
      }
      .privacy-shield {
        position: fixed;
        inset: 0;
        z-index: 999;
        background: #1a2b3c;
        display: flex;
        flex-direction: column;
        align-items: center;
        justify-content: center;
        gap: 16px;
        color: white;
        text-align: center;
        padding: 24px;
      }
      .privacy-shield .shield-glyph {
        font-size: 64px;
      }
      .privacy-shield p {
        margin: 0;
        font-size: 18px;
        opacity: 0.8;
      }
      .privacy-shield button {
        padding: 16px 32px;
        font-size: 18px;
        font-weight: 600;
        border-radius: 12px;
        border: 2px solid white;
        background: transparent;
        color: white;
        cursor: pointer;
      }

      #moves > details {
        margin-left: 1em;
      }

      .admin > div:first-child {
        margin-left: 0;
      }

      .admin > div {
        margin-left: 1em;
      }

      .card {
        position: relative;
      }
    `
  ];

  // View state - synced from Redux
  @property({ type: Number, attribute: false })
  requestedPlayer = 0;

  @property({ type: Object, attribute: false })
  game: GameFromServer | null = null;

  @property({ type: Number, attribute: false })
  viewingAsPlayer = 0;

  @property({ type: Boolean, attribute: false })
  autoCurrentPlayer = false;

  @property({ type: Object, attribute: false })
  moveForms: MoveForm[] | null = null;

  @property({ type: Boolean })
  selected = false;

  @property({ type: Boolean })
  promptedToJoin = false;

  // The current renderer, passed up from the gameRenderer, so we can pass
  // it to stateGameManager and readyForNextState.
  @property({ type: Object })
  activeRenderer: HostedGameRenderer | null = null;

  @property({ type: Boolean })
  socketActive = false;

  @property({ type: Number, attribute: false })
  _socketConnectionAttempts = 0;

  @property({ type: String, attribute: false })
  _socketError: string | null = null;

  @property({ type: Boolean })
  _firstStateBundle = true;

  @query('#manager')
  private _managerEle?: BoardgameGameStateManager;

  @query('#admin')
  private _adminEle?: BoardgameAdminControls;

  @query('#render')
  private _renderEle?: BoardgameRenderGame;

  @query('#player')
  private _playerEle?: BoardgamePlayerRoster;

  // Reactive properties - synced from Redux in stateChanged()
  @property({ type: Object, attribute: false })
  _currentState: ExpandedGameState | null = null;

  @property({ type: Object, attribute: false })
  _animationContext: import('./companion-sync.js').VersionAnimationContext | null = null;

  @property({ type: Object, attribute: false })
  _chest: GameChest | null = null;

  @property({ type: Array, attribute: false })
  _playersInfo: PlayerInfo[] = [];

  @property({ type: Boolean, attribute: false })
  _hasEmptySlots = false;

  @property({ type: Boolean, attribute: false })
  _open = false;

  @property({ type: Boolean, attribute: false })
  _visible = false;

  @property({ type: Boolean, attribute: false })
  _isOwner = false;

  @property({ type: Object, attribute: false })
  _companionInfo: import('../types/store').CompanionInfo | null = null;

	@property({ type: String, attribute: false })
  _moveInputSchemaFingerprint: string | null = null;

  @property({ type: Number, attribute: false })
  _moveSnapshotEpoch = 0;

  @property({ type: Number, attribute: false })
  _proposingAsPlayer = 0;

  private readonly _moveSubmissionGate = new MoveSubmissionGate();
  private readonly _moveTransport: MoveTransport = {
    submit: async request => {
      const route = this._gameRoute;
      if (!route) {
        return { kind: 'network-failure', error: 'The game route is unavailable' };
      }
      return submitMove(route, {
        ...request.arguments,
        MoveType: request.name,
        admin: request.proposingAsAdmin ? '1' : '0',
        player: String(request.proposingAsPlayer),
        ExpectedVersion: String(request.snapshotVersion),
      })(store.dispatch);
    },
  };
  private readonly _movePreviewTransport: MovePreviewTransport = {
    preview: async request => {
      const route = this._gameRoute;
      if (!route) {
        return { kind: 'failure', error: 'The game route is unavailable', retryable: false };
      }
      const response = await movePreview(
        route.name,
        route.id,
        request.name,
        { ...request.arguments, ExpectedVersion: String(request.snapshotVersion) },
        { player: request.proposingAsPlayer, admin: request.proposingAsAdmin ? 1 : 0 },
        request.signal,
      );
      if (response.error) {
        if (response.code === 'STALE_SNAPSHOT') {
          return {
            kind: 'stale-snapshot',
            expectedVersion: response.expectedVersion ?? request.snapshotVersion,
            actualVersion: response.actualVersion ?? this._lastFetchedVersion,
          };
        }
        return {
          kind: 'failure',
          error: response.error,
          friendlyError: response.friendlyError,
          retryable: response.status === 0,
        };
      }
      const form = response.data?.Form;
      if (!form) return { kind: 'failure', error: 'Move preview returned no form', retryable: true };
      return {
        kind: 'success',
        legal: form.LegalForPlayer ?? false,
        ...(form.LegalForPlayerError ? { error: form.LegalForPlayerError } : {}),
        ...(form.Preconditions ? { preconditions: form.Preconditions } : {}),
      };
    },
  };

  private readonly _targetPreviewTransport: TargetPreviewTransport = {
    previewTargets: async request => {
      const route = this._gameRoute;
      if (!route) return { kind: 'failure', error: 'The game route is unavailable', retryable: false };
      const response = await movePreviewBatch(
        route.name,
        route.id,
        request.name,
        request.candidates.map(candidate => ({ ID: candidate.id, Args: { ...candidate.arguments } })),
        { player: request.proposingAsPlayer, admin: request.proposingAsAdmin ? 1 : 0 },
        request.snapshotVersion,
        request.signal,
      );
      if (response.error) {
        if (response.code === 'STALE_SNAPSHOT') {
          return {
            kind: 'stale-snapshot',
            expectedVersion: response.expectedVersion ?? request.snapshotVersion,
            actualVersion: response.actualVersion ?? this._lastFetchedVersion,
          };
        }
        return {
          kind: 'failure',
          error: response.error,
          friendlyError: response.friendlyError,
          retryable: response.status === 0,
        };
      }
      const results: unknown = response.data?.Results;
      if (!Array.isArray(results)) {
        return { kind: 'failure', error: 'Target preview results must be an array', retryable: false };
      }
      const validated = [];
      for (const result of results) {
        if (typeof result !== 'object' || result === null) {
          return { kind: 'failure', error: 'Target preview returned a malformed result', retryable: false };
        }
        const item = result as Record<string, unknown>;
        if (typeof item['ID'] !== 'string' || !item['ID']
          || typeof item['Legal'] !== 'boolean'
          || (item['Error'] !== undefined && typeof item['Error'] !== 'string')) {
          return { kind: 'failure', error: 'Target preview returned a malformed result', retryable: false };
        }
        validated.push({
          id: item['ID'],
          legal: item['Legal'],
          ...(item['Error'] ? { error: item['Error'] } : {}),
        });
      }
      return {
        kind: 'success',
        results: validated,
      };
    },
  };

  @property({ type: String, attribute: false })
  _pageExtra = '';

  @property({ type: Object, attribute: false })
  _gameRoute: { id: string; name: string } | null = null;

  @property({ type: Boolean, attribute: false })
  _loggedIn = false;

  @property({ type: Boolean, attribute: false })
  _admin = false;

  @property({ type: String, attribute: false })
  _page = '';

  @property({ type: Number, attribute: false })
  _lastFetchedVersion = 0;

  @property({ type: Array, attribute: false })
  _playerColors: string[] = [];

  @property({ attribute: false })
  _playerPresentations: readonly PlayerPresentation[] = Object.freeze([]);

  private _presentationPlayersSource: readonly PlayerInfo[] | null = null;
  private _presentationColorsSource: readonly string[] | null = null;

  @property({ type: Array, attribute: false })
  _playerActivity: boolean[] = [];

  @property({ type: Array, attribute: false })
  _playerOrder: number[] | null = null;

  // The active companion surface ('table' | 'hand' | null), derived once
  // per game-route change in stateChanged — render() runs far too often
  // (every state bundle; every animation-frame tick while timers run) to
  // re-parse the query string + cookie jar each time.
  @property({ type: String, attribute: false })
  _companionSurface: 'table' | 'hand' | null = null;

  private _surfaceCachedGameId: string | null = null;
  private readonly _timerService = new TimerService();

  // Hide-my-hand privacy shield (hand surface only): when true, an opaque
  // full-viewport overlay covers the private hand so the player can set
  // the phone down or step away without shoulder-surfing risk. Purely
  // client-side and per-tab — game state keeps flowing underneath, so
  // revealing is instant and never misses an update.
  @property({ type: Boolean, attribute: false })
  _handHidden = false;

  // Mirrors boardgame-render-game's isAnimating (via the animating-changed
  // event, since #render is a plain @query reference, not a reactive
  // property source) so it can be threaded down to the admin move-form for
  // move auto-disable (#721).
  @property({ type: Boolean, attribute: false })
  _animating = false;

  constructor() {
    super();

    this.addEventListener('propose-move', (e: Event) => this._handleProposeMove(e as CustomEvent));
    this.addEventListener('refresh-info', (e: Event) => this._handleRefreshData(e));
    this.addEventListener('install-state-bundle', (e: Event) => this._handleStateBundle(e as CustomEvent));
    this.addEventListener('install-game-static-info', (e: Event) => this._handleGameStaticInfo(e as CustomEvent));
    this.addEventListener('all-animations-done', (e: Event) => this._handleAllAnimationsDone(e));
    this.addEventListener('set-animation-length', (e: Event) => this._handleSetAnimationLength(e as CustomEvent));
    this.addEventListener('animating-changed', (e: Event) => this._handleAnimatingChanged(e as CustomEvent));
    this.addEventListener(TIMER_SERVICE_REQUEST_EVENT, (event: Event) => {
      const request = event as CustomEvent<TimerServiceRequestDetail>;
      if (typeof request.detail?.accept !== 'function') {
        throw new Error('boardgame-game-view: malformed timer service request');
      }
      event.stopPropagation();
      request.detail.accept(this._timerService);
    });
  }

  override render() {
    // On companion surfaces the renderer owns the screen: the projector
    // (table) and phones (hand) hide the solo-flow chrome — roster with its
    // join buttons, admin controls, chat. Seating happens via the phone
    // join flow and identity lives on the avatar strip, so that chrome is
    // at best redundant and at worst contradicts the companion model. The
    // gathering panel stays: "Waiting for N more players" is useful on
    // both surfaces.
    const companionSurface = this._companionSurface;
    return html`
      ${companionSurface === 'hand' ? html`
        ${this._handHidden ? html`
          <div class="privacy-shield">
            <div class="shield-glyph">🙈</div>
            <p>Your hand is hidden.</p>
            <button @click=${() => { this._handHidden = false; }}>Show my hand</button>
          </div>
        ` : html`
          <button class="privacy-toggle" @click=${() => { this._handHidden = true; }}>🙈 Hide my hand</button>
        `}
      ` : ''}
      ${companionSurface ? '' : html`
      <div class="card">
        <boardgame-player-roster
          id="player"
          .loggedIn=${this._loggedIn}
          .gameRoute=${this._gameRoute}
          .viewingAsPlayer=${this.viewingAsPlayer}
          .hasEmptySlots=${this._hasEmptySlots}
          .gameOpen=${this._open}
          .gameVisible=${this._visible}
          .currentPlayerIndex=${this.game ? this.game.CurrentPlayerIndex : 0}
          .playersInfo=${this._playersInfo}
          .state=${this._currentState}
          .finished=${this.game ? this.game.Finished : false}
          .winners=${this.game ? this.game.Winners : []}
          .admin=${this._admin}
          .isOwner=${this._isOwner}
          .playerColors=${this._playerColors}
          .playerActivity=${this._playerActivity}
          .playerOrder=${this._playerOrder}
          .active=${this.selected}>
        </boardgame-player-roster>
      </div>
      `}
      <boardgame-gathering-panel
        .moveForms=${this.moveForms}
        .state=${this._currentState}
        .viewingAsPlayer=${this.viewingAsPlayer}
        .hasEmptySlots=${this._hasEmptySlots}
        .gameOpen=${this._open}
        .finished=${this.game ? this.game.Finished : false}
        .gameRoute=${this._gameRoute}
        .playersInfo=${this._playersInfo}
        .companionRoomCode=${this._companionInfo?.RoomCode || ''}>
      </boardgame-gathering-panel>
      <div class="card">
        <boardgame-render-game
          id="render"
          .state=${this._currentState}
          .animationContext=${this._animationContext}
          .diagram=${this.game ? this.game.Diagram : ''}
          .renderer=${this.activeRenderer}
          @renderer-changed=${this._handleRendererChanged}
          .gameName=${this._gameRoute ? this._gameRoute.name : ''}
          .gameId=${this._gameRoute ? this._gameRoute.id : ''}
          .gameVersion=${this.game ? this.game.Version : 0}
          .snapshotEpoch=${this._moveSnapshotEpoch}
          .proposingAsPlayer=${this._proposingAsPlayer}
          .proposingAsAdmin=${this._admin}
          .moveTransport=${this._moveTransport}
          .movePreviewTransport=${this._movePreviewTransport}
          .targetPreviewTransport=${this._targetPreviewTransport}
          .moveSubmissionGate=${this._moveSubmissionGate}
          .companionInfo=${this._companionInfo}
          .isOwner=${this._isOwner}
          .gameFinished=${this.game ? this.game.Finished : false}
          .gameWinners=${this.game ? this.game.Winners || [] : []}
          .playerPresentations=${this._playerPresentations}
          .viewingAsPlayer=${this.viewingAsPlayer}
          .currentPlayerIndex=${this.game ? this.game.CurrentPlayerIndex : 0}
          .previewAsPlayer=${this.requestedPlayer}
          .previewAsAdmin=${this._admin}
          .socketActive=${this.socketActive}
          .connectionAttempts=${this._socketConnectionAttempts}
          .connectionError=${this._socketError}
          @retry-connection=${this._handleRetryConnection}
          .active=${this.selected}
          .chest=${this._chest}
					.moveForms=${this.moveForms}
					.moveInputSchemaFingerprint=${this._moveInputSchemaFingerprint}>
        </boardgame-render-game>
      </div>
      <!-- Not chrome despite the name: admin-controls owns the move-form
           submission pipeline (_handleProposeMove forwards every proposed
           move through it), so it must exist on companion surfaces too.
           It renders nothing visible unless admin mode is active. -->
      <boardgame-admin-controls
        id="admin"
        .active=${this._admin}
        .game=${this.game}
        .viewingAsPlayer=${this.viewingAsPlayer}
        .moveForms=${this.moveForms}
        .gameRoute=${this._gameRoute}
        .chest=${this._chest}
        .currentState=${this._currentState}
        .animating=${this._animating}
        @requested-player-changed=${this._handleRequestedPlayerChanged}
        @auto-current-player-changed=${this._handleAutoCurrentPlayerChanged}
        @move-as-player-changed=${this._handleMoveAsPlayerChanged}>
      </boardgame-admin-controls>
      ${companionSurface ? '' : html`
      <boardgame-chat-panel
        .gameRoute=${this._gameRoute}
        .viewingAsPlayer=${this.viewingAsPlayer}
        .playersInfo=${this._playersInfo}>
      </boardgame-chat-panel>
      `}
      <boardgame-game-state-manager
        id="manager"
        .activeRenderer=${this.activeRenderer}
        .gameRoute=${this._gameRoute}
        .requestedPlayer=${this.requestedPlayer}
        .active=${this.selected}
        .admin=${this._admin}
        .gameFinished=${this.game ? this.game.Finished : false}
        .gameVersion=${this.game ? this.game.Version : 0}
        .loggedIn=${this._loggedIn}
        .autoCurrentPlayer=${this.autoCurrentPlayer}
        .viewingAsPlayer=${this.viewingAsPlayer}
        .socketActive=${this.socketActive}
        @socket-active-changed=${this._handleSocketActiveChanged}>
      </boardgame-game-state-manager>
    `;
  }

  // TODO: shouldUpdate should return false if selected is false. But if we do
  // that, then game-state-manager is never updated, so it never learns that
  // there was a time when it wasn't active. Once game-state-manager is done as
  // action creators then it should be fine.

  stateChanged(state: RootState) {
    this._timerService.update(selectGameTimerInfos(state));
    // Sync view state from Redux
    this.game = selectGame(state);
    this.viewingAsPlayer = selectViewingAsPlayer(state);
    this._admin = selectAdmin(state);
    if (!this._admin || this._adminEle?.makeMovesAsViewingAsPlayer !== false) {
      this._proposingAsPlayer = this.viewingAsPlayer;
    }
    this.requestedPlayer = selectRequestedPlayer(state);
    this.autoCurrentPlayer = selectAutoCurrentPlayer(state);
    this.moveForms = selectMoveForms(state);

    // Sync properties that were previously getters
    this._currentState = selectExpandedGameState(state);
    this._chest = selectGameChest(state);
    this._playersInfo = selectGamePlayersInfo(state);
    this._hasEmptySlots = selectGameHasEmptySlots(state);
    this._open = selectGameOpen(state);
    this._visible = selectGameVisible(state);
    this._isOwner = selectGameIsOwner(state);
    this._companionInfo = selectGameCompanionInfo(state);
    this._pageExtra = selectPageExtra(state);
    this._gameRoute = selectGameRoute(state);
    const surfaceGameId = this._gameRoute ? this._gameRoute.id : null;
    if (surfaceGameId !== this._surfaceCachedGameId) {
      this._surfaceCachedGameId = surfaceGameId;
      this._companionSurface = surfaceGameId ? surfaceForGame(surfaceGameId) : null;
    }
    this._loggedIn = selectLoggedIn(state);
    this._page = selectPage(state);
    this._lastFetchedVersion = selectLastFetchedVersion(state);
    this._playerColors = selectPlayerColors(state);
    if (this._presentationPlayersSource !== this._playersInfo
      || this._presentationColorsSource !== this._playerColors) {
      this._presentationPlayersSource = this._playersInfo;
      this._presentationColorsSource = this._playerColors;
      this._playerPresentations = playerPresentations(this._playersInfo, this._playerColors);
    }
    this._playerActivity = selectPlayerActivity(state);
    this._playerOrder = selectPlayerOrder(state);
    this._socketConnectionAttempts = selectSocketConnectionAttempts(state);
    this._socketError = selectSocketError(state);
  }

  private _handleRefreshData(e: Event) {
    // Dispatch Redux action directly instead of calling component method
    const gameRoute = this._gameRoute;
    const requestedPlayer = this.requestedPlayer;
    const admin = this._admin;
    const lastFetchedVersion = this._lastFetchedVersion;

    if (gameRoute) {
      store.dispatch(fetchGameInfo(gameRoute, requestedPlayer, admin, lastFetchedVersion));
    }
  }

  private _handleRequestedPlayerChanged(e: CustomEvent) {
    store.dispatch(setRequestedPlayer(e.detail.value));
  }

  private _handleAutoCurrentPlayerChanged(e: CustomEvent) {
    store.dispatch(setAutoCurrentPlayer(e.detail.value));
  }

  private _handleMoveAsPlayerChanged(e: CustomEvent) {
    this._proposingAsPlayer = e.detail.value;
  }

  private _handleSocketActiveChanged(e: CustomEvent) {
    this.socketActive = e.detail.value;
  }

  private _handleRetryConnection(e: Event) {
    e.stopPropagation();
    this._managerEle?.retryConnection();
  }

  private _handleRendererChanged(e: CustomEvent) {
    this.activeRenderer = e.detail.value;
  }

  private _handleProposeMove(e: CustomEvent) {
    // Swallow proposed moves while an animation cycle is in flight: the
    // rendered state on screen is mid-transition to what the server already
    // considers current, so a move proposed now would be judged against
    // stale-looking state from the user's perspective (#721). This also
    // guards the classic "double-click a move button" case with the default
    // animation length — the second click lands while isAnimating is still
    // true and is dropped rather than silently enqueuing a second move.
    // The gate re-opens either when animations finish normally or when the
    // 4s watchdog force-closes it (boardgame-render-game's
    // _notifyAnimationsDone), so this never permanently wedges move entry.
    if (this._renderEle?.isAnimating) {
      console.warn('[game-view] propose-move ignored while animations are running (#721)');
      return;
    }

    // Validate arguments against the move schema (dev-time safety net)
    warnOnInvalidMoveArgs(e.detail.name, e.detail.arguments || {}, this.moveForms);

    // Forward the propose-move event to the admin controls element
    // The admin element will handle it and forward to the move form
    if (this._adminEle) {
      this._adminEle.dispatchEvent(new CustomEvent('propose-move', {
        detail: { name: e.detail.name, arguments: e.detail.arguments },
        bubbles: true,
        composed: true
      }));
    }
  }

  override updated(changedProps: Map<PropertyKey, unknown>) {
    super.updated(changedProps);

    // Set CSS custom properties for player colors so game renderers can use them
    if (changedProps.has('_playerColors')) {
      // Remove stale properties from previous game (e.g. switching from 6-player to 2-player)
      const oldColors = changedProps.get('_playerColors') as string[] | undefined;
      const oldLen = oldColors?.length ?? 0;
      for (let i = this._playerColors.length; i < oldLen; i++) {
        this.style.removeProperty(`--player-${i}-color`);
      }
      this._playerColors.forEach((color, i) => {
        if (color) {
          this.style.setProperty(`--player-${i}-color`, color);
        }
      });
    }

    if (changedProps.has('_pageExtra') && this._page === PAGE_GAME) {
      store.dispatch(updateGameRoute(this._pageExtra));
    }
    if (changedProps.has('selected') && !this.selected) {
      this._resetState();
    }
    if (changedProps.has('_gameRoute')) {
      // reset this so the next time we get data set and notice that we COULD
      // login we prompt for it.
      this.promptedToJoin = false;
      this._resetState();
    }
  }

  private _handleStateBundle(e: CustomEvent) {
    this._installStateBundle(e.detail);
  }

  private _handleGameStaticInfo(e: CustomEvent) {
    const bundle = e.detail;
		this._moveInputSchemaFingerprint = bundle.moveInputSchemaFingerprint ?? null;
    store.dispatch(updateGameStaticInfo(bundle.chest, bundle.playersInfo, bundle.hasEmptySlots, bundle.open, bundle.visible, bundle.isOwner, bundle.companionInfo));
  }

  private _handleAllAnimationsDone(e: Event) {
    // Dispatch custom event for animation coordination
    // The manager element will listen for this and handle it
    if (this._managerEle) {
      this._managerEle.dispatchEvent(new CustomEvent('ready-for-next-state', {
        bubbles: true,
        composed: true
      }));
    }
  }

  private _handleSetAnimationLength(e: CustomEvent) {
    if (!this._renderEle) {
      throw new Error('boardgame-game-view: animation length arrived before the renderer host mounted');
    }
    this._renderEle.defaultAnimationLength = e.detail;
  }

  private _handleAnimatingChanged(e: CustomEvent) {
    this._animating = e.detail.value;
  }

  private _firstStateBundleInstalled() {
    // No roster on companion surfaces (@query returns null) — and no
    // join prompt either: phones join via the room code, not this dialog.
    if (!this._playerEle) return;
    if (this.selected && this._loggedIn && this._playerEle.showJoin && !this.promptedToJoin) {
      // Take note that we already prompted them, and don't prompt again unless the game changes.
      this.promptedToJoin = true;
      // Prompt the user to join!
      this._playerEle.openJoinDialog();
    }
  }

  private _resetState() {
    // Clear stale CSS custom properties for player colors
    for (let i = 0; i < this._playerColors.length; i++) {
      this.style.removeProperty(`--player-${i}-color`);
    }
    // Reset view state properties only
    // Computed properties (_currentState, _chest, etc.) are read from Redux selectors
    this.game = null;
    this.moveForms = null;
    this.viewingAsPlayer = 0;
    this._moveInputSchemaFingerprint = null;
    this._animationContext = null;
    this._moveSnapshotEpoch += 1;
    this._firstStateBundle = true;
  }

  private _installStateBundle(bundle: StateBundle) {
    // Set the version's animation context before publishing its state. The
    // render-game wrapper applies it to the shared animator before assigning
    // the new state to the game renderer.
    this._animationContext = bundle.animationContext ?? null;
    this._moveSnapshotEpoch += 1;
    store.dispatch(installGameState(bundle.game.CurrentState, bundle.game.ActiveTimers, bundle.originalWallClockStartTime));

    // Update view state in Redux (replaces direct property assignment)
    store.dispatch(updateViewState(bundle.game, bundle.viewingAsPlayer, bundle.moveForms));

    if (this._firstStateBundle) {
      this._firstStateBundleInstalled();
    }
    this._firstStateBundle = false;
  }
}

declare global {
  interface HTMLElementTagNameMap {
    'boardgame-game-view': BoardgameGameView;
  }
}
