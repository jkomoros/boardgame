/**
 * boardgame-game-board
 *
 * A visual board component that wraps boardgame-component-stack with
 * layout="board" and provides the board surface styling, checkerboard
 * cells, legal move highlighting, and click handling.
 *
 * Usage:
 * ```html
 * <boardgame-game-board
 *   rows="8" cols="8" checkerboard
 *   .stack="${this.state?.Game?.Spaces}"
 *   .highlightedSpaces="${this._legalMoves}"
 *   @space-tapped="${this._onSpaceTapped}">
 * </boardgame-game-board>
 * ```
 *
 * The component delegates all component rendering to an inner
 * boardgame-component-stack with layout="board". The board wrapper
 * provides two overlaid CSS Grid layers:
 *   1. Cell backgrounds (below) — checkerboard, highlights, click targets
 *   2. Component-stack (above) — the actual game pieces (pointer-events: none
 *      so clicks pass through to the cell layer)
 */
import { LitElement, html, css, nothing } from 'lit';
import { customElement, property } from 'lit/decorators.js';
import { repeat } from 'lit/directives/repeat.js';
import './boardgame-component-stack.js';

@customElement('boardgame-game-board')
export class BoardgameGameBoard extends LitElement {
  static styles = css`
    :host {
      display: block;
    }

    .board-surface {
      background: var(--board-surface, #2D5016);
      border-radius: 8px;
      box-shadow: inset 0 2px 8px rgba(0, 0, 0, 0.3),
                  0 4px 12px rgba(60, 40, 20, 0.2),
                  inset 0 1px 0 rgba(255, 255, 255, 0.05);
      padding: 6px;
      position: relative;
      overflow: hidden;
      box-sizing: border-box;
    }

    .board-area {
      position: relative;
    }

    .cell-layer {
      display: grid;
      position: absolute;
      inset: 0;
      z-index: 0;
    }

    .cell {
      aspect-ratio: 1;
      cursor: pointer;
      transition: filter 0.15s ease, box-shadow 0.15s ease;
      position: relative;
    }

    /* Non-checkerboard: subtle grid lines on the felt */
    :host(:not([checkerboard])) .cell {
      border: 1px solid rgba(255, 255, 255, 0.06);
    }

    /* Checkerboard colors — dark and light felt */
    :host([checkerboard]) .cell.dark {
      background: var(--board-dark-cell, #1B3A0A);
    }

    :host([checkerboard]) .cell.light {
      background: var(--board-light-cell, #3A6820);
    }

    /* Highlighted cells — brass glow for legal moves */
    .cell.highlighted {
      box-shadow: inset 0 0 0 2px var(--md-sys-color-tertiary, #8B7432);
      filter: brightness(1.25);
    }

    .cell.highlighted::after {
      content: '';
      position: absolute;
      inset: 0;
      border-radius: 50%;
      margin: 25%;
      background: var(--md-sys-color-tertiary, #8B7432);
      opacity: 0.3;
    }

    /* Disabled cells */
    .cell.disabled {
      opacity: 0.3;
      cursor: default;
    }

    /* Hover — subtle brightness on interactive cells */
    .cell:hover:not(.disabled) {
      filter: brightness(1.15);
    }

    .cell.highlighted:hover:not(.disabled) {
      filter: brightness(1.35);
    }

    /* Selected cell — strong brass border */
    .cell.selected {
      box-shadow: inset 0 0 0 3px var(--md-sys-color-tertiary, #8B7432);
      filter: brightness(1.3);
    }

    /* Component stack layer — above the cell backgrounds.
       pointer-events: none so all clicks reach the cell layer beneath. */
    boardgame-component-stack {
      position: absolute;
      inset: 0;
      width: auto;
      z-index: 1;
      pointer-events: none;
    }

    /* Coordinate labels */
    .labels-col {
      display: grid;
      position: absolute;
      right: -20px;
      top: 6px;
      bottom: 6px;
      width: 16px;
      font-size: 10px;
      color: var(--md-sys-color-on-surface-variant, #4A4539);
      font-family: var(--md-sys-typescale-label-small-font, 'Source Sans 3', sans-serif);
    }

    .labels-row {
      display: grid;
      position: absolute;
      bottom: -18px;
      left: 6px;
      right: 6px;
      height: 16px;
      font-size: 10px;
      color: var(--md-sys-color-on-surface-variant, #4A4539);
      font-family: var(--md-sys-typescale-label-small-font, 'Source Sans 3', sans-serif);
    }

    .label {
      display: flex;
      align-items: center;
      justify-content: center;
    }
  `;

  /** Number of rows on the board. */
  @property({ type: Number })
  rows = 8;

  /** Number of columns on the board. */
  @property({ type: Number })
  cols = 8;

  /** The SizedStack from game state, passed through to the inner component-stack. */
  @property({ type: Object })
  stack: any = null;

  /** Whether to render alternating dark/light cells. */
  @property({ type: Boolean, reflect: true })
  checkerboard = false;

  /** Array of space indices to highlight (e.g., legal move destinations). */
  @property({ type: Array })
  highlightedSpaces: number[] = [];

  /** Array of space indices to dim/disable. */
  @property({ type: Array })
  disabledSpaces: number[] = [];

  /** Index of the currently selected space (e.g., piece being moved). -1 = none. */
  @property({ type: Number })
  selectedSpace = -1;

  /** Attributes to forward to child components in the stack. */
  @property({ type: Object, attribute: false })
  componentAttrs: Record<string, any> = {};

  /** Whether to show coordinate labels (1-8, A-H). */
  @property({ type: Boolean })
  labels = false;

  // Precomputed Sets for O(1) lookup in _cellClass
  private _highlightedSet = new Set<number>();
  private _disabledSet = new Set<number>();

  protected willUpdate(changedProperties: Map<string, unknown>): void {
    if (changedProperties.has('highlightedSpaces')) {
      this._highlightedSet = new Set(this.highlightedSpaces);
    }
    if (changedProperties.has('disabledSpaces')) {
      this._disabledSet = new Set(this.disabledSpaces);
    }
  }

  private get _numSpaces(): number {
    return this.rows * this.cols;
  }

  private _cellClass(index: number): string {
    const row = Math.floor(index / this.cols);
    const col = index % this.cols;
    const classes: string[] = [];

    // Checkerboard pattern: (0,0) is dark (standard chess/checkers convention)
    if ((row + col) % 2 === 0) {
      classes.push('dark');
    } else {
      classes.push('light');
    }

    if (this._highlightedSet.has(index)) {
      classes.push('highlighted');
    }

    if (this._disabledSet.has(index)) {
      classes.push('disabled');
    }

    if (this.selectedSpace === index) {
      classes.push('selected');
    }

    return classes.join(' ');
  }

  private _onCellClick(index: number) {
    if (this._disabledSet.has(index)) return;

    this.dispatchEvent(new CustomEvent('space-tapped', {
      composed: true,
      bubbles: true,
      detail: {
        index,
        row: Math.floor(index / this.cols),
        col: index % this.cols,
      }
    }));
  }

  private _colLabel(col: number): string {
    return String.fromCharCode(65 + col); // A, B, C, ...
  }

  render() {
    const cells = Array.from({ length: this._numSpaces }, (_, i) => i);
    const colLabels = Array.from({ length: this.cols }, (_, i) => i);
    const rowLabels = Array.from({ length: this.rows }, (_, i) => i);

    return html`
      <div class="board-surface">
        <div class="board-area" style="aspect-ratio: ${this.cols} / ${this.rows}">
          <!-- Cell background layer -->
          <div class="cell-layer" style="grid-template-columns: repeat(${this.cols}, 1fr)">
            ${repeat(cells, (i) => i, (i) => html`
              <div class="cell ${this._cellClass(i)}"
                   @click="${() => this._onCellClick(i)}">
              </div>
            `)}
          </div>

          <!-- Component layer -->
          <boardgame-component-stack
            layout="board"
            .boardCols="${this.cols}"
            .boardRows="${this.rows}"
            .stack="${this.stack}"
            .componentAttrs="${this.componentAttrs}"
            no-default-spacer>
          </boardgame-component-stack>

          <!-- Optional coordinate labels -->
          ${this.labels ? html`
            <div class="labels-row" style="grid-template-columns: repeat(${this.cols}, 1fr)">
              ${repeat(colLabels, (i) => i, (i) => html`
                <div class="label">${this._colLabel(i)}</div>
              `)}
            </div>
            <div class="labels-col" style="grid-template-rows: repeat(${this.rows}, 1fr)">
              ${repeat(rowLabels, (i) => i, (i) => html`
                <div class="label">${this.rows - i}</div>
              `)}
            </div>
          ` : nothing}
        </div>
      </div>
    `;
  }
}

declare global {
  interface HTMLElementTagNameMap {
    'boardgame-game-board': BoardgameGameBoard;
  }
}
