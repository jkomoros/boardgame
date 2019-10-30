/*

Package build is comprised of two sub-packages, api and static, that are two
halves of building a functioning webapp given a config.json.

api builds the server binary that handles all of the game logic. It is a
headless binary that doesn't render any UI but serves as a REST endpoint for the
webapp.

static accumulates all of the various client-side resources (i.e. html, css, js)
necessary to render the server and the specific games it uses. The static server
output is all static; that is that the files don't change and it's designed to
be able to upload to static file hosting surfaces like firebase hosting.
Actually organizing all of the files on disk in the right configuration is a
pain, so the static.Build() is a huge help.

Both api and static Build() methods take a directory parameter, an optional
directory name to store ther results of their builds in. They will create sub
directories within that folder to store their results in a known location.
directory is optional; "" just defaults to the current directory.

Typically you don't have to use these libraries directly, but instead use
`boardgame-util build api`, `boardgame-util build static`, `boardgame-util
clean`, and `boardgame-util serve`.\

This build package contains only this documentation and no code. All code is in
the api and static sub-packages. See their package doc for more about each one.

*/
package build
