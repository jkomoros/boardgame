import { LitElement, html, css } from 'lit';
import { property, query, state } from 'lit/decorators.js';
import { repeat } from 'lit/directives/repeat.js';
import './boardgame-component-stack.js';
import type { ExpandedStack } from '../types/boardgame-types.js';
import type { TargetAction } from '../moves/target-action.js';
import {
  geometryFromSvg,
  parseTrustedBoardSvg,
  type BoardPiece,
  resolveBoardGeometry,
  type BoardGeometryFactory,
  type ResolvedBoardGeometry,
  type SpatialBoardKey,
} from './spatial-board-geometry.js';

/**
 * A reusable spatial board component that loads an SVG map and renders
 * game tokens at SVG space positions automatically.
 *
 * Author hit regions with data-board-space and an accessible label, then bind
 * the same TargetAction used by grid or custom markup. Optional focus and piece
 * anchors keep interaction, focus, and token placement as separate geometry.
 *
 * New renderers pass `.pieces=${piecesFromSizedStacks(...)}`. The stack/stacks,
 * disabledSpaces, spacePrefix, and space-tapped APIs remain migration adapters.
 *
 * Manual mode (backward compat): omit `.stack`/`.stacks` and use
 * `boxForSpace()`, `tokenPosition()` for manual positioning.
 *
 * Example (auto-rendering):
 * ```html
 * <boardgame-spatial-board
 *   svgUrl="game-src/mygame/board.svg"
 *   .pieces=${piecesFromSizedStacks(positionStacks, roomKeys)}
 *   .action=${this.move(MoveNames.MoveToRoom).targets(roomKeys, room => ({
 *     TargetLocation: room,
 *   }))}>
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

    [data-space].focused {
      filter: brightness(1.2) drop-shadow(0 0 4px Highlight);
      outline: 3px solid Highlight;
    }

    #token-overlay {
      position: absolute;
      inset: 0;
      pointer-events: none;
      overflow: hidden;
    }

    #focus-overlay {
      position: absolute;
      inset: 0;
      pointer-events: none;
    }

    .space-focus {
      appearance: none;
      position: absolute;
      width: 30px;
      height: 30px;
      padding: 0;
      border: 2px solid transparent;
      border-radius: 50%;
      background: transparent;
      transform: translate(-50%, -50%);
      pointer-events: none;
    }

    .space-focus:focus-visible {
      border-color: var(--md-sys-color-primary, #315c3b);
      outline: 3px solid var(--md-sys-color-surface, white);
      outline-offset: 1px;
    }

    #status {
      margin: 0.5rem;
      color: var(--board-on-surface, white);
    }

    #status[hidden] {
      display: none;
    }

    #space-list > div {
      display: flex;
      flex-wrap: wrap;
      gap: 0.35rem;
      margin-top: 0.5rem;
    }

    #space-list button[aria-disabled='true'] {
      opacity: 0.65;
    }

    #geometry-inspector {
      margin-top: 0.5rem;
      padding: 0.5rem;
      background: Canvas;
      color: CanvasText;
      font: 12px/1.4 ui-monospace, monospace;
      white-space: pre-wrap;
    }

    @media (prefers-reduced-motion: reduce) {
      [data-space] { transition: none; }
    }

    @media (forced-colors: active) {
      [data-space].disabled { outline: 2px dashed GrayText; }
      [data-space].focused { outline: 3px solid Highlight; }
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

  /** Headless legality/activation shared with grid and custom renderers. */
  @property({ type: Object, attribute: false })
  action: TargetAction<SpatialBoardKey> | null = null;

  /** Optional explicit sidecar. Omit to extract data-board-* attributes. */
  @property({ type: Object, attribute: false })
  geometry: BoardGeometryFactory<SpatialBoardKey> | null = null;

  /** Accessible name for the interactive board region. */
  @property({ type: String, attribute: 'board-label' })
  boardLabel = 'Game board';

  /** Optional extra context announced with the board region. */
  @property({ type: String, attribute: 'board-description' })
  boardDescription = '';

  /** Single SizedStack for auto-rendering (slot index = space index). */
  @property({ type: Object, attribute: false })
  stack: ExpandedStack<object, object> | null = null;

  /** Multiple SizedStacks for multi-token auto-rendering. */
  @property({ type: Array, attribute: false })
  stacks: readonly ExpandedStack<object, object>[] = [];

  /** Explicit piece-to-space projection; preferred over legacy slot coupling. */
  @property({ type: Array, attribute: false })
  pieces: readonly BoardPiece<SpatialBoardKey>[] = [];

  /** Size of token elements in pixels. */
  @property({ type: Number })
  tokenSize = 24;

  /** Attributes to forward to child components. */
  @property({ type: Object, attribute: false })
  componentAttrs: Record<string, unknown> = {};

  /** Development-only geometry report; never required for gameplay. */
  @property({ type: Boolean, attribute: 'geometry-inspector' })
  geometryInspector = false;

  /** True after the SVG has been loaded and inserted into the DOM. */
  @property({ type: Boolean })
  svgLoaded = false;

  /** Computed pixel positions per layer. Updated on state/resize changes. */
  @state()
  private _layerPositions: Array<Array<{ top: number; left: number } | null>> = [];

  @state()
  private _focusPositions: ReadonlyMap<string, { top: number; left: number }> = new Map();

  @state()
  private _inspection = '';

  @state()
  private _loadError: string | null = null;

  @state()
  private _resolvedGeometry: ResolvedBoardGeometry<SpatialBoardKey> | null = null;

  @query('#container')
  private _container!: HTMLDivElement;

  private _resizeObserver: ResizeObserver | null = null;
  private _loadController: AbortController | null = null;
  private _loadGeneration = 0;
  private _unsubscribeAction: (() => void) | null = null;
  private _candidatesByKey = new Map<string, TargetAction<SpatialBoardKey>['candidates'][number]>();

  override connectedCallback() {
    super.connectedCallback();
    this._resizeObserver = new ResizeObserver(() => {
      if (this.svgLoaded) this._recalculatePositions();
    });
    if (this.hasUpdated && this._container) this._resizeObserver.observe(this._container);
    this._subscribeAction();
    if (this.hasUpdated && this.svgUrl && !this.svgLoaded) void this._loadSvg();
  }

  override disconnectedCallback() {
    super.disconnectedCallback();
    this._resizeObserver?.disconnect();
    this._resizeObserver = null;
    this._loadController?.abort();
    this._loadController = null;
    this._unsubscribeAction?.();
    this._unsubscribeAction = null;
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

    if (changedProperties.has('svgUrl') && changedProperties.get('svgUrl') !== undefined) {
      if (this.svgUrl) void this._loadSvg();
      else this._clearSvg();
    }

    if (changedProperties.has('geometry') && changedProperties.get('geometry') !== undefined && this.svgUrl) {
      void this._loadSvg();
    }

    if (changedProperties.has('action')) this._subscribeAction();

    if (changedProperties.has('disabledSpaces') && this.svgLoaded) {
      this._applyDisabledSpaces();
    }

    if ((changedProperties.has('stack') || changedProperties.has('stacks')
      || changedProperties.has('pieces') || changedProperties.has('tokenSize')) && this.svgLoaded) {
      this._recalculatePositions();
    }
  }

  // ---- SVG loading ----

  private async _loadSvg() {
    const generation = ++this._loadGeneration;
    this._loadController?.abort();
    const controller = new AbortController();
    this._loadController = controller;
    this.svgLoaded = false;
    this._loadError = null;
    this._resolvedGeometry = null;
    this._layerPositions = [];
    this._focusPositions = new Map();
    this._inspection = '';
    this._container.querySelector('svg')?.remove();
    try {
      const response = await fetch(this.svgUrl, { signal: controller.signal });
      if (!response.ok) throw new Error(`request failed with HTTP ${response.status}`);
      const contentType = response.headers.get('content-type');
      if (contentType && !/(?:image\/svg\+xml|application\/xml|text\/xml)/i.test(contentType)) {
        throw new Error(`response content type ${JSON.stringify(contentType)} is not SVG/XML`);
      }
      const contentLength = Number(response.headers.get('content-length'));
      if (Number.isFinite(contentLength) && contentLength > 2 * 1024 * 1024) {
        throw new Error('SVG response exceeds the 2097152-byte limit');
      }
      const text = await response.text();
      const svgElement = parseTrustedBoardSvg(text);
      if (controller.signal.aborted || generation !== this._loadGeneration || !this.isConnected) return;

      // Legacy migration adapter. New artwork authors data-board-* directly.
      if (!svgElement.querySelector('[data-board-space]') && this.spacePrefix) {
        for (const element of [...svgElement.querySelectorAll('[id]')]) {
          const id = element.id;
          if (!id.startsWith(this.spacePrefix)) continue;
          const suffix = id.slice(this.spacePrefix.length);
          if (!/^\d+$/.test(suffix)) continue;
          element.setAttribute('data-board-space', suffix);
          element.setAttribute('data-board-label', `Space ${suffix}`);
        }
      }
      for (const element of svgElement.querySelectorAll('[data-board-space]')) element.setAttribute('data-space', '');
      svgElement.setAttribute('aria-hidden', 'true');
      svgElement.setAttribute('focusable', 'false');
      for (const element of svgElement.querySelectorAll('[tabindex]')) element.removeAttribute('tabindex');

      // Mount temporarily so browser geometry APIs can measure it. The catch
      // path removes it if any creator contract fails validation.
      const existing = this._container.querySelector('svg');
      if (existing) this._container.removeChild(existing);
      this._container.insertBefore(svgElement, this._container.firstChild);

      const geometry = this.geometry ? this.geometry(svgElement) : geometryFromSvg(svgElement);
      const resolvedGeometry = resolveBoardGeometry(geometry);
      this._resolvedGeometry = resolvedGeometry;
      this._validateActionKeys();
      this._validateRenderInputs();

      void this.action?.ensurePreview();

      this._applyDisabledSpaces();
      this.svgLoaded = true;
      this._recalculatePositions();
      this.dispatchEvent(new CustomEvent('svg-loaded-changed', {
        composed: true,
        detail: { value: true }
      }));
    } catch (err) {
      if (controller.signal.aborted || generation !== this._loadGeneration || !this.isConnected) return;
      this._container.querySelector('svg')?.remove();
      this._resolvedGeometry = null;
      this._loadError = err instanceof Error ? err.message : String(err);
      this.dispatchEvent(new CustomEvent('svg-load-error', {
        composed: true,
        detail: { message: this._loadError },
      }));
    }
  }

  private _clearSvg(): void {
    this._loadGeneration++;
    this._loadController?.abort();
    this._loadController = null;
    this._container?.querySelector('svg')?.remove();
    this.svgLoaded = false;
    this._resolvedGeometry = null;
    this._layerPositions = [];
    this._focusPositions = new Map();
    this._inspection = '';
    this._loadError = null;
  }

  private _subscribeAction(): void {
    this._unsubscribeAction?.();
    this._unsubscribeAction = this.isConnected && this.action
      ? this.action.subscribe(() => {
        this._refreshCandidates();
        this._applyDisabledSpaces();
        this.requestUpdate();
      })
      : null;
    this._refreshCandidates();
    if (this.action && this._resolvedGeometry) {
      try {
        this._validateActionKeys();
        this._loadError = null;
        void this.action.ensurePreview();
      } catch (error) {
        this._loadError = error instanceof Error ? error.message : String(error);
        this.dispatchEvent(new CustomEvent('svg-load-error', {
          composed: true,
          detail: { message: this._loadError },
        }));
      }
    }
  }

  private _refreshCandidates(): void {
    this._candidatesByKey = new Map((this.action?.candidates ?? []).map(candidate => [String(candidate.key), candidate]));
  }

  private _validateActionKeys(): void {
    if (!this.action || !this._resolvedGeometry) return;
    const geometryKeyList = this._resolvedGeometry.spaces.map(space => String(space.key));
    const targetKeyList = this.action.candidates.map(candidate => String(candidate.key));
    const geometryKeys = new Set(geometryKeyList);
    const targetKeys = new Set(targetKeyList);
    if (geometryKeys.size !== geometryKeyList.length || targetKeys.size !== targetKeyList.length) {
      throw new Error('boardgame-spatial-board: keys must remain unique when represented as SVG attribute strings');
    }
    const missing = [...geometryKeys].filter(key => !targetKeys.has(key));
    const unknown = [...targetKeys].filter(key => !geometryKeys.has(key));
    if (missing.length || unknown.length) {
      throw new Error(`boardgame-spatial-board: target keys do not match geometry; missing [${missing.join(', ')}], unknown [${unknown.join(', ')}]`);
    }
  }

  private _candidateForSpace(key: SpatialBoardKey) {
    return this._candidatesByKey.get(String(key));
  }

  private _activateSpace(space: ResolvedBoardGeometry<SpatialBoardKey>['spaces'][number]): void {
    const candidate = this._candidateForSpace(space.key);
    if (candidate) {
      if (candidate.action.canActivate) void candidate.action.activate();
      return;
    }
    const index = Number(space.key);
    if (!Number.isInteger(index) || this.disabledSpaces.includes(index)) return;
    this.dispatchEvent(new CustomEvent('space-tapped', {
      composed: true,
      bubbles: true,
      detail: { index },
    }));
  }

  // ---- Disabled spaces ----

  private _applyDisabledSpaces() {
    if (!this._resolvedGeometry) return;
    for (const space of this._resolvedGeometry.spaces) {
      const legacyDisabled = typeof space.key === 'number'
        ? this.disabledSpaces.includes(space.key)
        : this.disabledSpaces.includes(Number(space.key));
      const actionDisabled = this.action ? !(this._candidateForSpace(space.key)?.action.canActivate ?? false) : false;
      space.region.classList.toggle('disabled', legacyDisabled || actionDisabled);
    }
  }

  // ---- Click handling ----

  private _spaceTapped(e: Event) {
    if (!(e instanceof PointerEvent) && !(e instanceof MouseEvent)) return;
    const path = [
      ...e.composedPath(),
      ...(this.shadowRoot?.elementsFromPoint(e.clientX, e.clientY) ?? []),
    ];
    const spaces = this._resolvedGeometry?.spaces ?? [];
    const space = path.map(node => spaces.find(candidate =>
      candidate.region === node || (node instanceof Node && candidate.region.contains(node))))
      .find(candidate => candidate !== undefined);
    if (!space) return;
    this._activateSpace(space);
  }

  // ---- Coordinate conversion ----

  /**
   * Convert SVG user-space coordinates to pixel coordinates relative
   * to the #container div. Uses getScreenCTM() for accurate mapping
   * that respects viewBox, preserveAspectRatio, and CSS sizing.
   */
  private _elementCenterPixel(element: SVGGraphicsElement): { x: number; y: number } {
    const box = element.getBBox();
    const ctm = element.getScreenCTM();
    if (!ctm) throw new Error('boardgame-spatial-board: piece anchor has no screen transform');
    const point = new DOMPoint(box.x + box.width / 2, box.y + box.height / 2).matrixTransform(ctm);
    const containerRect = this._container.getBoundingClientRect();
    const result = {
      x: point.x - containerRect.left,
      y: point.y - containerRect.top,
    };
    if (!Number.isFinite(result.x) || !Number.isFinite(result.y)) {
      throw new Error('boardgame-spatial-board: piece anchor produced nonfinite screen coordinates');
    }
    return result;
  }

  // ---- Spatial queries (public, backward compat) ----

  boxForSpace(index: number): { x: number; y: number; width: number; height: number } {
    const result = { x: 0, y: 0, width: 0, height: 0 };
    if (!this.shadowRoot) return result;
    const space = this._resolvedGeometry?.spaces.find(candidate => String(candidate.key) === String(index))?.region ?? null;
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

  private get _effectiveStacks(): readonly ExpandedStack<object, object>[] {
    if (this.pieces.length) return [...new Set(this.pieces.map(piece => piece.stack))];
    if (this.stack) return [this.stack];
    return this.stacks;
  }

  /**
   * Recalculate pixel positions for all stacks/layers.
   * Called when SVG loads, state changes, or container resizes.
   */
  private _recalculatePositions() {
    this._validateRenderInputs();
    const focusPositions = new Map<string, { top: number; left: number }>();
    for (const space of this._resolvedGeometry?.spaces ?? []) {
      const center = this._elementCenterPixel(space.focusAnchor);
      focusPositions.set(String(space.key), { top: center.y, left: center.x });
    }
    this._focusPositions = focusPositions;
    this._inspection = this.geometryInspector ? this._geometryInspection() : '';
    const stacks = this._effectiveStacks;
    if (stacks.length === 0) {
      this._layerPositions = [];
      return;
    }

    const result: Array<Array<{ top: number; left: number } | null>> = [];

    for (let layerIndex = 0; layerIndex < stacks.length; layerIndex++) {
      const stack = stacks[layerIndex]!;
      const numSlots = stack?.Components?.length || 0;
      const positions: Array<{ top: number; left: number } | null> = [];
      const explicitPieces = this.pieces.filter(piece => piece.stack === stack);
      if (explicitPieces.length) {
        const seenSlots = new Set<number>();
        for (const piece of explicitPieces) {
          if (!Number.isInteger(piece.slot) || piece.slot < 0 || piece.slot >= numSlots) {
            throw new Error(`boardgame-spatial-board: piece ${JSON.stringify(piece.id)} has invalid stack slot ${piece.slot}`);
          }
          if (seenSlots.has(piece.slot)) throw new Error(`boardgame-spatial-board: duplicate piece slot ${piece.slot}`);
          seenSlots.add(piece.slot);
          if (stack.IDs[piece.slot] !== piece.id) {
            throw new Error(`boardgame-spatial-board: piece ${JSON.stringify(piece.id)} does not match stack slot ${piece.slot}`);
          }
          if (stack.Components[piece.slot] !== piece.component) {
            throw new Error(`boardgame-spatial-board: piece ${JSON.stringify(piece.id)} component does not match stack slot ${piece.slot}`);
          }
          if (!this._resolvedGeometry?.spaces.some(space => String(space.key) === String(piece.space))) {
            throw new Error(`boardgame-spatial-board: piece ${JSON.stringify(piece.id)} references unknown space ${JSON.stringify(piece.space)}`);
          }
        }
      }

      for (let slotIndex = 0; slotIndex < numSlots; slotIndex++) {
        const explicit = explicitPieces.find(piece => piece.slot === slotIndex);
        const key = explicit?.space ?? slotIndex;
        const space = this._resolvedGeometry?.spaces.find(candidate => String(candidate.key) === String(key));
        if (!space) {
          positions.push(null);
          continue;
        }

        // Each anchor uses its own full ancestor CTM, including nested transforms.
        const center = this._elementCenterPixel(space.pieceAnchor);

        // Apply jitter based on layer index to spread multiple tokens in same space
        const jitterX = this._deterministicJitter(slotIndex, layerIndex, 0) * this.tokenSize;
        const jitterY = this._deterministicJitter(slotIndex, layerIndex, 1) * this.tokenSize;

        // Keep the token inside the region's screen-space bounds. Explicit
        // piece anchors choose the center; jitter must not push it through a
        // narrow room wall.
        const regionRect = space.region.getBoundingClientRect();
        const containerRect = this._container.getBoundingClientRect();
        const minTop = regionRect.top - containerRect.top;
        const maxTop = regionRect.bottom - containerRect.top - this.tokenSize;
        const minLeft = regionRect.left - containerRect.left;
        const maxLeft = regionRect.right - containerRect.left - this.tokenSize;
        const intendedTop = center.y - this.tokenSize / 2 + jitterY;
        const intendedLeft = center.x - this.tokenSize / 2 + jitterX;

        positions.push({
          top: maxTop >= minTop ? Math.min(maxTop, Math.max(minTop, intendedTop)) : center.y - this.tokenSize / 2,
          left: maxLeft >= minLeft ? Math.min(maxLeft, Math.max(minLeft, intendedLeft)) : center.x - this.tokenSize / 2,
        });
      }

      result.push(positions);
    }

    this._layerPositions = result;
  }

  private _validateRenderInputs(): void {
    if (!Number.isFinite(this.tokenSize) || this.tokenSize <= 0) {
      throw new Error('boardgame-spatial-board: tokenSize must be a finite positive number');
    }
    const modes = [this.pieces.length > 0, this.stack !== null, this.stacks.length > 0].filter(Boolean).length;
    if (modes > 1) {
      throw new Error('boardgame-spatial-board: choose exactly one of pieces, stack, or stacks');
    }
  }

  private _geometryInspection(): string {
    const spaces = this._resolvedGeometry?.spaces ?? [];
    const boxes = spaces.map(space => ({ space, box: space.region.getBoundingClientRect() }));
    const overlaps: string[] = [];
    for (let left = 0; left < boxes.length; left++) {
      for (let right = left + 1; right < boxes.length; right++) {
        const a = boxes[left]!;
        const b = boxes[right]!;
        if (a.box.left < b.box.right && a.box.right > b.box.left
          && a.box.top < b.box.bottom && a.box.bottom > b.box.top) {
          overlaps.push(`${String(a.space.key)}↔${String(b.space.key)}`);
        }
      }
    }
    const lines = spaces.map(space => {
      const focus = this._focusPositions.get(String(space.key));
      const piece = this._elementCenterPixel(space.pieceAnchor);
      return `${space.order}: ${String(space.key)} — ${space.label}; region=<${space.region.localName}>; `
        + `focus=${focus ? `${focus.left.toFixed(1)},${focus.top.toFixed(1)}` : 'missing'}; `
        + `piece=${piece.x.toFixed(1)},${piece.y.toFixed(1)}`;
    });
    lines.push(`possible bounding-box overlaps: ${overlaps.length ? overlaps.join(', ') : 'none'}`);
    return lines.join('\n');
  }

  // ---- Render ----

  override render() {
    const stacks = this._effectiveStacks;
    const hasStacks = stacks.length > 0;

    return html`
      <div id="board-wrapper" role="region" aria-label=${this.boardLabel}
        aria-describedby=${this.boardDescription ? 'board-description' : undefined}>
        ${this.boardDescription ? html`<span id="board-description" hidden>${this.boardDescription}</span>` : ''}
        <p id="status" role="status" aria-live="polite" ?hidden=${!this._loadError}>
          ${this._loadError ? html`Board artwork could not be loaded: ${this._loadError}
            <button type="button" @click=${this._loadSvg}>Retry</button>` : ''}
        </p>
        <div id="container" @click="${this._spaceTapped}">
          <!-- SVG is loaded via fetch and inserted here -->
          ${this._resolvedGeometry ? html`
            <div id="focus-overlay">
              ${repeat(this._resolvedGeometry.spaces, space => space.key, space => {
                const position = this._focusPositions.get(String(space.key));
                const candidate = this._candidateForSpace(space.key);
                const disabled = candidate
                  ? !candidate.action.canActivate
                  : this.disabledSpaces.includes(Number(space.key));
                const reason = candidate?.action.reason?.message;
                return position ? html`<button
                  class="space-focus"
                  type="button"
                  style=${`left:${position.left}px;top:${position.top}px`}
                  aria-label=${reason
                    ? `${space.label}. ${candidate?.action.canActivate ? 'Retry available' : 'Unavailable'}: ${reason}`
                    : space.label}
                  aria-disabled=${String(disabled)}
                  @focus=${() => space.region.classList.add('focused')}
                  @blur=${() => space.region.classList.remove('focused')}
                  @click=${(event: Event) => { event.stopPropagation(); this._activateSpace(space); }}
                ></button>` : '';
              })}
            </div>
          ` : ''}
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
        ${this._resolvedGeometry ? html`
          <details id="space-list">
            <summary>Board spaces</summary>
            <div>${repeat(this._resolvedGeometry.spaces, space => space.key, space => {
              const candidate = this._candidateForSpace(space.key);
              const legacyDisabled = this.disabledSpaces.includes(Number(space.key));
              const disabled = candidate ? !candidate.action.canActivate : legacyDisabled;
              const reason = candidate?.action.reason?.message;
              return html`<button
                type="button"
                aria-disabled=${String(disabled)}
                title=${reason ?? ''}
                @focus=${() => space.region.classList.add('focused')}
                @blur=${() => space.region.classList.remove('focused')}
                @click=${(event: Event) => {
                  event.stopPropagation();
                  this._activateSpace(space);
                }}>${space.label}${reason ? ` — ${reason}` : ''}</button>`;
            })}</div>
          </details>
        ` : ''}
        ${this.geometryInspector && this._inspection ? html`
          <details id="geometry-inspector" open>
            <summary>Board geometry inspector</summary>
            ${this._inspection}
          </details>
        ` : ''}
      </div>
    `;
  }
}

customElements.define('boardgame-spatial-board', BoardgameSpatialBoard);

export { BoardgameSpatialBoard };
