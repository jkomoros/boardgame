import { LitElement, css, html, nothing } from 'lit';
import { property, query, state } from 'lit/decorators.js';
import type { SelectionOptionBinding } from '../moves/selection-draft.js';
import type { TargetKey } from '../moves/target-action.js';

/** Accessible toggle around arbitrary game-owned selection presentation. */
export class BoardgameSelectionOption extends LitElement {
  static override styles = css`
    :host { display: inline-block; }
    button {
      box-sizing: border-box;
      display: grid;
      min-width: 44px;
      min-height: 44px;
      width: 100%;
      padding: var(--boardgame-selection-option-padding, 0.35rem);
      border: var(--boardgame-selection-option-border, 1px solid currentColor);
      border-radius: var(--boardgame-selection-option-radius, 0.6rem);
      background: var(--boardgame-selection-option-background, transparent);
      color: inherit;
      font: inherit;
      cursor: pointer;
    }
    button[aria-pressed='true'] {
      border: var(--boardgame-selection-option-selected-border, 3px solid currentColor);
      background: var(--boardgame-selection-option-selected-background, color-mix(in srgb, currentColor 12%, transparent));
    }
    button:disabled { cursor: not-allowed; opacity: var(--boardgame-selection-option-disabled-opacity, 0.5); }
    button:focus-visible { outline: 3px solid Highlight; outline-offset: 3px; }
    @media (forced-colors: active) {
      button[aria-pressed='true'] { outline: 3px solid Highlight; outline-offset: -5px; }
    }
  `;

  @property({ attribute: false })
  option: SelectionOptionBinding<TargetKey> | null = null;

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
    const { selected, capacityBlocked } = this._validated();
    return html`
      <button
        part="button"
        type="button"
        aria-label=${this.label.trim()}
        aria-pressed=${String(selected)}
        ?disabled=${this.disabled || capacityBlocked}
        @click=${this._toggle}>
        <slot part="content" @slotchange=${this._contentChanged}></slot>
        ${this._hasContent ? nothing : html`<span part="fallback">${this.label.trim()}</span>`}
      </button>
    `;
  }

  private _validated(): SelectionOptionBinding<TargetKey> {
    if (!this.label.trim()) throw new Error('boardgame-selection-option: label must be non-empty');
    const option = this.option;
    if (!option) {
      throw new Error('boardgame-selection-option: .option must come from draft.option(choice)');
    }
    const choice = option?.choice;
    if ((typeof choice !== 'string' && typeof choice !== 'number')
      || (typeof choice === 'number' && !Number.isFinite(choice))) {
      throw new Error('boardgame-selection-option: .option must come from draft.option(choice)');
    }
    if (typeof option.selected !== 'boolean' || typeof option.capacityBlocked !== 'boolean'
      || typeof option.toggle !== 'function' || (option.selected && option.capacityBlocked)) {
      throw new Error('boardgame-selection-option: .option is malformed');
    }
    return option;
  }

  private readonly _toggle = (): void => {
    this._validated().toggle();
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
        throw new Error('boardgame-selection-option: slotted presentation cannot contain interactive content');
      }
    }
  }
}

customElements.define('boardgame-selection-option', BoardgameSelectionOption);

declare global {
  interface HTMLElementTagNameMap {
    'boardgame-selection-option': BoardgameSelectionOption;
  }
}
