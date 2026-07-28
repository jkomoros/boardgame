import { BoardgameComponent } from './boardgame-component.js';
import { html, css, CSSResult, TemplateResult, unsafeCSS } from 'lit';
import { customElement, property } from 'lit/decorators.js';
import { classMap } from 'lit/directives/class-map.js';
import { repeat } from 'lit/directives/repeat.js';
import { motionSilhouette } from '../motion/subject.js';
import type { MotionSubjectSnapshot } from '../motion/subject.js';
import { cssNumber } from '../solid/screen-frame.js';
import {
  ART_DEPTH,
  TOKEN_COLOR_FILTERS,
  artDrawnWidth,
  isTokenSolidShape,
  shadowOffsetEm,
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

/**
 * The depth treatment for the two shapes that keep their authored art, spelled
 * out from `ART_DEPTH` so that the numbers live next to the light they are
 * derived from rather than in a stylesheet nobody would think to check.
 *
 * Every offset here points along `SHADOW_DIRECTION`, which is `LIGHT`'s
 * in-plane part negated. Move the light in `token-solid.ts` and all of these
 * move with it, which is the only way "the same scene" survives an edit.
 */
const EDGE = shadowOffsetEm(ART_DEPTH.edgeEm);
const GROUND = shadowOffsetEm(ART_DEPTH.groundEm);
const ART_EDGE_FILTER =
  `drop-shadow(${EDGE.x.toFixed(4)}em ${EDGE.y.toFixed(4)}em 0 `
  + `rgba(0, 0, 0, ${ART_DEPTH.edgeAlpha}))`;

/** Every `type` a token accepts. Module-level so the rules below can be derived. */
const LEGAL_TYPES = Object.freeze([
  'token',
  'chip',
  'cube',
  'pawn',
  'meeple',
  'disc',
]);

/**
 * The types the depth treatment applies to: every legal type that is NOT a
 * solid — today `pawn` and `meeple`.
 *
 * Derived rather than listed, so that promoting a shape to a real solid stops
 * it being mirrored and leaned by deleting nothing: `SHAPES` in
 * `token-solid.ts` is the only list either half reads. A stale hardcoded pair
 * here would mirror a shape that had grown a mesh.
 */
const ART_SHAPES = LEGAL_TYPES.filter((type) => !isTokenSolidShape(type));

/** `#outer.meeple X, #outer.pawn X` for whatever `X` a rule needs. */
const artSelector = (suffix: string): string =>
  ART_SHAPES.map((type) => `#outer.${type} ${suffix}`).join(',\n      ');

/**
 * The contact shadow's width, per shape, because it is a fraction of the
 * PIECE's drawn width and not of the box.
 *
 * A meeple spans 0.90 of its square box and a pawn 0.43 — the box is square and
 * the SVG keeps its own proportions inside it. One shared width sized off the
 * box gave the pawn a shadow nearly twice as wide as the pawn, which reads as a
 * puddle rather than as the piece meeting the board.
 */
const GROUND_WIDTH_RULES = ART_SHAPES
  .map((type) => `#outer.${type} #inner::after { width: `
    + `${(artDrawnWidth(type) * ART_DEPTH.groundWidth).toFixed(4)}em; }`)
  .join('\n      ');

@customElement('boardgame-token')
export class BoardgameToken extends BoardgameComponent {
  static override styles: any = [
    BoardgameComponent.styles,
    css`
      #inner {
        height: var(--component-effective-height);
        width: var(--component-effective-width);
        /* The contact shadow below is positioned against this box. No offsets,
           so nothing moves; #inner's own transform still belongs entirely to
           the animation kernel and is never written here. */
        position: relative;
      }

      /*
       * DEPTH FOR THE TWO SHAPES THAT ARE NOT SOLIDS.
       *
       * meeple and pawn keep their authored SVG because a prism over a
       * non-convex outline paints its own back surface through its front and
       * CSS has no z-buffer to stop it -- measured at 2.8% of a meeple's
       * silhouette wrong at 75 degrees, and a comb-shaped control at 12.1%,
       * with explicit z-index sorting measured to make it WORSE rather than
       * better. So they are presented rather than modelled, and the whole of
       * the presentation is these three rules plus the mirror on the img.
       *
       * All of it is pure CSS keyed off the type's own class, which
       * _computeClasses derives from this.type on every render. Nothing is
       * written imperatively and nothing is remembered, so a pooled host that
       * arrives as a meeple having been a cube is a meeple -- the same
       * property the solid path is built around.
       *
       * #art exists so the edge shadow and the recolouring do not fight over
       * one filter slot: an #outer.COLOUR img rule is more specific than anything
       * that could be written for the img here, so an edge filter on the img
       * would be dropped for all nine non-red colours and survive only on red.
       * On the wrapper it composes -- the img is recoloured, then the wrapper's
       * drop-shadow is derived from the recoloured result's alpha.
       *
       * The font-size is what makes every length below scale with the token:
       * 1em IS the token's width, exactly as #solid's font-size is the solid's
       * size, so nothing has to be remeasured in JavaScript.
       */
      #art {
        position: relative;
        width: 100%;
        height: 100%;
        font-size: var(--component-effective-width);
      }

      #art img {
        height: 100%;
        width: 100%;
      }

      ${unsafeCSS(artSelector('#art'))} {
        filter: ${unsafeCSS(ART_EDGE_FILTER)};
        /* The tilt: a vertical scale about the FOOT, so the piece leans away
           rather than shrinking. Small on purpose -- fully reprojecting a
           standing piece to the scene's 50-degree camera would foreshorten it
           by 0.64 and lay it flat, which is the mesh work this design
           declines. A 2D scale, deliberately: a rotateX here would be a 3D
           transform, and a 3D transform is a composited layer. */
        transform: scaleY(${unsafeCSS(String(ART_DEPTH.lean))});
        transform-origin: 50% 100%;
      }

      /*
       * THE MIRROR, and it is the load-bearing half of this whole treatment.
       *
       * Both assets are lit from the upper RIGHT and every solid is lit from
       * the upper LEFT. Measured at 200px on white, mean luma per quadrant of
       * the drawn silhouette: a cube reads TL 43.0 / TR 42.1 and BL 34.0 /
       * BR 28.9 -- left brighter in both rows -- while a meeple reads TL 44.7 /
       * TR 51.9 and a pawn TL 36.2 / TR 53.6. A meeple and a pawn are
       * mirror-symmetric in silhouette, so this moves their light across
       * without touching their shape. It costs one declaration and it is the
       * difference between two art styles and one scene.
       */
      ${unsafeCSS(artSelector('#art img'))} {
        transform: scaleX(${unsafeCSS(ART_DEPTH.mirrored ? '-1' : '1')});
      }

      /*
       * The contact shadow: a soft ellipse under the piece's foot, offset along
       * the same SHADOW_DIRECTION every other shadow here uses.
       *
       * Only the standing shapes get one, and that is the point rather than an
       * omission: a disc lying on the board already meets it along a dark
       * bottom rim, and a pool of shadow under a lying piece would be a second
       * light source. The colour is the elevation shadows' own rgba(60,40,20),
       * so a token has one shadow palette rather than two.
       */
      ${unsafeCSS(artSelector('#inner::after'))} {
        content: '';
        position: absolute;
        left: 50%;
        bottom: 0;
        /* Its own font-size, so the offset below can be em and stay PARALLEL to
           the light. Expressed in percentages it would not be: a percentage
           translate resolves against the element's own box, and this box is
           0.78 wide by 0.15 tall, which would shear the direction by five to
           one. */
        font-size: var(--component-effective-width);
        /* Width is per shape; see GROUND_WIDTH_RULES below. */
        height: 0.15em;
        transform: translate(
          calc(-50% + ${unsafeCSS(`${GROUND.x.toFixed(4)}em`)}),
          ${unsafeCSS(`${GROUND.y.toFixed(4)}em`)});
        background: radial-gradient(closest-side,
          rgba(60, 40, 20, ${unsafeCSS(String(ART_DEPTH.groundAlpha))}),
          rgba(60, 40, 20, ${unsafeCSS(String(ART_DEPTH.groundAlpha * 0.35))}) 62%,
          rgba(60, 40, 20, 0) 100%);
        z-index: -1;
      }

      /*
       * THE BOX THE SOLID IS DRAWN IN -- and, deliberately, NO CAMERA.
       *
       * There used to be a perspective here and a preserve-3d on #inner,
       * with the camera reaching the facets from #outer the way
       * boardgame-card.ts's does. It is gone, and its absence is load-bearing:
       * Chromium hands EVERY element inside a live preserve-3d context its own
       * composited layer as soon as an ancestor transform starts animating, and
       * a stack's FLIP animates the component host on every single move.
       * Measured in pass, 55 tokens: 57 composited layers at rest, 1,047
       * during a move, 88.6 megapixels of layer area, 30fps against the flat
       * art's 59.6. Removing any two of {perspective, preserve-3d, the facets'
       * 3D transforms} still left ~1,000 layers; removing all three is what
       * restores 60fps.
       *
       * So the camera lives in JavaScript now. A token's pose is a constant, so
       * its projection is a constant: token-solid.ts applies the same
       * perspective divide once and emits flat, already-projected polygons. The
       * lens is unchanged -- still CAMERA_DEPTH_WIDTHS token widths, still
       * scale-invariant so that pass's two --component-scale values are drawn by
       * one lens -- it is just evaluated somewhere cheaper.
       *
       * The explicit box stays: #outer is a block that would otherwise stretch
       * to whatever cell the stack put it in, and #solid is sized off it.
       */
      #outer.solid {
        width: var(--component-effective-width);
        height: var(--component-effective-height);
      }

      /*
       * The solid's own element: the sizing, and nothing else. Its font-size IS
       * the solid's size, so every generated length below can be an em and the
       * whole thing follows --component-width with no JavaScript remeasurement.
       *
       * No transform. The pose is already in the polygons -- see
       * token-solid.ts's visibleFacetPolygons.
       */
      #solid {
        position: relative;
        width: 100%;
        height: 100%;
      }

      /*
       * One element per VISIBLE surface polygon, drawn as its own projected
       * outline. A back-facing facet is not hidden here, it was never built:
       * backface-visibility: hidden used to be the whole of the hidden-surface
       * removal, and the same fact that made it sufficient -- every shape that
       * renders as a solid is convex, so its camera-facing facets tile the
       * silhouette exactly once -- lets token-solid.ts cull them instead. That
       * is also why nothing here sorts: what is left cannot overlap.
       *
       * NO will-change, and now nothing that could provoke a promotion either.
       * The die promotes every facet because a facet that crosses from
       * back-facing to front-facing DURING a tumble would otherwise stay culled
       * for a frame or two and tear a hole in the solid. A token does not tumble.
       *
       * There is deliberately no .facet rule at all: position, box, margin,
       * clip-path and background are every one of them per facet, and they all
       * come from src/solid/flat-facets.ts.
       */

      /*
       * THERE ARE DELIBERATELY NO PER-SHAPE ASPECT RATIOS HERE.
       *
       * A rule setting --component-aspect-ratio to 2.0 on #outer.pawn, and one
       * setting it to 1.25 on #outer.meeple, used to sit here -- and neither
       * had ever once applied. A custom property that REFERENCES another is
       * substituted where it is DECLARED, and --component-effective-height is
       * declared at :host (boardgame-component.ts), above #outer. So it was
       * always substituted with the :host ratio of 1.0, whatever #outer said.
       * Measured: a pawn computed --component-aspect-ratio: 2.0 on #outer
       * while --component-effective-height computed to
       * calc(calc(1.0 * 30px) * 1.0), and every shape drew in a 30x30 box.
       *
       * They were deleted rather than made to work, for three measured
       * reasons.
       *
       * THE BOX IS THE LAYOUT CONTRACT, AND EVERY CONSUMER ASSUMES IT SQUARE.
       * The board layout puts aspect-ratio: 1 on every component host,
       * boardgame-spatial-board's tokenPosition centres a piece at
       * coords - tokenSize / 2 in BOTH axes, and the stack's spread/fan
       * margins and the FLIP scale ratio all key off the same box. A
       * stack-hosted component cannot reserve extra space -- the same rule
       * that makes a 3D token size by drawn extent rather than by
       * circumsphere.
       *
       * THE ART IS ALREADY IN TRUE PROPORTION. An SVG's default
       * preserveAspectRatio is xMidYMid meet, so token_pawn.svg (89.536 by
       * 207.215) draws at its own 0.432 inside whatever box it is handed. The
       * rules would not have removed that letterbox, only resized it.
       *
       * AND THE NUMBERS WERE WRONG ANYWAY: 2.0 against the pawn asset's 2.31,
       * 1.25 against the meeple's 1.11. Honouring 2.0 was rendered and looked
       * at -- at a 120px component width it draws a pawn 240px tall next to a
       * 120px cube, which is the opposite of "pieces from the same set".
       *
       * --component-aspect-ratio still works where it is meant to be set: at
       * :host or above, which is where --component-effective-height can see
       * it. What cannot work is setting it from inside this shadow tree.
       * tests/animations/parity/token-box.spec.ts pins the invariant both
       * ways: a shape may not declare a ratio its box does not have.
       */

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

      ${unsafeCSS(GROUND_WIDTH_RULES)}

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
    return [...LEGAL_TYPES];
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
        style="font-size:calc(var(--component-effective-width) * ${cssNumber(solid.fit)})">
        ${repeat(solid.facets, (facet) => facet.key, (facet) =>
          html`<div class="facet" style="${facet.style}"></div>`)}
      </div>
    `;
  }

  /**
   * The authored art, under the one wrapper the depth treatment needs.
   *
   * `#art` is not decoration: the edge shadow and the per-colour recolouring
   * would otherwise compete for the img's single `filter` slot, and
   * `#outer.<color> img` outranks anything the img could be given here — so the
   * edge would survive on red and vanish on the other nine. On the wrapper the
   * two compose in the right order: the img is recoloured, and the wrapper's
   * drop-shadow is derived from the recoloured result.
   *
   * Every shape still reaches this when it is a `spacer`, which has no item to
   * stand for and is `visibility: hidden`. That is deliberate and unchanged:
   * the img is what holds the slot's box open.
   */
  private _renderArt(): TemplateResult {
    return html`
      <div id="art"><img src="${this._computeAsset(this.type)}"></div>
    `;
  }

  override render(): TemplateResult {
    const solid = this._solid();
    return html`
      <div id="outer" class="${classMap(this._computeClasses())}" @click="${(e: Event) => this.handleTap(e)}" style="${this._outerStyle}">
        <div id="inner">
          ${solid ? this._renderSolid(solid) : this._renderArt()}
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
