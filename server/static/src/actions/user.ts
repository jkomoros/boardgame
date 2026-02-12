import type { ThunkAction } from 'redux-thunk';
import type { RootState, UserInfo } from '../types/store';

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
import firebase from 'firebase/compat/app';
import 'firebase/compat/auth';

import {
    selectUser,
    selectAdminAllowed,
    selectSignInDialogEmail,
    selectSignInDialogPassword,
    selectSignInDialogIsCreate
} from '../selectors.ts';

import {
    apiPath,
    postFetchParams
} from '../util.ts';

import {
    OFFLINE_DEV_MODE
} from './app.ts';

// Action type definitions
interface UpdateUserAction {
    type: typeof UPDATE_USER;
    user: UserInfo | null;
    adminAllowed: boolean;
}

interface VerifyingAuthAction {
    type: typeof VERIFYING_AUTH;
}

interface UpdateSignInErrorMessageAction {
    type: typeof UPDATE_SIGN_IN_ERROR_MESSAGE;
    error: string;
}

interface SetUserAdminAction {
    type: typeof SET_USER_ADMIN;
    admin: boolean;
}

interface ShowSignInDialogAction {
    type: typeof SHOW_SIGN_IN_DIALOG;
}

interface UpdateSignInDialogEmailAction {
    type: typeof UPDATE_SIGN_IN_DIALOG_EMAIL;
    email: string;
}

interface UpdateSignInDialogPasswordAction {
    type: typeof UPDATE_SIGN_IN_DIALOG_PASSWORD;
    password: string;
}

interface UpdateSignInDialogSelectedPageAction {
    type: typeof UPDATE_SIGN_IN_DIALOG_SELECTED_PAGE;
    selectedPage: number;
}

interface ShowSignInDialogEmailPageAction {
    type: typeof SHOW_SIGN_IN_DIALOG_EMAIL_PAGE;
    isCreate: boolean;
}

export type UserAction =
    | UpdateUserAction
    | VerifyingAuthAction
    | UpdateSignInErrorMessageAction
    | SetUserAdminAction
    | ShowSignInDialogAction
    | UpdateSignInDialogEmailAction
    | UpdateSignInDialogPasswordAction
    | UpdateSignInDialogSelectedPageAction
    | ShowSignInDialogEmailPageAction;

type UserThunk<ReturnType = void> = ThunkAction<ReturnType, RootState, unknown, UserAction>;

const fauxFirebaseEmailKey = "faux-firebase-email";
const fauxFirebaseDisplayNameKey = "faux-firebase-display-name"

class fauxFirebaseUser {
  email: string;
  displayName: string;
  uid: string;

  constructor(email: string | null, displayName: string | null) {
    this.email = email || "tester@gmail.com"
    this.displayName = displayName || "Mr. Tester"
    this.uid = this.email;
    localStorage.setItem(fauxFirebaseEmailKey, this.email);
    localStorage.setItem(fauxFirebaseDisplayNameKey, this.displayName);
  }

  getIdToken(force?: boolean): Promise<string> {
    return Promise.resolve("fake-token-value-for-offline-dev-mode");
  }
}

function recoverFauxUser(): fauxFirebaseUser | null {
  let email = localStorage.getItem(fauxFirebaseEmailKey);
  if (!email) return null;
  let displayName = localStorage.getItem(fauxFirebaseDisplayNameKey) || email;
  return new fauxFirebaseUser(email, displayName);
}

function fauxSignOut(): void {
  localStorage.removeItem(fauxFirebaseEmailKey);
  localStorage.removeItem(fauxFirebaseDisplayNameKey);
}

//firebaseUser isn't state that is rendered, and it can't go in the redux store
//anyway.
let firebaseUser: firebase.User | fauxFirebaseUser | null = null;
let lastValidatedFirebaseUserID = '';
const firebaseApp = firebase.initializeApp(CONFIG.firebase)


export const firebaseSignIn = (): UserThunk => (dispatch) => {
    if (OFFLINE_DEV_MODE) {
        dispatch(firebaseUserUpdated(recoverFauxUser()));
    } else {
        firebaseApp.auth().onAuthStateChanged(user => dispatch(firebaseUserUpdated(user)));
    }
};

let signedInAction: (() => void) | null = null;
export const setSignedInAction = (action: () => void): void => {
    signedInAction = action;
}

const updateSignInError = (err: { message: string }): UpdateSignInErrorMessageAction => {
    return {
        type: UPDATE_SIGN_IN_ERROR_MESSAGE,
        error: err.message,
    }
}

const fauxSignIn = (email: string | null, displayName: string | null): UserThunk => (dispatch) => {
    if (!OFFLINE_DEV_MODE) {
      console.error("OfflineDevMode not enabled")
      return;
    }
    dispatch(firebaseUserUpdated(new fauxFirebaseUser(email, displayName)));
}

export const signInWithGoogle = (): UserThunk => (dispatch) => {
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

export const signInOrCreateWithEmailAndPassword = (): UserThunk => (dispatch, getState) => {
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

const firebaseUserUpdated = (fUser: firebase.User | fauxFirebaseUser | null): UserThunk => (dispatch, getState) => {
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

export const signOut = (): UserThunk => (dispatch) => {
    if (OFFLINE_DEV_MODE) {
        fauxSignOut();
        dispatch(firebaseUserUpdated(null));
    } else {
        //This will call firebaseUserUpdate
        firebaseApp.auth().signOut();
    }
}

const validateCookie = (): UserThunk => (dispatch) => {
    if (!firebaseUser) {
        console.warn("No firebase user");
        return;
    }
    firebaseUser.getIdToken(true).then(token => dispatch(validateCookieWithToken(token)));
};

const validateCookieWithToken = (token: string): UserThunk<Promise<void>> => async (dispatch) => {
    //Reaches out to the auth endpoint to get a cookie set (or validate that our cookie is set).
    let uid = ""
    let email = ""
    let photoUrl = ""
    let displayName = ""

    if (firebaseUser) {

        uid = firebaseUser.uid || "";
        email = firebaseUser.email || "";
        photoUrl = (firebaseUser as firebase.User).photoURL || "";
        displayName = firebaseUser.displayName || "";

        if ((firebaseUser as firebase.User).providerData) {
        for (let i = 0; i < (firebaseUser as firebase.User).providerData.length; i++) {
            const provider = (firebaseUser as firebase.User).providerData[i];
            if (provider) {
                if (!email && provider.email) email = provider.email;
                if (!photoUrl && provider.photoURL) photoUrl = provider.photoURL;
                if (!displayName && provider.displayName) displayName = provider.displayName;
            }
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

    let authJSONResponse = await authResponse.json() as {
        Status: string;
        User?: UserInfo;
        AdminAllowed?: boolean;
    };

    if (authJSONResponse.Status != "Success") {
        //TODO: show an error here to user
        console.warn(authJSONResponse);
        dispatch(updateUser(null, false));
        return;
      }

      dispatch(updateUser(authJSONResponse.User || null, authJSONResponse.AdminAllowed || false));

      //Must have been a log out
      if (!authJSONResponse.User) return;
      if (!signedInAction) return;
      signedInAction();
      signedInAction = null;

};

const updateUser = (user: UserInfo | null, adminAllowed = false): UpdateUserAction => {
    return {
        type: UPDATE_USER,
        user,
        adminAllowed,
    }
}

export const setUserAdmin = (isAdmin: boolean): UserThunk => (dispatch, getState) => {
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

export const showSignInDialogEmailPage = (isCreate: boolean): ShowSignInDialogEmailPageAction => {
    return {
        type: SHOW_SIGN_IN_DIALOG_EMAIL_PAGE,
        isCreate
    }
}

export const showSignInDialog = (): ShowSignInDialogAction => {
    return {
        type: SHOW_SIGN_IN_DIALOG
    }
}

export const updateSignInDialogEmail = (email: string): UpdateSignInDialogEmailAction => {
    return {
        type: UPDATE_SIGN_IN_DIALOG_EMAIL,
        email
    }
}

export const updateSignInDialogPassword = (password: string): UpdateSignInDialogPasswordAction => {
    return {
        type: UPDATE_SIGN_IN_DIALOG_PASSWORD,
        password
    }
}


export const updateSignInDialogSelectedPage = (selectedPage: number): UpdateSignInDialogSelectedPageAction => {
    return {
        type: UPDATE_SIGN_IN_DIALOG_SELECTED_PAGE,
        selectedPage
    }
}