export const gamePath = (name, id, params) => {
    return apiPath("game/" + name + "/" + id + "/", params);
}

export const apiPath = (path, params) => {
    //API_HOST is defined in index.html
    const url = API_HOST + '/api/' + path;

    if (!params) return url;

    let parts = [];
    for (let [key, value] of Object.entries(params)) {
        parts.push(key + "=" + value);
    }

    return url + "?" + parts.join("&");
}