import { PolymerElement } from '@polymer/polymer/polymer-element.js';
import './boardgame-ajax.js';
import { html } from '@polymer/polymer/lib/utils/html-tag.js';

/*

  StateManager keeps track of fetching state bundles from the server and
  figuring out when it makes sense to have the game-view install them.

  When the game-view becomes active, the info bundle is feteched. This
  includes information about who the current viewing player is and also
  includes the initial state bundle.

  Once the first state bundle has been installed, a socket is connected to
  receive updates about when the state increases. When the state version
  increases, that increments TargetVersion, which changes the URL to fetch.

*/



class BoardgameGameStateManager extends PolymerElement {
  static get template() {
    return html`
    <boardgame-ajax id="version" game-path="[[effectiveGameVersionPath]]" game-route="[[gameRoute]]" handle-as="json" last-response="{{versionData}}"></boardgame-ajax>
    <boardgame-ajax id="info" game-path="[[gameViewPath]]" game-route="[[gameRoute]]" handle-as="json" last-response="{{infoData}}"></boardgame-ajax>
`;
  }

  static get is() {
    return "boardgame-game-state-manager"
  }

  static get properties() {
    return {
      gameRoute: Object,
      gameFinished: {
        type: Boolean,
        value: false,
      },
      chest: Object,
      admin: Boolean,
      autoCurrentPlayer: Boolean,
      active: {
        type: Boolean,
        observer: "_activeChanged"
      },
      loggedIn: {
        type: Boolean,
        observer: "_loggedInChanged",
      },
      targetVersion: {  
        type: Number,
        value: -1
      },
      gameVersionPath: {
        type: String,
        computed: "_computeGameVersionPath(active, requestedPlayer, admin, targetVersion, autoCurrentPlayer)",
        observer: "_gameVersionPathChanged"
      },
      gameViewPath : {
        type: String,
        computed: "_computeGameViewPath(requestedPlayer, admin, lastFetchedVersion)"
      },
      gameBasePath : String,
      //This is split out because lastFetchedVersion should be current
      //when it's sent, but when its value is changed it shouldn't be
      //considered a meaningful change that needs a refetch.
      effectiveGameVersionPath: {
        type: String,
        computed: "_computeEffectiveGameVersionPath(gameVersionPath, lastFetchedVersion, version)",
      },
      viewingAsPlayer: Number,
      requestedPlayer: {
        type: Number,
        value: 0
      },
      gameVersion: Number,
      //lastFetchedVersion is the last version we've fetched from the server.
      lastFetchedVersion: {
        type: Number,
        value: 0,
      },
      infoData: {
        type: Object,
        observer: '_infoDataChanged',
      },
      versionData: {
        type: Object,
        observer: "_versionDataChanged"
      },
      socketActive: {
        type: Boolean,
        notify: true,
        value: false,
      },
      _socketUrl: {
        type: String,
        computed: "_computeSocketUrl(active, _infoInstalled)",
        observer: "_socketUrlChanged",
      },
      _infoInstalled: {
        type: Boolean,
        value: false,
      },
      _socket: Object,
      _pendingStateBundles: Object,
      _stateBundleTimeout: Number,
      _stateBundleAnimationFrame: Number,
    }
  }

  ready() {
    super.ready();

    this._pendingStateBundles = [];

    this.updateData();
  }

  _computeEffectiveGameVersionPath(gameVersionPath, lastFetchedVersion, version) {
    if (!gameVersionPath) return "";
    //version is already part of gameVersionPath. However, often on first
    //load, version and lastFetchedVersion are the same, and we should skip
    //fetching because we already have that info. However in some cases the
    //info bundle will not have all of the most up to date stuff, and we still
    //do need to fetch.
    if (lastFetchedVersion == version) return "";
    return gameVersionPath + "&from=" + lastFetchedVersion
  }

  _computeGameVersionPath(active, requestedPlayer, admin, version, autoCurrentPlayer) {
    if (!active) return "";
    if (version < 0) return "";
    //TODO: factor this out with computeGameViewUrl a bit
    return "version/" + version + "?player=" + requestedPlayer+"&admin=" + (admin ? 1 : 0) + "&current=" + (autoCurrentPlayer ? 1 : 0);
  }

  _computeGameViewPath(requestedPlayer, admin, lastFetchedVersion){
    return "info?player=" + requestedPlayer+"&admin=" + (admin ? 1 : 0) + "&from=" + lastFetchedVersion;
  }

  _computeSocketUrl(active, infoInstalled) {
    if (!active) return "";
    if (!infoInstalled) return "";
    let result = this.$.version.gameBasePath + "socket";
    result = result.split("http:").join("ws:");
    result = result.split("https:").join("wss:");
    return result;
  }

  _loggedInChanged(newValue) {
    this.softReset();
  }

  _activeChanged(newValue) {
    if (newValue) {
      this.reset();
    } else {
      //If we don't clear this out when we deactivate then when we become
      //re-active there might be a brief period where our gameRoute is the
      //old one.
      this.gameRoute = null;
    }
  }

  _gameVersionPathChanged(newValue, oldValue) {
    if (!newValue) return;

    if (this.autoCurrentPlayer && this.requestedPlayer == this.viewingAsPlayer && this.targetVersion == this.gameVersion) {
      return
    }

    //TODO: the autoCurrent player stuff has to be done here...
    requestAnimationFrame(() => this.$.version.generateRequest());
  }

  _socketUrlChanged(newValue) {
    if (this._socket) {
      this._socket.close();
      this._socket = "";
    }

    this._connectSocket();

  }

  _connectSocket() {

    var theUrl = this._socketUrl;
    
    //Ifthere's no URL, don't establish a socket.
    if (!theUrl) return;

    this._socket = new WebSocket(theUrl);

    this._socket.onclose = e => this._socketClosed(e);
    this._socket.onerror = e => this._socketError(e);
    this._socket.onmessage = e => this._socketMessage(e);
    this._socket.onopen = e => this._socketOpened(e);
  }

  _socketMessage(e) {
    let version = parseInt(e.data);
    if (isNaN(version)) {
      return;
    }
    this.targetVersion = version;
  }

  _socketError(e) {
    //TOOD: do something more substantive
    console.warn("Socket error", e)
  }

  _socketOpened(e) {
    this.socketActive = true;
  }

  _socketClosed(e) {
    console.warn("Socket closed", e);
    this.socketActive = false;
    //We alawyas want a socket, so connect. Wait a bit so we don't just
    //busy spin if the server is down.

    //If we closed because we no longer have a valid URL, then
    //_connectSocket will just exit, and this loop won't be called.

    //TOOD: exponentional backoff on server connect.
    setTimeout(() => this._connectSocket(), 250);
  }

  updateData() {
    this.fetchInfo();
  }

  //When we should do a soft reset; that is, when we haven't flipped out and
  //back; it's still the same game we're viewing as before.
  softReset() {
    this.infoData = null;
    this._infoInstalled = false;
    window.requestAnimationFrame(() => this.updateData());
  }

  //When evertyhing should be reset
  reset() {
    this.lastFetchedVersion = 0;
    this._resetPendingStateBundles();
    this.softReset();
  }

  fetchInfo() {
    if (this.$.info.loading) {
      return
    }

    if (!this.active) {
      return
    }

    if (!this.gameRoute) {
      //The URL will be junk
      return
    }
    this.$.info.generateRequest();
  }

  _prepareStateBundle(game, moveForms, viewingAsPlayer, delay) {


    var bundle = {};

    bundle.gameState = JSON.stringify(game.CurrentState, null, 2);
    //Have to use this.set to get databinding system to notice the change.

    bundle.originalWallClockStartTime = Date.now();

    bundle.pathsToTick = this._expandState(game.CurrentState, game.ActiveTimers);

    bundle.game = game;
    bundle.moveForms = this._expandMoveForms(moveForms);
    bundle.viewingAsPlayer = viewingAsPlayer;
    bundle.delay = delay;

    return bundle;
  }

  _expandMoveForms(moveForms) {
    if (!moveForms) return null;
    for (let i = 0; i < moveForms.length; i++){
      let form = moveForms[i];
      //Some forms don't have fields and that's OK.
      if (!form.Fields) continue;
      for (let j = 0; j < form.Fields.length; j++) {
        let field = form.Fields[j];
        if (field.EnumName) {
          field.Enum = this.chest.Enums[field.EnumName];
        }
      }
    }
    return moveForms;
  }

  _expandState(currentState, timerInfos) {
    //Takes the currentState and returns an object where all of the Stacks are replaced by actual references to the component they reference.

    var pathsToTick = [];


    this._expandLeafState(currentState, currentState.Game, ["Game"], pathsToTick, timerInfos)
    for (var i = 0; i < currentState.Players.length; i++) {
      this._expandLeafState(currentState, currentState.Players[i], ["Players", i], pathsToTick, timerInfos)
    }

    return pathsToTick;

  }

  _expandLeafState(wholeState, leafState, pathToLeaf, pathsToTick, timerInfos) {
    //Returns an expanded version of leafState. leafState should have keys that are either bools, floats, strings, or Stacks.
    
    var entries = Object.entries(leafState);
    for (var i = 0; i < entries.length; i++) {
      let item = entries[i];
      let key = item[0];
      let val = item[1];
      //Note: null is typeof "object"
      if (val && typeof val == "object") {
        if (val.Deck) {
          this._expandStack(val, wholeState);
        } else if (val.IsTimer) {
          this._expandTimer(val, pathToLeaf.concat([key]), pathsToTick, timerInfos);
        }   
      }
    }

    //Copy in Player computed state if it exists, for convenience. Do it after expanding properties
    if (pathToLeaf && pathToLeaf.length == 2 && pathToLeaf[0] == "Players") {
      if (wholeState.Computed && wholeState.Computed.Players && wholeState.Computed.Players.length) {
        leafState.Computed = wholeState.Computed.Players[pathToLeaf[1]];
      }
    }
  }

  _expandStack(stack, wholeState) {
    if (!stack.Deck) {
      //Meh, I guess it's not a stack
      return;
    }

    var deck = this.chest.Decks[stack.Deck];

    var gameName = (this.gameRoute) ? this.gameRoute.name : "";

    var components = [];

    for (var i = 0; i < stack.Indexes.length; i++) {
      let index = stack.Indexes[i];
      if (index == -1) {
        components[i] = null;
        continue;
      }

      if(index == -2) {
        //TODO: to handle this appropriately we'd need to know how to
        //produce a GenericComponent for each Deck clientside.
        components[i] = {};
      } else {
        components[i] = this._componentForDeckAndIndex(stack.Deck, index, wholeState);
      }
      
      if (stack.Ids) {
        components[i].Id = stack.Ids[i];
      }
      components[i].Deck = stack.Deck;
      components[i].GameName = gameName;
    }

    stack.GameName = gameName;

    stack.Components = components;

  }

  _expandTimer(timer, pathToLeaf, pathsToTick, timerInfo) {

    //Always make sure these default to a number so databinding can use them.
    timer.TimeLeft = 0;
    timer.originalTimeLeft = 0;

    if (!timerInfo) return;

    let info = timerInfo[timer.Id];

    if (!info) return;
    timer.TimeLeft = info.TimeLeft;
    timer.originalTimeLeft = timer.TimeLeft;
    pathsToTick.push(pathToLeaf);
  }


  _componentForDeckAndIndex(deckName, index, wholeState) {
    let deck = this.chest.Decks[deckName];

    if (!deck) return null;

    let result = this._copyObj(deck[index]);

    if (wholeState && wholeState.Components) {
      if (wholeState.Components[deckName]) {
        result.DynamicValues = wholeState.Components[deckName][index];
      }
    }

    return result

  }

  _copyObj(obj) {
    let copy = {}
    for (let attr in obj) {
      if (obj.hasOwnProperty(attr)) copy[attr] = obj[attr]
    }
    return copy
  }

  //Called when gameView tells us to pass up the next state if we have one
  //(the animations are done).
  readyForNextState() {
    console.log("Ready for next state called");
  }

  //A new state bundle has been enqueued. Ensure that we're working ot fire a state bundle.
  _scheduleNextStateBundle() {

    if (this._stateBundleTimeout) return;
    if (this._stateBundleAnimationFrame) return;
    if (!this._pendingStateBundles.length) return;

    let nextBundle = this._pendingStateBundles[0];

    if (!nextBundle.delay) {
      //Fire it immediately but int he next micro task
      this._stateBundleAnimationFrame = window.requestAnimationFrame(() => this._fireNextStateBundle());
      return;
    }

    this._stateBundleTimeout = setTimeout(() => this._fireNextStateBundle(), nextBundle.delay);

  }

  _resetPendingStateBundles() {
    if (this._stateBundleTimeout) clearTimeout(this._stateBundleTimeout);
    this._stateBundleTimeout = 0;
    if (this._stateBundleAnimationFrame) window.cancelAnimationFrame(this._stateBundleAnimationFrame);
    this._stateBundleAnimationFrame = 0;
    this._pendingStateBundles = [];
  }

  _fireNextStateBundle() {
    //Called when the next state bundle should be installed NOW.
    this._stateBundleTimeout = 0;
    this._stateBundleAnimationFrame = 0;
    let bundle = this._pendingStateBundles.shift();
    if (bundle) {
      this.dispatchEvent(new CustomEvent('install-state-bundle', {composed: true, detail: bundle}));
    }
    this._scheduleNextStateBundle();
  }

  //Add the next state bundle to the end
  _enqueueStateBundle(bundle) {
    this._pendingStateBundles.push(bundle);
    this._scheduleNextStateBundle();    
  }

  _infoDataChanged(newValue, oldValue) {
    if (!newValue) {
      //Sometimes we set null, like when we select the view.
      return
    }

    this.chest = newValue.Chest;

    var gameInfo = {
      chest: newValue.Chest,
      playersInfo: newValue.Players,
      hasEmptySlots: newValue.HasEmptySlots,
      gameOpen: newValue.GameOpen,
      gameVisible: newValue.GameVisible,
      isOwner: newValue.IsOwner,
    }

    this.dispatchEvent(new CustomEvent("install-game-static-info", {composed: true, detail: gameInfo}))

    var bundle = this._prepareStateBundle(newValue.Game, newValue.Forms, newValue.ViewingAsPlayer, 0);
    this._enqueueStateBundle(bundle);

    this._infoInstalled = true;

    //We don't use newValue.Game.Version, because in some cases the current
    //state we're returning is not actually current state, but an old one to
    //force us to play animations for moves that are made before a player move
    //is.
    this.lastFetchedVersion = newValue.Game.CurrentState.Version;
    this.targetVersion = newValue.Game.Version;
  }

  _versionDataChanged(newValue) {
    if (!newValue) return;
    if (newValue.Error) {
      console.log("Version getter returned error: " + newValue.Error)
      return
    }

    let lastServerBundle = {};

    for (let i = 0; i < newValue.Bundles.length; i++) {
      let serverBundle = newValue.Bundles[i];
      let bundle = this._prepareStateBundle(serverBundle.Game, serverBundle.Forms, serverBundle.ViewingAsPlayer, serverBundle.Delay);
      this._enqueueStateBundle(bundle);
      lastServerBundle = serverBundle;
    }

    this.lastFetchedVersion = lastServerBundle.Game.Version;
  }
}

customElements.define(BoardgameGameStateManager.is, BoardgameGameStateManager);
