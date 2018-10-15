This document describes at high level how the client side game views are
architected.

Its primary audience is people who are working on the architecure themselves,
for example to add new features to the frameweork. People who are simply using
the framework shouldn't need to understand this document.

`boardgame-game-view` is the top-level view that renders the page for games.
Its responsibility is to identify the game type and game ID to render (based
on URL). It then passes that information to the `boardgame-render-player-
info`, and `boardgame-render-game`, which are responsible for dynamically
loading and instantiating the renderers for that game type.

Within `boardgame-game-view` is an instance of `boardgame-game-state-manager`.
Its job is to fetch state from the server and pass it up to the game-view to
render when a new one should show. It fetches informationa about the game in
general (number of players, their names, etc) and then also opens a socket to
the server which will receive a message whenever a new state is available on
the server. It will then fetch that state from the server (and any states
between what it last fetched and that new version). It will then pass bundles
of state up to game-view to render, one at a time. It waits until the game-
view asks for another one to render (because the previous one is done
animating). It also modifies the state objects as received from the server to
include additional information that is useful for databinding.

`boardgame-game-view` also listens for `propose-move` events emanating from
within the rendererd game, and then forwards them to the `boardgame-admin-
controls`, where the logic to actually serialize them and pass to the server
resides. If the move is successfully applied, the state manager will hear
about it via the socket, and that will kick off more states being downloaded.
(The same mechanism applies no matter if other players or the game itself
applied moves).

`boardgame-render-game`'s primary job is to a) instatiate the specific type of
renderer for this game type (and pass through any updated state it receives
from `boardgame-game-view`) and b) to coordinate animations of state. We will
come back to animations.

Most specific games' renderers inherit from `boardgame-base-game-renderer`.
Its job is entirely to listen for `tap` events on components within the
renderered game view that have attributes about move proposing, and when that
happens, to fire a new `propose-move` event upward for the `game-view` to
capture and process.

The primary goal of a given game renderer is to take a state object and data-
bind it into a well laid-out view. It typically does this by data-binding into
normal layout elements, as well as buttons, `boardgame-component` (and its
sub-types, `boardgame-card` and `boardgame-token`), and `boardgame-component-
stack`. Again, any item that has attributes that talk about propose-move will,
when they are tapped, emit an event that the base game renderer will catch and
then re-throw as a propose-move.

`boardgame-component` and sub-types are (almost?) always endered as children
of `boardgame-component-stack`. `boardgame-component-stack` has a few
responsibilities. It generally creates new `boardgame-components` base on the
data that is bound to it. It renders new components by stamping out copies of
whatever was defined for this deck in `boardgame-deck-defaults`. It also can
do advanced beahvior where for large stacks it only data-binds a few
components for real, and does faux components for others. Its primary job
though is to layout the children components according to its own layout
attributes. They can fan out cards, stack them, arranage in a grid, etc. They
also may perturb the exact position of the children to give a messier layout.
`boardgame-component-stack` also helps out with animations, generating faux
animating elements when necesary. More on that later.

The actual `boardgame-component` are generally either `boardgame-token` or
`boardgame-card`. The former is way simpler; it is just a simple object whose
appearance is defined by the attributes it has. Cards are way more
complicated; they can be tall or wide, rotated or not, and flipped or not. All
of those attributes, when changed, animate. If you select one in the DOM and
change one of those properties, you'll see them animate smoothly. Of course,
that doesn't happen literally in normal practice, because it's all one-way
databound statically from the state. These animations are referred to as
"internal" animations. They affect the layout properties of the component, but
they are based on information set internally.

Note that all animations of all types have a default length set by the CSS var
`--animation-length`. If you want to change the animation, you can target a
different CSS var at the item.

*TODO Describe in detail how external transforms for e.g. messy and fan work.
*(how do they interact with inverse transforms set by component animator? When
*does component-stack use literal layout vs transformed layout)

This is all pretty straightforward. However, the real benefit of the engine is
that it handles animations as components move between states well. At a high
level, the game logic on the server has decided how granularly to break up
moves. Correct animations can only happen between versions; the server game
logic thus decides where full animations MAY happen. It's up to the client to
actually calculate the animations to occur, set them in motion, and figure out
when they're done. _In the future it will also be possible for the client to
decide to skip certain states because it doesn't want to animate each state
change individually, by looking at a before and after state and choosing to
not databind the former._

At a high level, what we do is bind the first state, then bind the second
state as a totally separate item. Items that just so happen to be in the same
place might be re-used by Polymer's data-binding engine, but components that
have logically moved to a different location from state to state (for example,
a card that moved from the draw stack to the discard stack) are almost
certainly represented by different physical DOM nodes before and after.

Most of the magic is organized by `boardgame-component-animator`. Before a new
state is bound, it goes through and collects the current location and state of
all of the components, keeping track of which is which by comparing the "id".
Then it allows the new state to be bound. It then goes through each element
and sees where its new location is. (It does that in one pass before going on
to the next step to avoid layout thrashing).

Now for the hard part. It goes through and generates inverse transforms to
move each component, visually, back to where it was in the previous state, and
then applies a CSS transition to bring it back to the location it literally is
in the DOM, by reducing those transforms to 0. This transformation is referred
to as an "external" aniamtion, because the transforms are set on the
components from the outside. This is a very challenging calculation to do,
especially because components have internal animations that could change their
layout.

Components who are represented by a literal DOM element before and after are
(relatively) easy. Just calculate the inverse transform and apply it.

Slightly more difficult are cases where either before or after had a literal
DOM element, but the other end of the transition doesn't; perhaps it's going
to a `boardgame-component-stack` with so many elements that we print only a
handful of faux components instead of one per actual item. In those cases, we
ask the stack that contains the component to generate a fake component to
animate (the stack gives it a default position in the middle), that will act
like the literal element. When the animation is over, the faux animating
component is removed.

The hardest case is when there is a component who either before or after is
not known to be in a specific location in a stack. This happens, for exmaple,
when a component moves from a normal stack to one that's sanitized with
PolicyLen. That means that the actual list of component IDs is elided, and all
that's left is stack.IdsLastSeen. This captures that the last time the given
ID was seen was in this particular stack, but not _where_ in the stack it was
seen. In this case `boardgame-component-animator` does a behavior like the one
immediatley above. It creates a faux animating element. It positions the
component in the middle of the stack, and styles the element to be very small
and transparent, so as the component animates back to 0 state it's visually
clear which stack the component went to in general, but not where in the
component it went.

*TODO*: Animation-coordinator during prepare and start animator, listens for events to be emitted that are `will-animate` and stores them in a map of id to bool. Then later when `animation-done` is received, it removes each one from the map. It's the component's responsibiltiy to only fire one `animation-done` when it's fully done. When the map has zero entries, it tells the game-view to fetch and render another state from the state manager.

