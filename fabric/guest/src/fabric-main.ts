/**
 * Entry point of the guest when a Common Fabric pattern embeds it in a
 * `cf-iframe`. The renderer host and the bundled renderers are the same as
 * in the standalone guest (`guest-main.ts`); only the transport differs:
 * cells and streams through the pattern's bridge instead of `postMessage`.
 */
import { FABRIC_CSS, startFabricBridge } from './fabric-bridge.js';
import { bundledGames } from './games.js';
import { installGuestDocument } from './guest-document.js';
import { GuestHost } from './guest-host.js';
import { installAnimTestHooks } from '@boardgame/elements/utils/anim-test-hooks.js';

// See guest-main.ts: the pattern's two-user script reads these counters.
installAnimTestHooks();

const root = installGuestDocument(FABRIC_CSS);

document.addEventListener('securitypolicyviolation', event => {
  console.warn(`[boardgame guest] csp: ${event.effectiveDirective} blocked ${event.blockedURI || 'inline'}`);
});

startFabricBridge(root, post => new GuestHost(root, post, bundledGames));
