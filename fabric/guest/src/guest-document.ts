/**
 * The document-level things the webapp's index.html provides and a guest has
 * to bring itself: theme tokens, base typography, and the element the
 * renderer host mounts into. Shared by the standalone guest (`guest-main.ts`)
 * and the Common Fabric guest (`fabric-main.ts`).
 */
import themeCss from '@client/styles/md3-theme.css?inline';

/** Installs the theme and returns the element to mount the renderer host in. */
export function installGuestDocument(extraCss = ''): HTMLElement {
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
` + extraCss;
  document.head.append(theme);

  const root = document.createElement('div');
  root.id = 'guest-root';
  document.body.append(root);
  return root;
}
