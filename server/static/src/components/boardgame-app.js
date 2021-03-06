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

import { LitElement, html } from '@polymer/lit-element';

import { installRouter } from 'pwa-helpers/router.js';

import { connect } from 'pwa-helpers/connect-mixin.js';
import { store } from '../store.js';

import {
  selectPage,
  selectErrorShowing,
  selectErrorMessage,
  selectErrorFriendlyMessage,
  selectErrorTitle,
  selectAdminAllowed,
  selectAdmin
} from '../selectors.js';

import {
  navigated,
  navigatePathTo,
} from '../actions/app.js';

import {
  updateAndShowError,
  hideError
} from '../actions/error.js';

import {
  setUserAdmin,
  setSignedInAction,
  showSignInDialog
} from '../actions/user.js';

class BoardgameApp extends connect(store)(LitElement) {
  render() {
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
        <boardgame-user id="user"></boardgame-user>
        <paper-toggle-button .checked=${this._admin} @checked-changed=${this._handleAdminCheckedChanged} ?hidden=${!this._adminAllowed}>Admin Mode</paper-toggle-button>
        <app-toolbar>Menu</app-toolbar>
        <iron-selector .selected=${this._page} .attr-for-selected="name" class="drawer-list" role="navigation">
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

        <iron-pages .selected=${this._page} attr-for-selected="name" fallback-selection="view404" selected-attribute="selected" role="main">
          <boardgame-game-view name="game"></boardgame-game-view>
          <boardgame-list-games-view name="list-games"></boardgame-list-games-view>
          <boardgame-404-view name="view404"></boardgame-404-view>
        </iron-pages>
      </app-header-layout>
    </app-drawer-layout>
    <paper-dialog id="error" @opened-changed=${this._handleDialogOpenedChanged} .opened=${this._errorShowing}>
      <h2>${this._errorTitle}</h2>
      <p>${this._errorFriendlyMessage}</p>
      <p class="detail">${this._errorMessage}</p>
      <div class="buttons">
        <paper-button @tap=${this._handleDialogDismissTapped}>OK</paper-button>
      </div>
    </paper-dialog>
`;
  }

  static get properties() {
    return {
      _page: { type: String },
      _errorShowing: { type: Boolean },
      _errorMessage: { type: String },
      _errorFriendlyMessage: { type: String },
      _errorTitle: { type: String },
      _adminAllowed: { type: Boolean },
      _admin: { type: Boolean },
    }
  }

  firstUpdated() {
    this.addEventListener('navigate-to', e => this._handleNavigateTo(e));
    this.addEventListener('show-error', e => this._handleShowError(e));
    this.addEventListener('show-login', e => this._handleShowLogIn(e));
    installRouter((location) => store.dispatch(navigated(decodeURIComponent(location.pathname), decodeURIComponent(location.search))));
  }

  stateChanged(state) {
    this._page = selectPage(state);
    this._errorShowing = selectErrorShowing(state);
    this._errorTitle = selectErrorTitle(state);
    this._errorMessage = selectErrorMessage(state);
    this._errorFriendlyMessage = selectErrorFriendlyMessage(state);
    this._adminAllowed = selectAdminAllowed(state);
    this._admin = selectAdmin(state);
  }

  _handleAdminCheckedChanged(e) {
    store.dispatch(setUserAdmin(e.detail.value));
  }

  _handleNavigateTo(e) {
    store.dispatch(navigatePathTo(e.detail, false));
  }

  _handleShowError(e) {
    let details = e.detail;
    store.dispatch(updateAndShowError(details.title, details.friendlyMessage, details.message));
  }

  _handleDialogDismissTapped() {
    store.dispatch(hideError());
  }

  _handleDialogOpenedChanged(e) {
    //When the dialog is canceled by clicking on background or esc, we get a
    //different way. we coul dhave the dismiss button have `dialog-dismiss`, but
    //then it's not possible to distinguish that being clicked and just the
    //first time being updated
    if (!e.detail.value) {
      const dialogEle = this.shadowRoot.querySelector("#error");
      if (dialogEle && !dialogEle.closingReason.canceled) return;
      store.dispatch(hideError());
    }
  }

  _handleShowLogIn(e) {
      //Might be undefined, that's fine
      setSignedInAction(e.detail.nextAction);
      store.dispatch(showSignInDialog());
  }

}

customElements.define("boardgame-app", BoardgameApp);
