# Executable game scenarios

`boardgame-util/lib/scenario.Run` drives the real game engine from a small script.
Use it for an example that needs to prove rules, automatic moves, timers, hidden
information, and rendering agree. The result contains two separate artifacts:

- `History` uses the existing filesystem record format and can be replayed with
  `golden.Compare`. It includes authoritative hidden state.
- `Replay` contains API snapshots at settled decision boundaries for explicitly
  requested viewers. It can be rendered with `scenarioFixtureSnapshot` and the
  existing renderer fixture tools.

```go
result, err := scenario.Run(memory.NewDelegate(), scenario.Spec{
    Name: "Reveal one card", Seed: "memory-example", Players: 2,
    Viewers: []boardgame.PlayerIndex{boardgame.ObserverPlayerIndex, 0},
    Steps: []scenario.Step{{
        Label: "Reveal", Move: "Reveal Card", Player: 0,
        Input: map[string]interface{}{"CardIndex": 0},
    }},
})
```

Moves pass through `Game.MoveWithInput` and versioned proposal validation. The
binder accepts the declared creator inputs, rejects unknown/context-owned fields,
and preserves normal defaults. `Timer: true` fires the next committed timer;
`Seat: &player` uses the same seating rendezvous as golden replay. Each step
selects exactly one operation. No alternate rule evaluator is involved.

A seed fixes engine RNG identity and the runner uses a fixed manual timer clock.
Creator code must use `State.Rand()` for reproducibility. Explicit steps are
limited to 256; `MaxMoves` includes automatic moves and defaults to 256 (maximum
4096). These limits do not interrupt arbitrary creator callbacks. Runs close their
manager and return the completed prefix alongside an annotated error.

## Memory example

`examples/memory/scenarios.Mismatch` chooses a mismatching pair from an isolated
setup record and then emits an ordinary explicit script. Generate its visual
artifact and optional golden history with:

```sh
./scripts/go-local run ./examples/memory/scenarios/cmd/export \
  --replay examples/memory/client/scenario-replay.json \
  --history /tmp/memory-scenario.json
```

Start the normal development server and open `/scenario-review.html`. Step through
setup, first reveal, mismatch, and timer-driven hide; change the viewer to inspect
the same decision from another seat. Renderer clicks record fixture proposals;
the Next button advances to the next recorded state.

The client adapter shares live rendering's state expansion. It rejects a mismatched
game, missing viewer, invalid frame, or inconsistent state version. Default-bound
move legality has the same meaning as the ordinary move tray; it does not promise
that all possible inputs are legal.

A multi-viewer replay is a local review artifact: each individual snapshot is
sanitized, but combining seats can reveal their private information. Publish only
the viewer snapshots appropriate for the audience. Authoritative `History` is
never embedded in `Replay`.

For explicit bot decisions, set `Bot` and `Player` instead of `Move`/`Input` on a
step. `RunContext` passes cancellation to policies and checks it between steps.
See [observation bots](observation-bots.md) for the capability boundary and pilot.

## Native choices and other games

Replays include the API's actor-only candidate projections and default move-tray
legality, using the same visibility filtering as live `/info`. Recorded controls
use exact candidate inputs and explicit evidence for moves needing no input; inputs without recorded evidence stay disabled.
Declare an ordinary bounded choice with `moves.WithChoices` to make it available
both to native controls and replay previews. Handcrafted renderer fixtures retain
their explicit simulated preview behavior.

The review page accepts `?game=darwin` or another configured game with a
`client/scenario-replay.json`. Companion games can use `&surface=hand&viewer=0`
or `&surface=table`. Werewolf's `examples/werewolf/scenarios.TimedVote` includes
seating, a partial vote, and deadline resolution; generate it with:

```sh
./scripts/go-local run ./examples/werewolf/scenarios/cmd/export
```

The paired Games repository includes Darwin's simultaneous food commitments and
Secret Groups' clue-giver/guesser/observer comparison. Each is an ordinary Go
scenario, and each generated history is tested through golden replay.
