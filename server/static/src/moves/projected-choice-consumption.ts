import type { ReactiveController, ReactiveControllerHost } from 'lit';
import type { ProjectedMoveChoiceSet } from './projected-choices.js';

export const PROJECTED_CHOICE_CONSUMPTION_CHANGED = 'projected-choice-consumption-changed';

type AnyProjectedChoiceSet = ProjectedMoveChoiceSet<string, {
  readonly field: string;
  readonly value: string | number;
  readonly input: object;
}>;

const controllers = new WeakSet<object>();

export interface ProjectedChoiceConsumptionChangedDetail {
  readonly controller: ProjectedChoiceConsumptionController;
}

/**
 * Internal lifecycle proof used by framework-owned native controls. A claim is
 * evaluated live; registration alone never suppresses the generic fallback.
 */
export class ProjectedChoiceConsumptionController implements ReactiveController {
  readonly #host: ReactiveControllerHost & HTMLElement;
  readonly #currentSet: () => AnyProjectedChoiceSet | null;
  #resizeObserver: ResizeObserver | null = null;
  #attributeObserver: MutationObserver | null = null;
  readonly #observedSlots = new Set<HTMLSlotElement>();
  #invalidate: (() => void) | null = null;

  constructor(
    host: ReactiveControllerHost & HTMLElement,
    currentSet: () => AnyProjectedChoiceSet | null,
  ) {
    this.#host = host;
    this.#currentSet = currentSet;
    controllers.add(this);
    host.addController(this);
  }

  get hostElement(): HTMLElement {
    return this.#host;
  }

  get consumedSet(): AnyProjectedChoiceSet | null {
    if (!this.#isRendered()) return null;
    return this.#currentSet();
  }

  setInvalidator(invalidate: (() => void) | null): void {
    this.#invalidate = invalidate;
  }

  hostConnected(): void {
    if (typeof ResizeObserver === 'function') {
      this.#resizeObserver = new ResizeObserver(() => this.#notify());
      this.#resizeObserver.observe(this.#host);
    }
    if (typeof MutationObserver === 'function') {
      this.#attributeObserver = new MutationObserver(() => {
        this.#refreshVisibilityObservers();
        this.#notify();
      });
      this.#refreshVisibilityObservers();
    }
    queueMicrotask(() => this.#notify());
  }

  hostUpdated(): void {
    this.#refreshVisibilityObservers();
    this.#notify();
  }

  hostDisconnected(): void {
    this.#resizeObserver?.disconnect();
    this.#resizeObserver = null;
    this.#attributeObserver?.disconnect();
    this.#attributeObserver = null;
    this.#clearSlotObservers();
    this.#invalidate?.();
    this.#invalidate = null;
  }

  #notify(): void {
    if (!this.#host.isConnected) return;
    this.#host.dispatchEvent(new CustomEvent<ProjectedChoiceConsumptionChangedDetail>(
      PROJECTED_CHOICE_CONSUMPTION_CHANGED,
      { bubbles: true, composed: true, detail: { controller: this } },
    ));
  }

  #isRendered(): boolean {
    if (!this.#host.isConnected || this.#host.getClientRects().length === 0) return false;
    let element: Element | null = this.#host;
    while (element) {
      if (element.hasAttribute('hidden') || element.getAttribute('aria-hidden') === 'true') return false;
      const style = getComputedStyle(element);
      if (style.display === 'none' || style.visibility === 'hidden') return false;
      element = this.#composedParent(element);
    }
    return true;
  }

  #refreshVisibilityObservers(): void {
    if (!this.#attributeObserver) return;
    this.#attributeObserver.disconnect();
    this.#clearSlotObservers();
    let element: Element | null = this.#host;
    while (element) {
      this.#attributeObserver.observe(element, {
        attributes: true,
        attributeFilter: ['hidden', 'aria-hidden', 'class', 'style', 'inert'],
      });
      const assignedSlot = element.assignedSlot;
      if (assignedSlot && !this.#observedSlots.has(assignedSlot)) {
        assignedSlot.addEventListener('slotchange', this.#slotChanged);
        this.#observedSlots.add(assignedSlot);
      }
      element = this.#composedParent(element);
    }
  }

  #clearSlotObservers(): void {
    for (const slot of this.#observedSlots) {
      slot.removeEventListener('slotchange', this.#slotChanged);
    }
    this.#observedSlots.clear();
  }

  readonly #slotChanged = (): void => {
    this.#refreshVisibilityObservers();
    this.#notify();
  };

  #composedParent(element: Element): Element | null {
    if (element.assignedSlot) return element.assignedSlot;
    if (element.parentElement) return element.parentElement;
    const root = element.getRootNode();
    return root instanceof ShadowRoot ? root.host : null;
  }
}

export function isProjectedChoiceConsumptionController(
  value: unknown,
): value is ProjectedChoiceConsumptionController {
  return typeof value === 'object' && value !== null && controllers.has(value);
}
