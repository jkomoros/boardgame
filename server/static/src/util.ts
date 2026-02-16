//gamePath returns the absolute path to view the given game.
export const gamePath = (name: string, id: string): string => {
    return "/game/" + name + "/" + id + "/"
}

//gameAPIPath returns the absolute path to the API endpoint for the given name
export const gameAPIPath = (name: string, id: string, params?: Record<string, string | number>): string => {
    //gamePath has '/' but apiPath can strip it out
    return apiPath(gamePath(name, id), params);
}

export const apiPath = (path: string, params?: Record<string, string | number>): string => {

    if (path && path[0] === '/') path.slice(1);

    //API_HOST is defined in index.html
    const url = API_HOST + '/api/' + path;

    if (!params) return url;

    let parts: string[] = [];
    for (let [key, value] of Object.entries(params)) {
        parts.push(key + "=" + value);
    }

    return url + "?" + parts.join("&");
}

//postFetchParams returns the default params to use for a post fetch
export const postFetchParams = (body: string): RequestInit => {
    return {
        method: 'POST' as const,
        credentials: 'include' as const,
        headers: {
            'Content-Type': 'application/x-www-form-urlencoded'
        },
        mode: 'cors' as const,
        body,
    }
}

export const deepCopy = <T>(obj: T): T => {
    if (typeof obj != "object") return obj;
    if (!obj) return obj;
    const result: any = Array.isArray(obj) ? [] : {};
    for (let [key, val] of Object.entries(obj)) {
        result[key] = deepCopy(val);
    }
    return result;
}

//getProperty returns the given prop name out of the object. propName may be a
//string or an array of string/number fields. if propNames has a dot, then it will be
//split on those items. Numbers are for array indices. Handles undefined fine.
export const getProperty = (obj: any, propNames: string | (string | number)[]): any => {
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
export const setProperty = (obj: any, propNames: string | (string | number)[], val: any): boolean => {
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

//setPropertyInClone is like setProperty, but instead of modifying the given
//obj, it returns a clone that has the given property affected. The clone will
//duplicate as few things as possible, so for example sub-objects that are not
//affected by the property set will be returned as is. What this means in
//practice is that every object in the path down to propName will be a copy, but
//hopefully nothing else will be.
export const setPropertyInClone = (obj: any, propNames: string | (string | number)[], valToSet: any): any => {
    //Were at the end of the propNames chain, return val to "set" it in the
    //clone.
    if (!propNames || propNames.length == 0) return valToSet;
    if (!obj) return obj;
    if (typeof obj != "object") return obj;
    if (typeof propNames == "string") propNames = propNames.split(".");
    let propName = propNames[0];
    const result: any = Array.isArray(obj) ? [] : {};
    for (let [key, val] of Object.entries(obj)) {
        if (key != propName) {
            //The modification doesn't lie down this branch so we can just reuse
            //existing keys.
            result[key] = val;
            continue;
        }
        result[key] = setPropertyInClone(obj[propName], propNames.slice(1), valToSet)
    }
    return result;
}