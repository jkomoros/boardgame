export const SHOW_ERROR = 'SHOW_ERROR';
export const UPDATE_ERROR = 'UPDATE_ERROR';
export const HIDE_ERROR = 'HIDE_ERROR';

export const showError = () => {
    return {
        type: SHOW_ERROR,
    };
};

export const hideError = () => {
    return {
        type: HIDE_ERROR,
    };
};

export const updateAndShowError = (title, message, friendlyMessage) => (dispatch) => {
    dispatch(updateError(title, message, friendlyMessage));
    dispatch(showError());
}

export const updateError = (title, message, friendlyMessage) => {
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