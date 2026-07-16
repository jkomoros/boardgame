import type { RawGameState } from './game-state.js';

declare const rawState: RawGameState;

// The application shell transports creator state opaquely. Only generated
// per-game State types may advertise creator-owned property names.
// @ts-expect-error Core raw game state has no creator-specific fields.
rawState.Game.Score;
// @ts-expect-error Core raw player state has no creator-specific fields.
rawState.Players[0]?.Hand;
