/**
 * Entry point of the guest document. Wires `postMessage` to the GuestHost
 * and installs the document-level things the webapp's index.html provides:
 * theme tokens and base typography.
 */
import themeCss from '@client/styles/md3-theme.css?inline';

import {
  envelope,
  isGuestMessage,
  type GuestToHostMessage,
  type HostToGuestMessage,
  type MessageBody,
} from '../protocol.js';
import { bundledGames } from './games.js';
import { GuestHost } from './guest-host.js';
import { setServerCallReporter } from './shims/api.js';

// The host is whoever embedded this document. An opaque-origin guest cannot
// name its parent's origin, so messages are posted to '*'; they carry only
// what the person did, never state.
const hostWindow: Window = window.parent;
const HOST_MESSAGE_TYPES: ReadonlySet<string> = new Set(['init', 'state', 'move-result']);

function post(message: GuestToHostMessage): void {
  hostWindow.postMessage(message, '*');
}

function report(text: string): void {
  post(envelope<GuestToHostMessage>({ type: 'error', message: text } as MessageBody<GuestToHostMessage>));
}

// Theme tokens (FINDINGS F10). The webapp links this stylesheet from
// index.html; a guest has no document to inherit it from.
const theme = document.createElement('style');
theme.textContent = themeCss + `
  html, body { margin: 0; padding: 0; }
  body {
    /* Web fonts are omitted (FINDINGS F9). These stacks give the theme's
       named families a local fallback of the same class. */
    --md-sys-typescale-display-small-font: 'Crimson Text', 'Iowan Old Style', Georgia, 'Times New Roman', serif;
    font-family: 'Source Sans 3', system-ui, -apple-system, 'Segoe UI', Roboto, sans-serif;
  }
  /* The webapp's game view sets the default animation length on itself
     (FINDINGS F11). Without it every animation runs at a 250ms fallback. */
  #guest-root { padding: 12px; box-sizing: border-box; --animation-length: 0.5s; }
  .guest-seats { display: flex; flex-wrap: wrap; gap: 8px 24px; align-items: center; margin-bottom: 8px; }
  .guest-seats:empty { display: none; }
  .guest-seat { display: flex; gap: 8px; align-items: center; }
  .guest-seat-name { font-weight: 600; border-bottom: 2px solid var(--seat-color, transparent); }
`;
document.head.append(theme);

const root = document.createElement('div');
root.id = 'guest-root';
document.body.append(root);

const host = new GuestHost(root, post, bundledGames);

setServerCallReporter(name => report(`the client tried to reach the server (${name})`));
window.addEventListener('error', event => report(`uncaught: ${event.message}`));
window.addEventListener('unhandledrejection', event => report(`unhandled rejection: ${String(event.reason)}`));

// Under a Content-Security-Policy, say what was blocked instead of failing
// silently (FINDINGS, "Content Security Policy").
document.addEventListener('securitypolicyviolation', event => {
  report(`csp: ${event.effectiveDirective} blocked ${event.blockedURI || 'inline'}`);
});

window.addEventListener('message', (event: MessageEvent) => {
  // Only the embedding window may drive the guest.
  if (event.source !== hostWindow) return;
  if (!isGuestMessage(event.data)) return;
  // Ignore everything but the three host messages, silently. Opened as a
  // top-level document the guest is its own parent and hears what it posts;
  // answering an unknown type with an error would then never stop.
  if (!HOST_MESSAGE_TYPES.has(event.data.type)) return;
  try {
    host.handle(event.data as HostToGuestMessage);
  } catch (error) {
    report(`failed to handle ${event.data.type}: ${String(error)}`);
  }
});

post(envelope<GuestToHostMessage>({
  type: 'ready', stage: 'loaded', games: host.bundledGames,
} as MessageBody<GuestToHostMessage>));
