import '@polymer/polymer/polymer-element.js';
import '@polymer/iron-ajax/iron-ajax.js';
import {GamePathMixin} from './boardgame-game-path.js';

let IronAjax = customElements.get("iron-ajax");

class BoardgameAjax extends GamePathMixin(IronAjax) {

  static get is() {
    return "boardgame-ajax";
  }

  static get properties() {
    return {
      gameRoute: Object,
      url : {
        type: String,
        computed: "_computeUrl(path, gamePath, basePath, gameBasePath)"
      },

      //e.g. "http://api.boardgame.com/api/"
      basePath : {
        type: String,
        value: API_HOST + "/api/",
        readOnly: true
      },

      //e.g. "http://api.boardgame.com/api/game/memory/12345/"
      gameBasePath: {
        type: String,
        computed: "_computeGameBasePath(basePath,gameRoute)"
      },

      //Path for apis that aren't specific to a game, e.g. /api/auth, where
      //path would be 'auth'
      path: String,

      //Path for APIS that are part of a game, e.g.
      ///api/game/blackjack/123445/move, where gamePath would be `move`
      gamePath: String,
      //You almost always want withCredentials so set it to true.
      withCredentials: {
        type: Boolean,
        value: true,
      }
    }
  }

  _computeGameBasePath(basePath, gameRoute) {
    if (!gameRoute) {
      return "";
    }
    return basePath + this.GamePath(gameRoute.name, gameRoute.id);
  }

  _computeUrl(path, gamePath, basePath, gameBasePath) {
    if (path) {
      return basePath + path;
    }
    if (gamePath && gameBasePath) {
      return gameBasePath + gamePath;
    }
    return "";
  }

}

customElements.define(BoardgameAjax.is, BoardgameAjax);
