import { LitElement, html, css, nothing, type TemplateResult } from 'lit';
import { customElement, property, state } from 'lit/decorators.js';
import { isStatStack, statCapacity, statCount, type StatStack } from './stat-values.js';
import type { StatusTextAutoMessage, StatusTextValue } from './boardgame-status-text.js';
import './boardgame-status-text.js';

export type { StatStack } from './stat-values.js';

/**
 * A labelled value. Five of five game renderers hand-built one.
 *
 * The real shapes the audit found, all of which this replaces:
 *
 *   - `${name} · ${p.Score} pts`, assembled by string concatenation
 *   - `Supply<br>${n}`, using a `<br>` to stack the label over the value,
 *     duplicated verbatim in a second renderer in the same game
 *   - `Food ${n}/${m}` and `Fat ${n}/${m}` over a sized stack's own capacity
 *   - four bare `<div><strong>…</strong></div>` rows in one player-info panel
 *   - a `<boardgame-status-text>` followed by a literal `&nbsp;` (or returning
 *     `'\xa0'` from a helper) purely to stop the row collapsing when the value
 *     is empty -- which, MEASURED, it never did. An empty status-text renders
 *     18px tall, exactly as tall as a filled one; a bare empty span renders 0.
 *     The hack was cargo cult, and adopting this element deletes it
 *
 * ## Composed on boardgame-status-text, not built beside it
 *
 * `boardgame-status-text` already renders a scalar, announces changes politely
 * to assistive technology, and plays the fading-text callout when the number
 * moves. A second animating-number element would be exactly the duplication
 * this effort exists to delete, so the value here IS a status-text, and
 * `autoMessage` and `announce` pass straight through to it.
 *
 * ## Capacity comes from the stack, and only from the stack
 *
 * There is deliberately no way to pass a capacity NUMBER. `Food 3/6` is
 * `.stack=${player.Food}` and nothing else: the framework already knows both
 * halves (`stat-values.ts` documents the two traps that makes worth owning),
 * and an author who restates the 6 is an author whose 6 will one day disagree
 * with the stack it describes. Passing a number where the stack goes throws,
 * and says why.
 *
 * ## Escape hatches
 *
 * `icon` and `label` slots, and `icon` / `label` / `value` / `capacity` parts,
 * so a game can replace or restyle any piece without abandoning the component.
 * `--boardgame-stat-*` tokens cover the gap, the sizes and the colours.
 */
@customElement('boardgame-stat')
export class BoardgameStat extends LitElement {
  static override styles = css`
    :host {
      display: inline-flex;
      align-items: baseline;
      /*
       * A word space, not a designer's gap: replacing a hand-written
       * label-then-value pair with a stat must not move the text by even a
       * few pixels, and the thing it replaces was separated by one space.
       */
      gap: var(--boardgame-stat-gap, 0.25em);
    }

    :host([stacked]) {
      flex-direction: column;
      align-items: var(--boardgame-stat-align, center);
      gap: var(--boardgame-stat-gap, 0.1em);
    }

    /* An absent icon or label must not claim a gap of its own. */
    .empty {
      display: none;
    }

    .label {
      color: var(--boardgame-stat-label-color, var(--md-sys-color-on-surface-variant, #4A4539));
      font-size: var(--boardgame-stat-label-size, inherit);
    }

    .value {
      color: var(--boardgame-stat-value-color, inherit);
      font-size: var(--boardgame-stat-value-size, inherit);
      white-space: nowrap;
    }

    .capacity {
      color: var(--boardgame-stat-capacity-color, var(--md-sys-color-on-surface-variant, #4A4539));
    }
  `;

  /** Text or emoji shown before the label. Replace wholesale with the `icon` slot. */
  @property({ type: String })
  icon = '';

  /** What the value is. Replace wholesale with the `label` slot. */
  @property({ type: String })
  label = '';

  /**
   * The displayed value. Leave it unset to let `.stack` supply the count.
   * Property-only, exactly like `boardgame-status-text.value`.
   */
  @property({ attribute: false })
  value: StatusTextValue = undefined;

  /**
   * The stack this stat is about. Supplies the count when `.value` is unset,
   * and ALWAYS supplies the capacity when the stack has one. Never a number.
   */
  @property({ attribute: false })
  stack: StatStack | null = null;

  /** Passed to the underlying status-text; `diff-up` by default, as there. */
  @property({ type: String, attribute: 'auto-message' })
  autoMessage: StatusTextAutoMessage = 'diff-up';

  /** Passed to the underlying status-text. */
  @property({ type: Boolean })
  announce = true;

  /** Label above the value rather than beside it — the `Supply<br>${n}` shape. */
  @property({ type: Boolean, reflect: true })
  stacked = false;

  @state() private _iconSlotted = false;
  @state() private _labelSlotted = false;

  /** The value actually displayed, after the stack fallback. */
  get displayValue(): StatusTextValue {
    if (this.value !== undefined && this.value !== null) return this.value;
    if (this.stack) return statCount(this.stack);
    return this.value ?? null;
  }

  /** The capacity actually displayed, or undefined when the stack has none. */
  get displayCapacity(): number | undefined {
    return statCapacity(this.stack);
  }

  override render(): TemplateResult {
    this.#validateAuthoring();
    const capacity = this.displayCapacity;
    const iconEmpty = !this.icon && !this._iconSlotted;
    const labelEmpty = !this.label && !this._labelSlotted;
    return html`
      <span part="icon" class="icon ${iconEmpty ? 'empty' : ''}"
        ><slot name="icon" @slotchange=${this.#iconSlotChanged}>${this.icon || nothing}</slot
      ></span>
      <span part="label" class="label ${labelEmpty ? 'empty' : ''}"
        ><slot name="label" @slotchange=${this.#labelSlotChanged}>${this.label || nothing}</slot
      ></span>
      <span part="value" class="value"
        ><boardgame-status-text
          .value=${this.displayValue}
          .autoMessage=${this.autoMessage}
          .announce=${this.announce}
        ></boardgame-status-text
        >${capacity === undefined
          ? nothing
          : html`<span part="capacity" class="capacity">/${capacity}</span>`}</span
      >
    `;
  }

  readonly #iconSlotChanged = (event: Event): void => {
    this._iconSlotted = (event.target as HTMLSlotElement).assignedNodes().length > 0;
  };

  readonly #labelSlotChanged = (event: Event): void => {
    this._labelSlotted = (event.target as HTMLSlotElement).assignedNodes().length > 0;
  };

  readonly #validateAuthoring = (): void => {
    if (this.hasAttribute('value') || this.hasAttribute('stack')) {
      throw new Error('boardgame-stat: bind the typed properties with .value=${...} and .stack=${...}; value/stack attributes are not supported');
    }
    if (this.value !== undefined && this.value !== null
      && typeof this.value !== 'string' && typeof this.value !== 'number') {
      throw new Error('boardgame-stat: .value must be a string, number, null, or undefined');
    }
    // The load-bearing guard. `Food 3/6` is `.stack=${player.Food}`; a capacity
    // an author restates as a literal is a capacity that will one day disagree
    // with the stack it describes.
    if (this.stack !== null && this.stack !== undefined && !isStatStack(this.stack)) {
      throw new Error(
        typeof this.stack === 'number'
          ? 'boardgame-stat: .stack takes the stack itself, not a capacity number; the framework reads Size/MaxSize from it'
          : 'boardgame-stat: .stack must be a stack from renderer state',
      );
    }
  };
}

declare global {
  interface HTMLElementTagNameMap {
    'boardgame-stat': BoardgameStat;
  }
}
