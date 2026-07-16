import { LitElement, css, html } from 'lit';
import { property, query, state } from 'lit/decorators.js';

export interface BoardViewportChange {
  readonly scale: number;
  readonly x: number;
  readonly y: number;
}

/** Accessible pan/zoom shell for large board, map, SVG, or canvas scenes. */
class BoardgameBoardViewport extends LitElement {
  static override styles = css`
    :host {
      display: block;
      min-width: 0;
    }

    #toolbar {
      align-items: center;
      display: flex;
      gap: 0.35rem;
      justify-content: flex-end;
      margin-block-end: 0.4rem;
    }

    button {
      align-items: center;
      background: var(--board-viewport-control-background, Canvas);
      border: 1px solid var(--board-viewport-control-border, ButtonBorder);
      border-radius: var(--board-viewport-control-radius, 0.4rem);
      color: var(--board-viewport-control-color, CanvasText);
      cursor: pointer;
      display: inline-flex;
      font: inherit;
      justify-content: center;
      min-block-size: 2.5rem;
      min-inline-size: 2.5rem;
      padding: 0.35rem 0.65rem;
    }

    button:disabled {
      cursor: default;
      opacity: 0.5;
    }

    #viewport {
      cursor: grab;
      overflow: hidden;
      position: relative;
      touch-action: none;
    }

    #viewport:focus-visible {
      outline: 3px solid Highlight;
      outline-offset: 2px;
    }

    #viewport.dragging {
      cursor: grabbing;
      user-select: none;
    }

    #scene {
      min-width: 0;
      transform-origin: 0 0;
      will-change: transform;
    }

    #status {
      block-size: 1px;
      clip: rect(0 0 0 0);
      clip-path: inset(50%);
      inline-size: 1px;
      overflow: hidden;
      position: absolute;
      white-space: nowrap;
    }

    @media (forced-colors: active) {
      button { border-color: ButtonText; }
    }
  `;

  /** Accessible name for the navigable scene and its controls. */
  @property({ type: String })
  label = 'Board navigation';

  /** Largest allowed magnification. */
  @property({ type: Number, attribute: 'max-scale' })
  maxScale = 4;

  /** Magnification added or removed by buttons and keyboard controls. */
  @property({ type: Number, attribute: 'zoom-step' })
  zoomStep = 0.5;

  /** Show the built-in zoom/reset controls. Gestures and keyboard remain available. */
  @property({ type: Boolean, attribute: 'hide-controls' })
  hideControls = false;

  @state()
  private _scale = 1;

  @state()
  private _x = 0;

  @state()
  private _y = 0;

  @state()
  private _dragging = false;

  @query('#viewport')
  private _viewport!: HTMLDivElement;

  @query('#scene')
  private _scene!: HTMLDivElement;

  private _resizeObserver: ResizeObserver | null = null;
  private _pointers = new Map<number, { x: number; y: number }>();
  private _gestureMoved = false;
  private _gestureDistance = 0;
  private _suppressNextClick = false;

  /** Current immutable transform, useful for diagnostics and persistence. */
  get view(): BoardViewportChange {
    return Object.freeze({ scale: this._scale, x: this._x, y: this._y });
  }

  override connectedCallback(): void {
    super.connectedCallback();
    this._resizeObserver = new ResizeObserver(() => this._setView(this._scale, this._x, this._y));
    if (this.hasUpdated && this._viewport && this._scene) {
      this._resizeObserver.observe(this._viewport);
      this._resizeObserver.observe(this._scene);
    }
  }

  override disconnectedCallback(): void {
    super.disconnectedCallback();
    this._resizeObserver?.disconnect();
    this._resizeObserver = null;
    this._pointers.clear();
    this._dragging = false;
    this._gestureMoved = false;
    this._gestureDistance = 0;
    this._suppressNextClick = false;
  }

  override firstUpdated(): void {
    this._resizeObserver?.observe(this._viewport);
    this._resizeObserver?.observe(this._scene);
    this._viewport.addEventListener('click', this._captureClick, true);
  }

  override updated(changed: Map<PropertyKey, unknown>): void {
    if (changed.has('maxScale') || changed.has('zoomStep') || changed.has('label')) {
      this._validateConfiguration();
      this._setView(this._scale, this._x, this._y);
    }
  }

  /** Zoom in one configured step around the visual center. */
  zoomIn(): void {
    this._zoomAt(this._scale + this.zoomStep, this._viewportCenter());
  }

  /** Zoom out one configured step around the visual center. */
  zoomOut(): void {
    this._zoomAt(this._scale - this.zoomStep, this._viewportCenter());
  }

  /** Restore the full unpanned scene. */
  resetView(): void {
    this._setView(1, 0, 0);
  }

  /** Restore a previously captured view; out-of-bounds offsets are safely clamped. */
  setView(view: BoardViewportChange): void {
    if (!view || typeof view !== 'object'
      || !Number.isFinite(view.scale) || !Number.isFinite(view.x) || !Number.isFinite(view.y)) {
      throw new Error('boardgame-board-viewport: view must contain finite scale, x, and y numbers');
    }
    this._setView(view.scale, view.x, view.y);
  }

  /** Pan just enough to reveal an element inside the scene. */
  reveal(element: Element, padding = 24): void {
    if (!this._ownsElement(element)) {
      throw new Error('boardgame-board-viewport: reveal target must belong to the slotted scene');
    }
    if (!Number.isFinite(padding) || padding < 0) {
      throw new Error('boardgame-board-viewport: reveal padding must be a finite nonnegative number');
    }
    const viewport = this._viewport.getBoundingClientRect();
    const target = element.getBoundingClientRect();
    let x = this._x;
    let y = this._y;
    if (target.left < viewport.left + padding) x += viewport.left + padding - target.left;
    else if (target.right > viewport.right - padding) x -= target.right - (viewport.right - padding);
    if (target.top < viewport.top + padding) y += viewport.top + padding - target.top;
    else if (target.bottom > viewport.bottom - padding) y -= target.bottom - (viewport.bottom - padding);
    this._setView(this._scale, x, y);
  }

  private _ownsElement(element: Element): boolean {
    let node: Node | null = element;
    while (node) {
      if (node === this || this._scene.contains(node)) return true;
      const root: Node = node.getRootNode();
      node = root instanceof ShadowRoot ? root.host : null;
    }
    return false;
  }

  private _validateConfiguration(): void {
    if (typeof this.label !== 'string' || !this.label.trim()) {
      throw new Error('boardgame-board-viewport: label must be a non-empty string');
    }
    if (!Number.isFinite(this.maxScale) || this.maxScale < 1 || this.maxScale > 16) {
      throw new Error('boardgame-board-viewport: maxScale must be from 1 through 16');
    }
    if (!Number.isFinite(this.zoomStep) || this.zoomStep <= 0 || this.zoomStep > 4) {
      throw new Error('boardgame-board-viewport: zoomStep must be greater than 0 and at most 4');
    }
  }

  private _captureClick = (event: Event): void => {
    if (!this._suppressNextClick) return;
    this._suppressNextClick = false;
    event.preventDefault();
    event.stopImmediatePropagation();
  };

  private _viewportCenter(): { x: number; y: number } {
    const bounds = this._viewport.getBoundingClientRect();
    return { x: bounds.width / 2, y: bounds.height / 2 };
  }

  private _localPoint(clientX: number, clientY: number): { x: number; y: number } {
    const bounds = this._viewport.getBoundingClientRect();
    return { x: clientX - bounds.left, y: clientY - bounds.top };
  }

  private _zoomAt(nextScale: number, point: { x: number; y: number }): void {
    const scale = Math.min(this.maxScale, Math.max(1, nextScale));
    const contentX = (point.x - this._x) / this._scale;
    const contentY = (point.y - this._y) / this._scale;
    this._setView(scale, point.x - contentX * scale, point.y - contentY * scale);
  }

  private _setView(scale: number, x: number, y: number): void {
    if (!this._viewport || !this._scene) return;
    scale = Math.min(this.maxScale, Math.max(1, scale));
    const viewport = this._viewport.getBoundingClientRect();
    const width = this._scene.offsetWidth * scale;
    const height = this._scene.offsetHeight * scale;
    const minX = Math.min(0, viewport.width - width);
    const minY = Math.min(0, viewport.height - height);
    const nextX = width <= viewport.width ? (viewport.width - width) / 2 : Math.min(0, Math.max(minX, x));
    const nextY = height <= viewport.height ? (viewport.height - height) / 2 : Math.min(0, Math.max(minY, y));
    const changed = scale !== this._scale || nextX !== this._x || nextY !== this._y;
    this._scale = scale;
    this._x = nextX;
    this._y = nextY;
    if (changed) {
      this.dispatchEvent(new CustomEvent<BoardViewportChange>('board-viewport-change', {
        bubbles: true,
        composed: true,
        detail: this.view,
      }));
    }
  }

  private _pointerDown(event: PointerEvent): void {
    if (event.pointerType === 'mouse' && event.button !== 0) return;
    if (this._pointers.size === 0) {
      this._gestureMoved = false;
      this._gestureDistance = 0;
    }
    this._pointers.set(event.pointerId, this._localPoint(event.clientX, event.clientY));
    if (this._pointers.size === 2) {
      for (const pointerId of this._pointers.keys()) this._viewport.setPointerCapture(pointerId);
    }
  }

  private _pointerMove(event: PointerEvent): void {
    const previous = this._pointers.get(event.pointerId);
    if (!previous) return;
    const next = this._localPoint(event.clientX, event.clientY);
    if (this._pointers.size === 1) {
      this._pointers.set(event.pointerId, next);
      if (this._scale <= 1) return;
      const dx = next.x - previous.x;
      const dy = next.y - previous.y;
      this._gestureDistance += Math.hypot(dx, dy);
      if (this._gestureDistance >= 4) this._gestureMoved = true;
      if (!this._gestureMoved) return;
      if (!this._viewport.hasPointerCapture(event.pointerId)) this._viewport.setPointerCapture(event.pointerId);
      this._dragging = true;
      this._setView(this._scale, this._x + dx, this._y + dy);
      event.preventDefault();
      return;
    }
    const entries = [...this._pointers.entries()].slice(0, 2);
    const otherEntry = entries.find(([id]) => id !== event.pointerId);
    if (!otherEntry) return;
    const other = otherEntry[1];
    const oldDistance = Math.hypot(previous.x - other.x, previous.y - other.y);
    const newDistance = Math.hypot(next.x - other.x, next.y - other.y);
    const oldCenter = { x: (previous.x + other.x) / 2, y: (previous.y + other.y) / 2 };
    const newCenter = { x: (next.x + other.x) / 2, y: (next.y + other.y) / 2 };
    this._pointers.set(event.pointerId, next);
    if (oldDistance <= 0 || newDistance <= 0) return;
    const scale = Math.min(this.maxScale, Math.max(1, this._scale * newDistance / oldDistance));
    const contentX = (oldCenter.x - this._x) / this._scale;
    const contentY = (oldCenter.y - this._y) / this._scale;
    this._gestureMoved = true;
    this._dragging = true;
    this._setView(scale, newCenter.x - contentX * scale, newCenter.y - contentY * scale);
    event.preventDefault();
  }

  private _pointerEnd(event: PointerEvent): void {
    if (!this._pointers.delete(event.pointerId)) return;
    if (this._viewport.hasPointerCapture(event.pointerId)) this._viewport.releasePointerCapture(event.pointerId);
    if (this._pointers.size === 0) {
      this._dragging = false;
      this._suppressNextClick = this._gestureMoved;
      if (this._suppressNextClick) {
        setTimeout(() => { this._suppressNextClick = false; }, 0);
      }
      this._gestureMoved = false;
      this._gestureDistance = 0;
    }
  }

  private _wheel(event: WheelEvent): void {
    if (!event.ctrlKey && !event.metaKey) return;
    event.preventDefault();
    const direction = event.deltaY < 0 ? 1 : -1;
    this._zoomAt(this._scale + direction * this.zoomStep, this._localPoint(event.clientX, event.clientY));
  }

  private _keyDown(event: KeyboardEvent): void {
    if (event.target !== this._viewport) return;
    const pan = 40;
    if (event.key === '+' || event.key === '=') this.zoomIn();
    else if (event.key === '-' || event.key === '_') this.zoomOut();
    else if (event.key === '0' || event.key === 'Home') this.resetView();
    else if (event.key === 'ArrowLeft') this._setView(this._scale, this._x + pan, this._y);
    else if (event.key === 'ArrowRight') this._setView(this._scale, this._x - pan, this._y);
    else if (event.key === 'ArrowUp') this._setView(this._scale, this._x, this._y + pan);
    else if (event.key === 'ArrowDown') this._setView(this._scale, this._x, this._y - pan);
    else return;
    event.preventDefault();
  }

  override render() {
    this._validateConfiguration();
    const percentage = Math.round(this._scale * 100);
    return html`
      ${this.hideControls ? '' : html`
        <div id="toolbar" part="toolbar" role="group" aria-label=${`${this.label} controls`}>
          <button part="zoom-out" type="button" aria-label="Zoom out"
            ?disabled=${this._scale <= 1} @click=${this.zoomOut}>−</button>
          <button part="reset" type="button" @click=${this.resetView}
            ?disabled=${this._scale === 1 && this._x === 0 && this._y === 0}>Reset view</button>
          <button part="zoom-in" type="button" aria-label="Zoom in"
            ?disabled=${this._scale >= this.maxScale} @click=${this.zoomIn}>+</button>
        </div>
      `}
      <div id="viewport" part="viewport" class=${this._dragging ? 'dragging' : ''}
        role="region" aria-label=${this.label} tabindex="0"
        @pointerdown=${this._pointerDown} @pointermove=${this._pointerMove}
        @pointerup=${this._pointerEnd} @pointercancel=${this._pointerEnd}
        @wheel=${this._wheel} @keydown=${this._keyDown}>
        <div id="scene" part="scene"
          style=${`transform: translate(${this._x}px, ${this._y}px) scale(${this._scale})`}>
          <slot></slot>
        </div>
      </div>
      <span id="status" role="status" aria-live="polite">${percentage}% zoom</span>
    `;
  }
}

customElements.define('boardgame-board-viewport', BoardgameBoardViewport);

export { BoardgameBoardViewport };
