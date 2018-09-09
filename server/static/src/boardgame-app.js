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
import '@polymer/app-route/app-location.js';
import '@polymer/app-route/app-route.js';
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

class BoardgameApp extends PolymerElement {
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

    <app-location route="{{route}}"></app-location>
    <app-route route="{{route}}" pattern="/:page" data="{{routeData}}" tail="{{subroute}}"></app-route>
    <app-route route="{{route}}" pattern="/game/:name/:id" data="{{gameRoute}}" tail="{{gameSubRoute}}"></app-route>

    <app-drawer-layout fullbleed="">
      <!-- Drawer content -->
      <app-drawer slot="drawer" id="drawer">
        <boardgame-user id="user" logged-in="{{loggedIn}}" admin-allowed="{{adminAllowed}}"></boardgame-user>
        <paper-toggle-button checked="{{admin}}" hidden="{{!adminAllowed}}">Admin Mode</paper-toggle-button>
        <app-toolbar>Menu</app-toolbar>
        <iron-selector selected="[[page]]" attr-for-selected="name" class="drawer-list" role="navigation">
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

        <iron-pages selected="[[page]]" attr-for-selected="name" fallback-selection="view404" selected-attribute="selected" role="main">
          <boardgame-game-view logged-in="{{loggedIn}}" admin="{{admin}}" name="game" game-route="[[gameRoute]]"></boardgame-game-view>
          <boardgame-list-games-view name="list-games" logged-in="{{loggedIn}}" admin="{{admin}}"></boardgame-list-games-view>
          <boardgame-404-view name="404"></boardgame-404-view>
        </iron-pages>
      </app-header-layout>
    </app-drawer-layout>
    <paper-dialog id="error">
      <h2>{{errorTitle}}</h2>
      <p>{{friendlyErrorMessage}}</p>
      <p class="detail">{{errorMessage}}</p>
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
      page: {
        type: String,
        reflectToAttribute: true,
        observer: '_pageChanged',
      },
      route : Object,
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

  static get observers() {
    return [
      '_routePageChanged(routeData.page)',
    ] 
  }

  ready() {
    super.ready();
    this.addEventListener('navigate-to', e => this.handleNavigateTo(e));
    this.addEventListener('show-error', e => this.handleShowError(e));
    this.addEventListener('show-login', e => this.handleShowLogIn(e));
  }

  handleNavigateTo(e) {
    this.set('route.path',e.detail);
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

  _routePageChanged(page) {
    this.page = page || 'list-games';

    if (!this.$.drawer.persistent) {
      this.$.drawer.close();
    }
  }

  _pageChanged(page) {
    // Load page import on demand. Show 404 page if fails
    import('./boardgame-' + page + '-view.js').then(null, this._showPage404.bind(this));
  }

  _showPage404(err) {
    console.log(err);
    this.page = '404';
  }
}

customElements.define(BoardgameApp.is, BoardgameApp);