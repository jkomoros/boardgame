import '@polymer/polymer/polymer-element.js';
import '@polymer/iron-ajax/iron-ajax.js';
import { GamePathMixin } from './boardgame-game-path.js';

// Declare global API_HOST (defined in index.html or build config)
declare const API_HOST: string;

// Get the IronAjax constructor
const IronAjaxElement = customElements.get('iron-ajax') as any;

/**
 * BoardgameAjax extends IronAjax with game-specific URL construction.
 * Automatically constructs URLs for game-specific and general API endpoints.
 *
 * Usage:
 * - For general API calls: set `path` property (e.g., "auth")
 * - For game-specific calls: set `gamePath` property (e.g., "move")
 */
export class BoardgameAjax extends GamePathMixin(IronAjaxElement) {
  static get is(): string {
    return 'boardgame-ajax';
  }

  static get properties() {
    return {
      gameRoute: {
        type: Object
      },

      url: {
        type: String,
        computed: '_computeUrl(path, gamePath, basePath, gameBasePath)'
      },

      // e.g., "http://api.boardgame.com/api/"
      basePath: {
        type: String,
        value: typeof API_HOST !== 'undefined' ? API_HOST + '/api/' : '/api/',
        readOnly: true
      },

      // e.g., "http://api.boardgame.com/api/game/memory/12345/"
      gameBasePath: {
        type: String,
        computed: '_computeGameBasePath(basePath, gameRoute)'
      },

      // Path for APIs that aren't specific to a game, e.g., /api/auth,
      // where path would be 'auth'
      path: {
        type: String
      },

      // Path for APIs that are part of a game, e.g.,
      // /api/game/blackjack/123445/move, where gamePath would be 'move'
      gamePath: {
        type: String
      },

      // You almost always want withCredentials so set it to true
      withCredentials: {
        type: Boolean,
        value: true
      }
    };
  }

  // Property declarations for TypeScript
  gameRoute: { name: string; id: string } | null = null;
  url = '';
  basePath = '';
  gameBasePath = '';
  path = '';
  gamePath = '';
  withCredentials = true;

  private _computeGameBasePath(basePath: string, gameRoute: { name: string; id: string } | null): string {
    if (!gameRoute) {
      return '';
    }
    return basePath + this.GamePath(gameRoute.name, gameRoute.id);
  }

  private _computeUrl(path: string, gamePath: string, basePath: string, gameBasePath: string): string {
    if (path) {
      return basePath + path;
    }
    if (gamePath && gameBasePath) {
      return gameBasePath + gamePath;
    }
    return '';
  }
}

customElements.define(BoardgameAjax.is, BoardgameAjax as any);
