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
      this.#attributeObserver = new MutationObserver(() => this.#notify());
      this.#attributeObserver.observe(this.#host, {
        attributes: true,
        attributeFilter: ['hidden', 'aria-hidden', 'class', 'style'],
      });
    }
    queueMicrotask(() => this.#notify());
  }

  hostUpdated(): void {
    this.#notify();
  }

  hostDisconnected(): void {
    this.#resizeObserver?.disconnect();
    this.#resizeObserver = null;
    this.#attributeObserver?.disconnect();
    this.#attributeObserver = null;
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
    if (!this.#host.isConnected || this.#host.hidden
      || this.#host.closest('[hidden], [aria-hidden="true"]')) return false;
    const style = getComputedStyle(this.#host);
    return style.display !== 'none' && style.visibility !== 'hidden'
      && this.#host.getClientRects().length > 0;
  }
}

export function isProjectedChoiceConsumptionController(
  value: unknown,
): value is ProjectedChoiceConsumptionController {
  return typeof value === 'object' && value !== null && controllers.has(value);
}
