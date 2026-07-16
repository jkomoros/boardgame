# Werewolf Example

A simplified Mafia/Werewolf game demonstrating the **Table+Hand companion mode** with asymmetric hidden-role information.

## Companion Mode Touchpoints

This is the canonical example of how to add companion mode to a game. The key pieces:

### Go side

- **`state.go:47`** — `playerState` embeds `behaviors.PlayerRole`, which carries `sanitize:"other:hidden"` by default. This single line is what makes roles invisible on the projector (Table view) while visible on each player's own phone (Hand view).
- **`main.go`** — `ConfigureMoves()` does NOT register `moves.ForceFinishTurn` because werewolf uses simultaneous voting (`AnyPlayerIndex`). Turn-based games (like blackjack) should register it — see `examples/blackjack/main.go`.
- **`main_test.go`** — `TestRoleHiddenFromObserver` and `TestRoleVisibleToSelf` pin the privacy contract.

### Client side

- **`client/boardgame-render-game-werewolf-table.ts`** — extends the generated `TableRenderer`. Shows player tiles, vote tallies, and phase info. Roles are NOT shown because the Table connects as `ObserverPlayerIndex` and the sanitization hides them.
- **`client/boardgame-render-game-werewolf-hand.ts`** — extends the generated `HandRenderer`. Shows the player's own role prominently, fellow werewolves (if applicable), and phase-appropriate voting buttons.
- **`client/boardgame-render-game-werewolf.ts`** — fallback solo renderer.

### Opt-in convention

The game becomes companion-capable because `boardgame-util` detects both `*-table.ts` and `*-hand.ts` files in the client directory at build time. No other configuration is needed.

## Game Rules

- 4-7 players, roles: Villager (majority) vs Werewolf (1-2)
- Day: simultaneous public vote to eliminate a suspect
- Night: werewolves choose a target; villagers "sleep"
- Villagers win when all werewolves are eliminated; werewolves win when they equal or outnumber remaining villagers
