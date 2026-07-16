import { LitElement, css, html, nothing } from 'lit';
import { property, query, state } from 'lit/decorators.js';
import type { TargetKey } from '../moves/target-action.js';

export interface SelectionOptionBinding {
  readonly candidates: readonly TargetKey[];
  readonly selected: readonly TargetKey[];
  readonly maximumSelected: number;
  toggle(key: TargetKey): void;
  isSelected(key: TargetKey): boolean;
}

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
  draft: SelectionOptionBinding | null = null;

  @property({ attribute: false })
  choice: TargetKey | null = null;

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
    const { draft, choice, selected } = this._validated();
    const capacityBlocked = !selected && draft.selected.length >= draft.maximumSelected;
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

  private _validated(): { draft: SelectionOptionBinding; choice: TargetKey; selected: boolean } {
    if (!this.label.trim()) throw new Error('boardgame-selection-option: label must be non-empty');
    const draft = this.draft;
    if (!draft || !Array.isArray(draft.candidates) || !Array.isArray(draft.selected)
      || !Number.isSafeInteger(draft.maximumSelected) || draft.maximumSelected < 1
      || draft.selected.length > draft.maximumSelected
      || typeof draft.toggle !== 'function' || typeof draft.isSelected !== 'function') {
      throw new Error('boardgame-selection-option: .draft must be a SelectionDraftController binding');
    }
    const choice = this.choice;
    if ((typeof choice !== 'string' && typeof choice !== 'number')
      || (typeof choice === 'number' && !Number.isFinite(choice))) {
      throw new Error('boardgame-selection-option: .choice must be a finite string or number');
    }
    if (!draft.candidates.includes(choice)) {
      throw new Error(`boardgame-selection-option: choice ${JSON.stringify(choice)} is not a draft candidate`);
    }
    const selected = draft.isSelected(choice);
    if (selected !== draft.selected.includes(choice)) {
      throw new Error('boardgame-selection-option: draft selected state is inconsistent');
    }
    return { draft, choice, selected };
  }

  private readonly _toggle = (): void => {
    const { draft, choice } = this._validated();
    draft.toggle(choice);
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
