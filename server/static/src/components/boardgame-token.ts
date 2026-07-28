import { BoardgameComponent } from './boardgame-component.js';
import { html, css, CSSResult, TemplateResult, unsafeCSS } from 'lit';
import { customElement, property } from 'lit/decorators.js';
import { classMap } from 'lit/directives/class-map.js';
import { repeat } from 'lit/directives/repeat.js';
import { motionSilhouette } from '../motion/subject.js';
import type { MotionSubjectSnapshot } from '../motion/subject.js';
import { cssNumber } from '../solid/screen-frame.js';
import {
  CAMERA_DEPTH_WIDTHS,
  TOKEN_COLOR_FILTERS,
  isTokenSolidShape,
  tokenSolid,
  type TokenSolid,
} from './token-solid.js';

/**
 * The `#outer.<color> img` rules, generated from the one table that also feeds
 * the solids' arithmetic.
 *
 * They used to be authored here, and a 3D token would then have had a SECOND
 * list of colours that had to agree with this one by hand. Generating them means
 * the flat art and the solid are recoloured by the same nine strings — see
 * `token-solid.ts`'s `TOKEN_COLOR_FILTERS`. Red is absent on purpose: it is the
 * colour the art is drawn in, so it needs no filter and never had a rule.
 */
const COLOR_FILTER_RULES = Object.entries(TOKEN_COLOR_FILTERS)
  .map(([name, filter]) => `#outer.${name} img { filter: ${filter}; }`)
  .join('\n      ');

@customElement('boardgame-token')
export class BoardgameToken extends BoardgameComponent {
  static override styles: any = [
    BoardgameComponent.styles,
    css`
      #inner {
        height: var(--component-effective-height);
        width: var(--component-effective-width);
      }

      #inner img {
        height: 100%;
        width: 100%;
      }

      /*
       * THE CAMERA, on #outer, exactly where boardgame-card.ts puts its own.
       *
       * It cannot go on #inner: #inner is what motionTrackTarget('visual')
       * returns, so its transform belongs to the animation kernel, and a camera
       * written there would be replaced outright the first time anything plays a
       * visual track (the same argument boardgame-die.ts makes for its #orient).
       * It reaches the facets from #outer even though #outer carries the
       * alt-shadow filter: measured, the probe in
       * parity/component-3d-context.spec.ts reads 25px with this rule and 20px
       * without it.
       *
       * The depth is in TOKEN WIDTHS rather than the card's flat 1000px, so that
       * two tokens at different --component-scale values (pass puts 55 on screen
       * at two scales) are drawn by the same lens and not by two. See
       * CAMERA_DEPTH_WIDTHS.
       *
       * The explicit box is what keeps perspective-origin honest: it defaults to
       * the centre of #outer's own box, and #outer is a block that would
       * otherwise stretch to whatever cell the stack put it in, projecting the
       * solid about a point that is not the solid's centre.
       */
      #outer.solid {
        width: var(--component-effective-width);
        height: var(--component-effective-height);
        perspective: calc(var(--component-effective-width) * ${unsafeCSS(CAMERA_DEPTH_WIDTHS)});
      }

      /*
       * The 3D carrier. Nothing from here down may take a grouping property --
       * a filter, an opacity below 1, a clip-path, an overflow other than
       * visible -- because each of those forces transform-style back to flat and
       * collapses the solid into a pile of overlapping outlines. That is the
       * whole reason the elevation and the throb live on #outer (see
       * boardgame-component.ts's styles and _syncThrob below). The facets
       * themselves do carry clip-path, which is fine: they are leaves.
       */
      #outer.solid #inner {
        transform-style: preserve-3d;
      }

      /*
       * The solid's own element: the sizing (its font-size IS the solid's size,
       * so every generated length below can be an em and the whole thing follows
       * --component-width with no JavaScript remeasurement) and the resting pose,
       * both written from the template on every render.
       */
      #solid {
        position: relative;
        width: 100%;
        height: 100%;
        transform-style: preserve-3d;
      }

      /*
       * One element per surface polygon. backface-visibility is the whole of the
       * hidden-surface removal, and it is exactly right here because every shape
       * that renders as a solid is convex: its camera-facing facets tile the
       * silhouette once and the rest must be culled.
       *
       * NO will-change. The die promotes every facet because a facet that
       * crosses from back-facing to front-facing DURING a tumble would otherwise
       * stay culled for a frame or two and tear a hole in the solid -- and a
       * promotion has to be in place before the animation starts. A token does
       * not tumble; its pose is a constant. Declining promotion is what keeps 55
       * of these at 60fps, and it is measured: promotion changes the frame rate
       * at 24 sides by 0.4fps, i.e. not at all.
       */
      .facet {
        position: absolute;
        left: 50%;
        top: 50%;
        /* width/height/margin/transform/clip-path/background are per facet. */
        backface-visibility: hidden;
        -webkit-backface-visibility: hidden;
      }

      #outer.pawn {
        --component-aspect-ratio: 2.0;
      }

      #outer.meeple {
        --component-aspect-ratio: 1.25;
      }

      /* Declared on #outer, which is where the throb that reads them plays.
         They still inherit down to #inner, so a reader that has not caught up
         sees the same values. */
      #outer.active {
        --throb-color-from: rgba(136,136,38,1.0);
        --throb-color-to: rgba(136,136,38,0.5);
      }

      #outer.highlighted {
        --throb-color-from: rgba(0,0,0,1.0);
        --throb-color-to: rgba(0,0,0,0.5);
      }

      #outer.active.highlighted {
        --throb-color-from: rgba(255,255,0,1.0);
        --throb-color-to: rgba(255,255,0,0.0);
      }

      /* The per-colour filters for the FLAT art, generated from the same table
         the solids' base colours are computed from. Red is the colour the art is
         already drawn in, so it has no rule. */
      ${unsafeCSS(COLOR_FILTER_RULES)}
    `
  ];

  // Color to set. One of the colors returned by legalColors.
  @property({ type: String })
  color = 'red';

  // Active changes the styling to make it clear the thing is selected
  @property({ type: Boolean })
  active = false;

  // highlighted has a different visual style than active. Different
  // games will use it for different things.
  @property({ type: Boolean })
  highlighted = false;

  // The type of token. Supported values: "token" (default), "chip",
  // "cube", "pawn", "meeple"
  @property({ type: String })
  type = 'token';

  get legalTypes(): string[] {
    return [
      'token',
      'chip',
      'cube',
      'pawn',
      'meeple',
      'disc',
    ];
  }

  get legalColors(): string[] {
    return [
      'gray',
      'green',
      'teal',
      'purple',
      'pink',
      'red',
      'blue',
      'yellow',
      'orange',
      'black',
    ];
  }

  override motionSubjectSnapshot(): MotionSubjectSnapshot {
    return motionSilhouette(
      this.type === 'token' || this.type === 'chip' || this.type === 'disc'
        ? 'circle'
        : 'rounded-rectangle',
    );
  }

  override firstUpdated(_changedProperties: Map<PropertyKey, unknown>) {
    super.firstUpdated(_changedProperties);
    this.altShadow = true;
  }

  // The infinite highlight throb (#outer's drop-shadow pulse while
  // active/highlighted). Ambient decoration, not a state-arrival cue: it
  // must never hold the completion gate, so it is always started with
  // { gated: false }. Tracked here so a state change or disconnect can
  // cancel the prior instance before starting or discarding a new one.
  private _throb: Animation | null = null;

  private _computeAsset(type: string): string {
    return `src/assets/token_${type}.svg`;
  }

  // Override _computeClasses and add some more.
  //
  // The colour and the type name their own classes, so BOTH have to survive
  // being empty -- and a stack is entitled to pass an empty colour: checkers
  // passes `''` for every component whose colour this player may not see, which
  // on the first render pass is all of them, because the view runs before the
  // state arrives.
  //
  // An empty key here is not a harmless no-op. `classMap` records it as a class
  // it applied, and on the next update -- the one where the colour finally has
  // a name -- removes it with `classList.remove('')`, which throws a
  // SyntaxError from inside Lit's own update. That aborts `performUpdate`, and
  // Lit never retries: the class list, the item, the spacer flag and the
  // rendered content freeze at the poisoned pass's values permanently. It
  // emptied the entire checkers board, silently. See
  // parity/token-unnamed-color.spec.ts.
  protected override _computeClasses(): Record<string, boolean> {
    const result = super._computeClasses();
    const named: Record<string, boolean> = {};
    for (const name of [this.color.toLowerCase(), this.type]) {
      if (name) named[name] = true;
    }
    return {
      ...result,
      ...named,
      active: this.active,
      highlighted: this.highlighted,
      solid: this._solid() !== null,
    };
  }

  /**
   * The solid this token draws, or null for the flat art -- and it is a PURE
   * FUNCTION OF CURRENT STATE, deliberately, which is the single most important
   * fact about this component.
   *
   * A stack pools and reparents component hosts across membership changes and
   * nothing re-derives a pose on reuse, so a node carries whatever its previous
   * occupant's last write left on it. That is the failure `boardgame-die.ts`'s
   * `_clearRoll` exists to prevent -- a die that dropped a roll without clearing
   * the transform it wrote sat 60 to 106px outside its own slot, permanently --
   * and here it would be worse, because a token's motion carriers are `noAnimate`
   * and can never self-correct by playing anything.
   *
   * So there is no imperative write anywhere in this component's 3D path. The
   * pose, the size, the facet placements and the colours are all rendered from
   * the template, from this function, on every update; `token-solid.ts` holds no
   * state but a cache keyed by the same two strings it is called with. A pooled
   * host that arrives with a new `type` or `color` re-renders every one of them,
   * and one that arrives as a `spacer` drops the scene entirely.
   */
  private _solid(): TokenSolid | null {
    // A spacer has no item to stand for. It is `visibility: hidden` and exists
    // only to hold a slot open, so it must not build a scene at all -- 14 facet
    // elements per empty square of a board is exactly the kind of cost nobody
    // would ever see and everybody would pay.
    if (this.spacer) return null;
    if (!isTokenSolidShape(this.type)) return null;
    return tokenSolid(this.type, this.color);
  }

  protected override updated(changedProperties: Map<string, any>) {
    super.updated(changedProperties);
    if (changedProperties.has('active') || changedProperties.has('highlighted')) {
      this._syncThrob();
    }
  }

  // Start (or stop) the ambient throb to match active/highlighted. Always
  // cancels the prior instance first: the CSS custom properties that carry
  // the theme colors (--throb-color-from/-to) can change value across an
  // active<->highlighted transition even though the throb-state boolean
  // (active || highlighted) stays true the whole time, so a stale
  // in-flight animation would keep animating the OLD colors. WAAPI
  // keyframes cannot resolve var() portably, so the colors are read once
  // via getComputedStyle at (re)start time -- same restart-on-change
  // tradeoff the legacy CSS-variable-driven @keyframes had.
  //
  // The pulsed property is `filter`, and it plays on #outer, not on #inner,
  // for the same reason the alt-shadow elevation does (argued at length in
  // boardgame-component.ts's styles): a filter forces `transform-style: flat`
  // on the element that carries it, and #inner is what
  // motionTrackTarget('visual') returns -- the mount point for a
  // component-owned 3D scene. Nothing else about the throb changes: same
  // silhouette (#outer paints only #inner), same colors, same ungated
  // immediate infinite play, same start/cancel points.
  private _syncThrob(): void {
    this._throb?.cancel();
    this._throb = null;
    const outer = this.outerElement;
    if (outer) outer.style.filter = '';
    if (!this.active && !this.highlighted) return;
    if (!outer) return;
    const style = getComputedStyle(outer);
    const colorFrom = style.getPropertyValue('--throb-color-from').trim();
    const colorTo = style.getPropertyValue('--throb-color-to').trim();
    // Reduced motion: the highlight AFFORDANCE must survive even though
    // the pulse should not. The kernel would run the infinite play at
    // duration 0 (effectively suppressing the glow entirely, since with
    // fill 'none' nothing renders); the legacy shadow-scoped CSS ignored
    // the preference and kept pulsing — neither is right. Hold the strong
    // ('from') glow statically instead. (Phase 1 gate regression-critic
    // finding; declared in the token-throb evidence pack.)
    if (window.matchMedia('(prefers-reduced-motion: reduce)').matches) {
      outer.style.filter =
        `drop-shadow(0 0 0.25em ${colorFrom}) drop-shadow(0 0 0.25em ${colorFrom})`;
      return;
    }
    this._throb = this.play(outer, [
      { filter: `drop-shadow(0 0 0.25em ${colorTo}) drop-shadow(0 0 0.25em ${colorTo})` },
      // double the effect so it's darker
      { filter: `drop-shadow(0 0 0.25em ${colorFrom}) drop-shadow(0 0 0.25em ${colorFrom})` },
    ], {
      duration: 1000,
      easing: 'ease-in-out',
      direction: 'alternate',
      iterations: Infinity,
    }, { gated: false, timing: 'immediate' });
  }

  override connectedCallback(): void {
    super.connectedCallback();
    // Re-arm the ambient throb on (re)connect. The WAAPI throb -- unlike the
    // retired class-driven CSS @keyframes it replaced -- is cancelled in
    // disconnectedCallback and otherwise only (re)started when active or
    // highlighted CHANGE. Lit does not re-render on a reparent, so no
    // updated() fires and active/highlighted are unchanged: without this, a
    // still-highlighted token moved to a new container would lose its glow
    // forever. Safe on first connect -- outerElement is null pre-render, so
    // _syncThrob no-ops, and the first render's updated() starts it as before
    // (see the DOM-reparent test in token-throb.spec.ts).
    this._syncThrob();
  }

  override disconnectedCallback(): void {
    this._throb?.cancel();
    this._throb = null;
    super.disconnectedCallback();
  }

  /**
   * The last thing this host does before a stack orphans it -- which, because
   * `_insertNodes` pushes removed hosts straight into `_componentPool`, is also
   * the last thing it does before some OTHER component is rendered into it.
   *
   * Everything the scene is made of is rendered from the template and so is
   * re-derived for the next occupant. `#inner`'s own transform is not: it is
   * what `motionTrackTarget('visual')` returns, the animation kernel writes a
   * track's resting value there, and a value left behind would compose with the
   * next occupant's pose forever. A token compiles no visual tracks today, so
   * nothing writes it today -- and that is exactly the situation in which a
   * stale write is invisible until it is permanent. Measured precedent: a stale
   * write to a card's #inner transform was invisible for a whole flight and
   * became permanent the instant the animation ended, leaving the card stuck at
   * 45 degrees. One line here is cheaper than finding that again.
   */
  override beforeOrphaned(): void {
    if (this.innerElement) this.innerElement.style.transform = '';
    super.beforeOrphaned();
  }

  /**
   * The scene, as one element per polygon under one posed, sized carrier.
   *
   * Every value here is bound, so Lit rewrites all of them whenever the state
   * that produced them changes -- including on a recycled host, which is the
   * case that matters. Keyed by facet index so that switching a live token
   * between a 6-facet cube and a 14-facet prism reconciles rather than rebuilds.
   */
  private _renderSolid(solid: TokenSolid) {
    return html`
      <div id="solid"
        style="font-size:calc(var(--component-effective-width) * ${cssNumber(solid.fit)});transform:${solid.pose}">
        ${repeat(solid.facets, (facet) => facet.key, (facet) =>
          html`<div class="facet" style="${facet.style}"></div>`)}
      </div>
    `;
  }

  override render(): TemplateResult {
    const solid = this._solid();
    return html`
      <div id="outer" class="${classMap(this._computeClasses())}" @click="${(e: Event) => this.handleTap(e)}" style="${this._outerStyle}">
        <div id="inner">
          ${solid ? this._renderSolid(solid) : html`<img src="${this._computeAsset(this.type)}">`}
        </div>
      </div>
    `;
  }
}

declare global {
  interface HTMLElementTagNameMap {
    'boardgame-token': BoardgameToken;
  }
}
