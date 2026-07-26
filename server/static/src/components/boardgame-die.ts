import { BoardgameAnimatableItem } from './boardgame-animatable-item.js';
import { html, css } from 'lit';
import { property } from 'lit/decorators.js';
import { query } from 'lit/decorators.js';
import { repeat } from 'lit/directives/repeat.js';
import { isBoundMoveAction, type BoundMoveAction } from '../moves/action.js';
import { componentMotionTracks } from '../motion/component-track.js';
import type { ComponentMotionTarget } from '../motion/component-track.js';
import {
  cross,
  dieGeometry,
  dot,
  magnitude,
  normalize,
  scale as scaleVec,
  subtract,
  vec3,
  type DieFace,
  type DieGeometry,
  type Vec3,
} from '../motion/die-geometry.js';

/**
 * Drawing a die as a solid.
 *
 * `die-geometry.ts` hands over a closed surface — `[...faces, ...capFaces]` —
 * where every polygon carries its own outward `normal`, its `centroid` and its
 * vertex loop. Everything below turns ONE such polygon into ONE absolutely
 * positioned element, and it is deliberately a single routine: a d6's
 * rectangles, a d20's triangles, a d12's pentagons, a d10's kites and a
 * barrel's non-square side rectangles plus cap triangles all go through it
 * unchanged. Nothing here may assume a square facet — a d7's side face is
 * 2.37 long by 0.87 wide.
 *
 * Coordinate frames. The geometry is in the physics frame (+Y up, right
 * handed). CSS screen space is x right, y DOWN, z toward the viewer. The two
 * are related by a 180-degree rotation about X, `(x, y, z) -> (x, -y, -z)`,
 * which is a PROPER rotation: simply flipping Y would mirror the solid, and a
 * mirrored facet renders its glyphs backwards (task 9 paints those).
 *
 * Units. Lengths are emitted in `em`, and `#stage` sets `font-size:
 * var(--effective-die-size)`, so `1em` is the die's size and the whole solid
 * scales with the custom property with no JavaScript remeasurement. That is
 * the only reason a caller can set `--die-size` to anything (`120px`, `6rem`,
 * `10vmin`) and have the solid follow. Geometry units are scaled by
 * `0.5 / circumradius` so the die's bounding sphere is exactly `1em` across
 * whatever its face count — `die-geometry.ts` builds each solid at its own
 * natural scale (circumradius 1.000 for a d8, 1.902 for a d20) and documents
 * that consumers must normalize themselves.
 */

/** Screen up in CSS space: CSS y points down. */
const SCREEN_UP: Vec3 = vec3(0, -1, 0);

/**
 * Where the presented face is pointed, in CSS space, when the die is at rest.
 *
 * Not straight at the camera (`+Z`): a solid facing the viewer square-on
 * projects to a flat outline and reads as the 2D die this replaces. Pointing
 * the presented face slightly down and to the left instead puts the camera
 * above and to the right of it, so a d6 shows its presented face plus the
 * faces above and to its right — a die seen on a table. The physics-driven
 * resting pose replaces this when the roll is wired up.
 */
const RESTING_VIEW: Vec3 = normalize(vec3(-0.32, 0.26, 1));

/** Pip diameter as a fraction of the facet's shorter side (was 7px on 50px). */
const PIP_FRACTION = 0.14;

/** Numeral height as a fraction of the facet's shorter side. */
const GLYPH_FRACTION = 0.42;

/** Body frame (+Y up) to CSS frame (+Y down): a 180-degree turn about X. */
function toScreen(v: Vec3): Vec3 {
  return vec3(v[0], -v[1], -v[2]);
}

/** Short, stable decimal text: keeps generated style strings readable. */
function num(value: number): string {
  const rounded = Number(value.toFixed(5));
  return Object.is(rounded, -0) ? '0' : String(rounded);
}

/**
 * An orthonormal, right-handed basis `(u, v, w)` for a facet's own plane, with
 * `w` the outward normal. `u` is the facet's local +x (screen right when the
 * facet faces the camera) and `v` its local +y (screen down), so the CSS box
 * lands the same way up as the screen wherever the facet can be read.
 *
 * `det([u v w]) = u . ((w x u) x w) = 1`, so this is a rotation and never a
 * reflection, whichever branch is taken.
 */
function facetBasis(w: Vec3): { u: Vec3; v: Vec3 } {
  let u = cross(w, SCREEN_UP);
  // Degenerate exactly for the facets that point straight up or straight
  // down the screen (a d6's top and bottom): any perpendicular will do.
  if (magnitude(u) < 1e-6) u = cross(w, vec3(0, 0, 1));
  const unitU = normalize(u);
  return { u: unitU, v: cross(w, unitU) };
}

/** One facet of the solid, as the CSS needed to draw it. */
interface DieFacet {
  /** Stable key: index into `[...faces, ...capFaces]`. */
  readonly key: number;
  /** Index into `faces` (and so into the `faces` property), or -1 for a cap. */
  readonly faceIndex: number;
  readonly style: string;
}

/**
 * Place one surface polygon.
 *
 * The element is a plain box whose centre sits at the solid's centre (`left`/
 * `top` at 50% less half its size), so `transform-origin` — its own centre —
 * is the solid's origin. `translate3d(...) matrix3d(...)` then rotates the box
 * into the facet's plane about that origin and moves it out to the facet, the
 * translation being read in the PARENT's frame because CSS applies a transform
 * list left to right.
 *
 * The box is the facet's own bounding rectangle in its own plane — NOT a
 * square, and not centred on the polygon's centroid either (a triangle's
 * vertex mean is not the centre of its bounding box), so the translation
 * carries the bounding-box offset as well. `clip-path` then cuts the box down
 * to the actual polygon, which is what makes triangles, kites, pentagons and
 * rectangles one code path.
 */
function facetStyle(face: DieFace, unitsToEm: number): string {
  const w = normalize(toScreen(face.normal));
  const { u, v } = facetBasis(w);
  const centre = scaleVec(toScreen(face.centroid), unitsToEm);
  const points = face.polygon.map((point) => {
    const offset = subtract(scaleVec(toScreen(point), unitsToEm), centre);
    return { a: dot(offset, u), b: dot(offset, v) };
  });
  const minA = Math.min(...points.map((p) => p.a));
  const maxA = Math.max(...points.map((p) => p.a));
  const minB = Math.min(...points.map((p) => p.b));
  const maxB = Math.max(...points.map((p) => p.b));
  const width = maxA - minA;
  const height = maxB - minB;
  if (!(width > 0) || !(height > 0)) {
    throw new Error(`degenerate facet: ${width} x ${height} bounding box`);
  }
  // The box centre in the facet's plane, then out into the parent's frame.
  const boxA = (minA + maxA) / 2;
  const boxB = (minB + maxB) / 2;
  const t = [0, 1, 2].map((i) => centre[i] + u[i] * boxA + v[i] * boxB);
  const clip = points
    .map((p) => `${num(((p.a - minA) / width) * 100)}% ${num(((p.b - minB) / height) * 100)}%`)
    .join(', ');
  const shorter = Math.min(width, height);
  return [
    `width:${num(width)}em`,
    `height:${num(height)}em`,
    `margin-left:${num(-width / 2)}em`,
    `margin-top:${num(-height / 2)}em`,
    `transform:translate3d(${num(t[0])}em,${num(t[1])}em,${num(t[2])}em) `
      + `matrix3d(${[u, v, w].map((axis) => `${num(axis[0])},${num(axis[1])},${num(axis[2])},0`).join(',')},0,0,0,1)`,
    `clip-path:polygon(${clip})`,
    // Both are `em` against the facet's INHERITED font-size (the die's size,
    // set once on `#stage`). Nothing may set `font-size` on the facet itself:
    // its own `width`/`height` are in `em` too, and those resolve against the
    // element's own font-size, so a glyph size set here would resize the
    // facet. The numeral therefore sizes a child span instead.
    `--pip-size:${num(shorter * PIP_FRACTION)}em`,
    `--glyph-size:${num(shorter * GLYPH_FRACTION)}em`,
  ].join(';');
}

/**
 * The rotation that points `face`'s normal at `RESTING_VIEW`, as a CSS
 * `rotate3d`. The minimal rotation between two directions: axis `n x view`,
 * angle `atan2(|n x view|, n . view)`. CSS `rotate3d` is the right-handed
 * Rodrigues rotation in the same coordinate triple, so no sign fixing.
 */
function presentationTransform(face: DieFace): string {
  const n = normalize(toScreen(face.normal));
  const axis = cross(n, RESTING_VIEW);
  const sine = magnitude(axis);
  const cosine = dot(n, RESTING_VIEW);
  if (sine < 1e-9) {
    // Already there, or pointing exactly backwards: a half turn about any
    // perpendicular then does it, and `facetBasis` names one.
    if (cosine > 0) return 'none';
    const { u } = facetBasis(n);
    return `rotate3d(${num(u[0])},${num(u[1])},${num(u[2])},180deg)`;
  }
  const degrees = (Math.atan2(sine, cosine) * 180) / Math.PI;
  const unit = scaleVec(axis, 1 / sine);
  return `rotate3d(${num(unit[0])},${num(unit[1])},${num(unit[2])},${num(degrees)}deg)`;
}

/** The full facet list for a face count, plus the geometry it came from. */
interface DieSolid {
  readonly geometry: DieGeometry;
  readonly facets: readonly DieFacet[];
}

// Building a solid runs a convex hull for the closed-form shapes, so it is
// cached per face count. `null` records a face count that has no solid, so a
// malformed die does not retry the failure on every render pass.
const SOLID_CACHE = new Map<number, DieSolid | null>();

function dieSolid(faceCount: number): DieSolid | null {
  const cached = SOLID_CACHE.get(faceCount);
  if (cached !== undefined) return cached;
  let solid: DieSolid | null = null;
  try {
    const geometry = dieGeometry(faceCount);
    const unitsToEm = 0.5 / geometry.circumradius;
    const surface = [...geometry.faces, ...geometry.capFaces];
    solid = {
      geometry,
      facets: surface.map((face, key) => ({
        key,
        faceIndex: key < geometry.faces.length ? key : -1,
        style: facetStyle(face, unitsToEm),
      })),
    };
  } catch {
    // A face count with no solid (fewer than 3 faces, or a shape the geometry
    // module rejects) falls back to the reel rather than throwing mid-render.
    solid = null;
  }
  SOLID_CACHE.set(faceCount, solid);
  return solid;
}

class BoardgameDie extends BoardgameAnimatableItem {
  static override styles = [
    ...(BoardgameAnimatableItem.styles ? [BoardgameAnimatableItem.styles] : []),
    css`
      :host {
        --effective-die-scale: var(--die-scale, 1.0);
        /*
         * --die-size is the die's overall size, and is the property a caller
         * sets: any CSS length ('120px', '6rem', '10vmin'). It is the side of
         * the square box the die is laid out in AND the diameter of the
         * sphere the solid is inscribed in, so a die of any face count fits
         * its box in every orientation -- which is what lets a later task
         * tumble it without it escaping the layout.
         *
         * --effective-die-size is the resolved value everything in here
         * measures against; it is not part of the component's API.
         */
        --effective-die-size: var(--die-size, 50px);
        --pip-size: 7px;
      }

      #scaler {
        height: calc(var(--effective-die-size) * var(--effective-die-scale));
        width: calc(var(--effective-die-size) * var(--effective-die-scale));
        position: relative;
        display: flex;
        flex-direction: column;
        align-items: center;
        justify-content: center;
      }

      #main.disabled {
        cursor: default;
      }

      #main {
        height: var(--effective-die-size);
        width: var(--effective-die-size);
        border-radius: 6px;
        background: linear-gradient(135deg, #F5F0E8 0%, #E0D9CE 100%);
        overflow: hidden;
        cursor: pointer;
        box-shadow: 0 2px 4px 0 rgba(60, 40, 20, 0.18),
                    0 1px 5px 0 rgba(60, 40, 20, 0.12),
                    0 3px 1px -2px rgba(60, 40, 20, 0.2),
                    inset 0 1px 0 rgba(255, 255, 255, 0.4);
        transform: scale(var(--effective-die-scale));
        transition: box-shadow 0.28s cubic-bezier(0.4, 0, 0.2, 1);
        border: 0;
        padding: 0;
      }

      #main.interactive:hover {
        box-shadow: 0 8px 10px 1px rgba(60, 40, 20, 0.14),
                    0 3px 14px 2px rgba(60, 40, 20, 0.12),
                    0 5px 5px -3px rgba(60, 40, 20, 0.4);
      }

      /*
       * In solid mode the die's body is the facets themselves, so #main is
       * only the hit target and the 3D scene's positioning context. It must
       * give up overflow:hidden (which would slice the solid, and which
       * flattens any 3D context put on it) along with the flat card look.
       */
      #main.solid,
      #main.solid.interactive:hover {
        position: relative;
        overflow: visible;
        background: none;
        border-radius: 0;
        box-shadow: none;
      }

      #action-status {
        position: absolute;
        top: 100%;
        width: max-content;
        max-width: 16rem;
        margin-top: 0.25rem;
        color: var(--md-sys-color-error, #ba1a1a);
        font-size: 0.75rem;
      }

      /*
       * The solid: a perspective wrapper (#stage), the preserve-3d carrier
       * (#inner -- the element motionTrackTarget('visual') returns, so a later
       * task animates the tumble on it), a resting-pose carrier (#orient) and
       * one element per surface polygon.
       *
       * #orient exists because #inner's transform belongs to the animation
       * kernel: a resting pose written onto #inner would be replaced outright
       * the moment a spin plays (play() pins composite:'replace'), snapping
       * the solid flat mid-roll.
       *
       * Nothing from #stage down may carry a grouping property -- overflow
       * other than visible, a filter, opacity < 1, clip-path, mask -- because
       * each of those forces transform-style back to flat and collapses the
       * solid into a pile of overlapping outlines. That is why #main.solid
       * gives up the overflow:hidden the reel needs. The facets themselves
       * DO carry clip-path, which is fine: they are leaves of the 3D context.
       */
      #stage {
        position: absolute;
        inset: 0;
        /*
         * The one place the die's size becomes a font-size, so that every
         * generated length below can be an em unit and the whole solid follows
         * --die-size with no JavaScript remeasurement.
         */
        font-size: var(--effective-die-size);
        perspective: 6em;
        perspective-origin: 50% 50%;
      }

      #inner {
        position: relative;
        transform: translateY(calc(-1 * var(--effective-die-size) * var(--selected-face)));
        /* The spin is WAAPI-driven now; no CSS transform transition. */
      }

      #inner.solid {
        position: absolute;
        inset: 0;
        transform-style: preserve-3d;
        /*
         * There is no reel to scroll, so the reel step is zero. The rule above
         * and the WAAPI spin keyframes (_innerTransformForFace) both read
         * --effective-die-size, so zeroing it here makes the face-change spin
         * a no-op on the solid without touching either -- the roll is a real
         * tumble in a later task, and until then the die must not slide.
         */
        --effective-die-size: 0px;
      }

      #orient {
        position: absolute;
        inset: 0;
        transform-style: preserve-3d;
      }

      .facet {
        position: absolute;
        left: 50%;
        top: 50%;
        /* width/height/margin/transform/clip-path are generated per facet. */
        backface-visibility: hidden;
        background:
          linear-gradient(135deg, #F5F0E8 0%, #E0D9CE 100%);
        box-shadow: inset 0 0 0 1px rgba(60, 40, 20, 0.14);
        transition: background 0.2s ease-out;
      }

      #main.solid.interactive:hover .facet {
        background:
          linear-gradient(135deg, #FFFDF8 0%, #EFE9DE 100%);
      }

      .facet > span {
        font-size: var(--glyph-size, 1em);
        line-height: 1;
      }

      .face {
        display: flex;
        flex-direction: column;
        align-items: center;
        justify-content: center;
        font-family: var(--md-sys-typescale-body-large-font, 'Source Sans 3', sans-serif);
        font-weight: 500;
      }

      #inner.reel .face {
        height: var(--effective-die-size);
        width: var(--effective-die-size);
        position: relative;
        font-size: 20px;
        line-height: 28px;
      }

      .pip {
        background-color: var(--md-sys-color-on-surface, #1C1810);
        height: var(--pip-size);
        width: var(--pip-size);
        border-radius: calc(var(--pip-size) / 2);
        position: absolute;
        display: none;
        box-shadow: inset 0 1px 2px rgba(0, 0, 0, 0.4),
                    0 1px 0 rgba(255, 255, 255, 0.2);
      }

      .face.one span, .face.two span, .face.three span, .face.four span, .face.five span, .face.six span {
        display: none;
      }

      /*
       * Pips are placed relative to THEIR OWN FACE (50% of it), not to the
       * die's size. In the reel a face is exactly one die-size square, so this
       * is bit-for-bit the old placement; on a solid a facet is whatever shape
       * the geometry says, and a d6's facet is smaller than the box the die
       * is laid out in. Task 9 replaces this fixed one-to-six layout with a
       * computed one; until then --pip-size is scaled per facet.
       */
      .pip.mid {
        top: calc(50% - var(--pip-size) / 2);
      }

      .pip.center {
        left: calc(50% - var(--pip-size) / 2);
      }

      .pip.top {
        top: calc(50% - var(--pip-size) * 1.5 - var(--pip-size) / 2);
      }

      .pip.left {
        left: calc(50% - var(--pip-size) * 1.5 - var(--pip-size) / 2);
      }

      .pip.bottom {
        top: calc(50% + var(--pip-size) * 1.5 - var(--pip-size) / 2);
      }

      .pip.right {
        left: calc(50% + var(--pip-size) * 1.5 - var(--pip-size) / 2);
      }

      .face.one .pip.mid.center {
        display: block;
      }

      .face.two .pip.top.right, .face.two .pip.bottom.left {
        display: block;
      }

      .face.three .pip.top.right, .face.three .pip.mid.center, .face.three .pip.bottom.left {
        display: block;
      }

      .face.four .pip.top.right, .face.four .pip.top.left, .face.four .pip.bottom.left, .face.four .pip.bottom.right {
        display: block;
      }

      .face.five .pip.top.right, .face.five .pip.top.left, .face.five .pip.bottom.left, .face.five .pip.bottom.right, .face.five .pip.mid.center {
        display: block;
      }

      .face.six .pip.top.right, .face.six .pip.top.left, .face.six .pip.bottom.left, .face.six .pip.bottom.right, .face.six .pip.mid.left, .face.six .pip.mid.right {
        display: block;
      }
    `
  ];

  @property({ type: Object })
  item: any = null;

  @property({ type: Number })
  value = 0;

  @property({ type: Array })
  faces: number[] = [];

  @property({ type: Number })
  selectedFace = 0;

  @property({ type: Boolean })
  disabled = false;

  @property({ attribute: false })
  action: BoundMoveAction<string, object> | null = null;

  @query('#inner')
  private _innerElement?: HTMLElement;

  private _boundHandleClick?: (e: Event) => void;
  private _unsubscribeAction: (() => void) | null = null;

  override connectedCallback() {
    super.connectedCallback();
    this._boundHandleClick ??= (e: Event) => this._handleClick(e);
    this.renderRoot.addEventListener('click', this._boundHandleClick);
    this._subscribeAction();
  }

  override disconnectedCallback() {
    if (this._boundHandleClick) {
      this.renderRoot.removeEventListener('click', this._boundHandleClick);
    }
    this._unsubscribeAction?.();
    this._unsubscribeAction = null;
    super.disconnectedCallback();
  }

  override updated(changedProperties: Map<PropertyKey, unknown>) {
    super.updated(changedProperties);

    if (changedProperties.has('selectedFace')) {
      this._selectedFaceChanged(
        this.selectedFace,
        changedProperties.get('selectedFace') as number | undefined
      );
    }

    if (changedProperties.has('item')) {
      this._itemChanged(this.item);
    }

    if (changedProperties.has('action')) {
      this._subscribeAction();
    }
  }

  private _handleClick(e: Event) {
    if (!isBoundMoveAction(this.action)) {
      if (this.action !== null) e.stopPropagation();
      return;
    }
    if (this.disabled || !this.action.canActivate) {
      e.stopPropagation();
      return;
    }
    e.stopPropagation();
    void this.action.activate();
  }

  private _subscribeAction(): void {
    this._unsubscribeAction?.();
    this._unsubscribeAction = isBoundMoveAction(this.action)
      ? this.action.subscribe(() => this.requestUpdate())
      : null;
  }

  // _innerTransformForFace mirrors the CSS resting transform on #inner for
  // a given selectedFace: translateY(-1 * effective-die-size * face). The
  // #main element carries the --selected-face var that drives the CSS
  // resting position; here we build an explicit transform so WAAPI can
  // interpolate the spin instead of relying on a CSS transition.
  private _innerTransformForFace(face: number): string {
    return `translateY(calc(-1 * var(--effective-die-size) * ${face}))`;
  }

  protected override motionTrackTarget(target: ComponentMotionTarget): HTMLElement | null {
    return target === 'host' ? this : this._innerElement ?? null;
  }

  private _selectedFaceChanged(newValue: number, oldValue: number | undefined) {
    if (!this._innerElement) return;
    // On first render there's no meaningful spin to animate from.
    if (oldValue === undefined || oldValue === newValue) return;
    this.playMotionTracks(componentMotionTracks([{
      target: 'visual',
      property: 'transform',
      from: this._innerTransformForFace(oldValue),
      to: this._innerTransformForFace(newValue),
    }]));
  }

  private _itemChanged(newValue: any) {
    if (!newValue) {
      this.faces = [];
      this.selectedFace = 0;
      this.value = 0;
      return;
    }
    this.faces = newValue.Values.Faces;
    this.selectedFace = newValue.DynamicValues.SelectedFace;
    this.value = newValue.DynamicValues.Value;
  }

  private _classForFace(face: number): string {
    let str = '';
    switch (face) {
      case 1:
        str = 'one';
        break;
      case 2:
        str = 'two';
        break;
      case 3:
        str = 'three';
        break;
      case 4:
        str = 'four';
        break;
      case 5:
        str = 'five';
        break;
      case 6:
        str = 'six';
        break;
    }

    return 'face ' + str;
  }

  private _classes(disabled: boolean, solid: boolean): string {
    const pieces: string[] = [];
    pieces.push(disabled ? 'disabled' : 'interactive');
    pieces.push(solid ? 'solid' : 'reel');
    return pieces.join(' ');
  }

  /**
   * The solid this die should be drawn as, or `null` when it has none: fewer
   * than three faces, or a face list the geometry module refuses. Those fall
   * back to the reel rather than throwing during a render pass.
   */
  private _solid(): DieSolid | null {
    const faces = this.faces;
    if (!Array.isArray(faces) || faces.length < 3) return null;
    if (!faces.every((face) => Number.isFinite(face))) return null;
    return dieSolid(faces.length);
  }

  /**
   * Which FACE the die presents, as an index into `faces`.
   *
   * `selectedFace` is an index (the server sends `DynamicValues.SelectedFace`
   * alongside a separate `Values.Faces` list of face VALUES); reading it as a
   * value is the silent bug this component invites. Out-of-range values fall
   * back to the first face rather than rendering nothing.
   */
  private _presentedFaceIndex(faceCount: number): number {
    const index = Math.trunc(this.selectedFace);
    return Number.isFinite(index) && index >= 0 && index < faceCount ? index : 0;
  }

  // The face's own content. Values one to six draw pips (the .face.one ...
  // .face.six CSS below hides the numeral for exactly those); anything else
  // shows the numeral. Task 9 replaces this with a computed pip layout that
  // is not capped at six.
  private _faceContent(value: number) {
    return html`
      <span>${value}</span>
      <div class="pip mid center"></div>
      <div class="pip top left"></div>
      <div class="pip top right"></div>
      <div class="pip bottom left"></div>
      <div class="pip bottom right"></div>
      <div class="pip mid left"></div>
      <div class="pip mid right"></div>
    `;
  }

  // The degenerate fallback: the pre-3D vertical reel of flat faces, scrolled
  // to the selected one by #inner's translateY.
  private _renderReel() {
    return html`
      <div id="inner" class="reel">
        ${repeat(this.faces, (face) => face, (face) => html`
          <div class="${this._classForFace(face)}">${this._faceContent(face)}</div>
        `)}
      </div>
    `;
  }

  private _renderSolid(solid: DieSolid) {
    const presented = this._presentedFaceIndex(solid.geometry.faceCount);
    const orient = presentationTransform(solid.geometry.faces[presented]);
    return html`
      <div id="stage">
        <div id="inner" class="solid">
          <div id="orient" style="transform:${orient}">
            ${repeat(solid.facets, (facet) => facet.key, (facet) => facet.faceIndex < 0
              ? html`<div class="facet cap" style="${facet.style}"></div>`
              : html`<div
                    class="facet ${this._classForFace(this.faces[facet.faceIndex])}"
                    style="${facet.style}"
                    data-face-index="${facet.faceIndex}"
                    data-face-value="${this.faces[facet.faceIndex]}"
                  >${this._faceContent(this.faces[facet.faceIndex])}</div>`)}
          </div>
        </div>
      </div>
    `;
  }

  override render() {
    const action = this.action;
    const bound = isBoundMoveAction(action);
    const interactive = bound;
    const effectiveDisabled = this.disabled || !interactive || (bound && !action.canActivate);
    const baseReason = bound
      ? action.reason?.message
      : action ? 'Bind required move input with .with(...)' : null;
    const reason = bound && action.preview.kind === 'failed' && action.preview.retryable
      ? `${baseReason ?? 'Move legality check failed'}. Activate to retry.`
      : baseReason;
    const solid = this._solid();
    return html`
      <div id="scaler">
        <button
          id="main"
          type="button"
          aria-label=${interactive ? 'Roll die' : 'Die'}
          aria-describedby=${reason ? 'action-status' : ''}
          aria-busy=${String(bound && action.submission.kind === 'pending')}
          ?disabled=${effectiveDisabled}
          style="--selected-face:${this.selectedFace}"
          class="${this._classes(effectiveDisabled, solid !== null)}">
          ${solid ? this._renderSolid(solid) : this._renderReel()}
        </button>
        ${reason ? html`<span id="action-status" role="status">${reason}</span>` : ''}
      </div>
    `;
  }
}

customElements.define('boardgame-die', BoardgameDie);
