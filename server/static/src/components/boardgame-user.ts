import { LitElement, html, css } from 'lit';
import { customElement, property } from 'lit/decorators.js';
import { when } from 'lit/directives/when.js';
import '@material/web/dialog/dialog.js';
import '@material/web/button/filled-button.js';
import '@material/web/button/outlined-button.js';
import '@material/web/textfield/filled-text-field.js';
import '@material/web/progress/linear-progress.js';
import './boardgame-player-chip.ts';

import { connect } from 'pwa-helpers/connect-mixin.js';
import { store } from '../store.js';

import { OFFLINE_DEV_MODE } from '../actions/app.js';

import {
  firebaseSignIn,
  signOut,
  signInWithGoogle,
  signInOrCreateWithEmailAndPassword,
  showSignInDialog,
  updateSignInDialogEmail,
  updateSignInDialogPassword,
  showSignInDialogEmailPage
} from '../actions/user.js';

import {
  selectUser,
  selectVerifyingAuth,
  selectSignInErrorMessage,
  selectSignInDialogOpen,
  selectSignInDialogEmail,
  selectSignInDialogPassword,
  selectSignInDialogIsCreate,
  selectSignInDialogSelectedPage
} from '../selectors.js';

import type { UserInfo } from '../types/store';

@customElement('boardgame-user')
export class BoardgameUser extends connect(store)(LitElement) {
  static styles = css`
    :host {
      display: block;
      position: relative;
      padding: 16px;
    }

    a {
      cursor: pointer;
      color: var(--accent-color, #ff4081);
      text-decoration: none;
    }

    a:hover {
      text-decoration: underline;
    }

    md-dialog {
      min-width: 300px;
      min-height: 300px;
    }

    .verifying {
      font-style: italic;
    }

    .horizontal {
      display: flex;
      flex-direction: row;
    }

    .vertical {
      display: flex;
      flex-direction: column;
    }

    #offline {
      display: none;
      height: 5px;
      width: 5px;
      top: 16px;
      left: 16px;
      position: absolute;
      border-radius: 2.5px;
      background-color: red;
    }

    .offline #offline {
      display: block;
    }

    .page {
      display: none;
    }

    .page.selected {
      display: block;
    }

    .layout.vertical {
      gap: 8px;
    }

    md-filled-button,
    md-outlined-button {
      width: 100%;
    }

    md-filled-text-field {
      width: 100%;
      margin-bottom: 16px;
    }

    .buttons {
      display: flex;
      justify-content: flex-end;
      gap: 8px;
      margin-top: 16px;
    }

    .card {
      background: var(--md-sys-color-surface-container-low, #f7f2fa);
      padding: 16px;
      margin: 8px;
      border-radius: 12px;
      box-shadow: var(--md-sys-elevation-1, 0 1px 3px 1px rgba(0,0,0,.15), 0 1px 2px rgba(0,0,0,.3));
      color: var(--md-sys-color-on-surface, #1c1b1f);
    }

    md-linear-progress {
      width: 100%;
      margin-top: 16px;
    }
  `;

  @property({ type: Boolean })
  private _verifyingAuth = false;

  @property({ type: Object })
  private _user: UserInfo | null = null;

  @property({ type: String })
  private _errorMessage = '';

  @property({ type: Boolean })
  private _dialogOpen = false;

  @property({ type: String })
  private _email = '';

  @property({ type: String })
  private _password = '';

  @property({ type: Boolean })
  private _isCreate = false;

  @property({ type: Number })
  private _selectedPage = 0;

  stateChanged(state: any): void {
    this._user = selectUser(state);
    this._verifyingAuth = selectVerifyingAuth(state);
    this._errorMessage = selectSignInErrorMessage(state);
    this._dialogOpen = selectSignInDialogOpen(state);
    this._email = selectSignInDialogEmail(state);
    this._password = selectSignInDialogPassword(state);
    this._isCreate = selectSignInDialogIsCreate(state);
    this._selectedPage = selectSignInDialogSelectedPage(state);
  }

  protected firstUpdated(): void {
    store.dispatch(firebaseSignIn());
  }

  private _handleEmailChanged(e: Event): void {
    const target = e.composedPath()[0] as HTMLInputElement;
    store.dispatch(updateSignInDialogEmail(target.value));
  }

  private _handlePasswordChanged(e: Event): void {
    const target = e.composedPath()[0] as HTMLInputElement;
    store.dispatch(updateSignInDialogPassword(target.value));
  }

  render() {
    return html`
      <div class="${this._verifyingAuth ? 'verifying' : ''} ${OFFLINE_DEV_MODE ? 'offline' : ''}">
        <div id="offline"></div>
        <div class="horizontal">
          <boardgame-player-chip
            .photoUrl="${this._user ? this._user.PhotoURL : ''}"
            .displayName="${this._user ? this._user.DisplayName : ''}">
          </boardgame-player-chip>
          <div class="vertical">
            ${when(
              this._user,
              () => html`
                <div>${this._user!.DisplayName}</div>
                <a @click="${() => store.dispatch(signOut())}">Sign Out</a>
              `,
              () => html`
                <div>Not signed in</div>
                <a @click="${() => store.dispatch(showSignInDialog())}">Sign In</a>
              `
            )}
          </div>
        </div>
      </div>

      <md-dialog ?open="${this._dialogOpen}" @close="${this._handleDialogClose}">
        ${when(OFFLINE_DEV_MODE, () => html`
          <div slot="headline" style="color:red;">
            <strong>Offline Dev Mode enabled; login is faked</strong>
          </div>
        `)}

        <div slot="content">
          <div class="page ${this._selectedPage === 0 ? 'selected' : ''}">
            <h2>Sign In</h2>
            <p>You must sign in to use this app.</p>
            <div class="layout vertical">
              <md-filled-button @click="${() => store.dispatch(signInWithGoogle())}">
                Google
              </md-filled-button>
              <md-outlined-button @click="${() => store.dispatch(showSignInDialogEmailPage(false))}">
                Email/Password
              </md-outlined-button>
              <p style="text-align:center"><em>or</em></p>
              <md-filled-button @click="${() => store.dispatch(showSignInDialogEmailPage(true))}">
                Create an account
              </md-filled-button>
            </div>
          </div>

          <div class="page ${this._selectedPage === 1 ? 'selected' : ''}">
            <h2>${this._isCreate ? 'Create Account' : 'Sign In'}</h2>
            <md-filled-text-field
              label="Email"
              type="email"
              .value="${this._email}"
              @input="${this._handleEmailChanged}">
            </md-filled-text-field>
            <md-filled-text-field
              label="Password"
              type="password"
              .value="${this._password}"
              @input="${this._handlePasswordChanged}">
            </md-filled-text-field>
            <div class="buttons">
              <md-outlined-button @click="${() => store.dispatch(showSignInDialog())}">
                Cancel
              </md-outlined-button>
              <md-filled-button @click="${() => store.dispatch(signInOrCreateWithEmailAndPassword())}">
                ${this._isCreate ? 'Create Account' : 'Sign In'}
              </md-filled-button>
            </div>
          </div>

          <div class="page ${this._selectedPage === 2 ? 'selected' : ''}">
            <h2>Signing in...</h2>
            <md-linear-progress indeterminate></md-linear-progress>
          </div>

          <div class="page ${this._selectedPage === 3 ? 'selected' : ''}">
            <h2>Sign In Error</h2>
            <div>${this._errorMessage}</div>
            <div class="buttons">
              <md-filled-button @click="${() => store.dispatch(showSignInDialog())}">
                OK
              </md-filled-button>
            </div>
          </div>
        </div>
      </md-dialog>
    `;
  }

  private _handleDialogClose(): void {
    // Dialog can be closed by clicking outside or pressing ESC
    // We don't need special handling here as the dialog state is managed by Redux
  }
}

declare global {
  interface HTMLElementTagNameMap {
    'boardgame-user': BoardgameUser;
  }
}
