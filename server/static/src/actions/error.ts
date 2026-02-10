import type { ThunkAction } from 'redux-thunk';
import type { RootState } from '../types/store';

export const SHOW_ERROR = 'SHOW_ERROR';
export const UPDATE_ERROR = 'UPDATE_ERROR';
export const HIDE_ERROR = 'HIDE_ERROR';

// Action type definitions
interface ShowErrorAction {
    type: typeof SHOW_ERROR;
}

interface HideErrorAction {
    type: typeof HIDE_ERROR;
}

interface UpdateErrorAction {
    type: typeof UPDATE_ERROR;
    title: string;
    message: string;
    friendlyMessage: string;
}

export type ErrorAction =
    | ShowErrorAction
    | HideErrorAction
    | UpdateErrorAction;

type ErrorThunk<ReturnType = void> = ThunkAction<ReturnType, RootState, unknown, ErrorAction>;

export const showError = (): ShowErrorAction => {
    return {
        type: SHOW_ERROR,
    };
};

export const hideError = (): HideErrorAction => {
    return {
        type: HIDE_ERROR,
    };
};

export const updateAndShowError = (title: string, message: string, friendlyMessage: string): ErrorThunk => (dispatch) => {
    dispatch(updateError(title, message, friendlyMessage));
    dispatch(showError());
}

export const updateError = (title: string, message: string, friendlyMessage: string): UpdateErrorAction => {
    if (!title) title = 'Error';
    if (!friendlyMessage) friendlyMessage = "There was an error";
    if (message == friendlyMessage) message = "";
    return {
        type: UPDATE_ERROR,
        title,
        message,
        friendlyMessage,
    }
}