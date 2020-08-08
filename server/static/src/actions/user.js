
export const UPDATE_USER = 'UPDATE_USER';
export const VERIFYING_AUTH = 'VERIFYING_AUTH';
export const UPDATE_SIGN_IN_ERROR_MESSAGE = 'UPDATE_SIGN_IN_ERROR_MESSAGE';
export const SET_USER_ADMIN = 'SET_USER_ADMIN';
export const SHOW_SIGN_IN_DIALOG = 'SHOW_SIGN_IN_DIALOG';
export const UPDATE_SIGN_IN_DIALOG_EMAIL = "UPDATE_SIGN_IN_DIALOG_EMAIL";
export const UPDATE_SIGN_IN_DIALOG_PASSWORD = "UPDATE_SIGN_IN_DIALOG_PASSWORD";
export const UPDATE_SIGN_IN_DIALOG_SELECTED_PAGE = "UPDATE_SIGN_IN_DIALOG_SELECTED_PAGE";
export const SHOW_SIGN_IN_DIALOG_EMAIL_PAGE = "SHOW_SIGN_IN_DIALOG_EMAIL_PAGE";

// This import loads the firebase namespace along with all its type information.
import firebase from '@firebase/app';
import '@firebase/auth';

import {
    selectUser,
    selectAdminAllowed,
    selectSignInDialogEmail,
    selectSignInDialogIsCreate
} from '../selectors.js';

import {
    apiPath,
    postFetchParams
} from '../util.js';

import {
    OFFLINE_DEV_MODE
} from './app.js';

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

//firebaseUser isn't state that is rendered, and it can't go in the redux store
//anyway.
let firebaseUser = null;
let lastValidatedFirebaseUserID = '';
const firebaseApp = firebase.initializeApp(CONFIG.firebase)


export const firebaseSignIn = () => (dispatch) => {
    if (OFFLINE_DEV_MODE) {
        dispatch(firebaseUserUpdated(recoverFauxUser()));
    } else {
        firebaseApp.auth().onAuthStateChanged(user => dispatch(firebaseUserUpdated(user)));
    }
};

let signedInAction = null;
export const setSignedInAction = (action) => {
    signedInAction = action;
}

const updateSignInError = (err) => {
    return {
        type: UPDATE_SIGN_IN_ERROR_MESSAGE,
        error: err.message,
    }
}

const fauxSignIn = (email, displayName) => (dispatch) => {
    if (OFFLINE_DEV_MODE) {
      console.error("OfflineDevMode not enabled")
      return;
    }
    dispatch(firebaseUserUpdated(new fauxFirebaseUser(email, displayName)));
}

export const signInWithGoogle = () => (dispatch) => {
    if (OFFLINE_DEV_MODE) {
        let email = prompt("Fake email address to login with:");
        dispatch(fauxSignIn(email, email));
    } else {
        let provider = new firebase.auth.GoogleAuthProvider();
        provider.addScope("profile");
        provider.addScope("email");
        firebaseApp.auth().signInWithPopup(provider).catch(err => dispatch(updateSignInError(err)));
        dispatch(updateSignInDialogSelectedPage(2));
    }
};

export const signInOrCreateWithEmailAndPassword = () => (dispatch, getState) => {
    const state = getState();
    const email = selectSignInDialogEmail(state);
    const password = selectSignInDialogPassword(state);
    const isCreate = selectSignInDialogIsCreate(state);
    if (OFFLINE_DEV_MODE) {
        dispatch(fauxSignIn(email, email));
    } else {
        if (isCreate) {
            firebaseApp.auth().createUserWithEmailAndPassword(email, password).catch(err => dispatch(updateSignInError(err)));
        } else {
            firebaseApp.auth().signInWithEmailAndPassword(email, password).catch(err => dispatch(updateSignInError(err)));
        }
    };
    dispatch(updateSignInDialogSelectedPage(2));
};

const firebaseUserUpdated = (fUser) => (dispatch, getState) => {
    firebaseUser = fUser;
    const user = selectUser(getState());
    if (!user && !firebaseUser) return;
    dispatch({type: VERIFYING_AUTH});
    if (firebaseUser) {
      if (lastValidatedFirebaseUserID != firebaseUser.uid) {
        //User has changed!
       dispatch(validateCookie());
      }
      lastValidatedFirebaseUserID = firebaseUser.uid;
    } else {
      dispatch(validateCookieWithToken(""));
      lastValidatedFirebaseUserID = "";
    }
};

export const signOut = () => (dispatch) => {
    if (OFFLINE_DEV_MODE) {
        fauxSignOut();
        dispatch(firebaseUserUpdated(null));
    } else {
        //This will call firebaseUserUpdate
        firebaseApp.auth().signOut();
    }
}

const validateCookie = () => (dispatch) => {
    if (!firebaseUser) {
        console.warn("No firebase user");
        return;
    }
    firebaseUser.getIdToken(true).then(token => dispatch(validateCookieWithToken(token)));
};

const validateCookieWithToken = (token) => async (dispatch) => {
    //Reaches out to the auth endpoint to get a cookie set (or validate that our cookie is set).
    let uid = ""
    let email = ""
    let photoUrl = ""
    let displayName = ""

    if (firebaseUser) {

        uid = firebaseUser.uid || "";
        email = firebaseUser.email || "";
        photoUrl = firebaseUser.photoURL || "";
        displayName = firebaseUser.displayName || "";

        if (firebaseUser.providerData) {
        for (var i = 0; i < firebaseUser.providerData.length; i++) {
            var provider = firebaseUser.providerData[i];
            if (!email && provider.email) email = provider.email;
            if (!photoUrl && provider.photoURL) photoUrl = provider.photoURL;
            if (!displayName && provider.displayName) displayName = provider.displayName;
        }
        }
    }

    const body = "uid=" + uid + "&token=" + token + "&email=" + email + "&photo=" + photoUrl + "&displayname=" + displayName;

    let authResponse = await fetch(apiPath('auth'), postFetchParams(body));

    if (authResponse.status != 200) {
        //TODO: show an error here to user
        console.warn(authResponse);
        dispatch(updateUser(null, false));
        return;
    }

    let authJSONResponse = await authResponse.json();

    if (authJSONResponse.Status != "Success") {
        //TODO: show an error here to user
        console.warn(authJSONResponse);
        dispatch(updateUser(null, false));
        return;
      }
  
      dispatch(updateUser(authJSONResponse.User, authJSONResponse.AdminAllowed));

      //Must have been a log out
      if (!authJSONResponse.User) return;
      if (!signedInAction) return;
      signedInAction();
      signedInAction = null;

};

const updateUser = (user, adminAllowed = false) => {
    return {
        type: UPDATE_USER,
        user,
        adminAllowed,
    }
}

export const setUserAdmin = (isAdmin) => (dispatch, getState) => {
    const adminAllowed = selectAdminAllowed(getState());
    if (isAdmin && !adminAllowed) {
        console.warn("Can't set admin to true: admin not allowed");
        return;
    }
    dispatch({
        type: SET_USER_ADMIN,
        admin: isAdmin
    })
}

export const showSignInDialogEmailPage = (isCreate) => {
    return {
        type: SHOW_SIGN_IN_DIALOG_EMAIL_PAGE,
        isCreate
    }
}

export const showSignInDialog = () => {
    return {
        type: SHOW_SIGN_IN_DIALOG
    }
}

export const updateSignInDialogEmail = (email) => {
    return {
        type: UPDATE_SIGN_IN_DIALOG_EMAIL,
        email
    }
}

export const updateSignInDialogPassword = (password) => {
    return {
        type: UPDATE_SIGN_IN_DIALOG_EMAIL,
        password
    }
}


export const updateSignInDialogSelectedPage = (selectedPage) => {
    return {
        type: UPDATE_SIGN_IN_DIALOG_SELECTED_PAGE,
        selectedPage
    }
}