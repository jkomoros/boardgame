import { clearHistoricalPresentation, type HistoricalAppearance } from '../motion/historical-presentation.js';
import { BoardgameComponent } from './boardgame-component.js';
import { html, css, nothing, TemplateResult } from 'lit';
import { property, query, state } from 'lit/decorators.js';
import { classMap } from 'lit/directives/class-map.js';
import { artLayerStyle, isArtFit, type ArtFit } from './component-art.js';
import { motionSilhouette } from '../motion/subject.js';
import type { MotionSubjectSnapshot } from '../motion/subject.js';
import type { VisualMotionTrackInput } from '../motion/component-track.js';

interface CardHistoricalAppearance extends HistoricalAppearance {
  suit: string; rank: string; tall: boolean; aspectRatio: number;
  art: string; artFit: ArtFit; backArt: string;
  frontColor: string; inkColor: string; faceFontScale: number;
  faceShadow: string; footerColor: string; footerFontScale: number;
  noShadow: boolean; altShadow: boolean;
}

export class BoardgameCard extends BoardgameComponent {
  static override styles = [
    BoardgameComponent.styles,
    css`
      :host {
        /* Override component width for cards */
        --default-component-width: 100px;

        /* The DRAWN card's height/width, matching the aspectRatio property's
           default. _computeOuterStyle republishes the property's live value on
           #outer, which overrides this; the declaration is here so the very
           first paint -- before updated() has computed that inline style --
           still has a ratio to draw with. Deliberately NOT
           --component-aspect-ratio: that names the BOX's ratio and only works
           from :host or above. See the #inner rules. */
        --card-aspect-ratio: 0.6666666;

        /* Shadow elevation styles for rotated cards */
        --shadow-elevation-normal-rotated: 2px 0 2px 0 rgba(60, 40, 20, 0.14),
                                            1px 0 5px 0 rgba(60, 40, 20, 0.12),
                                            3px 0 1px -2px rgba(60, 40, 20, 0.2);

        --shadow-elevation-raised-rotated: 8px 0 10px 1px rgba(60, 40, 20, 0.14),
                                            3px 0 14px 2px rgba(60, 40, 20, 0.12),
                                            5px 0 5px -3px rgba(60, 40, 20, 0.4);

        --alt-shadow-elevation-normal-rotated: drop-shadow(2px 0 2px rgba(60, 40, 20, 0.14))
                                                drop-shadow(1px 0 5px rgba(60, 40, 20, 0.12))
                                                drop-shadow(3px 0 1px rgba(60, 40, 20, 0.2));

        --alt-shadow-elevation-raised-rotated: drop-shadow(8px 0 10px rgba(60, 40, 20, 0.14))
                                                drop-shadow(3px 0 14px rgba(60, 40, 20, 0.12))
                                                drop-shadow(5px 0 5px rgba(60, 40, 20, 0.4));
      }

      #outer {
        --card-effective-border-radius: 5px;
      }

      #outer div.fallback {
        display: none;
      }

      /* Historical motion carriers render cloned authored content through the
         fallback slot. Keep the shared face itself so its frame and other
         deck-wide skin travel with that content; only suppress the live face
         regions whose state belongs to the fresh carrier host. */
      #outer.no-content:not(.structured-history) #face.normal > * {
        display: none;
      }

      slot.historical-presentation { display: none; }
      #outer.structured-history slot.live-presentation { display: none; }
      #outer.structured-history slot.historical-presentation { display: contents; }

      #outer.no-content div.fallback {
        display: block;
      }

      #front {
        display: flex;
        flex-direction: column;
        align-items: center;
        justify-content: center;
      }

      #outer {
        height: var(--component-effective-height);
        width: var(--component-effective-width);
        display: flex;
        flex-direction: column;
        align-items: center;
        justify-content: center;
        perspective: 1000px;
      }

      #outer.tall {
        height: var(--component-effective-width);
        width: var(--component-effective-height);
      }

      #outer.rotated {
        height: var(--component-effective-width);
        width: var(--component-effective-height);
      }

      #outer.tall.rotated {
        height: var(--component-effective-height);
        width: var(--component-effective-width);
      }

      /*
       * THE CARD'S RATIO IS THE ART'S, NOT THE BOX'S, AND IT HAS ITS OWN NAME.
       *
       * --card-aspect-ratio used to be published as --component-aspect-ratio,
       * written inline on #outer from the aspectRatio property. Two things were
       * wrong with that, and only the second was doing any work.
       *
       * It never shaped the BOX. #outer's height comes from
       * --component-effective-height, and a custom property that REFERENCES
       * another is substituted where it is DECLARED --
       * --component-effective-height is declared at :host
       * (boardgame-component.ts), above #outer, so it was always substituted
       * with the :host ratio of 1.0 no matter what #outer said. Measured at
       * --component-width: 200px, #outer computed
       * --component-aspect-ratio: 0.6666666 and drew a 200x200 box; real
       * blackjack cards measured 105x105 hosts around a 103x71 card. This is
       * the same trap boardgame-token.ts describes at length for its deleted
       * per-shape rules.
       *
       * What it DID shape is this rule, which reads the ratio directly rather
       * than through --component-effective-height and so sees the value #outer
       * inherits down. That is the job worth keeping -- and the reason the
       * declaration was renamed instead of deleted the way the token's were.
       *
       * Under the old name it also SHADOWED the one place
       * --component-aspect-ratio is meant to be set and does work: :host or
       * above. A card handed --component-aspect-ratio: 1.5 on its host drew a
       * 200x300 box around a 100x66.7 card, box and art disagreeing, because
       * the inline write on #outer overwrote the author's value for everything
       * inside it. Under two names each value keeps its own job.
       *
       * The box itself stays square, for the reasons boardgame-token.ts spells
       * out: the board layout puts aspect-ratio: 1 on every component host,
       * boardgame-spatial-board's tokenPosition centres a piece at
       * coords - tokenSize / 2 in BOTH axes, and the stack's spread/fan margins
       * and the FLIP scale ratio all key off that one box. The card is already
       * drawn in true proportion inside it, exactly as a token's SVG is. See
       * tests/animations/parity/card-box.spec.ts.
       */
      #inner {
        width: var(--default-component-width);
        height: calc(var(--default-component-width) * var(--card-aspect-ratio));
        transform: scale(var(--component-effective-scale));
        border-radius: var(--card-effective-border-radius);
        transform-style: preserve-3d;
        position: absolute;
      }

      .tall #inner {
        height: var(--default-component-width);
        width: calc(var(--default-component-width) * var(--card-aspect-ratio));
      }

      #outer.shadow.rotated #inner {
        box-shadow: var(--shadow-elevation-normal-rotated);
      }

      #outer.shadow.interactive.rotated:hover #inner {
        box-shadow: var(--shadow-elevation-raised-rotated);
      }

      /* The rotated alt-shadow elevation, ON #outer -- never on #inner, for
         exactly the reason boardgame-component.ts spells out for the unrotated
         pair and boardgame-token.ts for its throb: motionTrackTarget('visual')
         returns #inner, so #inner is where a component-owned 3D scene mounts
         and where 'transform-style: preserve-3d' has to go, and a 'filter'
         forces 'transform-style: flat' on the element carrying it. A rotated
         card with altShadow set would therefore have been unable to host one.

         Nothing sets 'altShadow' on a card today -- not here, not in ../games
         -- so this pair was unreachable and the flattening was latent rather
         than live. That is the argument for moving it NOW: it costs nothing
         while nothing depends on its stacking, and #inner is exactly where a
         3D card would have to live.

         Visually inert for the same reason as the unrotated pair: #outer paints
         nothing of its own, so the alpha silhouette the drop-shadows derive
         from is the one #inner produced either way.

         .disabled's saturate is restated for the same reason
         boardgame-component.ts restates it: the two now share ONE filter slot
         on #outer, and this selector outranks '#outer.alt-shadow.disabled', so
         a disabled rotated card would otherwise silently stop looking
         disabled. Elevation first, then saturate -- the order the two-element
         version painted them. */
      #outer.alt-shadow.rotated {
        filter: var(--alt-shadow-elevation-normal-rotated);
      }

      #outer.alt-shadow.rotated.disabled {
        filter: var(--alt-shadow-elevation-normal-rotated) saturate(60%);
      }

      #outer.alt-shadow.interactive.rotated:hover {
        filter: var(--alt-shadow-elevation-raised-rotated);
      }

      #front,
      #back {
        height: 100%;
        width: 100%;
        position: absolute;
        top: 0;
        left: 0;
        backface-visibility: hidden;
        -webkit-backface-visibility: hidden;
        overflow: hidden;
        border-radius: var(--card-effective-border-radius);
      }

      /*
       * THE FACE, which is the half of a card this component never had.
       *
       * boardgame-card shipped the chrome -- face-up, rotated, the aspect
       * ratio, the back slot -- and a bare centred slot for the face. Three
       * games built the rest by hand, three different ways, and every one of
       * them rediscovered THIS box: something that fills the card, that content
       * can be pinned to the top, the middle and the bottom of.
       *
       *   - a button.card with 'grid-template-rows: auto auto 1fr', a
       *     radial-gradient corner pip, and an art band bled past the padding
       *     by 'width: calc(100% + 1.1rem); margin: -.4rem -.55rem 0'
       *   - a whole separate LitElement whose root rule is
       *     'position: absolute; height: 100%; width: 100%; top: 0; left: 0'
       *     followed by a 'flex: 1' middle and a 'flex-direction: row' bottom
       *   - '.card-face { width: 100%; height: 100%; display: grid }' with an
       *     'inset 0 0 0 4px' box-shadow standing in for a printed frame
       *
       * ## Why the padding is on the REGIONS and not here
       *
       * #top-rank, #bottom-rank and #center-rank are absolutely positioned, and
       * their containing block is the nearest positioned ancestor -- which, now
       * that this element exists, is #face rather than #front. Padding here
       * would move all three, and the classic card must render
       * pixel-identically to what it rendered before the face regions existed.
       * 'inset: 0' with no padding keeps that box exactly where it was; #center
       * and #footer carry the inset instead, and they contain only slotted
       * content. The regions are 'position: static' for the same reason.
       *
       * ## The rows
       *
       * Art, centre, footer, and every one of them names its row explicitly.
       *
       * AUTO-PLACEMENT IS WRONG HERE, and silently: a 'display: none' art band
       * is not a grid item at all, so auto-placement slides the centre up into
       * row 1 -- the 'auto' row -- and the card's whole middle collapses to the
       * height of its text. Measured: a 67px face with a 25px centre floating
       * at the top of it. Naming the rows is what makes the empty cases lay out
       * like the full one.
       *
       * The art row is a PERCENTAGE of the face, so it has to be a track and
       * not a height on the item: a percentage height on a grid item resolves
       * against its grid AREA, which in an 'auto' row is indefinite, so it
       * computed to zero and took the art with it. The track resolves against
       * the grid container, which has a definite height because '#face' is
       * absolutely positioned with 'inset: 0'.
       *
       * The middle row is 'minmax(0, 1fr)' rather than '1fr' because a grid
       * row's automatic minimum is its content's min-content size, so a long
       * trait name in the centre would otherwise push the footer off a 100px
       * card instead of wrapping inside it.
       *
       * The corner is out of flow on purpose: a cost or a rank in the corner of
       * a real card overlaps the art, and a corner that consumed a row would
       * shorten the art band for every card that has one.
       */
      #face {
        position: absolute;
        inset: 0;
        display: grid;
        grid-template-rows: 0 minmax(0, 1fr) auto;
        /* Proportional to the card's DRAWN width, which is
           --default-component-width: #inner is exactly that wide and is then
           scaled as a whole by --component-effective-scale. Sizing from
           --component-effective-width instead would apply the scale twice. */
        font-size: var(--card-face-font-size, calc(var(--default-component-width) * 0.13));
        line-height: 1.15;
        color: var(--card-ink-color, #1c2b22);
        text-align: center;
        min-width: 0;
        /*
         * The printed frame, as a TOKEN rather than only as ::part(face).
         *
         * ::part is the obvious escape hatch and it does not reach here: a
         * stack builds its component hosts inside its OWN shadow root, so a
         * game's 'boardgame-card::part(face)' rule has no card in scope to
         * match -- measured on sequenceforge, whose inset gold frame computed
         * to 'none' on every card. A custom property inherits through both
         * boundaries, so this is the hatch that actually works from a renderer.
         */
        box-shadow: var(--card-face-shadow, none);
      }

      /*
       * The art band, and the reason a game no longer needs negative margins.
       *
       * It is a grid row of #face, which has no padding, so the art already
       * runs to all three card edges it touches. The bleed the games hand-rolled
       * was only ever compensating for padding they had put on the face
       * themselves.
       */
      #face.has-art {
        grid-template-rows: var(--card-art-height, 45%) minmax(0, 1fr) auto;
      }

      #art-band {
        grid-row: 1;
        overflow: hidden;
        min-width: 0;
        min-height: 0;
      }

      /* An absent band must occupy no row at all -- not a zero-height one,
         which would still show a seam against a tinted centre. */
      #art-band.empty {
        display: none;
      }

      #art-image {
        height: 100%;
        width: 100%;
      }

      /* A game that slots its own <img slot="art"> gets the same box model
         without writing it: fill the band, crop rather than squash. */
      #art-band ::slotted(*) {
        display: block;
        height: 100%;
        width: 100%;
        object-fit: var(--card-art-fit, cover);
      }

      #center {
        grid-row: 2;
        display: flex;
        flex-direction: column;
        /*
         * STRETCH, not center, and the centring is #face's text-align instead.
         *
         * A flex column with align-items: center sizes every item to its own
         * content, so a slotted element WIDER than the card is simply wider
         * than the card and gets clipped by #front's overflow -- measured on a
         * murdermrmonroe card, a boardgame-stat reading "Room Winter Garden"
         * could not wrap because it was never told how much room it had.
         * Stretched, each item is the card's width and wraps inside it, and
         * text-align keeps short content looking exactly as centred as before.
         */
        align-items: stretch;
        justify-content: center;
        gap: var(--card-face-gap, 0.15em);
        padding: var(--card-face-padding, 0.4em);
        box-sizing: border-box;
        min-width: 0;
        min-height: 0;
        /* A 100px-wide card is narrower than plenty of single words a game
           will legitimately put on one -- "Cooperation" measured 8px wider
           than the card. Without this the word is simply clipped by #front's
           overflow, which reads as a rendering bug rather than as a long
           name. break-word rather than anywhere: only a word that cannot fit
           at all is broken. */
        overflow-wrap: break-word;
      }

      /* Slotted content may not be wider than the card.
         #center is a flex COLUMN with align-items: center, so a flex item's
         cross size is its own content width and nothing clamps it -- a
         boardgame-stat reading "Room Winter Garden" measured 168px inside a
         100px card and was clipped on both sides by #front's overflow, which
         reads as a rendering bug rather than as a long value. With the clamp it
         wraps instead. */
      #center ::slotted(*),
      #footer ::slotted(*),
      #corner ::slotted(*) {
        max-width: 100%;
        box-sizing: border-box;
      }

      #footer {
        grid-row: 3;
        padding: 0 var(--card-face-padding, 0.4em) var(--card-face-padding, 0.4em);
        box-sizing: border-box;
        font-size: var(--card-footer-font-size, 0.8em);
        color: var(--card-footer-color, var(--card-ink-color, #1c2b22));
        min-width: 0;
      }

      #corner {
        position: absolute;
        top: var(--card-face-padding, 0.4em);
        right: var(--card-face-padding, 0.4em);
        font-size: var(--card-corner-font-size, 0.95em);
        font-weight: 700;
        line-height: 1;
        z-index: 2;
      }

      #footer.empty,
      #corner.empty {
        display: none;
      }

      #back-art {
        position: absolute;
        inset: 0;
      }

      #top-rank,
      #bottom-rank {
        position: absolute;
        /* Scale corner indices with the card so they stay readable at any
           size (--component-width comes from the surrounding view). The
           old fixed 12px was illegible on phone-sized cards. */
        font-size: max(11px, calc(var(--component-effective-width) * 0.17));
        line-height: 1;
        font-weight: 600;
        color: var(--card-ink-color, #1c2b22);
      }

      #top-rank {
        bottom: 5px;
        left: 5px;
        transform: rotate(-90deg);
      }

      #bottom-rank {
        right: 5px;
        top: 5px;
        transform: rotate(90deg);
      }

      /* Big center pip so a hand card reads at arm's length. Rotated 90°
         for the same reason the corner indices are: the card's natural
         frame is landscape; views display it upright via the rotated
         attribute. */
      #center-rank {
        position: absolute;
        inset: 0;
        display: flex;
        align-items: center;
        justify-content: center;
        font-size: calc(var(--component-effective-width) * 0.48);
        line-height: 1;
        font-weight: 700;
        color: var(--card-ink-color, #1c2b22);
        transform: rotate(90deg);
      }

      /* Classic red suits: ♥/♦ fronts carry the .red-suit class. */
      .red-suit #top-rank,
      .red-suit #bottom-rank,
      .red-suit #center-rank {
        color: var(--card-red-ink-color, #B3362B);
      }

      #outer #front {
        background-color: var(--card-front-color, #D4E8DA);
        box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.5);
        z-index: 2;
        transform: rotateY(180deg);
      }

      #outer #back {
        background-color: var(--card-back-color, #2E6B4F);
        color: var(--card-back-text-color, rgba(255, 255, 255, 0.6));
        box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.15),
                    inset 0 -1px 0 rgba(0, 0, 0, 0.1);
        transform: rotateY(0deg);
      }

      #default-back {
        height: 100%;
        width: 120%;
        opacity: 0.2;
        font-size: 13.5px;
        line-height: 14px;
        overflow: hidden;
        text-overflow: clip;
        user-select: none;
        display: flex;
        flex-direction: column;
        align-items: center;
        justify-content: center;
      }

      .tall #default-back {
        width: 130%;
      }
    `
  ];

  @property({ type: String })
  suit = '';

  @property({ type: String })
  rank = '';

  @property({ type: Boolean, attribute: 'face-up' })
  faceUp = false;

  @property({ type: Boolean, reflect: true })
  rotated = false;

  @property({ type: Boolean, attribute: 'no-content' })
  noContent = false;

  @property({ type: Boolean })
  tall = false;

  @property({ type: Number, attribute: 'aspect-ratio' })
  aspectRatio = 0.6666666;

  /**
   * Art for the card FACE, drawn as a band across the top of it.
   *
   * The common case is one attribute and nothing else. A card whose art should
   * cover the whole face sets `--card-art-height: 100%`; a card that wants a
   * different picture per component slots its own `<img slot="art">` and gets
   * the same box model without writing it.
   */
  @property({ type: String })
  art = '';

  /** How `art` and any slotted art fills the band. `cover` crops, `contain` letterboxes. */
  @property({ type: String, attribute: 'art-fit' })
  artFit: ArtFit = 'cover';

  /**
   * Art for the card BACK.
   *
   * The back already had a `<slot name="back">`, and every game that wanted a
   * printed back still had to hand-write an element and a `background`
   * shorthand into it -- `darwin` renders a bare `<div class="deck-back">` with
   * its own width, aspect-ratio, radius, border and
   * `url('./assets/card-back.png') center / cover` and no card at all. The slot
   * stays; this is its default content, so the common case is an attribute and
   * the uncommon case is exactly as reachable as it was.
   */
  @property({ type: String, attribute: 'back-art' })
  backArt = '';

  /**
   * This card's face colour, when it is a fact about the COMPONENT rather than
   * about the deck.
   *
   * `--card-front-color` has always existed, and it has always been unreachable
   * from where the answer lives: a component view's `properties` callback can
   * set typed host properties but not custom properties, and a game's own
   * stylesheet cannot see one card's values. So `sequenceforge` renders a
   * `.card-face` div filling the whole slot purely to have something it can
   * paint, and switches its background between a cream fill and a radial
   * gradient on `Values.Value === 0`.
   *
   * Empty means "whatever the stylesheet says", so a deck with one face colour
   * keeps setting it in CSS exactly as before.
   */
  @property({ type: String, attribute: 'front-color' })
  frontColor = '';

  /** This card's ink colour, for the same reason and reaching `--card-ink-color`. */
  @property({ type: String, attribute: 'ink-color' })
  inkColor = '';

  /** @internal cardView opts into the shared, component-owned face snapshot. */
  structuredHistoricalPresentation = false;

  @state() private _historicalAppearance: CardHistoricalAppearance | null = null;
  @state() private _historyArtSlotted = false;
  @state() private _historyFooterSlotted = false;
  @state() private _historyCornerSlotted = false;

  private get _activeHistory(): CardHistoricalAppearance | null {
    return this.noContent ? this._historicalAppearance : null;
  }

  @query('#front-slot')
  private frontSlot!: HTMLSlotElement;

  @state() private _artSlotted = false;
  @state() private _footerSlotted = false;
  @state() private _cornerSlotted = false;

  override motionSubjectSnapshot(): MotionSubjectSnapshot {
    // Shape only: card face/back/content never crosses this boundary.
    return motionSilhouette('rounded-rectangle');
  }

  private _boundFrontChanged?: () => void;

  override connectedCallback() {
    super.connectedCallback();
    this._updateInnerTransform();
  }

  // Optimization opportunity: shouldUpdate() could be added here to prevent
  // unnecessary re-renders during animations. However, cards have complex state
  // with multiple interdependent properties (faceUp, rotated, noContent, tall,
  // aspectRatio) that affect visual output.
  // Conservative approach: Allow all renders to ensure correctness.
  // Future optimization: Skip renders when only non-visual properties change.

  protected override updated(changedProperties: Map<string, any>) {
    super.updated(changedProperties);

    if (changedProperties.has('faceUp') || changedProperties.has('rotated')) {
      this._updateInnerTransform();
    }

    if (changedProperties.has('rotated')) {
      this._rotatedChanged(this.rotated);
    }

    if (changedProperties.has('aspectRatio')) {
      this._outerStyle = this._computeOuterStyle(this.aspectRatio);
    }
  }

  override firstUpdated(_changedProperties: Map<PropertyKey, unknown>) {
    super.firstUpdated(_changedProperties);
    this._frontChanged();
    this._boundFrontChanged = () => this._frontChanged();
    if (this.frontSlot) {
      this.frontSlot.addEventListener('slotchange', this._boundFrontChanged);
    }
  }

  override disconnectedCallback() {
    super.disconnectedCallback();
    if (this.structuredHistoricalPresentation) clearHistoricalPresentation(this);
    if (this._boundFrontChanged && this.frontSlot) {
      this.frontSlot.removeEventListener('slotchange', this._boundFrontChanged);
    }
  }

  /**
   * Publishes the ART's ratio, under a name that is only the art's. See the
   * `#inner` rules for why this is not `--component-aspect-ratio`: that one
   * names the BOX's ratio, is only readable from `:host` or above, and this
   * used to overwrite it.
   */
  private _computeOuterStyle(aspectRatio: number): string {
    return `--card-aspect-ratio: ${aspectRatio};`;
  }

  override prepareMotionCarrier(
    defaults: Readonly<Record<string, unknown>>,
    stack?: any,
  ): void {
    if (this.prepareForBeingAnimatingComponent
      !== BoardgameCard.prototype.prepareForBeingAnimatingComponent) {
      this.prepareForBeingAnimatingComponent(stack);
      return;
    }
    this.noContent = true;
    this.rotated = !!defaults.rotated;
  }

  /** @deprecated Compatibility adapter for pre-motion component callers. */
  override prepareForBeingAnimatingComponent(stack: any): void {
    this.noContent = true;
    this.rotated = !!stack?.stackDefault?.('rotated');
  }

  override get animatingProperties(): string[] {
    return super.animatingProperties.concat(['rotated', 'faceUp']);
  }

  // _innerTransformFor computes the resting inner transform for a given
  // faceUp/rotated combination — the pure function behind what
  // _updateInnerTransform writes as the resting style.
  private _innerTransformFor(faceUp: boolean, rotated: boolean): string {
    return [
      'scale(var(--component-effective-scale))',
      faceUp ? 'rotateY(180deg)' : 'rotateY(0deg)',
      rotated ? 'rotate(90deg)' : 'rotate(0deg)',
    ].join(' ');
  }

  protected override propertyMotionTracks(
    before: Record<string, any>,
    after: Record<string, any>,
  ): readonly VisualMotionTrackInput[] {
    if (before.faceUp === after.faceUp && before.rotated === after.rotated) return [];
    return [{
      target: 'visual',
      property: 'transform',
      from: this._innerTransformFor(!!before.faceUp, !!before.rotated),
      to: this._innerTransformFor(!!after.faceUp, !!after.rotated),
    }];
  }

  /** @deprecated Compatibility adapter; framework playback uses planned tracks. */
  override playPropertyAnimation(
    before: Record<string, any>,
    after: Record<string, any>,
    delayMs: number = 0,
  ): void {
    if (before.faceUp === after.faceUp && before.rotated === after.rotated) return;
    if (!this.innerElement) return;
    this.play(this.innerElement, [
      { transform: this._innerTransformFor(!!before.faceUp, !!before.rotated) },
      { transform: this._innerTransformFor(!!after.faceUp, !!after.rotated) },
    ], { delay: delayMs });
  }

  protected override shouldPlayLegacyPropertyAnimation(): boolean {
    return this.playPropertyAnimation !== BoardgameCard.prototype.playPropertyAnimation;
  }

  override get historicalPresentationPolicy(): 'none' | 'clone-default-slot' | 'clone-default-slot-safe' {
    if (!this.cloneContent) return 'none';
    return this.structuredHistoricalPresentation ? 'clone-default-slot-safe' : 'clone-default-slot';
  }

  override get historicalPresentationSlots(): Readonly<Record<string, string>> | null {
    return this.structuredHistoricalPresentation ? {
      '': 'motion-history-center', art: 'motion-history-art',
      footer: 'motion-history-footer', corner: 'motion-history-corner',
      back: 'motion-history-back',
    } : null;
  }

  override captureHistoricalAppearance(): CardHistoricalAppearance | null {
    if (!this.structuredHistoricalPresentation || this.noContent) return null;
    const face = this.renderRoot.querySelector<HTMLElement>('#face');
    const front = this.renderRoot.querySelector<HTMLElement>('#front');
    const footer = this.renderRoot.querySelector<HTMLElement>('#footer');
    if (!face || !front || !footer) return null;
    const faceStyle = getComputedStyle(face);
    const inner = this.renderRoot.querySelector<HTMLElement>('#inner');
    const innerStyle = inner ? getComputedStyle(inner) : null;
    const basis = parseFloat((this.tall ? innerStyle?.height : innerStyle?.width) ?? '') || 100;
    const fontSize = parseFloat(faceStyle.fontSize) || 13;
    const footerStyle = getComputedStyle(footer);
    return Object.freeze({
      suit: this.suit, rank: this.rank, tall: this.tall, aspectRatio: this.aspectRatio,
      art: this.art, artFit: this.artFit, backArt: this.backArt,
      frontColor: getComputedStyle(front).backgroundColor, inkColor: faceStyle.color,
      faceFontScale: fontSize / basis, faceShadow: faceStyle.boxShadow,
      footerColor: footerStyle.color,
      footerFontScale: (parseFloat(footerStyle.fontSize) || fontSize * 0.8) / fontSize,
      noShadow: this.noShadow, altShadow: this.altShadow,
    });
  }

  override installHistoricalAppearance(appearance: HistoricalAppearance): () => void {
    const captured = appearance as CardHistoricalAppearance;
    this._historicalAppearance = captured;
    return () => {
      if (this._historicalAppearance === captured) this._historicalAppearance = null;
    };
  }

  /** @deprecated Compatibility adapter for pre-motion component callers. */
  override get cloneContent(): boolean {
    return !this.noContent;
  }

  override motionEndpointOrientation(
    state: Readonly<Record<string, unknown>>,
  ): 'natural' | 'quarter-turned' {
    return state.rotated ? 'quarter-turned' : 'natural';
  }

  /** @deprecated Compatibility adapter for pre-motion geometry callers. */
  override animationRotates(
    beforeProps: Record<string, any>,
    afterProps: Record<string, any>,
  ): boolean {
    return beforeProps.rotated !== afterProps.rotated;
  }

  /** Legacy subclass rotation policy is authoritative when overridden. */
  override legacyAnimationRotationRequested(
    beforeProps: Record<string, any>,
    afterProps: Record<string, any>,
  ): boolean | null {
    if (this.animationRotates === BoardgameCard.prototype.animationRotates) return null;
    return this.animationRotates(beforeProps, afterProps);
  }

  /**
   * Whether the LAST scan of the face slot found content claiming `tall`.
   *
   * The whole of the difference between "the content decides the card's shape"
   * and "the content may decide the card's shape". See `_frontChanged`.
   */
  private _tallFromContent = false;

  private _frontChanged() {
    if (!this.frontSlot) return;

    const nodes = this.frontSlot.assignedNodes();
    let newValue = false;
    for (let i = 0; i < nodes.length; i++) {
      const node = nodes[i];
      if (node.nodeType !== 1) continue;
      const element = node as Element;
      if (element.hasAttribute('tall')) {
        newValue = true;
      }
      if (element.hasAttribute('aspect-ratio')) {
        this.aspectRatio = parseFloat(element.getAttribute('aspect-ratio') || '0.6666666');
      }
    }
    /*
     * AN EMPTY SLOT MAKES NO CLAIM ABOUT THE CARD'S SHAPE.
     *
     * This used to be an unconditional `this.tall = newValue`, which meant the
     * `tall` PROPERTY could not be set by anybody: `firstUpdated` calls this
     * once before any content exists, so an authored `<boardgame-card tall>` or
     * a `cardView` `properties` callback returning `{ tall: true }` was
     * overwritten with `false` on the first render, every time, silently. The
     * only way to get a portrait card was to hang a `tall` attribute on the
     * element you slotted into the face -- which `debuganimations` does, and
     * which is a strange enough channel that three games with portrait cards
     * built the whole face by hand instead.
     *
     * The scan still wins whenever it has ever had something to say, so
     * `debuganimations` is bit-for-bit unchanged: its face content declares
     * `tall`, so the first scan finds it, and when the card flips face-down and
     * the content goes away the scan clears it exactly as before. What changes
     * is only the case where content NEVER declared it, where the scan now says
     * nothing rather than saying `false`.
     */
    if (newValue || this._tallFromContent) this.tall = newValue;
    this._tallFromContent = newValue;
  }

  private _rotatedChanged(_newValue: boolean) {
    this._updateInnerTransform();
  }

  private _updateInnerTransform() {
    if (!this.innerElement) return;
    this.innerElement.style.transform =
      this._innerTransformFor(this.faceUp, this.rotated) || 'none';
  }

  protected override _itemChanged(newValue: any) {
    if (newValue === undefined) return;
    if (newValue === null) {
      this.noContent = true;
      this.faceUp = false;
      super._itemChanged(newValue);
      return;
    }
    if (newValue.Values) {
      this.faceUp = true;
      this.noContent = false;
    } else {
      this.faceUp = false;
      this.noContent = true;
    }
    super._itemChanged(newValue);
  }

  // Override _computeClasses and add some more.
  protected override _computeClasses(): Record<string, boolean> {
    return {
      ...super._computeClasses(),
      card: true,
      shadow: !(this._activeHistory?.noShadow ?? this.noShadow) && !(this._activeHistory?.altShadow ?? this.altShadow),
      'alt-shadow': !(this._activeHistory?.noShadow ?? this.noShadow) && (this._activeHistory?.altShadow ?? this.altShadow),
      rotated: this.rotated,
      'no-content': this.noContent,
      tall: this._activeHistory?.tall ?? this.tall,
      wide: !(this._activeHistory?.tall ?? this.tall),
      'structured-history': !!this._activeHistory
    };
  }

  private _slotTracker(assign: (populated: boolean) => void) {
    return (event: Event): void => {
      const slot = event.target;
      if (!(slot instanceof HTMLSlotElement)) return;
      assign(slot.assignedNodes({ flatten: true })
        .some(node => node.nodeType !== Node.TEXT_NODE || (node.textContent ?? '').trim() !== ''));
    };
  }

  /**
   * The per-card colour overrides, written on `#outer` so they INHERIT down to
   * `#front` and to the corner indices rather than being set on either.
   *
   * Inheritance is the point: `--card-ink-color` is read by three separate
   * rules and by whatever a game slots into the face, and a custom property
   * declared on an ancestor is visible to all of them at once.
   */
  private _colorStyle(): string {
    let style = '';
    const history = this._activeHistory;
    const front = history?.frontColor ?? this.frontColor;
    const ink = history?.inkColor ?? this.inkColor;
    if (front) style += `--card-front-color: ${front};`;
    if (ink) style += `--card-ink-color: ${ink};`;
    if (history) {
      style += `--card-aspect-ratio: ${history.aspectRatio};`;
      style += `--card-face-font-size: calc(var(--default-component-width) * ${history.faceFontScale});`;
      style += `--card-face-shadow: ${history.faceShadow};`;
      style += `--card-footer-color: ${history.footerColor};`;
      style += `--card-footer-font-size: ${history.footerFontScale}em;`;
    }
    return style;
  }

  private _validateArt(): void {
    if (!isArtFit(this.artFit)) {
      throw new Error(`boardgame-card: art-fit must be "cover" or "contain", not ${JSON.stringify(this.artFit)}`);
    }
  }

  private _renderRanks(suit: string, rank: string): TemplateResult {
    return html`<div id="top-rank">${suit}${rank}</div>
      <div id="center-rank">${suit}</div>
      <div id="bottom-rank">${suit}${rank}</div>`;
  }

  override render(): TemplateResult {
    this._validateArt();
    const history = this._activeHistory;
    const art = history?.art ?? this.art;
    const artFit = history?.artFit ?? this.artFit;
    const backArt = history?.backArt ?? this.backArt;
    const suit = history?.suit ?? this.suit;
    const rank = history?.rank ?? this.rank;
    const hasArt = !!art || (history ? this._historyArtSlotted : this._artSlotted);
    const hasFooter = history ? this._historyFooterSlotted : this._footerSlotted;
    const hasCorner = history ? this._historyCornerSlotted : this._cornerSlotted;
    return html`
      <div id="outer" class="${classMap(this._computeClasses())}" @click="${this.handleTap}" style="${this._outerStyle}${this._colorStyle()}">
        <div id="inner">
          <div id="front">
            <div id="face" part="face" class="normal ${hasArt ? 'has-art' : ''} ${suit === '♥' || suit === '♦' ? 'red-suit' : ''}">
              <div id="art-band" part="art" class="${hasArt ? '' : 'empty'}">
                <slot class="live-presentation" name="art" @slotchange=${this._slotTracker(v => { this._artSlotted = v; })}
                  >${art
                    ? html`<div id="art-image" style="${artLayerStyle(art, artFit)}"></div>`
                    : nothing}</slot>
                <slot class="historical-presentation" name="motion-history-art"
                  @slotchange=${this._slotTracker(v => { this._historyArtSlotted = v; })}
                  >${history?.art ? html`<div style="${artLayerStyle(history.art, history.artFit)}; height: 100%; width: 100%"></div>` : nothing}</slot>
              </div>
              <div id="center" part="center">
                <slot class="live-presentation" id="front-slot">${!history ? this._renderRanks(suit, rank) : nothing}</slot>
                <slot class="historical-presentation" name="motion-history-center">${history ? this._renderRanks(suit, rank) : nothing}</slot>
              </div>
              <div id="footer" part="footer" class="${hasFooter ? '' : 'empty'}">
                <slot class="live-presentation" name="footer"
                  @slotchange=${this._slotTracker(v => { this._footerSlotted = v; })}></slot>
                <slot class="historical-presentation" name="motion-history-footer"
                  @slotchange=${this._slotTracker(v => { this._historyFooterSlotted = v; })}></slot>
              </div>
              <div id="corner" part="corner" class="${hasCorner ? '' : 'empty'}">
                <slot class="live-presentation" name="corner"
                  @slotchange=${this._slotTracker(v => { this._cornerSlotted = v; })}></slot>
                <slot class="historical-presentation" name="motion-history-corner"
                  @slotchange=${this._slotTracker(v => { this._historyCornerSlotted = v; })}></slot>
              </div>
            </div>
            <div class="fallback">
              <slot name="motion-history"><slot name="fallback"></slot></slot>
            </div>
          </div>
          <div id="back">
            <slot name="${history ? 'motion-history-back' : 'back'}">
              ${backArt
                ? html`<div id="back-art" part="back-art"
                    style="${artLayerStyle(backArt, artFit)}"></div>`
                : html`<div id="default-back">
                ★ ☆ ★ ☆ ★ ☆ ★ ☆ ★ ☆ ★ ☆ ★ ☆ ★ ☆ ★ ☆ ★ ☆ ★ ☆ ★ ☆ ★ ☆ ★ ☆ ★ ☆ ★ ☆ ★ ☆ ★ ☆
              </div>`}
            </slot>
          </div>
        </div>
      </div>
    `;
  }
}

customElements.define('boardgame-card', BoardgameCard);

declare global {
  interface HTMLElementTagNameMap {
    'boardgame-card': BoardgameCard;
  }
}
