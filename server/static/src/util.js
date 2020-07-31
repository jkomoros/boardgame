export const gamePath = (name, id) => {
    return apiPath("game/" + name + "/" + id + "/");
}

export const apiPath = (path) => {
    //API_HOST is defined in index.html
    return API_HOST + '/api/' + path;
}