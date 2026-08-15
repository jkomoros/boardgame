/**
 * The one place a piece of art becomes CSS.
 *
 * Issue #600 has been open since 2018: the framework has had an art pipeline
 * (`boardgame-util imagegen`, the shared `art/house-styles/tactile-naturalist`)
 * and no way to attach the result to a component. `darwin` ships 46 MB of
 * assets and applies three of them -- a card back, a round chip and a board mat,
 * all GENERIC component art -- as `background:` shorthands on hand-rolled divs,
 * because there was nowhere else to put them.
 *
 * This module is what the four art slots (`boardgame-card`'s `art` and
 * `back-art`, `boardgame-token`'s `art`, and the board/surface `art`) share, and
 * the reason it is a module rather than four copies of a `background:` string:
 *
 * ## The URL has to be escaped, and none of the hand-rolled copies did it
 *
 * Every game reaches its assets through
 * `new URL('./assets/x.png', import.meta.url).href`, which is a real URL that
 * has been through the bundler. Vite's asset hashing is quote-free, but a
 * project directory containing a `"` or a `)` -- or, far more likely, a
 * `?t=1234` dev-server cache-buster followed by anything -- lands unescaped
 * inside `url(...)` and silently breaks the declaration, taking the rest of the
 * shorthand with it. `cssUrl` produces the quoted, escaped form once.
 *
 * ## The wash is the mat's whole trick, and it is one string
 *
 * `darwin`'s board is `linear-gradient(#f4eddde8, #f4eddde8), url(mat) center /
 * cover` -- a flat translucent sheet over the photograph so that text stays
 * legible on top of it. That is the difference between "a texture" and "a
 * surface you can put a game on", it is not obvious, and it is the kind of thing
 * that gets rediscovered rather than reused. `artLayerStyle` takes the wash as
 * an argument.
 *
 * Everything here is a pure function of strings, deliberately, so
 * `component-art.test.ts` can run it under `node --test` with no DOM.
 */

/** How art fills the box it is given. The two values `background-size` means. */
export type ArtFit = 'cover' | 'contain';

const ART_FITS: readonly string[] = Object.freeze(['cover', 'contain']);

export function isArtFit(value: unknown): value is ArtFit {
  return typeof value === 'string' && ART_FITS.includes(value);
}

/**
 * `url("…")`, with the two characters that can end the string escaped and the
 * two that can end the DECLARATION removed.
 *
 * A newline inside a CSS string is a parse error rather than an escape problem,
 * so it is dropped rather than backslash-escaped: a URL cannot legally contain
 * one, and a caller that has managed to build one has a bug upstream that a
 * mangled-but-parsing declaration would hide.
 */
export function cssUrl(url: string): string {
  const escaped = url
    .replace(/[\\"]/g, (character) => `\\${character}`)
    .replace(/[\n\r\f]/g, '');
  return `url("${escaped}")`;
}

/**
 * The complete `background` shorthand for one piece of art, ready for a `style`
 * attribute.
 *
 * `no-repeat` is not optional and there is deliberately no knob for it: a
 * component's art is a picture OF that component, and a tiled second copy of a
 * card back peeking out of the corner is never what anybody meant. A game that
 * wants a repeating texture is describing a background, not component art, and
 * should say so in its own CSS.
 */
export function artLayerStyle(
  art: string,
  fit: ArtFit = 'cover',
  wash?: string,
): string {
  if (typeof art !== 'string' || art === '') {
    throw new Error('artLayerStyle(): art must be a non-empty URL string');
  }
  if (!isArtFit(fit)) {
    throw new Error(`artLayerStyle(): fit must be "cover" or "contain", not ${JSON.stringify(fit)}`);
  }
  // Every layer carries its own position/size/repeat. In the `background`
  // SHORTHAND those apply only to the layer they are written on, so the
  // hand-written `linear-gradient(w, w), url(x) center / cover no-repeat` form
  // leaves the wash at `auto` -- which happens to work for a gradient, whose
  // painting area is the whole box, and would quietly stop working the moment
  // the wash became anything with an intrinsic size.
  const placed = (layer: string): string => `${layer} center / ${fit} no-repeat`;
  const layers = wash
    ? `${placed(`linear-gradient(${wash}, ${wash})`)}, ${placed(cssUrl(art))}`
    : placed(cssUrl(art));
  return `background: ${layers};`;
}
