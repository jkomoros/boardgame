import { LitElement, css, html, nothing } from 'lit';
import { property, query, state } from 'lit/decorators.js';
import type { PlacementItemBinding } from '../moves/placement-draft.js';
import type { TargetKey } from '../moves/target-action.js';

/** Accessible selector around arbitrary game-owned placement-item visuals. */
export class BoardgamePlacementItem extends LitElement {
  static override styles = css`
    :host { display: inline-block; }
    button {
      box-sizing: border-box;
      display: grid;
      min-width: 44px;
      min-height: 44px;
      width: 100%;
      padding: var(--boardgame-placement-item-padding, 0.35rem);
      border: var(--boardgame-placement-item-border, 1px solid currentColor);
      border-radius: var(--boardgame-placement-item-radius, 0.6rem);
      background: var(--boardgame-placement-item-background, transparent);
      color: inherit;
      font: inherit;
      cursor: pointer;
    }
    button[data-placed] {
      background: var(--boardgame-placement-item-placed-background, color-mix(in srgb, currentColor 7%, transparent));
    }
    button[aria-pressed='true'] {
      border: var(--boardgame-placement-item-selected-border, 3px solid currentColor);
      background: var(--boardgame-placement-item-selected-background, color-mix(in srgb, currentColor 12%, transparent));
    }
    button:disabled { cursor: not-allowed; opacity: var(--boardgame-placement-item-disabled-opacity, 0.5); }
    button:focus-visible { outline: 3px solid Highlight; outline-offset: 3px; }
    .status { position: absolute; width: 1px; height: 1px; padding: 0; margin: -1px;
      overflow: hidden; clip: rect(0, 0, 0, 0); white-space: nowrap; border: 0; }
    @media (forced-colors: active) {
      button[aria-pressed='true'] { outline: 3px solid Highlight; outline-offset: -5px; }
      button[data-placed] { border-style: dashed; }
    }
  `;

  @property({ attribute: false })
  item: PlacementItemBinding<TargetKey, TargetKey> | null = null;

  /** Accessible name; also visible fallback when no presentation is slotted. */
  @property({ type: String })
  label = '';

  @property({ type: Boolean, reflect: true })
  disabled = false;

  @query('slot')
  private _slot!: HTMLSlotElement;

  @state()
  private _hasContent = false;

  protected override updated(): void {
    this._validateSlottedContent();
  }

  override render() {
    const item = this._validated();
    const placed = item.placedAt !== null;
    return html`
      <button
        part="button"
        type="button"
        aria-label=${this.label.trim()}
        aria-pressed=${String(item.selected)}
        aria-describedby=${placed ? 'status' : nothing}
        ?data-placed=${placed}
        ?disabled=${this.disabled || item.capacityBlocked}
        @click=${this._select}>
        <slot part="content" @slotchange=${this._contentChanged}></slot>
        ${this._hasContent ? nothing : html`<span part="fallback">${this.label.trim()}</span>`}
        ${placed ? html`<span id="status" class="status" part="status">Placed</span>` : nothing}
      </button>
    `;
  }

  private _validated(): PlacementItemBinding<TargetKey, TargetKey> {
    if (!this.label.trim()) throw new Error('boardgame-placement-item: label must be non-empty');
    const item = this.item;
    if (!item) throw new Error('boardgame-placement-item: .item must come from draft.item(key)');
    if ((typeof item.item !== 'string' && typeof item.item !== 'number')
      || (typeof item.item === 'number' && !Number.isFinite(item.item))
      || (item.placedAt !== null && typeof item.placedAt !== 'string' && typeof item.placedAt !== 'number')
      || (typeof item.placedAt === 'number' && !Number.isFinite(item.placedAt))
      || typeof item.selected !== 'boolean' || typeof item.capacityBlocked !== 'boolean'
      || typeof item.select !== 'function' || typeof item.remove !== 'function'
      || (item.selected && item.capacityBlocked) || (item.placedAt !== null && item.capacityBlocked)) {
      throw new Error('boardgame-placement-item: .item is malformed');
    }
    return item;
  }

  private readonly _select = (): void => {
    this._validated().select();
  };

  private readonly _contentChanged = (event: Event): void => {
    const slot = event.currentTarget as HTMLSlotElement;
    this._hasContent = slot.assignedNodes({ flatten: true }).some(node =>
      node.nodeType === Node.ELEMENT_NODE || Boolean(node.textContent?.trim()));
    this._validateSlottedContent();
  };

  private _validateSlottedContent(): void {
    const interactive = 'a[href],button,input,select,textarea,[contenteditable]:not([contenteditable="false"]),[tabindex]:not([tabindex="-1"])';
    for (const element of this._slot.assignedElements({ flatten: true })) {
      if (element.matches(interactive) || element.querySelector(interactive)) {
        throw new Error('boardgame-placement-item: slotted presentation cannot contain interactive content');
      }
    }
  }
}

customElements.define('boardgame-placement-item', BoardgamePlacementItem);

declare global {
  interface HTMLElementTagNameMap {
    'boardgame-placement-item': BoardgamePlacementItem;
  }
}
