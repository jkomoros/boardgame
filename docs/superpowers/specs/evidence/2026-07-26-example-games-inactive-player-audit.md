# Evidence pack: seated-player deadlock audit (checkers, tictactoe, debuganimations)

**Background.** `moves.SeatPlayer.Apply` marks every seated player INACTIVE
(seat_player.go:244) when the playerState embeds `behaviors.InactivePlayer`.
An inactive player makes `base.GameDelegate.PlayerMayBeActive` false
(base/game_delegate.go:733) → `PlayerIndex.Valid` false (state.go:473) →
`moves.CurrentPlayer.Legal` fails with "The specified target player is not
valid" (moves/current_player.go:61) — permanently, since only
`moves.ActivateInactivePlayer` clears the flag. Memory and pig were fixed
for this on 2026-07-25; this audit covers the remaining three.

**Why ordinary tests never caught it.** `manager.NewDefaultGame()` leaves
players unseated (SeatPlayer needs the server's rendezvous injection), so
nothing is ever marked inactive in-process and every move stays legal.

## Verdicts

| Game | InactivePlayer | CurrentPlayer-gated moves | Verdict |
|---|---|---|---|
| checkers | yes (state.go:24) | 3 (moves.go) | **DEADLOCKED → fixed** |
| tictactoe | yes (state.go:53) | 2 (moves.go) | **DEADLOCKED → fixed** |
| debuganimations | yes (state.go:50) | 0 (all 13 are `moves.Default`) | **no change needed** |

debuganimations is not merely argued safe: its moves were driven repeatedly
in real seated play throughout the animation-parity work (Swap, To Hidden,
Public Shuffle e2e scenarios) with auto-seated players and no deadlock.

## Empirical proof (live offline-dev server, real auto-seated play)

Reading `PlayerInactive` from the live client store after game creation:

- **Before the fix:** `tictactoe {"players":[{"inactive":true},{"inactive":true}]}`,
  `checkers {"players":[{"inactive":true},{"inactive":true}]}` — seated and
  permanently inactive.
- **After the fix:** both games report `inactive:false` for both players —
  ActivateInactivePlayer fires as a fix-up and play proceeds.

The before-state was produced by reverting both main.go files, rebuilding
`boardgame-util`, and restarting the server; the after-state by restoring,
rebuilding, and restarting. Same probe both times.

## Permanent regression test

`examples/{checkers,tictactoe,debuganimations}/inactive_player_test.go`
asserts the general invariant rather than a per-game constant: a game that
seats players AND gates any move on `moves.CurrentPlayer` MUST configure
`moves.ActivateInactivePlayer`. debuganimations passes on its own merits
(no CurrentPlayer-gated moves), so the test encodes the real rule and will
catch a future game that adds a CurrentPlayer move without activation. The
CurrentPlayer embed is detected structurally (reflection) because
`moves.CurrentPlayer`'s marker method is package-private.

Both fixes were watched fail-first with the exact deadlock message.
