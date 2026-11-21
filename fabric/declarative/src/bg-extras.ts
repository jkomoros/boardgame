/**
 * The small elements a board needs besides stacks. Each is a thin wrapper that
 * binds an existing element to the table by PATH, so its own attributes never
 * change from one state to the next.
 *
 *   <bg-callout>      boardgame-fading-text, triggered by a value in the state
 *   <bg-turn-status>  boardgame-turn-status, fed from the installed view
 *   <bg-action>       a button that proposes a move with fixed arguments
 *   <bg-seat>         a seat container that reflects player facts as attributes
 *   <bg-value>        boardgame-status-text, bound to a number or string in the state
 */
import { LitElement, html, css } from 'lit';
import { property } from 'lit/decorators.js';
import '../../../server/static/src/components/boardgame-fading-text.js';
import '../../../server/static/src/components/boardgame-turn-status.js';
import '../../../server/static/src/components/boardgame-status-text.js';
import { findTable } from './bg-table.js';
import type { BgTable, TableChild } from './bg-table.js';
import type { ProposeMoveDetail } from './bg-stack.js';

type Scalar = string | number | boolean | null;

abstract class TableBound extends LitElement implements TableChild {
  protected table: BgTable | null = null;

  override connectedCallback(): void {
    super.connectedCallback();
    this.table = findTable(this.parentNode ?? this);
    this.table?.register(this);
  }

  override disconnectedCallback(): void {
    super.disconnectedCallback();
    this.table?.unregister(this);
    this.table = null;
  }

  tableChanged(): void {
    this.requestUpdate();
  }
}

/**
 * A brief callout ("Match", "Busted!") that plays when a value changes.
 *
 * `boardgame-fading-text` is already declarative, but it decides whether to
 * play by comparing against the previous trigger it saw, which it keeps in a
 * private field. A recreated element has no previous value and never plays.
 * So the previous value is kept by the table, under this element's `name`.
 */
export class BgCallout extends TableBound {
  static override styles = css`
    :host { position: absolute; inset: 0; pointer-events: none; }
  `;

  @property({ type: String }) message = '';
  /** Path of the trigger value in the table's state. `.$count` and `.$length` are allowed. */
  @property({ type: String }) path = '';
  /** A trigger given directly, for callers that do not bind by path. */
  @property({ attribute: false }) trigger: Scalar = null;
  @property({ type: String, attribute: 'auto-message' }) autoMessage = 'fixed';
  @property({ type: String }) suppress = 'none';
  @property({ type: String }) name = '';

  private _shown: Scalar = null;
  private _primed = false;

  protected override willUpdate(): void {
    const raw = this.path ? this.table?.resolve(this.path) : this.trigger;
    const current = (raw === undefined || (raw !== null && typeof raw === 'object')
      ? null : raw) as Scalar;
    const key = `callout:${this.name || this.path}`;
    if (!this._primed && this.table) {
      this._primed = true;
      const previous = this.table.recall(key) as Scalar | undefined;
      if (previous !== undefined && previous !== current) {
        // Replay the remembered value first so the inner element sees a change.
        this._shown = previous;
        void this.updateComplete.then(() => {
          this._shown = current;
          this.requestUpdate();
        });
        this.table.remember(key, current);
        return;
      }
    }
    this._shown = current;
    this.table?.remember(key, current);
  }

  override render() {
    return html`<boardgame-fading-text
      .message=${this.message}
      .trigger=${this._shown}
      .autoMessage=${this.autoMessage as never}
      .suppress=${this.suppress as never}></boardgame-fading-text>`;
  }
}
customElements.define('bg-callout', BgCallout);

/** Whose turn it is, from the installed view. Hidden while motion is in flight. */
export class BgTurnStatus extends TableBound {
  /** Names for seats, as a JSON array attribute or property. */
  @property({ type: Array, attribute: 'player-labels' }) playerLabels: string[] = [];

  override render() {
    const view = this.table?.view;
    if (!view) return null;
    const turn = {
      currentPlayerIndex: view.CurrentPlayerIndex ?? -1,
      viewerPlayerIndex: this.table!.viewer,
      finished: !!view.Finished,
      animating: false,
    };
    return html`<boardgame-turn-status .turn=${turn} .playerLabels=${this.playerLabels}>
    </boardgame-turn-status>`;
  }
}
customElements.define('bg-turn-status', BgTurnStatus);

/**
 * A seat's container. It reflects facts about one player as attributes, so the
 * page styles them with ordinary CSS and nothing recomputes a class per state:
 *
 *   <bg-seat index="1" flags="Eliminated Stood">  ->  [active] [eliminated] [stood]
 */
export class BgSeat extends TableBound {
  static override styles = css`:host { display: block; position: relative; }`;

  @property({ type: Number }) index = 0;
  /** Boolean fields of `Players[index]` to reflect, space separated. */
  @property({ type: String }) flags = '';

  protected override willUpdate(): void {
    const view = this.table?.view;
    this.toggleAttribute('active', !!view && view.CurrentPlayerIndex === this.index);
    this.toggleAttribute('viewer', !!this.table && this.table.viewer === this.index);
    for (const flag of this.flags.split(/\s+/).filter(Boolean)) {
      if (!/^[A-Za-z][A-Za-z0-9]*$/.test(flag)) continue;
      this.toggleAttribute(flag.toLowerCase(),
        this.table?.resolve(`Players.${this.index}.${flag}`) === true);
    }
  }

  override render() {
    return html`<slot></slot>`;
  }
}
customElements.define('bg-seat', BgSeat);

/** One number or string from the state, with the standard "+3" change callout. */
export class BgValue extends TableBound {
  static override styles = css`:host { display: inline-block; }`;

  @property({ type: String }) path = '';

  override render() {
    const raw = this.table?.resolve(this.path);
    const value = typeof raw === 'number' || typeof raw === 'string' ? raw : null;
    return html`<boardgame-status-text .value=${value}></boardgame-status-text>`;
  }
}
customElements.define('bg-value', BgValue);

/** A button that proposes one move with constant arguments. */
export class BgAction extends TableBound {
  static override styles = css`
    button { font: inherit; padding: 0.4em 0.9em; border-radius: 999px; cursor: pointer; }
    button:disabled { cursor: default; opacity: 0.5; }
  `;

  @property({ type: String }) move = '';
  @property({ type: Object }) args: Record<string, unknown> = {};
  @property({ type: Boolean }) disabled = false;

  /** Disabled by attribute, or because the installed frame says the move is not legal. */
  private get _disabled(): boolean {
    return this.disabled || this.table?.legal(this.move) === false;
  }

  private readonly _click = (event: Event): void => {
    if (!event.isTrusted || this._disabled || !this.move) return;
    const detail: ProposeMoveDetail = Object.freeze({
      move: this.move,
      args: { ...this.args },
      version: this.table?.view?.Version ?? null,
    });
    this.dispatchEvent(new CustomEvent('propose-move', { detail, bubbles: true, composed: true }));
  };

  override render() {
    return html`<button part="button" ?disabled=${this._disabled} @click=${this._click}>
      <slot></slot></button>`;
  }
}
customElements.define('bg-action', BgAction);

declare global {
  interface HTMLElementTagNameMap {
    'bg-callout': BgCallout;
    'bg-turn-status': BgTurnStatus;
    'bg-action': BgAction;
    'bg-seat': BgSeat;
    'bg-value': BgValue;
  }
}
