import type { Reducer } from 'redux';
import type { UserState } from '../types/store';
import type { UserAction } from '../actions/user.js';
import {
    UPDATE_USER,
    VERIFYING_AUTH,
    UPDATE_SIGN_IN_ERROR_MESSAGE,
    SET_USER_ADMIN,
    SHOW_SIGN_IN_DIALOG,
    UPDATE_SIGN_IN_DIALOG_EMAIL,
    UPDATE_SIGN_IN_DIALOG_PASSWORD,
    UPDATE_SIGN_IN_DIALOG_SELECTED_PAGE,
    SHOW_SIGN_IN_DIALOG_EMAIL_PAGE
} from '../actions/user.js';

const INITIAL_STATE: UserState = {
    admin: false,
    adminAllowed: false,
    loggedIn: false,
    verifyingAuth: false,
    //the user object from OUR server
    user: null,
    errorMessage: "",
    dialogOpen: false,
    dialogEmail: "",
    dialogPassword: "",
    //The dialog can be either in sign-in mode, or create account mode
    dialogIsCreate: false,
    dialogSelectedPage: 0,
};

const user: Reducer<UserState, UserAction> = (state = INITIAL_STATE, action): UserState => {
	switch (action.type) {
	case UPDATE_USER:
        const loggedIn = action.user ? true : false;
		return {
			...state,
            user: action.user,
            adminAllowed: action.adminAllowed,
            verifyingAuth: false,
            dialogOpen: false,
            loggedIn,
        };
    case VERIFYING_AUTH:
        return {
            ...state,
            user: null,
            loggedIn: false,
            verifyingAuth: true,
            adminAllowed: false,
            //verifyingAuth means that the firebase auth token is valid, so no
            //error there.
            errorMessage: "",
        }
    case UPDATE_SIGN_IN_ERROR_MESSAGE:
        let selectedPage = action.error ? 3 : state.dialogSelectedPage
        return {
            ...state,
            errorMessage: action.error,
            dialogSelectedPage: selectedPage
        }
    case SET_USER_ADMIN:
        return {
            ...state,
            admin: action.admin
        }
    case SHOW_SIGN_IN_DIALOG:
        return {
            ...state,
            dialogOpen: true,
            dialogEmail: "",
            dialogPassword: "",
            dialogSelectedPage: 0
        }
    case UPDATE_SIGN_IN_DIALOG_EMAIL:
        return {
            ...state,
            dialogEmail: action.email
        }
    case UPDATE_SIGN_IN_DIALOG_PASSWORD:
        return {
            ...state,
            dialogPassword: action.password
        }
    case SHOW_SIGN_IN_DIALOG_EMAIL_PAGE:
        return {
            ...state,
            dialogIsCreate: action.isCreate,
            dialogSelectedPage: 1,
            dialogEmail: "",
            dialogPassword: ""
        }
    case UPDATE_SIGN_IN_DIALOG_SELECTED_PAGE:
        return{
            ...state,
            dialogSelectedPage: action.selectedPage
        }
	default:
		return state;
	}
};

export default user;
