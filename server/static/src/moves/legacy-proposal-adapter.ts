import {
  creatorMoveInputFromLegacyStrings,
  type MoveInputSchema,
} from './input.js';

/**
 * Quarantines the pre-typed `propose-move`/`data-arg-*` DOM convention.
 * New renderer code should use MoveAction; this exists only so old renderers
 * can cross the typed serialization boundary without bypassing validation.
 */
export class LegacyProposalAdapter {
  readonly #host: HTMLElement;
  readonly #schema: () => MoveInputSchema | null;
  readonly #propose: (moveName: string, nativeArguments: unknown) => void;
  #connected = false;

  constructor(
    host: HTMLElement,
    schema: () => MoveInputSchema | null,
    propose: (moveName: string, nativeArguments: unknown) => void,
  ) {
    this.#host = host;
    this.#schema = schema;
    this.#propose = propose;
  }

  connect(): void {
    if (this.#connected) return;
    this.#connected = true;
    this.#host.addEventListener('click', this.#handle);
    this.#host.addEventListener('component-tapped', this.#handle);
  }

  disconnect(): void {
    if (!this.#connected) return;
    this.#connected = false;
    this.#host.removeEventListener('click', this.#handle);
    this.#host.removeEventListener('component-tapped', this.#handle);
  }

  readonly #handle = (event: Event): void => {
    const proposalElement = event.composedPath().find(candidate => {
      if (!(candidate instanceof Element)) return false;
      return legacyMoveName(candidate) !== null;
    });
    if (!(proposalElement instanceof HTMLElement)) return;
    if (proposalElement.hasAttribute('boardgame-component') && event.type === 'click') return;

    const moveName = legacyMoveName(proposalElement);
    if (!moveName) return;
    const wireArguments: Record<string, string> = {};
    for (const [key, value] of Object.entries(proposalElement.dataset)) {
      if (!key.startsWith('arg') || key === 'arg' || value === undefined) continue;
      wireArguments[key.slice(3)] = value;
    }
    const schema = this.#schema();
    this.#propose(
      moveName,
      schema ? creatorMoveInputFromLegacyStrings(schema, moveName, wireArguments) : wireArguments,
    );
  };
}

function legacyMoveName(element: Element): string | null {
  const property = (element as Element & { proposeMove?: unknown }).proposeMove;
  return (typeof property === 'string' ? property : null) || element.getAttribute('propose-move');
}
