import '@polymer/paper-dialog/paper-dialog.js';
import '@polymer/paper-button/paper-button.js';
import '@polymer/paper-input/paper-input.js';
import '@polymer/iron-pages/iron-pages.js';
import '@polymer/paper-spinner/paper-spinner-lite.js';
import './boardgame-player-chip.js';

import {
  SharedStyles
} from './shared-styles-lit.js';

import { LitElement, html } from '@polymer/lit-element';

import { connect } from 'pwa-helpers/connect-mixin.js';
import { store } from '../store.js';

import {
  OFFLINE_DEV_MODE
} from '../actions/app.js';

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

class BoardgameUser extends connect(store)(LitElement) {
  render() {
    return html`
    ${ SharedStyles }
    <style>
      :host {
        display:block;
        position: relative;
        padding:16px;
      }

      a {
        cursor:pointer;
      }

      paper-dialog {
        min-width:300px;
        min-height:300px;
      }

      .verifying {
        font-style: italic;
      }

      .horizontal {
        display:flex;
        flex-direction: row;
      }

      .vertical {
        display: flex;
        flex-direction: column;
      }

      #offline {
        display:none;
        height: 5px;
        width: 5px;
        top: 16px;
        left: 16px;
        position: absolute;
        border-radius:2.5px;
        background-color:red;
      }

      .offline #offline {
        display:block;
      }
    </style>
    <div class="${this._verifyingAuth ? 'verifying' : ''} ${OFFLINE_DEV_MODE ? 'offline' : ''}">
      <div id="offline"></div>
      <div class="horizontal">
        <boardgame-player-chip .photoUrl=${this._user ? this._user.PhotoURL : ''} .displayName=${this._user ? this._user.DisplayName : ''}></boardgame-player-chip>
        <div class="vertical">
          ${
            this._user ?
            html`<div>${this._user.DisplayName}</div>
              <a @tap=${this.signOut}>Sign Out</a>` : 
            html`<div>Not signed in</div>
              <a @tap=${this.showSignInDialog}>Sign In</a>`
          }
        </div>
      </div>
    </div>
    <!-- TODO: ideall this would be modal, but given its position in DOM that doesn't work.
    See https://github.com/PolymerElements/paper-dialog/issues/7 -->

    <paper-dialog id="dialog" no-cancel-on-esc-key="" no-cancel-on-outside-click="" .opened=${this._dialogOpen}>
      <div ?hidden=${!OFFLINE_DEV_MODE}>
        <strong style="color:red;">Offline Dev Mode enabled; login is faked</strong>
      </div>
      <iron-pages id="pages" .selected=${this._selectedPage}>
        <div>
          <h2>Sign In</h2>
          <p>You must sign in to use this app.</p>
          <div class="layout vertical">
            <paper-button @tap=${this.signInWithGoogle}>Google</paper-button>
            <paper-button @tap=${this.showEmail}>Email/Password</paper-button>
            <p style="text-align:center"><em>or</em></p>
            <paper-button @tap=${this.createLogin}>Create an account</paper-button>
          </div>
        </div>
        <div>
          <paper-input id="email" label="Email" .value=${this._email} @change=${this._handleEmailChanged}></paper-input>
          <paper-input id="password" label="Password" type="password" .value=${this._password} @change=${this._handlePasswordChanged}></paper-input>
          <div class="buttons">
            <paper-button @tap=${this.cancel}>Cancel</paper-button>
            <paper-button @tap=${this.emailSubmitted} autofocus default>${this._isCreate ? 'Create Account' : 'Sign In'}</paper-button>
          </div>
        </div>
        <div>
          <h2>Signing in...</h2>
          <paper-spinner-lite active=""></paper-spinner-lite>
        </div>
        <div>
          <h2>Sign In Error</h2>
          <div>${this._errorMessage}</div>
          <div class="buttons">
            <paper-button @tap=${this.cancel} default="">OK</paper-button>
          </div>
        </div>
      </iron-pages>
    </paper-dialog>
`;
  }

  static get properties() {
    return {
      //set to true after firebase has a user but before our server has ack'd.
      _verifyingAuth : Boolean,
      //The confirmed user object from our server
      _user: {
        type: Object,
        value: null,
      },
      _errorMessage: { type: String },
      _dialogOpen: { type: Boolean },
      _email: { type: String },
      _password: { type: String },
      _isCreate: { type: Boolean },
      _selectedPage: { type: Number },
    }
  }

  stateChanged(state) {
    this._user = selectUser(state);
    this._verifyingAuth = selectVerifyingAuth(state);
    this._errorMessage = selectSignInErrorMessage(state);
    this._dialogOpen = selectSignInDialogOpen(state);
    this._email = selectSignInDialogEmail(state);
    this._password = selectSignInDialogPassword(state);
    this._isCreate = selectSignInDialogIsCreate(state);
    this._selectedPage = selectSignInDialogSelectedPage(state);
  }

  firstUpdated() {
    store.dispatch(firebaseSignIn());
  }

  _handleEmailChanged(e) {
    store.dispatch(updateSignInDialogEmail(e.composedPath()[0].value))
  }

  _handlePasswordChanged(e) {
    store.dispatch(updateSignInDialogPassword(e.composedPath()[0].value));
  }

  createLogin() {
    store.dispatch(showSignInDialogEmailPage(true));
  }

  showEmail() {
    store.dispatch(showSignInDialogEmailPage(false));
  }

  cancel() {
    store.dispatch(showSignInDialog());
  }

  emailSubmitted() {
    store.dispatch(signInOrCreateWithEmailAndPassword());
  }

  signInWithGoogle() {
    store.dispatch(signInWithGoogle());
  }

  showSignInDialog() {
    store.dispatch(showSignInDialog());
  }

  signOut() {
    store.dispatch(signOut());
  }
}

customElements.define("boardgame-user", BoardgameUser);
