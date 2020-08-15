import { LitElement, html } from '@polymer/lit-element';
import './boardgame-player-roster.js';
import './shared-styles.js';
import './boardgame-render-game.js';
import './boardgame-admin-controls.js';
import './boardgame-game-state-manager.js';
import { SharedStyles } from './shared-styles-lit.js';
import {
  deepCopy,
  getProperty,
  setProperty,
} from '../util.js';

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
  selectGameIsOwner
} from '../selectors.js';

import {
  PAGE_GAME
} from '../actions/app.js';

import {
  updateGameRoute,
  updateGameStaticInfo
} from '../actions/game.js';

import game from '../reducers/game.js';
store.addReducers({
	game
});

class BoardgameGameView extends connect(store)(LitElement) {
  render() {
    return html`
    ${ SharedStyles }
    <style>
      :host {
        display: block;
        --animation-length: 0.5s;
      }

      [hidden] {
        display:none !important;
      }

      #moves > details {
        margin-left:1em;
      }

      .admin > div:first-child {
        margin-left: 0;
      }

      .admin > div {
        margin-left:1em;
      }

      .card {
        position:relative;
      }
    </style>

    <div class="card">
      <boardgame-player-roster id="player" .loggedIn=${this._loggedIn} .gameRoute=${this._gameRoute} .viewingAsPlayer=${this.viewingAsPlayer} .hasEmptySlots=${this._hasEmptySlots} .gameOpen=${this._open} .gameVisible=${this._visible} .currentPlayerIndex=${this.game ? this.game.CurrentPlayerIndex : 0} .playersInfo=${this._playersInfo} .state=${this.currentState} .finished=${this.game ? this.game.Finished : false} .winners=${this.game ? this.game.Winners : []} .admin=${this._admin} .isOwner=${this._isOwner} .active=${this.selected}></boardgame-player-roster>
    </div>
    <div class="card">
      <boardgame-render-game id="render" .state=${this.currentState} .diagram=${this.game ? this.game.Diagram : ""} .renderer=${this.activeRenderer} @renderer-changed=${this._handleRendererChanged} .gameName=${this._gameRoute ? this._gameRoute.name : ""} .viewingAsPlayer=${this.viewingAsPlayer} .currentPlayerIndex=${this.game ? this.game.CurrentPlayerIndex : 0} .socketActive=${this.socketActive} .active=${this.selected} .chest=${this._chest}></boardgame-render-game>
    </div>
    <boardgame-admin-controls id="admin" .active=${this._admin} .game=${this.game} .viewingAsPlayer=${this.viewingAsPlayer} .moveForms=${this.moveForms} .gameRoute=${this._gameRoute} .chest=${this._chest} .currentState=${this.currentState} .requestedPlayer=${this.requestedPlayer} @requested-player-changed=${this._handleRequestedPlayerChanged} .autoCurrentPlayer=${this.autoCurrentPlayer} @auto-current-player-changed=${this._handleAutoCurrentPlayerChanged}></boardgame-admin-controls>
    <boardgame-game-state-manager id="manager" .activeRenderer=${this.activeRenderer} .gameRoute=${this._gameRoute} .requestedPlayer=${this.requestedPlayer} .active=${this.selected} .admin=${this._admin} .gameFinished=${this.game ? this.game.Finished : false} .gameVersion=${this.game ? this.game.Version : 0} .loggedIn=${this._loggedIn} .autoCurrentPlayer=${this.autoCurrentPlayer} .viewingAsPlayer=${this.viewingAsPlayer} .socketActive=${this.socketActive} @socket-active-changed=${this._handleSocketActiveChanged}></boardgame-game-state-manager>
`;
  }

  static get properties() {
    return {
      requestedPlayer: { type: Number },
      game: { type: Object },
      currentState: { type: Object },
      _chest: { type: Object },
      _playersInfo: { type: Array },
      _hasEmptySlots: { type: Boolean },
      _open: { type: Boolean },
      _visible: { type: Boolean },
      _isOwner: { type: Boolean },
      autoCurrentPlayer: { type: Boolean },
      selected: { type: Boolean },
      promptedToJoin: { type: Boolean },
      pathsToTick: { type: Array },
      originalWallClockStartTime: { type: Number },
      viewingAsPlayer: { type: Number },
      //The current renderer, passed up from the gameRenderer, so we can pass
      //it to stateGameManager and readyForNextState.
      activeRenderer : { type: Object },
      moveForms: { type: Object },
      socketActive: { type: Boolean },
      _firstStateBundle: { type: Boolean },
      _managerEle: { type: Object },
      _adminEle: { type: Object },
      _renderEle: { type: Object },
      _playerEle: { type: Object },
      _pageExtra: { type: String },
      _gameRoute: { type: Object },
      _loggedIn: { type: Boolean },
      _admin: { type: Boolean },
    }
  }

  //TODO: shouldUpdate should return false if selected is false. But if we do
  //that, then game-state-manager is never updated, so it never lerans that
  //there was a time when it wasn't active. Once game-state-manager is done as
  //action creators then it should be fine.

  stateChanged(state) {
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
  }

  constructor() {
    super();

    this.requestedPlayer = 0;
    this.promptedToJoin = false;
    this._firstStateBundle = true;
    this.viewingAsPlayer = 0;

    this.addEventListener('propose-move', e => this._handleProposeMove(e));
    this.addEventListener('refresh-info', e => this._handleRefreshData(e));
    this.addEventListener('install-state-bundle', e => this._handleStateBundle(e));
    this.addEventListener('install-game-static-info', e => this._handleGameStaticInfo(e));
    this.addEventListener('all-animations-done', e => this._handleAllAnimationsDone(e));
    this.addEventListener('set-animation-length', e => this._handleSetAnimationLength(e));
  }

  _handleRefreshData(e) {
    this._managerEle.fetchInfo();
  }

  _handleRequestedPlayerChanged(e) {
    this.requestedPlayer = e.detail.value;
  }

  _handleAutoCurrentPlayerChanged(e) {
    this.autoCurrentPlayer = e.detail.value;
  }

  _handleSocketActiveChanged(e) {
    this.socketActive = e.detail.value;
  }

  _handleRendererChanged(e) {
    this.activeRenderer = e.detail.value;
  }

  _handleProposeMove(e) {
    this._adminEle.proposeMove(e.detail.name, e.detail.arguments);
  }

  firstUpdated() {
    this._managerEle = this.shadowRoot.querySelector("#manager");
    this._adminEle = this.shadowRoot.querySelector("#admin");
    this._renderEle = this.shadowRoot.querySelector("#render");
    this._playerEle = this.shadowRoot.querySelector("#player");
  }

  updated(changedProps) {
    if (changedProps.has('_pageExtra') && this._page == PAGE_GAME) {
      store.dispatch(updateGameRoute(this._pageExtra));
    }
    if (changedProps.has('selected') && !this.selected) {
      this._resetState();
    }
    if (changedProps.has('_gameRoute')) {
      //reset this so the next time we get data set and notice that we COULD
      //login we prompt for it.
      this.promptedToJoin = false;
      this._resetState();
    }
  }

  _doTick() {
    this._tick();
    if (this.pathsToTick.length > 0) {
      window.requestAnimationFrame(this._doTick.bind(this));
    }
  }

  _tick() {

    if (!this.currentState) return;

    let newPaths = [];

    for (let i = 0; i < this.pathsToTick.length; i++) {
      let currentPath = this.pathsToTick[i];

      let timer = getProperty(this.currentState, currentPath);

      let now = Date.now();
      let difference = now - this.originalWallClockStartTime;

      let result = Math.max(0, timer.originalTimeLeft - difference);

      let newState = deepCopy(this.currentState);

      if (!setProperty(newState, currentPath.concat(["TimeLeft"]), result)) {
        console.warn("Failed to set property: ", newState, currentPath.concat("TimeLeft"), result);
      }

      //this should requestUpdate automatically since it's a copy
      this.currentState = newState;

      //If we still have time to tick on this, then make sure it's still
      //in the list of things to tick.
      if (timer.TimeLeft > 0) {
        newPaths.push(currentPath);
      }
    }

    this.pathsToTick = newPaths;
  }

  _handleStateBundle(e) {
    this._installStateBundle(e.detail);
  }

  _handleGameStaticInfo(e) {
    const bundle = e.detail;
    store.dispatch(updateGameStaticInfo(bundle.chest, bundle.playersInfo, bundle.hasEmptySlots, bundle.open, bundle.visible, bundle.isOwner));
  }

  _handleAllAnimationsDone(e) {
    this._managerEle.readyForNextState();
  }

  _handleSetAnimationLength(e) {
    this._renderEle.defaultAnimationLength = e.detail;
  }

  _firstStateBundleInstalled() {
    if (this.selected && this._loggedIn && this._playerEle.showJoin && !this.promptedToJoin) {

      //Take note that we already prompted them, and don't prompt again unless the game changes.
      this.promptedToJoin = true;
      //Prompt the user to join!
      this._playerEle.showDialog();
    }
  }

  _resetState() {
    this.game = null;
    this.currentState = null;
    this.moveForms = null;
    this.viewingAsPlayer = 0;
    this.originalWallClockStartTime = null;
    this.pathsToTick = null;
    this._firstStateBundle = true;
    this._chest = null;
    this._playersInfo = null;
    this._hasEmptySlots = false;
    this._open = false;
    this._visible = false;
    this._isOwner = false;
    this._firstStateBundle = true;
  }


  _installStateBundle(bundle) {

    //We only rerender once despite setting multiple properties at once
    this.game = bundle.game;
    this.currentState = bundle.game.CurrentState;
    this.moveForms = bundle.moveForms;
    this.viewingAsPlayer = bundle.viewingAsPlayer;
    this.originalWallClockStartTime = bundle.originalWallClockStartTime;
    this.pathsToTick = bundle.pathsToTick;

    if (this._firstStateBundle) {
      this._firstStateBundleInstalled();
    }
    this._firstStateBundle = false;

    window.requestAnimationFrame(() => this._doTick());
  }
}

customElements.define('boardgame-game-view', BoardgameGameView);
