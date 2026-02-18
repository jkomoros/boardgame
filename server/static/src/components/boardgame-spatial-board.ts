import { LitElement, html, css } from 'lit';
import { property, query } from 'lit/decorators.js';

/**
 * A reusable spatial board component that loads an SVG map and provides
 * spatial queries and interaction for board game spaces.
 *
 * The SVG must contain elements with IDs following the pattern
 * `{spacePrefix}{index}` (e.g., "Space-1", "Space-2", ...).
 *
 * Features:
 * - Loads SVG from a configurable URL
 * - Fires `space-tapped` events when spaces are clicked
 * - Provides `boxForSpace()` and `coordinatesForSpace()` for token positioning
 * - Supports disabling specific spaces via `disabledSpaces`
 * - Provides `tokenPosition()` with deterministic jitter for placing tokens
 *
 * Example:
 * ```html
 * <boardgame-spatial-board
 *   svgUrl="game-src/mygame/board.svg"
 *   spacePrefix="Room-"
 *   .disabledSpaces="${[8, 9, 10]}"
 *   @space-tapped="${this._onSpaceTapped}">
 * </boardgame-spatial-board>
 * ```
 */
class BoardgameSpatialBoard extends LitElement {

  static override styles = css`
    :host {
      display: block;
    }

    [data-space] {
      cursor: pointer;
    }

    [data-space].disabled {
      fill: #CCC !important;
      cursor: default !important;
    }
  `;

  /** URL to fetch the SVG board from. */
  @property({ type: String })
  svgUrl = '';

  /** ID prefix for space elements in the SVG (e.g., "Space-" matches "Space-1"). */
  @property({ type: String })
  spacePrefix = 'Space-';

  /** Array of space indices that should be visually disabled. */
  @property({ type: Array })
  disabledSpaces: number[] = [];

  /** True after the SVG has been loaded and inserted into the DOM. */
  @property({ type: Boolean })
  svgLoaded = false;

  @query('#container')
  private _container!: HTMLDivElement;

  override async firstUpdated(_changedProperties: Map<PropertyKey, unknown>) {
    super.firstUpdated(_changedProperties);
    if (this.svgUrl) {
      await this._loadSvg();
    }
  }

  override updated(changedProperties: Map<PropertyKey, unknown>) {
    super.updated(changedProperties);

    if (changedProperties.has('svgUrl') && this.svgUrl && changedProperties.get('svgUrl') !== undefined) {
      // svgUrl changed after initial render
      this._loadSvg();
    }

    if (changedProperties.has('disabledSpaces') && this.svgLoaded) {
      this._applyDisabledSpaces();
    }
  }

  private async _loadSvg() {
    try {
      const response = await fetch(this.svgUrl);
      const text = await response.text();
      const parser = new DOMParser();
      const doc = parser.parseFromString(text, 'image/svg+xml');
      const svgElement = doc.documentElement;

      // Clear previous SVG if any
      while (this._container.firstChild) {
        this._container.removeChild(this._container.firstChild);
      }

      this._container.appendChild(svgElement);

      // Mark space elements for CSS targeting
      const spaces = svgElement.querySelectorAll(`[id^="${this.spacePrefix}"]`);
      spaces.forEach(el => el.setAttribute('data-space', ''));

      this._applyDisabledSpaces();
      this.svgLoaded = true;
      this.dispatchEvent(new CustomEvent('svg-loaded-changed', {
        composed: true,
        detail: { value: true }
      }));
    } catch (err) {
      console.error('boardgame-spatial-board: Failed to load SVG:', err);
    }
  }

  private _applyDisabledSpaces() {
    if (!this.shadowRoot) return;

    // Remove disabled class from all space elements
    const allSpaces = this.shadowRoot.querySelectorAll('[data-space]');
    allSpaces.forEach(el => el.classList.remove('disabled'));

    // Add disabled class to specified spaces
    for (const index of this.disabledSpaces) {
      const el = this.shadowRoot.querySelector(`#${this.spacePrefix}${index}`);
      if (el) {
        el.classList.add('disabled');
      }
    }
  }

  private _spaceTapped(e: Event) {
    const target = e.target as HTMLElement;
    const id = target.id;

    if (!id.startsWith(this.spacePrefix)) return;

    const index = parseInt(id.substring(this.spacePrefix.length));

    if (isNaN(index)) return;

    // Don't fire events for disabled spaces
    if (this.disabledSpaces.includes(index)) return;

    this.dispatchEvent(new CustomEvent('space-tapped', {
      composed: true,
      bubbles: true,
      detail: { index }
    }));
  }

  /**
   * Returns the bounding box of a space element in the SVG.
   * Returns a zero-sized box if the space is not found.
   */
  boxForSpace(index: number): { x: number; y: number; width: number; height: number } {
    const result = { x: 0, y: 0, width: 0, height: 0 };

    if (!this.shadowRoot) return result;

    const space = this.shadowRoot.querySelector(`#${this.spacePrefix}${index}`) as SVGGraphicsElement | null;
    if (!space) {
      console.warn(`boardgame-spatial-board: Couldn't find space ${index}`);
      return result;
    }

    const box = space.getBBox();
    result.x = box.x;
    result.y = box.y;
    result.width = box.width;
    result.height = box.height;
    return result;
  }

  /**
   * Returns the center coordinates of a space element in the SVG.
   */
  coordinatesForSpace(index: number): [number, number] {
    const box = this.boxForSpace(index);
    return [box.x + (box.width / 2), box.y + (box.height / 2)];
  }

  /**
   * Computes the position for a token within a space, with deterministic
   * jitter to avoid tokens stacking on top of each other.
   *
   * @param spaceIndex - The space the token is in
   * @param tokenIndex - A unique index for this token (e.g., player index, or -1 for NPC)
   * @param tokenSize - The size of the token in pixels
   * @returns The top/left position, or undefined if the SVG isn't loaded
   */
  tokenPosition(
    spaceIndex: number,
    tokenIndex: number,
    tokenSize: number
  ): { top: number; left: number } | undefined {
    if (!this.svgLoaded) return undefined;

    const coords = this.coordinatesForSpace(spaceIndex);
    const box = this.boxForSpace(spaceIndex);

    if (box.width === 0 && box.height === 0) return undefined;

    // Deterministic jitter based on space and token indices
    const jitterX = this._deterministicJitter(spaceIndex, tokenIndex, 0) * tokenSize * 2;
    const jitterY = this._deterministicJitter(spaceIndex, tokenIndex, 1) * tokenSize * 2;

    let top = coords[1] - (tokenSize / 2) + jitterY;
    let left = coords[0] - (tokenSize / 2) + jitterX;

    // Clamp to space bounds
    if ((top + tokenSize) > (box.y + box.height)) top = box.y + box.height - tokenSize;
    if (top < box.y) top = box.y;
    if ((left + tokenSize) > (box.x + box.width)) left = box.x + box.width - tokenSize;
    if (left < box.x) left = box.x;

    return { top, left };
  }

  /**
   * Returns a deterministic pseudo-random number in [-1, 1] for the given
   * space index, token index, and axis. Uses a simple hash to ensure
   * consistent positioning across renders.
   */
  private _deterministicJitter(spaceIndex: number, tokenIndex: number, axis: number): number {
    // Simple hash combining space, token, and axis
    let hash = ((spaceIndex * 31 + tokenIndex) * 37 + axis) * 41;
    hash = ((hash >>> 16) ^ hash) * 0x45d9f3b;
    hash = ((hash >>> 16) ^ hash) * 0x45d9f3b;
    hash = (hash >>> 16) ^ hash;
    // Normalize to [-1, 1]
    return ((hash & 0xFFFF) / 0x7FFF) - 1;
  }

  override render() {
    return html`
      <div id="container" @click="${this._spaceTapped}">
        <!-- SVG is loaded via fetch and appended here -->
      </div>
    `;
  }
}

customElements.define('boardgame-spatial-board', BoardgameSpatialBoard);

export { BoardgameSpatialBoard };
