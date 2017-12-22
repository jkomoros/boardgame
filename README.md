boardgame is a work-in-progress package that aspires to make it easy to define multi-player boardgames that can be easily hosted in a high-quality web app with minimal configuration.

You formally define the rules and logic of your boardgame in **Go** by defining a state object and a series of moves that can be made. You also define a simple **Polymer-based web component** that renders a given game state client-side.

boardgame is under active development as a hobby project and different components of it vary in their completeness and polish.

## Getting started

A comprehensive getting started guide, including a walkthrough of all of the important concepts in a real-world example, is in the [tutorial](https://github.com/jkomoros/boardgame/blob/master/TUTORIAL.md).

## Design Goals

- **Don't Repeat Yourself** Write your normative logic that defines your game in Go a single time.
- **Minimize Code** Common operations, like "deal cards from the draw stack to each player's hand until each player has 3 cards" should be easy to accomplish with minimal error-prone code. Writing your game should feel like just transcribing the rules into a formal model, not like a challenging coding exercise.
- **Minimize Boilerplate** Structs with powerful defaults to anonymously embed in your structs, and easy-to-use code generation tools where required
- **Clean Layering** If you don't like the default behavior, override just a small portion to make it do what you want
- **Flexible** Powerful enough to model any real-world board or card game
- **Make Cheaters' Lives Hard** Don't rely on security by obscurity to protect secret game state: sanitize properties before transmitting to the client
- **Type-Safety** Minimize numbers of "leaps of faith" casts from `interface{}` so you can rely on the type checker in your logic
- **Fast** Minimal reliance on features like reflection at runtime
- **Minimize Javascript** Most client-side views are 10's of lines of templates and databinding, sometimes without any javascript at all
- **Rich animations and UI** Common operations like cards moving between stacks should make use of flexible, fast animations computed automatically

## Status

The library is currently relativley full-featured. Here are a few of the known gaps (and the issues that track their status):

- **Support of Multiple Browsers** (Issue #324) Currently the web-app only fully works in Chrome, with limited testing in other browsers
- **More Contorl over Animations** (Issue #396) Currently moves that are applied one after another don't pause to allow animations to play
- **Examples with a board** None of the example games in the main repo use a board, which means that tools aimed at board-based-games aren't fleshed out
- **Smooth Upgrading** (Issue #184) If you change the shape of the state objects in your game, there's currently no way to handle older versions of the game stored in your database.

Many more small things are still rough or not implemented. Please file issues or comment on existing issues in the tracker for things you're missing!


