# Bots with bounded observations

Use `bots.Policy` when a bot should reason from the same information a player
receives. Its `Decide(context.Context, bots.Observation, bots.Random)` method
returns an ordinary move name and creator inputs. It receives detached API JSON
and public component data, so it cannot reach the authoritative state through
`State.Game()` or mutate a component while choosing a move.

`bots.Play` observes a concrete player, calls the policy, binds the returned input
through `Game.MoveWithInput`, and submits against the observed version. A move
chosen before another commitment is rejected as stale. Unknown moves, fix-ups,
and context-owned input fields are rejected. Policy errors retain their cause and
include player/version context. Cancellation is checked before and after the
policy call; policies should also check their context while doing lengthy work.

The caller supplies randomness separately from the engine RNG. Scenario `Bot`
steps use a reproducible stream derived from the scenario seed. For example:

```go
scenario.Step{Label: "Choose a square", Player: 0, Bot: bot.RandomEmpty{}}
```

`examples/tictactoe/bot.RandomEmpty` is a small perfect-information example: it
reads empty squares from the observation and returns a Place Token input. Its
scenario is replayed as an ordinary golden. Memory tests verify that observations
hide card positions and that editing an observation cannot mutate engine state.

This API does not enumerate legal moves or promise a strong strategy. A game may
supply its own policy and memory. The existing `boardgame.Agent` interface remains
available; it receives full engine access and is not silently reinterpreted.
These are in-process creator callbacks. Cooperative context checks do not kill
stuck code, and submitting a move cannot be undone by later cancellation. Use a
separate process when hard execution deadlines or isolation are required.
