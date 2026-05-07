import { LitElement, html, css } from 'lit';
import { property, query, state } from 'lit/decorators.js';
import { repeat } from 'lit/directives/repeat.js';
import './boardgame-component-stack.js';

/**
 * A reusable spatial board component that loads an SVG map and renders
 * game tokens at SVG space positions automatically.
 *
 * The SVG must contain elements with IDs following the pattern
 * `{spacePrefix}{index}` (e.g., "Room-0", "Room-1", ...).
 *
 * Auto-rendering: pass `.stack` (single SizedStack) or `.stacks` (array
 * of SizedStacks) and the board renders tokens from boardgame-deck-defaults
 * at each space's SVG position. SizedStack slot index = SVG space index.
 *
 * Manual mode (backward compat): omit `.stack`/`.stacks` and use
 * `boxForSpace()`, `tokenPosition()` for manual positioning.
 *
 * Example (auto-rendering):
 * ```html
 * <boardgame-spatial-board
 *   svgUrl="game-src/mygame/board.svg"
 *   spacePrefix="Room-"
 *   .stacks="${[...playerPositions, npcPosition]}"
 *   @space-tapped="${this._onSpaceTapped}">
 * </boardgame-spatial-board>
 * ```
 */
class BoardgameSpatialBoard extends LitElement {

  static override styles = css`
    :host {
      display: block;
    }

    #board-wrapper {
      background: var(--board-surface, #2D5016);
      border-radius: 8px;
      box-shadow: inset 0 2px 8px rgba(0, 0, 0, 0.3),
                  0 4px 12px rgba(60, 40, 20, 0.2),
                  inset 0 1px 0 rgba(255, 255, 255, 0.05);
      padding: 8px;
      position: relative;
    }

    #container {
      position: relative;
    }

    #container svg {
      display: block;
      width: 100%;
      height: auto;
    }

    [data-space] {
      cursor: pointer;
      transition: filter 0.15s ease;
    }

    [data-space]:hover {
      filter: brightness(1.15);
    }

    [data-space].disabled {
      fill: var(--md-sys-color-surface-container-highest, #E0D9CE) !important;
      cursor: default !important;
      filter: none !important;
    }

    #token-overlay {
      position: absolute;
      inset: 0;
      pointer-events: none;
      overflow: hidden;
    }

    #token-overlay boardgame-component-stack {
      position: absolute;
      inset: 0;
    }
  `;

  /** URL to fetch the SVG board from. */
  @property({ type: String })
  svgUrl = '';

  /** ID prefix for space elements in the SVG (e.g., "Space-" matches "Space-0"). */
  @property({ type: String })
  spacePrefix = 'Space-';

  /** Array of space indices that should be visually disabled. */
  @property({ type: Array })
  disabledSpaces: number[] = [];

  /** Single SizedStack for auto-rendering (slot index = space index). */
  @property({ type: Object, attribute: false })
  stack: any = null;

  /** Multiple SizedStacks for multi-token auto-rendering. */
  @property({ type: Array, attribute: false })
  stacks: any[] = [];

  /** Size of token elements in pixels. */
  @property({ type: Number })
  tokenSize = 24;

  /** Attributes to forward to child components. */
  @property({ type: Object, attribute: false })
  componentAttrs: Record<string, any> = {};

  /** True after the SVG has been loaded and inserted into the DOM. */
  @property({ type: Boolean })
  svgLoaded = false;

  /** Computed pixel positions per layer. Updated on state/resize changes. */
  @state()
  private _layerPositions: Array<Array<{ top: number; left: number } | null>> = [];

  @query('#container')
  private _container!: HTMLDivElement;

  private _resizeObserver: ResizeObserver | null = null;

  override connectedCallback() {
    super.connectedCallback();
    this._resizeObserver = new ResizeObserver(() => {
      if (this.svgLoaded) this._recalculatePositions();
    });
  }

  override disconnectedCallback() {
    super.disconnectedCallback();
    this._resizeObserver?.disconnect();
    this._resizeObserver = null;
  }

  override async firstUpdated(_changedProperties: Map<PropertyKey, unknown>) {
    super.firstUpdated(_changedProperties);
    if (this.svgUrl) {
      await this._loadSvg();
    }
    if (this._container) {
      this._resizeObserver?.observe(this._container);
    }
  }

  override updated(changedProperties: Map<PropertyKey, unknown>) {
    super.updated(changedProperties);

    if (changedProperties.has('svgUrl') && this.svgUrl && changedProperties.get('svgUrl') !== undefined) {
      this._loadSvg();
    }

    if (changedProperties.has('disabledSpaces') && this.svgLoaded) {
      this._applyDisabledSpaces();
    }

    if ((changedProperties.has('stack') || changedProperties.has('stacks') || changedProperties.has('tokenSize')) && this.svgLoaded) {
      this._recalculatePositions();
    }
  }

  // ---- SVG loading ----

  private async _loadSvg() {
    try {
      const response = await fetch(this.svgUrl);
      const text = await response.text();
      const parser = new DOMParser();
      const doc = parser.parseFromString(text, 'image/svg+xml');
      const svgElement = doc.documentElement;

      // Clear previous SVG — only remove SVG children, keep overlay
      const existing = this._container.querySelector('svg');
      if (existing) this._container.removeChild(existing);

      // Insert SVG as first child (before overlay)
      this._container.insertBefore(svgElement, this._container.firstChild);

      // Mark space elements for CSS targeting
      const spaces = svgElement.querySelectorAll(`[id^="${this.spacePrefix}"]`);
      spaces.forEach(el => el.setAttribute('data-space', ''));

      this._applyDisabledSpaces();
      this.svgLoaded = true;
      this._recalculatePositions();
      this.dispatchEvent(new CustomEvent('svg-loaded-changed', {
        composed: true,
        detail: { value: true }
      }));
    } catch (err) {
      console.error('boardgame-spatial-board: Failed to load SVG:', err);
    }
  }

  // ---- Disabled spaces ----

  private _applyDisabledSpaces() {
    if (!this.shadowRoot) return;
    const allSpaces = this.shadowRoot.querySelectorAll('[data-space]');
    allSpaces.forEach(el => el.classList.remove('disabled'));
    for (const index of this.disabledSpaces) {
      const el = this.shadowRoot.querySelector(`#${this.spacePrefix}${index}`);
      if (el) el.classList.add('disabled');
    }
  }

  // ---- Click handling ----

  private _spaceTapped(e: Event) {
    const target = e.target as HTMLElement;
    const id = target.id;
    if (!id.startsWith(this.spacePrefix)) return;
    const index = parseInt(id.substring(this.spacePrefix.length));
    if (isNaN(index)) return;
    if (this.disabledSpaces.includes(index)) return;

    this.dispatchEvent(new CustomEvent('space-tapped', {
      composed: true,
      bubbles: true,
      detail: { index }
    }));
  }

  // ---- Coordinate conversion ----

  /**
   * Convert SVG user-space coordinates to pixel coordinates relative
   * to the #container div. Uses getScreenCTM() for accurate mapping
   * that respects viewBox, preserveAspectRatio, and CSS sizing.
   */
  private _svgToPixel(svgX: number, svgY: number): { x: number; y: number } {
    const svg = this._container?.querySelector('svg') as SVGSVGElement | null;
    if (!svg) return { x: 0, y: 0 };

    const ctm = svg.getScreenCTM();
    if (!ctm) return { x: 0, y: 0 };

    const point = new DOMPoint(svgX, svgY).matrixTransform(ctm);
    const containerRect = this._container.getBoundingClientRect();
    return {
      x: point.x - containerRect.left,
      y: point.y - containerRect.top,
    };
  }

  // ---- Spatial queries (public, backward compat) ----

  boxForSpace(index: number): { x: number; y: number; width: number; height: number } {
    const result = { x: 0, y: 0, width: 0, height: 0 };
    if (!this.shadowRoot) return result;
    const space = this.shadowRoot.querySelector(`#${this.spacePrefix}${index}`) as SVGGraphicsElement | null;
    if (!space) return result;
    const box = space.getBBox();
    return { x: box.x, y: box.y, width: box.width, height: box.height };
  }

  coordinatesForSpace(index: number): [number, number] {
    const box = this.boxForSpace(index);
    return [box.x + box.width / 2, box.y + box.height / 2];
  }

  tokenPosition(spaceIndex: number, tokenIndex: number, tokenSize: number): { top: number; left: number } | undefined {
    if (!this.svgLoaded) return undefined;
    const box = this.boxForSpace(spaceIndex);
    if (box.width === 0 && box.height === 0) return undefined;
    const coords = this.coordinatesForSpace(spaceIndex);
    const jitterX = this._deterministicJitter(spaceIndex, tokenIndex, 0) * tokenSize * 2;
    const jitterY = this._deterministicJitter(spaceIndex, tokenIndex, 1) * tokenSize * 2;
    let top = coords[1] - tokenSize / 2 + jitterY;
    let left = coords[0] - tokenSize / 2 + jitterX;
    if (top + tokenSize > box.y + box.height) top = box.y + box.height - tokenSize;
    if (top < box.y) top = box.y;
    if (left + tokenSize > box.x + box.width) left = box.x + box.width - tokenSize;
    if (left < box.x) left = box.x;
    return { top, left };
  }

  private _deterministicJitter(spaceIndex: number, tokenIndex: number, axis: number): number {
    let hash = ((spaceIndex * 31 + tokenIndex) * 37 + axis) * 41;
    hash = ((hash >>> 16) ^ hash) * 0x45d9f3b;
    hash = ((hash >>> 16) ^ hash) * 0x45d9f3b;
    hash = (hash >>> 16) ^ hash;
    return ((hash & 0xFFFF) / 0x7FFF) - 1;
  }

  // ---- Auto-rendering position computation ----

  private get _effectiveStacks(): any[] {
    if (this.stack) return [this.stack];
    return this.stacks;
  }

  /**
   * Recalculate pixel positions for all stacks/layers.
   * Called when SVG loads, state changes, or container resizes.
   */
  private _recalculatePositions() {
    const stacks = this._effectiveStacks;
    if (stacks.length === 0) {
      this._layerPositions = [];
      return;
    }

    const result: Array<Array<{ top: number; left: number } | null>> = [];

    for (let layerIndex = 0; layerIndex < stacks.length; layerIndex++) {
      const stack = stacks[layerIndex];
      const numSlots = stack?.Components?.length || 0;
      const positions: Array<{ top: number; left: number } | null> = [];

      for (let slotIndex = 0; slotIndex < numSlots; slotIndex++) {
        const box = this.boxForSpace(slotIndex);
        if (box.width === 0 && box.height === 0) {
          positions.push(null);
          continue;
        }

        // Convert SVG center of space to pixel coordinates
        const center = this._svgToPixel(
          box.x + box.width / 2,
          box.y + box.height / 2
        );

        // Apply jitter based on layer index to spread multiple tokens in same space
        const jitterX = this._deterministicJitter(slotIndex, layerIndex, 0) * this.tokenSize;
        const jitterY = this._deterministicJitter(slotIndex, layerIndex, 1) * this.tokenSize;

        positions.push({
          top: center.y - this.tokenSize / 2 + jitterY,
          left: center.x - this.tokenSize / 2 + jitterX,
        });
      }

      result.push(positions);
    }

    this._layerPositions = result;
  }

  // ---- Render ----

  override render() {
    const stacks = this._effectiveStacks;
    const hasStacks = stacks.length > 0;

    return html`
      <div id="board-wrapper">
        <div id="container" @click="${this._spaceTapped}">
          <!-- SVG is loaded via fetch and inserted here -->
          ${hasStacks ? html`
            <div id="token-overlay">
              ${repeat(stacks, (_, i) => i, (s, i) => html`
                <boardgame-component-stack
                  layout="spatial"
                  .stack="${s}"
                  .spatialPositions="${this._layerPositions[i] || []}"
                  .componentAttrs="${this.componentAttrs}"
                  no-default-spacer>
                </boardgame-component-stack>
              `)}
            </div>
          ` : ''}
        </div>
      </div>
    `;
  }
}

customElements.define('boardgame-spatial-board', BoardgameSpatialBoard);

export { BoardgameSpatialBoard };
