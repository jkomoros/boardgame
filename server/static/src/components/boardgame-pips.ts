import { LitElement, css, html, nothing, type TemplateResult } from 'lit';
import { customElement, property } from 'lit/decorators.js';

/**
 * A count rendered as repeated symbols — the thing printed on a card rather
 * than written next to it.
 *
 * `murdermrmonroe` builds one in a loop on every render:
 *
 * ```ts
 * private _prettyLuck(numLuck: number): string {
 *   let result = "";
 *   for (let i = 0; i < numLuck; i++) result += "☘";
 *   return result;
 * }
 * ```
 *
 * and `darwin` writes `☀ ${card.Suns} · ❄ ${card.Snowflakes}` because it had no
 * way to draw two suns and one snowflake, which is what its cards actually
 * depict.
 *
 * ## Why this is not a `<boardgame-stat>`
 *
 * A stat is a labelled NUMBER, and its value is a `boardgame-status-text` that
 * announces and animates when the number moves. Pips are the same count drawn
 * as a quantity you read at a glance without counting — three clovers, two suns
 * — and at four or five symbols that is genuinely faster than a numeral. They
 * are different renderings of the same fact, not a component and a variant of
 * it, and a stat can slot one of these where its value goes if a game wants
 * both.
 *
 * ## Screen readers get the number, never the symbols
 *
 * `☘☘☘` announced literally is "shamrock shamrock shamrock", and with the
 * emoji-presentation glyphs a game is likely to reach for it is worse. The
 * glyph row is `aria-hidden` decoration and the accessible name is the count —
 * or `label`, which is what makes it *3 Luck* rather than *3*.
 *
 * ## Escape hatches
 *
 * `part="pip"` and `part="empty"` reach every symbol from the game's own
 * stylesheet; `--boardgame-pips-size`, `-gap`, `-color` and `-empty-color`
 * cover the usual four. A game that needs a picture rather than a character
 * sets `glyph` to any string, including an emoji sequence.
 */
@customElement('boardgame-pips')
export class BoardgamePips extends LitElement {
  static override styles = css`
    :host {
      display: inline-flex;
      align-items: center;
      gap: var(--boardgame-pips-gap, 0.08em);
      font-size: var(--boardgame-pips-size, inherit);
      line-height: 1;
    }

    /* A suppressed pip row must not claim a flex gap from the row it sits in,
       which an empty 'display: inline-flex' host still would -- so the host is
       removed from layout rather than merely emptied. The attribute is written
       in willUpdate() so this is a plain CSS consequence of the state. */
    :host([empty]) {
      display: none;
    }

    .pip {
      color: var(--boardgame-pips-color, inherit);
    }

    /* The unfilled remainder of a capacity: present, so that "2 of 5" reads as
       a partly-filled row rather than as a shorter one, and quiet, so it does
       not read as a filled pip. */
    .empty {
      color: var(--boardgame-pips-empty-color, var(--md-sys-color-outline-variant, #CCC4B8));
    }

    .sr-only {
      position: absolute;
      width: 1px;
      height: 1px;
      padding: 0;
      margin: -1px;
      overflow: hidden;
      clip: rect(0, 0, 0, 0);
      white-space: nowrap;
      border: 0;
    }
  `;

  /** The symbol repeated `count` times. Any string: a character, or an emoji sequence. */
  @property({ type: String })
  glyph = '●';

  /** How many filled pips to draw. Fractions and negatives are floored and clamped. */
  @property({ type: Number })
  count = 0;

  /**
   * A capacity. When set, the row is always `max` symbols long and the
   * remainder is drawn in the empty colour — the difference between "3 food"
   * and "3 of the 5 this species can hold".
   */
  @property({ type: Number })
  max: number | null = null;

  /** The symbol used for the unfilled remainder. Defaults to `glyph`. */
  @property({ type: String, attribute: 'empty-glyph' })
  emptyGlyph = '';

  /** What the count is OF, for assistive technology. `3` becomes `3 Luck`. */
  @property({ type: String })
  label = '';

  /**
   * Render nothing at all when the count is zero.
   *
   * The `Room:`-style row `murdermrmonroe` suppresses by returning `''` from a
   * `_valueOrNothing` helper. Opt-in, because a zero is meaningful for most
   * counts and silently vanishing would be worse than showing it.
   */
  @property({ type: Boolean, attribute: 'hide-when-zero' })
  hideWhenZero = false;

  /** Whether this row is currently suppressed by `hide-when-zero`. */
  get suppressed(): boolean {
    return this.hideWhenZero && this.filled === 0 && this.unfilled === 0;
  }

  protected override willUpdate(): void {
    this.toggleAttribute('empty', this.suppressed);
  }

  /** The filled-pip count actually drawn, after flooring and clamping. */
  get filled(): number {
    if (!Number.isFinite(this.count)) return 0;
    const floored = Math.floor(this.count);
    if (floored <= 0) return 0;
    return this.max !== null && Number.isFinite(this.max)
      ? Math.min(floored, Math.max(0, Math.floor(this.max)))
      : floored;
  }

  /** The unfilled remainder actually drawn. Zero unless `max` is set. */
  get unfilled(): number {
    if (this.max === null || !Number.isFinite(this.max)) return 0;
    return Math.max(0, Math.floor(this.max) - this.filled);
  }

  override render(): TemplateResult | typeof nothing {
    this.#validateAuthoring();
    const filled = this.filled;
    const unfilled = this.unfilled;
    if (this.suppressed) return nothing;
    const empty = this.emptyGlyph || this.glyph;
    const spoken = this.label ? `${this.count} ${this.label}` : String(this.count);
    return html`<span class="sr-only">${spoken}</span
      ><span aria-hidden="true"
        >${Array.from({ length: filled }, () => html`<span part="pip" class="pip">${this.glyph}</span>`)}${
          Array.from({ length: unfilled }, () => html`<span part="empty" class="empty">${empty}</span>`)}</span
      >`;
  }

  readonly #validateAuthoring = (): void => {
    if (this.glyph === '') {
      throw new Error('boardgame-pips: glyph must be a non-empty symbol');
    }
    if (this.max !== null && (typeof this.max !== 'number' || !Number.isFinite(this.max) || this.max < 0)) {
      throw new Error('boardgame-pips: max must be a non-negative number or null');
    }
  };
}

declare global {
  interface HTMLElementTagNameMap {
    'boardgame-pips': BoardgamePips;
  }
}
