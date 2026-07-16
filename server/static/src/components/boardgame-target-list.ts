import { LitElement, css, html, nothing } from 'lit';
import { property } from 'lit/decorators.js';
import { isTargetListBinding, type TargetListBinding } from '../moves/target-list.js';
import type { TargetKey } from '../moves/target-action.js';
import './boardgame-action-button.js';

export type TargetListLayout = 'stack' | 'grid';

/** Accessible, preview-aware presentation for a typed target collection. */
export class BoardgameTargetList extends LitElement {
  static override styles = css`
    :host {
      display: block;
      min-width: 0;
      container-type: inline-size;
    }

    #heading {
      margin: 0 0 var(--boardgame-target-list-gap, 0.75rem);
      font: inherit;
      font-weight: 700;
    }

    #heading.visually-hidden {
      position: absolute;
      width: 1px;
      height: 1px;
      padding: 0;
      margin: -1px;
      overflow: hidden;
      clip: rect(0, 0, 0, 0);
      white-space: nowrap;
      border: 0;
    }

    ul {
      display: grid;
      grid-template-columns: 1fr;
      gap: var(--boardgame-target-list-gap, 0.75rem);
      margin: 0;
      padding: 0;
      list-style: none;
    }

    :host([layout='grid']) ul {
      grid-template-columns: repeat(
        auto-fit,
        minmax(min(100%, var(--boardgame-target-list-min-width, 12rem)), 1fr)
      );
    }

    boardgame-action-button {
      display: block;
      width: 100%;
      --boardgame-action-width: 100%;
    }

    #empty {
      box-sizing: border-box;
      padding: 1rem;
      border: 1px dashed color-mix(in srgb, currentColor 28%, transparent);
      border-radius: 0.75rem;
      color: color-mix(in srgb, currentColor 62%, transparent);
      font-style: italic;
    }
  `;

  @property({ attribute: false })
  choices: TargetListBinding<TargetKey> | null = null;

  @property({ type: String })
  label = 'Choices';

  @property({ type: Number, attribute: 'heading-level' })
  headingLevel = 2;

  @property({ type: Boolean, attribute: 'hide-heading' })
  hideHeading = false;

  @property({ type: String, attribute: 'empty-label' })
  emptyLabel = 'No choices available';

  @property({ type: String, reflect: true })
  layout: TargetListLayout = 'stack';

  #unsubscribe: (() => void) | null = null;

  override connectedCallback(): void {
    super.connectedCallback();
    this.#subscribeToTarget();
  }

  protected override updated(changedProperties: Map<PropertyKey, unknown>): void {
    super.updated(changedProperties);
    if (!changedProperties.has('choices')) return;
    const previous = changedProperties.get('choices');
    const previousTarget = isTargetListBinding(previous) ? previous.target : null;
    const nextTarget = isTargetListBinding(this.choices) ? this.choices.target : null;
    if (previousTarget === nextTarget && this.#unsubscribe) return;
    this.#unsubscribe?.();
    this.#unsubscribe = null;
    if (nextTarget) this.#subscribeToTarget();
  }

  override disconnectedCallback(): void {
    this.#unsubscribe?.();
    this.#unsubscribe = null;
    super.disconnectedCallback();
  }

  #subscribeToTarget(): void {
    if (!this.isConnected || this.#unsubscribe || !isTargetListBinding(this.choices)) return;
    this.#unsubscribe = this.choices.target.subscribe(() => this.requestUpdate());
  }

  override render() {
    this.#validate();
    const binding = this.choices!;
    return html`
      <section part="region" aria-labelledby="heading" aria-busy=${String(binding.target.preview.kind === 'checking')}>
        <div
          id="heading"
          part="heading"
          class=${this.hideHeading ? 'visually-hidden' : ''}
          role="heading"
          aria-level=${this.headingLevel}>
          ${this.label.trim()}
        </div>
        ${binding.choices.length ? html`
          <ul part="list">
            ${binding.choices.map(choice => html`
              <li part="choice">
                <boardgame-action-button .action=${choice.action}>${choice.label}</boardgame-action-button>
              </li>
            `)}
          </ul>
        ` : html`<div id="empty" part="empty">${this.emptyLabel.trim()}</div>`}
        ${binding.target.preview.kind === 'failed'
          ? html`<div part="status" role="status" aria-live="polite">${binding.target.preview.reason.message}</div>`
          : nothing}
      </section>
    `;
  }

  #validate(): void {
    if (!isTargetListBinding(this.choices)) {
      throw new Error('boardgame-target-list: .choices must come from targetList(move(...).targets(...), labelFor)');
    }
    if (!this.label.trim()) throw new Error('boardgame-target-list: label must be a non-empty choice collection name');
    if (!Number.isSafeInteger(this.headingLevel) || this.headingLevel < 1 || this.headingLevel > 6) {
      throw new Error('boardgame-target-list: headingLevel must be a safe integer from 1 through 6');
    }
    if (!this.emptyLabel.trim()) throw new Error('boardgame-target-list: emptyLabel must be non-empty');
    if (this.layout !== 'stack' && this.layout !== 'grid') {
      throw new Error(`boardgame-target-list: layout must be "stack" or "grid", not ${JSON.stringify(this.layout)}`);
    }
  }
}

customElements.define('boardgame-target-list', BoardgameTargetList);

declare global {
  interface HTMLElementTagNameMap {
    'boardgame-target-list': BoardgameTargetList;
  }
}
