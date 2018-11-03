/*

Package boardgame is a framework that makes it possible to build boardgames
with minimial fuss.

This package contains the core boardgame logic engine. Other packages extend
and build of of this base. package boardgame/base provides base
implementations of the various types of objects your game logic must provide.
boardgame/moves provides a collection of powerful move objects to base your
own logic off of. boardgame/server is a package that, given a game definition,
creates a powerful Progressive Web App, complete with automatically-generated
animations.

The boardgame/boardgame-util command is a powerful swiss army knife of
functionality to help create game packages and run serves based on that
automatically.

The documentation in this package is primarily detail about how the various
concepts wire together. For a high-level overview of how everything works and
tour of the main concepts, see TUTORIAL.md.

The primary entry point for use of this package is defining your own
GameDelegate. The methods and documentation from there will point to other
parts of this package.

*/
package boardgame
