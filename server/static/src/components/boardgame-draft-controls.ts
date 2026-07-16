import { LitElement, css, html, nothing } from 'lit';
import { property } from 'lit/decorators.js';
import { isBoundMoveAction, type BoundMoveAction } from '../moves/action.js';
import './boardgame-action-button.js';

export interface DraftControlsBinding {
  readonly count: number;
  readonly minimumCount: number;
  readonly maximumCount: number;
  readonly status: string;
  readonly action: BoundMoveAction<string, object> | null;
  readonly notice: { readonly message: string } | null;
  readonly canUndo: boolean;
  readonly canClear: boolean;
  undo(): void;
  clear(): void;
  dismissNotice(): void;
}

/** Shared accessible undo, clear, status, and exact-commit UI for local drafts. */
export class BoardgameDraftControls extends LitElement {
  static override styles = css`
    :host { display: block; container-type: inline-size; }
    #controls {
      display: grid;
      gap: var(--boardgame-draft-gap, 0.6rem);
      padding: var(--boardgame-draft-padding, 0.75rem);
      border: var(--boardgame-draft-border, 1px solid color-mix(in srgb, currentColor 18%, transparent));
      border-radius: var(--boardgame-draft-radius, 0.75rem);
      background: var(--boardgame-draft-background, color-mix(in srgb, currentColor 4%, transparent));
    }
    #heading { display: flex; align-items: baseline; justify-content: space-between; gap: 0.75rem; }
    #label { font-weight: 650; }
    #count { font-variant-numeric: tabular-nums; }
    #buttons { display: flex; flex-wrap: wrap; align-items: flex-start; gap: var(--boardgame-draft-button-gap, 0.5rem); }
    button {
      min-width: 44px; min-height: 44px; padding: 0.55rem 0.85rem;
      border: 1px solid currentColor; border-radius: 999px;
      background: transparent; color: inherit; font: inherit; cursor: pointer;
    }
    button:disabled { cursor: not-allowed; opacity: 0.5; }
    button:focus-visible { outline: 3px solid Highlight; outline-offset: 2px; }
    #status { min-height: 1.25em; color: var(--boardgame-draft-status-color, inherit); }
    #notice {
      display: flex; align-items: center; justify-content: space-between; gap: 0.75rem;
      color: var(--boardgame-draft-notice-color, var(--md-sys-color-error, #ba1a1a));
    }
    @container (max-width: 24rem) {
      #buttons { align-items: stretch; flex-direction: column; }
      #buttons > *, #buttons boardgame-action-button { width: 100%; --boardgame-action-width: 100%; }
    }
  `;

  @property({ attribute: false })
  draft: DraftControlsBinding | null = null;

  @property({ type: String }) label = 'Draft';
  @property({ type: String, attribute: 'commit-label' }) commitLabel = 'Commit';
  @property({ type: String, attribute: 'undo-label' }) undoLabel = 'Undo';
  @property({ type: String, attribute: 'clear-label' }) clearLabel = 'Clear';

  override render() {
    const draft = this.#validatedDraft();
    const needed = Math.max(0, draft.minimumCount - draft.count);
    const unboundReason = needed > 0
      ? `Add ${needed} more before committing`
      : 'The current draft cannot be committed';
    return html`
      <section id="controls" part="controls" aria-labelledby="label">
        <div id="heading" part="heading">
          <span id="label" part="label">${this.label.trim()}</span>
          <span id="count" part="count">${draft.count} / ${draft.maximumCount}</span>
        </div>
        <div id="status" part="status" role="status" aria-live="polite">${draft.status}</div>
        ${draft.notice ? html`
          <div id="notice" part="notice" role="status" aria-live="polite">
            <span>${draft.notice.message}</span>
            <button type="button" part="dismiss" @click=${draft.dismissNotice}>Dismiss</button>
          </div>
        ` : nothing}
        <div id="buttons" part="buttons" role="group" aria-label=${`${this.label.trim()} actions`}>
          <button type="button" part="undo" ?disabled=${!draft.canUndo} @click=${draft.undo}>${this.undoLabel.trim()}</button>
          <button type="button" part="clear" ?disabled=${!draft.canClear} @click=${draft.clear}>${this.clearLabel.trim()}</button>
          <boardgame-action-button
            part="commit"
            .action=${draft.action}
            .unboundReason=${unboundReason}>${this.commitLabel.trim()}</boardgame-action-button>
        </div>
      </section>
    `;
  }

  #validatedDraft(): DraftControlsBinding {
    if (!this.label.trim() || !this.commitLabel.trim() || !this.undoLabel.trim() || !this.clearLabel.trim()) {
      throw new Error('boardgame-draft-controls: labels must be non-empty');
    }
    const draft = this.draft;
    if (!draft
      || !Number.isSafeInteger(draft.count) || draft.count < 0
      || !Number.isSafeInteger(draft.minimumCount) || draft.minimumCount < 1
      || !Number.isSafeInteger(draft.maximumCount) || draft.maximumCount < draft.minimumCount
      || draft.maximumCount > 4096
      || draft.count > draft.maximumCount || typeof draft.status !== 'string' || !draft.status.trim()
      || (draft.action !== null && !isBoundMoveAction(draft.action))
      || (draft.notice !== null && (typeof draft.notice?.message !== 'string' || !draft.notice.message.trim()))
      || typeof draft.canUndo !== 'boolean' || typeof draft.canClear !== 'boolean'
      || typeof draft.undo !== 'function' || typeof draft.clear !== 'function'
      || typeof draft.dismissNotice !== 'function') {
      throw new Error('boardgame-draft-controls: .draft must be a placement or selection draft binding');
    }
    return draft;
  }
}

customElements.define('boardgame-draft-controls', BoardgameDraftControls);

declare global {
  interface HTMLElementTagNameMap {
    'boardgame-draft-controls': BoardgameDraftControls;
  }
}
