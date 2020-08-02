import {
    UPDATE_USER,
    VERIFYING_AUTH,
    UPDATE_SIGN_IN_ERROR_MESSAGE,
    SET_USER_ADMIN,
    SHOW_SIGN_IN_DIALOG,
    UPDATE_SIGN_IN_DIALOG_EMAIL,
    UPDATE_SIGN_IN_DIALOG_PASSWORD
} from '../actions/user.js';

const INITIAL_STATE = {
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
};

const user = (state = INITIAL_STATE, action) => {
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
        return {
            ...state,
            errorMessage: action.error,
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
            dialogPassword: ""
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
	default:
		return state;
	}
};

export default user;
