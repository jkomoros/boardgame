//gamePath returns the absolute path to view the given game.
export const gamePath = (name, id) => {
    return "/game/" + name + "/" + id + "/"
}

//gameAPIPath returns the absolute path to the API endpoint for the given name
export const gameAPIPath = (name, id, params) => {
    //gamePath has '/' but apiPath can strip it out
    return apiPath(gamePath(name, id), params);
}

export const apiPath = (path, params) => {

    if (path && path[0] === '/') path.slice(1);

    //API_HOST is defined in index.html
    const url = API_HOST + '/api/' + path;

    if (!params) return url;

    let parts = [];
    for (let [key, value] of Object.entries(params)) {
        parts.push(key + "=" + value);
    }

    return url + "?" + parts.join("&");
}

//postFetchParams returns the default params to use for a post fetch
export const postFetchParams = (body) => {
    return {
        method: 'POST',
        credentials: 'include',
        headers: {
            'Content-Type': 'application/x-www-form-urlencoded'
        },
        mode: 'cors',
        body,
    }
}

export const deepCopy = (obj) => {
    if (typeof obj != "object") return obj;
    if (!obj) return obj;
    const result = Array.isArray(obj) ? [] : {};
    for (let [key, val] of Object.entries(obj)) {
        result[key] = deepCopy(val);
    }
    return result;
}

//getProperty returns the given prop name out of the object. propName may be a
//string or an array of string fields. if propNames has a dot, then it will be
//split on those items. Handles undefined fine.
export const getProperty = (obj, propNames) => {
    if (!obj) return undefined;
    if (typeof obj != "object") return undefined;
    if (!propNames) return undefined;
    if (typeof propNames == "string") propNames = propNames.split(".");
    let propName = propNames[0];
    let prop = obj[propName];
    //That was the last name in the chain
    if (propNames.length == 1) return prop;
    return getProperty(prop, propNames.slice(1));
}

//setProperty sets the given propName to val. Will return true if it was
//successfully set, false if there was an error in the shape.
export const setProperty = (obj, propNames, val) => {
    if (!obj) return false;
    if (typeof obj != "object") return false;
    if (!propNames) return false;
    if (typeof propNames == "string") propNames = propNames.split(".");
    let propName = propNames[0];
    //That was the last name in the chain
    if (propNames.length == 1) {
        obj[propName] = val;
        return true;
    };
    return setProperty(obj[propName], propNames.slice(1), val);
}