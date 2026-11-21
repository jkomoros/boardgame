import { css, type CSSResultGroup } from 'lit';

/**
 * THE VOCABULARY EVERY RENDERER REBUILT.
 *
 * A line-by-line audit of every renderer in both repos found the same two
 * blocks copied into shadow root after shadow root:
 *
 *   - **Flex utilities, 5 of 5 games.** `.horizontal{display:flex;
 *     flex-direction:row}` appears verbatim in `pass`, `valentine`,
 *     `murdermrmonroe`, `blackjack` and `pig`, and EIGHT identical copies live
 *     in `debuganimations` alone. The framework's own chrome components
 *     (`boardgame-admin-controls`, `boardgame-create-game`, `boardgame-user`,
 *     the roster pair, `boardgame-game-item`) each declare it a ninth through
 *     fourteenth time.
 *   - **"This is selected", 4 of 5 games, four incompatible visual
 *     languages.** A `::part(panel)` background (valentine), `outline: 3px
 *     solid` in two colours (darwin), a wrapper background class
 *     (murdermrmonroe), a tinted border (sequenceforge). Plus four unrelated
 *     languages for "eliminated": `filter: saturate(.5) blur(1px)`, opacity
 *     with a strikethrough, a red bordered banner, and a bare `#c62828`.
 *
 * There is not a single `TODO`, `FIXME` or `HACK` in any client directory in
 * either repo. Nobody experienced writing these by hand as friction; they
 * experienced it as how you build a renderer. So the fix has to be a default
 * that arrives without being asked for, which is why these are attached to the
 * generated renderer base classes rather than left as an optional import.
 *
 * ## Why shared `CSSResult`s and not something cleverer
 *
 * Renderers are Lit elements with their own shadow roots, so a global
 * stylesheet does not reach them. Three mechanisms were considered:
 *
 *   1. **A global stylesheet.** Does not cross a shadow boundary. Rejected on
 *      the mechanics.
 *   2. **Custom properties alone.** These DO inherit through shadow roots, and
 *      the state colours below are custom properties for exactly that reason —
 *      a game retunes the whole language from one place, including inside a
 *      component's shadow tree. But a custom property cannot carry
 *      `display: flex`, so it cannot be the whole answer.
 *   3. **Shared `CSSResult`s composed into `static styles`.** This is what the
 *      repo already does (`components/shared-styles-lit.ts`, imported by
 *      `boardgame-game-view`), it is what Lit is designed for, and adopted
 *      stylesheets are shared between every instance rather than re-parsed.
 *
 * So: rules ship as `CSSResult`s, the colours inside them ship as custom
 * properties, and both are wired into the renderer base classes. A renderer
 * picks them up with no import at all as long as it COMPOSES `static styles`
 * rather than replacing it — `...(GameRenderer.styles ? [GameRenderer.styles] :
 * [])` first, which is what every example renderer writes and what the stub
 * generator emits, so a scaffolded game gets the vocabulary by default. They
 * are also re-exported from `src/client.js` for a renderer that replaces
 * `styles` outright instead of composing it.
 *
 * ## Why classes and not attribute selectors
 *
 * `[active]` and `[selected]` would read better than `class="active"` and would
 * kill the hand-rolled class-name concatenators three games wrote. They are
 * nonetheless wrong here: `boardgame-player-panel` declares
 * `@property({type: Boolean, reflect: true}) active`, and
 * `boardgame-selection-option` and `boardgame-placement-item` both reflect
 * `disabled`. An `[active]` rule in a renderer's shadow root would silently
 * start ringing every player panel a game binds `.active=${...}` on. Lit's own
 * `classMap` directive is the answer to the concatenator, and the tutorial says
 * so rather than this module inventing a second one.
 */

/**
 * Row/column/centre/flex. Deliberately boring; the point is that it stops being
 * copied.
 *
 * `.horizontal` and `.vertical` set direction only — no gap and no alignment —
 * because that is exactly the rule the five games copied, so adopting this
 * cannot move a pixel in any of them. `.center` is the modifier three of them
 * paired it with, and it means `align-items` in every single copy found.
 *
 * Gap is the one value that genuinely varied (16px, 12px, 8px, 4px), so `.gap`
 * reads `--boardgame-gap` rather than freezing a number: set the property on
 * the element (or any ancestor) to retune it. 16px is the default because it is
 * what seven of debuganimations' eight copies used.
 */
export const layoutStyles = css`
  .horizontal {
    display: flex;
    flex-direction: row;
  }

  .vertical {
    display: flex;
    flex-direction: column;
  }

  .center {
    align-items: center;
  }

  .justify-center {
    justify-content: center;
  }

  .space-between {
    justify-content: space-between;
  }

  .space-around {
    justify-content: space-around;
  }

  .wrap {
    flex-wrap: wrap;
  }

  .gap {
    gap: var(--boardgame-gap, 16px);
  }

  .flex {
    flex: 1;
  }
`;

/**
 * One state language, replacing four.
 *
 * **Two kinds of active player, both first class.** `murdermrmonroe`
 * distinguishes `.current` (whose turn it is) from `.current-luck` (who is
 * being asked to answer right now), and valentine's tertiary-container
 * highlight is the same second idea under a different name. A vocabulary with
 * only one "active" guarantees the second one keeps getting reinvented, so
 * there are two: `.active` for whose turn it is, and `.responding` for whoever
 * the game is currently waiting on. They deliberately read as different hues
 * (tertiary vs primary), matching the pair murdermrmonroe arrived at.
 *
 * **The ring is an `outline`, not a `border`.** `outline` is outside the box
 * model, so toggling a state cannot reflow its neighbours — which is the bug
 * `blackjack` worked around by giving every seat a permanent
 * `border: 2px solid transparent`. `--boardgame-state-ring-offset` defaults to
 * a NEGATIVE offset so the ring is drawn inside the element's own edge, where a
 * border would have been, rather than bleeding outward over a neighbour.
 *
 * **`.eliminated` dims the subject and says nothing.** The four hand-rolled
 * versions split into two different jobs: draining a tile of colour, and
 * writing the word "Eliminated" in red. Only the first is state. The words a
 * game uses are content and stay the game's business.
 *
 * Nothing here transitions or animates. Renderer state changes land in the
 * middle of the animation gate, and a stray CSS transition on a colour is
 * exactly the kind of thing that shows up later as an unexplainable frame.
 */
export const stateStyles = css`
  .active,
  .responding,
  .selected,
  .targetable {
    outline-width: var(--boardgame-state-ring-width, 2px);
    outline-offset: var(--boardgame-state-ring-offset, -2px);
    outline-style: solid;
    border-radius: var(--boardgame-state-radius, 12px);
  }

  /* Whose turn it is. */
  .active {
    outline-color: var(--boardgame-state-active-ring, var(--md-sys-color-tertiary, #7D5260));
    background-color: var(--boardgame-state-active-surface, transparent);
  }

  /* Who the game is waiting on right now, which is not always whose turn it
     is: a simultaneous response, a reaction window, a luck roll. */
  .responding {
    outline-color: var(--boardgame-state-responding-ring, var(--md-sys-color-primary, #6750A4));
    background-color: var(--boardgame-state-responding-surface, transparent);
  }

  /* This player picked this thing. */
  .selected {
    outline-color: var(--boardgame-state-selected-ring, var(--md-sys-color-secondary, #625B71));
    background-color: var(--boardgame-state-selected-surface, transparent);
  }

  /* You MAY pick this thing. An invitation, so it is dashed and tints nothing. */
  .targetable {
    outline-style: dashed;
    outline-color: var(--boardgame-state-targetable-ring, var(--md-sys-color-outline, #79747E));
  }

  .disabled {
    opacity: var(--boardgame-state-disabled-opacity, 0.5);
    cursor: not-allowed;
    pointer-events: none;
  }

  /*
   * One filter and one token, because opacity() is itself a filter
   * function: a game that wants a different treatment (blackjack's busted seats
   * blur as well as drain) replaces the whole thing with ONE declaration
   * instead of inventing a fifth private class. The default does NOT blur --
   * blur destroys the legibility of the score a player still needs to read.
   */
  .eliminated {
    filter: var(--boardgame-state-eliminated-filter, saturate(0.35) opacity(0.55));
  }
`;

/**
 * Both halves, in the order a renderer wants them: the vocabulary first so a
 * renderer's own rules win a specificity tie.
 */
export const rendererStyles: CSSResultGroup = [layoutStyles, stateStyles];
