export const UPDATE_GAME_ROUTE = 'UPDATE_GAME_ROUTE';

export const updateGameRoute = (pageExtra) => {
    const pieces = pageExtra.split("/");
    //remove the trailing slash
    if (!pieces[pieces.length - 1]) pieces.pop();
    if (pieces.length != 2) {
      console.warn("URL for game didn't have expected number of pieces");
      return null;
    }
    return {
        type: UPDATE_GAME_ROUTE,
        name: pieces[0],
        id: pieces[1],
    }
}