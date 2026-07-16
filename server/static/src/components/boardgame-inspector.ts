import { LitElement, css, html, nothing } from 'lit';
import { property, query, state } from 'lit/decorators.js';

export type InspectorChangeReason = 'backdrop' | 'close-button' | 'escape' | 'programmatic' | 'trigger';

export interface InspectorOpenChangedDetail {
  readonly open: boolean;
  readonly reason: InspectorChangeReason;
}

/** Zero-JavaScript thumbnail-to-modal inspection for cards, art, and boards. */
export class BoardgameInspector extends LitElement {
  static override styles = css`
    :host { display: inline-block; }

    #trigger {
      appearance: none;
      display: inline-grid;
      min-width: 44px;
      min-height: 44px;
      padding: var(--boardgame-inspector-trigger-padding, 0.25rem);
      border: var(--boardgame-inspector-trigger-border, 1px solid transparent);
      border-radius: var(--boardgame-inspector-trigger-radius, 0.5rem);
      background: var(--boardgame-inspector-trigger-background, transparent);
      color: inherit;
      font: inherit;
      cursor: zoom-in;
    }

    #trigger:focus-visible, #close:focus-visible {
      outline: 3px solid Highlight;
      outline-offset: 3px;
    }

    dialog {
      box-sizing: border-box;
      width: min(var(--boardgame-inspector-width, 48rem), calc(100vw - 2rem));
      max-width: none;
      max-height: min(var(--boardgame-inspector-max-height, 52rem), calc(100dvh - 2rem));
      padding: 0;
      border: var(--boardgame-inspector-border, 0);
      border-radius: var(--boardgame-inspector-radius, 1rem);
      background: var(--boardgame-inspector-background, var(--md-sys-color-surface-container-high, white));
      color: var(--boardgame-inspector-color, var(--md-sys-color-on-surface, #1d1b20));
      box-shadow: var(--boardgame-inspector-shadow, 0 1rem 3rem rgb(0 0 0 / 0.35));
      overflow: hidden;
    }

    dialog::backdrop {
      background: var(--boardgame-inspector-backdrop, rgb(0 0 0 / 0.62));
      backdrop-filter: blur(var(--boardgame-inspector-backdrop-blur, 2px));
    }

    #panel { display: grid; max-height: inherit; }

    #header {
      display: flex;
      align-items: flex-start;
      justify-content: space-between;
      gap: 1rem;
      padding: var(--boardgame-inspector-header-padding, 1rem 1rem 0.5rem);
    }

    #title { margin: 0; font: inherit; font-size: 1.25rem; font-weight: 700; }
    #description { margin: 0.25rem 0 0; color: var(--md-sys-color-on-surface-variant, #49454f); }

    #close {
      appearance: none;
      flex: none;
      min-width: 44px;
      min-height: 44px;
      border: 0;
      border-radius: 999px;
      background: transparent;
      color: inherit;
      font: inherit;
      font-size: 1.5rem;
      cursor: pointer;
    }

    #content {
      min-height: 0;
      overflow: auto;
      overscroll-behavior: contain;
      padding: var(--boardgame-inspector-content-padding, 0.5rem 1rem 1rem);
    }

    .visually-hidden {
      block-size: 1px;
      clip: rect(0 0 0 0);
      clip-path: inset(50%);
      inline-size: 1px;
      overflow: hidden;
      position: absolute;
      white-space: nowrap;
    }

    @media (max-width: 36rem) {
      dialog {
        width: 100vw;
        max-height: 100dvh;
        margin: auto 0 0;
        border-radius: var(--boardgame-inspector-mobile-radius, 1rem 1rem 0 0);
      }
    }

    @media (prefers-reduced-motion: reduce) {
      dialog::backdrop { backdrop-filter: none; }
    }

    @media (forced-colors: active) {
      #trigger, dialog { border: 1px solid CanvasText; }
    }
  `;

  /** Visible dialog title and default accessible trigger name. */
  @property({ type: String })
  label = '';

  @property({ type: String })
  description = '';

  @property({ type: String, attribute: 'trigger-label' })
  triggerLabel = '';

  @property({ type: Boolean, reflect: true })
  open = false;

  @property({ type: Boolean })
  dismissible = true;

  @query('dialog')
  private _dialog!: HTMLDialogElement;

  @query('slot[name="detail"]')
  private _detailSlot!: HTMLSlotElement;

  @query('slot[name="thumbnail"]')
  private _thumbnailSlot!: HTMLSlotElement;

  @state()
  private _hasThumbnail = false;

  private _nextReason: InspectorChangeReason = 'programmatic';

  override disconnectedCallback(): void {
    if (this._dialog?.open) this._dialog.close();
    this.open = false;
    super.disconnectedCallback();
  }

  protected override updated(changedProperties: Map<PropertyKey, unknown>): void {
    super.updated(changedProperties);
    this._validateLabels();
    this._validateThumbnail();
    if (!changedProperties.has('open')) return;
    if (this.open && !this._dialog.open) {
      if (!this._hasDetail()) {
        this.open = false;
        throw new Error('boardgame-inspector: provide non-empty slot="detail" content before opening');
      }
      this._dialog.showModal();
      this._emitChange(true, this._nextReason);
    } else if (!this.open && this._dialog.open) {
      this._dialog.close();
    }
  }

  show(reason: 'programmatic' | 'trigger' = 'programmatic'): void {
    if (!this.isConnected) throw new Error('boardgame-inspector: connect the element before calling show()');
    if (reason !== 'programmatic' && reason !== 'trigger') {
      throw new Error('boardgame-inspector: show reason must be "programmatic" or "trigger"');
    }
    this._nextReason = reason;
    this.open = true;
  }

  close(reason: 'backdrop' | 'close-button' | 'programmatic' = 'programmatic'): void {
    if (reason !== 'programmatic' && reason !== 'close-button' && reason !== 'backdrop') {
      throw new Error('boardgame-inspector: invalid close reason');
    }
    this._nextReason = reason;
    this.open = false;
  }

  override render() {
    const triggerName = this.triggerLabel.trim() || `Inspect ${this.label.trim()}`;
    return html`
      <button
        id="trigger"
        part="trigger"
        type="button"
        aria-label=${triggerName}
        aria-haspopup="dialog"
        aria-expanded=${String(this.open)}
        @click=${this._showFromTrigger}>
        <slot name="thumbnail" @slotchange=${this._thumbnailChanged}></slot>
        ${this._hasThumbnail ? nothing : html`<span>${triggerName}</span>`}
      </button>
      <dialog
        part="dialog"
        aria-labelledby="title"
        aria-describedby=${this.description.trim() ? 'description' : nothing}
        @cancel=${this._cancelled}
        @close=${this._closed}
        @click=${this._backdropClicked}>
        <div id="panel" part="panel">
          <header id="header" part="header">
            <div>
              <h2 id="title" part="title">${this.label.trim()}</h2>
              ${this.description.trim()
                ? html`<p id="description" part="description">${this.description.trim()}</p>`
                : nothing}
            </div>
            <button id="close" part="close" type="button" aria-label="Close" @click=${this._closeButton}>
              <span aria-hidden="true">×</span>
            </button>
          </header>
          <div id="content" part="content"><slot name="detail"></slot></div>
        </div>
      </dialog>
    `;
  }

  private readonly _showFromTrigger = (): void => this.show('trigger');
  private readonly _closeButton = (): void => this.close('close-button');

  private readonly _cancelled = (event: Event): void => {
    if (!this.dismissible) {
      event.preventDefault();
      return;
    }
    this._nextReason = 'escape';
  };

  private readonly _closed = (): void => {
    if (this.open) this.open = false;
    const reason = this._nextReason;
    this._nextReason = 'programmatic';
    this._emitChange(false, reason);
    queueMicrotask(() => this.shadowRoot?.querySelector<HTMLButtonElement>('#trigger')?.focus());
  };

  private readonly _backdropClicked = (event: MouseEvent): void => {
    if (event.target !== this._dialog || !this.dismissible) return;
    const bounds = this._dialog.getBoundingClientRect();
    const inside = event.clientX >= bounds.left && event.clientX <= bounds.right
      && event.clientY >= bounds.top && event.clientY <= bounds.bottom;
    if (!inside) this.close('backdrop');
  };

  private readonly _thumbnailChanged = (event: Event): void => {
    const slot = event.currentTarget as HTMLSlotElement;
    this._hasThumbnail = slot.assignedNodes({ flatten: true }).some(hasMeaningfulContent);
  };

  private _hasDetail(): boolean {
    return this._detailSlot.assignedNodes({ flatten: true }).some(hasMeaningfulContent);
  }

  private _validateLabels(): void {
    if (!this.label.trim()) throw new Error('boardgame-inspector: label must be a non-empty visible dialog title');
    if (this.triggerLabel && !this.triggerLabel.trim()) {
      throw new Error('boardgame-inspector: triggerLabel must be omitted or non-empty');
    }
  }

  private _validateThumbnail(): void {
    const interactive = 'a[href],button,input,select,textarea,[contenteditable]:not([contenteditable="false"]),[tabindex]:not([tabindex="-1"])';
    for (const element of this._thumbnailSlot.assignedElements({ flatten: true })) {
      if (element.matches(interactive) || element.querySelector(interactive)) {
        throw new Error('boardgame-inspector: slot="thumbnail" is already wrapped in a button and cannot contain interactive content');
      }
    }
  }

  private _emitChange(open: boolean, reason: InspectorChangeReason): void {
    this.dispatchEvent(new CustomEvent<InspectorOpenChangedDetail>('inspector-open-changed', {
      bubbles: true,
      composed: true,
      detail: Object.freeze({ open, reason }),
    }));
  }
}

function hasMeaningfulContent(node: Node): boolean {
  return node.nodeType === Node.ELEMENT_NODE || Boolean(node.textContent?.trim());
}

customElements.define('boardgame-inspector', BoardgameInspector);

declare global {
  interface HTMLElementTagNameMap {
    'boardgame-inspector': BoardgameInspector;
  }
  interface HTMLElementEventMap {
    'inspector-open-changed': CustomEvent<InspectorOpenChangedDetail>;
  }
}
