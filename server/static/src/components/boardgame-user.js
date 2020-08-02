/**
@license
Copyright (c) 2016 The Polymer Project Authors. All rights reserved.
This code may only be used under the BSD style license found at http://polymer.github.io/LICENSE.txt
The complete set of authors may be found at http://polymer.github.io/AUTHORS.txt
The complete set of contributors may be found at http://polymer.github.io/CONTRIBUTORS.txt
Code distributed by Google as part of the polymer project is also
subject to an additional IP rights grant found at http://polymer.github.io/PATENTS.txt
*/
import { PolymerElement } from '@polymer/polymer/polymer-element.js';

import '@polymer/polymer/lib/elements/dom-if.js';
import '@polymer/paper-dialog/paper-dialog.js';
import '@polymer/paper-button/paper-button.js';
import '@polymer/paper-input/paper-input.js';
import '@polymer/iron-pages/iron-pages.js';
import '@polymer/iron-flex-layout/iron-flex-layout-classes.js';
import '@polymer/paper-spinner/paper-spinner-lite.js';
import './boardgame-player-chip.js';
import './shared-styles.js';
import { html } from '@polymer/polymer/lib/utils/html-tag.js';

import { connect } from 'pwa-helpers/connect-mixin.js';
import { store } from '../store.js';

import {
  OFFLINE_DEV_MODE
} from '../actions/app.js';

import {
  firebaseSignIn,
  signOut,
  setSignedInAction,
  signInWithGoogle,
  signInWithEmailAndPassword,
  createUserWithEmailAndPassword,
  showSignInDialog
} from '../actions/user.js';

import {
  selectUser,
  selectVerifyingAuth,
  selectSignInErrorMessage,
  selectSignInDialogOpen
} from '../selectors.js';

class BoardgameUser extends connect(store)(PolymerElement) {
  static get template() {
    return html`
    <style is="custom-style" include="iron-flex shared-styles">
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
    <div class\$="[[_classForVerifyingAuth(_verifyingAuth, offlineDevMode)]]">
      <div id="offline"></div>
      <div class="horizontal layout">
        <boardgame-player-chip photo-url="[[_string(_user.PhotoURL)]]" display-name="[[_string(_user.DisplayName)]]"></boardgame-player-chip>
        <div class="vertical layout">
          <template is="dom-if" if="[[_user]]">
              <div>[[_user.DisplayName]]</div>
              <a on-tap="signOut">Sign Out</a>
          </template>
          <template is="dom-if" if="[[!_user]]">
            <div>Not signed in</div>
            <a on-tap="showSignInDialog">Sign In</a>
          </template>
        </div>
      </div>
    </div>
    <!-- TODO: ideall this would be modal, but given its position in DOM that doesn't work.
    See https://github.com/PolymerElements/paper-dialog/issues/7 -->

    <paper-dialog id="dialog" no-cancel-on-esc-key="" no-cancel-on-outside-click="" opened="[[_dialogOpen]]">
      <div hidden$="[[!offlineDevMode]]">
        <strong style="color:red;">Offline Dev Mode enabled; login is faked</strong>
      </div>
      <iron-pages id="pages">
        <div>
          <h2>Sign In</h2>
          <p>You must sign in to use this app.</p>
          <div class="layout vertical">
            <paper-button on-tap="signInWithGoogle">Google</paper-button>
            <paper-button on-tap="showEmail">Email/Password</paper-button>
            <p style="text-align:center"><em>or</em></p>
            <paper-button on-tap="createLogin">Create an account</paper-button>
          </div>
        </div>
        <div>
          <paper-input id="email" label="Email"></paper-input>
          <paper-input id="password" label="Password" type="password"></paper-input>
          <div class="buttons">
            <paper-button on-tap="cancel">Cancel</paper-button>
            <paper-button on-tap="emailSubmitted" autofocus="" default="">[[buttonText(_emailFormIsSignIn)]]</paper-button>
          </div>
        </div>
        <div>
          <h2>Signing in...</h2>
          <paper-spinner-lite active=""></paper-spinner-lite>
        </div>
        <div>
          <h2>Sign In Error</h2>
          <div>[[_errorMessage]]</div>
          <div class="buttons">
            <paper-button on-tap="cancel" default="">OK</paper-button>
          </div>
        </div>
      </iron-pages>
    </paper-dialog>
`;
  }

  static get is() {
    return "boardgame-user"
  }

  static get properties() {
    return {
      _emailFormIsSignIn: {
        type: Boolean,
        value: true,
      },
      //set to true after firebase has a user but before our server has ack'd.
      _verifyingAuth : Boolean,
      //The confirmed user object from our server
      _user: {
        type: Object,
        value: null,
      },
      _errorMessage: {
        type: String,
        observer: "_errorMessageChanged"
      },
      _dialogOpen: { type: Boolean },
    }
  }

  stateChanged(state) {
    this._user = selectUser(state);
    this._verifyingAuth = selectVerifyingAuth(state);
    this._errorMessage = selectSignInErrorMessage(state);
    this._dialogOpen = selectSignInDialogOpen(state);
  }

  get offlineDevMode() {
    //TODO; get rid of this getter when we switch to litelement. This is only
    //necessary for the template stamping
    return OFFLINE_DEV_MODE
  }

  ready() {
    super.ready();
    store.dispatch(firebaseSignIn());
  }

  buttonText(isSignIn) {
    return isSignIn ? "Sign In" : "Create Account";
  }

  _classForVerifyingAuth(verifyingAuth, offlineDevMode) {
    let pieces = [];
    if (offlineDevMode) pieces.push("offline");
    if (verifyingAuth) pieces.push("verifying");
    return pieces.join(" ");
  }

  _string(str) {
    //Necessary for a thing that might be undefined to make it the empty
    //string instead, because Polymer's databinding treats setting as
    //undefined as a signal that nothing has changed.
    return (str) ? str : ""
  }

  createLogin() {
    this._emailFormIsSignIn = false;
    this.showEmailPage();
  }

  cancel() {
    this.$.pages.selected = 0;
  }

  emailSubmitted() {
    let email = this.$.email.value;
    let password = this.$.password.value;

    if (this._emailFormIsSignIn) {
      this.signInWithEmailAndPassword(email, password);
    } else {
      this.createUserWithEmailAndPassword(email, password);
    }
  }

  _errorMessageChanged(newValue) {
    if (newValue) this.$.pages.selected = 3;
  }

  showEmail() {
    this._emailFormIsSignIn = true;
    this.showEmailPage();
  }

  signInWithGoogle() {
    store.dispatch(signInWithGoogle());
    this.$.pages.selected = 2;
  }

  signInWithEmailAndPassword(email, password) {
    store.dispatch(signInWithEmailAndPassword(email, password));
    this.$.pages.selected = 2;
  }

  createUserWithEmailAndPassword(email, password) {
    store.dispatch(createUserWithEmailAndPassword(email, password));
    this.$.pages.selected = 2;
  }

  showEmailPage() {
    this.$.email.value = "";
    this.$.password.value = "";
    this.$.pages.selected = 1;
  }

  showSignInDialog(e) {
    //Might be undefined, that's fine
    setSignedInAction(e.detail.nextAction);
    this.$.pages.selected = 0;
    store.dispatch(showSignInDialog());
  }

  signOut() {
    store.dispatch(signOut());
  }
}

customElements.define(BoardgameUser.is, BoardgameUser);
