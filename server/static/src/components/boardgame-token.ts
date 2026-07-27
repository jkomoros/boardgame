import { BoardgameComponent } from './boardgame-component.js';
import { html, css, CSSResult, TemplateResult } from 'lit';
import { customElement, property } from 'lit/decorators.js';
import { classMap } from 'lit/directives/class-map.js';
import { motionSilhouette } from '../motion/subject.js';
import type { MotionSubjectSnapshot } from '../motion/subject.js';

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

      #outer.gray img {
        filter: saturate(0.0) brightness(3.0);
      }

      #outer.green img {
        filter: hue-rotate(130deg) brightness(2.0);
      }

      #outer.teal img {
        filter: hue-rotate(185deg) brightness(2.4);
      }

      #outer.purple img {
        filter: hue-rotate(300deg) brightness(1.0);
      }

      #outer.pink img {
        filter: hue-rotate(-93deg) brightness(4) saturate(0.8);
      }

      /* red is the default color, no need for shifting */

      #outer.blue img {
        filter: hue-rotate(220deg) brightness(2.0) saturate(1.5);
      }

      #outer.orange img {
        filter: hue-rotate(50deg) brightness(2.5);
      }

      #outer.yellow img {
        filter: hue-rotate(70deg) brightness(4);
      }

      #outer.black img {
        filter: saturate(0.0) brightness(1.7);
      }
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
  protected override _computeClasses(): Record<string, boolean> {
    const result = super._computeClasses();
    return {
      ...result,
      [this.color.toLowerCase()]: true,
      active: this.active,
      highlighted: this.highlighted,
      [this.type]: true
    };
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
    // forever. Safe on first connect -- innerElement is null pre-render, so
    // _syncThrob no-ops, and the first render's updated() starts it as before
    // (see the DOM-reparent test in token-throb.spec.ts).
    this._syncThrob();
  }

  override disconnectedCallback(): void {
    this._throb?.cancel();
    this._throb = null;
    super.disconnectedCallback();
  }

  override render(): TemplateResult {
    const asset = this._computeAsset(this.type);
    return html`
      <div id="outer" class="${classMap(this._computeClasses())}" @click="${(e: Event) => this.handleTap(e)}" style="${this._outerStyle}">
        <div id="inner">
          <img src="${asset}">
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
