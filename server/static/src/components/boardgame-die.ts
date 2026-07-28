import {
  BoardgameAnimatableItem,
  type MotionTrackPlayResult,
} from './boardgame-animatable-item.js';
import { html, css, nothing } from 'lit';
import { property } from 'lit/decorators.js';
import { query } from 'lit/decorators.js';
import { repeat } from 'lit/directives/repeat.js';
import {
  isBoundMoveAction,
  moveActionReasonSeverity,
  type BoundMoveAction,
} from '../moves/action.js';
import { componentMotionTracks } from '../motion/component-track.js';
import type { ComponentMotionTarget } from '../motion/component-track.js';
import { dieGeometry, dot, type DieGeometry, type Vec3 } from '../motion/die-geometry.js';
import { assignFaceValues, presentedFaceIndex, resolveReadingRule } from '../motion/die-faces.js';
import {
  FRAME_MS,
  PERSPECTIVE_DEPTH_DIE_SIZES,
  dieRollTrajectory,
  rollScene,
  settledTrajectory,
} from '../motion/dice-roll.js';
import type { Component } from '../types/boardgame-types.js';
import { cssNumber as num } from '../solid/screen-frame.js';
import { solidFacets, type SolidFacet } from '../solid/facet-placement.js';
import { readingPoseTransform } from '../solid/reading-pose.js';
import {
  CORNER_GLYPH_HEIGHT,
  GLYPH_HEIGHT,
  MIN_LEGIBLE_GLYPH_PX,
  MIN_LEGIBLE_PIP_PX,
  PIP_DIAMETER,
  glyphScale,
  isPipValue,
  pipCells,
  type PipCell,
} from './die-face-marks.js';

/**
 * A die, drawn as a solid.
 *
 * This file is the DIE, and only the die. The two halves it is built on are
 * modules of their own, and the split is the one a second 3D component depends
 * on:
 *
 *   - `src/solid/` renders any closed convex surface — it places each polygon as
 *     a `clip-path`ed box, derives the content square that fits inside it, and
 *     computes the pose the solid is read in. It knows nothing about dice, and a
 *     3D `boardgame-token` is meant to import exactly these three modules.
 *   - `motion/dice-*` is the physics: `dice-sim.ts` throws the die, `dice-roll.ts`
 *     decides which throw this die is making and frames it, `dice-bake.ts` turns
 *     the resulting poses into compositable CSS.
 *
 * What is left here is what only a die has: which face carries which VALUE, pips
 * versus numerals (`die-face-marks.ts`), corner printing for a solid read from a
 * face nobody can see, the reel fallback for anything with fewer than three
 * faces, and the reactive glue that notices a throw and plays it.
 *
 * Units. The solid's lengths are `em`, and `#stage` sets `font-size:
 * var(--effective-die-size)`, so `1em` is the die's size and the whole solid
 * scales with the custom property with no JavaScript remeasurement. That is the
 * only reason a caller can set `--die-size` to anything (`120px`, `6rem`,
 * `10vmin`) and have the solid follow.
 */

/** The face VALUES a die's deck publishes. */
interface DieValues {
  readonly Faces: readonly number[];
}

/**
 * What the server says about THIS die right now.
 *
 * `SelectedFace` is an INDEX into `Values.Faces`; `Value` is the face VALUE it
 * selects, and `RollCount` is `components/dice`'s throw counter. All three are
 * optional here because a game is free to define its own die deck with fewer of
 * them: everything downstream treats a missing one as "not reported" rather
 * than as zero.
 */
interface DieDynamicValues {
  readonly SelectedFace?: number;
  readonly Value?: number;
  readonly RollCount?: number;
}

/**
 * A die's component, INCLUDING the sanitized form.
 *
 * A die in a hidden stack arrives as `{}` — a truthy object with no `Values` —
 * so every read of the item has to go through `isVisibleComponent` first. That
 * is what `Component` (as opposed to `VisibleComponent`) says in the type
 * system, and it is why this property is not `any`: bound to a card, or to a
 * component whose deck has no faces, this now fails to compile instead of
 * throwing out of the render pass.
 */
export type DieComponent = Component<DieValues, DieDynamicValues>;

/** What one face draws, and what it announces: see `_resolveFace`. */
interface ResolvedFace {
  readonly kind: 'symbol' | 'pips' | 'numeral';
  readonly text: string;
  readonly cells: readonly PipCell[];
  readonly label: string;
}

/**
 * Which face a die is READ from when `up` is its topmost direction, for a
 * solid that is not read from an up face at all — the one it is resting on,
 * i.e. the one whose normal is most opposed to `up`. `die-faces.ts` uses the
 * same rule for both of its non-`'up-face'` conventions, so this covers a d4
 * (read from the apex vertex) and an odd-sided barrel (read from the edge
 * pointing at the ceiling) with no case analysis.
 */
function faceReadFrom(geometry: DieGeometry, up: Vec3): number {
  let best = 0;
  let bestScore = Infinity;
  for (let index = 0; index < geometry.faces.length; index++) {
    const score = dot(geometry.faces[index].normal, up);
    if (score < bestScore) {
      bestScore = score;
      best = index;
    }
  }
  return best;
}

/** The full facet list for a face count, plus the geometry it came from. */
interface DieSolid {
  readonly geometry: DieGeometry;
  readonly facets: readonly SolidFacet[];
  /**
   * True when this solid presents the face it RESTS ON rather than one a
   * player can see — a d4 (`'top-vertex'`) or any odd-sided barrel
   * (`'down-face'`). Such a die prints its values at its corners.
   */
  readonly cornerPrinted: boolean;
}

/**
 * How much room the drawn solid needs, as a multiple of `--die-size`.
 *
 * `--die-size` sizes the solid's NOMINAL sphere (see `DieGeometry.nominalRadius`),
 * and for every closed-form die that is also its bounding sphere, so this is
 * 1.00 and the die's box is `--die-size`. A barrel is normalized by its SHORT
 * axis instead — which is what makes a d7's numerals legible rather than a 4.3px
 * smudge — so its bounding sphere is `boundingRadius / nominalRadius` times
 * larger: 1.37x for a d3, 2.37x for a d7, up to 2.63x. A tumble points that long
 * axis in every direction, so the room to reserve is the BOUNDING SPHERE's box,
 * not the resting silhouette's.
 *
 * The `d / sqrt(d^2 - r^2)` term is the camera: `#inner` projects the solid from
 * `PERSPECTIVE_DEPTH_DIE_SIZES` die-sizes away, which magnifies whatever is
 * nearest. `r` is the bounding radius in die-sizes, and the factor is the widest
 * a sphere of that radius can project to — 1.0035 for a closed-form solid (a
 * third of a percent, and the reason a d6's box is not exactly `--die-size` to
 * the last decimal), 1.020 for a d7. Without it a d7 at 100px reserves 237px and
 * draws 243, which is a 6px overlap that no test would ever explain.
 *
 * Returns the box's SIDE, not its half-width.
 */
export function solidExtent(geometry: DieGeometry): number {
  const radius = 0.5 * (geometry.boundingRadius / geometry.nominalRadius);
  const depth = PERSPECTIVE_DEPTH_DIE_SIZES;
  return (2 * radius * depth) / Math.sqrt(depth * depth - radius * radius);
}

// Building a solid runs a convex hull for the closed-form shapes, so it is
// cached per face count. `null` records a face count that has no solid, so a
// malformed die does not retry the failure on every render pass.
//
// Deliberately unbounded: the key is a die's face count, so the cache is
// bounded by the number of DISTINCT dice the loaded games define (a handful),
// not by the number of dice on the board or by anything a player can drive.
const SOLID_CACHE = new Map<number, DieSolid | null>();

function dieSolid(faceCount: number): DieSolid | null {
  const cached = SOLID_CACHE.get(faceCount);
  if (cached !== undefined) return cached;
  let solid: DieSolid | null = null;
  try {
    const geometry = dieGeometry(faceCount);
    // A d4 and every odd-sided barrel are read from the face they REST ON, so
    // painting the value only at each face's centre lands the result face-down
    // against the table. `die-faces.ts` owns which solids those are; a real d4
    // answers it by printing the value at the CORNERS of the faces that stay
    // visible, and each of those corners carries the value that is read when
    // that corner is the top of the die.
    const cornerPrinted = resolveReadingRule(geometry) !== 'up-face';
    solid = {
      geometry,
      cornerPrinted,
      facets: solidFacets(
        geometry,
        cornerPrinted ? { cornerOwner: (vertex) => faceReadFrom(geometry, vertex) } : {},
      ),
    };
  } catch (error) {
    // A face count with no solid (fewer than 3 faces, or a shape the geometry
    // module rejects) falls back to the reel rather than throwing mid-render.
    // Silently is not good enough: a bug in `facetPlacement` would land here too
    // and degrade a d20 into a 20-tall reel with nothing in the console to say
    // why, so say it — once per face count, since the result is then cached.
    console.warn(`boardgame-die: no solid for ${faceCount} faces; falling back to the reel`, error);
    solid = null;
  }
  SOLID_CACHE.set(faceCount, solid);
  return solid;
}

/** Everything the die reads off an item, validated, or null for "not a die". */
interface DieItem {
  readonly id: string;
  readonly faces: readonly number[];
  readonly selectedFace: number;
  /**
   * How many times the server says this die has been thrown, or `null` when it
   * does not say.
   *
   * `components/dice`'s `DynamicValue.RollCount` is incremented by `Roll()` and
   * by nothing else, so a change in it means exactly "this die was thrown" — the
   * one fact `SelectedFace` and `Value` cannot express, because a throw landing
   * on the face already showing leaves both of them alone. A die whose game does
   * not use that component reports nothing here, and falls back to the face
   * change; see `_itemChanged`.
   */
  readonly rollCount: number | null;
}

/**
 * Read an item, defensively, in ONE place.
 *
 * THE CASE THIS EXISTS FOR is the sanitized component. A die in a hidden stack
 * is not absent — `selectors.ts` renders an occupied but unreadable slot as
 * `{}`, a TRUTHY object with no `Values` — so reaching for `Values.Faces`
 * threw a TypeError straight out of `updated()`, where Lit does not catch it:
 * `updateComplete` rejects and the page gets an unhandled rejection on every
 * update, forever.
 *
 * Deliberately NOT `isVisibleComponent`, which is the framework's stricter test
 * of the same family. That one also demands `Index`, `Deck` and `GameName`,
 * three fields this component never reads, and requiring them would reject a
 * die driven by a hand-built item — which is how the renderer fixtures and any
 * game prototyping a die drive one. It is also not SUFFICIENT: a card passes it
 * and has no faces at all, so the face list has to be checked here regardless.
 * What a die needs from an item is a list of face values, and that is what this
 * asks for.
 */
function readDieItem(item: DieComponent | null | undefined): DieItem | null {
  if (!item || typeof item !== 'object') return null;
  const candidate = item as {
    ID?: unknown;
    Values?: { Faces?: unknown };
    DynamicValues?: { SelectedFace?: unknown; RollCount?: unknown };
  };
  const faces = candidate.Values?.Faces;
  if (!Array.isArray(faces) || !faces.every((face) => typeof face === 'number')) return null;
  const selected = candidate.DynamicValues?.SelectedFace;
  const rollCount = candidate.DynamicValues?.RollCount;
  return {
    id: typeof candidate.ID === 'string' ? candidate.ID : '',
    faces: faces as readonly number[],
    selectedFace:
      typeof selected === 'number' && Number.isFinite(selected) ? Math.trunc(selected) : 0,
    rollCount:
      typeof rollCount === 'number' && Number.isFinite(rollCount) ? rollCount : null,
  };
}

/**
 * What `--die-size` resolves to when a caller sets nothing, in px.
 *
 * 100, which is pig's -- the only shipping game with dice, and the size every
 * legibility number in this component and in `die-shape.spec.ts` is measured
 * at. It was 50, inherited from the flat die, and that number stopped being
 * right the moment `--die-size` became a bounding-SPHERE diameter rather than a
 * face's width: at 50 a d6 draws a 29px cube in a 50px box, and a d7's corner
 * numerals come out at 4.3px, which is a smudge. Both of the tutorial's
 * `<boardgame-die>` snippets set nothing, so copied verbatim they used to
 * produce exactly that.
 *
 * A default cannot be right for every board, and this one is deliberately at
 * the large end: a die is a game's primary button, an author who wants a
 * smaller one will say so, and an author who says nothing is far better served
 * by a die that is too big to miss than by one that is too small to read.
 */
const DEFAULT_DIE_SIZE_PX = 100;

/**
 * How long the landing beat runs, in ms. Long enough to be a beat and short
 * enough that it is over before a player has finished reading the number.
 */
const SETTLE_ACCENT_MS = 260;

/**
 * Face counts already warned about for illegible marks; see `_checkLegibility`.
 * Module-scoped, so a board full of the same shape produces one line and not
 * one per die.
 */
const WARNED_ILLEGIBLE = new Set<number>();

/** One planned roll: what to draw, and what to play. */
interface DieRoll {
  /** Face VALUES by face index, with the server's value on the landed face. */
  readonly faces: readonly number[];
  /** The face index the physics turned up, i.e. the one carrying the value. */
  readonly presented: number;
  /** The simulator could not settle this throw in eight attempts. */
  readonly cocked: boolean;
  readonly durationMs: number;
  readonly curve: (progress: number) => string;
  /** Byte-identical to `curve(1)`: see `_playRoll`. */
  readonly resting: string;
}

class BoardgameDie extends BoardgameAnimatableItem {
  static override styles = [
    ...(BoardgameAnimatableItem.styles ? [BoardgameAnimatableItem.styles] : []),
    css`
      :host {
        --effective-die-scale: var(--die-scale, 1.0);
        /*
         * --die-size IS THE DIAMETER OF THE SPHERE THE SOLID IS SIZED AGAINST,
         * and is the property a caller sets: any CSS length ('120px', '6rem',
         * '10vmin').
         *
         * IT IS NOT ALWAYS THE DIE'S FOOTPRINT, and that is the one thing about
         * it worth reading twice. For every solid with a closed form -- a d6, a
         * d20, a d12 -- the sphere it is sized against IS its bounding sphere,
         * so the die fits a --die-size box in every orientation and the two
         * numbers are the same. A BARREL (a d3, d5, d7, d9, d16, ... -- every
         * face count with no closed form) is 1.37 to 2.63 times longer than it
         * is wide, and it is deliberately sized by its WIDTH, because its
         * readable faces are its side faces and their content is bounded by the
         * width: sizing a d7 by its long diagonal instead put its numerals at
         * 4.3px on a default die, which cannot be read at all.
         *
         * So a barrel is LARGER than --die-size along its axis, and the
         * component reserves the room for it rather than overlapping whatever
         * is beside it: #scaler's box is --die-size * --solid-extent, which the
         * render pass computes per shape (see solidExtent). What a caller can
         * still rely on is that the die never draws outside the box it reserves
         * -- die-shape.spec.ts pins exactly that, for every shape -- and that
         * --die-size is the number every mark on the die is scaled from.
         *
         * A SPHERE, NOT A FACE. This is the one thing about the property worth
         * saying twice, because it changed meaning when the flat die became a
         * solid and nothing in the name says so: on the reel, --die-size was a
         * face's own width. It is not any more. A cube's face spans 1/sqrt(3) =
         * 57.7% of it, a d20's triangle less, a barrel's side face less again,
         * so a die set to 50px draws a 29px cube inside a 50px box. Sizing a
         * die by eye off the old number therefore produces something about
         * half the size the author meant. The default below is what a caller
         * who sets nothing gets, and it is chosen so that "nothing" is a
         * reasonable answer rather than a smudge.
         *
         * --effective-die-size is the resolved value everything in here
         * measures against; it is not part of the component's API.
         */
        --effective-die-size: var(--die-size, ${DEFAULT_DIE_SIZE_PX}px);
        /*
         * How far #inner scrolls per face of the REEL. One die-size, which is
         * a reel face's height -- except on a solid, which has no reel to
         * scroll and sets it to zero (see #inner.solid). It is a variable of
         * its own rather than a re-definition of --effective-die-size so that
         * zeroing it cannot silently zero anything else below #inner that
         * measures against the die's size.
         */
        --reel-step: var(--effective-die-size);
      }

      /*
       * THE SPACE THE DIE RESERVES, which is not always --die-size.
       *
       * --solid-extent is how wide the drawn solid's own bounding box is, as a
       * multiple of --die-size, and the component sets it per shape (see
       * solidExtent). For every solid whose nominal sphere IS its bounding
       * sphere -- a d6, a d20, every closed form -- it is 1.00, and this box is
       * --die-size exactly, as it always was. For a BARREL it is not: a d7 is
       * 2.37 times longer than it is wide, and since the barrel is deliberately
       * sized by its WIDTH so its numerals are legible, the solid is larger than
       * --die-size along its axis and can point that axis anywhere.
       *
       * So the barrel's box is the box its bounding SPHERE needs, and the die
       * takes that much room in a layout. That is the whole point: before this,
       * a d7 at --die-size 100px drew 243px wide inside a 100px box and simply
       * overlapped whatever was beside it, silently. #main -- the hit target,
       * the contact shadow's anchor and the 3D scene's positioning context --
       * stays --die-size and stays centred here, so nothing about how the die
       * LOOKS changes; only how much room it asks for.
       */
      #scaler {
        height: calc(var(--effective-die-size) * var(--effective-die-scale)
                     * var(--solid-extent, 1));
        width: calc(var(--effective-die-size) * var(--effective-die-scale)
                    * var(--solid-extent, 1));
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
        transition: box-shadow 0.28s cubic-bezier(0.4, 0, 0.2, 1),
                    transform 0.18s cubic-bezier(0.4, 0, 0.2, 1);
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
       * only the hit target, the 3D scene's positioning context, and the
       * element the contact shadow hangs off. It must give up overflow:hidden
       * (which would slice the solid, and which flattens any 3D context put on
       * it) along with the flat card look -- INCLUDING the box-shadow, which
       * would draw a rectangle around a solid that is not one.
       */
      #main.solid,
      #main.solid.interactive:hover {
        position: relative;
        overflow: visible;
        background: none;
        border-radius: 0;
        box-shadow: none;
      }

      /*
       * THE CONTACT SHADOW, which is the thing that puts the die on the table.
       *
       * The flat die it replaces carried a full elevation box-shadow and a
       * hover lift; the solid arrived with 'box-shadow: none' and nothing else,
       * so against pig's cream board it read as low-contrast and FLOATING --
       * for the game's primary button. The shadow cannot go on the solid: a
       * 'filter: drop-shadow' forces 'transform-style: flat' and would collapse
       * the whole 3D context (see #stage's comment), and a 'box-shadow' draws
       * the wrong outline. So it is a soft ellipse under the die instead, which
       * is also what a real die on a real table casts.
       *
       * z-index -1 keeps it behind the solid. #main carries a transform, so it
       * is a stacking context, and the negative layer is inside it.
       */
      #main.solid::after {
        content: '';
        position: absolute;
        z-index: -1;
        left: 50%;
        bottom: 2%;
        width: 62%;
        height: 12%;
        transform: translateX(-50%);
        border-radius: 50%;
        background: radial-gradient(closest-side,
                    rgba(60, 40, 20, 0.42) 0%,
                    rgba(60, 40, 20, 0.26) 45%,
                    rgba(60, 40, 20, 0) 100%);
        transition: width 0.28s cubic-bezier(0.4, 0, 0.2, 1),
                    height 0.28s cubic-bezier(0.4, 0, 0.2, 1),
                    bottom 0.28s cubic-bezier(0.4, 0, 0.2, 1),
                    opacity 0.28s cubic-bezier(0.4, 0, 0.2, 1);
      }

      /*
       * THE HOVER AFFORDANCE. The flat die lifted on a raised elevation
       * shadow; the solid's only hover cue was a barely-perceptible facet
       * brightening, which nobody reads as "this is a button". So the die
       * lifts and its shadow spreads and softens underneath it, which is the
       * same gesture the elevation shadow made and the one every other
       * interactive component in the app makes.
       *
       * The lift is a translate on #main, ABOVE the 3D scene, so it composes
       * with nothing the roll owns (#inner) and nothing the pose owns
       * (#orient).
       */
      #main.solid.interactive:hover {
        transform: scale(var(--effective-die-scale)) translateY(-4%);
      }

      #main.solid.interactive:hover::after {
        width: 72%;
        height: 9%;
        bottom: -2%;
        opacity: 0.72;
      }

      /*
       * The same status line <boardgame-game-board> renders, down to the class
       * name, the role and the :empty rule: one status element per interactive
       * surface, polite, and gone entirely when it has nothing to say. What is
       * PUT in it is the difference -- see _statusMessage: only a reason a
       * person has to act on, never a die that is merely busy rolling.
       */
      .interaction-status {
        position: absolute;
        top: 100%;
        width: max-content;
        max-width: 16rem;
        margin-top: 0.25rem;
        color: var(--md-sys-color-error, #ba1a1a);
        font-size: 0.75rem;
      }

      .interaction-status:empty {
        display: none;
      }

      /*
       * The die's result, for a screen reader only. A button's aria-label
       * changing is NOT announced, so the settled value never reached one; a
       * live region is the only thing that carries it. Visually hidden rather
       * than display:none, which would take it out of the accessibility tree
       * along with the rest.
       */
      .visually-hidden {
        position: absolute;
        width: 1px;
        height: 1px;
        margin: -1px;
        padding: 0;
        overflow: hidden;
        clip-path: inset(50%);
        white-space: nowrap;
        border: 0;
      }

      /*
       * The solid: a sizing wrapper (#stage), the preserve-3d carrier that also
       * carries the projection (#inner -- the element motionTrackTarget('visual')
       * returns, so the tumble animates on it), a resting-pose carrier (#orient)
       * and one element per surface polygon.
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
      }

      #inner {
        position: relative;
        transform: translateY(calc(-1 * var(--reel-step) * var(--selected-face)));
        /* The spin is WAAPI-driven now; no CSS transform transition. */
      }

      #inner.solid {
        position: absolute;
        inset: 0;
        transform-style: preserve-3d;
        /*
         * THE CAMERA, and it is on #inner rather than on #stage deliberately.
         *
         * A perspective() function projects everything to its RIGHT in the
         * transform list, so putting it here lets a roll write
         * "translate3d(travel) perspective(D) <the pose>" and have the die
         * projected about its OWN centre before it is moved -- the camera rides
         * with the solid. It has to, because backface-visibility culls a facet
         * by the sign of its accumulated m33 and never asks where the camera is:
         * with
         * the camera pinned to the middle of the box, a die 90px away had
         * facets the camera could see culled out of it, and a d20 spent 22% of
         * its frames with a see-through hole in it. See dice-roll.ts's
         * rollScene for the measurement and the arithmetic.
         *
         * A die that has never rolled has no inline transform, so this rule is
         * also what frames the resting solid; the depth is the same constant
         * either way (PERSPECTIVE_DEPTH_DIE_SIZES), in em so it follows
         * --die-size with no JavaScript.
         */
        transform: perspective(${PERSPECTIVE_DEPTH_DIE_SIZES}em);
        /*
         * There is no reel to scroll, so the reel step is zero. The #inner
         * rule above and the WAAPI spin keyframes (_innerTransformForFace)
         * both read --reel-step, so zeroing it here makes the face-change spin
         * a no-op on the solid without touching either -- the roll is a real
         * tumble in a later task, and until then the die must not slide.
         * Scoped to --reel-step and NOT to --effective-die-size, which
         * everything under here still needs at its true value.
         */
        --reel-step: 0px;
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
        /*
         * THE FACET'S OWN LAYER, and it is what keeps the solid closed while it
         * tumbles. Not a performance hint -- a correctness one.
         *
         * backface-visibility: hidden is the whole of this solid's
         * hidden-surface removal: every facet is opaque, the surface is closed
         * and convex, so the facets whose outward normal faces the camera tile
         * the silhouette exactly and the rest must be culled. Chromium decides
         * that per facet from the transform accumulated down the preserve-3d
         * chain -- which includes #inner's, the element the tumble animates.
         * While that animation RUNS, the decision is baked into the raster of
         * #inner's subtree rather than re-evaluated every frame, so a facet that
         * crosses from back-facing to front-facing mid-tumble stays culled for a
         * frame or two and the die is briefly see-through. Promoting each facet
         * gives it a transform node of its own, which cc re-evaluates per frame.
         *
         * MEASURED (12 seeded rolls per shape, every composited frame captured
         * and checked for background inside the silhouette's convex hull -- the
         * solid is convex, so any such pixel is a hole). Largest connected hole,
         * before -> after:
         *
         *   d20 @200px  22/593 frames, worst 5245px  ->  0/705 frames
         *   d7  @200px   4/415 frames, worst  6187px ->  0/494 frames
         *   d9  @200px  13/526 frames, worst  9075px ->  0/587 frames
         *   d12 @200px   2/562 frames, worst  4050px ->  0/610 frames
         *   d7  @100px   8/392 frames, worst  2900px ->  1/449, worst 447px
         *
         * Note what that table says and the branch's earlier claim did not: the
         * tear was never a barrel problem. The d20 is the WORST affected shape,
         * and a d6 -- whose facets are squares, so its facet boxes never overhang
         * into a neighbour -- is the only one that never tore at all.
         *
         * IT IS ONLY VISIBLE IN A ROLL THAT IS ACTUALLY PLAYING. The same
         * transforms written to #inner as an inline style, or reached by pausing
         * the animation and seeking to them, render clean: 2,413 static frames
         * over ten shapes produced one 12x47 sliver. So frame-stepping a paused
         * die cannot see this bug and cannot verify this fix; only sampling a
         * free-running roll can. die-shape.spec.ts does the latter.
         *
         * Why not simply drop the culling and let depth sorting hide the back
         * faces: measured, it does not. backface-visibility: visible also
         * takes the holes to zero, but it changes what is DRAWN -- over 1,831
         * static frames on ten shapes, every single frame differed from the
         * culled render, by up to a 2,604px connected region, with back faces'
         * numerals showing through the front of the solid. This rule is
         * pixel-identical instead: over the same 1,831 frames, two frames on a
         * d6 differ, by a 14x12 antialiasing blob on one seam.
         *
         * The cost is one composited layer per facet -- 6 for a d6, 32 for a
         * d32 -- which for a 200px die is well under a megabyte. It is declared
         * unconditionally rather than only while a roll plays because
         * will-change has to be in effect BEFORE the change it describes, and
         * a promotion applied in the same task that starts the animation is not
         * reliably in place for the first frames -- which are frames that tear.
         */
        will-change: transform;
        background:
          linear-gradient(135deg, #F5F0E8 0%, #E0D9CE 100%);
        box-shadow: inset 0 0 0 1px rgba(60, 40, 20, 0.14);
        transition: background 0.2s ease-out;
      }

      #main.solid.interactive:hover .facet {
        background:
          linear-gradient(135deg, #FFFDF8 0%, #EFE9DE 100%);
      }

      .facet,
      .face {
        font-family: var(--md-sys-typescale-body-large-font, 'Source Sans 3', sans-serif);
        font-weight: 500;
        --die-ink: var(--md-sys-color-on-surface, #1C1810);
        color: var(--die-ink);
      }

      /*
       * THE FACET THE PLAYER IS MEANT TO READ, said in ink rather than only in
       * geometry.
       *
       * The pose already guarantees the presented facet is the most square-on
       * one, but on a d20 the margin it can afford is small -- 0.946 towards
       * the camera against a runner-up at 0.891, which is about 2.8% of
       * projected area and below what an eye resolves. At 520px a neighbouring
       * numeral is just as readable, and the only thing distinguishing the
       * right one is that it happens to be in the middle.
       *
       * So the presented facet gets darker ink and a faintly warmer, brighter
       * face. Both are cheap, neither touches geometry, and -- this is the
       * point -- neither can be mistaken for the pose failing, because the
       * emphasis follows the face carrying the value even if the aim ever did
       * drift. The ink is derived from whatever ink the theme supplies rather
       * than hard-coded, so it stays a RELATIVE darkening in any palette.
       */
      .facet.presented {
        color: color-mix(in srgb, var(--die-ink) 76%, #000 24%);
        background:
          linear-gradient(135deg, #FFFCF3 0%, #E8DFCF 100%);
        box-shadow: inset 0 0 0 1px rgba(60, 40, 20, 0.26);
      }

      #main.solid.interactive:hover .facet.presented {
        background:
          linear-gradient(135deg, #FFFFFA 0%, #F1E9D9 100%);
      }

      /*
       * A reel face is one die-size square, so its content square is just a
       * centred fraction of it. 63% reproduces the flat die's original pip
       * geometry: on a 50px face a dot lands 10.5px off centre, exactly where
       * it used to, and measures 6.3px across where it used to be 7.
       */
      #inner.reel .face {
        height: var(--effective-die-size);
        width: var(--effective-die-size);
        position: relative;
        --content-left: 18.5%;
        --content-top: 18.5%;
        --content-width: 63%;
        --content-height: 63%;
        --content-size: calc(var(--effective-die-size) * 0.63);
      }

      /*
       * Every mark a face carries lives inside its CONTENT SQUARE, whose box
       * the facet supplies as four percentages of its own box plus the square's
       * side in em (--content-size) for the marks to size themselves against.
       * The square is the largest that fits inside the facet's polygon, so a
       * mark that fits the square cannot leave the facet -- which is what makes
       * one layout work on a cube's square, a d20's triangle, a d10's kite and
       * a d7's 2.7:1 barrel face alike.
       */
      .content {
        position: absolute;
        left: var(--content-left);
        top: var(--content-top);
        width: var(--content-width);
        height: var(--content-height);
        --pip-size: calc(var(--content-size) * ${PIP_DIAMETER});
        display: flex;
        align-items: center;
        justify-content: center;
        pointer-events: none;
      }

      /*
       * Text sizes itself in em against the facet's font-size (the die's size)
       * and is set per mark, because it depends on both the square it is in
       * and how many characters it has to fit.
       */
      .content > span,
      .corner > span {
        line-height: 1;
        white-space: nowrap;
      }

      /*
       * The corner marks of a die that is read from a face nobody can see.
       * Position and size are percentages of the FACET's box, never em: the
       * span inside sets a font-size, and an em on the same element would then
       * resolve against that instead of the die's size.
       */
      .corner {
        position: absolute;
        display: flex;
        align-items: center;
        justify-content: center;
        pointer-events: none;
        opacity: 0.85;
      }

      /*
       * Pips are placed on the 3x3 lattice of the content square: cell centres
       * at a sixth, a half and five sixths of it. The cells a value occupies
       * are computed (see pipCells), not enumerated in CSS -- which is what
       * used to cap the die at six faces.
       */
      .pip {
        background-color: currentColor;
        height: var(--pip-size);
        width: var(--pip-size);
        border-radius: 50%;
        position: absolute;
        box-shadow: inset 0 1px 2px rgba(0, 0, 0, 0.4),
                    0 1px 0 rgba(255, 255, 255, 0.2);
      }
    `
  ];

  @property({ type: Object })
  item: DieComponent | null = null;

  /**
   * The face VALUES this die carries, in face-index order — `[10, 20, 30]`, not
   * `[0, 1, 2]`.
   *
   * Set from `item.Values.Faces` when there is an item, and settable directly to
   * drive a die by hand. A value is what the face DRAWS and what the die
   * announces for it; nothing here reads a value as an index.
   */
  @property({ type: Array })
  faces: number[] = [];

  /**
   * Which face is presented, as an INDEX into `faces` — not a face value.
   *
   * `2` on a die with faces `[10, 20, 30]` presents the face showing 30. This is
   * the server's own convention (`DynamicValues.SelectedFace` is an index
   * alongside a separate `Values.Faces` list) and reading it as a value is the
   * silent bug this component invites: it is in range, it selects a face, and
   * the die shows the wrong number.
   */
  @property({ type: Number })
  selectedFace = 0;

  /**
   * Face VALUE to the name that face carries — `{ 3: 'Star' }`.
   *
   * The seam an enum plugs into. The framework's `enum` package already sends
   * its values to the client, so a later change supplies this from the enum a
   * die's faces are typed with and nothing else here moves: a name selects the
   * glyph out of `symbols`, and it is what the die announces. With no names
   * attached, a face's name is its own value written out, so a symbol set can
   * be keyed by plain integers and still work.
   */
  @property({ type: Object })
  faceNames: Record<string, string> | null = null;

  /**
   * Face NAME to the glyph drawn for it — `{ Star: '★' }`.
   *
   * The author-supplied symbol set, and the first thing face content resolves
   * to: a face with a glyph draws the glyph whatever its value would otherwise
   * have generated.
   */
  @property({ type: Object })
  symbols: Record<string, string> | null = null;

  /**
   * The version of the game state this die is showing.
   *
   * The FALLBACK half of a roll's identity — the other is the component's own
   * ID — used only by a die whose item carries no `DynamicValues.RollCount`
   * (one driven by hand, or served by an API binary built before the count
   * existed). Where the count is reported it is the seed, because it is the one
   * number that means "this die was thrown"; see `_rollIdentity`.
   *
   * Left unset, it is discovered from the nearest ancestor renderer's
   * `gameVersion`, the same ambient climb `animationContext` uses, so no game
   * has to wire it up; set it explicitly to drive a die that has no such
   * ancestor.
   */
  @property({ type: Number })
  stateVersion: number | null = null;

  @property({ type: Boolean })
  disabled = false;

  @property({ attribute: false })
  action: BoundMoveAction<string, object> | null = null;

  @query('#inner')
  private _innerElement?: HTMLElement;

  @query('#stage')
  private _stageElement?: HTMLElement;

  /** The roll the die is currently showing, or null before it has ever rolled. */
  private _roll: DieRoll | null = null;

  /**
   * A roll planned but not yet played. It is played from the update pass AFTER
   * the one that planned it, which is what puts the roll's new face values on
   * screen before its first frame — see `_startRoll`.
   */
  private _pendingRoll: DieRoll | null = null;

  /** How many times an `item` has been installed; the first one is not a roll. */
  private _itemInstalls = 0;

  /**
   * The roll count the current item reported, or null when it reported none.
   * A CHANGE in it is what a roll is; see `_itemChanged`.
   */
  private _rollCount: number | null = null;

  /** Set while the face change that the FIRST item install causes is in flight. */
  private _installingFace = false;

  /** The component ID the current item carries; half of a roll's seed. */
  private _componentId = '';

  /**
   * What the live region announces: the result of the roll that has just
   * settled, or '' before the die has finished one.
   *
   * Written when a roll ENDS, never during one — an aria-live region announces
   * every change it sees, and a die that narrated its own tumble would be worse
   * than one that said nothing.
   */
  private _announcement = '';

  /**
   * The `faceCount@sizePx` the legibility floor was last evaluated for, so an
   * unchanged die is not re-measured on every update pass.
   */
  private _legibilityCheckedFor: string | null = null;

  /**
   * Which roll is current. A roll that finishes after a LATER one has started
   * must not report itself as the die's result; see `_playRoll`.
   */
  private _rollGeneration = 0;

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

    // A roll planned during the PREVIOUS pass is played first, because the
    // render that has just finished is the one carrying its face assignment.
    const pending = this._pendingRoll;
    this._pendingRoll = null;
    if (pending) this._playRoll(pending);

    if (changedProperties.has('item')) {
      this._itemChanged(this.item);
    } else {
      // The face change an item install causes lands one pass later than the
      // install itself, and installing the first item is not a roll: the die is
      // being shown a state it was already in when it mounted.
      const installing = this._installingFace;
      this._installingFace = false;
      if (!installing && !pending && changedProperties.has('selectedFace')) {
        this._selectedFaceChanged(
          this.selectedFace,
          changedProperties.get('selectedFace') as number | undefined
        );
      }
    }

    if (changedProperties.has('action')) {
      this._subscribeAction();
    }

    // After the render, because it measures the rendered #stage; see there for
    // why it costs nothing on the passes where nothing has moved.
    const solid = this._solid();
    if (solid) this._checkLegibility(solid);
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
  // a given selectedFace: translateY(-1 * reel-step * face). The #main element
  // carries the --selected-face var that drives the CSS resting position; here
  // we build an explicit transform so WAAPI can interpolate the spin instead
  // of relying on a CSS transition.
  //
  // THE REEL FALLBACK ONLY. A solid rolls through a baked physics curve (see
  // _playRoll); this is what a die with no geometry -- fewer than three faces --
  // still does, and it is the pre-3D behaviour untouched.
  private _innerTransformForFace(face: number): string {
    return `translateY(calc(-1 * var(--reel-step) * ${face}))`;
  }

  protected override motionTrackTarget(target: ComponentMotionTarget): HTMLElement | null {
    return target === 'host' ? this : this._innerElement ?? null;
  }

  // A face change on a die that has already been mounted is a ROLL: the solid
  // tumbles through the physics, and only the degenerate reel still scrolls.
  //
  // This is the FALLBACK trigger, reached only by a die whose item carries no
  // `DynamicValues.RollCount` -- a die driven by hand, or by a game that does
  // not use `components/dice`. Where a roll count is reported it is the sole
  // trigger and `_itemChanged` throws the die directly; see there for why.
  private _selectedFaceChanged(newValue: number, oldValue: number | undefined) {
    if (!this._innerElement) return;
    // On first render there's no meaningful transition to animate from.
    if (oldValue === undefined || oldValue === newValue) return;
    if (this._solid()) {
      this._startRoll();
      return;
    }
    this.playMotionTracks(componentMotionTracks([{
      target: 'visual',
      property: 'transform',
      from: this._innerTransformForFace(oldValue),
      to: this._innerTransformForFace(newValue),
    }]));
  }

  /**
   * Install a new item, which is where a roll is noticed.
   *
   * SANITIZED ITEMS ARRIVE HERE, and are drawn as a die with no faces rather
   * than crashing the render pass; see `readDieItem`, which is the only place
   * this component reads the wire shape.
   */
  private _itemChanged(newValue: DieComponent | null | undefined) {
    const item = readDieItem(newValue);
    if (!item) {
      // A sanitized die is still a die, and a die with no item at all is not
      // one: both are drawn with no faces, and neither carries a roll.
      this.faces = [];
      this.selectedFace = 0;
      this._clearRoll();
      this._componentId = '';
      this._itemInstalls = 0;
      this._rollCount = null;
      this._announcement = '';
      return;
    }
    this.faces = [...item.faces];
    this.selectedFace = item.selectedFace;
    const componentId = item.id;
    // A DIFFERENT DIE on the same element, which is what a stack re-using a slot
    // does. The old roll's face assignment and resting pose belong to a
    // component that is no longer here, and keying that on the face COUNT alone
    // missed the case the framework actually produces: swap a d6 for another d6
    // and the die went on drawing -- and announcing -- the first one's numbers,
    // self-correcting only if a throw happened to arrive.
    const faceCount = this.faces.length;
    if (this._roll && (this._roll.faces.length !== faceCount || componentId !== this._componentId)) {
      this._clearRoll();
    }
    if (componentId !== this._componentId) {
      // A different component's install is a FIRST install: whether the die it
      // replaced had ever rolled says nothing about this one.
      this._itemInstalls = 0;
      this._rollCount = null;
      this._announcement = '';
    }
    this._componentId = componentId;
    const previousCount = this._rollCount;
    this._rollCount = item.rollCount;
    // The FIRST install is never a roll, whatever the counter says: the die is
    // being shown a state it was already in when it mounted, and a die that
    // tumbled because a page loaded would be lying about what had just
    // happened.
    // A THROW is what a roll is, and `DynamicValues.RollCount` is the server
    // saying one happened. Nothing else in the die's state can: a throw landing
    // on the face already showing leaves `SelectedFace` and `Value` untouched
    // (one throw in six for a d6), so the face change that used to be the
    // trigger silently skipped those rolls and the player saw the die not move.
    //
    // Deliberately not "the state version moved" either: the version moves for
    // every move any player makes -- a game view mounting installs this die
    // three times, at versions 0, 2 and 6, with the die untouched throughout --
    // so it says nothing about whether THIS die was thrown.
    //
    // Where a roll count is reported it is the SOLE trigger: the face-change
    // fallback below would otherwise fire a second time on the update pass the
    // install schedules, and a game that rewrote a die's face without throwing
    // it would get a tumble that never happened. `_installingFace` suppresses
    // that pass. The FIRST install is never a roll whatever the counter says --
    // the die is being shown a state it was already in when it mounted, and a
    // die that tumbled because a page loaded would be lying about what had just
    // happened.
    const install = this._itemInstalls++;
    if (this._rollCount !== null) {
      this._installingFace = true;
      if (install > 0 && previousCount !== null && this._rollCount !== previousCount) {
        this._startRoll();
      }
      return;
    }
    // A die whose game reports no roll count falls back to the face change,
    // which is the pre-`RollCount` behaviour and still right for a die driven by
    // hand: it arrives one update pass after the install, which is why
    // `_installingFace` exists.
    if (install === 0) this._installingFace = true;
  }

  /**
   * The state version this die's rolls are seeded from.
   *
   * The ambient renderer's `gameVersion` when there is one, exactly as
   * `animationContext` is discovered, so a game gets deterministic rolls without
   * wiring anything; the `stateVersion` property otherwise, for a die mounted on
   * its own. Ambient wins for the same reason it does there: the framework's
   * value is the authoritative one when the framework is present.
   */
  private _resolvedStateVersion(): number {
    const ambient = this._ambientLookup<number>(
      'gameVersion', (value) => typeof value === 'number' && Number.isFinite(value));
    if (ambient !== null) return ambient;
    const own = this.stateVersion;
    return typeof own === 'number' && Number.isFinite(own) ? own : 0;
  }

  /**
   * WHICH THROW this is: the other half of the seed, alongside the component ID.
   *
   * `DynamicValues.RollCount` when the server reports one, because it is the
   * only number in the die's state that means "this die was thrown" — it is
   * incremented by `Roll()` and by nothing else. That makes it the roll's real
   * identity: it changes exactly once per throw, and it does not change for any
   * other reason.
   *
   * NOT the state version, which was the seed before the count existed. The
   * version moves for every move any player makes, and a game view mounting
   * installs one die three times at three different versions with the die
   * untouched — so a component that remounts DURING a roll would re-derive a
   * DIFFERENT trajectory for the SAME throw, and the die would visibly change
   * its path mid-air. The count cannot do that: the same throw carries the same
   * count for as long as it is on screen.
   *
   * The version remains the fallback for a die whose game does not use
   * `components/dice` — one driven by hand, or served by an API binary built
   * before `RollCount` existed. That die is exactly as deterministic as it was
   * before, and exactly as exposed to the remount hazard; there is no better
   * signal available to it (see `_itemChanged`).
   */
  private _rollIdentity(): number {
    return this._rollCount ?? this._resolvedStateVersion();
  }

  /**
   * Throw the die, and schedule the tumble for the pass after this one.
   *
   * WHEN THE FACE VALUES SWAP. The assignment is recomputed for every roll, so
   * every face but the landed one carries a different number afterwards. That
   * swap has to happen either as the roll starts or as it ends, and this is
   * deliberately the start: `requestUpdate` re-renders with the new assignment
   * and `updated` plays the tumble on the NEXT pass, so the first frame anyone
   * sees is already both airborne and correctly numbered. Swapping at the end
   * instead would change a number under the eye of a player who has just watched
   * the die stop moving, which is the one moment they are certainly reading it.
   *
   * A roll that cannot be planned (no solid, no measurable size, a trajectory
   * the bake refuses) leaves `_roll` null and CLEARS the pose the last roll
   * left behind, so the die falls back to the deterministic presentation pose
   * rather than half-rendering a physics one; see `_clearRoll`.
   */
  private _startRoll(): void {
    const roll = this._planRoll();
    if (!roll) {
      this._clearRoll();
      this.requestUpdate();
      return;
    }
    this._rollGeneration++;
    this._roll = roll;
    this._pendingRoll = roll;
    // `roll-start` is dispatched by `_playRoll`, not here: this pass has only
    // PLANNED the throw, and the render carrying its face values has not run
    // yet. A listener that fired now would read the die's previous numbers off
    // the DOM.
    // _roll is not a reactive property: nothing re-renders without this.
    this.requestUpdate();
  }

  /**
   * Give up the current roll, INCLUDING the pose it left on the element.
   *
   * `#inner` carries the physics pose and `#orient` the resting one, and they
   * are mutually exclusive only because `_renderSolid` renders `#orient` as
   * `none` whenever `_roll` is set. That invariant is one-way: `#orient` is a
   * CHILD of `#inner`, so the two compose, and Lit re-uses the same static
   * `#inner` node across renders with whatever inline transform was last written
   * to it. Dropping `_roll` without this leaves the die posed by a throw it is
   * no longer showing -- measured over 900 seeded rolls, 60 to 106px outside its
   * own 100px slot, permanently, until some later roll happens to overwrite it.
   *
   * Both paths that reach it are real: a die whose item is replaced by a
   * different component, and a roll that cannot be planned (an unmeasurable
   * size, or a throw the simulator or the bake refuses).
   */
  private _clearRoll(): void {
    this._roll = null;
    this._pendingRoll = null;
    // The element only exists once the solid has rendered; a die that has never
    // had one has nothing to clear either.
    if (this._innerElement) this._innerElement.style.transform = '';
  }

  /** What a `roll-start`/`roll-end` event carries. */
  private _rollDetail(roll: DieRoll) {
    return {
      value: roll.faces[roll.presented],
      faceIndex: roll.presented,
      cocked: roll.cocked,
      durationMs: roll.durationMs,
    };
  }

  private _planRoll(): DieRoll | null {
    const solid = this._solid();
    if (!solid) return null;
    const geometry = solid.geometry;
    const faces = this.faces;
    const desired = faces[this._presentedFaceIndex(geometry.faceCount)];
    if (!Number.isFinite(desired)) return null;
    // HALF THE DIE'S BOX on screen, i.e. one `nominalRadius` in px — not one
    // bounding radius, which for a barrel is up to 2.63x larger; `dice-roll.ts`
    // documents on `posedPosition` why the travel is scaled by this one. Read
    // from #stage's font-size because that IS the die's size (the solid is built
    // at 1em across), and it is resolved to a NUMBER of pixels here rather than
    // left as a `calc()` over `--effective-die-size` in the keyframes.
    //
    // NOT because a `calc()` cannot composite — measured, a control animating
    // `translate3d(calc(var(--probe) * 80px), ...)` kept compositing and still
    // reported `ActiveTransformAnimation`, because Chromium substitutes a static
    // custom property when it composes the keyframes. The reason is that
    // `--die-size` is a caller's property and is not guaranteed static: a game
    // that retunes it mid-roll (a responsive rule, a theme switch, a `--die-scale`
    // transition) would invalidate a composited animation in mid-air. Resolving
    // it once, here, means the tumble depends on nothing it does not own. See
    // `dice-bake.ts`'s "Why literal `matrix3d`" for the frame-drop measurement
    // behind the guarantee itself.
    const stage = this._stageElement;
    const radiusPx = stage ? parseFloat(getComputedStyle(stage).fontSize) / 2 : NaN;
    if (!Number.isFinite(radiusPx) || radiusPx <= 0) return null;
    try {
      const trajectory = dieRollTrajectory(
        geometry, this._componentId, this._rollIdentity());
      // The tail of a throw is the simulator's rest-detection hold, i.e. a die
      // sitting perfectly still. Playing it would hold the gate open for ~300ms
      // of nothing; see `settledTrajectory`.
      const die = settledTrajectory(trajectory.dice[0]);
      const durationMs = die.samples[die.samples.length - 1].t;
      const presented = presentedFaceIndex(geometry, die.restingOrientation);
      // The scene owns the whole transform, projection included: see
      // `rollScene` for why the camera has to travel with the die.
      const scene = rollScene(geometry, die, presented, radiusPx, durationMs);
      return {
        faces: assignFaceValues(geometry, faces, presented, desired),
        presented,
        cocked: trajectory.cocked,
        durationMs,
        curve: scene.transform,
        // scene.transform(1) itself, so the two agree BYTE FOR BYTE. Animations
        // run with fill:'none', so the element renders this the instant the
        // tumble finishes -- or is finished early by the cycle sweep -- and a
        // single rounding digit of disagreement would show up as the die
        // twitching as it settles.
        resting: scene.resting,
      };
    } catch (error) {
      // A geometry the simulator or the bake refuses. Nothing here may throw
      // during a render pass, and a die that quietly stops rolling is worth a
      // line in the console.
      console.warn('boardgame-die: could not plan a roll; showing the value without one', error);
      return null;
    }
  }

  /**
   * Play the planned tumble, and see that the die ends up in its landed pose
   * WHETHER OR NOT it does.
   *
   * `playMotionTracks` has three early returns -- `missing-target`,
   * `not-started`, `playback-error` -- that all fire BEFORE it writes the
   * track's resting style. In every one of them `_roll` is already set, so
   * `#orient` renders `none` and `#inner` has nothing: the die would draw in its
   * raw body frame, showing whichever facet happens to have a +z normal, while
   * `aria-label` announced the value the physics landed. Announcing one number
   * and drawing another is the worst failure this component has, so the resting
   * pose is written by hand here when playback does not start. (`noAnimate` is
   * the only route that reaches it today; the kernel's own doc comment claims
   * this hole is closed, and this is what closes it.)
   */
  private _playRoll(roll: DieRoll): void {
    const generation = this._rollGeneration;
    let result: MotionTrackPlayResult | undefined;
    try {
      result = this.playMotionTracks(
        componentMotionTracks([{
          target: 'visual',
          property: 'transform',
          curve: roll.curve,
          resolution: Math.round(roll.durationMs / FRAME_MS),
          resting: roll.resting,
        }]),
        { duration: roll.durationMs },
        // REQUIRED. Under the default 'version' policy the kernel clamps the
        // duration into the cycle's slot -- a three-second bake played in 600ms is
        // geometrically faithful and physically absurd -- and can resolve to skip
        // outright, which reports 'not-started' and takes sibling tracks down with
        // it. The context is null in solo play, so both failures would appear only
        // in companion mode.
        { timing: 'immediate' },
      );
    } catch (error) {
      // componentMotionTracks compiles the curve OUTSIDE playMotionTracks' own
      // try, so a curve it refuses would escape into Lit's update pass and
      // reject updateComplete. Same treatment as a playback error: the die still
      // has to land.
      console.error('[boardgame-die] could not compile the roll:', error);
      result = undefined;
    }
    // Announced once the tumble is on screen -- or, below, once the die has
    // been put where the tumble would have left it.
    this._announcement = '';
    this.dispatchEvent(new CustomEvent('roll-start', {
      bubbles: true,
      composed: true,
      detail: this._rollDetail(roll),
    }));
    if (result?.status !== 'started') {
      if (this._innerElement) this._innerElement.style.transform = roll.resting;
      this._finishRoll(roll, generation);
      return;
    }
    // The animation's own settlement is the ground truth for "the die has
    // stopped", and it resolves for a tumble finished early by the cycle sweep
    // exactly as for one that ran to its end. A cancelled animation rejects; the
    // die is then no longer showing this roll and must not report it.
    void Promise.all(result.playbacks.map((playback) => playback.animation.finished))
      .then(() => this._finishRoll(roll, generation), () => undefined);
  }

  /**
   * The die has stopped: announce the result and say so.
   *
   * `roll-end` is what a game listens to in order to celebrate a number -- and
   * the reason it exists is that the alternative is celebrating at CYCLE start,
   * which fires an effect at the die's layout anchor while the solid is 60px
   * away in the air and finishes it while the roll still has a second to run.
   */
  private _finishRoll(roll: DieRoll, generation: number): void {
    // A roll superseded by a later one is not this die's result any more.
    if (generation !== this._rollGeneration || this._roll !== roll) return;
    const detail = this._rollDetail(roll);
    const values = this._faceValues();
    const label = values.length
      ? this._resolveFace(values[roll.presented], this._usesPips(this._solid())).label
      : String(detail.value);
    this._announcement = `Rolled ${label}`;
    this._playSettleAccent();
    this.requestUpdate();
    this.dispatchEvent(new CustomEvent('roll-end', {
      bubbles: true,
      composed: true,
      detail,
    }));
  }

  /**
   * THE LANDING BEAT: a short pop the instant the result arrives.
   *
   * Nothing used to mark the moment the die finished, and because the tail of
   * a throw decelerates there is no frame a player can point at as the one it
   * stopped on — the number simply becomes readable at some point. A quarter of
   * a second of overshoot-and-settle is enough to say "this is the answer", and
   * it is worth more than making the tumble itself bouncier.
   *
   * ON #STAGE, and deliberately: #inner is the roll's, #orient is the pose's,
   * and #main's transform is the hover lift's. #stage is above the whole 3D
   * scene and owns no transform otherwise, so a scale here composes with
   * nothing. It is also uniform, so it cannot change any of the ANGLES the
   * readability tests measure.
   *
   * PLAYED DIRECTLY, not through `play()`. It is not a gate participant — the
   * roll it punctuates has already settled, and holding the cycle open for a
   * flourish would delay every other player's board. Going through the kernel
   * would also declare a `will-animate` and record a `play` in the animation
   * hooks the parity goldens count, for an animation that is decoration. So
   * `noAnimate` and reduced motion are honoured here by hand instead.
   */
  private _playSettleAccent(): void {
    if (this.noAnimate) return;
    if (window.matchMedia('(prefers-reduced-motion: reduce)').matches) return;
    const stage = this._stageElement;
    if (!stage) return;
    stage.animate(
      [
        { transform: 'scale(1)', offset: 0 },
        { transform: 'scale(1.06)', offset: 0.3 },
        { transform: 'scale(0.99)', offset: 0.62 },
        { transform: 'scale(1)', offset: 1 },
      ],
      {
        duration: SETTLE_ACCENT_MS,
        easing: 'cubic-bezier(0.22, 1, 0.36, 1)',
        // fill: 'none' for the same reason every other play in this component
        // uses it: the element must be left in its own resting style, not
        // pinned to a keyframe.
        fill: 'none',
        composite: 'replace',
      },
    );
  }

  /**
   * Say so, ONCE, when this die's marks come out too small to read.
   *
   * Every other size assertion in the solid is an upper bound — stay inside the
   * facet, do not overrun the inscribed square — and all of them are satisfied
   * perfectly by drawing nothing at all. `die-shape.spec.ts` pins a floor for
   * the shapes that ship; this is the same floor applied to the shapes nobody
   * tested, at whatever size a game actually drew them. A game author who drops
   * a `d7` in at the default should not get silent 3px marks and be left to
   * guess why the die is a smudge.
   *
   * The size is DERIVED, from the facet's own content square times the die's
   * measured pixel size, never from a table of face counts: a change to how a
   * solid is proportioned moves the marks, and a hard-coded expectation here
   * would go quietly stale exactly when it mattered.
   *
   * Warned once per face count, and only once it has actually failed — a d7
   * drawn large somewhere else on the board records nothing, so the small one
   * still gets its warning. Bounded by the number of DISTINCT dice the loaded
   * games define, which is the same argument `SOLID_CACHE` makes.
   */
  private _checkLegibility(solid: DieSolid): void {
    const faceCount = solid.geometry.faceCount;
    if (WARNED_ILLEGIBLE.has(faceCount)) return;
    const stage = this._stageElement;
    if (!stage) return;
    const sizePx = parseFloat(getComputedStyle(stage).fontSize);
    if (!Number.isFinite(sizePx) || sizePx <= 0) return;
    // Re-measuring an unchanged die on every update pass would be a render-loop
    // cost for nothing; the answer only moves when the shape or the size does.
    const key = `${faceCount}@${Math.round(sizePx)}`;
    if (this._legibilityCheckedFor === key) return;
    this._legibilityCheckedFor = key;
    const values = this._faceValues();
    const usePips = this._usesPips(solid);
    let worst: { what: string; px: number } | null = null;
    const note = (what: string, px: number) => {
      if (!worst || px < worst.px) worst = { what, px };
    };
    for (const facet of solid.facets) {
      if (facet.faceIndex < 0 || facet.faceIndex >= values.length) continue;
      const content = this._resolveFace(values[facet.faceIndex], usePips);
      if (content.kind === 'pips') {
        const px = facet.contentSize * sizePx * PIP_DIAMETER;
        if (px < MIN_LEGIBLE_PIP_PX) note(`a pip on face ${values[facet.faceIndex]}`, px);
      } else {
        const px = facet.contentSize * sizePx * glyphScale(content.text, GLYPH_HEIGHT);
        if (px < MIN_LEGIBLE_GLYPH_PX) note(`"${content.text}" on face ${values[facet.faceIndex]}`, px);
      }
      for (const corner of facet.corners) {
        const mark = this._resolveFace(values[corner.faceIndex], false);
        const px = corner.size * sizePx * glyphScale(mark.text, CORNER_GLYPH_HEIGHT);
        if (px < MIN_LEGIBLE_GLYPH_PX) note(`the corner "${mark.text}"`, px);
      }
    }
    if (!worst) return;
    WARNED_ILLEGIBLE.add(faceCount);
    const { what, px } = worst as { what: string; px: number };
    console.warn(
      `boardgame-die: a d${faceCount} at --die-size ${Math.round(sizePx)}px draws ${what} at `
      + `${px.toFixed(1)}px, which is too small to read. --die-size is the diameter of the `
      + `SPHERE THE SOLID IS SIZED AGAINST, not a face's width, so a shape with small or `
      + `elongated faces needs a larger one than its face size suggests.`);
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
   * The face VALUES the die is currently drawing, by face index.
   *
   * The server's list until the die has rolled, and the roll's own assignment
   * afterwards — same multiset, permuted so the value the server chose sits on
   * the face the physics turned up.
   */
  private _faceValues(): readonly number[] {
    const faces = Array.isArray(this.faces) ? this.faces : [];
    const roll = this._roll;
    return roll && roll.faces.length === faces.length ? roll.faces : faces;
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

  /**
   * Which face the die is actually SHOWING: the one the physics landed once it
   * has rolled, and the selected one before that. The two carry the same value
   * either way — that is what the face assignment is for — but they are
   * different facets, and everything the player reads hangs off the facet.
   */
  private _shownFaceIndex(faceCount: number): number {
    const roll = this._roll;
    if (roll && roll.presented >= 0 && roll.presented < faceCount) return roll.presented;
    return this._presentedFaceIndex(faceCount);
  }

  /**
   * The name a face carries: the enum's string name once one is attached, and
   * otherwise the value written out. Also the die's accessible label for that
   * face, which is why it is one function and not two.
   */
  private _nameForValue(value: number): string {
    const names = this.faceNames;
    if (names && typeof names === 'object') {
      const name = names[String(value)];
      if (typeof name === 'string' && name.length > 0) return name;
    }
    return String(value);
  }

  /** The author-supplied glyph for a face name, or null when there is none. */
  private _glyphForName(name: string): string | null {
    const symbols = this.symbols;
    if (!symbols || typeof symbols !== 'object') return null;
    const glyph = symbols[name];
    return typeof glyph === 'string' && glyph.length > 0 ? glyph : null;
  }

  /**
   * Whether THIS DIE draws its unlettered faces as pips.
   *
   * A property of the whole die, not of each face: a die that mixed dots on
   * one face with a numeral on the next would read as two different dice, so
   * one value past the lattice's capacity (see `MAX_PIP_VALUE`) moves all of
   * them to numerals. That is what makes a d6 pipped and a d20 numbered
   * without either being named anywhere.
   *
   * Corner-printed dice (a d4, an odd barrel) are always numbered: their value
   * has to fit in a small square at a corner, where dots do not read, and a
   * face carrying pips in the middle and numerals at its corners reads as a
   * mistake.
   */
  private _usesPips(solid: DieSolid | null): boolean {
    if (solid?.cornerPrinted) return false;
    const faces = Array.isArray(this.faces) ? this.faces : [];
    return faces.every((value) =>
      isPipValue(value) || this._glyphForName(this._nameForValue(value)) !== null);
  }

  /**
   * A face's content, resolved in the one order this component documents:
   * author symbol set, then generated pips, then a numeral. `label` is what
   * the die ANNOUNCES for that face, and it always describes what is actually
   * drawn — which is the assertion that catches drawing one face's value while
   * announcing another's.
   *
   * `label` is published per facet as `data-face-label`, the deliberate
   * parallel to the pre-existing `data-face-value`: the pair says "this facet
   * carries value V and would be announced as L", and only the pair can be
   * checked against what the facet actually paints. `aria-label` cannot stand
   * in for it — it names the PRESENTED face only, so a die that announced the
   * right thing while labelling the other five faces wrongly would look
   * identical through the accessibility tree. It costs nothing per render:
   * both renderers resolve a face once and pass the result to `_faceContent`.
   */
  private _resolveFace(value: number, usePips: boolean): ResolvedFace {
    const name = this._nameForValue(value);
    const glyph = this._glyphForName(name);
    if (glyph !== null) {
      return { kind: 'symbol', text: glyph, cells: [], label: name };
    }
    // A named face with no glyph draws its number and announces both: the
    // number is what is on the facet, the name is what it means.
    const label = name === String(value) ? String(value) : `${name} (${value})`;
    if (usePips && isPipValue(value)) {
      return { kind: 'pips', text: '', cells: pipCells(value), label };
    }
    return { kind: 'numeral', text: String(value), cells: [], label };
  }

  /**
   * The content square of one face: its dots, or its glyph/numeral.
   *
   * Takes the ALREADY-resolved face rather than a value, so that the caller
   * that also needs the label (every caller) resolves it once.
   */
  private _faceContent(content: ResolvedFace) {
    if (content.kind === 'pips') {
      return html`<div class="content">
        ${content.cells.map(([col, row]) => html`<div
          class="pip"
          style="left:calc(${num(((col + 0.5) / 3) * 100)}% - var(--pip-size) / 2);top:calc(${num(((row + 0.5) / 3) * 100)}% - var(--pip-size) / 2)"
        ></div>`)}
      </div>`;
    }
    return html`<div class="content"><span
      style="font-size:calc(var(--content-size) * ${num(glyphScale(content.text, GLYPH_HEIGHT))})"
    >${content.text}</span></div>`;
  }

  /**
   * The values printed at a facet's corners, for a die read from a face nobody
   * can see. Each mark carries the value that would be READ if that corner
   * were the top of the die — so a d4 resting on face 2 shows a 2 at the apex
   * of the three faces still facing the player. Always a single glyph or
   * numeral, never dots: the square at a corner is a third of the face's.
   */
  private _cornerContent(facet: SolidFacet, values: readonly number[]) {
    return facet.corners.map((corner) => {
      const value = values[corner.faceIndex];
      const content = this._resolveFace(value, false);
      return html`<div
        class="corner"
        data-corner-face-index="${corner.faceIndex}"
        style="left:${num(corner.left)}%;top:${num(corner.top)}%;width:${num(corner.width)}%;height:${num(corner.height)}%"
      ><span
        style="font-size:${num(corner.size * glyphScale(content.text, CORNER_GLYPH_HEIGHT))}em"
      >${content.text}</span></div>`;
    });
  }

  // The degenerate fallback: the pre-3D vertical reel of flat faces, scrolled
  // to the selected one by #inner's translateY. It has no geometry, so no
  // corner marks -- and nothing with fewer than three faces is read from a
  // face it rests on anyway.
  private _renderReel() {
    const usePips = this._usesPips(null);
    return html`
      <div id="inner" class="reel">
        ${repeat(this.faces, (face) => face, (face) => {
          const content = this._resolveFace(face, usePips);
          return html`
          <div class="face" data-face-value="${face}" data-face-label="${content.label}"
            >${this._faceContent(content)}</div>
        `;
        })}
      </div>
    `;
  }

  private _renderSolid(solid: DieSolid) {
    const values = this._faceValues();
    const usePips = this._usesPips(solid);
    // Once the die has rolled, the whole of its pose is the scene's: #inner
    // holds the tumble (and, once it finishes, the trajectory's own resting
    // transform, written by the motion-track kernel), and that transform ends
    // in `landedReadingPose` -- the aim, the lean AND the roll that leaves the
    // landed numeral upright. So #orient must carry nothing at all, or the two
    // would compose into a third pose.
    //
    // It used to carry the upright roll, about the landed face's own normal,
    // because the scene structurally could not: a rotation about that normal
    // fixes it, so the presented facet stayed the most square-on one. What it
    // did NOT fix was every other facet's depth, and on a d4 that undid the
    // lean that makes a tetrahedron read as a solid at all. The roll is a roll
    // of the picture now, inside the scene, where it cannot (see `readingPose`).
    const orient = this._roll
      ? 'none'
      : readingPoseTransform(solid.geometry, this._presentedFaceIndex(solid.geometry.faceCount));
    // Which facet the player is meant to read: the one the physics landed once
    // the die has rolled, the selected one before that. It is emphasised in
    // ink; see `.facet.presented`.
    const shown = this._shownFaceIndex(solid.geometry.faceCount);
    return html`
      <div id="stage">
        <div id="inner" class="solid">
          <div id="orient" style="transform:${orient}">
            ${repeat(solid.facets, (facet) => facet.key, (facet) => {
              if (facet.faceIndex < 0) return html`<div class="facet cap" style="${facet.style}"></div>`;
              const content = this._resolveFace(values[facet.faceIndex], usePips);
              return html`<div
                    class="facet${facet.faceIndex === shown ? ' presented' : ''}"
                    style="${facet.style}"
                    data-face-index="${facet.faceIndex}"
                    data-face-value="${values[facet.faceIndex]}"
                    data-face-label="${content.label}"
                  >${this._faceContent(content)}${this._cornerContent(facet, values)}</div>`;
            })}
          </div>
        </div>
      </div>
    `;
  }

  /**
   * What the die announces: the label of the face it is presenting, so a
   * screen reader is told the same thing the facet draws. Absent entirely for
   * a die with no faces, which is what an unconfigured `<boardgame-die>` is.
   */
  private _ariaLabel(interactive: boolean, solid: DieSolid | null): string {
    const base = interactive ? 'Roll die' : 'Die';
    const values = this._faceValues();
    if (values.length === 0) return base;
    const shown = this._shownFaceIndex(values.length);
    return `${base} showing ${this._resolveFace(values[shown], this._usesPips(solid)).label}`;
  }

  /**
   * Why the die cannot be rolled, split into what is worth SHOWING and what is
   * only worth having.
   *
   * Every reason used to be drawn under the die in the error colour, which meant
   * that every roll in pig put red text under a tumbling die for a second and a
   * quarter: the client displays state N while N+1 exists, so the displayed
   * state's legality says the move is not possible (`move-not-possible`), and
   * the gate says an animation is running (`animation-running`). Neither is a
   * failure. `moveActionReasonSeverity` is where that judgement lives now, so
   * every control gets it and not just this one.
   *
   * A quiet reason is still worth having: it stays on the button's `title`,
   * where it is available on hover and to the accessibility tree, and where it
   * cannot flash a line of text on and off under a die that is mid-roll.
   */
  private _statusMessage(): { readonly text: string | null; readonly loud: boolean } {
    const action = this.action;
    if (!isBoundMoveAction(action)) {
      // Not a bound action at all: an authoring mistake, which is exactly the
      // kind of thing an error style is for.
      return { text: action ? 'Bind required move input with .with(...)' : null, loud: true };
    }
    const reason = action.reason;
    if (!reason) return { text: null, loud: false };
    const retryable = action.preview.kind === 'failed' && action.preview.retryable;
    const text = retryable
      ? `${reason.message ?? 'Move legality check failed'}. Activate to retry.`
      : reason.message;
    return { text, loud: moveActionReasonSeverity(reason) === 'error' };
  }

  override render() {
    const action = this.action;
    const bound = isBoundMoveAction(action);
    const interactive = bound;
    const effectiveDisabled = this.disabled || !interactive || (bound && !action.canActivate);
    const status = this._statusMessage();
    const shown = status.loud ? status.text : null;
    const solid = this._solid();
    return html`
      <div id="scaler"
        style=${solid ? `--solid-extent:${num(solidExtent(solid.geometry))}` : nothing}>
        <button
          id="main"
          type="button"
          aria-label=${this._ariaLabel(interactive, solid)}
          aria-describedby=${shown ? 'action-status' : nothing}
          title=${!status.loud && status.text ? status.text : nothing}
          aria-busy=${String(bound && action.submission.kind === 'pending')}
          ?disabled=${effectiveDisabled}
          data-cocked=${this._roll?.cocked ? 'true' : nothing}
          style="--selected-face:${this.selectedFace}"
          class="${this._classes(effectiveDisabled, solid !== null)}">
          ${solid ? this._renderSolid(solid) : this._renderReel()}
        </button>
        ${shown ? html`<span id="action-status" class="interaction-status"
          role="status" aria-live="polite">${shown}</span>` : nothing}
        <span class="visually-hidden" role="status" aria-live="polite">${this._announcement}</span>
      </div>
    `;
  }
}

customElements.define('boardgame-die', BoardgameDie);
