import { LitElement, html, css, svg } from 'lit';
import { property, query, state } from 'lit/decorators.js';
import { repeat } from 'lit/directives/repeat.js';
import './boardgame-component-stack.js';
import './boardgame-board-viewport.js';
import type { ExpandedStack } from '../types/boardgame-types.js';
import type { TargetAction } from '../moves/target-action.js';
import type { PlacementTargetBinding } from '../moves/placement-draft.js';
import type { TargetKey } from '../moves/target-action.js';
import type { ComponentView } from './component-view.js';
import type { BoardgameBoardViewport } from './boardgame-board-viewport.js';
import {
  geometryFromSvg,
  parseTrustedBoardSvg,
  rasterArtworkScene,
  type BoardPiece,
  type BoardPathOverlay,
  type BoardGeometry,
  resolveBoardGeometry,
  type BoardGeometryFactory,
  type RasterBoardArtwork,
  type ResolvedBoardGeometry,
  type SpatialBoardKey,
} from './spatial-board-geometry.js';

/** Placement-draft surface consumed by graphic-board destinations. */
export interface SpatialPlacementDraft {
  readonly targets: readonly SpatialBoardKey[];
  readonly selectedItem: TargetKey | null;
  target(target: SpatialBoardKey): PlacementTargetBinding<TargetKey, SpatialBoardKey>;
}

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

    [data-space].inactive {
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
      z-index: 2;
    }

    #focus-overlay {
      position: absolute;
      inset: 0;
      pointer-events: none;
      z-index: 3;
    }

    #path-overlay {
      inset: 0;
      overflow: visible;
      pointer-events: none;
      position: absolute;
      height: 100%;
      width: 100%;
      z-index: 1;
    }

    #path-overlay polyline {
      fill: none;
      stroke: var(--board-path-primary, #1565c0);
      stroke-linecap: round;
      stroke-linejoin: round;
      vector-effect: non-scaling-stroke;
    }

    #path-overlay polyline.secondary { stroke: var(--board-path-secondary, #7b1fa2); }
    #path-overlay polyline.danger { stroke: var(--board-path-danger, #b3261e); }
    #path-overlay polyline.muted { stroke: var(--board-path-muted, #616161); }

    #path-descriptions {
      block-size: 1px;
      clip: rect(0 0 0 0);
      clip-path: inset(50%);
      inline-size: 1px;
      overflow: hidden;
      position: absolute;
      white-space: nowrap;
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
  @property({ type: String, attribute: 'svg-url' })
  svgUrl = '';

  /** Raster board plus normalized interactive hotspots. Mutually exclusive with svgUrl. */
  @property({ type: Object, attribute: false })
  artwork: RasterBoardArtwork<SpatialBoardKey> | null = null;

  /** ID prefix for space elements in the SVG (e.g., "Space-" matches "Space-0"). */
  @property({ type: String, attribute: 'space-prefix' })
  spacePrefix = 'Space-';

  /** Array of space indices that should be visually disabled. */
  @property({ type: Array, attribute: 'disabled-spaces' })
  disabledSpaces: number[] = [];

  /** Headless legality/activation shared with grid and custom renderers. */
  @property({ type: Object, attribute: false })
  action: TargetAction<SpatialBoardKey> | null = null;

  /** Local draft destinations; mutually exclusive with action. */
  @property({ attribute: false })
  placementDraft: SpatialPlacementDraft | null = null;

  /** Exact geometry group targeted by action or placementDraft. Empty means every geometry key. */
  @property({ type: String, attribute: 'action-group' })
  actionGroup = '';

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

  /** Accessible, pointer-safe routes drawn through known piece anchors. */
  @property({ type: Array, attribute: false })
  pathOverlays: readonly BoardPathOverlay<SpatialBoardKey>[] = [];

  /** Size of token elements in pixels. */
  @property({ type: Number, attribute: 'token-size' })
  tokenSize = 24;

  /** Explicit escape hatch for untyped child properties. Prefer bound views. */
  @property({ type: Object, attribute: false })
  unsafeComponentAttrs: Record<string, unknown> = {};

  /** One renderer-scoped view shared by every spatial stack layer. */
  @property({ type: Object, attribute: false })
  componentView: ComponentView | null = null;

  /**
   * Heterogeneous escape hatch: one view (or explicit null) per effective
   * stack layer, in the same order as stack/stacks or first appearance in
   * pieces. Mutually exclusive with componentView and cardinality-checked.
   */
  @property({ type: Array, attribute: false })
  componentViews: readonly (ComponentView | null)[] = [];

  /** Development-only geometry report; never required for gameplay. */
  @property({ type: Boolean, attribute: 'geometry-inspector' })
  geometryInspector = false;

  /** Enable bounded pan/zoom controls and gestures for a large board scene. */
  @property({ type: Boolean, attribute: 'pan-zoom' })
  panZoom = false;

  /** Largest magnification when panZoom is enabled. */
  @property({ type: Number, attribute: 'max-zoom' })
  maxZoom = 4;

  /** True after the SVG has been loaded and inserted into the DOM. */
  @property({ type: Boolean, attribute: 'svg-loaded' })
  svgLoaded = false;

  /** Computed pixel positions per layer. Updated on state/resize changes. */
  @state()
  private _layerPositions: Array<Array<{ top: number; left: number } | null>> = [];

  @state()
  private _focusPositions: ReadonlyMap<string, { top: number; left: number }> = new Map();

  @state()
  private _inspection = '';

  @state()
  private _resolvedPathOverlays: readonly {
    readonly id: string;
    readonly label: string;
    readonly points: string;
    readonly tone: 'primary' | 'secondary' | 'danger' | 'muted';
    readonly width: number;
  }[] = [];

  @state()
  private _loadError: string | null = null;

  @state()
  private _resolvedGeometry: ResolvedBoardGeometry<SpatialBoardKey> | null = null;

  @query('#container')
  private _container!: HTMLDivElement;

  @query('boardgame-board-viewport')
  private _boardViewport?: BoardgameBoardViewport;

  private _resizeObserver: ResizeObserver | null = null;
  private _loadController: AbortController | null = null;
  private _loadGeneration = 0;
  private _unsubscribeAction: (() => void) | null = null;
  private _candidatesByKey = new Map<string, TargetAction<SpatialBoardKey>['candidates'][number]>();
  private _placementTargetsByKey = new Map<string, SpatialBoardKey>();

  override connectedCallback() {
    super.connectedCallback();
    this._resizeObserver = new ResizeObserver(() => {
      if (this.svgLoaded) this._recalculatePositions();
    });
    if (this.hasUpdated && this._container) this._resizeObserver.observe(this._container);
    this._subscribeAction();
    if (this.hasUpdated && (this.svgUrl || this.artwork) && !this.svgLoaded) void this._loadBoardSource();
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
    if (this.svgUrl || this.artwork) {
      await this._loadBoardSource();
    }
    if (this._container) {
      this._resizeObserver?.observe(this._container);
    }
  }

  protected override willUpdate(changedProperties: Map<PropertyKey, unknown>): void {
    if (changedProperties.has('placementDraft')) this._refreshPlacementTargets();
  }

  override updated(changedProperties: Map<PropertyKey, unknown>) {
    super.updated(changedProperties);

    if ((changedProperties.has('svgUrl') && changedProperties.get('svgUrl') !== undefined)
      || (changedProperties.has('artwork') && changedProperties.get('artwork') !== undefined)) {
      if (this.svgUrl || this.artwork) void this._loadBoardSource();
      else this._clearSvg();
    }

    if (changedProperties.has('geometry') && changedProperties.get('geometry') !== undefined
      && (this.svgUrl || this.artwork)) {
      void this._loadBoardSource();
    }

    if (changedProperties.has('action')) {
      this._subscribeAction();
      if (this.placementDraft && this._resolvedGeometry) this._revalidateActionConfiguration();
    }
    if (changedProperties.has('placementDraft')) {
      if (this._resolvedGeometry) this._revalidateActionConfiguration();
    }

    if (changedProperties.has('actionGroup') && this._resolvedGeometry) {
      this._revalidateActionConfiguration();
    }

    if (changedProperties.has('disabledSpaces') && this.svgLoaded) {
      if (this.placementDraft) this._revalidateActionConfiguration();
      else this._applyDisabledSpaces();
    }

    if ((changedProperties.has('stack') || changedProperties.has('stacks')
      || changedProperties.has('pieces') || changedProperties.has('tokenSize')
      || changedProperties.has('pathOverlays')
      || changedProperties.has('componentView') || changedProperties.has('componentViews')) && this.svgLoaded) {
      this._recalculatePositions();
    }
  }

  // ---- SVG loading ----

  private async _loadBoardSource(): Promise<void> {
    try {
      this._validateSourceConfiguration();
    } catch (error) {
      this._clearSvg();
      this._reportLoadError(error);
      return;
    }
    if (this.artwork) {
      await this._loadRasterArtwork();
      return;
    }
    if (this.svgUrl) {
      await this._loadSvg();
      return;
    }
    this._clearSvg();
  }

  private _beginLoad(): { readonly generation: number; readonly controller: AbortController } {
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
    this._resolvedPathOverlays = [];
    this._container.querySelector('svg')?.remove();
    return { generation, controller };
  }

  private _isStaleLoad(generation: number, controller: AbortController): boolean {
    return controller.signal.aborted || generation !== this._loadGeneration || !this.isConnected;
  }

  private _installScene(
    svgElement: SVGSVGElement,
    geometryForSvg: (svg: SVGSVGElement) => BoardGeometry<SpatialBoardKey>,
  ): void {
    const existing = this._container.querySelector('svg');
    if (existing) this._container.removeChild(existing);
    this._container.insertBefore(svgElement, this._container.firstChild);
    const resolvedGeometry = resolveBoardGeometry(geometryForSvg(svgElement));
    this._resolvedGeometry = resolvedGeometry;
    this._refreshPlacementTargets();
    this._validateActionGroup();
    this._validateActionKeys();
    this._validateRenderInputs();
    void this.action?.ensurePreview();
    this._applyDisabledSpaces();
    this.svgLoaded = true;
    this._recalculatePositions();
    this.dispatchEvent(new CustomEvent('svg-loaded-changed', {
      composed: true,
      detail: { value: true },
    }));
  }

  private _reportLoadError(error: unknown): void {
    this._container?.querySelector('svg')?.remove();
    this.svgLoaded = false;
    this._resolvedGeometry = null;
    this._loadError = error instanceof Error ? error.message : String(error);
    this.dispatchEvent(new CustomEvent('svg-load-error', {
      composed: true,
      detail: { message: this._loadError },
    }));
  }

  private async _loadRasterArtwork(): Promise<void> {
    const artwork = this.artwork;
    if (!artwork) return;
    const { generation, controller } = this._beginLoad();
    try {
      const source = new URL(artwork.src, document.baseURI);
      if (!['http:', 'https:', 'blob:', 'data:'].includes(source.protocol)) {
        throw new Error(`raster artwork uses unsupported URL protocol ${JSON.stringify(source.protocol)}`);
      }
      const dimensions = await this._decodeRasterDimensions(source.href, controller.signal);
      if (this._isStaleLoad(generation, controller)) return;
      const scene = rasterArtworkScene(artwork, dimensions.width, dimensions.height);
      this._installScene(scene.svg, () => scene.geometry);
    } catch (error) {
      if (this._isStaleLoad(generation, controller)) return;
      this._reportLoadError(error);
    }
  }

  private _decodeRasterDimensions(
    src: string,
    signal: AbortSignal,
  ): Promise<{ readonly width: number; readonly height: number }> {
    return new Promise((resolve, reject) => {
      const image = new Image();
      const cleanup = () => {
        signal.removeEventListener('abort', abort);
        image.onload = null;
        image.onerror = null;
      };
      const abort = () => {
        cleanup();
        image.src = '';
        reject(new DOMException('Raster image load aborted', 'AbortError'));
      };
      signal.addEventListener('abort', abort, { once: true });
      image.onload = () => {
        const dimensions = { width: image.naturalWidth, height: image.naturalHeight };
        cleanup();
        resolve(dimensions);
      };
      image.onerror = () => {
        cleanup();
        reject(new Error(`raster image ${JSON.stringify(src)} could not be decoded`));
      };
      image.src = src;
    });
  }

  private async _loadSvg() {
    const { generation, controller } = this._beginLoad();
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
      if (this._isStaleLoad(generation, controller)) return;

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

      this._installScene(svgElement, svg => this.geometry ? this.geometry(svg) : geometryFromSvg(svg));
    } catch (err) {
      if (this._isStaleLoad(generation, controller)) return;
      this._reportLoadError(err);
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
    this._resolvedPathOverlays = [];
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
    if (this._resolvedGeometry) {
      this._revalidateActionConfiguration();
      if (this.action && this._loadError === null) void this.action.ensurePreview();
    }
  }

  private _refreshCandidates(): void {
    this._candidatesByKey = new Map((this.action?.candidates ?? []).map(candidate => [String(candidate.key), candidate]));
  }

  private _refreshPlacementTargets(): void {
    this._placementTargetsByKey = new Map(
      (this.placementDraft?.targets ?? []).map(target => [String(target), target]),
    );
  }

  private _validateActionKeys(): void {
    if (this.action && this.placementDraft) {
      throw new Error('boardgame-spatial-board: action and placementDraft are mutually exclusive');
    }
    if (this.placementDraft && this.disabledSpaces.length > 0) {
      throw new Error('boardgame-spatial-board: placementDraft and disabledSpaces are mutually exclusive');
    }
    if ((!this.action && !this.placementDraft) || !this._resolvedGeometry) return;
    if (this.placementDraft
      && (!Array.isArray(this.placementDraft.targets) || typeof this.placementDraft.target !== 'function')) {
      throw new Error('boardgame-spatial-board: placementDraft must be a PlacementDraftController binding');
    }
    const geometryKeyList = this._actionGeometrySpaces.map(space => String(space.key));
    const targetKeyList = this.action
      ? this.action.candidates.map(candidate => String(candidate.key))
      : this.placementDraft!.targets.map(target => String(target));
    const geometryKeys = new Set(geometryKeyList);
    const targetKeys = new Set(targetKeyList);
    if (geometryKeys.size !== geometryKeyList.length || targetKeys.size !== targetKeyList.length) {
      throw new Error('boardgame-spatial-board: keys must remain unique when represented as SVG attribute strings');
    }
    const missing = [...geometryKeys].filter(key => !targetKeys.has(key));
    const unknown = [...targetKeys].filter(key => !geometryKeys.has(key));
    if (missing.length || unknown.length) {
      const scope = this.actionGroup ? ` group ${JSON.stringify(this.actionGroup)}` : '';
      throw new Error(`boardgame-spatial-board: target keys do not match geometry${scope}; missing [${missing.join(', ')}], unknown [${unknown.join(', ')}]`);
    }
  }

  private get _actionGeometrySpaces(): ResolvedBoardGeometry<SpatialBoardKey>['spaces'] {
    const spaces = this._resolvedGeometry?.spaces ?? [];
    return this.actionGroup ? spaces.filter(space => space.group === this.actionGroup) : spaces;
  }

  private _validateActionGroup(): void {
    if (!this._resolvedGeometry) return;
    if (typeof this.actionGroup !== 'string' || (this.actionGroup && this.actionGroup !== this.actionGroup.trim())) {
      throw new Error('boardgame-spatial-board: actionGroup must be empty or a non-empty string without surrounding whitespace');
    }
    if (this.actionGroup.length > 128 || /[\u0000-\u001f\u007f]/.test(this.actionGroup)) {
      throw new Error('boardgame-spatial-board: actionGroup must be at most 128 characters without control characters');
    }
    if (this.actionGroup && this._actionGeometrySpaces.length === 0) {
      const available = [...new Set(this._resolvedGeometry.spaces.map(space => space.group).filter(Boolean))];
      throw new Error(`boardgame-spatial-board: actionGroup ${JSON.stringify(this.actionGroup)} has no geometry; available groups [${available.join(', ')}]`);
    }
  }

  private _revalidateActionConfiguration(): void {
    try {
      this._validateActionGroup();
      this._validateActionKeys();
      this._loadError = null;
      this._applyDisabledSpaces();
      this.requestUpdate();
    } catch (error) {
      this._loadError = error instanceof Error ? error.message : String(error);
      this._applyDisabledSpaces();
      this.dispatchEvent(new CustomEvent('svg-load-error', {
        composed: true,
        detail: { message: this._loadError },
      }));
    }
  }

  private _candidateForSpace(key: SpatialBoardKey) {
    return this._candidatesByKey.get(String(key));
  }

  private _placementTargetForSpace(key: SpatialBoardKey) {
    if (!this.placementDraft) return null;
    const target = this._placementTargetsByKey.get(String(key));
    return target === undefined ? null : this.placementDraft.target(target);
  }

  private _activateSpace(space: ResolvedBoardGeometry<SpatialBoardKey>['spaces'][number]): void {
    const placement = this._placementTargetForSpace(space.key);
    if (placement) {
      if (placement.canPlace) placement.place();
      return;
    }
    if (this.placementDraft) return;
    const candidate = this._candidateForSpace(space.key);
    if (candidate) {
      if (candidate.action.canActivate) void candidate.action.activate();
      return;
    }
    if (this.action) return;
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
      const placement = this._placementTargetForSpace(space.key);
      const placementDisabled = Boolean(placement && !placement.canPlace);
      const inactive = Boolean((this.action && !this._candidateForSpace(space.key))
        || (this.placementDraft && !placement));
      space.region.classList.toggle('inactive', inactive);
      space.region.classList.toggle('disabled', legacyDisabled || ((actionDisabled || placementDisabled) && !inactive));
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
    const scale = this._boardViewport?.view.scale ?? 1;
    const result = {
      x: (point.x - containerRect.left) / scale,
      y: (point.y - containerRect.top) / scale,
    };
    if (!Number.isFinite(result.x) || !Number.isFinite(result.y)) {
      throw new Error('boardgame-spatial-board: piece anchor produced nonfinite screen coordinates');
    }
    return result;
  }

  /** Reset an enabled map viewport to its full-board view. */
  resetViewport(): void {
    this._boardViewport?.resetView();
  }

  /** Reveal a known board space in an enabled map viewport. */
  revealSpace(key: SpatialBoardKey, padding = 24): void {
    const space = this._resolvedGeometry?.spaces.find(candidate => String(candidate.key) === String(key));
    if (!space) throw new Error(`boardgame-spatial-board: cannot reveal unknown space ${JSON.stringify(key)}`);
    this._boardViewport?.reveal(space.focusAnchor, padding);
  }

  private _focusSpace(space: ResolvedBoardGeometry<SpatialBoardKey>['spaces'][number]): void {
    space.region.classList.add('focused');
    this._boardViewport?.reveal(space.focusAnchor);
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
    this._resolvedPathOverlays = this._resolvePathOverlays();
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
        const scale = this._boardViewport?.view.scale ?? 1;
        const minTop = (regionRect.top - containerRect.top) / scale;
        const maxTop = (regionRect.bottom - containerRect.top) / scale - this.tokenSize;
        const minLeft = (regionRect.left - containerRect.left) / scale;
        const maxLeft = (regionRect.right - containerRect.left) / scale - this.tokenSize;
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
    if (this.componentView !== null && this.componentViews.length > 0) {
      throw new Error('boardgame-spatial-board: choose componentView or componentViews, not both');
    }
    if (this.componentViews.length > 0 && this.componentViews.length !== this._effectiveStacks.length) {
      throw new Error(
        `boardgame-spatial-board: componentViews has ${this.componentViews.length} entries for `
        + `${this._effectiveStacks.length} effective stack layers`,
      );
    }
    if (this.panZoom && (!Number.isFinite(this.maxZoom) || this.maxZoom < 1 || this.maxZoom > 16)) {
      throw new Error('boardgame-spatial-board: maxZoom must be from 1 through 16');
    }
    this._validatePathOverlays();
  }

  private _validatePathOverlays(): void {
    if (!Array.isArray(this.pathOverlays)) {
      throw new Error('boardgame-spatial-board: pathOverlays must be an array');
    }
    if (this.pathOverlays.length > 256) {
      throw new Error('boardgame-spatial-board: pathOverlays exceeds the 256-path limit');
    }
    const ids = new Set<string>();
    const tones = new Set(['primary', 'secondary', 'danger', 'muted']);
    const knownSpaces = this._resolvedGeometry
      ? new Set(this._resolvedGeometry.spaces.map(space => String(space.key)))
      : null;
    let totalPoints = 0;
    for (const [index, path] of this.pathOverlays.entries()) {
      if (!path || typeof path !== 'object') throw new Error(`boardgame-spatial-board: path overlay ${index} must be an object`);
      if (typeof path.id !== 'string' || !path.id.trim() || path.id !== path.id.trim()
        || path.id.length > 128 || /[\u0000-\u001f\u007f]/.test(path.id)) {
        throw new Error(`boardgame-spatial-board: path overlay ${index} id must be a trimmed non-empty string of at most 128 characters`);
      }
      if (ids.has(path.id)) throw new Error(`boardgame-spatial-board: duplicate path overlay id ${JSON.stringify(path.id)}`);
      ids.add(path.id);
      if (typeof path.label !== 'string' || !path.label.trim() || path.label.length > 1024) {
        throw new Error(`boardgame-spatial-board: path overlay ${JSON.stringify(path.id)} requires an accessible label of at most 1024 characters`);
      }
      if (!Array.isArray(path.spaces) || path.spaces.length < 2 || path.spaces.length > 256) {
        throw new Error(`boardgame-spatial-board: path overlay ${JSON.stringify(path.id)} requires 2 through 256 spaces`);
      }
      totalPoints += path.spaces.length;
      if (totalPoints > 4096) {
        throw new Error('boardgame-spatial-board: pathOverlays exceeds the 4096-point total limit');
      }
      for (let point = 0; point < path.spaces.length; point++) {
        const key = path.spaces[point];
        if ((typeof key !== 'string' && typeof key !== 'number')
          || (typeof key === 'string' && !key.length) || (typeof key === 'number' && !Number.isFinite(key))) {
          throw new Error(`boardgame-spatial-board: path overlay ${JSON.stringify(path.id)} has an invalid key at point ${point}`);
        }
        if (point > 0 && String(path.spaces[point - 1]) === String(key)) {
          throw new Error(`boardgame-spatial-board: path overlay ${JSON.stringify(path.id)} repeats adjacent space ${JSON.stringify(key)}`);
        }
        if (knownSpaces && !knownSpaces.has(String(key))) {
          throw new Error(`boardgame-spatial-board: path overlay ${JSON.stringify(path.id)} references unknown space ${JSON.stringify(key)}`);
        }
      }
      const tone = path.tone ?? 'primary';
      if (!tones.has(tone)) {
        throw new Error(`boardgame-spatial-board: path overlay ${JSON.stringify(path.id)} has unknown tone ${JSON.stringify(tone)}`);
      }
      const width = path.width ?? 4;
      if (!Number.isFinite(width) || width < 1 || width > 32) {
        throw new Error(`boardgame-spatial-board: path overlay ${JSON.stringify(path.id)} width must be from 1 through 32`);
      }
    }
  }

  private _resolvePathOverlays(): readonly {
    readonly id: string;
    readonly label: string;
    readonly points: string;
    readonly tone: 'primary' | 'secondary' | 'danger' | 'muted';
    readonly width: number;
  }[] {
    if (!this._resolvedGeometry) return [];
    const spacesByKey = new Map(
      this._resolvedGeometry.spaces.map(space => [String(space.key), space] as const),
    );
    return Object.freeze(this.pathOverlays.map(path => Object.freeze({
      id: path.id,
      label: path.label.trim(),
      points: path.spaces.map(key => {
        const space = spacesByKey.get(String(key))!;
        const point = this._elementCenterPixel(space.pieceAnchor);
        return `${point.x},${point.y}`;
      }).join(' '),
      tone: path.tone ?? 'primary',
      width: path.width ?? 4,
    })));
  }

  private _validateSourceConfiguration(): void {
    if (this.svgUrl && this.artwork) {
      throw new Error('boardgame-spatial-board: choose svgUrl or artwork, not both');
    }
    if (this.artwork && this.geometry) {
      throw new Error('boardgame-spatial-board: geometry is only valid with svgUrl; artwork declares its own spaces');
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
      return `${space.order}: ${String(space.key)} — ${space.label}; group=${space.group ?? 'all'}; `
        + `region=<${space.region.localName}>; `
        + `focus=${focus ? `${focus.left.toFixed(1)},${focus.top.toFixed(1)}` : 'missing'}; `
        + `piece=${piece.x.toFixed(1)},${piece.y.toFixed(1)}`;
    });
    lines.push(`possible bounding-box overlaps: ${overlaps.length ? overlaps.join(', ') : 'none'}`);
    return lines.join('\n');
  }

  // ---- Render ----

  override render() {
    this._validateRenderInputs();
    const stacks = this._effectiveStacks;
    const hasStacks = stacks.length > 0;
    const boardScene = html`
      <div id="container" @click="${this._spaceTapped}">
        <!-- Board scene is loaded/generated and inserted here -->
        ${this._resolvedGeometry ? html`
          <div id="focus-overlay">
            ${repeat(this._resolvedGeometry.spaces, space => space.key, space => {
              const position = this._focusPositions.get(String(space.key));
              const candidate = this._candidateForSpace(space.key);
              const placement = this._placementTargetForSpace(space.key);
              if ((this.action && !candidate) || (this.placementDraft && !placement)) return '';
              const disabled = placement ? !placement.canPlace : candidate
                ? !candidate.action.canActivate
                : this.action || this.placementDraft ? true : this.disabledSpaces.includes(Number(space.key));
              const reason = placement?.reason ?? candidate?.action.reason?.message;
              return position ? html`<button
                class="space-focus"
                type="button"
                style=${`left:${position.left}px;top:${position.top}px`}
                aria-label=${reason
                  ? `${space.label}. ${candidate?.action.canActivate || placement?.canPlace ? 'Available' : 'Unavailable'}: ${reason}`
                  : space.label}
                aria-disabled=${String(disabled)}
                @focus=${() => this._focusSpace(space)}
                @blur=${() => space.region.classList.remove('focused')}
                @click=${(event: Event) => { event.stopPropagation(); this._activateSpace(space); }}
              ></button>` : '';
            })}
          </div>
        ` : ''}
        ${this._resolvedPathOverlays.length ? html`
          <svg id="path-overlay" part="path-overlay" aria-hidden="true">
            ${repeat(this._resolvedPathOverlays, path => path.id, path => svg`
              <polyline part="path" class=${path.tone} points=${path.points} stroke-width=${path.width}></polyline>
            `)}
          </svg>
          <div id="path-descriptions" role="list" aria-label="Board routes">
            ${repeat(this._resolvedPathOverlays, path => path.id, path => html`
              <span role="listitem">${path.label}</span>
            `)}
          </div>
        ` : ''}
        ${hasStacks ? html`
          <div id="token-overlay">
            ${repeat(stacks, (_, i) => i, (s, i) => html`
              <boardgame-component-stack
                layout="spatial"
                .stack="${s}"
                .spatialPositions="${this._layerPositions[i] || []}"
                .unsafeComponentAttrs="${this.unsafeComponentAttrs}"
                .componentView=${this.componentViews.length ? this.componentViews[i] : this.componentView}
                no-default-spacer>
              </boardgame-component-stack>
            `)}
          </div>
        ` : ''}
      </div>
    `;

    return html`
      <div id="board-wrapper" role="region" aria-label=${this.boardLabel}
        aria-describedby=${this.boardDescription ? 'board-description' : undefined}>
        ${this.boardDescription ? html`<span id="board-description" hidden>${this.boardDescription}</span>` : ''}
        <p id="status" role="status" aria-live="polite" ?hidden=${!this._loadError}>
          ${this._loadError ? html`Board artwork could not be loaded: ${this._loadError}
            <button type="button" @click=${this._loadBoardSource}>Retry</button>` : ''}
        </p>
        ${this.panZoom ? html`
          <boardgame-board-viewport
            label=${`${this.boardLabel} navigation`}
            .maxScale=${this.maxZoom}>
            ${boardScene}
          </boardgame-board-viewport>
        ` : boardScene}
        ${this._resolvedGeometry ? html`
          <details id="space-list">
            <summary>Board spaces</summary>
            <div>${repeat(this.action || this.placementDraft ? this._actionGeometrySpaces : this._resolvedGeometry.spaces, space => space.key, space => {
              const candidate = this._candidateForSpace(space.key);
              const placement = this._placementTargetForSpace(space.key);
              const legacyDisabled = this.disabledSpaces.includes(Number(space.key));
              const disabled = placement ? !placement.canPlace
                : candidate ? !candidate.action.canActivate : legacyDisabled;
              const reason = placement?.reason ?? candidate?.action.reason?.message;
              return html`<button
                type="button"
                aria-disabled=${String(disabled)}
                title=${reason ?? ''}
                @focus=${() => this._focusSpace(space)}
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

declare global {
  interface HTMLElementTagNameMap {
    'boardgame-spatial-board': BoardgameSpatialBoard;
  }
}
