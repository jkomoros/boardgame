import { BoardgameAnimatableItem } from './boardgame-animatable-item.js';
import { html, css, nothing } from 'lit';
import { property } from 'lit/decorators.js';
import { query } from 'lit/decorators.js';
import { repeat } from 'lit/directives/repeat.js';
import { isBoundMoveAction, type BoundMoveAction } from '../moves/action.js';
import { componentMotionTracks } from '../motion/component-track.js';
import type { ComponentMotionTarget } from '../motion/component-track.js';
import { dieGeometry, dot, type DieGeometry, type Vec3 } from '../motion/die-geometry.js';
import { assignFaceValues, presentedFaceIndex, resolveReadingRule } from '../motion/die-faces.js';
import { restingTransform, trajectoryCurve } from '../motion/dice-bake.js';
import {
  FRAME_MS,
  dieRollTrajectory,
  sceneTransform,
  settledTrajectory,
} from '../motion/dice-roll.js';
import { cssNumber as num } from '../solid/screen-frame.js';
import { solidFacets, type SolidFacet } from '../solid/facet-placement.js';
import { readingPoseTransform } from '../solid/reading-pose.js';
import {
  CORNER_GLYPH_HEIGHT,
  GLYPH_HEIGHT,
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
function itemRollCount(item: any): number | null {
  const raw = item?.DynamicValues?.RollCount;
  return typeof raw === 'number' && Number.isFinite(raw) ? raw : null;
}

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
        transform: translateY(calc(-1 * var(--reel-step) * var(--selected-face)));
        /* The spin is WAAPI-driven now; no CSS transform transition. */
      }

      #inner.solid {
        position: absolute;
        inset: 0;
        transform-style: preserve-3d;
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
        color: var(--md-sys-color-on-surface, #1C1810);
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
  item: any = null;

  @property({ type: Number })
  value = 0;

  @property({ type: Array })
  faces: number[] = [];

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

  private _itemChanged(newValue: any) {
    if (!newValue) {
      this.faces = [];
      this.selectedFace = 0;
      this.value = 0;
      this._roll = null;
      this._componentId = '';
      this._itemInstalls = 0;
      this._rollCount = null;
      return;
    }
    this.faces = newValue.Values.Faces;
    this.selectedFace = newValue.DynamicValues.SelectedFace;
    this.value = newValue.DynamicValues.Value;
    this._componentId = typeof newValue.ID === 'string' ? newValue.ID : '';
    // A different die entirely: the old roll's assignment and resting pose are
    // for a solid this one no longer is.
    const faceCount = Array.isArray(this.faces) ? this.faces.length : 0;
    if (this._roll && this._roll.faces.length !== faceCount) this._roll = null;
    const previousCount = this._rollCount;
    this._rollCount = itemRollCount(newValue);
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
   * the bake refuses) leaves `_roll` null, and the die falls back to the
   * deterministic presentation pose rather than half-rendering a physics one.
   */
  private _startRoll(): void {
    const roll = this._planRoll();
    this._roll = roll;
    this._pendingRoll = roll;
    // _roll is not a reactive property: nothing re-renders without this.
    this.requestUpdate();
  }

  private _planRoll(): DieRoll | null {
    const solid = this._solid();
    if (!solid) return null;
    const geometry = solid.geometry;
    const faces = this.faces;
    const desired = faces[this._presentedFaceIndex(geometry.faceCount)];
    if (!Number.isFinite(desired)) return null;
    // One circumradius on screen. Read from #stage's font-size because that IS
    // the die's size (the solid is built at 1em across), and it is a NUMBER of
    // pixels: interpolating a CSS variable into the matrix instead would forfeit
    // compositing for the whole tumble.
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
      const scene = sceneTransform(geometry, die, presented, radiusPx);
      const curve = trajectoryCurve(die, durationMs, { radiusPx });
      return {
        faces: assignFaceValues(geometry, faces, presented, desired),
        presented,
        cocked: trajectory.cocked,
        durationMs,
        curve: (progress: number) => `${scene} ${curve(progress)}`,
        // The same prefix in front of the same formatter's output as curve(1),
        // so the two agree BYTE FOR BYTE. Animations run with fill:'none', so
        // the element renders this the instant the tumble finishes -- or is
        // finished early by the cycle sweep -- and a single rounding digit of
        // disagreement would show up as the die twitching as it settles.
        resting: `${scene} ${restingTransform(die, { radiusPx })}`,
      };
    } catch (error) {
      // A geometry the simulator or the bake refuses. Nothing here may throw
      // during a render pass, and a die that quietly stops rolling is worth a
      // line in the console.
      console.warn('boardgame-die: could not plan a roll; showing the value without one', error);
      return null;
    }
  }

  private _playRoll(roll: DieRoll): void {
    this.playMotionTracks(
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
    // Once the die has rolled, its pose is the physics's ENTIRELY: #inner holds
    // the tumble (and, once it finishes, the trajectory's own resting transform,
    // written by the motion-track kernel), so #orient must contribute nothing or
    // the two poses would compose into a third. The presentation pose is what a
    // die that has never rolled is shown in, and only that.
    const orient = this._roll
      ? 'none'
      : readingPoseTransform(solid.geometry, this._presentedFaceIndex(solid.geometry.faceCount));
    return html`
      <div id="stage">
        <div id="inner" class="solid">
          <div id="orient" style="transform:${orient}">
            ${repeat(solid.facets, (facet) => facet.key, (facet) => {
              if (facet.faceIndex < 0) return html`<div class="facet cap" style="${facet.style}"></div>`;
              const content = this._resolveFace(values[facet.faceIndex], usePips);
              return html`<div
                    class="facet"
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
          aria-label=${this._ariaLabel(interactive, solid)}
          aria-describedby=${reason ? 'action-status' : ''}
          aria-busy=${String(bound && action.submission.kind === 'pending')}
          ?disabled=${effectiveDisabled}
          data-cocked=${this._roll?.cocked ? 'true' : nothing}
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
