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
import type { ComponentView } from './component-view.js';
import { MAX_TARGET_ACTION_CANDIDATES, type TargetAction } from '../moves/target-action.js';
import type { SourceDestinationBinding } from '../moves/source-destination.js';
import type { PlacementTargetBinding } from '../moves/placement-draft.js';
import type { TargetKey } from '../moves/target-action.js';
import type { ExpandedStack } from '../types/boardgame-types.js';

/** Placement-draft surface consumed by rectangular board destinations. */
export interface GridPlacementDraft {
  readonly targets: readonly number[];
  readonly selectedItem: TargetKey | null;
  target(target: number): PlacementTargetBinding<TargetKey, number>;
}

export interface GameBoardLabelContext {
  readonly index: number;
  readonly row: number;
  readonly col: number;
  readonly occupant: unknown;
}

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

    :host([labels]) .board-surface {
      padding-right: 26px;
      padding-bottom: 24px;
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

    .grid-row {
      display: grid;
      min-height: 0;
    }

    .grid-row > [role='gridcell'] {
      display: flex;
      min-width: 0;
      min-height: 0;
    }

    .cell {
      appearance: none;
      border: 0;
      border-radius: 0;
      background: transparent;
      color: inherit;
      font: inherit;
      margin: 0;
      padding: 0;
      width: 100%;
      height: 100%;
      box-sizing: border-box;
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

    .cell:focus-visible {
      outline: 3px solid var(--md-sys-color-primary, #675f00);
      outline-offset: -3px;
      z-index: 2;
    }

    .interaction-status {
      margin-top: 0.4rem;
      min-height: 1.25em;
      color: var(--md-sys-color-on-surface-variant, #4A4539);
      font-size: 0.875rem;
    }

    .interaction-status:empty {
      display: none;
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

    @media (prefers-reduced-motion: reduce) {
      .cell { transition: none; }
    }

    @media (forced-colors: active) {
      .cell { border: 1px solid CanvasText; }
      .cell.highlighted, .cell.selected { outline: 3px solid Highlight; outline-offset: -3px; }
      .cell.disabled { opacity: 1; }
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
  stack: ExpandedStack<object, object> | null = null;

  /** Typed per-cell move actions. Candidate keys must exactly match cell indexes. */
  @property({ attribute: false })
  action: TargetAction<number> | null = null;

  /** Composed source selection plus a typed destination action. */
  @property({ attribute: false })
  sourceDestination: SourceDestinationBinding<number> | null = null;

  /** Local draft destinations; mutually exclusive with action/sourceDestination. */
  @property({ attribute: false })
  placementDraft: GridPlacementDraft | null = null;

  /** Accessible name for the grid as a whole. */
  @property({ type: String, attribute: 'board-label' })
  boardLabel = 'Game board';

  /** Override the accessible name for a cell. Names must be non-empty and unique. */
  @property({ attribute: false })
  labelFor: ((context: GameBoardLabelContext) => string) | null = null;

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

  /** Explicit escape hatch for untyped child properties. Prefer a bound component view. */
  @property({ type: Object, attribute: false })
  unsafeComponentAttrs: Record<string, unknown> = {};

  /** Renderer-scoped component recipe passed to the board's inner stack. */
  @property({ attribute: false })
  componentView: ComponentView | null = null;

  /** Whether to show coordinate labels (1-8, A-H). */
  @property({ type: Boolean, reflect: true })
  labels = false;

  // Precomputed Sets for O(1) lookup in _cellClass
  private _highlightedSet = new Set<number>();
  private _disabledSet = new Set<number>();
  private _actionUnsubscribe: (() => void) | null = null;
  private _subscribedAction: TargetAction<number> | null = null;
  private _focusedIndex: number | null = null;
  private _hasFocusedCell = false;
  private _cellLabels: readonly string[] = [];

  override disconnectedCallback(): void {
    super.disconnectedCallback();
    this._actionUnsubscribe?.();
    this._actionUnsubscribe = null;
    this._subscribedAction = null;
  }

  override connectedCallback(): void {
    super.connectedCallback();
    this._subscribeToAction();
  }

  protected willUpdate(changedProperties: Map<string, unknown>): void {
    if (changedProperties.has('highlightedSpaces')) {
      this._highlightedSet = new Set(this.highlightedSpaces);
    }
    if (changedProperties.has('disabledSpaces')) {
      this._disabledSet = new Set(this.disabledSpaces);
    }
    this._validateConfiguration();
    if (!this._hasFocusedCell || this._focusedIndex === null || this._focusedIndex >= this._numSpaces) {
      this._focusedIndex = this._initialFocusIndex();
    }
  }

  protected override updated(changedProperties: Map<string, unknown>): void {
    if (!changedProperties.has('action') && !changedProperties.has('sourceDestination')) return;
    this._subscribeToAction();
  }

  private _subscribeToAction(): void {
    const action = this._targetAction;
    if (this._actionUnsubscribe && this._subscribedAction === action) return;
    this._actionUnsubscribe?.();
    this._actionUnsubscribe = null;
    this._subscribedAction = null;
    if (this.isConnected && action) {
      this._actionUnsubscribe = action.subscribe(() => this.requestUpdate());
      this._subscribedAction = action;
    }
  }

  private get _targetAction(): TargetAction<number> | null {
    return this.sourceDestination?.action ?? this.action;
  }

  private get _selectedSpace(): number {
    return this.sourceDestination?.selectedSource ?? this.selectedSpace;
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

    if (this._isDisabled(index)) {
      classes.push('disabled');
    }

    if (this._selectedSpace === index) {
      classes.push('selected');
    }

    if (this.sourceDestination && this._targetAction?.get(index)?.action.canPropose) {
      classes.push('highlighted');
    }

    return classes.join(' ');
  }

  private _isDisabled(index: number): boolean {
    if (this._disabledSet.has(index)) return true;
    const placement = this._placementTarget(index);
    if (placement) return !placement.canPlace;
    const candidate = this._targetAction?.get(index);
    if (candidate) return !candidate.action.canActivate;
    if (this.sourceDestination) return !this.sourceDestination.sources.includes(index);
    return false;
  }

  private _placementTarget(index: number) {
    if (!this.placementDraft || !this.placementDraft.targets.includes(index)) return null;
    return this.placementDraft.target(index);
  }

  private _initialFocusIndex(): number {
    if (this._targetAction?.preview.kind === 'ready') {
      const legal = this._targetAction.candidates.find(candidate => candidate.action.canPropose);
      if (legal) return legal.key;
    }
    return 0;
  }

  private async _onCellClick(index: number): Promise<void> {
    if (this._disabledSet.has(index)) return;
    const placement = this._placementTarget(index);
    if (placement) {
      if (placement.canPlace) placement.place();
      return;
    }
    if (this.placementDraft) return;
    const candidate = this._targetAction?.get(index);
    if (candidate) {
      if (candidate.action.canActivate
        || candidate.action.preview.kind === 'unchecked'
        || candidate.action.preview.kind === 'checking') {
        const result = await candidate.action.activate();
        if (result.kind === 'success') this.sourceDestination?.clear();
      }
      return;
    }
    if (this.sourceDestination) {
      if (this.sourceDestination.sources.includes(index)) {
        this.sourceDestination.selectSource(index);
      }
      return;
    }
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

  private _onCellKeyDown(event: KeyboardEvent, index: number): void {
    if (event.key === 'Escape' && this.sourceDestination && this.sourceDestination.selectedSource !== null) {
      event.preventDefault();
      this.sourceDestination?.clear();
      return;
    }
    let next = index;
    const rowStart = Math.floor(index / this.cols) * this.cols;
    const rowEnd = rowStart + this.cols - 1;
    switch (event.key) {
      case 'ArrowLeft': next = Math.max(rowStart, index - 1); break;
      case 'ArrowRight': next = Math.min(rowEnd, index + 1); break;
      case 'ArrowUp': next = Math.max(0, index - this.cols); break;
      case 'ArrowDown': next = Math.min(this._numSpaces - 1, index + this.cols); break;
      case 'Home': next = event.ctrlKey ? 0 : rowStart; break;
      case 'End': next = event.ctrlKey ? this._numSpaces - 1 : rowEnd; break;
      default: return;
    }
    event.preventDefault();
    this._focusedIndex = next;
    this.requestUpdate();
    void this.updateComplete.then(() => {
      this.renderRoot.querySelector<HTMLButtonElement>(`.cell[data-index="${next}"]`)?.focus();
    });
  }

  private _cellLabel(index: number): string {
    return this._cellLabels[index] ?? this._computeCellLabel(index);
  }

  private _computeCellLabel(index: number): string {
    const row = Math.floor(index / this.cols);
    const col = index % this.cols;
    const occupant = this.stack?.Components[index] ?? null;
    if (this.labelFor) {
      let label: unknown;
      try {
        label = this.labelFor({ index, row, col, occupant });
      } catch (error) {
        const detail = error instanceof Error ? `: ${error.message}` : '';
        throw new Error(`boardgame-game-board labelFor failed for cell ${index}${detail}`);
      }
      if (typeof label !== 'string') {
        throw new Error(`boardgame-game-board labelFor must return a string for cell ${index}`);
      }
      return label.trim();
    }
    return `${this._colLabel(col)}${row + 1}, ${occupant ? 'occupied' : 'empty'}`;
  }

  private _cellReason(index: number): string | null {
    return this._placementTarget(index)?.reason
      ?? this._targetAction?.get(index)?.action.reason?.message ?? null;
  }

  private _accessibleCellLabel(index: number): string {
    let label = this._cellLabel(index);
    if (this.sourceDestination?.sources.includes(index)) {
      label += this.sourceDestination.selectedSource === index
        ? '. Selected source; activate again to cancel'
        : '. Selectable source';
    }
    const reason = this._cellReason(index);
    if (!reason) return label;
    const retryable = !this._disabledSet.has(index)
      && (this._targetAction?.get(index)?.action.canActivate ?? false);
    return `${label}. ${retryable ? 'Retry available' : 'Unavailable'}: ${reason}`;
  }

  private _validateConfiguration(): void {
    if (!Number.isInteger(this.rows) || this.rows <= 0 || !Number.isInteger(this.cols) || this.cols <= 0) {
      throw new Error(`boardgame-game-board rows and cols must be positive integers; received ${this.rows}x${this.cols}`);
    }
    if (!this.boardLabel.trim()) throw new Error('boardgame-game-board board-label must be non-empty');
    if (this._numSpaces > 4096) throw new Error(`boardgame-game-board has ${this._numSpaces} cells; maximum is 4096`);
    const interactions = [this.action, this.sourceDestination, this.placementDraft]
      .filter(interaction => interaction !== null).length;
    if (interactions > 1) {
      throw new Error('boardgame-game-board action, sourceDestination, and placementDraft are mutually exclusive');
    }
    if (this.placementDraft && this.disabledSpaces.length > 0) {
      throw new Error('boardgame-game-board placementDraft and disabledSpaces are mutually exclusive');
    }
    const targetAction = this._targetAction;
    if (targetAction && this._numSpaces > MAX_TARGET_ACTION_CANDIDATES) {
      throw new Error(`boardgame-game-board target actions support at most ${MAX_TARGET_ACTION_CANDIDATES} cells`);
    }
    if (this.stack && !Array.isArray(this.stack.Components)) {
      throw new Error('boardgame-game-board stack.Components must be an array');
    }
    if (this.stack && this.stack.Components.length !== this._numSpaces) {
      throw new Error(`boardgame-game-board expected ${this._numSpaces} stack components but received ${this.stack.Components.length}`);
    }
    if (targetAction) {
      const keys = targetAction.candidates.map(candidate => candidate.key);
      const sorted = [...keys].sort((left, right) => left - right);
      if (!this.sourceDestination && (sorted.length !== this._numSpaces || sorted.some((key, index) => key !== index))) {
        throw new Error(`boardgame-game-board action keys must cover exactly 0 through ${this._numSpaces - 1}`);
      }
      const invalid = sorted.find(key => !Number.isInteger(key) || key < 0 || key >= this._numSpaces);
      if (invalid !== undefined) {
        throw new Error(`boardgame-game-board target key ${invalid} is outside 0 through ${this._numSpaces - 1}`);
      }
    }
    if (this.sourceDestination) {
      const invalidSource = this.sourceDestination.sources.find(
        key => !Number.isInteger(key) || key < 0 || key >= this._numSpaces,
      );
      if (invalidSource !== undefined) {
        throw new Error(`boardgame-game-board source key ${invalidSource} is outside 0 through ${this._numSpaces - 1}`);
      }
    }
    if (this.placementDraft) {
      if (!Array.isArray(this.placementDraft.targets) || typeof this.placementDraft.target !== 'function') {
        throw new Error('boardgame-game-board placementDraft must be a PlacementDraftController binding');
      }
      const sorted = [...this.placementDraft.targets].sort((left, right) => left - right);
      if (sorted.length !== this._numSpaces || sorted.some((key, index) => key !== index)) {
        throw new Error(`boardgame-game-board placementDraft targets must cover exactly 0 through ${this._numSpaces - 1}`);
      }
    }
    const labels = Array.from({ length: this._numSpaces }, (_, index) => this._computeCellLabel(index));
    const empty = labels.findIndex(label => !label);
    if (empty !== -1) throw new Error(`boardgame-game-board labelFor returned an empty label for cell ${empty}`);
    if (new Set(labels).size !== labels.length) throw new Error('boardgame-game-board accessible cell labels must be unique');
    this._cellLabels = Object.freeze(labels);
  }

  private _colLabel(col: number): string {
    let value = col + 1;
    let result = '';
    while (value > 0) {
      value--;
      result = String.fromCharCode(65 + (value % 26)) + result;
      value = Math.floor(value / 26);
    }
    return result;
  }

  render() {
    const colLabels = Array.from({ length: this.cols }, (_, i) => i);
    const rowLabels = Array.from({ length: this.rows }, (_, i) => i);

    return html`
      <div class="board-surface">
        <div class="board-area" style="aspect-ratio: ${this.cols} / ${this.rows}">
          <!-- Cell background layer -->
          <div class="cell-layer" role="grid"
               aria-label=${this.boardLabel}
               aria-rowcount=${this.rows} aria-colcount=${this.cols}
               aria-busy=${this._targetAction?.preview.kind === 'checking' ? 'true' : 'false'}
               style="grid-template-rows: repeat(${this.rows}, 1fr)">
            ${repeat(rowLabels, row => row, row => html`
              <div class="grid-row" role="row" aria-rowindex=${row + 1}
                   style="grid-template-columns: repeat(${this.cols}, 1fr)">
                ${repeat(colLabels, col => col, col => {
                  const i = row * this.cols + col;
                  return html`
                    <div role="gridcell" aria-colindex=${col + 1}
                         aria-selected=${this._selectedSpace === i ? 'true' : 'false'}>
                      <button type="button" class="cell ${this._cellClass(i)}" data-index=${i}
                         tabindex=${this._focusedIndex === i ? 0 : -1}
                         aria-label=${this._accessibleCellLabel(i)}
                         aria-disabled=${this._isDisabled(i) ? 'true' : 'false'}
                         title=${this._cellReason(i) ?? ''}
                         @focus=${() => { this._focusedIndex = i; this._hasFocusedCell = true; }}
                         @keydown=${(event: KeyboardEvent) => this._onCellKeyDown(event, i)}
                         @click=${() => this._onCellClick(i)}>
                      </button>
                    </div>
                  `;
                })}
              </div>
            `)}
          </div>

          <!-- Component layer -->
          <boardgame-component-stack
            aria-hidden="true"
            layout="board"
            .boardCols="${this.cols}"
            .boardRows="${this.rows}"
            .stack="${this.stack}"
            .componentView=${this.componentView}
            .unsafeComponentAttrs="${this.unsafeComponentAttrs}"
            no-default-spacer>
          </boardgame-component-stack>

          <!-- Optional coordinate labels -->
          ${this.labels ? html`
            <div class="labels-row" aria-hidden="true" style="grid-template-columns: repeat(${this.cols}, 1fr)">
              ${repeat(colLabels, (i) => i, (i) => html`
                <div class="label">${this._colLabel(i)}</div>
              `)}
            </div>
            <div class="labels-col" aria-hidden="true" style="grid-template-rows: repeat(${this.rows}, 1fr)">
              ${repeat(rowLabels, (i) => i, (i) => html`
                <div class="label">${i + 1}</div>
              `)}
            </div>
          ` : nothing}
        </div>
      </div>
      ${this._targetAction?.preview.kind === 'failed' ? html`
        <div class="interaction-status" role="status" aria-live="polite">
          ${this._targetAction.preview.reason.message}
        </div>
      ` : nothing}
    `;
  }
}

declare global {
  interface HTMLElementTagNameMap {
    'boardgame-game-board': BoardgameGameBoard;
  }
}
