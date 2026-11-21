/**
 * The renderers compiled into this guest. Each is imported from its game's
 * own client directory, unmodified, through a link under this package's
 * `game-src/` (the webapp's convention), so its `../../src/client.js` import
 * lands on this package's `src/client.ts` and from there on the renderer kit.
 *
 * Importing a renderer module registers its custom element. Adding a game
 * to the guest means linking its client directory under `game-src/`, and
 * adding its imports and one row here. Which one is mounted is decided at
 * run time by the `gameName` of the `init` message; only the solo surface of
 * each game is bundled.
 */
import '../game-src/memory/boardgame-render-game-memory.js';
import '../game-src/memory/boardgame-render-player-info-memory.js';
import { moveInputSchemaFingerprint as memoryFingerprint } from '../game-src/memory/_move_args.js';

import '../game-src/blackjack/boardgame-render-game-blackjack.js';
import '../game-src/blackjack/boardgame-render-player-info-blackjack.js';
import { moveInputSchemaFingerprint as blackjackFingerprint } from '../game-src/blackjack/_move_args.js';

import type { BundledGame } from './guest-host.js';

export const bundledGames: Readonly<Record<string, BundledGame>> = Object.freeze({
  blackjack: { moveInputSchemaFingerprint: blackjackFingerprint, hasPlayerInfo: true },
  memory: { moveInputSchemaFingerprint: memoryFingerprint, hasPlayerInfo: true },
});
