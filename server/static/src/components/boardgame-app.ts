import '@material/web/iconbutton/icon-button.js';
import '@material/web/dialog/dialog.js';
import '@material/web/button/filled-button.js';
import '@material/web/switch/switch.js';
import './boardgame-user.js';
import './my-icons.js';

import { LitElement, html, css } from 'lit';
import { customElement, property } from 'lit/decorators.js';

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

@customElement('boardgame-app')
export class BoardgameApp extends connect(store)(LitElement) {
  static styles = css`
    :host {
      display: block;
      height: 100vh;
      display: flex;
    }

    [hidden] {
      display: none !important;
    }

    /* Layout structure */
    .app-layout {
      display: flex;
      width: 100%;
      height: 100%;
      position: relative;
      background-color: var(--md-sys-color-background);
    }

    /* Drawer styles - Material Design 3 */
    .drawer {
      width: 280px;
      background: var(--md-sys-color-surface-container-low);
      border-right: 1px solid var(--md-sys-color-outline-variant);
      display: flex;
      flex-direction: column;
      overflow-y: auto;
      position: fixed;
      left: 0;
      top: 0;
      bottom: 0;
      transform: translateX(-100%);
      transition: transform 0.3s cubic-bezier(0.4, 0, 0.2, 1);
      z-index: 100;
      box-shadow: var(--md-sys-elevation-2);
    }

    .drawer.open {
      transform: translateX(0);
    }

    .drawer-backdrop {
      position: fixed;
      top: 0;
      left: 0;
      right: 0;
      bottom: 0;
      background: var(--md-sys-color-scrim);
      opacity: 0.32;
      z-index: 99;
      display: none;
    }

    .drawer-backdrop.visible {
      display: block;
    }

    /* Admin toggle */
    .admin-toggle {
      padding: 12px 16px;
      display: flex;
      align-items: center;
      gap: 12px;
      font-family: var(--md-sys-typescale-body-medium-font);
      font-size: var(--md-sys-typescale-body-medium-size);
      line-height: var(--md-sys-typescale-body-medium-line-height);
      color: var(--md-sys-color-on-surface);
    }

    .drawer-toolbar {
      padding: 24px 16px 16px 16px;
      font-family: var(--md-sys-typescale-title-large-font);
      font-size: var(--md-sys-typescale-title-large-size);
      line-height: var(--md-sys-typescale-title-large-line-height);
      font-weight: var(--md-sys-typescale-title-large-weight);
      color: var(--md-sys-color-on-surface);
      border-bottom: 1px solid var(--md-sys-color-outline-variant);
    }

    /* Navigation list - Material Design 3 */
    .drawer-list {
      margin: 8px;
      padding: 0;
    }

    .drawer-list a {
      display: flex;
      align-items: center;
      padding: 12px 16px;
      margin: 4px 0;
      text-decoration: none;
      color: var(--md-sys-color-on-surface);
      font-family: var(--md-sys-typescale-label-large-font);
      font-size: var(--md-sys-typescale-label-large-size);
      line-height: var(--md-sys-typescale-label-large-line-height);
      font-weight: var(--md-sys-typescale-label-large-weight);
      border-radius: 28px;
      cursor: pointer;
      transition: background-color 0.2s ease;
    }

    .drawer-list a:hover {
      background-color: var(--md-sys-color-surface-container-highest);
    }

    .drawer-list a.selected {
      background-color: var(--md-sys-color-secondary-container);
      color: var(--md-sys-color-on-secondary-container);
      font-weight: 700;
    }

    /* Main content area */
    .main-content {
      flex: 1;
      display: flex;
      flex-direction: column;
      width: 100%;
    }

    /* Header styles - Material Design 3 Top App Bar */
    .app-header {
      background-color: var(--md-sys-color-surface);
      color: var(--md-sys-color-on-surface);
      box-shadow: var(--md-sys-elevation-2);
      position: sticky;
      top: 0;
      z-index: 10;
    }

    .app-toolbar {
      display: flex;
      align-items: center;
      padding: 0 16px;
      height: 64px;
    }

    .app-toolbar md-icon-button {
      --md-icon-button-icon-color: var(--md-sys-color-on-surface);
    }

    .app-title {
      flex: 1;
      font-family: var(--md-sys-typescale-title-large-font);
      font-size: var(--md-sys-typescale-title-large-size);
      line-height: var(--md-sys-typescale-title-large-line-height);
      font-weight: var(--md-sys-typescale-title-large-weight);
      color: var(--md-sys-color-on-surface);
      margin-left: 16px;
    }

    /* Content area */
    .content-area {
      flex: 1;
      overflow-y: auto;
      position: relative;
      background-color: var(--md-sys-color-background);
    }

    /* Page visibility */
    .page {
      display: none;
    }

    .page.selected {
      display: block;
    }

    /* Error dialog styles */
    md-dialog {
      --md-dialog-container-color: var(--md-sys-color-surface-container-high);
    }

    md-dialog h2 {
      margin: 0 0 16px 0;
      font-family: var(--md-sys-typescale-headline-small-font);
      font-size: var(--md-sys-typescale-headline-small-size);
      line-height: var(--md-sys-typescale-headline-small-line-height);
      font-weight: var(--md-sys-typescale-headline-small-weight);
      color: var(--md-sys-color-on-surface);
    }

    md-dialog p {
      margin: 8px 0;
      font-family: var(--md-sys-typescale-body-medium-font);
      font-size: var(--md-sys-typescale-body-medium-size);
      line-height: var(--md-sys-typescale-body-medium-line-height);
      color: var(--md-sys-color-on-surface);
    }

    md-dialog p.detail {
      color: var(--md-sys-color-on-surface-variant);
      font-size: var(--md-sys-typescale-body-small-size);
    }

    md-dialog .buttons {
      display: flex;
      justify-content: flex-end;
      gap: 8px;
      margin-top: 16px;
    }

    /* Responsive - show drawer on larger screens */
    @media (min-width: 768px) {
      .drawer {
        position: static;
        transform: translateX(0);
      }

      .drawer-backdrop {
        display: none !important;
      }

      .menu-button {
        display: none;
      }
    }
  `;

  @property({ type: String })
  private _page = '';

  @property({ type: Boolean })
  private _errorShowing = false;

  @property({ type: String })
  private _errorMessage = '';

  @property({ type: String })
  private _errorFriendlyMessage = '';

  @property({ type: String })
  private _errorTitle = '';

  @property({ type: Boolean })
  private _adminAllowed = false;

  @property({ type: Boolean })
  private _admin = false;

  @property({ type: Boolean })
  private _drawerOpen = false;

  firstUpdated() {
    this.addEventListener('navigate-to', (e: Event) => this._handleNavigateTo(e as CustomEvent));
    this.addEventListener('show-error', (e: Event) => this._handleShowError(e as CustomEvent));
    this.addEventListener('show-login', (e: Event) => this._handleShowLogIn(e as CustomEvent));
    installRouter((location) => store.dispatch(navigated(decodeURIComponent(location.pathname), decodeURIComponent(location.search))));
  }

  stateChanged(state: any): void {
    this._page = selectPage(state);
    this._errorShowing = selectErrorShowing(state);
    this._errorTitle = selectErrorTitle(state);
    this._errorMessage = selectErrorMessage(state);
    this._errorFriendlyMessage = selectErrorFriendlyMessage(state);
    this._adminAllowed = selectAdminAllowed(state);
    this._admin = selectAdmin(state);
  }

  private _handleAdminChanged(e: Event): void {
    const target = e.target as any;
    store.dispatch(setUserAdmin(target.selected));
  }

  private _handleNavigateTo(e: CustomEvent): void {
    store.dispatch(navigatePathTo(e.detail, false));
  }

  private _handleShowError(e: CustomEvent): void {
    const details = e.detail;
    store.dispatch(updateAndShowError(details.title, details.friendlyMessage, details.message));
  }

  private _handleDialogDismissTapped(): void {
    store.dispatch(hideError());
  }

  private _handleDialogClosed(e: Event): void {
    const dialog = e.target as any;
    // When the dialog is canceled by clicking on background or esc
    if (dialog.returnValue === '' || dialog.returnValue === 'close') {
      store.dispatch(hideError());
    }
  }

  private _handleShowLogIn(e: CustomEvent): void {
    // Might be undefined, that's fine
    setSignedInAction(e.detail.nextAction);
    store.dispatch(showSignInDialog());
  }

  private _toggleDrawer(): void {
    this._drawerOpen = !this._drawerOpen;
  }

  private _closeDrawer(): void {
    this._drawerOpen = false;
  }

  private _handleNavClick(e: Event): void {
    e.preventDefault();
    const target = e.currentTarget as HTMLAnchorElement;
    const path = target.getAttribute('href');
    if (path) {
      window.history.pushState({}, '', path);
      store.dispatch(navigated(decodeURIComponent(path), ''));
    }
    this._closeDrawer();
  }

  render() {
    return html`
      <div class="app-layout">
        <!-- Drawer backdrop -->
        <div
          class="drawer-backdrop ${this._drawerOpen ? 'visible' : ''}"
          @click="${this._closeDrawer}">
        </div>

        <!-- Drawer -->
        <div class="drawer ${this._drawerOpen ? 'open' : ''}">
          <boardgame-user id="user"></boardgame-user>

          ${this._adminAllowed ? html`
            <div class="admin-toggle">
              <md-switch
                ?selected="${this._admin}"
                @change="${this._handleAdminChanged}">
              </md-switch>
              <span>Admin Mode</span>
            </div>
          ` : ''}

          <div class="drawer-toolbar">Menu</div>

          <nav class="drawer-list">
            <a
              href="/list-games"
              class="${this._page === 'list-games' ? 'selected' : ''}"
              @click="${this._handleNavClick}">
              List Games
            </a>
          </nav>
        </div>

        <!-- Main content -->
        <div class="main-content">
          <header class="app-header">
            <div class="app-toolbar">
              <md-icon-button
                class="menu-button"
                @click="${this._toggleDrawer}">
                <md-icon>menu</md-icon>
              </md-icon-button>
              <div class="app-title">Boardgame App</div>
            </div>
          </header>

          <main class="content-area">
            <boardgame-game-view
              class="page ${this._page === 'game' ? 'selected' : ''}"
              ?selected="${this._page === 'game'}">
            </boardgame-game-view>

            <boardgame-list-games-view
              class="page ${this._page === 'list-games' ? 'selected' : ''}"
              ?selected="${this._page === 'list-games'}">
            </boardgame-list-games-view>

            <boardgame-404-view
              class="page ${this._page === 'view404' || (!this._page && this._page !== 'game' && this._page !== 'list-games') ? 'selected' : ''}"
              ?selected="${this._page === 'view404'}">
            </boardgame-404-view>
          </main>
        </div>
      </div>

      <!-- Error Dialog -->
      <md-dialog
        ?open="${this._errorShowing}"
        @closed="${this._handleDialogClosed}">
        <div slot="headline">${this._errorTitle}</div>
        <div slot="content">
          <p>${this._errorFriendlyMessage}</p>
          <p class="detail">${this._errorMessage}</p>
        </div>
        <div slot="actions">
          <md-filled-button @click="${this._handleDialogDismissTapped}">
            OK
          </md-filled-button>
        </div>
      </md-dialog>
    `;
  }
}

declare global {
  interface HTMLElementTagNameMap {
    'boardgame-app': BoardgameApp;
  }
}
