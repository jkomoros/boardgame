/**
 * The renderer authoring facade, at the path every game client imports it
 * from. A game's client directory is linked at `game-src/<game>/`, so its
 * `../../src/client.js` lands here, exactly as it lands on
 * `server/static/src/client.ts` in the webapp.
 */
export * from '@boardgame/renderer-kit/client.js';
