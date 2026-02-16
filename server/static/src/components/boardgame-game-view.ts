import { LitElement, html, css } from 'lit';
import { customElement, property, query } from 'lit/decorators.js';
import './boardgame-player-roster.js';
import './boardgame-render-game.js';
import './boardgame-admin-controls.js';
import './boardgame-game-state-manager.js';
import { sharedStyles } from './shared-styles-lit.js';

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
  selectExpandedGameState,
  selectGame,
  selectViewingAsPlayer,
  selectRequestedPlayer,
  selectAutoCurrentPlayer,
  selectMoveForms,
  selectLastFetchedVersion
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
  fetchGameInfo
} from '../actions/game.js';

import type { StateBundle } from '../types/game-state';
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
  game: any = null;

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
  activeRenderer: any = null;

  @property({ type: Boolean })
  socketActive = false;

  @property({ type: Boolean })
  _firstStateBundle = true;

  @query('#manager')
  private _managerEle: any;

  @query('#admin')
  private _adminEle: any;

  @query('#render')
  private _renderEle: any;

  @query('#player')
  private _playerEle: any;

  // Reactive properties - synced from Redux in stateChanged()
  @property({ type: Object, attribute: false })
  _currentState: any = null;

  @property({ type: Object, attribute: false })
  _chest: any = null;

  @property({ type: Array, attribute: false })
  _playersInfo: any[] = [];

  @property({ type: Boolean, attribute: false })
  _hasEmptySlots = false;

  @property({ type: Boolean, attribute: false })
  _open = false;

  @property({ type: Boolean, attribute: false })
  _visible = false;

  @property({ type: Boolean, attribute: false })
  _isOwner = false;

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

  constructor() {
    super();

    this.addEventListener('propose-move', (e: Event) => this._handleProposeMove(e as CustomEvent));
    this.addEventListener('refresh-info', (e: Event) => this._handleRefreshData(e));
    this.addEventListener('install-state-bundle', (e: Event) => this._handleStateBundle(e as CustomEvent));
    this.addEventListener('install-game-static-info', (e: Event) => this._handleGameStaticInfo(e as CustomEvent));
    this.addEventListener('all-animations-done', (e: Event) => this._handleAllAnimationsDone(e));
    this.addEventListener('set-animation-length', (e: Event) => this._handleSetAnimationLength(e as CustomEvent));
  }

  override render() {
    return html`
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
          .active=${this.selected}>
        </boardgame-player-roster>
      </div>
      <div class="card">
        <boardgame-render-game
          id="render"
          .state=${this._currentState}
          .diagram=${this.game ? this.game.Diagram : ''}
          .renderer=${this.activeRenderer}
          @renderer-changed=${this._handleRendererChanged}
          .gameName=${this._gameRoute ? this._gameRoute.name : ''}
          .viewingAsPlayer=${this.viewingAsPlayer}
          .currentPlayerIndex=${this.game ? this.game.CurrentPlayerIndex : 0}
          .socketActive=${this.socketActive}
          .active=${this.selected}
          .chest=${this._chest}>
        </boardgame-render-game>
      </div>
      <boardgame-admin-controls
        id="admin"
        .active=${this._admin}
        .game=${this.game}
        .viewingAsPlayer=${this.viewingAsPlayer}
        .moveForms=${this.moveForms}
        .gameRoute=${this._gameRoute}
        .chest=${this._chest}
        .currentState=${this._currentState}
        @requested-player-changed=${this._handleRequestedPlayerChanged}
        @auto-current-player-changed=${this._handleAutoCurrentPlayerChanged}>
      </boardgame-admin-controls>
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

  stateChanged(state: any) {
    // Sync view state from Redux
    this.game = selectGame(state);
    this.viewingAsPlayer = selectViewingAsPlayer(state);
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
    this._pageExtra = selectPageExtra(state);
    this._gameRoute = selectGameRoute(state);
    this._loggedIn = selectLoggedIn(state);
    this._admin = selectAdmin(state);
    this._page = selectPage(state);
    this._lastFetchedVersion = selectLastFetchedVersion(state);
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

  private _handleSocketActiveChanged(e: CustomEvent) {
    this.socketActive = e.detail.value;
  }

  private _handleRendererChanged(e: CustomEvent) {
    this.activeRenderer = e.detail.value;
  }

  private _handleProposeMove(e: CustomEvent) {
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
    store.dispatch(updateGameStaticInfo(bundle.chest, bundle.playersInfo, bundle.hasEmptySlots, bundle.open, bundle.visible, bundle.isOwner));
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
    this._renderEle.defaultAnimationLength = e.detail;
  }

  private _firstStateBundleInstalled() {
    if (this.selected && this._loggedIn && this._playerEle.showJoin && !this.promptedToJoin) {
      // Take note that we already prompted them, and don't prompt again unless the game changes.
      this.promptedToJoin = true;
      // Prompt the user to join!
      this._playerEle.showDialog();
    }
  }

  private _resetState() {
    // Reset view state properties only
    // Computed properties (_currentState, _chest, etc.) are read from Redux selectors
    this.game = null;
    this.moveForms = null;
    this.viewingAsPlayer = 0;
    this._firstStateBundle = true;
  }

  private _installStateBundle(bundle: StateBundle) {
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
