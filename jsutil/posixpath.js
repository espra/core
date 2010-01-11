/*
 * Changes released into the Public Domain by tav <tav@espians.com>
 *
 * Adapted from posixpath.py in the Python Standard Library.
 *
 */

function join(p1, p2) {
    var path = p1;
    if (p2.charAt(0) === "/") {
        path = p2;
    } else if (path === "" || path.charAt(path.length - 1) === "/") {
        path += p2;
    } else {
        path += "/" + p2;
    }
    return path;
}

function split(path) {
    var i = path.lastIndexOf('/') + 1,
        head = path.slice(0, i),
        tail = path.slice(i);
    if (head && head !== ('/' * head.length)) {
        head = head.replace(/\/*$/g, "");
    }
    return [head, tail];
}

function dirname(path) {
    return split(path)[0];
}

// -----------------------------------------------------------------------------
// exports
// -----------------------------------------------------------------------------

exports.join = join;
exports.split = split;
exports.dirname = dirname;
