/*
Package conformance exports a complete, deterministic record of one game run
on the real Go engine, for holding another implementation of the engine to it.

Given a game delegate and either a scripted scenario (scenario.Spec) or a
recorded golden history (storage/filesystem/record), it replays the run and
emits one Document. For every state version from 0 to the end -- every fix-up
move included, not only settled decision points -- the Document holds:

  - the unsanitized state exactly as storage holds it (timers and secret-move
    counts included), with the game id and secret salt in the header, so a
    consumer can recompute component ids and shuffles;
  - the move record that produced the version;
  - for every viewer (admin, observer, and each player): the game object the
    server would send that viewer (sanitized state with computed properties
    and visibility facets, current player index), the sanitized move record,
    and the legality of each proposable move.

The header carries the chest (decks, enums, constants, legal message
templates), the variant configuration, the static move table and input
schemas, so a Document is self-contained. conformance/schema.ts at the
repository root describes the format for TypeScript consumers.

# How a run is driven

Both inputs reduce to a history. A scenario is first run by scenario.Run, and
its History is then replayed exactly like a golden file. Replay recreates the
game from the history's id, salt, player count and variant, then repeats each
move that starts a causal chain: a player move is proposed, a seat move is
injected through the seating rendezvous, and a timer is forced. Fix-up moves
are never proposed; the engine must produce them. Replay fails if any state or
move differs from the history, so a Document is never exported from a run the
engine no longer reproduces.

# What is captured live and what is state-pure

After each driving step the game is settled, and the head version is captured
exactly as the server would report it. The versions the step passed through
on the way are captured from Game.State(version): sanitization, computed
properties, current player and the sanitized move depend only on the state
and the stored move, so they are what a client fetching that version receives.
Legality at such a version is evaluated against that state with moves bound to
it (api.MoveFormsForViewerAtState). The server only reports legality for the
head, so Version.Settled says which rows the server would really have sent.
Projected move choices exist only at settled versions and only for a seated
player, as on the wire.

# Determinism and normalization

A Document is byte-identical across runs. Every field that is inherently
non-deterministic or depends on the environment is listed in Normalization,
which is copied into each Document as its "normalization" section. In short:

  - Move timestamps and the game's Created/Modified times are wall-clock and
    are omitted.
  - Timers run on a manual clock fixed at Clock. A timer's deadline is the
    clock plus its duration; the clock never advances, and a timer fires
    because the run says so. Timer ids come from State.Rand and are kept.
  - Diagram is a Go-rendered debug string and is omitted from each view.
  - ActiveTimers comes from the manager's process-local scheduler rather than
    a state version. It is rebuilt from the timer records of the version's
    stored state, and checked against the scheduler whenever the version is
    the head.
  - Game.Version, Finished and Winners in a view describe the version being
    viewed, not the head of the game at export time.
  - Form field order follows Go map iteration on the wire. Fields are reduced
    to a name-keyed "defaults" object.
  - All embedded JSON is re-encoded with sorted object keys.

# Regenerating the checked-in corpus

	./scripts/go-local run ./boardgame-util conformance export

writes the corpus defined by the corpus subpackage into conformance/ at the
repository root. The corpus package's test regenerates it in memory and fails
if a checked-in file differs.
*/
package conformance
