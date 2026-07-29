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
import type { BoardgameAnimatableItem } from './boardgame-animatable-item.js';
import type { BoardgameAdminControls } from './boardgame-admin-controls.js';
import type { BoardgameGameStateManager } from './boardgame-game-state-manager.js';
import { sharedStyles } from './shared-styles-lit.js';
import { warnOnInvalidMoveArgs } from '../utils/move-validation.js';
import {
  forgetSurfaceForGame,
  rememberSurfaceForGame,
  surfaceForGame,
  tableRecoveryDeviceID,
} from '../utils/companion-surface.js';
import { apiHttpPost, buildGameUrl } from '../api.js';
import {
  decodeTableLeaseAcquireResponse,
  isTableLeaseFailureCode,
  tableLeaseFailureMessage,
} from '../types/table-lease-response.js';
import {
  decodeTableTransferCancel,
  decodeTableTransferOffer,
  transferFailureMessage,
  type TableTransferOffer,
} from '../table-transfer/table-transfer.js';
import { decodeRematchResponse, type RematchResponse } from '../types/rematch-response.js';
import { gamePath } from '../util.js';
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
import type { ClientMove, MoveForm, ProjectedMoveChoicesWire } from '../types/api';

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

      .table-session-notice {
        box-sizing: border-box;
        margin: 12px;
        padding: 12px 16px;
        border: 2px solid #526475;
        border-radius: 10px;
        background: #f4f7fa;
        color: #17212b;
      }
      .table-session-notice p {
        margin: 0 0 10px;
      }
      .table-session-notice p:last-child {
        margin-bottom: 0;
      }
      .table-session-notice button {
        min-height: 44px;
        padding: 8px 16px;
        font: inherit;
        font-weight: 600;
        cursor: pointer;
      }
      .table-session-notice .table-session-error {
        color: #a11616;
        font-weight: 600;
      }
      .table-session-terminal {
        position: fixed;
        inset: 0;
        z-index: 1100;
        box-sizing: border-box;
        display: grid;
        place-items: center;
        padding: 24px;
        background: #17212b;
        color: white;
        text-align: center;
      }
      .table-session-terminal > section {
        max-width: 520px;
      }
      .table-session-terminal h1 {
        margin: 0 0 12px;
        font-size: clamp(28px, 5vw, 48px);
      }
      .table-session-terminal p {
        margin: 0 0 18px;
        font-size: 18px;
        line-height: 1.45;
      }
      .table-session-terminal button {
        min-height: 48px;
        padding: 10px 20px;
        border: 2px solid white;
        border-radius: 8px;
        background: white;
        color: #17212b;
        font: inherit;
        font-weight: 700;
        cursor: pointer;
      }
      .table-session-terminal button:disabled,
      .table-session-notice button:disabled {
        cursor: wait;
        opacity: 0.65;
      }
      .table-transfer-launch {
        position: fixed; right: var(--boardgame-table-transfer-right, max(16px, env(safe-area-inset-right)));
        bottom: var(--boardgame-table-transfer-bottom, max(16px, env(safe-area-inset-bottom)));
        z-index: var(--boardgame-table-transfer-z-index, 900);
        min-height: 44px; padding: 9px 14px; border: 1px solid #718090;
        border-radius: 999px; background: rgb(255 255 255 / 94%); color: #17212b;
        font: inherit; font-weight: 700; box-shadow: 0 3px 14px rgb(0 0 0 / 18%); cursor: pointer;
      }
      .table-transfer-dialog { width: min(680px, calc(100vw - 32px)); border: 0; border-radius: 16px; padding: 0; color: #17212b; }
      .table-transfer-dialog::backdrop { background: rgb(0 0 0 / 62%); }
      .table-transfer-content { padding: clamp(20px, 5vw, 36px); }
      .table-transfer-content h1 { margin-top: 0; }
      .table-transfer-offer { display: grid; grid-template-columns: minmax(160px, 230px) 1fr; gap: 24px; align-items: center; }
      .table-transfer-offer img { width: 100%; height: auto; }
      .table-transfer-code { font: 800 clamp(24px, 5vw, 38px) ui-monospace, monospace; letter-spacing: .08em; }
      .table-transfer-url { overflow-wrap: anywhere; font-size: .9em; }
      .table-transfer-actions { display: flex; flex-wrap: wrap; gap: 10px; margin-top: 22px; }
      .table-transfer-actions button { min-height: 44px; padding: 8px 16px; border: 0; border-radius: 8px; background: #245f94; color: white; font: inherit; font-weight: 700; cursor: pointer; }
      .table-transfer-actions button.secondary { background: #e4ebf1; color: #17212b; }
      .table-transfer-actions button:disabled { cursor: wait; opacity: .65; }
      .table-transfer-error { color: #a11616; font-weight: 650; }
      .rematch-panel {
        position: fixed; left: 50%; bottom: max(18px, env(safe-area-inset-bottom));
        z-index: 1050; box-sizing: border-box; width: min(560px, calc(100vw - 28px));
        transform: translateX(-50%); padding: 16px 20px; border-radius: 14px;
        background: rgb(23 33 43 / 96%); color: white; text-align: center;
        box-shadow: 0 10px 36px rgb(0 0 0 / 35%);
      }
      .rematch-panel p { margin: 0 0 12px; line-height: 1.4; }
      .rematch-panel button {
        min-height: 46px; padding: 9px 18px; border: 2px solid white;
        border-radius: 9px; background: white; color: #17212b;
        font: inherit; font-weight: 750; cursor: pointer;
      }
      .rematch-panel button:disabled { cursor: wait; opacity: .7; }
      .rematch-error { color: #ffd0d0; font-weight: 650; }
      @media (max-width: 540px) { .table-transfer-offer { grid-template-columns: 1fr; } .table-transfer-offer img { max-width: 220px; margin: auto; } }
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

  @property({ type: Boolean, attribute: 'prompted-to-join' })
  promptedToJoin = false;

  // The current renderer, passed up from the gameRenderer, so we can pass
  // it to stateGameManager and readyForNextState.
  @property({ type: Object, attribute: 'active-renderer' })
  activeRenderer: HostedGameRenderer | null = null;

  @property({ type: Boolean, attribute: 'socket-active' })
  socketActive = false;

  @property({ type: Number, attribute: false })
  _socketConnectionAttempts = 0;

  @property({ type: String, attribute: false })
  _socketError: string | null = null;

  @property({ type: Boolean, attribute: false })
  _firstStateBundle = true;

  @property({ type: Object, attribute: false })
  private _installedMove: ClientMove | null = null;

  @property({ type: Object, attribute: false })
  private _projectedMoveChoices: ProjectedMoveChoicesWire | null = null;

  @query('#manager')
  private _managerEle?: BoardgameGameStateManager;

  @query('#admin')
  private _adminEle?: BoardgameAdminControls;

  @query('#render')
  private _renderEle?: BoardgameRenderGame;

  @query('#player')
  private _playerEle?: BoardgamePlayerRoster;

  @query('#table-session-heading')
  private _tableSessionHeading?: HTMLElement;

  @query('#table-transfer-dialog')
  private _tableTransferDialog?: HTMLDialogElement;

  @query('#table-transfer-heading')
  private _tableTransferHeading?: HTMLElement;

  // Reactive properties - synced from Redux in stateChanged()
  @property({ type: Object, attribute: false })
  _currentState: ExpandedGameState | null = null;

  @property({ type: Object, attribute: false })
  _animationContext: import('./companion-sync.js').VersionAnimationContext | null = null;

  @property({ type: Boolean, attribute: false })
  _legacyAnimationOverlapConfigured = false;

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
  private _surfaceCachedCompanionMode: boolean | null = null;
  private readonly _timerService = new TimerService();

  // Hide-my-hand privacy shield (hand surface only): when true, an opaque
  // full-viewport overlay covers the private hand so the player can set
  // the phone down or step away without shoulder-surfing risk. Purely
  // client-side and per-tab — game state keeps flowing underneath, so
  // revealing is instant and never misses an update.
  @property({ type: Boolean, attribute: false })
  _handHidden = false;

  @property({ type: Boolean, attribute: false })
  _tableLeasePending = false;

  @property({ type: String, attribute: false })
  _tableLeaseError = '';

  private _tableLeaseRequest: AbortController | null = null;
  private _tableLeaseRefreshTimer: ReturnType<typeof setTimeout> | null = null;
  private _tableLeaseRefreshSignature = '';
  private _tableSessionStateSignature = '';
  private _focusedTableTerminalSignature = '';
  private _motionCycleId = 0;

  @property({ type: Object, attribute: false }) private _tableTransferOffer: TableTransferOffer | null = null;
  @property({ type: Boolean, attribute: false }) private _tableTransferOpen = false;
  @property({ type: Boolean, attribute: false }) private _tableTransferPending = false;
  @property({ type: String, attribute: false }) private _tableTransferError = '';
  @property({ type: String, attribute: false }) private _tableTransferCopyStatus = '';
  @property({ type: Number, attribute: false }) private _tableTransferSeconds = 0;
  @property({ type: Boolean, attribute: false }) private _rematchPending = false;
  @property({ type: String, attribute: false }) private _rematchError = '';
  private _rematchRequest: AbortController | null = null;
  private _rematchFollowTarget = '';
  private _tableTransferRequest: AbortController | null = null;
  private _tableTransferTimer: ReturnType<typeof setInterval> | null = null;
  private _tableTransferDeadline = 0;

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
    this.addEventListener('motion-cycle-release', (e: Event) => this._handleMotionCycleRelease(e as CustomEvent));
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
      ${this._renderRematch()}
      ${this._renderTableSessionRecovery()}
      ${this._renderTableTransfer()}
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
          .active=${this.selected}
          @will-animate=${(e: Event) => this._rosterWillAnimate(e as CustomEvent)}
          @animation-done=${(e: Event) => this._rosterAnimationDone(e as CustomEvent)}>
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
          .transitionMove=${this._installedMove}
          .diagram=${this.game ? this.game.Diagram : ''}
          .renderer=${this.activeRenderer}
          @renderer-changed=${this._handleRendererChanged}
          .gameName=${this._gameRoute ? this._gameRoute.name : ''}
          .gameId=${this._gameRoute ? this._gameRoute.id : ''}
          .gameVersion=${this.game ? this.game.Version : 0}
          .snapshotEpoch=${this._moveSnapshotEpoch}
          .projectedMoveChoicesWire=${this._projectedMoveChoices}
          .motionCycleId=${this._motionCycleId}
          .legacyAnimationOverlapConfigured=${this._legacyAnimationOverlapConfigured}
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

  override disconnectedCallback(): void {
    this._tableLeaseRequest?.abort();
    this._tableLeaseRequest = null;
    if (this._tableLeaseRefreshTimer !== null) clearTimeout(this._tableLeaseRefreshTimer);
    this._tableLeaseRefreshTimer = null;
    this._tableLeaseRefreshSignature = '';
    this._tableTransferRequest?.abort();
    this._tableTransferRequest = null;
    if (this._tableTransferTimer !== null) clearInterval(this._tableTransferTimer);
    this._tableTransferTimer = null;
    this._rematchRequest?.abort();
    this._rematchRequest = null;
    super.disconnectedCallback();
  }

  private _renderRematch() {
    if (!this.game?.Finished || !this._companionSurface || !this._companionInfo?.CompanionMode) return '';
    const target = this._companionInfo.RematchGameID;
    if (!target && !this._companionInfo.CanRematch && !this._rematchError) return '';
    return html`
      <aside class="rematch-panel" aria-live="polite" aria-busy=${this._rematchPending ? 'true' : 'false'}>
        <p>${target
          ? 'Your next game is ready. Keeping this seat and moving you there…'
          : 'Keep everyone in the same seats and start a fresh game.'}</p>
        ${!target && this._companionInfo.CanRematch ? html`
          <button type="button" ?disabled=${this._rematchPending} @click=${this._startRematch}>
            ${this._rematchPending ? 'Preparing rematch…' : 'Play again with the same players'}
          </button>
        ` : ''}
        ${this._rematchError ? html`<p class="rematch-error" role="alert">${this._rematchError}</p>` : ''}
      </aside>
    `;
  }

  private _navigateToRematch(rematch: Pick<RematchResponse, 'gameID' | 'gameName'>): void {
    const surface = this._companionSurface;
    const oldGameID = this._gameRoute?.id;
    if (!surface || !oldGameID) return;
    rememberSurfaceForGame(rematch.gameID, surface);
    forgetSurfaceForGame(oldGameID);
    window.location.replace(`${gamePath(rematch.gameName, rematch.gameID)}?display=${surface}`);
  }

  private readonly _startRematch = (): void => {
    void this._requestRematch();
  };

  private async _requestRematch(expectedGameID = ''): Promise<void> {
    const route = this._gameRoute;
    if (!route || !this._companionSurface || this._rematchPending) return;
    const request = new AbortController();
    this._rematchRequest?.abort();
    this._rematchRequest = request;
    this._rematchPending = true;
    this._rematchError = '';
    try {
      const response = await apiHttpPost(
        buildGameUrl(route.name, route.id, 'rematch'),
        {},
        { signal: request.signal },
      );
      if (request.signal.aborted || this._gameRoute?.id !== route.id) return;
      if (!response.data) {
        this._rematchError = response.error || response.friendlyError || 'The rematch could not be prepared. Please try again.';
        if (expectedGameID) this._rematchFollowTarget = '';
        return;
      }
      const rematch = decodeRematchResponse(response.data);
      if (expectedGameID && rematch.gameID !== expectedGameID) {
        throw new Error('Rematch response did not match the published successor');
      }
      this._navigateToRematch(rematch);
    } catch (error) {
      if (!request.signal.aborted) {
        console.error('[game-view] malformed rematch response', error);
        this._rematchError = 'The server returned an invalid rematch response. Please try again.';
        if (expectedGameID) this._rematchFollowTarget = '';
      }
    } finally {
      if (this._rematchRequest === request) {
        this._rematchRequest = null;
        this._rematchPending = false;
      }
    }
  }

  private _followPublishedRematch(gameID: string): void {
    const route = this._gameRoute;
    const surface = this._companionSurface;
    if (!route || !surface || !gameID || this._rematchFollowTarget === gameID) return;
    this._rematchFollowTarget = gameID;
    queueMicrotask(() => {
      if (this._gameRoute?.id !== route.id || this._companionInfo?.RematchGameID !== gameID) {
        this._rematchFollowTarget = '';
        return;
      }
      if (surface === 'table') {
        // The old Table capability is exchanged server-side for a credential
        // scoped to the successor before navigation.
        void this._requestRematch(gameID);
        return;
      }
      this._navigateToRematch({ gameID, gameName: route.name });
    });
  }

  private _renderTableTransfer() {
    const activeTable = !this.game?.Finished
      && this._companionSurface === 'table'
      && this._companionInfo?.TableSession.Status === 'active'
      && this._companionInfo.TableSession.IsThisTable;
    if (!activeTable && !this._tableTransferOpen) return '';
    const offer = this._tableTransferOffer;
    return html`
      ${activeTable && !this._tableTransferOpen ? html`
        <button type="button" class="table-transfer-launch" part="table-transfer-launch" @click=${this._openTableTransfer}>Move shared Table</button>
      ` : ''}
      <dialog id="table-transfer-dialog" class="table-transfer-dialog" aria-labelledby="table-transfer-heading" @cancel=${this._handleTableTransferCancelEvent}>
        <section class="table-transfer-content" aria-busy=${this._tableTransferPending ? 'true' : 'false'}>
          <h1 id="table-transfer-heading" tabindex="-1">Move the shared Table</h1>
          ${offer ? html`
            <p>The game stays active here until the other screen connects.</p>
            <div class="table-transfer-offer">
              <img src=${offer.qrDataURL} alt="QR code that opens this Table transfer on another screen">
              <div>
                <p>On the other screen, scan the QR code or open:</p>
                <p class="table-transfer-url">${offer.claimURL}</p>
                <p>Or go to <strong>${window.location.host}/table</strong> and enter room <strong>${this._companionInfo?.RoomCode}</strong> with transfer code:</p>
                <p class="table-transfer-code">${offer.manualCode}</p>
                <p aria-live="polite">${this._tableTransferSeconds > 0 ? `Expires in ${this._formatTransferTime(this._tableTransferSeconds)}.` : 'This transfer has expired.'}</p>
              </div>
            </div>
          ` : html`<p>Creating a short-lived, one-use connection for the other screen…</p>`}
          ${this._tableTransferError ? html`<p class="table-transfer-error" role="alert">${this._tableTransferError}</p>` : ''}
          <p aria-live="polite">${this._tableTransferCopyStatus}</p>
          <div class="table-transfer-actions">
            ${offer ? html`<button type="button" ?disabled=${this._tableTransferPending || this._tableTransferSeconds <= 0} @click=${this._copyTransferLink}>Copy link</button>` : ''}
            ${offer ? html`<button type="button" class="secondary" ?disabled=${this._tableTransferPending} @click=${this._cancelTableTransfer}>Cancel transfer</button>` : ''}
            <button type="button" class="secondary" @click=${this._dismissTableTransfer}>Close</button>
          </div>
        </section>
      </dialog>
    `;
  }

  private readonly _openTableTransfer = async (): Promise<void> => {
    const route = this._gameRoute;
    if (!route || this._tableTransferPending) return;
    this._tableTransferOpen = true;
    this._tableTransferPending = true;
    this._tableTransferError = '';
    this._tableTransferCopyStatus = '';
    this._tableTransferOffer = null;
    await this.updateComplete;
    const session = this._companionInfo?.TableSession;
    if (!this.selected || !this._tableTransferOpen || this._gameRoute?.id !== route.id
      || session?.Status !== 'active' || !session.IsThisTable) {
      this._tableTransferPending = false;
      return;
    }
    this._syncTableTransferDialog();
    const request = new AbortController();
    this._tableTransferRequest?.abort();
    this._tableTransferRequest = request;
    try {
      const response = await apiHttpPost(buildGameUrl(route.name, route.id, 'tableTransfer/create'), {}, { signal: request.signal });
      if (request.signal.aborted || this._gameRoute?.id !== route.id) return;
      if (!response.data) {
        this._tableTransferError = transferFailureMessage(response.code, response.error || response.friendlyError);
        return;
      }
      const offer = decodeTableTransferOffer(response.data);
      this._tableTransferOffer = offer;
      this._tableTransferDeadline = Date.now() + Math.max(0, offer.expiresAtMs - offer.serverNowMs);
      this._updateTableTransferCountdown();
      if (this._tableTransferTimer !== null) clearInterval(this._tableTransferTimer);
      this._tableTransferTimer = setInterval(() => this._updateTableTransferCountdown(), 1000);
    } catch (error) {
      if (!request.signal.aborted) {
        console.error('[game-view] malformed Table transfer offer', error);
        this._tableTransferError = 'The server returned an invalid Table transfer response. Please try again.';
      }
    } finally {
      if (this._tableTransferRequest === request) {
        this._tableTransferRequest = null;
        this._tableTransferPending = false;
      }
    }
  };

  private _updateTableTransferCountdown(): void {
    this._tableTransferSeconds = Math.max(0, Math.ceil((this._tableTransferDeadline - Date.now()) / 1000));
    if (this._tableTransferSeconds === 0 && this._tableTransferTimer !== null) {
      clearInterval(this._tableTransferTimer);
      this._tableTransferTimer = null;
    }
  }

  private _formatTransferTime(seconds: number): string {
    return `${Math.floor(seconds / 60)}:${String(seconds % 60).padStart(2, '0')}`;
  }

  private readonly _copyTransferLink = async (): Promise<void> => {
    if (!this._tableTransferOffer) return;
    try {
      await navigator.clipboard.writeText(this._tableTransferOffer.claimURL);
      this._tableTransferCopyStatus = 'Transfer link copied.';
    } catch {
      this._tableTransferError = 'The link could not be copied. Select the displayed link instead.';
    }
  };

  private readonly _handleTableTransferCancelEvent = (event: Event): void => {
    event.preventDefault();
    this._dismissTableTransfer();
  };

  private readonly _dismissTableTransfer = (): void => {
    this._tableTransferRequest?.abort();
    this._tableTransferRequest = null;
    this._tableTransferPending = false;
    this._closeTableTransfer();
  };

  private readonly _cancelTableTransfer = async (): Promise<void> => {
    const offer = this._tableTransferOffer;
    const route = this._gameRoute;
    if (this._tableTransferPending) return;
    if (!offer || !route) {
      this._closeTableTransfer();
      return;
    }
    this._tableTransferPending = true;
    this._tableTransferError = '';
    this._tableTransferCopyStatus = '';
    const request = new AbortController();
    this._tableTransferRequest?.abort();
    this._tableTransferRequest = request;
    try {
      const response = await apiHttpPost(
        buildGameUrl(route.name, route.id, 'tableTransfer/cancel'),
        { pairingID: offer.pairingID },
        { signal: request.signal },
      );
      if (request.signal.aborted || this._gameRoute?.id !== route.id || !this._tableTransferOpen) return;
      if (!response.data) {
        this._tableTransferError = transferFailureMessage(response.code, response.error || response.friendlyError);
        return;
      }
      decodeTableTransferCancel(response.data);
      this._closeTableTransfer();
    } catch (error) {
      if (!request.signal.aborted) {
        console.error('[game-view] malformed Table transfer cancellation', error);
        this._tableTransferError = 'The server returned an invalid cancellation response. Please try again.';
      }
    } finally {
      if (this._tableTransferRequest === request) {
        this._tableTransferRequest = null;
        this._tableTransferPending = false;
      }
    }
  };

  private _closeTableTransfer(): void {
    this._tableTransferOpen = false;
    this._tableTransferOffer = null;
    this._tableTransferError = '';
    if (this._tableTransferTimer !== null) clearInterval(this._tableTransferTimer);
    this._tableTransferTimer = null;
    this._tableTransferDialog?.close();
    void this.updateComplete.then(() => {
      this.renderRoot.querySelector<HTMLButtonElement>('.table-transfer-launch')?.focus();
    });
  }

  private _syncTableTransferDialog(): void {
    const dialog = this._tableTransferDialog;
    if (!dialog) return;
    if (this._tableTransferOpen && !dialog.open) dialog.showModal();
    if (!this._tableTransferOpen && dialog.open) dialog.close();
    if (this._tableTransferOpen) this._tableTransferHeading?.focus();
  }

  private _renderTableSessionRecovery() {
    const session = this._companionInfo?.TableSession;
    const surface = this._companionSurface;
    if (!this._companionInfo?.CompanionMode || !session || !surface || this.game?.Finished) return '';

    if (surface === 'table') {
      if (session.Status === 'active' && session.IsThisTable) return '';
      const available = session.Status === 'available';
      return html`
        <div class="table-session-terminal">
          <section aria-live="polite" aria-busy=${this._tableLeasePending ? 'true' : 'false'}>
            <h1 id="table-session-heading" tabindex="-1">
              ${available ? 'Restore the shared Table' : (session.DisplacedByTransfer ? 'The shared Table moved successfully' : 'This is no longer the shared Table')}
            </h1>
            <p>${available
              ? (session.CanTakeOver
                ? 'The previous Table is no longer connected. This screen can safely take its place.'
                : 'The previous Table is gone. A seated player can restore it from their Hand screen.')
              : (session.DisplacedByTransfer
                ? 'The game is now running on the new screen. This screen is safely paused.'
                : 'Another Table is active or reconnecting. This screen will remain paused to prevent two shared displays from controlling the game.')}</p>
            ${available && session.CanTakeOver ? html`
              <button type="button" ?disabled=${this._tableLeasePending} @click=${this._acquireTableLease}>
                ${this._tableLeasePending ? 'Restoring Table…' : 'Restore Table on this screen'}
              </button>
            ` : ''}
            ${this._tableLeaseError ? html`
              <p class="table-session-error" role="alert">${this._tableLeaseError}</p>
            ` : ''}
          </section>
        </div>
      `;
    }

    if (session.Status === 'active') return '';

    return html`
      <aside class="table-session-notice" aria-live="polite"
        aria-busy=${this._tableLeasePending ? 'true' : 'false'}>
        ${session.CanTakeOver ? html`
          <p>The shared Table disconnected. You can safely move it to this screen.</p>
          <button type="button" ?disabled=${this._tableLeasePending} @click=${this._acquireTableLease}>
            ${this._tableLeasePending ? 'Taking over…' : 'Take over shared Table'}
          </button>
        ` : html`
          <p>The shared Table is disconnected. A seated player can take it over.</p>
        `}
        ${this._tableLeaseError ? html`
          <p class="table-session-error" role="alert">${this._tableLeaseError}</p>
        ` : ''}
      </aside>
    `;
  }

  private readonly _acquireTableLease = async (): Promise<void> => {
    if (this._tableLeasePending) return;
    const route = this._gameRoute;
    const session = this._companionInfo?.TableSession;
    if (!this.selected || !route || session?.Status !== 'available' || !session.CanTakeOver) return;

    const request = new AbortController();
    this._tableLeaseRequest?.abort();
    this._tableLeaseRequest = request;
    this._tableLeasePending = true;
    this._tableLeaseError = '';
    try {
      const response = await apiHttpPost(
        buildGameUrl(route.name, route.id, 'tableLease/acquire'),
        { deviceID: tableRecoveryDeviceID(route.id) },
        { signal: request.signal },
      );
      if (request.signal.aborted || !this.selected
        || this._gameRoute?.id !== route.id
        || this._gameRoute?.name !== route.name) return;
      if (!response.data) {
        this._tableLeaseError = isTableLeaseFailureCode(response.code)
          ? tableLeaseFailureMessage(response.code)
          : response.error || response.friendlyError || 'The shared Table could not be restored.';
        this._refreshTableSessionNow();
        return;
      }
      decodeTableLeaseAcquireResponse(response.data);
      rememberSurfaceForGame(route.id, 'table');
      const target = new URL(window.location.href);
      target.searchParams.set('display', 'table');
      window.location.replace(target.pathname + target.search);
    } catch (error) {
      if (!request.signal.aborted) {
        console.error('[game-view] malformed Table lease response', error);
        this._tableLeaseError = 'The server returned an invalid Table recovery response. Please try again.';
      }
    } finally {
      if (this._tableLeaseRequest === request) {
        this._tableLeaseRequest = null;
        this._tableLeasePending = false;
      }
    }
  };

  private _refreshTableSessionNow(): void {
    const route = this._gameRoute;
    if (!route) return;
    store.dispatch(fetchGameInfo(route, this.requestedPlayer, this._admin, this._lastFetchedVersion));
  }

  private _scheduleTableSessionRefresh(): void {
    const route = this._gameRoute;
    const session = this._companionInfo?.TableSession;
    const needsRefresh = this.selected
      && this._companionInfo?.CompanionMode
      && !!this._companionSurface
	  && !!session;
    const signature = needsRefresh && route && session
      ? `${route.id}:${this._companionSurface}:${session.Status}:${session.RetryAfterMs}:${this.game?.Finished ? 1 : 0}:${this._companionInfo?.RematchGameID ?? ''}`
      : '';
    if (signature === this._tableLeaseRefreshSignature) return;
    if (this._tableLeaseRefreshTimer !== null) clearTimeout(this._tableLeaseRefreshTimer);
    this._tableLeaseRefreshTimer = null;
    this._tableLeaseRefreshSignature = signature;
    if (!signature || !session) return;
    // Add a small margin so the server clock has crossed the lease deadline.
	// Active leases refresh at the authoritative deadline; available state
	// polls gently so a takeover completed through another server instance
	// converges even without shared websocket fanout.
	const delay = this.game?.Finished
	  ? 2_000
	  : session.Status === 'active'
	  ? Math.min(Math.max(session.RetryAfterMs + 100, 250), 60_000)
	  : 5_000;
    this._tableLeaseRefreshTimer = setTimeout(() => {
      this._tableLeaseRefreshTimer = null;
      this._tableLeaseRefreshSignature = '';
      this._refreshTableSessionNow();
    }, delay);
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
    const previousRouteKey = this._gameRoute
      ? `${this._gameRoute.name}\u0000${this._gameRoute.id}`
      : '';
    this._gameRoute = selectGameRoute(state);
    const routeKey = this._gameRoute ? `${this._gameRoute.name}\u0000${this._gameRoute.id}` : '';
    if (previousRouteKey !== routeKey) {
      this._tableLeaseRequest?.abort();
      this._tableLeaseRequest = null;
      this._tableLeasePending = false;
      this._tableLeaseError = '';
      this._tableTransferRequest?.abort();
      this._tableTransferRequest = null;
      this._tableTransferPending = false;
      this._tableTransferOpen = false;
      this._tableTransferOffer = null;
      if (this._tableTransferTimer !== null) clearInterval(this._tableTransferTimer);
      this._tableTransferTimer = null;
      this._rematchRequest?.abort();
      this._rematchRequest = null;
      this._rematchPending = false;
      this._rematchError = '';
      this._rematchFollowTarget = '';
    }
    const tableSession = this._companionInfo?.TableSession;
    const tableSessionStateSignature = tableSession
      ? `${routeKey}:${tableSession.Status}:${tableSession.IsThisTable}:${tableSession.CanTakeOver}`
      : '';
    if (this._tableSessionStateSignature
      && this._tableSessionStateSignature !== tableSessionStateSignature) {
      this._tableLeaseError = '';
    }
    this._tableSessionStateSignature = tableSessionStateSignature;
    const surfaceGameId = this._gameRoute ? this._gameRoute.id : null;
    const companionMode = this._companionInfo?.CompanionMode ?? null;
    if (surfaceGameId !== this._surfaceCachedGameId
      || companionMode !== this._surfaceCachedCompanionMode) {
      this._surfaceCachedGameId = surfaceGameId;
      this._surfaceCachedCompanionMode = companionMode;
      this._companionSurface = surfaceGameId
        ? surfaceForGame(surfaceGameId, companionMode ?? undefined)
        : null;
      if (surfaceGameId && companionMode === false) forgetSurfaceForGame(surfaceGameId);
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
    this._scheduleTableSessionRefresh();
    if (this._companionInfo?.RematchGameID) {
      this._followPublishedRematch(this._companionInfo.RematchGameID);
    }
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

    if (changedProps.has('_tableTransferOpen')) this._syncTableTransferDialog();
    if (this._tableTransferOpen && changedProps.has('_companionInfo')) {
      const session = this._companionInfo?.TableSession;
      if (!session || session.Status !== 'active' || !session.IsThisTable) {
        this._tableTransferRequest?.abort();
        this._tableTransferRequest = null;
        this._tableTransferPending = false;
        this._closeTableTransfer();
      }
    }

    if (changedProps.has('selected')) {
      if (!this.selected) {
        this._tableLeaseRequest?.abort();
        this._tableLeaseRequest = null;
        this._tableLeasePending = false;
        this._tableLeaseError = '';
        this._tableTransferRequest?.abort();
        this._tableTransferRequest = null;
        this._tableTransferPending = false;
        this._tableTransferOpen = false;
        this._tableTransferOffer = null;
        this._tableTransferError = '';
        if (this._tableTransferTimer !== null) clearInterval(this._tableTransferTimer);
        this._tableTransferTimer = null;
        this._tableTransferDialog?.close();
        this._rematchRequest?.abort();
        this._rematchRequest = null;
        this._rematchPending = false;
      }
      this._scheduleTableSessionRefresh();
    }
    const tableSession = this._companionInfo?.TableSession;
    const terminalSignature = this._companionSurface === 'table'
      && tableSession
      && !(tableSession.Status === 'active' && tableSession.IsThisTable)
      ? `${this._gameRoute?.name ?? ''}:${this._gameRoute?.id ?? ''}:${tableSession.Status}`
      : '';
    if (terminalSignature && terminalSignature !== this._focusedTableTerminalSignature) {
      this._focusedTableTerminalSignature = terminalSignature;
      this._tableSessionHeading?.focus();
    } else if (!terminalSignature) {
      this._focusedTableTerminalSignature = '';
    }

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
    this._renderEle?.installCompanionInfo(bundle.companionInfo ?? null);
    store.dispatch(updateGameStaticInfo(bundle.chest, bundle.playersInfo, bundle.hasEmptySlots, bundle.open, bundle.visible, bundle.isOwner, bundle.companionInfo));
  }

  private _forwardCycleRelease(cycleId: unknown): void {
    if (!Number.isInteger(cycleId) || cycleId !== this._motionCycleId) return;
    // Dispatch custom event for animation coordination
    // The manager element will listen for this and handle it
    if (this._managerEle) {
      this._managerEle.dispatchEvent(new CustomEvent('ready-for-next-state', {
        bubbles: true,
        composed: true,
        detail: Object.freeze({ cycleId }),
      }));
    }
  }

  private _handleAllAnimationsDone(e: Event) {
    this._forwardCycleRelease((e as CustomEvent).detail?.cycleId);
  }

  private _handleMotionCycleRelease(e: CustomEvent) {
    // Companion slots are server-owned; local progress must never advance them.
    if (this._animationContext !== null) return;
    this._forwardCycleRelease(e.detail?.cycleId);
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

  // _rosterWillAnimate/_rosterAnimationDone (Task 10, #714's second Phase 2
  // gap): boardgame-player-roster is a DOM SIBLING of boardgame-render-game
  // (both rendered directly here), so a roster-hosted animatable's
  // (boardgame-status-text's nested boardgame-fading-text, ...)
  // will-animate/animation-done events bubble past render-game's own
  // listeners (installed on itself) and were previously silently un-gated
  // -- the literal #714 checklist ask ("verify that status-text and
  // friends in render-player-info will also be waited for"). Forwarding
  // pipes them into render-game's gate via its gateWillAnimate/
  // gateAnimationDone delegates.
  //
  // Direction guard (HARNESS-CRITIC REQUIREMENT, gap 3): will-animate is
  // forwarded ONLY while a board cycle is already open
  // (this._renderEle.isAnimating) -- a roster animation outside any cycle
  // (e.g. a hover-triggered fade) must NOT be able to open or queue a new
  // cycle; it simply has no effect on the gate. animation-done is ALWAYS
  // forwarded regardless of isAnimating: a participant admitted at open
  // must always be able to settle, and the gate kernel's animationDone()
  // is a safe no-op for an unregistered/unknown element (see
  // src/motion/animation-gate.ts), so forwarding an out-of-cycle settle
  // that was never registered cannot spuriously close anything.
  //
  // Orphan-settle done channel (#714 Phase 2 gate finding, evidence:
  // docs/superpowers/specs/evidence/2026-07-25-roster-orphan-settle.md):
  // the bubbled `animation-done` this class listens for on
  // boardgame-player-roster (wired up in the template, see render()) can
  // ONLY reach us while the animatable stays attached to the DOM -- a
  // detached node has no parent to bubble to. A roster item removed
  // mid-animation is still force-settled by BoardgameAnimatableItem's own
  // disconnectedCallback (see that file), but that settlement has to reach
  // this gate through a channel that does not depend on DOM presence.
  // `settled()` is exactly that channel: it resolves off the same
  // gated-count bookkeeping that drives the bubbled event, via a promise
  // that keeps its resolver regardless of whether the element is still
  // connected. Subscribing here (at will-animate time, while the element is
  // still live) rather than in a connectedCallback/registry hook keeps this
  // additive to the existing forward above rather than a replacement for
  // it -- in the normal (attached, never-removed) case both fire:
  // dispatchEvent's listeners run synchronously inside
  // BoardgameAnimatableItem's settlement handler, before it resolves the
  // settled() promise's continuations (a promise resolution is only
  // observable on a later microtask) -- so the bubbled path always closes
  // the gate first, and this settled() call arrives one microtask later to
  // find the kernel's animationDone() already a no-op for that element
  // (verified against src/motion/animation-gate.ts: a second call either
  // hits the size===0 early return, once the cycle already closed, or is a
  // harmless duplicate delete() on an already-absent key otherwise) --
  // confirmed by this file's non-wedging-guard test and the roster gate
  // test both staying green with this channel present. Only the orphaned
  // case (bubble impossible) actually depends on this second path.
  private _rosterWillAnimate(e: CustomEvent) {
    if (!this._renderEle?.isAnimating) return;
    this._renderEle.gateWillAnimate(e);
    const ele = e.detail?.ele as BoardgameAnimatableItem | undefined;
    if (ele && typeof ele.settled === 'function') {
      void ele.settled().then(() => {
        this._renderEle?.gateAnimationDone(new CustomEvent('animation-done', {
          detail: { ele },
        }));
      });
    }
  }

  private _rosterAnimationDone(e: CustomEvent) {
    this._renderEle?.gateAnimationDone(e);
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
    this._motionCycleId = 0;
    this._legacyAnimationOverlapConfigured = false;
    this._installedMove = null;
    this._projectedMoveChoices = null;
    this._moveSnapshotEpoch += 1;
    this._firstStateBundle = true;
  }

  private _installStateBundle(bundle: StateBundle) {
    // Set the version's animation context before publishing its state. The
    // render-game wrapper applies it to the shared animator before assigning
    // the new state to the game renderer.
    this._animationContext = bundle.animationContext ?? null;
    this._motionCycleId = bundle.motionCycleId ?? 0;
    this._legacyAnimationOverlapConfigured = bundle.legacyAnimationOverlapConfigured === true;
    this._installedMove = bundle.move;
    this._projectedMoveChoices = bundle.projectedMoveChoices;
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
