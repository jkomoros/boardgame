import { LitElement, html, css } from 'lit';
import { customElement, property, query } from 'lit/decorators.js';
import './boardgame-player-roster.js';
import './boardgame-render-game.js';
import './boardgame-admin-controls.js';
import './boardgame-game-state-manager.js';
import { SharedStyles } from './shared-styles-lit.js';

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
  selectGameCurrentState
} from '../selectors.js';

import {
  PAGE_GAME
} from '../actions/app.js';

import {
  updateGameRoute,
  updateGameStaticInfo,
  installGameState,
} from '../actions/game.js';

import game from '../reducers/game.js';
store.addReducers({
  game
});

@customElement('boardgame-game-view')
export class BoardgameGameView extends connect(store)(LitElement) {
  static override styles = css`
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
  `;

  @property({ type: Number })
  requestedPlayer = 0;

  @property({ type: Object })
  game: any = null;

  @property({ type: Object })
  _currentState: any = null;

  @property({ type: Object })
  _chest: any = null;

  @property({ type: Array })
  _playersInfo: any[] = [];

  @property({ type: Boolean })
  _hasEmptySlots = false;

  @property({ type: Boolean })
  _open = false;

  @property({ type: Boolean })
  _visible = false;

  @property({ type: Boolean })
  _isOwner = false;

  @property({ type: Boolean })
  autoCurrentPlayer = false;

  @property({ type: Boolean })
  selected = false;

  @property({ type: Boolean })
  promptedToJoin = false;

  @property({ type: Number })
  viewingAsPlayer = 0;

  // The current renderer, passed up from the gameRenderer, so we can pass
  // it to stateGameManager and readyForNextState.
  @property({ type: Object })
  activeRenderer: any = null;

  @property({ type: Object })
  moveForms: any = null;

  @property({ type: Boolean })
  socketActive = false;

  @property({ type: Boolean })
  _firstStateBundle = true;

  @property({ type: String })
  _pageExtra = '';

  @property({ type: Object })
  _gameRoute: any = null;

  @property({ type: Boolean })
  _loggedIn = false;

  @property({ type: Boolean })
  _admin = false;

  @property({ type: String })
  _page = '';

  @query('#manager')
  private _managerEle: any;

  @query('#admin')
  private _adminEle: any;

  @query('#render')
  private _renderEle: any;

  @query('#player')
  private _playerEle: any;

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
      ${SharedStyles}
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
        .requestedPlayer=${this.requestedPlayer}
        @requested-player-changed=${this._handleRequestedPlayerChanged}
        .autoCurrentPlayer=${this.autoCurrentPlayer}
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
    this._page = selectPage(state);
    this._pageExtra = selectPageExtra(state);
    this._gameRoute = selectGameRoute(state);
    this._loggedIn = selectLoggedIn(state);
    this._admin = selectAdmin(state);
    this._chest = selectGameChest(state);
    this._playersInfo = selectGamePlayersInfo(state);
    this._hasEmptySlots = selectGameHasEmptySlots(state);
    this._open = selectGameOpen(state);
    this._visible = selectGameVisible(state);
    this._isOwner = selectGameIsOwner(state);
    this._currentState = selectGameCurrentState(state);
  }

  private _handleRefreshData(e: Event) {
    this._managerEle.fetchInfo();
  }

  private _handleRequestedPlayerChanged(e: CustomEvent) {
    this.requestedPlayer = e.detail.value;
  }

  private _handleAutoCurrentPlayerChanged(e: CustomEvent) {
    this.autoCurrentPlayer = e.detail.value;
  }

  private _handleSocketActiveChanged(e: CustomEvent) {
    this.socketActive = e.detail.value;
  }

  private _handleRendererChanged(e: CustomEvent) {
    this.activeRenderer = e.detail.value;
  }

  private _handleProposeMove(e: CustomEvent) {
    this._adminEle.proposeMove(e.detail.name, e.detail.arguments);
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
    this._managerEle.readyForNextState();
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
    this.game = null;
    this._currentState = null;
    this.moveForms = null;
    this.viewingAsPlayer = 0;
    this._firstStateBundle = true;
    this._chest = null;
    this._playersInfo = [];
    this._hasEmptySlots = false;
    this._open = false;
    this._visible = false;
    this._isOwner = false;
    this._firstStateBundle = true;
  }

  private _installStateBundle(bundle: any) {
    store.dispatch(installGameState(bundle.game.CurrentState, bundle.game.ActiveTimers, bundle.originalWallClockStartTime));

    // We only rerender once despite setting multiple properties at once
    this.game = bundle.game;
    this.moveForms = bundle.moveForms;
    this.viewingAsPlayer = bundle.viewingAsPlayer;

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
