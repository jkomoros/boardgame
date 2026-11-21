/**
 * Compatibility entry point. The renderer authoring facade is
 * `@boardgame/renderer-kit/client.js` (packages/renderer-kit/src/client.ts).
 *
 * Game clients, the code generator's templates and the client checker all
 * name this file as `../../src/client.js`, relative to a game's directory
 * under `game-src/`. That specifier keeps working through this re-export.
 */
export * from '@boardgame/renderer-kit/client.js';
