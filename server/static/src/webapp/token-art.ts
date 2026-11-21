/**
 * Where the webapp serves the framework's shipped token art: `src/assets`,
 * relative to the document, which every webapp page serves from the site
 * root. This is the path boardgame-token used to hard-code; it is a fact
 * about this host, so it lives here and reaches the element through
 * setTokenArtResolver (components/component-art.ts).
 *
 * Deliberately import-free: the renderer fixture and the standalone demo
 * pages install it too.
 */
export function webappTokenArt(type: string): string {
  return `src/assets/token_${type}.svg`;
}
