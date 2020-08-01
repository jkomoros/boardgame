/**
@license
Copyright (c) 2016 The Polymer Project Authors. All rights reserved.
This code may only be used under the BSD style license found at http://polymer.github.io/LICENSE.txt
The complete set of authors may be found at http://polymer.github.io/AUTHORS.txt
The complete set of contributors may be found at http://polymer.github.io/CONTRIBUTORS.txt
Code distributed by Google as part of the polymer project is also
subject to an additional IP rights grant found at http://polymer.github.io/PATENTS.txt
*/
/* Statically import dynamic imports for the sake of modulizer */
/* end static linking */
/*
  FIXME(polymer-modulizer): the above comments were extracted
  from HTML and may be out of place here. Review them and
  then delete this comment!
*/
import { PolymerElement } from '@polymer/polymer/polymer-element.js';

import '@polymer/app-layout/app-drawer/app-drawer.js';
import '@polymer/app-layout/app-drawer-layout/app-drawer-layout.js';
import '@polymer/app-layout/app-header/app-header.js';
import '@polymer/app-layout/app-header-layout/app-header-layout.js';
import '@polymer/app-layout/app-scroll-effects/app-scroll-effects.js';
import '@polymer/app-layout/app-toolbar/app-toolbar.js';
import '@polymer/iron-pages/iron-pages.js';
import '@polymer/iron-selector/iron-selector.js';
import '@polymer/paper-icon-button/paper-icon-button.js';
import '@polymer/paper-toggle-button/paper-toggle-button.js';
import '@polymer/paper-dialog/paper-dialog.js';
import '@polymer/paper-button/paper-button.js';
import '@polymer/paper-styles/typography.js';
import '@polymer/paper-styles/default-theme.js';
import './boardgame-user.js';
import './my-icons.js';
import { html } from '@polymer/polymer/lib/utils/html-tag.js';

import { installRouter } from 'pwa-helpers/router.js';

import { connect } from 'pwa-helpers/connect-mixin.js';
import { store } from '../store.js';

import {
  selectPage,
  selectPageExtra
} from '../selectors.js';

import {
  navigated,
  navigatePathTo,
  PAGE_GAME,
} from '../actions/app.js';

class BoardgameApp extends connect(store)(PolymerElement) {
  static get template() {
    return html`
    <style>
      :host {
        --app-primary-color: #4285f4;
        --app-secondary-color: black;

        --paper-button-default-color: var(--app-primary-color);
        --paper-button-default-foreground-color: white;

        display: block;
      }

      [hidden] {
        display:none !important;
      }

      app-header {
        color: #fff;
        background-color: var(--app-primary-color);
      }
      app-header paper-icon-button {
        --paper-icon-button-ink-color: white;
      }

      paper-toggle-button {
        padding: 0 16px;
      }

      .drawer-list {
        margin: 0 20px;
      }

      .drawer-list a {
        display: block;
        padding: 0 16px;
        text-decoration: none;
        color: var(--app-secondary-color);
        line-height: 40px;
      }

      .drawer-list a.iron-selected {
        color: black;
        font-weight: bold;
      }

      #error p.detail {
        color: var(--disabled-text-color);
        @apply --paper-font-caption;
      }
    </style>

    <app-drawer-layout fullbleed="">
      <!-- Drawer content -->
      <app-drawer slot="drawer" id="drawer">
        <boardgame-user id="user" logged-in="{{loggedIn}}" admin-allowed="{{adminAllowed}}"></boardgame-user>
        <paper-toggle-button checked="{{admin}}" hidden="[[!adminAllowed]]">Admin Mode</paper-toggle-button>
        <app-toolbar>Menu</app-toolbar>
        <iron-selector selected="[[_page]]" attr-for-selected="name" class="drawer-list" role="navigation">
          <a name="list-games" href="/list-games">List Games</a>
        </iron-selector>
      </app-drawer>

      <!-- Main content -->
      <app-header-layout has-scrolling-region="">

        <app-header condenses="" reveals="" effects="waterfall">
          <app-toolbar>
            <paper-icon-button icon="my-icons:menu" drawer-toggle=""></paper-icon-button>
            <div main-title="">Boardgame App</div>
          </app-toolbar>
        </app-header>

        <iron-pages selected="[[_page]]" attr-for-selected="name" fallback-selection="view404" selected-attribute="selected" role="main">
          <boardgame-game-view logged-in="[[loggedIn]]" admin="[[admin]]" name="game" game-route="[[_gameRoute]]"></boardgame-game-view>
          <boardgame-list-games-view name="list-games" logged-in="[[loggedIn]]" admin="[[admin]]"></boardgame-list-games-view>
          <boardgame-404-view name="view404"></boardgame-404-view>
        </iron-pages>
      </app-header-layout>
    </app-drawer-layout>
    <paper-dialog id="error">
      <h2>[[errorTitle]]</h2>
      <p>[[friendlyErrorMessage]]</p>
      <p class="detail">[[errorMessage]]</p>
      <div class="buttons">
        <paper-button dialog-dismiss="">OK</paper-button>
      </div>
    </paper-dialog>
`;
  }

  static get is() {
    return "boardgame-app"
  }

  static get properties() {
    return {
      _page: { type: String },
      _pageExtra: { type: String },
      _gameRoute: {
        type: Object,
        computed: "_computeGameRoute(_page, _pageExtra)"
      },
      user: Object,
      loggedIn : Boolean,
      admin: {
        type: Boolean,
        value: false,
      },
      adminAllowed: {
        type: Boolean,
        value: false,
      }
    }
  }

  ready() {
    super.ready();
    this.addEventListener('navigate-to', e => this.handleNavigateTo(e));
    this.addEventListener('show-error', e => this.handleShowError(e));
    this.addEventListener('show-login', e => this.handleShowLogIn(e));
    installRouter((location) => store.dispatch(navigated(decodeURIComponent(location.pathname), decodeURIComponent(location.search))));
  }

  stateChanged(state) {
    this._page = selectPage(state);
    this._pageExtra = selectPageExtra(state);
  }

  //TODO: this logic should live in actions/app.js, and be kicked off by
  //game-view once it's a connected component
  _computeGameRoute(page, pageExtra) {
    if (!page || !pageExtra) return null;
    if (page != PAGE_GAME) return null;
    const pieces = pageExtra.split("/");
    //remove the trailing slash
    if (!pieces[pieces.length - 1]) pieces.pop();
    if (pieces.length != 2) {
      console.warn("URL for game didn't have expected number of pieces");
      return null;
    }
    return {
      name: pieces[0],
      id: pieces[1],
    }
  }

  handleNavigateTo(e) {
    store.dispatch(navigatePathTo(e.detail, false));
  }

  handleShowError(e) {
    let details = e.detail;
    this.showError(details.title, details.friendlyMessage, details.message);
  }

  showError(title, friendlyMessage, message) {
      this.errorTitle = (title || "Error");
      this.friendlyErrorMessage = (friendlyMessage || "There was an error");
      this.errorMessage = (message != friendlyMessage) ? message : "";
      this.$.error.open();
  }

  handleShowLogIn(e) {
    //The event might have things like a nextAction, so forward it.
    this.$.user.showSignInDialog(e);
  }

}

customElements.define(BoardgameApp.is, BoardgameApp);
