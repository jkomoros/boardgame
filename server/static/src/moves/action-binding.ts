import { noChange } from 'lit';
import {
  AsyncDirective,
  PartType,
  directive,
  type ElementPart,
  type PartInfo,
} from 'lit/async-directive.js';
import { isBoundMoveAction, type BoundMoveAction } from './action.js';

export interface MoveActionBindingOptions {
  /** A creator-owned constraint combined with the action's own availability. */
  readonly disabled?: boolean;
}

type OwnedAttribute = 'disabled' | 'aria-disabled' | 'aria-busy' | 'aria-description' | 'title';
const OWNED_ATTRIBUTES: readonly OwnedAttribute[] = [
  'disabled', 'aria-disabled', 'aria-busy', 'aria-description', 'title',
];

/**
 * Binds a MoveAction to an arbitrary button-like element, including Material
 * Web buttons. Prefer a framework component's `.action` property when one is
 * available because it can render a dedicated visible status region.
 * While connected this directive owns disabled/title/ARIA state; it restores
 * the element's original values on detach. Use the options argument for a
 * creator-owned disabled constraint instead of also binding `disabled`.
 *
 * @example html`<md-filled-button ${bindMoveAction(action)}>Roll</md-filled-button>`
 */
class MoveActionBindingDirective extends AsyncDirective {
  #element: HTMLElement | null = null;
  #action: BoundMoveAction<string, object> | unknown = null;
  #options: MoveActionBindingOptions = {};
  #unsubscribe: (() => void) | null = null;
  #originalAttributes: ReadonlyMap<OwnedAttribute, string | null> | null = null;
  #originalDisabled: boolean | null = null;

  constructor(partInfo: PartInfo) {
    super(partInfo);
    if (partInfo.type !== PartType.ELEMENT) {
      throw new Error('bindMoveAction() must be placed on an element');
    }
  }

  render(
    _action: BoundMoveAction<string, object>,
    _options: MoveActionBindingOptions = {},
  ): typeof noChange {
    return noChange;
  }

  override update(
    part: ElementPart,
    [action, options = {}]: [BoundMoveAction<string, object>, MoveActionBindingOptions?],
  ): typeof noChange {
    if (this.#element !== part.element || this.#action !== action) {
      this.#detach();
      this.#element = part.element as HTMLElement;
      this.#action = action;
      this.#captureOriginalState();
      this.#element.addEventListener('click', this.#activate);
      if (isBoundMoveAction(action)) {
        this.#unsubscribe = action.subscribe(() => this.#apply());
      }
    }
    this.#options = options;
    this.#apply();
    return noChange;
  }

  override disconnected(): void {
    this.#detach();
  }

  override reconnected(): void {
    if (!this.#element) return;
    this.#captureOriginalState();
    this.#element.addEventListener('click', this.#activate);
    if (isBoundMoveAction(this.#action)) {
      this.#unsubscribe = this.#action.subscribe(() => this.#apply());
    }
    this.#apply();
  }

  readonly #activate = (event: Event): void => {
    if (!isBoundMoveAction(this.#action) || !this.#action.canActivate || this.#options.disabled) {
      event.preventDefault();
      event.stopImmediatePropagation();
      return;
    }
    void this.#action.activate();
  };

  #apply(): void {
    const element = this.#element;
    if (!element) return;
    const action = this.#action;
    const bound = isBoundMoveAction(action);
    const disabled = Boolean(this.#options.disabled) || !bound || !action.canActivate;
    const baseReason = this.#options.disabled
      ? 'This action is disabled by the renderer'
      : bound
      ? action.reason?.message
      : action ? 'Bind required move input with .with(...)' : 'No move action supplied';
    const reason = bound && action.preview.kind === 'failed' && action.preview.retryable
      ? `${baseReason ?? 'Move legality check failed'}. Activate to retry.`
      : baseReason;
    if ('disabled' in element) {
      (element as HTMLElement & { disabled: boolean }).disabled = disabled;
    }
    element.toggleAttribute('disabled', disabled);
    element.setAttribute('aria-disabled', String(disabled));
    element.setAttribute('aria-busy', String(bound && action.submission.kind === 'pending'));
    if (reason) {
      element.setAttribute('aria-description', reason);
      element.setAttribute('title', reason);
    } else {
      element.removeAttribute('aria-description');
      element.removeAttribute('title');
    }
  }

  #detach(): void {
    this.#unsubscribe?.();
    this.#unsubscribe = null;
    this.#element?.removeEventListener('click', this.#activate);
    this.#restoreOriginalState();
  }

  #captureOriginalState(): void {
    const element = this.#element;
    if (!element) return;
    this.#originalAttributes = new Map(
      OWNED_ATTRIBUTES.map(attribute => [attribute, element.getAttribute(attribute)]),
    );
    this.#originalDisabled = 'disabled' in element
      ? Boolean((element as HTMLElement & { disabled: boolean }).disabled)
      : null;
  }

  #restoreOriginalState(): void {
    const element = this.#element;
    if (!element || !this.#originalAttributes) return;
    if (this.#originalDisabled !== null && 'disabled' in element) {
      (element as HTMLElement & { disabled: boolean }).disabled = this.#originalDisabled;
    }
    for (const [attribute, value] of this.#originalAttributes) {
      if (value === null) element.removeAttribute(attribute);
      else element.setAttribute(attribute, value);
    }
    this.#originalAttributes = null;
    this.#originalDisabled = null;
  }
}

export const bindMoveAction = directive(MoveActionBindingDirective);
