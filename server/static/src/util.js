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

//getProperty returns the given prop name out of the object. propName may be a
//string or an array of string fields. if propNames has a dot, then it will be
//split on those items. Handles undefined fine.
export const getProperty = (obj, propNames) => {
    if (!obj) return undefined;
    if (typeof obj != "object") return undefined;
    if (!propNames) return undefined;
    if (typeof propNames == "string") propNames = propNames.split(".");
    let propName = propNames.shift();
    let prop = obj[propName];
    //That was the last name in the chain
    if (propNames.length == 0) return prop;
    return getProperty(prop, propNames);
}