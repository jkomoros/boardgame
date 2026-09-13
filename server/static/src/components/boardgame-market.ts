import { LitElement, css, html, nothing } from 'lit';
import { property, query, state } from 'lit/decorators.js';
import { repeat } from 'lit/directives/repeat.js';
import { styleMap } from 'lit/directives/style-map.js';
import type { BoundMoveAction } from '../moves/action.js';
import type { SelectionDraftSelectionBinding } from '../moves/selection-draft.js';
import type { TargetAction } from '../moves/target-action.js';
import type { ExpandedStack } from '../types/boardgame-types.js';
import './boardgame-component-zone.js';
import './boardgame-deck.js';
import type { BoardgameComponentStack, StackSlotGeometryChangedDetail } from './boardgame-component-stack.js';
import type { ComponentView } from './component-view.js';

export type MarketAttachmentPosition = 'before' | 'after';

/** A draw deck beside one stable, non-wrapping row of visible market slots. */
export class BoardgameMarket extends LitElement {
  static override styles = css`
    :host {
      display: block;
      min-width: 0;
      container-type: inline-size;
    }

    #market { min-width: 0; }

    #label {
      display: block;
      margin-block-end: 0.35rem;
      font: inherit;
      font-weight: 650;
    }

    #body {
      display: grid;
      grid-template-columns: max-content minmax(0, 1fr);
      align-items: start;
      gap: var(--boardgame-market-gap, 0.75rem);
      min-width: 0;
    }

    #display-scroll {
      min-width: 0;
      max-width: 100%;
      overflow-x: auto;
      overscroll-behavior-inline: contain;
    }

    #display-track {
      box-sizing: border-box;
      width: max-content;
      min-width: 100%;
    }

    #display-zone {
      box-sizing: border-box;
      width: 100%;
      --boardgame-stack-flex-wrap: nowrap;
    }

    #attachments {
      display: grid;
      grid-template-columns: repeat(var(--boardgame-market-slots), var(--boardgame-market-slot-inline-size, max-content));
      width: max-content;
      min-width: 100%;
      align-items: start;
      box-sizing: border-box;
      padding-inline: var(--boardgame-market-content-inset, 0px);
    }

    #attachment-label {
      display: block;
      padding-inline: var(--boardgame-market-content-inset, 0px);
      font-size: 0.78rem;
      opacity: 0.72;
    }

    .attachment-cell {
      box-sizing: border-box;
      min-width: 0;
      width: var(--boardgame-market-slot-inline-size, max-content);
    }

    .attachment-cell boardgame-component-stack,
    ::slotted([slot^='attachment-']) {
      display: block;
      box-sizing: border-box;
      width: 100%;
      --component-width: var(--boardgame-market-component-inline-size);
      --component-scale: 1;
    }

    @container (max-width: 36rem) {
      #body {
        grid-template-columns: minmax(0, 1fr);
      }
    }
  `;

  @property({ type: String }) label = '';
  @property({ type: Number, attribute: 'heading-level' }) headingLevel = 2;
  @property({ type: String, attribute: 'source-label' }) sourceLabel = 'Deck';
  @property({ type: String, attribute: 'display-label' }) displayLabel = 'Market';
  @property({ type: String, attribute: 'source-empty-label' }) sourceEmptyLabel = 'Exhausted';
  @property({ attribute: false }) sourceStack: ExpandedStack | null | undefined = null;
  @property({ attribute: false }) sourceView: ComponentView | null = null;
  @property({ attribute: false }) stack: ExpandedStack | null | undefined = null;
  @property({ attribute: false }) componentView: ComponentView | null = null;
  @property({ type: Array, attribute: false })
  componentActions: readonly (BoundMoveAction<string, object> | null)[] = [];
  @property({ attribute: false }) action: TargetAction<number> | null = null;
  @property({ attribute: false })
  selection: SelectionDraftSelectionBinding<number> | SelectionDraftSelectionBinding<string> | null = null;
  @property({ attribute: false }) attachmentStack: ExpandedStack | null | undefined = null;
  @property({ attribute: false }) attachmentView: ComponentView | null = null;
  @property({ type: String, attribute: 'attachment-label' }) attachmentLabel = '';
  @property({ type: String, attribute: 'attachment-position' })
  attachmentPosition: MarketAttachmentPosition = 'after';

  @state() private _slotInlineSize = 0;
  @state() private _componentInlineSize = 0;
  @state() private _contentInset = 0;
  @query('#display-zone') private _displayZone!: HTMLElement & { stackElement?: BoardgameComponentStack | null };
  @query('#display-track') private _displayTrack!: HTMLElement;
  private _displayResizeObserver: ResizeObserver | null = null;

  /** The unchanged visible stack host, for specialized motion or presentation policy. */
  get stackElement(): BoardgameComponentStack | null {
    return this._displayZone?.stackElement ?? null;
  }

  private get _slotCount(): number {
    return this.stack?.Components.length ?? 0;
  }

  private _attachmentAt(index: number): ExpandedStack | null {
    const source = this.attachmentStack;
    if (!source || index >= source.Components.length) return null;
    const component = source.Components[index] ?? null;
    return Object.freeze({
      ...source,
      Indexes: Object.freeze(source.Indexes.slice(index, index + 1)),
      IDs: Object.freeze(source.IDs.slice(index, index + 1)),
      Components: Object.freeze([component]),
      Size: component === null ? 0 : 1,
      MaxSize: 1,
    });
  }

  private _displayGeometryChanged(event: CustomEvent<StackSlotGeometryChangedDetail>): void {
    const geometry = event.detail.geometry;
    this._slotInlineSize = geometry.inlineSize;
    this._componentInlineSize = geometry.componentInlineSize;
    this._measureDisplayInset();
  }

  override connectedCallback(): void {
    super.connectedCallback();
    if (this.hasUpdated) queueMicrotask(() => this._observeDisplaySize());
  }

  override firstUpdated(): void {
    this._observeDisplaySize();
  }

  private _observeDisplaySize(): void {
    this._displayResizeObserver?.disconnect();
    this._displayResizeObserver = new ResizeObserver(() => this._measureDisplayInset());
    this._displayResizeObserver.observe(this._displayZone);
    this._measureDisplayInset();
  }

  override disconnectedCallback(): void {
    this._displayResizeObserver?.disconnect();
    this._displayResizeObserver = null;
    super.disconnectedCallback();
  }

  private _measureDisplayInset(): void {
    const stack = this.stackElement;
    if (!stack || !this._displayTrack) return;
    const inset = stack.getBoundingClientRect().left - this._displayTrack.getBoundingClientRect().left;
    if (Number.isFinite(inset) && inset >= 0 && this._contentInset !== inset) this._contentInset = inset;
  }

  private _renderAttachments() {
    if (!this.attachmentStack && !this.attachmentLabel) return nothing;
    return html`
      ${this.attachmentLabel
        ? html`<span id="attachment-label">${this.attachmentLabel.trim()}</span>`
        : nothing}
      <div id="attachments" part="attachments" role="group"
        aria-labelledby=${this.attachmentLabel ? 'attachment-label' : nothing}>
        ${repeat(Array.from({ length: this._slotCount }, (_unused, index) => index), index => index, index => html`
          <div class="attachment-cell" part="attachment-cell">
            ${this.attachmentStack ? html`
              <boardgame-component-stack
                layout="stack"
                no-default-spacer
                components-disabled
                .stack=${this._attachmentAt(index)}
                .componentView=${this.attachmentView}>
              </boardgame-component-stack>
            ` : nothing}
            <slot name=${`attachment-${index}`}></slot>
          </div>
        `)}
      </div>
    `;
  }

  override render() {
    this._validateConfiguration();
    const attachments = this._renderAttachments();
    const childHeading = Math.min(6, this.headingLevel + 1);
    return html`
      <section id="market" part="market" aria-labelledby="label" style=${styleMap({
        '--boardgame-market-slots': String(this._slotCount),
        '--boardgame-market-slot-inline-size': this._slotInlineSize ? `${this._slotInlineSize}px` : undefined,
        '--boardgame-market-component-inline-size': this._componentInlineSize ? `${this._componentInlineSize}px` : undefined,
        '--boardgame-market-content-inset': this._contentInset ? `${this._contentInset}px` : undefined,
      })}>
        <span id="label" part="label" role="heading" aria-level=${this.headingLevel}>${this.label.trim()}</span>
        <div id="body" part="body">
          <boardgame-deck
            part="deck"
            .label=${this.sourceLabel}
            .emptyLabel=${this.sourceEmptyLabel}
            .headingLevel=${childHeading}
            .stack=${this.sourceStack}
            .componentView=${this.sourceView}>
          </boardgame-deck>
          <div id="display-scroll" part="scroller">
            <div id="display-track">
              ${this.attachmentPosition === 'before' ? attachments : nothing}
              <boardgame-component-zone id="display-zone" part="display"
                .label=${this.displayLabel}
                .headingLevel=${childHeading}
                layout="grid"
                hide-empty-state
                .stack=${this.stack}
                .componentView=${this.componentView}
                .componentActions=${this.componentActions}
                .action=${this.action}
                .selection=${this.selection}
                @stack-slot-geometry-changed=${this._displayGeometryChanged}>
              </boardgame-component-zone>
              ${this.attachmentPosition === 'after' ? attachments : nothing}
            </div>
          </div>
        </div>
      </section>
    `;
  }

  private _validateConfiguration(): void {
    if (!this.label.trim()) throw new Error('boardgame-market: label must be non-empty');
    if (!this.sourceLabel.trim()) throw new Error('boardgame-market: sourceLabel must be non-empty');
    if (!this.displayLabel.trim()) throw new Error('boardgame-market: displayLabel must be non-empty');
    if (!Number.isSafeInteger(this.headingLevel) || this.headingLevel < 1 || this.headingLevel > 6) {
      throw new Error('boardgame-market: headingLevel must be a safe integer from 1 through 6');
    }
    if (this.attachmentPosition !== 'before' && this.attachmentPosition !== 'after') {
      throw new Error('boardgame-market: attachmentPosition must be "before" or "after"');
    }
    if (this.attachmentStack && !this.attachmentView) {
      throw new Error('boardgame-market: attachmentStack requires attachmentView');
    }
    if ((this.attachmentStack?.Components.length ?? 0) > this._slotCount) {
      throw new Error('boardgame-market: attachmentStack cannot contain more slots than the visible market stack');
    }
  }
}

customElements.define('boardgame-market', BoardgameMarket);

declare global {
  interface HTMLElementTagNameMap {
    'boardgame-market': BoardgameMarket;
  }
}
