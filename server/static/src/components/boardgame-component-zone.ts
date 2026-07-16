import { LitElement, css, html } from 'lit';
import { property } from 'lit/decorators.js';
import type { BoundMoveAction } from '../moves/action.js';
import type { ExpandedStack } from '../types/boardgame-types.js';
import type { ComponentView } from './component-view.js';
import './boardgame-component-stack.js';
import type { StackLayout } from './boardgame-component-stack.js';

export type ComponentZoneLayout = Exclude<StackLayout, 'board' | 'spatial'>;

const zoneLayouts = new Set<ComponentZoneLayout>(['fan', 'grid', 'pile', 'spread', 'stack']);

/**
 * A semantic, responsive presentation wrapper for an ordinary component stack.
 * Board and spatial layouts remain on their dedicated lower-level components.
 */
export class BoardgameComponentZone extends LitElement {
  static override styles = css`
    :host {
      display: block;
      min-width: 0;
      container-type: inline-size;
    }

    #zone {
      box-sizing: border-box;
      min-width: 0;
      min-block-size: var(--boardgame-zone-min-block-size, 8rem);
      padding: var(--boardgame-zone-padding, 0.75rem);
      border: var(--boardgame-zone-border, 1px solid color-mix(in srgb, currentColor 18%, transparent));
      border-radius: var(--boardgame-zone-radius, 0.75rem);
      background: var(--boardgame-zone-background, color-mix(in srgb, currentColor 4%, transparent));
    }

    #heading {
      display: flex;
      align-items: center;
      justify-content: space-between;
      gap: 0.75rem;
      margin-block-end: var(--boardgame-zone-gap, 0.5rem);
    }

    #label {
      min-width: 0;
      font: inherit;
      font-weight: 650;
    }

    #count {
      box-sizing: border-box;
      min-width: 1.75rem;
      padding: 0.125rem 0.5rem;
      border-radius: 999px;
      text-align: center;
      font-variant-numeric: tabular-nums;
      background: var(--boardgame-zone-count-background, color-mix(in srgb, currentColor 12%, transparent));
    }

    #content,
    boardgame-component-stack {
      display: block;
      min-width: 0;
    }

    #empty {
      box-sizing: border-box;
      display: grid;
      min-block-size: 3rem;
      place-items: center;
      color: var(--boardgame-zone-empty-color, color-mix(in srgb, currentColor 62%, transparent));
      font-style: italic;
    }

    @container (max-width: 24rem) {
      #zone {
        padding: var(--boardgame-zone-compact-padding, 0.5rem);
      }
    }
  `;

  /** Visible and accessible zone name. */
  @property({ type: String })
  label = '';

  /** Heading depth for document-outline integration. */
  @property({ type: Number, attribute: 'heading-level' })
  headingLevel = 2;

  @property({ type: String })
  layout: ComponentZoneLayout = 'stack';

  @property({ attribute: false })
  stack: ExpandedStack | null | undefined = null;

  @property({ attribute: false })
  componentView: ComponentView | null = null;

  @property({ type: Array, attribute: false })
  componentActions: readonly (BoundMoveAction<string, object> | null)[] = [];

  @property({ type: Boolean })
  messy = false;

  @property({ type: Number })
  messiness = 1;

  @property({ type: Boolean, attribute: 'no-default-spacer' })
  noDefaultSpacer = false;

  @property({ type: Number, attribute: 'faux-components' })
  fauxComponents = 0;

  @property({ type: Number })
  stagger = 0;

  /** A useful empty state is automatic; opt out only when custom content replaces it. */
  @property({ type: Boolean, attribute: 'hide-empty-state' })
  hideEmptyState = false;

  @property({ type: String, attribute: 'empty-label' })
  emptyLabel = 'Empty';

  @property({ type: Boolean, attribute: 'hide-count' })
  hideCount = false;

  private get _occupiedCount(): number {
    return this.stack?.Components.reduce((count, component) => count + (component === null ? 0 : 1), 0) ?? 0;
  }

  override render() {
    this._validateConfiguration();
    const count = this._occupiedCount;
    return html`
      <section id="zone" part="zone" aria-labelledby="label">
        <div id="heading" part="heading">
          <span id="label" part="label" role="heading" aria-level=${this.headingLevel}>${this.label.trim()}</span>
          <slot name="heading-actions"></slot>
          ${this.hideCount ? null : html`<span id="count" part="count" aria-label=${`${count} items`}>${count}</span>`}
        </div>
        <div id="content" part="content">
          <boardgame-component-stack
            part="stack"
            .stack=${this.stack}
            .componentView=${this.componentView}
            .componentActions=${this.componentActions}
            .layout=${this.layout}
            .messy=${this.messy}
            .messiness=${this.messiness}
            .noDefaultSpacer=${this.noDefaultSpacer}
            .fauxComponents=${this.fauxComponents}
            .stagger=${this.stagger}
            .componentsDisabled=${this.componentActions.length === 0}>
          </boardgame-component-stack>
          ${count === 0 && !this.hideEmptyState
            ? html`<div id="empty" part="empty">${this.emptyLabel.trim()}</div>`
            : null}
        </div>
        <slot part="extras"></slot>
      </section>
    `;
  }

  private _validateConfiguration(): void {
    if (!this.label.trim()) {
      throw new Error('boardgame-component-zone: label must be a non-empty visible and accessible name');
    }
    if (!Number.isSafeInteger(this.headingLevel) || this.headingLevel < 1 || this.headingLevel > 6) {
      throw new Error('boardgame-component-zone: headingLevel must be a safe integer from 1 through 6');
    }
    if (!zoneLayouts.has(this.layout)) {
      throw new Error(`boardgame-component-zone: unknown layout "${this.layout}"; board and spatial layouts use their dedicated components`);
    }
    if (!this.hideEmptyState && !this.emptyLabel.trim()) {
      throw new Error('boardgame-component-zone: emptyLabel must be non-empty unless hideEmptyState is enabled');
    }
  }
}

customElements.define('boardgame-component-zone', BoardgameComponentZone);

declare global {
  interface HTMLElementTagNameMap {
    'boardgame-component-zone': BoardgameComponentZone;
  }
}
