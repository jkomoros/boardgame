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
import './boardgame-ajax.js';
// This import loads the firebase namespace along with all its type information.
import firebase from '@firebase/app';
import '@firebase/auth';
import './boardgame-player-chip.js';
import './shared-styles.js';
import { html } from '@polymer/polymer/lib/utils/html-tag.js';

const fauxFirebaseEmailKey = "faux-firebase-email";
const fauxFirebaseDisplayNameKey = "faux-firebase-display-name"

class fauxFirebaseUser {
  constructor(email, displayName) {
    this.email = email || "tester@gmail.com"
    this.displayName = displayName || "Mr. Tester"
    this.uid = this.email;
    localStorage.setItem(fauxFirebaseEmailKey, this.email);
    localStorage.setItem(fauxFirebaseDisplayNameKey, this.displayName);
  }

  getIdToken(force) {
    return Promise.resolve("fake-token-value-for-offline-dev-mode");
  }
}

function recoverFauxUser() {
  let email = localStorage.getItem(fauxFirebaseEmailKey);
  if (!email) return null;
  let displayName = localStorage.getItem(fauxFirebaseDisplayNameKey) || email;
  return new fauxFirebaseUser(email, displayName);
}

function fauxSignOut() {
  localStorage.removeItem(fauxFirebaseEmailKey);
  localStorage.removeItem(fauxFirebaseDisplayNameKey);
}

class BoardgameUser extends PolymerElement {
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
    <div class\$="{{_classForVerifyingAuth(verifyingAuth, offlineDevMode)}}">
      <div id="offline"></div>
      <div class="horizontal layout">
        <boardgame-player-chip photo-url="{{_string(user.PhotoUrl)}}" display-name="{{_string(user.DisplayName)}}"></boardgame-player-chip>
        <div class="vertical layout">
          <template is="dom-if" if="{{user}}">
              <div>{{user.DisplayName}}</div>
              <a on-tap="signOut">Sign Out</a>
          </template>
          <template is="dom-if" if="{{!user}}">
            <div>Not signed in</div>
            <a on-tap="showSignInDialog">Sign In</a>
          </template>
        </div>
      </div>
    </div>
    <!-- TODO: ideall this would be modal, but given its position in DOM that doesn't work.
    See https://github.com/PolymerElements/paper-dialog/issues/7 -->

    <paper-dialog id="dialog" no-cancel-on-esc-key="" no-cancel-on-outside-click="">
      <div hidden$="{{!offlineDevMode}}">
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
            <paper-button on-tap="emailSubmitted" autofocus="" default="">{{buttonText(emailFormIsSignIn)}}</paper-button>
          </div>
        </div>
        <div>
          <h2>Signing in...</h2>
          <paper-spinner-lite active=""></paper-spinner-lite>
        </div>
        <div>
          <h2>Sign In Error</h2>
          <div>{{errorText}}</div>
          <div class="buttons">
            <paper-button on-tap="cancel" default="">OK</paper-button>
          </div>
        </div>
      </iron-pages>
    </paper-dialog>
    <boardgame-ajax id="auth" path="auth" handle-as="json" last-response="{{authResponse}}" method="POST"></boardgame-ajax>
`;
  }

  static get is() {
    return "boardgame-user"
  }

  static get properties() {
    return {
      emailFormIsSignIn: {
        type: Boolean,
        value: true,
      },
      authResponse: {
        type: Object,
        observer: "_authResponseChanged",
      },
      adminAllowed: {
        type: Boolean,
        notify: true,
      },
      loggedIn : {
        type: Boolean,
        notify:true,
        computed: "_computeLoggedIn(user)",
      },
      //set to true after firebase has a user but before our server has ack'd.
      verifyingAuth : Boolean,
      _firebaseConfig: Object,
      //The firebaseUser object
      firebaseUser: {
        type: Object,
        observer: "_firebaseUserChanged",
      },
      //The confirmed user object from our server
      user: {
        type: Object,
        value: null,
      },
      lastUserId: String,
      //When the user signs in successfully, if this is not undefined, will be called.
      signedInAction: Object,
      _firebaseApp: Object,
    }
  }

  ready() {
    super.ready();
    if (this.offlineDevMode) {
      this.firebaseUser = recoverFauxUser();
    } else {
      this._firebaseApp = firebase.initializeApp(CONFIG.firebase);
      this._firebaseApp.auth().onAuthStateChanged(user => this.firebaseUser = user);
    }
  }

  get offlineDevMode() {
    if (!CONFIG) return false;
    return CONFIG.offline_dev_mode || false;
  }

  buttonText(isSignIn) {
    return isSignIn ? "Sign In" : "Create Account";
  }

  _computeLoggedIn(user) {
    if (!user) return false;
    return true;
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

  _authResponseChanged(newValue) {
    if (!newValue || newValue.Status != "Success") {
      this.user = null;
      return;
    }

    this.verifyingAuth = false;

    this.user = newValue.User;
    this.adminAllowed = newValue.AdminAllowed;

    //Must have been a log out
    if (!this.user) return;
    if (!this.signedInAction) return;

    this.signedInAction();

    this.signedInAction = null;
  }

  _firebaseUserChanged(user) {
    if (!this.user && !user) return;
    this.user = null;
    this.verifyingAuth = true;
    if (user) {
      this.$.dialog.close();
      if (this.lastUserId != user.uid) {
        //User has changed!
       this.validateCookie();
      }
      this.lastUserId = user.uid;
    } else {
      this.validateCookieWithToken("");
      this.lastUserId = "";
    }
  }

  validateCookie() {
    this.firebaseUser.getIdToken(true).then(this.validateCookieWithToken.bind(this));
  }

  validateCookieWithToken(token) {
    //Reaches out to the auth endpoint to get a cookie set (or validate that our cookie is set).
    let uid = ""
    let email = ""
    let photoUrl = ""
    let displayName = ""

    if (this.firebaseUser) {
      var user = this.firebaseUser

      uid = user.uid || "";
      email = user.email || "";
      photoUrl = user.photoURL || "";
      displayName = user.displayName || "";

      if (user.providerData) {
        for (var i = 0; i < user.providerData.length; i++) {
          var provider = user.providerData[i];
          if (!email && provider.email) email = provider.email;
          if (!photoUrl && provider.photoURL) photoUrl = provider.photoURL;
          if (!displayName && provider.displayName) displayName = provider.displayName;
        }
      }
    }

    this.$.auth.body = "uid=" + uid + "&token=" + token + "&email=" + email + "&photo=" + photoUrl + "&displayname=" + displayName;
    this.$.auth.generateRequest();
  }

  createLogin() {
    this.emailFormIsSignIn = false;
    this.showEmailPage();
  }

  cancel() {
    this.$.pages.selected = 0;
  }

  emailSubmitted() {
    let email = this.$.email.value;
    let password = this.$.password.value;

    if (this.emailFormIsSignIn) {
      this.signInWithEmailAndPassword(email, password);
    } else {
      this.createUserWithEmailAndPassword(email, password);
    }
  }

  handleSignInError(err) {
    this.errorText = err.message;
    this.$.pages.selected = 3;
  }

  showEmail() {
    this.emailFormIsSignIn = true;
    this.showEmailPage();
  }

  fauxSignIn(email, displayName) {
    if (!CONFIG || !CONFIG.offline_dev_mode) {
      console.error("OfflineDevMode not enabled")
      return;
    }
    this.firebaseUser = new fauxFirebaseUser(email, displayName);
  }

  signInWithGoogle() {
    if (this.offlineDevMode) {
      let email = prompt("Fake email address to login with:");
      this.fauxSignIn(email, email);
    } else {
      let provider = new firebase.auth.GoogleAuthProvider();
      provider.addScope("profile");
      provider.addScope("email");
      this._firebaseApp.auth().signInWithPopup(provider).catch(this.handleSignInError.bind(this));
    }
    this.$.pages.selected = 2;
  }

  signInWithEmailAndPassword(email, password) {
    if (this.offlineDevMode) {
      this.fauxSignIn(email, email);
    } else {
      this._firebaseApp.auth().signInWithEmailAndPassword(email, password).catch(this.handleSignInError.bind(this));
    };
    this.$.pages.selected = 2;
  }

  createUserWithEmailAndPassword(email, password) {
    if (this.offlineDevMode) {
      this.fauxSignIn(email, email);
    } else {
      this._firebaseApp.auth().createUserWithEmailAndPassword(email, password).catch(this.handleSignInError.bind(this));
    }
    this.$.pages.selected = 2;
  }

  showEmailPage() {
    this.$.email.value = "";
    this.$.password.value = "";
    this.$.pages.selected = 1;
  }

  showSignInDialog(e) {

    //Might be undefined, that's fine
    this.signedInAction = e.detail.nextAction;

    this.$.pages.selected = 0;
    this.$.dialog.open();
  }

  signOut(e) {
    if (this.offlineDevMode) {
      fauxSignOut();
      this.firebaseUser = null;
    } else {
      this._firebaseApp.auth().signOut();
    }
  }
}

customElements.define(BoardgameUser.is, BoardgameUser);
