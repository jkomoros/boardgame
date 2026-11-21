/**
 * The renderers compiled into this guest. Each is imported from its game's
 * own client directory, unmodified, through the same `game-src` symlink the
 * webapp uses, so its `../../src/client.js` imports resolve as they do there.
 *
 * Importing a renderer module registers its custom element. Adding a game
 * to the guest means adding its two imports and one row here.
 */
import '@game-src/memory/boardgame-render-game-memory.js';
import '@game-src/memory/boardgame-render-player-info-memory.js';
import { moveInputSchemaFingerprint as memoryFingerprint } from '@game-src/memory/_move_args.js';

import type { BundledGame } from './guest-host.js';

export const bundledGames: Readonly<Record<string, BundledGame>> = Object.freeze({
  memory: { moveInputSchemaFingerprint: memoryFingerprint, hasPlayerInfo: true },
});
